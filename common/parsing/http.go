package parsing

import (
	"errors"
	"log/slog"
	"net/url"
	"path"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"

	"github.com/ExplorViz/otel-collector/common/attrib"
)

const HTTPServerEntityType string = "httpserver"

// When an HTTP request is identified as fetching a static resource
// like an image or a style sheet as opposed to an API endpoint,
// then this name will be used as a catch-all endpoint for such requests.
// Otherwise, a lot of entities would be produced due to potentially high
// cardinality of resources in more complex applications.
const HTTPStaticResourceRoute = "Static Resource"

// An HTTPServerEntity represents a specific request route for a server's HTTP API.
type HTTPServerEntity struct {
	// Name of the service or application to which this API endpoint belongs.
	ServiceName string

	// Name of the instrumentation scope to which the entity belongs
	ScopeName string

	// The matched route template of the received request's URL path.
	// Dynamic segments in the path should be represented by placeholders.
	Route string

	// The HTTP request method type, e.g. "GET".
	Method string
}

func (h HTTPServerEntity) ID() string {
	return "httpserver" + "|" + h.ServiceName + "|" + h.ScopeName + "|" + h.Method + "|" + h.Route
}

func (h HTTPServerEntity) TelemetryKey() string {
	return "httpserver" + "|" + h.ServiceName + "|" + h.ScopeName + "|" + h.Route
}

func (h HTTPServerEntity) ToAttributes(attrs *pcommon.Map) {
	attrs.PutStr(string(attrib.ExplorVizAttributes.EntityType.Key), HTTPServerEntityType)
	attrs.PutStr(string(attrib.ExplorVizAttributes.ServiceName.Key), h.ServiceName)
	attrs.PutStr(string(attrib.ExplorVizAttributes.ScopeName.Key), h.ScopeName)
	attrs.PutStr(string(attrib.ExplorVizAttributes.HTTPRoute.Key), h.Route)
	attrs.PutStr(string(attrib.ExplorVizAttributes.HTTPMethod.Key), h.Method)
}

// httpServerEntityFromAttribs initializes a new [HTTPServerEntity] based on the entries of the provided map.
// If the provided map entries are incomplete (meaning a mandatory attribute is missing),
// then a zero-initialized HTTPServerEntity and an error is returned.
func httpServerEntityFromAttribs(m pcommon.Map) (HTTPServerEntity, error) {
	service, ok := m.Get(string(attrib.ExplorVizAttributes.ServiceName.Key))
	if !ok || service.Str() == "" {
		return HTTPServerEntity{}, errors.New("empty or missing string attribute for service name")
	}

	scope, ok := m.Get(string(attrib.ExplorVizAttributes.ScopeName.Key))
	if !ok || scope.Str() == "" {
		return HTTPServerEntity{}, errors.New("empty or missing string attribute for scope name")
	}

	route, ok := m.Get(string(attrib.ExplorVizAttributes.HTTPRoute.Key))
	if !ok || route.Str() == "" {
		return HTTPServerEntity{}, errors.New("empty or missing string attribute for http route")
	}

	method, _ := m.Get(string(attrib.ExplorVizAttributes.HTTPMethod.Key))

	return HTTPServerEntity{
		ServiceName: service.Str(),
		ScopeName:   scope.Str(),
		Route:       route.Str(),
		Method:      method.Str(),
	}, nil
}

const HTTPClientEntityType string = "httpclient"

// An HTTPClientEntity represents a client making requests to an HTTP API.
type HTTPClientEntity struct {
	// Name of the service or application to which this HTTP client belongs.
	ServiceName string

	// Name of the instrumentation scope to which the entity belongs
	ScopeName string
}

func (h HTTPClientEntity) ID() string {
	return "httpclient" + "|" + h.ServiceName + "|" + h.ScopeName
}

func (h HTTPClientEntity) TelemetryKey() string {
	return h.ID()
}

func (h HTTPClientEntity) ToAttributes(attrs *pcommon.Map) {
	attrs.PutStr(string(attrib.ExplorVizAttributes.EntityType.Key), HTTPClientEntityType)
	attrs.PutStr(string(attrib.ExplorVizAttributes.ServiceName.Key), h.ServiceName)
	attrs.PutStr(string(attrib.ExplorVizAttributes.ScopeName.Key), h.ScopeName)
}

// httpClientEntityFromAttribs initializes a new [HTTPClientEntity] based on the entries of the provided map.
// If the provided map entries are incomplete (meaning a mandatory attribute is missing),
// then a zero-initialized HTTPClientEntity and an error is returned.
func httpClientEntityFromAttribs(m pcommon.Map) (HTTPClientEntity, error) {
	service, ok := m.Get(string(attrib.ExplorVizAttributes.ServiceName.Key))
	if !ok || service.Str() == "" {
		return HTTPClientEntity{}, errors.New("empty or missing string attribute for service name")
	}

	scope, ok := m.Get(string(attrib.ExplorVizAttributes.ScopeName.Key))
	if !ok || scope.Str() == "" {
		return HTTPClientEntity{}, errors.New("empty or missing string attribute for scope name")
	}

	return HTTPClientEntity{
		ServiceName: service.Str(),
		ScopeName:   scope.Str(),
	}, nil
}

// ParseHTTPTelemetry parses telemetry describing HTTP / REST API calls by looking for attributes conforming
// to the [OTel semconv HTTP attributes]. For spans of kind [ptrace.SpanKindClient], an [HTTPClientEntity] will
// be extracted. Otherwise, an [HTTPServerEntity] is parsed.
//
// # Server entities
//
// For telemetry to be successfully parsed as a server entity, it needs to provide:
//   - a service name to which to match the API endpoint
//   - a description of the route path template for the requested endpoint (e.g. "/users/{userId}/profile")
//
// # Client entities
//
// For telemetry to be successfully parsed as a client entity, it needs to provide:
//   - a service name to which to match the HTTP client
//   - an instrumentation scope name
//   - attributes from the [OTel semconv HTTP attributes] indicating it represents an HTTP request
//
// [OTel semconv HTTP attributes]: https://opentelemetry.io/docs/specs/semconv/registry/attributes/http/
func ParseHTTPTelemetry(tr attrib.TelemetryReader) (Entity, error) {
	if tr.Span != nil && tr.Span.Kind() == ptrace.SpanKindClient {
		slog.Debug("http parser: encountered client span, attempting to parse as client")
		return parseHTTPClientTelemetry(tr)
	}
	return parseHTTPServerTelemetry(tr)
}

func parseHTTPServerTelemetry(tr attrib.TelemetryReader) (HTTPServerEntity, error) {
	if !isHTTPRequest(tr) {
		return HTTPServerEntity{}, errors.New("http server parser: no identifying http attributes found")
	}

	service := tr.ResourceStrAttrib(semconv.ServiceNameKey)
	if service == "" {
		return HTTPServerEntity{}, errors.New("http server parser: empty or missing service name attribute")
	}

	route := tr.StrAttrib(semconv.HTTPRouteKey)
	if route == "" {
		route = tr.StrAttrib(semconv.URLTemplateKey)
	}
	if route == "" {
		// Attempt to derive route from URL path instead
		urlPath := tr.StrAttrib(semconv.URLPathKey)
		if urlPath == "" {
			urlPath, _ = pathFromFullUrl(tr.StrAttrib(semconv.URLFullKey))
		}
		if urlPath == "" {
			// Use deprecated full URL attribute as fallback
			urlPath, _ = pathFromFullUrl(tr.StrAttrib("http.url"))
		}
		if urlPath == "" && tr.StrAttrib("http.target") != "" {
			// Use deprecated attribute giving the path + query portions of the URL as fallback
			urlPath, _ = pathFromFullUrl("http://example.com" + tr.StrAttrib("http.target"))
		}

		if path.Ext(urlPath) != "" {
			// A file extension at the end means this is likely a static resource, not an API endpoint.
			// Since these paths might be high cardinality, we group them under a single route
			route = HTTPStaticResourceRoute
		} else {
			route = urlPath
		}
	}
	if route == "" {
		return HTTPServerEntity{}, errors.New("http server parser: empty or missing route attribute")
	}

	method := tr.StrAttrib(semconv.HTTPRequestMethodKey)
	if method == "" {
		// Use deprecated attribute key as fallback
		method = tr.StrAttrib("http.method")
	}

	scope := tr.Scope.Name()

	return HTTPServerEntity{
		ServiceName: service,
		ScopeName:   scope,
		Route:       route,
		Method:      method,
	}, nil
}

func parseHTTPClientTelemetry(tr attrib.TelemetryReader) (HTTPClientEntity, error) {
	if !isHTTPRequest(tr) {
		return HTTPClientEntity{}, errors.New("http client parser: no identifying http attributes found")
	}

	service := tr.ResourceStrAttrib(semconv.ServiceNameKey)
	if service == "" {
		return HTTPClientEntity{}, errors.New("http client parser: empty or missing service name attribute")
	}

	scope := tr.Scope.Name()

	return HTTPClientEntity{ServiceName: service, ScopeName: scope}, nil
}

func isHTTPRequest(tr attrib.TelemetryReader) bool {
	return tr.HasAnyAttrKey([]attribute.Key{
		semconv.HTTPConnectionStateKey,
		semconv.HTTPRequestBodySizeKey,
		semconv.HTTPRequestMethodKey,
		semconv.HTTPRequestSizeKey,
		semconv.HTTPRouteKey,

		// Deprecated attributes
		attribute.Key("http.flavor"),
		attribute.Key("http.method"),
		attribute.Key("http.protocol"),
		attribute.Key("http.server_name"),
		attribute.Key("http.status_code"),
		attribute.Key("http.target"),
		attribute.Key("http.url"),
	})
}

func pathFromFullUrl(fullUrl string) (string, error) {
	if fullUrl == "" {
		return "", errors.New("received empty string as full url")
	}

	parsed, err := url.Parse(fullUrl)
	if err != nil {
		return "", err
	}

	if parsed.Path == "" {
		return "/", nil
	}

	return parsed.Path, nil
}

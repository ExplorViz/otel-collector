package parsing

import (
	"errors"

	"go.opentelemetry.io/collector/pdata/pcommon"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"

	"github.com/ExplorViz/otel-collector/common/attrib"
)

const HTTPEntityType string = "http"

// An HTTPEntity represents a specific request route for an HTTP API.
type HTTPEntity struct {
	// Name of the service or application to which this API endpoint belongs.
	ServiceName string

	// The matched route template of the request's URL path.
	// Dynamic segments in the path should be represented by placeholders.
	Route string

	// The HTTP request method type, e.g. "GET".
	Method string
}

func (h HTTPEntity) ID() string {
	return "http" + "|" + h.ServiceName + "|" + h.Method + "|" + h.Route
}

func (h HTTPEntity) TelemetryKey() string {
	return "http" + "|" + h.ServiceName + "|" + h.Route
}

func (h HTTPEntity) ToAttributes(attrs *pcommon.Map) {
	attrs.PutStr(string(attrib.ExplorVizAttributes.EntityType.Key), HTTPEntityType)
	attrs.PutStr(string(attrib.ExplorVizAttributes.ServiceName.Key), h.ServiceName)
	attrs.PutStr(string(attrib.ExplorVizAttributes.HTTPRoute.Key), h.Route)
	attrs.PutStr(string(attrib.ExplorVizAttributes.HTTPMethod.Key), h.Method)
}

// httpEntityFromAttribs initializes a new [HTTPEntity] based on the entries of the provided map.
// If the provided map entries are incomplete (meaning a mandatory attribute is missing),
// then a zero-initialized HTTPEntity and an error is returned.
func httpEntityFromAttribs(m pcommon.Map) (HTTPEntity, error) {
	service, ok := m.Get(string(attrib.ExplorVizAttributes.ServiceName.Key))
	if !ok || service.Str() == "" {
		return HTTPEntity{}, errors.New("empty or missing string attribute for service name")
	}

	route, ok := m.Get(string(attrib.ExplorVizAttributes.HTTPRoute.Key))
	if !ok || route.Str() == "" {
		return HTTPEntity{}, errors.New("empty or missing string attribute for http route")
	}

	method, _ := m.Get(string(attrib.ExplorVizAttributes.HTTPMethod.Key))

	return HTTPEntity{
		Route:  route.Str(),
		Method: method.Str(),
	}, nil
}

// ParseHTTPTelemetry parses telemetry describing HTTP / REST API calls by looking for attributes conforming
// to the [OTel semconv HTTP attributes]. For telemetry to be successfully parsed, it needs to provide:
//   - a service name to which to match the API endpoint
//   - a description of the route path template for the requested endpoint (e.g. "/users/{userId}/profile")
//
// [OTel semconv HTTP attributes]: https://opentelemetry.io/docs/specs/semconv/registry/attributes/http/
func ParseHTTPTelemetry(tr attrib.TelemetryReader) (Entity, error) {
	service := tr.ResourceStrAttrib(semconv.ServiceNameKey)
	if service == "" {
		return HTTPEntity{}, errors.New("http parser: empty or missing service name attribute")
	}

	route := tr.StrAttrib(semconv.HTTPRouteKey)
	if route == "" {
		route = tr.StrAttrib(semconv.URLTemplateKey)
	}
	if route == "" {
		return HTTPEntity{}, errors.New("http parser: empty or missing route attribute")
	}

	method := tr.StrAttrib(semconv.HTTPRequestMethodKey)

	return HTTPEntity{
		ServiceName: service,
		Route:       route,
		Method:      method,
	}, nil
}

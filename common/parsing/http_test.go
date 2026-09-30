package parsing

import (
	"testing"

	"go.opentelemetry.io/collector/pdata/pcommon"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"

	"github.com/stretchr/testify/assert"

	"github.com/ExplorViz/otel-collector/common/attrib"
)

func exampleResource() *pcommon.Resource {
	r := pcommon.NewResource()
	r.Attributes().PutStr(string(semconv.ServiceNameKey), "example-service")
	return &r
}

func exampleScope() *pcommon.InstrumentationScope {
	sc := pcommon.NewInstrumentationScope()
	sc.SetName("example-scope")
	return &sc
}

func exampleTelemetryReader() attrib.TelemetryReader {
	attrs := pcommon.NewMap()
	tr := attrib.TelemetryReader{
		Resource: exampleResource(),
		Scope:    exampleScope(),
		Attrs:    &attrs,
	}
	return tr
}

func TestParseHTTPServerTelemetry(t *testing.T) {
	tr := exampleTelemetryReader()
	tr.Attrs.PutStr("http.method", "GET")
	tr.Attrs.PutStr("http.route", "/api/products/[productId]")

	entity, err := parseHTTPServerTelemetry(tr)

	assert.Nil(t, err)
	assert.Equal(t, "example-service", entity.ServiceName)
	assert.Equal(t, "example-scope", entity.ScopeName)
	assert.Equal(t, "GET", entity.Method)
	assert.Equal(t, "/api/products/[productId]", entity.Route)
	assert.NotEmpty(t, entity.TelemetryKey())
}

func TestParseHTTPServerTelemetryNoMethod(t *testing.T) {
	tr := exampleTelemetryReader()
	tr.Attrs.PutStr("http.route", "/api/products/[productId]")

	entity, err := parseHTTPServerTelemetry(tr)

	assert.Nil(t, err)
	assert.Equal(t, "example-service", entity.ServiceName)
	assert.Equal(t, "example-scope", entity.ScopeName)
	assert.Equal(t, "", entity.Method)
	assert.Equal(t, "/api/products/[productId]", entity.Route)
	assert.NotEmpty(t, entity.TelemetryKey())
}

func TestParseHTTPServerTelemetryURLTemplate(t *testing.T) {
	tr := exampleTelemetryReader()
	tr.Attrs.PutStr("http.method", "GET")
	tr.Attrs.PutStr("url.template", "/api/products/[productId]")

	entity, err := parseHTTPServerTelemetry(tr)

	assert.Nil(t, err)
	assert.Equal(t, "example-service", entity.ServiceName)
	assert.Equal(t, "example-scope", entity.ScopeName)
	assert.Equal(t, "GET", entity.Method)
	assert.Equal(t, "/api/products/[productId]", entity.Route)
	assert.NotEmpty(t, entity.TelemetryKey())
}

func TestParseHTTPServerTelemetryURLPath(t *testing.T) {
	tr := exampleTelemetryReader()
	tr.Attrs.PutStr("http.method", "GET")
	tr.Attrs.PutStr("url.path", "/api/products/product123")

	entity, err := parseHTTPServerTelemetry(tr)

	assert.Nil(t, err)
	assert.Equal(t, "example-service", entity.ServiceName)
	assert.Equal(t, "example-scope", entity.ScopeName)
	assert.Equal(t, "GET", entity.Method)
	assert.Equal(t, "/api/products/product123", entity.Route)
	assert.NotEmpty(t, entity.TelemetryKey())

	tr.Attrs.PutStr("url.path", "/productImages/telescope.jpg")

	entity, err = parseHTTPServerTelemetry(tr)

	assert.Nil(t, err)
	assert.Equal(t, HTTPStaticResourceRoute, entity.Route)
}

func TestParseHTTPServerTelemetryURLFull(t *testing.T) {
	tr := exampleTelemetryReader()
	tr.Attrs.PutStr("http.method", "GET")
	tr.Attrs.PutStr("url.full", "http://localhost:8080/api/products/product123?q=somequerytoignore")

	entity, err := parseHTTPServerTelemetry(tr)

	assert.Nil(t, err)
	assert.Equal(t, "example-service", entity.ServiceName)
	assert.Equal(t, "example-scope", entity.ScopeName)
	assert.Equal(t, "GET", entity.Method)
	assert.Equal(t, "/api/products/product123", entity.Route)
	assert.NotEmpty(t, entity.TelemetryKey())

	tr.Attrs.PutStr("url.full", "http://localhost:8080/productImages/telescope.jpg?q=somequerytoignore")

	entity, err = parseHTTPServerTelemetry(tr)

	assert.Nil(t, err)
	assert.Equal(t, HTTPStaticResourceRoute, entity.Route)
}

func TestParseHTTPServerTelemetryDeprecatedHTTPUrl(t *testing.T) {
	tr := exampleTelemetryReader()
	tr.Attrs.PutStr("http.method", "GET")
	tr.Attrs.PutStr("http.url", "http://localhost:8080/api/products/product123?q=somequerytoignore")

	entity, err := parseHTTPServerTelemetry(tr)

	assert.Nil(t, err)
	assert.Equal(t, "example-service", entity.ServiceName)
	assert.Equal(t, "example-scope", entity.ScopeName)
	assert.Equal(t, "GET", entity.Method)
	assert.Equal(t, "/api/products/product123", entity.Route)
	assert.NotEmpty(t, entity.TelemetryKey())

	tr.Attrs.PutStr("http.url", "http://localhost:8080/productImages/telescope.jpg?q=somequerytoignore")

	entity, err = parseHTTPServerTelemetry(tr)

	assert.Nil(t, err)
	assert.Equal(t, HTTPStaticResourceRoute, entity.Route)
}

func TestParseHTTPServerTelemetryDeprecatedHTTPTarget(t *testing.T) {
	tr := exampleTelemetryReader()
	tr.Attrs.PutStr("http.method", "GET")
	tr.Attrs.PutStr("http.target", "/api/products/product123?q=somequerytoignore")

	entity, err := parseHTTPServerTelemetry(tr)

	assert.Nil(t, err)
	assert.Equal(t, "example-service", entity.ServiceName)
	assert.Equal(t, "example-scope", entity.ScopeName)
	assert.Equal(t, "GET", entity.Method)
	assert.Equal(t, "/api/products/product123", entity.Route)
	assert.NotEmpty(t, entity.TelemetryKey())

	tr.Attrs.PutStr("http.url", "/productImages/telescope.jpg?q=somequerytoignore")

	entity, err = parseHTTPServerTelemetry(tr)

	assert.Nil(t, err)
	assert.Equal(t, HTTPStaticResourceRoute, entity.Route)
}

func TestParseHTTPServerTelemetryMissingServiceName(t *testing.T) {
	tr := exampleTelemetryReader()
	tr.Resource.Attributes().Remove(string(semconv.ServiceNameKey))

	_, err := parseHTTPServerTelemetry(tr)

	assert.NotNil(t, err)
}

func TestParseHTTPServerTelemetryMissingRoute(t *testing.T) {
	tr := exampleTelemetryReader()
	tr.Attrs.PutStr("http.method", "GET")

	_, err := parseHTTPServerTelemetry(tr)

	assert.NotNil(t, err)
}

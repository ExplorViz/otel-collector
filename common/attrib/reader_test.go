package attrib

import (
	"testing"

	"go.opentelemetry.io/collector/pdata/pcommon"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	"github.com/stretchr/testify/assert"
)

const defaultTokenID string = "mytokenvalue"
const defaultTokenSecret string = "mytokensecret"

const defaultServiceName string = "trace-service"
const defaultFunctionFQN string = "net.explorviz.example.MyClass.myMethod"

func defaultTestResource() pcommon.Resource {
	r := pcommon.NewResource()
	r.Attributes().PutStr(string(ExplorVizAttributes.LandscapeTokenID.Key), defaultTokenID)
	r.Attributes().PutStr(string(ExplorVizAttributes.LandscapeTokenSecret.Key), defaultTokenSecret)
	r.Attributes().PutStr(string(semconv.ServiceNameKey), defaultServiceName)
	return r
}

func TestTokenID(t *testing.T) {
	attrs := pcommon.NewMap()
	sc := pcommon.NewInstrumentationScope()
	rs := defaultTestResource()
	tr := TelemetryReader{Attrs: &attrs, Scope: &sc, Resource: &rs}

	id := tr.LandscapeTokenID()
	assert.Equal(t, defaultTokenID, id)
}

func TestTokenIDInTelemetry(t *testing.T) {
	attrs := pcommon.NewMap()
	sc := pcommon.NewInstrumentationScope()
	rs := pcommon.NewResource()
	attrs.PutStr(string(ExplorVizAttributes.LandscapeTokenID.Key), defaultTokenID)
	tr := TelemetryReader{Attrs: &attrs, Scope: &sc, Resource: &rs}

	id := tr.LandscapeTokenID()
	assert.Equal(t, defaultTokenID, id)
}

func TestTokenIDMissing(t *testing.T) {
	attrs := pcommon.NewMap()
	sc := pcommon.NewInstrumentationScope()
	rs := pcommon.NewResource()
	tr := TelemetryReader{Attrs: &attrs, Scope: &sc, Resource: &rs}

	id := tr.LandscapeTokenID()
	assert.Empty(t, id)
}

func TestTokenSecret(t *testing.T) {
	attrs := pcommon.NewMap()
	sc := pcommon.NewInstrumentationScope()
	rs := defaultTestResource()
	tr := TelemetryReader{Attrs: &attrs, Scope: &sc, Resource: &rs}

	sec := tr.LandscapeTokenSecret()
	assert.Equal(t, defaultTokenSecret, sec)
}

func TestTokenSecretInTelemetry(t *testing.T) {
	attrs := pcommon.NewMap()
	sc := pcommon.NewInstrumentationScope()
	rs := pcommon.NewResource()
	attrs.PutStr(string(ExplorVizAttributes.LandscapeTokenID.Key), defaultTokenID)
	attrs.PutStr(string(ExplorVizAttributes.LandscapeTokenSecret.Key), defaultTokenSecret)
	tr := TelemetryReader{Attrs: &attrs, Scope: &sc, Resource: &rs}

	sec := tr.LandscapeTokenSecret()
	assert.Equal(t, defaultTokenSecret, sec)
}

func TestTokenSecretIDMissing(t *testing.T) {
	attrs := pcommon.NewMap()
	sc := pcommon.NewInstrumentationScope()
	rs := pcommon.NewResource()
	tr := TelemetryReader{Attrs: &attrs, Scope: &sc, Resource: &rs}

	sec := tr.LandscapeTokenSecret()
	assert.Empty(t, sec)
}

func TestStrAttrib(t *testing.T) {
	attrs := pcommon.NewMap()
	sc := pcommon.NewInstrumentationScope()
	rs := pcommon.NewResource()
	attrs.PutStr(string(semconv.CodeFunctionNameKey), defaultFunctionFQN)
	tr := TelemetryReader{Attrs: &attrs, Scope: &sc, Resource: &rs}

	fqn := tr.StrAttrib(semconv.CodeFunctionNameKey)
	assert.Equal(t, defaultFunctionFQN, fqn)
}

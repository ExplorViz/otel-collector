package trace

import (
	"encoding/hex"
	"testing"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	"github.com/stretchr/testify/assert"

	"github.com/ExplorViz/otel-collector/common/attrib"
)

const defaultSpanID string = "5fb397be34d26b51"
const defaultTraceID string = "5b8aa5a2d2c872e8321cf37308d69df2"
const defaultParentID string = "051581bf3cb55c13"
const defaultSpanName string = "hello-greetings"

const defaultTokenID string = "mytokenvalue"
const defaultTokenSecret string = "mytokensecret"

const defaultServiceName string = "trace-service"
const defaultFunctionFQN string = "net.explorviz.example.MyClass.myMethod"

func defaultTestSpan() ptrace.Span {
	bSpanID, _ := hex.DecodeString(defaultSpanID)
	bTraceID, _ := hex.DecodeString(defaultTraceID)
	bParentID, _ := hex.DecodeString(defaultParentID)

	s := ptrace.NewSpan()
	s.SetSpanID(pcommon.SpanID(bSpanID))
	s.SetTraceID(pcommon.TraceID(bTraceID))
	s.SetParentSpanID(pcommon.SpanID(bParentID))
	s.SetName(defaultSpanName)
	s.SetStartTimestamp(0)
	s.SetEndTimestamp(10)
	s.Attributes().PutStr(string(semconv.CodeFunctionNameKey), defaultFunctionFQN)

	return s
}

func defaultTestResource() pcommon.Resource {
	r := pcommon.NewResource()
	r.Attributes().PutStr(string(attrib.ExplorVizAttributes.LandscapeTokenID.Key), defaultTokenID)
	r.Attributes().PutStr(string(attrib.ExplorVizAttributes.LandscapeTokenSecret.Key), defaultTokenSecret)
	r.Attributes().PutStr(string(semconv.ServiceNameKey), defaultServiceName)

	return r
}

func TestSpanID(t *testing.T) {
	s := defaultTestSpan()
	sc := pcommon.NewInstrumentationScope()
	rs := pcommon.NewResource()
	sr := SpanReader{Span: &s, Scope: &sc, Resource: &rs}

	spanID := sr.SpanID()
	assert.Equal(t, defaultSpanID, spanID)
}

func TestSpanIDEmpty(t *testing.T) {
	s := ptrace.NewSpan()
	sc := pcommon.NewInstrumentationScope()
	rs := pcommon.NewResource()
	sr := SpanReader{Span: &s, Scope: &sc, Resource: &rs}

	spanID := sr.SpanID()
	assert.Empty(t, spanID)
}

func TestTraceID(t *testing.T) {
	s := defaultTestSpan()
	sc := pcommon.NewInstrumentationScope()
	rs := pcommon.NewResource()
	sr := SpanReader{Span: &s, Scope: &sc, Resource: &rs}

	traceID := sr.TraceID()
	assert.Equal(t, defaultTraceID, traceID)
}

func TestTraceIDEmpty(t *testing.T) {
	s := ptrace.NewSpan()
	sc := pcommon.NewInstrumentationScope()
	rs := pcommon.NewResource()
	sr := SpanReader{Span: &s, Scope: &sc, Resource: &rs}

	traceID := sr.TraceID()
	assert.Empty(t, traceID)
}

func TestParentID(t *testing.T) {
	s := defaultTestSpan()
	sc := pcommon.NewInstrumentationScope()
	rs := pcommon.NewResource()
	sr := SpanReader{Span: &s, Scope: &sc, Resource: &rs}

	parentID := sr.ParentSpanID()
	assert.Equal(t, defaultParentID, parentID)
}

func TestParentIDEmpty(t *testing.T) {
	s := ptrace.NewSpan()
	sc := pcommon.NewInstrumentationScope()
	rs := pcommon.NewResource()
	sr := SpanReader{Span: &s, Scope: &sc, Resource: &rs}

	parentID := sr.ParentSpanID()
	assert.Empty(t, parentID)
}

func TestTokenID(t *testing.T) {
	s := ptrace.NewSpan()
	sc := pcommon.NewInstrumentationScope()
	rs := defaultTestResource()
	sr := SpanReader{Span: &s, Scope: &sc, Resource: &rs}

	id := sr.LandscapeTokenID()
	assert.Equal(t, defaultTokenID, id)
}

func TestTokenIDInSpan(t *testing.T) {
	s := ptrace.NewSpan()
	sc := pcommon.NewInstrumentationScope()
	rs := pcommon.NewResource()
	s.Attributes().PutStr(string(attrib.ExplorVizAttributes.LandscapeTokenID.Key), defaultTokenID)
	sr := SpanReader{Span: &s, Scope: &sc, Resource: &rs}

	id := sr.LandscapeTokenID()
	assert.Equal(t, defaultTokenID, id)
}

func TestTokenIDMissing(t *testing.T) {
	s := ptrace.NewSpan()
	sc := pcommon.NewInstrumentationScope()
	rs := pcommon.NewResource()
	sr := SpanReader{Span: &s, Scope: &sc, Resource: &rs}

	id := sr.LandscapeTokenID()
	assert.Empty(t, id)
}

func TestTokenSecret(t *testing.T) {
	s := ptrace.NewSpan()
	sc := pcommon.NewInstrumentationScope()
	rs := defaultTestResource()
	sr := SpanReader{Span: &s, Scope: &sc, Resource: &rs}

	sec := sr.LandscapeTokenSecret()
	assert.Equal(t, defaultTokenSecret, sec)
}

func TestTokenSecretInSpan(t *testing.T) {
	s := ptrace.NewSpan()
	sc := pcommon.NewInstrumentationScope()
	rs := pcommon.NewResource()
	s.Attributes().PutStr(string(attrib.ExplorVizAttributes.LandscapeTokenID.Key), defaultTokenID)
	s.Attributes().PutStr(string(attrib.ExplorVizAttributes.LandscapeTokenSecret.Key), defaultTokenSecret)
	sr := SpanReader{Span: &s, Scope: &sc, Resource: &rs}

	sec := sr.LandscapeTokenSecret()
	assert.Equal(t, defaultTokenSecret, sec)
}

func TestTokenSecretIDMissing(t *testing.T) {
	s := ptrace.NewSpan()
	sc := pcommon.NewInstrumentationScope()
	rs := pcommon.NewResource()
	sr := SpanReader{Span: &s, Scope: &sc, Resource: &rs}

	sec := sr.LandscapeTokenSecret()
	assert.Empty(t, sec)
}

func TestParentSpanID(t *testing.T) {
	s := defaultTestSpan()
	sc := pcommon.NewInstrumentationScope()
	rs := pcommon.NewResource()
	sr := SpanReader{Span: &s, Scope: &sc, Resource: &rs}

	pID := sr.ParentSpanID()
	assert.Equal(t, defaultParentID, pID)
}

func TestParentSpanIDEmpty(t *testing.T) {
	s := ptrace.NewSpan()
	sc := pcommon.NewInstrumentationScope()
	rs := pcommon.NewResource()
	sr := SpanReader{Span: &s, Scope: &sc, Resource: &rs}

	pID := sr.ParentSpanID()
	assert.Empty(t, pID)
}

package trace

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

const defaultSpanID string = "5fb397be34d26b51"
const defaultTraceID string = "5b8aa5a2d2c872e8321cf37308d69df2"
const defaultParentID string = "051581bf3cb55c13"

func defaultTestSpan() ptrace.Span {
	bSpanID, _ := hex.DecodeString(defaultSpanID)
	bTraceID, _ := hex.DecodeString(defaultTraceID)
	bParentID, _ := hex.DecodeString(defaultParentID)

	s := ptrace.NewSpan()
	s.SetSpanID(pcommon.SpanID(bSpanID))
	s.SetTraceID(pcommon.TraceID(bTraceID))
	s.SetParentSpanID(pcommon.SpanID(bParentID))

	return s
}

func TestSpanID(t *testing.T) {
	s := defaultTestSpan()
	spanID := StrSpanID(&s)
	assert.Equal(t, defaultSpanID, spanID)
}

func TestSpanIDEmpty(t *testing.T) {
	s := ptrace.NewSpan()
	spanID := StrSpanID(&s)
	assert.Empty(t, spanID)
}

func TestTraceID(t *testing.T) {
	s := defaultTestSpan()
	traceID := StrTraceID(&s)
	assert.Equal(t, defaultTraceID, traceID)
}

func TestTraceIDEmpty(t *testing.T) {
	s := ptrace.NewSpan()
	traceID := StrTraceID(&s)
	assert.Empty(t, traceID)
}

func TestParentID(t *testing.T) {
	s := defaultTestSpan()
	parentID := StrParentSpanID(&s)
	assert.Equal(t, defaultParentID, parentID)
}

func TestParentIDEmpty(t *testing.T) {
	s := ptrace.NewSpan()
	parentID := StrParentSpanID(&s)
	assert.Empty(t, parentID)
}

func TestParentSpanID(t *testing.T) {
	s := defaultTestSpan()
	parentID := StrParentSpanID(&s)
	assert.Equal(t, defaultParentID, parentID)
}

func TestParentSpanIDEmpty(t *testing.T) {
	s := ptrace.NewSpan()
	parentID := StrParentSpanID(&s)
	assert.Empty(t, parentID)
}

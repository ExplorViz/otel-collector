// Package trace contains utilities for working with trace data.
package trace

import (
	"encoding/hex"

	"go.opentelemetry.io/collector/pdata/ptrace"
)

// StrTraceID returns the hex string representation of the provided span's trace ID
// Returns the empty string if the span is missing a trace ID.
func StrTraceID(s *ptrace.Span) string {
	b := s.TraceID()
	if b.IsEmpty() {
		return ""
	}
	return hex.EncodeToString(b[:])
}

// StrSpanID returns the hex string representation of the provided span's span ID.
// Returns the empty string if the span is missing a span ID.
func StrSpanID(s *ptrace.Span) string {
	b := s.SpanID()
	if b.IsEmpty() {
		return ""
	}
	return hex.EncodeToString(b[:])
}

// StrParentSpanID returns the hex string representation of the span's parent span ID.
// Returns the empty string if the span has no parent.
func StrParentSpanID(s *ptrace.Span) string {
	b := s.ParentSpanID()
	if b.IsEmpty() {
		return ""
	}
	return hex.EncodeToString(b[:])
}

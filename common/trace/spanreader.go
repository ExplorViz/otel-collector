// Package trace contains utilities for working with trace data.
package trace

import (
	"encoding/hex"

	"github.com/ExplorViz/otel-collector/common/attrib"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/otel/attribute"
)

// A SpanReader groups a [ptrace.Span] together with its [pcommon.InstrumentationScope]
// and [pcommon.Resource]. It provides helper methods for frequent attribute lookup operations.
type SpanReader struct {
	Span     *ptrace.Span
	Scope    *pcommon.InstrumentationScope
	Resource *pcommon.Resource
}

// SpanStrAttrib is a convenience function that looks for a string attribute with the
// given key within the underlying span. The string value of the attribute is returned.
// If the attribute does not exist or has a non-string type, the empty string is returned.
func (sr SpanReader) SpanStrAttrib(key attribute.Key) string {
	a, _ := sr.Span.Attributes().Get(string(key))
	return a.Str()
}

// ScopeStrAttrib is a convenience function that looks for a string attribute with the
// given key within the underlying instrumentation scope. The string value of the attribute is returned.
// If the attribute does not exist or has a non-string type, the empty string is returned.
func (sr SpanReader) ScopeStrAttrib(key attribute.Key) string {
	a, _ := sr.Scope.Attributes().Get(string(key))
	return a.Str()
}

// ResourceStrAttrib is a convenience function that looks for a string attribute with the
// given key within the underlying resource. The string value of the attribute is returned.
// If the attribute does not exist or has a non-string type, the empty string is returned.
func (sr SpanReader) ResourceStrAttrib(key attribute.Key) string {
	a, _ := sr.Resource.Attributes().Get(string(key))
	return a.Str()
}

// TraceID returns the hex string representation of the span's trace ID
// Returns the empty string if the span is missing a trace ID.
func (sr SpanReader) TraceID() string {
	b := sr.Span.TraceID()
	if b.IsEmpty() {
		return ""
	}
	return hex.EncodeToString(b[:])
}

// SpanID returns the hex string representation of the span's span ID.
// Returns the empty string if the span is missing a span ID.
func (sr SpanReader) SpanID() string {
	b := sr.Span.SpanID()
	if b.IsEmpty() {
		return ""
	}
	return hex.EncodeToString(b[:])
}

// ParentSpanID returns the hex string representation of the span's parent span ID.
// Returns the empty string if the span has no parent.
func (sr SpanReader) ParentSpanID() string {
	b := sr.Span.ParentSpanID()
	if b.IsEmpty() {
		return ""
	}
	return hex.EncodeToString(b[:])
}

// LandscapeTokenID looks for an attribute specifying the ID of an ExplorViz landscape token.
// The searched attribute key is given by [attrib.ExplorVizAttributes.LandscapeTokenID].
// The resource attributes are considered first. If no token ID can be extracted from the resource,
// then we look at the span attributes as a fallback. If this also fails, the empty string is returned.
func (sr SpanReader) LandscapeTokenID() string {
	resTokenID, ok := sr.Resource.Attributes().Get(string(attrib.ExplorVizAttributes.LandscapeTokenID.Key))
	if ok && resTokenID.Str() != "" {
		return resTokenID.Str()
	}

	spanTokenID, ok := sr.Span.Attributes().Get(string(attrib.ExplorVizAttributes.LandscapeTokenID.Key))
	if ok {
		return spanTokenID.Str()
	}

	return ""
}

// LandscapeTokenSecret looks for an attribute specifying the secret of an ExplorViz landscape token.
// The searched attribute key is given by [ExplorVizAttributes.LandscapeTokenSecret.Key].
// The resource attributes are considered first. If no token secret can be extracted from the resource,
// then we look at the span attributes as a fallback. If this also fails, the empty string is returned.
func (sr SpanReader) LandscapeTokenSecret() string {
	resTokenSec, ok := sr.Resource.Attributes().Get(string(attrib.ExplorVizAttributes.LandscapeTokenSecret.Key))
	if ok && resTokenSec.Str() != "" {
		return resTokenSec.Str()
	}

	spanTokenID, ok := sr.Span.Attributes().Get(string(attrib.ExplorVizAttributes.LandscapeTokenSecret.Key))
	if ok {
		return spanTokenID.Str()
	}

	return ""
}

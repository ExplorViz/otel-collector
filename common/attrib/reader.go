package attrib

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// A TelemetryReader groups the attributes of a telemetry unit (span, metric, log) together with its instrumentation scope
// and resource. It provides helper methods for frequent attribute lookup operations.
type TelemetryReader struct {
	Attrs    *pcommon.Map
	Scope    *pcommon.InstrumentationScope
	Resource *pcommon.Resource
}

// StrAttrib is a convenience function that looks for a string attribute with the
// given key within the underlying telemetry unit. The string value of the attribute is returned.
// If the attribute does not exist or has a non-string type, the empty string is returned.
func (tr TelemetryReader) StrAttrib(key attribute.Key) string {
	a, _ := tr.Attrs.Get(string(key))
	return a.Str()
}

// ScopeStrAttrib is a convenience function that looks for a string attribute with the
// given key within the underlying instrumentation scope. The string value of the attribute is returned.
// If the attribute does not exist or has a non-string type, the empty string is returned.
func (tr TelemetryReader) ScopeStrAttrib(key attribute.Key) string {
	a, _ := tr.Scope.Attributes().Get(string(key))
	return a.Str()
}

// ResourceStrAttrib is a convenience function that looks for a string attribute with the
// given key within the underlying resource. The string value of the attribute is returned.
// If the attribute does not exist or has a non-string type, the empty string is returned.
func (tr TelemetryReader) ResourceStrAttrib(key attribute.Key) string {
	a, _ := tr.Resource.Attributes().Get(string(key))
	return a.Str()
}

// LandscapeTokenID looks for an attribute specifying the ID of an ExplorViz landscape token.
// The searched attribute key is given by [attrib.ExplorVizAttributes.LandscapeTokenID].
// The resource attributes are considered first. If no token ID can be extracted from the resource,
// then we look at the telemetry attributes as a fallback. If this also fails, the empty string is returned.
func (tr TelemetryReader) LandscapeTokenID() string {
	tokenID, ok := tr.Resource.Attributes().Get(string(ExplorVizAttributes.LandscapeTokenID.Key))
	if ok && tokenID.Str() != "" {
		return tokenID.Str()
	}

	tokenID, ok = tr.Attrs.Get(string(ExplorVizAttributes.LandscapeTokenID.Key))
	if ok {
		return tokenID.Str()
	}

	return ""
}

// LandscapeTokenSecret looks for an attribute specifying the secret of an ExplorViz landscape token.
// The searched attribute key is given by [ExplorVizAttributes.LandscapeTokenSecret.Key].
// The resource attributes are considered first. If no token secret can be extracted from the resource,
// then we look at the telemetry attributes as a fallback. If this also fails, the empty string is returned.
func (tr TelemetryReader) LandscapeTokenSecret() string {
	tokenSec, ok := tr.Resource.Attributes().Get(string(ExplorVizAttributes.LandscapeTokenSecret.Key))
	if ok && tokenSec.Str() != "" {
		return tokenSec.Str()
	}

	tokenSec, ok = tr.Attrs.Get(string(ExplorVizAttributes.LandscapeTokenSecret.Key))
	if ok {
		return tokenSec.Str()
	}

	return ""
}

// GitCommitHash looks for an attribute specifying the checksum of the Git commit related to the telemetry.
// The searched attribute key is given by [semconv.VCSRefHeadRevisionKey].
// The resource attributes are considered first. If no commit hash can be extracted from the resource,
// then we look at the telemetry attributes as a fallback. If this also fails, the empty string is returned.
func (tr TelemetryReader) GitCommitHash() string {
	commitHash, ok := tr.Resource.Attributes().Get(string(semconv.VCSRefHeadRevisionKey))
	if ok && commitHash.Str() != "" {
		return commitHash.Str()
	}

	commitHash, ok = tr.Attrs.Get(string(semconv.VCSRefHeadRevisionKey))
	if ok {
		return commitHash.Str()
	}

	return ""
}

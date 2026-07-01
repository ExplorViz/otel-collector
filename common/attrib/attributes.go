// Package attrib is concerned with OpenTelemetry attributes.
// It defines ExplorViz-specific attributes and provides utilities
// for retrieving attributes from OTLP data.
package attrib

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/otel/attribute"
)

type ExplorVizAttribute struct {
	Key          attribute.Key
	DefaultValue pcommon.Value
}

// ExplorVizAttributes defines attributes required or written by ExplorViz itself.
var ExplorVizAttributes = struct {
	LandscapeTokenID     ExplorVizAttribute
	LandscapeTokenSecret ExplorVizAttribute
	EntityId             ExplorVizAttribute
	EntityDescriptor     ExplorVizAttribute
}{
	LandscapeTokenID: ExplorVizAttribute{
		Key:          "explorviz.token.id",
		DefaultValue: pcommon.NewValueStr("mytokenvalue"),
	},
	LandscapeTokenSecret: ExplorVizAttribute{
		Key:          "explorviz.token.secret",
		DefaultValue: pcommon.NewValueStr("mytokenvalue"),
	},
	EntityId: ExplorVizAttribute{
		Key:          "explorviz.entity.id",
		DefaultValue: pcommon.NewValueStr("unknown"),
	},
	EntityDescriptor: ExplorVizAttribute{
		Key:          "explorviz.entity.descriptor",
		DefaultValue: pcommon.NewValueMap(),
	},
}

var FallbackValues = struct {
	ServiceName string
}{
	ServiceName: "unknown-service",
}

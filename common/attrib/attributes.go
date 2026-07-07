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
	EntityID             ExplorVizAttribute
	EntityDescriptor     ExplorVizAttribute
	VizObjID             ExplorVizAttribute
}{
	LandscapeTokenID: ExplorVizAttribute{
		Key:          "explorviz.token.id",
		DefaultValue: pcommon.NewValueStr("mytokenvalue"),
	},
	LandscapeTokenSecret: ExplorVizAttribute{
		Key:          "explorviz.token.secret",
		DefaultValue: pcommon.NewValueStr("mytokenvalue"),
	},
	EntityID: ExplorVizAttribute{
		Key:          "explorviz.entity.id",
		DefaultValue: pcommon.NewValueStr("unknown-entity"),
	},
	EntityDescriptor: ExplorVizAttribute{
		Key:          "explorviz.entity.descriptor",
		DefaultValue: pcommon.NewValueMap(),
	},
	VizObjID: ExplorVizAttribute{
		Key:          "explorviz.vizobject.id",
		DefaultValue: pcommon.NewValueStr("unknown-vizobject"),
	},
}

var FallbackValues = struct {
	ServiceName string
}{
	ServiceName: "unknown-service",
}

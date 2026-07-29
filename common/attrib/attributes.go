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

	// Uniquely identifies an entity extracted from telemetry data. This should be a composite ID encoding the entity's type and properties,
	// which is subsequently hashed using the 64-bit xxHash algorithm and represented as a hexadecimal string.
	EntityID    ExplorVizAttribute
	EntityType  ExplorVizAttribute // Indicates the kind of entity extracted, e.g. "function" or "database".
	VizObjectID ExplorVizAttribute // Identifier / lookup key of the visualization object to which this telemetry belongs, e.g. building

	CodeFileID       ExplorVizAttribute
	CodeFilePath     ExplorVizAttribute
	CodeFunctionName ExplorVizAttribute
	CodeClassName    ExplorVizAttribute
	CodeLanguage     ExplorVizAttribute
}{
	LandscapeTokenID:     ExplorVizAttribute{Key: "explorviz.token.id", DefaultValue: pcommon.NewValueStr("mytokenvalue")},
	LandscapeTokenSecret: ExplorVizAttribute{Key: "explorviz.token.secret", DefaultValue: pcommon.NewValueStr("mytokenvalue")},

	EntityID:    ExplorVizAttribute{Key: "explorviz.entity.id", DefaultValue: pcommon.NewValueStr("unknown-entity-id")},
	EntityType:  ExplorVizAttribute{Key: "explorviz.entity.type", DefaultValue: pcommon.NewValueStr("unknown-entity-type")},
	VizObjectID: ExplorVizAttribute{Key: "explorviz.vizobject.id", DefaultValue: pcommon.NewValueStr("unknown-vizobject-id")},

	CodeFileID:       ExplorVizAttribute{Key: "explorviz.code.file.id", DefaultValue: pcommon.NewValueStr("unknown-file-id")},
	CodeFilePath:     ExplorVizAttribute{Key: "explorviz.code.file.path", DefaultValue: pcommon.NewValueStr("unknown-file-path")},
	CodeFunctionName: ExplorVizAttribute{Key: "explorviz.code.function.name", DefaultValue: pcommon.NewValueStr("unknown-function-name")},
	CodeClassName:    ExplorVizAttribute{Key: "explorviz.code.class.name", DefaultValue: pcommon.NewValueStr("unknown-class-name")},
	CodeLanguage:     ExplorVizAttribute{Key: "explorviz.code.language", DefaultValue: pcommon.NewValueStr("")},
}

// FallbackValues provides values for (non-ExplorViz specific) OTel attributes to use if the attribute is not provided within incoming telemetry.
var FallbackValues = struct {
	ServiceName string
}{
	ServiceName: "unknown-service",
}

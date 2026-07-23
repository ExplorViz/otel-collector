// Package parsing is concerned with the interpretation of telemetry attributes,
// with the goal of infering the entity within the system that the telemetry describes.
//
// Various parsing functions and result types are defined which rely on different
// attributes from the [OTel semantic conventions].
//
// [OTel semantic conventions]: https://opentelemetry.io/docs/specs/semconv/registry/attributes/code/
package parsing

import (
	"errors"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/ExplorViz/otel-collector/common/attrib"
)

// A TelemetryParser is a function which extracts a specific type of entity described by telemetry.
// Such a function has specific attributes it looks for within telemetry data, and if sufficient information
// is given to reliably classify the entity described by the telemetry, an [Entity] is returned.
// Otherwise, an error is returned and the returned Entity is invalid.
type TelemetryParser func(tr attrib.TelemetryReader) (Entity, error)

// An Entity is some visualization component for which telemetry may be created.
// It either directly corresponds to a visualization object or is contained within one.
// Examples include functions in code, HTTP endpoints, and databases.
//
// Note that these are distinct from the [OpenTelemetry definition of entities].
//
// [OpenTelemetry definition of entities]: https://opentelemetry.io/docs/specs/otel/entities/
type Entity interface {
	// ID returns a unique identifier for this entity.
	// It is unique in the sense that only one specific combination of attributes should produce that ID.
	// The identifier should be a composite key based on the entity's type and its attribute values.
	ID() string

	// VizObjectID returns a unique identifier for the visualization object to which this telemetry belongs, e.g. building.
	// The identifier should be a composite key based on the entity's type and its attribute values.
	VizObjectID() string

	// ToAttributes encodes the entity as OTel attributes and stores it within the provided attributes map.
	// Calling [FromMap] on the modified attributes map should yield an entity of the same type with identical field values.
	ToAttributes(attrs *pcommon.Map)
}

type invalidEntity struct{}

func (invalidEntity) ID() string {
	panic("ID() called on invalidEntity")
}

func (invalidEntity) VizObjectID() string {
	panic("VizObjectID() called on invalidEntity")
}

func (invalidEntity) ToAttributes(_ *pcommon.Map) {
	panic("ToAttributes() called on invalidEntity")
}

var parserChain = []TelemetryParser{
	ParseCodeTelemetry,
}

// ParseTelemetry applies a series of parsing functions to the given [attrib.TelemetryReader]
// which attempt to extract a [Entity] that the telemetry describes. If insufficient
// information is provided within the attributes for a parser to extract an entity, the next
// parser in the chain is used. The first parsing function to return a non-error value gives
// the final result. If all parsing functions fail, an error is returned and the returned entity is invalid.
// Calling any function on such an invalid entity will cause a panic.
func ParseTelemetry(tr attrib.TelemetryReader) (Entity, error) {
	for _, parser := range parserChain {
		entity, err := parser(tr)
		if err != nil {
			slog.Debug("parser failed", "error", err)
			continue
		}
		return entity, nil
	}
	return invalidEntity{}, errors.New("no matching telemetry parser found")
}

// FromAttributes returns a new [Entity] based on the entries in the provided attribute [pcommon.Map].
// It should be used to reconstruct entities previously encoded using [Entity.ToAttributes].
// If the map has insufficient entries or provides invalid values, then an invalid
// entity and an error will be returned.
func FromAttributes(m pcommon.Map) (Entity, error) {
	entityType, ok := m.Get(string(attrib.ExplorVizAttributes.EntityType.Key))
	if !ok {
		return invalidEntity{}, fmt.Errorf("cannot construct entity from attributes, missing attribute %s", attrib.ExplorVizAttributes.EntityType.Key)
	}

	var se Entity
	var err error
	switch entityType.Str() {
	case CodeEntityType:
		se, err = codeEntityFromAttribs(m)
	default:
		return invalidEntity{}, fmt.Errorf(`failed to initialize entity from attributes, unknown value "%s" for entity type`, entityType.Str())
	}
	if err != nil {
		return invalidEntity{}, fmt.Errorf("failed to initialize entity from attributes: %v", err)
	}
	return se, nil
}

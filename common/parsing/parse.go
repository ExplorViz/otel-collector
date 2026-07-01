// Package parsing is concerned with the interpretation of span attributes,
// with the goal of infering the entity within the system that the span describes.
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

	"github.com/ExplorViz/otel-collector/common/trace"
)

// A SpanParser is a function which extracts a specific type of entity described by spans.
// Such a function has specific attributes it looks for within spans, and if sufficient information
// is given to reliably classify the entity described by the span, a [SpanEntity] is returned.
// Otherwise, an error is returned and the returned SpanEntity is invalid.
type SpanParser func(sr trace.SpanReader) (SpanEntity, error)

// A SpanEntity is some visualizable component for which spans may be created.
// Examples include functions in code, HTTP endpoints, and databases.
type SpanEntity interface {
	// Id returns a unique identifier for this entity. The identifier should be
	// a composite key based on the entity's type and its attribute values.
	Id() string

	// ToMap encodes the entity as a [pcommon.Map] such that it can be stored within telemetry attributes.
	// Calling FromMap on the result should yield an entity with identical attribute values.
	ToMap() pcommon.Map
}

type invalidEntity struct{}

func (invalidEntity) Id() string {
	panic("Id() called on invalidEntity")
}

func (invalidEntity) ToMap() pcommon.Map {
	panic("ToMap() called on invalidEntity")
}

var parserChain = []SpanParser{
	ParseCodeSpan,
}

// ParseSpan applies a series of parsing functions to the given [trace.SpanReader]
// which attempt to extract a [SpanEntity] that the span describes. If insufficient
// information is provided within the attributes for a parser to extract an entity, the next
// parser in the chain is used. The first parsing function to return a non-error value gives
// the final result. If all parsing functions fail, an error is returned and the returned entity is invalid.
// Calling any function on such an invalid entity will cause a panic.
func ParseSpan(sr trace.SpanReader) (SpanEntity, error) {
	for _, parser := range parserChain {
		entity, err := parser(sr)
		if err != nil {
			slog.Debug("parser failed", "error", err)
			continue
		}
		return entity, nil
	}
	return invalidEntity{}, errors.New("no matching span parser found")
}

// FromMap returns a new [SpanEntity] based on the entries in the provided [pcommon.Map].
// It should be used to reconstruct entities previously encoded using [SpanEntity.ToMap].
// If the map has insufficient entries or provides invalid values, then an invalid
// entity and an error will be returned.
func FromMap(m pcommon.Map) (SpanEntity, error) {
	entityType, ok := m.Get("type")
	if !ok {
		return invalidEntity{}, errors.New(`cannot construct entity from map, missing key "type"`)
	}

	var se SpanEntity
	var err error
	switch entityType.Str() {
	case CodeSpanEntityType:
		se, err = codeSpanEntityFromMap(m)
	default:
		return invalidEntity{}, fmt.Errorf(`failed to initialize entity from map, unknown value "%s" for key "type"`, entityType.Str())
	}
	if err != nil {
		return invalidEntity{}, fmt.Errorf("failed to initialize entity from map: %v", err)
	}
	return se, nil
}

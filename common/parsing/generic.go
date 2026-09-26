package parsing

import (
	"errors"
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"

	"github.com/ExplorViz/otel-collector/common/attrib"
)

const GenericEntityType string = "generic"

// A GenericEntity represents an entity which cannot be classified apart from the name
// of the service and instrumentation scope it is a part of. It should only be used as
// a fallback if a more specific entity cannot be extracted.
type GenericEntity struct {
	ServiceName string
	ScopeName   string

	// The name of a generic entity is derived from the scope name it belongs to.
	// If the scope name is qualified using "/" or "::" as a separator, the name
	// is simplified to only include the portion after the last separator occurrence.
	Name string
}

func (gs GenericEntity) ID() string {
	return "generic" + "|" + gs.ServiceName + "|" + gs.ScopeName
}

func (gs GenericEntity) TelemetryKey() string {
	return gs.ID()
}

func (gs GenericEntity) ToAttributes(attrs *pcommon.Map) {
	attrs.PutStr(string(attrib.ExplorVizAttributes.EntityType.Key), GenericEntityType)
	attrs.PutStr(string(attrib.ExplorVizAttributes.ServiceName.Key), gs.ServiceName)
	attrs.PutStr(string(attrib.ExplorVizAttributes.ScopeName.Key), gs.ScopeName)
	attrs.PutStr(string(attrib.ExplorVizAttributes.GenericEntityName.Key), gs.Name)
}

// genericEntityFromAttribs initializes a new [GenericEntity] based on the entries of the provided map.
// If the provided map entries are incomplete (meaning a mandatory attribute is missing), then a
// zero-initialized GenericEntity and an error is returned.
func genericEntityFromAttribs(m pcommon.Map) (GenericEntity, error) {
	service, ok := m.Get(string(attrib.ExplorVizAttributes.ServiceName.Key))
	if !ok || service.Str() == "" {
		return GenericEntity{}, errors.New("empty or missing string attribute for service name")
	}

	scope, ok := m.Get(string(attrib.ExplorVizAttributes.ScopeName.Key))
	if !ok || scope.Str() == "" {
		return GenericEntity{}, errors.New("empty or missing string attribute for scope name")
	}

	name, ok := m.Get(string(attrib.ExplorVizAttributes.GenericEntityName.Key))
	if !ok || name.Str() == "" {
		return GenericEntity{}, errors.New("empty or missing string attribute for entity name")
	}

	return GenericEntity{
		ServiceName: service.Str(),
		ScopeName:   scope.Str(),
		Name:        name.Str(),
	}, nil
}

// ParseGenericTelemetry parses telemetry describing some generic entity by looking for attributes conforming
// to the [OTel semconv service attributes]. For telemetry to be successfully parsed, it needs to provide:
//   - a service name
//   - an instrumentation scope name
//
// [OTel semconv service attributes]: https://opentelemetry.io/docs/specs/semconv/registry/attributes/service/
func ParseGenericTelemetry(tr attrib.TelemetryReader) (Entity, error) {
	service := tr.ResourceStrAttrib(semconv.ServiceNameKey)
	if service == "" {
		return GenericEntity{}, errors.New("generic service parser: empty or missing service name attribute")
	}

	scope := tr.Scope.Name()
	name := dequalifyScopeName(scope)

	return GenericEntity{
		ServiceName: service,
		ScopeName:   scope,
		Name:        name,
	}, nil
}

func dequalifyScopeName(scopeName string) string {
	trimmed := scopeName
	for {
		var found1, found2 bool
		trimmed, found1 = strings.CutSuffix(trimmed, "/")
		trimmed, found2 = strings.CutSuffix(trimmed, "::")
		if !found1 && !found2 {
			break
		}
	}

	prefixEnd := 0
	for prefixEnd < len(trimmed) {
		if strings.HasPrefix(trimmed[prefixEnd:], "/") {
			prefixEnd += 1
		} else if strings.HasPrefix(trimmed[prefixEnd:], "::") {
			prefixEnd += 2
		} else {
			break
		}
	}

	slash := strings.LastIndex(trimmed, "/")
	colon := strings.LastIndex(trimmed, "::")

	if slash > colon && slash > prefixEnd {
		return scopeName[slash+1:]
	} else if colon > prefixEnd {
		return scopeName[colon+2:]
	}

	return scopeName
}

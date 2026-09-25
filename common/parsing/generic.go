package parsing

import (
	"errors"

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

	return GenericEntity{ServiceName: service.Str(), ScopeName: scope.Str()}, nil
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

	return GenericEntity{ServiceName: service, ScopeName: scope}, nil
}

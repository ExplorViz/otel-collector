package parsing

import (
	"errors"

	"go.opentelemetry.io/collector/pdata/pcommon"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"

	"github.com/ExplorViz/otel-collector/common/attrib"
)

const GenericServiceEntityType string = "genericservice"

// A GenericServiceEntity represents a service, identified only by its name.
// It should only be used as a fallback if a more specific entity cannot be extracted.
type GenericServiceEntity struct {
	ServiceName string
}

func (gs GenericServiceEntity) ID() string {
	return "genericservice" + "|" + gs.ServiceName
}

func (gs GenericServiceEntity) TelemetryKey() string {
	return gs.ID()
}

func (gs GenericServiceEntity) ToAttributes(attrs *pcommon.Map) {
	attrs.PutStr(string(attrib.ExplorVizAttributes.EntityType.Key), GenericServiceEntityType)
	attrs.PutStr(string(attrib.ExplorVizAttributes.ServiceName.Key), gs.ServiceName)
}

// genericServiceEntityFromAttribs initializes a new [GenericServiceEntity] based on the entries
// of the provided map. If the provided map entries are incomplete (meaning a mandatory attribute
// is missing), then a zero-initialized GenericServiceEntity and an error is returned.
func genericServiceEntityFromAttribs(m pcommon.Map) (GenericServiceEntity, error) {
	serviceName, ok := m.Get(string(attrib.ExplorVizAttributes.ServiceName.Key))
	if !ok || serviceName.Str() == "" {
		return GenericServiceEntity{}, errors.New(`empty or missing string attribute for service name`)
	}

	return GenericServiceEntity{ServiceName: serviceName.Str()}, nil
}

// ParseGenericServiceTelemetry parses telemetry describing some generic service by looking for attributes
// conforming to the [OTel semconv service attributes]. For telemetry to be successfully parsed, it needs to provide:
//   - a service name
//
// [OTel semconv service attributes]: https://opentelemetry.io/docs/specs/semconv/registry/attributes/service/
func ParseGenericServiceTelemetry(tr attrib.TelemetryReader) (Entity, error) {
	name := tr.ResourceStrAttrib(semconv.ServiceNameKey)
	if name == "" {
		return GenericServiceEntity{}, errors.New("generic service parser: empty or missing service name attribute")
	}

	return GenericServiceEntity{ServiceName: name}, nil
}

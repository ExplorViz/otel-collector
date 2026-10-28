package parsing

import (
	"errors"

	"go.opentelemetry.io/collector/pdata/pcommon"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	"github.com/ExplorViz/otel-collector/common/attrib"
)

const MessagingEntityType string = "messaging"

// A MessagingEntity represents a destination within a message broker system (e.g. Kafka, RabbitMQ),
type MessagingEntity struct {
	// Name of the message broker system, e.g. "kafka".
	SystemName string

	// Name of the destination (e.g. topic, queue name) within the system.
	DestinationName string
}

func (m MessagingEntity) ID() string {
	return "messaging" + "|" + m.SystemName + "|" + m.DestinationName
}

func (m MessagingEntity) TelemetryKey() string {
	return m.ID()
}

func (m MessagingEntity) ToAttributes(attrs *pcommon.Map) {
	attrs.PutStr(string(attrib.ExplorVizAttributes.EntityType.Key), MessagingEntityType)

	// Use system name as service name. Since messaging operations are usually only instrumented
	// client-side, the resource attributes will not specify the messaging system as its own service.
	attrs.PutStr(string(attrib.ExplorVizAttributes.ServiceName.Key), m.SystemName)

	attrs.PutStr(string(attrib.ExplorVizAttributes.MessagingSystemName.Key), m.SystemName)
	attrs.PutStr(string(attrib.ExplorVizAttributes.MessagingDestinationName.Key), m.DestinationName)
}

// messagingEntityFromAttribs initializes a new [MessagingEntity] based on the entries of the provided map.
// If the provided map entries are incomplete (meaning a mandatory attribute is missing),
// then a zero-initialized MessagingEntity and an error is returned.
func messagingEntityFromAttribs(m pcommon.Map) (MessagingEntity, error) {
	system, ok := m.Get(string(attrib.ExplorVizAttributes.MessagingSystemName.Key))
	if !ok || system.Str() == "" {
		return MessagingEntity{}, errors.New("empty or missing string attribute for messaging system name")
	}

	dest, ok := m.Get(string(attrib.ExplorVizAttributes.MessagingDestinationName.Key))
	if !ok || dest.Str() == "" {
		return MessagingEntity{}, errors.New("empty or missing string attribute for messaging destination name")
	}

	return MessagingEntity{
		SystemName:      system.Str(),
		DestinationName: dest.Str(),
	}, nil
}

// ParseMessagingTelemetry parses telemetry describing message broker operations by looking for attributes conforming
// to the [OTel semconv messaging attributes]. For telemetry to be successfully parsed as a messaging entity,
// it needs to provide:
//   - a name for the message broker system used
//   - a name for the message destination (e.g. topic, queue name)
//
// [OTel semconv messaging attributes]: https://opentelemetry.io/docs/specs/semconv/registry/attributes/messaging/
func ParseMessagingTelemetry(tr attrib.TelemetryReader) (Entity, error) {
	system := tr.StrAttrib(semconv.MessagingSystemKey)
	if system == "" {
		return MessagingEntity{}, errors.New("messaging parser: empty or missing messaging system name attribute")
	}

	dest := tr.StrAttrib(semconv.MessagingDestinationNameKey)
	if dest == "" {
		return MessagingEntity{}, errors.New("messaging parser: empty or missing message destination name attribute")
	}

	return MessagingEntity{
		SystemName:      system,
		DestinationName: dest,
	}, nil
}

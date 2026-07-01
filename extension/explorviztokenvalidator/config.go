package explorviztokenvalidator

import "go.opentelemetry.io/collector/component"

type Config struct {
	// Network endpoint of the Kafka broker to use.
	Broker string `mapstructure:"broker"`

	// Kafka topic to consume token events from.
	Topic string `mapstructure:"topic"`
}

var _ component.Config = (*Config)(nil)

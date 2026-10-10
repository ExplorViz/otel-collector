package explorvizexporter

import "go.opentelemetry.io/collector/component"

type Config struct {
	// Network endpoint of the Kafka broker to use.
	Broker string `mapstructure:"broker"`

	// Kafka topic to produce parsing results into.
	Topic string `mapstructure:"topic"`

	// Metrics-specific configuration options.
	MetricsCfg MetricsConfig `mapstructure:"metrics"`

	// Logs-specific configuration options.
	LogsCfg LogsConfig `mapstructure:"logs"`
}

type MetricsConfig struct {
	// Reports whether the exporter should allow sending entities of generic type,
	// meaning entities which could not be further classified based on the provided attributes.
	// Depending on the number of instrumented scopes, these can produce a lot of clutter in the
	// visualization while potentially never being involved in any communication.
	ProduceGenericEntities bool `mapstructure:"produce_generic_entities"`
}

type LogsConfig struct {
	// Reports whether the exporter should allow sending entities of generic type,
	// meaning entities which could not be further classified based on the provided attributes.
	// Depending on the number of instrumented scopes, these can produce a lot of clutter in the
	// visualization while potentially never being involved in any communication.
	ProduceGenericEntities bool `mapstructure:"produce_generic_entities"`
}

var _ component.Config = (*Config)(nil)

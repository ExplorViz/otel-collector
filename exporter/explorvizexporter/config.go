package explorvizexporter

type Config struct {
	// Network endpoint of the Kafka broker to use.
	Broker string `mapstructure:"broker"`

	// Kafka topic to produce parsing results into.
	Topic string `mapstructure:"topic"`
}

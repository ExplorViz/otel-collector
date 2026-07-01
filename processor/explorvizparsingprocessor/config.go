package explorvizparsingprocessor

import "go.opentelemetry.io/collector/component"

type Config struct {
	ValidateTokens bool `mapstructure:"validate_tokens"`
}

var _ component.Config = (*Config)(nil)

// Validate checks if the processor configuration is valid
func (*Config) Validate() error {
	return nil
}

package explorviztokenvalidator

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"

	"github.com/ExplorViz/otel-collector/extension/explorviztokenvalidator/internal/metadata"
)

func NewFactory() extension.Factory {
	return extension.NewFactory(
		metadata.Type,
		createDefaultConfig,
		createExtension,
		metadata.ExtensionStability,
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		Broker: "localhost:9091",
		Topic:  "tokens.events",
	}
}

func createExtension(_ context.Context, set extension.Settings, cfg component.Config) (extension.Extension, error) {
	c := cfg.(*Config)
	ext := newTokenValidatorExtension(c, set.Logger)
	return &ext, nil
}

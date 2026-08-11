package explorvizexporter

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"

	"github.com/ExplorViz/otel-collector/exporter/explorvizexporter/internal/metadata"
)

func NewFactory() exporter.Factory {
	return exporter.NewFactory(
		metadata.Type,
		createDefaultConfig,
		exporter.WithTraces(createTraceExporter, metadata.TracesStability),
		exporter.WithMetrics(createMetricsExporter, metadata.MetricsStability),
		exporter.WithLogs(createLogExporter, metadata.LogsStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		Broker: "localhost:9091",
		Topic:  "telemetry.entities",
	}
}

func createTraceExporter(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Traces, error) {
	c := cfg.(*Config)

	exp := newExplorVizExporter(c, set.Logger)

	return exporterhelper.NewTraces(
		ctx,
		set,
		cfg,
		exp.consumeTraces,
		exporterhelper.WithStart(exp.start),
		exporterhelper.WithShutdown(exp.shutdown),
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: true}),
	)
}

func createMetricsExporter(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Metrics, error) {
	c := cfg.(*Config)

	exp := newExplorVizExporter(c, set.Logger)

	return exporterhelper.NewMetrics(
		ctx,
		set,
		cfg,
		exp.consumeMetrics,
		exporterhelper.WithStart(exp.start),
		exporterhelper.WithShutdown(exp.shutdown),
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: true}),
	)
}

func createLogExporter(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Logs, error) {
	c := cfg.(*Config)

	exp := newExplorVizExporter(c, set.Logger)

	return exporterhelper.NewLogs(
		ctx,
		set,
		cfg,
		exp.consumeLogs,
		exporterhelper.WithStart(exp.start),
		exporterhelper.WithShutdown(exp.shutdown),
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: true}),
	)
}

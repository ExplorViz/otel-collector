package explorvizexporter

import (
	"context"
	"fmt"
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/ExplorViz/otel-collector/common/attrib"
	"github.com/ExplorViz/otel-collector/common/parsing"
	"github.com/ExplorViz/otel-collector/exporter/explorvizexporter/internal/encoding"
)

type explorVizExporter struct {
	logger *zap.Logger
	client *kgo.Client
	config Config
}

func newExplorVizExporter(cfg *Config, log *zap.Logger) explorVizExporter {
	return explorVizExporter{logger: log, config: *cfg}
}

func (e *explorVizExporter) start(ctx context.Context, host component.Host) error {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(e.config.Broker),
		kgo.DefaultProduceTopic(e.config.Topic),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize kgo client: %v", err)
	}
	e.client = cl

	return nil
}

func (e *explorVizExporter) shutdown(ctx context.Context) error {
	if e.client != nil {
		e.client.Close()
	}
	return nil
}

func (e *explorVizExporter) consumeReader(produceCtx context.Context, tr attrib.TelemetryReader, promise func(r *kgo.Record, err error)) error {
	entity, err := parsing.FromAttributes(*tr.Attrs)
	if err != nil {
		return fmt.Errorf("failed to reconstruct entity from descriptor map: %w", err)
	}

	if _, ok := entity.(parsing.GenericEntity); ok {
		if (tr.Metric != nil && !e.config.MetricsCfg.ProduceGenericEntities) || (tr.Log != nil && !e.config.LogsCfg.ProduceGenericEntities) {
			return nil
		}
	}

	pb, err := encoding.ToProtobuf(tr, entity)
	if err != nil {
		return fmt.Errorf("failed to convert to protobuf message: %w", err)
	}

	out, err := proto.Marshal(pb)
	if err != nil {
		return fmt.Errorf("failed to encode protobuf: %w", err)
	}

	e.client.Produce(produceCtx, &kgo.Record{
		Key:   []byte(pb.GetLandscapeTokenId()),
		Value: out,
	}, promise)

	return nil
}

func (e *explorVizExporter) consumeTraces(ctx context.Context, td ptrace.Traces) error {
	var (
		wg       sync.WaitGroup
		once     sync.Once
		firstErr error
	)
	produceCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	promise := func(r *kgo.Record, err error) {
		defer wg.Done()

		if err != nil {
			once.Do(func() {
				firstErr = err
				cancel()
			})
		}
	}

loop:
	for i := 0; i < td.ResourceSpans().Len(); i++ {
		rs := td.ResourceSpans().At(i)
		for j := 0; j < rs.ScopeSpans().Len(); j++ {
			ss := rs.ScopeSpans().At(j)
			for k := 0; k < ss.Spans().Len(); k++ {
				select {
				case <-produceCtx.Done():
					break loop
				default:
				}

				span := ss.Spans().At(k)
				attrs := span.Attributes()
				scope := ss.Scope()
				res := rs.Resource()

				tr := attrib.TelemetryReader{
					Attrs:    &attrs,
					Scope:    &scope,
					Resource: &res,
					Span:     &span,
				}

				err := e.consumeReader(produceCtx, tr, promise)
				if err != nil {
					e.logger.Warn("skipping entity export for span", zap.Error(err))
					continue
				}

				wg.Add(1)
			}
		}
	}

	wg.Wait()
	return firstErr
}

func (e *explorVizExporter) consumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	var (
		wg       sync.WaitGroup
		once     sync.Once
		firstErr error
	)
	produceCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	promise := func(r *kgo.Record, err error) {
		defer wg.Done()

		if err != nil {
			once.Do(func() {
				firstErr = err
				cancel()
			})
		}
	}

loop:
	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)
		res := rm.Resource()
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			scope := sm.Scope()
			for k := 0; k < sm.Metrics().Len(); k++ {
				select {
				case <-produceCtx.Done():
					break loop
				default:
				}

				m := sm.Metrics().At(k)

				getReader := func(attrs pcommon.Map) attrib.TelemetryReader {
					return attrib.TelemetryReader{
						Attrs:    &attrs,
						Scope:    &scope,
						Resource: &res,
						Metric:   &m,
					}
				}

				switch m.Type() {
				case pmetric.MetricTypeGauge:
					for l := 0; l < m.Gauge().DataPoints().Len(); l++ {
						dp := m.Gauge().DataPoints().At(l)
						tr := getReader(dp.Attributes())
						err := e.consumeReader(produceCtx, tr, promise)
						if err != nil {
							e.logger.Warn("skipping entity export for gauge datapoint", zap.Error(err))
							continue
						}
						wg.Add(1)
					}

				case pmetric.MetricTypeSum:
					for l := 0; l < m.Sum().DataPoints().Len(); l++ {
						dp := m.Sum().DataPoints().At(l)
						tr := getReader(dp.Attributes())
						err := e.consumeReader(produceCtx, tr, promise)
						if err != nil {
							e.logger.Warn("skipping entity export for sum datapoint", zap.Error(err))
							continue
						}
						wg.Add(1)
					}
				case pmetric.MetricTypeHistogram:
					for l := 0; l < m.Histogram().DataPoints().Len(); l++ {
						dp := m.Histogram().DataPoints().At(l)
						tr := getReader(dp.Attributes())
						err := e.consumeReader(produceCtx, tr, promise)
						if err != nil {
							e.logger.Warn("skipping entity export for histogram datapoint", zap.Error(err))
							continue
						}
						wg.Add(1)
					}
				case pmetric.MetricTypeExponentialHistogram:
					for l := 0; l < m.ExponentialHistogram().DataPoints().Len(); l++ {
						dp := m.ExponentialHistogram().DataPoints().At(l)
						tr := getReader(dp.Attributes())
						err := e.consumeReader(produceCtx, tr, promise)
						if err != nil {
							e.logger.Warn("skipping entity export for exponential histogram datapoint", zap.Error(err))
							continue
						}
						wg.Add(1)
					}
				}
			}
		}
	}

	wg.Wait()
	return firstErr
}

func (e *explorVizExporter) consumeLogs(ctx context.Context, ld plog.Logs) error {
	var (
		wg       sync.WaitGroup
		once     sync.Once
		firstErr error
	)
	produceCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	promise := func(r *kgo.Record, err error) {
		defer wg.Done()

		if err != nil {
			once.Do(func() {
				firstErr = err
				cancel()
			})
		}
	}

loop:
	for i := 0; i < ld.ResourceLogs().Len(); i++ {
		rl := ld.ResourceLogs().At(i)
		for j := 0; j < rl.ScopeLogs().Len(); j++ {
			sl := rl.ScopeLogs().At(j)
			for k := 0; k < sl.LogRecords().Len(); k++ {
				select {
				case <-produceCtx.Done():
					break loop
				default:
				}

				log := sl.LogRecords().At(k)
				attrs := log.Attributes()
				scope := sl.Scope()
				res := rl.Resource()

				tr := attrib.TelemetryReader{
					Attrs:    &attrs,
					Scope:    &scope,
					Resource: &res,
					Log:      &log,
				}

				err := e.consumeReader(produceCtx, tr, promise)
				if err != nil {
					e.logger.Warn("skipping entity export for log", zap.Error(err))
					continue
				}

				wg.Add(1)
			}
		}
	}

	wg.Wait()
	return firstErr
}

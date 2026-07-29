package explorvizexporter

import (
	"context"
	"fmt"
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/ExplorViz/otel-collector/common/attrib"
	"github.com/ExplorViz/otel-collector/common/parsing"
	"github.com/ExplorViz/otel-collector/exporter/explorvizexporter/internal/encoding"
)

type explorVizExporter struct {
	logger     *zap.Logger
	client     *kgo.Client
	seedBroker string
	topic      string
}

func newExplorVizExporter(cfg *Config, log *zap.Logger) explorVizExporter {
	return explorVizExporter{
		logger:     log,
		seedBroker: cfg.Broker,
		topic:      cfg.Topic,
	}
}

func (e *explorVizExporter) start(ctx context.Context, host component.Host) error {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(e.seedBroker),
		kgo.DefaultProduceTopic(e.topic),
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

				attrs := ss.Spans().At(k).Attributes()
				scope := ss.Scope()
				res := rs.Resource()

				entity, err := parsing.FromAttributes(attrs)
				if err != nil {
					e.logger.Warn("failed to reconstruct entity from descriptor map, skipping export", zap.Error(err))
					continue
				}

				tr := attrib.TelemetryReader{Attrs: &attrs, Scope: &scope, Resource: &res}

				pb, err := encoding.ToProtobuf(tr, entity)
				if err != nil {
					e.logger.Warn("failed to convert to protobuf message, skipping export", zap.Error(err))
					continue
				}

				out, err := proto.Marshal(pb)
				if err != nil {
					e.logger.Warn("failed to encode protobuf, skipping export", zap.Error(err))
					continue
				}

				wg.Add(1)
				e.client.Produce(produceCtx, &kgo.Record{
					Key:   []byte(pb.GetLandscapeTokenId()),
					Value: out,
				}, promise)
			}
		}
	}

	wg.Wait()
	return firstErr
}

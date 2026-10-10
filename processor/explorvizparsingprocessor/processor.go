package explorvizparsingprocessor

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"

	"github.com/cespare/xxhash/v2"

	"github.com/ExplorViz/otel-collector/common/attrib"
	"github.com/ExplorViz/otel-collector/common/parsing"
	"github.com/ExplorViz/otel-collector/common/token"
)

type tokenValidatorExtension interface {
	Validator() token.Validator
}

type parsingProcessor struct {
	logger         *zap.Logger
	tokenValidator token.Validator
}

func newParsingProcessor(cfg *Config, log *zap.Logger) parsingProcessor {
	p := parsingProcessor{
		logger: log,
	}
	if !cfg.ValidateTokens {
		p.tokenValidator = token.NoOpValidator{}
	}
	return p
}

func (p *parsingProcessor) Start(ctx context.Context, host component.Host) error {
	if p.tokenValidator == nil {
		extID := component.MustNewID("explorviz_token_validator")
		if ext, ok := host.GetExtensions()[extID]; ok {
			if tokenExt, ok := ext.(tokenValidatorExtension); ok {
				p.tokenValidator = tokenExt.Validator()
			} else {
				p.logger.Warn("extensions found but does not conform to interface, landscape token validation is disabled")
				p.tokenValidator = token.NoOpValidator{}
			}
		} else {
			p.logger.Warn("extensions not supported by host, landscape token validation is disabled")
			p.tokenValidator = token.NoOpValidator{}
		}
	}
	return nil
}

func (p *parsingProcessor) processReader(tr attrib.TelemetryReader) error {
	if err := p.validate(tr); err != nil {
		return fmt.Errorf("received invalid reader: %w", err)
	}

	entity, err := parsing.ParseTelemetry(tr)
	if err != nil {
		return fmt.Errorf("failed to parse reader: %w", err)
	}

	entity.ToAttributes(tr.Attrs)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, xxhash.Sum64String(entity.ID()))
	tr.Attrs.PutStr(string(attrib.ExplorVizAttributes.EntityID.Key), hex.EncodeToString(buf))
	binary.BigEndian.PutUint64(buf, xxhash.Sum64String(entity.TelemetryKey()))
	tr.Attrs.PutStr(string(attrib.ExplorVizAttributes.TelemetryKey.Key), hex.EncodeToString(buf))

	return nil
}

func (p *parsingProcessor) processTraces(ctx context.Context, td ptrace.Traces) (ptrace.Traces, error) {
	for i := 0; i < td.ResourceSpans().Len(); i++ {
		rs := td.ResourceSpans().At(i)
		for j := 0; j < rs.ScopeSpans().Len(); j++ {
			ss := rs.ScopeSpans().At(j)
			for k := 0; k < ss.Spans().Len(); k++ {
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

				err := p.processReader(tr)
				if err != nil {
					p.logger.Debug("failed to process span", zap.Error(err))
				}
			}
		}
	}
	return td, nil
}

func (p *parsingProcessor) processMetrics(ctx context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)
		res := rm.Resource()
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			scope := sm.Scope()
			for k := 0; k < sm.Metrics().Len(); k++ {
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
						err := p.processReader(tr)
						if err != nil {
							p.logger.Debug("failed to process gauge metric datapoint", zap.Error(err))
						}
					}

				case pmetric.MetricTypeSum:
					for l := 0; l < m.Sum().DataPoints().Len(); l++ {
						dp := m.Sum().DataPoints().At(l)
						tr := getReader(dp.Attributes())
						err := p.processReader(tr)
						if err != nil {
							p.logger.Debug("failed to process sum metric datapoint", zap.Error(err))
						}
					}
				case pmetric.MetricTypeHistogram:
					for l := 0; l < m.Histogram().DataPoints().Len(); l++ {
						dp := m.Histogram().DataPoints().At(l)
						tr := getReader(dp.Attributes())
						err := p.processReader(tr)
						if err != nil {
							p.logger.Debug("failed to process histogram metric datapoint", zap.Error(err))
						}
					}
				case pmetric.MetricTypeExponentialHistogram:
					for l := 0; l < m.ExponentialHistogram().DataPoints().Len(); l++ {
						dp := m.ExponentialHistogram().DataPoints().At(l)
						tr := getReader(dp.Attributes())
						err := p.processReader(tr)
						if err != nil {
							p.logger.Debug("failed to process exponential histogram metric datapoint", zap.Error(err))
						}
					}
				}
			}
		}
	}
	return md, nil
}

func (p *parsingProcessor) processLogs(ctx context.Context, ld plog.Logs) (plog.Logs, error) {
	for i := 0; i < ld.ResourceLogs().Len(); i++ {
		rl := ld.ResourceLogs().At(i)
		for j := 0; j < rl.ScopeLogs().Len(); j++ {
			sl := rl.ScopeLogs().At(j)
			for k := 0; k < sl.LogRecords().Len(); k++ {
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

				err := p.processReader(tr)
				if err != nil {
					p.logger.Debug("failed to process log", zap.Error(err))
				}
			}
		}
	}
	return ld, nil
}

func (p *parsingProcessor) validate(tr attrib.TelemetryReader) error {
	t := token.LandscapeToken{ID: tr.LandscapeTokenID(), Secret: tr.LandscapeTokenSecret()}

	// A landscape token ID is always required as we otherwise cannot match data to any landscape.
	// Whether a secret is required depends on whether token validation is enabled.
	if t.ID == "" {
		return errors.New("empty or missing landscape token ID attribute")
	}

	if err := p.tokenValidator.Validate(t); err != nil {
		return err
	}

	return nil
}

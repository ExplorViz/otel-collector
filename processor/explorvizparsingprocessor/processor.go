package explorvizparsingprocessor

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"

	"go.opentelemetry.io/collector/component"
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

func (p *parsingProcessor) processTraces(ctx context.Context, td ptrace.Traces) (ptrace.Traces, error) {
	for i := 0; i < td.ResourceSpans().Len(); i++ {
		rs := td.ResourceSpans().At(i)
		for j := 0; j < rs.ScopeSpans().Len(); j++ {
			ss := rs.ScopeSpans().At(j)
			for k := 0; k < ss.Spans().Len(); k++ {
				span := ss.Spans().At(k)
				scope := ss.Scope()
				res := rs.Resource()

				attrs := span.Attributes()
				tr := attrib.TelemetryReader{
					Attrs:    &attrs,
					Scope:    &scope,
					Resource: &res,
				}

				if err := p.validate(tr); err != nil {
					p.logger.Debug("received invalid span", zap.Error(err))
					continue
				}

				entity, err := parsing.ParseTelemetry(tr)
				if err != nil {
					p.logger.Debug("failed to parse span", zap.Error(err))
					continue
				}

				entity.ToAttributes(&attrs)
				buf := make([]byte, 8)
				binary.BigEndian.PutUint64(buf, xxhash.Sum64String(entity.ID()))
				attrs.PutStr(string(attrib.ExplorVizAttributes.EntityID.Key), hex.EncodeToString(buf))
				binary.BigEndian.PutUint64(buf, xxhash.Sum64String(entity.VizObjectID()))
				attrs.PutStr(string(attrib.ExplorVizAttributes.VizObjectID.Key), hex.EncodeToString(buf))
			}
		}
	}
	return td, nil
}

func (p *parsingProcessor) processMetrics(ctx context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			for k := 0; k < sm.Metrics().Len(); k++ {
				m := sm.Metrics().At(k)
				_ = m
				// TODO
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
				scope := sl.Scope()
				res := rl.Resource()

				attrs := log.Attributes()
				tr := attrib.TelemetryReader{
					Attrs:    &attrs,
					Scope:    &scope,
					Resource: &res,
				}

				if err := p.validate(tr); err != nil {
					p.logger.Debug("received invalid log", zap.Error(err))
					continue
				}

				entity, err := parsing.ParseTelemetry(tr)
				if err != nil {
					p.logger.Debug("failed to parse log", zap.Error(err))
					continue
				}

				entity.ToAttributes(&attrs)
				buf := make([]byte, 8)
				binary.BigEndian.PutUint64(buf, xxhash.Sum64String(entity.ID()))
				attrs.PutStr(string(attrib.ExplorVizAttributes.EntityID.Key), hex.EncodeToString(buf))
				binary.BigEndian.PutUint64(buf, xxhash.Sum64String(entity.VizObjectID()))
				attrs.PutStr(string(attrib.ExplorVizAttributes.VizObjectID.Key), hex.EncodeToString(buf))
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

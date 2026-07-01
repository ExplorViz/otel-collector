package explorvizparsingprocessor

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"

	"github.com/ExplorViz/otel-collector/common/attrib"
	"github.com/ExplorViz/otel-collector/common/parsing"
	"github.com/ExplorViz/otel-collector/common/token"
	"github.com/ExplorViz/otel-collector/common/trace"
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

				sr := trace.SpanReader{
					Span:     &span,
					Scope:    &scope,
					Resource: &res,
				}

				if err := p.validateSpan(sr); err != nil {
					p.logger.Debug("received invalid span", zap.Error(err))
					continue
				}

				se, err := parsing.ParseSpan(sr)
				if err != nil {
					p.logger.Debug("failed to parse span", zap.Error(err))
					continue
				}

				sr.Span.Attributes().PutStr(string(attrib.ExplorVizAttributes.EntityId.Key), se.Id())
				m := sr.Span.Attributes().PutEmptyMap(string(attrib.ExplorVizAttributes.EntityDescriptor.Key))
				se.ToMap().CopyTo(m)
			}
		}
	}
	return td, nil
}

func (p *parsingProcessor) validateSpan(sr trace.SpanReader) error {
	t := token.LandscapeToken{ID: sr.LandscapeTokenID(), Secret: sr.LandscapeTokenSecret()}

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

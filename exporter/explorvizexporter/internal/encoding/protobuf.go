package encoding

import (
	"errors"
	"fmt"

	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	"github.com/ExplorViz/otel-collector/common/attrib"
	"github.com/ExplorViz/otel-collector/common/genproto/spanpb"
	"github.com/ExplorViz/otel-collector/common/parsing"
	"github.com/ExplorViz/otel-collector/common/trace"
)

func ToProtobuf(sr trace.SpanReader, se parsing.SpanEntity) (*spanpb.ParsedSpan, error) {
	if se == nil {
		return &spanpb.ParsedSpan{}, errors.New("protobuf conversion: encountered nil entity")
	}

	appName := sr.ResourceStrAttrib(semconv.ServiceNameKey)
	if appName == "" {
		appName = attrib.FallbackValues.ServiceName
	}

	s := spanpb.ParsedSpan{
		LandscapeTokenId:     sr.LandscapeTokenID(),
		LandscapeTokenSecret: sr.LandscapeTokenSecret(),

		TraceId:  sr.TraceID(),
		SpanId:   sr.SpanID(),
		SpanName: sr.Span.Name(),
		ParentId: strOrNil(sr.ParentSpanID()),

		StartTime: uint64(sr.Span.StartTimestamp()),
		EndTime:   uint64(sr.Span.EndTimestamp()),

		ApplicationName: appName,

		EntityId:      se.VizObjectID(),
		GitCommitHash: strOrNil(sr.GitCommitHash()),
	}

	switch e := se.(type) {
	case parsing.CodeSpanEntity:
		s.EntityDescriptor = &spanpb.ParsedSpan_CodeDescriptor{
			CodeDescriptor: &spanpb.CodeDescriptor{
				FilePath:     e.FilePath,
				FunctionName: e.FuncName,
				ClassName:    strOrNil(e.ClassName),
				Language:     strOrNil(e.Language),
			},
		}
	default:
		return &spanpb.ParsedSpan{}, fmt.Errorf("protobuf conversion: encountered unhandled span entity type")
	}

	return &s, nil
}

func strOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

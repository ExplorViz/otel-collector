package encoding

import (
	"errors"
	"fmt"

	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	"github.com/ExplorViz/otel-collector/common/attrib"
	"github.com/ExplorViz/otel-collector/common/genproto/telemetrypb"
	"github.com/ExplorViz/otel-collector/common/parsing"
)

func ToProtobuf(tr attrib.TelemetryReader, se parsing.Entity) (*telemetrypb.TelemetryEntity, error) {
	if se == nil {
		return &telemetrypb.TelemetryEntity{}, errors.New("protobuf conversion: encountered nil entity")
	}

	entityID := tr.StrAttrib(attrib.ExplorVizAttributes.EntityID.Key)
	vizObjectID := tr.StrAttrib(attrib.ExplorVizAttributes.VizObjectID.Key)

	s := telemetrypb.TelemetryEntity{
		LandscapeTokenId:     tr.LandscapeTokenID(),
		LandscapeTokenSecret: tr.LandscapeTokenSecret(),

		GitCommitHash: strOrNil(tr.GitCommitHash()),
	}

	switch e := se.(type) {
	case parsing.CodeEntity:
		appName := tr.ResourceStrAttrib(semconv.ServiceNameKey)
		if appName == "" {
			appName = attrib.FallbackValues.ServiceName
		}

		s.EntityDescriptor = &telemetrypb.TelemetryEntity_CodeDescriptor{
			CodeDescriptor: &telemetrypb.CodeDescriptor{
				ApplicationName: appName,

				FileId:   vizObjectID,
				FilePath: e.FilePath,

				FunctionId:   entityID,
				FunctionName: e.FuncName,

				ClassName: strOrNil(e.ClassName),
				Language:  strOrNil(e.Language),
			},
		}
	default:
		return &telemetrypb.TelemetryEntity{}, fmt.Errorf("protobuf conversion: encountered unhandled entity type")
	}

	return &s, nil
}

func strOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

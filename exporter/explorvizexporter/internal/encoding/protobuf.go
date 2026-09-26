package encoding

import (
	"errors"
	"fmt"

	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	"github.com/ExplorViz/otel-collector/common/attrib"
	"github.com/ExplorViz/otel-collector/common/genproto/telemetrypb"
	"github.com/ExplorViz/otel-collector/common/parsing"
)

func ToProtobuf(tr attrib.TelemetryReader, entity parsing.Entity) (*telemetrypb.TelemetryEntity, error) {
	if entity == nil {
		return &telemetrypb.TelemetryEntity{}, errors.New("protobuf conversion: encountered nil entity")
	}

	entityID := tr.StrAttrib(attrib.ExplorVizAttributes.EntityID.Key)
	telemetryKey := tr.StrAttrib(attrib.ExplorVizAttributes.TelemetryKey.Key)

	te := telemetrypb.TelemetryEntity{
		LandscapeTokenId:     tr.LandscapeTokenID(),
		LandscapeTokenSecret: tr.LandscapeTokenSecret(),

		InstrumentationScope: tr.Scope.Name(),

		GitCommitHash: strOrNil(tr.GitCommitHash()),
	}

	switch e := entity.(type) {
	case parsing.CodeEntity:
		appName := tr.ResourceStrAttrib(semconv.ServiceNameKey)
		if appName == "" {
			appName = attrib.FallbackValues.ServiceName
		}

		te.EntityDescriptor = &telemetrypb.TelemetryEntity_CodeDescriptor{
			CodeDescriptor: &telemetrypb.CodeDescriptor{
				ApplicationName: appName,

				FileTelemetryKey: telemetryKey,
				FilePath:         e.FilePath,

				FunctionTelemetryKey: entityID,
				FunctionName:         e.FuncName,

				ClassName: strOrNil(e.ClassName),
				Language:  strOrNil(e.Language),
			},
		}
	case parsing.RPCServerEntity:
		te.EntityDescriptor = &telemetrypb.TelemetryEntity_RpcServerDescriptor{
			RpcServerDescriptor: &telemetrypb.RpcServerDescriptor{
				ApplicationName:     e.ApplicationName,
				ServiceTelemetryKey: telemetryKey,
				ServiceName:         e.ServiceName,
				MethodName:          e.MethodName,
				MethodTelemetryKey:  entityID,
				SystemName:          strOrNil(e.SystemName),
			},
		}
	case parsing.RPCClientEntity:
		te.EntityDescriptor = &telemetrypb.TelemetryEntity_RpcClientDescriptor{
			RpcClientDescriptor: &telemetrypb.RpcClientDescriptor{
				ApplicationName: e.ApplicationName,
				TelemetryKey:    telemetryKey,
				SystemName:      strOrNil(e.SystemName),
			},
		}
	case parsing.HTTPServerEntity:
		te.EntityDescriptor = &telemetrypb.TelemetryEntity_HttpServerDescriptor{
			HttpServerDescriptor: &telemetrypb.HttpServerDescriptor{
				ApplicationName: e.ServiceName,
				TelemetryKey:    telemetryKey,
				Route:           e.Route,
				Method:          e.Method,
			},
		}
	case parsing.HTTPClientEntity:
		te.EntityDescriptor = &telemetrypb.TelemetryEntity_HttpClientDescriptor{
			HttpClientDescriptor: &telemetrypb.HttpClientDescriptor{
				ApplicationName: e.ServiceName,
				TelemetryKey:    telemetryKey,
			},
		}
	case parsing.GenericEntity:
		te.EntityDescriptor = &telemetrypb.TelemetryEntity_GenericEntityDescriptor{
			GenericEntityDescriptor: &telemetrypb.GenericEntityDescriptor{
				ServiceName:  e.ServiceName,
				Name:         e.Name,
				TelemetryKey: telemetryKey,
			},
		}
	default:
		return &telemetrypb.TelemetryEntity{}, fmt.Errorf("protobuf conversion: encountered unhandled entity type")
	}

	return &te, nil
}

func strOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

package parsing

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"

	"github.com/ExplorViz/otel-collector/common/attrib"
)

const RPCServerEntityType string = "rpcserver"

// An RPCServerEntity represents a method within a server's remote procedure call interface (e.g. gRPC).
type RPCServerEntity struct {
	// Name of the application or service hosting the RPC server.
	ApplicationName string

	// Fully-qualified name of the logical RPC service containing the called method.
	// Package names before the service name must be separated using periods (".").
	ServiceName string

	// Unqualified name of the called method. This is the name from the RPC interface perspective,
	// meaning this does not necessarily correspond to the name of the implementing function / method.
	MethodName string

	// Name of the used RPC system, e.g. "grpc", "dubbo".
	SystemName string
}

func (r RPCServerEntity) ID() string {
	return "rpcserver" + "|" + r.ApplicationName + "|" + r.SystemName + "|" + r.ServiceName + "|" + r.MethodName
}

func (r RPCServerEntity) TelemetryKey() string {
	return "rpcserver" + "|" + r.ApplicationName + "|" + r.SystemName + "|" + r.ServiceName
}

func (r RPCServerEntity) ToAttributes(attrs *pcommon.Map) {
	attrs.PutStr(string(attrib.ExplorVizAttributes.EntityType.Key), RPCServerEntityType)
	attrs.PutStr(string(attrib.ExplorVizAttributes.ServiceName.Key), r.ApplicationName)
	attrs.PutStr(string(attrib.ExplorVizAttributes.RPCServiceName.Key), r.ServiceName)
	attrs.PutStr(string(attrib.ExplorVizAttributes.RPCMethodName.Key), r.MethodName)
	attrs.PutStr(string(attrib.ExplorVizAttributes.RPCSystemName.Key), r.SystemName)
}

// rpcServerEntityFromAttribs initializes a new [RPCServerEntity] based on the entries of the provided map.
// If the provided map entries are incomplete (meaning a mandatory attribute is missing),
// then a zero-initialized RPCServerEntity and an error is returned.
func rpcServerEntityFromAttribs(m pcommon.Map) (RPCServerEntity, error) {
	app, ok := m.Get(string(attrib.ExplorVizAttributes.ServiceName.Key))
	if !ok || app.Str() == "" {
		return RPCServerEntity{}, errors.New("empty or missing string attribute for application name")
	}

	service, ok := m.Get(string(attrib.ExplorVizAttributes.RPCServiceName.Key))
	if !ok || service.Str() == "" {
		return RPCServerEntity{}, errors.New("empty or missing string attribute for rpc service name")
	}

	method, ok := m.Get(string(attrib.ExplorVizAttributes.RPCMethodName.Key))
	if !ok || method.Str() == "" {
		return RPCServerEntity{}, errors.New("empty or missing string attribute for rpc method name")
	}

	system, _ := m.Get(string(attrib.ExplorVizAttributes.RPCSystemName.Key))

	return RPCServerEntity{
		ApplicationName: app.Str(),
		ServiceName:     service.Str(),
		MethodName:      method.Str(),
		SystemName:      system.Str(),
	}, nil
}

const RPCClientEntityType string = "rpcclient"

// An RPCClientEntity represents a client making remote procedure calls to a server (e.g. gRPC).
type RPCClientEntity struct {
	// Name of the application or service hosting the RPC server.
	ApplicationName string

	// Name of the instrumentation scope to which the entity belongs
	ScopeName string

	// Name of the used RPC system, e.g. "grpc", "dubbo".
	SystemName string
}

func (r RPCClientEntity) ID() string {
	return "rpcclient" + "|" + r.ApplicationName + "|" + r.ScopeName + "|" + r.SystemName
}

func (r RPCClientEntity) TelemetryKey() string {
	return r.ID()
}

func (r RPCClientEntity) ToAttributes(attrs *pcommon.Map) {
	attrs.PutStr(string(attrib.ExplorVizAttributes.EntityType.Key), RPCClientEntityType)
	attrs.PutStr(string(attrib.ExplorVizAttributes.ServiceName.Key), r.ApplicationName)
	attrs.PutStr(string(attrib.ExplorVizAttributes.ScopeName.Key), r.ScopeName)
	attrs.PutStr(string(attrib.ExplorVizAttributes.RPCSystemName.Key), r.SystemName)
}

// rpcClientEntityFromAttribs initializes a new [RPCClientEntity] based on the entries of the provided map.
// If the provided map entries are incomplete (meaning a mandatory attribute is missing),
// then a zero-initialized RPCClientEntity and an error is returned.
func rpcClientEntityFromAttribs(m pcommon.Map) (RPCClientEntity, error) {
	app, ok := m.Get(string(attrib.ExplorVizAttributes.ServiceName.Key))
	if !ok || app.Str() == "" {
		return RPCClientEntity{}, errors.New("empty or missing string attribute for application name")
	}

	scope, ok := m.Get(string(attrib.ExplorVizAttributes.ScopeName.Key))
	if !ok || scope.Str() == "" {
		return RPCClientEntity{}, errors.New("empty or missing string attribute for scope name")
	}

	system, _ := m.Get(string(attrib.ExplorVizAttributes.RPCSystemName.Key))

	return RPCClientEntity{
		ApplicationName: app.Str(),
		ScopeName:       scope.Str(),
		SystemName:      system.Str(),
	}, nil
}

// ParseRPCTelemetry parses telemetry describing remote procedure calls by looking for attributes conforming
// to the [OTel semconv RPC attributes]. For spans of kind [ptrace.SpanKindClient], an [RPCClientEntity] will
// be extracted. Otherwise, an [RPCServerEntity] is parsed.
//
// # Server entities
//
// For telemetry to be successfully parsed as a server entity, it needs to provide:
//   - a name for the application or service hosting the RPC server
//   - a name for the logical RPC service within which the RPC method resides
//   - a name of the called RPC method
//
// If an explicit name for the RPC service is provided, it will be used. Otherwise, the name will be
// extracted from the method name, which should always be fully-qualified according to OTel semconv.
// In this case, the method name is stripped to exclude any service or package information.
//
// # Client entities
//
// For telemetry to be successfully parsed as a client entity, it needs to provide:
//   - a service name to which to match the RPC client
//   - an instrumentation scope name
//   - attributes from the [OTel semconv RPC attributes] indicating it represents an RPC call
//
// [OTel semconv RPC attributes]: https://opentelemetry.io/docs/specs/semconv/registry/attributes/rpc/
func ParseRPCTelemetry(tr attrib.TelemetryReader) (Entity, error) {
	if tr.Span != nil && tr.Span.Kind() == ptrace.SpanKindClient {
		slog.Debug("rpc parser: encountered client span, attempting to parse as client")
		return parseRPCClientTelemetry(tr)
	}
	return parseRPCServerTelemetry(tr)
}

func parseRPCServerTelemetry(tr attrib.TelemetryReader) (Entity, error) {
	app := tr.ResourceStrAttrib(semconv.ServiceNameKey)
	if app == "" {
		return RPCServerEntity{}, errors.New("rpc parser: empty or missing application name attribute")
	}

	var service, method string

	if serviceAttrib := tr.StrAttrib("rpc.service"); serviceAttrib != "" {
		// If the now-deprecated "rpc.service" attribute is provided, we use it as the service name.
		// For such older versions of semconv, the "rpc.method" attribute will not be fully-qualified.
		service = serviceAttrib
		method = tr.StrAttrib("rpc.method")
		if method == "" {
			return RPCServerEntity{}, errors.New("rpc parser: empty or missing method name attribute")
		}
	} else {
		// Newer versions of semconv specify the method name as a fully-qualified name including the service.
		// We therefore extract service fully-qualified name and the method's name from the method fqn.
		// Example of the method fqn format: "oteldemo.ProductCatalogService/GetProduct"
		methodFqn := tr.StrAttrib(semconv.RPCMethodKey)
		if methodFqn == "" {
			// Attempt to use gRPC's own semantic conventions as a fallback.
			// See https://grpc.io/docs/guides/opentelemetry-metrics/
			methodFqn = tr.StrAttrib("grpc.method")
		}
		if methodFqn == "" {
			return RPCServerEntity{}, errors.New("rpc parser: empty or missing method name attribute")
		}
		// Some of the method attribute values in the OTel Demo start with "/" for some reason
		methodFqn = strings.TrimPrefix(methodFqn, "/")
		if strings.Count(methodFqn, "/") > 1 {
			return RPCServerEntity{}, fmt.Errorf("rpc parser: unknown method fqn format %s", methodFqn)
		}

		var found bool
		service, method, found = strings.Cut(methodFqn, "/")
		if !found {
			i := strings.LastIndex(service, ".")
			if i == -1 {
				return RPCServerEntity{}, errors.New("rpc parser: failed to extract rpc service name from method fqn")
			}
			method = service[i+1:]
			service = service[:i]
		}

		if method == "" {
			return RPCServerEntity{}, errors.New("rpc parser: failed to extract rpc method name from method fqn")
		}
	}

	system := tr.StrAttrib(semconv.RPCSystemNameKey)
	if system == "" {
		// Try to use deprecated attribute as fallback
		system = tr.StrAttrib("rpc.system")
	}

	return RPCServerEntity{
		ApplicationName: app,
		ServiceName:     service,
		MethodName:      method,
		SystemName:      system,
	}, nil
}

func parseRPCClientTelemetry(tr attrib.TelemetryReader) (RPCClientEntity, error) {
	service := tr.ResourceStrAttrib(semconv.ServiceNameKey)
	if service == "" {
		return RPCClientEntity{}, errors.New("rpc client parser: empty or missing service name attribute")
	}

	isRPCCall := tr.HasAnyAttrKey([]attribute.Key{
		semconv.RPCMethodKey,
		semconv.RPCMethodOriginalKey,
		semconv.RPCSystemNameKey,

		// Deprecated attributes
		attribute.Key("rpc.message.id"),
		attribute.Key("rpc.service"),
		attribute.Key("rpc.system"),
	})

	if !isRPCCall {
		return RPCClientEntity{}, errors.New("rpc client parser: no identifying rpc attributes found")
	}

	system := tr.StrAttrib(semconv.RPCSystemNameKey)
	if system == "" {
		// Try to use deprecated attribute as fallback
		system = tr.StrAttrib("rpc.system")
	}

	scope := tr.Scope.Name()

	return RPCClientEntity{
		ApplicationName: service,
		ScopeName:       scope,
		SystemName:      system,
	}, nil
}

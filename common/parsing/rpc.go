package parsing

import (
	"errors"
	"fmt"
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"

	"github.com/ExplorViz/otel-collector/common/attrib"
)

const RPCEntityType string = "rpc"

// An RPCEntity represents a method within a remote procedure call interface (e.g. gRPC).
type RPCEntity struct {
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

func (r RPCEntity) ID() string {
	return "rpc" + "|" + r.ApplicationName + "|" + r.SystemName + "|" + r.ServiceName + "|" + r.MethodName
}

func (r RPCEntity) TelemetryKey() string {
	return "rpc" + "|" + r.ApplicationName + "|" + r.SystemName + "|" + r.ServiceName
}

func (r RPCEntity) ToAttributes(attrs *pcommon.Map) {
	attrs.PutStr(string(attrib.ExplorVizAttributes.EntityType.Key), RPCEntityType)
	attrs.PutStr(string(attrib.ExplorVizAttributes.ServiceName.Key), r.ApplicationName)
	attrs.PutStr(string(attrib.ExplorVizAttributes.RPCServiceName.Key), r.ServiceName)
	attrs.PutStr(string(attrib.ExplorVizAttributes.RPCMethodName.Key), r.MethodName)
	attrs.PutStr(string(attrib.ExplorVizAttributes.RPCSystemName.Key), r.SystemName)
}

// rpcEntityFromAttribs initializes a new [RPCEntity] based on the entries of the provided map.
// If the provided map entries are incomplete (meaning a mandatory attribute is missing),
// then a zero-initialized RPCEntity and an error is returned.
func rpcEntityFromAttribs(m pcommon.Map) (RPCEntity, error) {
	app, ok := m.Get(string(attrib.ExplorVizAttributes.ServiceName.Key))
	if !ok || app.Str() == "" {
		return RPCEntity{}, errors.New("empty or missing string attribute for application name")
	}

	service, ok := m.Get(string(attrib.ExplorVizAttributes.RPCServiceName.Key))
	if !ok || service.Str() == "" {
		return RPCEntity{}, errors.New("empty or missing string attribute for rpc service name")
	}

	method, ok := m.Get(string(attrib.ExplorVizAttributes.RPCMethodName.Key))
	if !ok || method.Str() == "" {
		return RPCEntity{}, errors.New("empty or missing string attribute for rpc method name")
	}

	system, _ := m.Get(string(attrib.ExplorVizAttributes.RPCSystemName.Key))

	return RPCEntity{
		ApplicationName: app.Str(),
		ServiceName:     service.Str(),
		MethodName:      method.Str(),
		SystemName:      system.Str(),
	}, nil
}

// ParseRPCTelemetry parses telemetry describing remote procedure calls by looking for attributes conforming
// to the [OTel semconv RPC attributes]. For telemetry to be successfully parsed, it needs to provide:
//   - a name for the application or service hosting the RPC server
//   - a name for the logical RPC service within which the RPC method resides
//   - a name of the called RPC method
//
// If an explicit name for the RPC service is provided, it will be used. Otherwise, the name will be
// extracted from the method name, which should always be fully-qualified according to OTel semconv.
// In this case, the method name is stripped to exclude any service or package information.
//
// [OTel semconv RPC attributes]: https://opentelemetry.io/docs/specs/semconv/registry/attributes/rpc/
func ParseRPCTelemetry(tr attrib.TelemetryReader) (Entity, error) {
	app := tr.ResourceStrAttrib(semconv.ServiceNameKey)
	if app == "" {
		return RPCEntity{}, errors.New("rpc parser: empty or missing application name attribute")
	}

	var service, method string

	if serviceAttrib := tr.StrAttrib("rpc.service"); serviceAttrib != "" {
		// If the now-deprecated "rpc.service" attribute is provided, we use it as the service name.
		// For such older versions of semconv, the "rpc.method" attribute will not be fully-qualified.
		service = serviceAttrib
		method = tr.StrAttrib("rpc.method")
		if method == "" {
			return RPCEntity{}, errors.New("rpc parser: empty or missing method name attribute")
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
			return RPCEntity{}, errors.New("rpc parser: empty or missing method name attribute")
		}
		// Some of the method attribute values in the OTel Demo start with "/" for some reason
		methodFqn = strings.TrimPrefix(methodFqn, "/")
		if strings.Count(methodFqn, "/") > 1 {
			return RPCEntity{}, fmt.Errorf("rpc parser: unknown method fqn format %s", methodFqn)
		}

		var found bool
		service, method, found = strings.Cut(methodFqn, "/")
		if !found {
			i := strings.LastIndex(service, ".")
			if i == -1 {
				return RPCEntity{}, errors.New("rpc parser: failed to extract rpc service name from method fqn")
			}
			method = service[i+1:]
			service = service[:i]
		}

		if method == "" {
			return RPCEntity{}, errors.New("rpc parser: failed to extract rpc method name from method fqn")
		}
	}

	system := tr.StrAttrib(semconv.RPCSystemNameKey)

	return RPCEntity{
		ApplicationName: app,
		ServiceName:     service,
		MethodName:      method,
		SystemName:      system,
	}, nil
}

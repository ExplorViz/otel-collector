package parsing

import (
	"cmp"
	"errors"
	"fmt"

	"go.opentelemetry.io/collector/pdata/pcommon"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"

	"github.com/ExplorViz/otel-collector/common/attrib"
)

const CodeEntityType string = "code"

// A CodeEntity represents the execution of a function.
type CodeEntity struct {
	// FilePath is the path of the file within which the function is contained, with "/" as the separator.
	// The path should be a relative path that can uniquely identify the file within the application.
	FilePath string

	// FuncName is the name of the executed function, excluding its signature.
	FuncName string

	// ClassName is the name of the class within which the function is contained (if any).
	// Inner classes should be qualified against containing classes separated with ".", e.g. ("OuterClass.InnerClass").
	ClassName string

	// Language specifies the programming language runtime of the executed function. If applicable, the format
	// should match the well-known values specified by OpenTelemetry's telemetry.sdk.language attribute.
	Language string
}

func (c CodeEntity) ID() string {
	return "function" + "|" + c.FilePath + "|" + c.ClassName + "|" + c.FuncName
}

func (c CodeEntity) VizObjectID() string {
	return "file" + "|" + c.FilePath
}

func (c CodeEntity) ToAttributes(attrs *pcommon.Map) {
	attrs.PutStr(string(attrib.ExplorVizAttributes.EntityType.Key), CodeEntityType)
	attrs.PutStr(string(attrib.ExplorVizAttributes.CodeFilePath.Key), c.FilePath)
	attrs.PutStr(string(attrib.ExplorVizAttributes.CodeFunctionName.Key), c.FuncName)
	attrs.PutStr(string(attrib.ExplorVizAttributes.CodeClassName.Key), c.ClassName)
	attrs.PutStr(string(attrib.ExplorVizAttributes.CodeLanguage.Key), c.Language)
}

// codeEntityFromAttribs initializes a new [CodeEntity] based on the entries of the provided map.
// If the provided map entries are incomplete (meaning a mandatory attribute is missing),
// then a zero-initialized CodeEntity and an error is returned.
func codeEntityFromAttribs(m pcommon.Map) (CodeEntity, error) {
	filePath, ok := m.Get(string(attrib.ExplorVizAttributes.CodeFilePath.Key))
	if !ok || filePath.Str() == "" {
		return CodeEntity{}, errors.New(`empty or missing string attribute for file path`)
	}

	funcName, ok := m.Get(string(attrib.ExplorVizAttributes.CodeFunctionName.Key))
	if !ok || funcName.Str() == "" {
		return CodeEntity{}, errors.New(`empty or missing string attribute for function name`)
	}

	className, _ := m.Get(string(attrib.ExplorVizAttributes.CodeClassName.Key))
	lang, _ := m.Get(string(attrib.ExplorVizAttributes.CodeLanguage.Key))

	return CodeEntity{
		FilePath:  filePath.Str(),
		FuncName:  funcName.Str(),
		ClassName: className.Str(),
		Language:  lang.Str(),
	}, nil
}

// ParseCodeTelemetry parses telemetry describing function executions by looking for attributes conforming to the
// [OTel semconv code attributes]. For telemetry to be successfully parsed, it needs to provide:
//   - a relative path of the file containing the function, ideally relative to the repository root
//   - the name of the executed function
//
// Since relative paths for the file are preferred, it first attempts to parse the function fully-qualified
// name (FQN) for file path information. Only if this is insufficient will it look for an explicit file path.
//
// [OTel semconv code attributes]: https://opentelemetry.io/docs/specs/semconv/registry/attributes/code/
func ParseCodeTelemetry(tr attrib.TelemetryReader) (Entity, error) {
	fqn := tr.StrAttrib(semconv.CodeFunctionNameKey)
	if fqn == "" {
		return &CodeEntity{}, errors.New("code parser: empty or missing function name attribute")
	}

	lang := tr.StrAttrib(semconv.TelemetrySDKLanguageKey)

	parsedFQN := ParseFunctionFQN(fqn, lang)

	if parsedFQN.FuncName == "" {
		return &CodeEntity{}, errors.New("code parser: function name could not be extracted")
	}

	filePath := cmp.Or(parsedFQN.FilePath, tr.StrAttrib(semconv.CodeFilePathKey))

	if filePath == "" {
		return CodeEntity{}, fmt.Errorf("code parser: file path could not be extracted from FQN and %s not given", semconv.CodeFilePathKey)
	}

	return CodeEntity{
		FilePath:  filePath,
		FuncName:  parsedFQN.FuncName,
		ClassName: parsedFQN.ClassName,
		Language:  lang,
	}, nil
}

package parsing

import (
	"cmp"
	"errors"
	"fmt"

	"go.opentelemetry.io/collector/pdata/pcommon"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"

	"github.com/ExplorViz/otel-collector/common/trace"
)

const CodeSpanEntityType string = "code"

// A CodeSpanEntity represents the execution of a function.
type CodeSpanEntity struct {
	// FilePath is the path of the file within which the function is contained, with "/" as the separator.
	FilePath string

	// FuncName is the name of the executed function, excluding its signature.
	FuncName string

	// ClassName is the qualified name of the class within which the function is contained (if any).
	// Inner classes should be separated against containing classes using ".",e.g. ("OuterClass.InnerClass").
	ClassName string

	// Language specifies the programming language runtime of the executed function. If applicable, the format
	// should match the well-known values specified by OpenTelemetry's telemetry.sdk.language attribute.
	Language string

	// GitCommitHash can be specified if the function is contained within a file at a known commit.
	// This can be used to correlate data from runtime analysis with static analysis data.
	GitCommitHash string
}

func (c CodeSpanEntity) Id() string {
	return c.FilePath + " " + c.FuncName + " " + c.ClassName + " " + c.GitCommitHash
}

func (c CodeSpanEntity) ToMap() pcommon.Map {
	m := pcommon.NewMap()
	m.PutStr("type", CodeSpanEntityType)
	m.PutStr("filePath", c.FilePath)
	m.PutStr("funcName", c.FuncName)
	m.PutStr("className", c.ClassName)
	m.PutStr("language", c.Language)
	m.PutStr("gitCommitHash", c.GitCommitHash)
	return m
}

// codeSpanEntityFromMap initializes a new [CodeSpanEntity] based on the entries of the provided map.
// If the provided map entries are incomplete (meaning a mandatory field is missing),
// then a zero-initialized CodeSpanEntity and an error is returned.
func codeSpanEntityFromMap(m pcommon.Map) (CodeSpanEntity, error) {
	filePath, ok := m.Get("filePath")
	if !ok || filePath.Str() == "" {
		return CodeSpanEntity{}, errors.New(`empty or missing string map entry for key "filePath"`)
	}

	funcName, ok := m.Get("funcName")
	if !ok || funcName.Str() == "" {
		return CodeSpanEntity{}, errors.New(`empty or missing string map entry for key "funcName"`)
	}

	className, _ := m.Get("className")
	lang, _ := m.Get("language")
	gitHash, _ := m.Get("gitCommitHash")

	return CodeSpanEntity{
		FilePath:      filePath.Str(),
		FuncName:      funcName.Str(),
		ClassName:     className.Str(),
		Language:      lang.Str(),
		GitCommitHash: gitHash.Str(),
	}, nil
}

// ParseCodeSpan parses spans describing function executions by looking for attributes conforming to the
// [OTel semconv code attributes]. For a span to be successfully parsed, it needs to provide:
//   - a relative path of the file containing the function, ideally relative to the repository root
//   - the name of the executed function
//
// Since relative paths for the file are preferred, it first attempts to parse the function fully-qualified
// name (FQN) for file path information. Only if this is insufficient will it look for an explicit file path.
//
// [OTel semconv code attributes]: https://opentelemetry.io/docs/specs/semconv/registry/attributes/code/
func ParseCodeSpan(sr trace.SpanReader) (SpanEntity, error) {
	fqn := sr.SpanStrAttrib(semconv.CodeFunctionNameKey)
	if fqn == "" {
		return &CodeSpanEntity{}, errors.New("code parser: empty or missing function name attribute")
	}

	lang := sr.SpanStrAttrib(semconv.TelemetrySDKLanguageKey)

	parsedFQN := ParseFunctionFQN(fqn, lang)

	if parsedFQN.FuncName == "" {
		return &CodeSpanEntity{}, errors.New("code parser: function name could not be extracted")
	}

	filePath := cmp.Or(parsedFQN.FilePath, sr.SpanStrAttrib(semconv.CodeFilePathKey))

	if filePath == "" {
		return CodeSpanEntity{}, fmt.Errorf("code parser: file path could not be extracted from FQN and %s not given", semconv.CodeFilePathKey)
	}

	gitHash := sr.SpanStrAttrib(semconv.VCSRefHeadRevisionKey)

	return CodeSpanEntity{
		FilePath:      filePath,
		FuncName:      parsedFQN.FuncName,
		ClassName:     parsedFQN.ClassName,
		Language:      lang,
		GitCommitHash: gitHash}, nil
}

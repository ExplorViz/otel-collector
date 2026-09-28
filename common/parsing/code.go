package parsing

import (
	"cmp"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"unicode"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"

	"github.com/ExplorViz/otel-collector/common/attrib"
)

var extensionToLang = map[string]string{
	".c":    "c",
	".cpp":  "cpp",
	".cs":   "csharp",
	".go":   "go",
	".java": "java",
	".js":   "javascript",
	".kt":   "kotlin",
	".py":   "python",
	".php":  "php",
	".rs":   "rust",
	".ts":   "typescript",
}

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

func (c CodeEntity) TelemetryKey() string {
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
		return CodeEntity{}, errors.New("empty or missing string attribute for file path")
	}

	funcName, ok := m.Get(string(attrib.ExplorVizAttributes.CodeFunctionName.Key))
	if !ok || funcName.Str() == "" {
		return CodeEntity{}, errors.New("empty or missing string attribute for function name")
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

func langFromFilePath(path string) (string, error) {
	lastIndex := strings.LastIndex(path, ".")
	if lastIndex == -1 {
		return "", errors.New("file path does not seem to contain file extension")
	}
	ext := path[lastIndex:]
	lang, ok := extensionToLang[ext]
	if !ok {
		// Return the file extension as a fallback
		return path[lastIndex+1:], nil
	}
	return lang, nil
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
	if !tr.HasAnyAttrKey([]attribute.Key{
		semconv.CodeFunctionNameKey,
		semconv.CodeFilePathKey,
		"code.function",
		"code.filepath",
	}) {
		return CodeEntity{}, errors.New("code parser: no identifying code attributes found")
	}

	lang := tr.ResourceStrAttrib(semconv.TelemetrySDKLanguageKey)
	if lang == "" {
		lang = tr.StrAttrib(semconv.TelemetrySDKLanguageKey)
		if lang == "" {
			lang, _ = langFromFilePath(tr.StrAttrib(semconv.CodeFilePathKey))
		}
	}

	var parseResult codeParseResult
	var err error
	parseFunc, ok := langToParseFunc[strings.ToLower(lang)]
	if ok {
		parseResult, err = parseFunc(tr)
	} else {
		slog.Warn("unknown or missing SDK language, using generic fqn parser", string(semconv.TelemetrySDKLanguageKey), lang)
		parseResult, err = parseGenericCodeTelemetry(tr)
	}
	if err != nil {
		return CodeEntity{}, fmt.Errorf("code parser: %w", err)
	}

	return CodeEntity{
		FilePath:  strings.TrimPrefix(parseResult.FilePath, "/"),
		FuncName:  parseResult.FuncName,
		ClassName: parseResult.ClassName,
		Language:  lang,
	}, nil
}

// A codeParseResult represents the result of extracting function name,
// file path, and class name information from code telemetry attributes.
type codeParseResult struct {
	FilePath  string // Separated by "/"
	FuncName  string // Unqualified
	ClassName string // Inner classes may be specified with "." as separator
}

var langToParseFunc = map[string]func(attrib.TelemetryReader) (codeParseResult, error){
	"java": parseJavaCodeTelemetry,
}

// parseJavaMethodFqn extracts a [codeParseResult] from code telemetry describing Java method executions.
// If possible, the file path will be derived from the fully-qualified name (FQN) of the method, where the
// expected format is that specified by the [OTel semantic conventions]. Attempts to detect inner classes
// if multiple FQN segments are capitalized. Only looks for an explicit file path attribute if no file path
// can otherwise be extracted from the FQN.
//
// Example FQN:
//
//	"net.explorviz.app.MyConverter.MyInnerClass.convert"
//
// Example result:
//
//	["net/explorviz/app/MyConverter.java", "convert", "MyConverter.MyInnerClass"]
//
// [OTel semantic conventions]: https://opentelemetry.io/docs/specs/semconv/registry/attributes/code/
func parseJavaCodeTelemetry(tr attrib.TelemetryReader) (codeParseResult, error) {
	fqn := tr.StrAttrib(semconv.CodeFunctionNameKey)
	if fqn == "" {
		// Use deprecated attributes as fallback
		funcName := tr.StrAttrib("code.function")
		if funcName == "" {
			return codeParseResult{}, errors.New("java code parser: empty or missing attribute for function name")
		}

		if namespace := tr.StrAttrib("code.namespace"); namespace != "" {
			fqn = namespace + "." + funcName
		} else {
			fqn = funcName
		}
	}

	fqnResult := parseJavaMethodFqn(fqn)
	funcName := fqnResult.FuncName
	filePath := fqnResult.FilePath
	className := fqnResult.ClassName
	if funcName == "" {
		return codeParseResult{}, errors.New("java code parser: could not derive function name from fqn")
	}
	if filePath == "" {
		filePath = cmp.Or(tr.StrAttrib(semconv.CodeFilePathKey), tr.StrAttrib("code.filepath"))

		// Derive class name from file name
		splitFilePath := strings.Split(filePath, "/")
		className = strings.TrimSuffix(splitFilePath[len(splitFilePath)-1], ".java")
	}
	if filePath == "" {
		return codeParseResult{}, errors.New("java code parser: could not derive file path name from fqn and no explicit value provided")
	}

	return codeParseResult{FilePath: filePath, FuncName: funcName, ClassName: className}, nil
}

func parseJavaMethodFqn(fqn string) codeParseResult {
	const FileExtension = ".java"

	// Ignore lambda portion. If we explicitly want to support this, the data model would need to allow functions
	// to have child functions. For now, we simplify and treat this as a call of the containing function.
	fqnNoLambda, _, _ := strings.Cut(fqn, "$")
	separatedFQN := strings.Split(fqnNoLambda, ".")

	if len(separatedFQN) == 1 {
		return codeParseResult{FilePath: "", FuncName: fqnNoLambda, ClassName: ""}
	}

	lastPackageIndex := -1
	for i := len(separatedFQN) - 2; i >= 0; i-- {
		if !startsWithUpper(separatedFQN[i]) {
			lastPackageIndex = i
			break
		}
	}

	fileNameIndex := lastPackageIndex + 1
	filePath := strings.Join(separatedFQN[:fileNameIndex+1], "/") + FileExtension

	className := ""
	if fileNameIndex < len(separatedFQN)-1 {
		className = strings.Join(
			separatedFQN[fileNameIndex:len(separatedFQN)-1],
			".",
		)
	}

	methodName := separatedFQN[len(separatedFQN)-1]

	return codeParseResult{FilePath: filePath, FuncName: methodName, ClassName: className}
}

// parseGenericCodeTelemetry acts as a fallback generic parser for when the language is not known.
// If an explicit file path is provided, it is used. Otherwise, an attempt is made to derive it from
// the function fully-qualified name, which is assumed to be a dot-separated string where the last
// segment represents the function name and the prior segments represent the file's path. Assumes no
// classes are included in the FQN. No file extension is added to the file path.
//
// Example FQN:
//
//	"some.namespace.SomeFile.myFunc"
//
// Example result:
//
//	["some/namespace/SomeFile", "myFunc", ""]
func parseGenericCodeTelemetry(tr attrib.TelemetryReader) (codeParseResult, error) {
	fqn := tr.StrAttrib(semconv.CodeFunctionNameKey)
	if fqn == "" {
		// Look for deprecated attributes as fallback
		function := tr.StrAttrib("code.function")
		if function == "" {
			return codeParseResult{}, errors.New("generic code parser: empty or missing attribute for function name")
		}

		if namespace := tr.StrAttrib("code.namespace"); namespace != "" {
			fqn = namespace + "." + function
		} else {
			fqn = function
		}
	}

	separatedFQN := strings.Split(fqn, ".")
	funcName := separatedFQN[len(separatedFQN)-1]
	filePath := cmp.Or(tr.StrAttrib(semconv.CodeFilePathKey), tr.StrAttrib("code.filepath"))
	if filePath == "" {
		if len(separatedFQN) < 2 {
			return codeParseResult{}, errors.New("generic code parser: could not derive file path name from fqn and no explicit value provided")
		}
		filePath = strings.Join(separatedFQN[:len(separatedFQN)-1], "/")
	}

	return codeParseResult{
		FilePath: filePath,
		FuncName: funcName,
	}, nil
}

func startsWithUpper(s string) bool {
	for _, r := range s {
		return unicode.IsUpper(r)
	}
	return false
}

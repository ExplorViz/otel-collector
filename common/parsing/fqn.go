package parsing

import (
	"log/slog"
	"strings"
	"unicode"
)

// An FQNParseResult represents the result of extracting file path and
// class name information from a function's fully-qualified name (FQN).
type FQNParseResult struct {
	FilePath  string
	FuncName  string
	ClassName string
}

// ParseFunctionFQN parses the provided function fully-qualified name (FQN) into an [FQNParseResult].
// Since the representation of FQN values depend on the programming language runtime (see [OTel semantic conventions]),
// the language must additionally be specified. If the provided language is not known, a generic parser is used.
//
// [OTel semantic conventions]: https://opentelemetry.io/docs/specs/semconv/registry/attributes/code/
func ParseFunctionFQN(fqn string, language string) FQNParseResult {
	switch strings.ToLower(language) {
	case "java":
		return parseJavaFQN(fqn)
	default:
		slog.Warn("unknown or missing SDK language, using generic fqn parser", "telemetry.sdk.language", language)
		return parseGenericFQN(fqn)
	}
}

// ParseJavaFQN parses a Java method fully-qualified name (FQN) as specified by the [OTel semantic conventions].
// Attempts to detect inner classes if multiple FQN segments are capitalized.
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
func parseJavaFQN(fqn string) FQNParseResult {
	const FileExtension = ".java"

	// Ignore lambda portion. If we explicitly want to support this, the data model would need to allow functions
	// to have child functions. For now, we simplify and treat this as a call of the containing function.
	fqnNoLambda, _, _ := strings.Cut(fqn, "$")
	separatedFQN := strings.Split(fqnNoLambda, ".")

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

	return FQNParseResult{FilePath: filePath, FuncName: methodName, ClassName: className}
}

// ParseGenericFQN acts as a fallback generic parser for fully-qualified function names where the language is not known.
// Assumes dot-separated string, where the last segment represents the function name and the prior segments represent the
// file's path. Does not add any file extension to the file path. Assumes no classes are included in the FQN.
//
// Example FQN:
//
//	"src.main.SomeFile.myFunc"
//
// Example result:
//
//	["src/main/SomeFile", "myFunc", ""]
func parseGenericFQN(fqn string) FQNParseResult {
	separatedFQN := strings.Split(fqn, ".")
	return FQNParseResult{
		FilePath: strings.Join(separatedFQN[:len(separatedFQN)-1], "/"),
		FuncName: separatedFQN[len(separatedFQN)-1]}
}

func startsWithUpper(s string) bool {
	for _, r := range s {
		return unicode.IsUpper(r)
	}
	return false
}

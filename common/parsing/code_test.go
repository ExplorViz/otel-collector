package parsing

import (
	"testing"

	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/stretchr/testify/assert"

	"github.com/ExplorViz/otel-collector/common/attrib"
)

func TestParseJavaCodeTelemetry(t *testing.T) {
	attrs := pcommon.NewMap()
	tr := attrib.TelemetryReader{Attrs: &attrs}
	var res codeParseResult
	var err error

	attrs.Clear()
	attrs.PutStr("code.function.name", "com.example.MyClass.myMethod")
	res, err = parseJavaCodeTelemetry(tr)
	assert.Nil(t, err)
	assert.Equal(t, codeParseResult{
		FilePath:  "com/example/MyClass.java",
		FuncName:  "myMethod",
		ClassName: "MyClass",
	}, res)

	attrs.Clear()
	attrs.PutStr("code.function.name", "com.example.MyClass.MyInnerClass.myMethod")
	res, err = parseJavaCodeTelemetry(tr)
	assert.Nil(t, err)
	assert.Equal(t, codeParseResult{
		FilePath:  "com/example/MyClass.java",
		FuncName:  "myMethod",
		ClassName: "MyClass.MyInnerClass",
	}, res)

	attrs.Clear()
	attrs.PutStr("code.function.name", "myMethod")
	_, err = parseJavaCodeTelemetry(tr)
	assert.NotNil(t, err)

	attrs.Clear()
	attrs.PutStr("code.function.name", "myMethod")
	attrs.PutStr("code.file.path", "/var/app/com/example/MyClass.java")
	res, err = parseJavaCodeTelemetry(tr)
	assert.Nil(t, err)
	assert.Equal(t, codeParseResult{
		FilePath:  "/var/app/com/example/MyClass.java",
		FuncName:  "myMethod",
		ClassName: "MyClass",
	}, res)

	attrs.Clear()
	attrs.PutStr("code.function.name", "MyClass.myMethod")
	attrs.PutStr("code.file.path", "/var/app/com/example/MyClass.java")
	res, err = parseJavaCodeTelemetry(tr)
	assert.Nil(t, err)
	assert.Equal(t, codeParseResult{
		FilePath:  "MyClass.java",
		FuncName:  "myMethod",
		ClassName: "MyClass",
	}, res)

	attrs.Clear()
	attrs.PutStr("code.function", "myMethod")
	attrs.PutStr("code.namespace", "com.example.MyClass")
	res, err = parseJavaCodeTelemetry(tr)
	assert.Nil(t, err)
	assert.Equal(t, codeParseResult{
		FilePath:  "com/example/MyClass.java",
		FuncName:  "myMethod",
		ClassName: "MyClass",
	}, res)

	attrs.Clear()
	attrs.PutStr("code.function", "myMethod")
	attrs.PutStr("code.namespace", "com.example.MyClass.MyInnerClass")
	res, err = parseJavaCodeTelemetry(tr)
	assert.Nil(t, err)
	assert.Equal(t, codeParseResult{
		FilePath:  "com/example/MyClass.java",
		FuncName:  "myMethod",
		ClassName: "MyClass.MyInnerClass",
	}, res)

	attrs.Clear()
	attrs.PutStr("code.function", "myMethod")
	attrs.PutStr("code.filepath", "/var/app/com/example/MyClass.java")
	res, err = parseJavaCodeTelemetry(tr)
	assert.Nil(t, err)
	assert.Equal(t, codeParseResult{
		FilePath:  "/var/app/com/example/MyClass.java",
		FuncName:  "myMethod",
		ClassName: "MyClass",
	}, res)
}

func TestParseGenericCodeTelemetry(t *testing.T) {
	attrs := pcommon.NewMap()
	tr := attrib.TelemetryReader{Attrs: &attrs}
	var res codeParseResult
	var err error

	attrs.Clear()
	attrs.PutStr("code.function.name", "some.namespace.somefile.myFunc")
	res, err = parseGenericCodeTelemetry(tr)
	assert.Nil(t, err)
	assert.Equal(t, codeParseResult{
		FilePath: "some/namespace/somefile",
		FuncName: "myFunc",
	}, res)

	attrs.Clear()
	attrs.PutStr("code.function.name", "some.namespace.somefile.myFunc")
	attrs.PutStr("code.file.path", "path/to/file")
	res, err = parseGenericCodeTelemetry(tr)
	assert.Nil(t, err)
	assert.Equal(t, codeParseResult{
		FilePath: "path/to/file",
		FuncName: "myFunc",
	}, res)

	attrs.Clear()
	attrs.PutStr("code.function.name", "myFunc")
	_, err = parseGenericCodeTelemetry(tr)
	assert.NotNil(t, err)

	attrs.Clear()
	attrs.PutStr("code.file.path", "path/to/file")
	_, err = parseGenericCodeTelemetry(tr)
	assert.NotNil(t, err)

	attrs.Clear()
	attrs.PutStr("code.function", "myFunc")
	attrs.PutStr("code.filepath", "path/to/file")
	res, err = parseGenericCodeTelemetry(tr)
	assert.Nil(t, err)
	assert.Equal(t, codeParseResult{
		FilePath: "path/to/file",
		FuncName: "myFunc",
	}, res)

	attrs.Clear()
	attrs.PutStr("code.function", "myFunc")
	_, err = parseGenericCodeTelemetry(tr)
	assert.NotNil(t, err)

	attrs.Clear()
	attrs.PutStr("code.filepath", "path/to/file")
	_, err = parseGenericCodeTelemetry(tr)
	assert.NotNil(t, err)
}

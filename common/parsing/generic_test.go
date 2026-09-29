package parsing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

func TestToAttributes(t *testing.T) {
	before := GenericEntity{
		ServiceName: "myService",
		ScopeName:   "prefix/myScope",
		Name:        "myScope",
	}
	m := pcommon.NewMap()
	before.ToAttributes(&m)
	entity, err := FromAttributes(m)
	assert.Nil(t, err)

	after, ok := entity.(GenericEntity)
	assert.True(t, ok)

	assert.Equal(t, before, after)
}

func TestDequalifyScopeName(t *testing.T) {
	assert.Equal(t, "scope", dequalifyScopeName("scope"))
	assert.Equal(t, "scope", dequalifyScopeName("prefix/scope"))
	assert.Equal(t, "scope", dequalifyScopeName("prefix::scope"))
	assert.Equal(t, "scope", dequalifyScopeName("prefix1/prefix2/scope"))
	assert.Equal(t, "scope", dequalifyScopeName("prefix1::prefix2::scope"))
	assert.Equal(t, "scope", dequalifyScopeName("prefix1/prefix2::scope"))
	assert.Equal(t, "scope", dequalifyScopeName("prefix1::prefix2/scope"))
	assert.Equal(t, "scope", dequalifyScopeName("prefix1::prefix2/scope"))
	assert.Equal(t, "scope/", dequalifyScopeName("scope/"))
	assert.Equal(t, "scope::", dequalifyScopeName("scope::"))
	assert.Equal(t, "/scope", dequalifyScopeName("/scope"))
	assert.Equal(t, "::scope", dequalifyScopeName("::scope"))
	assert.Equal(t, "/scope/", dequalifyScopeName("/scope/"))
	assert.Equal(t, "::scope::", dequalifyScopeName("::scope::"))
	assert.Equal(t, "//scope/", dequalifyScopeName("//scope/"))
	assert.Equal(t, "::scope::", dequalifyScopeName("::scope::"))
	assert.Equal(t, "scope", dequalifyScopeName("/prefix::scope"))
	assert.Equal(t, "scope/", dequalifyScopeName("/prefix::scope/"))
	assert.Equal(t, "///::::///", dequalifyScopeName("///::::///"))
	assert.Equal(t, "notprefix:scope", dequalifyScopeName("notprefix:scope"))
	assert.Equal(t, ":scope:", dequalifyScopeName("prefix/:scope:"))
}

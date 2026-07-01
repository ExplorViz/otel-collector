package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInMemStore(t *testing.T) {
	ts := NewInMemStore()
	assert.NotNil(t, ts.ptr.Load())
}

func TestInMemStorePut(t *testing.T) {
	ts := NewInMemStore()
	tokens := []LandscapeToken{
		{"mytokenvalue", "mytokensecret"},
		{},
		{ID: "myothertokenvalue"},
		{Secret: "It's a secret to everybody."},
	}

	for _, token := range tokens {
		assert.Error(t, ts.Validate(token))
		ts.Put(token)
		assert.NoError(t, ts.Validate(token))
	}

	ts.Put(LandscapeToken{"dreamscape", "token"})
	ts.Put(LandscapeToken{"ignore", "me"})
	assert.NoError(t, ts.Validate(LandscapeToken{"dreamscape", "token"}))
}

func TestInMemStoreDelete(t *testing.T) {
	ts := NewInMemStore()
	token := LandscapeToken{"mytokenvalue", "mytokensecret"}
	tokenTwo := LandscapeToken{"myothervalue", "myothersecret"}

	ts.Put(token)
	assert.NoError(t, ts.Validate(token))
	ts.Delete(token.ID)
	assert.Error(t, ts.Validate(token))

	ts.Put(tokenTwo)
	ts.Put(token)
	assert.NoError(t, ts.Validate(tokenTwo))
	assert.NoError(t, ts.Validate(token))

	assert.NotPanics(t, func() { ts.Delete("nonexistent") })
	assert.Error(t, ts.Validate(LandscapeToken{"nonexistent", ""}))
}

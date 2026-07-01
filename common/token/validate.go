package token

import (
	"fmt"
	"sync/atomic"
)

// A Validator validates a landscape token, identified by its ID and secret.
type Validator interface {
	// Validate checks whether the given landscape token meets the requirements of this validator
	// and returns an error if this is not the case.
	Validate(token LandscapeToken) error
}

// A NoOpValidator stores no data and treats all landscape token as valid.
type NoOpValidator struct{}

// Validate returns nil on all inputs.
func (s NoOpValidator) Validate(t LandscapeToken) error {
	return nil
}

var _ Validator = NoOpValidator{}

// An InMemStore uses a map to keep track of landscape tokens in-memory.
// Tokens can be manually added and removed from the store.
// The store is safe to read from multiple goroutines, however there must be
// only a single writer to prevent lost writes.
//
// Should not be initialized directly; use [NewInMemStore] instead.
type InMemStore struct {
	ptr atomic.Pointer[map[string]string]
}

// NewInMemStore initializes a new, empty [InMemStore] along with its internal map.
func NewInMemStore() *InMemStore {
	ts := InMemStore{}
	m := make(map[string]string)
	ts.ptr.Store(&m)
	return &ts
}

// Validate checks whether the provided token currently exists in the store and returns an error if it does not.
func (s *InMemStore) Validate(t LandscapeToken) error {
	m := s.ptr.Load()
	sec, ok := (*m)[t.ID]
	if ok && sec == t.Secret {
		return nil
	}
	return fmt.Errorf("unknown landscape token or incorrect secret")
}

// Put writes the provided landscape token to the store. Existing tokens are updated.
func (s *InMemStore) Put(t LandscapeToken) {
	old := s.ptr.Load()
	newMap := make(map[string]string, len(*old)+1)
	for k, v := range *old {
		newMap[k] = v
	}
	newMap[t.ID] = t.Secret
	s.ptr.Store(&newMap)
}

// Delete removes the provided landscape token from the store. No-op if specified token ID is not in the store.
func (s *InMemStore) Delete(tokenID string) {
	old := s.ptr.Load()
	newMap := make(map[string]string, len(*old))
	for k, v := range *old {
		newMap[k] = v
	}
	delete(newMap, tokenID)
	s.ptr.Store(&newMap)
}

var _ Validator = (*InMemStore)(nil)

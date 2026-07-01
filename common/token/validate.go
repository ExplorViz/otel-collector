package token

import (
	"fmt"
	"sync"
)

// A Validator validates a landscape token, identified by its ID and secret.
type Validator interface {
	// Validate checks whether the given landscape token meets the requirements of this validator
	// and returns an error if this is not the case.
	Validate(token LandscapeToken) error
}

// An InMemStore uses a map to keep track of landscape tokens in-memory.
// Tokens can be manually added and removed from the store.
// The store is safe to read and write from multiple goroutines.
// Should not be initialized directly; use [NewInMemStore] instead.
type InMemStore struct {
	mu sync.RWMutex
	m  map[string]string
}

// NewInMemStore initializes a new, empty [InMemStore] along with its internal map.
func NewInMemStore() *InMemStore {
	return &InMemStore{
		m: make(map[string]string),
	}
}

// Validate checks whether the provided token currently exists in the store and returns an error if it does not.
func (s *InMemStore) Validate(t LandscapeToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, ok := s.m[t.ID]
	if ok && val == t.Secret {
		return nil
	}
	return fmt.Errorf("unknown landscape token or incorrect secret")
}

// Put writes the provided landscape token to the store.
func (s *InMemStore) Put(t LandscapeToken) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.m[t.ID] = t.Secret
}

// Delete removes the provided landscape token from the store.
func (s *InMemStore) Delete(tokenID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.m, tokenID)
}

var _ Validator = (*InMemStore)(nil)

// A NoOpValidator stores no data and treats all landscape token as valid.
type NoOpValidator struct{}

// Validate returns nil on all inputs.
func (s NoOpValidator) Validate(t LandscapeToken) error {
	return nil
}

var _ Validator = NoOpValidator{}

package appattest

import (
	"context"
	"sync"
)

// MemKeys is an in-memory Keys for tests and local development.
type MemKeys struct {
	mu sync.Mutex
	m  map[string]Key
}

// NewMemKeys returns an empty store.
func NewMemKeys() *MemKeys { return &MemKeys{m: map[string]Key{}} }

// Insert implements Keys.
func (m *MemKeys) Insert(_ context.Context, k Key) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.m[string(k.ID)]; ok {
		return ErrKeyExists
	}
	m.m[string(k.ID)] = k
	return nil
}

// Get implements Keys.
func (m *MemKeys) Get(_ context.Context, id []byte) (Key, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	k, ok := m.m[string(id)]
	return k, ok, nil
}

// AdvanceCounter implements Keys; it succeeds only when counter exceeds the stored one.
func (m *MemKeys) AdvanceCounter(_ context.Context, id []byte, counter uint32) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	k, ok := m.m[string(id)]
	if !ok || counter <= k.Counter {
		return false, nil
	}
	k.Counter = counter
	m.m[string(id)] = k
	return true, nil
}

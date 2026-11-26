package switches

import (
	"context"
	"sync"
)

// Reader answers whether a switch is on for a show.
type Reader interface {
	Enabled(ctx context.Context, key Key, showID string) bool
}

// Memory is an in-process Reader holding the latest Snapshot.
type Memory struct {
	mu   sync.RWMutex
	snap Snapshot
}

// NewMemory returns a Reader seeded with snap.
func NewMemory(snap Snapshot) *Memory { return &Memory{snap: snap} }

// Store replaces the snapshot atomically.
func (m *Memory) Store(snap Snapshot) {
	m.mu.Lock()
	m.snap = snap
	m.mu.Unlock()
}

// Enabled implements Reader.
func (m *Memory) Enabled(_ context.Context, key Key, showID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.snap.Enabled(key, showID)
}

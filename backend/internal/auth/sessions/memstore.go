package sessions

import (
	"context"
	"sync"
	"time"
)

// MemStore is an in-memory Store for tests and local development.
type MemStore struct {
	mu   sync.Mutex
	byID map[string]*Session
}

// NewMemStore returns an empty store.
func NewMemStore() *MemStore { return &MemStore{byID: map[string]*Session{}} }

// Insert implements Store.
func (m *MemStore) Insert(_ context.Context, s Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.byID[s.ID] = &s
	return nil
}

// ByRefreshHash implements Store.
func (m *MemStore) ByRefreshHash(_ context.Context, hash [32]byte) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.byID {
		if s.RefreshHash == hash {
			return *s, nil
		}
	}
	return Session{}, ErrInvalid
}

// MarkRotated implements Store as compare-and-set.
func (m *MemStore) MarkRotated(_ context.Context, id string, at time.Time) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.byID[id]
	if !ok || s.RotatedAt != nil {
		return false, nil
	}
	s.RotatedAt, s.LastUsedAt = &at, at
	return true, nil
}

// RevokeFamily implements Store.
func (m *MemStore) RevokeFamily(_ context.Context, familyID, reason string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.byID {
		if s.FamilyID == familyID && s.RevokedAt == nil {
			s.RevokedAt, s.RevokeReason = &at, reason
		}
	}
	return nil
}

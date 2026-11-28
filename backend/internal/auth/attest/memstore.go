package attest

import (
	"context"
	"sync"
	"time"

	"trivia.al/backend/internal/auth/otp"
)

type memTicket struct {
	dev     otp.Device
	expires time.Time
}

// MemStore is an in-memory Store for tests and local development.
type MemStore struct {
	mu      sync.Mutex
	chal    map[string]time.Time
	tickets map[string]memTicket
	Now     func() time.Time
}

// NewMemStore returns an empty MemStore.
func NewMemStore() *MemStore {
	return &MemStore{chal: map[string]time.Time{}, tickets: map[string]memTicket{}, Now: time.Now}
}

// PutChallenge implements Store.
func (m *MemStore) PutChallenge(_ context.Context, c []byte, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.chal[string(c)] = m.Now().Add(ttl)
	return nil
}

// ConsumeChallenge implements Store.
func (m *MemStore) ConsumeChallenge(_ context.Context, c []byte) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	exp, ok := m.chal[string(c)]
	delete(m.chal, string(c))
	return ok && !m.Now().After(exp), nil
}

// PutTicket implements Store.
func (m *MemStore) PutTicket(_ context.Context, tok string, dev otp.Device, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tickets[tok] = memTicket{dev: dev, expires: m.Now().Add(ttl)}
	return nil
}

// Ticket implements Store.
func (m *MemStore) Ticket(_ context.Context, tok string) (otp.Device, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tickets[tok]
	if !ok || m.Now().After(t.expires) {
		return otp.Device{}, false, nil
	}
	return t.dev, true, nil
}

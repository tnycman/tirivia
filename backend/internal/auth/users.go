package auth

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

// MemUsers is an in-memory user directory for tests and local development. Production uses the
// users table (lookup by phone_hash, a keyed HMAC of the E.164 number).
type MemUsers struct {
	mu      sync.Mutex
	byPhone map[string]string
}

// NewMemUsers returns an empty directory.
func NewMemUsers() *MemUsers { return &MemUsers{byPhone: map[string]string{}} }

// Exists implements otp.Directory.
func (m *MemUsers) Exists(_ context.Context, phoneHash string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.byPhone[phoneHash]
	return ok, nil
}

// FindOrCreate returns the user for phoneHash, creating one on first sign-in.
func (m *MemUsers) FindOrCreate(_ context.Context, phoneHash string) (string, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if id, ok := m.byPhone[phoneHash]; ok {
		return id, false, nil
	}
	id, err := uuid.NewV7()
	if err != nil {
		return "", false, err
	}
	m.byPhone[phoneHash] = id.String()
	return id.String(), true, nil
}

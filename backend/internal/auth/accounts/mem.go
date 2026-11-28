package accounts

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"trivia.al/backend/internal/auth/otp"
)

type memDevice struct{ id, userID string }

// Mem is an in-memory user and device directory for tests and local development.
type Mem struct {
	mu      sync.Mutex
	byPhone map[string]string
	byKey   map[string]memDevice
}

// NewMem returns an empty directory.
func NewMem() *Mem { return &Mem{byPhone: map[string]string{}, byKey: map[string]memDevice{}} }

// Exists implements otp.Directory.
func (m *Mem) Exists(_ context.Context, phoneHash string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.byPhone[phoneHash]
	return ok, nil
}

// FindOrCreate returns the user for u.PhoneHash, creating one on first sign-in.
func (m *Mem) FindOrCreate(_ context.Context, u NewUser) (string, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if id, ok := m.byPhone[u.PhoneHash]; ok {
		return id, false, nil
	}
	id, err := uuid.NewV7()
	if err != nil {
		return "", false, err
	}
	m.byPhone[u.PhoneHash] = id.String()
	return id.String(), true, nil
}

// Register returns the device for dev.JKT, rejecting a key bound to another user.
func (m *Mem) Register(_ context.Context, userID string, dev otp.Device) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if d, ok := m.byKey[dev.JKT]; ok {
		if d.userID != userID {
			return "", ErrKeyBound
		}
		return d.id, nil
	}
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	m.byKey[dev.JKT] = memDevice{id: id.String(), userID: userID}
	return id.String(), nil
}

package otp

import (
	"context"
	"crypto/hmac"
	"sync"
	"time"
)

// MemStore is an in-memory Store for tests and local development.
type MemStore struct {
	mu sync.Mutex
	m  map[string]Challenge
}

// NewMemStore returns an empty store.
func NewMemStore() *MemStore { return &MemStore{m: map[string]Challenge{}} }

// Put implements Store.
func (s *MemStore) Put(_ context.Context, c Challenge, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[c.ID] = c
	return nil
}

// Get implements Store.
func (s *MemStore) Get(_ context.Context, id string) (Challenge, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.m[id]
	if !ok {
		return Challenge{}, ErrExpired
	}
	return c, nil
}

// Verify implements Store atomically.
func (s *MemStore) Verify(_ context.Context, id, expectedJKT string, codeMAC []byte, now time.Time, maxAttempts int) (Challenge, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.m[id]
	if !ok {
		return Challenge{}, ErrExpired
	}
	if now.After(c.ExpiresAt) {
		delete(s.m, id)
		return Challenge{}, ErrExpired
	}
	if c.Device.JKT != expectedJKT {
		return Challenge{}, ErrBinding
	}
	c.Attempts++
	if c.Attempts > maxAttempts {
		delete(s.m, id)
		return Challenge{}, ErrTooManyAttempts
	}
	if !hmac.Equal(codeMAC, c.CodeMAC) {
		if c.Attempts == maxAttempts {
			delete(s.m, id)
		} else {
			s.m[id] = c
		}
		return c, ErrWrongCode
	}
	delete(s.m, id)
	return c, nil
}

// Delete implements Store.
func (s *MemStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, id)
	return nil
}

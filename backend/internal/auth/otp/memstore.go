package otp

import (
	"context"
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

// IncrementAttempts implements Store.
func (s *MemStore) IncrementAttempts(_ context.Context, id string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.m[id]
	if !ok {
		return 0, ErrExpired
	}
	c.Attempts++
	s.m[id] = c
	return c.Attempts, nil
}

// Delete implements Store.
func (s *MemStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, id)
	return nil
}

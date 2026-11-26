package otp

import (
	"context"
	"sync"
	"time"
)

// MemStore is an in-memory Store (tests, local development). Production: Valkey with TTL.
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

// MemLimiter is a fixed-window in-memory Limiter (tests, local development).
type MemLimiter struct {
	mu sync.Mutex
	m  map[string]struct {
		n     int
		until time.Time
	}
	now func() time.Time
}

// NewMemLimiter returns an empty limiter.
func NewMemLimiter() *MemLimiter {
	return &MemLimiter{m: map[string]struct {
		n     int
		until time.Time
	}{}, now: time.Now}
}

// Allow implements Limiter.
func (l *MemLimiter) Allow(_ context.Context, key string, limit int, window time.Duration) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	e := l.m[key]
	if now.After(e.until) {
		e.n, e.until = 0, now.Add(window)
	}
	e.n++
	l.m[key] = e
	return e.n <= limit, nil
}

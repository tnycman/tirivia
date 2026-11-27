package otp

import (
	"context"
	"sync"
	"time"
)

type window struct {
	n     int
	until time.Time
}

// MemLimiter is an in-memory fixed-window Limiter for tests and local development.
type MemLimiter struct {
	mu  sync.Mutex
	m   map[string]window
	Now func() time.Time
}

// NewMemLimiter returns an empty limiter.
func NewMemLimiter() *MemLimiter { return &MemLimiter{m: map[string]window{}, Now: time.Now} }

// Allow implements Limiter.
func (l *MemLimiter) Allow(_ context.Context, key string, limit int, d time.Duration) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.Now()
	w := l.m[key]
	if now.After(w.until) {
		w = window{until: now.Add(d)}
	}
	w.n++
	l.m[key] = w
	return w.n <= limit, nil
}

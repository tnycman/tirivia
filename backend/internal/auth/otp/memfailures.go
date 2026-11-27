package otp

import (
	"context"
	"sync"
)

// MemFailures is an in-memory Failures for tests and local development.
type MemFailures struct {
	mu sync.Mutex
	m  map[string]int
}

// NewMemFailures returns an empty counter.
func NewMemFailures() *MemFailures { return &MemFailures{m: map[string]int{}} }

// Count implements Failures.
func (f *MemFailures) Count(_ context.Context, k string) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.m[k], nil
}

// Increment implements Failures.
func (f *MemFailures) Increment(_ context.Context, k string) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.m[k]++
	return f.m[k], nil
}

// Reset implements Failures.
func (f *MemFailures) Reset(_ context.Context, k string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.m, k)
	return nil
}

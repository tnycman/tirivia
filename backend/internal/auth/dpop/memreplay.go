package dpop

import (
	"context"
	"sync"
	"time"
)

// MemoryReplayCache is an in-process ReplayCache for tests and local development.
type MemoryReplayCache struct {
	mu   sync.Mutex
	seen map[string]time.Time
	now  func() time.Time
}

// NewMemoryReplayCache returns an empty cache.
func NewMemoryReplayCache() *MemoryReplayCache {
	return &MemoryReplayCache{seen: map[string]time.Time{}, now: time.Now}
}

// Seen implements ReplayCache.
func (m *MemoryReplayCache) Seen(_ context.Context, key string, ttl time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := m.now()
	if exp, ok := m.seen[key]; ok && now.Before(exp) {
		return true, nil
	}
	m.seen[key] = now.Add(ttl)
	return false, nil
}

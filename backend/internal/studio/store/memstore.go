// Package store persists studio state.
package store

import (
	"context"
	"log/slog"
	"sync"

	"trivia.al/backend/internal/studio/api"
)

// MemStore keeps switch rows in memory and logs an audit line for every change.
type MemStore struct {
	mu   sync.Mutex
	rows map[string]api.Row
	log  *slog.Logger
}

// NewMemStore returns an empty store (every switch at its default; chat OFF).
func NewMemStore(log *slog.Logger) *MemStore {
	return &MemStore{rows: map[string]api.Row{}, log: log}
}

// Rows implements api.Store.
func (m *MemStore) Rows(context.Context) ([]api.Row, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]api.Row, 0, len(m.rows))
	for _, r := range m.rows {
		out = append(out, r)
	}
	return out, nil
}

// Set implements api.Store.
func (m *MemStore) Set(_ context.Context, r api.Row, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rows[string(r.Key)+"|"+r.ShowID] = r
	m.log.Info("audit: switch changed", "key", r.Key, "show_id", r.ShowID, "enabled", r.Enabled,
		"actor", r.ChangedBy, "reason", reason)
	return nil
}

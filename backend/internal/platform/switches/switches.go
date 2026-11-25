// Package switches resolves backend feature switches (kill switches) without an app release.
//
// Source of truth: Postgres table feature_switches (audited; changed from the studio console).
// Hot path reads a Valkey cache with a short TTL. Resolution: a show-scoped value overrides
// the global value; if neither exists the key's safe default applies.
//
// The chat switch (ADR-0009) reaches apps via JoinShowResponse.chat_enabled at join and
// Pulse.chat_enabled live, so flipping it shows/hides chat on every phone within one pulse.
package switches

import (
	"context"
	"sync"
)

// Key names a switch. Keys are stable strings stored in the database.
type Key string

// Kill switches named in Master Plan §8 (plus per-country eligibility, held separately).
const (
	// ChatEnabled turns live chat on/off (ADR-0009). Default: off.
	ChatEnabled Key = "chat.enabled"
	// ReactionsEnabled turns floating reactions on/off. Default: on.
	ReactionsEnabled Key = "reactions.enabled"
	// ExtraLivesEnabled turns extra lives on/off. Default: on.
	ExtraLivesEnabled Key = "extra_lives.enabled"
	// SignupsEnabled pauses new sign-ups (OTP start) when off. Default: on.
	SignupsEnabled Key = "signups.enabled"
)

// defaults apply when no row exists. Chat defaults OFF: it must be switched on deliberately.
var defaults = map[Key]bool{
	ChatEnabled:       false,
	ReactionsEnabled:  true,
	ExtraLivesEnabled: true,
	SignupsEnabled:    true,
}

// Known reports whether key is a recognised switch.
func Known(key Key) bool {
	_, ok := defaults[key]
	return ok
}

// Default returns the built-in value used when no row exists.
func Default(key Key) bool { return defaults[key] }

// Keys returns all recognised switches in a stable order.
func Keys() []Key {
	return []Key{ChatEnabled, ReactionsEnabled, ExtraLivesEnabled, SignupsEnabled}
}

// Reader answers "is this switch on for this show?". showID may be empty for global checks.
type Reader interface {
	Enabled(ctx context.Context, key Key, showID string) bool
}

// Snapshot is an immutable view of all switch values, refreshed from the store.
type Snapshot struct {
	Global  map[Key]bool
	PerShow map[string]map[Key]bool
}

// Enabled resolves show scope, then global scope, then the safe default.
func (s Snapshot) Enabled(key Key, showID string) bool {
	if showID != "" {
		if v, ok := s.PerShow[showID][key]; ok {
			return v
		}
	}
	if v, ok := s.Global[key]; ok {
		return v
	}
	return defaults[key]
}

// Memory is an in-process Reader holding the latest Snapshot. A refresher (Valkey/Postgres)
// calls Store; tests construct it directly.
type Memory struct {
	mu   sync.RWMutex
	snap Snapshot
}

// NewMemory returns a Reader seeded with snap.
func NewMemory(snap Snapshot) *Memory { return &Memory{snap: snap} }

// Store replaces the snapshot atomically.
func (m *Memory) Store(snap Snapshot) {
	m.mu.Lock()
	m.snap = snap
	m.mu.Unlock()
}

// Enabled implements Reader.
func (m *Memory) Enabled(_ context.Context, key Key, showID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.snap.Enabled(key, showID)
}

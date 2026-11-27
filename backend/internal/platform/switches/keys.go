// Package switches resolves backend feature switches: show scope, then global, then default (ADR-0009).
package switches

// Key names a switch.
type Key string

// Kill switches (Master Plan §8).
const (
	ChatEnabled       Key = "chat.enabled"
	ReactionsEnabled  Key = "reactions.enabled"
	ExtraLivesEnabled Key = "extra_lives.enabled"
	SignupsEnabled    Key = "signups.enabled"
)

var defaults = map[Key]bool{
	ChatEnabled:       false,
	ReactionsEnabled:  true,
	ExtraLivesEnabled: true,
	SignupsEnabled:    true,
}

// Known reports whether key is recognised.
func Known(key Key) bool {
	_, ok := defaults[key]
	return ok
}

// Default returns the value used when no row exists.
func Default(key Key) bool { return defaults[key] }

// Keys returns all switches in a stable order.
func Keys() []Key { return []Key{ChatEnabled, ReactionsEnabled, ExtraLivesEnabled, SignupsEnabled} }

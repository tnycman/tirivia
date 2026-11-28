// Package api exposes studio back-office endpoints (internal admin listener only, ADR-0009).
package api

import (
	"context"
	"errors"
	"time"

	"trivia.al/backend/internal/platform/switches"
)

// MaxReasonBytes bounds the audit reason (policy; OWASP Logging: limit input size).
const MaxReasonBytes = 500

// ErrNoActor means no authenticated operator session.
var ErrNoActor = errors.New("studio: no operator session")

// Row is one stored switch value; ShowID is empty for global rows.
type Row struct {
	Key       switches.Key
	ShowID    string
	Enabled   bool
	ChangedBy string
	ChangedAt time.Time
}

// Store persists switches; Set must write the switch and its audit record atomically.
type Store interface {
	Rows(ctx context.Context) ([]Row, error)
	Set(ctx context.Context, row Row, reason string) error
}

// ActorFunc returns the authenticated operator id.
type ActorFunc func(ctx context.Context) (string, error)

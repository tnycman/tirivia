package otp

import (
	"context"
	"time"
)

// Device is what a verified attestation ticket proves.
type Device struct {
	ID       string
	Platform string
	JKT      string
}

// Challenge is a pending code; only its HMAC is stored.
type Challenge struct {
	ID        string
	PhoneE164 string
	PhoneHash string
	Device    Device
	CodeMAC   []byte
	ExpiresAt time.Time
	Attempts  int
}

// Tickets validates attestation tickets.
type Tickets interface {
	Validate(ctx context.Context, ticket string) (Device, error)
}

// Limiter is a fixed-window counter.
type Limiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

// NumberIntel returns a number's line type and pumping-risk score.
type NumberIntel interface {
	Lookup(ctx context.Context, e164 string) (lineType string, pumpingRisk int, err error)
}

// Sender delivers a code through one provider.
type Sender interface {
	Name() string
	Send(ctx context.Context, e164, code, locale string) error
}

// Signups reports the signups.enabled switch.
type Signups interface {
	Enabled(ctx context.Context) bool
}

// Directory reports whether a phone already has an account.
type Directory interface {
	Exists(ctx context.Context, phoneHash string) (bool, error)
}

// Failures counts consecutive failed verifications per phone; a new code never resets it.
type Failures interface {
	Count(ctx context.Context, phoneHash string) (int, error)
	Increment(ctx context.Context, phoneHash string) (int, error)
	Reset(ctx context.Context, phoneHash string) error
}

// Store keeps pending challenges.
type Store interface {
	Put(ctx context.Context, c Challenge, ttl time.Duration) error
	Get(ctx context.Context, id string) (Challenge, error)
	IncrementAttempts(ctx context.Context, id string) (int, error)
	Delete(ctx context.Context, id string) error
}

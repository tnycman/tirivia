// Package sessions rotates DPoP-bound refresh tokens with reuse detection (RFC 9700 §4.14).
package sessions

import (
	"context"
	"errors"
	"time"
)

// InactivityTTL ends unused sessions (decision ID-4).
const InactivityTTL = 30 * 24 * time.Hour

// Errors; all map to unauthenticated.
var (
	ErrInvalid = errors.New("sessions: unknown refresh token")
	ErrReuse   = errors.New("sessions: refresh token reused; family revoked")
	ErrRevoked = errors.New("sessions: session revoked")
	ErrExpired = errors.New("sessions: expired after inactivity")
	ErrBinding = errors.New("sessions: DPoP key mismatch; family revoked")
)

// Revoke reasons.
const (
	ReasonReuse      = "refresh_reuse_detected"
	ReasonBinding    = "dpop_key_mismatch"
	ReasonInactivity = "inactivity"
	ReasonSignOut    = "sign_out"
)

// Session is one link in a refresh-token family.
type Session struct {
	ID, UserID, DeviceID, FamilyID, JKT string
	RefreshHash                         [32]byte
	CreatedAt, LastUsedAt               time.Time
	RotatedAt, RevokedAt                *time.Time
	RevokeReason                        string
}

// Store persists sessions.
type Store interface {
	Insert(ctx context.Context, s Session) error
	ByRefreshHash(ctx context.Context, hash [32]byte) (Session, error)
	MarkRotated(ctx context.Context, id string, at time.Time) (bool, error)
	RevokeFamily(ctx context.Context, familyID, reason string, at time.Time) error
}

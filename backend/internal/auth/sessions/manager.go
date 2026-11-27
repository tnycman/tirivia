package sessions

import (
	"context"
	"crypto/subtle"
	"time"

	"github.com/google/uuid"
)

// Manager applies the rotation rules.
type Manager struct {
	store Store
	now   func() time.Time
}

// NewManager returns a Manager over store.
func NewManager(store Store) *Manager { return &Manager{store: store, now: time.Now} }

// Create starts a new token family.
func (m *Manager) Create(ctx context.Context, userID, deviceID, jkt string) (string, Session, error) {
	return m.issue(ctx, userID, deviceID, jkt, "")
}

// Rotate exchanges a refresh token for a new one; proofJKT comes from the verified DPoP proof.
func (m *Manager) Rotate(ctx context.Context, refreshToken, proofJKT string) (string, Session, error) {
	s, err := m.store.ByRefreshHash(ctx, hashToken(refreshToken))
	if err != nil {
		return "", Session{}, ErrInvalid
	}
	if err := m.check(ctx, s, proofJKT); err != nil {
		return "", Session{}, err
	}
	won, err := m.store.MarkRotated(ctx, s.ID, m.now())
	if err != nil {
		return "", Session{}, err
	}
	if !won {
		_ = m.store.RevokeFamily(ctx, s.FamilyID, ReasonReuse, m.now())
		return "", Session{}, ErrReuse
	}
	return m.issue(ctx, s.UserID, s.DeviceID, s.JKT, s.FamilyID)
}

// Revoke ends the family of refreshToken; unknown tokens are ignored.
func (m *Manager) Revoke(ctx context.Context, refreshToken string) error {
	s, err := m.store.ByRefreshHash(ctx, hashToken(refreshToken))
	if err != nil {
		return nil //nolint:nilerr // unknown token: nothing to revoke
	}
	return m.store.RevokeFamily(ctx, s.FamilyID, ReasonSignOut, m.now())
}

func (m *Manager) check(ctx context.Context, s Session, proofJKT string) error {
	now := m.now()
	revoke := func(reason string, err error) error {
		_ = m.store.RevokeFamily(ctx, s.FamilyID, reason, now)
		return err
	}
	switch {
	case s.RevokedAt != nil:
		return ErrRevoked
	case s.RotatedAt != nil:
		return revoke(ReasonReuse, ErrReuse)
	case subtle.ConstantTimeCompare([]byte(s.JKT), []byte(proofJKT)) != 1:
		return revoke(ReasonBinding, ErrBinding)
	case now.Sub(s.LastUsedAt) > InactivityTTL:
		return revoke(ReasonInactivity, ErrExpired)
	}
	return nil
}

func (m *Manager) issue(ctx context.Context, userID, deviceID, jkt, familyID string) (string, Session, error) {
	tok, hash, err := newToken()
	if err != nil {
		return "", Session{}, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return "", Session{}, err
	}
	if familyID == "" {
		familyID = id.String()
	}
	now := m.now()
	s := Session{ID: id.String(), UserID: userID, DeviceID: deviceID, FamilyID: familyID, JKT: jkt,
		RefreshHash: hash, CreatedAt: now, LastUsedAt: now}
	if err := m.store.Insert(ctx, s); err != nil {
		return "", Session{}, err
	}
	return tok, s, nil
}

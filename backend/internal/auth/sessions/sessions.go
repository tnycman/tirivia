// Package sessions manages refresh tokens (ADR-0005, RFC 9700 §4.14 "Refresh Token Protection"):
//   - opaque, 256-bit random tokens; only a SHA-256 hash is stored
//   - rotated on every use; presenting an already-rotated token is treated as theft and revokes
//     the whole token family (reuse detection)
//   - sender-constrained: every refresh must carry a DPoP proof from the key bound at sign-in
//   - expire after 30 days of inactivity (decision ID-4)
package sessions

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// InactivityTTL ends sessions unused for this long (decision ID-4).
const InactivityTTL = 30 * 24 * time.Hour

// Errors. All map to CodeUnauthenticated; the app signs in again with an OTP.
var (
	ErrInvalid  = errors.New("sessions: unknown refresh token")
	ErrReuse    = errors.New("sessions: refresh token reused; family revoked")
	ErrRevoked  = errors.New("sessions: session revoked")
	ErrExpired  = errors.New("sessions: session expired after inactivity")
	ErrBinding  = errors.New("sessions: DPoP key does not match the session; family revoked")
	ErrConflict = errors.New("sessions: concurrent rotation")
)

// Revoke reasons stored for the audit trail.
const (
	ReasonReuse      = "refresh_reuse_detected"
	ReasonBinding    = "dpop_key_mismatch"
	ReasonInactivity = "inactivity"
	ReasonSignOut    = "sign_out"
)

// Session is one link in a refresh-token family.
type Session struct {
	ID, UserID, DeviceID, FamilyID string
	JKT                            string // DPoP key thumbprint bound at sign-in
	RefreshHash                    [32]byte
	CreatedAt, LastUsedAt          time.Time
	RotatedAt, RevokedAt           *time.Time
	RevokeReason                   string
}

// Store persists sessions (Postgres `sessions` table in production).
type Store interface {
	Insert(ctx context.Context, s Session) error
	ByRefreshHash(ctx context.Context, hash [32]byte) (Session, error) // ErrInvalid if absent
	// MarkRotated sets rotated_at only if it is still NULL (compare-and-set); false if lost race.
	MarkRotated(ctx context.Context, id string, at time.Time) (bool, error)
	RevokeFamily(ctx context.Context, familyID, reason string, at time.Time) error
}

// Manager implements the rotation rules.
type Manager struct {
	store Store
	now   func() time.Time
}

// NewManager returns a Manager over store.
func NewManager(store Store) *Manager { return &Manager{store: store, now: time.Now} }

func newToken() (string, [32]byte, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", [32]byte{}, err
	}
	tok := base64.RawURLEncoding.EncodeToString(raw[:])
	return tok, sha256.Sum256([]byte(tok)), nil
}

// Create starts a new family after a successful OTP sign-in.
func (m *Manager) Create(ctx context.Context, userID, deviceID, jkt string) (string, Session, error) {
	return m.issue(ctx, userID, deviceID, jkt, "")
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

// Rotate exchanges a refresh token for a new one. proofJKT is the thumbprint of the verified
// DPoP proof sent with the request.
func (m *Manager) Rotate(ctx context.Context, refreshToken, proofJKT string) (string, Session, error) {
	hash := sha256.Sum256([]byte(refreshToken))
	s, err := m.store.ByRefreshHash(ctx, hash)
	if err != nil {
		return "", Session{}, ErrInvalid
	}
	now := m.now()
	switch {
	case s.RevokedAt != nil:
		return "", Session{}, ErrRevoked
	case s.RotatedAt != nil:
		// RFC 9700 §4.14.2: a rotated token came back -> one copy was stolen. Kill the family.
		_ = m.store.RevokeFamily(ctx, s.FamilyID, ReasonReuse, now)
		return "", Session{}, ErrReuse
	case subtle.ConstantTimeCompare([]byte(s.JKT), []byte(proofJKT)) != 1:
		_ = m.store.RevokeFamily(ctx, s.FamilyID, ReasonBinding, now)
		return "", Session{}, ErrBinding
	case now.Sub(s.LastUsedAt) > InactivityTTL:
		_ = m.store.RevokeFamily(ctx, s.FamilyID, ReasonInactivity, now)
		return "", Session{}, ErrExpired
	}
	won, err := m.store.MarkRotated(ctx, s.ID, now)
	if err != nil {
		return "", Session{}, err
	}
	if !won {
		// Another request rotated this token first: indistinguishable from reuse.
		_ = m.store.RevokeFamily(ctx, s.FamilyID, ReasonReuse, now)
		return "", Session{}, ErrReuse
	}
	return m.issue(ctx, s.UserID, s.DeviceID, s.JKT, s.FamilyID)
}

// Revoke ends the family of refreshToken (sign-out). Unknown tokens are ignored.
func (m *Manager) Revoke(ctx context.Context, refreshToken string) error {
	s, err := m.store.ByRefreshHash(ctx, sha256.Sum256([]byte(refreshToken)))
	if err != nil {
		return nil //nolint:nilerr // sign-out of an unknown token is a no-op
	}
	return m.store.RevokeFamily(ctx, s.FamilyID, ReasonSignOut, m.now())
}

// MemStore is an in-memory Store for tests and local development.
type MemStore struct {
	mu   sync.Mutex
	byID map[string]*Session
}

// NewMemStore returns an empty store.
func NewMemStore() *MemStore { return &MemStore{byID: map[string]*Session{}} }

// Insert implements Store.
func (m *MemStore) Insert(_ context.Context, s Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := s
	m.byID[s.ID] = &cp
	return nil
}

// ByRefreshHash implements Store.
func (m *MemStore) ByRefreshHash(_ context.Context, hash [32]byte) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.byID {
		if s.RefreshHash == hash {
			return *s, nil
		}
	}
	return Session{}, ErrInvalid
}

// MarkRotated implements Store.
func (m *MemStore) MarkRotated(_ context.Context, id string, at time.Time) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.byID[id]
	if !ok || s.RotatedAt != nil {
		return false, nil
	}
	s.RotatedAt, s.LastUsedAt = &at, at
	return true, nil
}

// RevokeFamily implements Store.
func (m *MemStore) RevokeFamily(_ context.Context, familyID, reason string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.byID {
		if s.FamilyID == familyID && s.RevokedAt == nil {
			s.RevokedAt, s.RevokeReason = &at, reason
		}
	}
	return nil
}

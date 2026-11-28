package pgstore

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"trivia.al/backend/internal/auth/sessions"
	"trivia.al/backend/internal/auth/store/pgstore/pgq"
)

// Sessions stores refresh-token families; the DPoP key comes from the device row.
type Sessions struct{ q *pgq.Queries }

// NewSessions returns a Sessions store.
func NewSessions(db pgq.DBTX) *Sessions { return &Sessions{q: pgq.New(db)} }

// Insert implements sessions.Store.
func (s *Sessions) Insert(ctx context.Context, x sessions.Session) error {
	ids, err := parseIDs(x.ID, x.UserID, x.DeviceID, x.FamilyID)
	if err != nil {
		return err
	}
	return s.q.InsertSession(ctx, pgq.InsertSessionParams{ID: ids[0], UserID: ids[1], DeviceID: ids[2],
		RefreshFamilyID: ids[3], RefreshHash: x.RefreshHash[:], CreatedAt: x.CreatedAt, LastUsedAt: x.LastUsedAt})
}

// ByRefreshHash implements sessions.Store.
func (s *Sessions) ByRefreshHash(ctx context.Context, hash [32]byte) (sessions.Session, error) {
	r, err := s.q.SessionByRefreshHash(ctx, hash[:])
	if errors.Is(err, pgx.ErrNoRows) {
		return sessions.Session{}, sessions.ErrInvalid
	}
	if err != nil {
		return sessions.Session{}, err
	}
	x := sessions.Session{ID: r.ID.String(), UserID: r.UserID.String(), DeviceID: r.DeviceID.String(),
		FamilyID: r.RefreshFamilyID.String(), JKT: r.DpopJkt, RefreshHash: hash,
		CreatedAt: r.CreatedAt, LastUsedAt: r.LastUsedAt, RotatedAt: r.RotatedAt, RevokedAt: r.RevokedAt}
	if r.RevokeReason != nil {
		x.RevokeReason = *r.RevokeReason
	}
	return x, nil
}

// MarkRotated sets rotated_at once; only the first caller wins.
func (s *Sessions) MarkRotated(ctx context.Context, id string, at time.Time) (bool, error) {
	sid, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}
	n, err := s.q.MarkSessionRotated(ctx, pgq.MarkSessionRotatedParams{ID: sid, RotatedAt: &at})
	return n == 1, err
}

// RevokeFamily implements sessions.Store.
func (s *Sessions) RevokeFamily(ctx context.Context, familyID, reason string, at time.Time) error {
	fid, err := uuid.Parse(familyID)
	if err != nil {
		return err
	}
	if err := s.q.RecordSessionFamilyRevocation(ctx, pgq.RecordSessionFamilyRevocationParams{
		RefreshFamilyID: fid, RevokeReason: reason, RevokedAt: at,
	}); err != nil {
		return err
	}
	return s.q.ApplySessionFamilyRevocation(ctx, fid)
}

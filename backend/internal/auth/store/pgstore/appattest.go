package pgstore

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"trivia.al/backend/internal/auth/attest/appattest"
	"trivia.al/backend/internal/auth/store/pgstore/pgq"
)

// AppAttestKeys stores attested keys; the counter only moves forward, atomically.
type AppAttestKeys struct{ q *pgq.Queries }

// NewAppAttestKeys returns an AppAttestKeys store.
func NewAppAttestKeys(db pgq.DBTX) *AppAttestKeys { return &AppAttestKeys{q: pgq.New(db)} }

// Insert implements appattest.Keys.
func (s *AppAttestKeys) Insert(ctx context.Context, k appattest.Key) error {
	n, err := s.q.InsertAppAttestKey(ctx, pgq.InsertAppAttestKeyParams{KeyID: k.ID, PublicKey: k.PublicKey,
		PrizeEligible: k.PrizeEligible, Receipt: k.Receipt})
	if err != nil {
		return err
	}
	if n == 0 {
		return appattest.ErrKeyExists
	}
	return nil
}

// Get implements appattest.Keys.
func (s *AppAttestKeys) Get(ctx context.Context, id []byte) (appattest.Key, bool, error) {
	r, err := s.q.AppAttestKey(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return appattest.Key{}, false, nil
	}
	if err != nil {
		return appattest.Key{}, false, err
	}
	return appattest.Key{ID: id, PublicKey: r.PublicKey, Counter: uint32(r.Counter), //nolint:gosec // G115: CHECK keeps counter within uint32
		PrizeEligible: r.PrizeEligible, Receipt: r.Receipt}, true, nil
}

// AdvanceCounter implements appattest.Keys.
func (s *AppAttestKeys) AdvanceCounter(ctx context.Context, id []byte, counter uint32) (bool, error) {
	n, err := s.q.AdvanceAppAttestCounter(ctx, pgq.AdvanceAppAttestCounterParams{KeyID: id, Counter: int64(counter)})
	return n == 1, err
}

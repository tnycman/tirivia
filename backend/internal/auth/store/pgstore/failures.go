package pgstore

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/tirivia/backend/internal/auth/store/pgstore/pgq"
)

// Failures is the durable consecutive-failure counter (NIST SP 800-63B-4 §3.2.2).
type Failures struct{ q *pgq.Queries }

// NewFailures returns a Failures store.
func NewFailures(db pgq.DBTX) *Failures { return &Failures{q: pgq.New(db)} }

// Count implements otp.Failures.
func (f *Failures) Count(ctx context.Context, phoneHash string) (int, error) {
	h, err := phoneKey(phoneHash)
	if err != nil {
		return 0, err
	}
	n, err := f.q.FailureCount(ctx, h)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return int(n), err
}

// Increment implements otp.Failures.
func (f *Failures) Increment(ctx context.Context, phoneHash string) (int, error) {
	h, err := phoneKey(phoneHash)
	if err != nil {
		return 0, err
	}
	n, err := f.q.IncrementFailures(ctx, h)
	return int(n), err
}

// Reset implements otp.Failures.
func (f *Failures) Reset(ctx context.Context, phoneHash string) error {
	h, err := phoneKey(phoneHash)
	if err != nil {
		return err
	}
	return f.q.ResetFailures(ctx, h)
}

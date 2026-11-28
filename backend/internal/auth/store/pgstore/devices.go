package pgstore

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"trivia.al/backend/internal/auth/accounts"
	"trivia.al/backend/internal/auth/otp"
	"trivia.al/backend/internal/auth/store/pgstore/pgq"
)

// Devices binds each DPoP key to exactly one user.
type Devices struct{ q *pgq.Queries }

// NewDevices returns a Devices store.
func NewDevices(db pgq.DBTX) *Devices { return &Devices{q: pgq.New(db)} }

// Register returns the device for dev.JKT; a key owned by another user yields accounts.ErrKeyBound.
func (d *Devices) Register(ctx context.Context, userID string, dev otp.Device) (string, error) {
	user, err := uuid.Parse(userID)
	if err != nil {
		return "", err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	got, err := d.q.RegisterDevice(ctx, pgq.RegisterDeviceParams{ID: id, UserID: user,
		Platform: dev.Platform, AttestKeyID: &dev.ID, DpopJkt: dev.JKT, PrizeEligible: dev.PrizeEligible})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", accounts.ErrKeyBound
	}
	return got.String(), err
}

package pgstore

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"trivia.al/backend/internal/auth/accounts"
	"trivia.al/backend/internal/auth/store/pgstore/pgq"
	"trivia.al/backend/internal/platform/seal"
)

// Users is the accounts directory; phones are stored sealed, bound to their hash.
type Users struct {
	q    *pgq.Queries
	seal *seal.Sealer
}

// NewUsers returns a Users store.
func NewUsers(db pgq.DBTX, s *seal.Sealer) *Users { return &Users{q: pgq.New(db), seal: s} }

// Exists implements otp.Directory.
func (u *Users) Exists(ctx context.Context, phoneHash string) (bool, error) {
	h, err := phoneKey(phoneHash)
	if err != nil {
		return false, err
	}
	return u.q.UserExists(ctx, h)
}

// FindOrCreate inserts the user unless the phone hash exists; concurrent first sign-ins converge.
func (u *Users) FindOrCreate(ctx context.Context, nu accounts.NewUser) (string, bool, error) {
	h, err := phoneKey(nu.PhoneHash)
	if err != nil {
		return "", false, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return "", false, err
	}
	got, err := u.q.InsertUser(ctx, pgq.InsertUserParams{ID: id, PhoneHash: h,
		PhoneEnc: u.seal.Seal([]byte(nu.PhoneE164), h), CountryCode: nu.Country})
	if err == nil {
		return got.String(), true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", false, err
	}
	got, err = u.q.UserByPhone(ctx, h)
	return got.String(), false, err
}

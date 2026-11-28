package pgstore_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"trivia.al/backend/internal/auth/accounts"
	"trivia.al/backend/internal/auth/accounts/accountstest"
	"trivia.al/backend/internal/auth/attest/appattest/keystest"
	"trivia.al/backend/internal/auth/otp"
	"trivia.al/backend/internal/auth/otp/otptest"
	"trivia.al/backend/internal/auth/sessions/sessionstest"
	"trivia.al/backend/internal/auth/store/pgstore"
	"trivia.al/backend/internal/platform/pgdb"
	"trivia.al/backend/internal/platform/seal"
)

func pool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set (make store-test)")
	}
	db, err := pgdb.Open(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	return db
}

func random(t *testing.T, n int) []byte {
	t.Helper()
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return b
}

func sealer(t *testing.T) *seal.Sealer {
	t.Helper()
	s, err := seal.New(1, map[byte][]byte{1: random(t, seal.KeyBytes)})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

type directory struct {
	*pgstore.Users
	*pgstore.Devices
}

func TestDirectory(t *testing.T) {
	db := pool(t)
	d := directory{pgstore.NewUsers(db, sealer(t)), pgstore.NewDevices(db)}
	accountstest.Run(t, d)
	accountstest.RunRace(t, d)
}

func TestFailures(t *testing.T) { otptest.RunFailures(t, pgstore.NewFailures(pool(t))) }

func TestSessions(t *testing.T) {
	db := pool(t)
	users, devices := pgstore.NewUsers(db, sealer(t)), pgstore.NewDevices(db)
	seed := func(t *testing.T) (string, string, string) {
		ctx := context.Background()
		h, jkt := hex.EncodeToString(random(t, 32)), hex.EncodeToString(random(t, 32))
		u, _, err := users.FindOrCreate(ctx, accounts.NewUser{PhoneHash: h, PhoneE164: "+355691234567", Country: "AL"})
		if err != nil {
			t.Fatal(err)
		}
		d, err := devices.Register(ctx, u, otp.Device{ID: "ios:k", Platform: "ios", JKT: jkt})
		if err != nil {
			t.Fatal(err)
		}
		return u, d, jkt
	}
	sessionstest.Run(t, pgstore.NewSessions(db), seed)
	sessionstest.RunRace(t, pgstore.NewSessions(db), seed)
	sessionstest.RunFamilyRevocationRace(t, pgstore.NewSessions(db), seed)
}

func TestAppAttestKeys(t *testing.T) { keystest.Run(t, pgstore.NewAppAttestKeys(pool(t))) }

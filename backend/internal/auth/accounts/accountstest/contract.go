// Package accountstest is the behaviour every accounts directory must share.
package accountstest

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/tirivia/backend/internal/auth/accounts"
	"github.com/tirivia/backend/internal/auth/otp"
)

// Directory is the full surface the stores must provide.
type Directory interface {
	Exists(ctx context.Context, phoneHash string) (bool, error)
	FindOrCreate(ctx context.Context, u accounts.NewUser) (string, bool, error)
	Register(ctx context.Context, userID string, dev otp.Device) (string, error)
}

// Run checks d; phone hashes and keys are random so stores can be shared.
func Run(t *testing.T, d Directory) {
	t.Helper()
	ctx := context.Background()
	u := accounts.NewUser{PhoneHash: rnd(t), PhoneE164: "+355691234567", Country: "AL"}
	if ok, err := d.Exists(ctx, u.PhoneHash); err != nil || ok {
		t.Fatalf("exists before create: %v %v", ok, err)
	}
	id, isNew, err := d.FindOrCreate(ctx, u)
	if err != nil || !isNew {
		t.Fatalf("first sign-in: %v %v", isNew, err)
	}
	again, isNew, err := d.FindOrCreate(ctx, u)
	if err != nil || isNew || again != id {
		t.Fatalf("second sign-in must return the same user: %v %v %v", again, isNew, err)
	}
	other, _, err := d.FindOrCreate(ctx, accounts.NewUser{PhoneHash: rnd(t), PhoneE164: "+38344123456", Country: "XK"})
	if err != nil {
		t.Fatal(err)
	}
	dev := otp.Device{ID: "ios:key-1", Platform: "ios", JKT: rnd(t)}
	d1, err := d.Register(ctx, id, dev)
	if err != nil {
		t.Fatal(err)
	}
	if d2, err := d.Register(ctx, id, dev); err != nil || d2 != d1 {
		t.Fatalf("re-register must be idempotent: %v %v", d2, err)
	}
	if _, err := d.Register(ctx, other, dev); !errors.Is(err, accounts.ErrKeyBound) {
		t.Fatalf("key moved to another account: %v", err)
	}
}

func rnd(t *testing.T) string {
	t.Helper()
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(b)
}

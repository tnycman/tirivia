package otptest

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/tirivia/backend/internal/auth/otp"
)

// RunStore checks s: fields round-trip, attempts count up, deleted challenges are gone.
func RunStore(t *testing.T, s otp.Store) {
	t.Helper()
	ctx := context.Background()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	c := otp.Challenge{ID: id.String(), PhoneE164: "+355691234567", PhoneHash: hash(t), Region: "AL",
		Device: otp.Device{ID: "ios:k", Platform: "ios", JKT: "jkt"}, CodeMAC: []byte{1, 2, 3},
		ExpiresAt: time.Now().Add(otp.CodeTTL).UTC().Truncate(time.Second)}
	if err := s.Put(ctx, c, otp.CodeTTL); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(ctx, c.ID)
	if err != nil || got.PhoneE164 != c.PhoneE164 || got.PhoneHash != c.PhoneHash || got.Region != c.Region ||
		got.Device != c.Device || !bytes.Equal(got.CodeMAC, c.CodeMAC) || !got.ExpiresAt.Equal(c.ExpiresAt) {
		t.Fatalf("round trip: %+v %v", got, err)
	}
	for want := 1; want <= 2; want++ {
		if n, err := s.IncrementAttempts(ctx, c.ID); err != nil || n != want {
			t.Fatalf("attempt %d: %d %v", want, n, err)
		}
	}
	if err := s.Delete(ctx, c.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(ctx, c.ID); err == nil {
		t.Fatal("deleted challenge still readable")
	}
	if _, err := s.IncrementAttempts(ctx, c.ID); err == nil {
		t.Fatal("attempts counted on a deleted challenge")
	}
}

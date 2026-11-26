// Package attesttest is the behaviour every attest.Store must share.
package attesttest

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"

	"github.com/tirivia/backend/internal/auth/attest"
	"github.com/tirivia/backend/internal/auth/otp"
)

// RunStore checks s: challenges are single use, tickets are readable until they expire.
func RunStore(t *testing.T, s attest.Store) {
	t.Helper()
	ctx := context.Background()
	c := make([]byte, attest.ChallengeBytes)
	if _, err := rand.Read(c); err != nil {
		t.Fatal(err)
	}
	if err := s.PutChallenge(ctx, c, time.Minute); err != nil {
		t.Fatal(err)
	}
	for i, want := range []bool{true, false} {
		if ok, err := s.ConsumeChallenge(ctx, c); err != nil || ok != want {
			t.Fatalf("consume %d: %v %v", i+1, ok, err)
		}
	}
	tok := hex.EncodeToString(c)
	dev := otp.Device{ID: "android:x", Platform: "android", JKT: "jkt"}
	if err := s.PutTicket(ctx, tok, dev, 300*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if got, ok, err := s.Ticket(ctx, tok); err != nil || !ok || got != dev {
		t.Fatalf("ticket: %+v %v %v", got, ok, err)
	}
	if _, ok, err := s.Ticket(ctx, tok+"x"); err != nil || ok {
		t.Fatalf("unknown ticket: %v %v", ok, err)
	}
	time.Sleep(500 * time.Millisecond)
	if _, ok, err := s.Ticket(ctx, tok); err != nil || ok {
		t.Fatalf("expired ticket readable: %v %v", ok, err)
	}
}

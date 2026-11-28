package otptest

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"trivia.al/backend/internal/auth/otp"
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
	if _, err := s.Verify(ctx, c.ID, "other", c.CodeMAC, time.Now(), otp.MaxAttempts); !errors.Is(err, otp.ErrBinding) {
		t.Fatalf("binding mismatch: %v", err)
	}
	if got, err := s.Get(ctx, c.ID); err != nil || got.Attempts != 0 {
		t.Fatalf("binding mismatch altered challenge: %+v %v", got, err)
	}
	if got, err := s.Verify(ctx, c.ID, c.Device.JKT, []byte("wrong"), time.Now(), otp.MaxAttempts); !errors.Is(err, otp.ErrWrongCode) || got.Attempts != 1 {
		t.Fatalf("wrong code: %+v %v", got, err)
	}
	runConsumeRace(t, s, c)
	if _, err := s.Get(ctx, c.ID); err == nil {
		t.Fatal("consumed challenge still readable")
	}
}

func runConsumeRace(t *testing.T, s otp.Store, c otp.Challenge) {
	t.Helper()
	const racers = 16
	ready := make(chan struct{})
	var wg sync.WaitGroup
	var wins atomic.Int32
	for range racers {
		wg.Go(func() {
			<-ready
			if _, err := s.Verify(context.Background(), c.ID, c.Device.JKT, c.CodeMAC, time.Now(), otp.MaxAttempts); err == nil {
				wins.Add(1)
			}
		})
	}
	close(ready)
	wg.Wait()
	if wins.Load() != 1 {
		t.Fatalf("%d correct verifications succeeded; want 1", wins.Load())
	}
}

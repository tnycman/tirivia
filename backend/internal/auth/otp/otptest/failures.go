// Package otptest is the behaviour every otp store must share.
package otptest

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/tirivia/backend/internal/auth/otp"
)

// RunFailures checks f: counts accumulate per phone and only Reset clears them.
func RunFailures(t *testing.T, f otp.Failures) {
	t.Helper()
	ctx := context.Background()
	phone, other := hash(t), hash(t)
	if n, err := f.Count(ctx, phone); err != nil || n != 0 {
		t.Fatalf("initial count: %d %v", n, err)
	}
	for want := 1; want <= 3; want++ {
		if n, err := f.Increment(ctx, phone); err != nil || n != want {
			t.Fatalf("increment %d: %d %v", want, n, err)
		}
	}
	if n, _ := f.Count(ctx, other); n != 0 {
		t.Fatalf("count leaked across phones: %d", n)
	}
	if err := f.Reset(ctx, phone); err != nil {
		t.Fatal(err)
	}
	if n, err := f.Count(ctx, phone); err != nil || n != 0 {
		t.Fatalf("after reset: %d %v", n, err)
	}
	if err := f.Reset(ctx, other); err != nil {
		t.Fatalf("reset of unknown phone: %v", err)
	}
}

func hash(t *testing.T) string {
	t.Helper()
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(b)
}

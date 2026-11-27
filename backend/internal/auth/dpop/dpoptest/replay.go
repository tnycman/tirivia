// Package dpoptest is the behaviour every dpop.ReplayCache must share.
package dpoptest

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"

	"github.com/tirivia/backend/internal/auth/dpop"
)

// RunReplay checks c: the first sighting is new, the second is a replay, keys are independent.
func RunReplay(t *testing.T, c dpop.ReplayCache) {
	t.Helper()
	ctx := context.Background()
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	key := hex.EncodeToString(b)
	for i, want := range []bool{false, true} {
		if seen, err := c.Seen(ctx, key, time.Minute); err != nil || seen != want {
			t.Fatalf("sighting %d: seen=%v err=%v", i+1, seen, err)
		}
	}
	if seen, err := c.Seen(ctx, key+"-other", time.Minute); err != nil || seen {
		t.Fatalf("independent key: %v %v", seen, err)
	}
}

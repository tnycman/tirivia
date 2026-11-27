package sessionstest

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tirivia/backend/internal/auth/sessions"
)

// Racers is how many concurrent rotations RunRace attempts.
const Racers = 16

// RunRace checks that exactly one of many concurrent rotations of one token wins (RFC 9700 §4.14).
func RunRace(t *testing.T, s sessions.Store, seed Seed) {
	t.Helper()
	user, device, jkt := seed(t)
	x := session(t, user, device, jkt, "", time.Now().UTC())
	if err := s.Insert(context.Background(), x); err != nil {
		t.Fatal(err)
	}
	var wins atomic.Int32
	var wg sync.WaitGroup
	for range Racers {
		wg.Go(func() {
			if won, err := s.MarkRotated(context.Background(), x.ID, time.Now()); err == nil && won {
				wins.Add(1)
			}
		})
	}
	wg.Wait()
	if wins.Load() != 1 {
		t.Fatalf("%d concurrent rotations won; want exactly 1", wins.Load())
	}
}

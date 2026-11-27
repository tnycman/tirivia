package sessionstest

import (
	"context"
	"errors"
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

type blockingStore struct {
	sessions.Store
	started chan struct{}
	allow   chan struct{}
}

func (s *blockingStore) Insert(ctx context.Context, x sessions.Session) error {
	if x.ID != x.FamilyID {
		close(s.started)
		<-s.allow
	}
	return s.Store.Insert(ctx, x)
}

// RunFamilyRevocationRace forces reuse revocation before the winning child insert.
func RunFamilyRevocationRace(t *testing.T, store sessions.Store, seed Seed) {
	t.Helper()
	user, device, jkt := seed(t)
	b := &blockingStore{Store: store, started: make(chan struct{}), allow: make(chan struct{})}
	m := sessions.NewManager(b)
	token, _, err := m.Create(context.Background(), user, device, jkt)
	if err != nil {
		t.Fatal(err)
	}
	type result struct {
		token string
		err   error
	}
	winner := make(chan result, 1)
	go func() {
		child, _, rotateErr := m.Rotate(context.Background(), token, jkt)
		winner <- result{token: child, err: rotateErr}
	}()
	<-b.started
	if _, _, err := m.Rotate(context.Background(), token, jkt); !errors.Is(err, sessions.ErrReuse) {
		t.Fatalf("loser did not detect reuse: %v", err)
	}
	close(b.allow)
	r := <-winner
	if r.err != nil {
		t.Fatalf("winner failed to insert child: %v", r.err)
	}
	if _, _, err := m.Rotate(context.Background(), r.token, jkt); !errors.Is(err, sessions.ErrRevoked) {
		t.Fatalf("late child remained usable: %v", err)
	}
}

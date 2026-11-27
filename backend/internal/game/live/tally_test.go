package live

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/tirivia/backend/internal/game/domain"
)

// TestTallyExactlyOnce: concurrent and repeated submissions of the same answer count once, across buckets.
func TestTallyExactlyOnce(t *testing.T) {
	c := client(t)
	a, players := NewAnswers(c), NewPlayers(c, time.Hour)
	ctx, show, now := context.Background(), id(t), time.Now()
	st := openQuestion(1, now)
	users := make([]string, 200)
	for i := range users {
		users[i] = id(t)
		if _, err := players.Join(ctx, show, users[i], 0); err != nil {
			t.Fatal(err)
		}
	}
	var wg sync.WaitGroup
	for i, u := range users {
		opt := map[bool]string{true: "right", false: "wrong"}[i%4 != 0]
		for range 3 {
			wg.Go(func() { _, _ = a.Submit(ctx, show, u, st, "q1", opt, false, now) })
		}
	}
	wg.Wait()
	got, err := Sum(ctx, c, show, 1)
	if err != nil || got["right"] != 150 || got["wrong"] != 50 || len(got) != 2 {
		t.Fatalf("totals: %v %v", got, err)
	}
	if err := Verify(ctx, c, show, 1); err != nil {
		t.Fatal(err)
	}
}

// TestTallySkipsVoidedAndLate: voided answers and late joiners are recorded but never tallied (AC-3, AC-8).
func TestTallySkipsVoidedAndLate(t *testing.T) {
	c := client(t)
	a := NewAnswers(c)
	ctx, show, now := context.Background(), id(t), time.Now()
	voided, late := id(t), id(t)
	setPlayer(t, c, show, voided, domain.Participant{LastOK: 0})
	setPlayer(t, c, show, late, domain.Participant{JoinedAtQ: 1})
	if o, err := a.Submit(ctx, show, voided, openQuestion(1, now), "q1", "", true, now); err != nil || o.Counted {
		t.Fatalf("voided: %+v %v", o, err)
	}
	if o, err := a.Submit(ctx, show, late, openQuestion(2, now), "q2", "right", false, now); err != nil || o.Counted {
		t.Fatalf("late: %+v %v", o, err)
	}
	for q := 1; q <= 2; q++ {
		if got, err := Sum(ctx, c, show, q); err != nil || len(got) != 0 {
			t.Fatalf("Q%d tally: %v %v", q, got, err)
		}
	}
}

// TestVerifyDetectsDrift: a tally that disagrees with the stored answers is reported.
func TestVerifyDetectsDrift(t *testing.T) {
	c := client(t)
	ctx, show := context.Background(), id(t)
	if err := c.Do(ctx, c.B().Hincrby().Key(tallyKeyFor(show, 1)).Field("right").Increment(1).Build()).Error(); err != nil {
		t.Fatal(err)
	}
	if err := Verify(ctx, c, show, 1); err == nil {
		t.Fatal("drift not detected")
	}
}

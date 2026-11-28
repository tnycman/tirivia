package live

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"trivia.al/backend/internal/game/domain"
)

func TestIntakeMatchesDomainRules(t *testing.T) {
	c := client(t)
	a, players := NewAnswers(c), NewPlayers(c, time.Hour)
	ctx, show, now := context.Background(), id(t), time.Now()
	for joined := 0; joined <= domain.QuestionsPerShow; joined++ {
		for lastOK := 0; lastOK <= domain.QuestionsPerShow; lastOK++ {
			for q := 1; q <= domain.QuestionsPerShow; q++ {
				p := domain.Participant{JoinedAtQ: joined, LastOK: lastOK}
				user := id(t)
				setPlayer(t, c, show, user, p)
				out, err := a.Submit(ctx, show, user, openQuestion(q, now), "q"+strconv.Itoa(q), "right", false, now)
				if accepted := err == nil; accepted != p.CanAnswer(q) {
					t.Fatalf("joined=%d last_ok=%d q=%d: script accepted=%v (%v), domain=%v", joined, lastOK, q, accepted, err, p.CanAnswer(q))
				}
				if err != nil {
					if !errors.Is(err, ErrNotAlive) {
						t.Fatalf("unexpected rejection: %v", err)
					}
					continue
				}
				if out.Counted == p.PlaysAlong() || !out.Correct {
					t.Fatalf("joined=%d q=%d: outcome %+v", joined, q, out)
				}
				got, _ := players.Join(ctx, show, user, 0)
				if want := map[bool]int{true: lastOK, false: q}[p.PlaysAlong()]; got.LastOK != want {
					t.Fatalf("joined=%d q=%d: last_ok=%d want %d", joined, q, got.LastOK, want)
				}
			}
		}
	}
}

func TestIntakeRejects(t *testing.T) {
	c := client(t)
	a, players := NewAnswers(c), NewPlayers(c, time.Hour)
	ctx, show, user, now := context.Background(), id(t), id(t), time.Now()
	st := openQuestion(1, now)
	if _, err := a.Submit(ctx, show, user, st, "q1", "right", false, now); !errors.Is(err, ErrNotJoined) {
		t.Fatalf("not joined: %v", err)
	}
	if _, err := players.Join(ctx, show, user, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Submit(ctx, show, user, st, "q2", "right", false, now); !errors.Is(err, ErrNotQuestion) {
		t.Fatalf("other question: %v", err)
	}
	if _, err := a.Submit(ctx, show, user, st, "q1", "right", false, st.Deadline.Add(time.Millisecond)); !errors.Is(err, ErrClosed) {
		t.Fatalf("after deadline: %v", err)
	}
	if out, err := a.Submit(ctx, show, user, st, "q1", "wrong-option", false, st.Deadline); err != nil || out.Correct || !out.Counted {
		t.Fatalf("on the deadline: %+v %v", out, err)
	}
	if _, err := a.Submit(ctx, show, user, st, "q1", "right", false, now); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("second answer: %v", err)
	}
	if _, err := a.Submit(ctx, show, user, openQuestion(2, now), "q2", "right", false, now); !errors.Is(err, ErrNotAlive) {
		t.Fatalf("eliminated player answered Q2: %v", err)
	}
}

func TestConcurrentAnswersRecordOnce(t *testing.T) {
	c := client(t)
	a, players := NewAnswers(c), NewPlayers(c, time.Hour)
	ctx, show, user, now := context.Background(), id(t), id(t), time.Now()
	if _, err := players.Join(ctx, show, user, 0); err != nil {
		t.Fatal(err)
	}
	var ok atomic.Int32
	var wg sync.WaitGroup
	for i := range 16 {
		wg.Go(func() {
			opt := map[bool]string{true: "right", false: "wrong"}[i%2 == 0]
			if _, err := a.Submit(ctx, show, user, openQuestion(1, now), "q1", opt, false, now); err == nil {
				ok.Add(1)
			}
		})
	}
	wg.Wait()
	if ok.Load() != 1 {
		t.Fatalf("%d concurrent answers recorded; want 1", ok.Load())
	}
}

package live

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/tirivia/backend/internal/game/domain"
)

func TestLifeMatchesDomainRules(t *testing.T) {
	c := client(t)
	lives, players := NewLives(c), NewPlayers(c, time.Hour)
	ctx, show := context.Background(), id(t)
	for joined := 0; joined <= 3; joined++ {
		for lastOK := 0; lastOK <= domain.QuestionsPerShow; lastOK++ {
			for _, used := range []bool{false, true} {
				for q := 0; q <= domain.QuestionsPerShow; q++ {
					p := domain.Participant{JoinedAtQ: joined, LastOK: lastOK, LifeUsed: used}
					user := id(t)
					setPlayer(t, c, show, user, p)
					err := lives.Use(ctx, show, user, q, "q"+strconv.Itoa(q))
					if allowed := err == nil; allowed != p.CanUseExtraLife(q) {
						t.Fatalf("%+v q=%d: script allowed=%v (%v), domain=%v", p, q, allowed, err, p.CanUseExtraLife(q))
					}
					if err != nil && !errors.Is(err, ErrLifeNotAllowed) {
						t.Fatalf("unexpected: %v", err)
					}
					if err == nil {
						got, _ := players.Join(ctx, show, user, 0)
						if got.LastOK != q || !got.LifeUsed || got.CanUseExtraLife(q+1) {
							t.Fatalf("after life on Q%d: %+v", q, got)
						}
					}
				}
			}
		}
	}
}

func TestLifeRefund(t *testing.T) {
	c := client(t)
	lives, players := NewLives(c), NewPlayers(c, time.Hour)
	ctx, show, user := context.Background(), id(t), id(t)
	setPlayer(t, c, show, user, domain.Participant{LastOK: 4})
	for range 2 {
		if err := lives.Use(ctx, show, user, 5, "q5"); err != nil {
			t.Fatalf("use or retry on the same question: %v", err)
		}
	}
	if got, _ := Sum(ctx, c, show, 5); got[LifeOption] != 1 {
		t.Fatalf("a retried life counted twice: %v", got)
	}
	if err := lives.Use(ctx, show, user, 6, "q6"); !errors.Is(err, ErrLifeNotAllowed) {
		t.Fatalf("second life on another question: %v", err)
	}
	for range 2 {
		if err := lives.Refund(ctx, show, user, 5); err != nil {
			t.Fatal(err)
		}
	}
	if got, _ := players.Join(ctx, show, user, 0); got.LastOK != 4 || got.LifeUsed {
		t.Fatalf("after refund: %+v", got)
	}
	if got, _ := Sum(ctx, c, show, 5); got[LifeOption] != 0 {
		t.Fatalf("refund must remove the count exactly once: %v", got)
	}
	if err := lives.Use(ctx, id(t), id(t), 5, "q5"); !errors.Is(err, ErrNotJoined) {
		t.Fatalf("not joined: %v", err)
	}
}

package live

import (
	"bytes"
	"context"
	"testing"
	"time"

	"trivia.al/backend/internal/platform/livestate"
)

func TestJoinKeepsFirstJoin(t *testing.T) {
	c := client(t)
	players := NewPlayers(c, time.Hour)
	ctx, show := context.Background(), id(t)
	onTime, late := id(t), id(t)
	for _, step := range []struct {
		user string
		q    int
		want int
	}{{onTime, 0, 0}, {onTime, 5, 0}, {late, 3, 3}, {late, 0, 3}} {
		p, err := players.Join(ctx, show, step.user, step.q)
		if err != nil || p.JoinedAtQ != step.want || p.LastOK != 0 || p.LifeUsed {
			t.Fatalf("join at Q%d: %+v %v (want joined_at_q=%d)", step.q, p, err, step.want)
		}
	}
	ttl, err := c.Do(ctx, c.B().Pttl().Key(livestate.PlayerKey(show, onTime)).Build()).AsInt64()
	if err != nil || ttl <= 0 || ttl > time.Hour.Milliseconds() {
		t.Fatalf("player TTL: %d %v", ttl, err)
	}
}

// TestJoinSeedIsStable: the shuffle seed is 32 random bytes made at the first join and kept on rejoin.
func TestJoinSeedIsStable(t *testing.T) {
	c := client(t)
	players := NewPlayers(c, time.Hour)
	ctx, show, a, b := context.Background(), id(t), id(t), id(t)
	_, first, err := players.JoinSeeded(ctx, show, a, 0)
	if err != nil || len(first) != SeedBytes {
		t.Fatalf("first join: %d bytes, %v", len(first), err)
	}
	if _, again, err := players.JoinSeeded(ctx, show, a, 4); err != nil || !bytes.Equal(again, first) {
		t.Fatalf("rejoin changed the seed: %x -> %x (%v)", first, again, err)
	}
	if _, other, _ := players.JoinSeeded(ctx, show, b, 0); bytes.Equal(other, first) {
		t.Fatal("two players got the same seed")
	}
}

func TestFullShowWinner(t *testing.T) {
	c := client(t)
	players, a := NewPlayers(c, time.Hour), NewAnswers(c)
	ctx, show, now := context.Background(), id(t), time.Now()
	winner, loser := id(t), id(t)
	for _, u := range []string{winner, loser} {
		if _, err := players.Join(ctx, show, u, 0); err != nil {
			t.Fatal(err)
		}
	}
	for q := 1; q <= 12; q++ {
		st := openQuestion(q, now)
		if _, err := a.Submit(ctx, show, winner, st, st.QuestionID, "right", false, now); err != nil {
			t.Fatalf("winner Q%d: %v", q, err)
		}
		opt := "right"
		if q == 7 {
			opt = "wrong"
		}
		_, _ = a.Submit(ctx, show, loser, st, st.QuestionID, opt, false, now)
	}
	w, _ := players.Join(ctx, show, winner, 0)
	l, _ := players.Join(ctx, show, loser, 0)
	if !w.IsWinner() || l.IsWinner() || l.LastOK != 6 {
		t.Fatalf("winner %+v loser %+v", w, l)
	}
}

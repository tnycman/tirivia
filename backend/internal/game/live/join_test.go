package live

import (
	"context"
	"testing"
	"time"

	"github.com/tirivia/backend/internal/platform/livestate"
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
		if _, err := a.Submit(ctx, show, winner, st, st.QuestionID, "right", now); err != nil {
			t.Fatalf("winner Q%d: %v", q, err)
		}
		opt := "right"
		if q == 7 {
			opt = "wrong"
		}
		_, _ = a.Submit(ctx, show, loser, st, st.QuestionID, opt, now)
	}
	w, _ := players.Join(ctx, show, winner, 0)
	l, _ := players.Join(ctx, show, loser, 0)
	if !w.IsWinner() || l.IsWinner() || l.LastOK != 6 {
		t.Fatalf("winner %+v loser %+v", w, l)
	}
}

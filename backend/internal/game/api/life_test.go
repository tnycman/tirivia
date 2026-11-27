package api

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect/v2"

	cuev1 "github.com/tirivia/protocol/gen/go/tirivia/cue/v1"
	gamev1 "github.com/tirivia/protocol/gen/go/tirivia/game/v1"

	"github.com/tirivia/backend/internal/game/live"
)

func (g game) useLife(user string, q string) (*gamev1.UseExtraLifeResponse, error) {
	return g.client.UseExtraLife(as(user), &gamev1.UseExtraLifeRequest{ShowId: g.show, QuestionId: q})
}

func TestExtraLife(t *testing.T) {
	g := newGame(t)
	ctx := context.Background()
	saved, unlucky := rnd(t), rnd(t)
	for _, u := range []string{saved, unlucky} {
		if _, err := g.players.Join(ctx, g.show, u, 0); err != nil {
			t.Fatal(err)
		}
	}
	g.inv.n[saved], g.inv.n[unlucky] = 1, 1
	g.set(t, cuev1.Phase_PHASE_QUESTION_OPEN, 1, 10*time.Second)
	g.answer(t, saved, 1, "wrong", nil)
	g.answer(t, unlucky, 1, "wrong", nil)
	if _, err := g.useLife(saved, "q1"); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("life before the reveal: %v", err)
	}
	g.set(t, cuev1.Phase_PHASE_REVEAL, 1, 0)
	res, err := g.useLife(saved, "q1")
	if err != nil || res.GetPlayerState().GetStatus() != gamev1.PlayerStatus_PLAYER_STATUS_ALIVE || !res.GetPlayerState().GetExtraLifeUsedThisShow() {
		t.Fatalf("life during reveal: %v %v", res, err)
	}
	if g.inv.n[saved] != 0 {
		t.Fatalf("inventory not consumed: %d", g.inv.n[saved])
	}
	if _, err := g.useLife(saved, "q1"); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("second life: %v", err)
	}
	g.inv.failing = true
	if _, err := g.useLife(unlucky, "q1"); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("inventory failure: %v", err)
	}
	if st := g.status(t, unlucky).GetPlayerState(); st.GetStatus() != gamev1.PlayerStatus_PLAYER_STATUS_ELIMINATED || st.GetExtraLifeUsedThisShow() {
		t.Fatalf("refund must restore the player: %v", st)
	}
	if err := g.tally.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	if sum, _ := live.Sum(ctx, g.kv, g.show, 1); sum[live.LifeOption] != 1 {
		t.Fatalf("lives used on Q1: %v", sum)
	}
}

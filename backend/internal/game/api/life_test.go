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
	if _, err := g.useLife(saved, "q1"); err != nil || g.inv.n[saved] != 0 {
		t.Fatalf("a repeat on the same question is a retry, not a second debit: %v, inventory %d", err, g.inv.n[saved])
	}
	g.inv.failing = true
	if _, err := g.useLife(unlucky, "q1"); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("inventory failure: %v", err)
	}
	if st := g.status(t, unlucky).GetPlayerState(); st.GetStatus() != gamev1.PlayerStatus_PLAYER_STATUS_ELIMINATED || st.GetExtraLifeUsedThisShow() {
		t.Fatalf("refund must restore the player: %v", st)
	}
	if sum, _ := live.Sum(ctx, g.kv, g.show, 1); sum[live.LifeOption] != 1 {
		t.Fatalf("lives used on Q1 (the refund must cancel its count): %v", sum)
	}
}

// TestExtraLifeAmbiguousDebit: a lost inventory reply is retried, never refunded, and nothing is charged or counted twice.
func TestExtraLifeAmbiguousDebit(t *testing.T) {
	g := newGame(t)
	ctx := context.Background()
	retried, unread := rnd(t), rnd(t)
	for _, u := range []string{retried, unread} {
		if _, err := g.players.Join(ctx, g.show, u, 0); err != nil {
			t.Fatal(err)
		}
		g.inv.n[u] = 1
	}
	g.set(t, cuev1.Phase_PHASE_QUESTION_OPEN, 1, 10*time.Second)
	g.answer(t, retried, 1, "wrong", nil)
	g.answer(t, unread, 1, "wrong", nil)
	g.set(t, cuev1.Phase_PHASE_REVEAL, 1, 0)
	g.inv.lostReply = true
	if _, err := g.useLife(retried, "q1"); connect.CodeOf(err) != connect.CodeUnavailable {
		t.Fatalf("lost reply: %v, want Unavailable", err)
	}
	res, err := g.useLife(retried, "q1")
	if err != nil || res.GetPlayerState().GetStatus() != gamev1.PlayerStatus_PLAYER_STATUS_ALIVE {
		t.Fatalf("retry after lost reply: %v %v", res, err)
	}
	if g.inv.n[retried] != 0 {
		t.Fatalf("debited twice or refunded: %d", g.inv.n[retried])
	}
	late := rnd(t)
	if _, err := g.players.Join(ctx, g.show, late, 0); err != nil {
		t.Fatal(err)
	}
	g.inv.n[late] = 1
	g.set(t, cuev1.Phase_PHASE_QUESTION_OPEN, 1, 10*time.Second)
	g.answer(t, late, 1, "wrong", nil)
	g.set(t, cuev1.Phase_PHASE_REVEAL, 1, 0)
	g.inv.lostReply = true
	if _, err := g.useLife(late, "q1"); connect.CodeOf(err) != connect.CodeUnavailable {
		t.Fatalf("lost reply: %v", err)
	}
	g.set(t, cuev1.Phase_PHASE_QUESTION_OPEN, 2, 10*time.Second)
	if _, err := g.useLife(late, "q1"); err != nil {
		t.Fatalf("a retry after the show moved on must still settle the debit: %v", err)
	}
	if _, err := g.useLife(late, "q2"); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("a request for another question is not a retry: %v", err)
	}
	g.set(t, cuev1.Phase_PHASE_REVEAL, 1, 0)
	g.inv.readErr = true
	if _, err := g.useLife(unread, "q1"); err != nil {
		t.Fatalf("a spent life must not turn into an error when the state read fails: %v", err)
	}
	if sum, _ := live.Sum(ctx, g.kv, g.show, 1); sum[live.LifeOption] != 3 {
		t.Fatalf("each life counted once: %v", sum)
	}
}

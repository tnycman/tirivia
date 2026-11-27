package api

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect/v2"

	cuev1 "github.com/tirivia/protocol/gen/go/tirivia/cue/v1"
	gamev1 "github.com/tirivia/protocol/gen/go/tirivia/game/v1"

	"github.com/tirivia/backend/internal/game/live"
)

func (g game) join(user string) (*gamev1.JoinShowResponse, error) {
	return g.client.JoinShow(as(user), &gamev1.JoinShowRequest{ShowId: g.show})
}

// TestJoinShowWindow: joins open from LOBBY through FINAL (UF-16); a refused join leaves no player record.
func TestJoinShowWindow(t *testing.T) {
	g := newGame(t)
	user := rnd(t)
	if _, err := g.join(user); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("no live state: %v, want NotFound", err)
	}
	for _, ph := range []cuev1.Phase{cuev1.Phase_PHASE_SCHEDULED, cuev1.Phase_PHASE_SETTLED} {
		g.set(t, ph, 0, 0)
		if _, err := g.join(user); connect.CodeOf(err) != connect.CodeNotFound {
			t.Fatalf("%v: %v, want NotFound", ph, err)
		}
	}
	if _, err := g.players.Get(context.Background(), g.show, user); !errors.Is(err, live.ErrNotJoined) {
		t.Fatalf("a refused join created a player: %v", err)
	}
	for _, ph := range []cuev1.Phase{cuev1.Phase_PHASE_LOBBY, cuev1.Phase_PHASE_REVEAL, cuev1.Phase_PHASE_FINAL} {
		g.set(t, ph, 3, 0)
		if _, err := g.join(rnd(t)); err != nil {
			t.Fatalf("%v: %v", ph, err)
		}
	}
	if _, err := g.client.JoinShow(context.Background(), &gamev1.JoinShowRequest{ShowId: g.show}); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("anonymous join: %v", err)
	}
}

// TestJoinShowRejoinKeepsStanding: a rejoin mid-show returns the same seed and keeps the player on time.
func TestJoinShowRejoinKeepsStanding(t *testing.T) {
	g := newGame(t)
	user := rnd(t)
	g.set(t, cuev1.Phase_PHASE_LOBBY, 0, 0)
	first, err := g.join(user)
	if err != nil || len(first.GetShuffleSeed()) != live.SeedBytes || first.GetPlayerState().GetStatus() != gamev1.PlayerStatus_PLAYER_STATUS_ALIVE {
		t.Fatalf("lobby join: %v %v", first, err)
	}
	if first.GetReactionBatchIntervalMs() != 5000 || len(first.GetCueKeys()) != 1 || first.GetCueKeys()[0].GetKeyId() != testCueKeyID || first.GetChatEnabled() {
		t.Fatalf("server tunables: %v", first)
	}
	g.set(t, cuev1.Phase_PHASE_QUESTION_OPEN, 1, 10*time.Second)
	g.answer(t, user, 1, "right", nil)
	g.set(t, cuev1.Phase_PHASE_QUESTION_OPEN, 2, 10*time.Second)
	again, err := g.join(user)
	if err != nil || !bytes.Equal(again.GetShuffleSeed(), first.GetShuffleSeed()) || again.GetPlayerState().GetStatus() != gamev1.PlayerStatus_PLAYER_STATUS_ALIVE {
		t.Fatalf("rejoin: %v %v", again, err)
	}
	if got := g.answer(t, user, 2, "right", nil); got != gamev1.AnswerResult_ANSWER_RESULT_ACCEPTED {
		t.Fatalf("on-time player after rejoin: %v", got)
	}
}

// TestJoinShowLateJoiner: joining while Q2 is open plays along from Q3 on (AC-8).
func TestJoinShowLateJoiner(t *testing.T) {
	g := newGame(t)
	user := rnd(t)
	g.set(t, cuev1.Phase_PHASE_QUESTION_OPEN, 2, 10*time.Second)
	res, err := g.join(user)
	if err != nil || res.GetPlayerState().GetStatus() != gamev1.PlayerStatus_PLAYER_STATUS_PLAY_ALONG {
		t.Fatalf("late join: %v %v", res, err)
	}
	if got := g.answer(t, user, 2, "right", nil); got != gamev1.AnswerResult_ANSWER_RESULT_NOT_ALIVE {
		t.Fatalf("the question open at join must count as missed: %v", got)
	}
	g.set(t, cuev1.Phase_PHASE_QUESTION_OPEN, 3, 10*time.Second)
	if got := g.answer(t, user, 3, "right", nil); got != gamev1.AnswerResult_ANSWER_RESULT_ACCEPTED {
		t.Fatalf("next question: %v", got)
	}
}

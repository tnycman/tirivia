package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect/v2"

	cuev1 "github.com/tirivia/protocol/gen/go/tirivia/cue/v1"
	gamev1 "github.com/tirivia/protocol/gen/go/tirivia/game/v1"
	"github.com/tirivia/protocol/gen/go/tirivia/game/v1/gamev1connect"

	"github.com/tirivia/backend/internal/game/live"
	"github.com/tirivia/backend/internal/platform/livestate"
)

// TestWiredServerAuthenticates drives the production wiring with a real token and DPoP proof.
func TestWiredServerAuthenticates(t *testing.T) {
	w := newWired(t)
	req := &gamev1.GetShowStateRequest{ShowId: "no-such-show"}
	ctx := w.signed(t, gamev1connect.GameServiceGetShowStateProcedure)
	if _, err := w.client.GetShowState(ctx, req); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("authenticated call: %v, want NotFound (show not live)", err)
	}
	if _, err := w.client.GetShowState(ctx, req); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("replayed proof: %v, want Unauthenticated", err)
	}
	if _, err := w.client.GetShowState(context.Background(), req); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("anonymous call: %v, want Unauthenticated", err)
	}
}

// TestWiredAnswerIsTallied sends one answer through the production stores and the tally flush.
func TestWiredAnswerIsTallied(t *testing.T) {
	w := newWired(t)
	bg, show, now := context.Background(), "show-"+w.suffix, time.Now()
	err := w.deps.Shows.Write(bg, show, livestate.Show{Phase: int32(cuev1.Phase_PHASE_QUESTION_OPEN), QuestionIndex: 1,
		QuestionID: "q1", TCue: now, CloseAt: now.Add(time.Minute), Deadline: now.Add(time.Minute), CorrectOption: "right", CueSeq: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.deps.Players.Join(bg, show, w.user, 0); err != nil {
		t.Fatal(err)
	}
	res, err := w.client.SubmitAnswer(w.signed(t, gamev1connect.GameServiceSubmitAnswerProcedure),
		&gamev1.SubmitAnswerRequest{ShowId: show, QuestionId: "q1", OptionId: "right"})
	if err != nil || res.GetResult() != gamev1.AnswerResult_ANSWER_RESULT_ACCEPTED {
		t.Fatalf("answer: %v %v", res.GetResult(), err)
	}
	if err := w.deps.Tally.Flush(bg); err != nil {
		t.Fatal(err)
	}
	if got, err := live.Sum(bg, w.kv, show, 1); err != nil || got["right"] != 1 {
		t.Fatalf("tally = %v, %v; want right=1", got, err)
	}
}

func (w wired) signed(t *testing.T, procedure string) context.Context {
	t.Helper()
	proof, err := w.device.Proof(dpopProof(http.MethodPost, w.url+procedure, w.token))
	if err != nil {
		t.Fatal(err)
	}
	ctx, info := connect.NewClientContext(context.Background())
	info.RequestHeader().Set("Authorization", "DPoP "+w.token)
	info.RequestHeader().Set("DPoP", proof)
	return ctx
}

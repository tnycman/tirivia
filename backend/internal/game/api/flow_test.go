package api

import (
	"context"
	"strconv"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"google.golang.org/protobuf/proto"

	cuev1 "github.com/tirivia/protocol/gen/go/tirivia/cue/v1"
	gamev1 "github.com/tirivia/protocol/gen/go/tirivia/game/v1"

	"github.com/tirivia/backend/internal/game/live"
	"github.com/tirivia/backend/internal/platform/livestate"
)

func (g game) set(t *testing.T, phase cuev1.Phase, q int, closeIn time.Duration) {
	t.Helper()
	open, err := proto.Marshal(&cuev1.QuestionOpen{QuestionId: "q" + strconv.Itoa(q), Index: uint32(q), Text: "Pyetja?"}) //nolint:gosec // G115: q ≤ 12
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	err = g.shows.Write(context.Background(), g.show, livestate.Show{Phase: int32(phase), QuestionIndex: q, QuestionID: "q" + strconv.Itoa(q),
		TCue: now, CloseAt: now.Add(closeIn), Deadline: now.Add(closeIn + 1500*time.Millisecond), OpenCue: open, CorrectOption: "right", CueSeq: 1})
	if err != nil {
		t.Fatal(err)
	}
}

func (g game) answer(t *testing.T, user string, q int, option string, sig *gamev1.IntegritySignals) gamev1.AnswerResult {
	t.Helper()
	res, err := g.client.SubmitAnswer(as(user), &gamev1.SubmitAnswerRequest{ShowId: g.show, QuestionId: "q" + strconv.Itoa(q), OptionId: option, Signals: sig})
	if err != nil {
		t.Fatalf("%s Q%d: %v", user, q, err)
	}
	return res.GetResult()
}

func (g game) status(t *testing.T, user string) *gamev1.GetShowStateResponse {
	t.Helper()
	res, err := g.client.GetShowState(as(user), &gamev1.GetShowStateRequest{ShowId: g.show})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestAnswerFlow(t *testing.T) {
	g := newGame(t)
	ctx := context.Background()
	onTime, late := rnd(t), rnd(t)
	g.set(t, cuev1.Phase_PHASE_QUESTION_OPEN, 1, 10*time.Second)
	if _, err := g.players.Join(ctx, g.show, onTime, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := g.players.Join(ctx, g.show, late, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := g.client.SubmitAnswer(context.Background(), &gamev1.SubmitAnswerRequest{ShowId: g.show, QuestionId: "q1", OptionId: "right"}); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("anonymous answer: %v", err)
	}
	want := []struct {
		user, option string
		result       gamev1.AnswerResult
	}{
		{onTime, "right", gamev1.AnswerResult_ANSWER_RESULT_ACCEPTED},
		{onTime, "wrong", gamev1.AnswerResult_ANSWER_RESULT_DUPLICATE},
		{late, "right", gamev1.AnswerResult_ANSWER_RESULT_NOT_ALIVE},
		{rnd(t), "right", gamev1.AnswerResult_ANSWER_RESULT_NOT_ALIVE},
	}
	for _, w := range want {
		if got := g.answer(t, w.user, 1, w.option, nil); got != w.result {
			t.Errorf("%s %s: got %v want %v", w.user, w.option, got, w.result)
		}
	}
	s := g.status(t, onTime)
	if s.GetPlayerState().GetStatus() != gamev1.PlayerStatus_PLAYER_STATUS_ALIVE || s.GetOpenQuestion().GetText() != "Pyetja?" || s.GetRemainingMs() == 0 {
		t.Fatalf("state: %v", s)
	}
	if st := g.status(t, rnd(t)).GetPlayerState().GetStatus(); st != gamev1.PlayerStatus_PLAYER_STATUS_SPECTATOR {
		t.Fatalf("spectator: %v", st)
	}
	if st := g.status(t, late).GetPlayerState().GetStatus(); st != gamev1.PlayerStatus_PLAYER_STATUS_PLAY_ALONG {
		t.Fatalf("late joiner: %v", st)
	}
	if sum, _ := live.Sum(ctx, g.kv, g.show, 1); sum["right"] != 1 || len(sum) != 1 {
		t.Fatalf("tally must count only the on-time answer: %v", sum)
	}
	if err := live.Verify(ctx, g.kv, g.show, 1); err != nil {
		t.Fatal(err)
	}
	g.set(t, cuev1.Phase_PHASE_QUESTION_OPEN, 2, -2*time.Second)
	if got := g.answer(t, onTime, 2, "right", nil); got != gamev1.AnswerResult_ANSWER_RESULT_LATE {
		t.Fatalf("after deadline: %v", got)
	}
	g.set(t, cuev1.Phase_PHASE_REVEAL, 2, 0)
	if got := g.answer(t, onTime, 2, "right", nil); got != gamev1.AnswerResult_ANSWER_RESULT_QUESTION_NOT_OPEN {
		t.Fatalf("during reveal: %v", got)
	}
}

func TestFocusLossVoidsFinalThree(t *testing.T) {
	g := newGame(t)
	ctx, user := context.Background(), rnd(t)
	if _, err := g.players.Join(ctx, g.show, user, 0); err != nil {
		t.Fatal(err)
	}
	for q := 1; q <= 9; q++ {
		g.set(t, cuev1.Phase_PHASE_QUESTION_OPEN, q, 10*time.Second)
		sig := &gamev1.IntegritySignals{WentBackground: q == 9}
		if got := g.answer(t, user, q, "right", sig); got != gamev1.AnswerResult_ANSWER_RESULT_ACCEPTED {
			t.Fatalf("Q%d (rule starts at Q10): %v", q, got)
		}
	}
	g.set(t, cuev1.Phase_PHASE_QUESTION_OPEN, 10, 10*time.Second)
	if got := g.answer(t, user, 10, "right", &gamev1.IntegritySignals{LongestInactiveMs: 1600}); got != gamev1.AnswerResult_ANSWER_RESULT_VOIDED_FOCUS_LOSS {
		t.Fatalf("Q10 inactive 1.6 s: %v", got)
	}
	g.set(t, cuev1.Phase_PHASE_REVEAL, 10, 0)
	if st := g.status(t, user).GetPlayerState().GetStatus(); st != gamev1.PlayerStatus_PLAYER_STATUS_ELIMINATED {
		t.Fatalf("voided answer must count as missed: %v", st)
	}
}

package api

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect/v2"

	cuev1 "github.com/tirivia/protocol/gen/go/tirivia/cue/v1"
	gamev1 "github.com/tirivia/protocol/gen/go/tirivia/game/v1"
	"github.com/tirivia/protocol/gen/go/tirivia/game/v1/gamev1connect"

	"github.com/tirivia/backend/internal/game/domain"
	"github.com/tirivia/backend/internal/game/live"
	"github.com/tirivia/backend/internal/platform/livestate"
)

// voidedOption is recorded for a voided answer: it can never be correct, so the question counts as missed.
const voidedOption = ""

var answerErrors = []struct {
	err    error
	result gamev1.AnswerResult
}{
	{live.ErrDuplicate, gamev1.AnswerResult_ANSWER_RESULT_DUPLICATE},
	{live.ErrClosed, gamev1.AnswerResult_ANSWER_RESULT_LATE},
	{live.ErrNotAlive, gamev1.AnswerResult_ANSWER_RESULT_NOT_ALIVE},
	{live.ErrNotJoined, gamev1.AnswerResult_ANSWER_RESULT_NOT_ALIVE},
	{live.ErrNotQuestion, gamev1.AnswerResult_ANSWER_RESULT_QUESTION_NOT_OPEN},
	{livestate.ErrNoShow, gamev1.AnswerResult_ANSWER_RESULT_QUESTION_NOT_OPEN},
}

// SubmitAnswer records an answer against the server deadline; correctness is never returned before the reveal.
func (h *Handler) SubmitAnswer(ctx context.Context, req *gamev1.SubmitAnswerRequest) (*gamev1.SubmitAnswerResponse, error) {
	received := h.d.Now()
	p, err := h.caller(ctx, gamev1connect.GameServiceSubmitAnswerProcedure)
	if err != nil {
		return nil, err
	}
	if req.GetShowId() == "" || req.GetQuestionId() == "" || req.GetOptionId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, "show_id, question_id and option_id are required")
	}
	st, err := h.d.Shows.Read(ctx, req.GetShowId())
	if err == nil && !answering(st.Phase) {
		err = live.ErrNotQuestion
	}
	if err != nil {
		return answerResult(err)
	}
	sig := req.GetSignals()
	voided := domain.FocusLossVoids(st.QuestionIndex, sig.GetWentBackground(), time.Duration(sig.GetLongestInactiveMs())*time.Millisecond)
	option := req.GetOptionId()
	if voided {
		option = voidedOption
	}
	out, err := h.d.Answers.Submit(ctx, req.GetShowId(), p.UserID, st, req.GetQuestionId(), option, received)
	if err != nil {
		return answerResult(err)
	}
	if voided {
		return &gamev1.SubmitAnswerResponse{Result: gamev1.AnswerResult_ANSWER_RESULT_VOIDED_FOCUS_LOSS}, nil
	}
	if out.Counted {
		h.d.Tally.Add(req.GetShowId(), st.QuestionIndex, option)
	}
	return &gamev1.SubmitAnswerResponse{Result: gamev1.AnswerResult_ANSWER_RESULT_ACCEPTED}, nil
}

// answering covers QUESTION_CLOSED too: uploads are accepted until the deadline, after question_closed (SPEC §6.3).
func answering(phase int32) bool {
	return phase == int32(cuev1.Phase_PHASE_QUESTION_OPEN) || phase == int32(cuev1.Phase_PHASE_QUESTION_CLOSED)
}

func answerResult(err error) (*gamev1.SubmitAnswerResponse, error) {
	for _, m := range answerErrors {
		if errors.Is(err, m.err) {
			return &gamev1.SubmitAnswerResponse{Result: m.result}, nil
		}
	}
	return nil, connect.NewError(connect.CodeUnavailable, "answer not recorded; retry")
}

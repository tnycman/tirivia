package api

import (
	"context"
	"errors"

	"connectrpc.com/connect/v2"

	cuev1 "trivia.al/protocol/gen/go/trivia/cue/v1"
	gamev1 "trivia.al/protocol/gen/go/trivia/game/v1"
	"trivia.al/protocol/gen/go/trivia/game/v1/gamev1connect"

	"trivia.al/backend/internal/game/live"
)

var errNoLife = connect.NewError(connect.CodeFailedPrecondition, "no extra life can be used now")

// UseExtraLife applies UF-4 during the reveal: the rule is spent first, then the idempotent inventory debit.
// A retry for the question the life was already spent on settles the debit whatever the phase is now.
func (h *Handler) UseExtraLife(ctx context.Context, req *gamev1.UseExtraLifeRequest) (*gamev1.UseExtraLifeResponse, error) {
	p, err := h.caller(ctx, gamev1connect.GameServiceUseExtraLifeProcedure)
	if err != nil {
		return nil, err
	}
	if h.d.Inventory == nil {
		return nil, errNoLife
	}
	q, qid, spent, err := h.d.Lives.Spent(ctx, req.GetShowId(), p.UserID)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, "retry")
	}
	if spent && qid == req.GetQuestionId() {
		return h.settleLife(ctx, req.GetShowId(), p.UserID, q)
	}
	st, err := h.d.Shows.Read(ctx, req.GetShowId())
	if err != nil || st.Phase != int32(cuev1.Phase_PHASE_REVEAL) || st.QuestionID != req.GetQuestionId() {
		return nil, errNoLife
	}
	err = h.d.Lives.Use(ctx, req.GetShowId(), p.UserID, st.QuestionIndex, st.QuestionID)
	if errors.Is(err, live.ErrLifeNotAllowed) || errors.Is(err, live.ErrNotJoined) {
		return nil, errNoLife
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, "retry")
	}
	return h.settleLife(ctx, req.GetShowId(), p.UserID, st.QuestionIndex)
}

package api

import (
	"context"
	"errors"

	"connectrpc.com/connect/v2"

	cuev1 "github.com/tirivia/protocol/gen/go/tirivia/cue/v1"
	gamev1 "github.com/tirivia/protocol/gen/go/tirivia/game/v1"
	"github.com/tirivia/protocol/gen/go/tirivia/game/v1/gamev1connect"

	"github.com/tirivia/backend/internal/game/live"
)

var errNoLife = connect.NewError(connect.CodeFailedPrecondition, "no extra life can be used now")

// UseExtraLife applies UF-4 during the reveal: the rule is spent first, then the idempotent inventory debit.
// Only a definite refusal is refunded; an ambiguous error is left for the client's retry to resolve.
func (h *Handler) UseExtraLife(ctx context.Context, req *gamev1.UseExtraLifeRequest) (*gamev1.UseExtraLifeResponse, error) {
	p, err := h.caller(ctx, gamev1connect.GameServiceUseExtraLifeProcedure)
	if err != nil {
		return nil, err
	}
	st, err := h.d.Shows.Read(ctx, req.GetShowId())
	if err != nil || st.Phase != int32(cuev1.Phase_PHASE_REVEAL) || st.QuestionID != req.GetQuestionId() || h.d.Inventory == nil {
		return nil, errNoLife
	}
	err = h.d.Lives.Use(ctx, req.GetShowId(), p.UserID, st.QuestionIndex)
	if errors.Is(err, live.ErrLifeNotAllowed) || errors.Is(err, live.ErrNotJoined) {
		return nil, errNoLife
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, "retry")
	}
	ok, err := h.d.Inventory.Consume(ctx, p.UserID, req.GetShowId())
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, "retry")
	}
	if !ok {
		if rerr := h.d.Lives.Refund(ctx, req.GetShowId(), p.UserID, st.QuestionIndex); rerr != nil {
			return nil, connect.NewError(connect.CodeUnavailable, "retry")
		}
		return nil, errNoLife
	}
	ps, err := h.playerState(ctx, req.GetShowId(), p.UserID, st)
	if err != nil {
		return &gamev1.UseExtraLifeResponse{}, nil //nolint:nilerr // the life is spent; GetShowState resyncs the state
	}
	return &gamev1.UseExtraLifeResponse{PlayerState: ps}, nil
}

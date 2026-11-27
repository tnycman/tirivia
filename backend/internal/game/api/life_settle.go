package api

import (
	"context"

	"connectrpc.com/connect/v2"

	gamev1 "github.com/tirivia/protocol/gen/go/tirivia/game/v1"
)

// settleLife debits the inventory for a life spent on question q; only a definite refusal is refunded.
func (h *Handler) settleLife(ctx context.Context, show, user string, q int) (*gamev1.UseExtraLifeResponse, error) {
	ok, err := h.d.Inventory.Consume(ctx, user, show)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, "retry")
	}
	if !ok {
		if err := h.d.Lives.Refund(ctx, show, user, q); err != nil {
			return nil, connect.NewError(connect.CodeUnavailable, "retry")
		}
		return nil, errNoLife
	}
	st, err := h.d.Shows.Read(ctx, show)
	if err != nil {
		return &gamev1.UseExtraLifeResponse{}, nil //nolint:nilerr // the life is spent; GetShowState resyncs the state
	}
	ps, err := h.playerState(ctx, show, user, st)
	if err != nil {
		return &gamev1.UseExtraLifeResponse{}, nil //nolint:nilerr // the life is spent; GetShowState resyncs the state
	}
	return &gamev1.UseExtraLifeResponse{PlayerState: ps}, nil
}

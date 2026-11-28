package api

import (
	"context"
	"net/http"

	"connectrpc.com/connect/v2"

	"trivia.al/backend/internal/contracts"
)

var errUnauth = connect.NewError(connect.CodeUnauthenticated, "authentication failed")

func (h *Handler) caller(ctx context.Context, procedure string) (contracts.Principal, error) {
	info, ok := connect.CallInfoForServerContext(ctx)
	if !ok || h.d.Authn == nil {
		return contracts.Principal{}, errUnauth
	}
	p, err := h.d.Authn.Authenticate(ctx, info.RequestHeader(), http.MethodPost, h.d.PublicBaseURL+procedure)
	if err != nil {
		return contracts.Principal{}, errUnauth
	}
	return p, nil
}

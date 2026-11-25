// Package api exposes the game module over Connect (GameService). Handlers stay thin:
// authenticate → authorize → validate → rate limit → app use case → emit event.
package api

import (
	"context"

	"connectrpc.com/connect/v2"

	gamev1 "github.com/tirivia/protocol/gen/go/tirivia/game/v1"
	"github.com/tirivia/protocol/gen/go/tirivia/game/v1/gamev1connect"

	"github.com/tirivia/backend/internal/platform/switches"
)

// ChatTokenMinter is implemented by the chatmod module (Phase 3) via internal/contracts.
type ChatTokenMinter interface {
	MintChatToken(ctx context.Context, showID string) (token, endpoint string, err error)
}

// Handler implements gamev1connect.GameServiceHandler. Unimplemented RPCs return
// CodeUnimplemented until their phase lands (PLAN.md).
type Handler struct {
	gamev1connect.UnimplementedGameServiceHandler

	switches switches.Reader
	chat     ChatTokenMinter
}

// NewHandler wires the handler's dependencies.
func NewHandler(sw switches.Reader, chat ChatTokenMinter) *Handler {
	return &Handler{switches: sw, chat: chat}
}

// GetChatToken returns a chat token only while the backend chat switch is on (ADR-0009).
func (h *Handler) GetChatToken(ctx context.Context, req *gamev1.GetChatTokenRequest) (*gamev1.GetChatTokenResponse, error) {
	if req.GetShowId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, "show_id is required")
	}
	if !h.switches.Enabled(ctx, switches.ChatEnabled, req.GetShowId()) {
		return nil, connect.NewError(connect.CodeFailedPrecondition, "chat is switched off for this show")
	}
	if h.chat == nil {
		return nil, connect.NewError(connect.CodeUnimplemented, "chat token minting lands in Phase 3")
	}
	token, endpoint, err := h.chat.MintChatToken(ctx, req.GetShowId())
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, "chat temporarily unavailable")
	}
	return &gamev1.GetChatTokenResponse{ChatToken: token, ChatEndpoint: endpoint}, nil
}

// Package api exposes the game module as GameService.
package api

import (
	"context"
	"strings"
	"time"

	"connectrpc.com/connect/v2"

	gamev1 "trivia.al/protocol/gen/go/trivia/game/v1"
	"trivia.al/protocol/gen/go/trivia/game/v1/gamev1connect"

	"trivia.al/backend/internal/contracts"
	"trivia.al/backend/internal/game/live"
	"trivia.al/backend/internal/platform/livestate"
	"trivia.al/backend/internal/platform/switches"
)

// ChatTokenMinter mints chat tokens (chatmod module).
type ChatTokenMinter interface {
	MintChatToken(ctx context.Context, showID string) (token, endpoint string, err error)
}

// LifeInventory holds extra lives earned by referral (UF-5); Consume must be idempotent per (user, show), so a retry never debits twice.
type LifeInventory interface {
	Available(ctx context.Context, user string) (int, error)
	Consume(ctx context.Context, user, show string) (bool, error)
}

// Deps wires the handler; PublicBaseURL is the external origin DPoP htu must name.
type Deps struct {
	PublicBaseURL string
	Switches      switches.Reader
	Chat          ChatTokenMinter
	Authn         contracts.Authenticator
	Shows         *livestate.Shows
	Players       *live.Players
	Answers       *live.Answers
	Lives         *live.Lives
	Inventory     LifeInventory
	// CueKeys are the public keys apps accept for cue signatures (current + next).
	CueKeys []*gamev1.CueKey
	// ReactionBatchInterval is pushed to apps at join (TS-9, ADR-0006).
	ReactionBatchInterval time.Duration
	Now                   func() time.Time
}

// Handler implements gamev1connect.GameServiceHandler.
type Handler struct {
	gamev1connect.UnimplementedGameServiceHandler
	d Deps
}

// New returns a Handler.
func New(d Deps) *Handler {
	d.PublicBaseURL = strings.TrimRight(d.PublicBaseURL, "/")
	if d.Now == nil {
		d.Now = time.Now
	}
	return &Handler{d: d}
}

// GetChatToken mints a token only while the chat switch is on (ADR-0009).
func (h *Handler) GetChatToken(ctx context.Context, req *gamev1.GetChatTokenRequest) (*gamev1.GetChatTokenResponse, error) {
	if req.GetShowId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, "show_id is required")
	}
	if !h.d.Switches.Enabled(ctx, switches.ChatEnabled, req.GetShowId()) {
		return nil, connect.NewError(connect.CodeFailedPrecondition, "chat is switched off for this show")
	}
	if h.d.Chat == nil {
		return nil, connect.NewError(connect.CodeUnimplemented, "chat token minting lands in Phase 3")
	}
	token, endpoint, err := h.d.Chat.MintChatToken(ctx, req.GetShowId())
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, "chat temporarily unavailable")
	}
	return &gamev1.GetChatTokenResponse{ChatToken: token, ChatEndpoint: endpoint}, nil
}

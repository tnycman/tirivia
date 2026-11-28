package api

import (
	"context"

	"trivia.al/backend/internal/platform/switches"
)

// joinChat returns the chat switch and, if a mint succeeds, a token; otherwise the app calls GetChatToken (ADR-0009).
func (h *Handler) joinChat(ctx context.Context, show string) (enabled bool, token, endpoint string) {
	enabled = h.d.Switches.Enabled(ctx, switches.ChatEnabled, show)
	if !enabled || h.d.Chat == nil {
		return enabled, "", ""
	}
	token, endpoint, err := h.d.Chat.MintChatToken(ctx, show)
	if err != nil {
		return true, "", ""
	}
	return true, token, endpoint
}

package api

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect/v2"

	gamev1 "github.com/tirivia/protocol/gen/go/tirivia/game/v1"

	"github.com/tirivia/backend/internal/platform/switches"
)

type fakeMinter struct{}

func (fakeMinter) MintChatToken(context.Context, string) (string, string, error) {
	return "tok", "wss://edge.ivschat.eu-central-1.amazonaws.com", nil
}

func codeOf(err error) connect.Code {
	var ce *connect.Error
	if errors.As(err, &ce) {
		return ce.Code()
	}
	return connect.CodeUnknown
}

func TestGetChatTokenFollowsBackendSwitch(t *testing.T) {
	ctx := context.Background()
	sw := switches.NewMemory(switches.Snapshot{})
	h := NewHandler(sw, fakeMinter{})
	req := &gamev1.GetChatTokenRequest{ShowId: "show-1"}

	// Default: chat off.
	if _, err := h.GetChatToken(ctx, req); codeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("chat off: want FailedPrecondition, got %v", err)
	}

	// Operator switches chat on for this show from the studio console.
	sw.Store(switches.Snapshot{PerShow: map[string]map[switches.Key]bool{"show-1": {switches.ChatEnabled: true}}})
	res, err := h.GetChatToken(ctx, req)
	if err != nil || res.GetChatToken() == "" {
		t.Fatalf("chat on: want token, got %v / %v", res, err)
	}

	// Kill switch: global off does not override an explicit show-level on; show-level off wins.
	sw.Store(switches.Snapshot{PerShow: map[string]map[switches.Key]bool{"show-1": {switches.ChatEnabled: false}}})
	if _, err := h.GetChatToken(ctx, req); codeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("chat switched off again: want FailedPrecondition, got %v", err)
	}
}

func TestGetChatTokenValidates(t *testing.T) {
	h := NewHandler(switches.NewMemory(switches.Snapshot{}), fakeMinter{})
	if _, err := h.GetChatToken(context.Background(), &gamev1.GetChatTokenRequest{}); codeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("want InvalidArgument, got %v", err)
	}
}

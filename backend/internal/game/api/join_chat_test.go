package api

import (
	"context"
	"errors"
	"testing"

	"trivia.al/backend/internal/platform/switches"
)

type failingMinter struct{}

func (failingMinter) MintChatToken(context.Context, string) (string, string, error) {
	return "partial-token", "wss://edge", errors.New("mint failed")
}

// TestJoinChatDropsFailedMint: a mint error never leaks a token into JoinShow (Codex review of 6ef9893).
func TestJoinChatDropsFailedMint(t *testing.T) {
	ctx := context.Background()
	on := switches.NewMemory(switches.Snapshot{PerShow: map[string]map[switches.Key]bool{"s": {switches.ChatEnabled: true}}})
	cases := []struct {
		name      string
		minter    ChatTokenMinter
		wantToken string
	}{{"mint fails", failingMinter{}, ""}, {"mint works", fakeMinter{}, "tok"}, {"no minter yet", nil, ""}}
	for _, c := range cases {
		enabled, token, endpoint := New(Deps{Switches: on, Chat: c.minter}).joinChat(ctx, "s")
		if !enabled || token != c.wantToken || (token == "") != (endpoint == "") {
			t.Errorf("%s: enabled=%v token=%q endpoint=%q, want token %q", c.name, enabled, token, endpoint, c.wantToken)
		}
	}
	if enabled, token, _ := New(Deps{Switches: switches.NewMemory(switches.Snapshot{}), Chat: fakeMinter{}}).joinChat(ctx, "s"); enabled || token != "" {
		t.Errorf("chat off: enabled=%v token=%q", enabled, token)
	}
}

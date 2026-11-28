package api

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"connectrpc.com/connect/v2"

	adminv1 "trivia.al/protocol/gen/go/trivia/admin/v1"

	"trivia.al/backend/internal/platform/switches"
)

func invalid(msg string) error { return connect.NewError(connect.CodeInvalidArgument, msg) }

func validateSet(req *adminv1.SetSwitchRequest) (switches.Key, string, string, error) {
	key := switches.Key(req.GetKey())
	if !switches.Known(key) {
		return "", "", "", invalid("unknown switch key")
	}
	showID := strings.TrimSpace(req.GetShowId())
	switch req.GetScope() {
	case adminv1.SwitchScope_SWITCH_SCOPE_GLOBAL:
		if showID != "" {
			return "", "", "", invalid("show_id must be empty for global scope")
		}
	case adminv1.SwitchScope_SWITCH_SCOPE_SHOW:
		if showID == "" {
			return "", "", "", invalid("show_id is required for show scope")
		}
	default:
		return "", "", "", invalid("scope is required")
	}
	reason := strings.TrimSpace(req.GetReason())
	if reason == "" || len(reason) > MaxReasonBytes || !utf8.ValidString(reason) || strings.ContainsFunc(reason, unicode.IsControl) {
		return "", "", "", invalid("a reason (1-500 bytes, no control characters) is required")
	}
	return key, showID, reason, nil
}

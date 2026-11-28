package api

import (
	"context"
	"errors"

	"connectrpc.com/connect/v2"

	corev1 "trivia.al/protocol/gen/go/trivia/core/v1"
	"trivia.al/protocol/gen/go/trivia/core/v1/corev1connect"

	"trivia.al/backend/internal/auth/sessions"
)

// RefreshSession implements AuthServiceHandler.
func (h *Handler) RefreshSession(ctx context.Context, req *corev1.RefreshSessionRequest) (*corev1.RefreshSessionResponse, error) {
	p, err := h.proof(ctx, corev1connect.AuthServiceRefreshSessionProcedure)
	if err != nil {
		return nil, err
	}
	refresh, s, err := h.d.Sessions.Rotate(ctx, req.GetRefreshToken(), p.JKT)
	if err != nil {
		return nil, errUnauth
	}
	sess, err := h.session(s, refresh, "", false)
	if err != nil {
		return nil, err
	}
	return &corev1.RefreshSessionResponse{Session: sess}, nil
}

// SignOut implements AuthServiceHandler.
func (h *Handler) SignOut(ctx context.Context, req *corev1.SignOutRequest) (*corev1.SignOutResponse, error) {
	p, err := h.proof(ctx, corev1connect.AuthServiceSignOutProcedure)
	if err != nil {
		return nil, err
	}
	switch err := h.d.Sessions.Revoke(ctx, req.GetRefreshToken(), p.JKT); {
	case errors.Is(err, sessions.ErrBinding):
		return nil, errUnauth
	case err != nil:
		return nil, connect.NewError(connect.CodeUnavailable, "sign-out failed; try again")
	}
	return &corev1.SignOutResponse{}, nil
}

func (h *Handler) session(s sessions.Session, refresh, platform string, isNew bool) (*corev1.Session, error) {
	client := "trivia-app"
	if platform != "" {
		client = "trivia-" + platform
	}
	at, ttl, err := h.d.Tokens.Issue(s.UserID, s.DeviceID, client, s.JKT)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "token issue failed")
	}
	return &corev1.Session{AccessToken: at, AccessExpiresInSeconds: uint32(ttl.Seconds()),
		RefreshToken: refresh, UserId: s.UserID, IsNewUser: isNew}, nil
}

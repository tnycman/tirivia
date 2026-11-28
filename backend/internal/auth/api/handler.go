// Package api exposes the auth module as trivia.core.v1.AuthService.
package api

import (
	"context"
	"net/http"
	"strings"

	"connectrpc.com/connect/v2"

	"trivia.al/protocol/gen/go/trivia/core/v1/corev1connect"

	"trivia.al/backend/internal/auth/accounts"
	"trivia.al/backend/internal/auth/attest"
	"trivia.al/backend/internal/auth/dpop"
	"trivia.al/backend/internal/auth/otp"
	"trivia.al/backend/internal/auth/sessions"
	"trivia.al/backend/internal/auth/tokens"
)

// Users finds or creates the account for a verified phone.
type Users interface {
	FindOrCreate(ctx context.Context, u accounts.NewUser) (userID string, isNew bool, err error)
}

// Devices records the attested device and its DPoP key for a user.
type Devices interface {
	Register(ctx context.Context, userID string, dev otp.Device) (deviceID string, err error)
}

// Deps wires the handler; PublicBaseURL is the external origin DPoP htu must name.
type Deps struct {
	PublicBaseURL string
	Attest        *attest.Service
	OTP           *otp.Service
	DPoP          *dpop.Verifier
	Sessions      *sessions.Manager
	Tokens        *tokens.Issuer
	Users         Users
	Devices       Devices
}

// Handler implements corev1connect.AuthServiceHandler.
type Handler struct {
	corev1connect.UnimplementedAuthServiceHandler
	d Deps
}

// New returns a Handler.
func New(d Deps) *Handler {
	d.PublicBaseURL = strings.TrimRight(d.PublicBaseURL, "/")
	return &Handler{d: d}
}

var errUnauth = connect.NewError(connect.CodeUnauthenticated, "authentication failed")

func (h *Handler) proof(ctx context.Context, procedure string) (*dpop.Proof, error) {
	info, ok := connect.CallInfoForServerContext(ctx)
	if !ok {
		return nil, errUnauth
	}
	p, err := h.d.DPoP.Verify(ctx, info.RequestHeader().Get("DPoP"), http.MethodPost, h.d.PublicBaseURL+procedure, dpop.Options{})
	if err != nil {
		return nil, errUnauth
	}
	return p, nil
}

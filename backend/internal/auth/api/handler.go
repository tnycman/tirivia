// Package api exposes the auth module as tirivia.core.v1.AuthService.
package api

import (
	"context"
	"net/http"
	"strings"

	"connectrpc.com/connect/v2"

	"github.com/tirivia/protocol/gen/go/tirivia/core/v1/corev1connect"

	"github.com/tirivia/backend/internal/auth/attest"
	"github.com/tirivia/backend/internal/auth/dpop"
	"github.com/tirivia/backend/internal/auth/otp"
	"github.com/tirivia/backend/internal/auth/sessions"
	"github.com/tirivia/backend/internal/auth/tokens"
)

// Users finds or creates the account for a verified phone.
type Users interface {
	FindOrCreate(ctx context.Context, phoneHash string) (userID string, isNew bool, err error)
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

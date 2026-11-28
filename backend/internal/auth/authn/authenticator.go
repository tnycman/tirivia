// Package authn authenticates player API requests: access token plus DPoP proof bound to it (RFC 9449 §7).
package authn

import (
	"context"
	"strings"

	"trivia.al/backend/internal/auth/dpop"
	"trivia.al/backend/internal/auth/tokens"
	"trivia.al/backend/internal/contracts"
)

// Authenticator implements contracts.Authenticator.
type Authenticator struct {
	tokens *tokens.Validator
	dpop   *dpop.Verifier
}

// New returns an Authenticator.
func New(t *tokens.Validator, d *dpop.Verifier) *Authenticator {
	return &Authenticator{tokens: t, dpop: d}
}

// Authenticate requires "Authorization: DPoP <token>", a valid token and a proof with matching ath and key.
func (a *Authenticator) Authenticate(ctx context.Context, h contracts.Headers, method, url string) (contracts.Principal, error) {
	scheme, tok, ok := strings.Cut(h.Get("Authorization"), " ")
	if !ok || !strings.EqualFold(scheme, "DPoP") || tok == "" || len(h.Values("DPoP")) != 1 {
		return contracts.Principal{}, contracts.ErrUnauthenticated
	}
	c, err := a.tokens.Validate(tok)
	if err != nil || c.Subject == "" || c.Cnf.JKT == "" {
		return contracts.Principal{}, contracts.ErrUnauthenticated
	}
	if _, err := a.dpop.Verify(ctx, h.Get("DPoP"), method, url, dpop.Options{AccessToken: tok, ExpectedJKT: c.Cnf.JKT}); err != nil {
		return contracts.Principal{}, contracts.ErrUnauthenticated
	}
	return contracts.Principal{UserID: c.Subject, DeviceID: c.DeviceID}, nil
}

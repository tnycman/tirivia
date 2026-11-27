package tokens

import (
	"crypto/ed25519"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
)

// Validator verifies tokens against the published keys.
type Validator struct {
	keys             map[string]ed25519.PublicKey
	issuer, audience string
	leeway           time.Duration
	now              func() time.Time
}

// NewValidator copies keys; leeway is the allowed clock skew for exp/nbf.
func NewValidator(keys map[string]ed25519.PublicKey, issuer, audience string, leeway time.Duration) *Validator {
	cp := make(map[string]ed25519.PublicKey, len(keys))
	for k, v := range keys {
		cp[k] = v
	}
	return &Validator{keys: cp, issuer: issuer, audience: audience, leeway: leeway, now: time.Now}
}

// Validate parses and checks raw; only EdDSA is accepted.
func (v *Validator) Validate(raw string) (*Claims, error) {
	tok, err := jwt.ParseSigned(raw, []jose.SignatureAlgorithm{jose.EdDSA})
	if err != nil || len(tok.Headers) != 1 {
		return nil, ErrMalformed
	}
	h := tok.Headers[0]
	if typ, _ := h.ExtraHeaders[jose.HeaderType].(string); typ != TypeAccessToken {
		return nil, ErrType
	}
	key, ok := v.keys[h.KeyID]
	if !ok {
		return nil, ErrUnknownKey
	}
	var c Claims
	if err := tok.Claims(key, &c); err != nil {
		return nil, ErrMalformed
	}
	exp := jwt.Expected{Issuer: v.issuer, AnyAudience: jwt.Audience{v.audience}, Time: v.now()}
	if err := c.ValidateWithLeeway(exp, v.leeway); err != nil || c.Expiry == nil {
		return nil, ErrClaims
	}
	switch {
	case c.Subject == "" || c.ID == "":
		return nil, ErrMissingData
	case c.Cnf.JKT == "":
		return nil, ErrNotBound
	}
	return &c, nil
}

package dpop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-jose/go-jose/v4"
)

// Verifier checks proofs.
type Verifier struct {
	replay ReplayCache
	window time.Duration
	now    func() time.Time
}

// NewVerifier returns a Verifier accepting iat within ±window.
func NewVerifier(replay ReplayCache, window time.Duration) (*Verifier, error) {
	if replay == nil || window <= 0 {
		return nil, errors.New("dpop: replay cache and window are required")
	}
	return &Verifier{replay: replay, window: window, now: time.Now}, nil
}

// Verify runs RFC 9449 §4.3 for a request to method + requestURL.
func (v *Verifier) Verify(ctx context.Context, proof, method, requestURL string, o Options) (*Proof, error) {
	if proof == "" {
		return nil, ErrMissing
	}
	jws, err := jose.ParseSignedCompact(proof, AllowedAlgorithms)
	if err != nil || len(jws.Signatures) != 1 {
		return nil, ErrMalformed
	}
	jwk, err := proofKey(jws.Signatures[0].Protected)
	if err != nil {
		return nil, err
	}
	payload, err := jws.Verify(jwk.Key)
	if err != nil {
		return nil, ErrSignature
	}
	var c claims
	if err := json.Unmarshal(payload, &c); err != nil || c.JTI == "" || c.HTM == "" || c.HTU == "" || c.IAT == nil {
		return nil, ErrMalformed
	}
	if err := v.checkClaims(c, method, requestURL, o); err != nil {
		return nil, err
	}
	jkt, err := thumbprint(jwk)
	if err != nil {
		return nil, err
	}
	if o.ExpectedJKT != "" && jkt != o.ExpectedJKT {
		return nil, ErrBinding
	}
	seen, err := v.replay.Seen(ctx, "dpop:"+jkt+":"+c.JTI, 2*v.window)
	if err != nil {
		return nil, fmt.Errorf("dpop: replay cache: %w", err)
	}
	if seen {
		return nil, ErrReplay
	}
	return &Proof{JKT: jkt, JTI: c.JTI, IssuedAt: time.Unix(*c.IAT, 0)}, nil
}

package dpop

import (
	"crypto"
	"crypto/sha256"
	"encoding/base64"
	"time"

	"github.com/go-jose/go-jose/v4"
)

func proofKey(h jose.Header) (*jose.JSONWebKey, error) {
	if typ, _ := h.ExtraHeaders[jose.HeaderType].(string); typ != TypeDPoP {
		return nil, ErrType
	}
	if h.JSONWebKey == nil || !h.JSONWebKey.Valid() || !h.JSONWebKey.IsPublic() {
		return nil, ErrKey
	}
	return h.JSONWebKey, nil
}

func (v *Verifier) checkClaims(c claims, method, requestURL string, o Options) error {
	if c.HTM != method {
		return ErrMethod
	}
	if !sameURL(c.HTU, requestURL) {
		return ErrURL
	}
	if o.Nonce != "" && c.Nonce != o.Nonce {
		return ErrNonce
	}
	if d := v.now().Sub(time.Unix(*c.IAT, 0)); d > v.window || d < -v.window {
		return ErrTime
	}
	if o.AccessToken != "" {
		sum := sha256.Sum256([]byte(o.AccessToken))
		if c.ATH != base64.RawURLEncoding.EncodeToString(sum[:]) {
			return ErrAccessHash
		}
	}
	return nil
}

func thumbprint(jwk *jose.JSONWebKey) (string, error) {
	tp, err := jwk.Thumbprint(crypto.SHA256)
	if err != nil {
		return "", ErrKey
	}
	return base64.RawURLEncoding.EncodeToString(tp), nil
}

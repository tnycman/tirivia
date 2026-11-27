// Package dpopclient builds DPoP proofs (RFC 9449 §4.2); the reference for the mobile apps.
package dpopclient

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/google/uuid"
)

// Key is a P-256 device key.
type Key struct {
	signer jose.Signer
}

// New wraps priv; its public JWK is embedded in every proof.
func New(priv *ecdsa.PrivateKey) (*Key, error) {
	jwk := jose.JSONWebKey{Key: priv.Public(), Algorithm: string(jose.ES256)}
	opts := (&jose.SignerOptions{}).WithType("dpop+jwt").WithHeader("jwk", jwk)
	s, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.ES256, Key: priv}, opts)
	if err != nil {
		return nil, err
	}
	return &Key{signer: s}, nil
}

// ProofParams are one proof's claims.
type ProofParams struct {
	Method, URL, AccessToken, Nonce, JTI string
	IssuedAt                             time.Time
}

// Proof returns a compact proof JWT.
func (k *Key) Proof(p ProofParams) (string, error) {
	if p.IssuedAt.IsZero() {
		p.IssuedAt = time.Now()
	}
	if p.JTI == "" {
		id, err := uuid.NewV7()
		if err != nil {
			return "", err
		}
		p.JTI = id.String()
	}
	c := map[string]any{"jti": p.JTI, "htm": p.Method, "htu": p.URL, "iat": p.IssuedAt.Unix()}
	if p.AccessToken != "" {
		sum := sha256.Sum256([]byte(p.AccessToken))
		c["ath"] = base64.RawURLEncoding.EncodeToString(sum[:])
	}
	if p.Nonce != "" {
		c["nonce"] = p.Nonce
	}
	payload, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	jws, err := k.signer.Sign(payload)
	if err != nil {
		return "", err
	}
	return jws.CompactSerialize()
}

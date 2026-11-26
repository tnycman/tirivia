// Package dpopclient builds DPoP proofs the way the iOS and Android apps must (RFC 9449 §4.2).
// It is the reference for the mobile implementations and is used by tests and load tests.
package dpopclient

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/google/uuid"
)

// Key is a device key: P-256, as created in the Secure Enclave / Android Keystore.
type Key struct {
	priv   *ecdsa.PrivateKey
	signer jose.Signer
}

// New wraps a P-256 private key. The public JWK is embedded in every proof header.
func New(priv *ecdsa.PrivateKey) (*Key, error) {
	jwk := jose.JSONWebKey{Key: priv.Public(), Algorithm: string(jose.ES256)}
	opts := (&jose.SignerOptions{}).WithType("dpop+jwt").WithHeader("jwk", jwk)
	s, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.ES256, Key: priv}, opts)
	if err != nil {
		return nil, err
	}
	return &Key{priv: priv, signer: s}, nil
}

// ProofParams are the claims of one proof.
type ProofParams struct {
	Method      string
	URL         string // without query/fragment
	AccessToken string // set when calling with an access token (adds ath)
	Nonce       string // set when the server returned DPoP-Nonce
	IssuedAt    time.Time
	JTI         string // defaults to a fresh UUIDv7
}

// Proof returns a compact-serialised DPoP proof JWT.
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
	claims := map[string]any{"jti": p.JTI, "htm": p.Method, "htu": p.URL, "iat": p.IssuedAt.Unix()}
	if p.AccessToken != "" {
		sum := sha256.Sum256([]byte(p.AccessToken))
		claims["ath"] = base64.RawURLEncoding.EncodeToString(sum[:])
	}
	if p.Nonce != "" {
		claims["nonce"] = p.Nonce
	}
	payload, err := jsonMarshal(claims)
	if err != nil {
		return "", err
	}
	jws, err := k.signer.Sign(payload)
	if err != nil {
		return "", err
	}
	return jws.CompactSerialize()
}

package dpop

import (
	"context"
	"time"

	"github.com/go-jose/go-jose/v4"
)

// TypeDPoP is the required JWS typ.
const TypeDPoP = "dpop+jwt"

// AllowedAlgorithms is ES256 only: device keys are P-256 (Secure Enclave, Android Keystore).
var AllowedAlgorithms = []jose.SignatureAlgorithm{jose.ES256}

// ReplayCache records jti values; Seen must set-if-absent atomically.
type ReplayCache interface {
	Seen(ctx context.Context, key string, ttl time.Duration) (bool, error)
}

// Options are per-request expectations.
type Options struct {
	AccessToken string
	ExpectedJKT string
	Nonce       string
}

// Proof is a verified proof.
type Proof struct {
	JKT      string
	JTI      string
	IssuedAt time.Time
}

type claims struct {
	JTI   string `json:"jti"`
	HTM   string `json:"htm"`
	HTU   string `json:"htu"`
	IAT   *int64 `json:"iat"`
	ATH   string `json:"ath,omitempty"`
	Nonce string `json:"nonce,omitempty"`
}

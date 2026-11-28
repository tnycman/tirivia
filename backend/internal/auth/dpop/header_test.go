package dpop_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"

	"trivia.al/backend/internal/auth/dpop"
)

func signRaw(t *testing.T, payload []byte, opts *jose.SignerOptions, key any, alg jose.SignatureAlgorithm) string {
	t.Helper()
	s, err := jose.NewSigner(jose.SigningKey{Algorithm: alg, Key: key}, opts)
	if err != nil {
		t.Fatal(err)
	}
	jws, _ := s.Sign(payload)
	out, _ := jws.CompactSerialize()
	return out
}

func TestHeaderAttacks(t *testing.T) {
	ctx := context.Background()
	v := verifier(t)
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	priv2, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	_, edPriv, _ := ed25519.GenerateKey(rand.Reader)
	payload, _ := json.Marshal(map[string]any{"jti": "j", "htm": post, "htu": u, "iat": time.Now().Unix()})
	opts := func(typ string, jwk *jose.JSONWebKey) *jose.SignerOptions {
		o := (&jose.SignerOptions{}).WithType(jose.ContentType(typ))
		if jwk != nil {
			o = o.WithHeader("jwk", *jwk)
		}
		return o
	}
	pub := &jose.JSONWebKey{Key: priv.Public()}
	b64 := func(s string) string { return base64.RawURLEncoding.EncodeToString([]byte(s)) }

	privateInHeader := signRaw(t, payload, opts("dpop+jwt", &jose.JSONWebKey{Key: priv}), priv, jose.ES256)
	if _, err := v.Verify(ctx, privateInHeader, post, u, dpop.Options{}); !errors.Is(err, dpop.ErrMalformed) && !errors.Is(err, dpop.ErrKey) {
		t.Errorf("private key in header accepted: %v", err)
	}
	for name, tc := range map[string]struct {
		proof string
		want  error
	}{
		"typ JWT":         {signRaw(t, payload, opts("JWT", pub), priv, jose.ES256), dpop.ErrType},
		"no jwk":          {signRaw(t, payload, opts("dpop+jwt", nil), priv, jose.ES256), dpop.ErrKey},
		"signed by other": {signRaw(t, payload, opts("dpop+jwt", pub), priv2, jose.ES256), dpop.ErrSignature},
		"EdDSA":           {signRaw(t, payload, opts("dpop+jwt", &jose.JSONWebKey{Key: edPriv.Public()}), edPriv, jose.EdDSA), dpop.ErrMalformed},
		"HS256":           {signRaw(t, payload, opts("dpop+jwt", nil), []byte("0123456789abcdef0123456789abcdef"), jose.HS256), dpop.ErrMalformed},
		"alg none":        {b64(`{"alg":"none","typ":"dpop+jwt"}`) + "." + b64(string(payload)) + ".", dpop.ErrMalformed},
	} {
		if _, err := v.Verify(ctx, tc.proof, post, u, dpop.Options{}); !errors.Is(err, tc.want) {
			t.Errorf("%s: got %v want %v", name, err, tc.want)
		}
	}
}

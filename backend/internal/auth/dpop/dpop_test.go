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

	"github.com/tirivia/backend/internal/auth/dpop"
	"github.com/tirivia/backend/internal/auth/dpop/dpopclient"
)

const (
	post = "POST"
	u    = "https://api.tirivia.example/tirivia.core.v1.AuthService/RefreshSession"
)

func newKey(t *testing.T) *dpopclient.Key {
	t.Helper()
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	k, err := dpopclient.New(priv)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func proof(t *testing.T, k *dpopclient.Key, p dpopclient.ProofParams) string {
	t.Helper()
	if p.Method == "" {
		p.Method = post
	}
	if p.URL == "" {
		p.URL = u
	}
	s, err := k.Proof(p)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestValidProofAndThumbprint(t *testing.T) {
	ctx := context.Background()
	v := dpop.NewVerifier(dpop.NewMemoryReplayCache())
	k := newKey(t)
	got, err := v.Verify(ctx, proof(t, k, dpopclient.ProofParams{}), post, u, dpop.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.JKT) != 43 { // base64url(SHA-256) without padding
		t.Fatalf("thumbprint %q", got.JKT)
	}
	// Same key, new proof: same thumbprint (stable binding).
	again, err := v.Verify(ctx, proof(t, k, dpopclient.ProofParams{}), post, u+"?ignored=query#frag", dpop.Options{ExpectedJKT: got.JKT})
	if err != nil || again.JKT != got.JKT {
		t.Fatalf("stable binding: %v %v", again, err)
	}
}

func TestRFC9449Rejections(t *testing.T) {
	ctx := context.Background()
	k, other := newKey(t), newKey(t)
	now := time.Now()
	accessToken := "eyJ.access.token"

	replayed := proof(t, k, dpopclient.ProofParams{JTI: "fixed-jti"})
	v := dpop.NewVerifier(dpop.NewMemoryReplayCache())
	if _, err := v.Verify(ctx, replayed, post, u, dpop.Options{}); err != nil {
		t.Fatal(err)
	}
	otherJKT := func() string {
		p, _ := dpop.NewVerifier(dpop.NewMemoryReplayCache()).Verify(ctx, proof(t, other, dpopclient.ProofParams{}), post, u, dpop.Options{})
		return p.JKT
	}()

	cases := []struct {
		name  string
		proof string
		meth  string
		url   string
		opts  dpop.Options
		want  error
	}{
		{"missing", "", post, u, dpop.Options{}, dpop.ErrMissing},
		{"garbage", "not-a-jws", post, u, dpop.Options{}, dpop.ErrMalformed},
		{"wrong method", proof(t, k, dpopclient.ProofParams{Method: "GET"}), post, u, dpop.Options{}, dpop.ErrMethod},
		{"wrong url", proof(t, k, dpopclient.ProofParams{URL: "https://api.tirivia.example/other"}), post, u, dpop.Options{}, dpop.ErrURL},
		{"other host", proof(t, k, dpopclient.ProofParams{URL: "https://evil.example/tirivia.core.v1.AuthService/RefreshSession"}), post, u, dpop.Options{}, dpop.ErrURL},
		{"too old", proof(t, k, dpopclient.ProofParams{IssuedAt: now.Add(-2 * time.Minute)}), post, u, dpop.Options{}, dpop.ErrTime},
		{"from the future", proof(t, k, dpopclient.ProofParams{IssuedAt: now.Add(2 * time.Minute)}), post, u, dpop.Options{}, dpop.ErrTime},
		{"replayed jti", replayed, post, u, dpop.Options{}, dpop.ErrReplay},
		{"nonce required, absent", proof(t, k, dpopclient.ProofParams{}), post, u, dpop.Options{Nonce: "n-1"}, dpop.ErrNonce},
		{"nonce wrong", proof(t, k, dpopclient.ProofParams{Nonce: "n-0"}), post, u, dpop.Options{Nonce: "n-1"}, dpop.ErrNonce},
		{"ath missing", proof(t, k, dpopclient.ProofParams{}), post, u, dpop.Options{AccessToken: accessToken}, dpop.ErrAccessHash},
		{"ath for other token", proof(t, k, dpopclient.ProofParams{AccessToken: "other"}), post, u, dpop.Options{AccessToken: accessToken}, dpop.ErrAccessHash},
		{"stolen token, attacker key", proof(t, k, dpopclient.ProofParams{}), post, u, dpop.Options{ExpectedJKT: otherJKT}, dpop.ErrBinding},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := v.Verify(ctx, tc.proof, tc.meth, tc.url, tc.opts); !errors.Is(err, tc.want) {
				t.Fatalf("got %v want %v", err, tc.want)
			}
		})
	}
	// Positive: ath + nonce present and correct.
	ok := proof(t, k, dpopclient.ProofParams{AccessToken: accessToken, Nonce: "n-1"})
	if _, err := v.Verify(ctx, ok, post, u, dpop.Options{AccessToken: accessToken, Nonce: "n-1"}); err != nil {
		t.Fatalf("valid ath+nonce rejected: %v", err)
	}
}

// Hand-built proofs for header attacks the client library would never produce.
func TestHeaderAttacks(t *testing.T) {
	ctx := context.Background()
	v := dpop.NewVerifier(dpop.NewMemoryReplayCache())
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	payload, _ := json.Marshal(map[string]any{"jti": "j", "htm": post, "htu": u, "iat": time.Now().Unix()})

	sign := func(opts *jose.SignerOptions, key any, alg jose.SignatureAlgorithm) string {
		s, err := jose.NewSigner(jose.SigningKey{Algorithm: alg, Key: key}, opts)
		if err != nil {
			t.Fatal(err)
		}
		jws, _ := s.Sign(payload)
		out, _ := jws.CompactSerialize()
		return out
	}
	pubJWK := jose.JSONWebKey{Key: priv.Public()}
	privJWK := jose.JSONWebKey{Key: priv}
	wrongTyp := sign((&jose.SignerOptions{}).WithType("JWT").WithHeader("jwk", pubJWK), priv, jose.ES256)
	noJWK := sign((&jose.SignerOptions{}).WithType("dpop+jwt"), priv, jose.ES256)
	privateJWK := sign((&jose.SignerOptions{}).WithType("dpop+jwt").WithHeader("jwk", privJWK), priv, jose.ES256)
	// Header claims key A, signature made with key B.
	priv2, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	mismatched := sign((&jose.SignerOptions{}).WithType("dpop+jwt").WithHeader("jwk", pubJWK), priv2, jose.ES256)
	// Symmetric / non-allowed algorithms.
	_, edPriv, _ := ed25519.GenerateKey(rand.Reader)
	edProof := sign((&jose.SignerOptions{}).WithType("dpop+jwt").WithHeader("jwk", jose.JSONWebKey{Key: edPriv.Public()}), edPriv, jose.EdDSA)
	hmac := sign((&jose.SignerOptions{}).WithType("dpop+jwt"), []byte("0123456789abcdef0123456789abcdef"), jose.HS256)
	b64 := func(s string) string { return base64.RawURLEncoding.EncodeToString([]byte(s)) }
	algNone := b64(`{"alg":"none","typ":"dpop+jwt"}`) + "." + b64(string(payload)) + "."

	// go-jose already refuses a private key embedded in a JWS header at parse time (ErrMalformed);
	// our own IsPublic check (ErrKey) is defence in depth. Either rejection is correct.
	if _, err := v.Verify(ctx, privateJWK, post, u, dpop.Options{}); !errors.Is(err, dpop.ErrMalformed) && !errors.Is(err, dpop.ErrKey) {
		t.Errorf("private key in header accepted: %v", err)
	}
	for name, tc := range map[string]struct {
		proof string
		want  error
	}{
		"typ JWT":                {wrongTyp, dpop.ErrType},
		"no jwk header":          {noJWK, dpop.ErrKey},
		"signature by other key": {mismatched, dpop.ErrSignature},
		"EdDSA not allowed":      {edProof, dpop.ErrMalformed},
		"HS256 (symmetric)":      {hmac, dpop.ErrMalformed},
		"alg none":               {algNone, dpop.ErrMalformed},
	} {
		if _, err := v.Verify(ctx, tc.proof, post, u, dpop.Options{}); !errors.Is(err, tc.want) {
			t.Errorf("%s: got %v want %v", name, err, tc.want)
		}
	}
}

package dpop_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tirivia/backend/internal/auth/dpop"
	"github.com/tirivia/backend/internal/auth/dpop/dpopclient"
)

func TestValidProofStableThumbprint(t *testing.T) {
	ctx := context.Background()
	v, k := verifier(t), newKey(t)
	got, err := v.Verify(ctx, proof(t, k, dpopclient.ProofParams{}), post, u, dpop.Options{})
	if err != nil || len(got.JKT) != 43 {
		t.Fatalf("verify: %v %v", got, err)
	}
	again, err := v.Verify(ctx, proof(t, k, dpopclient.ProofParams{}), post, u+"?q=1#f", dpop.Options{ExpectedJKT: got.JKT})
	if err != nil || again.JKT != got.JKT {
		t.Fatalf("binding: %v %v", again, err)
	}
}

func TestNewVerifierRequiresConfig(t *testing.T) {
	if _, err := dpop.NewVerifier(nil, window); err == nil {
		t.Fatal("nil cache accepted")
	}
	if _, err := dpop.NewVerifier(dpop.NewMemoryReplayCache(), 0); err == nil {
		t.Fatal("zero window accepted")
	}
}

func TestRFC9449Rejections(t *testing.T) {
	ctx := context.Background()
	v, k := verifier(t), newKey(t)
	now := time.Now()
	at := "eyJ.access.token"
	replayed := proof(t, k, dpopclient.ProofParams{JTI: "fixed"})
	if _, err := v.Verify(ctx, replayed, post, u, dpop.Options{}); err != nil {
		t.Fatal(err)
	}
	other, _ := verifier(t).Verify(ctx, proof(t, newKey(t), dpopclient.ProofParams{}), post, u, dpop.Options{})

	for _, tc := range []struct {
		name, proof string
		opts        dpop.Options
		want        error
	}{
		{"missing", "", dpop.Options{}, dpop.ErrMissing},
		{"garbage", "x", dpop.Options{}, dpop.ErrMalformed},
		{"wrong method", proof(t, k, dpopclient.ProofParams{Method: "GET"}), dpop.Options{}, dpop.ErrMethod},
		{"wrong path", proof(t, k, dpopclient.ProofParams{URL: "https://api.tirivia.example/x"}), dpop.Options{}, dpop.ErrURL},
		{"other host", proof(t, k, dpopclient.ProofParams{URL: "https://evil.example/tirivia.core.v1.AuthService/RefreshSession"}), dpop.Options{}, dpop.ErrURL},
		{"too old", proof(t, k, dpopclient.ProofParams{IssuedAt: now.Add(-2 * window)}), dpop.Options{}, dpop.ErrTime},
		{"future", proof(t, k, dpopclient.ProofParams{IssuedAt: now.Add(2 * window)}), dpop.Options{}, dpop.ErrTime},
		{"replay", replayed, dpop.Options{}, dpop.ErrReplay},
		{"nonce absent", proof(t, k, dpopclient.ProofParams{}), dpop.Options{Nonce: "n1"}, dpop.ErrNonce},
		{"nonce wrong", proof(t, k, dpopclient.ProofParams{Nonce: "n0"}), dpop.Options{Nonce: "n1"}, dpop.ErrNonce},
		{"ath absent", proof(t, k, dpopclient.ProofParams{}), dpop.Options{AccessToken: at}, dpop.ErrAccessHash},
		{"ath other token", proof(t, k, dpopclient.ProofParams{AccessToken: "y"}), dpop.Options{AccessToken: at}, dpop.ErrAccessHash},
		{"other key", proof(t, k, dpopclient.ProofParams{}), dpop.Options{ExpectedJKT: other.JKT}, dpop.ErrBinding},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := v.Verify(ctx, tc.proof, post, u, tc.opts); !errors.Is(err, tc.want) {
				t.Fatalf("got %v want %v", err, tc.want)
			}
		})
	}
	ok := proof(t, k, dpopclient.ProofParams{AccessToken: at, Nonce: "n1"})
	if _, err := v.Verify(ctx, ok, post, u, dpop.Options{AccessToken: at, Nonce: "n1"}); err != nil {
		t.Fatalf("valid ath+nonce rejected: %v", err)
	}
}

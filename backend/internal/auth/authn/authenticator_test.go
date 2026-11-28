package authn

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"net/http"
	"testing"
	"time"

	"trivia.al/backend/internal/auth/dpop"
	"trivia.al/backend/internal/auth/dpop/dpopclient"
	"trivia.al/backend/internal/auth/tokens"
	"trivia.al/backend/internal/contracts"
)

const target = "https://api.example/trivia.game.v1.GameService/SubmitAnswer"

type fixture struct {
	authn *Authenticator
	issue func(jkt string) string
	key   func() (*dpopclient.Key, string)
}

func setup(t *testing.T) fixture {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	iss, err := tokens.NewIssuer("k1", priv, "https://api.example", "trivia-api")
	if err != nil {
		t.Fatal(err)
	}
	ver, err := dpop.NewVerifier(dpop.NewMemoryReplayCache(), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	probe, err := dpop.NewVerifier(dpop.NewMemoryReplayCache(), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	val := tokens.NewValidator(map[string]ed25519.PublicKey{"k1": pub}, "https://api.example", "trivia-api", 30*time.Second)
	return fixture{
		authn: New(val, ver),
		issue: func(jkt string) string {
			tok, _, err := iss.Issue("user-1", "device-1", "trivia-ios", jkt)
			if err != nil {
				t.Fatal(err)
			}
			return tok
		},
		key: func() (*dpopclient.Key, string) {
			p, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if err != nil {
				t.Fatal(err)
			}
			k, err := dpopclient.New(p)
			if err != nil {
				t.Fatal(err)
			}
			proof, _ := k.Proof(dpopclient.ProofParams{Method: http.MethodPost, URL: target})
			pr, err := probe.Verify(context.Background(), proof, http.MethodPost, target, dpop.Options{})
			if err != nil {
				t.Fatal(err)
			}
			return k, pr.JKT
		},
	}
}

func header(t *testing.T, scheme, tok string, k *dpopclient.Key, ath string) http.Header {
	t.Helper()
	h := http.Header{}
	h.Set("Authorization", scheme+" "+tok)
	proof, err := k.Proof(dpopclient.ProofParams{Method: http.MethodPost, URL: target, AccessToken: ath})
	if err != nil {
		t.Fatal(err)
	}
	h.Set("DPoP", proof)
	return h
}

func TestAuthenticate(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	k, jkt := f.key()
	tok := f.issue(jkt)
	good := header(t, "DPoP", tok, k, tok)
	p, err := f.authn.Authenticate(ctx, good, http.MethodPost, target)
	if err != nil || p.UserID != "user-1" || p.DeviceID != "device-1" {
		t.Fatalf("good request: %+v %v", p, err)
	}
	other, _ := f.key()
	twoProofs := header(t, "DPoP", tok, k, tok)
	twoProofs.Add("DPoP", twoProofs.Get("DPoP"))
	cases := map[string]http.Header{
		"replayed proof":       good,
		"bearer scheme":        header(t, "Bearer", tok, k, tok),
		"proof without ath":    header(t, "DPoP", tok, k, ""),
		"proof from other key": header(t, "DPoP", tok, other, tok),
		"forged token":         header(t, "DPoP", tok+"x", k, tok+"x"),
		"two proofs":           twoProofs,
		"no proof":             {"Authorization": []string{"DPoP " + tok}},
	}
	for name, h := range cases {
		if _, err := f.authn.Authenticate(ctx, h, http.MethodPost, target); !errors.Is(err, contracts.ErrUnauthenticated) {
			t.Errorf("%s: got %v", name, err)
		}
	}
	if _, err := f.authn.Authenticate(ctx, header(t, "DPoP", tok, k, tok), http.MethodGet, target); err == nil {
		t.Error("proof for another method accepted")
	}
}

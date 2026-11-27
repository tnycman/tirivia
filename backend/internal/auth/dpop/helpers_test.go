package dpop_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
	"time"

	"github.com/tirivia/backend/internal/auth/dpop"
	"github.com/tirivia/backend/internal/auth/dpop/dpopclient"
)

const (
	post   = "POST"
	u      = "https://api.tirivia.example/tirivia.core.v1.AuthService/RefreshSession"
	window = 60 * time.Second
)

func verifier(t *testing.T) *dpop.Verifier {
	t.Helper()
	v, err := dpop.NewVerifier(dpop.NewMemoryReplayCache(), window)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

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

package main

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"github.com/go-jose/go-jose/v4"

	gamev1 "github.com/tirivia/protocol/gen/go/tirivia/game/v1"
	"github.com/tirivia/protocol/gen/go/tirivia/game/v1/gamev1connect"

	"github.com/tirivia/backend/internal/auth/dpop/dpopclient"
	"github.com/tirivia/backend/internal/auth/tokens"
	gameapi "github.com/tirivia/backend/internal/game/api"
	"github.com/tirivia/backend/internal/platform/switches"
)

// TestWiredServerAuthenticates drives the production wiring with a real token and DPoP proof.
func TestWiredServerAuthenticates(t *testing.T) {
	url := os.Getenv("TEST_VALKEY_URL")
	if url == "" {
		t.Skip("TEST_VALKEY_URL not set (make store-test)")
	}
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	priv, pub := testKey(t)
	cfg, _, err := loadGameConfig(env(map[string]string{"VALKEY_URL": url, "PUBLIC_BASE_URL": ts.URL, //nolint:gosec // G101: a generated public test key
		"TOKEN_ISSUER": "https://core.test", "ACCESS_TOKEN_KEYS": "k1=" + pub, "PLAYER_STATE_TTL": "1h"}))
	if err != nil {
		t.Fatal(err)
	}
	deps, closeKV, err := playerDeps(cfg, true, switches.NewMemory(switches.Snapshot{}), slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(closeKV)
	srv := connect.NewServer()
	gamev1connect.RegisterGameServiceHandler(srv, gameapi.New(deps))
	connecthttp.Mount(mux, srv)
	client := gamev1connect.NewGameServiceClient(connect.NewClient(connecthttp.NewTransport(ts.Client(), ts.URL)))

	device, jkt := deviceKey(t)
	issuer, err := tokens.NewIssuer("k1", priv, "https://core.test", tokens.Audience)
	if err != nil {
		t.Fatal(err)
	}
	tok, _, err := issuer.Issue("user-1", "device-1", "ios", jkt)
	if err != nil {
		t.Fatal(err)
	}
	proof, err := device.Proof(dpopclient.ProofParams{Method: http.MethodPost, URL: ts.URL + gamev1connect.GameServiceGetShowStateProcedure, AccessToken: tok})
	if err != nil {
		t.Fatal(err)
	}
	req := &gamev1.GetShowStateRequest{ShowId: "no-such-show"}
	if _, err := client.GetShowState(signed(tok, proof), req); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("authenticated call: %v, want NotFound (show not live)", err)
	}
	if _, err := client.GetShowState(signed(tok, proof), req); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("replayed proof: %v, want Unauthenticated", err)
	}
	if _, err := client.GetShowState(context.Background(), req); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("anonymous call: %v, want Unauthenticated", err)
	}
}

func deviceKey(t *testing.T) (*dpopclient.Key, string) {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	k, err := dpopclient.New(priv)
	if err != nil {
		t.Fatal(err)
	}
	tp, err := (&jose.JSONWebKey{Key: &priv.PublicKey}).Thumbprint(crypto.SHA256)
	if err != nil {
		t.Fatal(err)
	}
	return k, base64.RawURLEncoding.EncodeToString(tp)
}

func signed(tok, proof string) context.Context {
	ctx, info := connect.NewClientContext(context.Background())
	info.RequestHeader().Set("Authorization", "DPoP "+tok)
	info.RequestHeader().Set("DPoP", proof)
	return ctx
}

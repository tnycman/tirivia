package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"github.com/go-jose/go-jose/v4"
	"github.com/valkey-io/valkey-go"

	"github.com/tirivia/protocol/gen/go/tirivia/game/v1/gamev1connect"

	"github.com/tirivia/backend/internal/auth/dpop/dpopclient"
	"github.com/tirivia/backend/internal/auth/tokens"
	gameapi "github.com/tirivia/backend/internal/game/api"
	"github.com/tirivia/backend/internal/platform/kv"
	"github.com/tirivia/backend/internal/platform/switches"
)

// wired is a game server built exactly as main builds it, plus a signed-in test device.
type wired struct {
	client           gamev1connect.GameServiceClient
	deps             gameapi.Deps
	kv               valkey.Client
	device           *dpopclient.Key
	url, token, user string
	suffix           string
}

func newWired(t *testing.T) wired {
	t.Helper()
	url := os.Getenv("TEST_VALKEY_URL")
	if url == "" {
		t.Skip("TEST_VALKEY_URL not set (make store-test)")
	}
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	priv, pub := testKey(t)
	cfg, enabled, err := loadGameConfig(env(map[string]string{"VALKEY_URL": url, "PUBLIC_BASE_URL": ts.URL, //nolint:gosec // G101: a generated public test key
		"TOKEN_ISSUER": "https://core.test", "ACCESS_TOKEN_KEYS": "k1=" + pub, "CUE_PUBLIC_KEYS": "1=" + pub, "PLAYER_STATE_TTL": "1h"}))
	if err != nil || !enabled {
		t.Fatalf("enabled=%v err=%v", enabled, err)
	}
	deps, closeKV, err := playerDeps(cfg, enabled, switches.NewMemory(switches.Snapshot{}), slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(closeKV)
	srv := connect.NewServer()
	gamev1connect.RegisterGameServiceHandler(srv, gameapi.New(deps))
	connecthttp.Mount(mux, srv)
	c, err := kv.Open(url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	w := wired{client: gamev1connect.NewGameServiceClient(connect.NewClient(connecthttp.NewTransport(ts.Client(), ts.URL))),
		deps: deps, kv: c, url: ts.URL, suffix: randomHex(t), user: "user-" + randomHex(t)}
	var jkt string
	w.device, jkt = deviceKey(t)
	issuer, err := tokens.NewIssuer("k1", priv, "https://core.test", tokens.Audience)
	if err != nil {
		t.Fatal(err)
	}
	if w.token, _, err = issuer.Issue(w.user, "device-1", "ios", jkt); err != nil {
		t.Fatal(err)
	}
	return w
}

func dpopProof(method, url, token string) dpopclient.ProofParams {
	return dpopclient.ProofParams{Method: method, URL: url, AccessToken: token}
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

func randomHex(t *testing.T) string {
	t.Helper()
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(b)
}

package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/tirivia/backend/internal/platform/httpserver"
)

func env(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

func testKey(t *testing.T) (ed25519.PrivateKey, string) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return priv, base64.RawURLEncoding.EncodeToString(pub)
}

func TestConfigDisabledWithoutValkey(t *testing.T) {
	_, enabled, err := loadGameConfig(env(map[string]string{"PUBLIC_BASE_URL": "https://x"}))
	if enabled || err != nil {
		t.Fatalf("enabled=%v err=%v, want disabled without error", enabled, err)
	}
}

func TestConfigPartialIsAnError(t *testing.T) {
	_, enabled, err := loadGameConfig(env(map[string]string{"VALKEY_URL": "valkey://localhost:6379"}))
	if !enabled || err == nil {
		t.Fatalf("enabled=%v err=%v, want an error", enabled, err)
	}
	for _, want := range []string{"PUBLIC_BASE_URL", "TOKEN_ISSUER", "ACCESS_TOKEN_KEYS", "PLAYER_STATE_TTL"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not name %s", err, want)
		}
	}
}

func TestConfigComplete(t *testing.T) {
	_, pub := testKey(t)
	cfg, enabled, err := loadGameConfig(env(map[string]string{"VALKEY_URL": "valkey://localhost:6379", //nolint:gosec // G101: a generated public test key
		"PUBLIC_BASE_URL": "https://api.example/", "TOKEN_ISSUER": "https://core.example",
		"ACCESS_TOKEN_KEYS": "k1=" + pub, "PLAYER_STATE_TTL": "6h"}))
	if !enabled || err != nil {
		t.Fatalf("enabled=%v err=%v", enabled, err)
	}
	if cfg.publicBaseURL != "https://api.example" || cfg.playerTTL != 6*time.Hour || len(cfg.keys["k1"]) != ed25519.PublicKeySize {
		t.Fatalf("cfg = %+v", cfg)
	}
}

func TestParseKeys(t *testing.T) {
	_, a := testKey(t)
	_, b := testKey(t)
	if keys, err := parseKeys("k1=" + a + ", k2=" + b); err != nil || len(keys) != 2 {
		t.Fatalf("two keys: %v %v", keys, err)
	}
	for name, bad := range map[string]string{"empty": "", "no kid": "=" + a, "short": "k1=AAAA",
		"not base64": "k1=***", "duplicate": "k1=" + a + ",k1=" + b} {
		if _, err := parseKeys(bad); err == nil {
			t.Errorf("%s: accepted %q", name, bad)
		}
	}
}

func TestDrainLeavesTimeForFinalFlush(t *testing.T) {
	if httpserver.DrainTimeout >= ecsStopTimeout {
		t.Fatalf("drain %v leaves no time before SIGKILL at %v", httpserver.DrainTimeout, ecsStopTimeout)
	}
}

func TestOrigin(t *testing.T) {
	for in, want := range map[string]string{"https://api.example": "https://api.example",
		"https://api.example/": "https://api.example", "http://localhost:8081": "http://localhost:8081"} {
		if got, err := origin("X", in); err != nil || got != want {
			t.Errorf("origin(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, bad := range []string{"", "api.example.com", "ftp://api.example", "https://u:p@api.example",
		"https://api.example/v1", "https://api.example?x=1", "https://api.example#f", "https://"} {
		if _, err := origin("X", bad); err == nil {
			t.Errorf("origin(%q) accepted", bad)
		}
	}
}

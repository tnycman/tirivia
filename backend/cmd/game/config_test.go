package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"
	"time"
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

func TestFinalFlushFitsInsideStopTimeout(t *testing.T) {
	if finalFlushTimeout <= 0 {
		t.Fatalf("drain leaves no time for the final flush: %v", finalFlushTimeout)
	}
}

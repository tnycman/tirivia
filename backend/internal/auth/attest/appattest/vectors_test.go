package appattest

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

type deviceVector struct {
	TeamID, BundleID string
	Attestation      struct{ Object, KeyID, ClientDataHash, Timestamp string }
	Assertion        struct {
		Object, ClientData string
		Counter            uint32
	}
}

func b64(t *testing.T, s string) []byte {
	t.Helper()
	b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(s))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func appleSample(t *testing.T) (Config, []byte, []byte, []byte, time.Time) {
	t.Helper()
	raw, err := os.ReadFile("testdata/apple_sample_attestation.b64")
	if err != nil {
		t.Fatal(err)
	}
	return Config{TeamID: "1234567890", BundleID: "com.example.myapp", Environment: Production},
		b64(t, string(raw)), b64(t, "zgSY9YSD+7TaDXssY6WlOPVS1K3Lmk+pFhlcSWE+ZV0="),
		[]byte("example_server_challenge"), time.Date(2026, 4, 21, 0, 0, 0, 0, time.UTC)
}

func device(t *testing.T) (deviceVector, Config) {
	t.Helper()
	raw, err := os.ReadFile("testdata/veehaitch_ios14_4.json")
	if err != nil {
		t.Fatal(err)
	}
	var v deviceVector
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	return v, Config{TeamID: v.TeamID, BundleID: v.BundleID, Environment: Development}
}

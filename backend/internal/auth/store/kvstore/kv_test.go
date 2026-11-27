package kvstore_test

import (
	"crypto/rand"
	"os"
	"testing"

	"github.com/valkey-io/valkey-go"

	"github.com/tirivia/backend/internal/auth/attest/attesttest"
	"github.com/tirivia/backend/internal/auth/dpop/dpoptest"
	"github.com/tirivia/backend/internal/auth/otp/otptest"
	"github.com/tirivia/backend/internal/auth/store/kvstore"
	"github.com/tirivia/backend/internal/platform/kv"
	"github.com/tirivia/backend/internal/platform/seal"
)

func client(t *testing.T) valkey.Client {
	t.Helper()
	url := os.Getenv("TEST_VALKEY_URL")
	if url == "" {
		t.Skip("TEST_VALKEY_URL not set (make store-test)")
	}
	c, err := kv.Open(url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	return c
}

func TestValkeyStores(t *testing.T) {
	c := client(t)
	k := make([]byte, seal.KeyBytes)
	if _, err := rand.Read(k); err != nil {
		t.Fatal(err)
	}
	s, err := seal.New(1, map[byte][]byte{1: k})
	if err != nil {
		t.Fatal(err)
	}
	otptest.RunStore(t, kvstore.NewChallenges(c, s))
	otptest.RunLimiter(t, kvstore.NewLimiter(c))
	otptest.RunLimiterRace(t, kvstore.NewLimiter(c))
	dpoptest.RunReplay(t, kvstore.NewReplay(c))
	attesttest.RunStore(t, kvstore.NewAttest(c))
}

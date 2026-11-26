package store_test

import (
	"testing"

	"github.com/tirivia/backend/internal/auth/attest"
	"github.com/tirivia/backend/internal/auth/attest/attesttest"
	"github.com/tirivia/backend/internal/auth/dpop"
	"github.com/tirivia/backend/internal/auth/dpop/dpoptest"
	"github.com/tirivia/backend/internal/auth/otp"
	"github.com/tirivia/backend/internal/auth/otp/otptest"
)

func TestMemStores(t *testing.T) {
	otptest.RunStore(t, otp.NewMemStore())
	otptest.RunLimiter(t, otp.NewMemLimiter())
	otptest.RunLimiterRace(t, otp.NewMemLimiter())
	dpoptest.RunReplay(t, dpop.NewMemoryReplayCache())
	attesttest.RunStore(t, attest.NewMemStore())
}

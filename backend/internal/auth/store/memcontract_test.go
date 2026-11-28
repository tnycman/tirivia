package store_test

import (
	"testing"

	"trivia.al/backend/internal/auth/attest"
	"trivia.al/backend/internal/auth/attest/attesttest"
	"trivia.al/backend/internal/auth/dpop"
	"trivia.al/backend/internal/auth/dpop/dpoptest"
	"trivia.al/backend/internal/auth/otp"
	"trivia.al/backend/internal/auth/otp/otptest"
)

func TestMemStores(t *testing.T) {
	otptest.RunStore(t, otp.NewMemStore())
	otptest.RunLimiter(t, otp.NewMemLimiter())
	otptest.RunLimiterRace(t, otp.NewMemLimiter())
	dpoptest.RunReplay(t, dpop.NewMemoryReplayCache())
	attesttest.RunStore(t, attest.NewMemStore())
}

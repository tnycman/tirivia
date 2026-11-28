package otp_test

import (
	"testing"

	"trivia.al/backend/internal/auth/otp"
	"trivia.al/backend/internal/auth/otp/otptest"
)

func TestMemFailuresContract(t *testing.T) { otptest.RunFailures(t, otp.NewMemFailures()) }

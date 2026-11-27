package otp_test

import (
	"testing"

	"github.com/tirivia/backend/internal/auth/otp"
	"github.com/tirivia/backend/internal/auth/otp/otptest"
)

func TestMemFailuresContract(t *testing.T) { otptest.RunFailures(t, otp.NewMemFailures()) }

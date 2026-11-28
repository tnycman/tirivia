package appattest_test

import (
	"testing"

	"trivia.al/backend/internal/auth/attest/appattest"
	"trivia.al/backend/internal/auth/attest/appattest/keystest"
)

func TestMemKeys(t *testing.T) { keystest.Run(t, appattest.NewMemKeys()) }

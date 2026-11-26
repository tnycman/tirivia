package appattest_test

import (
	"testing"

	"github.com/tirivia/backend/internal/auth/attest/appattest"
	"github.com/tirivia/backend/internal/auth/attest/appattest/keystest"
)

func TestMemKeys(t *testing.T) { keystest.Run(t, appattest.NewMemKeys()) }

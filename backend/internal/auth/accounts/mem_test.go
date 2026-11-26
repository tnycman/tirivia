package accounts_test

import (
	"testing"

	"github.com/tirivia/backend/internal/auth/accounts"
	"github.com/tirivia/backend/internal/auth/accounts/accountstest"
)

func TestMem(t *testing.T) {
	accountstest.Run(t, accounts.NewMem())
	accountstest.RunRace(t, accounts.NewMem())
}

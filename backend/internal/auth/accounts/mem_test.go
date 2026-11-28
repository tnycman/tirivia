package accounts_test

import (
	"testing"

	"trivia.al/backend/internal/auth/accounts"
	"trivia.al/backend/internal/auth/accounts/accountstest"
)

func TestMem(t *testing.T) {
	accountstest.Run(t, accounts.NewMem())
	accountstest.RunRace(t, accounts.NewMem())
}

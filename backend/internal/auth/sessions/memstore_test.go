package sessions_test

import (
	"testing"

	"trivia.al/backend/internal/auth/sessions"
	"trivia.al/backend/internal/auth/sessions/sessionstest"
)

func TestMemStoreContract(t *testing.T) {
	seed := func(*testing.T) (string, string, string) { return "u1", "d1", "jkt-1" }
	sessionstest.Run(t, sessions.NewMemStore(), seed)
	sessionstest.RunRace(t, sessions.NewMemStore(), seed)
	sessionstest.RunFamilyRevocationRace(t, sessions.NewMemStore(), seed)
}

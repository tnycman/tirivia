package accountstest

import (
	"context"
	"sync"
	"testing"

	"trivia.al/backend/internal/auth/accounts"
)

// RunRace checks that concurrent first sign-ins for one phone converge on one user.
func RunRace(t *testing.T, d Directory) {
	t.Helper()
	u := accounts.NewUser{PhoneHash: rnd(t), PhoneE164: "+355691234567", Country: "AL"}
	const n = 16
	ids := make([]string, n)
	created := make([]bool, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			id, isNew, err := d.FindOrCreate(context.Background(), u)
			if err != nil {
				t.Error(err)
			}
			ids[i], created[i] = id, isNew
		})
	}
	wg.Wait()
	news := 0
	for i := range n {
		if ids[i] != ids[0] {
			t.Fatalf("diverged: %s vs %s", ids[i], ids[0])
		}
		if created[i] {
			news++
		}
	}
	if news != 1 {
		t.Fatalf("%d callers saw a new user; want 1", news)
	}
}

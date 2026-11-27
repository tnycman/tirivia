package sessions

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

const jkt, attackerJKT = "device-key-thumbprint", "attacker-key-thumbprint"

func TestRotationChain(t *testing.T) {
	ctx := context.Background()
	m := NewManager(NewMemStore())
	r1, s1, err := m.Create(ctx, "user", "dev", jkt)
	if err != nil {
		t.Fatal(err)
	}
	r2, s2, err := m.Rotate(ctx, r1, jkt)
	if err != nil || r2 == r1 || s2.FamilyID != s1.FamilyID {
		t.Fatalf("rotate: %v", err)
	}
	if _, _, err := m.Rotate(ctx, r2, jkt); err != nil {
		t.Fatalf("second rotation: %v", err)
	}
}

// RFC 9700 §4.14.2: a rotated token presented again revokes the whole family, including the
// token the legitimate client currently holds.
func TestReuseRevokesFamily(t *testing.T) {
	ctx := context.Background()
	m := NewManager(NewMemStore())
	r1, _, _ := m.Create(ctx, "user", "dev", jkt)
	r2, _, _ := m.Rotate(ctx, r1, jkt) // legitimate client now holds r2
	if _, _, err := m.Rotate(ctx, r1, jkt); !errors.Is(err, ErrReuse) {
		t.Fatalf("reuse not detected: %v", err)
	}
	if _, _, err := m.Rotate(ctx, r2, jkt); !errors.Is(err, ErrRevoked) {
		t.Fatalf("family not revoked: %v", err)
	}
}

func TestStolenTokenWithOtherKey(t *testing.T) {
	ctx := context.Background()
	m := NewManager(NewMemStore())
	r1, _, _ := m.Create(ctx, "user", "dev", jkt)
	if _, _, err := m.Rotate(ctx, r1, attackerJKT); !errors.Is(err, ErrBinding) {
		t.Fatalf("other key accepted: %v", err)
	}
	if _, _, err := m.Rotate(ctx, r1, jkt); !errors.Is(err, ErrRevoked) {
		t.Fatalf("family must be revoked after a binding mismatch: %v", err)
	}
}

func TestInactivityExpiry(t *testing.T) {
	ctx := context.Background()
	m := NewManager(NewMemStore())
	start := time.Now()
	m.now = func() time.Time { return start }
	r1, _, _ := m.Create(ctx, "user", "dev", jkt)
	m.now = func() time.Time { return start.Add(InactivityTTL + time.Hour) }
	if _, _, err := m.Rotate(ctx, r1, jkt); !errors.Is(err, ErrExpired) {
		t.Fatalf("inactive session not expired: %v", err)
	}
}

func TestSignOutAndUnknown(t *testing.T) {
	ctx := context.Background()
	m := NewManager(NewMemStore())
	r1, _, _ := m.Create(ctx, "user", "dev", jkt)
	if err := m.Revoke(ctx, r1); err != nil {
		t.Fatal(err)
	}
	if _, _, err := m.Rotate(ctx, r1, jkt); !errors.Is(err, ErrRevoked) {
		t.Fatalf("signed-out session still works: %v", err)
	}
	if _, _, err := m.Rotate(ctx, "never-issued", jkt); !errors.Is(err, ErrInvalid) {
		t.Fatalf("unknown token: %v", err)
	}
}

// Two concurrent refreshes with the same token: exactly one wins, the other is treated as reuse.
func TestConcurrentRotationOneWinner(t *testing.T) {
	ctx := context.Background()
	m := NewManager(NewMemStore())
	r1, _, _ := m.Create(ctx, "user", "dev", jkt)
	var wg sync.WaitGroup
	results := make(chan error, 8)
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := m.Rotate(ctx, r1, jkt)
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	ok := 0
	for err := range results {
		if err == nil {
			ok++
		}
	}
	if ok > 1 {
		t.Fatalf("%d concurrent rotations succeeded; at most one may", ok)
	}
}

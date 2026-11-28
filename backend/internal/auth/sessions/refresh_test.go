package sessions

import (
	"context"
	"errors"
	"testing"
)

// ID-7: an old token with another key is refused before reuse detection, so the family stays live.
func TestOldTokenWithOtherKeyRefused(t *testing.T) {
	ctx := context.Background()
	m := NewManager(NewMemStore())
	r1, _, _ := m.Create(ctx, "user", "dev", jkt)
	r2, _, _ := m.Rotate(ctx, r1, jkt)
	if _, _, err := m.Rotate(ctx, r1, attackerJKT); !errors.Is(err, ErrBinding) {
		t.Fatalf("old token with other key: %v", err)
	}
	if _, _, err := m.Rotate(ctx, r2, jkt); err != nil {
		t.Fatalf("refused refresh revoked the family: %v", err)
	}
}

package sessions

import (
	"context"
	"errors"
	"testing"
)

// ID-7, RFC 7009 §2.1: sign-out with another key is refused and the family stays live.
func TestSignOutWithOtherKeyRefused(t *testing.T) {
	ctx := context.Background()
	m := NewManager(NewMemStore())
	r1, _, _ := m.Create(ctx, "user", "dev", jkt)
	if err := m.Revoke(ctx, r1, attackerJKT); !errors.Is(err, ErrBinding) {
		t.Fatalf("sign-out with other key: %v", err)
	}
	if _, _, err := m.Rotate(ctx, r1, jkt); err != nil {
		t.Fatalf("refused sign-out revoked the family: %v", err)
	}
}

// RFC 7009 §2.2: an unknown token is answered with success.
func TestSignOutUnknownTokenSucceeds(t *testing.T) {
	m := NewManager(NewMemStore())
	if err := m.Revoke(context.Background(), "never-issued", jkt); err != nil {
		t.Fatalf("unknown token: %v", err)
	}
}

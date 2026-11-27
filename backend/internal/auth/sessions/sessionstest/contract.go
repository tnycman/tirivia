// Package sessionstest is the behaviour every sessions.Store must share.
package sessionstest

import (
	"context"
	"crypto/rand"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/tirivia/backend/internal/auth/sessions"
)

// Seed returns an existing user, device and the device's DPoP thumbprint.
type Seed func(t *testing.T) (userID, deviceID, jkt string)

// Run checks s.
func Run(t *testing.T, s sessions.Store, seed Seed) {
	t.Helper()
	ctx := context.Background()
	user, device, jkt := seed(t)
	now := time.Now().UTC().Truncate(time.Microsecond)
	a := session(t, user, device, jkt, "", now)
	b := session(t, user, device, jkt, a.FamilyID, now)
	for _, x := range []sessions.Session{a, b} {
		if err := s.Insert(ctx, x); err != nil {
			t.Fatal(err)
		}
	}
	got, err := s.ByRefreshHash(ctx, a.RefreshHash)
	if err != nil || got.ID != a.ID || got.FamilyID != a.FamilyID || got.JKT != jkt || got.UserID != user || got.DeviceID != device {
		t.Fatalf("lookup: %+v %v", got, err)
	}
	if !got.LastUsedAt.Equal(now) || got.RotatedAt != nil || got.RevokedAt != nil {
		t.Fatalf("fresh session state: %+v", got)
	}
	var unknown [32]byte
	if _, err := s.ByRefreshHash(ctx, unknown); !errors.Is(err, sessions.ErrInvalid) {
		t.Fatalf("unknown hash: %v", err)
	}
	if won, err := s.MarkRotated(ctx, a.ID, now); err != nil || !won {
		t.Fatalf("first rotation must win: %v %v", won, err)
	}
	if won, err := s.MarkRotated(ctx, a.ID, now); err != nil || won {
		t.Fatalf("second rotation must lose: %v %v", won, err)
	}
	if err := s.RevokeFamily(ctx, a.FamilyID, sessions.ReasonReuse, now); err != nil {
		t.Fatal(err)
	}
	for _, x := range []sessions.Session{a, b} {
		got, err := s.ByRefreshHash(ctx, x.RefreshHash)
		if err != nil || got.RevokedAt == nil || got.RevokeReason != sessions.ReasonReuse {
			t.Fatalf("family member %s not revoked: %+v %v", x.ID, got, err)
		}
	}
	late := session(t, user, device, jkt, a.FamilyID, now)
	if err := s.Insert(ctx, late); err != nil {
		t.Fatal(err)
	}
	got, err = s.ByRefreshHash(ctx, late.RefreshHash)
	if err != nil || got.RevokedAt == nil || got.RevokeReason != sessions.ReasonReuse {
		t.Fatalf("late family member not born revoked: %+v %v", got, err)
	}
	if won, markErr := s.MarkRotated(ctx, late.ID, now); markErr != nil || won {
		t.Fatalf("revoked family member rotated: %v %v", won, markErr)
	}
}

func session(t *testing.T, user, device, jkt, family string, now time.Time) sessions.Session {
	t.Helper()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	if family == "" {
		family = id.String()
	}
	var h [32]byte
	if _, err := rand.Read(h[:]); err != nil {
		t.Fatal(err)
	}
	return sessions.Session{ID: id.String(), UserID: user, DeviceID: device, FamilyID: family, JKT: jkt,
		RefreshHash: h, CreatedAt: now, LastUsedAt: now}
}

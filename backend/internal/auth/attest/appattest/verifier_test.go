package appattest

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"trivia.al/backend/internal/auth/attest"
)

// The device vector signed clientData "wurzelpfropf" over challenge "wurzel", i.e. ClientData("wurzel", "pfropf").
func TestVerifierFlowOnDeviceVector(t *testing.T) {
	v, c := device(t)
	keys := NewMemKeys()
	ver, err := New(c, keys, []uint32{CategoryAppStore})
	if err != nil {
		t.Fatal(err)
	}
	ver.now = func() time.Time { return time.Date(2021, 1, 23, 12, 13, 34, 0, time.UTC) }
	ctx, challenge := context.Background(), []byte("wurzel")
	ev := attest.Evidence{Platform: "ios", AppAttestKeyID: v.Attestation.KeyID, AppAttestObject: b64(t, v.Attestation.Object), JKT: "pfropf"}

	res, err := ver.Verify(ctx, challenge, ev)
	if err != nil || res.PrizeEligible || res.DeviceRef != "ios:"+v.Attestation.KeyID || !slices.Contains(res.Signals, "aa.category=absent") {
		t.Fatalf("attestation (development: never prizes): %+v %v", res, err)
	}
	if _, err := ver.Verify(ctx, challenge, ev); !errors.Is(err, ErrKeyExists) {
		t.Fatalf("second attestation of one key: %v", err)
	}

	ev.AppAttestObject, ev.AppAttestAssertion = nil, b64(t, v.Assertion.Object)
	if _, err := ver.Verify(ctx, challenge, ev); err != nil {
		t.Fatalf("assertion: %v", err)
	}
	if _, err := ver.Verify(ctx, challenge, ev); !errors.Is(err, ErrCounter) {
		t.Fatalf("replayed assertion: %v", err)
	}
	other := ev
	other.JKT = "another-key"
	if _, err := ver.Verify(ctx, challenge, other); !errors.Is(err, ErrSignature) {
		t.Fatalf("assertion bound to another DPoP key: %v", err)
	}
	unknown := ev
	unknown.AppAttestKeyID = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	if _, err := ver.Verify(ctx, challenge, unknown); !errors.Is(err, ErrUnknownKey) {
		t.Fatalf("assertion for unknown key: %v", err)
	}
}

func TestCheckCategory(t *testing.T) {
	store := []uint32{CategoryAppStore}
	for _, tc := range []struct {
		e     Extensions
		prize bool
		err   error
	}{
		{Extensions{}, true, nil},
		{Extensions{HasCategory: true, Category: CategoryAppStore}, true, nil},
		{Extensions{HasCategory: true, Category: CategoryTestFlight}, false, nil},
		{Extensions{HasCategory: true, Category: 10}, false, nil},
		{Extensions{HasCategory: true, Category: 0x04000000}, false, nil},
		{Extensions{HasCategory: true, Category: 0}, false, ErrCategory},
		{Extensions{HasCategory: true, Category: 1}, false, ErrCategory},
		{Extensions{HasCategory: true, Category: 8}, false, ErrCategory},
	} {
		prize, err := CheckCategory(tc.e, store)
		if prize != tc.prize || !errors.Is(err, tc.err) {
			t.Errorf("%+v: prize=%v err=%v", tc.e, prize, err)
		}
	}
}

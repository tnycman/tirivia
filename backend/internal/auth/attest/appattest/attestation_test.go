package appattest

import (
	"errors"
	"testing"
	"time"
)

func TestAppleSampleAttestation(t *testing.T) {
	c, obj, keyID, cdh, now := appleSample(t)
	got, err := VerifyAttestation(c, obj, keyID, cdh, now)
	if err != nil {
		t.Fatal(err)
	}
	e := got.Extensions
	if got.PublicKey == nil || len(got.Receipt) == 0 || !e.HasCategory || e.Category != 1 || e.BundleVersion != "1" {
		t.Fatalf("attested: %+v", got)
	}
}

func TestDeviceAttestation(t *testing.T) {
	v, c := device(t)
	now, err := time.Parse(time.RFC3339, v.Attestation.Timestamp)
	if err != nil {
		t.Fatal(err)
	}
	got, err := VerifyAttestation(c, b64(t, v.Attestation.Object), b64(t, v.Attestation.KeyID), b64(t, v.Attestation.ClientDataHash), now)
	if err != nil || got.Extensions.HasCategory {
		t.Fatalf("iOS 14.4 device: %+v %v", got, err)
	}
}

func TestAttestationRejects(t *testing.T) {
	c, obj, keyID, cdh, now := appleSample(t)
	tampered := append([]byte{}, obj...)
	tampered[len(tampered)-40] ^= 1
	otherKey := append([]byte{}, keyID...)
	otherKey[0] ^= 1
	dev := c
	dev.Environment = Development
	otherApp := c
	otherApp.BundleID = "com.example.other"
	cases := []struct {
		name string
		c    Config
		obj  []byte
		key  []byte
		cdh  []byte
		now  time.Time
		want error
	}{
		{"other challenge", c, obj, keyID, []byte("another_challenge"), now, ErrNonce},
		{"other key id", c, obj, otherKey, cdh, now, ErrKeyID},
		{"other app", otherApp, obj, keyID, cdh, now, ErrRPID},
		{"development server", dev, obj, keyID, cdh, now, ErrAAGUID},
		{"expired leaf", c, obj, keyID, cdh, now.AddDate(0, 0, 7), ErrChain},
		{"not CBOR", c, []byte("junk"), keyID, cdh, now, ErrFormat},
	}
	for _, tc := range cases {
		if _, err := VerifyAttestation(tc.c, tc.obj, tc.key, tc.cdh, tc.now); !errors.Is(err, tc.want) {
			t.Errorf("%s: got %v want %v", tc.name, err, tc.want)
		}
	}
	if _, err := VerifyAttestation(c, tampered, keyID, cdh, now); err == nil {
		t.Error("tampered authData accepted")
	}
}

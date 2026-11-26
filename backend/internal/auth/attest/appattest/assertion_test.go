package appattest

import (
	"errors"
	"testing"
	"time"
)

func TestDeviceAssertion(t *testing.T) {
	v, c := device(t)
	now, _ := time.Parse(time.RFC3339, v.Attestation.Timestamp)
	att, err := VerifyAttestation(c, b64(t, v.Attestation.Object), b64(t, v.Attestation.KeyID), b64(t, v.Attestation.ClientDataHash), now)
	if err != nil {
		t.Fatal(err)
	}
	obj, clientData := b64(t, v.Assertion.Object), b64(t, v.Assertion.ClientData)
	counter, _, err := VerifyAssertion(c, obj, clientData, att.PublicKey, 0)
	if err != nil || counter != v.Assertion.Counter {
		t.Fatalf("assertion: counter=%d err=%v", counter, err)
	}
	if _, _, err := VerifyAssertion(c, obj, clientData, att.PublicKey, counter); !errors.Is(err, ErrCounter) {
		t.Fatalf("replayed assertion accepted: %v", err)
	}
	if _, _, err := VerifyAssertion(c, obj, []byte("other client data"), att.PublicKey, 0); !errors.Is(err, ErrSignature) {
		t.Fatalf("other client data accepted: %v", err)
	}
	other := c
	other.TeamID = "AAAAAAAAAA"
	if _, _, err := VerifyAssertion(other, obj, clientData, att.PublicKey, 0); !errors.Is(err, ErrRPID) {
		t.Fatalf("other app accepted: %v", err)
	}
}

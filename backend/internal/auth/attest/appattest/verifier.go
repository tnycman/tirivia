package appattest

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"

	"github.com/tirivia/backend/internal/auth/attest"
)

// Verifier implements attest.Verifier for iOS: attestation on first use, assertions after (Apple guidance).
type Verifier struct {
	cfg    Config
	keys   Keys
	prizes []uint32
	now    func() time.Time
}

// New returns a Verifier; prizeCategories is {CategoryAppStore} in production.
func New(cfg Config, keys Keys, prizeCategories []uint32) (*Verifier, error) {
	if cfg.TeamID == "" || cfg.BundleID == "" || keys == nil {
		return nil, errors.New("appattest: team id, bundle id and key store are required")
	}
	return &Verifier{cfg: cfg, keys: keys, prizes: prizeCategories, now: time.Now}, nil
}

// ClientData binds the evidence to the one-time challenge and the DPoP key; the app signs exactly these bytes.
func ClientData(challenge []byte, jkt string) []byte {
	return append(append([]byte{}, challenge...), jkt...)
}

// Verify implements attest.Verifier.
func (v *Verifier) Verify(ctx context.Context, challenge []byte, ev attest.Evidence) (attest.Result, error) {
	keyID, err := base64.StdEncoding.DecodeString(ev.AppAttestKeyID)
	if err != nil || len(keyID) != sha256.Size || ev.JKT == "" {
		return attest.Result{}, attest.ErrFailed
	}
	cd := ClientData(challenge, ev.JKT)
	ref := "ios:" + ev.AppAttestKeyID
	switch {
	case len(ev.AppAttestObject) > 0:
		return v.attestKey(ctx, keyID, ev.AppAttestObject, cd, ref)
	case len(ev.AppAttestAssertion) > 0:
		return v.assert(ctx, keyID, ev.AppAttestAssertion, cd, ref)
	}
	return attest.Result{}, attest.ErrFailed
}

func (v *Verifier) attestKey(ctx context.Context, keyID, obj, clientData []byte, ref string) (attest.Result, error) {
	cdh := sha256.Sum256(clientData)
	a, err := VerifyAttestation(v.cfg, obj, keyID, cdh[:], v.now())
	if err != nil {
		return attest.Result{}, err
	}
	prize, err := CheckCategory(a.Extensions, v.prizes)
	if err != nil {
		return attest.Result{}, err
	}
	prize = prize && v.cfg.Environment == Production
	point, err := a.PublicKey.Bytes()
	if err != nil {
		return attest.Result{}, err
	}
	if err := v.keys.Insert(ctx, Key{ID: keyID, PublicKey: point, PrizeEligible: prize, Receipt: a.Receipt}); err != nil {
		return attest.Result{}, err
	}
	return attest.Result{DeviceRef: ref, PrizeEligible: prize, Signals: a.Extensions.Signals()}, nil
}

func (v *Verifier) assert(ctx context.Context, keyID, assertion, clientData []byte, ref string) (attest.Result, error) {
	k, ok, err := v.keys.Get(ctx, keyID)
	if err != nil {
		return attest.Result{}, err
	}
	if !ok {
		return attest.Result{}, ErrUnknownKey
	}
	pub, err := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), k.PublicKey)
	if err != nil {
		return attest.Result{}, err
	}
	counter, ext, err := VerifyAssertion(v.cfg, assertion, clientData, pub, k.Counter)
	if err != nil {
		return attest.Result{}, err
	}
	prize, err := CheckCategory(ext, v.prizes)
	if err != nil {
		return attest.Result{}, err
	}
	won, err := v.keys.AdvanceCounter(ctx, keyID, counter)
	if err != nil {
		return attest.Result{}, err
	}
	if !won {
		return attest.Result{}, ErrCounter
	}
	return attest.Result{DeviceRef: ref, PrizeEligible: k.PrizeEligible && prize, Signals: ext.Signals()}, nil
}

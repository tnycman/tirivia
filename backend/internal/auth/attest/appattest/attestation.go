package appattest

import (
	"bytes"
	"crypto/ecdsa"
	"time"
)

// Attested is a verified attestation: the key to store, its receipt and the extension values.
type Attested struct {
	PublicKey  *ecdsa.PublicKey
	Receipt    []byte
	Extensions Extensions
}

// VerifyAttestation runs Apple's attestation steps 1-9 and reads 10-11 (CheckCategory applies the policy).
// clientDataHash is what the app passed to attestKey.
func VerifyAttestation(c Config, obj, keyID, clientDataHash []byte, now time.Time) (Attested, error) {
	var o attestationObject
	if err := decoder.Unmarshal(obj, &o); err != nil || o.Fmt != "apple-appattest" {
		return Attested{}, ErrFormat
	}
	cred, err := verifyChain(o.AttStmt.X5C, now)
	if err != nil {
		return Attested{}, err
	}
	if err := verifyNonce(cred, o.AuthData, clientDataHash); err != nil {
		return Attested{}, err
	}
	pub, err := verifyKeyID(cred, keyID)
	if err != nil {
		return Attested{}, err
	}
	a, err := parseAuthData(o.AuthData, true)
	if err != nil {
		return Attested{}, err
	}
	if err := verifyRPID(a.rpIDHash, c); err != nil {
		return Attested{}, err
	}
	switch {
	case a.counter != 0:
		return Attested{}, ErrCounter
	case !bytes.Equal(a.aaguid, aaguids[c.Environment]):
		return Attested{}, ErrAAGUID
	case !bytes.Equal(a.credID, keyID):
		return Attested{}, ErrKeyID
	}
	ext := readExtensions(a.extensions, "apple_validation_category_01", "apple_bundle_version_01")
	return Attested{PublicKey: pub, Receipt: o.AttStmt.Receipt, Extensions: ext}, nil
}

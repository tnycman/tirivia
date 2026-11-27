package appattest

import (
	"crypto/ecdsa"
	"crypto/sha256"
)

// VerifyAssertion runs Apple's assertion steps 1-5 and reads 7-8; the caller stores the counter atomically.
// The key signs nonce with ECDSA-SHA256, so the verified digest is SHA256(nonce) (SHA256withECDSA in the
// open-source reference verifier; confirmed by its real-device vector in testdata).
func VerifyAssertion(c Config, assertion, clientData []byte, pub *ecdsa.PublicKey, prevCounter uint32) (uint32, Extensions, error) {
	var o assertionObject
	if err := decoder.Unmarshal(assertion, &o); err != nil {
		return 0, Extensions{}, ErrFormat
	}
	clientDataHash := sha256.Sum256(clientData)
	nonce := sha256.Sum256(append(append([]byte{}, o.AuthenticatorData...), clientDataHash[:]...))
	digest := sha256.Sum256(nonce[:])
	if !ecdsa.VerifyASN1(pub, digest[:], o.Signature) {
		return 0, Extensions{}, ErrSignature
	}
	a, err := parseAuthData(o.AuthenticatorData, false)
	if err != nil {
		return 0, Extensions{}, err
	}
	if err := verifyRPID(a.rpIDHash, c); err != nil {
		return 0, Extensions{}, err
	}
	if a.counter <= prevCounter {
		return 0, Extensions{}, ErrCounter
	}
	return a.counter, readExtensions(a.extensions, "validationCategory", "bundleVersion"), nil
}

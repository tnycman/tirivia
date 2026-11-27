package appattest

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"time"
)

var nonceOID = asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 8, 2}

var aaguids = map[Environment][]byte{
	Development: []byte("appattestdevelop"),
	Production:  append([]byte("appattest"), 0, 0, 0, 0, 0, 0, 0),
}

// verifyChain is step 1; Apple leaves carry a private EKU, so any EKU is accepted under the pinned root.
func verifyChain(x5c [][]byte, now time.Time) (*x509.Certificate, error) {
	if len(x5c) < 2 {
		return nil, ErrChain
	}
	roots, err := appleRoot()
	if err != nil {
		return nil, err
	}
	certs := make([]*x509.Certificate, len(x5c))
	for i, der := range x5c {
		if certs[i], err = x509.ParseCertificate(der); err != nil {
			return nil, ErrChain
		}
	}
	inter := x509.NewCertPool()
	for _, c := range certs[1:] {
		inter.AddCert(c)
	}
	_, err = certs[0].Verify(x509.VerifyOptions{Roots: roots, Intermediates: inter, CurrentTime: now,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny}})
	if err != nil {
		return nil, ErrChain
	}
	return certs[0], nil
}

// verifyNonce is steps 2-4: the credCert extension holds SHA256(authData ‖ clientDataHash).
func verifyNonce(cred *x509.Certificate, rawAuthData, clientDataHash []byte) error {
	want := sha256.Sum256(append(append([]byte{}, rawAuthData...), clientDataHash...))
	for _, ext := range cred.Extensions {
		if !ext.Id.Equal(nonceOID) {
			continue
		}
		var seq struct {
			Nonce []byte `asn1:"tag:1,explicit"`
		}
		if rest, err := asn1.Unmarshal(ext.Value, &seq); err != nil || len(rest) != 0 {
			return ErrNonce
		}
		if bytes.Equal(seq.Nonce, want[:]) {
			return nil
		}
		return ErrNonce
	}
	return ErrNonce
}

// verifyKeyID is step 5: keyID = SHA256(uncompressed X9.62 point of the credCert key).
func verifyKeyID(cred *x509.Certificate, keyID []byte) (*ecdsa.PublicKey, error) {
	pub, ok := cred.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, ErrKeyID
	}
	point, err := pub.Bytes()
	if err != nil {
		return nil, ErrKeyID
	}
	sum := sha256.Sum256(point)
	if !bytes.Equal(sum[:], keyID) {
		return nil, ErrKeyID
	}
	return pub, nil
}

func verifyRPID(rpIDHash []byte, c Config) error {
	want := sha256.Sum256([]byte(c.appID()))
	if !bytes.Equal(rpIDHash, want[:]) {
		return ErrRPID
	}
	return nil
}

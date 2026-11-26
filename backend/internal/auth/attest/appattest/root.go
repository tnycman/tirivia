package appattest

import (
	"crypto/sha256"
	"crypto/x509"
	_ "embed"
	"encoding/hex"
	"encoding/pem"
	"errors"
)

// rootPEM is https://www.apple.com/certificateauthority/Apple_App_Attestation_Root_CA.pem (fetched 2026-11-25).
//
//go:embed apple_app_attestation_root_ca.pem
var rootPEM []byte

// rootSHA256 pins the DER of the root above (CN=Apple App Attestation Root CA, valid 2020-03-18 to 2045-03-15).
const rootSHA256 = "1cb9823ba28ba6ad2d33a006941de2ae4f513ef1d4e831b9f7e0fa7b6242c932"

func appleRoot() (*x509.CertPool, error) {
	block, _ := pem.Decode(rootPEM)
	if block == nil {
		return nil, errors.New("appattest: root PEM missing")
	}
	if sum := sha256.Sum256(block.Bytes); hex.EncodeToString(sum[:]) != rootSHA256 {
		return nil, errors.New("appattest: root certificate does not match the pinned fingerprint")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	pool.AddCert(cert)
	return pool, nil
}

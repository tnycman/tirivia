package otp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

func newCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", CodeDigits, n.Int64()), nil
}

func (s *Service) mac(parts ...string) []byte {
	m := hmac.New(sha256.New, s.cfg.Pepper)
	for _, p := range parts {
		m.Write([]byte(p))
		m.Write([]byte{0})
	}
	return m.Sum(nil)
}

func (s *Service) codeMAC(requestID, code string) []byte {
	return s.mac("otp", requestID, code)
}

// PhoneHash is the keyed hash used to look up users by phone.
func (s *Service) PhoneHash(e164 string) string {
	return hex.EncodeToString(s.mac("phone", e164))
}

package sessions

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

func newToken() (string, [32]byte, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", [32]byte{}, err
	}
	tok := base64.RawURLEncoding.EncodeToString(raw[:])
	return tok, hashToken(tok), nil
}

func hashToken(tok string) [32]byte { return sha256.Sum256([]byte(tok)) }

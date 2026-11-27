package playintegrity

import (
	"crypto/sha256"
	"encoding/base64"
)

// RequestHash binds a token to one attestation request: base64url(SHA-256(challenge ‖ jkt)), 43 chars (limit 500).
// The challenge has a fixed length, so the concatenation is unambiguous; no plaintext identifiers are sent.
func RequestHash(challenge []byte, jkt string) string {
	h := sha256.New()
	h.Write(challenge)
	h.Write([]byte(jkt))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

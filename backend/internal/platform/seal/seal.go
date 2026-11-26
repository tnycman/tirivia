// Package seal encrypts small fields with AES-256-GCM and random nonces (NIST SP 800-38D).
package seal

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

// KeyBytes selects AES-256.
const KeyBytes = 32

// Errors.
var (
	ErrKey    = errors.New("seal: keys must be 32 bytes and include the current key id")
	ErrSealed = errors.New("seal: unknown key id or tampered ciphertext")
)

// Sealer prefixes each ciphertext with its key id so keys can rotate.
type Sealer struct {
	current byte
	aeads   map[byte]cipher.AEAD
}

// New returns a Sealer that encrypts with keys[current] and decrypts with any key.
func New(current byte, keys map[byte][]byte) (*Sealer, error) {
	if _, ok := keys[current]; !ok {
		return nil, ErrKey
	}
	s := &Sealer{current: current, aeads: make(map[byte]cipher.AEAD, len(keys))}
	for id, k := range keys {
		if len(k) != KeyBytes {
			return nil, ErrKey
		}
		b, err := aes.NewCipher(k)
		if err != nil {
			return nil, err
		}
		a, err := cipher.NewGCMWithRandomNonce(b)
		if err != nil {
			return nil, err
		}
		s.aeads[id] = a
	}
	return s, nil
}

// Seal encrypts plaintext bound to aad.
func (s *Sealer) Seal(plaintext, aad []byte) []byte {
	return s.aeads[s.current].Seal([]byte{s.current}, nil, plaintext, aad) //nolint:gosec // G407: NewGCMWithRandomNonce generates the nonce; NonceSize is 0
}

// Open decrypts a Seal output with the same aad.
func (s *Sealer) Open(sealed, aad []byte) ([]byte, error) {
	if len(sealed) < 1 {
		return nil, ErrSealed
	}
	a, ok := s.aeads[sealed[0]]
	if !ok {
		return nil, ErrSealed
	}
	out, err := a.Open(nil, nil, sealed[1:], aad)
	if err != nil {
		return nil, ErrSealed
	}
	return out, nil
}

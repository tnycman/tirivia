package cuesign

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"

	"google.golang.org/protobuf/proto"

	cuev1 "github.com/tirivia/protocol/gen/go/tirivia/cue/v1"
)

// Signer produces signed text cues; safe for concurrent use.
type Signer struct {
	keyID uint32
	priv  ed25519.PrivateKey
}

// NewSigner returns a Signer for keyID.
func NewSigner(keyID uint32, priv ed25519.PrivateKey) (*Signer, error) {
	if len(priv) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("cuesign: private key must be %d bytes", ed25519.PrivateKeySize)
	}
	return &Signer{keyID: keyID, priv: priv}, nil
}

// Sign serializes body, signs it and returns the text cue.
func (s *Signer) Sign(body *cuev1.CueBody) (string, error) {
	raw, err := proto.MarshalOptions{Deterministic: true}.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("cuesign: marshal body: %w", err)
	}
	env, err := proto.Marshal(&cuev1.SignedCue{Body: raw, Signature: ed25519.Sign(s.priv, raw), KeyId: s.keyID})
	if err != nil {
		return "", fmt.Errorf("cuesign: marshal envelope: %w", err)
	}
	text := Prefix + base64.RawURLEncoding.EncodeToString(env)
	if len(text) > MaxTextBytes {
		return "", fmt.Errorf("%w: %d > %d bytes", ErrTooLarge, len(text), MaxTextBytes)
	}
	return text, nil
}

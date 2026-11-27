package cuesign

import (
	"crypto/ed25519"
	"encoding/base64"
	"strings"

	"google.golang.org/protobuf/proto"

	cuev1 "github.com/tirivia/protocol/gen/go/tirivia/cue/v1"
)

// Verifier checks cues against accepted public keys (current + next).
type Verifier struct {
	keys map[uint32]ed25519.PublicKey
}

// NewVerifier copies keys.
func NewVerifier(keys map[uint32]ed25519.PublicKey) *Verifier {
	cp := make(map[uint32]ed25519.PublicKey, len(keys))
	for id, k := range keys {
		cp[id] = k
	}
	return &Verifier{keys: cp}
}

// Verify decodes a text cue, checks its signature over the received body bytes, and returns the body.
func (v *Verifier) Verify(text string) (*cuev1.CueBody, error) {
	if len(text) > MaxTextBytes {
		return nil, ErrTooLarge
	}
	encoded, ok := strings.CutPrefix(text, Prefix)
	if !ok {
		return nil, ErrMalformed
	}
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, ErrMalformed
	}
	var env cuev1.SignedCue
	if err := proto.Unmarshal(raw, &env); err != nil {
		return nil, ErrMalformed
	}
	pub, ok := v.keys[env.GetKeyId()]
	if !ok || len(pub) != ed25519.PublicKeySize {
		return nil, ErrUnknownKey
	}
	if !ed25519.Verify(pub, env.GetBody(), env.GetSignature()) {
		return nil, ErrBadSignature
	}
	var body cuev1.CueBody
	if err := proto.Unmarshal(env.GetBody(), &body); err != nil {
		return nil, ErrMalformed
	}
	return &body, nil
}

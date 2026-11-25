// Package cuesign signs and verifies the timed-metadata cues carried in the IVS stream.
//
// Wire format (SPEC §6.4): "t1." + base64url_nopad(SignedCue), where SignedCue.body is the
// serialized CueBody and SignedCue.signature is Ed25519 over exactly those bytes. Verifiers
// never re-serialize before checking the signature.
//
// Hard limits (docs/research/01-live-video-ivs.md): IVS PutMetadata accepts at most 1 KB
// (not adjustable). Our budget is MaxTextBytes so the signature and framing always fit.
package cuesign

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"google.golang.org/protobuf/proto"

	cuev1 "github.com/tirivia/protocol/gen/go/tirivia/cue/v1"
)

const (
	// Prefix versions the text framing.
	Prefix = "t1."
	// MaxTextBytes is the budget for the full text cue (IVS limit: 1 KB).
	MaxTextBytes = 900
	// MaxQuestionTextBytes bounds question text in UTF-8 bytes so a worst-case question_open
	// cue stays under MaxTextBytes (proven in tests: 845 B). 240 bytes is roughly 140-200
	// Albanian characters (only ë/ç are 2-byte).
	MaxQuestionTextBytes = 240
	// MaxOptionTextBytes bounds each option's text in UTF-8 bytes.
	MaxOptionTextBytes = 64
	// OptionsPerQuestion is fixed by the show format (Master Plan §1).
	OptionsPerQuestion = 3
)

// Errors returned by Sign, Verify and ValidateQuestionContent.
var (
	// ErrTooLarge means the encoded cue exceeds MaxTextBytes.
	ErrTooLarge = errors.New("cuesign: cue exceeds size budget")
	// ErrMalformed means the cue is not "t1."+base64url(SignedCue).
	ErrMalformed = errors.New("cuesign: malformed cue")
	// ErrUnknownKey means the key id is not in the accepted set.
	ErrUnknownKey = errors.New("cuesign: unknown signing key")
	// ErrBadSignature means the Ed25519 signature does not verify (forged or tampered cue).
	ErrBadSignature = errors.New("cuesign: signature verification failed")
	// ErrContent means question content would not fit the cue budget.
	ErrContent = errors.New("cuesign: question content out of bounds")
)

// Signer produces signed text cues. It is safe for concurrent use.
type Signer struct {
	keyID uint32
	priv  ed25519.PrivateKey
}

// NewSigner returns a Signer for the given key. The private key comes from KMS/Secrets Manager.
func NewSigner(keyID uint32, priv ed25519.PrivateKey) (*Signer, error) {
	if len(priv) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("cuesign: private key must be %d bytes", ed25519.PrivateKeySize)
	}
	return &Signer{keyID: keyID, priv: priv}, nil
}

// Sign serializes body, signs it and returns the text cue for PutMetadata.
func (s *Signer) Sign(body *cuev1.CueBody) (string, error) {
	raw, err := proto.MarshalOptions{Deterministic: true}.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("cuesign: marshal body: %w", err)
	}
	envelope, err := proto.Marshal(&cuev1.SignedCue{
		Body:      raw,
		Signature: ed25519.Sign(s.priv, raw),
		KeyId:     s.keyID,
	})
	if err != nil {
		return "", fmt.Errorf("cuesign: marshal envelope: %w", err)
	}
	text := Prefix + base64.RawURLEncoding.EncodeToString(envelope)
	if len(text) > MaxTextBytes {
		return "", fmt.Errorf("%w: %d > %d bytes", ErrTooLarge, len(text), MaxTextBytes)
	}
	return text, nil
}

// Verifier checks text cues against the set of accepted public keys (current + next).
type Verifier struct {
	keys map[uint32]ed25519.PublicKey
}

// NewVerifier copies keys so callers cannot mutate the accepted set.
func NewVerifier(keys map[uint32]ed25519.PublicKey) *Verifier {
	cp := make(map[uint32]ed25519.PublicKey, len(keys))
	for id, k := range keys {
		cp[id] = k
	}
	return &Verifier{keys: cp}
}

// Verify decodes a text cue, checks its signature and returns the body.
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
	var envelope cuev1.SignedCue
	if err := proto.Unmarshal(raw, &envelope); err != nil {
		return nil, ErrMalformed
	}
	pub, ok := v.keys[envelope.GetKeyId()]
	if !ok || len(pub) != ed25519.PublicKeySize {
		return nil, ErrUnknownKey
	}
	if !ed25519.Verify(pub, envelope.GetBody(), envelope.GetSignature()) {
		return nil, ErrBadSignature
	}
	var body cuev1.CueBody
	if err := proto.Unmarshal(envelope.GetBody(), &body); err != nil {
		return nil, ErrMalformed
	}
	return &body, nil
}

// ValidateQuestionContent enforces the content budget at question-bank save time, long
// before air, so a question can never fail to fit in a cue during a live show.
func ValidateQuestionContent(text string, options []string) error {
	if !utf8.ValidString(text) || len(text) == 0 || len(text) > MaxQuestionTextBytes {
		return fmt.Errorf("%w: question text must be 1-%d UTF-8 bytes", ErrContent, MaxQuestionTextBytes)
	}
	if len(options) != OptionsPerQuestion {
		return fmt.Errorf("%w: exactly %d options required", ErrContent, OptionsPerQuestion)
	}
	for i, o := range options {
		if !utf8.ValidString(o) || len(o) == 0 || len(o) > MaxOptionTextBytes {
			return fmt.Errorf("%w: option %d must be 1-%d UTF-8 bytes", ErrContent, i, MaxOptionTextBytes)
		}
	}
	return nil
}

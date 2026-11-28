package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	gamev1 "trivia.al/protocol/gen/go/trivia/game/v1"
)

// parseKeys reads "kid=base64url(Ed25519 public key)" pairs separated by commas; several allow kid overlap on rotation.
func parseKeys(name, s string) (map[string]ed25519.PublicKey, error) {
	keys := map[string]ed25519.PublicKey{}
	for pair := range strings.SplitSeq(s, ",") {
		kid, enc, ok := strings.Cut(strings.TrimSpace(pair), "=")
		if !ok || kid == "" {
			return nil, errors.New(name + ": want kid=key pairs")
		}
		raw, err := base64.RawURLEncoding.DecodeString(enc)
		if err != nil || len(raw) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("%s: key %q is not a base64url Ed25519 public key", name, kid)
		}
		if _, dup := keys[kid]; dup {
			return nil, fmt.Errorf("%s: duplicate kid %q", name, kid)
		}
		keys[kid] = ed25519.PublicKey(raw)
	}
	return keys, nil
}

// cueKeys parses CUE_PUBLIC_KEYS: numeric key ids, as carried in signed cues.
func cueKeys(s string) ([]*gamev1.CueKey, error) {
	keys, err := parseKeys("CUE_PUBLIC_KEYS", s)
	if err != nil {
		return nil, err
	}
	out := make([]*gamev1.CueKey, 0, len(keys))
	for kid, pub := range keys {
		id, err := strconv.ParseUint(kid, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("CUE_PUBLIC_KEYS: key id %q is not a uint32", kid)
		}
		out = append(out, &gamev1.CueKey{KeyId: uint32(id), Ed25519PublicKey: pub})
	}
	return out, nil
}

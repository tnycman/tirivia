package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

// parseKeys reads "kid=base64url(Ed25519 public key)" pairs separated by commas; several allow kid overlap on rotation.
func parseKeys(s string) (map[string]ed25519.PublicKey, error) {
	keys := map[string]ed25519.PublicKey{}
	for pair := range strings.SplitSeq(s, ",") {
		kid, enc, ok := strings.Cut(strings.TrimSpace(pair), "=")
		if !ok || kid == "" {
			return nil, errors.New("ACCESS_TOKEN_KEYS: want kid=key pairs")
		}
		raw, err := base64.RawURLEncoding.DecodeString(enc)
		if err != nil || len(raw) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("ACCESS_TOKEN_KEYS: key %q is not a base64url Ed25519 public key", kid)
		}
		if _, dup := keys[kid]; dup {
			return nil, fmt.Errorf("ACCESS_TOKEN_KEYS: duplicate kid %q", kid)
		}
		keys[kid] = ed25519.PublicKey(raw)
	}
	return keys, nil
}

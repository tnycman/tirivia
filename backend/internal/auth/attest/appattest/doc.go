// Package appattest validates App Attest attestations and assertions (developer.apple.com/documentation/devicecheck).
package appattest

import "errors"

// Errors; each names the Apple validation step that failed.
var (
	ErrFormat     = errors.New("appattest: malformed object")
	ErrChain      = errors.New("appattest: certificate chain invalid (attestation step 1)")
	ErrNonce      = errors.New("appattest: nonce mismatch (attestation steps 2-4)")
	ErrKeyID      = errors.New("appattest: key identifier mismatch (attestation steps 5, 9)")
	ErrRPID       = errors.New("appattest: RP ID hash mismatch")
	ErrCounter    = errors.New("appattest: counter invalid")
	ErrAAGUID     = errors.New("appattest: wrong environment AAGUID (attestation step 8)")
	ErrCategory   = errors.New("appattest: validation category not acceptable (attestation step 10)")
	ErrSignature  = errors.New("appattest: assertion signature invalid (assertion step 3)")
	ErrUnknownKey = errors.New("appattest: key was never attested")
)

// Environment selects the expected AAGUID (attestation step 8).
type Environment int

// Environments.
const (
	Development Environment = iota
	Production
)

// Config identifies the app; AppID is "TEAMID.bundleID".
type Config struct {
	TeamID      string
	BundleID    string
	Environment Environment
}

func (c Config) appID() string { return c.TeamID + "." + c.BundleID }

// Package attest issues challenges, verifies platform attestation and issues OTP tickets (SPEC §7).
package attest

import (
	"context"
	"errors"
	"time"
)

// ChallengeBytes exceeds Apple's 16-byte minimum for App Attest challenges.
const ChallengeBytes = 32

// Errors.
var (
	ErrChallenge = errors.New("attest: unknown, used or expired challenge")
	ErrFailed    = errors.New("attest: attestation failed")
	ErrTicket    = errors.New("attest: unknown or expired ticket")
)

// Evidence is the platform payload from AttestRequest.
type Evidence struct {
	Platform           string
	AppAttestKeyID     string
	AppAttestObject    []byte
	AppAttestAssertion []byte
	PlayIntegrityToken string
}

// Result of a successful verification.
type Result struct {
	DeviceRef     string
	PrizeEligible bool
}

// Verifier checks platform evidence against a challenge.
type Verifier interface {
	Verify(ctx context.Context, challenge []byte, ev Evidence) (Result, error)
}

// TTLs configures challenge and ticket lifetimes.
type TTLs struct {
	Challenge time.Duration
	Ticket    time.Duration
}

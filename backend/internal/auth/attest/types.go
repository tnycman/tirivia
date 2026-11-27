// Package attest issues challenges, verifies platform attestation and issues OTP tickets (SPEC §7).
package attest

import (
	"context"
	"errors"
	"time"

	"github.com/tirivia/backend/internal/auth/otp"
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
	JKT                string
}

// Result of a successful verification; Signals feed the risk score (AC-7).
type Result struct {
	DeviceRef     string
	PrizeEligible bool
	Signals       []string
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

// Store keeps single-use challenges and tickets with expiry.
type Store interface {
	PutChallenge(ctx context.Context, challenge []byte, ttl time.Duration) error
	ConsumeChallenge(ctx context.Context, challenge []byte) (bool, error)
	PutTicket(ctx context.Context, token string, dev otp.Device, ttl time.Duration) error
	Ticket(ctx context.Context, token string) (otp.Device, bool, error)
}

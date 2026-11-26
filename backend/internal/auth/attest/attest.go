// Package attest issues attestation challenges, verifies platform attestation (App Attest on
// iOS, Play Integrity on Android) and issues short-lived tickets that the OTP pipeline requires
// (SPEC §7: no code is sent without a fresh attestation).
//
// Slice 1 ships the challenge/ticket machinery and a development verifier. The real App Attest
// (Apple's 9 server checks) and Play Integrity (server-side decode, verdict policy) verifiers
// land in Phase 1 slice 2 behind the same Verifier interface.
package attest

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/tirivia/backend/internal/auth/otp"
)

// TTLs.
const (
	ChallengeTTL = 5 * time.Minute
	TicketTTL    = 10 * time.Minute
	// ChallengeBytes: Apple requires a server challenge of at least 16 bytes; we use 32.
	ChallengeBytes = 32
)

// Errors.
var (
	ErrChallenge = errors.New("attest: unknown, used or expired challenge")
	ErrFailed    = errors.New("attest: attestation failed")
	ErrTicket    = errors.New("attest: unknown or expired ticket")
)

// Evidence is the platform payload from AttestRequest.
type Evidence struct {
	Platform           string // "ios" | "android"
	AppAttestKeyID     string
	AppAttestObject    []byte
	AppAttestAssertion []byte
	PlayIntegrityToken string
}

// Result of a successful platform verification.
type Result struct {
	// DeviceRef identifies the app instance: the App Attest key id on iOS, or an app-scoped
	// device id on Android. Used for per-device limits and the AC-6 prize binding.
	DeviceRef string
	// PrizeEligible is false when the device must play for points only (AC-5).
	PrizeEligible bool
}

// Verifier checks platform evidence against a challenge.
type Verifier interface {
	Verify(ctx context.Context, challenge []byte, ev Evidence) (Result, error)
}

// Service holds challenges and tickets (Valkey in production; memory here).
type Service struct {
	verifier Verifier
	mu       sync.Mutex
	chal     map[string]time.Time
	tickets  map[string]ticket
	now      func() time.Time
}

type ticket struct {
	dev     otp.Device
	expires time.Time
}

// New returns a Service using verifier.
func New(v Verifier) *Service {
	return &Service{verifier: v, chal: map[string]time.Time{}, tickets: map[string]ticket{}, now: time.Now}
}

// Challenge returns a fresh single-use challenge.
func (s *Service) Challenge() ([]byte, time.Duration, error) {
	b := make([]byte, ChallengeBytes)
	if _, err := rand.Read(b); err != nil {
		return nil, 0, err
	}
	s.mu.Lock()
	s.chal[string(b)] = s.now().Add(ChallengeTTL)
	s.mu.Unlock()
	return b, ChallengeTTL, nil
}

// Attest consumes the challenge, verifies evidence and issues a ticket bound to the DPoP key.
func (s *Service) Attest(ctx context.Context, challenge []byte, ev Evidence, jkt string) (string, time.Duration, Result, error) {
	s.mu.Lock()
	exp, ok := s.chal[string(challenge)]
	delete(s.chal, string(challenge)) // single use, even on failure
	s.mu.Unlock()
	if !ok || s.now().After(exp) {
		return "", 0, Result{}, ErrChallenge
	}
	res, err := s.verifier.Verify(ctx, challenge, ev)
	if err != nil {
		return "", 0, Result{}, ErrFailed
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", 0, Result{}, err
	}
	tok := base64.RawURLEncoding.EncodeToString(raw)
	s.mu.Lock()
	s.tickets[tok] = ticket{dev: otp.Device{ID: res.DeviceRef, Platform: ev.Platform, JKT: jkt}, expires: s.now().Add(TicketTTL)}
	s.mu.Unlock()
	return tok, TicketTTL, res, nil
}

// Validate implements otp.Tickets.
func (s *Service) Validate(_ context.Context, tok string) (otp.Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tickets[tok]
	if !ok || s.now().After(t.expires) {
		return otp.Device{}, ErrTicket
	}
	return t.dev, nil
}

// DevVerifier accepts any non-empty evidence. It exists ONLY for local development and tests;
// cmd/core refuses to start with it unless AUTH_DEV_MODE=1.
type DevVerifier struct{}

// Verify implements Verifier.
func (DevVerifier) Verify(_ context.Context, _ []byte, ev Evidence) (Result, error) {
	switch {
	case ev.Platform == "ios" && ev.AppAttestKeyID != "":
		return Result{DeviceRef: "ios:" + ev.AppAttestKeyID, PrizeEligible: true}, nil
	case ev.Platform == "android" && ev.PlayIntegrityToken != "":
		id, _ := uuid.NewV7()
		return Result{DeviceRef: "android:" + id.String(), PrizeEligible: true}, nil
	}
	return Result{}, ErrFailed
}

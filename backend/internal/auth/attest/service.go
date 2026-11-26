package attest

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"github.com/tirivia/backend/internal/auth/otp"
)

type ticket struct {
	dev     otp.Device
	expires time.Time
}

// Service holds single-use challenges and tickets.
type Service struct {
	verifier Verifier
	ttl      TTLs
	mu       sync.Mutex
	chal     map[string]time.Time
	tickets  map[string]ticket
	now      func() time.Time
}

// New returns a Service.
func New(v Verifier, ttl TTLs) (*Service, error) {
	if v == nil || ttl.Challenge <= 0 || ttl.Ticket <= 0 {
		return nil, errors.New("attest: verifier and TTLs are required")
	}
	return &Service{verifier: v, ttl: ttl, chal: map[string]time.Time{}, tickets: map[string]ticket{}, now: time.Now}, nil
}

// Challenge returns a fresh single-use challenge.
func (s *Service) Challenge() ([]byte, time.Duration, error) {
	b := make([]byte, ChallengeBytes)
	if _, err := rand.Read(b); err != nil {
		return nil, 0, err
	}
	s.mu.Lock()
	s.chal[string(b)] = s.now().Add(s.ttl.Challenge)
	s.mu.Unlock()
	return b, s.ttl.Challenge, nil
}

// Attest consumes the challenge, verifies evidence and issues a ticket bound to jkt.
func (s *Service) Attest(ctx context.Context, challenge []byte, ev Evidence, jkt string) (string, time.Duration, Result, error) {
	if !s.consume(challenge) {
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
	s.tickets[tok] = ticket{dev: otp.Device{ID: res.DeviceRef, Platform: ev.Platform, JKT: jkt}, expires: s.now().Add(s.ttl.Ticket)}
	s.mu.Unlock()
	return tok, s.ttl.Ticket, res, nil
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

func (s *Service) consume(challenge []byte) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	exp, ok := s.chal[string(challenge)]
	delete(s.chal, string(challenge))
	return ok && !s.now().After(exp)
}

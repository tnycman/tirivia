package attest

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/tirivia/backend/internal/auth/otp"
)

// TicketBytes is the random size of an attestation ticket.
const TicketBytes = 32

// Service issues challenges and tickets around a platform Verifier.
type Service struct {
	verifier Verifier
	store    Store
	ttl      TTLs
}

// New returns a Service.
func New(v Verifier, store Store, ttl TTLs) (*Service, error) {
	if v == nil || store == nil || ttl.Challenge <= 0 || ttl.Ticket <= 0 {
		return nil, errors.New("attest: verifier, store and TTLs are required")
	}
	return &Service{verifier: v, store: store, ttl: ttl}, nil
}

// Challenge returns a fresh single-use challenge.
func (s *Service) Challenge(ctx context.Context) ([]byte, time.Duration, error) {
	b := make([]byte, ChallengeBytes)
	if _, err := rand.Read(b); err != nil {
		return nil, 0, err
	}
	if err := s.store.PutChallenge(ctx, b, s.ttl.Challenge); err != nil {
		return nil, 0, err
	}
	return b, s.ttl.Challenge, nil
}

// Attest consumes the challenge, verifies evidence and issues a ticket bound to jkt.
func (s *Service) Attest(ctx context.Context, challenge []byte, ev Evidence, jkt string) (string, time.Duration, Result, error) {
	ok, err := s.store.ConsumeChallenge(ctx, challenge)
	if err != nil {
		return "", 0, Result{}, err
	}
	if !ok {
		return "", 0, Result{}, ErrChallenge
	}
	res, err := s.verifier.Verify(ctx, challenge, ev)
	if err != nil {
		return "", 0, Result{}, ErrFailed
	}
	raw := make([]byte, TicketBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", 0, Result{}, err
	}
	tok := base64.RawURLEncoding.EncodeToString(raw)
	dev := otp.Device{ID: res.DeviceRef, Platform: ev.Platform, JKT: jkt}
	if err := s.store.PutTicket(ctx, tok, dev, s.ttl.Ticket); err != nil {
		return "", 0, Result{}, err
	}
	return tok, s.ttl.Ticket, res, nil
}

// Validate implements otp.Tickets.
func (s *Service) Validate(ctx context.Context, tok string) (otp.Device, error) {
	dev, ok, err := s.store.Ticket(ctx, tok)
	if err != nil {
		return otp.Device{}, err
	}
	if !ok {
		return otp.Device{}, ErrTicket
	}
	return dev, nil
}

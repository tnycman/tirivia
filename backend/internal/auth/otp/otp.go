// Package otp runs the sign-in code pipeline (SPEC §7, stack/05-auth-identity).
//
// Order of checks before anything is sent (research 05): attestation ticket -> sign-up switch ->
// country allow-list -> rate limits (number, device, number range) -> line type (VoIP and
// premium blocked) -> SMS-pumping risk -> send on the user's chosen channel (decision ID-1),
// falling back to SMS across two providers.
//
// Codes (NIST SP 800-63B-4 §3.1.3.2, §3.2.2): 6 digits from a CSPRNG, single use, valid 5 min
// (NIST allows up to 10), at most 5 attempts per code; only an HMAC of the code is stored and
// comparison is constant time.
package otp

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Policy values (DEFAULTs from the decision register, ID-2).
const (
	CodeDigits     = 6
	CodeTTL        = 5 * time.Minute
	MaxAttempts    = 5
	ResendAfter    = 30 * time.Second
	PumpingBlockAt = 80 // risk score 0-100 at or above which sending is refused (DEFAULT)
)

// Channel is the delivery channel the user picked.
type Channel string

// Supported channels (decision ID-1: the user chooses; SMS is the fallback).
const (
	ChannelSMS      Channel = "sms"
	ChannelWhatsApp Channel = "whatsapp"
	ChannelViber    Channel = "viber"
)

// Outcomes of Start, mirroring tirivia.core.v1.StartOtpResult.
var (
	ErrAttestationRequired = errors.New("otp: valid attestation ticket required")
	ErrSignupsPaused       = errors.New("otp: new sign-ups are paused")
	ErrNumberNotAllowed    = errors.New("otp: number not allowed")
	ErrRateLimited         = errors.New("otp: rate limited")
	ErrSendFailed          = errors.New("otp: all providers failed")
	ErrExpired             = errors.New("otp: code expired or unknown")
	ErrWrongCode           = errors.New("otp: wrong code")
	ErrTooManyAttempts     = errors.New("otp: too many attempts")
)

// blockedLineTypes: VoIP, toll-free, premium and similar (research 05; Twilio Lookup naming).
var blockedLineTypes = map[string]bool{
	"fixedVoip": true, "nonFixedVoip": true, "tollFree": true, "premium": true,
	"sharedCost": true, "uan": true, "pager": true, "voicemail": true,
}

var e164 = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

// Device is what a verified attestation ticket proves.
type Device struct {
	ID       string
	Platform string // "ios" | "android"
	// JKT is the DPoP key thumbprint presented when the device attested. Sign-in must use the
	// same key, which links the session's device binding to the attested app instance (ADR-0005).
	JKT string
}

// Dependencies (Valkey, Postgres and vendor adapters in production; fakes in tests).
type (
	// Tickets validates the short-lived attestation ticket issued by AuthService.Attest.
	Tickets interface {
		Validate(ctx context.Context, ticket string) (Device, error)
	}
	// Limiter is a fixed-window counter (Valkey INCR + EXPIRE).
	Limiter interface {
		Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
	}
	// NumberIntel returns line type and SMS-pumping risk (e.g. Twilio Lookup v2).
	NumberIntel interface {
		Lookup(ctx context.Context, e164 string) (lineType string, pumpingRisk int, err error)
	}
	// Sender delivers a code. Each provider has its own pumping filter enabled (research 05).
	Sender interface {
		Name() string
		Send(ctx context.Context, e164, code, locale string) error
	}
	// Signups reports the signups.enabled kill switch.
	Signups interface {
		Enabled(ctx context.Context) bool
	}
	// Directory reports whether a phone already has an account (existing users can sign in
	// while new sign-ups are paused).
	Directory interface {
		Exists(ctx context.Context, phoneHash string) (bool, error)
	}
	// Store keeps pending challenges (Valkey with TTL).
	Store interface {
		Put(ctx context.Context, c Challenge, ttl time.Duration) error
		Get(ctx context.Context, id string) (Challenge, error)
		IncrementAttempts(ctx context.Context, id string) (int, error)
		Delete(ctx context.Context, id string) error
	}
)

// Challenge is a pending code. Only the HMAC of the code is kept.
type Challenge struct {
	ID        string
	PhoneE164 string
	PhoneHash string
	Device    Device
	CodeMAC   []byte
	ExpiresAt time.Time
	Attempts  int
}

// Config wires the service.
type Config struct {
	Pepper          []byte               // HMAC key from Secrets Manager (>= 32 bytes)
	AllowedPrefixes []string             // E.164 country codes allowed for SMS (anti-pumping)
	Channels        map[Channel][]Sender // ordered failover per channel; SMS must have >= 1
	Tickets         Tickets
	Limiter         Limiter
	Intel           NumberIntel
	Signups         Signups
	Directory       Directory
	Store           Store
}

// Service runs the pipeline.
type Service struct {
	cfg Config
	now func() time.Time
}

// New validates config.
func New(cfg Config) (*Service, error) {
	if len(cfg.Pepper) < 32 {
		return nil, errors.New("otp: pepper must be at least 32 bytes")
	}
	if len(cfg.Channels[ChannelSMS]) == 0 {
		return nil, errors.New("otp: at least one SMS provider is required (fallback channel)")
	}
	return &Service{cfg: cfg, now: time.Now}, nil
}

// StartResult describes what was sent.
type StartResult struct {
	RequestID   string
	ChannelUsed Channel
	Provider    string
	ResendAfter time.Duration
}

// PhoneHash returns the keyed hash used to look up users by phone (users.phone_hash).
func (s *Service) PhoneHash(e164 string) string {
	m := hmac.New(sha256.New, s.cfg.Pepper)
	m.Write([]byte("phone:" + e164))
	return fmt.Sprintf("%x", m.Sum(nil))
}

func (s *Service) codeMAC(requestID, code string) []byte {
	m := hmac.New(sha256.New, s.cfg.Pepper)
	m.Write([]byte("otp:" + requestID + ":" + code))
	return m.Sum(nil)
}

func newCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000)) // uniform 000000-999999
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", CodeDigits, n.Int64()), nil
}

// Start validates the request and sends a code.
func (s *Service) Start(ctx context.Context, phone, ticket string, channel Channel, locale string) (StartResult, error) {
	dev, err := s.cfg.Tickets.Validate(ctx, ticket)
	if err != nil {
		return StartResult{}, ErrAttestationRequired
	}
	if !e164.MatchString(phone) {
		return StartResult{}, ErrNumberNotAllowed
	}
	phoneHash := s.PhoneHash(phone)
	if !s.cfg.Signups.Enabled(ctx) {
		exists, err := s.cfg.Directory.Exists(ctx, phoneHash)
		if err != nil {
			return StartResult{}, err
		}
		if !exists {
			return StartResult{}, ErrSignupsPaused
		}
	}
	if !s.allowedCountry(phone) {
		return StartResult{}, ErrNumberNotAllowed
	}
	// Rate limits (DEFAULT values): resend spacing, per number, per device, per number range.
	for _, l := range []struct {
		key    string
		limit  int
		window time.Duration
	}{
		{"otp:resend:" + phoneHash, 1, ResendAfter},
		{"otp:phone:" + phoneHash, 5, time.Hour},
		{"otp:device:" + dev.ID, 5, time.Hour},
		{"otp:range:" + rangeKey(phone), 100, 10 * time.Minute},
	} {
		ok, err := s.cfg.Limiter.Allow(ctx, l.key, l.limit, l.window)
		if err != nil {
			return StartResult{}, err
		}
		if !ok {
			return StartResult{}, ErrRateLimited
		}
	}
	lineType, risk, err := s.cfg.Intel.Lookup(ctx, phone)
	if err != nil {
		return StartResult{}, fmt.Errorf("otp: number lookup: %w", err)
	}
	if blockedLineTypes[lineType] || risk >= PumpingBlockAt {
		return StartResult{}, ErrNumberNotAllowed
	}

	code, err := newCode()
	if err != nil {
		return StartResult{}, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return StartResult{}, err
	}
	ch := Challenge{ID: id.String(), PhoneE164: phone, PhoneHash: phoneHash, Device: dev,
		CodeMAC: s.codeMAC(id.String(), code), ExpiresAt: s.now().Add(CodeTTL)}
	if err := s.cfg.Store.Put(ctx, ch, CodeTTL); err != nil {
		return StartResult{}, err
	}
	used, provider, err := s.send(ctx, channel, phone, code, locale)
	if err != nil {
		_ = s.cfg.Store.Delete(ctx, ch.ID)
		return StartResult{}, err
	}
	return StartResult{RequestID: ch.ID, ChannelUsed: used, Provider: provider, ResendAfter: ResendAfter}, nil
}

// send tries the chosen channel's providers in order, then SMS providers (fallback).
func (s *Service) send(ctx context.Context, channel Channel, phone, code, locale string) (Channel, string, error) {
	order := []Channel{channel}
	if channel != ChannelSMS {
		order = append(order, ChannelSMS)
	}
	for _, c := range order {
		for _, p := range s.cfg.Channels[c] {
			if err := p.Send(ctx, phone, code, locale); err == nil {
				return c, p.Name(), nil
			}
		}
	}
	return "", "", ErrSendFailed
}

// Verify checks a code. Attempts are counted before comparison so parallel guesses cannot
// exceed MaxAttempts.
func (s *Service) Verify(ctx context.Context, requestID, code string) (Challenge, error) {
	ch, err := s.cfg.Store.Get(ctx, requestID)
	if err != nil || s.now().After(ch.ExpiresAt) {
		return Challenge{}, ErrExpired
	}
	n, err := s.cfg.Store.IncrementAttempts(ctx, requestID)
	if err != nil {
		return Challenge{}, err
	}
	if n > MaxAttempts {
		_ = s.cfg.Store.Delete(ctx, requestID)
		return Challenge{}, ErrTooManyAttempts
	}
	if !hmac.Equal(s.codeMAC(requestID, code), ch.CodeMAC) { // constant-time
		if n == MaxAttempts {
			_ = s.cfg.Store.Delete(ctx, requestID)
		}
		return Challenge{}, ErrWrongCode
	}
	_ = s.cfg.Store.Delete(ctx, requestID) // single use
	return ch, nil
}

func (s *Service) allowedCountry(phone string) bool {
	for _, p := range s.cfg.AllowedPrefixes {
		if strings.HasPrefix(phone, p) {
			return true
		}
	}
	return false
}

// rangeKey groups numbers by their first 7 digits (country code + operator block) so bursts
// against one number range (a pumping pattern) are throttled together.
func rangeKey(phone string) string {
	if len(phone) > 8 {
		return phone[:8]
	}
	return phone
}

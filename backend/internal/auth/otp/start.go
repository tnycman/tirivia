package otp

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// StartResult describes a sent code.
type StartResult struct {
	RequestID   string
	ChannelUsed Channel
	Provider    string
	ResendAfter time.Duration
}

// Start runs the pre-send checks in order, then sends a code.
func (s *Service) Start(ctx context.Context, phone, ticket string, channel Channel, locale string) (StartResult, error) {
	dev, err := s.cfg.Tickets.Validate(ctx, ticket)
	if err != nil {
		return StartResult{}, ErrAttestationRequired
	}
	region, ok := parsePhone(phone)
	if !ok {
		return StartResult{}, ErrNumberNotAllowed
	}
	phoneHash := s.PhoneHash(phone)
	l := s.limitsFor(region)
	if limited, err := s.rateLimited(ctx, l, phoneHash, dev.ID, phone); err != nil || limited {
		return StartResult{}, firstErr(err, ErrRateLimited)
	}
	if err := s.checkSignups(ctx, phoneHash); err != nil {
		return StartResult{}, err
	}
	if blocked, err := s.numberBlocked(ctx, l, phone); err != nil || blocked {
		return StartResult{}, firstErr(err, ErrNumberNotAllowed)
	}
	return s.issue(ctx, Challenge{PhoneE164: phone, PhoneHash: phoneHash, Region: region, Device: dev}, channel, locale, l.Resend.Window)
}

func (s *Service) checkSignups(ctx context.Context, phoneHash string) error {
	if s.cfg.Signups.Enabled(ctx) {
		return nil
	}
	exists, err := s.cfg.Directory.Exists(ctx, phoneHash)
	if err != nil {
		return err
	}
	if !exists {
		return ErrSignupsPaused
	}
	return nil
}

func (s *Service) issue(ctx context.Context, ch Challenge, channel Channel, locale string, resend time.Duration) (StartResult, error) {
	code, err := newCode()
	if err != nil {
		return StartResult{}, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return StartResult{}, err
	}
	ch.ID, ch.CodeMAC, ch.ExpiresAt = id.String(), s.codeMAC(id.String(), code), s.now().Add(CodeTTL)
	if err := s.cfg.Store.Put(ctx, ch, CodeTTL); err != nil {
		return StartResult{}, err
	}
	used, provider, err := s.send(ctx, channel, ch.PhoneE164, code, locale)
	if err != nil {
		_ = s.cfg.Store.Delete(ctx, ch.ID)
		return StartResult{}, err
	}
	return StartResult{RequestID: ch.ID, ChannelUsed: used, Provider: provider, ResendAfter: resend}, nil
}

func firstErr(err, fallback error) error {
	if err != nil {
		return fmt.Errorf("otp: %w", err)
	}
	return fallback
}

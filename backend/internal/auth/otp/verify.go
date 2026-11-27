package otp

import (
	"context"
	"crypto/hmac"
)

// Verify checks a code; attempts are counted before comparing.
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
	if !hmac.Equal(s.codeMAC(requestID, code), ch.CodeMAC) {
		if _, err := s.cfg.Failures.Increment(ctx, ch.PhoneHash); err != nil {
			return Challenge{}, err
		}
		if n == MaxAttempts {
			_ = s.cfg.Store.Delete(ctx, requestID)
		}
		return Challenge{}, ErrWrongCode
	}
	_ = s.cfg.Store.Delete(ctx, requestID)
	if err := s.cfg.Failures.Reset(ctx, ch.PhoneHash); err != nil {
		return Challenge{}, err
	}
	return ch, nil
}

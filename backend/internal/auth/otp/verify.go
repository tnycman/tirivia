package otp

import (
	"context"
	"errors"
)

// Verify checks and consumes a code for the expected DPoP key.
func (s *Service) Verify(ctx context.Context, requestID, code, expectedJKT string) (Challenge, error) {
	ch, err := s.cfg.Store.Verify(ctx, requestID, expectedJKT, s.codeMAC(requestID, code), s.now(), MaxAttempts)
	if errors.Is(err, ErrWrongCode) {
		if _, failureErr := s.cfg.Failures.Increment(ctx, ch.PhoneHash); failureErr != nil {
			return Challenge{}, failureErr
		}
		return Challenge{}, ErrWrongCode
	}
	if err != nil {
		return Challenge{}, err
	}
	if err := s.cfg.Failures.Reset(ctx, ch.PhoneHash); err != nil {
		return Challenge{}, err
	}
	return ch, nil
}

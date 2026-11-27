package otp

import "errors"

// Errors returned by Start and Verify.
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

package dpop

import "errors"

// Verification errors.
var (
	ErrMissing    = errors.New("dpop: proof missing")
	ErrMalformed  = errors.New("dpop: malformed proof")
	ErrType       = errors.New("dpop: typ must be dpop+jwt")
	ErrKey        = errors.New("dpop: jwk header missing, invalid or private")
	ErrSignature  = errors.New("dpop: signature invalid")
	ErrMethod     = errors.New("dpop: htm mismatch")
	ErrURL        = errors.New("dpop: htu mismatch")
	ErrTime       = errors.New("dpop: iat outside window")
	ErrReplay     = errors.New("dpop: jti already used")
	ErrNonce      = errors.New("dpop: nonce missing or wrong")
	ErrAccessHash = errors.New("dpop: ath mismatch")
	ErrBinding    = errors.New("dpop: key does not match binding")
)

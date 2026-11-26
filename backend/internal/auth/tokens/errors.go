package tokens

import "errors"

// Validation errors.
var (
	ErrMalformed   = errors.New("tokens: malformed or wrongly signed token")
	ErrType        = errors.New("tokens: wrong token type")
	ErrUnknownKey  = errors.New("tokens: unknown key id")
	ErrClaims      = errors.New("tokens: iss/aud/exp/nbf invalid")
	ErrNotBound    = errors.New("tokens: token is not sender-constrained")
	ErrMissingData = errors.New("tokens: missing subject or token id")
)

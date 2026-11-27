package contracts

import (
	"context"
	"errors"
)

// ErrUnauthenticated is returned for any failed request authentication.
var ErrUnauthenticated = errors.New("contracts: unauthenticated")

// Principal is the authenticated caller of a player API.
type Principal struct {
	UserID   string
	DeviceID string
}

// Headers is satisfied by net/http.Header and connect.Header.
type Headers interface {
	Get(key string) string
	Values(key string) []string
}

// Authenticator verifies a DPoP-bound access token request (RFC 9449 §7); auth implements it, cmd wires it.
type Authenticator interface {
	Authenticate(ctx context.Context, header Headers, method, url string) (Principal, error)
}

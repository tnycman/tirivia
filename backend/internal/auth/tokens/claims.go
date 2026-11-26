// Package tokens issues and validates RFC 9068 access tokens (RFC 8725 profile, ADR-0005).
package tokens

import (
	"time"

	"github.com/go-jose/go-jose/v4/jwt"
)

// TypeAccessToken is the RFC 9068 typ.
const TypeAccessToken = "at+jwt"

// DefaultTTL is the access-token lifetime (decision ID-4).
const DefaultTTL = 10 * time.Minute

// Confirmation is the RFC 7800 cnf claim holding the DPoP key thumbprint.
type Confirmation struct {
	JKT string `json:"jkt"`
}

// Claims are the access-token claims.
type Claims struct {
	jwt.Claims
	ClientID string       `json:"client_id"`
	DeviceID string       `json:"did"`
	Cnf      Confirmation `json:"cnf"`
}

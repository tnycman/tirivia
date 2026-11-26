// Package tokens issues and validates access tokens (ADR-0005, SPEC §7).
//
// Profile: RFC 9068 JWT access token (typ "at+jwt"), signed with EdDSA (Ed25519), following
// RFC 8725 (JWT BCP): the algorithm is pinned per key, "alg":"none" and every other algorithm are
// rejected at parse time, typ/iss/aud/exp/nbf are validated, and the token is sender-constrained
// to the device key with DPoP (RFC 9449) via the "cnf.jkt" claim.
package tokens

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/google/uuid"
)

const (
	// TypeAccessToken is the RFC 9068 media type for JWT access tokens.
	TypeAccessToken = "at+jwt"
	// DefaultTTL is the access-token lifetime (decision ID-4).
	DefaultTTL = 10 * time.Minute
	// Leeway tolerates small clock skew between server tasks. go-jose's default is one minute,
	// which is too generous for a 10-minute token.
	Leeway = 30 * time.Second
)

// Errors returned by Validate. Callers map all of them to Connect CodeUnauthenticated.
var (
	ErrMalformed   = errors.New("tokens: malformed or wrongly signed token")
	ErrType        = errors.New("tokens: wrong token type")
	ErrUnknownKey  = errors.New("tokens: unknown key id")
	ErrClaims      = errors.New("tokens: claims invalid (iss/aud/exp/nbf)")
	ErrNotBound    = errors.New("tokens: token is not sender-constrained")
	ErrMissingData = errors.New("tokens: missing subject or token id")
)

// Confirmation is the RFC 7800 "cnf" claim; jkt is the RFC 7638 SHA-256 thumbprint of the
// device's DPoP public key (base64url, no padding).
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

// Issuer signs access tokens with the current key.
type Issuer struct {
	signer   jose.Signer
	issuer   string
	audience string
	ttl      time.Duration
	now      func() time.Time
}

// NewIssuer creates an Issuer for key kid. The private key comes from KMS/Secrets Manager.
func NewIssuer(kid string, key ed25519.PrivateKey, issuer, audience string) (*Issuer, error) {
	if kid == "" || len(key) != ed25519.PrivateKeySize {
		return nil, errors.New("tokens: kid and a 64-byte Ed25519 key are required")
	}
	opts := (&jose.SignerOptions{}).WithType(TypeAccessToken).WithHeader(jose.HeaderKey("kid"), kid)
	s, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.EdDSA, Key: key}, opts)
	if err != nil {
		return nil, fmt.Errorf("tokens: signer: %w", err)
	}
	return &Issuer{signer: s, issuer: issuer, audience: audience, ttl: DefaultTTL, now: time.Now}, nil
}

// Issue returns a signed access token bound to the DPoP key thumbprint jkt.
func (i *Issuer) Issue(userID, deviceID, clientID, jkt string) (string, time.Duration, error) {
	if userID == "" || jkt == "" {
		return "", 0, errors.New("tokens: user and DPoP thumbprint are required")
	}
	now := i.now()
	jti, err := uuid.NewV7()
	if err != nil {
		return "", 0, err
	}
	c := Claims{
		Claims: jwt.Claims{
			Issuer: i.issuer, Subject: userID, Audience: jwt.Audience{i.audience},
			IssuedAt: jwt.NewNumericDate(now), NotBefore: jwt.NewNumericDate(now),
			Expiry: jwt.NewNumericDate(now.Add(i.ttl)), ID: jti.String(),
		},
		ClientID: clientID, DeviceID: deviceID, Cnf: Confirmation{JKT: jkt},
	}
	tok, err := jwt.Signed(i.signer).Claims(c).Serialize()
	if err != nil {
		return "", 0, fmt.Errorf("tokens: sign: %w", err)
	}
	return tok, i.ttl, nil
}

// Validator verifies access tokens against the published key set (current + previous keys
// during rotation).
type Validator struct {
	keys     map[string]ed25519.PublicKey
	issuer   string
	audience string
	now      func() time.Time
}

// NewValidator copies keys so the accepted set cannot be mutated by callers.
func NewValidator(keys map[string]ed25519.PublicKey, issuer, audience string) *Validator {
	cp := make(map[string]ed25519.PublicKey, len(keys))
	for k, v := range keys {
		cp[k] = v
	}
	return &Validator{keys: cp, issuer: issuer, audience: audience, now: time.Now}
}

// Validate parses and checks a token. Only EdDSA is accepted (algorithm pinning).
func (v *Validator) Validate(raw string) (*Claims, error) {
	tok, err := jwt.ParseSigned(raw, []jose.SignatureAlgorithm{jose.EdDSA})
	if err != nil || len(tok.Headers) != 1 {
		return nil, ErrMalformed
	}
	h := tok.Headers[0]
	if typ, _ := h.ExtraHeaders[jose.HeaderType].(string); typ != TypeAccessToken {
		return nil, ErrType
	}
	key, ok := v.keys[h.KeyID]
	if !ok {
		return nil, ErrUnknownKey
	}
	var c Claims
	if err := tok.Claims(key, &c); err != nil {
		return nil, ErrMalformed
	}
	exp := jwt.Expected{Issuer: v.issuer, AnyAudience: jwt.Audience{v.audience}, Time: v.now()}
	if err := c.ValidateWithLeeway(exp, Leeway); err != nil || c.Expiry == nil {
		return nil, ErrClaims
	}
	if c.Subject == "" || c.ID == "" {
		return nil, ErrMissingData
	}
	if c.Cnf.JKT == "" {
		return nil, ErrNotBound
	}
	return &c, nil
}

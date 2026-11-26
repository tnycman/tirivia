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

// Issuer signs access tokens.
type Issuer struct {
	signer           jose.Signer
	issuer, audience string
	ttl              time.Duration
	now              func() time.Time
}

// NewIssuer creates an Issuer signing with key under kid.
func NewIssuer(kid string, key ed25519.PrivateKey, issuer, audience string) (*Issuer, error) {
	if kid == "" || len(key) != ed25519.PrivateKeySize {
		return nil, errors.New("tokens: kid and a 64-byte Ed25519 key are required")
	}
	opts := (&jose.SignerOptions{}).WithType(TypeAccessToken).WithHeader("kid", kid)
	s, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.EdDSA, Key: key}, opts)
	if err != nil {
		return nil, fmt.Errorf("tokens: signer: %w", err)
	}
	return &Issuer{signer: s, issuer: issuer, audience: audience, ttl: DefaultTTL, now: time.Now}, nil
}

// Issue returns a token bound to the DPoP thumbprint jkt.
func (i *Issuer) Issue(userID, deviceID, clientID, jkt string) (string, time.Duration, error) {
	if userID == "" || jkt == "" {
		return "", 0, errors.New("tokens: user and DPoP thumbprint are required")
	}
	jti, err := uuid.NewV7()
	if err != nil {
		return "", 0, err
	}
	now := i.now()
	c := Claims{
		Claims: jwt.Claims{
			Issuer: i.issuer, Subject: userID, Audience: jwt.Audience{i.audience}, ID: jti.String(),
			IssuedAt: jwt.NewNumericDate(now), NotBefore: jwt.NewNumericDate(now), Expiry: jwt.NewNumericDate(now.Add(i.ttl)),
		},
		ClientID: clientID, DeviceID: deviceID, Cnf: Confirmation{JKT: jkt},
	}
	tok, err := jwt.Signed(i.signer).Claims(c).Serialize()
	if err != nil {
		return "", 0, fmt.Errorf("tokens: sign: %w", err)
	}
	return tok, i.ttl, nil
}

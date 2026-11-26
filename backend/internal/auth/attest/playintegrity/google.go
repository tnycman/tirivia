package playintegrity

import (
	"context"
	"errors"

	pi "google.golang.org/api/playintegrity/v1"
)

// ErrEmptyPayload means Google returned no token payload.
var ErrEmptyPayload = errors.New("playintegrity: empty token payload")

// Google calls decodeIntegrityToken with Application Default Credentials (the linked service account).
type Google struct {
	svc         *pi.Service
	packageName string
}

// NewGoogle returns a decoder for packageName.
func NewGoogle(ctx context.Context, packageName string) (*Google, error) {
	svc, err := pi.NewService(ctx)
	if err != nil {
		return nil, err
	}
	return &Google{svc: svc, packageName: packageName}, nil
}

// Decode implements Decoder.
func (g *Google) Decode(ctx context.Context, token string) (Verdict, error) {
	res, err := g.svc.V1.DecodeIntegrityToken(g.packageName, &pi.DecodeIntegrityTokenRequest{IntegrityToken: token}).Context(ctx).Do()
	if err != nil {
		return Verdict{}, err
	}
	if res.TokenPayloadExternal == nil {
		return Verdict{}, ErrEmptyPayload
	}
	return fromPayload(res.TokenPayloadExternal), nil
}

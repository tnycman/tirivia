package api

import (
	"context"

	"connectrpc.com/connect/v2"

	corev1 "github.com/tirivia/protocol/gen/go/tirivia/core/v1"
	"github.com/tirivia/protocol/gen/go/tirivia/core/v1/corev1connect"

	"github.com/tirivia/backend/internal/auth/attest"
)

var platforms = map[corev1.Platform]string{corev1.Platform_PLATFORM_IOS: "ios", corev1.Platform_PLATFORM_ANDROID: "android"}

// GetAttestationChallenge implements AuthServiceHandler.
func (h *Handler) GetAttestationChallenge(_ context.Context, _ *corev1.GetAttestationChallengeRequest) (*corev1.GetAttestationChallengeResponse, error) {
	c, ttl, err := h.d.Attest.Challenge()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "challenge unavailable")
	}
	return &corev1.GetAttestationChallengeResponse{Challenge: c, TtlSeconds: uint32(ttl.Seconds())}, nil
}

// Attest implements AuthServiceHandler.
func (h *Handler) Attest(ctx context.Context, req *corev1.AttestRequest) (*corev1.AttestResponse, error) {
	p, err := h.proof(ctx, corev1connect.AuthServiceAttestProcedure)
	if err != nil {
		return nil, err
	}
	platform := platforms[req.GetPlatform()]
	if platform == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, "platform is required")
	}
	ev := attest.Evidence{Platform: platform, AppAttestKeyID: req.GetAppAttestKeyId(), AppAttestObject: req.GetAppAttestObject(),
		AppAttestAssertion: req.GetAppAttestAssertion(), PlayIntegrityToken: req.GetPlayIntegrityToken()}
	tok, ttl, _, err := h.d.Attest.Attest(ctx, req.GetChallenge(), ev, p.JKT)
	if err != nil {
		return nil, connect.NewError(connect.CodePermissionDenied, "attestation failed")
	}
	return &corev1.AttestResponse{AttestationTicket: tok, TtlSeconds: uint32(ttl.Seconds())}, nil
}

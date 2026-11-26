// Package api exposes the auth module as tirivia.core.v1.AuthService (SPEC §7, ADR-0005).
//
// Flow: GetAttestationChallenge -> Attest (+DPoP proof: binds the device key to the attested
// app) -> StartOtp (ticket required) -> VerifyOtp (+DPoP proof with the same key) -> session.
// RefreshSession and SignOut require a DPoP proof from the session's key.
package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"connectrpc.com/connect/v2"

	corev1 "github.com/tirivia/protocol/gen/go/tirivia/core/v1"
	"github.com/tirivia/protocol/gen/go/tirivia/core/v1/corev1connect"

	"github.com/tirivia/backend/internal/auth/attest"
	"github.com/tirivia/backend/internal/auth/dpop"
	"github.com/tirivia/backend/internal/auth/otp"
	"github.com/tirivia/backend/internal/auth/sessions"
	"github.com/tirivia/backend/internal/auth/tokens"
)

// Users finds or creates the account for a verified phone.
type Users interface {
	FindOrCreate(ctx context.Context, phoneHash string) (userID string, isNew bool, err error)
}

// Deps wires the handler.
type Deps struct {
	// PublicBaseURL is the externally visible API origin (e.g. https://api.example.com), used to
	// build the htu each DPoP proof must name. Behind CloudFront/ALB the Host header differs.
	PublicBaseURL string
	Attest        *attest.Service
	OTP           *otp.Service
	DPoP          *dpop.Verifier
	Sessions      *sessions.Manager
	Tokens        *tokens.Issuer
	Users         Users
}

// Handler implements corev1connect.AuthServiceHandler.
type Handler struct {
	corev1connect.UnimplementedAuthServiceHandler
	d Deps
}

// New returns a Handler.
func New(d Deps) *Handler {
	d.PublicBaseURL = strings.TrimRight(d.PublicBaseURL, "/")
	return &Handler{d: d}
}

var errUnauth = connect.NewError(connect.CodeUnauthenticated, "authentication failed")

// proof verifies the DPoP header for procedure. All failures look the same to the client.
func (h *Handler) proof(ctx context.Context, procedure string, o dpop.Options) (*dpop.Proof, error) {
	info, ok := connect.CallInfoForServerContext(ctx)
	if !ok {
		return nil, errUnauth
	}
	p, err := h.d.DPoP.Verify(ctx, info.RequestHeader().Get("DPoP"), http.MethodPost, h.d.PublicBaseURL+procedure, o)
	if err != nil {
		return nil, errUnauth
	}
	return p, nil
}

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
	p, err := h.proof(ctx, corev1connect.AuthServiceAttestProcedure, dpop.Options{})
	if err != nil {
		return nil, err
	}
	platform := map[corev1.Platform]string{corev1.Platform_PLATFORM_IOS: "ios", corev1.Platform_PLATFORM_ANDROID: "android"}[req.GetPlatform()]
	if platform == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, "platform is required")
	}
	tok, ttl, _, err := h.d.Attest.Attest(ctx, req.GetChallenge(), attest.Evidence{
		Platform: platform, AppAttestKeyID: req.GetAppAttestKeyId(), AppAttestObject: req.GetAppAttestObject(),
		AppAttestAssertion: req.GetAppAttestAssertion(), PlayIntegrityToken: req.GetPlayIntegrityToken(),
	}, p.JKT)
	if err != nil {
		return nil, connect.NewError(connect.CodePermissionDenied, "attestation failed")
	}
	return &corev1.AttestResponse{AttestationTicket: tok, TtlSeconds: uint32(ttl.Seconds())}, nil
}

var channels = map[corev1.OtpChannel]otp.Channel{
	corev1.OtpChannel_OTP_CHANNEL_SMS: otp.ChannelSMS, corev1.OtpChannel_OTP_CHANNEL_WHATSAPP: otp.ChannelWhatsApp,
	corev1.OtpChannel_OTP_CHANNEL_VIBER: otp.ChannelViber,
}

// StartOtp implements AuthServiceHandler. Business outcomes are results, not RPC errors.
func (h *Handler) StartOtp(ctx context.Context, req *corev1.StartOtpRequest) (*corev1.StartOtpResponse, error) {
	ch, ok := channels[req.GetPreferredChannel()]
	if !ok {
		ch = otp.ChannelSMS
	}
	res, err := h.d.OTP.Start(ctx, req.GetPhoneE164(), req.GetAttestationTicket(), ch, req.GetLocale())
	out := &corev1.StartOtpResponse{}
	switch {
	case err == nil:
		out.Result, out.OtpRequestId = corev1.StartOtpResult_START_OTP_RESULT_SENT, res.RequestID
		out.ResendAfterSeconds = uint32(res.ResendAfter.Seconds())
		for k, v := range channels {
			if v == res.ChannelUsed {
				out.ChannelUsed = k
			}
		}
	case errors.Is(err, otp.ErrAttestationRequired):
		out.Result = corev1.StartOtpResult_START_OTP_RESULT_ATTESTATION_REQUIRED
	case errors.Is(err, otp.ErrRateLimited):
		out.Result = corev1.StartOtpResult_START_OTP_RESULT_RATE_LIMITED
	case errors.Is(err, otp.ErrNumberNotAllowed):
		out.Result = corev1.StartOtpResult_START_OTP_RESULT_NUMBER_NOT_ALLOWED
	case errors.Is(err, otp.ErrSignupsPaused):
		out.Result = corev1.StartOtpResult_START_OTP_RESULT_SIGNUPS_PAUSED
	default:
		return nil, connect.NewError(connect.CodeUnavailable, "could not send a code; try again")
	}
	return out, nil
}

// VerifyOtp implements AuthServiceHandler.
func (h *Handler) VerifyOtp(ctx context.Context, req *corev1.VerifyOtpRequest) (*corev1.VerifyOtpResponse, error) {
	p, err := h.proof(ctx, corev1connect.AuthServiceVerifyOtpProcedure, dpop.Options{})
	if err != nil {
		return nil, err
	}
	ch, err := h.d.OTP.Verify(ctx, req.GetOtpRequestId(), req.GetCode())
	switch {
	case errors.Is(err, otp.ErrTooManyAttempts):
		return nil, connect.NewError(connect.CodeResourceExhausted, "too many attempts; request a new code")
	case err != nil:
		return nil, connect.NewError(connect.CodeUnauthenticated, "wrong or expired code")
	}
	// The sign-in key must be the key that attested (ADR-0005).
	if ch.Device.JKT != p.JKT {
		return nil, errUnauth
	}
	userID, isNew, err := h.d.Users.FindOrCreate(ctx, ch.PhoneHash)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "sign-in failed")
	}
	refresh, s, err := h.d.Sessions.Create(ctx, userID, ch.Device.ID, p.JKT)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "sign-in failed")
	}
	sess, err := h.session(userID, s, refresh, ch.Device.Platform, isNew)
	if err != nil {
		return nil, err
	}
	return &corev1.VerifyOtpResponse{Session: sess}, nil
}

// RefreshSession implements AuthServiceHandler.
func (h *Handler) RefreshSession(ctx context.Context, req *corev1.RefreshSessionRequest) (*corev1.RefreshSessionResponse, error) {
	p, err := h.proof(ctx, corev1connect.AuthServiceRefreshSessionProcedure, dpop.Options{})
	if err != nil {
		return nil, err
	}
	refresh, s, err := h.d.Sessions.Rotate(ctx, req.GetRefreshToken(), p.JKT)
	if err != nil {
		return nil, errUnauth
	}
	sess, err := h.session(s.UserID, s, refresh, "", false)
	if err != nil {
		return nil, err
	}
	return &corev1.RefreshSessionResponse{Session: sess}, nil
}

// SignOut implements AuthServiceHandler: revokes the session family of the refresh token.
func (h *Handler) SignOut(ctx context.Context, req *corev1.SignOutRequest) (*corev1.SignOutResponse, error) {
	if _, err := h.proof(ctx, corev1connect.AuthServiceSignOutProcedure, dpop.Options{}); err != nil {
		return nil, err
	}
	if err := h.d.Sessions.Revoke(ctx, req.GetRefreshToken()); err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, "sign-out failed; try again")
	}
	return &corev1.SignOutResponse{}, nil
}

func (h *Handler) session(userID string, s sessions.Session, refresh, platform string, isNew bool) (*corev1.Session, error) {
	client := "tirivia-app"
	if platform != "" {
		client = "tirivia-" + platform
	}
	at, ttl, err := h.d.Tokens.Issue(userID, s.DeviceID, client, s.JKT)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "token issue failed")
	}
	return &corev1.Session{AccessToken: at, AccessExpiresInSeconds: uint32(ttl.Seconds()),
		RefreshToken: refresh, UserId: userID, IsNewUser: isNew}, nil
}

package api

import (
	"context"
	"errors"

	"connectrpc.com/connect/v2"

	corev1 "github.com/tirivia/protocol/gen/go/tirivia/core/v1"
	"github.com/tirivia/protocol/gen/go/tirivia/core/v1/corev1connect"

	"github.com/tirivia/backend/internal/auth/accounts"
	"github.com/tirivia/backend/internal/auth/otp"
)

var channels = map[corev1.OtpChannel]otp.Channel{
	corev1.OtpChannel_OTP_CHANNEL_SMS:      otp.ChannelSMS,
	corev1.OtpChannel_OTP_CHANNEL_WHATSAPP: otp.ChannelWhatsApp,
	corev1.OtpChannel_OTP_CHANNEL_VIBER:    otp.ChannelViber,
}

var startResults = []struct {
	err    error
	result corev1.StartOtpResult
}{
	{otp.ErrAttestationRequired, corev1.StartOtpResult_START_OTP_RESULT_ATTESTATION_REQUIRED},
	{otp.ErrRateLimited, corev1.StartOtpResult_START_OTP_RESULT_RATE_LIMITED},
	{otp.ErrNumberNotAllowed, corev1.StartOtpResult_START_OTP_RESULT_NUMBER_NOT_ALLOWED},
	{otp.ErrSignupsPaused, corev1.StartOtpResult_START_OTP_RESULT_SIGNUPS_PAUSED},
}

// StartOtp implements AuthServiceHandler; business outcomes are results, not RPC errors.
func (h *Handler) StartOtp(ctx context.Context, req *corev1.StartOtpRequest) (*corev1.StartOtpResponse, error) {
	ch, ok := channels[req.GetPreferredChannel()]
	if !ok {
		ch = otp.ChannelSMS
	}
	res, err := h.d.OTP.Start(ctx, req.GetPhoneE164(), req.GetAttestationTicket(), ch, req.GetLocale())
	if err == nil {
		out := &corev1.StartOtpResponse{Result: corev1.StartOtpResult_START_OTP_RESULT_SENT,
			OtpRequestId: res.RequestID, ResendAfterSeconds: uint32(res.ResendAfter.Seconds())}
		for k, v := range channels {
			if v == res.ChannelUsed {
				out.ChannelUsed = k
			}
		}
		return out, nil
	}
	for _, r := range startResults {
		if errors.Is(err, r.err) {
			return &corev1.StartOtpResponse{Result: r.result}, nil
		}
	}
	return nil, connect.NewError(connect.CodeUnavailable, "could not send a code; try again")
}

// VerifyOtp implements AuthServiceHandler; the sign-in key must be the attested key.
func (h *Handler) VerifyOtp(ctx context.Context, req *corev1.VerifyOtpRequest) (*corev1.VerifyOtpResponse, error) {
	p, err := h.proof(ctx, corev1connect.AuthServiceVerifyOtpProcedure)
	if err != nil {
		return nil, err
	}
	ch, err := h.d.OTP.Verify(ctx, req.GetOtpRequestId(), req.GetCode(), p.JKT)
	switch {
	case errors.Is(err, otp.ErrTooManyAttempts):
		return nil, connect.NewError(connect.CodeResourceExhausted, "too many attempts; request a new code")
	case err != nil:
		return nil, connect.NewError(connect.CodeUnauthenticated, "wrong or expired code")
	}
	userID, isNew, err := h.d.Users.FindOrCreate(ctx, accounts.NewUser{PhoneHash: ch.PhoneHash, PhoneE164: ch.PhoneE164, Country: ch.Region})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "sign-in failed")
	}
	deviceID, err := h.d.Devices.Register(ctx, userID, ch.Device)
	if err != nil {
		return nil, errUnauth
	}
	refresh, s, err := h.d.Sessions.Create(ctx, userID, deviceID, p.JKT)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "sign-in failed")
	}
	sess, err := h.session(s, refresh, ch.Device.Platform, isNew)
	if err != nil {
		return nil, err
	}
	return &corev1.VerifyOtpResponse{Session: sess}, nil
}

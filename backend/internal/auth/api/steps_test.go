package api_test

import (
	"context"
	"testing"

	corev1 "github.com/tirivia/protocol/gen/go/tirivia/core/v1"
	"github.com/tirivia/protocol/gen/go/tirivia/core/v1/corev1connect"

	"github.com/tirivia/backend/internal/auth/dpop/dpopclient"
)

func (h harness) attest(t *testing.T, k *dpopclient.Key) string {
	t.Helper()
	ch := must(h.client.GetAttestationChallenge(context.Background(), &corev1.GetAttestationChallengeRequest{}))
	att := must(h.client.Attest(h.withProof(t, k, corev1connect.AuthServiceAttestProcedure), &corev1.AttestRequest{
		Platform: corev1.Platform_PLATFORM_ANDROID, Challenge: ch.GetChallenge(), PlayIntegrityToken: "dev-token"}))
	return att.GetAttestationTicket()
}

func (h harness) start(t *testing.T, k *dpopclient.Key) *corev1.StartOtpResponse {
	t.Helper()
	return must(h.client.StartOtp(context.Background(), &corev1.StartOtpRequest{PhoneE164: phone,
		AttestationTicket: h.attest(t, k), PreferredChannel: corev1.OtpChannel_OTP_CHANNEL_SMS, Locale: "sq-AL"}))
}

func (h harness) signIn(t *testing.T, k *dpopclient.Key) *corev1.Session {
	t.Helper()
	st := h.start(t, k)
	if st.GetResult() != corev1.StartOtpResult_START_OTP_RESULT_SENT {
		t.Fatalf("start: %v", st.GetResult())
	}
	v := must(h.client.VerifyOtp(h.withProof(t, k, corev1connect.AuthServiceVerifyOtpProcedure),
		&corev1.VerifyOtpRequest{OtpRequestId: st.GetOtpRequestId(), Code: h.sms.code(phone)}))
	return v.GetSession()
}

func (h harness) refresh(t *testing.T, k *dpopclient.Key, token string) (*corev1.RefreshSessionResponse, error) {
	t.Helper()
	return h.client.RefreshSession(h.withProof(t, k, corev1connect.AuthServiceRefreshSessionProcedure),
		&corev1.RefreshSessionRequest{RefreshToken: token})
}

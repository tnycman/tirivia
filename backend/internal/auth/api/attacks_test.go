package api_test

import (
	"context"
	"net/http"
	"testing"

	"connectrpc.com/connect/v2"

	corev1 "github.com/tirivia/protocol/gen/go/tirivia/core/v1"
	"github.com/tirivia/protocol/gen/go/tirivia/core/v1/corev1connect"

	"github.com/tirivia/backend/internal/auth/dpop/dpopclient"
)

func TestNoCodeWithoutAttestation(t *testing.T) {
	h := newHarness(t)
	st, err := h.client.StartOtp(context.Background(), &corev1.StartOtpRequest{PhoneE164: phone})
	if err != nil || st.GetResult() != corev1.StartOtpResult_START_OTP_RESULT_ATTESTATION_REQUIRED || h.sms.code(phone) != "" {
		t.Fatalf("code sent without attestation: %v %v", st, err)
	}
}

func TestAttestRequiresProofAndSingleUseChallenge(t *testing.T) {
	h := newHarness(t)
	dev := deviceKey(t)
	ctx := context.Background()
	ch := must(h.client.GetAttestationChallenge(ctx, &corev1.GetAttestationChallengeRequest{}))
	req := &corev1.AttestRequest{Platform: corev1.Platform_PLATFORM_ANDROID, Challenge: ch.GetChallenge(), PlayIntegrityToken: "x"}
	if _, err := h.client.Attest(ctx, req); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("attest without DPoP: %v", err)
	}
	must(h.client.Attest(h.withProof(t, dev, corev1connect.AuthServiceAttestProcedure), req))
	if _, err := h.client.Attest(h.withProof(t, dev, corev1connect.AuthServiceAttestProcedure), req); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("challenge replay accepted: %v", err)
	}
}

func TestProofForOtherEndpointRejected(t *testing.T) {
	h := newHarness(t)
	ctx, info := connect.NewClientContext(context.Background())
	p := must(deviceKey(t).Proof(dpopclient.ProofParams{Method: http.MethodPost, URL: h.base + corev1connect.AuthServiceSignOutProcedure}))
	info.RequestHeader().Set("DPoP", p)
	if _, err := h.client.RefreshSession(ctx, &corev1.RefreshSessionRequest{RefreshToken: "x"}); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("proof for another endpoint accepted: %v", err)
	}
}

func TestStolenRefreshTokenFromOtherKey(t *testing.T) {
	h := newHarness(t)
	dev, attacker := deviceKey(t), deviceKey(t)
	s := h.signIn(t, dev)
	if _, err := h.refresh(t, attacker, s.GetRefreshToken()); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("attacker key accepted: %v", err)
	}
	if _, err := h.refresh(t, dev, s.GetRefreshToken()); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("family not revoked after theft attempt: %v", err)
	}
}

func TestCodeFromAnotherDeviceRejected(t *testing.T) {
	h := newHarness(t)
	dev, thief := deviceKey(t), deviceKey(t)
	st := h.start(t, dev)
	if _, err := h.client.VerifyOtp(h.withProof(t, thief, corev1connect.AuthServiceVerifyOtpProcedure),
		&corev1.VerifyOtpRequest{OtpRequestId: st.GetOtpRequestId(), Code: h.sms.code(phone)}); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("code accepted from another device key: %v", err)
	}
}

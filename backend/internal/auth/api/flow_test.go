package api_test

import (
	"testing"
	"time"

	"connectrpc.com/connect/v2"

	corev1 "github.com/tirivia/protocol/gen/go/tirivia/core/v1"
	"github.com/tirivia/protocol/gen/go/tirivia/core/v1/corev1connect"
)

func TestSignInAndRotate(t *testing.T) {
	h := newHarness(t)
	dev := deviceKey(t)
	s1 := h.signIn(t, dev)
	if !s1.GetIsNewUser() || s1.GetAccessExpiresInSeconds() != 600 {
		t.Fatalf("first sign-in: %+v", s1)
	}
	c, err := h.validator.Validate(s1.GetAccessToken())
	if err != nil || c.Subject != s1.GetUserId() || c.Cnf.JKT == "" {
		t.Fatalf("access token: %v", err)
	}
	r2, err := h.refresh(t, dev, s1.GetRefreshToken())
	if err != nil || r2.GetSession().GetRefreshToken() == s1.GetRefreshToken() {
		t.Fatalf("refresh: %v", err)
	}
	if _, err := h.refresh(t, dev, s1.GetRefreshToken()); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("rotated token reused: %v", err)
	}
	if _, err := h.refresh(t, dev, r2.GetSession().GetRefreshToken()); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("family survived reuse: %v", err)
	}
}

func TestResendWindowAndReturningUser(t *testing.T) {
	h := newHarness(t)
	dev := deviceKey(t)
	s1 := h.signIn(t, dev)
	if r := h.start(t, dev).GetResult(); r != corev1.StartOtpResult_START_OTP_RESULT_RATE_LIMITED {
		t.Fatalf("resend inside window: %v", r)
	}
	h.limiter.Now = func() time.Time { return time.Now().Add(testLimits.Resend.Window + time.Second) }
	s2 := h.signIn(t, dev)
	if s2.GetIsNewUser() || s2.GetUserId() != s1.GetUserId() {
		t.Fatalf("returning user: %+v", s2)
	}
}

func TestSignOut(t *testing.T) {
	h := newHarness(t)
	dev := deviceKey(t)
	s := h.signIn(t, dev)
	if _, err := h.client.SignOut(h.withProof(t, dev, corev1connect.AuthServiceSignOutProcedure),
		&corev1.SignOutRequest{RefreshToken: s.GetRefreshToken()}); err != nil {
		t.Fatal(err)
	}
	if _, err := h.refresh(t, dev, s.GetRefreshToken()); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("signed-out session refreshed: %v", err)
	}
}

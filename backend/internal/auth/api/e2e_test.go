package api_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	corev1 "github.com/tirivia/protocol/gen/go/tirivia/core/v1"
	"github.com/tirivia/protocol/gen/go/tirivia/core/v1/corev1connect"

	"github.com/tirivia/backend/internal/auth"
	"github.com/tirivia/backend/internal/auth/api"
	"github.com/tirivia/backend/internal/auth/attest"
	"github.com/tirivia/backend/internal/auth/dpop"
	"github.com/tirivia/backend/internal/auth/dpop/dpopclient"
	"github.com/tirivia/backend/internal/auth/otp"
	"github.com/tirivia/backend/internal/auth/sessions"
	"github.com/tirivia/backend/internal/auth/tokens"
)

const phone = "+355691234567"

type captureSender struct {
	mu   sync.Mutex
	last map[string]string
}

func (c *captureSender) Name() string { return "capture" }
func (c *captureSender) Send(_ context.Context, p, code, _ string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.last[p] = code
	return nil
}
func (c *captureSender) code(p string) string { c.mu.Lock(); defer c.mu.Unlock(); return c.last[p] }

// clockLimiter is a fixed-window limiter with a manual clock, so the test can step past the
// 30 s resend window instead of sleeping.
type clockLimiter struct {
	mu     sync.Mutex
	offset time.Duration
	m      map[string]struct {
		n     int
		until time.Time
	}
}

func (c *clockLimiter) advance(d time.Duration) { c.mu.Lock(); c.offset += d; c.mu.Unlock() }
func (c *clockLimiter) Allow(_ context.Context, key string, limit int, window time.Duration) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now().Add(c.offset)
	e := c.m[key]
	if now.After(e.until) {
		e.n, e.until = 0, now.Add(window)
	}
	e.n++
	c.m[key] = e
	return e.n <= limit, nil
}

type allowAll struct{}

func (allowAll) Lookup(context.Context, string) (string, int, error) { return "mobile", 1, nil }
func (allowAll) Enabled(context.Context) bool                        { return true }

type harness struct {
	client    corev1connect.AuthServiceClient
	sms       *captureSender
	limiter   *clockLimiter
	validator *tokens.Validator
	base      string
}

func newHarness(t *testing.T) harness {
	t.Helper()
	sms := &captureSender{last: map[string]string{}}
	lim := &clockLimiter{m: map[string]struct {
		n     int
		until time.Time
	}{}}
	users := auth.NewMemUsers()
	at := attest.New(attest.DevVerifier{})
	o, err := otp.New(otp.Config{
		Pepper: []byte(strings.Repeat("k", 32)), AllowedPrefixes: []string{"+355"},
		Channels: map[otp.Channel][]otp.Sender{otp.ChannelSMS: {sms}},
		Tickets:  at, Limiter: lim, Intel: allowAll{}, Signups: allowAll{}, Directory: users, Store: otp.NewMemStore(),
	})
	if err != nil {
		t.Fatal(err)
	}
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	issuer, _ := tokens.NewIssuer("k1", priv, "https://api.test", "tirivia-api")

	srv := connect.NewServer()
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	corev1connect.RegisterAuthServiceHandler(srv, api.New(api.Deps{
		PublicBaseURL: ts.URL, Attest: at, OTP: o, DPoP: dpop.NewVerifier(dpop.NewMemoryReplayCache()),
		Sessions: sessions.NewManager(sessions.NewMemStore()), Tokens: issuer, Users: users,
	}))
	connecthttp.Mount(mux, srv)
	client := corev1connect.NewAuthServiceClient(connect.NewClient(connecthttp.NewTransport(ts.Client(), ts.URL)))
	return harness{client: client, sms: sms, limiter: lim, validator: tokens.NewValidator(map[string]ed25519.PublicKey{"k1": pub}, "https://api.test", "tirivia-api"), base: ts.URL}
}

func deviceKey(t *testing.T) *dpopclient.Key {
	t.Helper()
	p, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	k, err := dpopclient.New(p)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

// withProof attaches a fresh DPoP proof for procedure, as the apps do on every call.
func (h harness) withProof(t *testing.T, k *dpopclient.Key, procedure string) context.Context {
	t.Helper()
	ctx, info := connect.NewClientContext(context.Background())
	p, err := k.Proof(dpopclient.ProofParams{Method: http.MethodPost, URL: h.base + procedure})
	if err != nil {
		t.Fatal(err)
	}
	info.RequestHeader().Set("DPoP", p)
	return ctx
}

func code(err error) connect.Code { return connect.CodeOf(err) }

// signIn runs challenge -> attest -> start -> verify with device key k.
func (h harness) signIn(t *testing.T, k *dpopclient.Key) *corev1.Session {
	t.Helper()
	ch, err := h.client.GetAttestationChallenge(context.Background(), &corev1.GetAttestationChallengeRequest{Platform: corev1.Platform_PLATFORM_ANDROID})
	if err != nil {
		t.Fatal(err)
	}
	att, err := h.client.Attest(h.withProof(t, k, corev1connect.AuthServiceAttestProcedure), &corev1.AttestRequest{
		Platform: corev1.Platform_PLATFORM_ANDROID, Challenge: ch.GetChallenge(), PlayIntegrityToken: "dev-token"})
	if err != nil {
		t.Fatalf("attest: %v", err)
	}
	st, err := h.client.StartOtp(context.Background(), &corev1.StartOtpRequest{PhoneE164: phone, AttestationTicket: att.GetAttestationTicket(),
		PreferredChannel: corev1.OtpChannel_OTP_CHANNEL_SMS, Locale: "sq-AL"})
	if err != nil || st.GetResult() != corev1.StartOtpResult_START_OTP_RESULT_SENT {
		t.Fatalf("start: %v %v", st, err)
	}
	v, err := h.client.VerifyOtp(h.withProof(t, k, corev1connect.AuthServiceVerifyOtpProcedure),
		&corev1.VerifyOtpRequest{OtpRequestId: st.GetOtpRequestId(), Code: h.sms.code(phone)})
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	return v.GetSession()
}

// startOnly attests and requests a code, returning the StartOtp result.
func (h harness) startOnly(t *testing.T, k *dpopclient.Key) (corev1.StartOtpResult, error) {
	t.Helper()
	ch, err := h.client.GetAttestationChallenge(context.Background(), &corev1.GetAttestationChallengeRequest{})
	if err != nil {
		return 0, err
	}
	att, err := h.client.Attest(h.withProof(t, k, corev1connect.AuthServiceAttestProcedure), &corev1.AttestRequest{
		Platform: corev1.Platform_PLATFORM_ANDROID, Challenge: ch.GetChallenge(), PlayIntegrityToken: "dev-token"})
	if err != nil {
		return 0, err
	}
	st, err := h.client.StartOtp(context.Background(), &corev1.StartOtpRequest{PhoneE164: phone, AttestationTicket: att.GetAttestationTicket()})
	return st.GetResult(), err
}

func TestFullSignInRefreshReuseAndSignOut(t *testing.T) {
	h := newHarness(t)
	dev := deviceKey(t)

	s1 := h.signIn(t, dev)
	if !s1.GetIsNewUser() || s1.GetAccessExpiresInSeconds() != 600 {
		t.Fatalf("first sign-in: %+v", s1)
	}
	claims, err := h.validator.Validate(s1.GetAccessToken())
	if err != nil || claims.Subject != s1.GetUserId() || claims.Cnf.JKT == "" {
		t.Fatalf("access token invalid or unbound: %v", err)
	}

	// Rotation.
	r2, err := h.client.RefreshSession(h.withProof(t, dev, corev1connect.AuthServiceRefreshSessionProcedure),
		&corev1.RefreshSessionRequest{RefreshToken: s1.GetRefreshToken()})
	if err != nil || r2.GetSession().GetRefreshToken() == s1.GetRefreshToken() {
		t.Fatalf("refresh: %v", err)
	}
	// Stolen refresh token used from an attacker's key: rejected, and the family is revoked.
	attacker := deviceKey(t)
	if _, err := h.client.RefreshSession(h.withProof(t, attacker, corev1connect.AuthServiceRefreshSessionProcedure),
		&corev1.RefreshSessionRequest{RefreshToken: r2.GetSession().GetRefreshToken()}); code(err) != connect.CodeUnauthenticated {
		t.Fatalf("attacker key accepted: %v", err)
	}
	if _, err := h.client.RefreshSession(h.withProof(t, dev, corev1connect.AuthServiceRefreshSessionProcedure),
		&corev1.RefreshSessionRequest{RefreshToken: r2.GetSession().GetRefreshToken()}); code(err) != connect.CodeUnauthenticated {
		t.Fatalf("family should be revoked after theft attempt: %v", err)
	}

	// An immediate new code for the same number is refused (30 s resend spacing)...
	if st, err := h.startOnly(t, dev); err != nil || st != corev1.StartOtpResult_START_OTP_RESULT_RATE_LIMITED {
		t.Fatalf("resend inside 30 s: %v %v", st, err)
	}
	// ...and allowed after the window. Returning user signs in again: same account, not new.
	h.limiter.advance(31 * time.Second)
	s3 := h.signIn(t, dev)
	if s3.GetIsNewUser() || s3.GetUserId() != s1.GetUserId() {
		t.Fatalf("returning user: %+v", s3)
	}
	// Old refresh token (already rotated) presented again -> reuse.
	if _, err := h.client.RefreshSession(h.withProof(t, dev, corev1connect.AuthServiceRefreshSessionProcedure),
		&corev1.RefreshSessionRequest{RefreshToken: s1.GetRefreshToken()}); code(err) != connect.CodeUnauthenticated {
		t.Fatalf("rotated token reused: %v", err)
	}
	// Sign-out kills the session.
	if _, err := h.client.SignOut(h.withProof(t, dev, corev1connect.AuthServiceSignOutProcedure),
		&corev1.SignOutRequest{RefreshToken: s3.GetRefreshToken()}); err != nil {
		t.Fatal(err)
	}
	if _, err := h.client.RefreshSession(h.withProof(t, dev, corev1connect.AuthServiceRefreshSessionProcedure),
		&corev1.RefreshSessionRequest{RefreshToken: s3.GetRefreshToken()}); code(err) != connect.CodeUnauthenticated {
		t.Fatalf("signed-out session still refreshes: %v", err)
	}
}

func TestAttackPaths(t *testing.T) {
	h := newHarness(t)
	dev := deviceKey(t)
	ctx := context.Background()

	// No ticket -> no code sent.
	st, err := h.client.StartOtp(ctx, &corev1.StartOtpRequest{PhoneE164: phone})
	if err != nil || st.GetResult() != corev1.StartOtpResult_START_OTP_RESULT_ATTESTATION_REQUIRED || h.sms.code(phone) != "" {
		t.Fatalf("code sent without attestation: %v %v", st, err)
	}
	// Attest without a DPoP proof.
	ch, _ := h.client.GetAttestationChallenge(ctx, &corev1.GetAttestationChallengeRequest{})
	if _, err := h.client.Attest(ctx, &corev1.AttestRequest{Platform: corev1.Platform_PLATFORM_ANDROID,
		Challenge: ch.GetChallenge(), PlayIntegrityToken: "x"}); code(err) != connect.CodeUnauthenticated {
		t.Fatalf("attest without DPoP: %v", err)
	}
	// Challenge is single use (consumed above even though the call failed DPoP? no: DPoP failed
	// first, so use it now, then replay it).
	if _, err := h.client.Attest(h.withProof(t, dev, corev1connect.AuthServiceAttestProcedure), &corev1.AttestRequest{
		Platform: corev1.Platform_PLATFORM_ANDROID, Challenge: ch.GetChallenge(), PlayIntegrityToken: "x"}); err != nil {
		t.Fatalf("attest: %v", err)
	}
	if _, err := h.client.Attest(h.withProof(t, dev, corev1connect.AuthServiceAttestProcedure), &corev1.AttestRequest{
		Platform: corev1.Platform_PLATFORM_ANDROID, Challenge: ch.GetChallenge(), PlayIntegrityToken: "x"}); code(err) != connect.CodePermissionDenied {
		t.Fatalf("challenge replay accepted: %v", err)
	}
	// Proof made for another procedure is rejected (htu binding).
	ctxWrong, info := connect.NewClientContext(ctx)
	p, _ := dev.Proof(dpopclient.ProofParams{Method: http.MethodPost, URL: h.base + corev1connect.AuthServiceSignOutProcedure})
	info.RequestHeader().Set("DPoP", p)
	if _, err := h.client.RefreshSession(ctxWrong, &corev1.RefreshSessionRequest{RefreshToken: "x"}); code(err) != connect.CodeUnauthenticated {
		t.Fatalf("proof for another endpoint accepted: %v", err)
	}
}

// The OTP must be verified with the same device key that attested: a code intercepted and
// replayed from another device fails even when correct.
func TestCodeFromAnotherDeviceRejected(t *testing.T) {
	h := newHarness(t)
	dev, thief := deviceKey(t), deviceKey(t)
	ctx := context.Background()
	ch, _ := h.client.GetAttestationChallenge(ctx, &corev1.GetAttestationChallengeRequest{})
	att, err := h.client.Attest(h.withProof(t, dev, corev1connect.AuthServiceAttestProcedure), &corev1.AttestRequest{
		Platform: corev1.Platform_PLATFORM_IOS, Challenge: ch.GetChallenge(), AppAttestKeyId: "key-1"})
	if err != nil {
		t.Fatal(err)
	}
	st, _ := h.client.StartOtp(ctx, &corev1.StartOtpRequest{PhoneE164: phone, AttestationTicket: att.GetAttestationTicket()})
	if _, err := h.client.VerifyOtp(h.withProof(t, thief, corev1connect.AuthServiceVerifyOtpProcedure),
		&corev1.VerifyOtpRequest{OtpRequestId: st.GetOtpRequestId(), Code: h.sms.code(phone)}); code(err) != connect.CodeUnauthenticated {
		t.Fatalf("code accepted from a different device key: %v", err)
	}
}

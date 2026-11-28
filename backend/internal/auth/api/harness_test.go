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

	"trivia.al/protocol/gen/go/trivia/core/v1/corev1connect"

	"trivia.al/backend/internal/auth/accounts"
	"trivia.al/backend/internal/auth/api"
	"trivia.al/backend/internal/auth/attest"
	"trivia.al/backend/internal/auth/dpop"
	"trivia.al/backend/internal/auth/dpop/dpopclient"
	"trivia.al/backend/internal/auth/otp"
	"trivia.al/backend/internal/auth/sessions"
	"trivia.al/backend/internal/auth/tokens"
)

const phone = "+355691234567"

var testLimits = otp.Limits{
	Resend: otp.Limit{Max: 1, Window: 30 * time.Second}, PerPhone: otp.Limit{Max: 5, Window: time.Hour},
	PerDevice: otp.Limit{Max: 5, Window: time.Hour}, PerRange: otp.Limit{Max: 100, Window: 10 * time.Minute},
	RangeDigits: 7, PumpingBlockAt: 90,
}

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

func (c *captureSender) code(p string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.last[p]
}

type allowAll struct{}

func (allowAll) Lookup(context.Context, string) (string, int, error) { return "mobile", 1, nil }
func (allowAll) Enabled(context.Context) bool                        { return true }

type harness struct {
	client    corev1connect.AuthServiceClient
	sms       *captureSender
	limiter   *otp.MemLimiter
	validator *tokens.Validator
	base      string
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func newHarness(t *testing.T) harness {
	t.Helper()
	sms := &captureSender{last: map[string]string{}}
	users, lim := accounts.NewMem(), otp.NewMemLimiter()
	at := must(attest.New(attest.DevVerifier{}, attest.NewMemStore(), attest.TTLs{Challenge: 5 * time.Minute, Ticket: 10 * time.Minute}))
	o := must(otp.New(otp.Config{
		Pepper: []byte(strings.Repeat("k", 32)), CoreRegions: []string{"AL"}, CoreLimits: testLimits, OtherLimits: testLimits,
		Failures: otp.NewMemFailures(),
		Channels: map[otp.Channel][]otp.Sender{otp.ChannelSMS: {sms}},
		Tickets:  at, Limiter: lim, Intel: allowAll{}, Signups: allowAll{}, Directory: users, Store: otp.NewMemStore(),
	}))
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	issuer := must(tokens.NewIssuer("k1", priv, "https://api.test", "trivia-api"))
	verifier := must(dpop.NewVerifier(dpop.NewMemoryReplayCache(), time.Minute))

	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	srv := connect.NewServer()
	corev1connect.RegisterAuthServiceHandler(srv, api.New(api.Deps{PublicBaseURL: ts.URL, Attest: at, OTP: o, DPoP: verifier,
		Sessions: sessions.NewManager(sessions.NewMemStore()), Tokens: issuer, Users: users, Devices: users}))
	connecthttp.Mount(mux, srv)
	return harness{
		client: corev1connect.NewAuthServiceClient(connect.NewClient(connecthttp.NewTransport(ts.Client(), ts.URL))),
		sms:    sms, limiter: lim, base: ts.URL,
		validator: tokens.NewValidator(map[string]ed25519.PublicKey{"k1": pub}, "https://api.test", "trivia-api", 30*time.Second),
	}
}

func deviceKey(t *testing.T) *dpopclient.Key {
	t.Helper()
	p, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	return must(dpopclient.New(p))
}

func (h harness) withProof(t *testing.T, k *dpopclient.Key, procedure string) context.Context {
	t.Helper()
	ctx, info := connect.NewClientContext(context.Background())
	info.RequestHeader().Set("DPoP", must(k.Proof(dpopclient.ProofParams{Method: http.MethodPost, URL: h.base + procedure})))
	return ctx
}

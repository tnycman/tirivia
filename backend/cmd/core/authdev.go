package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"log/slog"
	"os"
	"strings"

	"trivia.al/protocol/gen/go/trivia/core/v1/corev1connect"

	"trivia.al/backend/internal/auth/accounts"
	authapi "trivia.al/backend/internal/auth/api"
	"trivia.al/backend/internal/auth/attest"
	"trivia.al/backend/internal/auth/dpop"
	"trivia.al/backend/internal/auth/otp"
	"trivia.al/backend/internal/auth/sessions"
	"trivia.al/backend/internal/auth/tokens"
)

// authHandler is fail-closed: sign-in runs only with AUTH_DEV_MODE=1 until production adapters exist.
func authHandler(log *slog.Logger) corev1connect.AuthServiceHandler {
	if os.Getenv("AUTH_DEV_MODE") != "1" {
		log.Warn("auth disabled until production adapters land (AUTH_DEV_MODE=1 for local dev)")
		return corev1connect.UnimplementedAuthServiceHandler{}
	}
	log.Warn("AUTH_DEV_MODE=1: dev attestation and log-only code delivery; never use in production")
	base := strings.TrimRight(os.Getenv("PUBLIC_BASE_URL"), "/")
	if base == "" {
		base = "http://localhost:8081"
	}
	users := accounts.NewMem()
	at := must(attest.New(attest.DevVerifier{}, attest.NewMemStore(), attestTTLs))
	sender := devSender{log: log}
	o := must(otp.New(otp.Config{
		Pepper: randomBytes(32), CoreRegions: coreRegions, CoreLimits: coreLimits, OtherLimits: otherLimits,
		Channels: map[otp.Channel][]otp.Sender{otp.ChannelSMS: {sender}, otp.ChannelWhatsApp: {sender}, otp.ChannelViber: {sender}},
		Tickets:  at, Limiter: otp.NewMemLimiter(), Failures: otp.NewMemFailures(), Intel: devIntel{},
		Signups: on{}, Directory: users, Store: otp.NewMemStore(),
	}))
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	return authapi.New(authapi.Deps{
		PublicBaseURL: base, Attest: at, OTP: o, Users: users, Devices: users,
		DPoP:     must(dpop.NewVerifier(dpop.NewMemoryReplayCache(), dpopWindow)),
		Sessions: sessions.NewManager(sessions.NewMemStore()),
		Tokens:   must(tokens.NewIssuer("dev-1", key, base, tokens.Audience)),
	})
}

type devSender struct{ log *slog.Logger }

func (devSender) Name() string { return "dev-log" }

func (d devSender) Send(_ context.Context, phone, code, _ string) error {
	d.log.Info("DEV ONLY: sign-in code", "phone_last4", phone[len(phone)-4:], "code", code)
	return nil
}

type devIntel struct{}

func (devIntel) Lookup(context.Context, string) (string, int, error) { return "mobile", 0, nil }

type on struct{}

func (on) Enabled(context.Context) bool { return true }

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func randomBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return b
}

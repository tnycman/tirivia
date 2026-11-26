// Command core is the non-hot-path deployable: auth, profile, content, wallet, risk, notify
// and chatmod modules (SPEC §6.1).
package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	"github.com/tirivia/protocol/gen/go/tirivia/core/v1/corev1connect"

	"github.com/tirivia/backend/internal/auth"
	authapi "github.com/tirivia/backend/internal/auth/api"
	"github.com/tirivia/backend/internal/auth/attest"
	"github.com/tirivia/backend/internal/auth/dpop"
	"github.com/tirivia/backend/internal/auth/otp"
	"github.com/tirivia/backend/internal/auth/sessions"
	"github.com/tirivia/backend/internal/auth/tokens"
	"github.com/tirivia/backend/internal/platform/httpserver"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rpc := connect.NewServer()
	corev1connect.RegisterAuthServiceHandler(rpc, authHandler(log))
	// Handlers are replaced module by module as phases land (PLAN.md).
	corev1connect.RegisterProfileServiceHandler(rpc, corev1connect.UnimplementedProfileServiceHandler{})
	corev1connect.RegisterContentServiceHandler(rpc, corev1connect.UnimplementedContentServiceHandler{})
	corev1connect.RegisterWalletServiceHandler(rpc, corev1connect.UnimplementedWalletServiceHandler{})

	mux := http.NewServeMux()
	connecthttp.Mount(mux, rpc)
	health := &httpserver.Health{}
	health.Register(mux)

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8081"
	}
	if err := httpserver.Run(ctx, httpserver.New(addr, mux), health, 25*time.Second, log); err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

// authHandler is FAIL-CLOSED: until Phase 1 slice 2 provides real App Attest / Play Integrity
// verification, persistent stores, KMS-held signing keys and real SMS providers, sign-in is only
// available with AUTH_DEV_MODE=1 (local/dev), which uses in-memory stores, an ephemeral signing
// key, a development attestation verifier and a sender that writes codes to the log.
func authHandler(log *slog.Logger) corev1connect.AuthServiceHandler {
	if os.Getenv("AUTH_DEV_MODE") != "1" {
		log.Warn("auth disabled: production verifiers land in Phase 1 slice 2 (set AUTH_DEV_MODE=1 for local dev)")
		return corev1connect.UnimplementedAuthServiceHandler{}
	}
	log.Warn("AUTH_DEV_MODE=1: development attestation and log-only code delivery; NEVER use in production")
	base := os.Getenv("PUBLIC_BASE_URL")
	if base == "" {
		base = "http://localhost:8081"
	}
	_, signKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	issuer, err := tokens.NewIssuer("dev-1", signKey, base, "tirivia-api")
	if err != nil {
		panic(err)
	}
	pepper := make([]byte, 32)
	if _, err := rand.Read(pepper); err != nil {
		panic(err)
	}
	users := auth.NewMemUsers()
	at := attest.New(attest.DevVerifier{})
	logSender := devSender{log: log}
	o, err := otp.New(otp.Config{
		Pepper:          pepper,
		AllowedPrefixes: []string{"+355", "+383", "+389", "+39", "+30", "+49", "+41", "+43", "+44", "+1"},
		Channels:        map[otp.Channel][]otp.Sender{otp.ChannelSMS: {logSender}, otp.ChannelWhatsApp: {logSender}, otp.ChannelViber: {logSender}},
		Tickets:         at, Limiter: otp.NewMemLimiter(), Intel: devIntel{}, Signups: on{}, Directory: users, Store: otp.NewMemStore(),
	})
	if err != nil {
		panic(err)
	}
	return authapi.New(authapi.Deps{
		PublicBaseURL: base, Attest: at, OTP: o, DPoP: dpop.NewVerifier(dpop.NewMemoryReplayCache()),
		Sessions: sessions.NewManager(sessions.NewMemStore()), Tokens: issuer, Users: users,
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

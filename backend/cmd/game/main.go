// Command game is the hot-path deployable: game, studio and engage modules (SPEC §6.1).
// It is stateless and pre-scaled on a schedule before each show.
//
// Two listeners:
//   - ADDR (default :8080): public GameService behind CloudFront/WAF/ALB.
//   - ADMIN_ADDR (default :9090): internal-only back-office API (SwitchService) reached by the
//     web-consoles BFF over the private network. Never attach it to the public ALB.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	"github.com/tirivia/protocol/gen/go/tirivia/admin/v1/adminv1connect"
	"github.com/tirivia/protocol/gen/go/tirivia/game/v1/gamev1connect"

	gameapi "github.com/tirivia/backend/internal/game/api"
	"github.com/tirivia/backend/internal/platform/httpserver"
	"github.com/tirivia/backend/internal/platform/switches"
	studioapi "github.com/tirivia/backend/internal/studio/api"
	studiostore "github.com/tirivia/backend/internal/studio/store"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Live switch snapshot read by the hot path. Every switch starts at its default (chat OFF).
	// Phase 2b: Postgres-backed store + Valkey refresher shared by all tasks.
	sw := switches.NewMemory(switches.Snapshot{})

	// Public API.
	public := connect.NewServer()
	gamev1connect.RegisterGameServiceHandler(public, gameapi.NewHandler(sw, nil))
	publicMux := http.NewServeMux()
	connecthttp.Mount(publicMux, public)
	publicHealth := &httpserver.Health{}
	publicHealth.Register(publicMux)

	// Internal back-office API.
	admin := connect.NewServer()
	adminv1connect.RegisterSwitchServiceHandler(admin,
		studioapi.NewSwitchHandler(studiostore.NewMemStore(log), operatorFromBFF(), sw))
	adminMux := http.NewServeMux()
	connecthttp.Mount(adminMux, admin)
	adminHealth := &httpserver.Health{}
	adminHealth.Register(adminMux)

	var wg sync.WaitGroup
	run := func(name, addr string, mux *http.ServeMux, h *httpserver.Health) {
		defer wg.Done()
		if err := httpserver.Run(ctx, httpserver.New(addr, mux), h, 25*time.Second, log.With("listener", name)); err != nil {
			log.Error("server stopped", "listener", name, "err", err)
			stop()
		}
	}
	wg.Add(2)
	go run("public", envOr("ADDR", ":8080"), publicMux, publicHealth)
	go run("admin", envOr("ADMIN_ADDR", ":9090"), adminMux, adminHealth)
	wg.Wait()
}

// operatorFromBFF identifies the operator. Deny-by-default: until Phase 2b wires verification
// of the BFF's signed SSO assertion, identity is accepted only with ADMIN_DEV_MODE=1 (local and
// dev environments) from the X-Operator-Id header.
func operatorFromBFF() studioapi.ActorFunc {
	devMode := os.Getenv("ADMIN_DEV_MODE") == "1"
	return func(ctx context.Context) (string, error) {
		if !devMode {
			return "", studioapi.ErrNoActor
		}
		info, ok := connect.CallInfoForServerContext(ctx)
		if !ok {
			return "", studioapi.ErrNoActor
		}
		if id := info.RequestHeader().Get("X-Operator-Id"); id != "" {
			return id, nil
		}
		return "", studioapi.ErrNoActor
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

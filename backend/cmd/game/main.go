// Command game serves the hot path (game, studio, engage) and the internal admin listener.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	"trivia.al/protocol/gen/go/trivia/admin/v1/adminv1connect"
	"trivia.al/protocol/gen/go/trivia/game/v1/gamev1connect"

	gameapi "trivia.al/backend/internal/game/api"
	"trivia.al/backend/internal/platform/httpserver"
	"trivia.al/backend/internal/platform/switches"
	studioapi "trivia.al/backend/internal/studio/api"
	studiostore "trivia.al/backend/internal/studio/store"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, enabled, err := loadGameConfig(os.Getenv)
	if err != nil {
		log.Error("game config", "err", err)
		os.Exit(1)
	}
	sw := switches.NewMemory(switches.Snapshot{})
	deps, closeKV, err := playerDeps(cfg, enabled, sw, log)
	if err != nil {
		log.Error("game wiring", "err", err)
		os.Exit(1)
	}
	defer closeKV()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	public := connect.NewServer()
	gamev1connect.RegisterGameServiceHandler(public, gameapi.New(deps))
	admin := connect.NewServer()
	adminv1connect.RegisterSwitchServiceHandler(admin, studioapi.NewSwitchHandler(studiostore.NewMemStore(log), operatorFromBFF(), sw))

	var wg sync.WaitGroup
	for _, l := range []struct {
		name, addr string
		srv        *connect.Server
	}{
		{"public", envOr("ADDR", ":8080"), public},
		{"admin", envOr("ADMIN_ADDR", ":9090"), admin},
	} {
		wg.Go(func() { serve(ctx, stop, log.With("listener", l.name), l.addr, l.srv) })
	}
	wg.Wait()
}

func serve(ctx context.Context, stop context.CancelFunc, log *slog.Logger, addr string, srv *connect.Server) {
	mux := http.NewServeMux()
	connecthttp.Mount(mux, srv)
	health := &httpserver.Health{}
	health.Register(mux)
	if err := httpserver.Run(ctx, httpserver.New(addr, mux), health, httpserver.DrainTimeout, log); err != nil {
		log.Error("server stopped", "err", err)
		stop()
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

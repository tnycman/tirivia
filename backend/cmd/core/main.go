// Command core is the non-hot-path deployable: auth, profile, content, wallet, risk, notify
// and chatmod modules (SPEC §6.1).
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	"github.com/tirivia/protocol/gen/go/tirivia/core/v1/corev1connect"

	"github.com/tirivia/backend/internal/platform/httpserver"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rpc := connect.NewServer()
	// Handlers are replaced module by module as phases land (PLAN.md).
	corev1connect.RegisterAuthServiceHandler(rpc, corev1connect.UnimplementedAuthServiceHandler{})
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

// Package httpserver runs the HTTP server for both deployables with health endpoints and
// graceful shutdown (SIGTERM → stop accepting → drain → exit within ECS stopTimeout).
package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"
)

// New returns a server that speaks HTTP/1.1 and cleartext HTTP/2 (h2c) behind the ALB.
// ALB → target defaults to HTTP/1.1; Connect unary works on both (docs/research/02).
func New(addr string, h http.Handler) *http.Server {
	var p http.Protocols
	p.SetHTTP1(true)
	p.SetUnencryptedHTTP2(true)
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		Protocols:         &p,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}
}

// Health tracks readiness. Liveness is always OK while the process runs.
type Health struct{ ready atomic.Bool }

// SetReady flips readiness (false during drain so the ALB stops routing).
func (h *Health) SetReady(v bool) { h.ready.Store(v) }

// Register mounts /healthz (liveness) and /readyz (readiness).
func (h *Health) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		if !h.ready.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

// Run serves until ctx is cancelled, then drains for up to drain.
func Run(ctx context.Context, srv *http.Server, health *Health, drain time.Duration, log *slog.Logger) error {
	errCh := make(chan error, 1)
	go func() {
		log.Info("listening", "addr", srv.Addr)
		errCh <- srv.ListenAndServe()
	}()
	health.SetReady(true)

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	log.Info("shutting down", "drain", drain)
	health.SetReady(false)
	// ctx is already cancelled; keep its values but not its cancellation for the drain window.
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), drain)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	if err := <-errCh; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

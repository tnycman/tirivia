package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

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

package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/tirivia/backend/internal/game/live"
)

// finalFlush runs after the listeners drain and must finish before killAt, when ECS sends SIGKILL.
func finalFlush(t *live.Tally, killAt time.Time, log *slog.Logger) {
	ctx, cancel := context.WithDeadline(context.Background(), killAt)
	defer cancel()
	if err := t.Flush(ctx); err != nil {
		log.Error("final tally flush failed", "err", err)
	}
}

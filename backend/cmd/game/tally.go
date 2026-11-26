package main

import (
	"context"
	"log/slog"

	"github.com/tirivia/backend/internal/game/live"
)

// finalFlush runs after the listeners drain, so answers accepted during the drain are counted.
func finalFlush(t *live.Tally, log *slog.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), finalFlushTimeout)
	defer cancel()
	if err := t.Flush(ctx); err != nil {
		log.Error("final tally flush failed", "err", err)
	}
}

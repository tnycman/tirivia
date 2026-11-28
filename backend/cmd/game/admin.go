package main

import (
	"context"
	"os"

	"connectrpc.com/connect/v2"

	studioapi "trivia.al/backend/internal/studio/api"
)

// operatorFromBFF is deny-by-default; X-Operator-Id is trusted only with ADMIN_DEV_MODE=1.
func operatorFromBFF() studioapi.ActorFunc {
	devMode := os.Getenv("ADMIN_DEV_MODE") == "1"
	return func(ctx context.Context) (string, error) {
		info, ok := connect.CallInfoForServerContext(ctx)
		if !devMode || !ok {
			return "", studioapi.ErrNoActor
		}
		if id := info.RequestHeader().Get("X-Operator-Id"); id != "" {
			return id, nil
		}
		return "", studioapi.ErrNoActor
	}
}

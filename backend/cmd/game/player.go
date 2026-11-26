package main

import (
	"log/slog"

	gameapi "github.com/tirivia/backend/internal/game/api"
	"github.com/tirivia/backend/internal/platform/kv"
	"github.com/tirivia/backend/internal/platform/switches"
)

// playerDeps opens Valkey and wires the player API; disabled, it serves switches only and every player call is rejected.
func playerDeps(cfg gameConfig, enabled bool, sw switches.Reader, log *slog.Logger) (gameapi.Deps, func(), error) {
	if !enabled {
		log.Warn("player API disabled: VALKEY_URL unset")
		return gameapi.Deps{Switches: sw}, func() {}, nil
	}
	c, err := kv.Open(cfg.valkeyURL)
	if err != nil {
		return gameapi.Deps{}, nil, err
	}
	d, err := gameDeps(cfg, c, sw)
	if err != nil {
		c.Close()
		return gameapi.Deps{}, nil, err
	}
	log.Info("player API enabled; extra lives off until the referral inventory lands (UF-5)")
	return d, c.Close, nil
}

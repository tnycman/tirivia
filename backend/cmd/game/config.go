package main

import (
	"crypto/ed25519"
	"errors"
	"time"

	gamev1 "github.com/tirivia/protocol/gen/go/tirivia/game/v1"
)

// gameConfig is what the player API needs; every field is required once VALKEY_URL is set.
type gameConfig struct {
	valkeyURL, publicBaseURL, issuer string
	keys                             map[string]ed25519.PublicKey
	cueKeys                          []*gamev1.CueKey
	playerTTL                        time.Duration
}

// loadGameConfig is fail-closed: no VALKEY_URL disables the player API, a partial config is an error.
func loadGameConfig(getenv func(string) string) (cfg gameConfig, enabled bool, err error) {
	cfg.valkeyURL = getenv("VALKEY_URL")
	if cfg.valkeyURL == "" {
		return cfg, false, nil
	}
	var errs []error
	if cfg.publicBaseURL, err = origin("PUBLIC_BASE_URL", getenv("PUBLIC_BASE_URL")); err != nil {
		errs = append(errs, err)
	}
	if cfg.issuer, err = origin("TOKEN_ISSUER", getenv("TOKEN_ISSUER")); err != nil {
		errs = append(errs, err)
	}
	if cfg.keys, err = parseKeys("ACCESS_TOKEN_KEYS", getenv("ACCESS_TOKEN_KEYS")); err != nil {
		errs = append(errs, err)
	}
	if cfg.cueKeys, err = cueKeys(getenv("CUE_PUBLIC_KEYS")); err != nil {
		errs = append(errs, err)
	}
	if cfg.playerTTL, err = time.ParseDuration(getenv("PLAYER_STATE_TTL")); err != nil || cfg.playerTTL <= 0 {
		errs = append(errs, errors.New("PLAYER_STATE_TTL must be a positive Go duration"))
	}
	return cfg, true, errors.Join(errs...)
}

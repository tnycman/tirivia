package main

import (
	"crypto/ed25519"
	"errors"
	"time"
)

// gameConfig is what the player API needs; every field is required once VALKEY_URL is set.
type gameConfig struct {
	valkeyURL, publicBaseURL, issuer string
	keys                             map[string]ed25519.PublicKey
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
	if cfg.keys, err = parseKeys(getenv("ACCESS_TOKEN_KEYS")); err != nil {
		errs = append(errs, err)
	}
	if cfg.playerTTL, err = time.ParseDuration(getenv("PLAYER_STATE_TTL")); err != nil || cfg.playerTTL <= 0 {
		errs = append(errs, errors.New("PLAYER_STATE_TTL must be a positive Go duration"))
	}
	return cfg, true, errors.Join(errs...)
}

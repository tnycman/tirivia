package main

import (
	"github.com/valkey-io/valkey-go"

	"github.com/tirivia/backend/internal/auth/authn"
	"github.com/tirivia/backend/internal/auth/dpop"
	"github.com/tirivia/backend/internal/auth/store/kvstore"
	"github.com/tirivia/backend/internal/auth/tokens"
	gameapi "github.com/tirivia/backend/internal/game/api"
	"github.com/tirivia/backend/internal/game/live"
	"github.com/tirivia/backend/internal/platform/livestate"
	"github.com/tirivia/backend/internal/platform/switches"
)

// gameDeps wires the player API to Valkey; Inventory stays nil until referral lives land (UF-5).
func gameDeps(cfg gameConfig, c valkey.Client, sw switches.Reader) (gameapi.Deps, error) {
	verifier, err := dpop.NewVerifier(kvstore.NewReplay(c), dpopWindow)
	if err != nil {
		return gameapi.Deps{}, err
	}
	validator := tokens.NewValidator(cfg.keys, cfg.issuer, tokens.Audience, jwtLeeway)
	return gameapi.Deps{
		PublicBaseURL: cfg.publicBaseURL, Switches: sw, Authn: authn.New(validator, verifier),
		Shows: livestate.NewShows(c), Players: live.NewPlayers(c, cfg.playerTTL), Answers: live.NewAnswers(c),
		Lives: live.NewLives(c),
	}, nil
}

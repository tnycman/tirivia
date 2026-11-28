package main

import (
	"github.com/valkey-io/valkey-go"

	"trivia.al/backend/internal/auth/authn"
	"trivia.al/backend/internal/auth/dpop"
	"trivia.al/backend/internal/auth/store/kvstore"
	"trivia.al/backend/internal/auth/tokens"
	gameapi "trivia.al/backend/internal/game/api"
	"trivia.al/backend/internal/game/live"
	"trivia.al/backend/internal/platform/livestate"
	"trivia.al/backend/internal/platform/switches"
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
		Lives: live.NewLives(c), CueKeys: cfg.cueKeys, ReactionBatchInterval: reactionBatchInterval,
	}, nil
}

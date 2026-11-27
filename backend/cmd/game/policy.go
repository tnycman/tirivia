package main

import "time"

// Sources: stack/05-auth-identity "Sourced values"; reactionBatchInterval is TS-9 / ADR-0006.
var (
	dpopWindow            = 60 * time.Second
	jwtLeeway             = 30 * time.Second
	reactionBatchInterval = 5 * time.Second
)

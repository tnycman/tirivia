package main

import (
	"time"

	"github.com/tirivia/backend/internal/platform/httpserver"
)

// Sources: stack/05-auth-identity "Sourced values", SPEC §6.7, stack/03-backend-go "Server timeouts".
var (
	dpopWindow        = 60 * time.Second
	jwtLeeway         = 30 * time.Second
	tallyFlushEvery   = 100 * time.Millisecond
	ecsStopTimeout    = 30 * time.Second
	finalFlushTimeout = ecsStopTimeout - httpserver.DrainTimeout
)

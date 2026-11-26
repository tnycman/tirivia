package main

import "time"

// Sources: stack/05-auth-identity "Sourced values".
var (
	dpopWindow = 60 * time.Second
	jwtLeeway  = 30 * time.Second
)

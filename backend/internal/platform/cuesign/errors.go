package cuesign

import "errors"

// Errors.
var (
	ErrTooLarge     = errors.New("cuesign: cue exceeds size budget")
	ErrMalformed    = errors.New("cuesign: malformed cue")
	ErrUnknownKey   = errors.New("cuesign: unknown signing key")
	ErrBadSignature = errors.New("cuesign: signature verification failed")
	ErrContent      = errors.New("cuesign: question content out of bounds")
	ErrWrongShow    = errors.New("cuesign: cue is for another show")
	ErrReplay       = errors.New("cuesign: cue replayed or out of order")
)

// Package cuesign signs and verifies timed-metadata cues: "t1." + base64url(SignedCue), Ed25519 (SPEC §6.4).
package cuesign

// Limits (IVS PutMetadata max 1 KB; worst-case question measured at 845 B).
const (
	Prefix               = "t1."
	MaxTextBytes         = 900
	MaxQuestionTextBytes = 240
	MaxOptionTextBytes   = 64
	OptionsPerQuestion   = 3
)

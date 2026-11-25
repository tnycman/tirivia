// Package studio accepts producer commands, releases questions from the vault, signs cues and calls PutMetadata (<=4 TPS), and owns feature switches.
//
// Deployable: game. Lands in: Phase 2a/2b. Spec: SPEC §6.2-6.4.
// Layout: api/ (Connect handlers), app/ (use cases), domain/ (pure rules), store/ (Postgres/Valkey).
// Cross-module calls go only through internal/contracts (enforced by go-arch-lint and depguard).
package studio

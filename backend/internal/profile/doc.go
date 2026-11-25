// Package profile owns users, devices, stats, friends, referrals, leaderboards and push tokens.
//
// Deployable: core. Lands in: Phase 1/6. Spec: SPEC §6.6.
// Layout: api/ (Connect handlers), app/ (use cases), domain/ (pure rules), store/ (Postgres/Valkey).
// Cross-module calls go only through internal/contracts (enforced by go-arch-lint and depguard).
package profile

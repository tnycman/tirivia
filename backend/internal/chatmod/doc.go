// Package chatmod mints IVS Chat tokens, shards rooms and hosts the review-handler logic; it honours the chat.enabled backend switch.
//
// Deployable: core. Lands in: Phase 3. Spec: SPEC §6.10, ADR-0009.
// Layout: api/ (Connect handlers), app/ (use cases), domain/ (pure rules), store/ (Postgres/Valkey).
// Cross-module calls go only through internal/contracts (enforced by go-arch-lint and depguard).
package chatmod

// Package risk scores fraud from the Kinesis event stream into clear/review/block.
//
// Deployable: core. Lands in: Phase 4. Spec: SPEC §8, stack/13-anti-cheat.
// Layout: api/ (Connect handlers), app/ (use cases), domain/ (pure rules), store/ (Postgres/Valkey).
// Cross-module calls go only through internal/contracts (enforced by go-arch-lint and depguard).
package risk

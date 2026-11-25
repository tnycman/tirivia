// Package engage ingests batched reactions, aggregates sharded counters, counts viewers server-side and builds pulse content.
//
// Deployable: game. Lands in: Phase 3. Spec: SPEC §2, ADR-0006.
// Layout: api/ (Connect handlers), app/ (use cases), domain/ (pure rules), store/ (Postgres/Valkey).
// Cross-module calls go only through internal/contracts (enforced by go-arch-lint and depguard).
package engage

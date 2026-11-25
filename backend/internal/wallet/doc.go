// Package wallet is the double-entry ledger (append-only, zero-sum), prize credit, claims and expiry.
//
// Deployable: core. Lands in: Phase 5. Spec: SPEC §3, stack/04-data.
// Layout: api/ (Connect handlers), app/ (use cases), domain/ (pure rules), store/ (Postgres/Valkey).
// Cross-module calls go only through internal/contracts (enforced by go-arch-lint and depguard).
package wallet

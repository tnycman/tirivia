// Package notify schedules and delivers push via APNs HTTP/2 and FCM HTTP v1 (tokens only, no topics).
//
// Deployable: core. Lands in: Phase 6. Spec: SPEC §6.11.
// Layout: api/ (Connect handlers), app/ (use cases), domain/ (pure rules), store/ (Postgres/Valkey).
// Cross-module calls go only through internal/contracts (enforced by go-arch-lint and depguard).
package notify

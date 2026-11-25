// Package content owns replays, clips, the winners wall, question of the day, likes, views and share cards.
//
// Deployable: core. Lands in: Phase 6/8. Spec: SPEC §2.2.
// Layout: api/ (Connect handlers), app/ (use cases), domain/ (pure rules), store/ (Postgres/Valkey).
// Cross-module calls go only through internal/contracts (enforced by go-arch-lint and depguard).
package content

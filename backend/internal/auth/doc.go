// Package auth handles attestation (App Attest, Play Integrity), the OTP pipeline, EdDSA access tokens, DPoP proofs and refresh rotation.
//
// Deployable: core. Lands in: Phase 1. Spec: SPEC §7, stack/05-auth-identity.
// Layout: api/ (Connect handlers), app/ (use cases), domain/ (pure rules), store/ (Postgres/Valkey).
// Cross-module calls go only through internal/contracts (enforced by go-arch-lint and depguard).
package auth

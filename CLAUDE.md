# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Trivia is a scheduled, live-hosted Albanian trivia show (12 questions; everyone who gets all 12 right splits the prize) played in native iOS and Android apps. It is a monorepo of six self-contained top-level projects, each meant to become its own repository: `protocol/`, `backend/`, `ios/`, `android/`, `web-consoles/`, `infra/`, plus `design-tokens/`.

Sources of truth, in order of precedence:
- `Master Plan/Trivia-Master-Plan-v1.1.pdf`: product scope and locked decisions.
- `docs/decision-register.md`: every choice that must be settled before building (IDs such as AC-6, ID-2, UF-4, LG-1). Settle a decision here before you build anything that depends on it. If a change departs from a decision, it needs an ADR in `docs/adr/`.
- `SPEC.md`: the technical spec. §16 lists every deviation from the Master Plan, each with its evidence. `PLAN.md` holds phases 0–8 with tasks and exit tests. Commits are named after them (e.g. "Phase 2a slice 2").
- `stack/NN-*/README.md`: one folder per function, with pinned versions, cited best practices, "Rules for this codebase", and tables of sourced values.
- `docs/build-standards.md`: the 20-point PR acceptance bar.

## Rules for every agent and subagent (owner-set)

These rules bind the main session and every agent it starts, including research agents, Codex and eval runs. The full workflow is in `.claude/skills/building-trivia-changes/SKILL.md`.

1. **Two corrections, then escalate.** When the same thing has needed correcting twice (a fix that fails review twice, a test that keeps failing, the owner correcting the same point twice), stop. Report what was tried, show the evidence and ask the owner; don't start a third round.
2. **Fresh reviewers get closed criteria.** Give any reviewer without this session's context the exact commit, numbered checks, what is out of scope, and what "sound" means. Have it answer HOLDS or FAILS for each check. Never send an open "find problems" brief.
3. **Report artifacts produced, not documents revised.** List commits, new files, test and lint output, measurements, research files and register row IDs, each with its evidence. Don't narrate edits.
4. **Status before reporting.** The owner runs `pwsh -File .claude/skills/building-trivia-changes/scripts/status.ps1 -Base <branch below>`. Its numbers go in the report: size against the base branch, Go files over 100 lines, correction count. Two corrections on one branch triggers rule 1.
5. **The owner runs every shell command.** Give git, go, docker and make commands as copy-paste PowerShell (or one `pwsh -File` line) and claim only what the pasted output shows. Agents that must run commands themselves, like Codex, need the owner's OK and stay read-only.

## Working rules (owner-set)

- **No guessed values.** Every constant traces to a primary source (an RFC, NIST, Apple, Google, AWS or vendor docs) recorded in `stack/` or `docs/research/`. Constants with no published value are config, not literals (see `stack/05-auth-identity` "Sourced values" → `backend/cmd/core/policy.go`). Use researched practice only.
- **Editorial values are per-show studio settings**, e.g. time slot, timers and themes. Never hardcode them, and don't ask the owner for them.
- **Small single-purpose files and one-line comments.** Rationale lives in `stack/` and `docs/`, not in code comments. Most Go files are under 100 lines.
- **Chat is fully built, switched from the back office and OFF by default** (ADR-0009). Never descope it.
- One concern per PR. The owner approves every PR.
- Protocol: never renumber or reuse a field (`reserved` the number and the name). Never send the correct answer to a client before `reveal`. Every new RPC or cue type ships with test vectors in the same PR.
- Every mutating RPC is idempotent. Money is integer minor units in an append-only, zero-sum ledger. Hot-path changes add no Postgres reads and no unsharded hot keys.

## Commands

### Backend (`backend/`, Go 1.27)

| Task | Command |
|---|---|
| Build | `make build` (`go build ./...`) |
| Unit tests | `make test` (`go test -race -count=1 ./...`). `-race` needs cgo; on this Windows machine drop `-race` |
| Single test | `go test -count=1 -run TestName ./internal/game/domain/` |
| Store and live-state tests | `make store-test` (Docker: Postgres 17 on :55433, Valkey 9.1 on :56379). These tests skip unless `TEST_DATABASE_URL` / `TEST_VALKEY_URL` are set |
| Lint | `make lint` (golangci-lint v2.15.0, includes depguard module boundaries) |
| Architecture rules | `make arch` (go-arch-lint) |
| Schema invariants | `make db-test`: expects exactly 10 rejected statements from `migrations/tests/invariants.sql` |
| sqlc | `make sqlc` to regenerate; CI runs `make sqlc-check` |
| Test vectors | `make vectors` (Go generator `-check` plus the independent Node runner) |
| Everything CI runs | `make ci` |
| Run locally | `go run ./cmd/game` (:8080 public, :9090 admin), `go run ./cmd/core` (:8081) |

### Protocol (`protocol/`, buf 1.75.0)

- Install the pinned tools with `sh scripts/install-tools.sh`.
- Lint and format check: `buf lint && buf format -d --exit-code`.
- After any proto change, regenerate Go with `buf generate --template buf.gen.go.yaml`. The backend consumes `trivia.al/protocol/gen` through a `replace` directive in `backend/go.mod`. CI fails if `gen/` is stale.
- `buf breaking` runs against the PR base. A deliberate break needs the `buf skip breaking` label.

### Other projects

- `design-tokens/`: `npm ci && npm run check`, which runs the WCAG 2.2 AA contrast gate and checks that the generated Swift, Compose and CSS are current. `npm run build` regenerates them.
- `web-consoles/` (npm workspaces): `npm run generate` (TS client from protos, needs buf), `npm run typecheck`, `npm run build`.
- `infra/` (Terraform 1.17): `terraform -chdir=infra fmt -check -recursive`, plus `init -backend=false && validate` per stack. CI also runs TFLint, Trivy and Checkov.
- `ios/` and `android/` cannot be compiled on this Windows workspace (no Xcode or JDK). macOS and Linux CI compile them.

Use `git --no-pager` in non-interactive shells, because the pager hangs them.

## Backend architecture

**Two deployables, one Go module.**
- `cmd/game` is the hot path: the game, studio and engage modules. It is pre-scaled before each show and holds the live state in Valkey.
- `cmd/core` runs auth/OTP, profile, content, wallet, risk, notify and chatmod against Aurora Postgres.

**Module boundaries are enforced** by `.go-arch-lint.yml` and depguard:
- Each `internal/<module>/` may depend only on itself, `internal/platform/*` and `internal/contracts/`.
- Cross-module calls go only through `contracts/` interfaces.
- `cmd/` is the only place that wires modules together.
- Modules use the layers `api/` (Connect handlers), `app/` (use cases), `domain/` (pure rules, no I/O) and `store/` (Postgres/Valkey). Not every module has every layer yet.

**Handlers take their dependencies through a `Deps` struct**, e.g. `gameapi.New(gameapi.Deps{...})`. Tests build a full `Deps`, so a passing handler test does not prove that `cmd/*/main.go` wires the real dependencies. Check the wiring when you add a dependency.

**The handler pipeline:** authenticate with DPoP, authorize at the object level (player keys come from the authenticated subject), validate, rate limit, run the use case, emit an event. Return Connect error codes, never raw error strings.

**Live game flow:**
1. The studio signs timed-metadata cues (`platform/cuesign`, Ed25519, 900-byte budget) and injects them into the Amazon IVS stream: `question_open`, `reveal`, `pulse`, `burst`, `final`.
2. The apps verify each cue and start their timers from it.
3. Answers go to `GameService` over Connect RPC.
4. Answer intake and extra lives are atomic Valkey operations (Lua) in `game/live`. The intake script counts each new answer into its bucket's tally in the same atomic step (64 buckets sharing a hash tag with the player keys, `platform/livestate`, TS-16); `live.Verify` checks tallies against a recount.
5. Players remaining = correct answers + extra lives used (the `_life` tally field).

**Auth (`internal/auth`):**
- Our own OTP over SMS, WhatsApp or Viber (rules in decision ID-2).
- App Attest (iOS) and Play Integrity (Android). A device that fails attestation plays for points only (AC-5).
- EdDSA JWT access tokens bound to a device key with DPoP (RFC 9449, `cnf.jkt`).
- Refresh-token families with rotation and reuse detection.

**Feature and kill switches** (`platform/switches`, owned by `studio`) are changed only through the internal admin listener, which is not exposed on the public port. Every change is audited.

**Shared conformance suite:** `protocol/testvectors/` holds language-neutral JSON vectors for cues and per-player option shuffle. Go, Node, Swift and Kotlin must all agree on them. The keys in them are test-only and derived from a public seed.

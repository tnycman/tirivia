# backend: Go modular monolith

Two deployables from a single Go module. Design is in [SPEC §6](../SPEC.md#6-system-design); the stack and best practices are in [`stack/03-backend-go`](../stack/03-backend-go/README.md).

```
backend/
├── cmd/
│   ├── game/                 # hot path: game + studio + engage (pre-scaled before each show)
│   └── core/                 # auth, profile, content, wallet, risk, notify, chatmod
├── internal/
│   ├── game/                 # state machine, answers, tallies, elimination, lives
│   │   ├── api/              #   Connect handlers (GameService)
│   │   └── domain/           #   pure rules: transitions, deadlines, IsAlive (tested)
│   ├── studio/               # producer commands, question vault, cue signing, PutMetadata, feature switches
│   ├── engage/               # reactions, viewer count, pulse
│   ├── auth/ profile/ content/ wallet/ risk/ notify/ chatmod/
│   ├── contracts/            # the ONLY shared interfaces between modules
│   └── platform/             # cross-cutting technical packages
│       ├── cuesign/          #   Ed25519 cue signing, 900 B budget (tested: worst case 845 B)
│       ├── shuffle/          #   per-player option order (shared with iOS/Android via test vectors)
│       ├── switches/         #   backend feature/kill switches (chat.enabled, ...)
│       └── httpserver/       #   h2c server, health, graceful drain
├── migrations/
│   ├── core/00001_init.sql   # goose; Aurora PostgreSQL 17
│   └── tests/invariants.sql  # ledger zero-sum, append-only, four-eyes, eligibility, switches
├── loadtest/k6/              # 600k-player show scenario
├── .golangci.yml             # golangci-lint v2 + depguard module boundaries
├── .go-arch-lint.yml         # modules depend only on platform + contracts
├── sqlc.yaml  Dockerfile  Makefile
```

## Module layout

Every module under `internal/<module>/` has the same four layers:

| Folder | Contents |
|---|---|
| `api/` | Connect handlers |
| `app/` | use cases |
| `domain/` | pure rules, no I/O |
| `store/` | Postgres and Valkey access |

## Commands

| Task | Command |
|---|---|
| Build | `make build` |
| Unit tests | `make test` (uses `-race`, which needs cgo; Linux CI runs it) |
| Lint | `make lint` (golangci-lint v2.15.0) |
| Architecture rules | `make arch` |
| Vulnerability scan | `make vuln` (govulncheck) |
| Schema invariants | `make db-test` (Postgres 17 in Docker, then goose, then `invariants.sql`) |
| Run locally | `go run ./cmd/game` (`:8080`), `go run ./cmd/core` (`:8081`) |

## Verified in this workspace on 2026-11-25

| Check | Result |
|---|---|
| `go build ./...` and `go vet ./...` with Go 1.27.4 | clean |
| `go test ./...` | 18 tests pass: cue signing and size budget, shuffle uniformity, show state machine, deadlines and elimination, chat switch on the game API, back-office SwitchService (auth, validation, audit, live propagation) |
| End-to-end (real binary + curl) | chat OFF by default → operator switches it ON through the admin API → game API honours it → OFF again; unauthenticated change rejected; admin API absent from the public port |
| goose v3.30.1 against `postgres:17-alpine` | `00001_init.sql` applies; all 12 checks in `invariants.sql` behave as designed |

## Code generation

`protocol/` generates the Go SDK (`github.com/tirivia/protocol/gen`). During local development the backend consumes it through a `replace` directive in `go.mod`. After any proto change, regenerate it:

```sh
cd ../protocol && buf generate --template buf.gen.go.yaml
```

# 03 Backend (Go modular monolith)

## Role
All server logic. The hot path covers the show state machine, answers, reactions and cues. The core path covers auth, profile, content, wallet, risk, notifications and chat moderation.

## Decision
Go, built as a **modular monolith in one Go module** with two binaries (Master Plan §5):

| Binary | Modules | Scaling |
|---|---|---|
| `cmd/game` | `game`, `studio`, `engage` | Stateless; scaled on a schedule before each show |
| `cmd/core` | `auth`, `profile`, `content`, `wallet`, `risk`, `notify`, `chatmod` | Steady baseline |

## Pinned versions (2026-11-25)

| Tool | Version | Source |
|---|---|---|
| Go | 1.27.4 (`go 1.27` in go.mod) | https://go.dev/doc/devel/release |
| connect-go | v2.0.2 | https://github.com/connectrpc/connect-go/releases |
| pgx | v5.13.0 | https://pkg.go.dev/github.com/jackc/pgx/v5 |
| sqlc | 1.31.1 | https://github.com/sqlc-dev/sqlc/releases |
| goose | 3.30.1 | https://github.com/pressly/goose/releases |
| valkey-go / go-redis | 1.0.80 / 9.24.0 | https://github.com/valkey-io/valkey-go/releases |
| OTel Go | v1.47.0 | https://github.com/open-telemetry/opentelemetry-go/releases |
| golangci-lint | 2.15.0 | https://github.com/golangci/golangci-lint/releases |
| go-arch-lint | latest stable (version UNVERIFIED) | https://github.com/fe3dback/go-arch-lint |

## Best practices
1. **Layout follows the official Go guide.**
   - Commands go in `cmd/<name>`; all logic goes in `internal/`.
   - `golang-standards/project-layout` is explicitly not an official standard.
   - ([go.dev/doc/modules/layout](https://go.dev/doc/modules/layout))
2. **Module boundaries are enforced in CI**, not by convention:
   - `go-arch-lint`: each module may depend only on `platform/*` and `contracts`.
   - `depguard`: denies importing another module's `store` or `domain` packages.
3. **Style precedence:**
   1. [Google Go Style Guide](https://google.github.io/styleguide/go/)
   2. [Effective Go](https://go.dev/doc/effective_go)
   3. [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) for team conventions
4. **Idempotency** follows the IETF draft-09 semantics and uses Redis `SET NX EX`. For answers, the natural key (show, question, user) is the real guard. ([draft](https://www.ietf.org/archive/id/draft-ietf-httpapi-idempotency-key-header-09.html), [Stripe](https://stripe.com/blog/idempotency))
5. **Hot counters are pre-aggregated.**
   - Each task holds in-memory counters and flushes them every 100–250 ms with pipelined `HINCRBY` into 64 shards, each shard under its own hash tag.
   - One key per option would hit the per-slot ceiling.
   - ([ElastiCache troubleshooting](https://docs.aws.amazon.com/AmazonElastiCache/latest/dg/wwe-troubleshooting.html))
6. **The hot path never waits on Postgres.**
   - Answers go to Valkey and Kinesis.
   - Postgres is written after each round.
7. **Retries use exponential backoff with jitter** for every outbound call: PutMetadata, SMS, APNs, FCM.
8. **Postgres access goes through RDS Proxy** to avoid session pinning:
   - pgx in `QueryExecModeExec` mode
   - no `SET` outside `SET LOCAL`
   - UUIDv7 keys
   - ([pinning](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/rds-proxy-pinning.html))
9. **Graceful shutdown.**
   - `SIGTERM` → stop accepting traffic → drain in-flight requests → flush counters → exit.
   - The drain must finish inside the ECS `stopTimeout`.

## Rules for this codebase
- `internal/<module>/` contains `api/` (Connect handlers), `app/` (use cases), `domain/` (pure types and rules) and `store/` (Postgres/Valkey).
- Cross-module calls go only through interfaces in `internal/contracts/`.
- Every handler: authenticate (DPoP), authorize (object level), validate, rate limit, run the use case, emit an event.
- Return a Connect error code, never a raw error string, to clients.
- Every operation is a context-aware function with a deadline.

## Changes vs Master Plan
- The tally design uses per-task pre-aggregation plus 64 shards spread across slots.
- The `studio` module also owns **feature switches**, which include the chat switch.

## Code
[`backend/`](../../backend/)

## Sources
[docs/research/02-backend-protocol-data.md](../../docs/research/02-backend-protocol-data.md)

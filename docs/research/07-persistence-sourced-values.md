# 07 Persistence: sourced facts (verified 2026-11-25)

These facts back the storage adapters in `backend/internal/auth/store/{pgstore,kvstore}` and in `platform/{pgdb,kv,seal}`. Each fact was read from its primary source. **UNVERIFIED** means no source states it.

## Postgres behind RDS Proxy

| Fact | Source |
|---|---|
| **What pins a PostgreSQL session.** `SET`; `PREPARE`/`DISCARD`/`DEALLOCATE`/`EXECUTE`; temporary tables, sequences and views; cursors; `LISTEN`; `nextval`/`setval`; `pg_advisory_lock`; any statement over 16 KB. | [RDS Proxy pinning](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/rds-proxy-pinning.html) |
| **`SET LOCAL` does not pin inside `BEGIN`…`COMMIT`.** `set_config(..., true)` does pin. | Same page |
| **Extended-protocol prepared statements are multiplexed and do not pin.** SQL-level `PREPARE` still pins. | [AWS Database Blog, 2023-11-10](https://aws.amazon.com/blogs/database/amazon-rds-proxy-multiplexing-support-for-postgresql-extended-query-protocol/) (blog, not the user guide) |
| **pgx `QueryExecModeCacheStatement` can fail on schema change.** "If the database schema is modified … the first execution of a previously cached query may fail." | pgx v5.13.0 `conn.go` |
| **pgx `QueryExecModeExec`** uses the extended protocol with no named statements, in one round trip. pgx recommends it for proxies without prepared-statement support. | pgx v5.13.0 `conn.go`, `doc.go` |

**Our choice: `QueryExecModeExec`.**
- RDS Proxy would multiplex cached statements, but rolling deploys run migrations while old tasks are still serving, which is the documented failure case for cached statements.
- The schema has **no sequences**; keys are UUIDv7 generated in the app, so no `nextval` pinning.
- No session-level `SET`.

## sqlc

- **v1.31.1** with `sql_package: pgx/v5` ([guide](https://docs.sqlc.dev/en/latest/guides/using-go-and-pgx.html)).
- On Windows the Postgres parser is a WebAssembly build, so no cgo is needed (`internal/engine/postgresql/parse_wasi.go`).
- `sqlc diff` fails CI when the generated code is stale.

## Valkey

| Use | Primitive | Source |
|---|---|---|
| Fixed-window limiter | Lua `INCR`, then `PEXPIRE` when the count is 1. This is the Redis "rate limiter 2" pattern, which fixes the leaked-key race in pattern 1 | [INCR](https://redis.io/docs/latest/commands/incr/) |
| DPoP replay | `SET NX PX`. Null reply = key exists = replay | [SET](https://valkey.io/commands/set/) |
| Single-use attestation challenge | `GETDEL`, since 6.2.0 (read from the 9.1.4 server's `COMMAND DOCS`) | Server docs |
| Atomicity | Scripts and transactions run as "a single isolated operation". Single-key operations are cluster-safe | [Transactions](https://valkey.io/topics/transactions/), [cluster spec](https://valkey.io/topics/cluster-spec/) |
| Client | valkey-go v1.0.80. Lua scripts are preferred over dedicated MULTI connections for throughput. Client-side caching is disabled because nothing calls `DoCache` | [README](https://github.com/valkey-io/valkey-go) |

**OPEN:** `ShardsRefreshInterval` is needed so the client notices ElastiCache replica scaling. Set it when the cluster endpoint is wired (stack/04).

## Field encryption

- AES-256-GCM via `cipher.NewGCMWithRandomNonce`, which uses a 96-bit random nonce.
- The Go docs cap use at "2^32 messages" per key. This follows NIST SP 800-38D §8.3.
- Each ciphertext is prefixed with a key ID so keys can rotate.
- The phone hash is the AAD, which binds each ciphertext to its row.
- Where the key comes from (Secrets Manager or a KMS data key) is wired with the production adapters.

## CI

- **Service containers** follow GitHub's documented `services:` pattern with health checks ([Postgres](https://docs.github.com/en/actions/tutorials/use-containerized-services/create-postgresql-service-containers)). GitHub has no Valkey example; the `valkey-cli ping` health check is adapted from its Redis example.
- **`buf breaking`:**
  - It compares against the PR base commit. Buf's CI docs warn that `#branch=main` gives false positives when branches land one after another.
  - It is skipped only when the visible `buf skip breaking` label is on the PR ([buf GitHub Action](https://buf.build/docs/bsr/ci-cd/github-actions/)).
  - The category stays FILE, the documented default.

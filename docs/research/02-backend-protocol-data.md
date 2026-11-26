# Research 02: Go backend, protocol, data (Postgres, Valkey, Kinesis), observability, load testing

> Checked against primary sources on **2026-11-25**. **UNVERIFIED** marks a claim we could not confirm. "(yr n/s)" means the GitHub release page gave only day and month; the year is presumed to be 2026.
> Feeds: [`stack/01-protocol`](../../stack/01-protocol/README.md), [`stack/03-backend-go`](../../stack/03-backend-go/README.md), [`stack/04-data`](../../stack/04-data/README.md), [`stack/11-observability`](../../stack/11-observability/README.md), [`stack/14-testing-quality`](../../stack/14-testing-quality/README.md).

## 1. Verified facts and limits

| Fact | Value | Source |
|---|---|---|
| Go stable | **go1.27.4** (2026-11-10); 1.27.0 on 2026-08-19; 1.26.x still patched | https://go.dev/doc/devel/release |
| Official layout | Server logic goes under `internal/`, commands under `cmd/` | https://go.dev/doc/modules/layout |
| golang-standards/project-layout | "NOT an official standard defined by the core Go dev team" | https://github.com/golang-standards/project-layout |
| buf CLI | **v1.75.0** (22 Oct) | https://github.com/bufbuild/buf/releases |
| buf v2 defaults | lint STANDARD; breaking FILE (strictest). **WIRE_JSON is the "recommended minimum" with Connect or JSON** | https://buf.build/docs/configuration/v2/buf-yaml/ · https://buf.build/docs/breaking/rules/ |
| Connect protocol | Unary is a POST over HTTP/1.1 or HTTP/2, `application/proto` or JSON, no trailers. GET is allowed for `NO_SIDE_EFFECTS`. Deadlines use `Connect-Timeout-Ms` | https://connectrpc.com/docs/protocol/ |
| connect-go | **v2.0.2** (21 Oct), path `connectrpc.com/connect/v2`. v1.21.1 is "supported indefinitely". It serves Connect, gRPC and gRPC-Web | https://github.com/connectrpc/connect-go/releases |
| connect-swift / connect-kotlin | **1.2.3** / **0.10.0** | https://github.com/connectrpc/connect-swift/releases · https://github.com/connectrpc/connect-kotlin/releases |
| Swift/Kotlin codegen | Swift uses `connectrpc/swift` with `apple/swift` (URLSession). Kotlin uses `connectrpc/kotlin` with javalite/kotlin-lite (OkHttp) | https://connectrpc.com/docs/swift/getting-started/ · https://connectrpc.com/docs/kotlin/getting-started/ |
| swift-protobuf / protobuf-es | **1.41.0** / **2.17.0** | https://github.com/apple/swift-protobuf/releases · https://github.com/bufbuild/protobuf-es/releases |
| connectrpc/conformance | **v1.0.7**. It tests *protocol implementations*, **not your API** | https://github.com/connectrpc/conformance |
| Idempotency-Key draft | **draft-ietf-httpapi-idempotency-key-header-09** (21 Nov 2026), still an I-D. 400 if missing, **409 while in flight, 422 on payload mismatch** | https://www.ietf.org/archive/id/draft-ietf-httpapi-idempotency-key-header-09.html |
| Stripe idempotency | Key ≤255 chars, pruned after 24 h or later. Stores the first response (including 500s) and rejects mismatched params | https://docs.stripe.com/api/idempotent_requests |
| RDS PostgreSQL | **18.6** is the latest GA minor; 19 is preview | https://docs.aws.amazon.com/AmazonRDS/latest/PostgreSQLReleaseNotes/postgresql-versions.html |
| Aurora PostgreSQL | **18.6.1 / 17.11.1** (15 Oct 2026) | https://docs.aws.amazon.com/AmazonRDS/latest/AuroraPostgreSQLReleaseNotes/AuroraPostgreSQL.Updates.html |
| RDS Multi-AZ DB cluster | 1 writer and 2 readable standbys; failover "typically under 35 s"; d-class instances only | https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/multi-az-db-clusters-concepts.html |
| Aurora HA | 6 copies across 3 AZs, up to 15 replicas; failover "often less than 30 s"; RDS Proxy cuts failover time by up to 66% | https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/Concepts.AuroraHighAvailability.html |
| RDS Proxy pinning (PG) | Pins on `SET` (except `SET LOCAL` in a transaction), SQL `PREPARE`, temp tables, cursors, `LISTEN`, `nextval/setval`, advisory locks, and statements over 16 KB | https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/rds-proxy-pinning.html |
| pgx | **v5.13.0**. Uses named prepared statements by default; `QueryExecModeExec` is recommended behind poolers | https://pkg.go.dev/github.com/jackc/pgx/v5 |
| sqlc / goose / Atlas | **1.31.1** / **3.30.1** / **1.4.0** | https://github.com/sqlc-dev/sqlc/releases · https://github.com/pressly/goose/releases · https://github.com/ariga/atlas/releases |
| ElastiCache for Valkey | GA since Oct 2024. **Valkey 9.1** is the newest on ElastiCache | https://docs.aws.amazon.com/AmazonElastiCache/latest/dg/engine-versions.html |
| Valkey pricing | Node-based is **20% cheaper** than Redis OSS; Serverless is 33% cheaper | https://aws.amazon.com/elasticache/pricing/ |
| Cluster hash tags | 16384 slots. Only the `{…}` part is hashed. Multi-key commands, transactions and Lua must use one slot (`CROSSSLOT`) | https://valkey.io/topics/cluster-spec/ |
| Serverless per-slot ceiling | **30,000 ECPU/s per slot**. Serverless takes about 10–12 min to double | https://docs.aws.amazon.com/AmazonElastiCache/latest/dg/wwe-troubleshooting.html |
| go-redis / valkey-go | **v9.24.0** / **v1.0.80** (requires Go ≥1.26) | https://github.com/redis/go-redis/releases · https://github.com/valkey-io/valkey-go/releases |
| KDS shard write | 1 MB/s or 1,000 rec/s per shard. PutRecords takes ≤500 records and 10 MiB | https://docs.aws.amazon.com/streams/latest/dev/service-sizes-and-limits.html |
| KDS on-demand | Absorbs 2× the 30-day peak; going **over 2× within 15 min throttles**. Capacity-mode switches: 2 per 24 h | https://docs.aws.amazon.com/streams/latest/dev/how-do-i-size-a-stream.html |
| KDS native delivery | Direct to S3 (raw) or Iceberg streaming tables (Parquet; needs Glue Schema Registry). On-demand streams only | https://docs.aws.amazon.com/streams/latest/dev/data-delivery.html |
| Firehose Parquet conversion | **Input must be JSON** | https://docs.aws.amazon.com/firehose/latest/dev/record-format-conversion.html |
| Firehose dynamic partitioning | 500 active partitions by default (up to 2,500) | https://docs.aws.amazon.com/firehose/latest/dev/limits.html |
| Athena partition projection | Computes partitions from table properties, so no crawlers are needed | https://docs.aws.amazon.com/athena/latest/ug/partition-projection.html |
| OTel Go | Traces and metrics stable. Logs stable in **v1.47.0**, although the docs page still says RC | https://github.com/open-telemetry/opentelemetry-go/releases |
| OTel semconv | **v1.45.0**. HTTP and DB spans stable | https://opentelemetry.io/docs/specs/semconv/ |
| ADOT collector | **v0.51.0**, run as an ECS sidecar | https://aws-otel.github.io/docs/setup/ecs |
| **Fargate quotas** | **Default On-Demand vCPU: 6.** Launch rate: burst 100, then 20 tasks/s per account; **500 tasks/min per service** (not adjustable); 5,000 tasks per service | https://docs.aws.amazon.com/AmazonECS/latest/developerguide/throttling.html · https://docs.aws.amazon.com/general/latest/gr/ecs-service.html |
| **ALB LCU reservation** | Minimum 100 LCU. Usually ready in minutes but **"can take up to a few hours"**. Can be decreased only twice a day | https://docs.aws.amazon.com/elasticloadbalancing/latest/application/capacity-unit-reservation.html |
| k6 | **v2.4.0**. About 30–40k VUs per instance | https://github.com/grafana/k6/releases · https://grafana.com/docs/k6/latest/testing-guides/running-large-tests/ |
| Distributed Load Testing on AWS | JMeter, k6 or Locust on Fargate; about 200 users per task by default | https://docs.aws.amazon.com/solutions/latest/distributed-load-testing-on-aws/solution-overview.html |
| golangci-lint | **v2.15.0** (depguard included) | https://github.com/golangci/golangci-lint/releases |

## 2. Best practices

**Go structure**
1. Use one Go module with `cmd/game` and `cmd/core`; keep all logic in `internal/`. ([go.dev layout](https://go.dev/doc/modules/layout))
2. Enforce module boundaries in CI with **go-arch-lint** (`components`/`mayDependOn`) and **depguard** deny rules. ([go-arch-lint](https://github.com/fe3dback/go-arch-lint), [golangci depguard](https://golangci-lint.run/docs/linters/configuration/))
3. Style precedence: [Google Go Style Guide](https://google.github.io/styleguide/go/), then [Effective Go](https://go.dev/doc/effective_go), then the [Uber guide](https://github.com/uber-go/guide/blob/master/style.md).

**Protobuf and Connect**
4. Use one buf v2 workspace with lint STANDARD and breaking FILE (never below WIRE_JSON). Run `buf breaking --against` main in CI.
5. Never reuse or renumber a field; add `reserved` for the number and the name. Give every enum an `*_UNSPECIFIED = 0`. Use an enum rather than a bool where the value may grow states. Use `Timestamp`/`Duration`. Put hot fields in tags 1–15. ([proto3](https://protobuf.dev/programming-guides/proto3/), [dos & don'ts](https://protobuf.dev/best-practices/dos-donts/), [style](https://protobuf.dev/programming-guides/style/))
6. Use versioned packages (`albtrivia.game.v1`). Mark read-only RPCs `idempotency_level = NO_SIDE_EFFECTS` so they can be served by cacheable GET.
7. The BSR is optional; local plugins with `buf generate` work.

**Idempotency**
8. Follow the draft's semantics (400/409/422) plus a payload fingerprint. For answers, the **natural key (show, question, user) is the real uniqueness guard**, so a retry with a new random key still cannot answer twice.
9. In Redis: `SET idem:{user}:<key> <state> NX EX <ttl>`. A nil reply means a duplicate: return the stored result, or 409 if the first request is still in flight. Clients back off exponentially with jitter. ([Stripe blog](https://stripe.com/blog/idempotency))

**Ledger**
10. Entries are append-only and every transaction sums to zero. Correct mistakes with reversal entries. ([Fowler](https://martinfowler.com/eaaDev/AccountingNarrative.html), [Square Books](https://developer.squareup.com/blog/books-an-immutable-double-entry-accounting-database-service/), [Modern Treasury](https://www.moderntreasury.com/journal/enforcing-immutability-in-your-double-entry-ledger))
11. Store money as `bigint` minor units with one currency per account. Never use float or the `money` type. ([PG docs](https://www.postgresql.org/docs/current/datatype-money.html))
12. Follow the pgledger pattern: account `balance` and `version`; each entry records previous and current balance; transfers carry idempotency keys. ([pgledger](https://github.com/pgr0ss/pgledger))
13. REVOKE UPDATE/DELETE on entries and check balance with a trigger. This is general practice, not from one primary source (UNVERIFIED).

**Valkey counters**
14. Split hot keys across slots. ([ElastiCache perf pillar](https://docs.aws.amazon.com/AmazonElastiCache/latest/dg/PerformanceEfficiencyPillar.html))
15. Shard each tally as `tally:{s<show>:q<n>:<shard>}`, shard 0..63, with a **different hash tag per shard**, and HINCRBY the option field. Readers sum the shards. *(No source was cited for the shard count. Superseded 2026-11-26 by TS-16: buckets shared with player keys; see [research 09](09-exact-tallies.md).)*
16. **Pre-aggregate in each `game` task**: in-memory counters flushed every 100–250 ms with pipelined HINCRBY. *(No source was cited. Superseded 2026-11-26 by TS-16: it loses counts on a crash and doubles them on a retry; see [research 09](09-exact-tallies.md).)*

**Kinesis**
17. Use a high-cardinality partition key (user_id), never show_id. Use PutRecords batches and retry the records that failed.
18. Use Athena partition projection on `dt/hour` and `show_id`; skip crawlers.

**Operations**
19. A scheduled action raises MinCapacity at T-20; target tracking still runs above that floor. ([App Auto Scaling](https://docs.aws.amazon.com/autoscaling/application/userguide/application-auto-scaling-scheduled-scaling.html))
20. Pin the ADOT image tag. Start the app container after the collector with `dependsOn START`.

## 3. Conflicts with Master Plan v1.1 and adopted fixes

| # | Issue | Adopted fix |
|---|---|---|
| 1 | One tally key per option takes 30k writes/s on a single slot, which hits the ceiling. | Pre-aggregate in each task, plus 32–64 shards spread across slots. Use **node-based** Valkey cluster mode, not Serverless. |
| 2 | The plan's "conformance test suite" is not the same thing as `connectrpc/conformance`. | Create a **`protocol/testvectors/`** corpus (golden binary/JSON encodings, cue signatures, error codes) run in Go, Swift and Kotlin. Also run `connectrpc/conformance` against the server. |
| 3 | **Reactions are the biggest request stream**: 2 s batches mean 150k req/s at 300k players and 300k at 600k. | Batch interval of **5 s with jitter, sent only when there are taps**. Size the ALB reservation and the fleet for reactions. Make the interval server-tunable through the join response. See [ADR-0006](../adr/0006-reaction-batching.md). |
| 4 | Firehose cannot convert protobuf to Parquet. | The analytics producer writes **JSON event envelopes** to Kinesis. Firehose converts them to Parquet with partitions `dt/hour/show_id`. |
| 5 | KDS on-demand throttles above 2× the 30-day peak within 15 min. | Use **provisioned mode with `UpdateShardCount` before the show** in the pre-show scaling job, or On-Demand Advantage warm throughput. Decide after load test. |
| 6 | The default Fargate vCPU quota is 6. | Quota request is a phase-0 task. The cap of 500 tasks/min per service fits a 20-minute pre-scale. Avoid Cloud Map on hot services. |
| 7 | ALB LCU reservation can take hours. | **Standing reservation on show days**, set at T-4 h or earlier and sized from the 600k test's `PeakLCUs`. |
| 8 | RDS Proxy pinning. | Use **UUIDv7 keys generated in the app** (no `nextval`), pgx `QueryExecModeExec`, and session settings in the proxy's init query. Alarm on `DatabaseConnectionsCurrentlySessionPinned`. |
| 9 | "RDS/Aurora" left open. | **Aurora PostgreSQL 17** (with an 18.x upgrade path) for `core`. The game hot path never waits on Postgres. |
| 10 | connect-go v2 is about 6 weeks old. | Pin v2.0.2. Fallback is v1.21.1. |
| 11 | OTel Logs status is inconsistent across docs. | Ship stdout JSON logs to CloudWatch, with the OTLP log bridge optional. |
| 12 | DLT's default of about 200 users per task means about 3,000 tasks. | Run **k6 at a high VU count per runner** (k6-operator on EKS or EC2 fleet, about 20 runners × 30k VUs). Follow the AWS simulated-events testing policy. |

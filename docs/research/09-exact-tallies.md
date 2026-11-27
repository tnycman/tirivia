# 09. Exact live tallies: counting answers with no losses and no double counts

**Question.** Up to 600k players answer within a ~10 s window. At each reveal the show needs exact per-option counts and the survivor count. A count must never be lost when a `game` task crashes, and never doubled when a write is retried after its reply was lost. This report checks the current design (SPEC §6.7; research 02 items 15–16) against primary sources. It was triggered by Codex review findings #4 and #5 (2026-11-26).

**Method.** Primary sources only: Valkey docs (read as Markdown source from `valkey-io/valkey-doc`) and source code, AWS docs, first-party engineering write-ups, papers, and our Valkey client's own source (valkey-go v1.0.80 in the Go module cache). Quotes are verbatim unless marked. **[INFERENCE]** marks our reasoning, and **UNVERIFIED** marks anything no primary source confirmed. Researched 2026-11-26.

> **Gap in earlier research.** Research 02 items 15 (64 shards) and 16 (pre-aggregate in memory, flush every 100–250 ms) cite no source. Only item 14 ("split hot keys") has one. The shard count and flush interval are therefore policy, not sourced values.

## 1. What Valkey guarantees

| Topic | Primary source says | Source |
|---|---|---|
| Script atomicity | "Valkey guarantees the script's atomic execution… all of the script's effects either have yet to happen or had already happened." | [eval-intro](https://valkey.io/topics/eval-intro/) |
| One slot per script | Scripts "are not able to run commands accessing keys of different hash slots." AWS: "Lua scripts involving multiple keys are allowed only if all the keys involved are in the same hash slot." | [lua-api](https://valkey.io/topics/lua-api/), [ElastiCache troubleshooting](https://docs.aws.amazon.com/AmazonElastiCache/latest/dg/wwe-troubleshooting.html) |
| Hash tags | "only what is between the first occurrence of `{` and the following first occurrence of `}` is hashed." | [cluster-spec](https://valkey.io/topics/cluster-spec/) |
| No rollback | "Interrupting the execution of a script has the potential of leaving the dataset with half-written changes." "Valkey does not support rollbacks of transactions." Rollback of a script that errors after a write: **UNVERIFIED** (no doc says it happens). | [programmability](https://valkey.io/topics/programmability/), [transactions](https://valkey.io/topics/transactions/) |
| Acknowledged writes can be lost | "Usually there are small windows where acknowledged writes can be lost." "Valkey Cluster does not guarantee strong consistency." | [cluster-spec](https://valkey.io/topics/cluster-spec/), [cluster-tutorial](https://valkey.io/topics/cluster-tutorial/) |
| WAIT does not fix it | "`WAIT` does not make Valkey a strongly consistent store… it is possible to still lose a write synchronously replicated to multiple replicas." | [WAIT](https://valkey.io/commands/wait/) |
| ElastiCache failover | "Valkey and Redis OSS replication is asynchronous… a small amount of data might be lost due to replication lag. For clusters with durability enabled, this limitation does not apply." | [AutoFailover](https://docs.aws.amazon.com/AmazonElastiCache/latest/dg/AutoFailover.html) |
| No idempotent increment | HINCRBY "Increments the number stored at `field`". XADD has no idempotency option in Valkey (`src/commands/xadd.json`). `XADD IDMP` exists only in Redis 8.6. Open request: "it is only safe to retry idempotent commands." | [HINCRBY](https://valkey.io/commands/hincrby/), [xadd.json](https://github.com/valkey-io/valkey/blob/unstable/src/commands/xadd.json), [Redis XADD](https://redis.io/docs/latest/commands/xadd/), [valkey#1087](https://github.com/valkey-io/valkey/issues/1087) |
| Our client does not retry writes | "All read-only commands are automatically retried on failures by default." Write commands retry only with `ToRetryable()`: "Make sure you only use this feature with idempotent operations." | valkey-go v1.0.80 `README.md` "Disable Auto Retry"; `internal/cmds/cmds.go` (`readonly = … | retryableTag`) |

**[INFERENCE]** The library never repeats a tally write on its own. The double count in Codex #5 comes from our `Tally.Flush`, which restores failed counts to memory and adds them again on the next flush.

## 2. ElastiCache durability (new since SPEC was written)

- **What it offers:** "synchronous writes designed for zero data loss at single-digit millisecond write latency, and asynchronous writes for microsecond write latency." ([durability](https://docs.aws.amazon.com/AmazonElastiCache/latest/dg/durability.html); announced [2 June 2026](https://aws.amazon.com/about-aws/whats-new/2026/06/durability-amazon-elasticache/))
- **Synchronous mode:** writes "are durably stored in the Multi-AZ transactional log before returning to clients… no acknowledged write operations will be lost in the event of a failure." **Asynchronous mode:** "up to the last 10 seconds of successful write operations may be lost." ([options](https://docs.aws.amazon.com/AmazonElastiCache/latest/dg/Durability.Options.html))
- **Consistency:** "Primary nodes are strongly consistent… preserved across primary failovers." ([consistency](https://docs.aws.amazon.com/AmazonElastiCache/latest/dg/Durability.Consistency.html))
- **Limits:**
  - Valkey 9.0 or later
  - node-based, cluster-mode-enabled only
  - Multi-AZ with at least one replica per shard
  - Graviton R/M/C families only
  - "up to 100 MBps of write throughput per primary node"
  - no Global Datastore

  ([limitations](https://docs.aws.amazon.com/AmazonElastiCache/latest/dg/Durability.Limitations.html))
- **UNVERIFIED:**
  - whether Lua/EVAL behaves differently on durable clusters
  - durability on 9.1 specifically (the docs say "9.0 or later")
- **Open-source Valkey 9.0/9.1** has no equivalent feature (release notes; [valkey#1355](https://github.com/valkey-io/valkey/issues/1355), [valkey#3037](https://github.com/valkey-io/valkey/issues/3037) are open).

**[INFERENCE]** This applies to the answer records themselves, not just the tallies. SPEC §6.7 calls Valkey state "rebuildable", from the last cue plus the Kinesis/answers stream. That stream is not built yet.

## 3. How real systems count

| System | What they do | Exact? | Source |
|---|---|---|---|
| Discord polls | Live counts "may not be perfectly accurate"; "after a poll is finished there is a background job which performs a final, accurate tally of votes" (`is_finalized`). | Live approximate; final exact by recount | [discord-api-docs poll.mdx](https://github.com/discord/discord-api-docs/blob/main/developers/resources/poll.mdx) |
| Google MillWheel | "including this unique ID for the record in the same atomic write as the state modification. If the same record is later retried… discard and ACK the duplicate". Hot keys: "build a two-phase aggregator." | Exact (revenue use) | [VLDB 2013](https://static.googleusercontent.com/media/research.google.com/en//pubs/archive/41378.pdf) |
| Google Photon | Lost events came from "absence of transactional atomicity between recording the event in the IdRegistry and recording in the output". Fixed with retry tokens plus "a verification system which takes each event in the input and checks its presence in the output." | Near-exact live, exact eventually | [SIGMOD 2013](https://static.googleusercontent.com/media/research.google.com/en//pubs/archive/41318.pdf) |
| Uber ad events | Kafka plus Flink exactly-once, "A record UUID is used as an idempotency key"; "The main bottleneck for time is the Flink checkpoint interval." | Exact at minute windows | [Uber blog](https://www.uber.com/blog/real-time-exactly-once-ad-event-processing/) |
| Disney+ Hotstar emojis | "Write the message to a local buffer and respond with success"; "data loss in rare scenarios is not a big concern". | Lossy by design (our current pattern) | [author repost](https://highscalability.com/capturing-a-billion-emo-j-i-ons/) (original blog returned 403) |
| AWS IVS polling demo | One DynamoDB item per (poll_id, requestor_id); tallies by reading the items. | Exact, demo scale | [AWS Media blog](https://aws.amazon.com/blogs/media/part-2-using-amazon-interactive-video-service-timed-metadata-2/) |
| AWS live-trivia sample | Conditional writes reject repeats. "These writes are not one atomic transaction across all players; failure recovery remains a sample limitation." | Not exact under failure | [aws-samples](https://github.com/aws-samples/sample-aws-lambda-web-functions-live-trivia) |

No first-party write-up was found for HQ Trivia, Twitch, Kahoot, Facebook Live, X or Netflix live voting (**UNVERIFIED**).

## 4. Patterns and retry guidance

- **DynamoDB, on counters:** "With an atomic counter, the updates are not idempotent… An atomic counter would not be appropriate where overcounting or undercounting can't be tolerated… it is safer to use a conditional update." ([WorkingWithItems](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/WorkingWithItems.html))
- **AWS Builders' Library:** "A timeout or failure doesn't necessarily mean that side effects haven't happened." "APIs with side effects aren't safe to retry unless they provide idempotency." ([timeouts and retries](https://builder.aws.com/content/3EumjoZascWd1oZiEgL8ORlv3qE/timeouts-retries-and-backoff-with-jitter))
- **Transactional outbox:** "Messages are guaranteed to be sent if and only if the database transaction commits"; the relay "might publish a message more than once", so consumers must be idempotent. ([microservices.io](https://microservices.io/patterns/data/transactional-outbox.html))
- **Kafka exactly-once** covers Kafka's own processing only. External sinks need the consumer to store "its offset in the same place as its output." ([Kafka design](https://kafka.apache.org/42/design/design/))
- **MillWheel on batch IDs:** a rebuilt batch can be "bit-wise distinct from its predecessor but corresponded to the same logical window". Batch IDs must be deterministic. (VLDB 2013 above)
- **CRDT counters** do not fix retries. The op-based counter's increment is a bare `i := i + 1` that assumes reliable broadcast, and a client retry is a new update. ([Shapiro et al. 2011, INRIA RR-7506](https://inria.hal.science/inria-00555588))

## 5. Options compared

| | **A. Count inside the answer script** | **B. Memory + durable log with batch IDs** | **C. Recount from answer records** |
|---|---|---|---|
| How | The intake script records the answer under its natural key and, only if new, increments the tally in the same script | Keep in-memory counts; log each batch with a deterministic ID; an idempotent consumer applies them | No live counters; recount the stored answers after the window closes |
| Task crash | Nothing lost: count and record commit together | Lost unless the log is written with the record | Nothing to lose |
| Retry after lost reply | Not doubled: the natural key guard (MillWheel, DynamoDB conditional write) | Only safe with deterministic batch IDs | Not doubled: one record per player |
| Requirement | Record and counter in one slot, so player keys move into the tally buckets | Relay, consumer and dedup store; adds latency before the reveal | O(N) read at reveal; window must be closed; latency **UNVERIFIED** |
| Precedent | MillWheel; DynamoDB conditional updates | Uber, Netflix patent, Kafka/Flink | Discord final tally; Photon verification; AWS IVS demo |

**What none of the options fixes.** Without ElastiCache synchronous durability, a Valkey failover can lose acknowledged answer records and their counts together (§2). That is a separate decision from the tally design.

**Must be measured, not assumed.**
- Lua intake throughput on the chosen node type (no official EVAL/HINCRBY benchmark exists).
- The number of buckets.
- Synchronous-durability latency under 600k load.

The AWS Performance Efficiency pillar marks a baseline load test as Required (PE1, [pillar](https://docs.aws.amazon.com/AmazonElastiCache/latest/dg/PerformanceEfficiencyPillar.html)).

## Measured locally (2026-11-26)

| What | Result | Setup |
|---|---|---|
| `live.Recount` of 600k counted answers (TS-15 load target) | **2.56 s** per run (5 runs) | `BenchmarkRecount`; one Valkey 9.1.4 node in Docker Desktop on an Intel Core Ultra 9 185H laptop (22 vCPU), 64 buckets scanned in parallel, default HSCAN `COUNT` |
| Keys of one bucket share a slot | player, tally and counted keys of a bucket all hash to slot 10130 | `CLUSTER KEYSLOT` on a cluster-enabled Valkey 9.1 node |

A local single node is not ElastiCache: re-measure on the chosen node type in the load test before setting any `COUNT` or putting the recount on the reveal's critical path.

## Still UNVERIFIED

- Internal designs of HQ Trivia, Twitch, Kahoot, Facebook Live, X and Netflix voting.
- Script behaviour after an error that follows a write.
- Lua on durable clusters.
- Valkey EVAL throughput.
- Reveal-time recount latency at 600k.
- Some quotes in §3–4 were retrieved through a summarizing fetcher. Spot-check them before they go into the decision register: the Uber blog, highscalability, microservices.io, Kafka/Confluent and the AWS Builders' Library pages.

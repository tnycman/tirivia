# 14 Testing and quality

## Role
Prove correctness on every pull request, and prove scale and resilience before each audience-ramp step (Master Plan §8 test gates).

## Tooling (pinned 2026-11-25)

| Purpose | Tool | Version | Source |
|---|---|---|---|
| Protocol | `protocol/testvectors` runners (Go, XCTest, JUnit) + `buf breaking` | buf 1.75.0 | [01-protocol](../01-protocol/) |
| Protocol (server) | connectrpc/conformance | 1.0.7 | https://github.com/connectrpc/conformance |
| Go unit/integration | `go test -race`, testcontainers for Postgres/Valkey | Go 1.27 | https://go.dev/doc/devel/release |
| Load | **k6** (k6-operator or EC2 runner fleet) | 2.4.0 | https://github.com/grafana/k6/releases |
| Chaos | **AWS Fault Injection Service** | — | https://docs.aws.amazon.com/fis/latest/userguide/fis-actions-reference.html |
| Android perf | Macrobenchmark `FrameTimingMetric`, Baseline Profile generator | AGP 9.5 | https://developer.android.com/topic/performance/baselineprofiles/overview |
| iOS perf | XCTest metrics (`XCTOSSignpostMetric`), Instruments | Xcode 27.2 | — |
| Web | vitest, Playwright | — | — |

## Gates

| Gate | When | Pass criteria |
|---|---|---|
| Unit and component | Every PR | Green, with coverage reported (not gated) |
| Protocol vectors | Every PR, all 3 platforms | Every vector passes; the vector count is asserted |
| `buf breaking` | Every PR | No breaking change against `main` |
| Load | Phase 4, then before each ramp step | **600k simulated players** through a full show with reactions on. Answer ack p99 < 300 ms, 0 lost accepted answers, no Valkey shard above 70% CPU, no KDS throttling |
| Animation | Nightly + release | Reaction layer at **55 fps or more** on the low-end Android reference phone while video plays (OPEN: pick the phone) |
| Chaos | Phase 4, then quarterly game days | Survives task loss, Valkey AZ power interruption, Aurora failover, primary encoder drop, **region failover to Ireland in 5 min or less** |
| Rehearsal shows | Weekly from the first working build | Full run of show; measures cue-to-render latency p99, which sets `L_max` |
| Penetration test | Before prizes go live | ASVS L2/L3 and MASVS scope; no open highs |
| Real devices | Each release | Device matrix including the Android reference phone and an iOS 16 device (TestFlight; Xcode 27 cannot debug iOS 16) |

## Load-test design
- **Scale.** About 20 runners at 30k VUs each (about 30–40k VUs per k6 instance; [large tests](https://grafana.com/docs/k6/latest/testing-guides/running-large-tests/)). The AWS Distributed Load Testing solution defaults to about 200 users per task, which is too few at this scale.
- **Scenario.** Lobby ramp → 12 questions (answers arrive in a burst shaped like the real timer distribution) → 5 s reaction batches → reveal → `GetShowState` storms that simulate reconnects.
- **What to record.** `PeakLCUs`, which sizes the ALB reservation. Fargate task count. Valkey per-slot CPU. KDS `WriteProvisionedThroughputExceeded`.
- **Approvals.** Follow AWS's simulated-events policy. Pre-approve with CloudFront if a test exceeds 500k rps through it.

## Sources
[docs/research/02-backend-protocol-data.md](../../docs/research/02-backend-protocol-data.md), [docs/research/03-aws-infra-devops.md](../../docs/research/03-aws-infra-devops.md)

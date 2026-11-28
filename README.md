# Trivia: live Albanian trivia show

A scheduled, live-hosted trivia show in Albanian, played in native iOS and Android apps. There are 12 questions per show, and everyone who answers all 12 correctly splits the prize. The audience is Albania, Kosovo, North Macedonia and the diaspora, and prizes are awarded and collected only in Albania.

> `trivia` is a working codename. The product name and brand are still OPEN (Master Plan §11).

## Start here

| Document | What it is |
|---|---|
| [Master Plan v1.1](Master%20Plan/Trivia-Master-Plan-v1.1.pdf) | Product source of truth (scope, locked decisions, DEFAULTs) |
| [**SPEC.md**](SPEC.md) | Technical specification. §16 lists every change from the Master Plan, each with its evidence |
| [**PLAN.md**](PLAN.md) | Build plan: long-lead items, then phases 0–8 with tasks and exit tests |
| [**stack/**](stack/) | The tech stack in 15 folders, one per function, each with pinned versions, cited best practices and codebase rules |
| [docs/research/](docs/research/) | Five research reports. Every fact links to a primary source (AWS, Apple, Google, NIST, IETF RFCs, OWASP, vendor docs) |
| [docs/adr/](docs/adr/) | Architecture decision records |
| [docs/open-items.md](docs/open-items.md) | Decisions and outside inputs still needed |
| [docs/runbooks/](docs/runbooks/) | Pre-show checklist, encoder failover, region failover |
| [docs/build-standards.md](docs/build-standards.md) | The 20-point PR acceptance bar |

## Repository layout

The Master Plan §8 calls for six repositories. Here they are six top-level folders, each self-contained so it can become its own repository.

```
.
├── protocol/        Protobuf + Connect contract, signed video cues, cross-platform test vectors
├── backend/         Go 1.27 modular monolith: cmd/game (hot path) + cmd/core
├── ios/             Swift 6 / SwiftUI, iOS 16+, SPM feature packages
├── android/         Kotlin 2.4 / Compose, minSdk 26, Now-in-Android modules
├── web-consoles/    React 19 + Vite back office (studio, question bank, claims, risk, clips) + Astro site
├── infra/           Terraform 1.17 / AWS provider 6.x, multi-account, Frankfurt + Ireland
├── design-tokens/   W3C DTCG 2025.10 tokens → Swift / Compose / CSS, WCAG 2.2 AA contrast gate
├── stack/           Tech-stack documentation, one folder per function
├── docs/            research/ adr/ runbooks/ open-items, build standards
└── .github/         CI workflows, every action SHA-pinned
```

## Architecture

```
 Camera+host → vMix/OBS → SRT (primary) / RTMPS (backup) → Amazon IVS Low-Latency (ADVANCED_HD)
                                                        │   signed timed-metadata cues:
                                                        │   question_open · reveal · pulse · burst · final
                                                        ▼
 iOS / Android apps ── play video, render cues, start timer on cue ──┐
        │  Connect RPC (protobuf over HTTPS), DPoP-bound tokens        │
        ▼                                                              │
 CloudFront + WAF + Shield → ALB → ECS Fargate (ARM64)                 │
        ├── game  (state machine, answers, reactions, studio, switches) ─► Valkey 9.1 (sharded live state)
        └── core  (auth/OTP, profile, content, wallet, risk, notify, chat-mod)
                     ├─► Aurora PostgreSQL 17 (ledger, claims, audit)
                     └─► Kinesis → Firehose (Parquet) → S3 → Athena → risk scoring
 Studio console (back office) ──BFF──► internal admin API: feature switches (chat on/off, …)
 Region: Frankfurt primary · Ireland warm standby
```

## What was verified in this workspace (2026-11-25)

| Area | Check | Result |
|---|---|---|
| protocol | `buf lint`, `buf format -d` (buf 1.75.0) | clean |
| protocol | Go codegen with connect-go v2.0.2 | builds |
| backend | `go vet`, golangci-lint v2.15.0 | 0 issues |
| backend | `go test` (Go 1.27.4) | 6 test packages pass: cue signing (worst case **845 B** vs 900 B budget), shuffle uniformity, state machine, deadlines, chat switch, back-office switch API |
| backend | Real binary, end-to-end with curl | chat OFF by default → operator switches ON → game API honours it → OFF again. Unauthenticated changes rejected, admin API not exposed on the public port, every change audited |
| database | goose 3.30.1 on Postgres 17 + `invariants.sql` | 12/12 checks behave as designed (ledger zero-sum, append-only, four-eyes, eligibility needs a counsel reference, switches) |
| web | npm install, protobuf-es codegen, `tsc` 7.0.2, `vite build` 8.3.8 | builds |
| infra | `terraform fmt`, `init`, `validate` (1.17.2, AWS provider 6.74.0) | 4/4 stacks valid |
| CI | actionlint | clean; all actions pinned to commit SHAs |
| ios / android | Swift and Kotlin code **not compiled here** (no Xcode or JDK on this Windows machine) | Phase 0 compiles them in CI |

## Changes from the Master Plan that need your confirmation

Research found facts that contradict parts of v1.1. The main ones:

1. **Standby region: Ireland instead of Milan.** Milan has no IVS and no IVS Chat.
2. ~~Minimum iOS 17~~ **Kept at iOS 16 (owner decision, ADR-0004).**
3. **Device-bound tokens and step-up before claims.**
4. **Reactions batched every 5 s.**
5. ~~Kosovo, North Macedonia and the diaspora play for points~~ **All countries eligible from launch (owner decision LG-1). The per-country kill switch remains.**

The full list with evidence is in [SPEC §16](SPEC.md#16-changes-vs-master-plan-v11).

**Chat** is fully built and testable. It is switched on and off from the back-office studio console, is **off by default**, and launching to production without it needs no code change ([ADR-0009](docs/adr/0009-chat-backend-switch.md)).

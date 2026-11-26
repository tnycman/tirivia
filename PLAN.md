# PLAN: build plan

The phases follow Master Plan v1.1 §9. Each one is broken into workstream tasks and ends with the plan's exit test. Technical detail is in [SPEC.md](SPEC.md) and decisions are in [docs/adr/](docs/adr/).

**Team assumption (Master Plan):** 2 mobile, 2 backend, 1 web/infra, plus part-time design and QA. Roughly **6 months to closed beta**. This is an estimate only.
**Workstream tags:** `IOS` `AND` `BE` `WEB` `INF` `PROTO` `VID` (studio/video) `SEC` `QA` `LEGAL` `OPS`

---

## Long-lead items: start in week 1

Each item below either takes calendar time outside our control or has a default limit that blocks a later phase. The research cites the default limits ([docs/research/](docs/research/)).

| # | Item | Default → needed | Blocks | Tag |
|---|---|---|---|---|
| L1 | **Fargate On-Demand vCPU quota** (prod + staging, Frankfurt + Ireland) | 6 vCPU → about 1,000+ (size from the load test) | Phase 4 | INF |
| L2 | **ALB reserved LCU quota** | 0 → per the load-test `PeakLCUs` | Phase 4 | INF |
| L3 | **IVS concurrent views** (Frankfurt + Ireland) | 15,000 → 600,000 | Phase 4 | VID |
| L4 | **IVS Chat connections** | 50,000 → 30% of peak (SPEC §6.10) | Phase 3 | VID |
| L5 | **Play Integrity daily requests** | 10,000/day → peak daily joins plus sign-ups (approval takes up to a week) | Phase 1 beta | AND |
| L6 | **Albanian counsel**: classification, prize tax, official rules; AMLF position | OPEN | Phase 7 | LEGAL |
| L7 | **Counsel for MK, XK and the main diaspora countries** | All countries are eligible from launch (LG-1). Counsel must confirm the MK permit + 18% fee and the XK position **before the first cash show** | First cash show | LEGAL |
| L8 | **Albanian bank partner** for payouts | OPEN | Phase 5 | OPS |
| L9 | **SMS / WhatsApp / Viber delivery and price test** in AL, XK and MK | OPEN | Phase 1 | BE |
| L10 | **KYC vendor proof of concept** (Veriff, Sumsub) with real AL/XK/MK IDs; ISO 30107-3 letters | — | Phase 5 | SEC |
| L11 | **Low-end Android reference phones**: buy a Samsung Galaxy A07 4G and a Xiaomi Redmi A5 (OP-4) | Chosen | Phase 3 | QA |
| L12 | **DPIA + EU Art. 27 representative** (GDPR, diaspora) | — | Phase 7 | LEGAL |
| L13 | Apple Developer Program, Google Play Console, Facebook App ID (Instagram Stories), TikTok client key | — | Phase 6 | OPS |
| L14 | AWS Business Support (needed for the Shield Advanced response team) | — | Phase 4 | INF |
| L15 | Studio, host and production team; launch sponsors | OPEN | Phase 2b | OPS |
| L16 | **TV broadcast partner** (if the show is simulcast): agree the feed hand-off, a **broadcast delay above the app p99 latency** (~8–10 s to start), who controls it on air, and QR/"play in the app" graphics (ADR-0010) | OPEN | Phase 2b | OPS |
| L17 | **Counsel: TV prize-game rules.** Does the Audiovisual Media Authority (AMA) regulate prize games shown on TV? (UNVERIFIED) | OPEN | Phase 7 | LEGAL |

---

## Phase 0: Setup

**Exit test:** a change flows from a pull request to `dev` automatically.

| Task | Tag |
|---|---|
| Control Tower landing zone: OUs (Security, Infrastructure, Workloads/SDLC, Workloads/Prod, Business Continuity, Sandbox, Policy Staging); accounts `log-archive`, `audit`, `network`, `shared-services`, `dev`, `staging`, `prod`, `backup-vault`; Identity Center with FIDO2 | INF |
| SCPs: region allow-list (eu-central-1, eu-west-1, us-east-1 for global services), deny leaving the org or disabling CloudTrail/GuardDuty | INF |
| `infra/bootstrap`: state bucket per account (KMS, versioning, `use_lockfile`), GitHub OIDC provider, separate plan and apply roles | INF |
| `infra/modules/network` and `live/dev/eu-central-1/network` | INF |
| CI baseline for each repo: lint, test, build, CodeQL, Scorecard, dependency review, workflow lockfile; Renovate | INF |
| `protocol/`: Buf workspace, `buf lint` and `buf breaking` in CI, generation for Go/Swift/Kotlin/TS, **test-vector generator plus runners on all three platforms** | PROTO |
| `backend/`: module skeleton, `go-arch-lint`, `golangci-lint` (depguard), health endpoints, OTel wiring, container build (ARM64) with attestation, deploy to dev on ECS | BE |
| `ios/` and `android/` shells: module structure, design-system package, CI build, Sentry | IOS AND |
| `web-consoles/` workspace: pnpm + Vite + TS7 (TS6 alias for lint), shared `ui`, `auth` (BFF) and `api-client` packages | WEB |
| **Done:** design-tokens v0 (DTCG 2025.10 → Swift/Compose/CSS via Style Dictionary 5.8), WCAG 2.2 AA contrast gate (25 pairings), light/dark/overlay, Albanian glyph rule. Brand re-skin follows LG-3 | Design |
| File long-lead items L1–L5 and L13 | INF AND VID |

## Phase 1: Foundation

**Exit test:** sign-up works on real devices in all three countries.

| Task | Tag |
|---|---|
| `auth`: challenge, App Attest server verification (9 checks plus assertions), Play Integrity standard-request decode | BE |
| `auth`: OTP pipeline (attestation → limits → line type → pumping score → send), two SMS providers with failover, WhatsApp/Viber templates, conversion-rate metric | BE |
| `auth`: EdDSA `at+jwt`, DPoP proof validation and nonces, refresh rotation with family revocation | BE |
| `profile`: users, devices, profile, push-token registration | BE |
| Aurora PG 17 in dev and staging, goose migrations, sqlc, RDS Proxy (`QueryExecModeExec`) | BE INF |
| iOS: App Attest key lifecycle, Secure Enclave DPoP key, OTP screens, in-app language picker (sq/en), String Catalogs | IOS |
| Android: Play Integrity warm-up and requests, Keystore DPoP key, OTP screens (WhatsApp zero-tap), per-app locales | AND |
| Home screen: next-show countdown (`GetNextShow`) | IOS AND |
| WAF on the auth path: ACFP, targeted Bot Control, rate rules keyed by IP and session | INF SEC |

## Phase 2a: Game loop

**Exit test:** a full 12-question game runs on test video.

| Task | Tag |
|---|---|
| IVS ADVANCED_HD channel (dev), looping test stream (ffmpeg), recording configuration | VID INF |
| `studio`: cue builder, Ed25519 signing, 900 B size guard, ≤ 4 TPS token bucket, retry on 404/429, `T_cue` anchoring | BE |
| `game`: state machine, `JoinShow`, `SubmitAnswer` (natural-key `SET NX`, server deadline, correctness on intake), sharded tallies counted inside the intake script (TS-16), `last_ok` elimination, `GetShowState` | BE |
| `JoinShow` slice 2: `playback_url` and `reaction_set` from the studio's per-show settings (needs Phase 2b show scheduling); `reaction_max_per_batch` stays unset and OPEN until the reactions design (TS-18) | BE |
| `JoinShow` slice 3: prize eligibility at join (attestation ticket, AC-5; one prize account per device, AC-6; country, LG-1) recorded in `participations` | BE |
| Valkey 9.1 cluster (dev), key design per SPEC §6.7 | BE INF |
| Question bank v1: CRUD, content limits (question ≤ 240 B, option ≤ 64 B UTF-8; `cuesign.ValidateQuestionContent`), **two-person approval**, KMS vault | WEB BE |
| Live screen on both apps: IVS player, cue parsing and verification, seq handling, timer from cue render, per-player shuffle, answer bars, resync after a gap | IOS AND |
| **Verify the UNVERIFIED IVS behaviours**: cues across rendition switches, late join, measured cue-to-render latency (sets `L_max`) | VID QA |
| Verify whether IVS cues are carried in the **audio-only rendition** (decides the TV companion mode implementation, ADR-0010) | VID QA |

## Phase 2b: Studio

**Exit test:** an internal rehearsal show runs end to end.

| Task | Tag |
|---|---|
| Studio console: **show scheduling (producers set day/time; weekly show)**, run of show, question release, reveal, burst, **feature switches (chat, reactions, lives, sign-ups, countries)**, rehearsal mode | WEB BE |
| Reveal runs the tally check `live.Verify` after the answer window closes (TS-16), and only once every accepted answer request has finished: an answer that passed the deadline check can still be running its script, so Verify waits at least the server `WriteTimeout` (15 s, stack/03 "Server timeouts") past the deadline, or uses a closing barrier (Codex review of 4584663). The action on a mismatch (alert only, or publish the recount) is decided before this task | BE |
| Encoder chain: vMix/OBS → SRT primary (`priority=1`) and RTMPS backup (`priority=2`), identical profiles, failover slate, reconnect window | VID |
| **TV simulcast path** (if a TV partner is confirmed): studio feed hand-off to the station; broadcast delay set above the app p99 latency; QR/"play in the app" graphics in vMix; delay value recorded in the run-of-show settings | VID OPS |
| Runbooks: pre-show checklist, encoder failover, cue desync | VID OPS |

## Phase 3: Live engagement

**Exit test:** reactions and counts are live in a rehearsal show.

| Task | Tag |
|---|---|
| `engage`: reaction intake (5 s batches, per-user cap, paused while a question is open), sharded counters, server-side viewer count, `pulse` builder | BE |
| Burst moments with a city/country meter | BE IOS AND |
| Reaction layer: iOS (SwiftUI Canvas/TimelineView), Android (single Canvas, particle pool, pre-rasterised bitmaps), sampled at ≤ 30 per second | IOS AND |
| **TV companion mode** on both apps: a "Watching on TV" toggle mutes and minimises the video (or uses the audio-only rendition) while still receiving cues | IOS AND |
| **Chat**: `chatmod` token minting, room sharding, review Lambda (200 ms timeout, provisioned concurrency, sq/en lists), slow mode, **backend switch delivered at join and live via `pulse.chat_enabled`**, chat panel with lazy connection on both apps | BE IOS AND |
| Macrobenchmark frame-timing gate on the reference phone; Baseline Profiles | AND QA |

## Phase 4: Scale and safety

**Exit test:** 600k simulated players complete a full show.

| Task | Tag |
|---|---|
| Staging at production shape. **Pre-show orchestrator**: when a producer schedules or reschedules a show, it books one-time actions from the show start: LCU reservation at T-4 h; MinCapacity, KDS re-shard, Valkey scale-out and Ireland warm-up at T-60; scale-down after SETTLED | INF BE |
| k6 load suite: players (join → answers → reactions → state), about 20 runners × 30k VUs; record `PeakLCUs`, per-slot Valkey CPU, answer p99 | QA BE |
| `risk`: Kinesis consumer, features, scoring, clustering, review queue | BE |
| Kill switches end to end | BE |
| Dashboards and alarms (SPEC §11), on-call rota | INF OPS |
| FIS experiments: task loss, Valkey AZ power interruption, Aurora failover, encoder drop, **region failover to Ireland within 5 min** | INF QA |
| Standby region: Ireland IVS channel and chat, warm Fargate, Aurora Global secondary, Secrets replicas, ARC routing controls | INF |
| Shield Advanced enrolment | INF |

## Phase 5: Prizes

**Exit test:** a test prize is claimed through each collection option.

| Task | Tag |
|---|---|
| `wallet`: ledger (append-only, zero-sum trigger, idempotent transfers), prize credit after risk review, claims, expiry and reminders, €20 minimum | BE |
| KYC integration (selected vendor), tiers A/B/C, video verification scheduling, KYC re-check after a number rebind | BE SEC |
| KYC storage: CMK, Object Lock (governance), Macie, retention lifecycle | INF SEC |
| Claims console and risk console with four-eyes approval on payout release and on unblocking | WEB |
| Wallet and claim card UI with step-up (biometric unlock and a DPoP nonce) | IOS AND |

## Phase 6: Growth features

**Exit test:** a feature-complete build.

| Task | Tag |
|---|---|
| Extra lives and referrals (granted after the invitee verifies and completes a game) | BE IOS AND |
| Leaderboards (country, city, diaspora vs homeland) | BE IOS AND |
| Share cards: server rendering and native share (Instagram Stories with the FB App ID, TikTok Share Kit, WhatsApp via the share sheet) | BE IOS AND |
| `notify`: APNs (time-sensitive reminders, Live Activity push-to-start), FCM HTTP v1 to tokens | BE IOS AND |
| Public website (Astro: sq/en/mk, schedule, countdown, winners, store links) | WEB |

## Phase 7: Launch readiness

**Exit test:** the go/no-go checklist is signed.

Penetration test (ASVS L2/L3 scope), a **TV simulcast rehearsal** if a TV partner is confirmed (measure the TV delay per platform: terrestrial, cable, IPTV; confirm TV always trails the app p99), store review (official rules and the "Apple is not a sponsor" statement in the app), legal sign-off (L6, L7, L12), privacy policy and consents, the country eligibility table set (all countries eligible per LG-1, with any country moved to `play_for_points` if counsel advises), and a **chat go/no-go**: decide whether the production chat switch goes on. Chat is fully built and tested by then but defaults to OFF, so launching without it needs no code change (ADR-0009).

## Phase 8: Post-launch

**Exit test:** shipped during the audience ramp.

Replays and clips (EventBridge Recording End → clip pipeline), winners wall (with opt-out), question of the day, friends in the game.

### Audience ramp (after Phase 7)

| Stage | Cap | Prizes |
|---|---|---|
| Closed beta | 1,000 | Points |
| Soft launch | 10,000 | Small sponsor prizes |
| Ramp | 50,000, then 150,000 | Cash |
| Open | None | Full |

A cap rises only after a clean run at the previous one. IVS, chat and Play Integrity quotas must be confirmed **before** each step.

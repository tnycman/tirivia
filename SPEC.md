# SPEC: Tirivia Live Trivia (technical specification)

| | |
|---|---|
| **Status** | Draft 1, 2026-11-25 |
| **Source of truth for product scope** | [Master Plan v1.1](Master%20Plan/Trivia-Master-Plan-v1.1.pdf) |
| **Research basis** | [`docs/research/`](docs/research/): five domain reports, every fact cited to a primary source |
| **Decisions** | [`docs/adr/`](docs/adr/) |
| **Codename** | `tirivia` (the product name and brand are still OPEN, Master Plan §11) |

**Markers used in this document:**

- **DEFAULT**: a proposed value carried over from the Master Plan; the owner can change it.
- **OPEN**: needs outside input.
- **CHANGED**: differs from Master Plan v1.1 because research contradicted it. Every CHANGED item is listed in [§16](#16-changes-vs-master-plan-v11) with its evidence.
- **NEW**: missing from the plan; added because research showed it was required.

---

## 1. Product

A scheduled, live-hosted trivia show in Albanian, played on native iOS and Android apps. Players answer 12 questions, and everyone who gets all 12 right splits the prize.

**Locked decisions (Master Plan §1):**

- A live host on camera.
- Cash and sponsor prizes.
- Prizes are offered, awarded and collected **only in Albania**.
- Audience: Albania, Kosovo, North Macedonia and the diaspora.
- Fully native apps.
- Entry is always free, and extra lives are never sold while prizes are on offer.

### 1.1 Show format (DEFAULT)

| Item | Value |
|---|---|
| Schedule | **One show per week** (owner decision 2026-11-25, replacing the Master Plan daily + weekend default). **Producers set the day and time of each show in the studio console.** Nothing in code or infrastructure fixes a time slot. Pre-show scaling is booked automatically from each show's start time (§10) |
| Length | 12 questions, 3 options each, about 15–20 min |
| Timer | Q1–4: 10 s. Q5–8: 8 s. Q9–12: 6–7 s (7 s used until rehearsal data says otherwise). The timer starts when the full question is on screen, i.e. when the cue renders in the player's own video |
| Extra life | One per show; not usable on Q12 |
| Winner selection | **Skill only.** Everyone who answers all 12 correctly splits the pot. **No random tie-breaks or draws** (NEW, see [research 05 §3](docs/research/05-auth-security-compliance-web.md)) |

### 1.2 Release scope

| Release | Features |
|---|---|
| **Launch** | Live show with overlay, timer and results. Viewer count, players remaining, answer bars. Floating reactions with a running total. Phone sign-in, profile and stats. Extra lives and referrals. Country, city and diaspora-vs-homeland leaderboards. **Moderated chat with slow mode: built and testable, toggled from the back office, default OFF. Using it in production is a go/no-go decision (§6.10).** Share cards (Instagram, TikTok, WhatsApp). Push reminders. Wallet, claim card and collection options. Sponsor shows and prize drops. Albanian first, English second. Light website |
| **Right after launch** | Replays and clips with views and likes. Winners wall. Question of the day. Friends in the game |
| **Later** | Themed shows. Private rooms (no prizes). Teams, leagues, tournaments, streak rewards |

---

## 2. Social and engagement layer

### 2.1 During the show

| Element | Behaviour | Mechanism |
|---|---|---|
| Viewer count | "128.4K watching" | Counted on the server from joins and stream sessions; carried in `pulse` |
| Players remaining | Shown after each question | `reveal.players_remaining` and `pulse` |
| Floating reactions | A tap or double-tap sends a reaction that floats up and fades. Everyone sees everyone's reactions | Client batches → `SubmitReactions` → Valkey → totals in `pulse` → each phone renders a sample of at most ~30 per second |
| Reaction total | Running count next to the viewer count | `pulse.reaction_totals` |
| Answer bars | Share of players per option, shown at reveal | `reveal.option_counts` |
| Burst moments | The host starts one ("send hearts for Prishtina") and a live meter shows by city or country | `burst_start` / `burst_end` cues |
| Themed reactions | Icon set per show | `JoinShowResponse.reaction_set` |

### 2.2 Between shows

Replays and clips (view and like counts), winners wall (with opt-out), question of the day, share cards, profile stats.

### 2.3 Rules

1. Reactions, likes and views are **cosmetic**. They never earn prizes, lives or rank.
2. Displayed counts are real.
3. **Reactions pause while a question is open.**
4. Each user has a reaction rate cap enforced on the server.
5. A user can like an item once.

---

## 3. Prizes and claims

### 3.1 Claim flow

1. After the show, `risk` reviews the result and the server confirms the winners.
2. Each winner gets a push notification and a claim card showing the amount and the deadline.
3. The winner is verified according to their tier (§3.2).
4. The winner picks how to collect:
   - bank transfer to an Albanian account in their own name
   - in person at the office
   - through a representative holding a notarised power of attorney
   - as a donation to a listed charity
   - for sponsor prizes, pickup or delivery within Albania
5. The claim window is **60 days**, with reminders at 30, 7 and 1 days before it closes. Unclaimed prizes go back into a future pot. **If nobody answers all 12 correctly, the pot rolls over to the next show (PR-3).**
6. The cash-out minimum is **€20** (DEFAULT). Smaller balances accumulate, and the claim window starts once the minimum is reached.
7. **Step-up before claim (NEW):**
   - The player must unlock their device-bound key with biometrics.
   - The request must carry a fresh DPoP proof with a server nonce ([ADR-0005](docs/adr/0005-device-bound-tokens.md)).
   - **If the player's phone number was re-bound in the last 30 days (DEFAULT), KYC is re-checked**, because no SIM-swap data is available for AL, XK or MK.

### 3.2 Verification tiers (DEFAULT thresholds)

| Tier | Prize | Verification |
|---|---|---|
| A | under €50 | Clear risk score, verified phone, credited to the wallet; ID checked at cash-out |
| B | €50–500 | Government ID, selfie match (ISO/IEC 30107-3 PAD-tested vendor), hold period, manual review |
| C | over €500, or any flagged score | Tier B plus live video verification with staff, including 2–3 fresh questions on the same timer, within 48 h. Jackpot finalists can be verified on air |

KYC vendor: run a proof of concept with **Veriff and Sumsub** using real AL, XK and MK documents. Entrust Instant covers only the Albanian *passport* ([research 05](docs/research/05-auth-security-compliance-web.md)).

---

## 4. Legal and compliance controls (not legal advice)

| Control | Spec |
|---|---|
| Governing law | Official rules governed by Albanian law and courts. Prizes awarded and collected in Albania |
| Country eligibility switch | Each country is set to `eligible` / `play_for_points` / `blocked` on the backend; changes need no app release and every change carries a `decision_ref`. **Launch: all countries `eligible` (owner decision LG-1, 2026-11-25).** The switch is the kill switch if a regulator or counsel raises an issue. **Open risk for counsel before the first cash show:** North Macedonia's giveaway permit and 18% fee (in force 14 Jul 2026), and Kosovo's lack of a legal basis for prize games ([research 05 §3](docs/research/05-auth-security-compliance-web.md)) |
| Age | 18+ only |
| Store rules | Official rules in the app; statement that Apple and Google are not sponsors (Apple 5.3.2); no IAP linked to prizes (Apple 5.3.3) |
| Privacy | Albanian Law 124/2024 (GDPR-aligned) applies. **GDPR Art. 3(2) applies to EU-resident diaspora**, so an Art. 27 EU representative and a **DPIA done now** are required (NEW). Consent covers ID documents, verification recordings, and winners-wall/replay display |
| OPEN | AL counsel opinion and AMLF position. Prize tax and withholding. Counsel views for MK, XK and the main diaspora countries |

---

## 5. Technology stack

Each row links to its `stack/` folder, which holds the pinned versions, best practices and sources.

| Layer | Choice | Folder |
|---|---|---|
| Protocol | Protobuf + **Connect RPC** over HTTPS; Buf v2 workspace; shared test vectors | [`stack/01-protocol`](stack/01-protocol/) |
| Live video | Amazon IVS Low-Latency, **ADVANCED_HD** channel (720p top, server-side ABR). Questions and counts sent as signed timed metadata | [`stack/02-live-video`](stack/02-live-video/) |
| Backend | Go 1.27, modular monolith, deployables `game` and `core` | [`stack/03-backend-go`](stack/03-backend-go/) |
| Data | Aurora PostgreSQL 17 (Multi-AZ), ElastiCache **Valkey 9.1** (node-based, cluster mode), Kinesis → Firehose (JSON → Parquet) → S3 → Athena | [`stack/04-data`](stack/04-data/) |
| Auth | Own OTP service, two SMS providers + WhatsApp/Viber, App Attest / Play Integrity, **EdDSA access JWT + rotating refresh + DPoP device binding** | [`stack/05-auth-identity`](stack/05-auth-identity/) |
| iOS | Swift 6 mode (Xcode 27.2), SwiftUI with `ObservableObject`, **minimum iOS 16.0** (owner decision; research recommended 17), IVS Player 1.58.1, App Attest | [`stack/06-ios`](stack/06-ios/) |
| Android | Kotlin 2.4, Compose (BOM 2026.11.01, M3), Hilt, **minSdk 26 / targetSdk 36 / compileSdk 37**, IVS Player 1.58.1, Play Integrity standard requests | [`stack/07-android`](stack/07-android/) |
| Web consoles | React 19.3 + TypeScript 7 + Vite 8 SPAs behind a BFF; Astro 7 public site | [`stack/08-web-consoles`](stack/08-web-consoles/) |
| Infrastructure | AWS (Control Tower multi-account), Terraform 1.17 + AWS provider 6.x, ECS Fargate (ARM64), CloudFront + WAF + Shield Advanced | [`stack/09-infrastructure-aws`](stack/09-infrastructure-aws/) |
| CI/CD | GitHub Actions with OIDC, workflow lockfile / SHA pins, CodeQL, Scorecard, Renovate, artifact attestations | [`stack/10-ci-cd`](stack/10-ci-cd/) |
| Observability | OpenTelemetry (Go SDK 1.47), ADOT collector, CloudWatch + Managed Prometheus/Grafana, Sentry for mobile | [`stack/11-observability`](stack/11-observability/) |
| Security | OWASP ASVS 5.0 L2 (L3 for wallet, KYC, vault, studio), MASVS 2.1, API Top 10 2023 | [`stack/12-security`](stack/12-security/) |
| Anti-cheat | Four layers: question design, client integrity, risk scoring, winner verification | [`stack/13-anti-cheat`](stack/13-anti-cheat/) |
| Testing | Test vectors, k6 at 600k, FIS chaos, Macrobenchmark, rehearsals | [`stack/14-testing-quality`](stack/14-testing-quality/) |
| Compliance | Law 124/2024, GDPR, store policies, gaming law inputs | [`stack/15-compliance-privacy`](stack/15-compliance-privacy/) |
| Regions | **Frankfurt (eu-central-1) primary, Ireland (eu-west-1) standby** (CHANGED from Milan) | [ADR-0003](docs/adr/0003-standby-region-ireland.md) |

---

## 6. System design

### 6.1 Deployables and modules

Everything lives in one Go module with two binaries. The modules are packages under `backend/internal/`, and their boundaries are enforced by `go-arch-lint` and `depguard`.

| Module | Responsibility | Deployable |
|---|---|---|
| `game` | Show state machine, answer intake, tallies, elimination, extra lives, winners | **game** |
| `studio` | Producer commands, question release from the vault, cue signing and `PutMetadata`, bursts, rehearsal mode, **feature switches** | **game** |
| `engage` | Reaction intake and aggregation, viewer count, `pulse` content | **game** |
| `auth` | Attestation, OTP, sessions, DPoP, device binding | **core** |
| `profile` | Users, stats, friends, referrals, leaderboards, push tokens | **core** |
| `content` | Replays, clips, winners wall, QOTD, likes and views, share cards | **core** |
| `wallet` | Double-entry ledger, prize credit, claims, expiry | **core** |
| `risk` | Fraud scoring from the event stream, clustering, review queue | **core** |
| `notify` | Push scheduling and delivery (APNs HTTP/2, FCM HTTP v1) | **core** |
| `chatmod` | Chat token minting, room sharding, IVS Chat review Lambda | **core** (+ Lambda) |

### 6.2 Show state machine

```
SCHEDULED → LOBBY → QUESTION_OPEN → QUESTION_CLOSED → REVEAL ─┬→ QUESTION_OPEN (next)
                                                             └→ FINAL → SETTLED
```

1. Only `studio` can change the state. Each transition is written to `audit_log` with the operator, server time, cue `seq` and the IVS stream position.
2. A transition is accepted only from the state listed before it in the diagram. Anything else returns `FAILED_PRECONDITION`.
3. `LOBBY` opens at T-20 min (DEFAULT). That is also when chat tokens can start being minted (§6.10).
4. In rehearsal mode the same machine runs, but on a separate show flagged `rehearsal=true`: no prizes, and no ledger writes.

### 6.3 Question round

1. **Release.** The producer releases question *n*.
   - `studio` decrypts that one question from the vault (KMS, per-question data key).
   - It writes the correct `option_id` into the show-state hash, which only `game` tasks can read.
   - It emits a **signed** `question_open` cue through `PutMetadata`.
   - **The correct answer is never sent to clients.** Reactions pause.
2. **Anchor.** `studio` records `T_cue`, the moment `PutMetadata` returned success.
   - IVS gives no guarantee about when a cue is inserted ([research 01](docs/research/01-live-video-ivs.md)), so **every deadline is set on the server**.
   - If the call fails with `404 ChannelNotBroadcasting` or `429`, it retries with backoff under the token bucket (≤4 TPS per channel).
3. **Client.** When the cue renders, the app starts the timer and shuffles the options. It then submits `SubmitAnswer` with:
   - an `Idempotency-Key` header
   - the `option_id`
   - `client_elapsed_ms`
   - integrity signals (focus loss, capture, overlay)

   It retries with jittered backoff until it gets an acknowledgement or the window closes.
4. **Acceptance.** The server accepts the answer if all three hold:
   - it arrives at or before `T_cue + L_max + D_q + U_grace`
     - `D_q`: the question timer
     - `L_max`: allowance for stream latency, DEFAULT 6 s; tuned from rehearsal data on p99 cue-to-render latency
     - `U_grace`: allowance for upload, DEFAULT 1.5 s
   - the player is alive
   - no answer exists yet for (show, question, user)

   Correctness is computed on intake, because the server already holds the answer. The result goes to Valkey and Kinesis.
5. **Close and reveal.**
   - `question_closed` is sent at `T_cue + L_max + D_q`.
   - `reveal` follows with the correct option, per-option counts and players remaining.
   - Reactions resume.
   - Results are persisted to Postgres asynchronously, after the round.
6. **Focus-loss rule (AC-3, settled).**
   - On Q10–12, focus loss voids the answer only on `.background` on iOS or `onStop` on Android, or on an `inactive` gap longer than 1.5 s.
   - Shorter `inactive` events, such as a pulled-down notification shade, are only risk signals.
   - The app reports `went_background` and `longest_inactive_ms` in `IntegritySignals`. A voided answer is stored as no option, so the question counts as missed (`domain.FocusLossVoids`).
   - Before the reveal, a right and a wrong answer show the same player status, so `GetShowState` never leaks correctness (tested).
   - The rule is explained before every show.
7. **Per-player shuffle.**
   - `shuffle_seed` is 32 random bytes from a cryptographic generator, created at the player's first join and returned unchanged on every rejoin. RFC 2104 §2 recommends an HMAC key of at least L bytes (the hash output, 32 for SHA-256 per RFC 6234 §6.2), and §3 requires random keys.
   - `k = HMAC-SHA256(key = shuffle_seed, msg = question_id)`.
   - For `i` from 2 down to 1: `j = k[2 - i] mod (i + 1)`; swap `options[i]` and `options[j]`.
   - The client always submits an `option_id`, never a position.
   - Vectors are in `protocol/testvectors/answer/`.

**Elimination without a scan.**

- Each player has a hash `player:{show}:{user}` with `last_ok` = the index of the last question they passed.
- **Prize eligibility** is fixed at join (`participations.prize_eligible`). It is false if the device failed attestation (AC-5), if another account already holds this device's prize binding (AC-6), or if the country is not `eligible`. Ineligible players play normally but for points.
- **Late joiners (AC-8):** a player who joins after Q1 opened gets `PLAY_ALONG` status. They answer from the next question on, for fun and stats, are excluded from tallies and players remaining, and can never win.
- An on-time player is alive for question *q* iff `last_ok == q - 1`.
- A correct answer sets `last_ok = q`. An extra life also sets `last_ok = q`. The player taps to use it during the `REVEAL` window: once per show, never on Q12, covering a wrong or unanswered question, but **only for a question the player was present for when it opened (UF-4)**. Rules are in `backend/internal/game/domain` (tested).
- **Players remaining after *q*** = the summed tally for the correct option, plus lives used on *q*.
- **Winners** are the players with `last_ok == 12` once the show reaches `FINAL`.

### 6.4 Timed-metadata cues

- **Schema:** [`protocol/proto/tirivia/cue/v1/cue.proto`](protocol/proto/tirivia/cue/v1/cue.proto).
- **Wire format:** `"t1." + base64url(SignedCue)`, where `SignedCue = {body, Ed25519 signature, key_id}`.

| Rule | Value | Why |
|---|---|---|
| Max text size | **900 bytes** (CI-tested at the maximum content length) | Hard IVS limit is 1 KB, and the signature counts toward it |
| Rate | ≤ 4 TPS per channel, one shared token bucket; burst events coalesced | Hard limit is 5 TPS per channel |
| Question content limits | Question ≤ **240 UTF-8 bytes** (about 140–200 Albanian characters), options ≤ **64 UTF-8 bytes**, enforced when a question is saved | The worst case (every character a 2-byte `ë`/`ç`) measures **845 B** in `cuesign_test.go`. A 140/40 *character* limit would have measured about 958 B, over budget |
| Ordering | `seq` strictly increasing; duplicates ignored; a gap triggers `GetShowState` | Delivery is best-effort and late joiners miss earlier cues |
| Late join | `pulse` every 3–5 s carries phase, question index, `remaining_ms`, counts, and the reaction and chat switches | Lets a player resynchronise without a stored cue |
| Replay | VOD playback ignores gameplay cues | IVS keeps cues in recordings |
| Signature | Apps reject unsigned cues, cues with unknown keys, and cues for another show. Keys: current plus next, delivered at join | Prevents forged cues |

### 6.5 API (Connect RPC)

- **Transport:** Connect protocol, `POST` with `application/proto` over HTTPS.
- **Read-only calls:** `GET` (Connect `NO_SIDE_EFFECTS`).
- **Auth:** every authenticated call carries `Authorization: DPoP <jwt>` and a `DPoP` proof header.

| Master Plan call | RPC | Deployable |
|---|---|---|
| `POST /auth/otp/start`, `/verify` | `AuthService.StartOtp`, `VerifyOtp` | core |
| `POST /auth/attest` | `AuthService.GetAttestationChallenge`, `Attest` | core |
| (NEW) token refresh | `AuthService.RefreshSession` (rotation and reuse detection) | core |
| `GET /shows/next` | `GameService.GetNextShow` | game |
| `POST /shows/{id}/join` | `GameService.JoinShow` | game |
| `POST /shows/{id}/answers` | `GameService.SubmitAnswer` | game |
| `POST /shows/{id}/reactions` | `GameService.SubmitReactions` | game |
| `GET /shows/{id}/state` | `GameService.GetShowState` | game |
| `POST /shows/{id}/extra-life` | `GameService.UseExtraLife` | game |
| (NEW) chat token after a mid-show switch-on | `GameService.GetChatToken` | game |
| (NEW) back-office feature switches (chat, reactions, extra lives, sign-ups) | `admin.v1.SwitchService.ListSwitches`, `SetSwitch`: **internal admin listener only**, operator SSO via the BFF, reason required, audited | game (admin port) |
| `GET /feed` | `ContentService.GetFeed` | core |
| `POST /items/{id}/like`, `/view` | `ContentService.LikeItem`, `RecordView` | core |
| `GET /share-cards/{id}` | `ContentService.GetShareCard` | core |
| `GET /wallet`, `POST /claims` | `WalletService.GetWallet`, `SubmitClaim` | core |
| `GET /leaderboards` | `ProfileService.GetLeaderboard` | core |

**Idempotency** follows `draft-ietf-httpapi-idempotency-key-header-09`:

- missing key → 400
- request still in flight → 409
- key reused with a different payload → 422

For answers, the natural key (show, question, user) is the final guard against double answers.

### 6.6 Data model (Aurora PostgreSQL)

**General rules:**

- Primary keys are app-generated **UUIDv7**. There is no `serial` or `nextval`, which avoids RDS Proxy session pinning.
- Money is `bigint` minor units plus an ISO 4217 currency code. Never float or `money`.
- All timestamps are `timestamptz`.

| Table | Key columns / notes |
|---|---|
| `users` | id, phone_e164_hash, phone_enc, display_name, city, country_code, is_diaspora, locale, birth_year_verified, status, created_at |
| `devices` | id, user_id, platform, attest_key_id, dpop_jkt, push_token(s), first_seen, last_seen, risk_metric |
| `sessions` | id, user_id, device_id, refresh_family_id, refresh_hash, rotated_at, revoked_at |
| `shows` | id, starts_at, state, sponsor_id, prize_plan_id, rehearsal, ivs_channel_arn, region |
| `questions` | id, type (general / show_dependent / audio / visual / sequence), difficulty, ciphertext, data_key_enc, author_id, approved_by[2], status |
| `show_questions` | show_id, index (1–12), question_id, timer_ms, released_at, t_cue, cue_seq |
| `participations` | show_id, user_id, joined_at, last_ok, extra_life_used_q, outcome |
| `answers` | (show_id, question_id, user_id) **PK**, option_id, correct, server_elapsed_ms, client_elapsed_ms, signals jsonb, accepted |
| `extra_lives` | user_id, source (referral / …), granted_at, used_show_id |
| `referrals` | referrer_id, referee_id, status (pending → verified → completed_game), granted_at |
| `friendships` | user_id, friend_id, created_at |
| `reaction_sets` | id, show_id, kinds[], themed_icon_url |
| `show_stats` | show_id, peak_viewers, players_per_q[], reactions_total, … |
| `clips` | id, show_id, s3_key, duration, views, likes |
| `item_likes` | (item_id, user_id) **PK** (one like per account) |
| `item_views` | item_id, user_id, viewed_at (aggregated into counters) |
| `winners_wall` | show_id, user_id, first_name, city, amount_minor, opted_out |
| `ledger_accounts` | id, owner, currency, balance, version, allow_negative |
| `ledger_transactions` | id, idempotency_key **unique**, kind, created_at |
| `ledger_entries` | id, transaction_id, account_id, amount_minor, prev_balance, new_balance, account_version. **Append-only** (UPDATE/DELETE revoked; trigger checks that each transaction sums to zero) |
| `claims` | id, user_id, amount_minor, tier, deadline, status, option, iban_enc |
| `kyc_records` | id, user_id, vendor, vendor_ref, result, liveness_score, s3_key_enc, retention_until |
| `risk_flags` | id, user_id, show_id, score, verdict (clear / review / block), reasons[] |
| `audit_log` | id, actor, action, entity, before, after, approver_id, ip, device, request_id, prev_hash, hash (**hash-chained**) |
| `sponsors`, `prize_plans` | Sponsor branding; pot, split rules, sponsor prizes |
| `feature_switches` (NEW) | key, scope (global / show / country), value, changed_by, changed_at (audited) |
| `country_eligibility` (NEW) | country_code, eligibility, decision_ref (owner decision or counsel opinion; required), changed_by, changed_at |

### 6.7 Live state (Valkey 9.1, cluster mode)

| Key | Type | Notes |
|---|---|---|
| `show:{<show>}:state` | hash | phase, q_index, t_cue, deadline, correct_option (game-only), cue_seq |
| `player:{<show>:<b>}:<user>` | hash | joined_at_q, last_ok, life_used, life_q, and `a<q>` = the answer to question q. **The answer field is the natural-key guard** (`HEXISTS` then `HSET` in one script). `b` = FNV-1a(user) mod 64 is the player's bucket; the player key shares its hash tag with that bucket's tally and counted keys, so the intake script touches one slot. Valkey cluster rejects scripts that touch several slots ([cluster spec](https://valkey.io/topics/cluster-spec/)) |
| `idem:<user>:<key>` | string | Idempotency-Key record (state plus response hash) |
| `tally:{<show>:<b>}:<q>` | hash | option → count, plus `_life`. **The intake script increments it in the same atomic step as a new answer**, so a crash cannot lose a count and a retry cannot double one (TS-16, [research 09](docs/research/09-exact-tallies.md)). 64 buckets with different hash tags; readers sum the buckets. The bucket count is a policy value proven in the 600k load test |
| `counted:{<show>:<b>}:<q>` | hash | user → option for every answer the tally counted. After the reveal, a cursored `HSCAN` recount is checked against the tally (`live.Verify`) |
| `react:{<show>:<shard>}` | hash | kind → count, sharded the same way |
| `viewers:<show>` | HyperLogLog or a per-task counter | Viewer count on the server |
| `rl:<user>:<bucket>` | string | Per-user rate limits for answers, reactions and OTP |
| `switch:*` | string | Feature-switch cache (TTL 5 s; the source of truth is Postgres) |

Live-state scripts validate key types before their first write, because Valkey does not roll a script back. A change to the live-state key format is deployed only between shows: readers do not fall back to an older format. Asynchronous replication can lose acknowledged writes on failover ([cluster spec](https://valkey.io/topics/cluster-spec/)); ElastiCache synchronous durability is being evaluated in the 600k load test (TS-17). Valkey holds state that can be rebuilt. If a node is lost, the state is recovered from the last cue plus the Kinesis/answers stream, which is acceptable because Postgres persists results after every round.

### 6.8 Event stream

- **Envelope:** `events.v1.Event`, written as **canonical proto3 JSON**. Firehose's Parquet conversion needs JSON input.
- **Partition key:** `user_id`.
- **Path:** KDS → Firehose (Parquet, partitioned by `dt/hour/show_id`) → S3 → Athena with partition projection.
- **Capacity:** KDS runs in **provisioned mode, re-sharded by the pre-show scaling job**, or on-demand with warm throughput. The choice is made after the load test, because on-demand throttles above 2× its 30-day peak.
- `risk` consumes the stream directly.

### 6.9 Video and studio chain

```
Camera + host ─► vMix / OBS (graphics) ─┬► primary encoder  ── SRT  :9000  priority=1 ─┐
                                        └► backup encoder   ── RTMPS :443  priority=2 ─┴► IVS ADVANCED_HD channel (Frankfurt)
                                                                                          ├► auto-record → S3 → CloudFront (replays, clips)
                                                                                          └► standby channel pre-created in Ireland
```

- **Encoder profile** (both encoders must match exactly for stream takeover):
  - H.264 Main, 720p30, CBR, 2 s keyframe interval
  - B-frames 0, scene-change detection off
  - AAC-LC 48 kHz
- **Failover:** Low-Latency has no automatic ingest failover. Instead:
  - stream takeover by priority
  - the failover slate
  - `recordingReconnectWindowSeconds` = 90 (DEFAULT)
  - Runbooks: [`docs/runbooks/encoder-failover.md`](docs/runbooks/encoder-failover.md).
- **Clips:** cut from byte-range HLS playlists, or produced with MediaConvert after the EventBridge "Recording End" event.

### 6.9a TV simulcast ([ADR-0010](docs/adr/0010-tv-simulcast.md))

The show may also air on a TV channel. The rule is: **the app is the game, and TV runs deliberately behind it.**

```
                 ┌─► Amazon IVS ──► app video (2–5 s behind live, carries the cues) ─► players answer here
Studio feed ─────┤
                 └─► TV station broadcast delay (≥ app p99 latency, ~8–10 s to start) ─► TV viewers watch only
```

| Item | Spec |
|---|---|
| Where you play | Only in the app. Questions, timers, answers and reveals ride the IVS stream. Nothing depends on the TV signal |
| TV delay | The TV feed is delayed **above the app's measured p99 latency** (~8–10 s to start), so TV never shows a question before an app player's timer starts. The value is tuned from rehearsal data and stored in the run-of-show settings. The TV station controls it on air |
| Reveal safety | The host reveals only after the server has closed the question (§6.3), so even an undelayed TV feed cannot leak an answer inside anyone's window |
| Companion mode | A "Watching on TV" toggle in the apps mutes and minimises the video, or plays the audio-only rendition, while still receiving cues. If cues are not carried in audio-only (to verify in Phase 2a), the app plays the lowest rendition muted and hidden |
| Acquisition | The TV feed carries a QR code and "play in the app" graphics, set by producers in vMix |
| Not used | Audio sync via the phone microphone (the Master Plan forbids microphone monitoring). Questions pushed independently of the video (would break fairness) |
| Legal | Prize games on TV may also fall under Albania's Audiovisual Media Authority (AMA). UNVERIFIED; on counsel's list |

### 6.10 Chat: fully built and testable, toggled from the back office

The owner decided on 2026-11-25:

- **Chat is implemented in full and kept as a working, testable function.** It is switched on and off from the back-office admin (the studio console).
- **It defaults to OFF.** Turn it on for testing, rehearsals and any show where it is wanted.
- **Whether production launches with chat on is a separate go/no-go decision.** If launch goes ahead without chat, nothing changes in the apps or the backend: the switch simply stays off.
- **Chat delivery costs nothing while the switch is off**, because no tokens are minted and no messages are delivered (§12).

**Switch**

- `feature_switches` key `chat.enabled`, with scopes **global → per show** (the most specific scope wins).
- Toggled from the studio console by an authorised operator. Every change is audit-logged. The kill switch is the same switch.

**Delivery**

- `JoinShowResponse.chat_enabled` gives the state at join.
- **`Pulse.chat_enabled` carries live changes**: apps show or hide the chat panel within one pulse (≤5 s), with no app release and no rejoin.

**When OFF**

- The server mints no tokens.
- `GetChatToken` returns `FAILED_PRECONDITION`.
- Apps disconnect from chat and hide the panel.
- The review Lambda's `fallbackResult` is set to DENY.

**When ON**

- **The chat panel shows by default, but a player connects only when they open it.** On open, the app gets a token and is assigned a **room shard** (DEFAULT ≤ 5,000 members per room).

  The verified IVS Chat limits ([research 01](docs/research/01-live-video-ivs.md)) make the lazy connection necessary:
  - `CreateChatToken` is capped at **200 TPS** (fixed), so minting 300k tokens takes about 25 min.
  - The default limit is **50k connections**.
  - Delivery costs **$0.008 per 1,000 messages delivered**, so a single message to 300k connections costs about $2.40.
- Tokens use 180-min sessions, so one token covers the whole show.

**Moderation**

- Slow mode per room via `maximumMessageRatePerSecond`; also capped at 100 msg/s per room (fixed).
- Review Lambda in the same region as the room:
  - **200 ms timeout**
  - provisioned concurrency
  - Albanian and English word lists
  - `fallbackResult = DENY` under load
- Host highlights: moderators can pin messages for everyone.

**Quotas**

- Ask for chat connections sized to the expected share of players who open chat (DEFAULT plan: 30% of peak).

### 6.11 Push

- **APNs:** token-based `.p8` key over HTTP/2.
  - Show reminders use `timeSensitive` and a `collapse-id`.
  - **Live Activity countdown:** started locally on iOS 16.1+; push-to-start on iOS 17.2+; broadcast-channel updates on iOS 18+.
- **FCM:** HTTP v1, **sent to individual tokens only**. No server-side topics or device groups; those InstanceId APIs end on 2027-09-29.
- **Android 13+:** ask for `POST_NOTIFICATIONS` in context (on "Remind me").

### 6.12 Regions and disaster recovery (CHANGED)

| Item | Spec |
|---|---|
| Primary | **eu-central-1 (Frankfurt)**: all planes |
| Standby | **eu-west-1 (Ireland)**: IVS channel and chat rooms pre-created; Fargate services deployed and **warm-scaled to 30–50% (DEFAULT) during the show window**; Aurora Global Database secondary (non-headless); Valkey rebuildable or Global Datastore; Secrets Manager replicas; S3 CRR for KYC and media |
| Why not Milan | No IVS, no IVS Chat, no Managed Grafana. Fargate launch rate is 125 tasks/min per service (fixed), which cannot meet a 5-min cold RTO ([ADR-0003](docs/adr/0003-standby-region-ireland.md)) |
| Failover | Route 53 ARC routing controls (data-plane only). CloudFront cannot fail POST requests over between origins, so failover is at DNS level. Apps get `playback_url` and the chat endpoint from `JoinShow`, never hard-coded |
| Target | Show restart within **5 min** (DEFAULT), rehearsed every quarter as an FIS game day |

---

## 7. Identity and auth

Source: [ADR-0005](docs/adr/0005-device-bound-tokens.md), [research 05](docs/research/05-auth-security-compliance-web.md).

1. **Install:**
   - **iOS:** generate the App Attest key once, attest it, and send assertions after that. Apple asks to keep attestation under 100 requests/s and to ramp up gradually. Both sign `clientData = challenge ‖ DPoP jkt` (attestation uses its SHA-256).
   - **Android:** Play Integrity standard request, with `requestHash = base64url(SHA-256(challenge ‖ DPoP jkt))`. The check runs before the phone is entered, so it binds the one-time challenge and the device key (docs/research/08).
   - Generate a DPoP key in the Secure Enclave or Keystore, linked to the attested key.
2. **OTP request**, checked in this order:
   1. attestation ticket
   2. per-device, per-number and per-range limits
   3. line type (block VoIP, toll-free, premium and similar)
   4. pumping-risk score
   5. send

   Channel: **the user picks SMS, WhatsApp or Viber on the sign-in screen (ID-1)**. Button order is set after the price/delivery test. If the chosen channel fails, SMS is offered. SMS runs on two providers, each with its own pumping filter.
3. **OTP code:**
   - 6 digits from a CSPRNG
   - valid 5 min, single use
   - stored as a hash and compared in constant time
   - **5 attempts per code**
   - alphanumeric sender ID
4. **Access token:**
   - JWT with `typ=at+jwt`, **EdDSA**, a pinned `kid`, 10 min lifetime
   - validated on iss, aud, exp, nbf and typ
   - sender-constrained by **DPoP** (`cnf.jkt`)
5. **Refresh token:**
   - opaque and rotated on every use
   - **reuse of an old token revokes the whole family**
   - 30-day inactivity expiry
6. **Step-up:** a biometric unlock of the DPoP key plus a server nonce for claims and payout changes.
7. **SMS is a "restricted authenticator"** (NIST SP 800-63B-4 §3.2.9). The risk notice and the migration path to passkeys are documented in [`stack/05-auth-identity`](stack/05-auth-identity/).

---

## 8. Anti-cheat

**Principle:** make cheating unprofitable. Risk scores never eliminate anyone mid-show; they decide what happens at prize time. Details: [`stack/13-anti-cheat`](stack/13-anti-cheat/).

| Layer | Controls |
|---|---|
| L1 Question design | **At least half of the 12 questions (6+) are lookup-resistant (AC-1):** show-dependent (what the host said, wore, held or did), audio, brief visual or memory, weighted toward the late questions. The rest are pan-Albanian general knowledge. AI may draft only general-knowledge candidates. **Two-person approval** (NIST 800-53 AC-3(2)). Encrypted vault, released one question at a time. Producers set per-question timers (6–10 s) |
| L2 Client integrity | **Attestation:** attest at sign-in, assertion at join. **A failed attestation means the player plays for points only (AC-5).** **One prize-eligible account per device (AC-6):** iOS DeviceCheck bit plus a binding table; Android app-scoped ANDROID_ID (Play Integrity device recall evaluated in Phase 1). Focus-loss rule (§6.3.6). **Screen capture blocked while a question is open (AC-4):** Android sets `FLAG_SECURE` from `question_open` to `question_closed`; iOS has no screenshot-blocking API, so the question overlay is hidden while recording or mirroring is detected (`isCaptured`, `sceneCaptureState` on 17+), and screenshots are logged. Overlay and accessibility-automation signals are logged (Android `appAccessRiskVerdict`). **Accessibility users are never penalised.** Signed answers with server-side deadlines. One account per phone number; VoIP numbers blocked. No camera, microphone or location permission |
| L3 Risk scoring | From the event stream: answer timing, accuracy vs difficulty, identical answer sequences across linked accounts, focus and capture events, account age, sign-up velocity, referral chains, App Attest risk metric. Output: `clear` / `review` / `block` |
| L4 Winner verification | Tiers A, B, C (§3.2) |

| Other threat | Control |
|---|---|
| SIM farms / OTP pumping | Attestation before OTP; limits; line type; pumping score; provider filters; WAF ACFP on sign-up; conversion-rate alarm |
| Referral farming | Extra life granted only after the invitee verifies **and** completes a game |
| Reaction/like farming | Cosmetic only; per-user cap; one like per account |
| Claim fraud | Bank account name must match the ID; notarised power of attorney; hold period; KYC re-check after a number rebind |
| Insider leaks | Vault, per-question release, two-person approval, hash-chained audit log |
| Forged cues | Ed25519-signed cues; key rotation |
| Answer spam | Per-player limits in Valkey. WAF is only a coarse layer because it reacts in about 30 s or more |

---

## 9. Security baseline

- **Standards:**
  - OWASP **ASVS 5.0 Level 2** everywhere; **Level 3** for wallet, KYC, question vault and studio
  - **MASVS 2.1** for the apps
  - **API Security Top 10 2023**, especially API1 (BOLA) and API6 (sensitive business flows)
- **Secrets:** AWS Secrets Manager, replicated to the standby region. Nothing in code or app bundles. Terraform uses ephemeral and write-only arguments so secrets stay out of state.
- **KYC data:**
  - dedicated KMS CMK, SSE-KMS with Bucket Keys
  - Object Lock in **Governance** mode
  - Macie
  - CloudTrail data events
  - lifecycle rules matching the DPIA retention schedule
- **TLS pinning:** only our own API hosts, never IVS or CDN hosts (MASVS-NETWORK-2).
- **Consoles:**
  - SSO with **FIDO2/passkey-only** MFA
  - BFF with `__Host-` session cookies
  - strict nonce-based CSP
  - 5 min idle timeout on high-value consoles
- **Kill switches** (all in `feature_switches`, audited): chat, reactions, extra lives, new sign-ups, per-country eligibility.

---

## 10. Non-functional requirements

| NFR | Target | Notes |
|---|---|---|
| Concurrent players | 300k design, **600k load-tested** | Assumption (Master Plan) |
| Answer intake | ~30k req/s peak at 300k (60k at 600k) | Stateless `game` fleet, one atomic Valkey script per answer that also counts it (TS-16) |
| Reaction intake | **≤ 60k req/s at 300k**: 5 s batching with jitter, sent only when the user has tapped, interval tunable from the server | CHANGED from 2 s ([ADR-0006](docs/adr/0006-reaction-batching.md)); at 2 s this would have been the largest request stream (150k/s) |
| Answer ack latency | p99 < 300 ms server-side at peak | |
| Cue-to-render | p99 ≤ L_max (6 s DEFAULT) | Measured in rehearsal |
| Glass-to-glass latency | < 5 s (IVS Low-Latency) | |
| Reaction rendering | ≥ 55 fps on the low-end Android reference phone during playback, ~30 reactions/s drawn | OPEN: pick the reference phone |
| Show restart | ≤ 5 min | §6.12 |
| Pre-scale | **Booked automatically from each show's scheduled start** by the studio pre-show orchestrator: ALB LCU reservation at **T-4 h**; Fargate MinCapacity, KDS re-shard, Valkey scale-out and warm standby at **T-60** (healthy by T-20); back to baseline after SETTLED. Rescheduling a show re-books all of these | CHANGED: an LCU reservation can take hours to apply |
| Availability (show window) | 99.95% of shows complete | |

---

## 11. Observability

- **Tracing and metrics:** OpenTelemetry Go SDK. ADOT collector as an ECS sidecar (pinned image).
- **Logs:** stdout JSON to CloudWatch. The OTel Logs SDK is stable in Go 1.47, but the docs disagree on its status.
- **Live-show dashboard:**
  - joins
  - answers/s
  - acceptance rate
  - reactions/s
  - cue `PutMetadata` latency and errors
  - IVS `ConcurrentViews`
  - stream health
  - Valkey CPU per shard
  - pinned RDS Proxy sessions
  - error rates
- **Hosting:** in Frankfurt, with a CloudWatch cross-region view, so it is still usable if Frankfurt goes down. Managed Grafana is not available in every region.
- **Mobile:** Sentry (Cocoa 9.35.0, Android 8.65.1) with PII scrubbing.
- **Alerts:** on-call alarms on SLO burn, quota headroom (IVS views, chat connections, Fargate vCPU, Play Integrity daily quota), and the OTP conversion rate.

---

## 12. Cost model (order of magnitude; Europe list prices verified 2026-11-25)

| Line | Basis | Figure |
|---|---|---|
| Video delivery | IVS EU: SD $0.036, HD $0.072 per viewer-hour (first tier) | 300k × 0.5 h at HD ≈ **$9.4k per show** (upper bound; lower in practice with SD viewers and volume tiers) |
| Video input | ADVANCED_HD $0.85/h | Negligible |
| **Chat delivery (NEW)** | $0.56 per 1k sent, $0.008 per 1k delivered | **$0 while the back-office switch is off (the default).** When on, sharding and lazy connection keep it bounded. Example: 90k chatters in 5k-member rooms at 1 msg/s per room ≈ $0.04/s, **≈ $50 per show**. Without sharding, one message to 300k viewers would cost $2.40 |
| Compute | Fargate ARM $0.03725/vCPU-h | ~300 × 2 vCPU tasks × 1 h ≈ **$27 per show** (≈ $110–135/month at one show per week), plus the warm standby |
| Valkey | r7g.xlarge $0.42/h | 6 nodes always on ≈ **$1.8k/month** (about double with a cross-region datastore). **With one show a week, run a small baseline cluster and scale shards out before each show and back in after.** Validate the online-resharding time in the Phase 4 load test |
| Shield Advanced | $3,000/month, 1-year commitment | Largest fixed infrastructure line |
| SMS OTP | per sign-up | OPEN: delivery and price test |
| Prizes | product decision | Largest line overall |

---

## 13. Engineering practice

- **Repositories:** `ios`, `android`, `backend`, `web-consoles`, `infra`, `protocol` (Master Plan §8). This workspace holds all six as top-level folders laid out so each can become its own repository.
- **Commits:** one concern per commit; the owner approves and pushes.
- **Definition of done:** the build-standards checklist ([`docs/build-standards.md`](docs/build-standards.md)).
- **CI on every pull request:** lint, unit tests, **protocol test vectors**, `buf breaking`, build, CodeQL, dependency review, secret scan.
- **Supply chain:** workflow lockfile or SHA pins, OIDC to AWS, artifact attestations.

### 13.1 Test gates

| Type | Gate |
|---|---|
| Unit and component | Every PR |
| Protocol vectors (Go/Swift/Kotlin) + `buf breaking` | Every PR |
| `connectrpc/conformance` against the server | Nightly |
| Load | **600k** simulated players through a full show with reactions on (k6, ~20 runners × 30k VUs) |
| Animation | Reaction layer ≥ 55 fps on the low-end Android reference phone (Macrobenchmark `FrameTimingMetric`) |
| Chaos (FIS) | Task loss, Valkey AZ power interruption, Aurora failover, primary encoder drop during rehearsal, **region failover to Ireland** |
| Rehearsal shows | Weekly from the first working build |
| Penetration test | Before prizes go live (ASVS L2/L3 scope) |
| Real devices | Including the low-end Android reference phone (OPEN) |

---

## 14. Build phases

See **[PLAN.md](PLAN.md)** for the task-level plan.

---

## 15. Open items

See **[docs/open-items.md](docs/open-items.md)**.

---

## 16. Changes vs Master Plan v1.1

Every change rests on a cited primary source in `docs/research/`. Each one needs the owner's confirmation.

| # | Plan v1.1 | This spec | Evidence |
|---|---|---|---|
| 1 | Milan standby | **Ireland standby** | No IVS or IVS Chat in Milan; Milan Fargate launch limits too low (research 01, 03); [ADR-0003](docs/adr/0003-standby-region-ireland.md) |
| 2 | Minimum iOS 16 | **Kept at iOS 16 by owner decision** (research recommended 17.0; trade-offs in ADR-0004) | `@Observable` and `sceneCaptureState` need 17; Xcode 27 cannot debug on 16 (research 04); [ADR-0004](docs/adr/0004-mobile-minimum-os.md) |
| 3 | Reactions posted every 2 s | **5 s + jitter, only when tapped, server-tunable** | At 2 s reactions become a 150k req/s stream (research 02); [ADR-0006](docs/adr/0006-reaction-batching.md) |
| 4 | "Short-lived JWT + refresh, device binding" | **DPoP-bound EdDSA JWT, rotating refresh with family revocation, step-up before claims** | NIST 800-63B-4 classes SMS as restricted; OWASP; RFC 9700, RFC 9449 (research 05); [ADR-0005](docs/adr/0005-device-bound-tokens.md) |
| 5 | "IVS Low-Latency over WebRTC" (wording) | Low-Latency is HLS with ID3 cues. Decision unchanged | research 01 |
| 6 | Real-time "caps at 25,000" | 10k default, 25k per stage on request. Decision unchanged | research 01 |
| 7 | Backup encoder (implied automatic failover) | Priority-based stream takeover, identical encode settings, failover slate, runbook | research 01 |
| 8 | Adaptive renditions with SRT | **ADVANCED_HD** server-side transcoding (multitrack needs E-RTMP) | research 01 |
| 9 | "Conformance test suite" | **`protocol/testvectors`** corpus run on all three platforms, plus `connectrpc/conformance` | research 02 |
| 10 | Kinesis → S3 → Athena | Events in JSON for Firehose's Parquet conversion; KDS capacity pre-scaled | research 02 |
| 11 | Scale up 20 min before | MinCapacity at T-60; **LCU reservation ≥ T-4 h**; quota requests in phase 0 | research 02, 03 |
| 12 | Country switch | **Every country `eligible` from launch (owner decision LG-1).** The per-country switch is kept as a kill switch, and each change records a `decision_ref` | Owner decision 2026-11-25. Research 05 legal risk (MK permit + 18% fee; XK no legal basis) goes to counsel |
| 13 | Chat (IVS Chat) | **Backend on/off switch, live via pulse**; lazy connection; sharded rooms; chat cost line | Owner instruction (2026-11-25); IVS Chat quotas and pricing (research 01) |
| 14 | Dependabot (implied) | Renovate for `infra/` | Dependabot supports Terraform only up to 1.15 (research 03) |
| 15 | "RDS/Aurora" | **Aurora PostgreSQL 17**, UUIDv7 keys | RDS Proxy pinning, failover times (research 02) |
| 16 | (not in plan) TV simulcast | **App is the game; TV delayed behind the app; companion mode** | Owner request 2026-11-25; ADR-0010 |

# Decision register

Every choice that has to be settled before the build continues. Each row cites where it comes from: the Master Plan (MP), SPEC.md, an ADR, or a research report. Rows are updated as decisions are made.

**Status values**

| Status | Meaning |
|---|---|
| LOCKED | Fixed by the owner in the Master Plan |
| DEFAULT | Proposed in the Master Plan |
| CHANGED | Research contradicted the plan |
| NEW | A gap that research found |
| OPEN | Needs an owner call or outside input |
| **SETTLED** | Decided in a review session (date given) |

---

## A. Anti-cheat

| ID | Decision | Options | Recommendation | Status |
|---|---|---|---|---|
| AC-1 | Question mix per show | **At least half of the 12 questions (6+) are lookup-resistant: show-dependent, audio, brief visual or memory**, weighted toward the late questions. The rest are pan-Albanian general knowledge | Owner chose a stronger anti-AI mix than MP. More studio prep per show | **SETTLED 2026-11-25** |
| AC-2 | Who writes questions | AI may draft general-knowledge candidates only. Humans fact-check every question; two-person approval before air. Show-dependent, audio and visual questions are written by the show team | As MP | **SETTLED 2026-11-25** |
| AC-3 | Focus-loss rule on Q10–12 | **The answer is voided only if the app goes to background or is inactive > 1.5 s.** Shorter blips are a risk signal only. The rule is announced before every show | Refines MP | **SETTLED 2026-11-25** |
| AC-4 | Screenshots / screen recording | **Blocked while a question is open; normal at all other times.** Android: FLAG_SECURE set at `question_open` and cleared at `question_closed`. iOS has no screenshot-blocking API, so while recording or mirroring is detected (`sceneCaptureState`) the question overlay is hidden, and every screenshot is logged as a risk signal. Capture events are logged on both platforms | Owner choice; differs from MP log-only. Sharing and streaming still work outside question windows | **SETTLED 2026-11-25** |
| AC-5 | Device fails attestation | **Can play, but for points only; never wins prizes** | — | **SETTLED 2026-11-25** |
| AC-6 | Accounts per device | **One prize-eligible account per device**; other accounts on the same device play for points | — | **SETTLED 2026-11-25** |
| AC-7 | Risk score effect | MP: never eliminates mid-show; decides prize release: `clear` → pay, `review` → hold for manual review, `block` → withhold (unblock needs four-eyes) | As MP | **SETTLED 2026-11-25** |
| AC-8 | Late joiners | **Can join at any time and answer from the current question on** (play along). A question that opened before they joined counts as missed, and **an extra life cannot cover it**, so late joiners cannot win (see UF-4) | Owner choice | **SETTLED 2026-11-25** |
| AC-9 | Answer timing tolerance (`L_max`) | 6 s latency allowance + 1.5 s upload grace, tuned from rehearsal data | As proposed; revisit after Phase 2a measurements | **SETTLED 2026-11-25** |
| AC-10 | Jackpot finalists | **Private live video call with staff within 48 h of the show** (Tier C): ID check plus 2–3 fresh questions on the same timer. The producer may choose on-air verification for sponsor shows | — | **SETTLED 2026-11-25** |

## B. Identity and sign-in

| ID | Decision | Options | Recommendation | Status |
|---|---|---|---|---|
| ID-1 | OTP channels | **The user chooses on the sign-in screen: SMS, WhatsApp or Viber.** SMS runs on 2 providers with failover. If the chosen channel fails, offer SMS. The order of the buttons is set after the price/delivery test (PLAN L9) | Owner choice | **SETTLED 2026-11-25** |
| ID-2 | OTP code rules | 6 digits, valid 5 min, 5 tries per code, resend after 30 s | As proposed (NIST 800-63B-4) | **SETTLED 2026-11-25** |
| ID-3 | Device-bound sessions + biometric step-up before claims (ADR-0005) | **Yes.** Tokens are bound to a key in the phone's secure hardware. Face ID or fingerprint is required before claims and payout changes | ADR-0005 → Accepted | **SETTLED 2026-11-25** |
| ID-4 | Session length | access token 10 min; signed in for 30 days of inactivity | As proposed | **SETTLED 2026-11-25** |
| ID-5 | Age check | **Self-declared 18+ at sign-up; age is verified by government ID at Tier B/C and at any cash-out** | — | **SETTLED 2026-11-25** |
| ID-6 | Phone number changes | Re-binding a number triggers a KYC re-check before the next payout (no SIM-swap data exists for AL/XK/MK) | As proposed | **SETTLED 2026-11-25** |

## C. Prizes and claims

| ID | Decision | Options | Recommendation | Status |
|---|---|---|---|---|
| PR-1 | Prize currency | **Decided together with the bank partner (PLAN L8).** Until then, the ledger stays multi-currency-ready: every account carries an ISO 4217 code | Owner choice: defer | **SETTLED 2026-11-25** (deferred to bank) |
| PR-2 | Winner selection | **Skill only: all-correct winners split the pot equally; no random tie-breaks or draws** | As proposed. Keeps the show out of "game of chance" territory (research 05) | **SETTLED 2026-11-25** |
| PR-3 | Nobody wins | **The pot rolls over to the next show** | — | **SETTLED 2026-11-25** |
| PR-4 | Claim window | 60 days, with reminders at 30, 7 and 1 days | As MP | **SETTLED 2026-11-25** |
| PR-5 | Cash-out minimum | €20 equivalent | As MP | **SETTLED 2026-11-25** |
| PR-6 | Verification tiers | A < €50 · B €50–500 · C > €500 or flagged | As MP | **SETTLED 2026-11-25** |
| PR-7 | Collection options | Bank transfer to an account in the winner's own name · office pickup · notarised representative · charity donation · sponsor prize pickup or delivery | As MP | **SETTLED 2026-11-25** |
| PR-8 | KYC vendor | PoC with Veriff and Sumsub on real AL/XK/MK IDs | As proposed | **SETTLED 2026-11-25** |

## D. Show format and user features

| ID | Decision | Options | Recommendation | Status |
|---|---|---|---|---|
| UF-1 | Schedule | **One live show per week. Producers set the day and time per show in the studio console; the system never hard-codes a slot.** Between shows, engagement relies on question of the day, replays and clips, leaderboards and push | Owner decision; replaces the MP daily + weekend default | **SETTLED 2026-11-25** |
| UF-2 | Format | 12 questions × 3 options, 15–20 min | As MP | **SETTLED 2026-11-25** |
| UF-3 | Timers | **Set per show by the producers in the studio console**, within safe bounds of 6–10 s per question. Default ladder: 10 s (Q1–4), 8 s (Q5–8), 7 s (Q9–12) | Producer-configurable | **SETTLED 2026-11-25** |
| UF-4 | Extra life | **1 per show, never on Q12.** The player taps "Use extra life" during the reveal. It saves a wrong or unanswered question **only if the player was in the game when that question opened** | — | **SETTLED 2026-11-25** |
| UF-5 | Extra-life sources | referrals only (after the invitee verifies and plays a game); never sold | As MP | **SETTLED 2026-11-25** |
| UF-6 | Reaction set | heart, thumbs up, fire, laugh, clap, Albanian flag, + themed per show | As MP proposal | **SETTLED 2026-11-25** |
| UF-7 | Chat | Built and testable; back-office switch, off by default; go/no-go before launch | — | **SETTLED 2026-11-25** |
| UF-8 | Chat when on | Panel visible; a player connects when they open it; rooms of up to 5k. **Producers set slow mode per show** (default 1 message per 10 s per user) | Producer-configurable | **SETTLED 2026-11-25** |
| UF-9 | Leaderboards | country, city, diaspora vs homeland; weekly + all-time | As MP | **SETTLED 2026-11-25** |
| UF-10 | Winners wall | **Opt-in, asked when the winner claims.** Shown (first name, city, amount) only with consent | — | **SETTLED 2026-11-25** |
| UF-11 | App language | Albanian first, English second; in-app picker; default Albanian when the region is AL/XK/MK | As proposed (iOS has no Albanian system language) | **SETTLED 2026-11-25** |
| UF-12 | Website languages | sq (default), en, mk | As proposed | **SETTLED 2026-11-25** |
| UF-13 | Launch features | as MP §1 "Features at launch" | As MP | **SETTLED 2026-11-25** |
| UF-14 | Post-launch / later features | as MP §1 | As MP | **SETTLED 2026-11-25** |
| UF-15 | TV simulcast | **The app is the only place to play; TV is promotion.** The TV feed is deliberately delayed above the app p99 latency (~8–10 s to start), so TV never shows a question before an app timer starts. Apps get a "Watching on TV" companion mode. No microphone sync and no server-pushed questions | ADR-0010 | **SETTLED 2026-11-25** |

## E. Tech stack

| ID | Decision | Options | Recommendation | Status |
|---|---|---|---|---|
| TS-1 | Fully native apps | — | — | LOCKED |
| TS-2 | iOS minimum | **iOS 16** (owner kept the MP default). Trade-offs: `ObservableObject` state model, `isCaptured` fallback for capture blocking, iOS 16 test device gets TestFlight builds because Xcode 27 can't debug on it. Revisit yearly | ADR-0004 → Accepted | **SETTLED 2026-11-25** |
| TS-3 | Android minimum | **8.0 / API 26** · 9.0 / API 28 | 26 (99% reach) | **SETTLED 2026-11-25** |
| TS-4 | API protocol | **Connect RPC (protobuf over HTTPS)** · plain REST + protobuf · gRPC | Connect (ADR-0007) | **SETTLED 2026-11-25** |
| TS-5 | Backend | Go modular monolith, `game` + `core` | As MP | **SETTLED 2026-11-25** |
| TS-6 | Database | **Aurora PostgreSQL 17** · RDS Multi-AZ cluster | Aurora (Global DB for DR) | **SETTLED 2026-11-25** |
| TS-7 | Live state | **Valkey 9.1 node-based cluster** · Serverless | Node-based (Serverless scales too slowly for the spike) | **SETTLED 2026-11-25** |
| TS-8 | Video | **IVS Low-Latency ADVANCED_HD (720p top)** · STANDARD with 1080p | ADVANCED_HD ($0.85/h input, no 1080p delivery charges) | **SETTLED 2026-11-25** |
| TS-9 | Reaction batching | **5 s ± 1 s, sent only when tapped, server-tunable** | ADR-0006 → Accepted | **SETTLED 2026-11-25** |
| TS-10 | Standby region | **Ireland (eu-west-1) for backend, video and chat.** All 8 EU regions were compared (research 06). Only Frankfurt and Ireland run IVS and IVS Chat. Zurich, Milan and Spain also have capped server launch rates. Paris is closer but would need a split standby | ADR-0003 → Accepted | **SETTLED 2026-11-25** |
| TS-11 | Standby readiness | **Warm at 30–50% of primary during the show window** | — | **SETTLED 2026-11-25** |
| TS-12 | IaC tool | **Terraform 1.17** · OpenTofu | Terraform (ADR-0008) | **SETTLED 2026-11-25** |
| TS-13 | Web consoles | **React SPA + BFF** · Next.js | SPA + BFF (no React Server Components attack surface) | **SETTLED 2026-11-25** |
| TS-14 | Crash / performance | Sentry | As MP | **SETTLED 2026-11-25** |
| TS-15 | Capacity target | 300k concurrent players; load test 600k | As MP | **SETTLED 2026-11-25** |

## F. Operations, cost and governance

| ID | Decision | Options | Recommendation | Status |
|---|---|---|---|---|
| OP-1 | Shield Advanced ($3,000/month, 1-year commitment) | **Enable when cash prizes start (50k ramp stage).** Until then, use Shield Standard and WAF | — | **SETTLED 2026-11-25** |
| OP-2 | AWS Business Support | **Enable before prizes go live** | — | **SETTLED 2026-11-25** |
| OP-3 | GitHub plan | **GitHub Team now, with compensating controls:** protected main, CODEOWNERS, and an OIDC trust pinned to `environment:prod` on main. **Move to Enterprise Cloud before prizes go live** | — | **SETTLED 2026-11-25** |
| OP-4 | Low-end Android reference phones | **Primary: Samsung Galaxy A07 4G (SM-A075F, Helio G99, 4 GB)**, about €120 in Albania. Samsung leads Android in AL/XK/MK with 37–42%, and this was the #1 Android phone worldwide in Q1 2026. **Worst case: Xiaomi Redmi A5 (Unisoc T7250, Android Go)**, about 8,890 L at Vodafone AL; Xiaomi holds 15% in MK. Buy both while retail stock lasts | Research: StatCounter Oct 2026, Counterpoint Q1 2026, Shpresa and Vodafone AL listings | **SETTLED 2026-11-25** |
| OP-5 | Who builds it | **Claude Code agents write the code phase by phase. The owner reviews and approves every PR** against the 20-point build standards | — | **SETTLED 2026-11-25** |
| OP-6 | Audience ramp | 1k points → 10k small sponsor prizes → 50k/150k cash → open | As MP | **SETTLED 2026-11-25** |

## G. Legal and privacy (inputs for counsel; not legal advice)

| ID | Decision | Options | Recommendation | Status |
|---|---|---|---|---|
| LG-1 | Launch eligibility | **All countries eligible to win from launch.** Owner decision; it overrides the "AL only" recommendation. Prizes are still awarded and collected only in Albania (MP, locked). The per-country switch remains a kill switch: any country can be moved to `play_for_points` from the back office instantly. **Open risk:** North Macedonia's 2026 giveaway permit and 18% fee, and Kosovo having no legal basis for prize games. Counsel should confirm both before the first cash show (open items) | Owner choice; research 05 flagged the legal risk | **SETTLED 2026-11-25** |
| LG-2 | GDPR for the EU diaspora | **Appoint an EU Art. 27 representative and complete the DPIA now**, before any ID document or selfie is processed | — | **SETTLED 2026-11-25** |
| LG-3 | Product name and brand | — | — | OPEN |

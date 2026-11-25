# 13 Anti-cheat

## Role
Make cheating unprofitable (Master Plan §7).
- **Cheap controls run on everyone. Expensive controls run only on players who are about to be paid.**
- Risk scores never eliminate anyone mid-show. They decide what happens at prize time.

## Layers

### L1: Question design (show team and question-bank console)

| Type | Per show | Why it resists lookup |
|---|---|---|
| Show-dependent (what the host said, wore, held or did) | 2–3, placed late | Not in any database or AI model |
| Audio (song, voice, sound) | 1–2 | Slow to relay to another device |
| Visual shown briefly, then removed | 1–2 | Needs a capture inside a short window |
| Sequence or memory | 1 | Needs attention to the stream |
| Fast pan-Albanian general knowledge | remainder | Fun and accessible |

- AI may draft general-knowledge candidates only.
- **Two-person approval** of every question: the author cannot approve their own question, and the two approvers must be different people (NIST AC-3(2)).
- **Encrypted vault.** Each question is encrypted separately with KMS and decrypted only at release by `studio`.
- Content limits are in UTF-8 bytes: question ≤ 240 B, options ≤ 64 B each. The worst-case cue is 845 B, inside the 900 B budget (tested).

### L2: Client integrity

| Signal | iOS | Android | Use |
|---|---|---|---|
| Device attestation | App Attest: attest at sign-in, assertion at join | Play Integrity standard request + `appAccessRiskVerdict` | Required to join a prize show |
| Focus loss | `scenePhase` | lifecycle `onStop` / window focus | Risk signal. **Voids the answer on Q10–12** only on background, or when inactive for more than 1.5 s (DEFAULT) |
| Screen capture | `UIScreen.isCaptured` (iOS 16), `sceneCaptureState` (iOS 17+) | `addScreenRecordingCallback` (API 35+) | Risk signal |
| Screenshot | `userDidTakeScreenshotNotification` | `ScreenCaptureCallback` (API 34+) | Risk signal |
| Overlays | — | `setHideOverlayWindows` (API 31+), Integrity `*_OVERLAYS` | Hide overlays and log |
| Automation / accessibility | — | Enabled-services list, Integrity `*_CONTROLLING` | Risk signal. **Never penalise accessibility users** |
| Signed answers | DPoP proof on every request | same | The server sets the deadline |

The apps never use the camera, microphone or location.

### L3: Risk scoring (`backend/internal/risk`)

**Features** (computed from the Kinesis stream):
- `server_elapsed_ms` distribution against question difficulty, and accuracy against difficulty
- identical answer sequences across linked accounts (device, IP/ASN, referral graph, payout IBAN)
- focus and capture events
- account age and sign-up velocity
- referral chains
- App Attest risk metric
- Play Integrity device-recognition level

**Output:** `clear`, `review` or `block`, with reasons, written to `risk_flags`.

**Effect:**
- `review` holds payouts until manual review in the risk console.
- Reversing a `block` needs four-eyes approval.

### L4: Winner verification
Tiers A, B and C (SPEC §3.2). Tier C includes live video with 2–3 fresh questions on the same timer.

## Other threats

| Threat | Control |
|---|---|
| SIM farms / OTP pumping | Attestation before OTP, limits, line-type block, pumping score, provider filters, ACFP, conversion-rate alarm |
| Referral farming | Extra life only after the invitee is verified **and** completes a game |
| Reaction/like farming | Cosmetic only, per-user cap, one like per account |
| Claim fraud | Bank account name must match the ID, notarised POA, hold period, KYC re-check after a number rebind |
| Insider leaks | Vault, per-question release, four-eyes, hash-chained audit |
| Forged cues | Ed25519 signatures, key rotation, apps reject unsigned cues |
| Answer spam / bots | Per-player rate limits in Valkey, natural-key idempotency, attestation |

## Residual risk
A player with a second phone and an AI assistant can sometimes share a small split pot. Late show-dependent questions, risk scoring and live verification stop them from reliably taking significant money.

## Sources
[Master Plan §7](../../Master%20Plan/Trivia-Master-Plan-v1.1.pdf), [docs/research/04-native-mobile.md](../../docs/research/04-native-mobile.md), [docs/research/05-auth-security-compliance-web.md](../../docs/research/05-auth-security-compliance-web.md)

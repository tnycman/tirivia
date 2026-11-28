# 05 Auth and identity

## Role
- Phone sign-in.
- Device attestation before any OTP is sent.
- Sessions bound to the device.
- Step-up authentication before money moves.

## Decision
- Our own OTP service (Master Plan §5) with **two SMS providers**, plus WhatsApp and Viber templates where those are cheaper.
- Attestation: **App Attest** on iOS and **Play Integrity standard requests** on Android.
- Tokens: **EdDSA `at+jwt` access tokens, DPoP-bound**, with **rotating opaque refresh tokens** ([ADR-0005](../../docs/adr/0005-device-bound-tokens.md)).

## Standards and versions

| Item | Version / spec | Source |
|---|---|---|
| NIST SP 800-63B | Rev 4 (final, Jul 2025). SMS is a *restricted* authenticator | https://pages.nist.gov/800-63-4/sp800-63b.html |
| JWT BCP | RFC 8725 | https://www.rfc-editor.org/rfc/rfc8725.html |
| JWT access token profile | RFC 9068 | https://www.rfc-editor.org/rfc/rfc9068.html |
| OAuth Security BCP | RFC 9700 | https://www.rfc-editor.org/rfc/rfc9700.html |
| DPoP | RFC 9449 | https://www.rfc-editor.org/rfc/rfc9449.html |
| Play Integrity lib | 1.6.0 | https://developer.android.com/google/play/integrity/setup |

## Best practices
1. **Check order before any send:**
   1. Attestation ticket.
   2. Device, number and range limits.
   3. **Line type.** Block `fixedVoip`, `nonFixedVoip`, `tollFree`, `premium`, `sharedCost`, `uan`, `pager` and `voicemail` ([Line Type Intelligence](https://www.twilio.com/docs/lookup/v2-api/line-type-intelligence)).
   4. Pumping-risk score ([SMS Pumping Risk](https://www.twilio.com/docs/lookup/v2-api/sms-pumping-risk)).
   5. Send.
2. **OTP code:**
   - 6 digits from a CSPRNG, single use.
   - **Valid for 5 minutes.** NIST allows up to 10.
   - Store a hash only and compare in constant time.
   - **5 attempts per code.** NIST's hard ceiling is 100 ([NIST §3.1.3.2, §3.2.2](https://pages.nist.gov/800-63-4/sp800-63b.html)).
3. **SMS is a restricted authenticator.** We must notify users of the risk, offer an alternative (the device-bound key) and keep a migration plan ([NIST §3.2.9](https://pages.nist.gov/800-63-4/sp800-63b.html)).
4. **SMS pumping defences:**
   - Allow-list destination countries.
   - Turn on each provider's own filter: Twilio SMS Pumping Protection, AWS End User Messaging SMS Protect in Filter mode, or Vonage AIT. **Fraud Guard works only inside Twilio Verify.**
   - **Alarm when the send-to-verify conversion rate drops.**
5. **SMS sending setup:**
   - Use a branded alphanumeric sender. Generic IDs such as "Verify" and "INFO" are blocked in MK.
   - Two-way SMS is unsupported in AL, XK and MK.
   - Sources: [AL](https://www.twilio.com/en-us/guidelines/al/sms), [MK](https://www.twilio.com/en-us/guidelines/mk/sms), [XK](https://www.twilio.com/en-us/guidelines/xk/sms).
6. **WhatsApp OTP:** use authentication templates (zero-tap or one-tap on Android) with an expiry of 5 minutes or less ([Meta](https://developers.facebook.com/docs/whatsapp/business-management-api/authentication-templates)).
7. **App Attest:**
   - **Attest once per install** and use **assertions** after that. Keep `attestKey` under 100 requests/s and ramp up gradually.
   - The server performs all 11 attestation checks and 8 assertion checks, binding `clientData = challenge ‖ jkt` ([docs/research/08](../../docs/research/08-device-attestation.md)).
   - ([Apple](https://developer.apple.com/documentation/devicecheck/validating-apps-that-connect-to-your-server.md))
8. **Play Integrity:**
   - Use standard requests with `requestHash = base64url(SHA-256(challenge ‖ DPoP jkt))`. The check runs before the phone is entered; Google says never to put plaintext identifiers in the hash.
   - Decode verdicts on the server and never cache them.
   - The **default quota is 10k per day**, so request an increase in Phase 0 ([standard](https://developer.android.com/google/play/integrity/standard)).
9. **Access JWT:**
   - EdDSA (Ed25519), one algorithm per `kid`; reject `none`.
   - Validate `iss`, `aud`, `exp`, `nbf` and `typ=at+jwt`.
   - **Lifetime 10 minutes**, with `cnf.jkt` for DPoP binding.
10. **Refresh tokens:**
    - Opaque, rotated on every use.
    - **If an old refresh token is reused, revoke the whole token family.**
    - **A refresh or sign-out whose DPoP key differs from the token's key is refused; the family stays live** ([ID-7](../../docs/decision-register.md), [RFC 7009 §2.1](https://www.rfc-editor.org/rfc/rfc7009.html#section-2.1), [RFC 9449 §5](https://www.rfc-editor.org/rfc/rfc9449.html#section-5)).
    - Expire after 30 days of inactivity ([RFC 9700 §4.14](https://www.rfc-editor.org/rfc/rfc9700.html)).
11. **Step-up:**
    - Claims and payout changes need a biometric unlock of the DPoP key plus a server nonce.
    - **If the phone number was re-bound within the last 30 days, re-check KYC before payout.** No SIM-swap data exists for AL, XK or MK ([Twilio SIM Swap coverage](https://www.twilio.com/docs/lookup/v2-api/sim-swap)).

## Rules for this codebase
- Signing keys live in KMS, or Secrets Manager plus an HSM-backed KMS envelope. Rotate them every 90 days with a `kid` overlap.
- Every auth decision emits an event: `OtpRequested` and `AttestationResult`.
- **Kill switch: `signups.enabled`.**

## Changes vs Master Plan
- Tokens are bound to the device key with DPoP.
- Step-up is required before claims.
- KYC is re-checked after a phone-number re-bind.
- Each provider's own pumping filter is used.

## Code
`backend/internal/auth`, `ios/Packages/Core/Sources/CoreAttestation`, `android/core/attestation`

## Sources
[docs/research/05-auth-security-compliance-web.md](../../docs/research/05-auth-security-compliance-web.md), [docs/research/04-native-mobile.md](../../docs/research/04-native-mobile.md)

## Sourced values (used by `backend/cmd/core/policy.go` and `backend/cmd/game/policy.go`, verified 2026-11-25)

**Standard** means the value is fixed by a standard or a vendor's documented default. **Policy** means no published standard exists; the value is our design decision, tuned from production data.

| Value | Setting | Kind | Source |
|---|---|---|---|
| Code length, validity | 6 digits, 5 min (NIST max 10 min) | Standard | NIST SP 800-63B-4 §3.1.3.2 |
| Attempts per code | 5 | Standard | Twilio Verify error 60202 ("Max (5) verification check attempts") |
| Consecutive failures per phone | 100; **a new code never resets it**; success resets it | Standard | NIST SP 800-63B-4 §3.2.2 ("SHALL NOT reset the failed authentication count") |
| Sends per phone (core regions) | 5 per 10 min | Standard | Twilio Verify error 60203 (default send limit) |
| Resend spacing | 30 s (core) / 60 s (other) | Policy | No published value (NIST, OWASP, Twilio) |
| Sends per device | 5 / 3 per 10 min | Policy | No published value; mirrors the per-phone default |
| Number-range throttle | first 7 digits after `+`; 100 / 20 per 10 min | Policy | Twilio recommends prefix throttling but publishes no numbers ([toll fraud](https://www.twilio.com/docs/verify/preventing-toll-fraud)) |
| Pumping-risk block | ≥ 90 core ("High: Don't send"), ≥ 75 other ("Moderate: treat as suspicious") | Standard | [Twilio SMS Pumping Risk bands](https://www.twilio.com/docs/lookup/v2-api/sms-pumping-risk) |
| Number validation | libphonenumber `IsValidNumber`; only `MOBILE` and `FIXED_LINE_OR_MOBILE`; canonical E.164 | Standard | ITU-T E.164 (max 15 digits); [nyaruka/phonenumbers v1.8.1](https://github.com/nyaruka/phonenumbers) (libphonenumber port). Probed: AL, XK (+383), MK, IT, US, DE |
| Countries | **All allowed** (LG-1). Core regions get the standard limits; others get stricter policy limits | Policy | Core list: AL, XK, MK + diaspora destinations named in INSTAT/World Bank 2025 and the Kosovo Migration Profile 2023 (IT, GR, DE, GB, US, FR, CH, SI, HR, FI). The UN DESA 2024 ranking is **OPEN** (its site blocked automated fetch; download by hand) |
| DPoP `iat` window | ±60 s; jti kept for the window | Standard range | RFC 9449 §11.1 ("on the order of seconds or minutes") |
| JWT clock leeway | 30 s | Standard range | RFC 7519 §4.1.4 ("usually no more than a few minutes") |
| App Attest challenge / ticket lifetime | 5 min / 10 min | Policy | Apple: one-time challenge; no lifetime published |
| Play Integrity freshness | Token no older than the attestation challenge TTL (5 min); at most 30 s in the future (the JWT leeway) | Derived | Google gives only a placeholder `ALLOWED_WINDOW_MILLIS`. The request hash includes the challenge, so a valid token cannot be older than the challenge |
| Play Integrity prize eligibility | `PLAY_RECOGNIZED` + `LICENSED` + `MEETS_DEVICE_INTEGRITY` + our package and signing certificate. Anything weaker plays for points only (AC-5) | Standard | Google sample checks ([verdicts](https://developer.android.com/google/play/integrity/verdicts)) |

**Open follow-ups:**

- **Friction for "Mild/Moderate" pumping scores.** Twilio advises adding friction such as a CAPTCHA. Not built yet.
- **Diaspora ranking.** Rank the diaspora tier from the UN DESA 2024 workbook.

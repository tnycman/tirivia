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
   - The server performs all 9 attestation checks.
   - ([Apple](https://developer.apple.com/documentation/devicecheck/validating-apps-that-connect-to-your-server.md))
8. **Play Integrity:**
   - Use standard requests with `requestHash = SHA-256(phone ‖ server_nonce)`.
   - Decode verdicts on the server and never cache them.
   - The **default quota is 10k per day**, so request an increase in Phase 0 ([standard](https://developer.android.com/google/play/integrity/standard)).
9. **Access JWT:**
   - EdDSA (Ed25519), one algorithm per `kid`; reject `none`.
   - Validate `iss`, `aud`, `exp`, `nbf` and `typ=at+jwt`.
   - **Lifetime 10 minutes**, with `cnf.jkt` for DPoP binding.
10. **Refresh tokens:**
    - Opaque, rotated on every use.
    - **If an old refresh token is reused, revoke the whole token family.**
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

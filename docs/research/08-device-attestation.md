# 08 Device attestation: sourced facts (verified 2026-11-25)

These facts back `backend/internal/auth/attest/{playintegrity,appattest}`. Each was read from its primary source. **UNVERIFIED** means no source states it.

## Android: Play Integrity (standard requests)

| Fact | Source |
|---|---|
| **Only Google's servers can decrypt standard tokens.** Call `POST playintegrity.googleapis.com/v1/{package}:decodeIntegrityToken` with the scope `…/auth/playintegrity` | [standard](https://developer.android.com/google/play/integrity/standard), [Go client v0.317.0](https://pkg.go.dev/google.golang.org/api/playintegrity/v1) |
| **Authenticate as a service account** in the linked Cloud project. The Go client uses Application Default Credentials | standard page; pkg.go.dev |
| **Check `requestDetails` first**: package, `requestHash` and timestamp. Then check `PLAY_RECOGNIZED`, `LICENSED` and `MEETS_DEVICE_INTEGRITY` | [verdicts](https://developer.android.com/google/play/integrity/verdicts) |
| **No freshness value is published.** Google gives only the placeholder `ALLOWED_WINDOW_MILLIS` | verdicts page |
| **`requestHash`**: a digest of the request ("SHA256 of a stable request serialization"), at most 500 bytes. "Never put any sensitive information as plain-text" | standard page |
| **Android 13+ (May 2025)**: `MEETS_DEVICE_INTEGRITY` needs hardware-backed verified boot. Rooted devices can still get `MEETS_BASIC_INTEGRITY` | [improvements](https://developer.android.com/google/play/integrity/improvements) |
| **Tiered enforcement is recommended** over allow/deny. Verdicts should be "used alongside other signals". "Avoid caching integrity verdicts" | [overview](https://developer.android.com/google/play/integrity/overview) |
| **Decoding a token twice** returns cleared verdicts, not an error | standard page, "Automatic replay protection" |
| **No device identifier exists.** The terms forbid using the API "to fingerprint or track individual users or devices" | [terms](https://developer.android.com/google/play/integrity/terms) |
| **Device recall** (3 bits that survive reinstall and reset) is a beta. It is open only to apps with production Play Integrity **and Play Billing** usage above minimum thresholds | [device recall](https://developer.android.com/google/play/integrity/device-recall) |
| **Default quota** is 10,000 decodes per day. Increases need a linked Cloud project and the request form (up to a week) | [quota](https://developer.android.com/google/play/integrity/quota-limits) |
| **Client library** `com.google.android.play:integrity:1.6.0`. Warm-ups are limited to 5 per minute per app instance | [setup](https://developer.android.com/google/play/integrity/setup) |

**Our policy:**
- **Rejected:** a missing token, a token that will not decode, or a token whose package, hash or timestamp does not match. No ticket is issued.
- **`requestHash`** = base64url(SHA-256(challenge ‖ jkt)).
  - The challenge has a fixed length, so the concatenation is unambiguous.
  - No phone number or other plaintext is included.
  - The token is bound to the one-time challenge and to the device's DPoP key.
- **Freshness:** the token may be no older than the challenge TTL and at most the JWT leeway in the future. Since the hash includes the challenge, an older token cannot be valid anyway.
- **Prize-eligible** only when Google's sample checks all pass, our signing-certificate digest matches, and the response is not a testing response. Anything weaker plays for points (AC-5).
- **Other signals go to the risk score (AC-7)**, not to the eligibility decision: strong integrity, SDK level, activity level, Play Protect and app-access risk.

**OPEN (owner):** AC-6 ("one prize account per device") **cannot be enforced across reinstalls on Android**.
- Device recall is unavailable without Play Billing volume, and fingerprinting is forbidden.
- Today we enforce it per install (the hardware-backed DPoP key), and through the ID check at Tier B/C and at cash-out (ID-5).

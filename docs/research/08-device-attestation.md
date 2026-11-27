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

## iOS: App Attest

| Fact | Source |
|---|---|
| **Attestation has 11 steps:** chain to the App Attest root; nonce = SHA256(authData ‖ clientDataHash) in credCert extension OID 1.2.840.113635.100.8.2; keyID = SHA256(uncompressed point); RP ID = SHA256("TEAMID.bundleID"); counter 0; AAGUID; credentialId = keyID; `apple_validation_category_01`; `apple_bundle_version_01`. **Assertion has 8 steps** | [Validating apps that connect to your server](https://developer.apple.com/documentation/devicecheck/validating-apps-that-connect-to-your-server) |
| **Validation categories:** 0 invalid, 1 OS executable, 2 TestFlight, 3 development, 4 App Store, 5 enterprise/ad hoc, 6 Developer ID, 7–9 "aren't appropriate", 10 other | Same page |
| **Root CA:** `Apple_App_Attestation_Root_CA.pem`, valid 2020-03-18 to 2045-03-15. SHA-256 `1cb9823b…6242c932`, computed by us (Apple publishes no fingerprint) | [apple.com/certificateauthority/private](https://www.apple.com/certificateauthority/private/) |
| **Challenge:** one time, at least 16 bytes. Keys do not survive reinstall, migration or restore, so the app re-attests | [Establishing your app's integrity](https://developer.apple.com/documentation/devicecheck/establishing-your-app-s-integrity) |
| **Rollout:** keep `attestKey` under 100 requests/s across all installs, and ramp to at most 10M users per day | [Preparing to use App Attest](https://developer.apple.com/documentation/devicecheck/preparing-to-use-the-app-attest-service) |
| **Receipt** is PKCS #7 under Apple Root CA - G3. The fraud metric (attested keys per device in 30 days) comes from `data.appattest.apple.com` with a DeviceCheck JWT | [Assessing fraud risk](https://developer.apple.com/documentation/devicecheck/assessing-fraud-risk) |
| **CBOR:** fxamacker/cbor v2.10.0 with duplicate keys rejected, indefinite lengths forbidden, and the library's minimum container limits | [README](https://github.com/fxamacker/cbor) |

**Findings from the test vectors, which differ from Apple's prose:**
- **Apple's own sample uses the raw challenge as clientDataHash.** The verifier therefore takes clientDataHash directly.
- **Apple's flags don't follow WebAuthn.**
  - Extensions appear without the ED flag (`0x80`).
  - Assertions carry the AT flag (`0x40`) with no credential data.
  - So the object type decides the layout, not the flags.
- **The assertion signature is ECDSA-SHA256 over the nonce**, i.e. the digest is SHA256(nonce). This is confirmed by a real iOS 14.4 assertion (veehaitch/devicecheck-appattest, Apache-2.0).
- **A real development attestation carries the AAGUID `appattestdevelop`.** That settles the "appattestsandbox" mention on Apple's preparation page.
- **Validation category byte order is OPEN.** It arrives as the 4-byte string `01 00 00 00`, read little-endian: big-endian lands outside Apple's table. Values outside the table play for points, so an error shows up in rehearsal, never as a silent prize. Confirm on the first real TestFlight build.

**Our policy:**
- **Binding:** clientData = challenge ‖ jkt. Attestation uses SHA256(clientData); assertions sign clientData itself. This binds the evidence to the one-time challenge and the DPoP key, the same binding as Android's requestHash.
- **Rejected:** categories 0, 1 and 7–9.
- **Prize-eligible** only with the production AAGUID and either category 4 or no category (older iOS). The category cannot be stripped, because the Apple-signed nonce covers it. Everything else plays for points (AC-5).
- **Counters** only move forward, atomically (`counter < $new`). A replayed assertion fails even under concurrent requests.
- **OPEN:**
  - The fraud-metric receipt exchange. Receipts are stored; the DeviceCheck key is needed.
  - AC-6 on iOS is also per install: App Attest keys do not survive reinstall.

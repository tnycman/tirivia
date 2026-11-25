# Research 05: Auth/OTP/tokens, KYC, privacy and legal, web consoles

> Checked against primary sources on **2026-11-25**. **UNVERIFIED** marks claims we could not confirm. **None of the legal content below is legal advice.** Each legal point is an input for Albanian, North Macedonian and Kosovar counsel (Master Plan §11).
> Feeds into: [`stack/05-auth-identity`](../../stack/05-auth-identity/README.md), [`stack/12-security`](../../stack/12-security/README.md), [`stack/15-compliance-privacy`](../../stack/15-compliance-privacy/README.md), [`stack/08-web-consoles`](../../stack/08-web-consoles/README.md), [ADR-0005](../adr/0005-device-bound-tokens.md).

## 1. Verified facts

| Fact | Value | Source |
|---|---|---|
| NIST SP 800-63B-4 | **Final, July 2025** | https://csrc.nist.gov/pubs/sp/800/63/b/4/final |
| SMS/PSTN OTP | **"Restricted" authenticator** (§3.1.3.3), the only one in that category | https://pages.nist.gov/800-63-4/sp800-63b.html |
| Restricted duties (§3.2.9) | Offer an unrestricted alternative, notify users of the risk, carry out a risk assessment, and keep a migration plan | same |
| OTP | At least 6 digits, valid ≤10 min, single use (§3.1.3.2). No more than 100 consecutive failures (§3.2.2) | same |
| Risk signals before sending SMS | SHOULD check device swap, SIM change and number porting | same |
| AAL2 | SHALL offer a phishing-resistant option. Reauthentication ≤24 h overall, ≤1 h inactivity | same |
| OWASP on SMS | Don't use SMS for apps handling PII or carrying financial risk. If used, document that the risk is accepted, rate-limit, and monitor SIM swap | https://cheatsheetseries.owasp.org/cheatsheets/Multifactor_Authentication_Cheat_Sheet.html |
| RFC 8725 (BCP 225) | Algorithm allow-list, one algorithm per key, validate iss/sub/aud, explicit `typ` | https://www.rfc-editor.org/rfc/rfc8725.html |
| RFC 9068 | `typ=at+jwt`; iss, exp, aud, sub, client_id, iat, jti | https://www.rfc-editor.org/rfc/rfc9068.html |
| RFC 9700 (BCP 240, Jan 2025) | Public-client refresh tokens MUST be sender-constrained or rotated. Reuse means revoke. Inactivity expiry | https://www.rfc-editor.org/rfc/rfc9700.html |
| RFC 9449 DPoP | Binds access and refresh tokens to a client key (`cnf.jkt`, server nonce) | https://www.rfc-editor.org/rfc/rfc9449.html |
| Twilio Fraud Guard | **Verify only, SMS only** | https://www.twilio.com/docs/verify/preventing-toll-fraud/sms-fraud-guard |
| Twilio SMS Pumping Protection | For Programmable Messaging; GA in Apr 2024 | https://www.twilio.com/docs/messaging/features/sms-pumping-protection-programmable-messaging |
| AWS SMS Protect | Allow/Block/Monitor/Filter per country; AIT filter $0.01 per message | https://docs.aws.amazon.com/sms-voice/latest/userguide/filter-and-monitor-messages.html |
| Vonage Fraud Defender | Standard tier free; Advanced adds AIT protection | https://developer.vonage.com/en/dashboard/monitor/fraud-defender/fraud-defender |
| Twilio Line Type Intelligence | Mobile, landline, fixedVoip, nonFixedVoip, tollFree, premium…; worldwide | https://www.twilio.com/docs/lookup/v2-api/line-type-intelligence |
| Twilio SMS Pumping Risk lookup | Score 0–100; worldwide | https://www.twilio.com/docs/lookup/v2-api/sms-pumping-risk |
| **Twilio SIM Swap** | **Does not cover AL, XK or MK** | https://www.twilio.com/docs/lookup/v2-api/sim-swap |
| SMS rules for Albania | Dynamic alphanumeric sender, no registration needed. No two-way SMS, no short codes | https://www.twilio.com/en-us/guidelines/al/sms |
| SMS rules for North Macedonia | Alphanumeric sender; avoid generic IDs such as "Verify" or "INFO" | https://www.twilio.com/en-us/guidelines/mk/sms |
| SMS rules for Kosovo | Alphanumeric and international long code; no two-way SMS | https://www.twilio.com/en-us/guidelines/xk/sms |
| WhatsApp auth templates | Copy-code, one-tap and zero-tap (Android); expiry 1–90 min | https://developers.facebook.com/docs/whatsapp/business-management-api/authentication-templates |
| Viber OTP | Via partners (Infobip, Sinch); charged only when delivered; SMS failover; requires user opt-in | https://www.infobip.com/viber-business/otp |
| KYC coverage: Veriff | AL: passport, ID card, driving licence. XK/MK: the same plus residence permit | https://www.veriff.com/supported-countries |
| KYC coverage: Sumsub | AL: passport, Letërnjoftim (back side required), driving licence, residence permit | https://docs.sumsub.com/docs/global-supported-documents |
| KYC coverage: Entrust/Onfido Instant | **AL: passport only.** XK: ID card and passport. MK: ID card | https://documentation.identity.entrust.com/guide/supported-documents-instant/ |
| Liveness standard | ISO/IEC 30107-3:2023 | https://webstore.iec.ch/publication/81714 |
| **Albania Law 124/2024** | In force 31 Jan 2025; described as fully aligned with GDPR. Biometric data is sensitive data. 72 h breach notice. DPO required for large-scale sensitive data. Fines up to ALL 2 bn or 4% of turnover. **DPIA duty deferred to about Jan 2027** | https://kpmg.com/al/en/insights/2025/02/new-law-on--personal-data-protection-.html · https://iapp.org/news/a/albania-s-personal-data-protection-law-a-legal-framework-harmonized-with-the-gdpr |
| Regulator (AL) | Commissioner for the Right to Information and Personal Data Protection (IDP) | https://idp.al/ |
| Kosovo data law | Law 06/L-082; Information and Privacy Agency | https://aip.rks-gov.net/?p=393 |
| North Macedonia data law | Official Gazette 42/20; Agency for Personal Data Protection (AZLP); GDPR-aligned | https://azlp.mk/en/pdpa/regulations-and-documents/laws/ |
| EU AI Act | 1:1 biometric verification excluded from Annex III 1(a) | https://artificialintelligenceact.eu/annex/3/ |
| Albania gambling | Law 155/2015 as amended by 18/2024: online sports betting only, max 10 licences, AMLF | https://www.karanovicpartners.com/news/albanian-new-law-on-gambling/ |
| **North Macedonia giveaways** | New law in force **14 Jul 2026**: Ministry of Finance permit plus **18% of the prize fund**, paid before the game starts (law-firm summary) | https://bdkadvokati.com/north-macedonia-adopts-new-act-on-games-of-chance-and-entertainment-games |
| Kosovo | Law 06/L-155 bans games of chance (2019). Prize games have had no legal basis since then (GAP Institute) | https://www.institutigap.org/news/3228 |
| Apple 5.3 | 5.3.1 the developer is the sponsor; **5.3.2 official rules in the app stating Apple is not a sponsor**; 5.3.3 no IAP for real-money gaming | https://developer.apple.com/app-store/review/guidelines/ |
| Google Play | No real-money entry for prizes. Promotions need official rules, number of winners and deadlines. Free-entry contests are not explicitly covered | https://support.google.com/googleplay/android-developer/answer/9877032 |
| OWASP ASVS | **5.0.0** (May 2025). L2 for most apps, L3 for high-value | https://asvs.owasp.org/ |
| OWASP API Top 10 | 2023 edition | https://api-security.owasp.org/editions/2023/en/0x11-t10 |
| React | **19.3.1** | https://react.dev/versions |
| RSC CVE | **CVE-2025-55182** (CVSS 10, RCE in React Server Components) | https://react.dev/blog/2025/12/03/critical-security-vulnerability-in-react-server-components |
| TypeScript | **7.0.2** (Go-based). **No programmatic API**, so typescript-eslint and Astro need TS 6 via `@typescript/typescript6` | https://devblogs.microsoft.com/typescript/announcing-typescript-7.0/ |
| Vite / Astro / TanStack | **8.3.8 / 7.4.13 / Query 5.108.4 / Router 1.170.46** | npm registry |
| Identity Center MFA | FIDO2 and TOTP. **Not available when an external IdP is used** (enforce MFA at the IdP) | https://docs.aws.amazon.com/singlesignon/latest/userguide/enable-mfa.html |
| NIST SP 800-53 r5 | AC-5 separation of duties; **AC-3(2) dual authorization** | https://csf.tools/reference/nist-sp-800-53/r5/ac/ac-3/ac-3-2/ |

## 2. Auth, OTP and tokens: adopted practices

1. **OTP**: 6 digits from a CSPRNG, single use, **valid 5 min** (NIST allows up to 10). Store a hash only and compare in constant time. **5 attempts per code**, plus per-number, per-device and per-range limits.
2. **SMS is a restricted authenticator.** Users see a risk notice. The **device-bound key (DPoP, attested) is the long-lived factor**, and there is a migration path to passkeys.
3. **Checks run in this order before any send**:
   1. attestation (Play Integrity `requestHash` = H(phone, nonce); iOS assertion)
   2. device and number limits
   3. line type (block VoIP, toll-free, premium and similar)
   4. pumping-risk score
   5. prefix reputation
   6. send
4. **Anti-pumping**:
   - Country allow-list: AL, XK, MK and the diaspora countries.
   - Provider-side filter on both providers.
   - Per-country burst throttles.
   - **Alert on drops in the send-to-verify conversion rate.**
5. **Channels**: WhatsApp auth templates (zero-tap or one-tap on Android) and Viber OTP where cheaper, with SMS as failover. Use a branded alphanumeric sender and no reply-based flows.
6. **SIM-swap data is unavailable** for AL, XK and MK. Use rebind cooldowns and device binding. **Re-check KYC before any payout after a number rebind.**
7. **Access token**:
   - JWT with `typ=at+jwt`, **EdDSA (Ed25519)**.
   - One algorithm per `kid`; reject `none`.
   - Validate iss, aud, exp, nbf and typ.
   - **10 min lifetime.**
8. **Refresh token**:
   - Opaque, rotated on every use.
   - **On reuse, revoke the whole family.**
   - 30-day inactivity expiry.
9. **DPoP (RFC 9449)**:
   - Key in the Secure Enclave or Android Keystore, linked to the attested key id.
   - A proof on every request.
   - A server nonce is required for claims and payouts.
   - **Step-up** (biometric unlock of the key) before a claim.
10. **Revocation**: a denylist keyed by `jti` and `iss`.

## 3. KYC, privacy and legal: adopted practices (not legal advice)

1. **KYC vendor proof of concept with real AL, XK and MK documents.** Shortlist **Veriff and Sumsub** (verified ID-card coverage); Entrust Instant covers only the Albanian passport. Require ISO/IEC 30107-3 PAD test letters plus injection and virtual-camera detection.
2. **Biometric and ID data is sensitive** (Law 124/2024; GDPR Art. 9):
   - explicit consent
   - field-level envelope encryption (KMS)
   - stored apart from the ledger
   - short retention for raw selfies; longer retention for results and the audit trail (periods set in the DPIA)
3. **Do the DPIA now.** The Albanian duty is deferred, but **GDPR Art. 35 applies today** to EU-resident diaspora.
4. **GDPR Art. 3(2) targeting applies to the EU diaspora**, so an **Art. 27 EU representative** is required.
5. **Skill-only winner selection. No random tie-breakers or draws.** The prize is split. This keeps chance out of winner selection (AL, MK) *(reasoning, for counsel to confirm)*.
6. **North Macedonia** (in force 14 Jul 2026): permit plus 18%. Until counsel clears it, **MK = `play_for_points`**.
7. **Kosovo**: there is no legal basis for prize games. Until counsel clears it, **XK = `play_for_points`**.
8. **Stores**: official rules in the app with the "Apple is not a sponsor" statement; no IAP linked to prizes. Confirm free-entry contests with Play policy support (UNVERIFIED).

## 4. Web consoles: adopted practices

1. **Vite single-page apps** (React 19.3, TanStack Router/Query) behind a **backend-for-frontend (BFF)**. No React Server Components, which avoids the CVE-2025-55182 attack surface.
2. **Server-side sessions in `__Host-` cookies** (Secure, HttpOnly, SameSite=Strict). No JWTs held in the browser. Idle timeout of 5 min on claims, risk, vault and studio.
3. **SSO with phishing-resistant MFA only** (FIDO2/passkeys), through IAM Identity Center or an OIDC IdP with MFA enforced at the IdP.
4. **Strict nonce-based CSP** with `strict-dynamic`, `object-src 'none'`, `base-uri 'none'` and `frame-ancestors 'none'`. Roll it out Report-Only first.
5. **Four-eyes approval enforced on the server** (AC-5, AC-3(2)): the author cannot approve, two distinct approvers are needed, and approvers rotate. It applies to: question publish, payout release, KYC override, and reversing a risk "block".
6. **Question vault**: KMS encryption per question. Decryption happens only in the live studio session, inside a time window.
7. **Audit log**: append-only and hash-chained, plus S3 Object Lock. Each entry records who, what, before/after, approver, IP, device and request id.
8. **ASVS 5.0 L2 baseline; L3 for wallet, KYC, vault and studio.** Test against API Top 10 2023, especially API1 (BOLA) and API6 (sensitive business flows).
9. **Public website**: Astro static build with i18n (sq default, en, mk).
10. **TypeScript 7** for type checking, with the `@typescript/typescript6` alias for lint and Astro tooling.

## 5. Conflicts with Master Plan v1.1 and adopted fixes

| Plan | Finding | Adopted fix |
|---|---|---|
| Phone OTP + JWT is the only factor in front of a cash wallet | NIST: SMS is restricted. OWASP: don't use it where there is financial risk | **The device-bound DPoP key is the long-lived factor**, with step-up before claims ([ADR-0005](../adr/0005-device-bound-tokens.md)) |
| "Two SMS providers with failover" with Fraud Guard implied | Fraud Guard works only inside Verify | Enable each provider's own pumping filter (Twilio SMS Pumping Protection, AWS Protect Filter, Vonage AIT) |
| SIM-swap defence | No SIM-swap API covers AL, XK or MK | Rebind cooldown and a KYC re-check before payout |
| KYC vendor not named | Coverage differs by vendor | PoC with Veriff and Sumsub |
| DPIA | Albanian duty deferred | Do it now (GDPR) |
| `eligible` / `play_for_points` per country | Legal status of MK and XK | **AL = eligible; XK, MK and all diaspora countries = `play_for_points` at launch**, until each is cleared by counsel |
| Web consoles in React | React Server Components CVE; TS 7 tooling breaks | Vite SPA plus BFF; TS 6 alias for tooling |
| Web console SSO | Identity Center MFA does not apply with an external IdP | FIDO2-only MFA, enforced at whichever IdP is used |

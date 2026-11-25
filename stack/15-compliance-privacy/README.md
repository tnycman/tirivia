# 15 Compliance and privacy

> **This is not legal advice.** Each item below is an input for counsel (Master Plan §4, §11).

## Role
Covers the legal requirements of the show: how it is classified, who is eligible, how personal and biometric data is handled, and the app-store rules. All are applied through backend switches, so a change never needs an app release.

## Facts (verified 2026-11-25)

| Topic | Finding | Source |
|---|---|---|
| Albania privacy | **Law 124/2024** has been in force since 31 Jan 2025 and is described as GDPR-aligned. Biometric data is sensitive data. A breach must be reported within 72 h. A DPO is required for large-scale processing of sensitive data. Fines reach ALL 2 bn or 4% of turnover. The DPIA duty is deferred to about Jan 2027 | https://kpmg.com/al/en/insights/2025/02/new-law-on--personal-data-protection-.html |
| Regulator | Commissioner for the Right to Information and Personal Data Protection (IDP) | https://idp.al/ |
| GDPR reach | Art. 3(2) applies when the service targets EU residents (the diaspora), so an Art. 27 EU representative is needed | https://www.edpb.europa.eu/our-work-tools/our-documents/guidelines/guidelines-32018-territorial-scope-gdpr-article-3-version_en |
| Kosovo privacy | Law 06/L-082; Information and Privacy Agency | https://aip.rks-gov.net/?p=393 |
| North Macedonia privacy | Official Gazette 42/20; AZLP | https://azlp.mk/en/pdpa/regulations-and-documents/laws/ |
| EU AI Act | 1:1 biometric verification is excluded from Annex III 1(a) | https://artificialintelligenceact.eu/annex/3/ |
| Albania gambling | Law 155/2015 as amended by 18/2024 permits online sports betting only (AMLF). How a free-entry quiz is classified is **OPEN** | https://www.karanovicpartners.com/news/albanian-new-law-on-gambling/ |
| North Macedonia | In force 14 Jul 2026: giveaways need a Ministry of Finance permit plus **18% of the prize fund** (law-firm summary) | https://bdkadvokati.com/north-macedonia-adopts-new-act-on-games-of-chance-and-entertainment-games |
| Kosovo | Games of chance banned since 2019. Prize games have had no legal basis since then | https://www.institutigap.org/news/3228 |
| Apple 5.3 | Developer is the sponsor. **Official rules in the app stating Apple is not a sponsor.** No IAP for real-money gaming | https://developer.apple.com/app-store/review/guidelines/ |
| Google Play | No real-money entry. Promotions need official rules. Free-entry contests are not explicitly addressed (confirm with Play) | https://support.google.com/googleplay/android-developer/answer/9877032 |

## Controls we build

| Control | Implementation |
|---|---|
| Country eligibility | `country_eligibility` table (`eligible` / `play_for_points` / `blocked`). Every change needs a `decision_ref` (owner decision or counsel opinion) and is audited. **Launch: all countries eligible (LG-1).** Counsel should confirm the North Macedonia permit/fee and the Kosovo position before the first cash show. The switch moves any country to `play_for_points` instantly |
| Skill-only winners | No random tie-break or draw anywhere in the code. The pot is split equally |
| 18+ | Age declared at sign-up; verified by ID at Tier B/C and at cash-out |
| Official rules | In the app (sq/en), versioned, with the "Apple and Google are not sponsors" statement; acceptance logged per user and version |
| Consent | Separate, versioned consents for ID documents and biometrics, verification recordings, and winners-wall and replay display |
| DPIA | Done **now** (GDPR applies), stored in `docs/compliance/dpia.md`, reviewed before each new data flow |
| Retention | Raw selfies and liveness media kept short (the DPIA sets the period). Verification results and audit records kept longer. S3 lifecycle enforces both |
| Data subject rights | Export and erasure endpoints. Erasure respects legal holds (ledger and AML records). Object Lock **Governance** mode, not Compliance |
| Breach response | 72 h notification runbook (IDP + EU lead authority via the Art. 27 representative) |
| Winners wall | Shows first name + city + amount only, with an opt-out |

## Sources
[docs/research/05-auth-security-compliance-web.md](../../docs/research/05-auth-security-compliance-web.md)

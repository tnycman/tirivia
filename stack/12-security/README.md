# 12 Security baseline

## Role
Protect a system that handles cash, government IDs and biometrics, and that draws a large audience at one predictable time every day.

## Standards

| Standard | Level / scope | Source |
|---|---|---|
| OWASP ASVS 5.0.0 (May 2025) | **L2** everywhere; **L3** for wallet, KYC, question vault, studio | https://asvs.owasp.org/ |
| OWASP MASVS 2.1.0 | iOS and Android apps | https://github.com/OWASP/masvs |
| OWASP API Security Top 10 2023 | All APIs; focus on API1 (BOLA) and API6 (sensitive business flows: OTP, claims) | https://api-security.owasp.org/editions/2023/en/0x11-t10 |
| NIST SP 800-63B-4 | Authentication | https://pages.nist.gov/800-63-4/sp800-63b.html |
| NIST SP 800-53 r5 AC-5 / AC-3(2) | Separation of duties, dual authorization | https://csf.tools/reference/nist-sp-800-53/r5/ac/ac-3/ac-3-2/ |

## Controls

**Identity and access**
- Detail is in [05-auth-identity](../05-auth-identity/).
- People use AWS only through IAM Identity Center with FIDO2. There are no IAM users.
- Break-glass access is monitored.

**Data**
- KMS customer-managed keys (CMKs) for each data class: KYC, PII columns, the question vault, and the audit log.
- Envelope encryption for PII columns.
- Phone numbers are looked up by HMAC.
- Macie scans S3.

**Secrets**
- Stored in AWS Secrets Manager, replicated to the standby region, and rotated.
- Never in code, app bundles or Terraform state.

**Network**
- Private subnets for compute and data.
- VPC endpoints for AWS APIs.
- The only public entry point is CloudFront → ALB, behind WAF.

**Edge**
- WAF: managed rules, rate rules keyed by session header and IP, ACFP, and targeted Bot Control.
- Shield Advanced.

**Mobile (MASVS)**

| Control | What we do |
|---|---|
| STORAGE-1/2 | Keychain/Keystore; no PII in logs or Sentry |
| NETWORK-1/2 | TLS everywhere; pinning only on our own API hosts |
| PLATFORM-3 | Overlay protection on the live and wallet screens |
| CODE | Minimum OS version, forced update |
| RESILIENCE-1..4 | App Attest / Play Integrity; anti-tamper |

**Consoles**
- SSO with FIDO2.
- BFF sessions.
- Strict CSP.
- Four-eyes approval.
- Hash-chained audit log.

**Supply chain**
- Detail is in [10-ci-cd](../10-ci-cd/).

**Kill switches**
- Chat, reactions, extra lives, new sign-ups, and per-country eligibility.
- All live in `feature_switches`, are audited, and take effect without an app release.

**Incident response**
- Runbooks in `docs/runbooks/`.
- Breach notification within **72 h** under Albania Law 124/2024 and GDPR.

## Gates
- **Before prizes go live:** a penetration test scoped to ASVS L2/L3 plus the MASVS app checks.
- **Before every release:** threat model updated for new flows (STRIDE per feature, kept in `docs/threat-models/`).
- **Every PR:** dependency and vulnerability scanning (govulncheck, Trivy image scan, CodeQL).

## Sources
[docs/research/05-auth-security-compliance-web.md](../../docs/research/05-auth-security-compliance-web.md), [docs/research/03-aws-infra-devops.md](../../docs/research/03-aws-infra-devops.md), [docs/research/04-native-mobile.md](../../docs/research/04-native-mobile.md)

# 08 Web consoles and public website

## Role

**Internal consoles** (staff only):

- **studio**: runs the live show and holds the **feature switches**, chat included
- **question-bank**: question entry with two-person approval
- **claims**
- **risk**
- **clips**

**Public website**: show schedule, countdown, winners and store links. There is no web play.

## Decision

- Consoles are **React + TypeScript single-page apps built with Vite**, served behind a **backend-for-frontend (BFF)**. They do not use React Server Components.
- The public site is built with **Astro**: static pages in sq, en and mk.

## Pinned versions (2026-11-25)

| Tool | Version | Source |
|---|---|---|
| React | 19.3.1 | https://react.dev/versions |
| TypeScript | 7.0.2 (`tsc`) plus the `@typescript/typescript6` alias for lint and Astro tooling | https://devblogs.microsoft.com/typescript/announcing-typescript-7.0/ |
| Vite | 8.3.8 | https://registry.npmjs.org/vite/latest |
| TanStack Router / Query | 1.170.46 / 5.108.4 | npm registry |
| protobuf-es / connect-es | 2.17.0 / matching major | https://github.com/bufbuild/protobuf-es/releases |
| Astro | 7.4.13 | https://registry.npmjs.org/astro/latest |

## Best practices

1. **SPA + BFF instead of a React Server Components framework.** CVE-2025-55182 (CVSS 10) was a remote code execution flaw in React Server Components ([react.dev](https://react.dev/blog/2025/12/03/critical-security-vulnerability-in-react-server-components)). The consoles do not need server components, so they avoid that attack surface entirely.
2. **Sessions:**
   - Session state lives on the server.
   - The cookie is `__Host-` prefixed, `Secure`, `HttpOnly` and `SameSite=Strict`.
   - The session is regenerated whenever privilege changes.
   - High-value consoles (claims, risk, vault, studio) **sign out after 5 minutes idle**.
   - No JWTs are kept in the browser.
   - Source: [OWASP Session](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html).
3. **Single sign-on:** IAM Identity Center or an OIDC identity provider, with **FIDO2/passkey-only MFA**. If an external IdP is used, Identity Center's own MFA does not apply, so MFA must be enforced at that IdP ([AWS](https://docs.aws.amazon.com/singlesignon/latest/userguide/enable-mfa.html)).
4. **Content Security Policy:**
   - Nonce-based, with `strict-dynamic`.
   - `object-src 'none'`, `base-uri 'none'`, `frame-ancestors 'none'`.
   - Roll it out in Report-Only mode first.
   - Source: [OWASP CSP](https://cheatsheetseries.owasp.org/cheatsheets/Content_Security_Policy_Cheat_Sheet.html).
5. **Two-person approval** is enforced on the server (NIST 800-53 AC-5 and AC-3(2)):
   - the author cannot approve their own change
   - two distinct approvers are required
   - approvers rotate

   It applies to: publishing a question, releasing a payout, overriding KYC, and reversing a risk `block` decision.
6. **Audit log:** every console action is written to a hash-chained audit log with who, what, before and after values, approver, IP, device and request ID.
7. **ASVS 5.0:** Level 2 baseline. **Level 3** for studio, vault, claims and KYC views ([ASVS](https://asvs.owasp.org/)).
8. **API Security Top 10 2023:** test every console endpoint for API1 (broken object-level authorization) and API6 (sensitive business flows) ([OWASP API](https://api-security.owasp.org/editions/2023/en/0x11-t10)).
9. **Website i18n:** Astro's built-in i18n, with `sq` as the default, plus `en` and `mk` ([Astro i18n](https://docs.astro.build/en/guides/internationalization/)).

## Rules for this codebase

- **Feature switches UI (studio):**
  - The chat, reactions, extra lives, sign-ups and country eligibility switches each have a confirm dialog.
  - Every change is audit-logged.
  - The **live state of each switch is shown next to it**.
- The API client is generated from `protocol/` with protobuf-es. Never write API types by hand.
- No secrets go into console bundles. The BFF holds all credentials.

## Code

[`web-consoles/`](../../web-consoles/)

## Sources

[docs/research/05-auth-security-compliance-web.md](../../docs/research/05-auth-security-compliance-web.md)

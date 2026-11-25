# ADR-0005: Device-bound tokens (DPoP), refresh rotation, step-up before claims

- **Status:** Accepted (owner, 2026-11-25)
- **Date:** 2026-11-25

## Context ([research 05](../research/05-auth-security-compliance-web.md))
- NIST SP 800-63B-4 (final, July 2025) classes SMS/PSTN OTP as a **restricted** authenticator. Using one requires:
  - offering an alternative
  - notifying users of the risk
  - doing a risk assessment
  - having a migration plan
- OWASP advises against relying on SMS where there is financial risk. This app has a cash wallet.
- RFC 9700 (BCP 240): refresh tokens issued to public clients MUST be sender-constrained or rotated, with reuse detection.
- RFC 9449 (DPoP) binds tokens to a key held by the client.
- There is no SIM-swap data for Albania, Kosovo or North Macedonia.

## Decision
1. **Device key.** Each install creates a non-exportable key in the Secure Enclave or Android Keystore. The key is linked to the identity attested by App Attest or Play Integrity.
2. **Access tokens.** `at+jwt`, signed with EdDSA, valid for 10 minutes, bound to the device key with DPoP (`cnf.jkt`).
3. **Refresh tokens.**
   - Opaque, and rotated on every use.
   - Reuse of an old token revokes the whole token family.
   - Expire after 30 days of inactivity.
4. **Step-up before claims and payout changes.** The user unlocks the device key with biometrics, and the request carries a DPoP proof with a server nonce.
5. **Re-bound phone numbers.** If the number was re-bound in the last 30 days, KYC is re-checked before payout.
6. **SMS OTP** only bootstraps the device key.

## Consequences
- Stolen access or refresh tokens cannot be used from another device.
- DPoP has to be implemented on both clients and the server, with DPoP vectors added to the test suite.
- A lost device means signing in again with OTP. That re-bind triggers a KYC re-check before the next payout.

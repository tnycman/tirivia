# 11. Sign-out and the refresh token's DPoP key binding

Codex finding #9: `SignOut` checks that the request carries a valid DPoP proof, but never checks that the proof's key is the key the refresh token is bound to (`backend/internal/auth/api/session.go`, `sessions.Manager.Revoke`). Anyone holding a copy of a refresh token can therefore end that user's session from another device.

## Primary sources (verbatim, read 2026-11-27)

| Source | Text |
|---|---|
| [RFC 9449 §5](https://www.rfc-editor.org/rfc/rfc9449.html#section-5) | "the refresh token MUST be bound to the respective public key." |
| RFC 9449 §5 | "The binding MUST be validated when the refresh token is later presented to get new access tokens." |
| RFC 9449 §5 | "The implementation details of the binding of the refresh token are at the discretion of the authorization server." |
| RFC 9449 §1–13 | No sentence mentions token revocation or RFC 7009 (§11.8–13 read as a summary only, UNVERIFIED verbatim). |
| [RFC 7009 §2.1](https://www.rfc-editor.org/rfc/rfc7009.html#section-2.1) | "...then verifies whether the token was issued to the client making the revocation request." |
| RFC 7009 §2.1 | "If this validation fails, the request is refused" |
| [RFC 7009 §2.2](https://www.rfc-editor.org/rfc/rfc7009.html#section-2.2) | "The authorization server responds with HTTP status code 200 if the token has been revoked successfully ... or if the client submitted an invalid token." |
| RFC 7009 §2.1 | "the authorization server SHOULD also invalidate all access tokens based on the same authorization grant" (only "if the authorization server supports the revocation of access tokens") |
| [RFC 9700 §2.2.2](https://www.rfc-editor.org/rfc/rfc9700.html#section-2.2.2) | "Refresh tokens for public clients MUST be sender-constrained or use refresh token rotation as described in Section 4.14." |
| [RFC 9700 §4.14.2](https://www.rfc-editor.org/rfc/rfc9700.html#section-4.14.2) | Sender-constrained: "the authorization server cryptographically binds the refresh token to a certain client instance, e.g., by utilizing [RFC8705] or [RFC9449]." |

## What the sources settle

1. For a public client the DPoP key is what identifies "the client making the revocation request". RFC 7009 §2.1 says the server verifies the token was issued to that client and refuses the request if not. So sign-out must compare the proof's key thumbprint with the session's `jkt` and refuse on a mismatch.
2. An unknown refresh token gets a success response (RFC 7009 §2.2). The current `Revoke` already does this.
3. A refused request is not a revocation. RFC 7009 says "refused", so the family stays live on a mismatch.

## What the sources do not settle (OPEN)

- **Refresh with the wrong key.** RESOLVED by ID-7: refused without revoking, and checked before reuse detection. Before that fix, `Manager.check` revoked the whole family when the key does not match (`ReasonBinding`). RFC 9449 §5 requires the binding to be validated but does not say to revoke. ADR-0005, stack/05 rule 10 and research 05 §8 call for family revocation only on **reuse**. So this behaviour has no recorded source. It also lets someone holding a stolen refresh token, without the device key, end the user's session through `RefreshSession`. That is the same effect finding #9 describes, through a different RPC.
- **Access tokens after sign-out.** RFC 7009 says SHOULD, and only where access-token revocation exists. The access JWT lives 10 minutes (stack/05 rule 9), and the `jti` denylist (research 05 §10) is not built yet. This is outside finding #9.

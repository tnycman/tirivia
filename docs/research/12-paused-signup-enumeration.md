# 12. Paused sign-ups and phone-number enumeration

Codex finding #7. When `signups.enabled` is off, `otp.Service.Start` (`backend/internal/auth/otp/start.go`) answers `SIGNUPS_PAUSED` for a number with no account and `SENT` for a registered number. That answer tells the caller whether the number is registered.

## What the code does today

| Fact | Where |
|---|---|
| The switch check runs **before** the rate limits (`checkSignups`, then `rateLimited`), so these probes never count against the per-phone, per-device or per-range limits | `otp/start.go` `Start` |
| The attestation ticket is not consumed by `StartOtp`; it stays valid until its TTL, so one attested device can probe many numbers | `attest/service.go` `Validate` |
| An unregistered number returns without calling the SMS provider; a registered one waits for the send, so the response times differ as well as the result | `otp/start.go` `issue` |
| With sign-ups on, every allowed number gets `SENT`; new or returning is only revealed after the code is verified (`is_new_user`) | `otp/start.go`, `api` VerifyOtp |

## Primary sources (verbatim, read 2026-11-27)

| Source | Text |
|---|---|
| [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html), "Authentication and Error Messages" | "An application should respond (both HTTP and HTML) in a generic manner." |
| same, "Authentication Responses" | "The objective is to prevent the creation of a discrepancy factor, allowing an attacker to mount a user enumeration action against the application." |
| same | Processing time "can bring a discrepancy factor"; the fix is code that goes "through the same process no matter what the user or the password is" |
| same | Brute-force protection "prevents an attacker from applying the enumeration at scale"; "Usage of CAPTCHA can be applied to a feature for which a generic error message cannot be returned" |
| [OWASP Forgot Password Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html), "Forgot Password Request" | "Return a consistent message for both existent and non-existent accounts." |
| same | "Ensure that responses return in a consistent amount of time to prevent an attacker enumerating which accounts exist." This "could be achieved by using asynchronous calls or by making sure that the same logic is followed" |
| same | "Implement protections against excessive automated submissions such as rate-limiting on a per-account basis" |
| [NIST SP 800-63B-4](https://pages.nist.gov/800-63-4/sp800-63b.html) | No sentence on enumeration or account existence in the Abstract to §4.1 (read verbatim). The remainder was not read verbatim: UNVERIFIED |
| [NIST SP 800-63A-4](https://pages.nist.gov/800-63-4/sp800-63a.html) | No such sentence in the first ~40k characters. The remainder is UNVERIFIED |

The OWASP Testing Guide page on account enumeration redirected to the guide root and was not read.

## What the sources settle

1. A different answer for registered and unregistered numbers is an enumeration flaw (OWASP, both cheat sheets).
2. Where the answer can be made the same, it should be, and its timing too.
3. Where a distinct answer is kept, rate limiting is the stated control that stops enumeration "at scale".

## What the sources do not settle (owner decision)

Decided 2026-11-27 as ID-8 (option B): the distinct answer stays, and the switch check moves after the limits.

- Whether the paused state keeps its own answer (clear to a new player, rate-limited) or answers like a sent code (no leak; the new player waits for a code that never comes unless the app shows the paused state up front, from the public switch value, not per number).
- Equal timing needs the provider call off the request path for every number. That changes how send failures and the channel fallback (ID-1) are reported today.

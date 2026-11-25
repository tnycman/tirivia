# Build standards: the acceptance bar for every pull request

Master Plan §8 refers to a "standing 20-point checklist", but that checklist was not provided with the plan. This file is a **proposed** version, assembled from the research in `docs/research/`. Replace it if the owner already has one.

Reviewers check each item. CI enforces the ones marked ⚙.

**Scope and process**
1. The PR addresses one concern only, and the owner approves before merging (Master Plan §8).
2. ⚙ CI is green on lint, unit tests, protocol vectors, build and the security scan.
3. The PR description links the task in PLAN.md and any ADR or SPEC section it touches.

**Correctness**

4. New behaviour has tests, including the failure path. Bug fixes include a regression test.
5. ⚙ Protocol changes pass `buf lint` and `buf breaking`, include test vectors, and never send the correct answer before `reveal`.
6. Every mutating RPC is idempotent: it uses the natural key or an `Idempotency-Key` header.
7. Money is stored as integer minor units. The ledger is only ever appended to, and every transaction balances to zero.
8. ⚙ Migrations only expand, then contract. They never hold long locks, and none run in a show window.

**Security and privacy**

9. Every new endpoint authenticates and authorizes at the object level (API1) and is rate-limited.
10. ⚙ No secrets in code, config or bundles. Secrets Manager only.
11. No PII in logs, metrics labels, Sentry or analytics. PII columns are envelope-encrypted.
12. Sensitive flows (OTP, claims, payouts, switches) are covered by four-eyes approval, an audit log or step-up authentication as SPEC requires.
13. ⚙ Dependencies are pinned. Actions are pinned by SHA or lockfile. Dependency review and CodeQL report no high findings.

**Operability**

14. New code paths emit OTel traces and metrics. Alerts are added or updated where an SLO is affected.
15. A risky feature has a kill switch, or is covered by an existing one (`feature_switches`).
16. ⚙ Terraform passes fmt, validate, TFLint, Trivy and Checkov. The plan is posted on the PR. No click-ops.

**Performance**

17. Hot-path changes add no Postgres reads and no unsharded hot keys. Load-test impact is noted.
18. Mobile UI changes on the live screen keep the reaction layer at 55 fps or more on the reference phone (Macrobenchmark).

**Product quality**

19. User-facing strings are in Albanian and English through String Catalogs or Android resources. Accessibility labels are present. Accessibility tools are never penalised.
20. Docs are updated: README, SPEC, stack folder or runbook. If the change departs from a decision, it has an ADR.

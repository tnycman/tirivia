# 10 CI/CD and supply chain

## Role
Every pull request in the six repos runs lint, unit tests, protocol test vectors, a build, and a security scan (Master Plan §8). Releases are signed, and every artifact can be traced back to the source it was built from.

## Decision
- **GitHub Actions**, authenticating to AWS through **OIDC**, so no long-lived keys exist.
- Third-party actions are pinned through the **workflow lockfile** (GA since 28 Oct 2026), or by SHA.
- **CodeQL**, **OpenSSF Scorecard**, **dependency review**.
- **Renovate** for dependency updates.
- **Artifact attestations** for provenance.

## Facts (verified 2026-11-25)

| Item | Value | Source |
|---|---|---|
| OIDC trust | provider `token.actions.githubusercontent.com`, aud `sts.amazonaws.com`, job needs `id-token: write`. IAM rejects trust policies with no `sub` condition | https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/configuring-openid-connect-in-amazon-web-services |
| Workflow lockfile | `.github/workflows/actions.lock`. GitHub calls it "strongest protection… supersedes SHA pinning" | https://github.blog/changelog/2026-10-28-github-actions-workflow-lockfiles-are-generally-available |
| Trivy action compromise (Mar 2026) | ~75 of 76 tags rewritten. SHA-pinned users were not exposed | https://github.com/aquasecurity/trivy/security/advisories/GHSA-69fq-xp46-6x23 |
| Artifact attestations | `actions/attest@v4` reaches SLSA v1.0 Build L2; L3 needs a reusable workflow | https://docs.github.com/en/actions/concepts/security/artifact-attestations |
| Environment required reviewers | **Not available for private repos on Free/Pro/Team** | https://docs.github.com/en/actions/reference/workflows-and-actions/deployments-and-environments |
| Dependabot | Supports Terraform only up to 1.15.x | https://docs.github.com/en/code-security/dependabot/ecosystems-supported-by-dependabot/supported-ecosystems-and-repositories |
| CodeQL | Swift, Kotlin/Java, JS/TS, Go, Actions. Private repos need GitHub Code Security | https://docs.github.com/en/code-security/code-scanning/introduction-to-code-scanning/about-code-scanning-with-codeql |

## Best practices
1. **Token permissions:** set `permissions: {}` at workflow level and grant the minimum per job.
2. **Pinning:** keep the workflow lockfile committed, or pin every action by SHA and enforce it with the org policy.
3. **OIDC roles:** each account has a read-only *plan* role and a write *apply* role. The trust policy pins `sub` to `repo:<org>/<repo>:environment:<env>` and the `ref` to `refs/heads/main` for apply.
4. **Untrusted input:** never interpolate it (PR titles, branch names) into `run:` steps. Never use `pull_request_target` with a checkout of PR code. Scorecard's Dangerous-Workflow check covers both.
5. **Build provenance:** images are built in a reusable workflow and attested with `actions/attest@v4`. The deploy step checks the attestation with `gh attestation verify`.
6. **Renovate:** run it with a cooldown (minimum release age) so newly published packages are not picked up immediately.
7. **Secret scanning:** push protection enabled on every repo.
8. **Show-window freeze:** CI blocks deploys and migrations to prod from T-2 h until the show is `SETTLED`.

## Pipelines per repo

| Repo | PR gates | Main |
|---|---|---|
| protocol | `buf lint`, `buf breaking`, `buf generate` diff-clean, vector generation | Publish generated SDKs |
| backend | golangci-lint, go-arch-lint, `go test -race`, vectors, govulncheck, CodeQL, image build | Attest, deploy to dev, then staging (manual promote to prod) |
| ios | SwiftLint, unit and snapshot tests, vectors, build | TestFlight |
| android | lint, detekt, unit tests, vectors, Macrobenchmark (nightly) | Play internal track |
| web-consoles | eslint (TS6 alias), `tsc` (TS7), vitest, Playwright | Deploy to S3/CloudFront |
| infra | fmt, validate, TFLint, `trivy config`, Checkov, plan | Apply per environment |

## Changes vs Master Plan
- Approval gates: GitHub plan is an open decision. Required reviewers need Enterprise Cloud; the alternative is branch rules plus an OIDC trust pinned to `environment:prod`.
- Dependency updates: Renovate instead of Dependabot for `infra/`.

## Code
[`.github/`](../../.github/)

## Sources
[docs/research/03-aws-infra-devops.md](../../docs/research/03-aws-infra-devops.md)

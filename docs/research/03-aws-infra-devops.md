# Research 03: AWS platform, IaC, CI/CD, scaling, edge security, DR, data protection

> Verified against primary sources on **2026-11-25**. **UNVERIFIED** = could not confirm; *(reasoning)* = our inference rather than a documented statement.
> Feeds: [`stack/09-infrastructure-aws`](../../stack/09-infrastructure-aws/README.md), [`stack/10-ci-cd`](../../stack/10-ci-cd/README.md), [`stack/12-security`](../../stack/12-security/README.md), [ADR-0003](../adr/0003-standby-region-ireland.md).

## 1. Verified facts and limits

| Fact | Value | Source |
|---|---|---|
| Recommended OUs | Security, Infrastructure, Workloads (+ Sandbox, Policy Staging, Suspended, Deployments, Business Continuity) | https://docs.aws.amazon.com/whitepapers/latest/organizing-your-aws-environment/recommended-ous-and-accounts.html |
| Control Tower | Log Archive + Audit accounts; IAM Identity Center directory | https://docs.aws.amazon.com/controltower/latest/userguide/how-control-tower-works.html |
| Landing Zone Accelerator | AWS recommends layering it on top of Control Tower; sample config ≈ $430/month | https://docs.aws.amazon.com/solutions/latest/landing-zone-accelerator-on-aws/solution-overview.html |
| SCPs | Grant no permissions; do not affect the management account | https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_scps.html |
| Terraform licence | BUSL 1.1 from v1.6.0 | https://discuss.hashicorp.com/t/hashicorp-projects-changing-license-to-business-source-license-v1-1/57106 |
| OpenTofu | CNCF Sandbox (Apr 2025), MPL 2.0, native state encryption | https://www.cncf.io/projects/opentofu/ |
| S3 backend locking | `use_lockfile = true`; **DynamoDB locking deprecated** | https://developer.hashicorp.com/terraform/language/backend/s3 |
| AWS Prescriptive Guidance (PG) | Terraform ≥1.11, AWS provider 6.x; one backend per environment in its own account; SSE-KMS state; `allowed_account_ids` | https://docs.aws.amazon.com/prescriptive-guidance/latest/terraform-aws-provider-best-practices/backend.html |
| Secrets out of state | Ephemeral resources (1.10+), write-only `*_wo` args (1.11+) | https://docs.aws.amazon.com/prescriptive-guidance/latest/terraform-aws-provider-best-practices/security.html |
| tfsec | Merged into Trivy | https://github.com/aquasecurity/tfsec |
| **Trivy Action compromise** | Mar 2026: about 75 of 76 tags rewritten; **SHA-pinned workflows were not exposed** | https://github.com/aquasecurity/trivy/security/advisories/GHSA-69fq-xp46-6x23 |
| GitHub OIDC → AWS | `token.actions.githubusercontent.com`, aud `sts.amazonaws.com`, `id-token: write`; IAM rejects trust policies without a `sub` condition | https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/configuring-openid-connect-in-amazon-web-services |
| Workflow lockfiles | GA 28 Oct 2026 (`.github/workflows/actions.lock`); GitHub says they supersede SHA pinning | https://github.blog/changelog/2026-10-28-github-actions-workflow-lockfiles-are-generally-available |
| **Env required reviewers** | Private repos on Free/Pro/Team: **not available** (public repos only) | https://docs.github.com/en/actions/reference/workflows-and-actions/deployments-and-environments |
| Artifact attestations | `actions/attest@v4` = SLSA v1.0 Build L2; L3 via reusable workflow | https://docs.github.com/en/actions/concepts/security/artifact-attestations |
| CodeQL | Swift, Kotlin/Java, JS/TS, Go, Actions; private repos need GitHub Code Security | https://docs.github.com/en/code-security/code-scanning/introduction-to-code-scanning/about-code-scanning-with-codeql |
| Dependabot | gradle, swift, npm, docker, actions; **Terraform only ≤1.15.x** | https://docs.github.com/en/code-security/dependabot/ecosystems-supported-by-dependabot/supported-ecosystems-and-repositories |
| Fargate vCPU quota | **Default 6 per region** | https://docs.aws.amazon.com/general/latest/gr/ecs-service.html |
| Fargate launch rate eu-central-1 | Burst 100, sustained 20/s; 500 tasks/min per service | same |
| **Fargate launch rate eu-south-1 (Milan)** | **Burst 25, sustained 5/s; 125 tasks/min per service (not adjustable)** | same |
| SOCI / ARM64 | ARM64 supported; SOCI lazy loading only worth it above ~250 MiB images | https://docs.aws.amazon.com/AmazonECS/latest/developerguide/fargate-tasks-services.html |
| ALB LCU reservation | Min 100, max 15,000 per ALB; **regional default quota 0**; can take hours; decrease only twice a day | https://docs.aws.amazon.com/elasticloadbalancing/latest/application/capacity-unit-reservation.html |
| CloudFront for POST APIs | POST is not cached; default origin keep-alive 5 s; **origin failover only for GET/HEAD/OPTIONS**; Origin Shield not suited to dynamic traffic | https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/high_availability_origin_failover.html |
| WAF rate-based rules | Windows 60/120/300/600 s; detection lag "usually below 30 s"; aggregate by IP, header, JA3/JA4, ASN | https://docs.aws.amazon.com/waf/latest/developerguide/waf-rule-statement-type-rate-based-caveats.html |
| WAF ACFP / Bot Control | Fake account-creation detection; Bot Control Targeted can be scoped | https://docs.aws.amazon.com/waf/latest/developerguide/waf-acfp.html |
| SMS Protect | AWS End User Messaging: Block/Filter/Monitor, $0.01 per message | https://docs.aws.amazon.com/sms-voice/latest/userguide/filter-and-monitor-messages.html |
| Shield Advanced | $3,000/month, 1-year commitment; SRT needs Business/Enterprise Support | https://aws.amazon.com/shield/pricing/ |
| DR strategies | Backup/restore → pilot light → warm standby → hot/active-active; fail over with **data-plane** operations (ARC routing controls); raise quotas in the DR region | https://docs.aws.amazon.com/whitepapers/latest/disaster-recovery-workloads-on-aws/disaster-recovery-options-in-the-cloud.html |
| Aurora Global DB | Lag typically <1 s; switchover RPO 0 (<30 s on qualifying versions); managed failover "within a few minutes" | https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-global-database-disaster-recovery.html |
| Aurora Global DB limits | **No Secrets Manager managed master password; no Auto Scaling on secondaries** | https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-global-database.html |
| ElastiCache Global Datastore | Valkey supported; **manual promote, same account only** | https://docs.aws.amazon.com/AmazonElastiCache/latest/dg/Redis-Global-Datastores-Getting-Started.html |
| **IVS regions** | Frankfurt and Ireland in the EU; **not Milan** | https://docs.aws.amazon.com/general/latest/gr/ivs.html |
| S3 Object Lock | Compliance mode cannot be bypassed even by root; Governance can be bypassed with a permission | https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lock.html |
| FIS actions | ecs stop-task, Valkey AZ power interruption, Aurora failover, network disruption | https://docs.aws.amazon.com/fis/latest/userguide/fis-actions-reference.html |
| Game days | REL12-BP05 | https://docs.aws.amazon.com/wellarchitected/latest/reliability-pillar/rel_testing_resiliency_game_days_resiliency.html |
| **Managed Grafana regions** | Frankfurt yes, **Milan no** | https://docs.aws.amazon.com/general/latest/gr/grafana-service.html |
| **Games Industry Lens** | Published Dec 2025: GAMEOPS03-BP03 "Load test early and often", **GAMEOPS03-BP05 "Pre-scale infrastructure"** | https://docs.aws.amazon.com/wellarchitected/latest/games-industry-lens/gameops03-bp05.html |
| Fargate price eu-central-1 | x86 $0.04656/vCPU-h; **ARM $0.03725/vCPU-h**, $0.00409/GB-h | https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonECS/current/eu-central-1/index.csv |
| ElastiCache price eu-central-1 | r7g.xlarge Valkey $0.42/h vs Redis $0.525/h | https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonElastiCache/current/eu-central-1/index.csv |

## 2. Best practices adopted

**Accounts**
1. Control Tower (Organizations + Identity Center). No workloads in the management account. Humans sign in only through Identity Center, with no IAM users.
2. OUs: Security (log-archive, audit), Infrastructure (network, shared-services), Workloads/SDLC (dev, staging), Workloads/Prod (prod), Business Continuity (backup-vault), Sandbox, Policy Staging.
3. **dev/staging/prod alone is not enough.** You also need tamper-resistant logs, a delegated security admin, and a place to test SCPs.
4. SCPs: deny all regions except the ones in use (eu-central-1, eu-west-1, plus us-east-1 for CloudFront/WAF/ACM). Deny leaving the org, disabling CloudTrail/GuardDuty, and governance-retention bypass outside break-glass.

**IaC**
5. Pin Terraform `~> 1.17` and AWS provider `~> 6.74`, and commit `.terraform.lock.hcl`.
6. S3 backend with `use_lockfile`: one KMS-encrypted, versioned, TLS-only state bucket per account. Set `allowed_account_ids` on every provider.
7. Keep secrets out of state with ephemeral resources and write-only arguments.
8. Static checks: fmt/validate, TFLint (AWS ruleset), `trivy config`, Checkov. Scheduled drift plan.
9. One module per capability, not per resource. No provider blocks inside modules.

**CI/CD**
10. OIDC with per-account roles. Pin `sub` (repo + environment) and `ref`. Use separate plan (read-only) and apply roles.
11. `permissions: {}` at workflow level; grant per job.
12. Commit the workflow lockfile, or SHA-pin every action and enforce it with org policy.
13. Run Scorecard, CodeQL and Renovate (Terraform ≥1.16 support, cooldowns).
14. Add build provenance with `actions/attest@v4` from a reusable workflow (SLSA L3). Verify it at deploy time.

**Compute / scaling**
15. Scheduled MinCapacity raise at T-60 (images pulled, health checks green by T-20); target tracking on top.
16. Request quota increases in every account and region that will run at scale.
17. Use ARM64/Graviton (≈20% cheaper), small images, and larger tasks (2–4 vCPU) so fewer launches are needed.
18. Reserve ALB LCUs from the load test's `PeakLCUs`, several hours before the show.

**Edge / security**
19. CloudFront + WAF in front of the API: Shield Standard, pooled keep-alive (raise from 5 s), HTTP/2 to origin, no Origin Shield.
20. WAF rate limits keyed on a signed session header plus IP. Albanian mobile carriers likely use CGNAT, so IP-only limits would hit innocent players *(reasoning)*. ACFP on sign-up. Bot Control Targeted only on auth paths.
21. Per-player answer and reaction limits are enforced in the app and Valkey, because WAF takes about 30 s or more to react.
22. Shield Advanced is justified for prod: all value is concentrated in one daily 20-minute window. It requires Business Support.

**DR**: see §3 and ADR-0003.

**Data protection**
23. KYC bucket: SSE-KMS with a dedicated CMK and Bucket Keys; key policy limited to the KYC service role; deny unencrypted puts; Object Lock in **Governance** mode (Compliance mode would block lawful erasure, *(reasoning)*); lifecycle rules that match the retention schedule; Macie; CloudTrail data events.
24. Secrets Manager with multi-region replication and rotation.
25. Quarterly FIS game days with CloudWatch alarm stop conditions.

## 3. Conflicts with Master Plan v1.1 and adopted fixes

| # | Issue | Adopted fix |
|---|---|---|
| 1 | **Milan cannot meet a cold 5-minute RTO.** Fargate there launches 125 tasks/min per service (fixed) and 5/s per account. There is also no IVS, no IVS Chat and no Managed Grafana. | **Standby = Ireland (eu-west-1) for all planes** ([ADR-0003](../adr/0003-standby-region-ireland.md)). Ireland has full launch rates, IVS and Chat. KDS on-demand reaches 10 GB/s there. Aurora Global DB and Global Datastore are both supported. **Warm standby during the show window**: Ireland is pre-scaled to about 30–50% on the same schedule. *Owner must confirm: plan §5 names Milan.* |
| 2 | Default quotas: 6 Fargate vCPU, 0 reserved LCUs, 15k IVS views, 50k chat connections. | Quota requests in prod and staging (for the 600k test) go into **Phase 0**. |
| 3 | GitHub required reviewers are unavailable for private repos on Team. | Either use GitHub Enterprise Cloud, or protect deployments with deployment-branch rules, an IAM trust pinned to `environment:prod` + `ref:main`, and a CODEOWNERS-protected main. **Owner decision** (open item). |
| 4 | Global Datastore is same-account only. | The standby region lives **inside the prod account** (separate VPC). Backups are isolated in the backup-vault account. |
| 5 | Aurora Global DB does not support Secrets Manager managed passwords. | Self-managed secret, rotation Lambda, `password_wo`, secret replicated to Ireland. |
| 6 | Aurora secondaries do not auto-scale; Global Datastore promotion is manual. | Size the standby statically. Script the promotion in the region-failover runbook, or treat Valkey as rebuildable (live state is reconstructed from the last cue plus Kinesis). |
| 7 | Dependabot does not cover Terraform 1.17. | Use Renovate for `infra/`. |
| 8 | Mutable action tags (Trivy incident). | Workflow lockfile or SHA pins; Trivy run as a pinned binary or image digest. |
| 9 | WAF too slow for per-request abuse. | App-level limits in Valkey; WAF is the coarse layer. |
| 10 | Terraform or OpenTofu. | **Terraform 1.17.x**: it matches the AWS PG examples, and internal use is permitted under BUSL (the FAQ wording is UNVERIFIED). Switch to OpenTofu 1.13.x only if an open-source licence is required. Do not mix the two. |
| 11 | CloudFront cannot fail over POST requests between origins. | Regional failover of the API is **DNS-level (Route 53 ARC routing controls)**, not a CloudFront origin group. |

## 4. Versions

| Tool | Version | Source |
|---|---|---|
| Terraform | 1.17.2 | https://github.com/hashicorp/terraform/releases |
| OpenTofu (alt) | 1.13.10 | https://github.com/opentofu/opentofu/releases |
| AWS provider | 6.74.0 | https://github.com/hashicorp/terraform-provider-aws/releases |
| TFLint | 0.66.0 | https://github.com/terraform-linters/tflint/releases |
| Trivy | 0.76.2 | https://github.com/aquasecurity/trivy/releases |
| Checkov | 3.4.10 | https://github.com/bridgecrewio/checkov/releases |
| actions/attest | v4 | https://docs.github.com/en/actions/concepts/security/artifact-attestations |

## 5. Cost sanity (Frankfurt prices verified; task counts are assumptions)

- Fargate: 300 ARM tasks × 2 vCPU / 4 GB × 1 h per show ≈ **$27 per show, about $820/month**. A warm standby in Ireland adds 30–50% of that.
- ElastiCache: 6 × r7g.xlarge Valkey always on ≈ **$1,840/month**, roughly doubled with a Global Datastore in Ireland.
- Shield Advanced: **$3,000/month**, probably the largest fixed infra line.

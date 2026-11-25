# 09 Cloud infrastructure (AWS)

## Role
All cloud resources:
- accounts and landing zone
- network
- compute and data
- edge and DR

Everything is managed as code.

## Decision
- **AWS from day one** (Master Plan §5).
- **Control Tower** multi-account landing zone.
- **Terraform 1.17** for infrastructure as code ([ADR-0008](../../docs/adr/0008-terraform-not-opentofu.md)).
- **ECS Fargate on ARM64** for compute.
- **CloudFront + WAF + Shield Advanced** at the edge.
- **Frankfurt is primary, Ireland is the warm standby** ([ADR-0003](../../docs/adr/0003-standby-region-ireland.md)).

## Pinned versions (2026-11-25)

| Tool | Version | Source |
|---|---|---|
| Terraform | 1.17.2 (`~> 1.17`) | https://github.com/hashicorp/terraform/releases |
| AWS provider | 6.74.0 (`~> 6.74`) | https://github.com/hashicorp/terraform-provider-aws/releases |
| TFLint | 0.66.0 | https://github.com/terraform-linters/tflint/releases |
| Trivy | 0.76.2 (pinned binary or digest, **never a mutable action tag**) | https://github.com/aquasecurity/trivy/releases |
| Checkov | 3.4.10 | https://github.com/bridgecrewio/checkov/releases |

## Account structure
Based on the AWS whitepaper *Organizing Your AWS Environment*:

| OU | Accounts |
|---|---|
| Security | `log-archive`, `audit` (delegated admin for GuardDuty, Security Hub, Macie) |
| Infrastructure | `network`, `shared-services` (ECR, Route 53) |
| Workloads/SDLC | `dev`, `staging` |
| Workloads/Prod | `prod` (Frankfurt + Ireland standby in the same account) |
| Business Continuity | `backup-vault` |
| Sandbox, Policy Staging | as needed |

A dev/staging/prod split alone is **not enough**. It leaves no place for tamper-resistant logs, no delegated security admin, and nowhere to test SCPs.

Source: [whitepaper](https://docs.aws.amazon.com/whitepapers/latest/organizing-your-aws-environment/recommended-ous-and-accounts.html)

## Best practices
1. **State:**
   - S3 backend with `use_lockfile = true`. DynamoDB locking is deprecated.
   - One bucket per account, with KMS encryption, versioning, a TLS-only policy and Block Public Access.

   Source: [S3 backend](https://developer.hashicorp.com/terraform/language/backend/s3)
2. **Provider guards:**
   - Set `allowed_account_ids` in every provider block.
   - Keep secrets out of state with ephemeral resources and `*_wo` write-only arguments.

   Source: [AWS Prescriptive Guidance](https://docs.aws.amazon.com/prescriptive-guidance/latest/terraform-aws-provider-best-practices/security.html)
3. **Modules:**
   - One module per capability (e.g. `kyc-storage`), not per resource.
   - No provider blocks inside modules.
   - Keep nesting flat.

   Source: [style](https://developer.hashicorp.com/terraform/language/style)
4. **SCPs:**
   - Region allow-list: eu-central-1, eu-west-1, plus us-east-1 for CloudFront, WAF and ACM.
   - Deny leaving the org.
   - Deny disabling CloudTrail or GuardDuty.
   - Deny `s3:BypassGovernanceRetention` except for the break-glass role.
5. **Pre-scale before each show** (Games Industry Lens **GAMEOPS03-BP05**):
   - Scheduled action raises MinCapacity at **T-60**.
   - Target tracking stays on top of that floor.
   - ALB LCU reservation goes in at **≥ T-4 h**, because it can take hours.
   - Source: [scheduled scaling](https://docs.aws.amazon.com/autoscaling/application/userguide/application-auto-scaling-scheduled-scaling.html), [LCU reservation](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/capacity-unit-reservation.html)
6. **Quotas** to request in every account and region that runs at scale:

   | Quota | Default |
   |---|---|
   | Fargate vCPU | 6 |
   | ALB reserved LCU | 0 |
   | IVS concurrent views | 15k |
   | IVS chat connections | 50k |
   | Play Integrity requests | 10k/day |

7. **Compute:** ARM64/Graviton, which is about 20% cheaper ($0.03725 vs $0.04656 per vCPU-hour in Frankfurt). Use 2–4 vCPU tasks so fewer launches are needed. Avoid Cloud Map on hot services, because it caps a service at 1,000 tasks.
8. **Edge:**
   - CloudFront in front of the API: TLS termination, Shield Standard, pooled keep-alive with origin keep-alive raised above 5 s, HTTP/2 to the origin.
   - **No Origin Shield for the API.**
   - **CloudFront cannot fail over POST requests**, so regional failover is done in DNS with ARC.

   Source: [origin failover](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/high_availability_origin_failover.html)
9. **WAF:**
   - Rate-based rules keyed by session header + IP. Albanian mobile networks use CGNAT, so IP alone would block real users.
   - ACFP on sign-up.
   - Bot Control Targeted, scoped to auth paths only.
   - WAF reacts in ≥ 30 s, so per-request limits live in the app, not in WAF.
10. **Shield Advanced:** justified because all revenue is concentrated in one 20-minute window. Costs $3,000/month and requires Business Support.
11. **DR (warm standby):**
    - Aurora Global DB secondary that is not headless.
    - Secrets replicated to the standby region.
    - S3 cross-region replication for KYC and media.
    - IVS channel and chat pre-created in Ireland.
    - Quarterly **FIS game days**.

    Source: [DR whitepaper](https://docs.aws.amazon.com/whitepapers/latest/disaster-recovery-workloads-on-aws/disaster-recovery-options-in-the-cloud.html)
12. **KYC bucket:**
    - Dedicated CMK.
    - Deny unencrypted puts.
    - **Object Lock in Governance mode.** Compliance mode would block lawful erasure.
    - Macie and CloudTrail data events.

## Rules for this codebase
- No console click-ops in staging or prod. Break-glass use is logged and reviewed.
- Every stack in `live/` has its own state file.
- `terraform plan` output is posted on every pull request. Apply happens only from `main` through OIDC.
- Tags on every resource: `app`, `env`, `owner`, `cost-center`, `data-class`.

## Changes vs Master Plan
- The standby region is Ireland, not Milan.
- The landing zone has more accounts than dev/staging/prod.
- MinCapacity is raised at T-60 instead of T-20, plus an LCU reservation.
- Shield Advanced is recommended.

## Code
[`infra/`](../../infra/)

## Sources
[docs/research/03-aws-infra-devops.md](../../docs/research/03-aws-infra-devops.md)

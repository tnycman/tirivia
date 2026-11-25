# infra: AWS infrastructure as code

Uses **Terraform 1.17** and **AWS provider 6.x** ([ADR-0008](../docs/adr/0008-terraform-not-opentofu.md)) on a Control Tower multi-account landing zone. Frankfurt is the primary region and Ireland is the warm standby ([ADR-0003](../docs/adr/0003-standby-region-ireland.md)). Best practices are in [`stack/09-infrastructure-aws`](../stack/09-infrastructure-aws/README.md).

```
infra/
├── bootstrap/                    # run once per account, using local state
│   ├── state-backend/            #   KMS-encrypted, versioned, TLS-only state bucket (native S3 locking)
│   └── github-oidc/              #   OIDC provider + read-only plan role + apply role (main + environment)
├── org/                          # management account: OUs, SCPs (region allow-list), Identity Center
├── modules/                      # one module per capability
│   ├── ivs-channel/              #   ADVANCED_HD channel + recording config (per region)
│   ├── scheduled-scaling/        #   pre-show MinCapacity at T-60 + target tracking
│   ├── network/  ecs-service/  aurora-global/  valkey/  edge-api/ (CloudFront + WAF)
│   ├── kyc-storage/ (CMK, Object Lock governance, Macie)  media-buckets/  kinesis-analytics/
│   └── observability/  fis-experiments/  dr-routing/ (Route 53 ARC)
├── live/<account>/<region>/<stack>/   # one state file per stack
│   ├── prod/eu-central-1/        #   primary
│   ├── prod/eu-west-1/           #   warm standby (IVS + chat pre-created; same account for Global Datastore)
│   └── prod/global/              #   CloudFront, WAF (us-east-1 alias), Shield Advanced
├── policies/                     # Checkov custom policies
├── .tflint.hcl  renovate.json
```

## Verified on 2026-11-25

Ran `terraform fmt -check`, `init` and `validate` with `hashicorp/terraform:1.17.2` in Docker:

| Stack | Result |
|---|---|
| `modules/ivs-channel` | valid against `hashicorp/aws v6.74.0` |
| `modules/scheduled-scaling` | valid |
| `bootstrap/state-backend` | valid |
| `bootstrap/github-oidc` | valid |

## Rules

- **Accounts:** every provider block sets `allowed_account_ids`.
- **Lock files:** commit `.terraform.lock.hcl` for every stack.
- **Secrets:** they never enter Terraform state. Use ephemeral resources and write-only `*_wo` arguments.
- **Changes:** no console click-ops in staging or prod. `plan` runs on the PR with the read-only role; `apply` runs only from `main` through the apply role.
- **Freeze:** no applies to prod from T-2 h before a show until the show is SETTLED.
- **Phase 0 quota requests:** Fargate vCPU, reserved LCUs, IVS concurrent views, and chat connections. See PLAN L1–L4.

# ADR-0008: Use Terraform 1.17, not OpenTofu

- **Status:** Accepted (owner, 2026-11-25)
- **Date:** 2026-11-25

## Context ([research 03](../research/03-aws-infra-devops.md))
- **Licensing:** Terraform has been under BUSL 1.1 since v1.6. OpenTofu is the MPL-2.0 fork and a CNCF Sandbox project.
- **AWS guidance:** AWS Prescriptive Guidance targets Terraform 1.11 or later with AWS provider 6.x.
- **State locking:** the S3 backend now locks natively with `use_lockfile`. DynamoDB locking is deprecated.
- **Dependency updates:** Dependabot does not support Terraform versions above 1.15.

## Decision
- Pin Terraform `~> 1.17` and the AWS provider `~> 6.74`.
- Store state in the S3 backend with `use_lockfile`.
- Let Renovate keep Terraform and provider versions current.
- Do not mix Terraform and OpenTofu in this codebase.
- Revisit this decision if we come to need an open-source licence or native state encryption.

## Consequences
- **Licence:** BUSL permits internal use. We did not verify the exact wording of HashiCorp's licensing FAQ.
- **State encryption:** we rely on SSE-KMS on the state bucket, since Terraform has no native state encryption (OpenTofu does).

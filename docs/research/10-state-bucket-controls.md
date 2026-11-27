# Research 10: Terraform state bucket controls (Checkov findings)

Read 2026-11-27. Decision: [OP-7](../decision-register.md). Trigger: Checkov 3.4.10 failed five checks on `infra/bootstrap/state-backend` in PR #4.

## Sources and what they say

| Topic | Quote | Source |
|---|---|---|
| Versioning | "highly recommended that you enable Bucket Versioning on the S3 bucket to allow for state recovery" | [HashiCorp S3 backend](https://developer.hashicorp.com/terraform/language/backend/s3) |
| Best-practice status | "treat them as helpful recommendations rather than prescriptions" | [S3 security best practices](https://docs.aws.amazon.com/AmazonS3/latest/userguide/security-best-practices.html) |
| Replication | "Consider using S3 Cross-Region Replication"; "CRR requires both the source and destination S3 buckets to have versioning enabled" | same |
| Lifecycle | "identify buckets that don't have S3 Lifecycle rules to abort incomplete multipart uploads" | same |
| Default key policy | "gives the AWS account that owns the KMS key permission to use IAM policies to allow access to all AWS KMS operations" | [KMS default key policy](https://docs.aws.amazon.com/kms/latest/developerguide/key-policy-default.html) |
| Access logs | "delivered on a best-effort basis"; destination "must be in the same AWS Region and AWS account"; S3 destination supports SSE-S3 only | [S3 server access logging](https://docs.aws.amazon.com/AmazonS3/latest/userguide/ServerLogs.html) |
| KMS objects | "By default, Amazon S3 doesn't replicate objects that are encrypted with SSE-KMS"; the replica key must be "in the same AWS Region as the destination bucket" | [Replicating KMS objects](https://docs.aws.amazon.com/AmazonS3/latest/userguide/replication-config-for-kms-objects.html) |
| KMS role grants | "grant `kms:Decrypt` permissions for the AWS KMS key that protects the source object"; "grant `kms:Encrypt` permissions for the AWS KMS key that will protect the object replica" | same |
| Bucket Keys | the replica key policy "must include the `kms:Decrypt` permission for the calling principal"; the encryption context is "the bucket's Amazon Resource Name (ARN)" | same |
| Role actions | source bucket: `s3:GetReplicationConfiguration`, `s3:ListBucket`; source objects: `s3:GetObjectVersionForReplication`, `s3:GetObjectVersionAcl`, `s3:GetObjectVersionTagging`; replica objects: `s3:ReplicateObject`, `s3:ReplicateDelete` (no separate `s3:ReplicateTags` needed) | [Replication permissions](https://docs.aws.amazon.com/AmazonS3/latest/userguide/setting-repl-config-perm-overview.html) |
| Delete markers | by default "Amazon S3 adds a delete marker in the source bucket only" | [Delete marker replication](https://docs.aws.amazon.com/AmazonS3/latest/userguide/delete-marker-replication.html) |
| Terraform | `filter {}` is required for a rule with no filter; `delete_marker_replication` needs a V2 rule (with `filter`); provider v6 sets the region per resource | [aws_s3_bucket_replication_configuration](https://raw.githubusercontent.com/hashicorp/terraform-provider-aws/main/website/docs/r/s3_bucket_replication_configuration.html.markdown) |
| Skips | `checkov:skip=<check_id>:<suppression_comment>` inside the resource block | [Checkov suppressions](https://www.checkov.io/2.Basics/Suppressing%20and%20Skipping%20Policies.html) |

## How each check decides (Checkov source)

From [checkov graph_checks/aws](https://github.com/bridgecrewio/checkov/tree/main/checkov/terraform/checks/graph_checks/aws):

- **CKV2_AWS_64** (`KmsKeyPolicyIsDefined.yaml`): passes when `policy` is set or an `aws_kms_key_policy` exists.
- **CKV2_AWS_61** (`S3BucketLifecycle.yaml`): passes when any lifecycle configuration is attached; rule content is not inspected.
- **CKV_AWS_144** (`S3BucketReplicationConfiguration.yaml`): passes when an attached replication rule has `status = Enabled`.
- **CKV2_AWS_62** (`S3BucketEventNotifications.yaml`): passes when an `aws_s3_bucket_notification` is attached.
- **CKV_AWS_18** (`S3BucketLogging.yaml`): passes when logging is configured.

## Not stated by any source read

- Costs of replication, access logging or CloudTrail data events.
- A recommended day count for aborting incomplete uploads (so it is the variable `abort_incomplete_upload_days`).
- Event notifications for a state bucket (HashiCorp and AWS are silent).

## Reasoning (not sourced)

- During a Frankfurt outage Terraform still needs its state to change the Ireland standby, so the state is replicated there.
- `use_lockfile` deletes its lock object after each run. Replicating delete markers keeps stale locks out of the replica.

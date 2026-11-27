# State replica in var.standby_region, written only by S3 replication (OP-7, ADR-0003).

resource "aws_kms_key" "replica" {
  region              = var.standby_region
  description         = "Terraform state replica (${var.env})"
  enable_key_rotation = true
  policy              = local.replica_key_policy
}

resource "aws_s3_bucket" "replica" {
  #checkov:skip=CKV_AWS_144:This bucket is the cross-region replica of the state bucket (OP-7)
  #checkov:skip=CKV2_AWS_62:No consumer for state-bucket events; neither AWS nor HashiCorp recommends them here (OP-7)
  #checkov:skip=CKV_AWS_18:Audited by CloudTrail data events from the account baseline, not best-effort access logs (OP-7)
  region = var.standby_region
  bucket = "tirivia-tfstate-${var.env}-${var.account_id}-${var.standby_region}"
}

resource "aws_s3_bucket_versioning" "replica" {
  region = var.standby_region
  bucket = aws_s3_bucket.replica.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "replica" {
  region = var.standby_region
  bucket = aws_s3_bucket.replica.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "aws:kms"
      kms_master_key_id = aws_kms_key.replica.arn
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_public_access_block" "replica" {
  region                  = var.standby_region
  bucket                  = aws_s3_bucket.replica.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_policy" "replica" {
  region = var.standby_region
  bucket = aws_s3_bucket.replica.id
  policy = data.aws_iam_policy_document.tls_only["replica"].json
}

resource "aws_s3_bucket_lifecycle_configuration" "replica" {
  region = var.standby_region
  bucket = aws_s3_bucket.replica.id
  rule {
    id     = "abort-incomplete-uploads"
    status = "Enabled"
    filter {}
    abort_incomplete_multipart_upload {
      days_after_initiation = var.abort_incomplete_upload_days
    }
  }
}

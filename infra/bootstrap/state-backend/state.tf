# Primary state bucket in var.region (OP-7).

resource "aws_kms_key" "state" {
  description         = "Terraform state (${var.env})"
  enable_key_rotation = true
  policy              = local.root_key_policy
}

resource "aws_s3_bucket" "state" {
  #checkov:skip=CKV2_AWS_62:No consumer for state-bucket events; neither AWS nor HashiCorp recommends them here (OP-7)
  #checkov:skip=CKV_AWS_18:Audited by CloudTrail data events from the account baseline, not best-effort access logs (OP-7)
  bucket = "tirivia-tfstate-${var.env}-${var.account_id}"
}

resource "aws_s3_bucket_versioning" "state" {
  bucket = aws_s3_bucket.state.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "state" {
  bucket = aws_s3_bucket.state.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "aws:kms"
      kms_master_key_id = aws_kms_key.state.arn
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_public_access_block" "state" {
  bucket                  = aws_s3_bucket.state.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_policy" "state" {
  bucket = aws_s3_bucket.state.id
  policy = data.aws_iam_policy_document.tls_only["state"].json
}

resource "aws_s3_bucket_lifecycle_configuration" "state" {
  bucket = aws_s3_bucket.state.id
  rule {
    id     = "abort-incomplete-uploads"
    status = "Enabled"
    filter {}
    abort_incomplete_multipart_upload {
      days_after_initiation = var.abort_incomplete_upload_days
    }
  }
}

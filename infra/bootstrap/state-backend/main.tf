# One-time bootstrap per account: Terraform state bucket with native S3 locking (use_lockfile; DynamoDB locking is deprecated).

terraform {
  required_version = "~> 1.17"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.74"
    }
  }
}

variable "account_id" {
  description = "Guard rail: the provider refuses to run in any other account."
  type        = string
}

variable "region" {
  type    = string
  default = "eu-central-1"
}

variable "env" {
  description = "Account name, e.g. dev, staging, prod."
  type        = string
}

provider "aws" {
  region              = var.region
  allowed_account_ids = [var.account_id]
  default_tags {
    tags = { app = "tirivia", env = var.env, owner = "platform", "data-class" = "internal" }
  }
}

resource "aws_kms_key" "state" {
  description         = "Terraform state (${var.env})"
  enable_key_rotation = true
}

resource "aws_s3_bucket" "state" {
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

data "aws_iam_policy_document" "tls_only" {
  statement {
    sid     = "DenyInsecureTransport"
    effect  = "Deny"
    actions = ["s3:*"]
    resources = [
      aws_s3_bucket.state.arn,
      "${aws_s3_bucket.state.arn}/*",
    ]
    principals {
      type        = "*"
      identifiers = ["*"]
    }
    condition {
      test     = "Bool"
      variable = "aws:SecureTransport"
      values   = ["false"]
    }
  }
}

resource "aws_s3_bucket_policy" "state" {
  bucket = aws_s3_bucket.state.id
  policy = data.aws_iam_policy_document.tls_only.json
}

output "backend_config" {
  description = "Paste into each stack's backend block."
  value = {
    bucket       = aws_s3_bucket.state.id
    region       = var.region
    encrypt      = true
    kms_key_id   = aws_kms_key.state.arn
    use_lockfile = true
  }
}

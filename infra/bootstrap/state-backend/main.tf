# One-time bootstrap per account: Terraform state bucket with native S3 locking, replicated to the standby region (OP-7).

terraform {
  required_version = "~> 1.17"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.74"
    }
  }
}

provider "aws" {
  region              = var.region
  allowed_account_ids = [var.account_id]
  default_tags {
    tags = { app = "tirivia", env = var.env, owner = "platform", "data-class" = "internal" }
  }
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

output "standby_backend_config" {
  description = "Backend block for running Terraform against the replica during a regional failover (ADR-0003)."
  value = {
    bucket       = aws_s3_bucket.replica.id
    region       = var.standby_region
    encrypt      = true
    kms_key_id   = aws_kms_key.replica.arn
    use_lockfile = true
  }
}

# Key and bucket policies (OP-7). The root statement is AWS's default key policy, written out explicitly.

locals {
  root_statement = {
    Sid       = "Enable IAM User Permissions"
    Effect    = "Allow"
    Principal = { AWS = "arn:aws:iam::${var.account_id}:root" }
    Action    = "kms:*"
    Resource  = "*"
  }
  root_key_policy = jsonencode({ Version = "2012-10-17", Statement = [local.root_statement] })
  replica_key_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [local.root_statement, {
      Sid       = "AllowStateReplication"
      Effect    = "Allow"
      Principal = { AWS = aws_iam_role.replication.arn }
      Action    = ["kms:Encrypt", "kms:Decrypt"]
      Resource  = "*"
    }]
  })
  buckets = { state = aws_s3_bucket.state.arn, replica = aws_s3_bucket.replica.arn }
}

data "aws_iam_policy_document" "tls_only" {
  for_each = local.buckets
  statement {
    sid       = "DenyInsecureTransport"
    effect    = "Deny"
    actions   = ["s3:*"]
    resources = [each.value, "${each.value}/*"]
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

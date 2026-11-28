# S3 replication of the state bucket to the standby replica, KMS objects and delete markers included (OP-7).

resource "aws_iam_role" "replication" {
  name = "trivia-tfstate-replication-${var.env}"
  assume_role_policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [{ Effect = "Allow", Principal = { Service = "s3.amazonaws.com" }, Action = "sts:AssumeRole" }]
  })
}

data "aws_iam_policy_document" "replication" {
  statement {
    sid       = "SourceBucket"
    actions   = ["s3:GetReplicationConfiguration", "s3:ListBucket"]
    resources = [aws_s3_bucket.state.arn]
  }
  statement {
    sid       = "SourceObjects"
    actions   = ["s3:GetObjectVersionForReplication", "s3:GetObjectVersionAcl", "s3:GetObjectVersionTagging"]
    resources = ["${aws_s3_bucket.state.arn}/*"]
  }
  statement {
    sid       = "ReplicaObjects"
    actions   = ["s3:ReplicateObject", "s3:ReplicateDelete"]
    resources = ["${aws_s3_bucket.replica.arn}/*"]
  }
  statement {
    sid       = "DecryptSource"
    actions   = ["kms:Decrypt"]
    resources = [aws_kms_key.state.arn]
    condition {
      test     = "StringLike"
      variable = "kms:ViaService"
      values   = ["s3.${var.region}.amazonaws.com"]
    }
    condition {
      test     = "StringLike"
      variable = "kms:EncryptionContext:aws:s3:arn"
      values   = [aws_s3_bucket.state.arn]
    }
  }
  statement {
    sid       = "EncryptReplica"
    actions   = ["kms:Encrypt"]
    resources = [aws_kms_key.replica.arn]
    condition {
      test     = "StringLike"
      variable = "kms:ViaService"
      values   = ["s3.${var.standby_region}.amazonaws.com"]
    }
    condition {
      test     = "StringLike"
      variable = "kms:EncryptionContext:aws:s3:arn"
      values   = [aws_s3_bucket.replica.arn]
    }
  }
}

resource "aws_iam_role_policy" "replication" {
  name   = "state-replication"
  role   = aws_iam_role.replication.id
  policy = data.aws_iam_policy_document.replication.json
}

resource "aws_s3_bucket_replication_configuration" "state" {
  depends_on = [aws_s3_bucket_versioning.state, aws_s3_bucket_versioning.replica]
  role       = aws_iam_role.replication.arn
  bucket     = aws_s3_bucket.state.id
  rule {
    id     = "state-to-standby"
    status = "Enabled"
    filter {}
    delete_marker_replication {
      status = "Enabled"
    }
    source_selection_criteria {
      sse_kms_encrypted_objects {
        status = "Enabled"
      }
    }
    destination {
      bucket = aws_s3_bucket.replica.arn
      encryption_configuration {
        replica_kms_key_id = aws_kms_key.replica.arn
      }
    }
  }
}

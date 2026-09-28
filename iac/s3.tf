data "aws_caller_identity" "current" {}

# Bucket names are globally unique across ALL AWS accounts, same reason we
# suffixed the state bucket with the account ID.
resource "aws_s3_bucket" "uploads" {
  bucket = "${var.project_name}-uploads-${data.aws_caller_identity.current.account_id}"

  tags = {
    Name = "${var.project_name}-uploads"
  }
}

resource "aws_s3_bucket_public_access_block" "uploads" {
  bucket = aws_s3_bucket.uploads.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_server_side_encryption_configuration" "uploads" {
  bucket = aws_s3_bucket.uploads.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# Matches the app's own PDPA requirement (docs/security.md): uploaded
# files are never kept past 1 day.
resource "aws_s3_bucket_lifecycle_configuration" "uploads" {
  bucket = aws_s3_bucket.uploads.id

  rule {
    id     = "expire-after-1-day"
    status = "Enabled"

    filter {}

    expiration {
      days = 1
    }

    # Cleans up storage from uploads that started but never finished
    # (browser closed mid-upload, network drop, etc).
    abort_incomplete_multipart_upload {
      days_after_initiation = 1
    }
  }
}

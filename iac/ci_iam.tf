# Identity for the Jenkins `terraform plan` stage (and the nightly drift
# check). It can read state and describe resources but cannot change
# anything. The access key is created by hand so its secret never enters
# Terraform state:
#   aws iam create-access-key --user-name tf-readonly
resource "aws_iam_user" "tf_readonly" {
  name = "tf-readonly"
}

data "aws_iam_policy_document" "tf_readonly" {
  # Plan runs with -lock=false, so state is only ever read.
  statement {
    sid       = "ReadState"
    actions   = ["s3:GetObject"]
    resources = ["arn:aws:s3:::thai-gov-processor-tfstate-${data.aws_caller_identity.current.account_id}/thai-gov-processor/terraform.tfstate"]
  }

  statement {
    sid       = "ListStateBucket"
    actions   = ["s3:ListBucket"]
    resources = ["arn:aws:s3:::thai-gov-processor-tfstate-${data.aws_caller_identity.current.account_id}"]
  }

  # Bucket settings only, never s3:GetObject: the uploads bucket holds users'
  # documents and photos.
  statement {
    sid = "ReadUploadsBucketConfig"
    actions = [
      "s3:ListBucket",
      "s3:GetBucket*",
      "s3:GetLifecycleConfiguration",
      "s3:GetEncryptionConfiguration",
      "s3:GetAccelerateConfiguration",
      "s3:GetReplicationConfiguration",
    ]
    resources = [aws_s3_bucket.uploads.arn]
  }

  statement {
    sid       = "ReadBudget"
    actions   = ["budgets:ViewBudget", "budgets:ListTagsForResource"]
    resources = ["*"]
  }

  statement {
    sid = "ReadRemainingInfra"
    actions = [
      "ec2:Describe*",
      "ecr:Describe*",
      "ecr:GetLifecyclePolicy",
      "ecr:GetRepositoryPolicy",
      "ecr:ListTagsForResource",
      "iam:Get*",
      "iam:List*",
    ]
    resources = ["*"] # Describe/List actions do not support resource-level scoping
  }
}

resource "aws_iam_user_policy" "tf_readonly" {
  name   = "${var.project_name}-tf-readonly"
  user   = aws_iam_user.tf_readonly.name
  policy = data.aws_iam_policy_document.tf_readonly.json
}

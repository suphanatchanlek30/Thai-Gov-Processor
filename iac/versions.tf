terraform {
  required_version = ">= 1.10"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # State lives in the bucket we created by hand (see docs/infrastructure.md
  # for why it can't be created by Terraform itself). use_lockfile needs
  # Terraform >= 1.10 and replaces the older DynamoDB lock-table pattern —
  # S3 now supports native conditional writes, so no second AWS service is
  # needed just to prevent two people running `apply` at the same time.
  backend "s3" {
    bucket       = "thai-gov-processor-tfstate-334177992720"
    key          = "thai-gov-processor/terraform.tfstate"
    region       = "ap-southeast-1"
    use_lockfile = true
    encrypt      = true
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project   = "thai-gov-processor"
      ManagedBy = "terraform"
    }
  }
}

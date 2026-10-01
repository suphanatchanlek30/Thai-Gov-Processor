variable "aws_region" {
  type    = string
  default = "ap-southeast-1"
}

variable "project_name" {
  type    = string
  default = "thai-gov-processor"
}

# Your own IP, so port 6443 (K3s API) is reachable from your machine but
# nowhere else on the internet. Set in terraform.tfvars (gitignored) — never
# hardcode a real IP into a committed .tf file.
variable "admin_cidr" {
  type = string
}

# Pinned so a rebuild months later gets the same cluster version instead of
# whatever "latest" is that day. Upgrade by bumping this and re-applying.
variable "k3s_version" {
  type    = string
  default = "v1.36.4+k3s1"
}

variable "ecr_provider_version" {
  type    = string
  default = "v1.37.0"
}

# Pinned checksum so a tampered or replaced binary fails the boot script
# instead of running as root-adjacent code on the node. Update together
# with ecr_provider_version (the .sha256 file sits next to the binary).
variable "ecr_provider_sha256" {
  type    = string
  default = "842e0fd8159f5ed8df2f38e6521e1e4f0a1ee80cf1e1ac119bbf20bf9c20681c"
}

locals {
  ecr_provider_url = "https://storage.googleapis.com/k8s-artifacts-prod/binaries/cloud-provider-aws/${var.ecr_provider_version}/linux/amd64/ecr-credential-provider-linux-amd64"
}

variable "instance_type" {
  type    = string
  default = "m7i-flex.large"
}

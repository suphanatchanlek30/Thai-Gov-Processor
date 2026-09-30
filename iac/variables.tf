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

variable "instance_type" {
  type    = string
  default = "t3a.large"
}

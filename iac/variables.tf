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

variable "instance_type" {
  type    = string
  default = "t3a.large"
}

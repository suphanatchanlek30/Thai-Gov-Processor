resource "aws_security_group" "app" {
  name        = "${var.project_name}-app"
  description = "K3s node: public web/API, admin-only K3s API, no SSH (SSM instead)"
  vpc_id      = aws_vpc.main.id

  tags = {
    Name = "${var.project_name}-app-sg"
  }
}

# Split into one resource per rule (the provider's newer style, AWS
# provider v5+) instead of inline ingress{}/egress{} blocks on the SG
# itself — adding or removing a single rule no longer forces Terraform to
# recreate the whole security group.
resource "aws_vpc_security_group_ingress_rule" "http" {
  security_group_id = aws_security_group.app.id
  description       = "HTTP (also the ACME HTTP-01 challenge path)"
  cidr_ipv4         = "0.0.0.0/0"
  from_port         = 80
  to_port           = 80
  ip_protocol       = "tcp"
}

resource "aws_vpc_security_group_ingress_rule" "https" {
  security_group_id = aws_security_group.app.id
  description       = "HTTPS"
  cidr_ipv4         = "0.0.0.0/0"
  from_port         = 443
  to_port           = 443
  ip_protocol       = "tcp"
}

resource "aws_vpc_security_group_ingress_rule" "k3s_api" {
  security_group_id = aws_security_group.app.id
  description       = "K3s API (kubectl) - admin IP only"
  cidr_ipv4         = var.admin_cidr
  from_port         = 6443
  to_port           = 6443
  ip_protocol       = "tcp"
}

resource "aws_vpc_security_group_egress_rule" "all" {
  security_group_id = aws_security_group.app.id
  description       = "Outbound: ECR pulls, SSM agent, ACME/Lets Encrypt, apt"
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "-1"
}

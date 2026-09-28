# Look up the AMI at plan time instead of hardcoding an ID - AMI IDs are
# region- and time-specific (a new one is published regularly), so a
# hardcoded ID would silently go stale or not exist in another region.
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd*/ubuntu-noble-24.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# Allocated standalone (not tied to the instance yet) so its address is
# already known before the instance - and its user_data - is created.
resource "aws_eip" "app" {
  domain = "vpc"

  tags = {
    Name = "${var.project_name}-eip"
  }
}

resource "aws_instance" "app" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = var.instance_type
  subnet_id              = aws_subnet.public.id
  vpc_security_group_ids = [aws_security_group.app.id]
  iam_instance_profile   = aws_iam_instance_profile.ec2.name

  # IMDSv2 only (http_tokens = required blocks the older, SSRF-prone
  # IMDSv1). hop_limit = 2, not the default 1, because requests to the
  # metadata service from inside a container cross one extra network hop
  # (host -> container bridge) - at hop_limit 1 those requests silently
  # time out and pods can't pick up the instance's IAM role.
  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 2
  }

  root_block_device {
    volume_type = "gp3"
    volume_size = 35
    encrypted   = true
  }

  user_data = templatefile("${path.module}/templates/user_data.sh.tftpl", {
    elastic_ip = aws_eip.app.public_ip
  })

  tags = {
    Name = "${var.project_name}-app"
  }
}

resource "aws_eip_association" "app" {
  instance_id   = aws_instance.app.id
  allocation_id = aws_eip.app.id
}

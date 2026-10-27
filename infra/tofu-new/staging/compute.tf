variable "ami_id" {
  type = string
}

variable "app_subnet_id" {
  type = string
}

variable "instance_type" {
  type    = string
  default = "t3.micro"
}

resource "aws_instance" "app" {
  ami                         = var.ami_id
  instance_type               = var.instance_type
  subnet_id                   = var.app_subnet_id
  associate_public_ip_address = true
  vpc_security_group_ids      = [aws_security_group.app.id]
  key_name                    = aws_key_pair.admin.key_name
  user_data                   = file("${path.module}/cloud-init.yml")

  root_block_device {
    volume_type           = "gp3"
    volume_size           = 16
    encrypted             = true
    delete_on_termination = true
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
  }

  credit_specification {
    cpu_credits = "standard"
  }

  lifecycle {
    prevent_destroy = true
    ignore_changes  = [user_data]
  }

  tags = {
    Name        = "ezvacss-staging"
    Environment = "staging"
  }
}

output "app_public_ip" {
  value = aws_instance.app.public_ip
}
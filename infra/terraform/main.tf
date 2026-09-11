terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }
}

# Configure the AWS Provider
provider "aws" {
  region = "eu-central-1"
  access_key                  = "test"        # Dummy key for LocalStack
  secret_key                  = "test"        # Dummy key for LocalStack
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  s3_use_path_style           = true
  # Direct AWS API calls to LocalStack
  endpoints {
    ec2            = "http://localhost:4566"
    s3             = "http://localhost:4566"
    sqs            = "http://localhost:4566"
    secretsmanager = "http://localhost:4566"
  }
}

# -------------------------------------------------------------------
# 2. SECURITY GROUP (FIREWALL RULES)
# -------------------------------------------------------------------
resource "aws_security_group" "web_sg" {
  name        = "ezvacss-xyz-web-sg"
  description = "Allow HTTP, Web App, and SSH traffic"

  # Allow HTTP (Port 80)
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow Go App (Port 8080)
  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow SSH (Port 22)
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Outbound rule: Allow all traffic
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# -------------------------------------------------------------------
# 3. EC2 INSTANCE RESOURCE
# -------------------------------------------------------------------
resource "aws_instance" "web_server" {
  ami           = "ami-00000000"              # Dummy AMI for LocalStack
  instance_type = "t2.micro"

  vpc_security_group_ids = [aws_security_group.web_sg.id]

  tags = {
    Name = "ezvacss.xyz"
  }
}

# -------------------------------------------------------------------
# 4. OUTPUTS (PRINT IP AFTER CREATION)
# -------------------------------------------------------------------
output "instance_id" {
  value = aws_instance.web_server.id
}

output "instance_public_ip" {
  value = aws_instance.web_server.public_ip
}
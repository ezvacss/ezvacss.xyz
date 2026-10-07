terraform {
  backend "s3" {
    bucket       = "ezvacss.xyz-opentofu-state"
    key          = "ezvacss.xyz/staging/terraform.tfstate"
    region       = "eu-north-1"
    encrypt      = true
    use_lockfile = true
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0, < 7.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}
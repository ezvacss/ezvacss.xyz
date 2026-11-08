data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }

  filter {
    name   = "default-for-az"
    values = ["true"]
  }
}

output "vpc_id" {
  value = data.aws_vpc.default.id
}

output "default_subnet_ids" {
  value = sort(data.aws_subnets.default.ids)
}
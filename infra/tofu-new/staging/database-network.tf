resource "aws_subnet" "database" {
  for_each = {
    "eu-north-1a" = "172.31.128.0/24"
    "eu-north-1b" = "172.31.129.0/24"
  }

  vpc_id                  = data.aws_vpc.default.id
  availability_zone       = each.key
  cidr_block              = each.value
  map_public_ip_on_launch = false

  tags = {
    Name = "ezvacss-xyz-staging-db-${each.key}"
  }
}

resource "aws_route_table" "database" {
  vpc_id = data.aws_vpc.default.id

  tags = {
    Name = "ezvacss-xyz-staging-db-private"
  }
}

resource "aws_route_table_association" "database" {
  for_each = aws_subnet.database

  subnet_id      = each.value.id
  route_table_id = aws_route_table.database.id
}

resource "aws_db_subnet_group" "database" {
  name        = "ezvacss-xyz-staging-db"
  description = "Private subnets for staging PostgreSQL"
  subnet_ids  = [for subnet in aws_subnet.database : subnet.id]
}
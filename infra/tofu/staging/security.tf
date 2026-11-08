resource "aws_security_group" "app" {
  name        = "ezvacss-staging-app"
  description = "Staging application server"
  vpc_id      = data.aws_vpc.default.id
}

resource "aws_vpc_security_group_ingress_rule" "web" {
  for_each = toset(["80", "443"])

  security_group_id = aws_security_group.app.id
  description       = "Public web traffic on port ${each.value}"
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "tcp"
  from_port         = tonumber(each.value)
  to_port           = tonumber(each.value)
}

resource "aws_vpc_security_group_ingress_rule" "ssh" {
  security_group_id = aws_security_group.app.id
  description       = "SSH access"
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "tcp"
  from_port         = 22
  to_port           = 22
}

resource "aws_vpc_security_group_egress_rule" "outbound" {
  security_group_id = aws_security_group.app.id
  description       = "Outbound IPv4 traffic"
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "-1"
}
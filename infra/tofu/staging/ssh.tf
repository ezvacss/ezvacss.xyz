variable "ssh_public_key_path" {
  description = "Local path to the administrator SSH public key"
  type        = string
}

resource "aws_key_pair" "admin" {
  key_name   = "ezvacss-staging-admin"
  public_key = trimspace(file(pathexpand(var.ssh_public_key_path)))
}
resource "aws_db_parameter_group" "database" {
  name   = "ezvacss-xyz-staging-postgres"
  family = "postgres17"

  parameter {
    name         = "rds.force_ssl"
    value        = "1"
    apply_method = "pending-reboot"
  }
}

resource "aws_db_instance" "database" {
  identifier     = "ezvacss-xyz-staging-db"
  engine         = "postgres"
  engine_version = "17.11"
  instance_class = "db.t4g.micro"

  db_name                     = "ezvacss"
  username                    = "dbadmin"
  manage_master_user_password = true

  allocated_storage = 20
  storage_type      = "gp3"
  storage_encrypted = true

  db_subnet_group_name   = aws_db_subnet_group.database.name
  vpc_security_group_ids = [aws_security_group.database.id]
  parameter_group_name   = aws_db_parameter_group.database.name
  publicly_accessible    = false
  multi_az               = false
  availability_zone      = "eu-north-1b"

  backup_retention_period     = 1
  backup_window               = "03:00-04:00"
  maintenance_window          = "sun:04:00-sun:05:00"
  auto_minor_version_upgrade  = true
  allow_major_version_upgrade = false
  engine_lifecycle_support    = "open-source-rds-extended-support-disabled"

  deletion_protection       = true
  skip_final_snapshot       = false
  final_snapshot_identifier = "ezvacss-xyz-staging-db-final"

  lifecycle {
    prevent_destroy = true
  }
}

output "database_address" {
  value = aws_db_instance.database.address
}
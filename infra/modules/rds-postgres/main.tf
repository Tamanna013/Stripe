resource "aws_db_subnet_group" "this" {
  name       = "${var.name}-subnet-group"
  subnet_ids = var.subnet_ids
}

resource "aws_security_group" "rds" {
  name_prefix = "${var.name}-rds-"
  vpc_id      = var.vpc_id
  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = var.allowed_security_group_ids
  }
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_db_instance" "this" {
  identifier                  = var.name
  engine                      = "postgres"
  engine_version               = "16.4"
  instance_class               = var.instance_class
  allocated_storage             = var.allocated_storage_gb
  storage_encrypted             = true
  db_subnet_group_name          = aws_db_subnet_group.this.name
  vpc_security_group_ids        = [aws_security_group.rds.id]
  multi_az                      = var.multi_az
  username                      = var.master_username
  manage_master_user_password   = true
  backup_retention_period       = 7
  deletion_protection           = var.deletion_protection
  skip_final_snapshot           = !var.deletion_protection
}

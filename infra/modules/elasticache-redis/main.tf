resource "aws_elasticache_subnet_group" "this" {
  name       = "${var.name}-subnet-group"
  subnet_ids = var.subnet_ids
}

resource "aws_security_group" "redis" {
  name_prefix = "${var.name}-redis-"
  vpc_id      = var.vpc_id
  ingress {
    from_port       = 6379
    to_port         = 6379
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

resource "aws_elasticache_replication_group" "this" {
  replication_group_id       = var.name
  description                = "FlowGuard Redis - ${var.name}"
  engine                     = "redis"
  engine_version              = "7.1"
  node_type                  = var.node_type
  num_cache_clusters          = var.num_cache_clusters
  automatic_failover_enabled  = var.num_cache_clusters > 1
  subnet_group_name           = aws_elasticache_subnet_group.this.name
  security_group_ids          = [aws_security_group.redis.id]
  at_rest_encryption_enabled   = true
  transit_encryption_enabled   = true
}

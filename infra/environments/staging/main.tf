module "vpc" {
  source = "../../modules/vpc"

  name                 = "flowguard-staging"
  cidr_block           = "10.0.0.0/16"
  azs                  = ["us-east-1a", "us-east-1b"]
  private_subnet_cidrs = ["10.0.1.0/24", "10.0.2.0/24"]
  public_subnet_cidrs  = ["10.0.101.0/24", "10.0.102.0/24"]
}

module "eks" {
  source = "../../modules/eks"

  cluster_name        = "flowguard-staging"
  kubernetes_version  = "1.30"
  subnet_ids          = module.vpc.private_subnet_ids
  node_instance_types = ["m6i.large"]
  desired_node_count  = 3
  min_node_count      = 2
  max_node_count      = 6
  oidc_thumbprint     = "9e99a48a9960b14926bb7f3b02e22da2b0ab7280"
}

module "rds_postgres" {
  source = "../../modules/rds-postgres"

  name                       = "flowguard-staging-db"
  vpc_id                     = module.vpc.vpc_id
  subnet_ids                 = module.vpc.private_subnet_ids
  instance_class             = "db.t4g.medium"
  allocated_storage_gb       = 20
  multi_az                   = false
  master_username            = "flowguardadmin"
  deletion_protection        = false
  allowed_security_group_ids = []
}

module "elasticache_redis" {
  source = "../../modules/elasticache-redis"

  name                       = "flowguard-staging-redis"
  vpc_id                     = module.vpc.vpc_id
  subnet_ids                 = module.vpc.private_subnet_ids
  node_type                  = "cache.t4g.small"
  num_cache_clusters         = 2
  allowed_security_group_ids = []
}

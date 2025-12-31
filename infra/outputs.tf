# Outputs
# These values are needed for GitHub secrets and DNS configuration

output "alb_dns_name" {
  description = "ALB DNS name - access your app here"
  value       = aws_lb.main.dns_name
}

output "ecr_registry" {
  description = "ECR registry URL (for GitHub secrets)"
  value       = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${var.aws_region}.amazonaws.com"
}

output "ecr_api_repository" {
  description = "ECR repository for API"
  value       = aws_ecr_repository.api.repository_url
}

output "ecr_web_repository" {
  description = "ECR repository for Web"
  value       = aws_ecr_repository.web.repository_url
}

output "ecr_ml_repository" {
  description = "ECR repository for ML"
  value       = aws_ecr_repository.ml.repository_url
}

output "database_endpoint" {
  description = "RDS endpoint"
  value       = aws_db_instance.main.endpoint
}

output "database_url" {
  description = "Full database URL (for GitHub secrets)"
  value       = "postgresql://${var.db_username}:${random_password.db_password.result}@${aws_db_instance.main.endpoint}/${var.db_name}"
  sensitive   = true
}

output "s3_bucket" {
  description = "S3 bucket for audio storage"
  value       = aws_s3_bucket.audio.id
}

output "aws_region" {
  description = "AWS region"
  value       = var.aws_region
}

output "ecs_cluster_name" {
  description = "ECS cluster name"
  value       = aws_ecs_cluster.main.name
}

output "api_service_name" {
  description = "API ECS service name"
  value       = aws_ecs_service.api.name
}

output "web_service_name" {
  description = "Web ECS service name"
  value       = aws_ecs_service.web.name
}

output "ml_service_name" {
  description = "ML ECS service name"
  value       = aws_ecs_service.ml.name
}

# GitHub Secrets Summary
output "github_secrets_summary" {
  description = "Summary of values to add to GitHub secrets"
  value       = <<-EOT

    ============================================
    ADD THESE TO GITHUB REPOSITORY SECRETS:
    ============================================

    AWS_REGION          = ${var.aws_region}
    ECR_REGISTRY        = ${data.aws_caller_identity.current.account_id}.dkr.ecr.${var.aws_region}.amazonaws.com
    ECS_CLUSTER         = ${aws_ecs_cluster.main.name}
    API_SERVICE_NAME    = ${aws_ecs_service.api.name}
    WEB_SERVICE_NAME    = ${aws_ecs_service.web.name}
    ML_SERVICE_NAME     = ${aws_ecs_service.ml.name}
    S3_BUCKET           = ${aws_s3_bucket.audio.id}

    Run 'terraform output database_url' to get DATABASE_URL (sensitive)

    You also need AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
    for an IAM user with ECR push and ECS deploy permissions.

    ============================================
  EOT
}

# Data source for account ID
data "aws_caller_identity" "current" {}

output "alb_dns_name" {
  description = "DNS name of the Application Load Balancer (use for Slack app Request URL)"
  value       = aws_lb.clew.dns_name
}

output "ecr_repository_url" {
  description = "ECR repository URL for docker push"
  value       = aws_ecr_repository.clew.repository_url
}

output "oidc_role_arn" {
  description = "IAM role ARN for GitHub Actions OIDC authentication"
  value       = aws_iam_role.clew_github_actions.arn
}

output "ecs_cluster_name" {
  description = "ECS cluster name (used in CI pipeline and CLI commands)"
  value       = aws_ecs_cluster.clew.name
}

output "ecs_service_name" {
  description = "ECS service name (used in CI pipeline and CLI commands)"
  value       = aws_ecs_service.clew.name
}

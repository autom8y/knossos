# -----------------------------------------------------------------------------
# Required Variables
# -----------------------------------------------------------------------------

variable "aws_account_id" {
  description = "AWS account ID used for IAM ARN construction"
  type        = string

  validation {
    condition     = can(regex("^\\d{12}$", var.aws_account_id))
    error_message = "aws_account_id must be a 12-digit AWS account number."
  }
}

variable "vpc_id" {
  description = "VPC ID where Clew resources are deployed"
  type        = string

  validation {
    condition     = can(regex("^vpc-", var.vpc_id))
    error_message = "vpc_id must start with 'vpc-'."
  }
}

variable "public_subnet_ids" {
  description = "List of public subnet IDs for ALB and ECS tasks (minimum 2 for ALB)"
  type        = list(string)

  validation {
    condition     = length(var.public_subnet_ids) >= 2
    error_message = "At least 2 public subnets are required for the ALB."
  }
}

variable "acm_certificate_arn" {
  description = "ARN of the ACM certificate for HTTPS on the ALB"
  type        = string

  validation {
    condition     = can(regex("^arn:aws:acm:", var.acm_certificate_arn))
    error_message = "acm_certificate_arn must be a valid ACM ARN."
  }
}

# -----------------------------------------------------------------------------
# Optional Variables
# -----------------------------------------------------------------------------

variable "aws_region" {
  description = "AWS region for all resources"
  type        = string
  default     = "us-east-1"
}

variable "github_org" {
  description = "GitHub organization for OIDC trust policy"
  type        = string
  default     = "autom8y"
}

variable "github_repo" {
  description = "GitHub repository name for OIDC trust policy"
  type        = string
  default     = "knossos"
}

variable "otel_endpoint" {
  description = "OpenTelemetry OTLP HTTP endpoint. Empty string disables tracing (noop exporter)"
  type        = string
  default     = ""
}

variable "image_tag" {
  description = "Docker image tag to deploy. CI uses the git SHA; 'latest' for manual deploys"
  type        = string
  default     = "latest"
}

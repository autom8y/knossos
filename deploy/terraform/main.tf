# Clew Infrastructure — AWS ECS Fargate deployment
#
# Provisions: ECR repository, ECS cluster + service, ALB with HTTPS,
# IAM roles (execution, task, OIDC for GitHub Actions), Secrets Manager
# placeholders, CloudWatch log group, and security groups.
#
# Prerequisites:
#   - Existing VPC with public subnets (passed via variables)
#   - ACM certificate for HTTPS (passed via variable)
#   - Terraform >= 1.5

terraform {
  required_version = ">= 1.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0"
    }
  }

  # Local backend for MVP. Migrate to S3 + DynamoDB when ready for team use.
  backend "local" {
    path = "terraform.tfstate"
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project   = "clew"
      ManagedBy = "terraform"
    }
  }
}

# ---------------------------------------------------------------------------
# Data Sources
# ---------------------------------------------------------------------------

data "aws_vpc" "selected" {
  id = var.vpc_id
}

data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

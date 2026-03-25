# IAM Roles — ECS execution, ECS task, and GitHub Actions OIDC.

# ---------------------------------------------------------------------------
# 1. ECS Execution Role — used by ECS agent to pull images, fetch secrets, push logs
# ---------------------------------------------------------------------------

resource "aws_iam_role" "clew_execution" {
  name = "clew-ecs-execution"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

# Standard ECS execution policy (ECR pull + CloudWatch Logs)
resource "aws_iam_role_policy_attachment" "clew_execution_ecs" {
  role       = aws_iam_role.clew_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# Secrets Manager access for the 3 Clew secrets
resource "aws_iam_role_policy" "clew_execution_secrets" {
  name = "clew-secrets-access"
  role = aws_iam_role.clew_execution.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue"
        ]
        Resource = [
          aws_secretsmanager_secret.slack_signing_secret.arn,
          aws_secretsmanager_secret.slack_bot_token.arn,
          aws_secretsmanager_secret.anthropic_api_key.arn,
        ]
      }
    ]
  })
}

# ---------------------------------------------------------------------------
# 2. ECS Task Role — assumed by the running container (minimal for MVP)
# ---------------------------------------------------------------------------

resource "aws_iam_role" "clew_task" {
  name = "clew-ecs-task"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

# No additional policies for MVP — Clew talks to external APIs (Slack, Claude)
# over the internet, not to AWS services from within the container.

# ---------------------------------------------------------------------------
# 3. GitHub Actions OIDC Role — assumed by CI to push images and deploy
# ---------------------------------------------------------------------------

# OIDC provider for GitHub Actions (one per account, may already exist)
resource "aws_iam_openid_connect_provider" "github_actions" {
  url             = "https://token.actions.githubusercontent.com"
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = ["ffffffffffffffffffffffffffffffffffffffff"] # GitHub-managed, thumbprint validation not used for OIDC
}

resource "aws_iam_role" "clew_github_actions" {
  name = "clew-github-actions"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = aws_iam_openid_connect_provider.github_actions.arn
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
          }
          StringLike = {
            "token.actions.githubusercontent.com:sub" = "repo:${var.github_org}/${var.github_repo}:ref:refs/heads/main"
          }
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "clew_github_actions" {
  name = "clew-ci-deploy"
  role = aws_iam_role.clew_github_actions.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "ECRAuth"
        Effect = "Allow"
        Action = [
          "ecr:GetAuthorizationToken"
        ]
        Resource = "*"
      },
      {
        Sid    = "ECRPush"
        Effect = "Allow"
        Action = [
          "ecr:BatchCheckLayerAvailability",
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage",
          "ecr:PutImage",
          "ecr:InitiateLayerUpload",
          "ecr:UploadLayerPart",
          "ecr:CompleteLayerUpload",
        ]
        Resource = aws_ecr_repository.clew.arn
      },
      {
        Sid    = "ECSDeploy"
        Effect = "Allow"
        Action = [
          "ecs:DescribeServices",
          "ecs:UpdateService",
          "ecs:DescribeTaskDefinition",
          "ecs:RegisterTaskDefinition",
          "ecs:ListTaskDefinitions",
          "ecs:DescribeTasks",
          "ecs:ListTasks",
        ]
        Resource = "*"
      },
      {
        Sid    = "PassRole"
        Effect = "Allow"
        Action = "iam:PassRole"
        Resource = [
          aws_iam_role.clew_execution.arn,
          aws_iam_role.clew_task.arn,
        ]
      }
    ]
  })
}

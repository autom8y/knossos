# ECS Cluster + Service — runs Clew on Fargate.

resource "aws_ecs_cluster" "clew" {
  name = "clew-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

resource "aws_ecs_task_definition" "clew" {
  family                   = "clew"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.clew_execution.arn
  task_role_arn            = aws_iam_role.clew_task.arn

  container_definitions = jsonencode([
    {
      name      = "clew"
      image     = "${aws_ecr_repository.clew.repository_url}:${var.image_tag}"
      essential = true

      portMappings = [
        {
          containerPort = 8080
          hostPort      = 8080
          protocol      = "tcp"
        }
      ]

      healthCheck = {
        command     = ["CMD-SHELL", "wget -qO- http://localhost:8080/health || exit 1"]
        interval    = 10
        timeout     = 5
        retries     = 3
        startPeriod = 90
      }

      environment = [
        { name = "PORT", value = "8080" },
        { name = "LOG_LEVEL", value = "INFO" },
        { name = "MAX_CONCURRENT", value = "10" },
        { name = "OTEL_EXPORTER_OTLP_ENDPOINT", value = var.otel_endpoint },
        { name = "OTEL_SERVICE_NAME", value = "clew" },
      ]

      secrets = [
        {
          name      = "SLACK_SIGNING_SECRET"
          valueFrom = aws_secretsmanager_secret.slack_signing_secret.arn
        },
        {
          name      = "SLACK_BOT_TOKEN"
          valueFrom = aws_secretsmanager_secret.slack_bot_token.arn
        },
        {
          name      = "ANTHROPIC_API_KEY"
          valueFrom = aws_secretsmanager_secret.anthropic_api_key.arn
        },
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.clew.name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "clew"
        }
      }

      stopTimeout = 30
    }
  ])
}

resource "aws_ecs_service" "clew" {
  name            = "clew-service"
  cluster         = aws_ecs_cluster.clew.id
  task_definition = aws_ecs_task_definition.clew.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = var.public_subnet_ids
    security_groups  = [aws_security_group.clew_ecs.id]
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.clew.arn
    container_name   = "clew"
    container_port   = 8080
  }

  deployment_circuit_breaker {
    enable   = true
    rollback = true
  }

  # Allow ECS to stabilize during first deploy before Terraform times out.
  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 100

  # Ignore changes to task_definition and desired_count so CI can update
  # the service without Terraform reverting it.
  lifecycle {
    ignore_changes = [task_definition, desired_count]
  }

  depends_on = [aws_lb_listener.clew_https]
}

# CloudWatch Logs — centralized logging for Clew ECS tasks.

resource "aws_cloudwatch_log_group" "clew" {
  name              = "/ecs/clew"
  retention_in_days = 30
}

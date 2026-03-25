# Security Groups — network access control for ALB and ECS tasks.

# ---------------------------------------------------------------------------
# ALB Security Group — accepts HTTPS from the internet
# ---------------------------------------------------------------------------

resource "aws_security_group" "clew_alb" {
  name        = "clew-alb-sg"
  description = "Allow HTTPS inbound to Clew ALB"
  vpc_id      = var.vpc_id
}

resource "aws_vpc_security_group_ingress_rule" "alb_https" {
  security_group_id = aws_security_group.clew_alb.id
  description       = "HTTPS from internet"
  from_port         = 443
  to_port           = 443
  ip_protocol       = "tcp"
  cidr_ipv4         = "0.0.0.0/0"
}

resource "aws_vpc_security_group_ingress_rule" "alb_http" {
  security_group_id = aws_security_group.clew_alb.id
  description       = "HTTP from internet (redirects to HTTPS)"
  from_port         = 80
  to_port           = 80
  ip_protocol       = "tcp"
  cidr_ipv4         = "0.0.0.0/0"
}

resource "aws_vpc_security_group_egress_rule" "alb_to_ecs" {
  security_group_id            = aws_security_group.clew_alb.id
  description                  = "Forward traffic to ECS tasks"
  from_port                    = 8080
  to_port                      = 8080
  ip_protocol                  = "tcp"
  referenced_security_group_id = aws_security_group.clew_ecs.id
}

# ---------------------------------------------------------------------------
# ECS Security Group — accepts traffic from ALB only, egress to internet
# ---------------------------------------------------------------------------

resource "aws_security_group" "clew_ecs" {
  name        = "clew-ecs-sg"
  description = "Allow inbound from ALB to Clew ECS tasks"
  vpc_id      = var.vpc_id
}

resource "aws_vpc_security_group_ingress_rule" "ecs_from_alb" {
  security_group_id            = aws_security_group.clew_ecs.id
  description                  = "HTTP from ALB"
  from_port                    = 8080
  to_port                      = 8080
  ip_protocol                  = "tcp"
  referenced_security_group_id = aws_security_group.clew_alb.id
}

resource "aws_vpc_security_group_egress_rule" "ecs_to_internet" {
  security_group_id = aws_security_group.clew_ecs.id
  description       = "Allow all outbound (Slack API, Claude API, ECR, Secrets Manager)"
  ip_protocol       = "-1"
  cidr_ipv4         = "0.0.0.0/0"
}

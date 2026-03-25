# Application Load Balancer — HTTPS termination for Clew.

resource "aws_lb" "clew" {
  name               = "clew-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.clew_alb.id]
  subnets            = var.public_subnet_ids

  enable_deletion_protection = false # MVP — enable in production
}

resource "aws_lb_target_group" "clew" {
  name        = "clew-tg"
  port        = 8080
  protocol    = "HTTP"
  vpc_id      = var.vpc_id
  target_type = "ip" # Required for Fargate awsvpc

  health_check {
    enabled             = true
    path                = "/ready"
    port                = "traffic-port"
    protocol            = "HTTP"
    healthy_threshold   = 2
    unhealthy_threshold = 3
    timeout             = 5
    interval            = 30
    matcher             = "200"
  }

  deregistration_delay = 30 # Match container stopTimeout
}

resource "aws_lb_listener" "clew_https" {
  load_balancer_arn = aws_lb.clew.arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-2021-06"
  certificate_arn   = var.acm_certificate_arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.clew.arn
  }
}

# Redirect HTTP to HTTPS
resource "aws_lb_listener" "clew_http_redirect" {
  load_balancer_arn = aws_lb.clew.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

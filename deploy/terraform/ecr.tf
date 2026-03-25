# ECR Repository — stores Clew Docker images.

resource "aws_ecr_repository" "clew" {
  name                 = "clew"
  image_tag_mutability = "MUTABLE" # Allow :latest tag overwrites
  force_delete         = false

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_lifecycle_policy" "clew" {
  repository = aws_ecr_repository.clew.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 10 tagged images"
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = 10
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}

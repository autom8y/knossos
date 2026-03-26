---
domain: feat/ci-cd-infrastructure
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./.github/workflows/*.yml"
  - "./deploy/terraform/**/*.tf"
  - "./deploy/ecs-task-definition.json"
  - "./deploy/Dockerfile"
  - "./.goreleaser.yaml"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.90
format_version: "1.0"
---

# CI/CD Pipeline Infrastructure

## Purpose and Design Rationale

Two distinct product delivery paths operate independently: the **ari CLI distribution pipeline** (GitHub Releases + Homebrew) and the **Clew deployment pipeline** (AWS ECS Fargate). A third set of workflows provides quality gates for framework content. OIDC authentication (no static AWS credentials), path filtering for cost control, and tests gating deploys are the key design decisions.

## Conceptual Model

**Chain 1 (ari CLI):** ariadne-tests.yml (gate) -> release.yml (GoReleaser, 4 platforms) -> e2e-distribution.yml (macOS + Linux validation) -> Homebrew tap update.

**Chain 2 (Clew):** deploy-clew.yml (test -> Docker build -> ECR push -> ECS task definition render via envsubst -> ECS deploy -> health check verification).

**Three-layer infrastructure:** Compute (ECS Fargate, 256 CPU/512MB, desired_count=1, circuit breaker + rollback), Network/TLS (ALB with HTTPS termination, TLS 1.3+), Identity/Secrets (3 IAM roles, 3 Secrets Manager entries).

## Implementation Map

6 workflow files in `.github/workflows/`, 13 Terraform files in `deploy/terraform/`, ECS task definition with envsubst rendering, GoReleaser config for 4 cross-compilation targets, multi-stage distroless Dockerfile with pre-baked registry + content.

## Boundaries and Failure Modes

- No explicit rollback workflow (circuit breaker only)
- Single environment (no staging)
- Terraform state committed locally (MVP)
- Content freshness requires manual operator step + rebuild
- No secret rotation configured
- Race detector tests are `continue-on-error: true` (non-blocking)

## Knowledge Gaps

1. CI task definition template relationship to Terraform-managed task definition
2. `InstrumentedPipeline` only wraps `Query`, not `QueryStream`
3. `containerInsights` interaction with EMF metrics

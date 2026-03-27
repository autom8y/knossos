---
domain: release/platform-profile
generated_at: "2026-03-27T11:15:00Z"
expires_after: "30d"
source_scope:
  - "./.know/release/"
generator: cartographer
source_hash: "722d4ef"
confidence: 0.95
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

## Repository Ecosystem

- **Repo**: autom8y-data (autom8y/autom8y-data)
- **Ecosystem**: python_uv
- **Distribution**: AWS CodeArtifact registry
- **Publish URL**: `https://autom8y-696318035277.d.codeartifact.us-east-1.amazonaws.com/pypi/autom8y-python/`
- **Auth**: UV_PUBLISH_USERNAME (aws) + UV_PUBLISH_PASSWORD (CodeArtifact token via `aws codeartifact get-authorization-token`)
- **Note**: pyproject.toml `[[tool.uv.index]]` has no `publish-url` — must use `--publish-url` flag directly
- **Build**: `uv build` produces whl + sdist in `dist/`

## Pipeline Chain

- **Type**: dispatch_chain, depth 4, cross-repo
- **Stage 1**: test.yml (autom8y/autom8y-data) — push to main — CI gates (~8 min)
- **Stage 2**: satellite-dispatch.yml (autom8y/autom8y-data) — workflow_run: Test completed (success) (~10s)
- **Stage 3**: satellite-receiver.yml (autom8y/autom8y) — repository_dispatch: satellite-deploy — includes Docker build, ECR push, ECS canary deploy (~27 min)
- **Stage 4**: service-deploy.yml (autom8y/autom8y) — embedded in Stage 3 (no separate workflow fires)
- **Terminal stage**: Stage 3 satellite-receiver.yml contains full deployment
- **ECS**: cluster=autom8y-cluster, service=autom8y-data-service, Fargate, 2 tasks, CANARY deployment
- **Smoke test**: https://data.api.autom8y.io/health
- **Typical chain duration**: ~35 minutes total (Stage 1: 8m, Stage 2: 0.2m, Stage 3: 27m)
- **CI waiter**: 80 polls * 30s = 40 min budget (fixed by autom8y/autom8y#26, 2026-03-27)

## Health Check Architecture

| Component | Path | Owner | Purpose |
|-----------|------|-------|---------|
| ALB target group | `/health` | Terraform (alb-target-group) | Liveness — keeps targets registered during transient issues |
| ECS container | Terraform-owned | Terraform (ecs-fargate-service) | Readiness — restarts on real failures |
| Dockerfile | `/health` | Docker (fallback only) | Overridden by ECS task definition |

- **Health check grace period**: 120s (data service override, covers 70-80s DuckDB warmup)
- **CI no longer overrides container health check** (fixed by autom8y/autom8y#26)

## CI Gates (in order)

1. ruff format --check
2. ruff check
3. mypy strict targets: `src/autom8_data/clients src/autom8_data/api/auth src/autom8_data/utils`
4. mypy advisory: `src/autom8_data`
5. semgrep: `.semgrep.yml` (3 architecture rules including layer violation checks)
6. pytest (unit)
7. pytest (integration, push-only)

## Sibling Repos (no dependents)

Scanned: autom8y, autom8y-ads, autom8y-asana, autom8y-auth, autom8y-config, autom8y-log, autom8y-telemetry
None depend on autom8-data. has_dependents=false. PATCH complexity.

## Available Commands (Justfile)

test, lint, fmt, type-check, coverage (no publish target)

## Known Issues (Resolved)

- ~~**ALB listener rule misconfiguration** (discovered v0.2.1)~~: RESOLVED by terraform commit 7947c06 (Defect A+B fix). Test listener rules provisioned at +10000 priority offset. Validated with v0.2.2 deployment.
- ~~**CI waiter timeout** (discovered v0.2.2)~~: RESOLVED by autom8y/autom8y#26. Waiter extended from 15 min to 40 min. Validated with v0.2.3 (first clean CI green deployment).

## Active Issues

- Semgrep rule `autom8y.data-no-services-imports-api` enforces Layer 3 (services/) cannot import Layer 4 (api/). Request schemas shared between layers must live in core/schemas/ with re-exports in api/schemas/.
- Re-exports for mypy require explicit `X as X` syntax to be recognized as public exports.

---
domain: release/platform-profile
generated_at: "2026-03-15T02:45:00Z"
expires_after: "30d"
source_scope:
  - "./.know/release/"
generator: cartographer
source_hash: "07a7524"
confidence: 0.85
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
- **Auth**: UV_INDEX_AUTOM8Y_USERNAME (aws) + UV_INDEX_AUTOM8Y_PASSWORD (CodeArtifact token)
- **Note**: pyproject.toml `[[tool.uv.index]]` has no `publish-url` — must use `--publish-url` flag directly

## Pipeline Chain

- **Type**: dispatch_chain, depth 3, cross-repo
- **Stage 1**: test.yml (autom8y/autom8y-data) — push to main
- **Stage 2**: satellite-dispatch.yml (autom8y/autom8y-data) — workflow_run: Test completed (success)
- **Stage 3**: satellite-receiver.yml (autom8y/autom8y) — repository_dispatch: satellite-deploy
- **Terminal stage**: satellite-receiver.yml with Docker build, ECR push, ECS deploy, smoke test
- **ECS**: cluster=autom8y-cluster, service=autom8y-data-service, Fargate

## CI Gates (in order)

1. ruff format --check
2. ruff check
3. mypy strict targets: `src/autom8_data/clients src/autom8_data/api/auth src/autom8_data/utils`
4. mypy advisory: `src/autom8_data`
5. semgrep: `.semgrep.yml`
6. pytest (unit)
7. pytest (integration, push-only)

## Sibling Repos (no dependents)

Scanned: autom8y, autom8y-ads, autom8y-asana, autom8y-auth, autom8y-config, autom8y-log, autom8y-telemetry
None depend on autom8-data. has_dependents=false. PATCH complexity.

## Available Commands (Justfile)

test, lint, fmt, type-check, coverage (no publish target)

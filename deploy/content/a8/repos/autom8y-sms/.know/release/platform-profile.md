---
domain: release/platform-profile
generated_at: "2026-03-15T00:36:30Z"
expires_after: "30d"
source_scope:
  - "./.know/release/"
generator: cartographer
source_hash: "25efd0c"
confidence: 0.85
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

## Repository Ecosystem Map

| Repo | Ecosystem | Version | Distribution | Manifest |
|------|-----------|---------|--------------|----------|
| autom8y-sms | python_uv | 0.1.0 | container | pyproject.toml |

- **Package name**: `autom8y-sms`
- **Python requirement**: >=3.12
- **Build backend**: hatchling
- **has_dependents**: false (no other repos in the org depend on this package)

## Build Tooling

- **justfile**: yes
  - Key targets: `test`, `test-cov`, `lint`, `fmt`, `fmt-check`, `check`, `build`, `push`, `deploy`
- **Makefile**: no
- **Dockerfile**: yes (container distribution)

## Pipeline Chain Discovery

Single dispatch chain: `autom8y-sms:Satellite Dispatch`

| Stage | Repo | Workflow | Trigger | Classification |
|-------|------|----------|---------|----------------|
| 1 | autom8y/autom8y-sms | Test | push to main / PR | ci |
| 2 | autom8y/autom8y-sms | Satellite Dispatch | workflow_run (Test success) | dispatch |
| 3 | autom8y/autom8y | Satellite Receiver | repository_dispatch (satellite-deploy) | build |
| 4 | autom8y/autom8y | service-deploy-lambda.yml | workflow_call (from stage 3) | deploy |

- **Chain type**: dispatch_chain, depth 4, cross-repo
- **Target repos**: autom8y/autom8y
- **Terminal stage**: Terraform apply (Lambda image tag update)
- **Health check**: none (deployment success inferred from Terraform exit code)
- **Typical chain duration**: ~5 minutes

## Configuration Artifacts

- CI delegates to `autom8y/autom8y` reusable workflows via `workflow_call`
- Satellite Dispatch fires `repository_dispatch` event type `satellite-deploy`
- Lambda-only service (ECS deploy is skipped in Satellite Receiver)
- Build attestation (SBOM + Trivy scan) runs before deploy

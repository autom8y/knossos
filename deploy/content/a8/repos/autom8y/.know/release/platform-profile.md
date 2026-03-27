---
domain: release/platform-profile
generated_at: "2026-03-07T12:00:00Z"
expires_after: "30d"
source_scope:
  - "./.know/release/"
generator: cartographer
source_hash: "c9d08d0"
confidence: 0.85
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

## Repository Ecosystem Map

Single monorepo workspace at `/Users/tomtenuta/Code/autom8y/`.

### SDK Packages (sdks/python/)

| Package | Version | Ecosystem | Distribution |
|---------|---------|-----------|-------------|
| autom8y-config | 1.2.0 | python_uv | CodeArtifact |
| autom8y-log | 0.5.6 | python_uv | CodeArtifact |
| autom8y-http | 0.5.0 | python_uv | CodeArtifact |
| autom8y-core | 1.1.1 | python_uv | CodeArtifact |
| autom8y-auth | 1.1.1 | python_uv | CodeArtifact |
| autom8y-stripe | 1.3.1 | python_uv | CodeArtifact |
| autom8y-ai | 1.1.0 | python_uv | CodeArtifact |
| autom8y-telemetry | 0.5.2 | python_uv | CodeArtifact |
| autom8y-cache | 0.4.0 | python_uv | CodeArtifact |
| autom8y-slack | 0.2.0 | python_uv | CodeArtifact |
| autom8y-meta | 0.2.1 | python_uv | CodeArtifact |
| autom8y-interop | 1.0.0 | python_uv | CodeArtifact |
| autom8y-events | 0.1.0 | python_uv | CodeArtifact |

### Services (services/)

| Service | Type | Key SDK Consumers |
|---------|------|-------------------|
| pull-payments | ECS/Lambda | autom8y-stripe>=1.3.1, autom8y-config>=1.2.0, autom8y-core>=1.1.1, autom8y-http>=0.5.0, autom8y-interop>=0.3.0, autom8y-log>=0.5.6, autom8y-telemetry>=0.4.0 |
| reconcile-ads | Lambda | autom8y-config, autom8y-log, autom8y-http |
| reconcile-spend | Lambda | autom8y-config, autom8y-log |
| slack-alert | Lambda | autom8y-slack>=0.2.0, autom8y-config>=1.2.0, autom8y-log>=0.5.6, autom8y-telemetry>=0.4.0 |
| auth-mysql-sync | ECS | autom8y-config>=1.2.0, autom8y-log>=0.5.6, autom8y-telemetry>=0.4.0 |
| sms-performance-report | Lambda | autom8y-config, autom8y-log |

## Pipeline Chain Discovery

### sdk-publish-v2.yml (4-stage sequential chain)

1. **detect-changes** -- path-filtered on `sdks/python/**`, supports `force_publish` bypass
2. **Test** -- ruff lint + mypy strict (src/ and tests/) + pytest
3. **Version Check** -- pyproject.toml diff gate (version-check)
4. **Publish** -- `uv build --no-sources --out-dir dist/` + twine validate + `uv publish` to CodeArtifact

Trigger: push to main (path match) or workflow_dispatch with `sdk` + `force_publish` inputs.

### Service Deployment Pipelines

| Pipeline | Trigger | Chain | Terminal Stage |
|----------|---------|-------|----------------|
| service-deploy-dispatch.yml | push to services/** or workflow_dispatch | dispatch -> ci -> build -> deploy-lambda | service-deploy-lambda.yml |
| satellite-receiver.yml | repository_dispatch from satellite | validate -> checkout -> build -> deploy-ecs/deploy-lambda | service-deploy.yml or service-deploy-lambda.yml |

**Dispatch options (workflow_dispatch):** auth, auth-mysql-sync, pull-payments, reconcile-spend, slack-alert

### Satellite Repos

| Satellite | Deploy Method | Dispatch Chain |
|-----------|--------------|----------------|
| autom8y-asana | ECS + Lambda | test.yml -> satellite-dispatch.yml -> satellite-receiver.yml -> service-deploy.yml + service-deploy-lambda.yml |
| autom8y-data | ECS | test.yml -> satellite-dispatch.yml -> satellite-receiver.yml -> service-deploy.yml |
| autom8y-sms | Lambda | test.yml -> satellite-dispatch.yml -> satellite-receiver.yml -> service-deploy-lambda.yml |
| autom8y-ads | ECS | test.yml -> satellite-dispatch.yml -> satellite-receiver.yml -> service-deploy.yml |

**Architecture note:** Stage 4 (service-deploy) runs as workflow_call jobs INSIDE the satellite-receiver.yml run, not as separate top-level runs. Track via receiver run ID.

### Satellite SDK Dependencies (verified 2026-03-07)

| Satellite | Production SDK Dependencies |
|-----------|---------------------------|
| autom8y-ads | autom8y-auth>=1.1.0, autom8y-config>=1.2.0, autom8y-http>=0.5.0, autom8y-interop>=1.0.0, autom8y-log>=0.5.6, autom8y-meta>=0.2.0, autom8y-telemetry>=0.5.0 |
| autom8y-asana | autom8y-config>=1.2.0, autom8y-http>=0.5.0, autom8y-cache>=0.4.0, autom8y-log>=0.5.6, autom8y-core>=1.1.1, autom8y-telemetry>=0.5.0 (+ optional: autom8y-auth>=1.1.0) |
| autom8y-data | autom8y-config>=1.2.0, autom8y-log>=0.5.6, autom8y-core>=1.1.1, autom8y-http>=0.5.0, autom8y-cache>=0.4.0 (+ optional: autom8y-auth>=1.1.0, autom8y-telemetry>=0.5.0) |
| autom8y-sms | autom8y-ai>=0.1.0, autom8y-config>=1.2.0, autom8y-core>=1.1.1, autom8y-http>=0.5.0, autom8y-interop>=1.0.0, autom8y-log>=0.5.6, autom8y-telemetry>=0.5.0 |

## Configuration Artifacts

- Root workspace: `pyproject.toml` (package=false, all SDKs as workspace members)
- Build system: `uv_build>=0.7,<1`
- CodeArtifact: domain=autom8y, repository=autom8y-python
- Publish command (local): `just ca-publish {sdk-name}`

### Services Not in services.yaml

- reconcile-ads: has code + Justfile, missing from manifest
- sms-performance-report: has code, no Justfile/Dockerfile, missing from manifest

## Known Issues

- autom8y-sms pins autom8y-ai>=0.1.0 vs published 1.1.0 (wide floor across major version boundary -- recommend tightening)
- autom8y-events 0.1.0 exists but has zero satellite consumers

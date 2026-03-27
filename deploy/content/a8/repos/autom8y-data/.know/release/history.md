---
domain: release/history
generated_at: "2026-03-27T11:15:00Z"
source_scope:
  - "./.know/release/"
generator: pipeline-monitor
source_hash: "722d4ef"
confidence: 0.95
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 0
---

## Release Log

### v0.2.3 — 2026-03-27

- **Version**: 0.2.3
- **Commit**: 722d4ef
- **Tag**: v0.2.3
- **Complexity**: PATCH
- **Outcome**: PASS (first clean CI deployment success in project history)
- **Attempts**: 1 (first-attempt success)
- **CI Status**: All stages GREEN. Stage 3 completed at poll 54/80 (27m 11s of 40m budget).
- **Registry**: Docker image in ECR, deployed
- **Deployment**: task-def:214, autom8y-cluster/autom8y-data-service, 2 tasks healthy
- **Duration**: ~35 minutes total (CI 8m, dispatch 0.2m, build+deploy 27m)
- **Changes**: Version bump only (SRE validation deployment)
- **Notable**: First deployment to succeed with CI reporting green. Prior releases all had false-negative CI failures due to 15-min waiter timeout on ~27-min canary deployments. Fixed by autom8y/autom8y#26 (waiter 40 min, health check ownership to terraform).

### v0.2.2 — 2026-03-27

- **Version**: 0.2.2
- **Commit**: be20482
- **Tag**: v0.2.2
- **Complexity**: PATCH
- **Outcome**: PASS (deployed successfully, CI reported false-negative FAIL)
- **Attempts**: 1 (deployment succeeded, CI waiter timed out at 15 min)
- **CI Status**: Stages 1-2 GREEN. Stage 3 CI waiter timed out at 15 min; ECS deployment actually completed at ~26 min.
- **Registry**: Docker image in ECR (be20482), deployed
- **Deployment**: task-def:213, autom8y-cluster/autom8y-data-service, 2 tasks healthy
- **Duration**: ~35 minutes total (CI 8m, deploy 26m)
- **Changes**: Health endpoint version fix (importlib.metadata instead of hardcoded 0.1.0), analytics budget CTE fan-out fix
- **Notable**: Confirmed root cause of all prior deployment "failures" — canary lifecycle (~26 min) exceeded CI waiter (15 min). Health endpoint now correctly reports version from package metadata. ALB Defect A+B fixes validated (terraform applied successfully).

### v0.2.1 — 2026-03-26

- **Version**: 0.2.1
- **Commit**: 083b0d0 (format fix after e19b384 version bump)
- **Tag**: v0.2.1 (SHA 1a36056a)
- **Complexity**: PATCH
- **Outcome**: PASS (retrospective — deployed via v0.2.2 image push to :latest tag)
- **Attempts**: 2 (1 ruff format, 1 ALB/ECS deployment failure then eventual success)
- **CI Status**: Stages 1-2 GREEN. Stage 3 FAILED (CI waiter timeout). ECS eventually completed canary.
- **Registry**: v0.2.1 Docker image in ECR (sha256:cc669635ae63)
- **Deployment**: Superseded by v0.2.2 deployment
- **Duration**: ~35 minutes (CI ~8 min, deploy attempt ~27 min)
- **Changes**: 10 commits since v0.2.0 — gcal_calendar_id field, insights freshness metadata, ad-persistence decomposition, hygiene fixes
- **Notable**: Original diagnosis was ALB listener rule misconfiguration (Defect B). Terraform fix applied (commit 7947c06). Retrospective analysis revealed the ALB fix worked but CI waiter timeout masked the success. ECS events confirm deployment completed hours after CI reported failure.

### v0.2.0 — 2026-03-25

- **Version**: 0.2.0
- **Commit**: c3c132f (fix commit after Semgrep failure on f20b7d4)
- **Complexity**: PATCH
- **Outcome**: PASS
- **Attempts**: 2 (1 Semgrep layer violation, 1 success)
- **Registry**: autom8_data-0.2.0 on AWS CodeArtifact (whl + sdist)
- **Deployment**: task-def:206, autom8y-cluster/autom8y-data-service, 2 tasks healthy
- **Duration**: ~45 minutes (including Semgrep fix cycle)
- **Changes**: 12 commits since v0.1.0 — ads-persistence DDL/endpoints, solid_scheds redesign, LOCAL mode connection reuse, enrichment-aware dimension resolution, CI fixes, test additions
- **Notable**: Semgrep caught services/ importing api/ schemas (1 finding). Fix: moved request schemas to core/schemas/ with re-exports. Re-exports needed explicit `as X` syntax for mypy. Stage 4 (service-deploy.yml) is embedded in Stage 3 (satellite-receiver.yml) — not a separate workflow.

### v0.1.0 — 2026-03-15

- **Version**: 0.1.0 (first release)
- **Commit**: 07a7524
- **Complexity**: PATCH
- **Outcome**: PASS (verified via ECS API after CI wait timeout)
- **Attempts**: 7 (4 code/lint/CI, 2 deployment, 1 success)
- **Hotfixes**: 14 (HF-1 through HF-14)
- **Registry**: autom8_data-0.1.0 on AWS CodeArtifact (whl + sdist)
- **Deployment**: task-def:195, autom8y-cluster/autom8y-data-service, 2 tasks healthy
- **Duration**: ~2.5 hours (including all fail-forward cycles)
- **Notable**: RF-C1 refactoring introduced regressions across 4 layers (code, lint, CI config, deployment config). Each CI step masked the next — ruff format hid lint, lint hid mypy, mypy hid semgrep. ECS deployment required REDIS_URL backward-compat and slowapi state fix.

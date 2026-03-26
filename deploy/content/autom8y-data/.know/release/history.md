---
domain: release/history
generated_at: "2026-03-15T02:45:00Z"
source_scope:
  - "./.know/release/"
generator: pipeline-monitor
source_hash: "07a7524"
confidence: 0.90
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 0
---

## Release Log

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

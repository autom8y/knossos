---
domain: release/history
generated_at: "2026-03-09T14:22:00Z"
source_scope:
  - "./.know/release/"
generator: pipeline-monitor
source_hash: "db104cd"
confidence: 0.90
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 0
---

## Release Log

### v0.1.4 — 2026-03-09

- **Repos**: a8
- **Version**: v0.1.4
- **Tag SHA**: 40065dfbbe2654a7d9b4149331f98b3dad1c59a7
- **Complexity**: PATCH
- **Outcome**: PASS (with manual E2E trigger)
- **Duration**: ~12 minutes (tag push to final E2E green)
- **Commits**:
  - `fix(ci): bump golangci-lint-action to v7 for golangci-lint v2 compat`
  - `fix(devenv): add global config fallback to _a8_resolve_config`
- **CI runs**:
  - release.yml: 22857489969 (green, 175s)
  - e2e-distribution.yml: 22857905033 (green, 68s, manual dispatch)
- **Assets**: 4/4 verified (darwin_amd64, darwin_arm64, linux_amd64, checksums.txt)
- **Homebrew tap**: updated at 14:12:50Z
- **Issues**: `release.published` event failed to dispatch `e2e-distribution.yml` (2nd consecutive occurrence). Manual `workflow_dispatch` required.

### v0.1.3 — 2026-03-09

- **Repos**: a8
- **Version**: v0.1.3
- **Tag SHA**: 88575470cdb6dc504935bf38db6bd34cfd224232
- **Complexity**: PATCH
- **Outcome**: PARTIAL (e2e-distribution.yml dispatch_not_received)
- **CI runs**:
  - release.yml: 22856533784 (green, 182s)
  - e2e-distribution.yml: not triggered (dispatch_not_received)
- **Assets**: 4/4 verified
- **Homebrew tap**: updated
- **Issues**: First occurrence of `release.published` dispatch failure

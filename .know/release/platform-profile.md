---
domain: release/platform-profile
generated_at: "2026-03-07T11:18:30Z"
expires_after: "30d"
source_scope:
  - "./.know/release/"
generator: cartographer
source_hash: "2b4a43c"
confidence: 0.85
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

## Repository Ecosystem Map

| Property | Value |
|----------|-------|
| Repo | knossos |
| Module | github.com/autom8y/knossos |
| Ecosystem | go_mod |
| Distribution | binary (GoReleaser v2) |
| Project Name | ari |
| Remote | git@github.com:autom8y/knossos.git |
| Default Branch | main |
| Has Dependents | false |

### Build Targets

- darwin/amd64
- darwin/arm64
- linux/amd64
- linux/arm64

### GoReleaser Configuration

- Config: `.goreleaser.yaml` (v2)
- Brew tap: `autom8y/homebrew-tap` (via `HOMEBREW_TAP_TOKEN`)
- Release repo: `autom8y/knossos`
- Asset pattern: `ari_{version}_{os}_{arch}.tar.gz` + `checksums.txt`
- Archive extra files: README.md, LICENSE*

### Build & Test Tooling

| Tool | Targets |
|------|---------|
| Justfile | default, build, build-verbose, test, test-verbose, test-sails, audit-frontmatter, lint, clean, install, info |
| Makefile | e2e-linux, e2e-local |

## Pipeline Chain Discovery

### Chain: autom8y/knossos:release.yml

- Type: trigger_chain (single-repo)
- Depth: 2
- Cross-repo: false

| Stage | Workflow | Trigger | Classification |
|-------|----------|---------|---------------|
| 1 | release.yml | push (tags: v*) | build |
| 2 | e2e-distribution.yml | release.published | deploy |

Terminal stage: e2e-distribution.yml (has health check)

### CI Workflows

| Workflow | Triggers | Classification |
|----------|----------|---------------|
| ariadne-tests.yml | pull_request, push, workflow_dispatch | ci |
| release.yml | push_tags_v_star | build |
| e2e-distribution.yml | release_published, workflow_dispatch | deploy |
| validate-orchestrators.yml | pull_request, push, workflow_dispatch | ci |
| verify-doctrine.yml | push, pull_request, workflow_dispatch | ci |
| verify-formal-specs.yml | pull_request, push, workflow_dispatch | ci |

## Known Issues

- None. Full pipeline chain verified green as of v0.6.1 (2026-03-07).
- 3 pre-existing informational workflow failures (non-blocking): ariadne-tests (golangci-lint schema), verify-doctrine (missing ariadne/ dir), verify-formal-specs (logs unavailable).

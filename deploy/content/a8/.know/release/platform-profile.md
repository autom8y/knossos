---
domain: release/platform-profile
generated_at: "2026-03-09T14:22:00Z"
expires_after: "30d"
source_scope:
  - "./.know/release/"
generator: cartographer
source_hash: "db104cd"
confidence: 0.85
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

## Repository Ecosystem Map

| Repo | Ecosystem | Distribution | Manifest | Build Tool |
|------|-----------|-------------|----------|------------|
| a8 | go_mod | binary | go.mod | GoReleaser v2 |

- Single repo: `autom8y/a8`
- Path: `/Users/tomtenuta/Code/a8`
- Branch: `main`
- No cross-repo dependents (`has_dependents: false`)
- Complexity: PATCH (single repo, binary release)

## GoReleaser Configuration

- Config: `.goreleaser.yaml` (version 2)
- Project name: `a8`
- GOOS: darwin, linux
- GOARCH: amd64, arm64 (linux/arm64 excluded via ignore block)
- Archive name template: `{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}` (no version in filename)
- Expected assets: `a8_darwin_amd64.tar.gz`, `a8_darwin_arm64.tar.gz`, `a8_linux_amd64.tar.gz`, `checksums.txt`
- Brew tap: `autom8y/homebrew-tap` (token env: `HOMEBREW_TAP_TOKEN`)
- Release repo: `autom8y/a8`

## Pipeline Chain Discovery

**Chain: `autom8y/a8:release.yml`** (trigger_chain, depth 2, intra-repo)

| Stage | Workflow | Trigger | Classification |
|-------|----------|---------|----------------|
| 1 | `release.yml` | `push: tags: v*` | build |
| 2 | `e2e-distribution.yml` | `release.published` | deploy |

- Terminal stage: `e2e-distribution.yml` (has health check: macOS + Linux E2E)
- Cross-repo: false
- `e2e-distribution.yml` also supports `workflow_dispatch` (manual fallback)

### Known Issue: Stage 2 Dispatch Failure

The `release.published` event has failed to trigger `e2e-distribution.yml` in 2 consecutive releases (v0.1.3, v0.1.4). Root cause suspected: GITHUB_TOKEN used by GoReleaser lacks scope to emit the webhook that triggers downstream workflows. Manual `workflow_dispatch` is the current workaround.

## Available Commands

- **Justfile**: `build`, `test`, `test-verbose`, `lint`, `clean`, `install`, `bootstrap`, `status`
- **Go**: `go build ./cmd/a8/`, `go test ./...`, `go mod tidy`, `go install ./cmd/a8/`
- **GoReleaser**: `goreleaser release --clean` (CI only), `goreleaser release --snapshot --clean` (local test)
- **CI workflows**: `release.yml`, `go-ci.yml`, `manifest-validate.yml`, `e2e-distribution.yml`

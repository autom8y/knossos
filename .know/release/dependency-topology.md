---
domain: release/dependency-topology
generated_at: "2026-03-17T23:21:30Z"
expires_after: "30d"
source_scope:
  - "./.know/release/"
generator: dependency-resolver
source_hash: "caac5baa"
confidence: 0.90
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

## Dependency Topology

Single repo (knossos) with no downstream dependents. No cross-repo DAG.

### Publish Order

| Phase | Repo | Action |
|-------|------|--------|
| 1 | knossos | Tag push triggers GoReleaser (binary distribution) |

### Version Constraints

- Go module: `github.com/autom8y/knossos`
- Version source: git tags (no embedded version in go.mod)
- No consumers to bump after publish

### Blast Radius

- has_dependents: false
- No auto-escalation triggers
- PATCH complexity sufficient for single-repo release

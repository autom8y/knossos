---
domain: feat/org-management
generated_at: "2026-03-03T21:30:00Z"
expires_after: "14d"
source_scope:
  - "./internal/cmd/org/**/*.go"
  - "./internal/materialize/orgscope/**/*.go"
  - "./internal/materialize/source/resolver.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.88
format_version: "1.0"
---

# Organization-Level Resource Management

## Purpose and Design Rationale

Org-level resource management solves the multi-project sharing problem: teams using Knossos across multiple repositories share agents, mena, and rites via a named org at tier 4 in the 6-tier resolution hierarchy.

**Key decisions**: Org commands do NOT require project context. Active org via `$KNOSSOS_ORG` env var or `$XDG_CONFIG_HOME/knossos/active-org` file. Graceful no-org behavior (returns `status: "skipped"`). Separate `ORG_PROVENANCE_MANIFEST.yaml`.

## Conceptual Model

### Org Directory Structure

```
$XDG_DATA_HOME/knossos/orgs/<org-name>/
  org.yaml          ← metadata
  rites/            ← org-level rite definitions
  agents/           ← synced to ~/.claude/agents/
  mena/             ← synced to ~/.claude/skills/
```

### Active Org Resolution

1. `$KNOSSOS_ORG` env var (CI/CD override)
2. `$XDG_CONFIG_HOME/knossos/active-org` file (developer workstation)

## Implementation Map

| Command | File | Purpose |
|---------|------|---------|
| `ari org init` | `internal/cmd/org/init.go` | Create org directory + `org.yaml` |
| `ari org set` | `internal/cmd/org/set.go` | Write active-org config |
| `ari org list` | `internal/cmd/org/list.go` | Enumerate orgs |
| `ari org current` | `internal/cmd/org/current.go` | Display `ActiveOrg()` result |

Sync engine: `/Users/tomtenuta/Code/knossos/internal/materialize/orgscope/sync.go` — copies agents + mena, tracks provenance. Phase 1.5 in the materialize pipeline.

## Boundaries and Failure Modes

- Flat-files-only constraint: subdirectories in `agents/` and `mena/` silently skipped
- Org mena routed entirely to `~/.claude/skills/` (NOT split between commands/skills)
- `ActiveOrg()` reads file on every call (not cached like `KnossosHome()`)

## Knowledge Gaps

1. `--org` flag wiring in sync CLI not confirmed.
2. Org mena routing intent (all to skills/) undocumented.
3. Multi-org scenarios not supported (single-active-org model).

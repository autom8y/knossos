---
domain: feat/project-status-dashboard
generated_at: "2026-03-03T21:30:00Z"
expires_after: "14d"
source_scope:
  - "./internal/cmd/status/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.88
format_version: "1.0"
---

# Project Health Dashboard (ari status)

## Purpose and Design Rationale

`ari status` provides a single-command read-only health snapshot of all five Knossos directory trees (`.claude/`, `.knossos/`, `.know/`, `.ledge/`, `.sos/`) in one unified view.

**Key design decisions**:
- Single collector function with five independent sub-collectors
- Read-only by design (explicitly documented in help text)
- Exit code 1 on unhealthy (`.claude/` missing)
- Sync recency from `PROVENANCE_MANIFEST.yaml`, not `state.json`
- Active rite from `ACTIVE_RITE` file, not `state.json`

## Conceptual Model

### HealthDashboard Structure

Five sub-health types: `ClaudeHealth`, `KnossosHealth`, `KnowHealth`, `LedgeHealth`, `SOSHealth`, plus `Healthy bool` and `Errors []string`.

**Healthy = `.claude/` exists only.** Other directories not existing is non-fatal.

### Relationship to Other Features

- Consumes `know.ReadMeta()` for `.know/` freshness
- Consumes `provenance.Load()` for last sync timestamp
- Sibling to `ari knows` (richer per-domain detail) and `ari session status` (full session detail)

## Implementation Map

Single file: `/Users/tomtenuta/Code/knossos/internal/cmd/status/status.go` (457 lines), tests in `status_test.go` (420 lines, 76.8% coverage).

### Data Flow

```
ari status → collect(resolver) → 5 sub-collectors → HealthDashboard → printer.Print()
```

## Boundaries and Failure Modes

- Does NOT modify any files
- Does NOT validate agent frontmatter or ledge artifact schemas
- Silent error discarding pattern: `collectX` functions return partial results on read failures
- In worktrees: all five directories may be absent, only `.claude/` triggers unhealthy

## Knowledge Gaps

1. No ADR for `ari status` creation.
2. Silent error discard policy is undocumented.
3. `collect()` orchestrator has no dedicated test.

---
domain: feat/mena-lint
generated_at: "2026-03-03T21:30:00Z"
expires_after: "14d"
source_scope:
  - "./internal/cmd/lint/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.88
format_version: "1.0"
---

# Mena and Agent Lint

## Purpose and Design Rationale

`ari lint` is a pre-sync validation gate for knossos source files — mena (dromena + legomena) and agent prompts. It catches structural and semantic errors in source artifacts *before* they are projected into `.claude/` and silently break Claude Code behavior.

**Why it exists**: The materialization pipeline is one-directional and silent about semantic errors. SCAR-017 (195+ `@skill-name` references silently ignored by CC), SCAR-019 (invalid agent colors silently dropped), and SCAR-027 (session artifact persisted permanently in shared mena) each established a lint rule.

**Design position**: Lint operates on *source* artifacts under `rites/*/agents/`, `rites/*/mena/`, and `mena/` — not on materialized `.claude/` output.

## Conceptual Model

### Severity Hierarchy

| Level | Meaning |
|-------|---------|
| `CRIT` | Structural failure — file unreadable, frontmatter absent |
| `HIGH` | Semantic failure — CC will silently ignore or misroute |
| `MED` | Convention violation — deviation from platform patterns |
| `LOW` | Style gap — minor reference inconsistency |

### Rule Taxonomy

**Agent rules**: frontmatter validation, type/description checks, `agent-invalid-color` (SCAR-019), `skill-at-syntax` (SCAR-017), `maxTurns-deviation`, `agent-oversized`.

**Dromena rules**: frontmatter, `context-fork-expected/unexpected` (SCAR-018), `name-collision`, `skill-at-syntax`, source path leak detection.

**Legomena rules**: frontmatter, `triggers-missing`, `legomen-oversized`, `skill-at-syntax`, source path leaks.

**Cross-cutting**: `session-artifact-in-shared-mena` (SCAR-027).

### The Fork/Inline Allowlist

The `expectedForkState` map at `/Users/tomtenuta/Code/knossos/internal/cmd/lint/lint.go:46-88` is a compile-time registry of every known dromenon's deliberate fork classification.

## Implementation Map

Single-file implementation: `/Users/tomtenuta/Code/knossos/internal/cmd/lint/lint.go` (852 lines), tests in `lint_test.go` (275 lines).

### Key Dependencies

- `internal/frontmatter` — YAML frontmatter extraction
- `internal/mena` — `MenaSource`, `Walk()` for filesystem enumeration
- `internal/output` — text/JSON/YAML output

## Boundaries and Failure Modes

### What lint does NOT cover

- Cross-rite agents (`agents/` at project root) — only scans `rites/*/agents/`
- User-scope agents (`~/.claude/agents/`)
- Materialized `.claude/` output
- Embedded FS sources
- Manifest validity or hook configurations

### Known Limitations

- `expectedForkState` staleness: hardcoded at compile time, new dromena get LOW finding
- No exit code differentiation by severity (lint is informational, not a hard gate)
- Agent discovery is rite-scoped only

## Knowledge Gaps

1. Cross-rite agent lint coverage gap is undocumented as intentional or oversight.
2. Exit code behavior under findings has no test coverage.
3. Agent-specific rules (`maxTurns-deviation`, `agent-oversized`) have no dedicated tests.

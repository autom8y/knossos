---
domain: feat/procession-system
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/procession/**/*.go"
  - "./internal/cmd/procession/**/*.go"
  - "./internal/materialize/procession/**/*.go"
  - "./processions/**/*.yaml"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.88
format_version: "1.0"
---

# Cross-Rite Procession Workflow System

## Purpose and Design Rationale

Coordinates predetermined sequences of rite-scoped stations for workflows spanning 3-4 rite boundaries (e.g., security remediation). CC constraint is foundational: every rite switch rematerializes `.claude/agents/` but CC doesn't hot-reload, so every cross-rite station boundary is a CC session boundary. The handoff artifact is the SOLE structured context surviving close/reopen. ADR-0030 is the primary design decision. Property of a session, not a replacement for it.

## Conceptual Model

**Template:** YAML in `processions/` defining ordered stations, rite assignments, artifact directory, optional loop-back. **Station:** name, rite, alt_rite, goal, produces, loop_to. **Instance:** `session.Procession` struct in SESSION_CONTEXT.md (schema v2.3). **Handoff artifact:** Markdown with 8 required frontmatter fields, validated by ValidateProcessionHandoffFields. **Transitions:** create -> proceed -> (complete or recede). `completed_stations` is append-only.

## Implementation Map

`internal/procession/template.go` (Template, Station, Validate, LoadTemplate). `internal/materialize/procession/` (resolver, renderer, archetype_data). `internal/cmd/procession/` (create, proceed, recede, abandon, status, list -- 6 subcommands). `internal/validation/procession.go`. Embedded via `EmbeddedProcessions`. Archetype templates: `procession-workflow.md.tpl`, `procession-ref.md.tpl`.

## Boundaries and Failure Modes

Template re-resolved on every proceed/recede (no snapshot at create time). Same-day ID collision not guarded. Invalid template at render silently falls back to embedded. Rite not discoverable: warning but not blocked. Completed procession nil-ed out, no re-entry. Orchestrator template section from ADR-0030 not found in current codebase.

## Knowledge Gaps

1. handoff-procession.schema.json not read
2. resolution.ProcessionChain shadowing semantics inferred
3. execution-protocol.md and transition-protocol.md not read

---
domain: feat/session-forking
generated_at: "2026-03-03T21:30:00Z"
expires_after: "14d"
source_scope:
  - "./internal/cmd/session/fray*.go"
  - "./internal/session/context.go"
  - "./docs/decisions/ADR-0006*.md"
  - "./.know/architecture.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.85
format_version: "1.0"
---

# Session Forking (Fray)

## Purpose and Design Rationale

Enables a single session to spawn parallel workstreams without violating the single-active-session-per-terminal constraint. Creates a child session inheriting parent context, auto-parks the parent, provides each child with its own directory and lifecycle.

**ADR-0006**: Hybrid parallelization pattern (serial Layer 1, parallel Layer 2). 60% wall-time reduction. **ADR-0001**: FSM that fray respects (ACTIVE→PARKED). **ADR-0010**: `--seed` complement. **ADR-0029**: Worktree environment contract.

## Conceptual Model

### Parent-Child Relationship

- Parent: ACTIVE → PARKED, `parked_reason: "Frayed to {childID}"`, `strands: [childID]`
- Child: `frayed_from: parentID`, `fray_point: parentPhase`, `schema_version: "2.2"`, inherits initiative/complexity/rite/phase

### Key Events

- `session.frayed` on parent at fork time
- `session.strand_resolved` on parent when child wraps

## Implementation Map

Primary: `/Users/tomtenuta/Code/knossos/internal/cmd/session/fray.go` (entry point, `fraySession()` function). Supporting: `internal/session/context.go` (`FrayedFrom`/`Strands` fields), `internal/cmd/session/wrap.go:246-256` (strand_resolved back-emission).

Tests: `fray_test.go` — 4 tests. No integration test for strand_resolved roundtrip.

## Boundaries and Failure Modes

- Does NOT detect conflicting changes between strands
- Does NOT auto-resume parent when strands resolve
- Worktree creation is best-effort (non-fatal on failure)
- Events emission is non-fatal by design
- Fraying a frayed child creates a chain (no cycle detection)
- No integration test for fray→wrap→strand_resolved path

## Knowledge Gaps

1. Worktree lifecycle unmanaged (created at `/tmp/`, never tracked or cleaned).
2. v2 vs v3 event path divergence in fray (v3 constructor exists but unused).
3. No dedicated ADR for fray; schema_version 2.2 bump lacks rationale.

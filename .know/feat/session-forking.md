---
domain: feat/session-forking
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/cmd/session/fray.go"
  - "./internal/session/context.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.82
format_version: "1.0"
---

# Session Forking (Fray)

## Purpose and Design Rationale

Enables parallel workstreams within a single initiative by creating a child session inheriting parent context while allowing independent work. Auto-parks parent (prevents concurrent mutation), fail-open for worktree creation, child inherits state but not identity, strand tracking on parent (ACTIVE -> LANDED), event emission non-fatal. ADR-0006 referenced but not found on disk.

## Conceptual Model

**Directed parent-child graph:** Parent (PARKED, strands list) -> Child (ACTIVE, FrayedFrom back-pointer, FrayPoint phase). **Strand lifecycle:** ACTIVE -> LANDED (on child wrap). **Two event types:** session.frayed (on parent at fork), session.strand_resolved (on parent at child wrap). **Worktree isolation:** /tmp/knossos-fray-* via createWorktree (30s timeout). **Schema:** strandList polymorphic YAML deserializer handles v2.1 []string and v2.3 []Strand.

## Implementation Map

Entry: `internal/cmd/session/fray.go` -- fraySession() 11-step pipeline: resolve parent -> exclusive lock -> load context -> FSM validate -> create child -> write child -> update parent (park + append strand) -> write parent -> optional worktree -> emit events -> return FrayOutput. Child wrap integration in wrap.go: emit strand_resolved, update parent strand status to LANDED.

## Boundaries and Failure Modes

FSM constraint: fray only from ACTIVE -> PARKED. Lock protocol: parent locked exclusively (5-min stale threshold). Partial failure: child dir removed on parent save failure. No ADR-0006 on disk. Strand abandon path unimplemented. Multi-level fray trees untested. Worktree cleanup on wrap not handled. FrameRef on Strand never set.

## Knowledge Gaps

1. ADR-0006 missing from disk
2. Strand abandon resolution not implemented
3. Nested fray (grandchild) behavior untested

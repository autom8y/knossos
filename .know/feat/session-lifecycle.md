---
domain: feat/session-lifecycle
generated_at: "2026-03-03T21:30:00Z"
expires_after: "14d"
source_scope:
  - "./internal/session/**/*.go"
  - "./internal/cmd/session/**/*.go"
  - "./internal/lock/**/*.go"
  - "./docs/decisions/ADR-0001*.md"
  - "./docs/decisions/ADR-0022*.md"
  - "./docs/decisions/ADR-0027*.md"
  - "./.know/architecture.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.91
format_version: "1.0"
---

# Session Lifecycle Management

## Purpose and Design Rationale

Solves context loss and uncoordinated state mutations in multi-agent CC workflows. Three foundational ADRs: **ADR-0001** (FSM with 3 states, advisory flock locking, TLA+ spec), **ADR-0022** (scan-based discovery, lock format v2, SQLite rejected), **ADR-0027** (unified event system, CC session map, relaxed single-ACTIVE to per-CC-instance).

### Rejected Alternatives

Multiple ACTIVE per repo (complexity), configurable single/multi (two code paths), SQLite (CGO), event sourcing (overhead), process-scoped sessions (PID accumulation).

## Conceptual Model

### State Machine

```
NONE ──[create]──► ACTIVE ──[park]──► PARKED
                     │                   │
                     │[wrap]             │[resume]
                     ▼                   ▼
                  ARCHIVED ◄──[wrap]── ACTIVE
```

Phase transitions FORWARD-ONLY: requirements → design → implementation → validation → complete.

### Layer Model

User/Agent → Moirai (convention) → `ari session` CLI (enforcement) → `internal/session` (FSM/serialization) → `internal/lock` (flock) → filesystem

### Core Types

- `Context` — SESSION_CONTEXT.md parsed form (status, initiative, complexity, phase, fray fields)
- `FSM` — transition validation matrix
- `Status` — ACTIVE/PARKED/ARCHIVED + `NormalizeStatus()` for phantom values (SCAR-014)
- CC session map — `.sos/sessions/.cc-map/{cc-session-id}` → Knossos session ID

### Resolution Priority Chain

1. `--session-id` flag → 2. CC map lookup → 3. Smart scan (`FindActiveSessions`)

## Implementation Map

`internal/session/` (22 files, 12 test files): status.go, fsm.go, context.go, id.go, discovery.go, resolve.go, rotation.go, events_read.go, timeline.go, snapshot.go.

`internal/cmd/session/` (35 files, 17 test files): create, park, resume, wrap, fray, transition, recover, gc, list, audit, snapshot, migrate, log, field, lock, integration tests.

`internal/lock/` (4 files): advisory flock with stale detection.

### Key Flows

- **Create**: sentinel lock → FindActiveSession guard → FSM validate → NewContext → Save → emit event
- **Wrap**: pre-lock archived check → exclusive lock → FSM validate → generate sails (BLACK blocks) → archive → graduate artifacts → cleanup locks/CC map → move to archive
- **Hook resolution**: stdin JSON → `ResolveSession()` priority chain → session operations

### SCAR Evidence

SCAR-001 (stale lock reclamation), SCAR-011 (.current-session deprecated), SCAR-012 (archived session denial), SCAR-013 (wrap edge cases), SCAR-014 (phantom status normalization), SCAR-020 (session ID threading).

### Test Coverage

~300 test functions in `internal/session/`, ~450 in `internal/cmd/session/`. Notable: `moirai_integration_test.go` (28 tests, golden path), FSM exhaustive transition tests.

## Boundaries and Failure Modes

- Lock does NOT work on NFS/SMB (advisory flock is local-only)
- `BufferedEventWriter.Write()` is fire-and-forget (use `Flush()+FlushError()` for confirmation)
- Rotation operates on body only (frontmatter always preserved)
- Snapshot generation never fails (graceful degradation to empty)
- Phase transitions are forward-only (no reversal mechanism)
- `--seed` mode bypasses single-session constraint via ephemeral worktree

### Key Error Paths

- Lock timeout → `ErrLockTimeout` (directs to `ari session recover`)
- FSM violation → `ErrLifecycleViolation`
- BLACK sails at wrap → `CodeQualityGateFailed` (override with `--force`)
- Multiple active sessions → error from `FindActiveSession()`

## Knowledge Gaps

1. `migrate.go` (267 lines) has no dedicated test file (HIGH priority gap).
2. `park.go` and `resume.go` have no dedicated test files.
3. CC map cleanup in `recover.go` implementation not confirmed.
4. `docs/specs/session-fsm.tla` TLA+ spec not read.
5. ADR-0027 Phase 3/4 `.current-session` removal completeness unconfirmed.

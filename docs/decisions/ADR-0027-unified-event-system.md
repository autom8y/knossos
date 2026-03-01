# ADR-0027: Unified Event System and CC Session Map

| Field | Value |
|-------|-------|
| **Status** | Accepted |
| **Date** | 2026-02-09 |
| **Deciders** | Architecture Team |
| **Supersedes** | Partially supersedes ADR-0001 (relaxes single-ACTIVE invariant) |
| **Superseded by** | N/A |

## Part 1: Unified Event System

### Context

Two independent event systems write to the same `events.jsonl` file using incompatible schemas. This creates a dual-schema problem where consumers cannot reliably parse the event log without per-line type detection.

**System A: `session.Event`** (`internal/session/events.go`)

9 lifecycle event types using SCREAMING_CASE naming and the following JSON fields:

| Field | Type | Description |
|-------|------|-------------|
| `timestamp` | string (RFC3339) | Event timestamp |
| `event` | string | Event type (e.g., `SESSION_CREATED`) |
| `from` | string | Source state |
| `to` | string | Target state |
| `from_phase` | string | Source phase (phase transitions) |
| `to_phase` | string | Target phase (phase transitions) |
| `metadata` | object | Additional metadata |

System A event types: `SESSION_CREATED`, `SESSION_PARKED`, `SESSION_RESUMED`, `SESSION_ARCHIVED`, `PHASE_TRANSITIONED`, `LOCK_ACQUIRED`, `LOCK_RELEASED`, `SCHEMA_MIGRATED`, `SESSION_FRAYED`.

System A is backed by `EventEmitter`, which writes synchronously and also emits to a separate global audit log (`EmitToAudit`).

**System B: `clewcontract.Event`** (`internal/hook/clewcontract/event.go`)

15 event types using snake_case naming and the following JSON fields:

| Field | Type | Description |
|-------|------|-------------|
| `ts` | string (RFC3339 with millis) | Event timestamp |
| `type` | string | Event type (e.g., `tool_call`) |
| `tool` | string | Tool name (optional) |
| `path` | string | File path (optional) |
| `summary` | string | One-line summary |
| `meta` | object | Additional metadata |

System B event types: `tool_call`, `file_change`, `decision`, `context_switch`, `sails_generated`, `task_start`, `task_end`, `session_start`, `session_end`, `artifact_created`, `error`, `handoff_prepared`, `handoff_executed`, `session_frayed`, `strand_resolved`.

System B is backed by `BufferedEventWriter`, which buffers events in memory and flushes to disk every 5 seconds via a background goroutine.

**The Problem**

1. `ReadEvents()` in `session/events.go` deserializes into `session.Event` structs. System B events (with `ts` instead of `timestamp`, `type` instead of `event`) deserialize into zero-valued fields and are silently dropped as "malformed lines."
2. The schemas overlap semantically but diverge structurally. Both systems record session lifecycle events (e.g., `SESSION_CREATED` vs. `session_start`), creating duplicate but incompatible records.
3. Two separate writer implementations maintain their own concurrency control. `EventEmitter` opens and closes the file per write with no locking. `BufferedEventWriter` uses mutex-guarded batching with re-queue on failure.
4. `EmitToAudit()` writes to a second file (the global audit log) in a different format (pipe-delimited plaintext). This is redundant with structured JSONL events.

### Decision

Converge on `clewcontract.Event` as the sole event schema. Deprecate and remove `session.EventEmitter`.

#### Namespace Taxonomy

All event types adopt a dot-separated `category.action` naming convention. This replaces both the SCREAMING_CASE lifecycle types and the flat snake_case clew types.

21 total event types organized into 5 categories:

| Category | Event Type | Description | Origin |
|----------|-----------|-------------|--------|
| **session** | `session.created` | New session created (`ari session create`) | System A `SESSION_CREATED` |
| | `session.started` | Session entered CC context (`SessionStart` hook) | System B `session_start` |
| | `session.parked` | Session parked (`ari session park`) | System A `SESSION_PARKED` |
| | `session.resumed` | Session resumed (`ari session resume`) | System A `SESSION_RESUMED` |
| | `session.archived` | Session archived/wrapped (`ari session wrap`) | System A `SESSION_ARCHIVED` |
| | `session.ended` | Session left CC context (park/wrap/autopark) | System B `session_end` |
| | `session.frayed` | Session forked (`ari session fray`) | Both (merge) |
| | `session.strand_resolved` | Frayed child resolved (wrap on child) | System B `strand_resolved` |
| | `session.schema_migrated` | Schema migration occurred | System A `SCHEMA_MIGRATED` |
| **phase** | `phase.transitioned` | Phase change within session | System A `PHASE_TRANSITIONED` |
| **tool** | `tool.call` | Tool invocation (`PostToolUse` hook) | System B `tool_call` |
| | `tool.file_change` | File modification (`PostToolUse` supplemental) | System B `file_change` |
| | `tool.artifact_created` | Semantic artifact created | System B `artifact_created` |
| | `tool.error` | Error occurred | System B `error` |
| **agent** | `agent.task_start` | Agent task begun | System B `task_start` |
| | `agent.task_end` | Agent task completed | System B `task_end` |
| | `agent.decision` | Workflow decision made | System B `decision` |
| | `agent.handoff_prepared` | Handoff preparation | System B `handoff_prepared` |
| | `agent.handoff_executed` | Handoff execution | System B `handoff_executed` |
| **quality** | `quality.sails_generated` | White Sails quality signal | System B `sails_generated` |
| **lock** | `lock.acquired` | Moirai lock acquired | System A `LOCK_ACQUIRED` |
| | `lock.released` | Moirai lock released | System A `LOCK_RELEASED` |

#### Unified Event Struct

The `clewcontract.Event` struct becomes the single event representation. No structural changes to the struct are needed -- the existing fields (`ts`, `type`, `tool`, `path`, `summary`, `meta`) accommodate all 21 event types. The `meta` map provides extensibility for type-specific data (state transitions use `meta.from`/`meta.to`, phase transitions use `meta.from_phase`/`meta.to_phase`).

#### Migration Strategy

The migration proceeds in five ordered steps:

1. **Add lifecycle constructors to `clewcontract`**. Create `NewSessionCreatedEvent()`, `NewSessionParkedEvent()`, `NewSessionResumedEvent()`, `NewSessionArchivedEvent()`, `NewPhaseTransitionedEvent()`, `NewLockAcquiredEvent()`, `NewLockReleasedEvent()`, `NewSchemaMigratedEvent()`. These replace the `EventEmitter.EmitCreated()`, `EventEmitter.EmitParked()`, etc. methods with functions that return `clewcontract.Event` values using the namespaced type names.

2. **Rename existing clew types to namespaced form**. Update the `EventType` constants in `clewcontract/event.go`: `tool_call` becomes `tool.call`, `file_change` becomes `tool.file_change`, `decision` becomes `agent.decision`, and so on per the taxonomy table above.

3. **Migrate callers from `EventEmitter` to `BufferedEventWriter`**. Each CLI command that currently calls `EventEmitter.EmitX()` is rewritten to construct the corresponding `clewcontract.Event` and pass it to `BufferedEventWriter.Write()`. The callers are:
   - `internal/cmd/session/create.go` -- `EmitCreated()`
   - `internal/cmd/session/park.go` -- `EmitParked()`
   - `internal/cmd/session/resume.go` -- `EmitResumed()`
   - `internal/cmd/session/wrap.go` -- `EmitArchived()`
   - `internal/cmd/session/transition.go` -- `EmitPhaseTransition()`
   - `internal/cmd/session/fray.go` -- `EmitFrayed()`
   - `internal/lock/lock.go` -- `EmitLockAcquired()`, `EmitLockReleased()`
   - `internal/cmd/session/recover.go` -- `EmitSchemaMigrated()` (if applicable)

4. **Delete `EventEmitter`**. Remove `session.EventEmitter`, `session.EventType`, `session.Event`, and the `ReadEvents()` / `FilterEvents()` functions from `internal/session/events.go`. Add a `ReadEvents()` function to `clewcontract` that deserializes `clewcontract.Event` structs.

5. **Old `events.jsonl` files are read-only history**. No backward compatibility shims are provided for pre-migration event files. System A events in old files will not parse under the new `ReadEvents()` -- this is acceptable because historical events are archival data, not operational inputs.

#### Writer Consolidation

`BufferedEventWriter` becomes the sole writer. Key properties:

| Property | `EventEmitter` (removed) | `BufferedEventWriter` (retained) |
|----------|-------------------------|----------------------------------|
| Write signature | `Emit(Event) error` | `Write(Event)` (void) |
| Concurrency | File open/close per write, no mutex | Mutex-guarded buffer, periodic flush |
| Error visibility | Synchronous (caller sees error) | Asynchronous (`FlushError()` for diagnostic) |
| Audit log | Separate `EmitToAudit()` to plaintext | Dropped (structured JSONL is the audit log) |
| Batch I/O | No | Yes (buffer swap + `WriteMultiple`) |

The tradeoff: callers lose synchronous error visibility on write. `EventEmitter.Emit()` returned an error that callers could handle. `BufferedEventWriter.Write()` returns void. In exchange, callers gain batched I/O performance and thread-safe buffering. For callers that need immediate error feedback (e.g., short-lived hook processes), `Flush()` followed by `FlushError()` provides synchronous error checking -- this pattern is already used in `RecordToolEvent()` and `RecordStamp()`.

The global audit log (`EmitToAudit()`) is dropped. Its pipe-delimited plaintext format is inferior to structured JSONL. Any consumer that reads the audit log can instead read `events.jsonl` directly with a type filter.

#### Rotation Constants

`SESSION_CONTEXT.md` rotation uses named constants per trigger context. These are documented here because the event system records rotation as observable state change:

| Constant | MaxLines | KeepLines | Trigger |
|----------|----------|-----------|---------|
| `RotationPreCompact` | 200 | 80 | `PreCompact` hook fires (aggressive, frequent) |
| `RotationPark` | 1000 | 100 | `ari session park` (generous, preserves context for resume) |
| `RotationTransition` | 200 | 80 | Phase transition (standard, phase boundary cleanup) |

Note: currently all three contexts use `DefaultMaxLines=200` / `DefaultKeepLines=80`. The park context should use the more generous constants to preserve context for session resume. This is a follow-up implementation item.

`events.jsonl` does NOT have rotation. Each session gets its own `events.jsonl` at `.sos/sessions/<session-id>/events.jsonl`. Session-scoped files are naturally bounded by session lifetime.

### Consequences

**Positive**

1. Single event schema across the entire system. Consumers parse one struct, one set of field names, one timestamp format.
2. `ReadEvents()` in `clewcontract` returns all events without silent drops.
3. Namespace taxonomy groups related events (`session.*`, `tool.*`, `agent.*`) for filtering and aggregation.
4. `BufferedEventWriter` provides better throughput than `EventEmitter` for high-frequency tool events during active sessions.
5. Audit log consolidation reduces file count per session from 2 (events.jsonl + audit) to 1.

**Negative**

1. Breaking change for any external consumer that parses `session.Event` JSON format. Mitigation: no known external consumers; `events.jsonl` is an internal artifact.
2. Callers that relied on synchronous `Emit()` error returns must adapt to the async `Write()` + `FlushError()` pattern.
3. Historical `events.jsonl` files with System A events become partially unreadable under the new `ReadEvents()`. Mitigation: these are archival data and no operational code replays them.

**Neutral**

1. The `clewcontract` package grows from 15 to 21 event constructors. Package size increase is proportional to capability increase.
2. `context_switch` remains deferred (not in the 21 types) per Sprint 6 decision -- it requires cross-event state tracking in a stateless hook process.

---

## Part 2: CC Session Map

### Context

Knossos uses a singleton `.current-session` file (at `.sos/sessions/.current-session`) to track the active session. ADR-0022 demoted this file from authoritative source to TTL cache, with scan-based discovery (`FindActiveSession()`) as the source of truth. However, the fundamental model remains one-active-session-per-repo.

This creates a conflict when multiple Claude Code instances work on the same project simultaneously. CC provides a stable `session_id` per conversation, delivered via stdin JSON in every hook invocation (see `StdinPayload.SessionID` in `internal/hook/env.go`). This session ID is the CC conversation identifier -- it persists across resume, compact, and clear events within the same CC conversation.

The current resolution chain for "which Knossos session am I in?" is:

1. Scan `.sos/sessions/` for `status: ACTIVE` directories
2. Expect exactly 0 or 1 results
3. If 2+ results, error ("multiple active sessions found")

This fails when two CC conversations each create their own Knossos session. Both are legitimately ACTIVE, but the scan finds 2 and errors.

### Decision

Implement a CC-to-Knossos session mapping system. Deprecate `.current-session`.

#### Architecture

**Storage**: Directory-based mapping at `.sos/sessions/.cc-map/{cc-session-id}`. Each file contains the Knossos session ID as its sole content (plaintext, no JSON wrapper). Directory-based storage avoids the single-file contention that `.current-session` suffered from.

**Resolution function**: `ResolveSession(ccSessionID string, explicitID string) (string, error)` with the following precedence:

| Priority | Input | Behavior |
|----------|-------|----------|
| 1 | `--session-id` flag (explicitID) | Return it directly. No lookup needed. |
| 2 | CC session ID from hook context (ccSessionID) | Look up `.cc-map/{ccSessionID}` file. Return the Knossos session ID stored inside. |
| 3 | Neither provided | Scan all session directories for ACTIVE status. 0 results = no session. 1 result = use it. 2+ results = error with list of active sessions. |

Priority 3 is the existing `FindActiveSession()` behavior from ADR-0022, preserved as the fallback for CLI usage without CC context.

**Mapping semantics**: The `.cc-map/` entry is a live pointer, not a log. When a CC conversation wraps session K1 and creates session K2 (wrap-and-restart), the mapping is overwritten: the file `.cc-map/X` changes content from `K1` to `K2`. There is no history of previous mappings -- the event log in `events.jsonl` provides that history.

**Single-ACTIVE relaxation**: ADR-0001 established a single-ACTIVE invariant ("only one session can be ACTIVE per `.claude/` directory"). ADR-0022 Decision 1 reaffirmed this. This ADR relaxes the invariant to "single-ACTIVE-per-CC-instance." Multiple ACTIVE sessions are valid when they belong to different CC conversations. The mapping system makes this safe: each CC instance resolves to its own Knossos session without ambiguity.

#### Edge Cases

| # | Scenario | CC Session | Knossos Session | Mapping Behavior |
|---|----------|-----------|-----------------|------------------|
| 1 | Normal single-session | X | K1 | `SessionStart` hook creates `.cc-map/X` containing `K1` |
| 2 | CC resumes (source=resume/compact/clear) | X (same) | K1 (same) | Lookup `.cc-map/X` returns `K1`. No mapping change. |
| 3 | Wrap-and-restart | X | K1 archived, K2 created | `.cc-map/X` overwritten: `K1` becomes `K2` |
| 4 | Park-and-start-new | X | K1 parked, K2 created | `.cc-map/X` overwritten: `K1` becomes `K2` |
| 5 | Parallel CC instances | X, Y | K1, K2 | `.cc-map/X` contains `K1`, `.cc-map/Y` contains `K2`. Both ACTIVE simultaneously. |
| 6 | Fray | X | K1 (parent), K2 (child) | `.cc-map/X` overwritten: `K1` becomes `K2` (child becomes active in this conversation) |
| 7 | CLI without CC context | none | scan-based | Smart scan fallback: 0=no session, 1=unambiguous, 2+=error with session list |
| 8 | Stale mapping | X (CC gone) | K1 (may be PARKED/ARCHIVED) | Lookup returns K1. Caller checks session status. If PARKED/ARCHIVED, handle gracefully (suggest resume or start new). Stale `.cc-map/` files accumulate harmlessly. |

**Edge case 8 detail**: Stale mappings occur when a CC conversation ends without triggering a `SessionEnd` hook (e.g., user closes terminal). The mapping file persists but the Knossos session may have been auto-parked. This is not a bug -- the mapping is a cache, not a lock. If CC session X starts a new conversation with the same ID (unlikely but possible), the lookup returns the old Knossos session. The session's status (PARKED/ARCHIVED) tells the caller it needs to resume or create new. Cleanup of `.cc-map/` files is handled by `ari session recover`, which removes entries pointing to ARCHIVED sessions.

#### Mapping Lifecycle

**Creation**: The `SessionStart` hook handler creates or updates the mapping. When CC fires `SessionStart`, the hook receives `session_id` via stdin JSON. The handler:
1. Extracts CC `session_id` from `StdinPayload`
2. Resolves or creates a Knossos session
3. Writes `.cc-map/{cc-session-id}` with the Knossos session ID

**Update**: CLI commands that change the active session (`create`, `resume`, `fray`) update the mapping if a CC session ID is available in the environment.

**Deletion**: `ari session recover` cleans up stale mappings. The `wrap` command removes the mapping entry when archiving a session (the CC conversation may continue, but the session is done -- a new session will create a new mapping).

#### `.current-session` Deprecation

`.current-session` is fully removed. The migration path:

1. **Phase 1**: Add `.cc-map/` directory support and `ResolveSession()` function. `.current-session` continues to work as fallback (Priority 3 scan).
2. **Phase 2**: Migrate all callers of `SetCurrentSessionID()` and `ClearCurrentSessionID()` (in `internal/cmd/common/context.go`) to use `ResolveSession()`.
3. **Phase 3**: Remove `SetCurrentSessionID()`, `ClearCurrentSessionID()`, `GetCurrentSessionID()`. Remove `.current-session` file creation from all code paths. The `CacheTTL` constant and cache validation logic in `context.go` are removed.
4. **Phase 4**: `ari session recover` deletes any remaining `.current-session` files as part of its cleanup.

### Consequences

**Positive**

1. `.current-session` TOCTOU race (ADR-0022 D2 Risk #1) is eliminated by removing `.current-session` entirely. The CC session map uses per-file storage (one file per CC instance), so concurrent CC instances never contend on the same file.
2. Multiple parallel sessions become first-class. Two developers (or two CC conversations on the same machine) can each have an ACTIVE session without worktree isolation.
3. Hook resolution becomes deterministic. Given a CC session ID (always available in hook context), the lookup is O(1) -- read one file. No scan needed.
4. Worktrees are reserved for git-level isolation only. Session-level isolation no longer requires worktrees, simplifying the developer experience.

**Negative**

1. `.cc-map/` directory accumulates files over time. Each CC conversation creates one file. Cleanup relies on `ari session recover`. For a typical developer, this is 1-5 files; not a practical concern.
2. CLI commands without CC context (e.g., `ari session list` from a plain terminal) fall back to scan-based discovery. With multiple ACTIVE sessions, the scan returns 2+ results and the user must specify `--session-id`. This is a UX change from the current unambiguous single-session model.
3. The single-ACTIVE invariant relaxation means that tooling which assumed at most one ACTIVE session must be updated. Known callers: `FindActiveSession()`, `ari session create` (reject-if-active check), autopark logic.

**Neutral**

1. No schema changes to `SESSION_CONTEXT.md`. Session state, FSM transitions, and serialization are unchanged.
2. The `.cc-map/` directory lives inside `.sos/sessions/`, which is already gitignored. No git implications.
3. Tests that set up `.current-session` will be migrated to use `.cc-map/` instead. The test surface area is moderate (approximately 10 test files reference `.current-session`).

---

## Relationship to Prior ADRs

| ADR | Relationship |
|-----|-------------|
| ADR-0001 (Session State Machine) | **Partially superseded**. The single-ACTIVE invariant ("only one ACTIVE session per `.claude/` directory") is relaxed to "single-ACTIVE-per-CC-instance." The FSM itself (ACTIVE/PARKED/ARCHIVED states and valid transitions) is unchanged. |
| ADR-0022 (Session Model) | **Risk #1 resolved**. The `.current-session` TOCTOU race identified in D2 is eliminated by removing `.current-session` entirely. Decision 1 (single-session-per-repo) is superseded by the per-CC-instance model. Decision 2 (scan-based discovery) is retained as fallback for non-CC contexts. |
| ADR-0005 (Moirai Centralized State Authority) | **Unaffected**. Moirai continues to be the convention-layer authority for agent-driven mutations. The event system changes do not alter Moirai's role. |
| ADR-0013 (Moirai Consolidation) | **Unaffected**. CLI remains the enforcement layer. |

## Alternatives Considered

### Alternative 1: Keep Two Event Systems, Add a Unified Reader

Add a `ReadAllEvents()` function that detects the schema of each line (by checking for `timestamp` vs. `ts` field) and normalizes into a common struct.

**Rejected**: This solves the read problem but perpetuates the write problem. Two writer implementations, two schema definitions, two sets of constructors. Every new event type must be added to both systems. The maintenance cost grows linearly with event type count. Convergence on one system has constant maintenance cost.

### Alternative 2: Converge on `session.Event` Instead of `clewcontract.Event`

Make `session.Event` the sole schema and migrate clew events into it.

**Rejected**: `session.Event` has a narrower schema (no `tool`, `path`, or `summary` fields -- these would all need to be crammed into `metadata`). Its SCREAMING_CASE type names are inconsistent with the snake_case used by all 15 clew types. `clewcontract.Event` already has constructors for tool events, decision stamps, artifact creation, handoffs, and quality signals. Starting from `session.Event` would require more migration work with a less capable result.

### Alternative 3: CC Session Map via Environment Variable Instead of File

Store the CC-to-Knossos mapping in an environment variable (e.g., `KNOSSOS_SESSION_ID`) that is set by the `SessionStart` hook and inherited by subsequent hooks.

**Rejected**: CC hooks are stateless processes. Each hook invocation is a separate `ari` process execution. Environment variables set in one hook invocation do not propagate to the next. Only filesystem-based state persists between hook invocations.

### Alternative 4: Use SQLite for CC Session Map

Store mappings in a SQLite database instead of individual files.

**Rejected**: Same rationale as ADR-0022's SQLite rejection -- requires `CGO_ENABLED=1`, contradicts static binary requirement. A directory of small files provides the same O(1) lookup with no dependencies.

### Alternative 5: Keep `.current-session` Alongside CC Map

Maintain `.current-session` as a convenience for the "default single-session" case while adding `.cc-map/` for multi-session scenarios.

**Rejected**: Two resolution mechanisms create ambiguity. If `.current-session` says K1 but `.cc-map/X` says K2, which wins? Defining precedence rules adds complexity. Removing `.current-session` entirely produces a simpler, unambiguous system.

## References

| Reference | Location |
|-----------|----------|
| System A Event Types | `internal/session/events.go` -- 9 `EventType` constants, `Event` struct, `EventEmitter` |
| System B Event Types | `internal/hook/clewcontract/event.go` -- 15 `EventType` constants, `Event` struct |
| BufferedEventWriter | `internal/hook/clewcontract/writer.go` -- async writer with flush loop |
| RecordToolEvent | `internal/hook/clewcontract/record.go` -- hook integration point |
| StdinPayload | `internal/hook/env.go` -- CC session_id delivery via stdin JSON |
| .current-session Cache | `internal/cmd/common/context.go` -- `GetCurrentSessionID()`, `SetCurrentSessionID()` |
| Scan-Based Discovery | `internal/session/discovery.go` -- `FindActiveSession()` |
| Rotation Constants | `internal/session/rotation.go` -- `DefaultMaxLines`, `DefaultKeepLines` |
| Session Paths | `internal/paths/paths.go` -- `CurrentSessionFile()`, `SessionsDir()` |

## Version History

| Date | Author | Change |
|------|--------|--------|
| 2026-02-09 | Claude Opus 4.6 (Context Architect) | Initial proposal -- unified event system and CC session map |

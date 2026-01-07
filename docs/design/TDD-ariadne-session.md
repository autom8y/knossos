# TDD: Ariadne Session Domain

> Technical Design Document for the session domain of the Ariadne Go CLI

**Status**: Draft
**Author**: Architect Agent
**Date**: 2026-01-04
**PRD**: docs/requirements/PRD-ariadne.md
**Spike**: docs/spikes/SPIKE-ariadne-go-cli-architecture.md

---

## 1. Overview

This Technical Design Document specifies the implementation of the **session domain** for Ariadne (`ari`), the Go binary replacement for the roster bash script harness. The session domain encompasses 11 commands that manage the "thread" -- the session lifecycle that enables deterministic return paths through complex multi-agent workflows.

### 1.1 Context

| Reference | Location |
|-----------|----------|
| PRD | `docs/requirements/PRD-ariadne.md` |
| Spike | `docs/spikes/SPIKE-ariadne-go-cli-architecture.md` |
| Gap Analysis | `docs/analysis/GAP-ariadne-prd-review.md` |
| Current Implementation | `.claude/hooks/lib/session-manager.sh`, `.claude/hooks/lib/session-fsm.sh` |
| Session Schema | `schemas/artifacts/session-context.schema.json` |
| Common Schema | `schemas/artifacts/common.schema.json` |

### 1.2 Scope

**In Scope**:
- 11 session commands: create, status, list, park, resume, wrap, transition, migrate, audit, lock, unlock
- Internal packages: `cmd/session/`, `lock/`, `paths/`, `validation/`, `output/`
- Error handling with exit codes per PRD Section 5.1
- Concurrency model with flock() and stale detection per PRD Section 5.3
- Resolution of GAP-3 (audit output), GAP-4 (dry-run scope), GAP-5 (error codes)

**Out of Scope**:
- Team, manifest, and sync domains (separate TDDs)
- State-mate agent integration (documented, but implementation is in state-mate.md)
- Shell completion (deferred to v1.1+)

### 1.3 Design Goals

1. **Behavioral Specification**: Define interfaces that satisfy PRD requirements independent of bash quirks
2. **Type Safety**: Leverage Go's type system to prevent the string parsing errors in bash
3. **Concurrency Safety**: Implement advisory locking with stale detection
4. **Testability**: Enable unit, integration, and race condition testing
5. **Clean Contracts**: JSON output contracts that state-mate can reliably parse

---

## 2. Architecture

### 2.1 Package Structure

```
ariadne/
├── cmd/
│   └── ari/
│       └── main.go                 # Entry point (minimal)
├── internal/
│   ├── cmd/
│   │   ├── root.go                 # Root command with global flags
│   │   └── session/
│   │       ├── session.go          # Parent command registration
│   │       ├── create.go           # ari session create
│   │       ├── status.go           # ari session status
│   │       ├── list.go             # ari session list
│   │       ├── park.go             # ari session park
│   │       ├── resume.go           # ari session resume
│   │       ├── wrap.go             # ari session wrap
│   │       ├── transition.go       # ari session transition
│   │       ├── migrate.go          # ari session migrate
│   │       ├── audit.go            # ari session audit
│   │       ├── lock.go             # ari session lock
│   │       └── unlock.go           # ari session unlock
│   ├── lock/
│   │   ├── lock.go                 # Advisory locking implementation
│   │   ├── flock.go                # flock-based locking (Linux/macOS)
│   │   └── stale.go                # Stale lock detection
│   ├── paths/
│   │   ├── paths.go                # Path resolution and discovery
│   │   ├── xdg.go                  # XDG directory helpers
│   │   └── project.go              # Project root discovery
│   ├── validation/
│   │   ├── validator.go            # Schema validation engine
│   │   ├── session.go              # Session-specific validation
│   │   └── loader.go               # Schema loading from embedded FS
│   ├── output/
│   │   ├── printer.go              # Format-aware output
│   │   ├── json.go                 # JSON output formatting
│   │   ├── text.go                 # Text/table output formatting
│   │   └── verbose.go              # JSON lines verbose logging
│   ├── session/
│   │   ├── context.go              # SESSION_CONTEXT.md parsing/writing
│   │   ├── fsm.go                  # Finite state machine logic
│   │   ├── id.go                   # Session ID generation
│   │   └── events.go               # Event emission to JSONL
│   └── errors/
│       └── errors.go               # Domain-specific error types
├── schemas/
│   ├── session-context.schema.json # Embedded session schema
│   └── common.schema.json          # Embedded common definitions
└── go.mod
```

### 2.2 Dependency Graph

```
                    ┌─────────────────────────────┐
                    │  cmd/ari/main.go            │
                    └─────────────┬───────────────┘
                                  │
                                  v
                    ┌─────────────────────────────┐
                    │  internal/cmd/root.go       │
                    │  (global flags, config)     │
                    └─────────────┬───────────────┘
                                  │
                    ┌─────────────┴───────────────┐
                    │                             │
                    v                             v
         ┌─────────────────────┐      ┌─────────────────────┐
         │ internal/cmd/session│      │ internal/cmd/{team, │
         │ (11 commands)       │      │  manifest, sync}    │
         └─────────┬───────────┘      └─────────────────────┘
                   │
     ┌─────────────┼─────────────┬─────────────┬─────────────┐
     │             │             │             │             │
     v             v             v             v             v
┌─────────┐  ┌─────────┐  ┌─────────┐  ┌───────────┐  ┌─────────┐
│ lock/   │  │ paths/  │  │validate/│  │ output/   │  │ session/│
│         │  │         │  │         │  │           │  │         │
└────┬────┘  └────┬────┘  └────┬────┘  └───────────┘  └────┬────┘
     │            │            │                           │
     └────────────┴────────────┴───────────────────────────┘
                              │
                              v
                    ┌─────────────────────────────┐
                    │  adrg/xdg, cobra, viper,    │
                    │  jsonschema, json-patch     │
                    └─────────────────────────────┘
```

### 2.3 External Dependencies

Per spike recommendations (confirmed in PRD Section 3.2):

| Purpose | Library | Version | Import Path |
|---------|---------|---------|-------------|
| CLI Framework | spf13/cobra | v1.8+ | `github.com/spf13/cobra` |
| Config | spf13/viper | v1.18+ | `github.com/spf13/viper` |
| JSON Schema | santhosh-tekuri/jsonschema | v6+ | `github.com/santhosh-tekuri/jsonschema/v6` |
| XDG Paths | adrg/xdg | v0.5+ | `github.com/adrg/xdg` |
| JSON Merge | evanphx/json-patch | v5+ | `github.com/evanphx/json-patch/v5` |
| YAML | gopkg.in/yaml.v3 | v3 | `gopkg.in/yaml.v3` |

---

## 3. Interface Contracts

### 3.1 Command Summary

| Command | Description | Requires Lock | Modifies State |
|---------|-------------|---------------|----------------|
| `create` | Create new session (NONE -> ACTIVE) | Yes (exclusive) | Yes |
| `status` | Show session state | Yes (shared) | No |
| `list` | List sessions with filters | No | No |
| `park` | Suspend session (ACTIVE -> PARKED) | Yes (exclusive) | Yes |
| `resume` | Resume session (PARKED -> ACTIVE) | Yes (exclusive) | Yes |
| `wrap` | Complete session (ACTIVE/PARKED -> ARCHIVED) | Yes (exclusive) | Yes |
| `transition` | Change workflow phase | Yes (exclusive) | Yes |
| `migrate` | Upgrade session schema (v1 -> v2) | Yes (exclusive) | Yes |
| `audit` | Show session event history | Yes (shared) | No |
| `lock` | Manually acquire session lock | Yes (exclusive) | No |
| `unlock` | Manually release session lock | No | No |

### 3.2 Command: `ari session create`

Creates a new session, transitioning from NONE to ACTIVE state.

**Signature**:
```
ari session create <initiative> [--complexity=MODULE] [--team=NAME]
```

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--complexity` | `-c` | string | `MODULE` | Complexity level: PATCH, MODULE, SYSTEM, INITIATIVE, MIGRATION |
| `--team` | `-t` | string | (from ACTIVE_RITE) | Team pack to activate |

**Output (JSON)**:
```json
{
  "session_id": "session-20260104-160414-563c681e",
  "session_dir": ".claude/sessions/session-20260104-160414-563c681e",
  "status": "ACTIVE",
  "initiative": "Ariadne Go CLI",
  "complexity": "MODULE",
  "team": "10x-dev",
  "created_at": "2026-01-04T16:04:14Z",
  "schema_version": "2.1"
}
```

**Output (text)**:
```
Created session: session-20260104-160414-563c681e
Initiative: Ariadne Go CLI
Complexity: MODULE
Team: 10x-dev
```

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Session created successfully |
| 1 | Session already exists |
| 3 | Lock timeout |
| 4 | Schema validation failed |
| 9 | No .claude/ directory found |

**Error Response (JSON)**:
```json
{
  "error": {
    "code": "SESSION_EXISTS",
    "message": "Session already active. Use 'ari session park' first or 'ari session wrap' to finalize.",
    "details": {
      "existing_session": "session-20260104-150000-abcd1234",
      "status": "ACTIVE"
    }
  }
}
```

**Implementation Notes**:
- Generates session ID: `session-YYYYMMDD-HHMMSS-{8-hex}`
- Creates session directory at `.claude/sessions/{session-id}/`
- Creates SESSION_CONTEXT.md with schema_version 2.1
- Sets `.claude/sessions/.current-session` to new session ID
- Emits `SESSION_CREATED` event to `events.jsonl`
- Acquires exclusive lock during creation, releases after

### 3.3 Command: `ari session status`

Returns current session state with comprehensive metadata.

**Signature**:
```
ari session status [--session-id=ID]
```

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--session-id` | `-s` | string | (current) | Session ID to query (overrides current) |

**Output (JSON)**:
```json
{
  "session_id": "session-20260104-160414-563c681e",
  "session_dir": ".claude/sessions/session-20260104-160414-563c681e",
  "has_session": true,
  "status": "ACTIVE",
  "initiative": "Ariadne Go CLI",
  "complexity": "MODULE",
  "current_phase": "design",
  "active_rite": "10x-dev",
  "execution_mode": "orchestrated",
  "created_at": "2026-01-04T16:04:14Z",
  "schema_version": "2.1",
  "git_branch": "feature/ariadne",
  "git_changes": 3
}
```

**Output (text)**:
```
Session: session-20260104-160414-563c681e
Status: ACTIVE
Initiative: Ariadne Go CLI
Phase: design
Team: 10x-dev
Mode: orchestrated
Branch: feature/ariadne (3 changes)
```

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Status retrieved successfully |
| 6 | Session not found (FILE_NOT_FOUND) |
| 9 | No .claude/ directory found |

**Implementation Notes**:
- Acquires shared lock for consistent read
- Returns `has_session: false` if no current session (not an error)
- Derives `execution_mode` from session state and team configuration

### 3.4 Command: `ari session list`

Lists sessions with optional filtering.

**Signature**:
```
ari session list [--all] [--status=STATUS] [--limit=N]
```

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--all` | `-a` | bool | false | Include archived sessions |
| `--status` | | string | | Filter by status: ACTIVE, PARKED, ARCHIVED |
| `--limit` | `-n` | int | 20 | Maximum sessions to return |

**Output (JSON)**:
```json
{
  "sessions": [
    {
      "session_id": "session-20260104-160414-563c681e",
      "status": "ACTIVE",
      "initiative": "Ariadne Go CLI",
      "complexity": "MODULE",
      "created_at": "2026-01-04T16:04:14Z",
      "current": true
    },
    {
      "session_id": "session-20260103-100000-deadbeef",
      "status": "PARKED",
      "initiative": "Documentation Update",
      "complexity": "PATCH",
      "created_at": "2026-01-03T10:00:00Z",
      "parked_at": "2026-01-03T12:00:00Z",
      "current": false
    }
  ],
  "total": 2,
  "current_session": "session-20260104-160414-563c681e"
}
```

**Output (text)**:
```
SESSION ID                              STATUS   INITIATIVE              CREATED
* session-20260104-160414-563c681e      ACTIVE   Ariadne Go CLI          2026-01-04
  session-20260103-100000-deadbeef      PARKED   Documentation Update    2026-01-03

Total: 2 sessions (* = current)
```

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | List retrieved successfully (even if empty) |
| 9 | No .claude/ directory found |

**Implementation Notes**:
- Scans `.claude/sessions/session-*` directories
- With `--all`, also scans `.claude/.archive/sessions/`
- No locking required (reads directory listing)
- Sorts by created_at descending (most recent first)

### 3.5 Command: `ari session park`

Suspends the current session (ACTIVE -> PARKED).

**Signature**:
```
ari session park [--reason=TEXT] [--session-id=ID]
```

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--reason` | `-r` | string | "Manual park" | Reason for parking |
| `--session-id` | `-s` | string | (current) | Session to park |

**Output (JSON)**:
```json
{
  "session_id": "session-20260104-160414-563c681e",
  "status": "PARKED",
  "previous_status": "ACTIVE",
  "parked_at": "2026-01-04T18:00:00Z",
  "parked_reason": "End of day",
  "git_status": "uncommitted changes"
}
```

**Output (text)**: Silent on success (exit 0, no output)

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Session parked successfully |
| 3 | Lock timeout |
| 5 | Lifecycle violation (already parked or archived) |
| 6 | Session not found |

**Error Response (JSON)**:
```json
{
  "error": {
    "code": "LIFECYCLE_VIOLATION",
    "message": "Cannot park session: already parked",
    "details": {
      "current_status": "PARKED",
      "requested_transition": "ACTIVE -> PARKED"
    }
  }
}
```

**Implementation Notes**:
- Validates transition ACTIVE -> PARKED via FSM
- Records git status (clean/uncommitted changes) in event
- Emits `SESSION_PARKED` event to `events.jsonl`
- Updates `status` field in SESSION_CONTEXT.md

### 3.6 Command: `ari session resume`

Resumes a parked session (PARKED -> ACTIVE).

**Signature**:
```
ari session resume [--session-id=ID]
```

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--session-id` | `-s` | string | (current) | Session to resume |

**Output (JSON)**:
```json
{
  "session_id": "session-20260104-160414-563c681e",
  "status": "ACTIVE",
  "previous_status": "PARKED",
  "resumed_at": "2026-01-05T09:00:00Z"
}
```

**Output (text)**: Silent on success

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Session resumed successfully |
| 3 | Lock timeout |
| 5 | Lifecycle violation (not parked) |
| 6 | Session not found |

**Implementation Notes**:
- Validates transition PARKED -> ACTIVE via FSM
- Sets `.current-session` to resumed session ID
- Emits `SESSION_RESUMED` event to `events.jsonl`

### 3.7 Command: `ari session wrap`

Completes a session, transitioning to ARCHIVED state.

**Signature**:
```
ari session wrap [--session-id=ID] [--skip-checks] [--no-archive]
```

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--session-id` | `-s` | string | (current) | Session to wrap |
| `--skip-checks` | | bool | false | Skip quality gate checks |
| `--no-archive` | | bool | false | Don't move to archive directory |

**Output (JSON)**:
```json
{
  "session_id": "session-20260104-160414-563c681e",
  "status": "ARCHIVED",
  "previous_status": "ACTIVE",
  "wrapped_at": "2026-01-04T20:00:00Z",
  "archived": true,
  "archive_path": ".claude/.archive/sessions/session-20260104-160414-563c681e"
}
```

**Output (text)**: Silent on success

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Session wrapped successfully |
| 3 | Lock timeout |
| 5 | Lifecycle violation (already archived) |
| 6 | Session not found |

**Implementation Notes**:
- Valid transitions: ACTIVE -> ARCHIVED, PARKED -> ARCHIVED
- Clears `.current-session` file
- Moves session directory to `.claude/.archive/sessions/` unless `--no-archive`
- Emits `SESSION_ARCHIVED` event to `events.jsonl`

### 3.8 Command: `ari session transition`

Transitions between workflow phases within an active session.

**Signature**:
```
ari session transition <phase> [--session-id=ID] [--force]
```

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `phase` | Yes | Target phase: requirements, design, implementation, validation, complete |

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--session-id` | `-s` | string | (current) | Session to transition |
| `--force` | `-f` | bool | false | Skip artifact validation |

**Output (JSON)**:
```json
{
  "session_id": "session-20260104-160414-563c681e",
  "from_phase": "requirements",
  "to_phase": "design",
  "transitioned_at": "2026-01-04T17:00:00Z",
  "artifacts_validated": true
}
```

**Output (text)**: Silent on success

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Phase transition successful |
| 3 | Lock timeout |
| 5 | Lifecycle violation (session not ACTIVE, or missing required artifacts) |
| 6 | Session not found |

**Error Response (JSON)**:
```json
{
  "error": {
    "code": "LIFECYCLE_VIOLATION",
    "message": "Cannot transition to design: missing required artifacts",
    "details": {
      "from_phase": "requirements",
      "to_phase": "design",
      "missing_artifacts": ["PRD: No PRD found in docs/requirements/"]
    }
  }
}
```

**Implementation Notes**:
- Validates artifact requirements per phase:
  - `design` requires PRD in `docs/requirements/`
  - `implementation` requires TDD in `docs/design/`
  - `complete` requires test plan in `docs/testing/`
- `--force` bypasses artifact validation (logged in event)
- Emits `PHASE_TRANSITIONED` event to `events.jsonl`

### 3.9 Command: `ari session migrate`

Migrates session(s) from v1 to v2 schema format.

**Signature**:
```
ari session migrate [--session-id=ID] [--all] [--dry-run]
```

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--session-id` | `-s` | string | | Specific session to migrate |
| `--all` | `-a` | bool | false | Migrate all v1 sessions |
| `--dry-run` | | bool | false | Preview changes without applying |

**Output (JSON)**:
```json
{
  "migrated": [
    {
      "session_id": "session-20251230-100000-old12345",
      "from_version": "1.0",
      "to_version": "2.1",
      "status_derived": "PARKED",
      "fields_migrated": ["session_state -> status", "parked_at -> events.jsonl"],
      "backup_path": ".claude/sessions/session-20251230-100000-old12345/SESSION_CONTEXT.v1.backup"
    }
  ],
  "skipped": [
    {
      "session_id": "session-20260104-160414-563c681e",
      "reason": "Already v2"
    }
  ],
  "failed": [],
  "total_migrated": 1,
  "total_skipped": 1,
  "total_failed": 0,
  "dry_run": false
}
```

**Output (text)**:
```
Migrating session-20251230-100000-old12345...
  Status derived: PARKED (from parked_at field)
  Migrated: session_state -> status
  Migrated: parked_at -> events.jsonl
  Backup: SESSION_CONTEXT.v1.backup
  Done.

Skipped session-20260104-160414-563c681e (already v2)

Migration complete: 1 migrated, 1 skipped, 0 failed
```

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Migration completed (even with skips) |
| 1 | One or more migrations failed |
| 9 | No .claude/ directory found |

**Implementation Notes**:
- Creates backup at `SESSION_CONTEXT.v1.backup` before migration
- Derives `status` from v1 fields (`parked_at` -> PARKED, `completed_at` -> ARCHIVED, else ACTIVE)
- Moves park metadata to `events.jsonl`
- Validates migrated file against v2 schema; rollback on failure
- `--dry-run` supported per GAP-4 resolution

### 3.10 Command: `ari session audit`

Displays session event history (resolves GAP-3).

**Signature**:
```
ari session audit [--session-id=ID] [--limit=N] [--event-type=TYPE] [--since=TIMESTAMP]
```

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--session-id` | `-s` | string | (current) | Session to audit |
| `--limit` | `-n` | int | 50 | Maximum events to return |
| `--event-type` | `-e` | string | | Filter by event type |
| `--since` | | string | | Only events after this ISO8601 timestamp |

**Output (JSON)** (GAP-3 Resolution):
```json
{
  "session_id": "session-20260104-160414-563c681e",
  "events": [
    {
      "timestamp": "2026-01-04T16:04:14Z",
      "event": "SESSION_CREATED",
      "from": "NONE",
      "to": "ACTIVE",
      "metadata": {
        "initiative": "Ariadne Go CLI",
        "complexity": "MODULE",
        "team": "10x-dev"
      }
    },
    {
      "timestamp": "2026-01-04T17:00:00Z",
      "event": "PHASE_TRANSITIONED",
      "from_phase": "requirements",
      "to_phase": "design",
      "metadata": {
        "artifacts_validated": true
      }
    },
    {
      "timestamp": "2026-01-04T18:00:00Z",
      "event": "SESSION_PARKED",
      "from": "ACTIVE",
      "to": "PARKED",
      "metadata": {
        "reason": "End of day",
        "git_status": "uncommitted changes"
      }
    }
  ],
  "total": 3,
  "filters_applied": {
    "limit": 50,
    "event_type": null,
    "since": null
  }
}
```

**Output (text)**:
```
TIMESTAMP                EVENT               FROM -> TO          DETAILS
2026-01-04T16:04:14Z     SESSION_CREATED     NONE -> ACTIVE      initiative=Ariadne Go CLI
2026-01-04T17:00:00Z     PHASE_TRANSITIONED  requirements->design artifacts_validated=true
2026-01-04T18:00:00Z     SESSION_PARKED      ACTIVE -> PARKED    reason=End of day

Total: 3 events
```

**Event Types** (GAP-3 Resolution):

| Event Type | Description | Metadata Fields |
|------------|-------------|-----------------|
| `SESSION_CREATED` | Session initialized | initiative, complexity, team |
| `SESSION_PARKED` | Session suspended | reason, git_status |
| `SESSION_RESUMED` | Session resumed | - |
| `SESSION_ARCHIVED` | Session completed | - |
| `PHASE_TRANSITIONED` | Workflow phase changed | from_phase, to_phase, artifacts_validated |
| `LOCK_ACQUIRED` | Manual lock acquired | pid, timestamp |
| `LOCK_RELEASED` | Manual lock released | - |
| `SCHEMA_MIGRATED` | Schema version upgraded | from_version, to_version |

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Audit retrieved successfully |
| 6 | Session not found |
| 9 | No .claude/ directory found |

**Implementation Notes**:
- Reads from `{session-dir}/events.jsonl`
- Also reads from `.claude/sessions/.audit/transitions.log` for global view
- Supports filtering by event type and timestamp range

### 3.11 Command: `ari session lock`

Manually acquires an exclusive lock on a session.

**Signature**:
```
ari session lock [--session-id=ID] [--timeout=SECONDS]
```

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--session-id` | `-s` | string | (current) | Session to lock |
| `--timeout` | `-t` | int | 10 | Lock acquisition timeout in seconds |

**Output (JSON)**:
```json
{
  "session_id": "session-20260104-160414-563c681e",
  "locked": true,
  "lock_path": ".claude/sessions/.locks/session-20260104-160414-563c681e.lock",
  "holder_pid": 12345,
  "acquired_at": "2026-01-04T19:00:00Z"
}
```

**Output (text)**: Silent on success

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Lock acquired successfully |
| 3 | Lock timeout (LOCK_TIMEOUT) |
| 6 | Session not found |

**Implementation Notes**:
- Primarily for debugging and external tooling
- Lock is held until `unlock` command or process termination
- Emits `LOCK_ACQUIRED` event to `events.jsonl`

### 3.12 Command: `ari session unlock`

Manually releases a session lock.

**Signature**:
```
ari session unlock [--session-id=ID] [--force]
```

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--session-id` | `-s` | string | (current) | Session to unlock |
| `--force` | `-f` | bool | false | Force unlock even if not owner |

**Output (JSON)**:
```json
{
  "session_id": "session-20260104-160414-563c681e",
  "unlocked": true,
  "was_stale": false
}
```

**Output (text)**: Silent on success

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Lock released successfully |
| 1 | Not lock owner (without --force) |
| 6 | Session not found |

**Implementation Notes**:
- `--force` allows removing locks held by other processes (e.g., dead processes)
- Validates PID ownership unless `--force` specified
- Emits `LOCK_RELEASED` event to `events.jsonl`

---

## 4. Error Handling

### 4.1 Error Code Taxonomy (GAP-5 Resolution)

Extending PRD Section 5.1 with session-domain-specific codes:

| Code | Exit | Name | Description |
|------|------|------|-------------|
| `SUCCESS` | 0 | Success | Operation completed successfully |
| `GENERAL_ERROR` | 1 | General Error | Unspecified error |
| `USAGE_ERROR` | 2 | Usage Error | Invalid arguments or flags |
| `LOCK_TIMEOUT` | 3 | Lock Timeout | Could not acquire lock within timeout |
| `LOCK_STALE` | 3 | Lock Stale | Lock holder process dead (auto-recovered) |
| `SCHEMA_INVALID` | 4 | Schema Invalid | Data failed schema validation |
| `LIFECYCLE_VIOLATION` | 5 | Lifecycle Violation | Invalid state transition |
| `FILE_NOT_FOUND` | 6 | File Not Found | Required file missing |
| `SESSION_NOT_FOUND` | 6 | Session Not Found | Session directory does not exist |
| `PERMISSION_DENIED` | 7 | Permission Denied | Cannot read/write file |
| `MERGE_CONFLICT` | 8 | Merge Conflict | Three-way merge has conflicts |
| `PROJECT_NOT_FOUND` | 9 | Project Not Found | No .claude/ directory found |
| `SESSION_EXISTS` | 10 | Session Exists | Session already active (for create) |
| `MIGRATION_FAILED` | 11 | Migration Failed | Schema migration failed |

### 4.2 Error Response Structure

All errors follow the PRD Section 4.4 contract:

```go
type ErrorResponse struct {
    Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}
```

Example:
```json
{
  "error": {
    "code": "LIFECYCLE_VIOLATION",
    "message": "Cannot park session: already parked",
    "details": {
      "session_id": "session-20260104-160414-563c681e",
      "current_status": "PARKED",
      "requested_status": "PARKED"
    }
  }
}
```

### 4.3 Partial Write Recovery

All mutations follow this pattern:

```go
func mutateWithRecovery(ctx context.Context, sessionID string, mutationFn func() error) error {
    // 1. Acquire exclusive lock
    if err := lock.Acquire(sessionID, lock.Exclusive, timeout); err != nil {
        return errors.Wrap(err, "LOCK_TIMEOUT")
    }
    defer lock.Release(sessionID)

    // 2. Create backup
    backupPath, err := backup(sessionID)
    if err != nil {
        return errors.Wrap(err, "BACKUP_FAILED")
    }

    // 3. Execute mutation
    if err := mutationFn(); err != nil {
        restore(sessionID, backupPath)
        return err
    }

    // 4. Validate result
    if err := validate(sessionID); err != nil {
        restore(sessionID, backupPath)
        return errors.Wrap(err, "SCHEMA_INVALID")
    }

    // 5. Remove backup on success
    os.Remove(backupPath)
    return nil
}
```

---

## 5. Concurrency Model

### 5.1 Locking Strategy

Per PRD Section 5.3, using `flock()` with stale detection:

```go
package lock

import (
    "os"
    "syscall"
    "time"
)

type LockType int

const (
    Shared    LockType = syscall.LOCK_SH
    Exclusive LockType = syscall.LOCK_EX
)

type Lock struct {
    sessionID string
    file      *os.File
    lockType  LockType
}

// Acquire attempts to acquire a lock with timeout
func Acquire(sessionID string, lockType LockType, timeout time.Duration) (*Lock, error) {
    lockPath := lockFilePath(sessionID)

    // Ensure lock directory exists
    if err := os.MkdirAll(filepath.Dir(lockPath), 0755); err != nil {
        return nil, err
    }

    // Open lock file
    file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
    if err != nil {
        return nil, err
    }

    // Attempt lock with timeout
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        err := syscall.Flock(int(file.Fd()), int(lockType)|syscall.LOCK_NB)
        if err == nil {
            // Success - write PID for debugging
            if lockType == Exclusive {
                file.Truncate(0)
                file.Seek(0, 0)
                fmt.Fprintf(file, "%d\n", os.Getpid())
            }
            return &Lock{sessionID: sessionID, file: file, lockType: lockType}, nil
        }

        // Check for stale lock
        if isStale(lockPath) {
            // Force remove stale lock and retry
            os.Remove(lockPath)
            continue
        }

        time.Sleep(100 * time.Millisecond)
    }

    file.Close()
    return nil, ErrLockTimeout
}

// isStale checks if the lock holder process is dead
func isStale(lockPath string) bool {
    data, err := os.ReadFile(lockPath)
    if err != nil {
        return false
    }

    var pid int
    if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
        return false
    }

    // Check if process exists
    process, err := os.FindProcess(pid)
    if err != nil {
        return true // Can't find process
    }

    // On Unix, FindProcess always succeeds; check with signal 0
    if err := process.Signal(syscall.Signal(0)); err != nil {
        return true // Process doesn't exist
    }

    return false
}

// Release releases the lock
func (l *Lock) Release() error {
    if l.file == nil {
        return nil
    }
    syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
    l.file.Close()
    l.file = nil
    return nil
}
```

### 5.2 Lock File Location

```
.claude/sessions/.locks/
    {session-id}.lock       # flock target file
```

### 5.3 Concurrency Properties

From TLA+ specification (referenced in TDD-session-state-machine.md):

1. **MutualExclusion**: At most one process holds exclusive lock per session
2. **LockEventuallyGranted**: With fair scheduling, every request completes or times out
3. **NoDeadlock**: Single lock per session prevents circular wait
4. **StaleDetection**: Dead process locks are automatically recovered

---

## 6. Data Model

### 6.1 SESSION_CONTEXT.md (v2.1 Schema)

```yaml
---
schema_version: "2.1"
session_id: "session-20260104-160414-563c681e"
status: "ACTIVE"                    # Single source of truth
created_at: "2026-01-04T16:04:14Z"
initiative: "Ariadne Go CLI"
complexity: "MODULE"
active_rite: "10x-dev"
team: "10x-dev"               # Explicit null for cross-cutting
current_phase: "design"
---

# Session: Ariadne Go CLI

## Artifacts
- PRD: docs/requirements/PRD-ariadne.md
- TDD: pending

## Blockers
None yet.

## Next Steps
1. Complete TDD for session domain
```

### 6.2 Session ID Format

Pattern: `session-YYYYMMDD-HHMMSS-{8-hex}`

```go
func GenerateSessionID() string {
    now := time.Now()
    hex := make([]byte, 4)
    rand.Read(hex)
    return fmt.Sprintf("session-%s-%x",
        now.Format("20060102-150405"),
        hex,
    )
}
```

Regex for validation: `^session-[0-9]{8}-[0-9]{6}-[a-f0-9]{8}$`

### 6.3 Event Log Format (events.jsonl)

```jsonl
{"timestamp":"2026-01-04T16:04:14Z","event":"SESSION_CREATED","from":"NONE","to":"ACTIVE","metadata":{"initiative":"Ariadne Go CLI"}}
{"timestamp":"2026-01-04T17:00:00Z","event":"PHASE_TRANSITIONED","from_phase":"requirements","to_phase":"design","metadata":{"artifacts_validated":true}}
{"timestamp":"2026-01-04T18:00:00Z","event":"SESSION_PARKED","from":"ACTIVE","to":"PARKED","metadata":{"reason":"End of day","git_status":"uncommitted changes"}}
```

### 6.4 Global Audit Log (.audit/transitions.log)

```
2026-01-04T16:04:14Z | session-20260104-160414-563c681e | SESSION_CREATED | NONE -> ACTIVE
2026-01-04T17:00:00Z | session-20260104-160414-563c681e | PHASE_TRANSITIONED | requirements -> design
2026-01-04T18:00:00Z | session-20260104-160414-563c681e | SESSION_PARKED | ACTIVE -> PARKED
```

---

## 7. Internal Package Design

### 7.1 Package: `internal/session`

Core session domain logic, independent of CLI.

```go
package session

// Context represents parsed SESSION_CONTEXT.md
type Context struct {
    SchemaVersion string    `yaml:"schema_version"`
    SessionID     string    `yaml:"session_id"`
    Status        Status    `yaml:"status"`
    CreatedAt     time.Time `yaml:"created_at"`
    Initiative    string    `yaml:"initiative"`
    Complexity    string    `yaml:"complexity"`
    ActiveTeam    string    `yaml:"active_rite"`
    CurrentPhase  string    `yaml:"current_phase"`
}

// Status represents session lifecycle state
type Status string

const (
    StatusNone     Status = "NONE"
    StatusActive   Status = "ACTIVE"
    StatusParked   Status = "PARKED"
    StatusArchived Status = "ARCHIVED"
)

// FSM validates and executes state transitions
type FSM struct {
    transitions map[Status][]Status
}

func NewFSM() *FSM {
    return &FSM{
        transitions: map[Status][]Status{
            StatusNone:   {StatusActive},
            StatusActive: {StatusParked, StatusArchived},
            StatusParked: {StatusActive, StatusArchived},
            // StatusArchived has no valid transitions (terminal)
        },
    }
}

func (f *FSM) CanTransition(from, to Status) bool {
    validTargets, ok := f.transitions[from]
    if !ok {
        return false
    }
    for _, target := range validTargets {
        if target == to {
            return true
        }
    }
    return false
}
```

### 7.2 Package: `internal/lock`

Advisory locking with stale detection.

```go
package lock

type Manager interface {
    Acquire(sessionID string, lockType LockType, timeout time.Duration) (*Lock, error)
    Release(sessionID string) error
    IsLocked(sessionID string) bool
    GetHolder(sessionID string) (pid int, err error)
}

type FlockManager struct {
    locksDir string
}

func NewFlockManager(locksDir string) *FlockManager {
    return &FlockManager{locksDir: locksDir}
}
```

### 7.3 Package: `internal/paths`

Path resolution and project discovery.

```go
package paths

import "github.com/adrg/xdg"

type Resolver struct {
    projectRoot string
}

// FindProjectRoot walks up from cwd looking for .claude/
func FindProjectRoot() (string, error) {
    dir, err := os.Getwd()
    if err != nil {
        return "", err
    }

    for {
        claudeDir := filepath.Join(dir, ".claude")
        if info, err := os.Stat(claudeDir); err == nil && info.IsDir() {
            return dir, nil
        }

        parent := filepath.Dir(dir)
        if parent == dir {
            return "", ErrProjectNotFound
        }
        dir = parent
    }
}

// SessionsDir returns the path to sessions directory
func (r *Resolver) SessionsDir() string {
    return filepath.Join(r.projectRoot, ".claude", "sessions")
}

// ConfigDir returns XDG config directory for ariadne
func ConfigDir() string {
    return filepath.Join(xdg.ConfigHome, "ariadne")
}
```

### 7.4 Package: `internal/validation`

Schema validation using embedded schemas.

```go
package validation

import (
    "embed"
    "github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed schemas/*.json
var schemaFS embed.FS

type Validator struct {
    compiler *jsonschema.Compiler
    schemas  map[string]*jsonschema.Schema
}

func NewValidator() (*Validator, error) {
    compiler := jsonschema.NewCompiler()

    // Register embedded schemas
    entries, _ := schemaFS.ReadDir("schemas")
    for _, entry := range entries {
        data, _ := schemaFS.ReadFile("schemas/" + entry.Name())
        compiler.AddResource("embed:///"+entry.Name(), bytes.NewReader(data))
    }

    return &Validator{
        compiler: compiler,
        schemas:  make(map[string]*jsonschema.Schema),
    }, nil
}

func (v *Validator) ValidateSession(data []byte) error {
    schema, err := v.getSchema("session-context.schema.json")
    if err != nil {
        return err
    }

    var parsed interface{}
    if err := json.Unmarshal(data, &parsed); err != nil {
        return fmt.Errorf("invalid JSON: %w", err)
    }

    return schema.Validate(parsed)
}
```

### 7.5 Package: `internal/output`

Format-aware output printing.

```go
package output

type Format string

const (
    FormatText Format = "text"
    FormatJSON Format = "json"
    FormatYAML Format = "yaml"
)

type Printer struct {
    format  Format
    out     io.Writer
    verbose bool
}

func NewPrinter(format Format, out io.Writer, verbose bool) *Printer {
    return &Printer{format: format, out: out, verbose: verbose}
}

func (p *Printer) Print(data interface{}) error {
    switch p.format {
    case FormatJSON:
        enc := json.NewEncoder(p.out)
        enc.SetIndent("", "  ")
        return enc.Encode(data)
    case FormatYAML:
        enc := yaml.NewEncoder(p.out)
        return enc.Encode(data)
    default:
        return p.printText(data)
    }
}

// VerboseLog writes JSON lines to stderr for debugging
func (p *Printer) VerboseLog(level, msg string, fields map[string]interface{}) {
    if !p.verbose {
        return
    }
    entry := map[string]interface{}{
        "level": level,
        "msg":   msg,
        "ts":    time.Now().UTC().Format(time.RFC3339),
    }
    for k, v := range fields {
        entry[k] = v
    }
    json.NewEncoder(os.Stderr).Encode(entry)
}
```

---

## 8. --dry-run Scope (GAP-4 Resolution)

### 8.1 Decision

`--dry-run` is **command-specific**, not a global flag for the session domain.

**Rationale**:
- Most session commands are either read-only (status, list, audit) or inherently stateful (park, resume, wrap)
- Only `migrate` benefits from dry-run (preview schema changes without applying)
- Making it global would be misleading for commands where it doesn't apply

### 8.2 Commands Supporting --dry-run

| Command | --dry-run Support | Behavior |
|---------|-------------------|----------|
| `migrate` | Yes | Preview migrations without writing |
| All others | No | N/A |

### 8.3 Dry-Run Output

When `--dry-run` is specified:

```json
{
  "dry_run": true,
  "would_migrate": [
    {
      "session_id": "session-20251230-100000-old12345",
      "from_version": "1.0",
      "to_version": "2.1",
      "status_would_be": "PARKED",
      "fields_would_migrate": ["session_state -> status"]
    }
  ]
}
```

---

## 9. Test Strategy

### 9.1 Unit Tests

Location: `internal/*_test.go`

| Package | Test Focus | Coverage Target |
|---------|-----------|-----------------|
| `session` | FSM transitions, context parsing | 100% |
| `lock` | Lock acquisition, stale detection | 100% |
| `paths` | Project discovery, path resolution | 100% |
| `validation` | Schema validation, error messages | 100% |
| `output` | Format rendering | 90% |

Example test:

```go
func TestFSM_CanTransition(t *testing.T) {
    fsm := session.NewFSM()

    tests := []struct {
        from, to session.Status
        want     bool
    }{
        {session.StatusNone, session.StatusActive, true},
        {session.StatusActive, session.StatusParked, true},
        {session.StatusActive, session.StatusArchived, true},
        {session.StatusParked, session.StatusActive, true},
        {session.StatusParked, session.StatusArchived, true},
        {session.StatusArchived, session.StatusActive, false},
        {session.StatusArchived, session.StatusParked, false},
        {session.StatusActive, session.StatusNone, false},
    }

    for _, tt := range tests {
        t.Run(fmt.Sprintf("%s->%s", tt.from, tt.to), func(t *testing.T) {
            got := fsm.CanTransition(tt.from, tt.to)
            if got != tt.want {
                t.Errorf("CanTransition(%s, %s) = %v, want %v",
                    tt.from, tt.to, got, tt.want)
            }
        })
    }
}
```

### 9.2 Integration Tests

Location: `tests/integration/session_test.go`

| Test ID | Description | TLA+ Property |
|---------|-------------|---------------|
| `int_001` | Create session generates valid ID and context | TypeInvariant |
| `int_002` | Park-resume cycle preserves session | ValidTransition |
| `int_003` | Wrap moves to archive | ArchivedIsTerminal |
| `int_004` | List filters by status correctly | - |
| `int_005` | Migrate upgrades v1 to v2 | - |
| `int_006` | Concurrent writes are serialized | MutualExclusion |

### 9.3 Concurrency Tests

Location: `tests/concurrency/lock_test.go`

```go
func TestConcurrentPark(t *testing.T) {
    // Setup: Create active session
    sessionID := createTestSession(t)

    // Run two parallel park operations
    var wg sync.WaitGroup
    results := make(chan error, 2)

    for i := 0; i < 2; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            err := parkSession(sessionID, fmt.Sprintf("worker-%d", n))
            results <- err
        }(i)
    }

    wg.Wait()
    close(results)

    // Exactly one should succeed, one should fail
    var successes, failures int
    for err := range results {
        if err == nil {
            successes++
        } else {
            failures++
        }
    }

    assert.Equal(t, 1, successes, "exactly one park should succeed")
    assert.Equal(t, 1, failures, "exactly one park should fail")

    // Final state should be PARKED
    status := getSessionStatus(t, sessionID)
    assert.Equal(t, session.StatusParked, status)
}
```

Run with: `go test -race ./...`

### 9.4 Test Fixtures

```
ariadne/
└── testdata/
    ├── sessions/
    │   ├── v1-active/              # v1 schema, ACTIVE state
    │   │   └── SESSION_CONTEXT.md
    │   ├── v1-parked/              # v1 schema, PARKED state
    │   │   └── SESSION_CONTEXT.md
    │   ├── v2-active/              # v2 schema, ACTIVE state
    │   │   └── SESSION_CONTEXT.md
    │   └── v2-archived/            # v2 schema, ARCHIVED state
    │       └── SESSION_CONTEXT.md
    └── schemas/
        └── session-context.schema.json
```

---

## 10. Migration from Bash

### 10.1 Behavioral Parity

The Go implementation follows the **specification** (PRD), not bash quirks. Known divergences:

| Behavior | Bash Implementation | Go Implementation |
|----------|---------------------|-------------------|
| Empty reason | Defaults to "Manual park" | Same |
| Lock timeout | 10 seconds | 10 seconds (configurable) |
| Status field | Reads from multiple fields | Single source: `status` |
| Event emission | Appends to JSONL | Same |

### 10.2 Integration Path

During migration, bash scripts call `ari`:

```bash
# session-manager.sh (bridge)
case "$1" in
  create) ari session create "${@:2}" ;;
  park)   ari session park "${@:2}" ;;
  *)      echo "Unknown command: $1" >&2; exit 1 ;;
esac
```

### 10.3 Post-v1.0 Cleanup

After v1.0 ships:
- Delete `session-manager.sh`
- Delete `session-fsm.sh`
- Delete `session-migrate.sh`
- Update state-mate.md to invoke `ari session *` directly

---

## 11. Implementation Guidance

### 11.1 Recommended Order

1. **Foundation** (Week 1)
   - `internal/paths` - Project discovery
   - `internal/lock` - Advisory locking
   - `internal/validation` - Schema validation with embedded schemas
   - `internal/output` - Format-aware printing

2. **Core Session** (Week 2)
   - `internal/session` - Context parsing, FSM, events
   - `cmd/session/create.go`
   - `cmd/session/status.go`
   - `cmd/session/list.go`

3. **State Transitions** (Week 3)
   - `cmd/session/park.go`
   - `cmd/session/resume.go`
   - `cmd/session/wrap.go`
   - `cmd/session/transition.go`

4. **Utilities** (Week 4)
   - `cmd/session/migrate.go`
   - `cmd/session/audit.go`
   - `cmd/session/lock.go`
   - `cmd/session/unlock.go`

5. **Integration** (Week 5)
   - Integration tests
   - Concurrency tests with `-race`
   - Bash bridge wiring

### 11.2 Dependency Injection

For testability, all I/O operations go through interfaces:

```go
type SessionStore interface {
    Create(ctx Context) (string, error)
    Read(sessionID string) (*Context, error)
    Update(sessionID string, ctx Context) error
    Delete(sessionID string) error
    List(filter ListFilter) ([]Context, error)
}

type FileSystemStore struct {
    sessionsDir string
    validator   *validation.Validator
    locker      lock.Manager
}
```

---

## 12. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| flock unavailable on some systems | Low (macOS has it) | Medium | Test on target platforms in CI |
| Lock contention under high concurrency | Low | Medium | Short critical sections, configurable timeout |
| Schema validation performance | Low | Low | Cache compiled schemas |
| Migration corrupts v1 sessions | Low | High | Create backups, validate post-migration, provide rollback |
| Behavioral parity gaps | Medium | High | Spec-based testing, integration tests |

---

## 13. ADRs

| ADR | Status | Topic |
|-----|--------|-------|
| ADR-ariadne-001 | Proposed | Cobra CLI framework selection |
| ADR-ariadne-002 | Proposed | flock-based locking strategy |
| ADR-ariadne-003 | Proposed | Embedded schema validation |

---

## 14. Handoff Criteria

Ready for Implementation when:

- [x] All 11 session commands have interface contracts
- [x] Internal package boundaries defined
- [x] Test scenarios cover critical paths
- [x] GAP-3 (audit output) resolved - Section 3.10
- [x] GAP-4 (dry-run scope) resolved - Section 8
- [x] GAP-5 (error codes) resolved - Section 4.1
- [ ] Principal Engineer can implement without architectural questions
- [ ] All artifacts verified via Read tool

---

## 15. Artifact Attestation

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-ariadne-session.md` | Read |
| PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-ariadne.md` | Read |
| Spike | `/Users/tomtenuta/Code/roster/docs/spikes/SPIKE-ariadne-go-cli-architecture.md` | Read |
| Gap Analysis | `/Users/tomtenuta/Code/roster/docs/analysis/GAP-ariadne-prd-review.md` | Read |
| Session Schema | `/Users/tomtenuta/Code/roster/schemas/artifacts/session-context.schema.json` | Read |
| Common Schema | `/Users/tomtenuta/Code/roster/schemas/artifacts/common.schema.json` | Read |
| Current session-manager.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-manager.sh` | Read |
| Current session-fsm.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-fsm.sh` | Read |

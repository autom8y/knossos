# ADR-0001: Session State Machine Redesign

| Field | Value |
|-------|-------|
| **Status** | Accepted |
| **Date** | 2025-12-31 |
| **Deciders** | Architecture Team |
| **Supersedes** | N/A (foundational) |
| **Superseded by** | N/A |

## Context

The roster session management system has evolved organically over multiple iterations, resulting in several architectural problems that create confusion, bugs, and maintenance burden.

### Current System Issues

#### 1. Field Naming Chaos

The SESSION_CONTEXT.md schema contains duplicate fields with inconsistent naming:

| Canonical Field | Legacy Field | Issue |
|-----------------|--------------|-------|
| `status` | `session_state` | Both defined in schema, unclear precedence |
| `parked_reason` | `park_reason` | Code handles both (lines 583-590 session-manager.sh) |
| `parked_git_status` | `git_status_at_park` | Same pattern |

The `extract_session_fields()` function (lines 60-87, session-manager.sh) demonstrates this confusion:
```bash
# Read status field (canonical), fallback to session_state (legacy)
local explicit_state=$({ grep -m1 "^status:" "$ctx_file" 2>/dev/null || \
    grep -m1 "^session_state:" "$ctx_file" 2>/dev/null || true; } | ...)
```

#### 2. Dual State Determination

State is determined through two incompatible mechanisms:

1. **Explicit field reading**: `extract_session_fields()` reads `status`/`session_state` field
2. **Presence-based inference**: `get_session_state()` (session-state.sh, lines 21-52) ignores explicit fields entirely:

```bash
# Note: This infers state from presence of park fields, not from reading the status field.
get_session_state() {
    # Check for auto_parked_at first (more specific)
    if grep -q "^auto_parked_at:" "$session_file" 2>/dev/null; then
        echo "AUTO_PARKED"
        return 0
    fi
    # Check for parked_at
    if grep -q "^parked_at:" "$session_file" 2>/dev/null; then
        echo "PARKED"
        return 0
    fi
    echo "ACTIVE"
}
```

This creates a situation where:
- The `status` field might say `ACTIVE`
- But `parked_at` field exists
- Different functions return different states for the same session

#### 3. Race Conditions

The locking strategy is inconsistent:

| Operation | Lock Type | Lock Scope |
|-----------|-----------|------------|
| `cmd_status()` | None | Reads unlocked |
| `cmd_create()` | `mkdir .create.lock` | Session creation only |
| `cmd_mutate()` | `mkdir .mutate.lock` | Full mutation |
| `get_session_state()` | None | Reads unlocked |
| `atomic_team_update()` | `acquire_session_lock` | Separate lock mechanism |

A typical race condition scenario:
1. Process A calls `cmd_status()` (unlocked read)
2. Process B calls `cmd_mutate park` (acquires lock)
3. Process B updates `parked_at` field
4. Process A returns stale state (shows ACTIVE when actually PARKED)

#### 4. Undefined State Transitions

No formal definition of valid state transitions exists. The code implicitly assumes certain sequences but doesn't enforce them:

```
ACTIVE -> PARKED: Allowed (mutate_park checks not already parked)
PARKED -> ACTIVE: Allowed (mutate_resume checks is parked)
ACTIVE -> ARCHIVED: Allowed (mutate_wrap)
PARKED -> ARCHIVED: Allowed? (code doesn't prevent it)
ARCHIVED -> ACTIVE: Undefined (no mechanism exists)
```

#### 5. Substate Confusion

Workflow phases (requirements, design, implementation, validation) are treated as separate from session state, but they're actually substates of ACTIVE. The current model doesn't represent this hierarchy.

### Forces

- **Backward Compatibility**: Existing sessions must migrate cleanly
- **Multi-Process Safety**: Multiple Claude Code instances can run concurrently
- **Debuggability**: State should be deterministic and inspectable
- **Extensibility**: New states/substates should be addable without redesign
- **Performance**: State checks happen frequently (every hook invocation)

## Decision

We will implement a full redesign of the session state machine with the following characteristics:

### 1. Explicit Finite State Machine

Replace implicit state determination with an explicit FSM model:

```
                    ┌─────────────────────────────────────────────┐
                    │                   ACTIVE                     │
                    │  ┌─────────────────────────────────────────┐ │
                    │  │ substates (from ACTIVE_WORKFLOW.yaml):  │ │
    ┌───────────────┼──│  requirements → design → implementation │─┼──► ARCHIVED
    │               │  │  → validation                           │ │
    │               │  └─────────────────────────────────────────┘ │
    │               └──────────────────────────────────────────────┘
    │                     │                      ▲
    │                     │ /park                │ /resume
    │                     ▼                      │
    │               ┌─────────────────────────────────────────────┐
    │               │                  PARKED                      │
    │               │  (preserves last_active_phase for resume)   │
    └───────────────┼─────────────────────────────────────────────┘
     stale/cleanup  │
                    ▼
              ┌─────────────────────────────────────────────┐
              │                 ARCHIVED                     │
              │  (terminal state, read-only)                │
              └─────────────────────────────────────────────┘
```

### 2. Minimal Top-Level States

Only three top-level states:

| State | Description | Valid Transitions |
|-------|-------------|-------------------|
| `ACTIVE` | Session is in progress | PARKED, ARCHIVED |
| `PARKED` | Session is suspended, resumable | ACTIVE, ARCHIVED |
| `ARCHIVED` | Session is complete, immutable | None (terminal) |

Rationale: The existing `COMPLETED` status is redundant with `ARCHIVED`. A completed session is archived immediately.

### 3. Hierarchical Substates

Substates are derived from `ACTIVE_WORKFLOW.yaml` and only valid within `ACTIVE`:

```yaml
# Example: 10x-dev-pack workflow defines these substates
substates:
  - requirements
  - design
  - implementation
  - validation
```

Substate transitions are governed by workflow rules, not the FSM itself.

### 4. Single Source of Truth

The `status` field is the ONLY authority for state:

```yaml
# SESSION_CONTEXT.md (v2)
---
session_id: "session-20251231-120000-abcd1234"
status: "ACTIVE"                    # ONLY this field determines state
current_phase: "design"             # Substate within ACTIVE
# parked_at: removed (status=PARKED is sufficient)
# session_state: removed (duplicate)
---
```

Park/resume metadata becomes event-sourced in an audit log rather than state fields.

### 5. Advisory File Locking

All state reads and writes use advisory locking:

```bash
# Read pattern
flock -s "$LOCK_FILE" cat "$SESSION_FILE"  # Shared lock for reads

# Write pattern
flock -x "$LOCK_FILE" write_session        # Exclusive lock for writes
```

Platform compatibility:
- **macOS/Linux**: Use `flock` (available via coreutils on macOS)
- **Fallback**: `mkdir`-based locking (current approach) with timeout

### 6. Runtime Schema Validation

All writes to SESSION_CONTEXT.md are validated against JSON Schema before commit:

```bash
# Before any write
validate_against_schema "$content" "session-context.schema.json" || {
    emit_event "VALIDATION_FAILED" "$content"
    fail_fast "Session context validation failed"
}
```

### 7. Fail-Fast with Event Emission

Invalid state transitions fail immediately with clear messages:

```bash
# Example: attempting ARCHIVED -> ACTIVE
{
  "error": "INVALID_TRANSITION",
  "from_state": "ARCHIVED",
  "to_state": "ACTIVE",
  "message": "Cannot resume archived session. ARCHIVED is a terminal state.",
  "session_id": "session-20251231-120000-abcd1234"
}
```

Events are emitted for observability (future: metrics, alerting).

### 8. state-mate as Policy Engine

The `state-mate` agent remains the centralized authority for mutations. Hooks serve as fast-path executors for common operations with fallback to state-mate for complex cases:

```
┌─────────────────────────────────────────────────────────────────┐
│                          Request                                 │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
              ┌───────────────────────────────┐
              │  Hook: Is this a simple case? │
              └───────────────────────────────┘
                     │              │
                   Yes              No
                     │              │
                     ▼              ▼
         ┌───────────────────┐  ┌───────────────────┐
         │ Fast-path execute │  │ Delegate to       │
         │ (validated)       │  │ state-mate agent  │
         └───────────────────┘  └───────────────────┘
```

## Consequences

### Positive

1. **Deterministic State**: Single source of truth eliminates ambiguity
2. **Formal Verification**: TLA+ model enables proof of deadlock/livelock freedom
3. **Clear Transitions**: Invalid state changes rejected at boundary
4. **Concurrency Safety**: Advisory locking prevents race conditions
5. **Auditability**: Event emission provides complete state history
6. **Extensibility**: New substates added via workflow YAML, not code changes

### Negative

1. **Breaking Changes**: Existing sessions require migration
2. **Migration Complexity**: Auto-migration must handle all field naming variants
3. **Lock Contention**: Advisory locking adds overhead for high-concurrency scenarios
4. **Verification Overhead**: TLA+ model must be maintained alongside code

### Neutral

1. **Learning Curve**: Team must understand FSM model and TLA+ specs
2. **Tooling Requirements**: Formal verification requires TLC and Alloy Analyzer

## Migration Strategy

### 1. Automatic Migration with Backup

```bash
# Migration script (run once per session)
migrate_session() {
    local session_dir="$1"
    local ctx_file="$session_dir/SESSION_CONTEXT.md"

    # Create backup
    cp "$ctx_file" "$ctx_file.v1.backup"

    # Determine canonical state from current chaos
    local new_status="ACTIVE"
    if grep -qE "^(parked_at|auto_parked_at):" "$ctx_file"; then
        new_status="PARKED"
    fi
    if grep -q "^completed_at:" "$ctx_file"; then
        new_status="ARCHIVED"
    fi

    # Rewrite with v2 schema
    # ... (details in implementation)
}
```

### 2. Schema Version Tracking

```yaml
---
schema_version: "2.0"  # Enables format detection
status: "ACTIVE"
# ...
---
```

### 3. Rollback Capability

Backup files retained for 30 days. Rollback via:
```bash
mv "$ctx_file.v1.backup" "$ctx_file"
```

## Formal Specifications

This ADR is accompanied by formal specifications:

| Spec | Location | Purpose |
|------|----------|---------|
| TLA+ FSM | `docs/specs/session-fsm.tla` | State machine model with concurrency |
| TLA+ Config | `docs/specs/session-fsm.cfg` | TLC model checking configuration |
| Alloy Permissions | `docs/specs/session-permissions.als` | Capability and role model |

These specifications are normative: implementation must satisfy all invariants.

### Verification Strategy (Hybrid Approach)

Formal verification uses a hybrid approach to balance rigor with developer experience:

| Environment | Verification Level | Tools |
|-------------|-------------------|-------|
| **Local (default)** | Lightweight syntax validation | `./scripts/verify-specs.sh` |
| **Local (optional)** | Full model checking | `./scripts/verify-specs.sh --full` (requires Java) |
| **CI (mandatory)** | Full model checking | TLC + Alloy Analyzer via GitHub Actions |

**Local Verification**:
```bash
# Quick syntax check (no external tools required)
./scripts/verify-specs.sh

# Full model checking (requires Java 17+, downloads tools automatically)
./scripts/verify-specs.sh --full
```

**CI Verification**:
- Triggered on changes to `docs/specs/*.tla` or `docs/specs/*.als`
- Runs TLC model checker with `session-fsm.cfg` configuration
- Runs Alloy Analyzer for all assertions
- Workflow: `.github/workflows/verify-formal-specs.yml`

**Expected Model Checking Results**:
- `Safety` invariant: PASS
- `LockEventuallyGranted` property: PASS
- `NoStaleReads` invariant: FAIL (intentional - demonstrates current race condition bug)

## Related Decisions

- **ADR-0005**: state-mate Centralized State Authority (referenced, not superseded)
- **Future**: ADR for event sourcing architecture
- **Future**: ADR for workflow extensibility model

## Implementation Status

**Status**: Implemented (Sprint 1-5 Complete, Sprint 6 Documentation Complete)

### Implementation Artifacts

| Artifact | Location | Lines | Description |
|----------|----------|-------|-------------|
| Core FSM | `user-hooks/lib/session-fsm.sh` | 830 | State machine, locking, validation, events |
| Migration | `user-hooks/lib/session-migrate.sh` | 681 | v1 to v2 schema migration |
| TLA+ Spec | `docs/specs/session-fsm.tla` | 401 | Formal state machine model |
| TDD | `docs/design/TDD-session-state-machine.md` | 1145 | Technical design document |
| Tests | `tests/session-fsm/*.bats` | 73 tests | BATS test suite with fixtures |
| Documentation | `docs/session-fsm/` | - | Architecture and operations guides |

### What Was Implemented

1. **Explicit Finite State Machine** (as designed)
   - Three states: ACTIVE, PARKED, ARCHIVED
   - Transition matrix enforced via `_fsm_is_valid_transition()`
   - ARCHIVED is terminal (no transitions out)

2. **Single Source of Truth** (as designed)
   - `status` field is the only state authority
   - Legacy fields removed during migration
   - v2 schema with `schema_version: "2.0"`

3. **Advisory Locking** (as designed)
   - Primary: `flock` for Linux/macOS
   - Fallback: `mkdir`-based locking (portable)
   - Stale lock detection via PID tracking

4. **Event Emission** (as designed)
   - JSONL event logs per session
   - Global audit logs for transitions and errors
   - Event types: SESSION_CREATED, SESSION_PARKED, SESSION_RESUMED, SESSION_ARCHIVED

5. **Migration Engine** (as designed)
   - Auto-migration on first access
   - Batch migration support
   - Rollback capability via `.v1.backup` files
   - Field canonicalization and event extraction

### Deviations from Original Design

| Design Element | Deviation | Rationale |
|----------------|-----------|-----------|
| JSON Schema validation | Lightweight bash validation instead of ajv-cli | Reduces external dependencies |
| flock availability | Added mkdir fallback | macOS may not have flock installed |
| Park metadata in events only | Also kept in SESSION_CONTEXT.md | Easier debugging, backward compat |

### Lessons Learned

1. **Formal Verification Value**: The TLA+ specification caught the race condition bug before implementation, proving the value of formal methods for concurrent systems.

2. **Migration Complexity**: Field naming chaos in v1 required extensive canonicalization logic. Future schema changes should include version tracking from the start.

3. **Lock Timeout Tuning**: The default 10-second timeout works well for normal operation but may need adjustment for slow systems or high-concurrency scenarios.

4. **Backward Compatibility**: Supporting both v1 and v2 sessions during migration period adds complexity but ensures smooth rollout.

### Future Work

- [ ] Full JSON Schema validation in CI (ajv-cli integration)
- [ ] Event log rotation for long-running sessions
- [ ] Metrics emission (Prometheus/StatsD)
- [ ] Workflow extensibility model (new ADR needed)

## References

- Core FSM implementation: `user-hooks/lib/session-fsm.sh`
- Migration engine: `user-hooks/lib/session-migrate.sh`
- CLI interface: `user-hooks/lib/session-manager.sh`
- Schema definition: `schemas/artifacts/session-context.schema.json`
- Workflow definition: `.claude/ACTIVE_WORKFLOW.yaml`
- Architecture documentation: `docs/session-fsm/ARCHITECTURE.md`
- Operations guide: `docs/session-fsm/OPERATIONS.md`

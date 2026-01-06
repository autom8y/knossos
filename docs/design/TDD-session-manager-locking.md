# TDD: Session Manager Lock Granularity Model

## Overview

This Technical Design Document specifies a comprehensive fix architecture for 14 bugs identified in the session management infrastructure. The design introduces a unified locking model, post-operation verification patterns, and validation consolidation to eliminate race conditions, phantom sessions, and lock scope mismatches.

## Context

| Reference | Location |
|-----------|----------|
| Gap Analysis | `docs/analysis/GAP-session-manager-concurrency.md` |
| Session FSM TDD | `docs/design/TDD-session-state-machine.md` |
| TLA+ Spec | `docs/specs/session-fsm.tla` |
| Current Implementation | `.claude/hooks/lib/session-manager.sh` |
| FSM Implementation | `.claude/hooks/lib/session-fsm.sh` |
| Core Functions | `.claude/hooks/lib/session-core.sh` |
| State Functions | `.claude/hooks/lib/session-state.sh` |
| ADR | `docs/decisions/ADR-0005-state-mate-centralized-state-authority.md` |

### Bug Summary

| Severity | Count | Bug IDs |
|----------|-------|---------|
| CRITICAL | 4 | LOCK-001, RACE-001, VALID-001, STATE-001 |
| HIGH | 5 | LOCK-002, RACE-002, RACE-003, RACE-004, STATE-002 |
| MEDIUM | 3 | LOCK-003, LOCK-004, VALID-002 |
| LOW | 2 | PARSE-001, PARSE-002 |

### Root Cause Patterns

1. **Lock Scope Mismatch**: `session-manager.sh` and `session-fsm.sh` have independent locking that does not coordinate
2. **Missing Post-Operation Verification**: Operations report success without verifying artifacts exist
3. **Validation Over-Restriction**: Session ID validation rejects legitimate IDs with non-standard formats

---

## Architecture Overview

### Current State (Problematic)

```
session-manager.sh                     session-fsm.sh
      |                                       |
      +-- mkdir .create.lock ----+            |
      |                          |            |
      +-- fsm_create_session() --+---> (no coordination)
      |                          |            |
      +-- release lock ----------+            +-- _fsm_lock_exclusive()
                                              +-- (independent lock)
                                              +-- _fsm_unlock()
```

**Problem**: Two independent locking mechanisms that do not coordinate.

### Target State (Fixed)

```
session-manager.sh
      |
      +-- Global Operation Lock (create/wrap only)
      |         |
      |         +-- Delegates to FSM with lock coordination
      |                   |
session-fsm.sh            |
      |                   v
      +-- Session-Scoped Lock (all mutations)
      |         |
      |         +-- Holds lock through entire operation
      |         |
      +-- Post-Operation Verification
      |         |
      +-- Lock Release
```

**Solution**: Hierarchical locking with clear scope boundaries.

---

## Lock Model Design

### Lock Hierarchy

The system uses a two-tier lock hierarchy:

| Lock Type | Scope | Operations | Location |
|-----------|-------|------------|----------|
| **Global Create Lock** | `.claude/sessions/.create.lock` | `cmd_create`, `cmd_wrap` | `session-manager.sh` |
| **Session Lock** | `.claude/sessions/.locks/${session_id}.lock` | All FSM operations | `session-fsm.sh` |

### Lock Acquisition Order

To prevent deadlocks, locks MUST be acquired in this order:

```
1. Global Create Lock (if needed)
   |
   v
2. Session-Scoped Lock (always for mutations)
```

**Rationale**: Global lock is only needed for operations that create/destroy sessions. Session-scoped locks are sufficient for all other operations.

### Lock Acquisition Protocol

```bash
# CORRECT: Create operation acquires global lock, then session lock
cmd_create() {
    acquire_global_lock || fail
    trap 'release_global_lock' EXIT

    session_id=$(generate_session_id)

    # FSM operation INHERITS global lock scope
    # FSM acquires session lock internally
    fsm_create_session "$initiative" "$complexity" "$team"

    # Verify BEFORE releasing global lock
    verify_session_created "$session_id" || {
        cleanup_partial_session "$session_id"
        fail
    }
}

# CORRECT: Park operation uses only session lock
mutate_park_fsm() {
    # No global lock needed - session already exists
    # FSM handles session-scoped locking internally
    fsm_transition "$session_id" "PARKED" "$metadata"
}
```

### Lock Release Guarantees (RAII Pattern for Bash)

The current trap quoting bug (LOCK-002) demonstrates the need for reliable lock cleanup. The fix uses a trap factory pattern:

```bash
# INCORRECT (current - LOCK-002)
trap "rm -rf '$lockfile'" EXIT  # Single quotes prevent expansion

# CORRECT (fixed)
_setup_lock_cleanup() {
    local lockfile="$1"
    # Double-quote the entire trap string; escape inner quotes
    trap "rm -rf \"$lockfile\"" EXIT INT TERM
}

# Alternative: Indirect variable reference
_setup_lock_cleanup_indirect() {
    local lockfile="$1"
    _CLEANUP_LOCKFILE="$lockfile"
    trap '_cleanup_lock "$_CLEANUP_LOCKFILE"' EXIT INT TERM
}

_cleanup_lock() {
    rm -rf "$1" 2>/dev/null || true
}
```

### Fallback Strategy for Non-flock Systems

The current implementation falls back to `mkdir`-based locking when `flock` is unavailable. This fallback has a limitation: shared locks are treated as exclusive (LOCK-004).

**Design Decision**: Accept this conservative behavior. Document it.

**Rationale**:
- Correctness is preserved (no data corruption)
- Performance impact is limited (most systems have flock)
- Implementing shared mkdir locking adds significant complexity

---

## Fix Specifications

### LOCK-001: Create Lock Scope Mismatch

**Severity**: CRITICAL
**Component**: `session-manager.sh:295-308`
**Root Cause**: `cmd_create()` acquires a lock but FSM creates session directory without coordination.

**Current Code**:
```bash
while ! mkdir "$lockfile" 2>/dev/null; do
    # ...wait loop...
done
trap "rm -rf '$lockfile'" EXIT
# FSM call happens OUTSIDE effective lock scope
session_id=$(fsm_create_session "$initiative" "$complexity" "$team")
```

**Fix Design**:

1. Move lock acquisition to use the new `_setup_lock_cleanup()` pattern
2. Verify session creation BEFORE releasing lock
3. FSM already has internal locking; let it handle session-scoped operations

**Implementation**:
```bash
cmd_create() {
    local initiative="${1:-unnamed}"
    local complexity="${2:-MODULE}"
    local team="${3:-$(cat ".claude/ACTIVE_RITE" 2>/dev/null || echo "none")}"

    local lockfile="$SESSIONS_DIR/.create.lock"
    local lock_timeout=10
    local waited=0
    mkdir -p "$SESSIONS_DIR" 2>/dev/null

    # Acquire global create lock
    while ! mkdir "$lockfile" 2>/dev/null; do
        if [ $waited -ge $lock_timeout ]; then
            echo '{"success": false, "error": "Timeout waiting for session lock"}' >&2
            exit 1
        fi
        sleep 1
        ((waited++)) || true  # Prevent errexit on zero increment
    done

    # FIXED: Use proper trap quoting
    trap "rm -rf \"$lockfile\"" EXIT INT TERM

    # ... validation ...

    # Create via FSM (FSM handles session-scoped locking)
    local session_id
    session_id=$(fsm_create_session "$initiative" "$complexity" "$team")

    if [[ -z "$session_id" || "$session_id" == *"error"* ]]; then
        echo '{"success": false, "error": "Failed to create session via FSM"}' >&2
        exit 1
    fi

    local session_dir="$SESSIONS_DIR/$session_id"

    # FIXED: Post-creation verification (addresses RACE-001)
    if [[ ! -d "$session_dir" ]] || [[ ! -f "$session_dir/SESSION_CONTEXT.md" ]]; then
        echo '{"success": false, "error": "Session creation failed: directory or context file missing"}' >&2
        exit 1
    fi

    # ... rest of function ...
}
```

**Files Changed**:
- `.claude/hooks/lib/session-manager.sh`: Lines 295-366

**Backward Compatibility**: COMPATIBLE - No external interface changes.

---

### LOCK-002: Trap Single Quote Escaping Bug

**Severity**: HIGH
**Component**: `session-manager.sh:308`
**Root Cause**: Single quotes inside double quotes prevent variable expansion.

**Current Code**:
```bash
trap "rm -rf '$lockfile'" EXIT
```

**Fix Design**: Single-line fix with proper quoting.

**Implementation**:
```bash
trap "rm -rf \"$lockfile\"" EXIT INT TERM
```

**Files Changed**:
- `.claude/hooks/lib/session-manager.sh`: Line 308

**Backward Compatibility**: COMPATIBLE - Bug fix only.

---

### LOCK-003: FD Number Collision Risk

**Severity**: MEDIUM
**Component**: `session-fsm.sh:84, 121`
**Root Cause**: FD calculation uses `cksum % 50`, allowing only 50 unique FDs.

**Current Code**:
```bash
fd=$((200 + $(echo "$session_id" | cksum | cut -d' ' -f1) % 50))
```

**Fix Design**: Use deterministic per-session FD allocation with larger range.

**Implementation**:
```bash
# Expand FD range from 50 to 200 (FDs 200-399)
_fsm_get_fd() {
    local session_id="$1"
    local hash
    hash=$(echo "$session_id" | cksum | cut -d' ' -f1)
    echo $((200 + hash % 200))
}
```

**Alternative Design** (Recommended): Use named FD with automatic allocation.

```bash
_fsm_lock_exclusive() {
    local session_id="$1"
    local lock_file
    lock_file=$(_fsm_lock_file "$session_id")
    local timeout="${FSM_LOCK_TIMEOUT:-10}"

    mkdir -p "$(dirname "$lock_file")" 2>/dev/null || return 1

    if command -v flock >/dev/null 2>&1; then
        # Use exec with automatic FD allocation via /dev/fd
        if exec {fd}>"$lock_file"; then
            if flock -x -w "$timeout" "$fd" 2>/dev/null; then
                echo "$$" >&"$fd"
                _FSM_LOCK_FDS["$session_id"]="$fd"
                return 0
            else
                exec {fd}>&- 2>/dev/null || true
                return 1
            fi
        fi
        return 1
    else
        # Fallback: mkdir-based locking (unchanged)
        ...
    fi
}
```

**Trade-off Analysis**:
- `{fd}` syntax requires bash 4.1+ (available on all target platforms)
- Eliminates collision risk entirely
- Slightly more complex error handling

**Files Changed**:
- `.claude/hooks/lib/session-fsm.sh`: Lines 84-87, 121-124

**Backward Compatibility**: COMPATIBLE - Internal implementation change.

---

### RACE-001: Phantom Session Creation

**Severity**: CRITICAL
**Component**: `session-manager.sh:332-366`
**Root Cause**: `cmd_create()` reports success without verifying artifacts.

**Current Code**:
```bash
session_id=$(fsm_create_session "$initiative" "$complexity" "$team")
if [[ -z "$session_id" || "$session_id" == *"error"* ]]; then
    # failure path
fi
# SUCCESS PATH - no directory verification!
```

**Fix Design**: Add post-creation verification before reporting success.

**Implementation**:
```bash
session_id=$(fsm_create_session "$initiative" "$complexity" "$team")

# Check return value
if [[ -z "$session_id" || "$session_id" == *"error"* ]]; then
    echo '{"success": false, "error": "Failed to create session via FSM"}' >&2
    exit 1
fi

local session_dir="$SESSIONS_DIR/$session_id"

# FIXED: Post-creation verification
_verify_session_artifacts() {
    local dir="$1"
    [[ -d "$dir" ]] || return 1
    [[ -f "$dir/SESSION_CONTEXT.md" ]] || return 1
    # Validate frontmatter has required fields
    grep -q "^session_id:" "$dir/SESSION_CONTEXT.md" 2>/dev/null || return 1
    grep -q "^status:" "$dir/SESSION_CONTEXT.md" 2>/dev/null || return 1
    return 0
}

if ! _verify_session_artifacts "$session_dir"; then
    # Cleanup partial creation
    rm -rf "$session_dir" 2>/dev/null
    echo '{"success": false, "error": "Session creation incomplete: missing artifacts"}' >&2
    exit 1
fi
```

**Files Changed**:
- `.claude/hooks/lib/session-manager.sh`: Lines 332-366

**Backward Compatibility**: COMPATIBLE - Adds verification, no interface change.

---

### STATE-001: .current-session Directory Creation Race

**Severity**: CRITICAL
**Component**: `session-core.sh:86-94`
**Root Cause**: `atomic_write()` does not check if destination is a directory.

**Current Code**:
```bash
if ! atomic_write "$current_file" "$session_id"; then
    echo "Error: Failed to write current session file" >&2
    return 1
fi
```

**Problem**: If `.current-session` is a directory (from race condition), `mv` moves the temp file INTO the directory instead of replacing it.

**Fix Design**: Check destination type and handle directory case.

**Implementation**:
```bash
set_current_session() {
    local session_id="$1"

    if [ -z "$session_id" ]; then
        echo "Error: session_id required" >&2
        return 1
    fi

    # Validate session ID format (relaxed - see VALID-001)
    if [[ ! "$session_id" =~ ^session- ]]; then
        echo "Error: Invalid session_id format: $session_id" >&2
        return 1
    fi

    local project_dir="${CLAUDE_PROJECT_DIR:-.}"
    local sessions_dir="$project_dir/.claude/sessions"
    local current_file="$sessions_dir/.current-session"

    mkdir -p "$sessions_dir" 2>/dev/null || {
        echo "Error: Cannot create sessions directory" >&2
        return 1
    }

    # FIXED: Handle case where .current-session is a directory
    if [[ -d "$current_file" ]]; then
        # This is an error state - remove the directory
        rm -rf "$current_file" 2>/dev/null || {
            echo "Error: .current-session is a directory and cannot be removed" >&2
            return 1
        }
    fi

    # FIXED: Use atomic file replacement
    # Create temp file and use mv -f to replace even if target is a symlink
    local temp_file
    temp_file=$(mktemp "$sessions_dir/.current-session.XXXXXX") || {
        echo "Error: Cannot create temp file" >&2
        return 1
    }

    printf '%s' "$session_id" > "$temp_file" || {
        rm -f "$temp_file"
        echo "Error: Cannot write to temp file" >&2
        return 1
    }

    # Use mv -f to force replacement
    mv -f "$temp_file" "$current_file" || {
        rm -f "$temp_file"
        echo "Error: Failed to write current session file" >&2
        return 1
    }

    return 0
}
```

**Files Changed**:
- `.claude/hooks/lib/session-core.sh`: Lines 69-100

**Backward Compatibility**: COMPATIBLE - Handles edge case, no interface change.

---

### VALID-001: Session ID Format Rejection

**Severity**: CRITICAL
**Component**: `session-core.sh:78, 125`
**Root Cause**: Regex `^session-[0-9]{8}-[0-9]{6}-[a-f0-9]{8}$` is too strict.

**Current Code**:
```bash
if [[ ! "$session_id" =~ ^session-[0-9]{8}-[0-9]{6}-[a-f0-9]{8}$ ]]; then
    echo "Error: Invalid session_id format: $session_id" >&2
    return 1
fi
```

**Problem**: Rejects valid sessions like `session-meta-locking-20260104-022313-9310167d`.

**Fix Design**: Relax validation to accept any `session-*` format, or use directory existence as validation.

**Option A: Relaxed Pattern** (Recommended)
```bash
# Accept: session-<anything>
if [[ ! "$session_id" =~ ^session-.+ ]]; then
    echo "Error: Invalid session_id format: $session_id" >&2
    return 1
fi
```

**Option B: Directory Existence**
```bash
# Validate by checking directory exists
local session_dir="$project_dir/.claude/sessions/$session_id"
if [[ ! -d "$session_dir" ]]; then
    echo "Error: Session not found: $session_id" >&2
    return 1
fi
```

**Trade-off Analysis**:
- Option A: Allows creation of new sessions with custom names
- Option B: Requires session to exist first (prevents set_current_session for new sessions)

**Decision**: Use Option A for `set_current_session()`, but add directory check in `get_current_session()` (already exists at line 133-138).

**Implementation**:
```bash
# session-core.sh: set_current_session()
if [[ ! "$session_id" =~ ^session-.+ ]]; then
    echo "Error: Invalid session_id format: $session_id" >&2
    return 1
fi

# session-core.sh: get_current_session() - existing check is sufficient
# Line 125: Keep strict validation for reading (filters old corrupt data)
# OR relax to match set_current_session
if [[ ! "$session_id" =~ ^session-.+ ]]; then
    rm -f "$current_file" 2>/dev/null
    echo ""
    return 0
fi

# session-state.sh: validate_session_id_format() - update to match
validate_session_id_format() {
    local session_id="$1"
    [[ "$session_id" =~ ^session-.+ ]]
}
```

**Files Changed**:
- `.claude/hooks/lib/session-core.sh`: Lines 78, 125
- `.claude/hooks/lib/session-state.sh`: Line 167

**Backward Compatibility**: COMPATIBLE - Relaxing validation accepts more inputs.

---

### STATE-002: Inconsistent State Sources

**Severity**: HIGH
**Component**: `session-state.sh` vs `session-fsm.sh`
**Root Cause**: Two functions with different logic for determining state.

**Current State**:

| Function | Location | Logic |
|----------|----------|-------|
| `get_session_state()` | session-state.sh:21 | Infers from `parked_at` field presence |
| `fsm_get_state()` | session-fsm.sh:495 | Reads `status` field directly |

**Fix Design**: Consolidate to single source of truth using FSM as authoritative.

**Implementation**:
```bash
# session-state.sh: Delegate to FSM

# Get current session state
# Returns: ACTIVE, PARKED, ARCHIVED, or NONE
# Canonical implementation - uses FSM as single source of truth
get_session_state() {
    local session_id="${1:-$(get_session_id)}"

    if [ -z "$session_id" ]; then
        echo "NONE"
        return 0
    fi

    # Delegate to FSM (authoritative)
    # FSM handles both v1 (inference) and v2 (status field) sessions
    if declare -F fsm_get_state >/dev/null 2>&1; then
        fsm_get_state "$session_id"
    else
        # Fallback if FSM not sourced (should not happen in production)
        _legacy_get_session_state "$session_id"
    fi
}

# Legacy implementation for backward compatibility
_legacy_get_session_state() {
    local session_id="$1"
    local project_dir="${CLAUDE_PROJECT_DIR:-.}"
    local session_file="$project_dir/.claude/sessions/$session_id/SESSION_CONTEXT.md"

    if [ ! -f "$session_file" ]; then
        echo "NONE"
        return 0
    fi

    if grep -q "^auto_parked_at:" "$session_file" 2>/dev/null; then
        echo "AUTO_PARKED"
    elif grep -q "^parked_at:" "$session_file" 2>/dev/null; then
        echo "PARKED"
    else
        echo "ACTIVE"
    fi
}
```

**Note**: FSM's `fsm_get_state()` already handles v1 fallback (lines 522-531), so delegation is safe.

**Files Changed**:
- `.claude/hooks/lib/session-state.sh`: Lines 21-52

**Backward Compatibility**: COMPATIBLE - Same return values, different implementation.

---

### VALID-002: Session Context Field Mismatch

**Severity**: MEDIUM
**Component**: `session-state.sh:143` vs `session-fsm.sh:215`
**Root Cause**: Different required field lists for validation.

**Current State**:

| Function | Required Fields |
|----------|-----------------|
| `validate_session_context()` | session_id, created_at, initiative, complexity, active_team, current_phase |
| `_fsm_validate_context()` | schema_version, session_id, status, created_at, initiative, complexity, active_team, current_phase |

**Fix Design**: Consolidate to single validation function in FSM.

**Implementation**:
```bash
# session-state.sh: Delegate validation to FSM

validate_session_context() {
    local file="$1"

    [ -f "$file" ] || { echo "File not found: $file" >&2; return 1; }

    # Check schema version to determine validation path
    local version
    version=$(grep -m1 "^schema_version:" "$file" 2>/dev/null | cut -d: -f2- | tr -d ' "')

    if [[ "$version" == "2.0" || "$version" == "2.1" ]]; then
        # v2: Use FSM validation (stricter)
        if declare -F _fsm_validate_context >/dev/null 2>&1; then
            _fsm_validate_context "$file"
            return $?
        fi
    fi

    # v1 or FSM not available: Basic validation
    local required_fields=("session_id" "created_at" "initiative" "complexity" "active_team" "current_phase")
    local missing=()

    for field in "${required_fields[@]}"; do
        if ! grep -q "^$field:" "$file" 2>/dev/null; then
            missing+=("$field")
        fi
    done

    if [ ${#missing[@]} -gt 0 ]; then
        echo "Missing required fields: ${missing[*]}" >&2
        return 1
    fi

    return 0
}
```

**Files Changed**:
- `.claude/hooks/lib/session-state.sh`: Lines 141-160

**Backward Compatibility**: COMPATIBLE - v1 sessions use original validation.

---

### RACE-002: Set Current Session TOCTOU

**Severity**: HIGH
**Component**: `session-core.sh:69-100`
**Root Cause**: Gap between validation and write allows race.

**Fix Design**: Use atomic file operations with pre-write check.

**Implementation**: Already addressed by STATE-001 fix (directory check + atomic mv).

---

### RACE-003: FSM Transition Backup File Collision

**Severity**: MEDIUM
**Component**: `session-fsm.sh:600`
**Root Cause**: Static backup name without PID suffix.

**Current Code**:
```bash
local backup_file="${ctx_file}.backup"
```

**Fix Design**: Add PID suffix consistent with `_fsm_safe_mutate()`.

**Implementation**:
```bash
# session-fsm.sh: fsm_transition() at line 600
local backup_file="${ctx_file}.backup.$$"
```

**Files Changed**:
- `.claude/hooks/lib/session-fsm.sh`: Line 600

**Backward Compatibility**: COMPATIBLE - Internal change.

---

### RACE-004: Lock Release Before Cleanup

**Severity**: HIGH (Re-evaluated as addressed)
**Component**: `session-fsm.sh:636-637`

**Analysis**: Re-examination shows the order is correct (cleanup before unlock). The real issue is backup collision (RACE-003).

**Status**: Addressed by RACE-003 fix.

---

### LOCK-004: Shared Lock Fallback to Exclusive

**Severity**: LOW
**Component**: `session-fsm.sh:97-100`
**Root Cause**: `mkdir`-based locking cannot support shared mode.

**Fix Design**: Document behavior, no code change.

**Documentation**:
```bash
# session-fsm.sh: _fsm_lock_shared()
# DESIGN NOTE: On systems without flock, shared locks fall back to exclusive.
# This is conservative-correct (no data corruption) but may serialize reads.
# Impact: Performance degradation on non-flock systems.
# Justification: Implementing shared mkdir locking adds significant complexity
# for minimal benefit (most systems have flock).
```

**Files Changed**: None (documentation only in existing comments)

**Backward Compatibility**: N/A

---

### PARSE-001: Main Dispatch Positional Argument Leak

**Severity**: LOW
**Component**: `session-manager.sh:793-800`
**Root Cause**: Flags after command interpreted as positional args.

**Current Code**:
```bash
case "${1:-help}" in
    create)     cmd_create "${2:-}" "${3:-}" "${4:-}" ;;
```

**Fix Design**: Parse global flags before dispatch.

**Implementation**:
```bash
# Parse global flags
DRY_RUN=false
VERBOSE=false

while [[ "${1:-}" == --* ]]; do
    case "$1" in
        --dry-run) DRY_RUN=true; shift ;;
        --verbose) VERBOSE=true; shift ;;
        --) shift; break ;;
        *) echo "Unknown flag: $1" >&2; exit 1 ;;
    esac
done

case "${1:-help}" in
    create)     cmd_create "${2:-}" "${3:-}" "${4:-}" ;;
    ...
```

**Files Changed**:
- `.claude/hooks/lib/session-manager.sh`: Lines 792-807

**Backward Compatibility**: COMPATIBLE - New flags, existing behavior unchanged.

---

### PARSE-002: Mutate Subcommand Shift Bug

**Severity**: MEDIUM (Re-evaluated as LOW)
**Component**: `session-manager.sh:517-519`

**Analysis**: The `shift || true` pattern is correct for preventing errexit. The real issue is flag parsing in handlers.

**Fix Design**: Add flag parsing to mutate handlers.

**Implementation**:
```bash
mutate_park_fsm() {
    local session_id="$1"
    shift

    local reason="Manual park"

    while [[ "${1:-}" == --* ]]; do
        case "$1" in
            --reason) reason="$2"; shift 2 ;;
            --reason=*) reason="${1#--reason=}"; shift ;;
            *) shift ;;  # Skip unknown flags
        esac
    done

    # Use positional arg as reason if provided and no flag
    [[ -n "${1:-}" ]] && reason="$1"

    # ... rest of function
}
```

**Files Changed**:
- `.claude/hooks/lib/session-manager.sh`: Lines 591-634

**Backward Compatibility**: COMPATIBLE - New flags, existing usage unchanged.

---

## Implementation Plan

### Phase 1: Quick Wins (Day 1)

**Single-line or minimal fixes with immediate impact.**

| Bug | Fix | LOC |
|-----|-----|-----|
| LOCK-002 | Fix trap quoting | 1 |
| RACE-003 | Add PID suffix to backup file | 1 |
| VALID-001 | Relax session ID regex | 3 |

**Verification**:
```bash
# LOCK-002: Verify trap expansion
bash -c 'lockfile="/tmp/test lock"; trap "rm -rf \"$lockfile\"" EXIT; mkdir "$lockfile"; exit 0'
ls "/tmp/test lock"  # Should not exist

# VALID-001: Verify relaxed validation
source .claude/hooks/lib/session-core.sh
set_current_session "session-meta-locking-20260104-022313-9310167d"
echo $?  # Should be 0
```

### Phase 2: Lock Consolidation (Day 2-3)

**Coordinate locking between session-manager.sh and session-fsm.sh.**

| Bug | Fix | Complexity |
|-----|-----|------------|
| LOCK-001 | Restructure cmd_create lock scope | Medium |
| LOCK-003 | Use automatic FD allocation | Low |
| STATE-001 | Handle .current-session directory case | Low |

**Verification**:
```bash
# LOCK-001: Parallel create test
for i in {1..5}; do
    .claude/hooks/lib/session-manager.sh create "test-$i" MODULE &
done
wait
# Verify: At most 1 succeeds, no orphan locks
ls .claude/sessions/.create.lock 2>/dev/null && echo "FAIL: orphan lock"
```

### Phase 3: Verification Layer (Day 4-5)

**Add post-operation verification.**

| Bug | Fix | Complexity |
|-----|-----|------------|
| RACE-001 | Add post-creation verification | Medium |
| RACE-002 | Atomic file operations | Low (addressed by STATE-001) |
| STATE-002 | Consolidate state queries | Medium |
| VALID-002 | Unify validation | Low |

**Verification**:
```bash
# RACE-001: Interrupt create mid-operation
timeout 0.1 .claude/hooks/lib/session-manager.sh create "test" MODULE &
sleep 0.5
# Verify: No partial session directories
```

### Phase 4: Polish (Day 6)

**Argument parsing and documentation.**

| Bug | Fix | Complexity |
|-----|-----|------------|
| PARSE-001 | Global flag parsing | Low |
| PARSE-002 | Handler flag parsing | Low |
| LOCK-004 | Document fallback behavior | Documentation |

---

## Test Matrix

### Concurrency Test Scenarios

| Test ID | Scenario | Expected Outcome | TLA+ Property |
|---------|----------|------------------|---------------|
| CONC-001 | 5 parallel creates | 1 succeeds, 4 fail with "already exists" | MutualExclusion |
| CONC-002 | Create during park | Both complete without corruption | MutualExclusion |
| CONC-003 | 10 parallel status reads | All return consistent state | LockedReadsAreConsistent |
| CONC-004 | Lock timeout scenario | Error returned, no hang | NoDeadlock |
| CONC-005 | Stale lock cleanup | Old lock removed, new operation succeeds | HolderNotInQueue |

### Regression Test Cases

| Test ID | Bug | Test Command | Success Condition |
|---------|-----|--------------|-------------------|
| REG-001 | LOCK-002 | Trap with spaces in path | Lock cleaned up |
| REG-002 | VALID-001 | Set non-standard session ID | Returns 0 |
| REG-003 | STATE-001 | Create dir at .current-session | Handled, file created |
| REG-004 | RACE-001 | Create with immediate interrupt | No orphan directories |
| REG-005 | RACE-003 | Parallel transitions | No backup collision |

### Performance Benchmarks

| Metric | Baseline | Target | Method |
|--------|----------|--------|--------|
| Create latency | ~100ms | <150ms | time session-manager.sh create |
| Lock acquisition | ~10ms | <20ms | time with flock timing |
| Status query | ~50ms | <75ms | time session-manager.sh status |
| Parallel creates (5) | N/A | <2s total | time for all to complete |

---

## Migration Considerations

### Breaking Changes

**None**. All fixes are internal implementation changes or relaxations of validation.

### Backward Compatibility Requirements

| Requirement | How Addressed |
|-------------|---------------|
| v1 sessions continue to work | FSM has v1 fallback (lines 522-531) |
| Existing scripts work unchanged | No API changes |
| Older session IDs remain valid | VALID-001 relaxes validation |

### Rollback Strategy

```bash
# If issues discovered, rollback specific files:
git checkout HEAD~1 -- .claude/hooks/lib/session-manager.sh
git checkout HEAD~1 -- .claude/hooks/lib/session-fsm.sh
git checkout HEAD~1 -- .claude/hooks/lib/session-core.sh
git checkout HEAD~1 -- .claude/hooks/lib/session-state.sh
```

---

## Integration Test Specifications

### Satellite Diversity Testing

| Satellite Type | Characteristics | Test Focus |
|----------------|-----------------|------------|
| **baseline** | Standard roster setup | Regression |
| **minimal** | No local settings | Basic operation |
| **complex** | Nested arrays, custom hooks | Edge cases |

### Test Execution Matrix

| Satellite | CONC-001 | CONC-002 | REG-001-005 | Expected |
|-----------|----------|----------|-------------|----------|
| baseline | Pass | Pass | Pass | Baseline regression |
| minimal | Pass | Pass | Pass | No special config needed |
| complex | Pass | Pass | Pass | Complex configs handled |

### CI Integration

```yaml
# .github/workflows/session-tests.yml
session-concurrency:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Install bats
      run: npm install -g bats
    - name: Run concurrency tests
      run: bats tests/integration/session-concurrency.bats
    - name: Run regression tests
      run: bats tests/unit/session-fixes.bats
```

---

## Summary

This TDD specifies fixes for 14 bugs in the session management infrastructure:

| Phase | Bugs Fixed | Effort |
|-------|------------|--------|
| Phase 1 (Quick Wins) | LOCK-002, RACE-003, VALID-001 | 1 day |
| Phase 2 (Lock Consolidation) | LOCK-001, LOCK-003, STATE-001 | 2 days |
| Phase 3 (Verification) | RACE-001, RACE-002, STATE-002, VALID-002 | 2 days |
| Phase 4 (Polish) | PARSE-001, PARSE-002, LOCK-004 | 1 day |

**Total Estimated Effort**: 6 days

**Key Design Decisions**:
1. Two-tier lock hierarchy (global create + session-scoped)
2. FSM as single source of truth for state
3. Relaxed session ID validation (accept `session-*`)
4. Post-operation verification pattern
5. Automatic FD allocation for flock

**No Breaking Changes**: All fixes are backward compatible.

---

## Artifact Attestation

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-session-manager-locking.md` | Created |
| Gap Analysis | `/Users/tomtenuta/Code/roster/docs/analysis/GAP-session-manager-concurrency.md` | Read |
| Session Manager | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-manager.sh` | Read |
| Session FSM | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-fsm.sh` | Read |
| Session Core | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-core.sh` | Read |
| Session State | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-state.sh` | Read |
| TLA+ Spec | `/Users/tomtenuta/Code/roster/docs/specs/session-fsm.tla` | Read |
| Existing TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-session-state-machine.md` | Read |
| ADR-0005 | `/Users/tomtenuta/Code/roster/docs/decisions/ADR-0005-state-mate-centralized-state-authority.md` | Read |

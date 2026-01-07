# TDD: Session Manager Ecosystem Audit

## Overview

This Technical Design Document expands the session management locking architecture (TDD-session-manager-locking.md) to include the full ecosystem: hook system coordination, state-mate agent contract, v1-v2 migration, and comprehensive test coverage. It provides Integration Engineer with a complete, implementation-ready specification.

## Context

| Reference | Location |
|-----------|----------|
| Session Locking TDD | `docs/design/TDD-session-manager-locking.md` |
| Gap Analysis | `docs/analysis/GAP-session-manager-concurrency.md` |
| Session FSM TDD | `docs/design/TDD-session-state-machine.md` |
| TLA+ Spec | `docs/specs/session-fsm.tla` |
| state-mate ADR | `docs/decisions/ADR-0005-state-mate-centralized-state-authority.md` |
| Hook Settings | `.claude/settings.local.json` |
| Session Manager | `.claude/hooks/lib/session-manager.sh` |
| Session FSM | `.claude/hooks/lib/session-fsm.sh` |

### Scope

**IN SCOPE:**
- Hook system locking and coordination (11 hook scripts)
- state-mate agent coordination contract
- v1-v2 schema migration algorithm with rollback
- Extended test matrix including hook ordering and concurrent access
- Risk boundaries and protected subsystems

**OUT OF SCOPE:**
- CEM infrastructure (roster-sync) - separate initiative
- state-mate agent implementation - coordinate only
- Global commands refactoring - audit only, commands are API consumers
- Team pack structure changes

---

## Part 1: Hook System Architecture

### 1.1 Current Hook Inventory

| Event | Hook Script | Purpose | Mutates Session? |
|-------|-------------|---------|-----------------|
| **SessionStart** | `session-context.sh` | Injects SESSION_ID, ACTIVE_RITE context | No (read-only) |
| **SessionStart** | `orchestrated-mode.sh` | Detects orchestrator presence | No (read-only) |
| **Stop** | `auto-park.sh` | Auto-parks session on stop | **Yes** (writes parked_at) |
| **PreToolUse/Bash** | `command-validator.sh` | Validates bash commands, rites | No |
| **PreToolUse/Edit,Write** | `session-write-guard.sh` | Blocks direct writes to *_CONTEXT.md | No (blocker) |
| **PreToolUse/Edit,Write** | `delegation-check.sh` | Warns on orchestrator bypass | No (warning) |
| **PreToolUse/Task** | `orchestrator-bypass-check.sh` | Warns on missing orchestrator consultation | No (warning) |
| **PostToolUse/Write** | `artifact-tracker.sh` | Tracks PRD/TDD/ADR artifacts | **Yes** (appends to SESSION_CONTEXT) |
| **PostToolUse/Write** | `session-audit.sh` | Logs session mutations | No (audit log only) |
| **PostToolUse/Bash** | `commit-tracker.sh` | Logs git commits | **Yes** (appends to SESSION_CONTEXT) |
| **UserPromptSubmit** | `orchestrator-router.sh` | Routes /start,/sprint,/task to orchestrator | No |
| **UserPromptSubmit** | `start-preflight.sh` | Pre-checks for /start command | **Yes** (creates session) |

### 1.2 Hook Execution Model

**Current Behavior:**
```
Event Trigger
    |
    v
settings.local.json lookup
    |
    v
Array iteration (index 0, 1, 2...)
    |
    v
For each hook in array:
    - Fork process
    - Execute with timeout
    - Capture stdout/stderr
    - If exit != 0 && type == "block": stop tool execution
```

**Key Observation:** Hooks are executed **sequentially** within an event type, but **no explicit locking** exists between hooks. Multiple events (e.g., PostToolUse firing for two rapid Write operations) could execute concurrently.

### 1.3 Hook Locking Design Decision

**Question:** Should hooks use the same session-scoped locking primitives as session-manager.sh?

**Decision:** **Preserve implicit ordering with documentation; do NOT add locking to hooks.**

**Rationale:**
1. **Session manager provides serialization**: Hooks that mutate session state call into `session-manager.sh` or `session-fsm.sh`, which already provide locking.
2. **Hook timeout acts as implicit serialization**: Each hook has a timeout (3-10s). Sequential execution within an event type prevents parallel execution of same-type hooks.
3. **Complexity vs. benefit**: Adding flock to 11 hooks increases complexity and maintenance burden for minimal gain.
4. **Risk isolation**: Hooks are categorized as DEFENSIVE (never crash) or RECOVERABLE (degrade gracefully). This classification already handles failures.

**However:** Two hooks that mutate SESSION_CONTEXT.md (`artifact-tracker.sh`, `commit-tracker.sh`) should use atomic operations:

```bash
# CURRENT (artifact-tracker.sh:69-84) - uses sed -i
sed -i.bak "s|^- PRD:.*|- PRD: $FILE_PATH|" "$SESSION_CONTEXT"

# REQUIRED: Use atomic_write pattern
# 1. Read current content
# 2. Apply transformation
# 3. Write via atomic_write() from primitives.sh
```

### 1.4 Hook Execution Order Specification

**Preserve current implicit ordering.** Document dependencies:

| Event | Order | Dependencies |
|-------|-------|--------------|
| **SessionStart** | 1. session-context.sh 2. orchestrated-mode.sh | orchestrated-mode depends on session-context having run |
| **Stop** | 1. auto-park.sh | None (single hook) |
| **PreToolUse/Edit,Write** | 1. session-write-guard.sh 2. delegation-check.sh | session-write-guard must run first (blocking); delegation-check is warning-only |
| **PostToolUse/Write** | 1. artifact-tracker.sh 2. session-audit.sh | Independent, but audit should run second to capture final state |

**Do NOT introduce explicit priority numbering.** The current array-order approach is sufficient and matches Claude Code's hook execution model.

### 1.5 Hook-to-FSM Coordination

**Problem:** `auto-park.sh` directly writes to SESSION_CONTEXT.md without using FSM:

```bash
# auto-park.sh:43-49 - Direct awk mutation
UPDATED_CONTENT=$(awk -v ts="$TIMESTAMP" '...' "$SESSION_FILE")
atomic_write "$SESSION_FILE" "$UPDATED_CONTENT"
```

**Solution:** Refactor `auto-park.sh` to use FSM transition:

```bash
# auto-park.sh (FIXED)
# Instead of direct mutation, use FSM
local session_id
session_id=$(basename "$SESSION_DIR")

# Attempt FSM transition to PARKED state
local result
result=$(fsm_transition "$session_id" "PARKED" '{"reason":"auto-park","auto":true}')

if [[ "$result" == *'"success": true'* ]]; then
    echo '{"systemMessage": "Session auto-parked."}'
else
    # FSM failed (maybe already parked) - log and continue
    echo '{"systemMessage": "Session auto-park skipped (FSM transition failed)."}'
fi
```

**Files Changed:**
- `.claude/hooks/session-guards/auto-park.sh`: Lines 39-55

**Backward Compatibility:** COMPATIBLE - FSM handles v1 sessions via fallback logic.

---

## Part 2: state-mate Coordination Contract

### 2.1 Current Architecture

Per ADR-0005, state-mate is the sole authority for `*_CONTEXT.md` mutations. The enforcement flow is:

```
Write/Edit Tool
    |
    v
PreToolUse Hook (session-write-guard.sh)
    |
    +-- Match: *_CONTEXT.md? ---> BLOCK with guidance
    |
    +-- No match: ALLOW
```

When state-mate is invoked:
```
Task(moirai, "mutation request")
    |
    v
state-mate agent reads SESSION_CONTEXT.md
    |
    v
state-mate validates operation via FSM
    |
    v
state-mate executes Write/Edit
    |
    v
session-write-guard intercepts ---> ???
```

**Problem:** The current design creates a circular block. state-mate cannot write because the hook blocks all writes.

### 2.2 state-mate Bypass Mechanism

**Design Decision:** state-mate bypasses the write guard via an environment marker.

**Implementation:**

```bash
# session-write-guard.sh (UPDATED)
# Check for state-mate bypass marker
if [[ "${STATE_MATE_BYPASS:-}" == "true" ]]; then
    exit 0  # Allow write
fi
```

state-mate sets this environment variable before writing:

```markdown
<!-- state-mate.md agent prompt -->
## Write Protocol

When writing to *_CONTEXT.md files, set the bypass marker:

1. Validate operation via FSM transition check
2. Set environment: `STATE_MATE_BYPASS=true`
3. Execute Write tool
4. Environment automatically cleared after tool execution
```

**Alternative Considered:** Check agent name in hook. Rejected because:
- Hook doesn't have reliable access to calling agent identity
- Environment variable is simpler and explicit

### 2.3 state-mate Locking Contract

**Question:** Should state-mate use new locking primitives?

**Answer:** **No new locking required.** state-mate already provides serialization:

1. **Task tool serialization**: When state-mate is invoked via Task tool, Claude Code serializes agent execution. Only one state-mate invocation runs at a time within a session.

2. **FSM internal locking**: When state-mate calls FSM functions (directly or via session-manager.sh), FSM acquires session-scoped locks:
   ```bash
   fsm_transition() {
       _fsm_lock_exclusive "$session_id"  # Already implemented
       ...
       _fsm_unlock "$session_id"
   }
   ```

3. **Audit trail guarantees**: state-mate logs all mutations to `.claude/sessions/.audit/session-mutations.log`. This audit trail is append-only (no concurrent write corruption risk).

### 2.4 state-mate Operations and FSM Mapping

| state-mate Operation | FSM Function | Lock Required |
|---------------------|--------------|---------------|
| `park_session` | `fsm_transition(id, "PARKED", meta)` | Yes (FSM handles) |
| `resume_session` | `fsm_transition(id, "ACTIVE", meta)` | Yes (FSM handles) |
| `wrap_session` | `fsm_transition(id, "ARCHIVED", meta)` | Yes (FSM handles) |
| `mark_complete` | Direct write to task status | No (single-writer) |
| `transition_phase` | `session-manager.sh transition` | Yes (session-manager handles) |
| `update_field` | Direct write | No (single-writer) |

### 2.5 Coordination with Orchestrator

When in orchestrated mode, the flow is:

```
User: /start <initiative>
    |
    v
UserPromptSubmit: orchestrator-router.sh injects CONSULTATION_REQUEST
    |
    v
Main thread invokes: Task(orchestrator, "CONSULTATION_REQUEST...")
    |
    v
Orchestrator returns: Work breakdown + first phase directive
    |
    v
Main thread invokes: Task(specialist, "phase work...")
    |
    v
Specialist completes, returns artifact path
    |
    v
Orchestrator marks phase complete via Task(moirai, "...")
    |
    v
Hooks track: artifact-tracker.sh logs artifact
             session-audit.sh logs mutation
```

**Key Constraint:** During orchestrated workflows, **only hooks and state-mate** should mutate SESSION_CONTEXT.md. The main thread and specialists should NOT call state-mate directly for phase transitions - the orchestrator coordinates this.

---

## Part 3: v1 to v2 Migration Design

### 3.1 Schema Differences

| Field | v1 (Legacy) | v2 (Current) |
|-------|-------------|--------------|
| `schema_version` | Absent | `"2.0"` or `"2.1"` |
| `session_id` | Present | Present |
| `status` | Absent (inferred) | `"ACTIVE"`, `"PARKED"`, `"ARCHIVED"` |
| `created_at` | Present | Present |
| `parked_at` | Present when parked | Present when parked |
| `auto_parked_at` | Present when auto-parked | Merged into `parked_at` with `auto: true` |
| `team` | Absent | Present (null for cross-cutting) |

### 3.2 Migration Algorithm

**Location:** `.claude/hooks/lib/session-migrate.sh` (existing file, extend)

**Algorithm (Pseudocode):**

```
FUNCTION migrate_v1_to_v2(session_id):
    ctx_file = sessions_dir / session_id / "SESSION_CONTEXT.md"

    # Check if already v2
    IF has_field(ctx_file, "schema_version"):
        RETURN "already_migrated"

    # Create backup
    backup_file = ctx_file + ".v1.backup"
    copy(ctx_file, backup_file)

    # Determine current state from v1 fields
    status = "ACTIVE"
    IF has_field(ctx_file, "parked_at") OR has_field(ctx_file, "auto_parked_at"):
        status = "PARKED"
    ELSE IF has_field(ctx_file, "completed_at") OR has_field(ctx_file, "archived_at"):
        status = "ARCHIVED"

    # Determine team
    team = read_field(ctx_file, "active_team")
    IF team == "" OR team == "none":
        team_value = "null"
    ELSE:
        team_value = quoted(team)

    # Build v2 frontmatter additions
    additions = [
        "schema_version: \"2.1\"",
        "status: \"" + status + "\"",
        "team: " + team_value
    ]

    # Merge auto_parked fields if present
    IF has_field(ctx_file, "auto_parked_at"):
        auto_ts = read_field(ctx_file, "auto_parked_at")
        IF NOT has_field(ctx_file, "parked_at"):
            additions.append("parked_at: " + auto_ts)
            additions.append("parked_auto: true")
        remove_field(ctx_file, "auto_parked_at")
        remove_field(ctx_file, "auto_parked_reason")

    # Insert additions before closing ---
    insert_before_closing_frontmatter(ctx_file, additions)

    # Validate result
    IF NOT fsm_validate_context(ctx_file):
        # Rollback
        copy(backup_file, ctx_file)
        delete(backup_file)
        RETURN "validation_failed"

    # Log migration
    log_to_audit("MIGRATED_V1_TO_V2", session_id)

    RETURN "success"
```

### 3.3 Migration Implementation

```bash
# session-migrate.sh (EXTENDED)

# Migrate a v1 session to v2 schema
# Usage: migrate_session_v1_to_v2 <session_id>
# Returns: 0 on success, 1 on failure
# Output: JSON result object
migrate_session_v1_to_v2() {
    local session_id="$1"

    if [[ -z "$session_id" ]]; then
        echo '{"success": false, "error": "session_id required"}'
        return 1
    fi

    local ctx_file="$SESSIONS_DIR/$session_id/SESSION_CONTEXT.md"

    if [[ ! -f "$ctx_file" ]]; then
        echo '{"success": false, "error": "Session not found"}'
        return 1
    fi

    # Check if already v2
    local version
    version=$(grep -m1 "^schema_version:" "$ctx_file" 2>/dev/null | cut -d: -f2- | tr -d ' "')
    if [[ -n "$version" ]]; then
        echo "{\"success\": true, \"status\": \"already_v2\", \"version\": \"$version\"}"
        return 0
    fi

    # Create backup
    local backup_file="${ctx_file}.v1.backup"
    cp "$ctx_file" "$backup_file" || {
        echo '{"success": false, "error": "Failed to create backup"}'
        return 1
    }

    # Determine current state from v1 fields
    local status="ACTIVE"
    if grep -qE "^(parked_at|auto_parked_at):" "$ctx_file" 2>/dev/null; then
        status="PARKED"
    elif grep -qE "^(completed_at|archived_at):" "$ctx_file" 2>/dev/null; then
        status="ARCHIVED"
    fi

    # Determine team value
    local active_team
    active_team=$(grep -m1 "^active_team:" "$ctx_file" 2>/dev/null | cut -d: -f2- | tr -d ' "')
    local team_value
    if [[ -z "$active_team" || "$active_team" == "none" ]]; then
        team_value="null"
    else
        team_value="\"$active_team\""
    fi

    # Handle auto_parked fields
    local auto_parked_at
    auto_parked_at=$(grep -m1 "^auto_parked_at:" "$ctx_file" 2>/dev/null | cut -d: -f2- | tr -d ' "')
    local has_parked_at
    has_parked_at=$(grep -c "^parked_at:" "$ctx_file" 2>/dev/null || echo "0")

    # Build new content with v2 fields
    local temp_file
    temp_file=$(mktemp) || {
        echo '{"success": false, "error": "Failed to create temp file"}'
        return 1
    }

    # Process file: add v2 fields before closing ---
    awk -v status="$status" -v team="$team_value" -v auto_ts="$auto_parked_at" -v has_parked="$has_parked_at" '
        BEGIN { frontmatter_count = 0; added = 0 }
        /^---$/ {
            frontmatter_count++
            if (frontmatter_count == 2 && added == 0) {
                print "schema_version: \"2.1\""
                print "status: \"" status "\""
                print "team: " team
                if (auto_ts != "" && has_parked == "0") {
                    print "parked_at: \"" auto_ts "\""
                    print "parked_auto: true"
                }
                added = 1
            }
        }
        # Skip auto_parked fields (merged into parked_at)
        /^auto_parked_at:/ { next }
        /^auto_parked_reason:/ { next }
        { print }
    ' "$ctx_file" > "$temp_file"

    # Replace original with transformed content
    mv "$temp_file" "$ctx_file" || {
        cp "$backup_file" "$ctx_file"
        rm -f "$backup_file"
        echo '{"success": false, "error": "Failed to write migrated content"}'
        return 1
    }

    # Validate result using FSM validator
    if ! _fsm_validate_context "$ctx_file" 2>/dev/null; then
        # Rollback
        cp "$backup_file" "$ctx_file"
        rm -f "$backup_file"
        echo '{"success": false, "error": "Validation failed after migration"}'
        return 1
    fi

    # Log migration to audit trail
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    local audit_file="$SESSIONS_DIR/.audit/migrations.log"
    mkdir -p "$(dirname "$audit_file")" 2>/dev/null
    echo "$timestamp | $session_id | V1_TO_V2 | SUCCESS" >> "$audit_file"

    echo "{\"success\": true, \"status\": \"migrated\", \"from_version\": \"1.0\", \"to_version\": \"2.1\"}"
    return 0
}

# Rollback a v2 session to v1 (disaster recovery)
# Usage: rollback_session_v2_to_v1 <session_id>
# Returns: 0 on success, 1 on failure
rollback_session_v2_to_v1() {
    local session_id="$1"

    if [[ -z "$session_id" ]]; then
        echo '{"success": false, "error": "session_id required"}'
        return 1
    fi

    local ctx_file="$SESSIONS_DIR/$session_id/SESSION_CONTEXT.md"
    local backup_file="${ctx_file}.v1.backup"

    if [[ ! -f "$backup_file" ]]; then
        echo '{"success": false, "error": "No v1 backup found"}'
        return 1
    fi

    # Restore backup
    cp "$backup_file" "$ctx_file" || {
        echo '{"success": false, "error": "Failed to restore backup"}'
        return 1
    }

    # Log rollback
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    local audit_file="$SESSIONS_DIR/.audit/migrations.log"
    echo "$timestamp | $session_id | V2_TO_V1_ROLLBACK | SUCCESS" >> "$audit_file"

    echo '{"success": true, "status": "rolled_back"}'
    return 0
}

# Dry-run migration to preview changes
# Usage: migrate_session_dry_run <session_id>
migrate_session_dry_run() {
    local session_id="$1"

    if [[ -z "$session_id" ]]; then
        echo '{"success": false, "error": "session_id required"}'
        return 1
    fi

    local ctx_file="$SESSIONS_DIR/$session_id/SESSION_CONTEXT.md"

    if [[ ! -f "$ctx_file" ]]; then
        echo '{"success": false, "error": "Session not found"}'
        return 1
    fi

    # Check if already v2
    local version
    version=$(grep -m1 "^schema_version:" "$ctx_file" 2>/dev/null | cut -d: -f2- | tr -d ' "')
    if [[ -n "$version" ]]; then
        echo "{\"dry_run\": true, \"status\": \"already_v2\", \"changes\": []}"
        return 0
    fi

    # Compute what would change
    local status="ACTIVE"
    if grep -qE "^(parked_at|auto_parked_at):" "$ctx_file" 2>/dev/null; then
        status="PARKED"
    elif grep -qE "^(completed_at|archived_at):" "$ctx_file" 2>/dev/null; then
        status="ARCHIVED"
    fi

    local active_team
    active_team=$(grep -m1 "^active_team:" "$ctx_file" 2>/dev/null | cut -d: -f2- | tr -d ' "')

    local auto_parked_at
    auto_parked_at=$(grep -m1 "^auto_parked_at:" "$ctx_file" 2>/dev/null | cut -d: -f2- | tr -d ' "')

    cat <<EOF
{
  "dry_run": true,
  "status": "would_migrate",
  "changes": [
    {"field": "schema_version", "action": "add", "value": "2.1"},
    {"field": "status", "action": "add", "value": "$status"},
    {"field": "team", "action": "add", "value": "${active_team:-null}"}
EOF

    if [[ -n "$auto_parked_at" ]]; then
        cat <<EOF
    ,{"field": "auto_parked_at", "action": "merge_to_parked_at"},
    {"field": "auto_parked_reason", "action": "remove"}
EOF
    fi

    echo '  ]'
    echo '}'
}

# Auto-migrate on first access (called from session-manager.sh status)
auto_migrate_if_needed() {
    local session_id="$1"

    # Quick check: if schema_version exists, skip
    local ctx_file="$SESSIONS_DIR/$session_id/SESSION_CONTEXT.md"
    if grep -q "^schema_version:" "$ctx_file" 2>/dev/null; then
        return 0
    fi

    # Perform silent migration
    migrate_session_v1_to_v2 "$session_id" >/dev/null 2>&1
    return $?
}
```

### 3.4 Migration Script CLI

```bash
# session-migrate.sh CLI extension

_migrate_main() {
    local cmd="${1:-help}"
    shift || true

    case "$cmd" in
        migrate)
            migrate_session_v1_to_v2 "$@"
            ;;
        rollback)
            rollback_session_v2_to_v1 "$@"
            ;;
        dry-run)
            migrate_session_dry_run "$@"
            ;;
        list-v1)
            # List all v1 sessions needing migration
            for dir in "$SESSIONS_DIR"/session-*; do
                [[ -d "$dir" ]] || continue
                local ctx="$dir/SESSION_CONTEXT.md"
                [[ -f "$ctx" ]] || continue
                if ! grep -q "^schema_version:" "$ctx" 2>/dev/null; then
                    basename "$dir"
                fi
            done
            ;;
        migrate-all)
            # Migrate all v1 sessions
            local count=0
            local failed=0
            for dir in "$SESSIONS_DIR"/session-*; do
                [[ -d "$dir" ]] || continue
                local session_id
                session_id=$(basename "$dir")
                local result
                result=$(migrate_session_v1_to_v2 "$session_id")
                if [[ "$result" == *'"success": true'* ]]; then
                    ((count++))
                else
                    ((failed++))
                    echo "Failed: $session_id - $result" >&2
                fi
            done
            echo "{\"migrated\": $count, \"failed\": $failed}"
            ;;
        help|--help|-h)
            cat <<EOF
session-migrate.sh - Session Schema Migration

Commands:
  migrate <session_id>     Migrate v1 session to v2 schema
  rollback <session_id>    Rollback v2 session to v1 backup
  dry-run <session_id>     Preview migration without applying
  list-v1                  List all v1 sessions needing migration
  migrate-all              Migrate all v1 sessions (batch)

Migration creates .v1.backup file for rollback capability.
EOF
            ;;
        *)
            echo "Unknown command: $cmd" >&2
            return 1
            ;;
    esac
}
```

---

## Part 4: Extended Test Matrix

### 4.1 Concurrency Tests (from TDD-session-manager-locking.md)

| Test ID | Scenario | Expected Outcome | TLA+ Property |
|---------|----------|------------------|---------------|
| CONC-001 | 5 parallel session creates | 1 succeeds, 4 fail with "already exists" | MutualExclusion |
| CONC-002 | Create during park | Both complete without corruption | MutualExclusion |
| CONC-003 | 10 parallel status reads | All return consistent state | LockedReadsAreConsistent |
| CONC-004 | Lock timeout scenario | Error returned, no hang | NoDeadlock |
| CONC-005 | Stale lock cleanup | Old lock removed, new operation succeeds | HolderNotInQueue |

### 4.2 Hook Ordering Tests

| Test ID | Scenario | Hooks Involved | Expected Outcome |
|---------|----------|----------------|------------------|
| HOOK-001 | SessionStart with parked session | session-context, orchestrated-mode | Parked status shown in context |
| HOOK-002 | PreToolUse Edit to SESSION_CONTEXT.md | session-write-guard, delegation-check | Blocked with state-mate guidance |
| HOOK-003 | PreToolUse Write to implementation file during workflow | session-write-guard, delegation-check | Warning emitted, write allowed |
| HOOK-004 | PostToolUse Write PRD artifact | artifact-tracker, session-audit | Artifact logged, audit entry created |
| HOOK-005 | Stop event after active session | auto-park | Session transitions to PARKED via FSM |

**Implementation (BATS):**

```bash
# tests/integration/hook-ordering.bats

@test "HOOK-001: SessionStart shows parked status" {
    # Setup: Create and park a session
    local session_id
    session_id=$(.claude/hooks/lib/session-manager.sh create "test" MODULE | jq -r '.session_id')
    .claude/hooks/lib/session-fsm.sh transition "$session_id" PARKED

    # Execute: Run session-context hook
    run .claude/hooks/context-injection/session-context.sh

    # Assert: Output contains parked indicator
    [[ "$output" == *"PARKED"* ]]
}

@test "HOOK-002: session-write-guard blocks CONTEXT writes" {
    # Setup: Create active session
    export CLAUDE_HOOK_TOOL_NAME="Write"
    export CLAUDE_HOOK_FILE_PATH=".claude/sessions/test/SESSION_CONTEXT.md"

    # Execute
    run .claude/hooks/session-guards/session-write-guard.sh

    # Assert: Blocked
    [[ "$status" -eq 1 ]]
    [[ "$output" == *'"decision": "block"'* ]]
}

@test "HOOK-003: delegation-check warns on impl file during workflow" {
    # Setup: Create session with active workflow
    local session_id
    session_id=$(.claude/hooks/lib/session-manager.sh create "test" MODULE | jq -r '.session_id')
    # Set workflow.active in context
    echo "workflow:" >> ".claude/sessions/$session_id/SESSION_CONTEXT.md"
    echo "  active: true" >> ".claude/sessions/$session_id/SESSION_CONTEXT.md"

    export CLAUDE_PROJECT_DIR="."
    echo '{"tool_name": "Write", "tool_input": {"file_path": "src/main.ts"}}' | run .claude/hooks/validation/delegation-check.sh

    # Assert: Warning emitted but allowed
    [[ "$status" -eq 0 ]]
    [[ "$output" == *"[DELEGATION]"* ]] || [[ -z "$output" ]]
}

@test "HOOK-005: auto-park transitions via FSM" {
    # Setup: Create active session
    local session_id
    session_id=$(.claude/hooks/lib/session-manager.sh create "test" MODULE | jq -r '.session_id')

    # Execute: Run auto-park hook
    export SESSION_DIR=".claude/sessions/$session_id"
    run .claude/hooks/session-guards/auto-park.sh

    # Assert: State is now PARKED
    local state
    state=$(.claude/hooks/lib/session-fsm.sh get-state "$session_id")
    [[ "$state" == "PARKED" ]]
}
```

### 4.3 v1-v2 Migration Tests

| Test ID | Scenario | Expected Outcome |
|---------|----------|------------------|
| MIGRATE-001 | Migrate ACTIVE v1 session | schema_version=2.1, status=ACTIVE |
| MIGRATE-002 | Migrate PARKED v1 session (parked_at) | status=PARKED preserved |
| MIGRATE-003 | Migrate auto-parked v1 session | auto_parked merged to parked_at+parked_auto |
| MIGRATE-004 | Migrate cross-cutting session (no team) | team=null |
| MIGRATE-005 | Migrate already v2 session | No-op, returns already_migrated |
| MIGRATE-006 | Rollback v2 to v1 | Original v1 content restored |
| MIGRATE-007 | Dry-run preview | Shows changes without applying |
| MIGRATE-008 | Validation failure during migration | Rollback to backup, error returned |

**Implementation (BATS):**

```bash
# tests/unit/session-migrate.bats

setup() {
    export SESSIONS_DIR="$BATS_TMPDIR/sessions"
    mkdir -p "$SESSIONS_DIR"
    source .claude/hooks/lib/session-migrate.sh
}

@test "MIGRATE-001: v1 ACTIVE session migrates to v2" {
    # Create v1 session
    local session_id="session-20260101-120000-abcd1234"
    mkdir -p "$SESSIONS_DIR/$session_id"
    cat > "$SESSIONS_DIR/$session_id/SESSION_CONTEXT.md" <<EOF
---
session_id: "$session_id"
created_at: "2026-01-01T12:00:00Z"
initiative: "Test"
complexity: "MODULE"
active_team: "test-pack"
current_phase: "requirements"
---
# Test Session
EOF

    # Execute migration
    run migrate_session_v1_to_v2 "$session_id"

    # Assert success
    [[ "$status" -eq 0 ]]
    [[ "$output" == *'"success": true'* ]]

    # Verify v2 fields added
    grep -q "^schema_version:" "$SESSIONS_DIR/$session_id/SESSION_CONTEXT.md"
    grep -q '^status: "ACTIVE"' "$SESSIONS_DIR/$session_id/SESSION_CONTEXT.md"
}

@test "MIGRATE-003: auto_parked merged to parked_at" {
    # Create v1 auto-parked session
    local session_id="session-20260101-120000-efgh5678"
    mkdir -p "$SESSIONS_DIR/$session_id"
    cat > "$SESSIONS_DIR/$session_id/SESSION_CONTEXT.md" <<EOF
---
session_id: "$session_id"
created_at: "2026-01-01T12:00:00Z"
initiative: "Test"
complexity: "MODULE"
active_team: "none"
current_phase: "requirements"
auto_parked_at: "2026-01-01T13:00:00Z"
auto_parked_reason: "Session stopped"
---
EOF

    # Execute migration
    run migrate_session_v1_to_v2 "$session_id"

    # Assert auto_parked fields removed, parked_at added
    ! grep -q "^auto_parked_at:" "$SESSIONS_DIR/$session_id/SESSION_CONTEXT.md"
    grep -q "^parked_at:" "$SESSIONS_DIR/$session_id/SESSION_CONTEXT.md"
    grep -q "^parked_auto: true" "$SESSIONS_DIR/$session_id/SESSION_CONTEXT.md"
    grep -q '^status: "PARKED"' "$SESSIONS_DIR/$session_id/SESSION_CONTEXT.md"
}

@test "MIGRATE-006: rollback restores v1" {
    # Create and migrate a session
    local session_id="session-20260101-120000-ijkl9012"
    mkdir -p "$SESSIONS_DIR/$session_id"
    cat > "$SESSIONS_DIR/$session_id/SESSION_CONTEXT.md" <<EOF
---
session_id: "$session_id"
created_at: "2026-01-01T12:00:00Z"
initiative: "Test"
complexity: "MODULE"
active_team: "test-pack"
current_phase: "requirements"
---
EOF

    migrate_session_v1_to_v2 "$session_id"

    # Verify v2
    grep -q "^schema_version:" "$SESSIONS_DIR/$session_id/SESSION_CONTEXT.md"

    # Execute rollback
    run rollback_session_v2_to_v1 "$session_id"

    # Assert v1 restored
    [[ "$status" -eq 0 ]]
    ! grep -q "^schema_version:" "$SESSIONS_DIR/$session_id/SESSION_CONTEXT.md"
}
```

### 4.4 state-mate Coordination Tests

| Test ID | Scenario | Expected Outcome |
|---------|----------|------------------|
| MATE-001 | state-mate bypasses write guard | Write allowed, no block |
| MATE-002 | Regular agent blocked by write guard | Write blocked with guidance |
| MATE-003 | state-mate invokes FSM transition | Lock acquired, state changed |
| MATE-004 | state-mate concurrent invocation | Task tool serializes, no conflict |

**Implementation (BATS):**

```bash
# tests/integration/state-mate-coordination.bats

@test "MATE-001: STATE_MATE_BYPASS allows write" {
    export CLAUDE_HOOK_TOOL_NAME="Write"
    export CLAUDE_HOOK_FILE_PATH=".claude/sessions/test/SESSION_CONTEXT.md"
    export STATE_MATE_BYPASS="true"

    run .claude/hooks/session-guards/session-write-guard.sh

    [[ "$status" -eq 0 ]]
}

@test "MATE-002: Regular agent blocked" {
    export CLAUDE_HOOK_TOOL_NAME="Write"
    export CLAUDE_HOOK_FILE_PATH=".claude/sessions/test/SESSION_CONTEXT.md"
    unset STATE_MATE_BYPASS

    run .claude/hooks/session-guards/session-write-guard.sh

    [[ "$status" -eq 1 ]]
}
```

### 4.5 Performance Benchmarks

| Metric | Baseline | Target | Method |
|--------|----------|--------|--------|
| Session create latency | ~100ms | <150ms | `time session-manager.sh create` |
| Lock acquisition (flock) | ~10ms | <20ms | Instrumented timing |
| Lock acquisition (mkdir fallback) | ~50ms | <100ms | Instrumented timing |
| Status query | ~50ms | <75ms | `time session-manager.sh status` |
| Parallel creates (5) | N/A | <2s total | All complete with 1 success |
| Hook execution (session-context) | ~80ms | <100ms | `time session-context.sh` |
| Migration (single session) | N/A | <200ms | `time session-migrate.sh migrate` |

---

## Part 5: Risk Boundaries and Protected Systems

### 5.1 What NOT to Change

| Component | Reason | Alternative |
|-----------|--------|-------------|
| CEM (roster-sync) | Out of scope; separate initiative | Flag for future work |
| Team pack structure | Affects all satellites | Document as frozen |
| ACTIVE_RITE file format | Simple string, no need to change | Keep as-is |
| ACTIVE_WORKFLOW.yaml schema | Workflow engine separate concern | Keep as-is |
| Hook event types | Claude Code API | Use existing events only |
| Schema validation rules (beyond v2.1) | Breaking change risk | Additive only |

### 5.2 Rollback Strategy

**Granular Rollback:**

```bash
# Rollback session management fixes
git checkout HEAD~1 -- .claude/hooks/lib/session-manager.sh
git checkout HEAD~1 -- .claude/hooks/lib/session-fsm.sh
git checkout HEAD~1 -- .claude/hooks/lib/session-core.sh
git checkout HEAD~1 -- .claude/hooks/lib/session-state.sh

# Rollback hook changes
git checkout HEAD~1 -- .claude/hooks/session-guards/auto-park.sh
git checkout HEAD~1 -- .claude/hooks/session-guards/session-write-guard.sh

# Rollback migration (sessions remain, backups available)
# For each migrated session:
.claude/hooks/lib/session-migrate.sh rollback <session_id>
```

**Full Rollback:**

```bash
# Revert entire commit
git revert <commit-sha>
```

### 5.3 Disaster Recovery Procedures

**Scenario 1: Lock file stuck (orphan)**

```bash
# Detect
ls -la .claude/sessions/.locks/

# Clean stale locks (process dead)
for lock in .claude/sessions/.locks/*.lock.d; do
    pid=$(cat "$lock/pid" 2>/dev/null || echo "")
    if [[ -n "$pid" ]] && ! kill -0 "$pid" 2>/dev/null; then
        rm -rf "$lock"
        echo "Removed stale lock: $lock"
    fi
done
```

**Scenario 2: SESSION_CONTEXT.md corrupted**

```bash
# Check for backup
ls -la .claude/sessions/<session_id>/SESSION_CONTEXT.md.*

# Restore from v1 backup
cp .claude/sessions/<session_id>/SESSION_CONTEXT.md.v1.backup \
   .claude/sessions/<session_id>/SESSION_CONTEXT.md

# Or restore from FSM backup
cp .claude/sessions/<session_id>/SESSION_CONTEXT.md.backup \
   .claude/sessions/<session_id>/SESSION_CONTEXT.md
```

**Scenario 3: All sessions inaccessible**

```bash
# Remove current session pointer (forces re-selection)
rm .claude/sessions/.current-session

# List all sessions
ls .claude/sessions/

# Manually set current session
echo "session-XXXXXXXX-XXXXXX-XXXXXXXX" > .claude/sessions/.current-session
```

### 5.4 Monitoring Checklist

Post-implementation, verify:

- [ ] `session-manager.sh status` returns valid JSON
- [ ] `session-manager.sh create` succeeds with new session
- [ ] `session-fsm.sh transition <id> PARKED` works
- [ ] Hooks execute without errors in Claude Code
- [ ] No orphan lock files after normal operations
- [ ] Audit logs populate in `.claude/sessions/.audit/`
- [ ] v1 sessions auto-migrate on first access

---

## Implementation Plan

### Phase 1: Quick Wins (1 day)
From TDD-session-manager-locking.md:
- LOCK-002: Fix trap quoting (1 line)
- RACE-003: Add PID suffix to backup file (1 line)
- VALID-001: Relax session ID regex (3 lines)

### Phase 2: Lock Consolidation (2 days)
From TDD-session-manager-locking.md:
- LOCK-001: Restructure cmd_create lock scope
- LOCK-003: Use automatic FD allocation
- STATE-001: Handle .current-session directory case

### Phase 3: Hook Integration (1 day)
From this document:
- Refactor `auto-park.sh` to use FSM transition
- Add `STATE_MATE_BYPASS` check to `session-write-guard.sh`
- Update `artifact-tracker.sh` to use atomic_write

### Phase 4: Migration (1 day)
From this document:
- Implement `migrate_session_v1_to_v2()`
- Implement `rollback_session_v2_to_v1()`
- Add `auto_migrate_if_needed()` to session-manager.sh status

### Phase 5: Verification & Polish (1 day)
From both documents:
- Run extended test matrix
- Verify performance benchmarks
- Document rollback procedures

**Total Estimated Effort**: 6 days

---

## Artifact Attestation

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-session-manager-ecosystem-audit.md` | Created |
| Session Locking TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-session-manager-locking.md` | Read |
| Gap Analysis | `/Users/tomtenuta/Code/roster/docs/analysis/GAP-session-manager-concurrency.md` | Read |
| state-mate ADR | `/Users/tomtenuta/Code/roster/docs/decisions/ADR-0005-state-mate-centralized-state-authority.md` | Read |
| Hook Settings | `/Users/tomtenuta/Code/roster/.claude/settings.local.json` | Read |
| Session FSM | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-fsm.sh` | Read |
| Session Manager | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-manager.sh` | Read |
| auto-park.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/session-guards/auto-park.sh` | Read |
| session-write-guard.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/session-guards/session-write-guard.sh` | Read |
| TLA+ Spec | `/Users/tomtenuta/Code/roster/docs/specs/session-fsm.tla` | Read |

---

## Appendix A: Hook Script Audit Summary

| Hook | Category | Locking Needed | Mutation Type | Fix Required |
|------|----------|----------------|---------------|--------------|
| session-context.sh | RECOVERABLE | No | None | None |
| orchestrated-mode.sh | RECOVERABLE | No | None | None |
| auto-park.sh | RECOVERABLE | FSM provides | Writes parked_at | Use FSM transition |
| command-validator.sh | DEFENSIVE | No | None | None |
| session-write-guard.sh | DEFENSIVE | No | Blocker | Add STATE_MATE_BYPASS |
| delegation-check.sh | DEFENSIVE | No | Warning | None |
| orchestrator-bypass-check.sh | DEFENSIVE | No | Warning | None |
| artifact-tracker.sh | RECOVERABLE | No | Appends | Use atomic_write |
| session-audit.sh | RECOVERABLE | No | Appends (audit) | None |
| commit-tracker.sh | RECOVERABLE | No | Appends | Use atomic_write |
| orchestrator-router.sh | RECOVERABLE | No | None | None |
| start-preflight.sh | RECOVERABLE | FSM provides | Creates session | None (uses session-manager) |

---

## Appendix B: Design Decision Log

| Decision | Rationale | Alternatives Considered |
|----------|-----------|------------------------|
| No hook-level locking | Session manager provides serialization; complexity vs. benefit | Per-hook flock (rejected: maintenance burden) |
| Preserve implicit hook ordering | Array order is sufficient; explicit priorities add complexity | Priority numbers (rejected: over-engineering) |
| STATE_MATE_BYPASS env var | Simple, explicit bypass mechanism | Agent name check (rejected: unreliable identity) |
| Auto-migrate on first access | Seamless user experience | Manual migration command (rejected: friction) |
| Keep v1 backup after migration | Safety for rollback | Delete backup (rejected: no recovery path) |
| Use FSM for auto-park | Consistency with locking model | Direct write (rejected: bypasses locking) |

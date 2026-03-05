# TDD: Hybrid Session Model Implementation

## Overview

This Technical Design Document specifies the implementation of the hybrid session model for roster. The design introduces explicit execution mode detection (native, orchestrated, cross-cutting) based on session state and rite context, enabling the system to behave appropriately across all three modes without ambiguity.

## Context

| Reference | Location |
|-----------|----------|
| PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-hybrid-session-model.md` |
| Session Manager | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-manager.sh` |
| Session FSM | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-fsm.sh` |
| Session State | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-state.sh` |
| Delegation Check Hook | `/Users/tomtenuta/Code/roster/.claude/hooks/delegation-check.sh` |
| Execution Mode Skill | `/Users/tomtenuta/Code/roster/.claude/skills/orchestration/execution-mode.md` |
| Session Context Hook | `/Users/tomtenuta/Code/roster/.claude/hooks/context-injection/session-context.sh` |
| Related TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-session-state-machine.md` |

### Problem Statement

The roster ecosystem operates in a hybrid mode alongside native Claude Code, but the relationship between these modes is implicit:

1. **No Mode Detection Function**: Components cannot programmatically determine execution mode
2. **Inappropriate Hook Warnings**: `delegation-check.sh` warns in native mode when no orchestration is active
3. **Session-Without-Team Undefined**: No defined behavior for sessions created without a team
4. **Documentation Gap**: CLAUDE.md lacks explicit mode determination rules
5. **Parked Session Ambiguity**: Parked sessions with teams are incorrectly treated as orchestrated

### Design Goals

1. Single function for execution mode detection usable by all hooks and components
2. Mode detection must complete in <50ms (NFR-1)
3. Graceful fallback to cross-cutting mode on detection failures (NFR-2)
4. Full backward compatibility with existing sessions (NFR-4)
5. Clear documentation of mode behaviors in CLAUDE.md

---

## System Design

### Architecture Overview

```
                    +---------------------------+
                    |   Claude Code CLI         |
                    +-----------+---------------+
                                |
                                v
                    +---------------------------+
                    |   CLAUDE.md Routing       |
                    |   (Execution Mode Section)|
                    +-----------+---------------+
                                |
                    +-----------+---------------+
                    |                           |
                    v                           v
        +-----------------+         +-----------------------+
        | Hooks           |         | Skills / Agents       |
        | - delegation-   |         | - execution-mode.md   |
        |   check.sh      |         | - consult-ref.md      |
        | - session-      |         | - start-ref.md        |
        |   context.sh    |         +-----------------------+
        +--------+--------+
                 |
                 v
        +---------------------------+
        |   execution_mode()        |
        |   session-manager.sh      |
        +---------------------------+
                 |
    +------------+-------------+
    |            |             |
    v            v             v
+--------+  +----------+  +------------+
| native |  | orchestr.|  | cross-     |
|        |  |          |  | cutting    |
+--------+  +----------+  +------------+
```

### Execution Mode Decision Tree

Per FR-1.1 from PRD:

```
User Intent
    |
    +-- No session active?
    |       |
    |       +-- Native Claude Mode
    |           Direct execution, no orchestration, no session tracking
    |
    +-- Session active?
            |
            +-- Has team AND session status = ACTIVE?
            |       |
            |       +-- Orchestrated Mode
            |           Main thread = Coach, delegates via Task tool
            |           (Note: Parked sessions are NOT orchestrated)
            |
            +-- No team, OR session parked/not-active?
                    |
                    +-- Cross-Cutting Mode
                        Main agent executes directly
                        Session tracking active
                        /consult available for routing
```

### Mode Detection Logic

| Condition | Mode | Behavior |
|-----------|------|----------|
| No session file exists | `native` | Direct execution, no session tracking |
| Session exists, status=ACTIVE, rite configured | `orchestrated` | Coach pattern, delegate via Task tool |
| Session exists, status=PARKED (regardless of rite) | `cross-cutting` | Direct execution with session tracking |
| Session exists, status=ACTIVE, no rite (null/none) | `cross-cutting` | Direct execution with session tracking |
| Session file corrupted or unreadable | `cross-cutting` | Graceful degradation per NFR-2 |

---

## Interface Contracts

### execution_mode Function

**Location**: `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-manager.sh`

**Signature**:
```bash
# Get current execution mode based on session state and rite context
# Returns: "native" | "orchestrated" | "cross-cutting"
# Exit code: 0 always (graceful fallback on errors)
# Performance: <50ms required (NFR-1)
execution_mode() -> string
```

**Implementation Specification**:

```bash
# =============================================================================
# Execution Mode Detection (FR-1.2)
# =============================================================================
# Returns: "native" | "orchestrated" | "cross-cutting"
# Performance requirement: <50ms
# Error handling: Falls back to "cross-cutting" on any detection failure (NFR-2)

execution_mode() {
    local session_id
    session_id=$(get_session_id 2>/dev/null) || {
        echo "native"
        return 0
    }

    # No session = native mode
    if [[ -z "$session_id" ]]; then
        echo "native"
        return 0
    fi

    # Session exists - check if directory and context file exist
    local session_dir="$SESSIONS_DIR/$session_id"
    local ctx_file="$session_dir/SESSION_CONTEXT.md"

    if [[ ! -f "$ctx_file" ]]; then
        # Session ID set but no context file = corrupted, fallback to native
        echo "native"
        return 0
    fi

    # Get session status via FSM (authoritative)
    local status
    status=$(fsm_get_state "$session_id" 2>/dev/null) || status="ACTIVE"

    # Parked sessions are cross-cutting regardless of team
    if [[ "$status" == "PARKED" ]]; then
        echo "cross-cutting"
        return 0
    fi

    # Archived sessions should not be active (edge case)
    if [[ "$status" == "ARCHIVED" ]]; then
        echo "native"
        return 0
    fi

    # Session is ACTIVE - check team configuration
    local active_rite
    active_rite=$(cat ".knossos/ACTIVE_RITE" 2>/dev/null || echo "none")

    # Also check team field in session context for cross-cutting sessions
    if [[ "$active_rite" == "none" || -z "$active_rite" ]]; then
        # Check if team is explicitly null in SESSION_CONTEXT
        local ctx_team
        ctx_team=$(grep -m1 "^active_rite:" "$ctx_file" 2>/dev/null | cut -d: -f2- | tr -d ' "')
        if [[ -z "$ctx_team" || "$ctx_team" == "none" || "$ctx_team" == "null" ]]; then
            echo "cross-cutting"
            return 0
        fi
        active_rite="$ctx_team"
    fi

    # Verify rite exists
    local rite_dir="${ROSTER_HOME:-$HOME/.config/roster}/rites/$active_rite"
    if [[ ! -d "$rite_dir" ]]; then
        # Team configured but pack missing - error state, but fallback gracefully
        echo "cross-cutting"
        return 0
    fi

    # All conditions met: ACTIVE session + valid team = orchestrated
    echo "orchestrated"
    return 0
}
```

**Test Cases**:

| Test ID | Precondition | Expected Output |
|---------|--------------|-----------------|
| `mode_001` | No session file, no ACTIVE_RITE | `native` |
| `mode_002` | Session ACTIVE, team=10x-dev, pack exists | `orchestrated` |
| `mode_003` | Session ACTIVE, team=none | `cross-cutting` |
| `mode_004` | Session PARKED, team=10x-dev | `cross-cutting` |
| `mode_005` | Session ACTIVE, team=missing-pack (pack dir missing) | `cross-cutting` |
| `mode_006` | Session ARCHIVED | `native` |
| `mode_007` | SESSION_CONTEXT.md corrupted (parse error) | `cross-cutting` |
| `mode_008` | ACTIVE_RITE file exists but empty | `cross-cutting` |

---

## SESSION_CONTEXT Schema Changes

### New Optional `team` Field

Per FR-2.3, add an optional `team` field to SESSION_CONTEXT schema. This field:
- Is distinct from `active_rite` (which syncs with ACTIVE_RITE file)
- When null/absent indicates cross-cutting mode at session creation
- Preserved through session lifecycle

**Schema Update** (`schemas/artifacts/session-context.schema.json`):

```json
{
  "properties": {
    "team": {
      "type": ["string", "null"],
      "description": "Team at session creation. Null indicates cross-cutting session. Distinct from active_rite which may change."
    }
  }
}
```

**SESSION_CONTEXT.md v2.1 Example**:

```yaml
---
schema_version: "2.1"
session_id: "session-20260102-120000-abcd1234"
status: "ACTIVE"
created_at: "2026-01-02T12:00:00Z"
initiative: "Cross-cutting refactor"
complexity: "MODULE"
active_rite: "none"
team: null                    # NEW: Explicitly null for cross-cutting
current_phase: "implementation"
---
```

### Backward Compatibility

Existing sessions without `team` field:
- Continue to work unchanged (NFR-4)
- Mode detection uses `active_rite` from ACTIVE_RITE file or SESSION_CONTEXT
- No migration required; field is optional

**Validation Update** (in `session-fsm.sh`):

```bash
# Update _fsm_validate_context to accept v2.0 and v2.1 schemas
local version
version=$(grep -m1 "^schema_version:" "$ctx_file" 2>/dev/null | cut -d: -f2- | tr -d ' "' || echo "")
case "$version" in
    2.0|2.1) ;;  # Accept both
    *)
        echo "Validation failed: Unsupported schema version: '$version'" >&2
        return 1
        ;;
esac
```

---

## Hook Behavior Matrix

### delegation-check.sh Updates

**Location**: `/Users/tomtenuta/Code/roster/.claude/hooks/delegation-check.sh`

**Current Behavior**: Checks `workflow.active` field in SESSION_CONTEXT, warns on Edit/Write when true.

**New Behavior**: Use `execution_mode()` function, only warn in orchestrated mode.

| Mode | Edit/Write on Code Files | Behavior |
|------|--------------------------|----------|
| `native` | Allowed | Silent (no warning) |
| `orchestrated` | Discouraged | Warning emitted |
| `cross-cutting` | Allowed | Silent (no warning) |

**Updated Implementation**:

```bash
#!/bin/bash
# PreToolUse (Edit/Write) hook - warn on direct implementation during orchestrated workflow
# Emits WARNING (not block) to preserve human override

set -euo pipefail

SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"
HOOKS_LIB="${CLAUDE_PROJECT_DIR:-.}/.claude/hooks/lib"

source "$HOOKS_LIB/logging.sh" 2>/dev/null && log_init "delegation-check" && log_start || true
source "$HOOKS_LIB/session-utils.sh" 2>/dev/null || { log_end 1 2>/dev/null; exit 0; }
source "$HOOKS_LIB/session-manager.sh" 2>/dev/null || { exit 0; }

INPUT=$(cat)
TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // empty' 2>/dev/null)

# Only check Edit and Write tools
if [[ "$TOOL_NAME" != "Edit" ]] && [[ "$TOOL_NAME" != "Write" ]]; then
    exit 0
fi

PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"
cd "$PROJECT_DIR" 2>/dev/null || true

# Get execution mode (FR-5.1, FR-5.2)
MODE=$(execution_mode 2>/dev/null || echo "cross-cutting")

# Only warn in orchestrated mode
if [[ "$MODE" != "orchestrated" ]]; then
    exit 0
fi

# Get file being modified
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // .tool_input.path // "unknown"' 2>/dev/null)

# Allow session/artifact files by main thread
ALLOWED_PATHS="SESSION_CONTEXT|sessions/|docs/requirements|docs/design|docs/testing"
if echo "$FILE_PATH" | grep -qE "$ALLOWED_PATHS"; then
    exit 0
fi

# Get workflow name for context (optional enhancement)
WORKFLOW_NAME=$(cat ".knossos/ACTIVE_WORKFLOW.yaml" 2>/dev/null | grep "^name:" | awk '{print $2}' || echo "active")

# Emit condensed warning
cat >&2 <<EOF
[DELEGATION] Mode: orchestrated ($WORKFLOW_NAME): $TOOL_NAME on $FILE_PATH
  -> Use Task tool to delegate, or proceed if intentional override.
  -> See: .claude/skills/orchestration/execution-mode.md
EOF

exit 0
```

### session-context.sh Updates

**Location**: `/Users/tomtenuta/Code/roster/.claude/hooks/context-injection/session-context.sh`

**Addition**: Include mode indicator in output per FR-S.3.

**Updated Output Format**:

```bash
# In output_condensed_context():
local mode
mode=$(execution_mode 2>/dev/null || echo "unknown")

cat <<EOF
## Session Context

| | |
|---|---|
| **Team** | $ACTIVE_RITE |
| **Mode** | $mode |
| **Session** | $session_display |
| **Initiative** | $initiative_display |
| **Git** | $GIT_DISPLAY |
EOF
```

### Hook Behavior Summary

| Hook | Native | Orchestrated | Cross-Cutting |
|------|--------|--------------|---------------|
| `delegation-check.sh` | Silent | Warning | Silent |
| `session-context.sh` | Mode: native | Mode: orchestrated | Mode: cross-cutting |
| `coach-mode.sh` | Inactive | Active (if exists) | Inactive |

---

## CLAUDE.md Updates

### New Execution Mode Section (FR-3.1)

Add after "## Execution Mode" header in `.claude/CLAUDE.md`:

```markdown
## Execution Mode

The roster ecosystem operates in three execution modes based on session state and rite context:

| Mode | Session | Team | Main Agent Behavior |
|------|---------|------|---------------------|
| **Native** | No | - | Direct execution, no session tracking |
| **Cross-Cutting** | Yes | No (or parked) | Direct execution with session tracking |
| **Orchestrated** | Yes (ACTIVE) | Yes | Coach pattern, delegate via Task tool |

**Mode Detection**:
- No session file -> Native
- Session ACTIVE + rite configured -> Orchestrated
- Session exists but no rite OR parked -> Cross-Cutting

**Active workflow?** (Orchestrated mode) MUST delegate via Task tool. See `orchestration/execution-mode.md`.
**Cross-cutting?** Execute directly with session tracking. `/consult` available for routing guidance.
**Native?** Execute directly, no session overhead.
```

### Agent Routing Section Update (FR-3.2)

Update the "## Agent Routing" section:

```markdown
## Agent Routing

**Orchestrated?** Delegate via Task tool. Main thread is Coach.
**Cross-cutting?** Execute directly with session tracking.
**Native?** Execute directly, no session.
**Unsure?** Route to `/consult` for guidance.
```

---

## /start Command Updates

### Session-Without-Team Support (FR-2.1, FR-2.2)

Allow `/start <initiative>` to create a session without requiring rite specification.

**Location**: Update `/start` skill (`.claude/skills/session-lifecycle/start-ref.md`) and `session-manager.sh`.

**Updated cmd_create**:

```bash
cmd_create() {
    local initiative="${1:-unnamed}"
    local complexity="${2:-MODULE}"
    local team="${3:-}"  # Changed: No default, allow empty

    # If team not specified, check ACTIVE_RITE file
    if [[ -z "$team" ]]; then
        team=$(cat ".knossos/ACTIVE_RITE" 2>/dev/null || echo "")
    fi

    # Normalize empty/none to explicit null for cross-cutting
    if [[ -z "$team" || "$team" == "none" ]]; then
        team="none"  # Explicit marker for cross-cutting
    fi

    # ... rest of creation logic ...

    # Create SESSION_CONTEXT.md with team field
    cat > "$ctx_file" <<CONTEXT
---
schema_version: "2.1"
session_id: "$session_id"
status: "ACTIVE"
created_at: "$timestamp"
initiative: "$initiative"
complexity: "$complexity"
active_rite: "$team"
team: $([ "$team" == "none" ] && echo "null" || echo "\"$team\"")
current_phase: "requirements"
---
CONTEXT

    # ... validation and response ...
}
```

### New --no-team Flag (FR-S.1)

```bash
# Parse arguments in /start skill
if [[ "$1" == "--no-team" ]]; then
    EXPLICIT_NO_TEAM=true
    shift
fi

# Pass to session-manager.sh create
if [[ "$EXPLICIT_NO_TEAM" == "true" ]]; then
    session-manager.sh create "$INITIATIVE" "$COMPLEXITY" "none"
else
    session-manager.sh create "$INITIATIVE" "$COMPLEXITY"
fi
```

---

## Edge Case Handling

### Edge Case Implementation Matrix

| Case | Detection | Behavior | Implementation |
|------|-----------|----------|----------------|
| Session file corrupted | `fsm_get_state` returns error | `cross-cutting` (graceful) | `execution_mode()` catches error, returns fallback |
| ACTIVE_RITE exists but pack missing | Check `$ROSTER_HOME/rites/$team` dir | `cross-cutting` + warn | Log warning, return cross-cutting |
| Session with team, then team file deleted | ACTIVE_RITE file absent | Downgrade to `cross-cutting` | Mode detection checks both sources |
| Session-without-team receives `/handoff` | No orchestrator available | Error with guidance | `/handoff` skill checks mode first |
| Native mode receives `/park` | No session to park | Error message | `/park` skill checks session existence |
| `/consult` in native mode | Mode detection | Works, may suggest `/start` | `/consult` adapts response to mode |
| Parked session with team | status=PARKED | `cross-cutting` | Mode detection prioritizes status |

### /handoff in Cross-Cutting Mode (Edge Case Handler)

```bash
# In /handoff skill implementation
MODE=$(execution_mode 2>/dev/null || echo "cross-cutting")

if [[ "$MODE" != "orchestrated" ]]; then
    cat >&2 <<EOF
Error: /handoff requires orchestrated mode.

Current mode: $MODE
To enable orchestration:
  1. Use /team <pack-name> to activate a team
  2. Or start a new session with a team: /start "initiative" --team <pack>

Alternatively, in cross-cutting mode:
  - Execute directly (no delegation needed)
  - Use /consult for routing guidance
EOF
    exit 1
fi
```

---

## /consult Enhancement (FR-4.1, FR-4.2)

### Cross-Cutting Mode Response

Update `/consult` skill (`.claude/skills/consult-ref.md`):

```markdown
## Mode-Aware Response

When invoked, check execution mode first:

### Cross-Cutting Mode Response Template

```
Current Mode: Cross-Cutting

You're in a session without team orchestration. In this mode:
- Direct execution is valid (Edit/Write allowed)
- Session tracking is active (artifacts, blockers recorded)
- No delegation required

Options:
1. Continue directly - You can implement this yourself
2. Switch to orchestrated mode: /team <pack-name>
3. Get routing advice - Describe your task and I'll suggest an approach
```

### Native Mode Response Template

```
Current Mode: Native

No session is active. Options:
1. Execute directly - For quick tasks, just do it
2. Start a tracked session: /start "<initiative>"
3. Start with a team: /start "<initiative>" --team <pack>
```
```

---

## Performance Requirements

### NFR-1: Mode Detection <50ms

**Implementation Strategy**:

1. **No External Commands**: Use bash builtins where possible
2. **Single File Reads**: Read SESSION_CONTEXT once, cache relevant fields
3. **Early Exit**: Return as soon as mode is determined
4. **No Network**: All detection is local filesystem

**Benchmark Test**:

```bash
@test "mode_perf: execution_mode completes in <50ms" {
    # Setup: Create active session with team
    local session_id
    session_id=$(fsm_create_session "Perf Test" "MODULE" "10x-dev")

    # Measure time
    local start_ms end_ms duration_ms
    start_ms=$(date +%s%3N)

    local mode
    mode=$(execution_mode)

    end_ms=$(date +%s%3N)
    duration_ms=$((end_ms - start_ms))

    # Assert
    [ "$mode" = "orchestrated" ]
    [ "$duration_ms" -lt 50 ]
}
```

### NFR-2: Graceful Fallback

**Implementation**: All error paths in `execution_mode()` return `cross-cutting`:

```bash
# Any parse error, missing file, or unexpected state
# falls back to cross-cutting (preserves session tracking)
```

---

## Test Strategy

### Unit Tests

**Location**: `tests/unit/execution-mode.bats`

| Test ID | Description | Requirement |
|---------|-------------|-------------|
| `mode_001` | No session returns native | FR-1.1 |
| `mode_002` | Active session + team = orchestrated | FR-1.1 |
| `mode_003` | Active session + no team = cross-cutting | FR-1.1, FR-2.2 |
| `mode_004` | Parked session + team = cross-cutting | FR-1.1 |
| `mode_005` | Corrupted session = cross-cutting | NFR-2 |
| `mode_006` | Performance <50ms | NFR-1 |
| `mode_007` | Archived session = native | FR-1.1 |
| `mode_008` | Rite missing = cross-cutting | Edge case |

### Integration Tests

**Location**: `tests/integration/hybrid-session.bats`

| Test ID | Description | Requirement |
|---------|-------------|-------------|
| `int_001` | delegation-check silent in native mode | FR-5.1 |
| `int_002` | delegation-check warns in orchestrated mode | FR-5.1 |
| `int_003` | delegation-check silent in cross-cutting mode | FR-5.1 |
| `int_004` | /start without team creates cross-cutting session | FR-2.1 |
| `int_005` | /start --no-team explicit cross-cutting | FR-S.1 |
| `int_006` | /handoff errors in cross-cutting mode | Edge case |
| `int_007` | /consult adapts response to mode | FR-4.1, FR-4.2 |
| `int_008` | Mode indicator in session-context output | FR-S.3 |
| `int_009` | Existing sessions continue working | NFR-4 |
| `int_010` | /team upgrade from cross-cutting to orchestrated | FR-S.4 |

### Hook Behavior Tests

```bash
@test "int_003: delegation-check silent in cross-cutting mode" {
    # Setup: Create session without team
    local session_id
    session_id=$(session-manager.sh create "Cross-cutting test" "MODULE" "none" | jq -r '.session_id')

    # Verify mode
    local mode
    mode=$(execution_mode)
    [ "$mode" = "cross-cutting" ]

    # Simulate Edit tool invocation
    local hook_output
    hook_output=$(echo '{"tool_name":"Edit","tool_input":{"file_path":"src/main.py"}}' | \
                  bash .claude/hooks/delegation-check.sh 2>&1)

    # Assert: No warning emitted
    [ -z "$hook_output" ]
}
```

---

## Implementation Phases

### Phase 1: Core Mode Detection

1. Add `execution_mode()` function to `session-manager.sh`
2. Add unit tests for mode detection logic
3. Validate performance requirement (<50ms)

**Files Modified**:
- `.claude/hooks/lib/session-manager.sh`: Add `execution_mode()` function

### Phase 2: Hook Updates

1. Update `delegation-check.sh` to use `execution_mode()`
2. Update `session-context.sh` to display mode
3. Add hook behavior tests

**Files Modified**:
- `.claude/hooks/delegation-check.sh`: Use mode detection
- `.claude/hooks/context-injection/session-context.sh`: Add mode display

### Phase 3: Schema Updates

1. Add optional `team` field to SESSION_CONTEXT schema
2. Update `fsm_create_session` to accept null team
3. Update `cmd_create` in session-manager.sh
4. Add schema version bump to 2.1 (backward compatible)

**Files Modified**:
- `schemas/artifacts/session-context.schema.json`: Add team field
- `.claude/hooks/lib/session-fsm.sh`: Accept v2.1
- `.claude/hooks/lib/session-manager.sh`: Update create command

### Phase 4: Skill Updates

1. Update `execution-mode.md` skill with three modes
2. Update `consult-ref.md` with mode-aware responses
3. Update `start-ref.md` with --no-team flag
4. Update `handoff-ref.md` with mode check

**Files Modified**:
- `.claude/skills/orchestration/execution-mode.md`
- `.claude/skills/consult-ref.md`
- `.claude/skills/session-lifecycle/start-ref.md`
- `.claude/skills/session-lifecycle/handoff-ref.md`

### Phase 5: CLAUDE.md Updates

1. Add Execution Mode section
2. Update Agent Routing section
3. Add quick reference table

**Files Modified**:
- `.claude/CLAUDE.md`

### Phase 6: Integration Testing

1. Run full test suite
2. Verify backward compatibility
3. Performance validation

---

## Backward Compatibility

### Guaranteed Behaviors

1. **Existing Sessions**: Sessions without `team` field continue working
2. **Schema Validation**: Both v2.0 and v2.1 schemas accepted
3. **ACTIVE_RITE File**: Continues to be primary team source
4. **Hook Behavior**: Only warnings changed, never blocking

### Migration Strategy

**No Migration Required**: The `team` field is optional. Existing sessions:
- Are detected via ACTIVE_RITE file or active_rite frontmatter field
- Mode detection works with or without explicit `team` field
- Schema version 2.0 remains valid

---

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Mode detection logic diverges | Medium | High | Single `execution_mode()` function, all components use it |
| Performance regression in hooks | Low | Medium | Benchmark tests, early-exit optimization |
| Confusion with three modes | Medium | Medium | Clear CLAUDE.md documentation, /consult guidance |
| Breaking existing workflows | Low | High | Extensive backward compatibility testing |
| Parked session edge cases | Low | Medium | FSM status is authoritative, not team presence |

---

## Success Criteria (From PRD)

- [ ] Decision tree for execution mode documented in CLAUDE.md
- [ ] `execution_mode` function exists in hook library and returns correct mode
- [ ] `/start` works without team specification (creates cross-cutting session)
- [ ] delegation-check.sh does not warn in native or cross-cutting mode
- [ ] `/consult` acknowledges cross-cutting mode and offers routing
- [ ] Session-without-team allows direct Edit/Write without warnings
- [ ] Session-with-team enforces delegation pattern (existing behavior)
- [ ] All existing tests pass (backwards compatibility)
- [ ] New tests cover three-mode detection logic
- [ ] Main agent correctly identifies its role in each mode

---

## Open Items

| Item | Status | Owner | Notes |
|------|--------|-------|-------|
| Team upgrade/downgrade mid-session | Deferred | FR-S.4 | Implement in follow-up |
| `/mode` command | Deferred | FR-C.1 | Nice-to-have |
| Visual mode indicator | Deferred | FR-C.2 | Nice-to-have |

---

## Artifact Attestation

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-hybrid-session-model.md` | Created |
| PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-hybrid-session-model.md` | Read |
| Session Manager | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-manager.sh` | Read |
| Session FSM | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-fsm.sh` | Read |
| Delegation Check Hook | `/Users/tomtenuta/Code/roster/.claude/hooks/delegation-check.sh` | Read |
| Execution Mode Skill | `/Users/tomtenuta/Code/roster/.claude/skills/orchestration/execution-mode.md` | Read |
| Session Context Hook | `/Users/tomtenuta/Code/roster/.claude/hooks/context-injection/session-context.sh` | Read |

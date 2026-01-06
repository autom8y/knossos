# TDD: Auto-Orchestration Hook Enhancement

## Overview

This Technical Design Document specifies the implementation of automatic session bootstrap and consultation request injection for the roster hook system. The design modifies two existing hooks (`orchestrator-router.sh` and `start-preflight.sh`) to reduce session initialization friction from 3-5 manual steps to 1-2 steps.

## Context

| Reference | Location |
|-----------|----------|
| PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-auto-orchestration.md` |
| orchestrator-router.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/validation/orchestrator-router.sh` |
| start-preflight.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/session-guards/start-preflight.sh` |
| session-manager.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-manager.sh` |
| Session Context | `/Users/tomtenuta/Code/roster/.claude/sessions/session-20260104-022401-5552866f/SESSION_CONTEXT.md` |

### Problem Statement

The current session initialization workflow requires excessive manual intervention:

1. **orchestrator-router.sh** outputs a YAML CONSULTATION_REQUEST that cannot be directly executed
2. **start-preflight.sh** creates sessions but the output is informational, not actionable
3. Users must manually construct Task tool invocations with session context
4. No single hook produces a copy-paste ready orchestrator invocation

### Design Goals

1. Reduce friction from 3-5 steps to 1-2 steps
2. Output ready-to-execute Task tool invocations
3. Maintain backward compatibility with existing session management
4. Preserve user control (no auto-execution of Task invocations)
5. Follow Google Shell Style Guide conventions

---

## System Design

### Architecture Overview

```
                    /start "Initiative Name"
                            |
                            v
                +---------------------------+
                |   UserPromptSubmit Event  |
                +-------------+-------------+
                              |
              +---------------+---------------+
              |                               |
              v                               v
    +-------------------+           +-------------------+
    | orchestrator-     |           | start-preflight   |
    | router.sh (P:5)   |           | .sh (P:10)        |
    +-------------------+           +-------------------+
              |                               |
              v                               v
    - Check for orchestrator          - Check session state
    - Prepare consultation context    - Auto-create if needed
    - Output Task invocation          - Output status
              |                               |
              +---------------+---------------+
                              |
                              v
                +---------------------------+
                |   Combined Hook Output    |
                |   - Session status        |
                |   - Task(orchestrator...) |
                +---------------------------+
```

### Dual-Agent Runtime Architecture

After `/start`, the session operates with two parallel agents:

```
                  +---------------------------+
                  |     Active Session        |
                  |  SESSION_CONTEXT.md       |
                  +-------------+-------------+
                                |
          +---------------------+---------------------+
          |                                           |
          v                                           v
+-------------------+                       +-------------------+
|   orchestrator    |                       |    state-mate     |
| (workflow agent)  |                       |  (state agent)    |
+-------------------+                       +-------------------+
          |                                           |
          | CONSULTATION_REQUEST                      | State Mutations
          | CONSULTATION_RESPONSE                     | (via ADR-0005)
          |                                           |
          v                                           v
+-------------------+                       +-------------------+
| Subagent Tasks    |                       | SESSION_CONTEXT   |
| - context-arch    |                       | - session_state   |
| - integration-eng |                       | - current_phase   |
| - ecosystem-ana   |                       | - tasks[]         |
+-------------------+                       | - audit trail     |
          |                                 +-------------------+
          |                                           ^
          +-------------------------------------------+
                    Orchestrator reads state
                    (never writes directly)
```

### Agent Responsibility Matrix

| Concern | orchestrator | state-mate | session-manager.sh |
|---------|--------------|------------|-------------------|
| Session creation | No | No | Yes (CLI tool) |
| Workflow planning | Yes | No | No |
| Subagent delegation | Yes | No | No |
| Phase transitions | Requests | Executes | No |
| Task completion | Requests | Records | No |
| SESSION_CONTEXT writes | Never | Always | Initial only |
| Audit logging | No | Yes | Basic only |
| Schema validation | No | Full | Basic |

### Coordination Protocol

When orchestrator needs a state change:

```
1. Orchestrator determines phase transition needed

2. Orchestrator outputs instruction (NOT direct mutation):
   "Phase complete. To transition:
   Task(moirai, "transition_phase from=requirements to=design

   Session Context:
   - Session ID: session-xyz
   - Session Path: .claude/sessions/session-xyz/SESSION_CONTEXT.md")"

3. Main thread (or user) invokes state-mate

4. state-mate validates and executes transition

5. state-mate returns JSON confirmation

6. Orchestrator can continue with next phase
```

This separation ensures:
- Orchestrator focuses purely on workflow coordination
- state-mate maintains data integrity through validation
- Audit trail captures all state changes with reasoning
- No direct writes bypass the guard hook (ADR-0005)

### Hook Execution Order

| Priority | Hook | Role |
|----------|------|------|
| 5 | `orchestrator-router.sh` | Check orchestrator, output Task invocation |
| 10 | `start-preflight.sh` | Session creation, status output |

**Key insight**: `orchestrator-router.sh` runs BEFORE `start-preflight.sh`. This means the router must read session info that preflight creates, OR we need to coordinate differently.

**Resolution**: `start-preflight.sh` already creates the session. `orchestrator-router.sh` should read the just-created session and include context in its output.

### Coordination Strategy

Since hooks run in priority order and cannot share state directly:

1. **start-preflight.sh (P:10)**: Creates session, outputs session creation status
2. **orchestrator-router.sh (P:5)**: Runs BEFORE preflight, but we need session ID

**Problem**: Router runs at P:5 before preflight at P:10, but needs session info.

**Solution Options**:

A. **Move session creation to router** (P:5) - Router creates session if needed
B. **Change priorities** - Make preflight P:4 so it runs first
C. **Router reads after preflight** - Not possible with current hook model
D. **Router generates placeholder, preflight fills in** - Complex

**Recommended**: Option A - Move session creation responsibility to `orchestrator-router.sh` for `/start` commands. This consolidates the auto-orchestration flow in one place.

---

## Interface Contracts

### orchestrator-router.sh Updates

**Location**: `/Users/tomtenuta/Code/roster/.claude/hooks/validation/orchestrator-router.sh`

**Event**: UserPromptSubmit
**Priority**: 5 (runs before start-preflight.sh)
**Matcher**: `^/(start|sprint|task)`

**Current Output Format**:
```yaml
type: initial
initiative:
  name: "Initiative Name"
  complexity: "MODULE"
state:
  current_phase: null
context_summary: |
  User invoked /start. Assess complexity...
```

**New Output Format**:
```markdown
---
## Orchestrator Routing Required

Session created: **session-20260104-022401-abcd1234**

### Next Step

Execute the following Task invocation:

\`\`\`
Task(orchestrator, "Break down initiative into phases and tasks

Session Context:
- Session ID: session-20260104-022401-abcd1234
- Session Path: .claude/sessions/session-20260104-022401-abcd1234/SESSION_CONTEXT.md
- Initiative: Auto-Orchestration Hook Enhancement
- Complexity: MODULE
- Team: ecosystem-pack")
\`\`\`

Copy the Task invocation above and execute it, or use `/consult` for manual routing.

---
```

**Implementation Specification**:

```bash
#!/bin/bash
# orchestrator-router.sh - UserPromptSubmit hook for /start, /sprint, /task routing
# Injects ready-to-execute Task invocation when orchestrator is present
#
# Event: UserPromptSubmit
# Priority: 5 (before start-preflight.sh at 10)
# Matcher: ^/(start|sprint|task)

set -euo pipefail

SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"

# Library Resolution - per ADR-0002
HOOKS_LIB="${CLAUDE_PROJECT_DIR:-.}/.claude/hooks/lib"
source "$HOOKS_LIB/logging.sh" 2>/dev/null && log_init "orchestrator-router" && log_start || true
source "$HOOKS_LIB/session-utils.sh" 2>/dev/null || exit 0

PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"
cd "$PROJECT_DIR" 2>/dev/null || exit 0

# Get user prompt
USER_PROMPT="${CLAUDE_USER_PROMPT:-}"

# Check if this is a workflow command
if [[ ! "$USER_PROMPT" =~ ^/(start|sprint|task) ]]; then
    exit 0
fi

# Extract command
COMMAND=$(echo "$USER_PROMPT" | grep -oE '^/(start|sprint|task)' | tr -d '/')

# Check if orchestrator is present in active team
if [[ ! -f "$PROJECT_DIR/.claude/agents/orchestrator.md" ]]; then
    # No orchestrator = direct execution is valid
    log_end 0 2>/dev/null || true
    exit 0
fi

# Extract initiative from prompt (everything after the command)
INITIATIVE=$(echo "$USER_PROMPT" | sed -E "s|^/$COMMAND[[:space:]]*||" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
[[ -z "$INITIATIVE" ]] && INITIATIVE="Unnamed initiative"

# Extract complexity if provided (last word if it's a valid complexity)
COMPLEXITY="MODULE"
if [[ "$INITIATIVE" =~ (.+)[[:space:]]+(FUNCTION|MODULE|SERVICE|PLATFORM)[[:space:]]*$ ]]; then
    COMPLEXITY="${BASH_REMATCH[2]}"
    INITIATIVE="${BASH_REMATCH[1]}"
fi

# Clean up initiative (remove quotes if present)
INITIATIVE=$(echo "$INITIATIVE" | sed 's/^"//;s/"$//' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')

# Get active team
ACTIVE_RITE=$(cat ".claude/ACTIVE_RITE" 2>/dev/null || echo "none")

# Check for existing session
SESSION_ID=$(get_session_id 2>/dev/null || echo "")
SESSION_CREATED="false"

# For /start command, create session if none exists
if [[ "$COMMAND" == "start" && -z "$SESSION_ID" ]]; then
    # Create session via session-manager.sh
    SESSION_RESULT=$("$HOOKS_LIB/session-manager.sh" create "$INITIATIVE" "$COMPLEXITY" "$ACTIVE_RITE" 2>&1) || true

    if [[ "$SESSION_RESULT" == *'"success": true'* ]]; then
        SESSION_ID=$(echo "$SESSION_RESULT" | grep -o '"session_id": *"[^"]*"' | cut -d'"' -f4)
        SESSION_CREATED="true"
    fi
fi

# Get session ID if we don't have it yet (for sprint/task or if create failed)
if [[ -z "$SESSION_ID" ]]; then
    SESSION_ID=$(get_session_id 2>/dev/null || echo "")
fi

# Build session path
SESSION_PATH=""
if [[ -n "$SESSION_ID" ]]; then
    SESSION_PATH=".claude/sessions/$SESSION_ID/SESSION_CONTEXT.md"
fi

# Determine request type based on command
REQUEST_TYPE="initial"
case "$COMMAND" in
    start)
        REQUEST_TYPE="initial"
        ;;
    sprint|task)
        REQUEST_TYPE="checkpoint"
        ;;
esac

# Escape special characters in initiative for Task invocation
INITIATIVE_ESCAPED=$(echo "$INITIATIVE" | sed 's/"/\\"/g')

# Build routing context with ready-to-execute Task invocation
if [[ "$SESSION_CREATED" == "true" ]]; then
    SESSION_MSG="Session created: **$SESSION_ID**"
else
    SESSION_MSG="Using existing session: **$SESSION_ID**"
fi

cat <<EOF

---
## Orchestrator Routing Required

$SESSION_MSG

### Next Step

Execute the following Task invocation:

\`\`\`
Task(orchestrator, "Break down initiative into phases and tasks

Session Context:
- Session ID: $SESSION_ID
- Session Path: $SESSION_PATH
- Initiative: $INITIATIVE_ESCAPED
- Complexity: $COMPLEXITY
- Team: $ACTIVE_RITE
- Request Type: $REQUEST_TYPE")
\`\`\`

Copy the Task invocation above and execute it, or use \`/consult\` for manual routing.

---

EOF

log_end 0 2>/dev/null || true
exit 0
```

### start-preflight.sh Updates

**Location**: `/Users/tomtenuta/Code/roster/.claude/hooks/session-guards/start-preflight.sh`

**Changes Required**:

Since `orchestrator-router.sh` now handles session creation for `/start` commands when an orchestrator is present, `start-preflight.sh` needs to:

1. Check if session was already created by `orchestrator-router.sh`
2. Skip redundant session creation if session exists
3. Maintain existing behavior for non-orchestrated teams

**Updated Logic**:

```bash
# In start-preflight.sh, for /start command handling:

# Handle /start specifically
if [[ "$USER_PROMPT" =~ ^/start ]]; then
    if [[ "$HAS_SESSION" == "true" ]]; then
        # Session exists (may have been just created by orchestrator-router.sh)
        if [[ "$PARKED" == "true" ]]; then
            # Existing parked session - show options
            cat <<EOF
---
**Preflight Check**: Session exists (parked)
...
EOF
        else
            # Session is active - could be just-created or pre-existing
            # Check if orchestrator.md exists - if so, router already handled it
            if [[ -f "$PROJECT_DIR/.claude/agents/orchestrator.md" ]]; then
                # Orchestrator present - router created session, skip duplicate output
                exit 0
            fi
            # No orchestrator - this is pre-existing active session
            cat <<EOF
---
**Preflight Check**: Session already active
...
EOF
        fi
    else
        # No session - create one (only if no orchestrator, otherwise router handles it)
        if [[ -f "$PROJECT_DIR/.claude/agents/orchestrator.md" ]]; then
            # Orchestrator should have created session - this is an error state
            # Router runs at P:5 before preflight at P:10, so if session missing, creation failed
            exit 0  # Let router's error message stand
        fi

        # No orchestrator - create session and output status
        # ... existing session creation logic ...
    fi
fi
```

---

## Integration Points

### session-manager.sh (No Changes Required)

The existing `cmd_create` function already provides all needed functionality:

- Accepts initiative, complexity, and team parameters
- Returns JSON with session_id on success
- Handles locking for race condition prevention
- Validates no existing session

**Interface**:
```bash
session-manager.sh create "$INITIATIVE" "$COMPLEXITY" "$TEAM"
# Returns: {"success": true, "session_id": "session-...", ...}
```

### session-utils.sh (No Changes Required)

Existing utilities used:
- `get_session_id()` - Returns current session ID
- `generate_session_id()` - Generates new session ID format

### state-mate Integration (Reference Only)

state-mate is an existing agent (per ADR-0005) that handles all SESSION_CONTEXT.md mutations. This TDD does not modify state-mate, but defines how hooks coordinate with it.

**state-mate Role in Auto-Orchestration:**

| Stage | state-mate Involvement |
|-------|----------------------|
| `/start` invocation | Not invoked (session-manager.sh creates session) |
| Orchestrator planning | Not invoked (orchestrator reads state only) |
| Phase transitions | Invoked via Task tool to execute transition |
| Task completion | Invoked via Task tool to record completion |
| Session park/resume/wrap | Invoked via Task tool for lifecycle transitions |

**Coordination Pattern in Hook Output:**

When the orchestrator needs state changes, hooks may optionally include guidance:

```markdown
---
## Orchestrator Routing Required

Session created: **session-20260104-022401-abcd1234**

### Next Steps

1. Execute orchestrator invocation:
\`\`\`
Task(orchestrator, "Break down initiative into phases and tasks
...
")
\`\`\`

2. When orchestrator indicates phase transition, use:
\`\`\`
Task(moirai, "transition_phase from=<current> to=<next>

Session Context:
- Session ID: session-20260104-022401-abcd1234
- Session Path: .claude/sessions/session-20260104-022401-abcd1234/SESSION_CONTEXT.md")
\`\`\`

---
```

**Why Hooks Don't Invoke state-mate Directly:**

1. **Separation of concerns**: Hooks handle routing, agents handle execution
2. **User control**: User decides when to transition state
3. **Orchestrator authority**: Orchestrator determines WHEN transitions happen
4. **Audit clarity**: state-mate invocations are explicit and traceable

---

## Test Strategy

### Phase 1: Session Bootstrap Tests

| Test ID | Description | Requirement |
|---------|-------------|-------------|
| `boot_001` | `/start "Test"` creates session when none exists | FR-1.1 |
| `boot_002` | `/start` with existing session does not create duplicate | FR-1.2 |
| `boot_003` | Parallel `/start` commands don't race (lock test) | FR-1.3 |
| `boot_004` | Session creation failure produces graceful error | NFR-2 |

**Test Implementation** (`tests/integration/auto-orchestration.bats`):

```bash
#!/usr/bin/env bats

load '../test_helper'

setup() {
    setup_test_project
    echo "ecosystem-pack" > ".claude/ACTIVE_RITE"
    mkdir -p ".claude/agents"
    echo "# Orchestrator" > ".claude/agents/orchestrator.md"
}

teardown() {
    cleanup_test_project
}

@test "boot_001: /start creates session when none exists" {
    # Ensure no session
    rm -f ".claude/sessions/.current-session"

    # Simulate /start command via orchestrator-router.sh
    export CLAUDE_USER_PROMPT='/start "Test Initiative"'
    run bash .claude/hooks/validation/orchestrator-router.sh

    # Check session was created
    [ "$status" -eq 0 ]
    [[ "$output" == *"Session created:"* ]]
    [[ "$output" == *"Task(orchestrator"* ]]

    # Verify session file exists
    local session_id
    session_id=$(cat ".claude/sessions/.current-session" 2>/dev/null)
    [ -n "$session_id" ]
    [ -f ".claude/sessions/$session_id/SESSION_CONTEXT.md" ]
}

@test "boot_002: /start with existing session reuses it" {
    # Create existing session
    local existing_id
    existing_id=$(session-manager.sh create "Existing" "MODULE" | jq -r '.session_id')

    # Simulate /start command
    export CLAUDE_USER_PROMPT='/start "New Initiative"'
    run bash .claude/hooks/validation/orchestrator-router.sh

    # Should use existing session
    [ "$status" -eq 0 ]
    [[ "$output" == *"Using existing session:"* ]]
    [[ "$output" == *"$existing_id"* ]]
}

@test "boot_003: parallel /start commands don't race" {
    rm -f ".claude/sessions/.current-session"

    # Run two /start commands in parallel
    export CLAUDE_USER_PROMPT='/start "Parallel Test 1"'
    bash .claude/hooks/validation/orchestrator-router.sh &
    pid1=$!

    export CLAUDE_USER_PROMPT='/start "Parallel Test 2"'
    bash .claude/hooks/validation/orchestrator-router.sh &
    pid2=$!

    wait $pid1
    wait $pid2

    # Only one session should exist
    local session_count
    session_count=$(ls -1 .claude/sessions/session-* 2>/dev/null | wc -l)
    [ "$session_count" -eq 1 ]
}
```

### Phase 2: Consultation Request Tests

| Test ID | Description | Requirement |
|---------|-------------|-------------|
| `cons_001` | Task invocation includes session ID | FR-2.2 |
| `cons_002` | Task invocation includes session path | FR-2.2 |
| `cons_003` | Task invocation includes initiative | FR-2.2 |
| `cons_004` | Task invocation includes complexity | FR-2.2 |
| `cons_005` | Task invocation is syntactically valid | FR-2.4 |
| `cons_006` | Special characters in initiative are escaped | Edge case |

```bash
@test "cons_001: Task invocation includes session ID" {
    export CLAUDE_USER_PROMPT='/start "Test"'
    run bash .claude/hooks/validation/orchestrator-router.sh

    [ "$status" -eq 0 ]
    [[ "$output" == *"Session ID: session-"* ]]
}

@test "cons_005: Task invocation is syntactically valid" {
    export CLAUDE_USER_PROMPT='/start "Test Initiative"'
    run bash .claude/hooks/validation/orchestrator-router.sh

    # Extract Task invocation
    local task_invocation
    task_invocation=$(echo "$output" | sed -n '/```/,/```/p' | grep -v '```')

    # Validate format: Task(agent, "message")
    [[ "$task_invocation" == Task\(orchestrator,* ]]
    [[ "$task_invocation" == *\"\) ]]
}

@test "cons_006: special characters in initiative are escaped" {
    export CLAUDE_USER_PROMPT='/start "Test \"quoted\" initiative"'
    run bash .claude/hooks/validation/orchestrator-router.sh

    [ "$status" -eq 0 ]
    # Quotes should be escaped
    [[ "$output" == *'Test \"quoted\" initiative'* ]]
}
```

### Phase 3: Friction Measurement Tests

| Test ID | Description | Requirement |
|---------|-------------|-------------|
| `fric_001` | Measure baseline steps (before changes) | FR-3.5 |
| `fric_002` | Measure target steps (after changes) | FR-3.5 |
| `fric_003` | End-to-end: start to orchestrator invocation | FR-3.5 |

```bash
@test "fric_003: end-to-end friction is 2 steps" {
    # Step 1: User types /start
    export CLAUDE_USER_PROMPT='/start "E2E Test"'
    run bash .claude/hooks/validation/orchestrator-router.sh

    # Verify output is actionable
    [ "$status" -eq 0 ]
    [[ "$output" == *"Task(orchestrator"* ]]

    # Step 2 would be: User copies Task invocation
    # (Cannot automate Claude Code execution in test)

    # Total steps: 2 (type /start, copy Task invocation)
    # Target: 1-2 steps - PASS
}
```

### Hook Behavior Tests

| Test ID | Description | Requirement |
|---------|-------------|-------------|
| `hook_001` | orchestrator-router.sh runs at priority 5 | Design |
| `hook_002` | start-preflight.sh skips when orchestrator created session | FR-1.2 |
| `hook_003` | No orchestrator = start-preflight.sh handles creation | Backward compat |
| `hook_004` | Hook execution under 100ms | NFR-1 |

### Phase 4: State-Mate Coordination Tests

| Test ID | Description | Requirement |
|---------|-------------|-------------|
| `state_001` | Orchestrator output contains NO Write/Edit to *_CONTEXT.md | FR-4.5 |
| `state_002` | Hook output includes state-mate guidance when appropriate | FR-4.3 |
| `state_003` | state-mate invocation pattern is syntactically valid | FR-4.2 |
| `state_004` | Direct write to SESSION_CONTEXT.md is blocked by guard hook | ADR-0005 |

```bash
@test "state_001: orchestrator output contains no direct context writes" {
    # This test validates that orchestrator respects ADR-0005
    # by never outputting Write/Edit operations to *_CONTEXT.md

    # Simulate orchestrator Task output (mocked)
    local orchestrator_output="Task completed. Phase transition needed.
To record completion:
Task(moirai, \"mark_complete task-001 artifact=docs/PRD.md

Session Context:
- Session ID: session-test-123
- Session Path: .claude/sessions/session-test-123/SESSION_CONTEXT.md\")"

    # Verify no direct writes in output
    ! [[ "$orchestrator_output" == *"Write("*"SESSION_CONTEXT"* ]]
    ! [[ "$orchestrator_output" == *"Edit("*"SESSION_CONTEXT"* ]]

    # Verify state-mate delegation is present
    [[ "$orchestrator_output" == *"Task(moirai"* ]]
}

@test "state_004: direct write to SESSION_CONTEXT is blocked" {
    # Ensure the PreToolUse guard hook blocks direct writes
    # This test assumes session-write-guard.sh is installed

    if [[ ! -f ".claude/hooks/session-guards/session-write-guard.sh" ]]; then
        skip "session-write-guard.sh not installed"
    fi

    # Simulate PreToolUse hook invocation
    export TOOL_NAME="Edit"
    export FILE_PATH=".claude/sessions/session-test/SESSION_CONTEXT.md"

    run bash .claude/hooks/session-guards/session-write-guard.sh

    # Hook should block the operation
    [ "$status" -ne 0 ] || [[ "$output" == *"block"* ]]
}
```

```bash
@test "hook_004: hook execution under 100ms" {
    export CLAUDE_USER_PROMPT='/start "Perf Test"'

    local start_ms end_ms duration_ms
    start_ms=$(gdate +%s%3N 2>/dev/null || date +%s%3N)

    run bash .claude/hooks/validation/orchestrator-router.sh

    end_ms=$(gdate +%s%3N 2>/dev/null || date +%s%3N)
    duration_ms=$((end_ms - start_ms))

    [ "$status" -eq 0 ]
    [ "$duration_ms" -lt 100 ]
}
```

---

## Coding Standards

### Google Shell Style Guide Compliance

All bash scripts must follow [Google Shell Style Guide](https://google.github.io/styleguide/shellguide.html):

1. **Shebang**: `#!/bin/bash` at top
2. **set options**: `set -euo pipefail` for safety
3. **Variable naming**: `UPPER_CASE` for constants, `lower_case` for locals
4. **Quoting**: Always quote variables: `"$VAR"` not `$VAR`
5. **Command substitution**: Use `$(command)` not backticks
6. **Functions**: Use `function_name() { }` format
7. **Comments**: Explain why, not what

### Error Handling Patterns

From existing hooks:

```bash
# Graceful source with fallback
source "$HOOKS_LIB/session-utils.sh" 2>/dev/null || exit 0

# Safe command execution
SESSION_RESULT=$("$HOOKS_LIB/session-manager.sh" create ... 2>&1) || true

# Check result before parsing
if [[ "$SESSION_RESULT" == *'"success": true'* ]]; then
    # Parse JSON
fi
```

### Idempotent Operations

Hooks must be idempotent:

```bash
# Check before acting
if [[ -z "$SESSION_ID" ]]; then
    # Create session
fi

# Don't duplicate output
if [[ "$SESSION_CREATED" == "true" ]]; then
    echo "Session created: ..."
else
    echo "Using existing session: ..."
fi
```

---

## Implementation Notes

### Task 1: Modify orchestrator-router.sh

**File**: `/Users/tomtenuta/Code/roster/.claude/hooks/validation/orchestrator-router.sh`

**Changes**:
1. Add session creation logic for `/start` command
2. Source `session-utils.sh` for session utilities
3. Replace YAML output with Task tool format
4. Include session context in Task invocation
5. Add execution guidance text

**Key Implementation Details**:
- Use existing `session-manager.sh create` for session creation
- Parse JSON result to extract session_id
- Escape special characters in initiative name
- Include all context fields in Task invocation

### Task 2: Update start-preflight.sh

**File**: `/Users/tomtenuta/Code/roster/.claude/hooks/session-guards/start-preflight.sh`

**Changes**:
1. Detect if orchestrator exists in active team
2. Skip session creation output if orchestrator handled it
3. Maintain existing behavior for non-orchestrated teams

**Key Implementation Details**:
- Check for `.claude/agents/orchestrator.md` existence
- Exit early if orchestrator present and session active
- Preserve all existing edge case handling

### Task 3: Test Session Auto-Creation

**Test Scenarios**:
1. Clean state: `/start "Test"` creates new session
2. Existing session: `/start` reuses, shows appropriate message
3. Parallel: Lock prevents race conditions

### Task 4: Test Task Invocation Format

**Test Scenarios**:
1. All context fields present
2. Special character escaping
3. Valid Task tool syntax

### Task 5: Integration Testing

**Test Scenarios**:
1. End-to-end: `/start` to Task invocation output
2. Hook priority ordering works correctly
3. Non-orchestrated teams still work

### Task 6: Verify Friction Reduction

**Measurement**:
- Baseline: Document current 3-5 step flow
- Target: Verify 1-2 step flow works
- Document: Update workflow documentation

### Task 7: Documentation Updates

**Files to Update**:
1. Workflow documentation with new simplified flow
2. Hook documentation with updated behavior
3. Session lifecycle documentation

### Task 8: Validate State-Mate Coordination

**Validation Steps**:
1. Verify orchestrator.md does NOT include Write/Edit tools for *_CONTEXT.md
2. Confirm session-write-guard.sh blocks direct context writes
3. Test that orchestrator outputs state-mate delegation instructions
4. Verify state-mate can execute transitions from orchestrator guidance

**Files to Review** (no changes, validation only):
- `/Users/tomtenuta/Code/roster/user-agents/state-mate.md` - existing agent definition
- `/Users/tomtenuta/Code/roster/docs/decisions/ADR-0005-state-mate-centralized-state-authority.md` - authority definition
- `.claude/hooks/session-guards/session-write-guard.sh` - PreToolUse guard

---

## Backward Compatibility

### Guaranteed Behaviors

1. **Non-orchestrated teams**: Continue using start-preflight.sh for session creation
2. **Existing sessions**: Not affected by changes
3. **Hook priorities**: No changes to execution order
4. **Session schema**: No changes to SESSION_CONTEXT format

### Migration Path

No migration required. Changes are:
- Output format change (YAML to Task format) - purely cosmetic
- Session creation location change - transparent to users

---

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Hook timing creates race condition | Low | High | Use existing session-manager.sh locking |
| Task format not recognized by Claude | Low | High | Test with actual Claude Code before merge |
| User confusion from changed output | Medium | Low | Include brief guidance in output |
| Performance regression | Low | Medium | Benchmark test, session creation already fast |

---

## Success Criteria (From PRD)

- [ ] `/start` command auto-creates SESSION_CONTEXT.md (no manual session-manager.sh call)
- [ ] Hook outputs ready-to-execute Task tool invocation (no manual construction)
- [ ] Task invocation includes all required context (session ID, path, initiative, complexity)
- [ ] Friction reduced from 3-5 manual steps to 1-2 steps
- [ ] No regression in existing session management features
- [ ] Proper locking prevents race conditions
- [ ] All existing tests pass (backwards compatibility)
- [ ] New tests cover auto-orchestration flow

---

## Open Items

| Item | Status | Owner | Notes |
|------|--------|-------|-------|
| Auto-execution of Task invocation | Deferred | Future PRD | Safety review needed |
| Telemetry for friction measurement | Deferred | FR-C.2 | Nice-to-have |
| `/start --quiet` flag | Deferred | FR-S.3 | Nice-to-have |

---

## Artifact Attestation

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-auto-orchestration.md` | Updated |
| PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-auto-orchestration.md` | Updated |
| orchestrator-router.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/validation/orchestrator-router.sh` | Read |
| start-preflight.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/session-guards/start-preflight.sh` | Read |
| session-manager.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-manager.sh` | Read |
| SESSION_CONTEXT.md | `/Users/tomtenuta/Code/roster/.claude/sessions/session-20260104-022401-5552866f/SESSION_CONTEXT.md` | Read |
| state-mate agent | `/Users/tomtenuta/Code/roster/user-agents/state-mate.md` | Read |
| ADR-0005 | `/Users/tomtenuta/Code/roster/docs/decisions/ADR-0005-state-mate-centralized-state-authority.md` | Read |

#!/bin/bash
# session-write-guard.sh - PreToolUse hook that blocks direct writes to *_CONTEXT.md files
# Category: DEFENSIVE - must never crash Claude's tool flow
#
# Blocks: Write, Edit operations to *_CONTEXT.md files
# Allows: Read operations, operations to other files
# Response: Workflow-aware error messages
#
# This hook enforces centralized state management through the state-mate agent,
# preventing unguarded writes that could corrupt session/sprint state.
#
# OPTIMIZATION: Lazy init - check tool type BEFORE sourcing hooks-init.sh
# This saves ~35ms on 90%+ of invocations (Read, Bash, Glob, Grep, etc.)

HOOKS_LIB="${CLAUDE_PROJECT_DIR:-.}/.claude/hooks/lib"
PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"

# =============================================================================
# EARLY EXIT CHECK (before heavy initialization)
# =============================================================================

# Environment variables from Claude Code hook framework
TOOL_NAME="${CLAUDE_HOOK_TOOL_NAME:-}"
FILE_PATH="${CLAUDE_HOOK_FILE_PATH:-}"

# Early exit: Not Write or Edit operation (vast majority of calls)
[[ "$TOOL_NAME" != "Write" && "$TOOL_NAME" != "Edit" ]] && exit 0

# Early exit: Not a context file
[[ ! "$FILE_PATH" =~ _CONTEXT\.md$ ]] && exit 0

# =============================================================================
# STATE-MATE BYPASS DETECTION (per ADR-0010 Section 5)
# =============================================================================
# Priority 1: Agent name detection (future Claude Code enhancement)
# Priority 2: Environment marker (current mechanism)
# Both mechanisms log to audit trail for traceability

# Source fail-open library for bypass logging
source "$HOOKS_LIB/fail-open.sh" 2>/dev/null || true

# Priority 1: Check if invoked by state-mate agent via CLAUDE_TASK_AGENT_NAME
# Note: This variable is a proposed Claude Code enhancement. Until implemented,
# the STATE_MATE_BYPASS environment marker remains the primary mechanism.
AGENT_NAME="${CLAUDE_TASK_AGENT_NAME:-}"
if [[ "$AGENT_NAME" == "state-mate" ]]; then
    # Log bypass for audit trail
    BYPASS_CONTEXT=$(build_fail_open_context "agent_name" "$AGENT_NAME" "tool_name" "$TOOL_NAME" 2>/dev/null || echo '{}')
    log_bypass "session-write-guard.sh" "agent_name" "$FILE_PATH" "$BYPASS_CONTEXT" 2>/dev/null || true
    exit 0  # Allow write - state-mate is authorized
fi

# Priority 2: Environment marker (current/fallback mechanism)
if [[ "${STATE_MATE_BYPASS:-}" == "true" ]]; then
    # Log bypass for audit trail
    BYPASS_CONTEXT=$(build_fail_open_context "env_var" "STATE_MATE_BYPASS" "tool_name" "$TOOL_NAME" 2>/dev/null || echo '{}')
    log_bypass "session-write-guard.sh" "env_var" "$FILE_PATH" "$BYPASS_CONTEXT" 2>/dev/null || true
    exit 0  # Allow write
fi

# =============================================================================
# FULL INITIALIZATION (only for Write/Edit to *_CONTEXT.md)
# =============================================================================

# Absolute fallback if hooks-init.sh itself fails
source "$HOOKS_LIB/hooks-init.sh" 2>/dev/null || exit 0
hooks_init "session-write-guard" "DEFENSIVE"

# Source session utilities for workflow detection
source "$HOOKS_LIB/session-utils.sh" 2>/dev/null || true

# Check for active workflow and orchestrator presence
has_active_workflow() {
    local session_dir
    session_dir=$(get_session_dir 2>/dev/null || echo "")
    [[ -z "$session_dir" ]] && return 1

    local ctx_file="$session_dir/SESSION_CONTEXT.md"
    [[ ! -f "$ctx_file" ]] && return 1

    # Check if workflow.active is true or current_phase is set
    if grep -qE "^current_phase:" "$ctx_file" 2>/dev/null; then
        return 0
    fi

    # Also check for explicit workflow.active field
    if grep -A5 "^workflow:" "$ctx_file" 2>/dev/null | grep -q "active: true"; then
        return 0
    fi

    return 1
}

has_orchestrator() {
    [[ -f "$PROJECT_DIR/.claude/agents/orchestrator.md" ]]
}

# Build appropriate error message based on context
if has_active_workflow && has_orchestrator; then
    # Active workflow with orchestrator = hooks handle state
    cat >&2 <<'EOF'

## State Mutation Blocked

State mutations are handled **automatically by hooks** during active workflows.

**Why?** The orchestrator coordinates phase transitions, and hooks invoke state-mate to maintain the audit trail.

**If you need an explicit mutation**, use the appropriate command:
- `/park` - Pause current session
- `/wrap` - Complete and archive session
- `/handoff` - Transfer to another agent

**Do not** call `Task(state-mate, ...)` directly during orchestrated workflows.

EOF
else
    # No workflow or no orchestrator = suggest state-mate
    cat >&2 <<'EOF'

## State Mutation Blocked

Direct writes to `*_CONTEXT.md` files are not allowed.

**Use state-mate for all session/sprint mutations:**

```
Task(state-mate, "<your mutation request>")
```

**Examples:**
- `Task(state-mate, "mark task-001 complete")`
- `Task(state-mate, "transition to design phase")`
- `Task(state-mate, "register artifact docs/PRD-foo.md")`

See `~/.claude/agents/state-mate.md` for full documentation (synced from roster/user-agents/).

EOF
fi

# Block the operation with structured instruction
cat <<'EOF'
{
  "decision": "block",
  "reason": "Direct writes to *_CONTEXT.md files are blocked"
}
EOF

hooks_finalize 1
exit 1

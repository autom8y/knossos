#!/bin/bash
# session-write-guard.sh - PreToolUse hook that blocks direct writes to *_CONTEXT.md files
#
# Blocks: Write, Edit operations to *_CONTEXT.md files
# Allows: Read operations, operations to other files
# Response: JSON with error and instruction to use state-mate agent
#
# This hook enforces centralized state management through the state-mate agent,
# preventing unguarded writes that could corrupt session/sprint state.

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "$SCRIPT_DIR/lib/logging.sh" 2>/dev/null && log_init "session-write-guard" || true

# Environment variables from Claude Code hook framework
TOOL_NAME="${CLAUDE_HOOK_TOOL_NAME:-}"
FILE_PATH="${CLAUDE_HOOK_FILE_PATH:-}"

# Only intercept Write and Edit operations
if [[ "$TOOL_NAME" != "Write" && "$TOOL_NAME" != "Edit" ]]; then
    exit 0
fi

# Check if target is a context file (pattern: *_CONTEXT.md)
if [[ ! "$FILE_PATH" =~ _CONTEXT\.md$ ]]; then
    exit 0
fi

# Block the operation with structured instruction
cat <<'EOF'
{
  "decision": "block",
  "reason": "Direct writes to *_CONTEXT.md files are not allowed. Use state-mate agent for all session/sprint state mutations.",
  "instruction": "Use the Task tool to invoke state-mate: Task(state-mate, 'your mutation request')",
  "examples": [
    "Task(state-mate, 'update_field status=completed')",
    "Task(state-mate, 'mark_complete task-001 artifact=docs/design/TDD-foo.md')",
    "Task(state-mate, 'park_session reason=\"Taking a break\"')",
    "Task(state-mate, 'transition_phase from=design to=implementation')",
    "Task(state-mate, '--dry-run mark_complete task-001 artifact=...')"
  ],
  "documentation": ".claude/agents/state-mate.md"
}
EOF

exit 1

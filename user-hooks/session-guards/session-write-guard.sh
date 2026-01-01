#!/bin/bash
# session-write-guard.sh - PreToolUse hook that blocks direct writes to *_CONTEXT.md files
# Category: DEFENSIVE - must never crash Claude's tool flow
#
# Blocks: Write, Edit operations to *_CONTEXT.md files
# Allows: Read operations, operations to other files
# Response: JSON with error and instruction to use state-mate agent
#
# This hook enforces centralized state management through the state-mate agent,
# preventing unguarded writes that could corrupt session/sprint state.

# Library Resolution - per ADR-0002
HOOKS_LIB="${CLAUDE_PROJECT_DIR:-.}/.claude/hooks/lib"

# Absolute fallback if hooks-init.sh itself fails
source "$HOOKS_LIB/hooks-init.sh" 2>/dev/null || exit 0

hooks_init "session-write-guard" "DEFENSIVE"

# Environment variables from Claude Code hook framework
TOOL_NAME="${CLAUDE_HOOK_TOOL_NAME:-}"
FILE_PATH="${CLAUDE_HOOK_FILE_PATH:-}"

# Only intercept Write and Edit operations
if [[ "$TOOL_NAME" != "Write" && "$TOOL_NAME" != "Edit" ]]; then
    hooks_finalize 0
    exit 0
fi

# Check if target is a context file (pattern: *_CONTEXT.md)
if [[ ! "$FILE_PATH" =~ _CONTEXT\.md$ ]]; then
    hooks_finalize 0
    exit 0
fi

# Block the operation with structured instruction (condensed)
cat <<'EOF'
{
  "decision": "block",
  "reason": "Direct writes to *_CONTEXT.md files are blocked. Use state-mate for state mutations.",
  "instruction": "Task(state-mate, 'your mutation request')",
  "example": "Task(state-mate, 'mark_complete task-001 artifact=docs/design/TDD-foo.md')",
  "documentation": "user-agents/state-mate.md"
}
EOF

hooks_finalize 1
exit 1

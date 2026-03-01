---
title: Hook Exit Codes Reference
category: reference
audience: hook-authors
last_updated: 2026-01-01
related:
  - ADR-0002 Hook Library Resolution Architecture
  - docs/ORCHESTRATOR-CI-IMPLEMENTATION.md
---

# Hook Exit Codes Reference

> Comprehensive guide to hook exit code behavior in Claude Code and roster

## Overview

Claude Code hooks communicate success/failure via **bash exit codes**. Different hook types have different behaviors based on exit codes. Understanding these behaviors is critical for writing reliable hooks.

**Key Principle**: Exit codes are the **primary control mechanism** for hooks. They determine whether operations proceed or are blocked.

---

## Claude Code Hook Behavior

### Exit Code Interpretation

| Exit Code | Meaning | Claude Code Behavior |
|-----------|---------|----------------------|
| `0` | Success | Hook completed successfully. Proceed with operation. |
| `1` | Error/Rejection | Generic error. Operation rejected (for PreToolUse) or logged (for PostToolUse). |
| `2` | Invalid Input | Schema/validation error. Operation blocked or logged. |
| `3+` | Reserved | Treated same as `1` (generic error). Avoid using. |

**Important**: Claude Code does **not** differentiate between exit codes `1`, `2`, `3`, etc. - all non-zero codes are treated as failures. The convention of using `2` for validation errors is for **human readability** in logs.

### Hook Type Behaviors

| Hook Type | Exit 0 | Exit Non-Zero | Can Block Operation? |
|-----------|--------|---------------|----------------------|
| **PreToolUse** | Allow tool execution | Block tool execution | ✅ YES |
| **PostToolUse** | Success logged | Error logged | ❌ NO (tool already ran) |
| **SessionStart** | Context injected | Error logged, session continues | ❌ NO |
| **SessionStop** | Success logged | Error logged | ❌ NO |
| **UserPromptSubmit** | Message injected | Error logged | ❌ NO (informational only) |

**Critical Insight**: Only **PreToolUse** hooks can prevent operations. All other hooks are **informational** - they add context or log results but cannot block actions.

---

## Hook Timeout Behavior

**Timeout**: All hooks have a **60-second execution timeout** (per ADR-0002).

### What Happens on Timeout

1. **Process Killed**: Hook process receives SIGTERM, then SIGKILL if not terminated
2. **Treated as Failure**: Timeout = non-zero exit code
3. **PreToolUse**: Operation **blocked** (same as exit 1)
4. **PostToolUse**: Error logged, operation already completed
5. **SessionStart**: Context injection **skipped**, session continues

### Timeout Prevention Strategies

```bash
#!/bin/bash
# Example: Early bailout for non-critical operations

set -euo pipefail

# Fast path: Skip expensive checks if condition met
[[ "$SIMPLE_CASE" == "true" ]] && exit 0

# Timeout wrapper for external command (30s max)
timeout 30s expensive_validation.sh || {
    echo "Validation timed out, allowing operation" >&2
    exit 0  # Graceful degradation
}

# Rest of hook logic
```

**Best Practice**: Design hooks to **fail open** (exit 0) on timeout for non-critical validations. Fail closed (exit 1) only for critical security/data integrity checks.

---

## Per-Hook-Type Standards

### PreToolUse Hooks

**Purpose**: Validate and optionally block tool execution before it runs.

**Exit Code Standards**:

| Exit Code | Use Case | Example |
|-----------|----------|---------|
| `0` | Tool execution allowed | Safe command, validation passed |
| `1` | Tool execution blocked (generic error) | Team pack not found, session conflict |
| `2` | Tool execution blocked (validation failure) | Invalid YAML, missing required field |

**Example: session-write-guard.sh**
```bash
#!/bin/bash
# PreToolUse hook - blocks direct writes to *_CONTEXT.md

set -euo pipefail

TOOL_NAME="${CLAUDE_HOOK_TOOL_NAME:-}"
FILE_PATH="${CLAUDE_HOOK_FILE_PATH:-}"

# Allow non-Write/Edit operations
[[ "$TOOL_NAME" != "Write" && "$TOOL_NAME" != "Edit" ]] && exit 0

# Allow non-context files
[[ ! "$FILE_PATH" =~ _CONTEXT\.md$ ]] && exit 0

# BLOCK: Direct write to context file
cat <<'EOF'
{
  "decision": "block",
  "reason": "Direct writes to *_CONTEXT.md files are blocked.",
  "instruction": "Use state-mate agent for mutations"
}
EOF

exit 1  # Block the Write/Edit operation
```

**Output Format**: PreToolUse hooks can output JSON with `decision: block` or `decision: allow`. Exit code takes precedence - a hook that outputs `decision: allow` but exits with `1` will still **block** the operation.

**Common Validations**:
- Team pack existence (`exit 2` if not found)
- Workflow schema validation (`exit 2` for invalid YAML)
- Session state conflicts (`exit 1` for active session)
- Security checks (`exit 1` for blocked operation)

### PostToolUse Hooks

**Purpose**: Audit, log, or track operations after they complete. **Cannot block**.

**Exit Code Standards**:

| Exit Code | Use Case | Example |
|-----------|----------|---------|
| `0` | Audit/logging succeeded | Session mutation logged successfully |
| `1` | Audit/logging failed | Log file write failed (non-critical) |

**Example: session-audit.sh**
```bash
#!/bin/bash
# PostToolUse hook - audit SESSION_CONTEXT mutations

set -euo pipefail

FILE_PATH="${CLAUDE_HOOK_FILE_PATH:-}"

# Only audit session files
[[ ! "$FILE_PATH" =~ ^\.sos/sessions/session-.* ]] && exit 0

# Extract session ID, log mutation
SESSION_ID=$(echo "$FILE_PATH" | grep -o 'session-[^/]*' | head -1)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
echo "$TIMESTAMP | $SESSION_ID | WRITE | file=$(basename "$FILE_PATH")" >> .sos/sessions/.audit/mutations.log

# Always exit 0 - don't block on logging failure
exit 0
```

**Best Practice**: PostToolUse hooks should **always exit 0** unless logging is critical. Failures here are purely informational - the tool already executed.

**Common Use Cases**:
- Audit trail logging (`exit 0` even on log failure)
- Artifact tracking (`exit 0` if tracking optional)
- Metric collection (`exit 0` always)
- Integrity validation (`exit 1` to flag corruption, but operation already completed)

### SessionStart Hooks

**Purpose**: Inject project/session context when Claude starts. **Cannot block session start**.

**Exit Code Standards**:

| Exit Code | Use Case | Example |
|-----------|----------|---------|
| `0` | Context injected successfully | Session context displayed |
| `1` | Context injection failed | Session utilities unavailable, fallback used |

**Example: session-context.sh**
```bash
#!/bin/bash
# SessionStart hook - inject session context

set -euo pipefail

PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"
cd "$PROJECT_DIR" 2>/dev/null || exit 0

# Library Resolution - graceful fallback
HOOKS_LIB="${CLAUDE_PROJECT_DIR:-.}/.claude/hooks/lib"
source "$HOOKS_LIB/session-utils.sh" 2>/dev/null || {
    echo "## Session Context (fallback mode)"
    echo "- Session utilities not initialized"
    exit 0  # Graceful degradation
}

# Get session status
SESSION_JSON=$(session-manager.sh status 2>/dev/null || echo '{}')
HAS_SESSION=$(echo "$SESSION_JSON" | jq -r '.has_session // false')

# Output context
if [[ "$HAS_SESSION" == "true" ]]; then
    echo "## Session Context"
    echo "| Team | $(echo "$SESSION_JSON" | jq -r '.active_rite') |"
    echo "| Session | $(echo "$SESSION_JSON" | jq -r '.session_id') |"
else
    echo "## Session Context"
    echo "No active session. Use \`/start\` to begin."
fi

exit 0
```

**Best Practice**: Always implement **graceful fallback**. If libraries unavailable, output minimal context and exit 0. Never block session start.

**Failure Impact**: If SessionStart hook exits non-zero, context is **not injected** but Claude session **still starts**. User sees no context but can proceed.

### SessionStop Hooks

**Purpose**: Cleanup, auto-save, or finalization when Claude session ends. **Cannot block stop**.

**Exit Code Standards**:

| Exit Code | Use Case | Example |
|-----------|----------|---------|
| `0` | Cleanup succeeded | Session auto-parked successfully |
| `1` | Cleanup failed | Auto-park failed (non-critical) |

**Example: auto-park.sh**
```bash
#!/bin/bash
# SessionStop hook - auto-park active sessions

set -euo pipefail

HOOKS_LIB="${CLAUDE_PROJECT_DIR:-.}/.claude/hooks/lib"
source "$HOOKS_LIB/session-utils.sh" 2>/dev/null || exit 0

SESSION_ID=$(get_session_id)
[[ -z "$SESSION_ID" ]] && exit 0

# Auto-park if active
if is_session_active; then
    session-manager.sh park --reason "auto" 2>/dev/null || {
        echo "Auto-park failed" >&2
        exit 1  # Log failure, but session still stops
    }
fi

exit 0
```

**Best Practice**: Exit 0 unless critical cleanup failed. Session **always stops** regardless of exit code.

### UserPromptSubmit Hooks

**Purpose**: Inject preflight context or warnings before Claude processes user prompt. **Cannot block**.

**Exit Code Standards**:

| Exit Code | Use Case | Example |
|-----------|----------|---------|
| `0` | Preflight context injected | Warning about active session displayed |
| `1` | Preflight check failed | Session detection failed (non-critical) |

**Example: start-preflight.sh**
```bash
#!/bin/bash
# UserPromptSubmit hook - preflight checks for /start

set -euo pipefail

USER_PROMPT="${CLAUDE_USER_PROMPT:-}"

# Only act on /start commands
[[ ! "$USER_PROMPT" =~ ^/start ]] && exit 0

# Check for active session
HOOKS_LIB="${CLAUDE_PROJECT_DIR:-.}/.claude/hooks/lib"
source "$HOOKS_LIB/session-utils.sh" 2>/dev/null || exit 0

if has_active_session; then
    cat <<EOF

---
**Preflight Check**: Session already active

You have an active session. Consider:
- \`/park\` to pause current work
- \`/wrap\` to finalize session

---
EOF
fi

exit 0
```

**Best Practice**: Always exit 0. This hook is **informational only** - it adds context to the prompt but never blocks user input.

---

## Exit Code Conventions

### Standard Roster Conventions

| Exit Code | Semantic Meaning | When to Use |
|-----------|------------------|-------------|
| `0` | Success / Allow | Operation completed successfully, or validation passed |
| `1` | Generic Error / Reject | Team not found, session conflict, blocked operation |
| `2` | Validation Failure | Schema error, missing required field, invalid YAML |

**Example Usage**:
```bash
# Exit 0: Allow operation
[[ "$COMMAND" =~ ^git[[:space:]]+status ]] && exit 0

# Exit 1: Block operation (generic error)
if [[ ! -d "rites/$TEAM" ]]; then
    echo "Team pack '$TEAM' not found" >&2
    exit 1
fi

# Exit 2: Block operation (validation error)
if ! yq eval "$WORKFLOW_FILE" >/dev/null 2>&1; then
    echo "Invalid YAML syntax in $WORKFLOW_FILE" >&2
    exit 2
fi
```

### CI/CD Exit Codes (Pre-Commit Hooks)

For **git pre-commit hooks** (not Claude Code hooks), additional conventions apply:

| Exit Code | Meaning | Git Behavior |
|-----------|---------|--------------|
| `0` | All checks passed | Commit allowed |
| `1` | Validation failed | Commit blocked (override with `--no-verify`) |
| `2` | Schema error | Commit blocked (should never occur with valid input) |

**Example: .githooks/pre-commit-orchestrator**
```bash
#!/bin/bash
# Git pre-commit hook - validate orchestrator.yaml

set -euo pipefail

# Validate YAML schema
if ! yq eval "$YAML_FILE" >/dev/null 2>&1; then
    echo "ERROR: Invalid YAML syntax" >&2
    exit 2  # Schema error
fi

# Validate required fields
if ! jq -e '.routing' "$YAML_FILE" >/dev/null 2>&1; then
    echo "ERROR: Missing required field: routing" >&2
    exit 1  # Validation error
fi

echo "OK: Pre-commit validation passed"
exit 0  # Allow commit
```

---

## Hook Error Handling Patterns

### Pattern 1: Graceful Degradation

**When**: Hook provides non-critical functionality (logging, metrics, optional context)

```bash
#!/bin/bash
set -euo pipefail

# Attempt operation, but don't fail if unavailable
source "$HOOKS_LIB/session-utils.sh" 2>/dev/null || {
    echo "Session utilities unavailable, skipping" >&2
    exit 0  # Graceful degradation
}

# Continue with enhanced functionality
```

**Result**: Hook succeeds even when dependencies unavailable. User gets reduced functionality but no blocking errors.

### Pattern 2: Fail Closed (Security)

**When**: Hook enforces critical security or data integrity rules

```bash
#!/bin/bash
set -euo pipefail

# Critical validation - fail if cannot verify
if ! verify_signature "$FILE"; then
    echo "ERROR: Signature verification failed" >&2
    exit 1  # Block operation
fi

# Only proceed if verification succeeded
exit 0
```

**Result**: Operation blocked if validation fails. Security/integrity preserved.

### Pattern 3: Fail Open (Availability)

**When**: Hook provides optional guardrails but shouldn't break workflows

```bash
#!/bin/bash
set -euo pipefail

# Optional validation - allow on failure
if ! timeout 5s validate_workflow.sh 2>/dev/null; then
    echo "WARNING: Workflow validation failed, allowing operation" >&2
    exit 0  # Fail open - availability over validation
fi

exit 0
```

**Result**: Operation proceeds even if validation fails/times out. Availability prioritized.

### Pattern 4: Fast Path Optimization

**When**: Hook validates many operations but most are safe (command-validator.sh pattern)

```bash
#!/bin/bash
set -euo pipefail

auto_approve() {
    local reason="$1"
    cat <<EOF
{
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "allow",
    "permissionDecisionReason": "$reason"
  }
}
EOF
    exit 0
}

# FAST PATH: Auto-approve common safe operations
[[ "$COMMAND" =~ ^git[[:space:]]+status ]] && auto_approve "Safe git read"
[[ "$COMMAND" =~ ^ls[[:space:]] ]] && auto_approve "Safe ls"

# SLOW PATH: Complex validation for rare operations
validate_complex_operation "$COMMAND"
```

**Result**: Common operations exit immediately (< 1ms). Only rare operations trigger expensive validation. Keeps hook overhead minimal.

---

## Debugging Hook Exit Codes

### Enable Hook Logging

All roster hooks use the `logging.sh` library. Enable debug logging:

```bash
# In hook script
HOOKS_LIB="${CLAUDE_PROJECT_DIR:-.}/.claude/hooks/lib"
source "$HOOKS_LIB/logging.sh" 2>/dev/null && log_init "hook-name" || true

# Logs written to: .claude/hooks/.logs/<hook-name>.log
```

**View logs**:
```bash
tail -f .claude/hooks/.logs/session-write-guard.log
tail -f .claude/hooks/.logs/command-validator.log
```

### Test Hooks in Isolation

Run hooks manually to test exit codes:

```bash
# Test PreToolUse hook with sample input
echo '{"tool_name": "Write", "tool_input": {"file_path": "SESSION_CONTEXT.md"}}' | \
  CLAUDE_HOOK_TOOL_NAME=Write \
  CLAUDE_HOOK_FILE_PATH=SESSION_CONTEXT.md \
  CLAUDE_PROJECT_DIR=/Users/tomtenuta/Code/roster \
  .claude/hooks/session-write-guard.sh

echo "Exit code: $?"  # Should be 1 (blocked)

# Test SessionStart hook
CLAUDE_PROJECT_DIR=/Users/tomtenuta/Code/roster \
  .claude/hooks/session-context.sh

echo "Exit code: $?"  # Should be 0
```

### Common Failure Modes

| Symptom | Likely Cause | Debug Approach |
|---------|--------------|----------------|
| Hook always blocks | Exit code never 0 | Check fast path logic, ensure early `exit 0` for common cases |
| Hook always allows | Exit code never 1 | Check validation logic, ensure `exit 1` on failure conditions |
| Hook times out | Infinite loop or slow operation | Add `timeout` wrappers, implement early bailout |
| Hook crashes | Missing library, `set -e` triggered | Check library resolution, add `2>/dev/null` to sourcing |
| Hook output not visible | Exit 1 before output | Move output **before** `exit 1`, ensure flushed to stdout |

---

## Examples by Scenario

### Scenario: Block Invalid Team Swap

**Hook**: PreToolUse (`command-validator.sh`)

```bash
#!/bin/bash
set -euo pipefail

COMMAND="${CLAUDE_HOOK_COMMAND:-}"

# Extract team from swap-rite.sh command
if [[ "$COMMAND" =~ swap-rite\.sh[[:space:]]+([a-z0-9-]+-pack) ]]; then
    TARGET_TEAM="${BASH_REMATCH[1]}"

    # Validate team exists
    if [[ ! -d "rites/$TARGET_TEAM" ]]; then
        echo "Team pack '$TARGET_TEAM' not found" >&2
        echo "Available teams:" >&2
        ls -1 rites/ | sed 's/^/  - /' >&2
        exit 1  # Block operation
    fi
fi

exit 0  # Allow operation
```

**Expected**:
- Valid team: Exit 0, operation proceeds
- Invalid team: Exit 1, operation blocked, user sees error message

### Scenario: Log Artifact Creation

**Hook**: PostToolUse (`artifact-tracker.sh`)

```bash
#!/bin/bash
set -euo pipefail

FILE_PATH="${CLAUDE_HOOK_FILE_PATH:-}"
TOOL_NAME="${CLAUDE_HOOK_TOOL_NAME:-}"

# Only track Write operations to docs/
[[ "$TOOL_NAME" != "Write" ]] && exit 0
[[ ! "$FILE_PATH" =~ ^docs/(requirements|design)/ ]] && exit 0

# Extract artifact type
ARTIFACT_TYPE=$(basename "$(dirname "$FILE_PATH")")
ARTIFACT_NAME=$(basename "$FILE_PATH")

# Log to artifact tracker
echo "$(date -u +"%Y-%m-%dT%H:%M:%SZ") | $ARTIFACT_TYPE | $ARTIFACT_NAME" \
  >> .claude/artifacts/.tracker.log

# Always exit 0 - logging is non-critical
exit 0
```

**Expected**:
- Artifact created: Exit 0, logged to tracker
- Log write fails: Exit 0, operation succeeded (tool already ran)

### Scenario: Inject Session Context

**Hook**: SessionStart (`session-context.sh`)

```bash
#!/bin/bash
set -euo pipefail

# Graceful fallback if libraries unavailable
HOOKS_LIB="${CLAUDE_PROJECT_DIR:-.}/.claude/hooks/lib"
source "$HOOKS_LIB/session-utils.sh" 2>/dev/null || {
    echo "## Session Context (minimal)"
    echo "- Status: Session utilities unavailable"
    exit 0  # Graceful degradation
}

# Get session status
SESSION_ID=$(get_session_id)

if [[ -n "$SESSION_ID" ]]; then
    echo "## Session Context"
    echo "| Session | $SESSION_ID |"
    echo "| Status | ACTIVE |"
else
    echo "## Session Context"
    echo "No active session. Use \`/start\` to begin."
fi

exit 0
```

**Expected**:
- Libraries available: Exit 0, full context displayed
- Libraries missing: Exit 0, minimal context displayed
- Never blocks session start

---

## Related Documentation

- **ADR-0002**: Hook Library Resolution Architecture
- **docs/ORCHESTRATOR-CI-IMPLEMENTATION.md**: CI/CD validation and git pre-commit hooks
- **user-hooks/*/**: Reference implementations for all hook types
- **Claude Code Hooks Documentation**: https://code.claude.com/docs/en/hooks

---

## Quick Reference

### Exit Code Decision Tree

```
Is this a PreToolUse hook?
├─ YES: Can block operations
│   ├─ Validation passed? → exit 0 (allow)
│   ├─ Generic error? → exit 1 (block)
│   └─ Validation failed? → exit 2 (block)
└─ NO: Cannot block (informational only)
    ├─ Operation succeeded? → exit 0
    ├─ Non-critical failure? → exit 0 (graceful degradation)
    └─ Critical failure? → exit 1 (log error, operation already completed)
```

### Hook Type Matrix

| Hook Type | Blocks? | Timeout Effect | Best Practice |
|-----------|---------|----------------|---------------|
| PreToolUse | ✅ YES | Blocks operation | Fail closed for security, fail open for availability |
| PostToolUse | ❌ NO | Logs error | Always exit 0 unless critical |
| SessionStart | ❌ NO | Skips context | Graceful fallback always |
| SessionStop | ❌ NO | Logs error | Exit 0 unless critical cleanup |
| UserPromptSubmit | ❌ NO | Skips message | Always exit 0 (informational) |

### Common Patterns

```bash
# Pattern: Fast path (auto-approve)
[[ "$SAFE_CONDITION" ]] && exit 0

# Pattern: Graceful degradation
source "$LIB" 2>/dev/null || { echo "Fallback"; exit 0; }

# Pattern: Fail closed (security)
verify_security || exit 1

# Pattern: Fail open (availability)
timeout 5s validate || { echo "WARN: Allowing"; exit 0; }
```

---

**File Path**: `/Users/tomtenuta/Code/roster/docs/reference/hook-exit-codes.md`

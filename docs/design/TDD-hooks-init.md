# TDD: hooks-init.sh Initialization Pattern (Sprint 002 - Task 001)

## Overview

This Technical Design Document specifies the `hooks-init.sh` initialization script that standardizes hook setup, error handling, and shared utilities across the roster hook ecosystem. The design resolves the fundamental conflict between fail-fast (`set -euo pipefail`) and defensive (no set -e) error handling strategies by categorizing hooks by failure tolerance requirements.

## Context

| Reference | Location |
|-----------|----------|
| Sprint | sprint-002 "Hooks Standardization" |
| Session | session-20260101-195557-1b27d7dc |
| Existing Primitives | `/Users/tomtenuta/Code/roster/user-hooks/lib/primitives.sh` |
| Existing Config | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/config.sh` |
| Existing Logging | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/logging.sh` |
| Hook Registration | `/Users/tomtenuta/Code/roster/user-hooks/base_hooks.yaml` |

### Problem Statement

Current hooks exhibit **conflicting error handling strategies**:

1. **session-context.sh** (line 6): Uses `set -euo pipefail`
2. **command-validator.sh** (lines 5-6): Explicitly forbids this:
   ```
   # CRITICAL: NO set -e, NO set -euo pipefail - hooks must NEVER crash
   # CRITICAL: NO sourcing session-utils.sh - it can fail and crash the hook
   ```

Both approaches have merit for their specific contexts:
- **Fail-fast** catches errors early, prevents silent corruption
- **Defensive** ensures hooks never break Claude's tool flow

Without a principled resolution, hook authors make ad-hoc decisions, leading to inconsistent behavior and unpredictable failures.

### Design Goals

1. Resolve error handling conflict with principled categorization
2. Create unified initialization script (`hooks-init.sh`) for all hooks
3. Standardize library sourcing, logging, and error recovery
4. Maintain backward compatibility with existing hook implementations
5. Document which hooks belong to which category and why

---

## Design Decisions

### Decision 1: Hook Failure Tolerance Categories

**Options Considered**:

| Option | Description | Pros | Cons |
|--------|-------------|------|------|
| A. Uniform fail-fast | All hooks use `set -euo pipefail` | Simple, catches errors | PreToolUse crashes break Claude's flow |
| B. Uniform defensive | No hook uses `set -e` | Never crashes | Silent failures, stale context possible |
| C. Event-based categories | Different strategies by event type | Matches failure impact | More complex |
| D. Consequence-based categories | Strategy based on failure consequence | Precise matching | Requires per-hook analysis |

**Selected**: Option D - Consequence-based categories

**Rationale**: Event type alone is insufficient. For example, PostToolUse hooks like `artifact-tracker.sh` can fail silently (tracking loss is recoverable), but `session-audit.sh` failing silently could mean undetected corruption. The failure *consequence* determines the appropriate strategy:

- **Critical-path hooks**: Failures that leave Claude in an inconsistent state or break tool flow must be defended against.
- **Best-effort hooks**: Failures that lose non-critical data or tracking can fail gracefully.

### Decision 2: Two-Tier Error Handling Model

**Tier 1: DEFENSIVE (Must Never Crash)**
- No `set -e`, no `set -u`, no `set -o pipefail`
- Every external command wrapped with `|| true` or explicit error handling
- Return 0 on any internal error (log error, continue)
- PreToolUse hooks that return non-zero can block Claude's tools

**Tier 2: RECOVERABLE (Fail-Fast with Recovery)**
- Uses `set -euo pipefail` for early error detection
- Wraps entire script in error trap that logs and exits 0
- SessionStart hooks that fail may produce stale context but don't break Claude

**Rationale for Two Tiers**:
- Three tiers (adding "strict fail-fast") was considered but rejected: no current hook should crash Claude entirely.
- The distinction is between "must not affect tool availability" (DEFENSIVE) vs. "can detect errors but must degrade gracefully" (RECOVERABLE).

### Decision 3: hooks-init.sh API Surface

**Options Considered**:

| Option | Description | Pros | Cons |
|--------|-------------|------|------|
| A. Monolithic init | One function: `hooks_init "hook-name" "category"` | Simple API | Less flexible |
| B. Modular functions | Separate: `hooks_set_error_mode`, `hooks_init_logging`, etc. | Flexible | More boilerplate |
| C. Declarative header | Hook declares category in comment, init reads it | Self-documenting | Magic comments are fragile |

**Selected**: Option A - Monolithic init with category parameter

**Rationale**: Single function call reduces boilerplate and ensures consistent initialization order. Category is an explicit parameter, not magic comment.

```bash
source "$HOOKS_LIB/hooks-init.sh"
hooks_init "session-context" "RECOVERABLE"
```

### Decision 4: Integration with Existing Libraries

**Selected Architecture**:

```
hooks-init.sh
    |
    +-- sources config.sh (paths, timeouts, patterns)
    |
    +-- sources logging.sh (log_init, log_start, etc.)
    |
    +-- sources primitives.sh (md5_portable, get_yaml_field, etc.)
    |
    +-- provides hooks_init() function
    |
    +-- provides error mode setup based on category
    |
    +-- provides safe_source() for optional dependencies
```

**Rationale**: Hooks currently source these libraries individually with inconsistent patterns. `hooks-init.sh` becomes the single entry point that orchestrates correct sourcing order and error handling.

### Decision 5: Backward Compatibility Strategy

**Approach**: Incremental adoption with deprecation warnings

- Phase 1: Create `hooks-init.sh`, document new pattern
- Phase 2: Migrate hooks to new pattern (task-002 through task-010)
- Phase 3: Add deprecation warnings to direct library sourcing
- Phase 4: (Future sprint) Remove direct sourcing support

**Rationale**: Existing hooks must continue to work during migration. No breaking changes in this sprint.

---

## Error Handling Strategy

### Category Assignment Criteria

| Criterion | DEFENSIVE | RECOVERABLE |
|-----------|-----------|-------------|
| Non-zero exit blocks Claude tool? | Yes | No |
| Failure leaves stale/corrupt state? | No (ephemeral) | Acceptable (logged) |
| External command dependencies | Many, any can fail | Few, controlled |
| Output format | JSON (Claude parses) | Markdown (display only) |

### Hook Category Assignments

| Hook | Event | Category | Rationale |
|------|-------|----------|-----------|
| **command-validator.sh** | PreToolUse | DEFENSIVE | Non-zero exit blocks Bash tool; auto-approves safe commands |
| **session-write-guard.sh** | PreToolUse | DEFENSIVE | Non-zero exit blocks Write/Edit; must return valid JSON |
| **delegation-check.sh** | PreToolUse | DEFENSIVE | Non-zero exit blocks Edit/Write; warning-only, never blocks |
| **session-context.sh** | SessionStart | RECOVERABLE | Failure shows stale context; error trap logs and exits 0 |
| **coach-mode.sh** | SessionStart | RECOVERABLE | Failure omits reminder; non-critical |
| **start-preflight.sh** | UserPromptSubmit | RECOVERABLE | Failure shows no preflight; non-critical guidance |
| **auto-park.sh** | Stop | RECOVERABLE | Failure means no auto-park; logged, retryable |
| **artifact-tracker.sh** | PostToolUse | RECOVERABLE | Failure loses tracking; logged, non-critical |
| **commit-tracker.sh** | PostToolUse | RECOVERABLE | Failure loses tracking; logged, non-critical |
| **session-audit.sh** | PostToolUse | RECOVERABLE | Failure loses audit entry; logged, non-critical |

### Error Handling Patterns by Category

#### DEFENSIVE Pattern

```bash
#!/bin/bash
# Hook: command-validator.sh
# Category: DEFENSIVE - must never crash, must never return non-zero unexpectedly
#
# CRITICAL: NO set -e, NO set -euo pipefail

source "$HOOKS_LIB/hooks-init.sh" 2>/dev/null || {
    # Absolute fallback if init itself fails
    exit 0
}
hooks_init "command-validator" "DEFENSIVE"

# All commands use || patterns
INPUT=$(cat 2>/dev/null) || INPUT=""
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // empty' 2>/dev/null) || COMMAND=""

# Empty = nothing to validate
[[ -z "$COMMAND" ]] && exit 0

# ... validation logic with explicit || true on all externals
```

#### RECOVERABLE Pattern

```bash
#!/bin/bash
# Hook: session-context.sh
# Category: RECOVERABLE - can detect errors, must degrade gracefully

source "$HOOKS_LIB/hooks-init.sh"
hooks_init "session-context" "RECOVERABLE"

# Now safe to use pipefail - error trap will catch and exit 0
# ... main logic using standard bash patterns
```

---

## hooks-init.sh API Design

### Function: `hooks_init`

```bash
# Initialize hook with appropriate error handling mode
# Usage: hooks_init <hook_name> <category>
# Categories: DEFENSIVE | RECOVERABLE
#
# Effects:
#   - Sets HOOK_NAME and HOOK_CATEGORY globals
#   - Configures shell options based on category
#   - Initializes logging
#   - Sources required libraries
#   - Sets up error trap (RECOVERABLE only)
#
# Returns: 0 always (errors are logged, not propagated)

hooks_init() {
    local hook_name="${1:-unknown}"
    local category="${2:-RECOVERABLE}"

    # Export for use in error messages
    export HOOK_NAME="$hook_name"
    export HOOK_CATEGORY="$category"

    # Source dependencies (order matters)
    source "$(dirname "${BASH_SOURCE[0]}")/config.sh" 2>/dev/null || true
    source "$(dirname "${BASH_SOURCE[0]}")/logging.sh" 2>/dev/null || true
    source "$(dirname "${BASH_SOURCE[0]}")/primitives.sh" 2>/dev/null || true

    # Initialize logging
    log_init "$hook_name" 2>/dev/null || true
    log_start 2>/dev/null || true

    # Category-specific setup
    case "$category" in
        DEFENSIVE)
            # Explicitly disable strict modes
            set +e +u +o pipefail 2>/dev/null || true
            ;;
        RECOVERABLE)
            # Enable strict modes with recovery trap
            _hooks_setup_recovery_trap
            set -euo pipefail
            ;;
        *)
            # Unknown category defaults to DEFENSIVE
            log_warn "Unknown hook category: $category, defaulting to DEFENSIVE"
            set +e +u +o pipefail 2>/dev/null || true
            ;;
    esac

    return 0
}
```

### Function: `_hooks_setup_recovery_trap`

```bash
# Internal: Set up error trap for RECOVERABLE hooks
# Catches any error, logs it, and exits 0 to prevent hook from crashing Claude

_hooks_setup_recovery_trap() {
    trap '_hooks_handle_error $? $LINENO "$BASH_COMMAND"' ERR
}

_hooks_handle_error() {
    local exit_code="$1"
    local line_number="$2"
    local command="$3"

    # Log the error
    log_error "Hook failed at line $line_number: $command (exit $exit_code)" 2>/dev/null || true

    # Log completion with error
    log_end "$exit_code" 2>/dev/null || true

    # Exit 0 to prevent hook from blocking Claude
    exit 0
}
```

### Function: `safe_source`

```bash
# Safely source optional dependency with fallback
# Usage: safe_source <file_path> [fallback_action]
#
# Returns: 0 if sourced successfully, 1 if not (never crashes)

safe_source() {
    local file_path="$1"
    local fallback="${2:-}"

    if [[ -f "$file_path" ]]; then
        source "$file_path" 2>/dev/null
        return $?
    else
        if [[ -n "$fallback" ]]; then
            log_debug "Optional dependency not found: $file_path, using fallback"
            eval "$fallback" 2>/dev/null || true
        fi
        return 1
    fi
}
```

### Function: `hooks_finalize`

```bash
# Call at end of hook to log completion
# Usage: hooks_finalize [exit_code]
#
# Note: For DEFENSIVE hooks that explicitly manage their exit
# RECOVERABLE hooks auto-finalize via trap

hooks_finalize() {
    local exit_code="${1:-0}"
    log_end "$exit_code" 2>/dev/null || true
}
```

---

## Environment Variables

| Variable | Source | Description | Default |
|----------|--------|-------------|---------|
| `HOOK_NAME` | hooks_init | Current hook identifier | "unknown" |
| `HOOK_CATEGORY` | hooks_init | DEFENSIVE or RECOVERABLE | "RECOVERABLE" |
| `HOOKS_LIB` | config.sh | Path to hooks library directory | `$CLAUDE_PROJECT_DIR/.claude/hooks/lib` |
| `CLAUDE_PROJECT_DIR` | Claude Code | Project root directory | `.` |
| `HOOK_TIMEOUT` | config.sh | Default timeout in seconds | 5 |
| `CLAUDE_HOOK_DEBUG` | User | Enable debug logging if "1" | "0" |
| `ROSTER_VERBOSE` | User | Enable verbose output if "1" | "0" |

---

## Integration with Existing Infrastructure

### Primitives Reuse

The following functions from `primitives.sh` are available after `hooks_init`:

| Function | Purpose | Category Usage |
|----------|---------|----------------|
| `md5_portable` | Cross-platform MD5 hash | Both |
| `get_yaml_field` | Parse YAML frontmatter | Both |
| `atomic_write` | Safe file writes | Both |
| `json_extract` | Parse JSON with jq fallback | Both |
| `auto_approve` | PreToolUse permission response | DEFENSIVE |

### Logging Integration

After `hooks_init`, these logging functions are available:

| Function | Purpose | When to Use |
|----------|---------|-------------|
| `log_hook` | Log message with timestamp | General logging |
| `log_start` | Log hook start | Called by hooks_init |
| `log_end` | Log completion with exit code | Called by finalize/trap |
| `log_error` | Log error message | Error conditions |
| `log_warn` | Log warning message | Non-fatal issues |
| `log_debug` | Log debug (if CLAUDE_HOOK_DEBUG=1) | Verbose tracing |
| `get_time_ms` | Get current time in ms | Duration tracking |
| `calc_duration_ms` | Calculate elapsed time | Performance metrics |

### Config Integration

After `hooks_init`, these config values are available:

| Variable | Purpose |
|----------|---------|
| `CLAUDE_PROJECT_DIR` | Project root |
| `ROSTER_HOME` | Roster repository path |
| `SESSIONS_DIR` | Session storage path |
| `CURRENT_SESSION_FILE` | Active session pointer |
| `SAFE_READ_COMMANDS` | Auto-approve patterns |
| `SAFE_GIT_COMMANDS` | Auto-approve patterns |

---

## Migration Path

### Current State (Before Migration)

```bash
#!/bin/bash
# session-context.sh - current pattern

set -euo pipefail  # or NOT set, inconsistent

SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"
HOOKS_LIB="${CLAUDE_PROJECT_DIR:-.}/.claude/hooks/lib"
source "$HOOKS_LIB/logging.sh" 2>/dev/null && log_init "session-context" && log_start || true
source "$HOOKS_LIB/session-utils.sh" 2>/dev/null || { ... }
# ... hook logic
```

### Target State (After Migration)

```bash
#!/bin/bash
# session-context.sh - standardized pattern
# Category: RECOVERABLE

HOOKS_LIB="${CLAUDE_PROJECT_DIR:-.}/.claude/hooks/lib"
source "$HOOKS_LIB/hooks-init.sh"
hooks_init "session-context" "RECOVERABLE"

# Optional: source session-utils for session-specific functions
safe_source "$HOOKS_LIB/session-utils.sh"

# ... hook logic (now protected by error trap)
```

### Migration Steps Per Hook

1. Replace manual library sourcing with `source "$HOOKS_LIB/hooks-init.sh"`
2. Replace `set -euo pipefail` / no-set with `hooks_init "name" "CATEGORY"`
3. Replace manual log_init/log_start with hooks_init (does it internally)
4. Use `safe_source` for optional dependencies like session-utils.sh
5. Remove explicit error handling for RECOVERABLE hooks (trap handles it)
6. Keep explicit `|| true` patterns for DEFENSIVE hooks

### Backward Compatibility

During migration period:
- Hooks using old pattern continue to work unchanged
- New pattern is opt-in via sourcing hooks-init.sh
- No deprecation warnings in Sprint 002 (documentation only)

---

## File Structure

```
.claude/hooks/lib/
    |
    +-- config.sh           # Existing: paths, timeouts
    +-- logging.sh          # Existing: log functions
    +-- primitives.sh       # Existing: utilities
    +-- session-utils.sh    # Existing: session functions (shim)
    +-- session-core.sh     # Existing: core session ops
    +-- session-state.sh    # Existing: state queries
    |
    +-- hooks-init.sh       # NEW: unified initialization

user-hooks/
    +-- lib/
        +-- (symlinks to .claude/hooks/lib/)
    +-- context-injection/
        +-- session-context.sh    # RECOVERABLE
        +-- coach-mode.sh         # RECOVERABLE
    +-- validation/
        +-- command-validator.sh  # DEFENSIVE
        +-- delegation-check.sh   # DEFENSIVE
    +-- session-guards/
        +-- auto-park.sh          # RECOVERABLE
        +-- start-preflight.sh    # RECOVERABLE
        +-- session-write-guard.sh # DEFENSIVE
    +-- tracking/
        +-- artifact-tracker.sh   # RECOVERABLE
        +-- commit-tracker.sh     # RECOVERABLE
        +-- session-audit.sh      # RECOVERABLE
```

---

## Test Matrix

### Unit Tests for hooks-init.sh

| Test ID | Description | Expected Outcome |
|---------|-------------|------------------|
| init_001 | DEFENSIVE category disables strict modes | `set +o` shows errexit=off |
| init_002 | RECOVERABLE category enables strict modes | `set +o` shows errexit=on |
| init_003 | RECOVERABLE trap catches error | Exit 0 after intentional failure |
| init_004 | Unknown category defaults to DEFENSIVE | Warning logged, errexit=off |
| init_005 | Missing config.sh doesn't crash init | Init completes, uses defaults |
| init_006 | safe_source returns 1 for missing file | Return code is 1, no crash |
| init_007 | safe_source executes fallback | Fallback action runs |
| init_008 | hooks_finalize logs completion | Log entry with exit code |

### Integration Tests Per Hook Category

| Test ID | Category | Test | Expected |
|---------|----------|------|----------|
| cat_001 | DEFENSIVE | command-validator with malformed JSON | Exit 0, no output |
| cat_002 | DEFENSIVE | session-write-guard with missing env vars | Exit 0, valid JSON |
| cat_003 | RECOVERABLE | session-context with missing session-utils | Exit 0, fallback output |
| cat_004 | RECOVERABLE | artifact-tracker with jq unavailable | Exit 0, logged error |

---

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Trap doesn't catch all errors | Low | Medium | Extensive testing, fallback exit 0 |
| Library sourcing order issues | Low | High | Explicit order in hooks_init, tested |
| Performance overhead from init | Low | Low | Init is ~10ms, within timeout budgets |
| Migration breaks existing hooks | Low | High | Backward compatible, opt-in adoption |
| RECOVERABLE mask real bugs | Medium | Medium | Errors always logged, monitor logs |

---

## Success Criteria

- [ ] hooks-init.sh created and implements documented API
- [ ] All 10 hooks assigned to DEFENSIVE or RECOVERABLE category
- [ ] Category assignments documented with rationale
- [ ] Error trap catches and gracefully handles failures in RECOVERABLE hooks
- [ ] DEFENSIVE hooks explicitly handle all error paths
- [ ] Backward compatibility maintained (old hooks still work)
- [ ] Unit tests pass for hooks-init.sh functions
- [ ] Integration tests pass for hook categories

---

## Implementation Guidance

### Recommended Implementation Order

1. **Create hooks-init.sh** with `hooks_init`, `safe_source`, `hooks_finalize`
2. **Add unit tests** for hooks-init.sh functions
3. **Migrate one DEFENSIVE hook** (command-validator.sh) as reference
4. **Migrate one RECOVERABLE hook** (session-context.sh) as reference
5. **Document patterns** in hook header comments
6. **Migrate remaining hooks** (tasks 003-010)

### Code Style for hooks-init.sh

- Follow existing logging.sh patterns for consistency
- Use `local` for all function variables
- Export only HOOK_NAME, HOOK_CATEGORY
- All error handling must use `2>/dev/null || true` pattern
- Comments document "why" not "what"

---

## ADRs

No new ADRs required. This design implements standardization within the existing hook architecture established in ADR-0002.

---

## Artifact Attestation

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| This TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-hooks-init.md` | Created |
| primitives.sh | `/Users/tomtenuta/Code/roster/user-hooks/lib/primitives.sh` | Read |
| config.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/config.sh` | Read |
| logging.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/logging.sh` | Read |
| base_hooks.yaml | `/Users/tomtenuta/Code/roster/user-hooks/base_hooks.yaml` | Read |
| session-context.sh | `/Users/tomtenuta/Code/roster/user-hooks/context-injection/session-context.sh` | Read |
| command-validator.sh | `/Users/tomtenuta/Code/roster/user-hooks/validation/command-validator.sh` | Read |
| All 10 hooks | Various paths in user-hooks/ | Read |

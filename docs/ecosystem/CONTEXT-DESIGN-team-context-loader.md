---
title: "Context Design: Per-Team Hook Context Injection"
type: context-design
complexity: MODULE
created_at: "2026-01-02T16:00:00Z"
status: ready-for-implementation
gap_analysis: "Gap Analysis from ecosystem-analyst task-001 (session reference)"
affected_systems:
  - roster
  - skeleton
author: context-architect
backward_compatible: true
migration_required: false
work_packages:
  - id: WP1
    name: "Team Context Loader Library"
    description: "Create shell library that discovers and loads team-specific context injection scripts"
    files:
      - path: ".claude/hooks/lib/team-context-loader.sh"
        action: create
        description: "Core library with load_team_context() function"
    estimated_effort: "1 hour"
  - id: WP2
    name: "Session Context Hook Integration"
    description: "Integrate team context loader into existing session-context.sh hook"
    files:
      - path: ".claude/hooks/context-injection/session-context.sh"
        action: modify
        description: "Add team context section after session info output"
    dependencies: [WP1]
    estimated_effort: "30 minutes"
  - id: WP3
    name: "Ecosystem-Pack Context Script"
    description: "Create prototype context-injection script for ecosystem-pack team"
    files:
      - path: "teams/ecosystem-pack/context-injection.sh"
        action: create
        description: "Team-specific context: CEM sync, skeleton ref, drift status"
    dependencies: [WP1]
    estimated_effort: "1 hour"
schema_version: "1.0"
---

## Executive Summary

This Context Design addresses the gap identified in task-001: team-specific context is only available in verbose mode (session-context.sh lines 261-267). The solution uses a compose pattern where teams can optionally provide a `context-injection.sh` script that the SessionStart hook discovers and executes, injecting team-relevant context into every session start.

## Design Decisions

### Decision 1: Compose Pattern via Team Scripts

**Options Considered**:
1. **Hardcode all team context in session-context.sh** - Rejected: requires roster changes for every team, doesn't scale
2. **Dynamic detection via workflow.yaml parsing** - Rejected: complex, inflexible, can't customize output format
3. **Team-owned context-injection.sh scripts (Compose Pattern)** - Selected: teams own their context, no roster changes needed for new teams

**Selected**: Compose Pattern with team-owned scripts

**Rationale**:
- Teams know what context matters for their workflow
- Ecosystem-pack needs CEM sync status; 10x-dev-pack needs different context
- Adding new team context requires no roster changes
- Script-based approach allows dynamic content (git queries, file checks)
- Graceful degradation: teams without scripts get no extra context (not an error)

### Decision 2: Script Location in Team Directory

**Options Considered**:
1. **`teams/$TEAM/hooks/context-injection.sh`** - Rejected: too nested, hooks/ already has .gitkeep placeholder
2. **`teams/$TEAM/context.sh`** - Rejected: ambiguous name, could be config
3. **`teams/$TEAM/context-injection.sh`** - Selected: clear purpose, flat in team dir

**Selected**: `teams/$TEAM/context-injection.sh`

**Rationale**:
- Direct sibling to workflow.yaml and agents/
- Name clearly indicates purpose (injection into session context)
- Simple discovery: `$ROSTER_HOME/teams/$ACTIVE_RITE/context-injection.sh`

### Decision 3: Script Interface Contract

**Options Considered**:
1. **Source script and call function** - Selected: explicit contract, named function documents purpose
2. **Execute script and capture stdout** - Rejected: no error handling, harder to debug
3. **Source script with side effects** - Rejected: implicit behavior, maintenance nightmare

**Selected**: Source script, call `inject_team_context` function

**Rationale**:
- Function name is self-documenting
- Caller controls when output happens
- Function can return non-zero to indicate partial failure (logged, not fatal)
- Allows team script to access hook library functions if sourced

### Decision 4: Output Placement in Session Context

**Options Considered**:
1. **Always show team context (condensed + verbose)** - Selected: team context is always relevant when team is active
2. **Only in verbose mode** - Rejected: defeats purpose, that's the current broken state
3. **Separate hook entirely** - Rejected: adds hook latency, fragmentscontext

**Selected**: Always show team context in condensed mode when team is active

**Rationale**:
- If a team is active, its context matters
- Teams can decide what to include (keep it concise)
- User can still use `--verbose` for full detail
- Maintains single "Session Context" section

### Decision 5: Error Handling Strategy

**Pattern**: RECOVERABLE category per ADR-0002

- Team script not found: Silent (not all teams need context)
- Team script exists but `inject_team_context` function missing: Log warning, continue
- Function returns non-zero: Log warning, continue (partial failure)
- Function outputs nothing: Normal (team may have nothing to report)

**Rationale**: Hook must never fail and block Claude. Team context is enhancement, not requirement.

---

## Team Context Loader Library Specification

### File: `.claude/hooks/lib/team-context-loader.sh`

```bash
#!/bin/bash
# Team Context Loader - discovers and executes team-specific context injection
# Part of Per-Team Hook Context Injection feature
#
# Usage:
#   source "$HOOKS_LIB/team-context-loader.sh"
#   output=$(load_team_context)
#   [[ -n "$output" ]] && echo "$output"

# =============================================================================
# Configuration
# =============================================================================

# Team context script name (convention)
readonly TEAM_CONTEXT_SCRIPT_NAME="context-injection.sh"

# Function name teams must export
readonly TEAM_CONTEXT_FUNCTION_NAME="inject_team_context"

# =============================================================================
# Main Function
# =============================================================================

# Load team-specific context if available
# Arguments: None (uses ACTIVE_RITE file and ROSTER_HOME)
# Output: Markdown content to stdout (may be empty)
# Returns: 0 always (errors logged, not propagated)
#
# Contract:
#   - Reads ACTIVE_RITE from .claude/ACTIVE_RITE
#   - Looks for $ROSTER_HOME/teams/$ACTIVE_RITE/context-injection.sh
#   - Sources script and calls inject_team_context()
#   - Returns function output on stdout
#   - Never fails (RECOVERABLE pattern)

load_team_context() {
    local active_team
    local team_script
    local output=""

    # Read active team
    active_team=$(cat ".claude/ACTIVE_RITE" 2>/dev/null || echo "")
    if [[ -z "$active_team" || "$active_team" == "none" ]]; then
        # No team active - nothing to inject
        return 0
    fi

    # Resolve team context script path
    local roster_home="${ROSTER_HOME:-$HOME/Code/roster}"
    team_script="$roster_home/teams/$active_team/$TEAM_CONTEXT_SCRIPT_NAME"

    # Check if team has context script
    if [[ ! -f "$team_script" ]]; then
        # Team has no context script - normal, not an error
        log_debug "Team $active_team has no context script at $team_script" 2>/dev/null || true
        return 0
    fi

    # Check if script is executable (warning if not)
    if [[ ! -x "$team_script" ]]; then
        log_warning "Team context script exists but not executable: $team_script" 2>/dev/null || true
        # Try sourcing anyway - bash doesn't require +x for sourcing
    fi

    # Source the team script
    # Use subshell to isolate any side effects
    output=$(
        # Source team script
        source "$team_script" 2>/dev/null || {
            log_warning "Failed to source team context script: $team_script" 2>/dev/null || true
            exit 0
        }

        # Check if function exists
        if ! declare -f "$TEAM_CONTEXT_FUNCTION_NAME" >/dev/null 2>&1; then
            log_warning "Team context script missing function: $TEAM_CONTEXT_FUNCTION_NAME" 2>/dev/null || true
            exit 0
        fi

        # Call the function
        "$TEAM_CONTEXT_FUNCTION_NAME" 2>/dev/null || {
            log_warning "$TEAM_CONTEXT_FUNCTION_NAME returned non-zero" 2>/dev/null || true
        }
    )

    # Output result (may be empty)
    echo "$output"
    return 0
}

# =============================================================================
# Utility Functions for Team Scripts
# =============================================================================

# Teams can use these helpers in their context-injection.sh

# Format a key-value pair for team context table
# Usage: team_context_row "Key" "Value"
team_context_row() {
    local key="$1"
    local value="$2"
    echo "| **$key** | $value |"
}

# Check if a file is newer than N minutes
# Usage: is_file_stale "/path/to/file" 60  # true if older than 60 minutes
is_file_stale() {
    local file="$1"
    local max_age_minutes="${2:-60}"

    if [[ ! -f "$file" ]]; then
        return 0  # Non-existent = stale
    fi

    local now file_time age_seconds max_age_seconds
    now=$(date +%s)
    file_time=$(stat -f %m "$file" 2>/dev/null || stat -c %Y "$file" 2>/dev/null || echo 0)
    age_seconds=$((now - file_time))
    max_age_seconds=$((max_age_minutes * 60))

    [[ $age_seconds -gt $max_age_seconds ]]
}
```

### Integration Points

The library:
1. Sources via `$HOOKS_LIB/team-context-loader.sh`
2. Depends on `config.sh` for `ROSTER_HOME`
3. Depends on `logging.sh` for `log_debug`, `log_warning`
4. Provides utility functions teams can use

---

## Session Context Hook Changes

### File: `.claude/hooks/context-injection/session-context.sh`

**Current behavior** (lines 259-269):
```bash
# Team routing context (if team is active)
# Note: ROSTER_HOME is defined in config.sh (sourced via session-utils.sh)
if [[ -f ".claude/ACTIVE_RITE" ]]; then
    local TEAM_CONTEXT=$("$ROSTER_HOME/generate-team-context.sh" 2>/dev/null || echo "")
    if [[ -n "$TEAM_CONTEXT" ]]; then
        echo ""
        echo "$TEAM_CONTEXT"
    fi
fi
```

**Change 1**: Add source statement (after line 34, with other library sources)

```bash
# Source team context loader
safe_source "$HOOKS_LIB/team-context-loader.sh" || true
```

**Change 2**: Add team context to condensed output (in `output_condensed_context`, after commands line ~177)

```bash
# Team-specific context (if team has context script)
local TEAM_CONTEXT
TEAM_CONTEXT=$(load_team_context 2>/dev/null || echo "")
if [[ -n "$TEAM_CONTEXT" ]]; then
    echo ""
    echo "### Team Context"
    echo ""
    echo "$TEAM_CONTEXT"
fi
```

**Change 3**: Update verbose output to also use the loader (lines 259-269)

Replace the existing `generate-team-context.sh` call with:
```bash
# Team-specific context (compose pattern)
local TEAM_CONTEXT
TEAM_CONTEXT=$(load_team_context 2>/dev/null || echo "")
if [[ -n "$TEAM_CONTEXT" ]]; then
    echo ""
    echo "### Team Context"
    echo ""
    echo "$TEAM_CONTEXT"
fi

# Team routing context (workflow phases) - keep existing
if [[ -f ".claude/ACTIVE_RITE" ]]; then
    local ROUTING_CONTEXT=$("$ROSTER_HOME/generate-team-context.sh" 2>/dev/null || echo "")
    if [[ -n "$ROUTING_CONTEXT" ]]; then
        echo ""
        echo "$ROUTING_CONTEXT"
    fi
fi
```

**Rationale**: Keep both team-specific context (new, from context-injection.sh) and routing context (existing, from generate-team-context.sh). They serve different purposes:
- Team context: Status information specific to the team's domain
- Routing context: Workflow phases and agent routing table

---

## Ecosystem-Pack Context Script Specification

### File: `teams/ecosystem-pack/context-injection.sh`

```bash
#!/bin/bash
# Ecosystem-Pack Context Injection
# Provides CEM sync status, skeleton reference, and drift detection for ecosystem work
#
# Called by: session-context.sh via team-context-loader.sh
# Output: Markdown table with ecosystem status

# Required function name (per team-context-loader.sh contract)
inject_team_context() {
    local project_dir="${CLAUDE_PROJECT_DIR:-.}"
    local output=""

    # Start table
    output="| | |"$'\n'
    output+="|---|---|"$'\n'

    # CEM Sync Status
    local cem_sync_file="$project_dir/.claude/.cem-sync"
    local cem_status="unknown"
    local cem_timestamp="never"

    if [[ -f "$cem_sync_file" ]]; then
        cem_timestamp=$(stat -f "%Sm" -t "%Y-%m-%d %H:%M" "$cem_sync_file" 2>/dev/null || \
                        stat -c "%y" "$cem_sync_file" 2>/dev/null | cut -d'.' -f1 || \
                        echo "unknown")
        # Check staleness (>24h = stale)
        if is_file_stale "$cem_sync_file" 1440 2>/dev/null; then
            cem_status="stale"
        else
            cem_status="synced"
        fi
    else
        cem_status="never synced"
    fi
    output+="| **CEM Sync** | $cem_status ($cem_timestamp) |"$'\n'

    # Skeleton Reference
    local skeleton_ref="unknown"
    local skeleton_home="${SKELETON_HOME:-$HOME/Code/skeleton_claude}"
    if [[ -d "$skeleton_home/.git" ]]; then
        skeleton_ref=$(cd "$skeleton_home" && git rev-parse --short HEAD 2>/dev/null || echo "unknown")
        local skeleton_branch=$(cd "$skeleton_home" && git branch --show-current 2>/dev/null || echo "detached")
        skeleton_ref="$skeleton_branch@$skeleton_ref"
    fi
    output+="| **Skeleton Ref** | $skeleton_ref |"$'\n'

    # Drift Detection
    local drift_status="clean"
    # Check if local .claude/ differs from skeleton (simplified check)
    if [[ -f "$project_dir/.claude/.local-overrides" ]]; then
        local override_count=$(wc -l < "$project_dir/.claude/.local-overrides" 2>/dev/null | tr -d ' ')
        drift_status="$override_count local overrides"
    fi
    output+="| **Drift Status** | $drift_status |"$'\n'

    # Test Satellites (for compatibility testing context)
    local satellites_dir="${ROSTER_HOME:-$HOME/Code/roster}/test-satellites"
    local satellite_count=0
    if [[ -d "$satellites_dir" ]]; then
        satellite_count=$(ls -1d "$satellites_dir"/*/ 2>/dev/null | wc -l | tr -d ' ')
    fi
    output+="| **Test Satellites** | $satellite_count available |"$'\n'

    echo "$output"
}

# Helper function (provided by team-context-loader.sh, but define fallback)
if ! declare -f is_file_stale >/dev/null 2>&1; then
    is_file_stale() {
        local file="$1"
        local max_age_minutes="${2:-60}"
        [[ ! -f "$file" ]] && return 0
        local now file_time age_seconds max_age_seconds
        now=$(date +%s)
        file_time=$(stat -f %m "$file" 2>/dev/null || stat -c %Y "$file" 2>/dev/null || echo 0)
        age_seconds=$((now - file_time))
        max_age_seconds=$((max_age_minutes * 60))
        [[ $age_seconds -gt $max_age_seconds ]]
    }
fi
```

### Expected Output (Example)

When ecosystem-pack is active, session context will include:

```markdown
### Team Context

| | |
|---|---|
| **CEM Sync** | synced (2026-01-02 15:30) |
| **Skeleton Ref** | main@a1b2c3d |
| **Drift Status** | clean |
| **Test Satellites** | 3 available |
```

---

## Backward Compatibility

**Classification**: COMPATIBLE

This change is fully backward compatible:
- Teams without `context-injection.sh` see no change
- Existing `generate-team-context.sh` output is preserved (routing table)
- New team context appears as additional section
- No schema changes to SESSION_CONTEXT.md
- No changes to session-manager.sh

### Impact by Team Type

| Team | Impact |
|------|--------|
| ecosystem-pack | Gets new context section (CEM sync, skeleton ref) |
| 10x-dev-pack | No change (no context script, can add later) |
| Other teams | No change (can add context script when needed) |
| Satellites | No change (receive updated session-context.sh via CEM) |

---

## Test Matrix

| Scenario | Setup | Expected Outcome |
|----------|-------|------------------|
| Team with context script | ecosystem-pack active | Team Context section appears with CEM/skeleton status |
| Team without context script | 10x-dev-pack active | No Team Context section (normal) |
| No team active | ACTIVE_RITE = none | No Team Context section (normal) |
| Script exists but not executable | chmod -x context-injection.sh | Warning logged, graceful skip |
| Function missing from script | inject_team_context not defined | Warning logged, graceful skip |
| Function returns error | inject_team_context returns 1 | Warning logged, partial output shown |
| Function outputs nothing | Empty inject_team_context | No Team Context section (normal) |
| ROSTER_HOME not set | Env variable missing | Falls back to ~/Code/roster |
| Verbose mode | --verbose flag | Both Team Context and routing table shown |
| Condensed mode (default) | No flags | Team Context shown (new), routing table hidden |

---

## Handoff Criteria

- [x] Solution architecture documented with rationale
- [x] All design decisions have documented rationale
- [x] team-context-loader.sh fully specified with function signatures
- [x] session-context.sh changes specified at line level
- [x] ecosystem-pack/context-injection.sh template provided
- [x] Backward compatibility assessed: COMPATIBLE
- [x] No migration required
- [x] Test matrix complete with expected outcomes
- [x] No "TBD", "TODO", or "maybe" in document
- [x] Error handling strategy documented (RECOVERABLE pattern)

---

## Notes for Integration Engineer

### Implementation Order

1. Create `team-context-loader.sh` library first (WP1)
2. Test library in isolation: `source team-context-loader.sh; load_team_context`
3. Create ecosystem-pack context script (WP3)
4. Test script in isolation: `source context-injection.sh; inject_team_context`
5. Integrate into session-context.sh (WP2)
6. Test full flow: `/start` or new session

### Key Files Reference

| File | Purpose |
|------|---------|
| `/Users/tomtenuta/Code/roster/.claude/hooks/lib/team-context-loader.sh` | New library |
| `/Users/tomtenuta/Code/roster/.claude/hooks/context-injection/session-context.sh` | Hook to modify |
| `/Users/tomtenuta/Code/roster/.claude/hooks/lib/hooks-init.sh` | Pattern reference |
| `/Users/tomtenuta/Code/roster/.claude/hooks/lib/config.sh` | ROSTER_HOME definition |
| `/Users/tomtenuta/Code/roster/teams/ecosystem-pack/context-injection.sh` | New team script |
| `/Users/tomtenuta/Code/roster/generate-team-context.sh` | Existing routing generator (keep) |

### Testing Commands

```bash
# Test library function
cd /path/to/satellite
source .claude/hooks/lib/team-context-loader.sh
load_team_context

# Test ecosystem-pack script directly
source ~/Code/roster/teams/ecosystem-pack/context-injection.sh
inject_team_context

# Test full session context hook
ROSTER_VERBOSE=1 .claude/hooks/context-injection/session-context.sh
```

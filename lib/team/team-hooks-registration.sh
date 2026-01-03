#!/usr/bin/env bash
#
# team-hooks-registration.sh - Hook Registration for settings.local.json
#
# Parses hooks.yaml files and generates Claude Code hook registrations
# in settings.local.json while preserving user-defined hooks.
#
# Part of: roster team-swap infrastructure
#
# Usage:
#   source "$ROSTER_HOME/lib/team/team-hooks-registration.sh"
#   swap_hook_registrations "team-name"
#
# Dependencies:
#   - yq v4+ (for YAML parsing)
#   - jq (for JSON manipulation)
#   - Logging functions (log, log_debug, log_warning, log_error)
#
# Environment:
#   ROSTER_HOME - Path to roster installation
#   DRY_RUN_MODE - If set to 1, preview changes without writing

# Guard against re-sourcing
[[ -n "${_TEAM_HOOKS_REGISTRATION_LOADED:-}" ]] && return 0
readonly _TEAM_HOOKS_REGISTRATION_LOADED=1

# ============================================================================
# Logging Stubs (overridden when sourced from swap-team.sh)
# ============================================================================

# These stub implementations provide basic logging when team-hooks-registration.sh
# is used standalone (e.g., in unit tests). When sourced from swap-team.sh,
# these are overridden by the full logging implementation.

if ! type log >/dev/null 2>&1; then
    log() {
        echo "[Hook Registration] $*"
    }
fi

if ! type log_debug >/dev/null 2>&1; then
    log_debug() {
        echo "[DEBUG] $*" >&2
    }
fi

if ! type log_warning >/dev/null 2>&1; then
    log_warning() {
        echo "[WARNING] $*" >&2
    }
fi

if ! type log_error >/dev/null 2>&1; then
    log_error() {
        echo "[ERROR] $*" >&2
    }
fi

# ============================================================================
# Validation
# ============================================================================

# Check if yq v4+ is available
# Returns: 0 if yq v4+ available, 1 otherwise
# Side effects: Logs error if not available
require_yq() {
    # TODO: Implement in RF-018
    return 1
}

# ============================================================================
# YAML Parsing
# ============================================================================

# Parse hooks.yaml file and emit JSON-lines format
# Parameters:
#   $1 - yaml_file: Path to hooks.yaml file
# Output: One JSON object per line to stdout
#   Format: {"event":"...","matcher":"...","path":"...","timeout":N}
# Returns: 0 always (empty output for missing/invalid file)
# Side effects: Logs warnings for invalid entries
parse_hooks_yaml() {
    # TODO: Implement in RF-019
    return 0
}

# ============================================================================
# JSON Extraction
# ============================================================================

# Extract non-roster hooks from existing settings.local.json
# These are hooks whose command does NOT contain ".claude/hooks/"
# Parameters:
#   $1 - settings_file: Path to settings.local.json
# Output: JSON object with preserved hooks by event type to stdout
# Returns: 0 always (empty {} for missing file)
extract_non_roster_hooks() {
    # TODO: Implement in RF-020
    echo "{}"
    return 0
}

# ============================================================================
# Data Merge
# ============================================================================

# Merge hook registrations (base first, team appended)
# Parameters:
#   $1 - base_registrations: JSON-lines format (from base hooks)
#   $2 - team_registrations: JSON-lines format (from team hooks)
# Output: Combined JSON-lines to stdout (base first, then team)
# Returns: 0 always
merge_hook_registrations() {
    # TODO: Implement in RF-021
    return 0
}

# Merge generated hooks with preserved user hooks
# Parameters:
#   $1 - generated_json: Generated hooks JSON object
#   $2 - preserved_json: Preserved user hooks JSON object
# Output: Combined hooks JSON object to stdout
# Returns: 0 always
merge_with_preserved() {
    # TODO: Implement in RF-023
    return 0
}

# ============================================================================
# JSON Generation
# ============================================================================

# Generate Claude Code hooks JSON format from registrations
# Parameters:
#   $1 - registrations: JSON-lines format
# Output: Claude Code settings.local.json hooks object to stdout
# Returns: 0 always (empty {} for no registrations)
generate_hooks_json() {
    # TODO: Implement in RF-022
    echo "{}"
    return 0
}

# ============================================================================
# Main Orchestrator
# ============================================================================

# Sync hook registrations to settings.local.json
# Called after swap_hooks() syncs the actual hook files
# Parameters:
#   $1 - team_name: Name of team being activated
# Returns: 0 on success, 1 on error
# Side effects:
#   - Updates .claude/settings.local.json hooks section
#   - Preserves non-roster hooks in settings
#   - Creates settings.local.json if missing
#   - Backs up corrupted settings.local.json
# Environment:
#   ROSTER_HOME - Must be set
#   DRY_RUN_MODE - If 1, prints preview without writing
swap_hook_registrations() {
    # TODO: Implement in RF-024
    return 1
}

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
    if ! command -v yq &>/dev/null; then
        log_error "yq is required but not installed"
        log_error "Install with: brew install yq (macOS) or pip install yq"
        return 1
    fi

    # Check for yq v4+ (mikefarah/yq)
    local yq_version
    yq_version=$(yq --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+' | head -1)
    local major_version
    major_version=$(echo "$yq_version" | cut -d. -f1)

    if [[ -z "$major_version" ]] || [[ "$major_version" -lt 4 ]]; then
        log_error "yq v4+ is required (found: $yq_version)"
        log_error "Install with: brew install yq"
        return 1
    fi

    return 0
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
    local yaml_file="$1"

    # File doesn't exist - return empty
    if [[ ! -f "$yaml_file" ]]; then
        return 0
    fi

    # Validate schema version
    local schema_version
    schema_version=$(yq -r '.schema_version // ""' "$yaml_file" 2>/dev/null)
    if [[ -n "$schema_version" ]] && [[ "$schema_version" != "1.0" ]]; then
        log_warning "Unknown schema version: $schema_version (expected 1.0)"
    fi

    # Get hook count
    local hook_count
    hook_count=$(yq -r '.hooks | length' "$yaml_file" 2>/dev/null)
    if [[ -z "$hook_count" ]] || [[ "$hook_count" -eq 0 ]]; then
        return 0
    fi

    # Process each hook entry
    local i
    for ((i=0; i<hook_count; i++)); do
        local event matcher path timeout

        event=$(yq -r ".hooks[$i].event // \"\"" "$yaml_file")
        matcher=$(yq -r ".hooks[$i].matcher // \"\"" "$yaml_file")
        path=$(yq -r ".hooks[$i].path // \"\"" "$yaml_file")
        timeout=$(yq -r ".hooks[$i].timeout // 5" "$yaml_file")

        # Validate event type
        case "$event" in
            SessionStart|Stop|PreToolUse|PostToolUse|UserPromptSubmit)
                ;;
            *)
                log_warning "Invalid event type: $event (skipping)"
                continue
                ;;
        esac

        # Validate matcher requirement for PreToolUse and PostToolUse
        if [[ "$event" == "PreToolUse" || "$event" == "PostToolUse" ]]; then
            if [[ -z "$matcher" ]]; then
                log_warning "Event $event requires matcher (skipping: $path)"
                continue
            fi
        fi

        # Validate path is provided
        if [[ -z "$path" ]]; then
            log_warning "Hook entry $i missing path (skipping)"
            continue
        fi

        # Validate matcher syntax (check regex compiles without error)
        if [[ -n "$matcher" ]]; then
            # Use grep -E with a test string to validate regex syntax
            # We check exit code 0 or 1 (valid regex), 2 means syntax error
            echo "test" | grep -E "$matcher" >/dev/null 2>&1
            local grep_exit=$?
            if [[ $grep_exit -eq 2 ]]; then
                log_warning "Invalid matcher regex: $matcher (skipping: $path)"
                continue
            fi
        fi

        # Clamp timeout to valid range
        if [[ "$timeout" -gt 60 ]]; then
            log_warning "Timeout $timeout exceeds 60s limit, clamping to 60 (hook: $path)"
            timeout=60
        fi
        if [[ "$timeout" -lt 1 ]]; then
            timeout=5
        fi

        # Emit registration record (JSON-lines format)
        # Use jq to properly escape strings
        jq -n -c \
            --arg event "$event" \
            --arg matcher "$matcher" \
            --arg path "$path" \
            --argjson timeout "$timeout" \
            '{event: $event, matcher: $matcher, path: $path, timeout: $timeout}'
    done
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
    local settings_file="$1"

    # File doesn't exist - return empty object
    if [[ ! -f "$settings_file" ]]; then
        echo "{}"
        return 0
    fi

    # Read current hooks section
    local current_hooks
    current_hooks=$(jq '.hooks // {}' "$settings_file" 2>/dev/null)
    if [[ -z "$current_hooks" ]] || [[ "$current_hooks" == "null" ]]; then
        echo "{}"
        return 0
    fi

    # For each event type, filter out roster-managed hooks
    # Roster hooks contain ".claude/hooks/" in the command path
    local preserved="{}"
    local events=("SessionStart" "Stop" "PreToolUse" "PostToolUse" "UserPromptSubmit")

    for event in "${events[@]}"; do
        local event_entries
        event_entries=$(echo "$current_hooks" | jq -c ".\"$event\" // []")

        local entry_count
        entry_count=$(echo "$event_entries" | jq 'length')
        [[ "$entry_count" -eq 0 ]] && continue

        local filtered_entries="[]"
        local i
        for ((i=0; i<entry_count; i++)); do
            local entry
            entry=$(echo "$event_entries" | jq -c ".[$i]")

            # Filter hooks array within entry to exclude roster-managed ones
            local filtered_hooks
            filtered_hooks=$(echo "$entry" | jq -c '[.hooks // [] | .[] | select(.command | contains(".claude/hooks/") | not)]')

            local filtered_count
            filtered_count=$(echo "$filtered_hooks" | jq 'length')

            if [[ "$filtered_count" -gt 0 ]]; then
                # Update entry with filtered hooks
                local new_entry
                new_entry=$(echo "$entry" | jq -c ".hooks = $filtered_hooks")
                filtered_entries=$(echo "$filtered_entries" | jq -c ". + [$new_entry]")
            fi
        done

        local filtered_len
        filtered_len=$(echo "$filtered_entries" | jq 'length')
        if [[ "$filtered_len" -gt 0 ]]; then
            preserved=$(echo "$preserved" | jq -c ".\"$event\" = $filtered_entries")
        fi
    done

    echo "$preserved"
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
    local base_registrations="$1"
    local team_registrations="$2"

    # Combine all registrations (base first, team second)
    printf '%s\n%s' "$base_registrations" "$team_registrations" | grep -v '^$' || true
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

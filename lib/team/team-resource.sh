#!/usr/bin/env bash
#
# team-resource.sh - Generic Team Resource Operations
#
# Consolidates backup, removal, orphan detection, and team membership
# checks for commands, skills, and hooks into parameterized functions.
#
# Part of: roster team-swap infrastructure
#
# Usage:
#   source "$ROSTER_HOME/lib/team/team-resource.sh"
#   backup_team_resource "commands" ".claude/commands" ".team-commands" "f"
#   detect_resource_orphans "commands" ".claude/commands" "my-team" "f" "*.md"
#
# Functions:
#   is_resource_from_team     - Check if resource belongs to any team
#   get_resource_team         - Get team name that owns a resource
#   backup_team_resource      - Backup team-owned resources before swap
#   remove_team_resource      - Remove team-owned resources
#   detect_resource_orphans   - Detect orphaned resources from other teams
#   remove_resource_orphans   - Remove orphaned resources with backup

# Guard against re-sourcing
[[ -n "${_TEAM_RESOURCE_LOADED:-}" ]] && return 0
readonly _TEAM_RESOURCE_LOADED=1

# ============================================================================
# Logging Stubs (overridden when sourced from swap-team.sh)
# ============================================================================

# These stub implementations provide basic logging when team-resource.sh
# is used standalone (e.g., in unit tests). When sourced from swap-team.sh,
# these are overridden by the full logging implementation.

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

# ============================================================================
# Team Membership Checks
# ============================================================================

# Check if a resource belongs to ANY team pack in ROSTER_HOME/teams/
#
# Parameters:
#   $1 - resource_name: basename of resource (e.g., "commit.md", "qa-ref")
#   $2 - resource_type: "commands" | "skills" | "hooks"
#   $3 - find_type:     "f" (file) | "d" (directory)
#
# Returns: 0 if resource is from a team, 1 otherwise
#
# Requires: ROSTER_HOME environment variable
is_resource_from_team() {
    local resource_name="$1"
    local resource_type="$2"
    local find_type="$3"

    find "$ROSTER_HOME/teams" -path "*/${resource_type}/$resource_name" -type "$find_type" 2>/dev/null | grep -q .
}

# Get the team name that owns a specific resource
#
# Parameters:
#   $1 - resource_name: basename of resource
#   $2 - resource_type: "commands" | "skills" | "hooks"
#   $3 - find_type:     "f" (file) | "d" (directory)
#
# Outputs: team name to stdout, empty if not found
#
# Requires: ROSTER_HOME environment variable
get_resource_team() {
    local resource_name="$1"
    local resource_type="$2"
    local find_type="$3"
    local match

    match=$(find "$ROSTER_HOME/teams" -path "*/${resource_type}/$resource_name" -type "$find_type" 2>/dev/null | head -1)
    if [[ -n "$match" ]]; then
        echo "$match" | sed 's|.*/teams/\([^/]*\)/'"$resource_type"'/.*|\1|'
    fi
}

# ============================================================================
# Backup Operations
# ============================================================================

# Backup team-owned resources to a .backup directory before swap
#
# Parameters:
#   $1 - resource_type: "commands" | "skills" | "hooks"
#   $2 - resource_dir:  ".claude/commands" | ".claude/skills" | ".claude/hooks"
#   $3 - marker_file:   ".team-commands" | ".team-skills" | ".team-hooks"
#   $4 - find_type:     "f" (file) | "d" (directory)
#
# Returns: 0 on success, 0 if nothing to backup
#
# Side effects:
#   - Creates ${resource_dir}.backup/ directory
#   - Copies team resources to backup
#   - Logs via log_debug()
backup_team_resource() {
    local resource_type="$1"
    local resource_dir="$2"
    local marker_file="$3"
    local find_type="$4"

    log_debug "Checking for team ${resource_type} to backup"

    local backup_dir="${resource_dir}.backup"
    local marker_path="${resource_dir}/${marker_file}"

    # Check if any team resources exist (marked by marker file)
    if [[ ! -d "$resource_dir" ]] || [[ ! -f "$marker_path" ]]; then
        log_debug "No team ${resource_type} to backup"
        return 0
    fi

    # Remove old backup if exists
    if [[ -d "$backup_dir" ]]; then
        log_debug "Removing old ${resource_type} backup"
        rm -rf "$backup_dir" || {
            log_warning "Failed to remove old ${resource_type} backup"
        }
    fi

    # Read list of team resources and backup
    mkdir -p "$backup_dir"
    while IFS= read -r resource_name; do
        [[ -z "$resource_name" ]] && continue
        local resource_path="${resource_dir}/${resource_name}"

        if [[ "$find_type" == "d" ]] && [[ -d "$resource_path" ]]; then
            # For directories (skills), use recursive copy with preservation
            cp -rp "$resource_path" "${backup_dir}/${resource_name}"
            log_debug "Backed up ${resource_type%s}: $resource_name"
        elif [[ "$find_type" == "f" ]] && [[ -f "$resource_path" ]]; then
            # For files (commands, hooks), use simple copy
            cp "$resource_path" "${backup_dir}/${resource_name}"
            log_debug "Backed up ${resource_type%s}: $resource_name"
        fi
    done < "$marker_path"

    log_debug "Team ${resource_type} backed up"
}

# ============================================================================
# Removal Operations
# ============================================================================

# Remove team-owned resources listed in marker file
#
# Parameters:
#   $1 - resource_type: "commands" | "skills" | "hooks"
#   $2 - resource_dir:  ".claude/commands" | ".claude/skills" | ".claude/hooks"
#   $3 - marker_file:   ".team-commands" | ".team-skills" | ".team-hooks"
#   $4 - find_type:     "f" (file) | "d" (directory)
#
# Returns: 0 on success
#
# Side effects:
#   - Removes resources listed in marker file
#   - Removes marker file itself
#   - Logs via log_debug()
remove_team_resource() {
    local resource_type="$1"
    local resource_dir="$2"
    local marker_file="$3"
    local find_type="$4"

    log_debug "Removing team ${resource_type} from previous team"

    local marker_path="${resource_dir}/${marker_file}"

    if [[ ! -f "$marker_path" ]]; then
        log_debug "No team ${resource_type} marker found"
        return 0
    fi

    # Read list and remove each resource
    while IFS= read -r resource_name; do
        [[ -z "$resource_name" ]] && continue
        local resource_path="${resource_dir}/${resource_name}"

        if [[ "$find_type" == "d" ]] && [[ -d "$resource_path" ]]; then
            # For directories (skills), use rm -rf
            rm -rf "$resource_path"
            log_debug "Removed team ${resource_type%s}: $resource_name"
        elif [[ "$find_type" == "f" ]] && [[ -f "$resource_path" ]]; then
            # For files (commands, hooks), use rm -f
            rm -f "$resource_path"
            log_debug "Removed team ${resource_type%s}: $resource_name"
        fi
    done < "$marker_path"

    # Remove the marker file
    rm -f "$marker_path"

    log_debug "Team ${resource_type} removed"
}

# ============================================================================
# Orphan Detection
# ============================================================================

# Detect orphaned resources from other teams that shouldn't be present
#
# Parameters:
#   $1 - resource_type:     "commands" | "skills" | "hooks"
#   $2 - resource_dir:      ".claude/commands" | ".claude/skills" | ".claude/hooks"
#   $3 - incoming_team:     name of team being swapped in
#   $4 - find_type:         "f" (file) | "d" (directory)
#   $5 - glob_pattern:      "*.md" | "*/" | "*" (for find pattern)
#
# Outputs: One "resource_name:origin_team" per line to stdout
#
# Returns: 0 always (empty output means no orphans)
#
# Note: Uses stdout instead of global arrays for bash 3.2 portability
detect_resource_orphans() {
    local resource_type="$1"
    local resource_dir="$2"
    local incoming_team="$3"
    local find_type="$4"
    local glob_pattern="$5"

    # Stub - to be implemented in RF-005
    return 0
}

# ============================================================================
# Orphan Removal
# ============================================================================

# Remove orphaned resources based on orphan mode
#
# Parameters:
#   $1 - resource_type:  "commands" | "skills" | "hooks"
#   $2 - resource_dir:   ".claude/commands" | ".claude/skills" | ".claude/hooks"
#   $3 - orphan_mode:    "remove" | "keep" | ""
#   $4 - find_type:      "f" (file) | "d" (directory)
#   stdin:               orphan list (one "name:team" per line)
#
# Returns: 0 on success
#
# Side effects:
#   - Creates ${resource_dir}.orphan-backup/ if removing
#   - Backs up and removes orphaned resources
#   - Logs removals via log()
remove_resource_orphans() {
    local resource_type="$1"
    local resource_dir="$2"
    local orphan_mode="$3"
    local find_type="$4"

    # Stub - to be implemented in RF-006
    return 0
}

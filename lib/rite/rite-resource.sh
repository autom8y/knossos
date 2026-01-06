#!/usr/bin/env bash
#
# rite-resource.sh - Generic Rite Resource Operations
#
# Consolidates backup, removal, orphan detection, and rite membership
# checks for commands, skills, and hooks into parameterized functions.
#
# Part of: roster rite-swap infrastructure
#
# Usage:
#   source "$ROSTER_HOME/lib/rite/rite-resource.sh"
#   backup_rite_resource "commands" ".claude/commands" ".rite-commands" "f"
#   detect_resource_orphans "commands" ".claude/commands" "my-rite" "f" "*.md"
#
# Functions:
#   is_resource_from_team     - Check if resource belongs to any team
#   get_resource_team         - Get team name that owns a resource
#   backup_team_resource      - Backup team-owned resources before swap
#   remove_team_resource      - Remove team-owned resources
#   detect_resource_orphans   - Detect orphaned resources from other teams
#   remove_resource_orphans   - Remove orphaned resources with backup

# Guard against re-sourcing
[[ -n "${_RITE_RESOURCE_LOADED:-}" ]] && return 0
readonly _RITE_RESOURCE_LOADED=1

# ============================================================================
# Logging Stubs (overridden when sourced from swap-rite.sh)
# ============================================================================

# These stub implementations provide basic logging when rite-resource.sh
# is used standalone (e.g., in unit tests). When sourced from swap-rite.sh,
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

# Check if a resource belongs to a team pack in ROSTER_HOME/teams/
#
# Parameters:
#   $1 - resource_name: basename of resource (e.g., "commit.md", "qa-ref")
#   $2 - resource_type: "commands" | "skills" | "hooks"
#   $3 - find_type:     "f" (file) | "d" (directory)
#   $4 - team_scope:    (optional) space-separated list of team names to check
#                       If empty, checks ALL teams (legacy behavior)
#
# Returns: 0 if resource is from a team, 1 otherwise
#
# Requires: ROSTER_HOME environment variable
is_resource_from_team() {
    local resource_name="$1"
    local resource_type="$2"
    local find_type="$3"
    local team_scope="${4:-}"

    if [[ -z "$team_scope" ]]; then
        # Legacy behavior: check ALL teams
        find "$ROSTER_HOME/teams" -path "*/${resource_type}/$resource_name" -type "$find_type" 2>/dev/null | grep -q .
    else
        # Scoped behavior: check only specified teams
        local team
        for team in $team_scope; do
            local check_path="$ROSTER_HOME/teams/$team/$resource_type/$resource_name"
            if [[ "$find_type" == "d" ]] && [[ -d "$check_path" ]]; then
                return 0
            elif [[ "$find_type" == "f" ]] && [[ -f "$check_path" ]]; then
                return 0
            fi
        done
        return 1
    fi
}

# Get the team name that owns a specific resource
#
# Parameters:
#   $1 - resource_name: basename of resource
#   $2 - resource_type: "commands" | "skills" | "hooks"
#   $3 - find_type:     "f" (file) | "d" (directory)
#   $4 - team_scope:    (optional) space-separated list of team names to check
#                       If empty, checks ALL teams (legacy behavior)
#
# Outputs: team name to stdout, empty if not found
#
# Requires: ROSTER_HOME environment variable
get_resource_team() {
    local resource_name="$1"
    local resource_type="$2"
    local find_type="$3"
    local team_scope="${4:-}"

    if [[ -z "$team_scope" ]]; then
        # Legacy behavior: check ALL teams
        local match
        match=$(find "$ROSTER_HOME/teams" -path "*/${resource_type}/$resource_name" -type "$find_type" 2>/dev/null | head -1)
        if [[ -n "$match" ]]; then
            echo "$match" | sed 's|.*/teams/\([^/]*\)/'"$resource_type"'/.*|\1|'
        fi
    else
        # Scoped behavior: check only specified teams
        local team
        for team in $team_scope; do
            local check_path="$ROSTER_HOME/teams/$team/$resource_type/$resource_name"
            if [[ "$find_type" == "d" ]] && [[ -d "$check_path" ]]; then
                echo "$team"
                return 0
            elif [[ "$find_type" == "f" ]] && [[ -f "$check_path" ]]; then
                echo "$team"
                return 0
            fi
        done
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
#   $6 - previous_team:     (optional) name of the previous team
#                           If provided, only flags orphans from this team
#                           If empty, checks ALL teams (legacy behavior)
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
    local previous_team="${6:-}"

    local incoming_resource_dir="$ROSTER_HOME/teams/$incoming_team/$resource_type"
    local orphan_count=0

    # Build team scope for orphan detection
    # If previous_team is provided, only check that team (scoped detection)
    # Otherwise, check all teams (legacy behavior)
    local team_scope=""
    if [[ -n "$previous_team" ]]; then
        team_scope="$previous_team"
        log_debug "Orphan detection scoped to previous team: $previous_team"
    fi

    # Return if resource directory doesn't exist
    [[ -d "$resource_dir" ]] || return 0

    # Iterate over resources using glob pattern
    for resource_path in $resource_dir/$glob_pattern; do
        # Check if path exists (glob may not match anything)
        if [[ "$find_type" == "d" ]]; then
            [[ -d "$resource_path" ]] || continue
        else
            [[ -f "$resource_path" ]] || continue
        fi

        local resource_name
        resource_name=$(basename "$resource_path")

        # Skip hidden files/dirs and special directories
        [[ "$resource_name" == .* ]] && continue
        [[ "$resource_name" == "lib" ]] && continue

        # Is this resource in the incoming team? If so, skip (not an orphan)
        if [[ "$find_type" == "d" ]] && [[ -d "$incoming_resource_dir/$resource_name" ]]; then
            continue
        elif [[ "$find_type" == "f" ]] && [[ -f "$incoming_resource_dir/$resource_name" ]]; then
            continue
        fi

        # Is this resource from a team pack (scoped or all)?
        if is_resource_from_team "$resource_name" "$resource_type" "$find_type" "$team_scope"; then
            local origin_team
            origin_team=$(get_resource_team "$resource_name" "$resource_type" "$find_type" "$team_scope")
            echo "$resource_name:$origin_team"
            orphan_count=$((orphan_count + 1))
            log_debug "Orphan ${resource_type%s} detected: $resource_name (from $origin_team)"
        fi
    done

    log_debug "Total orphan ${resource_type}: $orphan_count"
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

    local backup_dir="${resource_dir}.orphan-backup"
    local orphan_count=0

    # Read orphan list from stdin (one "name:team" per line)
    while IFS=: read -r resource_name origin_team; do
        # Skip empty lines
        [[ -z "$resource_name" ]] && continue

        orphan_count=$((orphan_count + 1))
        local resource_path="${resource_dir}/${resource_name}"

        case "$orphan_mode" in
            "remove")
                # Create backup directory if it doesn't exist
                mkdir -p "$backup_dir"

                # Backup and remove based on resource type
                if [[ "$find_type" == "d" ]] && [[ -d "$resource_path" ]]; then
                    # For directories (skills), use recursive copy
                    cp -rp "$resource_path" "$backup_dir/$resource_name"
                    rm -rf "$resource_path"
                    log "Removed orphan ${resource_type%s}: $resource_name (was from $origin_team)"
                elif [[ "$find_type" == "f" ]] && [[ -f "$resource_path" ]]; then
                    # For files (commands, hooks), use simple copy
                    cp "$resource_path" "$backup_dir/$resource_name"
                    rm -f "$resource_path"
                    log "Removed orphan ${resource_type%s}: $resource_name (was from $origin_team)"
                fi
                ;;
            "keep")
                log "Keeping orphan ${resource_type%s}: $resource_name (from $origin_team)"
                ;;
            *)
                # Default: keep silently (no explicit mode set)
                log_debug "Keeping orphan ${resource_type%s}: $resource_name (no disposition)"
                ;;
        esac
    done

    # Log summary if backups were created
    if [[ "$orphan_mode" == "remove" ]] && [[ -d "$backup_dir" ]]; then
        log "Orphan ${resource_type%s} backups saved to: $backup_dir"
    fi

    return 0
}

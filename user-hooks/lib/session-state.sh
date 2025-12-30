#!/bin/bash
# Session state management - validation, staleness, team sync
# Single Responsibility: Session state queries and business logic
#
# Addresses: DRY-001, DRY-002 (consolidate validation/team sync)
# Part of Ecosystem v2 refactoring (RF-005)

# Source session-core (which sources primitives -> config)
# Provides: get_session_id, get_session_dir, get_current_session, atomic_write, etc.
# shellcheck source=session-core.sh
source "$(dirname "${BASH_SOURCE[0]}")/session-core.sh"

# =============================================================================
# Session State Queries
# =============================================================================

# Get current session state
# Returns: ACTIVE, PARKED, AUTO_PARKED, or NONE
# Note: This infers state from presence of park fields, not from reading the status field.
# The status field (canonical) is managed separately and may have values like "ACTIVE", "PARKED", "ARCHIVED".
get_session_state() {
    local session_id="${1:-$(get_session_id)}"

    if [ -z "$session_id" ]; then
        echo "NONE"
        return 0
    fi

    local project_dir="${CLAUDE_PROJECT_DIR:-.}"
    local session_dir="$project_dir/.claude/sessions/$session_id"
    local session_file="$session_dir/SESSION_CONTEXT.md"

    # Check session exists
    if [ ! -f "$session_file" ]; then
        echo "NONE"
        return 0
    fi

    # Check for auto_parked_at first (more specific)
    if grep -q "^auto_parked_at:" "$session_file" 2>/dev/null; then
        echo "AUTO_PARKED"
        return 0
    fi

    # Check for parked_at
    if grep -q "^parked_at:" "$session_file" 2>/dev/null; then
        echo "PARKED"
        return 0
    fi

    echo "ACTIVE"
}

# Get a field from SESSION_CONTEXT.md frontmatter
# Usage: get_session_field "field_name" [session_id]
get_session_field() {
    local field="$1"
    local session_id="${2:-$(get_session_id)}"

    if [ -z "$session_id" ]; then
        return 1
    fi

    local project_dir="${CLAUDE_PROJECT_DIR:-.}"
    local session_file="$project_dir/.claude/sessions/$session_id/SESSION_CONTEXT.md"

    if [ ! -f "$session_file" ]; then
        return 1
    fi

    # Use get_yaml_field from primitives.sh
    get_yaml_field "$session_file" "$field"
}

# Set a field in SESSION_CONTEXT.md frontmatter
# Usage: set_session_field "field_name" "value" [session_id]
set_session_field() {
    local field="$1"
    local value="$2"
    local session_id="${3:-$(get_session_id)}"

    if [ -z "$session_id" ] || [ -z "$field" ]; then
        return 1
    fi

    local project_dir="${CLAUDE_PROJECT_DIR:-.}"
    local session_file="$project_dir/.claude/sessions/$session_id/SESSION_CONTEXT.md"

    if [ ! -f "$session_file" ]; then
        return 1
    fi

    local temp_content

    # Update or add field in frontmatter
    if grep -q "^${field}:" "$session_file" 2>/dev/null; then
        # Update existing field
        temp_content=$(sed "s/^${field}:.*/${field}: \"${value}\"/" "$session_file")
    else
        # Insert before the closing --- of frontmatter
        temp_content=$(awk -v f="$field" -v v="$value" '
            /^---$/ && ++count == 2 {
                print f ": \"" v "\""
            }
            { print }
        ' "$session_file")
    fi

    # Write atomically
    atomic_write "$session_file" "$temp_content"
}

# Check if current session is parked
# Returns: 0 if parked, 1 if not parked or no session
is_parked() {
    local state
    state=$(get_session_state "$@")
    [[ "$state" == "PARKED" || "$state" == "AUTO_PARKED" ]]
}

# Get session initiative name
# Returns: initiative name or empty string
get_initiative() {
    get_session_field "initiative" "$@"
}

# Get session complexity level
# Returns: complexity level or empty string
get_complexity() {
    get_session_field "complexity" "$@"
}

# =============================================================================
# Session Validation
# =============================================================================
# This is THE ONLY place for session validation (consolidates 4 locations)

# Validate SESSION_CONTEXT.md has required fields
# Usage: validate_session_context "path/to/SESSION_CONTEXT.md"
# Returns: 0 if valid, 1 if missing required fields
validate_session_context() {
    local file="$1"
    local required_fields=("session_id" "created_at" "initiative" "complexity" "active_team" "current_phase")
    local missing=()

    [ -f "$file" ] || { echo "File not found: $file" >&2; return 1; }

    for field in "${required_fields[@]}"; do
        if ! grep -q "^$field:" "$file" 2>/dev/null; then
            missing+=("$field")
        fi
    done

    if [ ${#missing[@]} -gt 0 ]; then
        echo "Missing required fields: ${missing[*]}" >&2
        return 1
    fi

    return 0
}

# Validate session ID format
# Usage: validate_session_id_format "session-id"
# Returns: 0 if valid format, 1 if invalid
validate_session_id_format() {
    local session_id="$1"
    [[ "$session_id" =~ ^session-[0-9]{8}-[0-9]{6}-[a-f0-9]{8}$ ]]
}

# =============================================================================
# Session Staleness Detection
# =============================================================================

# Update last_accessed_at in SESSION_CONTEXT.md
# Uses atomic_write to prevent corruption on interrupted writes (STATE-004)
touch_session() {
    local session_dir
    session_dir=$(get_session_dir)
    local session_file="$session_dir/SESSION_CONTEXT.md"
    [ -f "$session_file" ] || return 1

    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    local updated_content

    # Update or add last_accessed_at in frontmatter
    if grep -q "^last_accessed_at:" "$session_file" 2>/dev/null; then
        updated_content=$(sed "s/^last_accessed_at:.*/last_accessed_at: \"$timestamp\"/" "$session_file")
    else
        # Insert before the closing --- of frontmatter
        updated_content=$(awk -v ts="$timestamp" '
            /^---$/ && ++count == 2 {
                print "last_accessed_at: \"" ts "\""
            }
            { print }
        ' "$session_file")
    fi

    # Write atomically to prevent corruption
    atomic_write "$session_file" "$updated_content"
}

# Check if session is stale (not accessed in N hours, default 24)
is_session_stale() {
    local hours="${1:-24}"
    local session_dir
    session_dir=$(get_session_dir)
    local session_file="$session_dir/SESSION_CONTEXT.md"
    [ -f "$session_file" ] || return 1

    # Use get_yaml_field for proper handling of quoted/unquoted timestamps
    local last_access
    last_access=$(get_yaml_field "$session_file" "last_accessed_at")
    [ -z "$last_access" ] && return 0  # No timestamp = stale

    # Convert to epoch and compare (portable across macOS/Linux)
    local last_epoch now_epoch
    if date -j -f "%Y-%m-%dT%H:%M:%SZ" "$last_access" +%s >/dev/null 2>&1; then
        # macOS
        last_epoch=$(date -j -f "%Y-%m-%dT%H:%M:%SZ" "$last_access" +%s 2>/dev/null)
    else
        # Linux
        last_epoch=$(date -d "$last_access" +%s 2>/dev/null)
    fi
    now_epoch=$(date +%s)

    local stale_seconds=$((hours * 3600))
    local age=$((now_epoch - last_epoch))

    [ "$age" -gt "$stale_seconds" ]
}

# =============================================================================
# Session Listing
# =============================================================================

# List all session directories (excluding .tty-map)
list_sessions() {
    find .claude/sessions -maxdepth 1 -type d -name "session-*" 2>/dev/null | sort
}

# List sessions that have parked_at in their SESSION_CONTEXT.md
list_parked_sessions() {
    for dir in $(list_sessions); do
        if grep -qE "^(parked_at|auto_parked_at):" "$dir/SESSION_CONTEXT.md" 2>/dev/null; then
            echo "$dir"
        fi
    done
}

# List stale sessions (for cleanup suggestions)
list_stale_sessions() {
    local hours="${1:-24}"
    for dir in $(list_sessions); do
        local session_file="$dir/SESSION_CONTEXT.md"
        [ -f "$session_file" ] || continue

        # Check parked and stale
        if grep -qE "^(parked_at|auto_parked_at):" "$session_file" 2>/dev/null; then
            # Check staleness for parked sessions
            if is_session_stale "$hours"; then
                echo "$dir"
            fi
        fi
    done
}

# =============================================================================
# Cleanup
# =============================================================================

# Remove TTY mappings that are orphaned OR older than 24 hours
# Addresses STATE-002: No cleanup for orphaned TTY mappings
cleanup_stale_mappings() {
    local max_age_hours="${1:-24}"
    local project_dir="${CLAUDE_PROJECT_DIR:-.}"
    local map_dir="$project_dir/.claude/sessions/.tty-map"
    local cleaned=0

    [ -d "$map_dir" ] || return 0

    # Calculate cutoff time
    local now_epoch cutoff_epoch
    now_epoch=$(date +%s)
    cutoff_epoch=$((now_epoch - (max_age_hours * 3600)))

    for map_file in "$map_dir"/*; do
        [ -f "$map_file" ] || continue

        local should_remove=0
        local session_id
        session_id=$(cat "$map_file" 2>/dev/null)

        # Check 1: Points to non-existent session (orphaned)
        if [ -n "$session_id" ] && [ ! -d "$project_dir/.claude/sessions/$session_id" ]; then
            should_remove=1
        fi

        # Check 2: Map file is older than max_age_hours
        if [ "$should_remove" -eq 0 ]; then
            local file_mtime
            if [ "$(uname)" = "Darwin" ]; then
                file_mtime=$(stat -f%m "$map_file" 2>/dev/null)
            else
                file_mtime=$(stat -c%Y "$map_file" 2>/dev/null)
            fi
            if [ -n "$file_mtime" ] && [ "$file_mtime" -lt "$cutoff_epoch" ]; then
                should_remove=1
            fi
        fi

        if [ "$should_remove" -eq 1 ]; then
            rm -f "$map_file"
            ((cleaned++)) || true
        fi
    done

    echo "$cleaned"
}

# =============================================================================
# Team State Synchronization
# =============================================================================
# This is THE ONLY place for team updates (consolidates 3 locations)

# Atomic team update: sync ACTIVE_TEAM and SESSION_CONTEXT.active_team together
# Usage: atomic_team_update "new_team_name"
# Returns: 0 on success, 1 on failure
atomic_team_update() {
    local new_team="$1"
    local project_dir="${CLAUDE_PROJECT_DIR:-.}"
    local active_team_file="$project_dir/.claude/ACTIVE_TEAM"
    local session_dir
    session_dir=$(get_session_dir)

    if [ -z "$new_team" ]; then
        echo "Error: Team name required" >&2
        return 1
    fi

    # Acquire lock to prevent race conditions
    if ! acquire_session_lock "team_update"; then
        echo "Error: Could not acquire lock for team update" >&2
        return 1
    fi

    # Trap to ensure lock release on error
    trap 'release_session_lock "team_update"' EXIT

    local success=1

    # Update ACTIVE_TEAM atomically
    if atomic_write "$active_team_file" "$new_team"; then
        success=0

        # Update SESSION_CONTEXT.md if session exists
        if [ -n "$session_dir" ] && [ -f "$session_dir/SESSION_CONTEXT.md" ]; then
            local session_file="$session_dir/SESSION_CONTEXT.md"
            local temp_content

            # Read current content and update active_team field
            if grep -q "^active_team:" "$session_file" 2>/dev/null; then
                temp_content=$(sed "s/^active_team:.*/active_team: \"$new_team\"/" "$session_file")
            else
                # Insert active_team before closing --- of frontmatter
                temp_content=$(awk -v team="$new_team" '
                    /^---$/ && ++count == 2 {
                        print "active_team: \"" team "\""
                    }
                    { print }
                ' "$session_file")
            fi

            # Write updated content atomically
            if ! atomic_write "$session_file" "$temp_content"; then
                # SESSION_CONTEXT update failed, but ACTIVE_TEAM succeeded
                # Log warning but don't fail the operation
                echo "Warning: ACTIVE_TEAM updated but SESSION_CONTEXT update failed" >&2
            fi
        fi
    else
        success=1
    fi

    # Release lock and clear trap
    trap - EXIT
    release_session_lock "team_update"

    return $success
}

# =============================================================================
# Worktree Utilities
# =============================================================================

# Detect if current directory is a git worktree (not main working tree)
is_worktree() {
    local git_dir
    git_dir=$(git rev-parse --git-dir 2>/dev/null) || return 1
    # If .git is a file (not directory) containing "gitdir:", it's a worktree
    [[ -f "$git_dir" ]] && grep -q "^gitdir:" "$git_dir" 2>/dev/null
}

# Get worktree metadata as JSON (or empty object if not in worktree)
get_worktree_meta() {
    if is_worktree && [[ -f ".claude/.worktree-meta.json" ]]; then
        cat ".claude/.worktree-meta.json"
    else
        echo "{}"
    fi
}

# Get specific field from worktree metadata
# Usage: get_worktree_field "worktree_id"
get_worktree_field() {
    local field="$1"
    local meta
    meta=$(get_worktree_meta)
    if command -v jq >/dev/null 2>&1; then
        echo "$meta" | jq -r ".$field // empty" 2>/dev/null
    else
        # Fallback: grep-based parsing
        echo "$meta" | grep -o "\"$field\": *\"[^\"]*\"" 2>/dev/null | cut -d'"' -f4
    fi
}

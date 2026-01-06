#!/bin/bash
# Session schema migration: v1 -> v2
#
# Migrates legacy SESSION_CONTEXT.md files to the v2 schema with:
#   - Single source of truth status field
#   - Field canonicalization (removes duplicates)
#   - Metadata extraction to event log
#   - Schema version upgrade to 2.0
#   - Backup and rollback support
#
# Reference: TDD-session-state-machine.md (Migration Design section)
#
# Usage:
#   ./session-migrate.sh migrate [--dry-run] [--batch] [session_id]
#   ./session-migrate.sh rollback [--batch] [session_id]
#   ./session-migrate.sh status [session_id]
#   ./session-migrate.sh help

set -euo pipefail

# =============================================================================
# Configuration
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"

# Source session-fsm for validation
if [[ -f "$SCRIPT_DIR/session-fsm.sh" ]]; then
    # shellcheck source=session-fsm.sh
    source "$SCRIPT_DIR/session-fsm.sh"
fi

# Override sessions directory if not set
SESSIONS_DIR="${FSM_SESSIONS_DIR:-$PROJECT_DIR/.claude/sessions}"

# Migration state
MIGRATE_DRY_RUN="${MIGRATE_DRY_RUN:-false}"
MIGRATE_VERBOSE="${MIGRATE_VERBOSE:-false}"

# =============================================================================
# Utility Functions
# =============================================================================

# Log with timestamp
_log() {
    local level="$1"
    shift
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    echo "[$timestamp] [$level] $*" >&2
}

_log_info() { _log "INFO" "$@"; }
_log_warn() { _log "WARN" "$@"; }
_log_error() { _log "ERROR" "$@"; }
_log_debug() { [[ "$MIGRATE_VERBOSE" == "true" ]] && _log "DEBUG" "$@" || true; }

# Extract YAML field value from file
_get_field() {
    local file="$1"
    local field="$2"
    grep -m1 "^${field}:" "$file" 2>/dev/null | cut -d: -f2- | tr -d ' "' || true
}

# Extract YAML field value preserving spaces (for reason fields)
_get_field_with_spaces() {
    local file="$1"
    local field="$2"
    grep -m1 "^${field}:" "$file" 2>/dev/null | cut -d: -f2- | sed 's/^ *//; s/^"//; s/"$//' || true
}

# Check if file is v1 schema (missing schema_version or not 2.0/2.1)
_is_v1_session() {
    local ctx_file="$1"

    if [[ ! -f "$ctx_file" ]]; then
        return 1
    fi

    local version
    version=$(_get_field "$ctx_file" "schema_version")

    [[ -z "$version" || ("$version" != "2.0" && "$version" != "2.1") ]]
}

# =============================================================================
# Migration Core
# =============================================================================

# Migrate a single session from v1 to v2 schema
# Usage: migrate_session <session_id>
# Returns: 0 on success, 1 on failure
migrate_session() {
    local session_id="$1"
    local session_dir="$SESSIONS_DIR/$session_id"
    local ctx_file="$session_dir/SESSION_CONTEXT.md"
    local backup_file="$session_dir/SESSION_CONTEXT.md.v1.backup"
    local events_file="$session_dir/events.jsonl"

    # Validate session exists
    if [[ ! -f "$ctx_file" ]]; then
        _log_error "Session not found: $session_id"
        return 1
    fi

    # Skip if already v2
    if ! _is_v1_session "$ctx_file"; then
        _log_info "Already v2: $session_id (skipping)"
        return 0
    fi

    _log_info "Migrating: $session_id"

    # Dry run: just report what would happen
    if [[ "$MIGRATE_DRY_RUN" == "true" ]]; then
        local derived_status
        derived_status=$(_derive_v1_status "$ctx_file")
        echo "Would migrate $session_id: status=$derived_status"
        return 0
    fi

    # Create backup
    if ! cp "$ctx_file" "$backup_file"; then
        _log_error "Failed to create backup: $backup_file"
        return 1
    fi
    _log_debug "Backup created: $backup_file"

    # Determine canonical status from v1 fields
    local new_status
    new_status=$(_derive_v1_status "$ctx_file")

    # Extract park metadata for event log (before removing fields)
    _extract_park_metadata_to_events "$ctx_file" "$events_file"

    # Transform file: remove legacy fields, add v2 fields
    if ! _transform_to_v2 "$ctx_file" "$new_status"; then
        _log_error "Transform failed, rolling back: $session_id"
        mv "$backup_file" "$ctx_file"
        return 1
    fi

    # Validate result
    if ! _validate_migrated_session "$ctx_file"; then
        _log_error "Validation failed, rolling back: $session_id"
        mv "$backup_file" "$ctx_file"
        return 1
    fi

    # Emit migration event
    _emit_migration_event "$session_id" "$events_file" "$new_status"

    _log_info "Migrated: $session_id (status=$new_status)"
    return 0
}

# Derive v1 session status from legacy fields
# Priority: completed_at (ARCHIVED) > parked_at/auto_parked_at (PARKED) > default (ACTIVE)
_derive_v1_status() {
    local ctx_file="$1"

    # Check for completed_at first (terminal state)
    if grep -q "^completed_at:" "$ctx_file" 2>/dev/null; then
        echo "ARCHIVED"
        return
    fi

    # Check for park fields
    if grep -qE "^(parked_at|auto_parked_at):" "$ctx_file" 2>/dev/null; then
        echo "PARKED"
        return
    fi

    # Default to ACTIVE
    echo "ACTIVE"
}

# Extract park metadata to event log
_extract_park_metadata_to_events() {
    local ctx_file="$1"
    local events_file="$2"

    # Check if any park fields exist
    if ! grep -qE "^(parked_at|auto_parked_at):" "$ctx_file" 2>/dev/null; then
        return 0
    fi

    local parked_at=""
    local park_reason=""
    local git_status=""

    # Extract parked_at (try both variants)
    parked_at=$(_get_field "$ctx_file" "parked_at")
    [[ -z "$parked_at" ]] && parked_at=$(_get_field "$ctx_file" "auto_parked_at")

    # Extract park_reason (try all variants, preserving spaces)
    park_reason=$(_get_field_with_spaces "$ctx_file" "parked_reason")
    [[ -z "$park_reason" ]] && park_reason=$(_get_field_with_spaces "$ctx_file" "park_reason")
    [[ -z "$park_reason" ]] && park_reason=$(_get_field_with_spaces "$ctx_file" "auto_parked_reason")

    # Extract git_status (try both variants)
    git_status=$(_get_field "$ctx_file" "parked_git_status")
    [[ -z "$git_status" ]] && git_status=$(_get_field "$ctx_file" "git_status_at_park")

    # Build and write event JSON
    if [[ -n "$parked_at" ]]; then
        local event="{\"timestamp\":\"$parked_at\",\"event\":\"SESSION_PARKED\",\"source\":\"migration\""
        [[ -n "$park_reason" ]] && event+=",\"reason\":\"$park_reason\""
        [[ -n "$git_status" ]] && event+=",\"git_status\":\"$git_status\""
        event+="}"

        echo "$event" >> "$events_file"
        _log_debug "Wrote park event to: $events_file"
    fi
}

# Transform v1 context file to v2.1 schema
_transform_to_v2() {
    local ctx_file="$1"
    local new_status="$2"
    local temp_file="${ctx_file}.tmp.$$"

    # Determine rite value from active_rite field (or legacy active_team for backward compat)
    local active_rite
    active_rite=$(_get_field "$ctx_file" "active_rite")
    [[ -z "$active_rite" ]] && active_rite=$(_get_field "$ctx_file" "active_team")
    local rite_value="null"
    if [[ -n "$active_rite" && "$active_rite" != "none" ]]; then
        rite_value="\"$active_rite\""
    fi

    # Check for auto_parked_at to merge
    local auto_parked_at
    auto_parked_at=$(_get_field "$ctx_file" "auto_parked_at")
    local has_parked_at
    has_parked_at=$(grep -c "^parked_at:" "$ctx_file" 2>/dev/null || echo "0")

    # Use awk to process the file:
    # - Add schema_version (2.1), status, and rite fields
    # - Merge auto_parked_at into parked_at if needed
    # - Remove legacy fields
    # - Preserve body content
    awk -v status="$new_status" -v rite="$rite_value" -v auto_ts="$auto_parked_at" -v has_parked="$has_parked_at" '
    BEGIN {
        in_frontmatter = 0
        frontmatter_count = 0
        version_written = 0
        status_written = 0
        rite_written = 0
    }

    /^---$/ {
        frontmatter_count++
        if (frontmatter_count == 1) {
            in_frontmatter = 1
            print
            next
        }
        if (frontmatter_count == 2) {
            # Add v2.1 fields before closing ---
            if (!version_written) print "schema_version: \"2.1\""
            if (!status_written) print "status: \"" status "\""
            if (!rite_written) print "rite: " rite
            # If auto_parked_at exists and no parked_at, merge it
            if (auto_ts != "" && has_parked == "0") {
                print "parked_at: \"" auto_ts "\""
                print "parked_auto: true"
            }
            in_frontmatter = 0
            print
            next
        }
    }

    in_frontmatter {
        # Skip legacy fields that are being removed
        if (/^session_state:/) next
        if (/^auto_parked_at:/) next
        if (/^auto_parked_reason:/) next
        if (/^git_status_at_park:/) next
        if (/^parked_git_status:/) next

        # Track if we see existing v2 fields
        if (/^schema_version:/) { version_written = 1 }
        if (/^status:/) { status_written = 1 }
        if (/^rite:/) { rite_written = 1 }
        if (/^team:/) { rite_written = 1 }  # Also count legacy team field
    }

    { print }
    ' "$ctx_file" > "$temp_file"

    # Atomic move
    if mv "$temp_file" "$ctx_file"; then
        return 0
    else
        rm -f "$temp_file"
        return 1
    fi
}

# Validate migrated session meets v2.1 requirements
_validate_migrated_session() {
    local ctx_file="$1"

    # Check schema_version is 2.0 or 2.1
    local version
    version=$(_get_field "$ctx_file" "schema_version")
    if [[ "$version" != "2.0" && "$version" != "2.1" ]]; then
        _log_error "Missing schema_version: 2.0 or 2.1"
        return 1
    fi

    # Check status is valid enum
    local status
    status=$(_get_field "$ctx_file" "status")
    case "$status" in
        ACTIVE|PARKED|ARCHIVED)
            ;;
        *)
            _log_error "Invalid status value: $status"
            return 1
            ;;
    esac

    # Check required fields (rite is optional for v2.1, accepts both active_rite and legacy active_team)
    local required_fields=("session_id" "created_at" "initiative" "complexity" "current_phase")
    for field in "${required_fields[@]}"; do
        if ! grep -q "^${field}:" "$ctx_file" 2>/dev/null; then
            _log_error "Missing required field: $field"
            return 1
        fi
    done

    # Check for active_rite OR active_team (backward compat)
    if ! grep -q "^active_rite:" "$ctx_file" 2>/dev/null && ! grep -q "^active_team:" "$ctx_file" 2>/dev/null; then
        _log_error "Missing required field: active_rite"
        return 1
    fi

    return 0
}

# Emit migration event to event log
_emit_migration_event() {
    local session_id="$1"
    local events_file="$2"
    local new_status="$3"
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    local event="{\"timestamp\":\"$timestamp\",\"event\":\"SCHEMA_MIGRATED\",\"from_version\":\"1.0\",\"to_version\":\"2.1\",\"derived_status\":\"$new_status\"}"
    echo "$event" >> "$events_file"

    # Also log to audit trail
    local audit_dir="$SESSIONS_DIR/.audit"
    mkdir -p "$audit_dir" 2>/dev/null
    echo "$timestamp | MIGRATE | $session_id | v1 -> v2.1 | status=$new_status" >> "$audit_dir/migrations.log"
}

# =============================================================================
# Rollback
# =============================================================================

# Rollback a single session from v2 to v1
# Usage: rollback_session <session_id>
rollback_session() {
    local session_id="$1"
    local session_dir="$SESSIONS_DIR/$session_id"
    local ctx_file="$session_dir/SESSION_CONTEXT.md"
    local backup_file="$session_dir/SESSION_CONTEXT.md.v1.backup"

    if [[ ! -f "$backup_file" ]]; then
        _log_error "No backup found for rollback: $session_id"
        return 1
    fi

    if [[ "$MIGRATE_DRY_RUN" == "true" ]]; then
        echo "Would rollback: $session_id"
        return 0
    fi

    if mv "$backup_file" "$ctx_file"; then
        _log_info "Rolled back: $session_id"

        # Log to audit trail
        local audit_dir="$SESSIONS_DIR/.audit"
        local timestamp
        timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
        mkdir -p "$audit_dir" 2>/dev/null
        echo "$timestamp | ROLLBACK | $session_id | v2 -> v1" >> "$audit_dir/migrations.log"

        return 0
    else
        _log_error "Rollback failed: $session_id"
        return 1
    fi
}

# =============================================================================
# Batch Operations
# =============================================================================

# Migrate all v1 sessions
migrate_all_sessions() {
    local success=0
    local failed=0
    local skipped=0

    _log_info "Starting batch migration..."

    # Find all session directories
    for session_dir in "$SESSIONS_DIR"/session-*; do
        [[ -d "$session_dir" ]] || continue

        local session_id
        session_id=$(basename "$session_dir")
        local ctx_file="$session_dir/SESSION_CONTEXT.md"

        # Skip if no context file
        if [[ ! -f "$ctx_file" ]]; then
            _log_debug "Skipping (no context): $session_id"
            continue
        fi

        # Skip if already v2
        if ! _is_v1_session "$ctx_file"; then
            ((skipped++))
            continue
        fi

        if migrate_session "$session_id"; then
            ((success++))
        else
            ((failed++))
        fi
    done

    _log_info "Batch migration complete: $success succeeded, $failed failed, $skipped skipped (already v2)"

    # Return success only if no failures
    [[ $failed -eq 0 ]]
}

# Rollback all migrated sessions
rollback_all_sessions() {
    local success=0
    local failed=0
    local skipped=0

    _log_info "Starting batch rollback..."

    for session_dir in "$SESSIONS_DIR"/session-*; do
        [[ -d "$session_dir" ]] || continue

        local session_id
        session_id=$(basename "$session_dir")
        local backup_file="$session_dir/SESSION_CONTEXT.md.v1.backup"

        # Skip if no backup
        if [[ ! -f "$backup_file" ]]; then
            ((skipped++))
            continue
        fi

        if rollback_session "$session_id"; then
            ((success++))
        else
            ((failed++))
        fi
    done

    _log_info "Batch rollback complete: $success succeeded, $failed failed, $skipped skipped (no backup)"

    [[ $failed -eq 0 ]]
}

# =============================================================================
# Status Reporting
# =============================================================================

# Report migration status
report_status() {
    local session_id="${1:-}"

    if [[ -n "$session_id" ]]; then
        _report_single_status "$session_id"
    else
        _report_all_status
    fi
}

_report_single_status() {
    local session_id="$1"
    local session_dir="$SESSIONS_DIR/$session_id"
    local ctx_file="$session_dir/SESSION_CONTEXT.md"
    local backup_file="$session_dir/SESSION_CONTEXT.md.v1.backup"

    if [[ ! -f "$ctx_file" ]]; then
        echo "{\"session_id\": \"$session_id\", \"exists\": false}"
        return 1
    fi

    local version
    version=$(_get_field "$ctx_file" "schema_version")
    local status
    status=$(_get_field "$ctx_file" "status")
    local has_backup="false"
    [[ -f "$backup_file" ]] && has_backup="true"

    local schema="v1"
    [[ "$version" == "2.0" || "$version" == "2.1" ]] && schema="v2"

    cat <<EOF
{
  "session_id": "$session_id",
  "exists": true,
  "schema": "$schema",
  "schema_version": "${version:-null}",
  "status": "${status:-unknown}",
  "has_backup": $has_backup,
  "can_rollback": $has_backup
}
EOF
}

_report_all_status() {
    local v1_count=0
    local v2_count=0
    local with_backup=0
    local total=0

    for session_dir in "$SESSIONS_DIR"/session-*; do
        [[ -d "$session_dir" ]] || continue

        local ctx_file="$session_dir/SESSION_CONTEXT.md"
        [[ -f "$ctx_file" ]] || continue

        ((total++))

        if _is_v1_session "$ctx_file"; then
            ((v1_count++))
        else
            ((v2_count++))
        fi

        if [[ -f "$session_dir/SESSION_CONTEXT.md.v1.backup" ]]; then
            ((with_backup++))
        fi
    done

    cat <<EOF
{
  "total_sessions": $total,
  "v1_sessions": $v1_count,
  "v2_sessions": $v2_count,
  "with_backup": $with_backup,
  "migration_needed": $v1_count
}
EOF
}

# =============================================================================
# Auto-Migration Support
# =============================================================================

# Migrate session on first access (called by session-manager.sh)
# Returns: 0 if migration succeeded or not needed, 1 on failure
auto_migrate_if_needed() {
    local session_id="$1"
    local session_dir="$SESSIONS_DIR/$session_id"
    local ctx_file="$session_dir/SESSION_CONTEXT.md"

    # Skip if not v1
    if [[ ! -f "$ctx_file" ]] || ! _is_v1_session "$ctx_file"; then
        return 0
    fi

    _log_info "Auto-migrating v1 session: $session_id"
    migrate_session "$session_id"
}

# =============================================================================
# CLI Interface
# =============================================================================

_print_help() {
    cat <<EOF
session-migrate.sh - Session schema migration (v1 -> v2)

Commands:
  migrate [options] [session_id]   Migrate session(s) to v2 schema
  rollback [options] [session_id]  Rollback to v1 from backup
  status [session_id]              Show migration status
  help                             Show this help

Options:
  --dry-run    Show what would be done without making changes
  --batch      Process all sessions (used when no session_id specified)
  --verbose    Enable debug output

Examples:
  # Check migration status
  ./session-migrate.sh status

  # Dry run migration of all v1 sessions
  ./session-migrate.sh migrate --dry-run --batch

  # Migrate all sessions
  ./session-migrate.sh migrate --batch

  # Migrate specific session
  ./session-migrate.sh migrate session-20251231-120000-abcd1234

  # Rollback specific session
  ./session-migrate.sh rollback session-20251231-120000-abcd1234

  # Rollback all sessions
  ./session-migrate.sh rollback --batch

Migration Details:
  v1 -> v2.1 Field Changes:
    - Adds: schema_version: "2.1", status: "{ACTIVE|PARKED|ARCHIVED}", team: {value|null}
    - Removes: session_state, auto_parked_at, auto_parked_reason,
               git_status_at_park, parked_git_status
    - Merges: auto_parked_at → parked_at (with parked_auto: true)
    - Preserves: All other fields (including parked_at, parked_reason) and body content
    - Creates: events.jsonl with park events (if applicable)
    - Creates: SESSION_CONTEXT.md.v1.backup

  State Derivation (v1 -> v2.1):
    - completed_at present -> ARCHIVED
    - parked_at or auto_parked_at present -> PARKED
    - Neither -> ACTIVE
EOF
}

_main() {
    local cmd="${1:-help}"
    shift || true

    # Parse global options
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --dry-run)
                MIGRATE_DRY_RUN="true"
                shift
                ;;
            --verbose|-v)
                MIGRATE_VERBOSE="true"
                shift
                ;;
            --batch)
                local batch_mode="true"
                shift
                ;;
            -*)
                _log_error "Unknown option: $1"
                _print_help
                exit 1
                ;;
            *)
                break
                ;;
        esac
    done

    local session_id="${1:-}"

    case "$cmd" in
        migrate)
            if [[ -n "$session_id" ]]; then
                migrate_session "$session_id"
            elif [[ "${batch_mode:-}" == "true" ]]; then
                migrate_all_sessions
            else
                _log_error "Specify session_id or use --batch for all sessions"
                exit 1
            fi
            ;;
        rollback)
            if [[ -n "$session_id" ]]; then
                rollback_session "$session_id"
            elif [[ "${batch_mode:-}" == "true" ]]; then
                rollback_all_sessions
            else
                _log_error "Specify session_id or use --batch for all sessions"
                exit 1
            fi
            ;;
        status)
            report_status "$session_id"
            ;;
        help|--help|-h)
            _print_help
            ;;
        *)
            _log_error "Unknown command: $cmd"
            _print_help
            exit 1
            ;;
    esac
}

# Run main if script is executed directly (not sourced)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    _main "$@"
fi

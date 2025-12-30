#!/bin/bash
# Unified session management - all operations in one place
# Usage: session-manager.sh <command> [args...]
#
# Commands:
#   status              - Output JSON with full session state
#   create <init> <complexity> [team] - Create new session
#   exists              - Exit 0 if session exists, 1 otherwise
#   tty-hash            - Output TTY hash for current terminal
#   suggest-id          - Output suggested session ID
#   cleanup             - Remove orphaned TTY mappings

set -euo pipefail

# Get script directory and source utilities
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"
cd "$PROJECT_DIR" 2>/dev/null || exit 1

# Source session utilities
# shellcheck source=session-utils.sh
source "$SCRIPT_DIR/session-utils.sh" 2>/dev/null || {
    echo '{"error": "Failed to source session-utils.sh"}' >&2
    exit 1
}

SESSIONS_DIR=".claude/sessions"
TTY_MAP_DIR="$SESSIONS_DIR/.tty-map"

# Check if current terminal has an active session
has_session() {
    local session_id
    session_id=$(get_session_id)
    [[ -n "$session_id" && -d "$SESSIONS_DIR/$session_id" ]]
}

# NOTE: is_parked() provided by session-state.sh via session-utils.sh

# Get workflow entry agent from ACTIVE_WORKFLOW.yaml
get_workflow_entry() {
    if [[ -f ".claude/ACTIVE_WORKFLOW.yaml" ]]; then
        # Entire pipeline protected: any failure returns empty string
        { grep -A2 "^entry_point:" .claude/ACTIVE_WORKFLOW.yaml 2>/dev/null | grep "agent:" | head -1 | awk '{print $2}'; } 2>/dev/null || echo ""
    else
        echo ""
    fi
}

# Get workflow name
get_workflow_name() {
    if [[ -f ".claude/ACTIVE_WORKFLOW.yaml" ]]; then
        # Entire pipeline protected: any failure returns empty string
        { grep "^name:" .claude/ACTIVE_WORKFLOW.yaml 2>/dev/null | awk '{print $2}'; } 2>/dev/null || echo ""
    else
        echo ""
    fi
}

# -----------------------------------------------------------------------------
# Helper: Extract session state from context file
# Returns: session_state initiative complexity current_phase parked (space-separated)
# -----------------------------------------------------------------------------
extract_session_fields() {
    local ctx_file="$1"
    local session_state="ACTIVE"
    local initiative=""
    local complexity=""
    local current_phase=""
    local parked="false"

    if [[ -f "$ctx_file" ]]; then
        initiative=$({ grep -m1 "^initiative:" "$ctx_file" 2>/dev/null || true; } | cut -d: -f2- | tr -d ' "')
        complexity=$({ grep -m1 "^complexity:" "$ctx_file" 2>/dev/null || true; } | cut -d: -f2- | tr -d ' "')
        current_phase=$({ grep -m1 "^current_phase:" "$ctx_file" 2>/dev/null || true; } | cut -d: -f2- | tr -d ' "')

        # Read status field (canonical), fallback to session_state (legacy) for backward compatibility
        local explicit_state=$({ grep -m1 "^status:" "$ctx_file" 2>/dev/null || grep -m1 "^session_state:" "$ctx_file" 2>/dev/null || true; } | cut -d: -f2- | tr -d ' "')
        if [[ -n "$explicit_state" ]]; then
            session_state="$explicit_state"
        fi

        # Override with PARKED if park fields present
        if grep -qE "^(parked_at|auto_parked_at):" "$ctx_file" 2>/dev/null; then
            parked="true"
            session_state="PARKED"
        fi
    fi

    echo "$session_state" "$initiative" "$complexity" "$current_phase" "$parked"
}

# -----------------------------------------------------------------------------
# Helper: Extract worktree info
# Returns: is_worktree worktree_id worktree_name worktree_team (space-separated)
# -----------------------------------------------------------------------------
extract_worktree_fields() {
    local in_worktree="false"
    local wt_id=""
    local wt_name=""
    local wt_team=""

    if is_worktree; then
        in_worktree="true"
        wt_id=$(get_worktree_field worktree_id)
        wt_name=$(get_worktree_field name)
        wt_team=$(get_worktree_field team)
    fi

    echo "$in_worktree" "$wt_id" "$wt_name" "$wt_team"
}

# -----------------------------------------------------------------------------
# Helper: Format value for JSON (null or quoted string)
# -----------------------------------------------------------------------------
json_string() {
    local val="$1"
    if [[ -z "$val" ]]; then
        echo "null"
    else
        echo "\"$val\""
    fi
}

# Output full session status as JSON
cmd_status() {
    local tty_hash
    tty_hash=$(get_tty_hash)
    local session_id
    session_id=$(get_session_id)
    local has_session="false"
    local session_state="IDLE"
    local session_dir=""
    local initiative="" complexity="" current_phase="" parked="false"

    # Session state
    if [[ -n "$session_id" && -d "$SESSIONS_DIR/$session_id" ]]; then
        has_session="true"
        session_dir="$SESSIONS_DIR/$session_id"
        read -r session_state initiative complexity current_phase parked \
            < <(extract_session_fields "$session_dir/SESSION_CONTEXT.md")
    fi

    # Team and workflow
    local active_team
    active_team=$(cat ".claude/ACTIVE_TEAM" 2>/dev/null || echo "none")
    local workflow_name
    workflow_name=$(get_workflow_name)
    local workflow_entry
    workflow_entry=$(get_workflow_entry)

    # Git status
    local git_branch
    git_branch=$(git branch --show-current 2>/dev/null || echo "not a git repo")
    local git_status_count
    git_status_count=$(git status --short 2>/dev/null | wc -l | tr -d ' ')

    # Worktree info
    local in_worktree wt_id wt_name wt_team
    read -r in_worktree wt_id wt_name wt_team < <(extract_worktree_fields)

    # Generate JSON
    cat <<EOF
{
  "tty_hash": "$tty_hash",
  "session_id": $(json_string "$session_id"),
  "session_dir": $(json_string "$session_dir"),
  "has_session": $has_session,
  "status": "$session_state",
  "initiative": $(json_string "$initiative"),
  "complexity": $(json_string "$complexity"),
  "current_phase": $(json_string "$current_phase"),
  "parked": $parked,
  "active_team": "$active_team",
  "workflow_name": "${workflow_name:-null}",
  "workflow_entry": "${workflow_entry:-null}",
  "git_branch": "$git_branch",
  "git_changes": $git_status_count,
  "is_worktree": $in_worktree,
  "worktree_id": $(json_string "$wt_id"),
  "worktree_name": $(json_string "$wt_name"),
  "worktree_team": $(json_string "$wt_team"),
  "suggested_session_id": "$(generate_session_id)",
  "sessions_dir": "$SESSIONS_DIR"
}
EOF
}

# Create new session
cmd_create() {
    local initiative="${1:-unnamed}"
    local complexity="${2:-MODULE}"
    local team="${3:-$(cat ".claude/ACTIVE_TEAM" 2>/dev/null || echo "none")}"

    # Acquire lock to prevent race conditions during session creation
    local lockfile="$SESSIONS_DIR/.create.lock"
    local lock_timeout=10
    local waited=0
    mkdir -p "$SESSIONS_DIR" 2>/dev/null
    while ! mkdir "$lockfile" 2>/dev/null; do
        if [ $waited -ge $lock_timeout ]; then
            echo '{"success": false, "error": "Timeout waiting for session lock"}' >&2
            exit 1
        fi
        sleep 1
        ((waited++))
    done
    # Ensure lock is released on exit
    trap 'rm -rf "$lockfile"' EXIT

    # Validate no existing session
    if has_session; then
        if is_parked; then
            cat >&2 <<EOF
{
  "success": false,
  "error": "Session already exists (parked). Use /continue to resume or /wrap to finalize.",
  "hint": "continue"
}
EOF
        else
            cat >&2 <<EOF
{
  "success": false,
  "error": "Session already active. Use /park first or /wrap to finalize.",
  "hint": "park"
}
EOF
        fi
        exit 1
    fi

    local session_id
    session_id=$(generate_session_id)
    local session_dir="$SESSIONS_DIR/$session_id"
    local tty_hash
    tty_hash=$(get_tty_hash)
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Create directories atomically - rollback on failure
    mkdir -p "$session_dir" || {
        echo '{"success": false, "error": "Failed to create session directory"}' >&2
        exit 1
    }

    mkdir -p "$TTY_MAP_DIR" || {
        rm -rf "$session_dir"
        echo '{"success": false, "error": "Failed to create TTY map directory"}' >&2
        exit 1
    }

    # Map TTY to session
    echo "$session_id" > "$TTY_MAP_DIR/$tty_hash" || {
        rm -rf "$session_dir"
        echo '{"success": false, "error": "Failed to create TTY mapping"}' >&2
        exit 1
    }

    # Create SESSION_CONTEXT.md
    cat > "$session_dir/SESSION_CONTEXT.md" <<CONTEXT
---
session_id: "$session_id"
created_at: "$timestamp"
initiative: "$initiative"
complexity: "$complexity"
active_team: "$team"
current_phase: "requirements"
---

# Session: $initiative

## Artifacts
- PRD: pending
- TDD: pending

## Blockers
None yet.

## Next Steps
1. Complete requirements gathering
CONTEXT

    # Validate the created file
    if ! validate_session_context "$session_dir/SESSION_CONTEXT.md" 2>/dev/null; then
        rm -rf "$session_dir"
        rm -f "$TTY_MAP_DIR/$tty_hash"
        echo '{"success": false, "error": "Failed to validate SESSION_CONTEXT.md"}' >&2
        exit 1
    fi

    # Set as current session (file-based, stable across CLI invocations)
    if ! set_current_session "$session_id"; then
        # Non-fatal: session was created, just current-session tracking failed
        # Log warning but don't fail the operation
        echo "Warning: Session created but current-session file could not be set" >&2
    fi

    # Get workflow entry for response
    local entry_agent
    entry_agent=$(get_workflow_entry)

    cat <<EOF
{
  "success": true,
  "session_id": "$session_id",
  "session_dir": "$session_dir",
  "tty_hash": "$tty_hash",
  "initiative": "$initiative",
  "complexity": "$complexity",
  "team": "$team",
  "entry_agent": "${entry_agent:-requirements-analyst}"
}
EOF
}

# Check if session exists (for conditionals)
cmd_exists() {
    if has_session; then
        echo '{"exists": true}'
        exit 0
    else
        echo '{"exists": false}'
        exit 1
    fi
}

# Output TTY hash
cmd_tty_hash() {
    echo "{\"tty_hash\": \"$(get_tty_hash)\"}"
}

# Output suggested session ID
cmd_suggest_id() {
    echo "{\"suggested_id\": \"$(generate_session_id)\"}"
}

# Cleanup stale TTY mappings
cmd_cleanup() {
    local cleaned
    cleaned=$(cleanup_stale_mappings)
    echo "{\"cleaned\": $cleaned}"
}

# Phase transition handler with artifact validation
# Usage: session-manager.sh transition <from_phase> <to_phase>
cmd_transition() {
    local from_phase="${1:-}"
    local to_phase="${2:-}"

    if [[ -z "$from_phase" || -z "$to_phase" ]]; then
        echo '{"success": false, "error": "Usage: transition <from_phase> <to_phase>"}' >&2
        exit 1
    fi

    local session_id
    session_id=$(get_session_id)
    if [[ -z "$session_id" ]]; then
        echo '{"success": false, "error": "No active session found"}' >&2
        exit 1
    fi

    local session_dir
    session_dir=$(get_session_dir)
    local ctx_file="$session_dir/SESSION_CONTEXT.md"

    if [[ ! -f "$ctx_file" ]]; then
        echo '{"success": false, "error": "SESSION_CONTEXT.md not found"}' >&2
        exit 1
    fi

    # Validate artifacts required for transition
    local missing_artifacts=()
    local docs_dir="docs"

    case "$to_phase" in
        design)
            # Transitioning to design requires approved PRD
            if ! ls $docs_dir/requirements/PRD-*.md >/dev/null 2>&1; then
                missing_artifacts+=("PRD: No PRD found in $docs_dir/requirements/")
            fi
            ;;
        implementation)
            # Transitioning to implementation requires approved TDD
            if ! ls $docs_dir/design/TDD-*.md >/dev/null 2>&1; then
                missing_artifacts+=("TDD: No TDD found in $docs_dir/design/")
            fi
            ;;
        validation)
            # Transitioning to validation requires implementation complete
            # Check that we're not in requirements or design phase
            if [[ "$from_phase" == "requirements" || "$from_phase" == "design" ]]; then
                missing_artifacts+=("Implementation: Cannot transition to validation from $from_phase")
            fi
            ;;
        complete)
            # Transitioning to complete requires test plan or validation
            if ! ls $docs_dir/testing/TP-*.md >/dev/null 2>&1; then
                missing_artifacts+=("Test Plan: No test plan found in $docs_dir/testing/ (validation incomplete)")
            fi
            ;;
    esac

    # If artifacts are missing, return error with actionable message
    if [[ ${#missing_artifacts[@]} -gt 0 ]]; then
        local error_msg="Cannot transition from '$from_phase' to '$to_phase'. Missing required artifacts:"
        local artifacts_json="["
        local first=true
        for artifact in "${missing_artifacts[@]}"; do
            if [[ "$first" == "true" ]]; then
                first=false
            else
                artifacts_json+=","
            fi
            artifacts_json+="\"$artifact\""
        done
        artifacts_json+="]"

        cat >&2 <<EOF
{
  "success": false,
  "error": "$error_msg",
  "missing_artifacts": $artifacts_json,
  "from_phase": "$from_phase",
  "to_phase": "$to_phase",
  "hint": "Complete required artifacts before transitioning to next phase"
}
EOF
        exit 1
    fi

    # Update current_phase in SESSION_CONTEXT
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Use sed to update the current_phase field in the frontmatter
    if grep -q "^current_phase:" "$ctx_file"; then
        # Update existing field
        sed -i.bak "s/^current_phase:.*$/current_phase: \"$to_phase\"/" "$ctx_file"
        rm -f "$ctx_file.bak"
    else
        # Add field after active_team (insert before closing ---)
        sed -i.bak "/^active_team:/a\\
current_phase: \"$to_phase\"
" "$ctx_file"
        rm -f "$ctx_file.bak"
    fi

    # Log transition to audit trail
    local audit_log="$session_dir/audit.log"
    echo "$timestamp | PHASE_TRANSITION | $from_phase -> $to_phase" >> "$audit_log"

    # Also append to SESSION_CONTEXT for visibility
    cat >> "$ctx_file" <<TRANSITION

## Phase Transition: $from_phase -> $to_phase
**Timestamp**: $timestamp
**Status**: Transition successful

TRANSITION

    cat <<EOF
{
  "success": true,
  "session_id": "$session_id",
  "from_phase": "$from_phase",
  "to_phase": "$to_phase",
  "timestamp": "$timestamp",
  "message": "Phase transition completed successfully"
}
EOF
}

# Mutate session with atomic operations and rollback
# Usage: session-manager.sh mutate <operation> [args...]
# Operations: park, resume, wrap, handoff
cmd_mutate() {
    local operation="${1:-}"
    shift || true

    if [[ -z "$operation" ]]; then
        echo '{"success": false, "error": "Operation required: park|resume|wrap|handoff"}' >&2
        exit 1
    fi

    # Acquire lock to prevent race conditions
    local lockfile="$SESSIONS_DIR/.mutate.lock"
    local lock_timeout=10
    local waited=0
    mkdir -p "$SESSIONS_DIR" 2>/dev/null
    while ! mkdir "$lockfile" 2>/dev/null; do
        if [ $waited -ge $lock_timeout ]; then
            echo '{"success": false, "error": "Timeout waiting for mutation lock"}' >&2
            exit 1
        fi
        sleep 1
        ((waited++))
    done
    # Ensure lock is released on exit
    trap 'rm -rf "$lockfile"' EXIT

    # Get current session
    local session_id
    session_id=$(get_session_id)
    if [[ -z "$session_id" ]]; then
        echo '{"success": false, "error": "No active session for this terminal"}' >&2
        exit 1
    fi

    local session_dir="$SESSIONS_DIR/$session_id"
    local session_file="$session_dir/SESSION_CONTEXT.md"

    if [[ ! -f "$session_file" ]]; then
        echo '{"success": false, "error": "SESSION_CONTEXT.md not found"}' >&2
        exit 1
    fi

    # Create backup before mutation
    local backup_file="$session_dir/.SESSION_CONTEXT.backup"
    cp "$session_file" "$backup_file" || {
        echo '{"success": false, "error": "Failed to create backup"}' >&2
        exit 1
    }

    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    local audit_log="$SESSIONS_DIR/.audit/session-mutations.log"
    mkdir -p "$SESSIONS_DIR/.audit" 2>/dev/null

    # Execute operation
    local result=""
    case "$operation" in
        park)
            local reason="${1:-Manual park}"
            result=$(mutate_park "$session_file" "$reason" "$timestamp")
            ;;
        resume)
            result=$(mutate_resume "$session_file" "$timestamp")
            ;;
        wrap)
            local archive="${1:-true}"
            result=$(mutate_wrap "$session_file" "$session_dir" "$archive" "$timestamp")
            ;;
        handoff)
            local from_agent="${1:-}"
            local to_agent="${2:-}"
            local notes="${3:-}"
            result=$(mutate_handoff "$session_file" "$from_agent" "$to_agent" "$notes" "$timestamp")
            ;;
        *)
            echo '{"success": false, "error": "Unknown operation: '"$operation"'"}' >&2
            rm -f "$backup_file"
            exit 1
            ;;
    esac

    # Validate result
    if ! validate_session_context "$session_file" 2>/dev/null; then
        # Rollback on validation failure
        mv "$backup_file" "$session_file"
        echo "$timestamp | $session_id | $operation | ROLLBACK | VALIDATION_FAILED" >> "$audit_log"
        echo '{"success": false, "error": "Validation failed, rolled back changes"}' >&2
        exit 1
    fi

    # Log to audit trail
    echo "$timestamp | $session_id | $operation | COMPLETE | SUCCESS" >> "$audit_log"

    # Remove backup on success
    rm -f "$backup_file"

    echo "$result"
}

# Mutation operations

mutate_park() {
    local file="$1"
    local reason="$2"
    local timestamp="$3"

    # Check not already parked
    if grep -qE "^(parked_at|auto_parked_at):" "$file" 2>/dev/null; then
        echo '{"success": false, "error": "Session already parked"}' >&2
        return 1
    fi

    # Get git status
    local git_status="clean"
    if [[ -n "$(git status --short 2>/dev/null)" ]]; then
        git_status="uncommitted changes"
    fi

    # Add park metadata to frontmatter
    # Insert before closing ---
    awk -v ts="$timestamp" -v reason="$reason" -v git="$git_status" '
        /^---$/ && ++count == 2 {
            print "parked_at: \"" ts "\""
            print "parked_reason: \"" reason "\""
            print "parked_git_status: \"" git "\""
        }
        { print }
    ' "$file" > "${file}.tmp" && mv "${file}.tmp" "$file"

    echo '{"success": true, "operation": "park", "timestamp": "'"$timestamp"'", "reason": "'"$reason"'"}'
}

mutate_resume() {
    local file="$1"
    local timestamp="$2"

    # Check is parked
    if ! grep -qE "^(parked_at|auto_parked_at):" "$file" 2>/dev/null; then
        echo '{"success": false, "error": "Session not parked"}' >&2
        return 1
    fi

    # Remove park metadata and add resumed_at
    # Delete both old (git_status_at_park, park_reason) and new (parked_git_status, parked_reason) field names
    sed -i.bak \
        -e '/^parked_at:/d' \
        -e '/^park_reason:/d' \
        -e '/^parked_reason:/d' \
        -e '/^git_status_at_park:/d' \
        -e '/^parked_git_status:/d' \
        -e '/^auto_parked_at:/d' \
        -e '/^auto_parked_reason:/d' \
        "$file" && rm -f "${file}.bak"

    # Add resumed_at
    awk -v ts="$timestamp" '
        /^---$/ && ++count == 2 {
            print "resumed_at: \"" ts "\""
        }
        { print }
    ' "$file" > "${file}.tmp" && mv "${file}.tmp" "$file"

    # Ensure file-based current session is set (may have been cleared by another session)
    local session_id
    session_id=$(grep "^session_id:" "$file" 2>/dev/null | cut -d'"' -f2)
    if [ -n "$session_id" ]; then
        set_current_session "$session_id" 2>/dev/null || true
    fi

    echo '{"success": true, "operation": "resume", "timestamp": "'"$timestamp"'"}'
}

mutate_wrap() {
    local file="$1"
    local session_dir="$2"
    local archive="$3"
    local timestamp="$4"

    # Add completed_at to frontmatter
    awk -v ts="$timestamp" '
        /^---$/ && ++count == 2 {
            print "completed_at: \"" ts "\""
        }
        { print }
    ' "$file" > "${file}.tmp" && mv "${file}.tmp" "$file"

    # Archive if requested
    local session_id
    session_id=$(basename "$session_dir")
    if [[ "$archive" == "true" ]]; then
        mkdir -p ".claude/.archive/sessions" 2>/dev/null
        # Only archive if not already there
        if [[ ! -d ".claude/.archive/sessions/$session_id" ]]; then
            mv "$session_dir" ".claude/.archive/sessions/$session_id"
        fi
    fi

    # Clear TTY mapping (legacy)
    local tty_hash
    tty_hash=$(get_tty_hash)
    rm -f "$SESSIONS_DIR/.tty-map/$tty_hash"

    # Clear file-based current session
    clear_current_session

    echo '{"success": true, "operation": "wrap", "timestamp": "'"$timestamp"'", "archived": '"$archive"'}'
}

mutate_handoff() {
    local file="$1"
    local from_agent="$2"
    local to_agent="$3"
    local notes="$4"
    local timestamp="$5"

    # Update last_agent in frontmatter
    if grep -q "^last_agent:" "$file" 2>/dev/null; then
        sed -i.bak "s/^last_agent:.*/last_agent: \"$from_agent\"/" "$file" && rm -f "${file}.bak"
    else
        awk -v agent="$from_agent" '
            /^---$/ && ++count == 2 {
                print "last_agent: \"" agent "\""
            }
            { print }
        ' "$file" > "${file}.tmp" && mv "${file}.tmp" "$file"
    fi

    # Increment handoff_count
    local current_count
    current_count=$({ grep "^handoff_count:" "$file" 2>/dev/null || true; } | awk '{print $2}')
    [[ -z "$current_count" ]] && current_count=0
    local new_count=$((current_count + 1))
    if grep -q "^handoff_count:" "$file" 2>/dev/null; then
        sed -i.bak "s/^handoff_count:.*/handoff_count: $new_count/" "$file" && rm -f "${file}.bak"
    else
        awk -v count="$new_count" '
            /^---$/ && ++count_marker == 2 {
                print "handoff_count: " count
            }
            { print }
        ' "$file" > "${file}.tmp" && mv "${file}.tmp" "$file"
    fi

    # Add handoff note to body
    local handoff_note="
## Handoff: $from_agent → $to_agent
Time: $timestamp
Notes: $notes
"
    echo "$handoff_note" >> "$file"

    echo '{"success": true, "operation": "handoff", "from": "'"$from_agent"'", "to": "'"$to_agent"'", "timestamp": "'"$timestamp"'"}'
}

# Show help
cmd_help() {
    cat <<EOF
session-manager.sh - Unified session management

Commands:
  status              Show full session state as JSON
  create <init> <complexity> [team]   Create new session
  exists              Check if session exists (exit code)
  transition <from> <to>   Transition between workflow phases with validation
  tty-hash            Show TTY hash for this terminal
  suggest-id          Generate new session ID
  cleanup             Remove orphaned TTY mappings
  mutate <op> [args]  Atomic session mutations (park|resume|wrap|handoff)
  help                Show this help

Examples:
  session-manager.sh status
  session-manager.sh create "Add dark mode" MODULE
  session-manager.sh create "New API" SERVICE 10x-dev-pack
  session-manager.sh transition requirements design
  session-manager.sh transition design implementation
  session-manager.sh mutate park "Going to lunch"
  session-manager.sh mutate resume
  session-manager.sh mutate handoff architect principal-engineer "Design approved"
  session-manager.sh cleanup
EOF
}

# Main dispatch
case "${1:-help}" in
    status)     cmd_status ;;
    create)     cmd_create "${2:-}" "${3:-}" "${4:-}" ;;
    exists)     cmd_exists ;;
    transition) cmd_transition "${2:-}" "${3:-}" ;;
    tty-hash)   cmd_tty_hash ;;
    suggest-id) cmd_suggest_id ;;
    cleanup)    cmd_cleanup ;;
    mutate)     shift; cmd_mutate "$@" ;;
    resume|park|wrap|handoff) cmd_mutate "$1" "${2:-}" "${3:-}" ;;
    help|--help|-h) cmd_help ;;
    *)
        echo "{\"error\": \"Unknown command: $1\"}" >&2
        cmd_help >&2
        exit 1
        ;;
esac

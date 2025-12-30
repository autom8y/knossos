#!/bin/bash
# Core session primitives - identification, locks, atomic operations
# Single Responsibility: Session identity resolution and low-level operations
#
# Addresses: SRP-001, ISP-001 (session-utils.sh decomposition)
# Part of Ecosystem v2 refactoring (RF-004)

# Source primitives (which sources config.sh)
# Provides: md5_portable, get_tty_hash, get_shell_session_id, atomic_write, atomic_write_stdin
# shellcheck source=$(dirname
source "$(dirname "${BASH_SOURCE[0]}")/primitives.sh"

# =============================================================================
# Session ID Resolution
# =============================================================================

# Get current session ID using priority chain:
# 1. CLAUDE_SESSION_ID environment variable (explicit override)
# 2. File-based .current-session (stable across CLI invocations)
# 3. TTY-based mapping (legacy, for backward compatibility)
get_session_id() {
  # Priority 1: Explicit environment variable (always highest priority)
  if [ -n "${CLAUDE_SESSION_ID:-}" ]; then
    echo "$CLAUDE_SESSION_ID"
    return
  fi

  # Priority 2: File-based current session (stable across CLI invocations)
  local current_session
  current_session=$(get_current_session)
  if [ -n "$current_session" ]; then
    echo "$current_session"
    return
  fi

  # Priority 3: TTY-based mapping (legacy, for backward compatibility)
  # Note: This is unreliable in Claude Code due to PPID instability
  # Retained for environments where TERM_SESSION_ID is stable (e.g., VS Code)
  local tty_hash=$(get_tty_hash)
  local project_dir="${CLAUDE_PROJECT_DIR:-.}"
  local tty_map="$project_dir/.claude/sessions/.tty-map/$tty_hash"
  if [ -f "$tty_map" ]; then
    cat "$tty_map"
    return
  fi

  # No session found
  echo ""
}

# Get session directory path for current session
get_session_dir() {
  local sid=$(get_session_id)
  if [ -n "$sid" ]; then
    echo ".claude/sessions/$sid"
  else
    echo ""
  fi
}

# Get path to .current-session file
get_current_session_file() {
    local project_dir="${CLAUDE_PROJECT_DIR:-.}"
    echo "$project_dir/.claude/sessions/.current-session"
}

# Generate a new unique session ID
generate_session_id() {
  echo "session-$(date +%Y%m%d-%H%M%S)-$(openssl rand -hex 4)"
}

# =============================================================================
# File-Based Current Session (Primary Session Association)
# =============================================================================
# These functions provide stable session association that persists across
# Claude Code invocations. Unlike TTY-based mapping (which changes with each
# `claude` CLI run), file-based association uses the filesystem as stable state.

# Set the current session for this project
# Usage: set_current_session "session-20251227-145523-dafb3260"
# Returns: 0 on success, 1 on failure
set_current_session() {
    local session_id="$1"

    if [ -z "$session_id" ]; then
        echo "Error: session_id required" >&2
        return 1
    fi

    # Validate session ID format
    if [[ ! "$session_id" =~ ^session-[0-9]{8}-[0-9]{6}-[a-f0-9]{8}$ ]]; then
        echo "Error: Invalid session_id format: $session_id" >&2
        return 1
    fi

    local project_dir="${CLAUDE_PROJECT_DIR:-.}"
    local sessions_dir="$project_dir/.claude/sessions"
    local current_file="$sessions_dir/.current-session"

    # Ensure sessions directory exists
    mkdir -p "$sessions_dir" 2>/dev/null || {
        echo "Error: Cannot create sessions directory" >&2
        return 1
    }

    # Write atomically using existing atomic_write function
    if ! atomic_write "$current_file" "$session_id"; then
        echo "Error: Failed to write current session file" >&2
        return 1
    fi

    return 0
}

# Get the current session for this project
# Returns: Session ID to stdout, or empty string if none/invalid
get_current_session() {
    local project_dir="${CLAUDE_PROJECT_DIR:-.}"
    local current_file="$project_dir/.claude/sessions/.current-session"

    # Check file exists
    if [ ! -f "$current_file" ]; then
        echo ""
        return 0
    fi

    # Read and trim content
    local session_id
    session_id=$(cat "$current_file" 2>/dev/null | tr -d '[:space:]')

    # Validate non-empty
    if [ -z "$session_id" ]; then
        echo ""
        return 0
    fi

    # Validate format
    if [[ ! "$session_id" =~ ^session-[0-9]{8}-[0-9]{6}-[a-f0-9]{8}$ ]]; then
        # Invalid format - clear stale file
        rm -f "$current_file" 2>/dev/null
        echo ""
        return 0
    fi

    # Validate session directory exists
    local session_dir="$project_dir/.claude/sessions/$session_id"
    if [ ! -d "$session_dir" ]; then
        # Session was deleted/archived - clear stale pointer
        rm -f "$current_file" 2>/dev/null
        echo ""
        return 0
    fi

    echo "$session_id"
}

# Clear the current session for this project
# Returns: 0 always (rm -f semantics)
clear_current_session() {
    local project_dir="${CLAUDE_PROJECT_DIR:-.}"
    local current_file="$project_dir/.claude/sessions/.current-session"
    rm -f "$current_file" 2>/dev/null
    return 0
}

# =============================================================================
# TTY-Based Session Mapping (Legacy)
# =============================================================================
# These functions are retained for backward compatibility but file-based
# association (above) is preferred.

# Map TTY to session (legacy)
# Usage: map_tty_to_session "session-id"
map_tty_to_session() {
  local session_id="$1"
  local tty_hash=$(get_tty_hash)
  local project_dir="${CLAUDE_PROJECT_DIR:-.}"
  mkdir -p "$project_dir/.claude/sessions/.tty-map"
  echo "$session_id" > "$project_dir/.claude/sessions/.tty-map/$tty_hash"
}

# Alias for backward compatibility
set_session_for_tty() {
  map_tty_to_session "$@"
}

# Remove TTY mapping (legacy)
unmap_tty() {
  local tty_hash=$(get_tty_hash)
  local project_dir="${CLAUDE_PROJECT_DIR:-.}"
  rm -f "$project_dir/.claude/sessions/.tty-map/$tty_hash"
}

# Alias for backward compatibility
clear_session_for_tty() {
  unmap_tty
}

# =============================================================================
# Session State Checks
# =============================================================================

# Check if a session exists and is active (not parked)
# Usage: is_session_active [session_id]
# Returns: 0 if active, 1 if not active or doesn't exist
is_session_active() {
    local session_id="${1:-$(get_session_id)}"

    if [ -z "$session_id" ]; then
        return 1
    fi

    local project_dir="${CLAUDE_PROJECT_DIR:-.}"
    local session_dir="$project_dir/.claude/sessions/$session_id"
    local session_file="$session_dir/SESSION_CONTEXT.md"

    # Check session directory exists
    if [ ! -d "$session_dir" ]; then
        return 1
    fi

    # Check SESSION_CONTEXT.md exists
    if [ ! -f "$session_file" ]; then
        return 1
    fi

    # Check if parked (parked_at or auto_parked_at field exists)
    if grep -qE "^(parked_at|auto_parked_at):" "$session_file" 2>/dev/null; then
        return 1  # Session is parked, not active
    fi

    return 0  # Session is active
}

# =============================================================================
# File Locking
# =============================================================================
# Note: LOCK_DIR and LOCK_TIMEOUT are defined in config.sh

# Acquire a session lock
# Usage: acquire_session_lock "lock_name" [timeout_seconds]
# Returns: 0 on success, 1 on timeout/failure
acquire_session_lock() {
    local lock_name="$1"
    local timeout="${2:-$LOCK_TIMEOUT}"
    local lock_file="$LOCK_DIR/${lock_name}.lock"

    mkdir -p "$LOCK_DIR" 2>/dev/null || return 1

    # Try flock if available (preferred method)
    if command -v flock >/dev/null 2>&1; then
        # Create lock file descriptor and try to acquire lock
        exec 200>"$lock_file"
        if flock -w "$timeout" 200 2>/dev/null; then
            # Store PID for debugging
            echo "$$" >&200
            return 0
        else
            exec 200>&-
            return 1
        fi
    fi

    # Fallback: mkdir-based locking (portable but less elegant)
    local lock_marker="$lock_file.d"
    local elapsed=0
    local sleep_interval=0.1

    while [ "$elapsed" -lt "$timeout" ]; do
        # mkdir is atomic - if it succeeds, we have the lock
        if mkdir "$lock_marker" 2>/dev/null; then
            # Store PID for stale lock detection
            echo "$$" > "$lock_marker/pid"
            return 0
        fi

        # Check if existing lock is stale (owner process dead)
        if [ -f "$lock_marker/pid" ]; then
            local owner_pid
            owner_pid=$(cat "$lock_marker/pid" 2>/dev/null)
            if [ -n "$owner_pid" ] && ! kill -0 "$owner_pid" 2>/dev/null; then
                # Owner is dead, remove stale lock
                rm -rf "$lock_marker" 2>/dev/null
                continue
            fi
        fi

        sleep "$sleep_interval"
        elapsed=$((elapsed + 1))  # Approximate, each sleep ~0.1s
        [ "$elapsed" -ge $((timeout * 10)) ] && break
    done

    return 1  # Timeout
}

# Release a session lock
# Usage: release_session_lock "lock_name"
release_session_lock() {
    local lock_name="$1"
    local lock_file="$LOCK_DIR/${lock_name}.lock"
    local lock_marker="$lock_file.d"

    # Release flock if we're using file descriptor
    if command -v flock >/dev/null 2>&1; then
        exec 200>&- 2>/dev/null || true
    fi

    # Remove mkdir-based lock if present
    rm -rf "$lock_marker" 2>/dev/null || true

    return 0
}

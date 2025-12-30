#!/bin/bash
# Platform-portable primitives - no dependencies except config
# Pure utility functions with no session-specific logic
#
# Addresses: SRP-001 (partial - session-utils.sh decomposition)
# Part of Ecosystem v2 refactoring (RF-002)

# Source configuration first
source "$(dirname "${BASH_SOURCE[0]}")/config.sh"

# =============================================================================
# Portable Hash Functions
# =============================================================================

# Portable MD5 hash function (works on macOS and Linux)
md5_portable() {
  local input="$1"
  if command -v md5 >/dev/null 2>&1; then
    # macOS
    echo -n "$input" | md5 -q
  elif command -v md5sum >/dev/null 2>&1; then
    # Linux
    echo -n "$input" | md5sum | cut -d' ' -f1
  else
    # Fallback: use first 32 chars of base64
    echo -n "$input" | base64 | tr -d '\n' | cut -c1-32
  fi
}

# =============================================================================
# YAML Parsing
# =============================================================================

# Get YAML field from frontmatter (portable, handles quotes and colons in values)
# Usage: get_yaml_field "file.md" "field_name"
get_yaml_field() {
  local file="$1"
  local field="$2"
  [ -f "$file" ] || return 1

  # Use yq if available (most reliable), otherwise grep-based parsing
  if command -v yq >/dev/null 2>&1; then
    # Extract frontmatter (first YAML document) and get field
    # Uses mikefarah/yq v4 syntax with document_index selector
    yq "select(document_index == 0) | .$field // \"\"" "$file" 2>/dev/null
  else
    # Fallback: grep-based parsing with proper quote handling
    # Handles: field: value, field: "value", field: 'value'
    grep -m1 "^${field}:" "$file" 2>/dev/null | \
      sed "s/^${field}:[[:space:]]*//" | \
      sed 's/^["'"'"']//' | \
      sed 's/["'"'"']$//' | \
      tr -d '\r'
  fi
}

# =============================================================================
# TTY/Terminal Identification
# =============================================================================

# Get or create a unique session identifier for the current shell
# This persists for the life of the shell and prevents TTY reuse collisions
get_shell_session_id() {
  # If we already have a shell session ID, use it
  if [ -n "${_CLAUDE_SHELL_SID:-}" ]; then
    echo "$_CLAUDE_SHELL_SID"
    return
  fi

  # Generate a new one based on shell PID and start time
  # This is unique per shell instance, even if TTY is reused
  local shell_pid="${PPID:-$$}"
  local shell_start=""

  # Get shell start time (cross-platform)
  if [ "$(uname)" = "Darwin" ]; then
    # macOS: use ps to get process start time
    shell_start=$(ps -p "$shell_pid" -o lstart= 2>/dev/null | tr -d ' ' || echo "")
  else
    # Linux: use /proc for process start time
    if [ -f "/proc/$shell_pid/stat" ]; then
      shell_start=$(cut -d' ' -f22 "/proc/$shell_pid/stat" 2>/dev/null || echo "")
    fi
  fi

  # Combine PID + start time for uniqueness
  local combined="${shell_pid}-${shell_start:-$(date +%s)}"
  echo "$combined"
}

# Get TTY hash for terminal identification
# Includes shell session ID to prevent collision on terminal reuse
get_tty_hash() {
  # Use TTY path or terminal session ID, hashed for filesystem safety
  # Include shell session ID (PID + start time) to prevent collision on terminal reuse
  #
  # The hash includes:
  # - TTY device path (or TERM_SESSION_ID)
  # - Parent PID (shell PID)
  # - Shell start time (via get_shell_session_id)
  #
  # This ensures that even if a terminal window is closed and reopened with
  # the same TTY path, the hash will be different because either:
  # - The shell PID is different, OR
  # - The shell start time is different
  local tty_id="${TTY:-${TERM_SESSION_ID:-unknown}}"
  local shell_session
  shell_session=$(get_shell_session_id)
  md5_portable "${tty_id}-${shell_session}"
}

# =============================================================================
# Atomic File Operations
# =============================================================================

# Atomic write: temp file + mv pattern for safe file updates
# Usage: atomic_write "destination_file" "content"
# Returns: 0 on success, 1 on failure
atomic_write() {
    local dest_file="$1"
    local content="$2"
    local dest_dir
    dest_dir=$(dirname "$dest_file")

    # Ensure destination directory exists
    mkdir -p "$dest_dir" 2>/dev/null || {
        echo "Error: Cannot create directory $dest_dir" >&2
        return 1
    }

    # Create temp file in same directory (ensures same filesystem for atomic mv)
    local temp_file
    temp_file=$(mktemp "${dest_dir}/.tmp.XXXXXX") || {
        echo "Error: Cannot create temp file in $dest_dir" >&2
        return 1
    }

    # Write content to temp file
    if ! printf '%s' "$content" > "$temp_file" 2>/dev/null; then
        rm -f "$temp_file" 2>/dev/null
        echo "Error: Cannot write to temp file" >&2
        return 1
    fi

    # Atomic move (on POSIX systems, mv within same filesystem is atomic)
    if ! mv "$temp_file" "$dest_file" 2>/dev/null; then
        rm -f "$temp_file" 2>/dev/null
        echo "Error: Cannot move temp file to $dest_file" >&2
        return 1
    fi

    return 0
}

# =============================================================================
# JSON Extraction
# =============================================================================

# Extract value from JSON string with automatic jq/grep fallback
# Usage: json_extract "$json_string" ".path.to.field"
# Returns: extracted value or empty string
json_extract() {
    local json="$1"
    local path="$2"

    # Try jq first (fast and reliable)
    if command -v jq >/dev/null 2>&1; then
        echo "$json" | jq -r "$path // empty" 2>/dev/null || echo ""
        return
    fi

    # Fallback: grep-based for simple paths
    # Handles: .field_name (top-level) and .parent.child (nested)
    local field="${path##*.}"  # Get last component
    echo "$json" | grep -o "\"$field\": *\"[^\"]*\"" 2>/dev/null | head -1 | cut -d'"' -f4 || echo ""
}

# =============================================================================
# Hook Permission Helpers
# =============================================================================

# Output PreToolUse auto-approve JSON and exit
# Usage: auto_approve "reason" [log_function]
# Note: Calls exit 0 - does not return
auto_approve() {
    local reason="$1"
    local log_func="${2:-}"

    cat <<EOF
{
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "allow",
    "permissionDecisionReason": "$reason"
  }
}
EOF

    # Call optional log function (e.g., log_end)
    if [[ -n "$log_func" ]] && declare -F "$log_func" >/dev/null 2>&1; then
        "$log_func" 0 2>/dev/null || true
    fi

    exit 0
}

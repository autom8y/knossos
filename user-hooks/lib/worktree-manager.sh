#!/bin/bash
# Worktree management for per-session isolation
# Usage: worktree-manager.sh <command> [args...]
#
# Lifecycle Commands:
#   create [name] [--team=PACK]  - Create new worktree with full isolation
#   list                          - List all worktrees with status
#   status [id]                   - Detailed status of worktree(s)
#   remove <id>                   - Remove specific worktree
#   cleanup [--force]             - Remove stale worktrees
#   gc                            - Garbage collect (prune orphaned refs)
#
# Merge/Transfer Commands:
#   diff <id> [--to=BRANCH]       - Preview changes (excludes .claude/)
#   merge <id> [--to=BRANCH]      - Merge worktree to target branch
#   cherry-pick <id> <commits...> - Apply specific commits

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Source configuration (provides SKELETON_HOME, ROSTER_HOME, etc.)
# shellcheck source=config.sh
source "$SCRIPT_DIR/config.sh"

# Source session-state for is_worktree (with fallback)
# shellcheck source=session-state.sh
source "$SCRIPT_DIR/session-state.sh" 2>/dev/null || {
    # Minimal fallback if session-state.sh unavailable
    is_worktree() {
        local git_dir
        git_dir=$(git rev-parse --git-dir 2>/dev/null) || return 1
        [[ -f "$git_dir" ]] && grep -q "^gitdir:" "$git_dir" 2>/dev/null
    }
}

# Ensure we're in a git repository
ensure_git_repo() {
    if ! git rev-parse --git-dir >/dev/null 2>&1; then
        echo '{"error": "Not a git repository"}' >&2
        exit 1
    fi
}

# Get project root (main worktree)
get_project_root() {
    git rev-parse --show-toplevel 2>/dev/null
}

# Generate worktree ID
generate_worktree_id() {
    echo "wt-$(date +%Y%m%d-%H%M%S)-$(openssl rand -hex 4)"
}

# Get worktrees directory
get_worktrees_dir() {
    local root
    root=$(get_project_root)
    echo "$root/worktrees"
}

# =============================================================================
# Helper Functions for Merge/Diff/Cherry-pick Operations
# =============================================================================

# Resolve worktree ID to path, validating existence
# Returns: worktree path on success, exits with JSON error on failure
resolve_worktree() {
    local wt_id="$1"
    local wt_dir
    wt_dir=$(get_worktrees_dir)
    local wt_path="$wt_dir/$wt_id"

    # Validate ID format
    if [[ ! "$wt_id" =~ ^wt-[0-9]{8}-[0-9]{6}-[a-f0-9]+$ ]]; then
        echo '{"error": "Invalid worktree ID format: '"$wt_id"'. Expected: wt-YYYYMMDD-HHMMSS-hex"}' >&2
        return 1
    fi

    # Check directory exists
    if [[ ! -d "$wt_path" ]]; then
        echo '{"error": "Worktree not found: '"$wt_id"'"}' >&2
        return 1
    fi

    # Verify it's a valid git worktree
    if ! git -C "$wt_path" rev-parse --git-dir >/dev/null 2>&1; then
        echo '{"error": "Invalid git worktree: '"$wt_id"'"}' >&2
        return 1
    fi

    echo "$wt_path"
    return 0
}

# Check for uncommitted changes in worktree
# Returns: 0 if clean, 1 if dirty (with count in output)
check_uncommitted() {
    local wt_path="$1"
    local status_output

    status_output=$(git -C "$wt_path" status --porcelain 2>/dev/null)

    if [[ -n "$status_output" ]]; then
        local count
        count=$(echo "$status_output" | wc -l | tr -d ' ')
        echo "$count"
        return 1
    fi

    echo "0"
    return 0
}

# Count .claude/ changes between target and worktree HEAD
# Arguments: target_ref, worktree_head_ref
# Returns: count of files changed in .claude/
count_claude_changes() {
    local target="$1"
    local wt_head="$2"

    local count
    count=$(git diff --name-only "$target".."$wt_head" -- .claude/ 2>/dev/null | wc -l | tr -d ' ')
    echo "$count"
}

# Get pathspec exclusion string based on --include-claude flag
# Arguments: include_claude (true/false)
# Returns: pathspec string (empty if include, exclusion pattern if not)
get_pathspec_exclusion() {
    local include_claude="${1:-false}"

    if [[ "$include_claude" == "true" ]]; then
        echo ""
    else
        echo ":(exclude).claude/"
    fi
}

# Ensure worktrees directory exists with .gitignore
ensure_worktrees_dir() {
    local wt_dir
    wt_dir=$(get_worktrees_dir)
    mkdir -p "$wt_dir"

    if [[ ! -f "$wt_dir/.gitignore" ]]; then
        echo '*' > "$wt_dir/.gitignore"
    fi
}

# Create new worktree with full isolation
cmd_create() {
    ensure_git_repo

    # CRITICAL: Prevent nested worktree creation
    if is_worktree; then
        echo '{"error": "Cannot create worktree from within a worktree. Navigate to main project first."}' >&2
        exit 1
    fi

    local name=""
    local team=""
    local from_ref="HEAD"
    local complexity="MODULE"

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --team=*)
                team="${1#--team=}"
                shift
                ;;
            --from=*)
                from_ref="${1#--from=}"
                shift
                ;;
            --complexity=*)
                complexity="${1#--complexity=}"
                shift
                ;;
            -*)
                echo '{"error": "Unknown option: '"$1"'"}' >&2
                exit 1
                ;;
            *)
                name="$1"
                shift
                ;;
        esac
    done

    # Default name if not provided
    name="${name:-unnamed}"

    # Validate --from=REF exists before proceeding
    if ! git rev-parse "$from_ref" >/dev/null 2>&1; then
        echo '{"error": "Invalid git ref: '"$from_ref"'. Branch, tag, or commit not found."}' >&2
        exit 1
    fi

    # Get current team if not specified
    if [[ -z "$team" ]]; then
        local root
        root=$(get_project_root)
        team=$(cat "$root/.claude/ACTIVE_TEAM" 2>/dev/null || echo "")
    fi

    local wt_id
    wt_id=$(generate_worktree_id)
    local wt_dir
    wt_dir=$(get_worktrees_dir)
    local wt_path="$wt_dir/$wt_id"
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    ensure_worktrees_dir

    # Create worktree with detached HEAD (no branch pollution)
    if ! git worktree add --detach "$wt_path" "$from_ref" 2>&1; then
        echo '{"error": "Failed to create git worktree from ref '"$from_ref"'"}' >&2
        exit 1
    fi

    # Sync CEM in worktree (worktrees inherit .claude/ from parent, so use sync not init)
    if [[ ! -x "$SKELETON_HOME/cem" ]]; then
        git worktree remove --force "$wt_path" 2>/dev/null || true
        echo '{"error": "CEM not found at '"$SKELETON_HOME/cem"'. Cannot create worktree without ecosystem. Run: chmod +x $SKELETON_HOME/cem"}' >&2
        exit 1
    fi

    # Worktrees inherit .claude/ from git, so check if CEM is already initialized
    # If yes, use sync. If no manifest exists, use init.
    local cem_output
    if [[ -f "$wt_path/.claude/.cem/manifest.json" ]]; then
        # Already initialized - just sync to get latest
        if ! cem_output=$(cd "$wt_path" && "$SKELETON_HOME/cem" sync 2>&1); then
            # Sync failed - try force reinit
            if ! cem_output=$(cd "$wt_path" && "$SKELETON_HOME/cem" init --force 2>&1); then
                git worktree remove --force "$wt_path" 2>/dev/null || true
                echo '{"error": "CEM sync/init failed in worktree. Details: '"${cem_output:-unknown}"'"}' >&2
                exit 1
            fi
        fi
    else
        # Not initialized - run init
        if ! cem_output=$(cd "$wt_path" && "$SKELETON_HOME/cem" init 2>&1); then
            git worktree remove --force "$wt_path" 2>/dev/null || true
            echo '{"error": "CEM init failed in worktree. Details: '"${cem_output:-unknown}"'"}' >&2
            exit 1
        fi
    fi

    # Set team if specified
    if [[ -n "$team" && "$team" != "none" ]]; then
        if [[ -x "$ROSTER_HOME/swap-team.sh" ]]; then
            (cd "$wt_path" && "$ROSTER_HOME/swap-team.sh" "$team" 2>/dev/null) || {
                echo '{"warning": "Failed to set team '"$team"'. Worktree created with default team."}' >&2
            }
        fi
    fi

    # Create worktree metadata
    mkdir -p "$wt_path/.claude"
    cat > "$wt_path/.claude/.worktree-meta.json" <<EOF
{
    "worktree_id": "$wt_id",
    "created_at": "$timestamp",
    "name": "$name",
    "from_ref": "$from_ref",
    "team": "${team:-none}",
    "complexity": "$complexity",
    "parent_project": "$(get_project_root)"
}
EOF

    # Create initial session in worktree
    if [[ -x "$wt_path/.claude/hooks/lib/session-manager.sh" ]]; then
        (cd "$wt_path" && .claude/hooks/lib/session-manager.sh create "$name" "$complexity" "${team:-none}" 2>/dev/null) || true
    fi

    # Output success
    cat <<EOF
{
    "success": true,
    "worktree_id": "$wt_id",
    "path": "$wt_path",
    "name": "$name",
    "team": "${team:-none}",
    "from_ref": "$from_ref",
    "instructions": "cd $wt_path && claude"
}
EOF
}

# List all worktrees
cmd_list() {
    ensure_git_repo

    local wt_dir
    wt_dir=$(get_worktrees_dir)
    local count=0

    echo "{"
    echo '  "worktrees": ['

    local first=true
    for wt in "$wt_dir"/wt-*; do
        [[ -d "$wt" ]] || continue
        ((count++)) || true

        local wt_id
        wt_id=$(basename "$wt")
        local meta_file="$wt/.claude/.worktree-meta.json"
        local name="unknown"
        local team="unknown"
        local created_at="unknown"

        if [[ -f "$meta_file" ]]; then
            name=$({ grep -o '"name": *"[^"]*"' "$meta_file" 2>/dev/null | cut -d'"' -f4; } || echo "unknown")
            team=$({ grep -o '"team": *"[^"]*"' "$meta_file" 2>/dev/null | cut -d'"' -f4; } || echo "unknown")
            created_at=$({ grep -o '"created_at": *"[^"]*"' "$meta_file" 2>/dev/null | cut -d'"' -f4; } || echo "unknown")
        fi

        # Check for uncommitted changes
        local has_changes="false"
        if (cd "$wt" && ! git diff --quiet 2>/dev/null); then
            has_changes="true"
        fi

        # Check session status
        local session_status="none"
        local session_dir
        session_dir=$(find "$wt/.claude/sessions" -maxdepth 1 -type d -name "session-*" 2>/dev/null | head -1)
        if [[ -n "$session_dir" ]]; then
            if grep -q "parked_at:\|auto_parked_at:" "$session_dir/SESSION_CONTEXT.md" 2>/dev/null; then
                session_status="parked"
            else
                session_status="active"
            fi
        fi

        [[ "$first" == "true" ]] || echo ","
        first=false

        cat <<EOF
    {
      "id": "$wt_id",
      "path": "$wt",
      "name": "$name",
      "team": "$team",
      "created_at": "$created_at",
      "has_changes": $has_changes,
      "session_status": "$session_status"
    }
EOF
    done

    echo ""
    echo "  ],"
    echo "  \"count\": $count"
    echo "}"
}

# Detailed status
cmd_status() {
    ensure_git_repo

    local target_id="${1:-}"
    local wt_dir
    wt_dir=$(get_worktrees_dir)

    if [[ -n "$target_id" ]]; then
        # Specific worktree
        local wt_path="$wt_dir/$target_id"
        if [[ ! -d "$wt_path" ]]; then
            echo '{"error": "Worktree not found: '"$target_id"'"}' >&2
            exit 1
        fi

        local meta_file="$wt_path/.claude/.worktree-meta.json"
        if [[ -f "$meta_file" ]]; then
            cat "$meta_file"
        else
            echo '{"worktree_id": "'"$target_id"'", "path": "'"$wt_path"'", "metadata": "missing"}'
        fi
    else
        # All worktrees summary
        cmd_list
    fi
}

# Remove specific worktree
cmd_remove() {
    ensure_git_repo

    local target_id="${1:-}"
    local force="${2:-}"

    if [[ -z "$target_id" ]]; then
        echo '{"error": "Worktree ID required. Use: worktree-manager.sh remove <id>"}' >&2
        exit 1
    fi

    local wt_dir
    wt_dir=$(get_worktrees_dir)
    local wt_path="$wt_dir/$target_id"

    if [[ ! -d "$wt_path" ]]; then
        echo '{"error": "Worktree not found: '"$target_id"'"}' >&2
        exit 1
    fi

    # Check for uncommitted changes
    if (cd "$wt_path" && ! git diff --quiet 2>/dev/null); then
        if [[ "$force" != "--force" ]]; then
            echo '{"error": "Worktree has uncommitted changes. Use --force to override.", "worktree_id": "'"$target_id"'"}' >&2
            exit 1
        fi
    fi

    # Remove worktree
    if git worktree remove ${force:+--force} "$wt_path" 2>/dev/null; then
        echo '{"success": true, "removed": "'"$target_id"'"}'
    else
        echo '{"error": "Failed to remove worktree"}' >&2
        exit 1
    fi
}

# Cleanup stale worktrees
cmd_cleanup() {
    ensure_git_repo

    local force="${1:-}"
    local wt_dir
    wt_dir=$(get_worktrees_dir)
    local cutoff_days=7
    local cutoff_seconds=$((cutoff_days * 24 * 60 * 60))
    local now
    now=$(date +%s)
    local removed=0
    local skipped=0
    local skipped_reasons=""

    for wt in "$wt_dir"/wt-*; do
        [[ -d "$wt" ]] || continue

        local wt_id
        wt_id=$(basename "$wt")

        # Get modification time
        local mtime
        if [[ "$(uname)" == "Darwin" ]]; then
            mtime=$(stat -f %m "$wt" 2>/dev/null || echo "$now")
        else
            mtime=$(stat -c %Y "$wt" 2>/dev/null || echo "$now")
        fi

        local age=$((now - mtime))

        if [[ $age -gt $cutoff_seconds ]]; then
            local skip_reason=""

            # Check for uncommitted tracked changes
            if ! (cd "$wt" && git diff --quiet 2>/dev/null); then
                skip_reason="uncommitted changes"
            fi

            # Check for untracked files (CRITICAL: prevents data loss)
            if [[ -z "$skip_reason" ]]; then
                local untracked
                untracked=$(cd "$wt" && git status --porcelain 2>/dev/null | { grep -c "^??" || echo "0"; })
                if [[ "$untracked" -gt 0 ]]; then
                    skip_reason="$untracked untracked files"
                fi
            fi

            # Check for active (non-parked) sessions
            if [[ -z "$skip_reason" ]]; then
                local session_dir
                session_dir=$(find "$wt/.claude/sessions" -maxdepth 1 -type d -name "session-*" 2>/dev/null | head -1)
                if [[ -n "$session_dir" && -f "$session_dir/SESSION_CONTEXT.md" ]]; then
                    if ! grep -qE "^(parked_at|auto_parked_at):" "$session_dir/SESSION_CONTEXT.md" 2>/dev/null; then
                        skip_reason="active session"
                    fi
                fi
            fi

            # Apply force override or skip
            if [[ -n "$skip_reason" && "$force" != "--force" ]]; then
                ((skipped++)) || true
                skipped_reasons="${skipped_reasons}${wt_id}: ${skip_reason}\n"
            else
                if git worktree remove ${force:+--force} "$wt" 2>/dev/null; then
                    ((removed++)) || true
                else
                    ((skipped++)) || true
                    skipped_reasons="${skipped_reasons}${wt_id}: removal failed\n"
                fi
            fi
        fi
    done

    # Run git worktree prune
    git worktree prune 2>/dev/null || true

    # Output with skip reasons if any
    if [[ -n "$skipped_reasons" && "$skipped" -gt 0 ]]; then
        echo '{"removed": '$removed', "skipped": '$skipped', "cutoff_days": '$cutoff_days', "skip_reasons": "'"$(echo -e "$skipped_reasons" | tr '\n' '; ' | sed 's/; $//')"'"}'
    else
        echo '{"removed": '$removed', "skipped": '$skipped', "cutoff_days": '$cutoff_days'}'
    fi
}

# Garbage collection
cmd_gc() {
    ensure_git_repo

    # Prune orphaned worktree refs
    git worktree prune 2>/dev/null || true

    # Count remaining
    local wt_dir
    wt_dir=$(get_worktrees_dir)
    local count=0
    for wt in "$wt_dir"/wt-*; do
        [[ -d "$wt" ]] && ((count++)) || true
    done

    echo '{"pruned": true, "remaining_worktrees": '$count'}'
}

# =============================================================================
# Merge/Diff/Cherry-pick Commands
# =============================================================================

# Show diff between worktree and target branch
cmd_diff() {
    ensure_git_repo

    # CRITICAL: Prevent running from within a worktree
    if is_worktree; then
        echo '{"error": "Cannot diff from within a worktree. Navigate to main project first."}' >&2
        exit 1
    fi

    local wt_id=""
    local target="main"
    local include_claude="false"

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --to=*)
                target="${1#--to=}"
                shift
                ;;
            --include-claude)
                include_claude="true"
                shift
                ;;
            -*)
                echo '{"error": "Unknown option: '"$1"'"}' >&2
                exit 1
                ;;
            *)
                if [[ -z "$wt_id" ]]; then
                    wt_id="$1"
                fi
                shift
                ;;
        esac
    done

    # Validate worktree ID provided
    if [[ -z "$wt_id" ]]; then
        echo '{"error": "Worktree ID required. Use: worktree-manager.sh diff <id> [--to=BRANCH]"}' >&2
        exit 1
    fi

    # Resolve worktree
    local wt_path
    if ! wt_path=$(resolve_worktree "$wt_id"); then
        exit 1
    fi

    # Verify target branch exists
    if ! git rev-parse --verify "$target" >/dev/null 2>&1; then
        echo '{"error": "Target branch '"'$target'"' not found. Create it first or use --to to specify existing branch."}' >&2
        exit 1
    fi

    # Get worktree HEAD
    local wt_head
    wt_head=$(git -C "$wt_path" rev-parse HEAD 2>/dev/null)
    if [[ -z "$wt_head" ]]; then
        echo '{"error": "Could not determine worktree HEAD"}' >&2
        exit 1
    fi

    # Count .claude/ changes for reporting
    local claude_changes=0
    if [[ "$include_claude" == "false" ]]; then
        claude_changes=$(count_claude_changes "$target" "$wt_head")
    fi

    # Build and execute diff command
    local pathspec
    pathspec=$(get_pathspec_exclusion "$include_claude")

    # Output diff to stdout
    if [[ -n "$pathspec" ]]; then
        git diff "$target"..."$wt_head" -- . "$pathspec" 2>/dev/null
    else
        git diff "$target"..."$wt_head" 2>/dev/null
    fi

    # Calculate summary statistics
    local stat_output
    if [[ -n "$pathspec" ]]; then
        stat_output=$(git diff --stat "$target"..."$wt_head" -- . "$pathspec" 2>/dev/null | tail -1)
    else
        stat_output=$(git diff --stat "$target"..."$wt_head" 2>/dev/null | tail -1)
    fi

    # Parse stats (format: "X files changed, Y insertions(+), Z deletions(-)")
    local files_changed=0
    local insertions=0
    local deletions=0

    if [[ -n "$stat_output" ]]; then
        files_changed=$(echo "$stat_output" | grep -oE '[0-9]+ files? changed' | grep -oE '[0-9]+' || echo "0")
        insertions=$(echo "$stat_output" | grep -oE '[0-9]+ insertions?' | grep -oE '[0-9]+' || echo "0")
        deletions=$(echo "$stat_output" | grep -oE '[0-9]+ deletions?' | grep -oE '[0-9]+' || echo "0")
    fi

    # Ensure numeric values
    files_changed="${files_changed:-0}"
    insertions="${insertions:-0}"
    deletions="${deletions:-0}"

    # Output summary JSON to stderr
    local exclusion_note=""
    if [[ "$claude_changes" -gt 0 && "$include_claude" == "false" ]]; then
        exclusion_note="$claude_changes files in .claude/ excluded from merge (use --include-claude to include)"
    fi

    cat >&2 <<EOF
{
    "worktree_id": "$wt_id",
    "target": "$target",
    "files_changed": $files_changed,
    "insertions": $insertions,
    "deletions": $deletions,
    "claude_files_excluded": $claude_changes,
    "claude_exclusion_note": "$exclusion_note"
}
EOF
}

# Merge worktree commits to target branch
cmd_merge() {
    ensure_git_repo

    # CRITICAL: Prevent running from within a worktree
    if is_worktree; then
        echo '{"error": "Cannot merge from within a worktree. Navigate to main project first."}' >&2
        exit 1
    fi

    local wt_id=""
    local target="main"
    local include_claude="false"
    local no_cleanup="false"
    local force="false"
    local yes="false"
    local dry_run="false"

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --to=*)
                target="${1#--to=}"
                shift
                ;;
            --include-claude)
                include_claude="true"
                shift
                ;;
            --no-cleanup)
                no_cleanup="true"
                shift
                ;;
            --force)
                force="true"
                shift
                ;;
            --yes)
                yes="true"
                shift
                ;;
            --dry-run)
                dry_run="true"
                shift
                ;;
            -*)
                echo '{"error": "Unknown option: '"$1"'"}' >&2
                exit 1
                ;;
            *)
                if [[ -z "$wt_id" ]]; then
                    wt_id="$1"
                fi
                shift
                ;;
        esac
    done

    # Validate worktree ID provided
    if [[ -z "$wt_id" ]]; then
        echo '{"error": "Worktree ID required. Use: worktree-manager.sh merge <id> [--to=BRANCH]"}' >&2
        exit 1
    fi

    # Resolve worktree
    local wt_path
    if ! wt_path=$(resolve_worktree "$wt_id"); then
        exit 1
    fi

    # Check for uncommitted changes (use if/! to avoid errexit)
    local uncommitted_count
    if ! uncommitted_count=$(check_uncommitted "$wt_path"); then
        # Has uncommitted changes
        if [[ "$force" == "true" ]]; then
            # Discard uncommitted changes
            git -C "$wt_path" reset --hard HEAD 2>/dev/null
            git -C "$wt_path" clean -fd 2>/dev/null
        else
            echo '{"error": "Worktree has uncommitted changes. Commit or stash before merging.", "worktree_id": "'"$wt_id"'", "uncommitted_files": '"$uncommitted_count"'}' >&2
            exit 1
        fi
    fi

    # Verify target branch exists
    if ! git rev-parse --verify "$target" >/dev/null 2>&1; then
        echo '{"error": "Target branch '"'$target'"' not found. Create it first or use --to to specify existing branch."}' >&2
        exit 1
    fi

    # Get worktree HEAD
    local wt_head
    wt_head=$(git -C "$wt_path" rev-parse HEAD 2>/dev/null)
    if [[ -z "$wt_head" ]]; then
        echo '{"error": "Could not determine worktree HEAD"}' >&2
        exit 1
    fi

    # Count commits to be merged (from merge-base to wt_head)
    local merge_base
    merge_base=$(git merge-base "$target" "$wt_head" 2>/dev/null)
    local commits_count=0
    if [[ -n "$merge_base" ]]; then
        commits_count=$(git rev-list --count "$merge_base".."$wt_head" 2>/dev/null || echo "0")
    fi

    # Count .claude/ changes for reporting
    local claude_changes=0
    if [[ "$include_claude" == "false" ]]; then
        claude_changes=$(count_claude_changes "$target" "$wt_head")
    fi

    # Count files to be changed (excluding .claude/ if not included)
    local files_count=0
    local pathspec
    pathspec=$(get_pathspec_exclusion "$include_claude")
    if [[ -n "$pathspec" ]]; then
        files_count=$(git diff --name-only "$target"..."$wt_head" -- . "$pathspec" 2>/dev/null | wc -l | tr -d ' ')
    else
        files_count=$(git diff --name-only "$target"..."$wt_head" 2>/dev/null | wc -l | tr -d ' ')
    fi

    # Include-claude confirmation (unless --yes)
    if [[ "$include_claude" == "true" && "$yes" == "false" && "$dry_run" == "false" ]]; then
        echo '{"warning": "--include-claude specified. This will propagate session-specific artifacts to main. Use --yes to confirm.", "worktree_id": "'"$wt_id"'", "claude_files": '"$claude_changes"'}' >&2
        exit 1
    fi

    # Dry run output
    if [[ "$dry_run" == "true" ]]; then
        cat <<EOF
{
    "dry_run": true,
    "command": "merge",
    "worktree_id": "$wt_id",
    "would_merge_to": "$target",
    "commits_to_merge": $commits_count,
    "files_to_change": $files_count,
    "claude_files_excluded": $claude_changes,
    "would_remove_worktree": $([ "$no_cleanup" == "true" ] && echo "false" || echo "true")
}
EOF
        exit 0
    fi

    # Store current branch to return to it on failure
    local original_branch
    original_branch=$(git branch --show-current 2>/dev/null || git rev-parse --short HEAD)

    # Perform the merge
    # Step 1: Checkout target branch
    if ! git checkout "$target" 2>/dev/null; then
        echo '{"error": "Failed to checkout target branch '"'$target'"'"}' >&2
        exit 1
    fi

    # Step 2: Merge with no-commit, no-ff (per ADR-0001)
    if ! git merge --no-commit --no-ff "$wt_head" 2>/dev/null; then
        # Merge conflict - abort and restore
        git merge --abort 2>/dev/null || true
        git checkout "$original_branch" 2>/dev/null || true
        echo '{"error": "Merge conflict occurred. Resolve manually or use cherry-pick for selective commits.", "worktree_id": "'"$wt_id"'", "target": "'"$target"'"}' >&2
        exit 1
    fi

    # Step 3: Exclude .claude/ if not included (per ADR-0002)
    if [[ "$include_claude" == "false" ]]; then
        # Reset .claude/ to target branch state
        git reset HEAD -- .claude/ 2>/dev/null || true
        git checkout -- .claude/ 2>/dev/null || true
    fi

    # Step 4: Commit the merge
    local commit_msg="Merge worktree $wt_id"
    if [[ "$include_claude" == "false" && "$claude_changes" -gt 0 ]]; then
        commit_msg="$commit_msg (excluding $claude_changes .claude/ files)"
    fi

    if ! git commit -m "$commit_msg" 2>/dev/null; then
        # Check if there's nothing to commit (all changes were in .claude/)
        if git diff --cached --quiet 2>/dev/null; then
            # Nothing to commit - still a success, just no changes to merge
            git merge --abort 2>/dev/null || true
        else
            git merge --abort 2>/dev/null || true
            git checkout "$original_branch" 2>/dev/null || true
            echo '{"error": "Failed to commit merge"}' >&2
            exit 1
        fi
    fi

    # Step 5: Cleanup worktree (unless --no-cleanup)
    local worktree_removed="false"
    local cleanup_note=""
    if [[ "$no_cleanup" == "false" ]]; then
        if git worktree remove --force "$wt_path" 2>/dev/null; then
            worktree_removed="true"
        else
            cleanup_note="cleanup skipped: worktree removal failed"
        fi
    else
        cleanup_note="cleanup skipped per --no-cleanup"
    fi

    # Output success
    cat <<EOF
{
    "success": true,
    "command": "merge",
    "worktree_id": "$wt_id",
    "merged_to": "$target",
    "commits_merged": $commits_count,
    "files_changed": $files_count,
    "claude_files_excluded": $claude_changes,
    "worktree_removed": $worktree_removed,
    "note": "$cleanup_note"
}
EOF
}

# Cherry-pick specific commits from worktree
cmd_cherry_pick() {
    ensure_git_repo

    # CRITICAL: Prevent running from within a worktree
    if is_worktree; then
        echo '{"error": "Cannot cherry-pick from within a worktree. Navigate to main project first."}' >&2
        exit 1
    fi

    local wt_id=""
    local include_claude="false"
    local yes="false"
    local commits=()

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --include-claude)
                include_claude="true"
                shift
                ;;
            --yes)
                yes="true"
                shift
                ;;
            -*)
                echo '{"error": "Unknown option: '"$1"'"}' >&2
                exit 1
                ;;
            *)
                if [[ -z "$wt_id" ]]; then
                    wt_id="$1"
                else
                    commits+=("$1")
                fi
                shift
                ;;
        esac
    done

    # Validate worktree ID provided
    if [[ -z "$wt_id" ]]; then
        echo '{"error": "Worktree ID required. Use: worktree-manager.sh cherry-pick <id> <commit...>"}' >&2
        exit 1
    fi

    # Validate commits provided
    if [[ ${#commits[@]} -eq 0 ]]; then
        echo '{"error": "At least one commit required. Use: worktree-manager.sh cherry-pick <id> <commit...>"}' >&2
        exit 1
    fi

    # Resolve worktree
    local wt_path
    if ! wt_path=$(resolve_worktree "$wt_id"); then
        exit 1
    fi

    # Validate all commits exist in worktree
    for commit in "${commits[@]}"; do
        if ! git -C "$wt_path" cat-file -t "$commit" >/dev/null 2>&1; then
            echo '{"error": "Commit '"'$commit'"' not found in worktree history"}' >&2
            exit 1
        fi

        # Check if it's a merge commit
        local parent_count
        parent_count=$(git -C "$wt_path" cat-file -p "$commit" 2>/dev/null | grep -c "^parent " || echo "0")
        if [[ "$parent_count" -gt 1 ]]; then
            echo '{"error": "Cannot cherry-pick merge commit '"'$commit'"'. Use --mainline to specify parent."}' >&2
            exit 1
        fi
    done

    # Include-claude confirmation (unless --yes)
    if [[ "$include_claude" == "true" && "$yes" == "false" ]]; then
        echo '{"warning": "--include-claude specified. This will propagate session-specific artifacts. Use --yes to confirm.", "worktree_id": "'"$wt_id"'"}' >&2
        exit 1
    fi

    # Apply each commit
    local applied_commits=()
    local total_claude_excluded=0

    for commit in "${commits[@]}"; do
        # Get commit info
        local short_hash
        local subject
        short_hash=$(git -C "$wt_path" rev-parse --short "$commit" 2>/dev/null)
        subject=$(git -C "$wt_path" log -1 --format="%s" "$commit" 2>/dev/null)

        # Cherry-pick with no-commit
        if ! git cherry-pick --no-commit "$commit" 2>/dev/null; then
            # Conflict - abort and report
            git cherry-pick --abort 2>/dev/null || true
            local conflicting_files
            conflicting_files=$(git diff --name-only --diff-filter=U 2>/dev/null | head -5 | tr '\n' ', ' | sed 's/,$//')

            cat >&2 <<EOF
{
    "error": "Cherry-pick conflict",
    "conflict": true,
    "commit": "$short_hash",
    "subject": "$subject",
    "conflicting_files": "$conflicting_files",
    "guidance": "Resolve conflicts manually then run: git cherry-pick --continue"
}
EOF
            exit 1
        fi

        # Exclude .claude/ if not included
        local commit_claude_changes=0
        if [[ "$include_claude" == "false" ]]; then
            # Check if this commit has .claude/ changes
            commit_claude_changes=$(git diff --cached --name-only -- .claude/ 2>/dev/null | wc -l | tr -d ' ')
            if [[ "$commit_claude_changes" -gt 0 ]]; then
                git reset HEAD -- .claude/ 2>/dev/null || true
                git checkout -- .claude/ 2>/dev/null || true
                ((total_claude_excluded += commit_claude_changes)) || true
            fi
        fi

        # Commit with original message
        if ! git commit -C "$commit" 2>/dev/null; then
            # Check if there's nothing to commit (all changes were in .claude/)
            if git diff --cached --quiet 2>/dev/null; then
                # Skip this commit - it only had .claude/ changes
                applied_commits+=("{\"hash\": \"$short_hash\", \"subject\": \"$subject\", \"skipped\": true, \"reason\": \"only .claude/ changes\"}")
                continue
            else
                git cherry-pick --abort 2>/dev/null || true
                echo '{"error": "Failed to commit cherry-pick for '"$short_hash"'"}' >&2
                exit 1
            fi
        fi

        applied_commits+=("{\"hash\": \"$short_hash\", \"subject\": \"$subject\"}")
    done

    # Build applied commits JSON array
    local commits_json="["
    local first=true
    for c in "${applied_commits[@]}"; do
        if [[ "$first" == "true" ]]; then
            first=false
        else
            commits_json+=", "
        fi
        commits_json+="$c"
    done
    commits_json+="]"

    # Output success
    cat <<EOF
{
    "success": true,
    "command": "cherry-pick",
    "worktree_id": "$wt_id",
    "commits_applied": $commits_json,
    "claude_files_excluded": $total_claude_excluded
}
EOF
}

# Show help
cmd_help() {
    cat <<EOF
worktree-manager.sh - Per-session worktree isolation

Commands:
  create [name] [--team=PACK] [--from=REF]   Create isolated worktree
  list                                        List all worktrees
  status [id]                                 Detailed status
  remove <id> [--force]                       Remove worktree
  cleanup [--force]                           Remove stale (7+ days)
  gc                                          Prune orphaned refs

Merge/Transfer Commands:
  diff <id> [--to=BRANCH] [--include-claude]
      Preview changes between worktree and target branch.
      Excludes .claude/ by default. Output: diff to stdout, summary JSON to stderr.

  merge <id> [--to=BRANCH] [--include-claude] [--no-cleanup] [--force] [--yes] [--dry-run]
      Merge worktree commits to target branch (default: main).
      Excludes .claude/ by default. Auto-removes worktree after merge.
      --force: Discard uncommitted changes
      --no-cleanup: Keep worktree after merge
      --dry-run: Show what would happen without executing

  cherry-pick <id> <commit...> [--include-claude] [--yes]
      Apply specific commits from worktree to current branch.
      Excludes .claude/ from cherry-picked commits by default.

Common Options:
  --include-claude   Include .claude/ directory in operation (requires --yes)
  --to=BRANCH        Target branch (default: main)
  --yes              Skip confirmation prompts
  --force            Override safety checks

  help                                        Show this help

Examples:
  worktree-manager.sh create "auth-sprint" --team=10x-dev-pack
  worktree-manager.sh list
  worktree-manager.sh diff wt-20251224-143052-abc
  worktree-manager.sh merge wt-20251224-143052-abc --to=develop
  worktree-manager.sh cherry-pick wt-20251224-143052-abc abc1234 def5678
  worktree-manager.sh remove wt-20251224-143052-abc
  worktree-manager.sh cleanup --force
EOF
}

# Main dispatch
case "${1:-help}" in
    create)     shift; cmd_create "$@" ;;
    list)       cmd_list ;;
    status)     cmd_status "${2:-}" ;;
    remove)     cmd_remove "${2:-}" "${3:-}" ;;
    cleanup)    cmd_cleanup "${2:-}" ;;
    gc)         cmd_gc ;;
    diff)       shift; cmd_diff "$@" ;;
    merge)      shift; cmd_merge "$@" ;;
    cherry-pick) shift; cmd_cherry_pick "$@" ;;
    help|--help|-h) cmd_help ;;
    *)
        echo '{"error": "Unknown command: '"$1"'"}' >&2
        cmd_help >&2
        exit 1
        ;;
esac

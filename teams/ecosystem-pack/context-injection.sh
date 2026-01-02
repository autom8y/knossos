#!/bin/bash
# Ecosystem-Pack Context Injection
# Provides CEM sync status, skeleton reference, and drift detection for ecosystem work
#
# Called by: session-context.sh via team-context-loader.sh
# Output: Markdown table with ecosystem status

# Required function name (per team-context-loader.sh contract)
inject_team_context() {
    local project_dir="${CLAUDE_PROJECT_DIR:-.}"
    local output=""

    # Start table
    output="| | |"$'\n'
    output+="|---|---|"$'\n'

    # CEM Sync Status
    local cem_sync_file="$project_dir/.claude/.cem-sync"
    local cem_status="unknown"
    local cem_timestamp="never"

    if [[ -f "$cem_sync_file" ]]; then
        cem_timestamp=$(stat -f "%Sm" -t "%Y-%m-%d %H:%M" "$cem_sync_file" 2>/dev/null || \
                        stat -c "%y" "$cem_sync_file" 2>/dev/null | cut -d'.' -f1 || \
                        echo "unknown")
        # Check staleness (>24h = stale)
        if is_file_stale "$cem_sync_file" 1440 2>/dev/null; then
            cem_status="stale"
        else
            cem_status="synced"
        fi
    else
        cem_status="never synced"
    fi
    output+="| **CEM Sync** | $cem_status ($cem_timestamp) |"$'\n'

    # Skeleton Reference
    local skeleton_ref="unknown"
    local skeleton_home="${SKELETON_HOME:-$HOME/Code/skeleton_claude}"
    if [[ -d "$skeleton_home/.git" ]]; then
        skeleton_ref=$(cd "$skeleton_home" && git rev-parse --short HEAD 2>/dev/null || echo "unknown")
        local skeleton_branch=$(cd "$skeleton_home" && git branch --show-current 2>/dev/null || echo "detached")
        skeleton_ref="$skeleton_branch@$skeleton_ref"
    fi
    output+="| **Skeleton Ref** | $skeleton_ref |"$'\n'

    # Drift Detection
    local drift_status="clean"
    # Check if local .claude/ differs from skeleton (simplified check)
    if [[ -f "$project_dir/.claude/.local-overrides" ]]; then
        local override_count=$(wc -l < "$project_dir/.claude/.local-overrides" 2>/dev/null | tr -d ' ')
        drift_status="$override_count local overrides"
    fi
    output+="| **Drift Status** | $drift_status |"$'\n'

    # Test Satellites (for compatibility testing context)
    local satellites_dir="${ROSTER_HOME:-$HOME/Code/roster}/test-satellites"
    local satellite_count=0
    if [[ -d "$satellites_dir" ]]; then
        satellite_count=$(ls -1d "$satellites_dir"/*/ 2>/dev/null | wc -l | tr -d ' ')
    fi
    output+="| **Test Satellites** | $satellite_count available |"$'\n'

    echo "$output"
}

# Helper function (provided by team-context-loader.sh, but define fallback)
if ! declare -f is_file_stale >/dev/null 2>&1; then
    is_file_stale() {
        local file="$1"
        local max_age_minutes="${2:-60}"
        [[ ! -f "$file" ]] && return 0
        local now file_time age_seconds max_age_seconds
        now=$(date +%s)
        file_time=$(stat -f %m "$file" 2>/dev/null || stat -c %Y "$file" 2>/dev/null || echo 0)
        age_seconds=$((now - file_time))
        max_age_seconds=$((max_age_minutes * 60))
        [[ $age_seconds -gt $max_age_seconds ]]
    }
fi

#!/bin/bash
# Ecosystem-Pack Context Injection
# Provides CEM sync status, roster reference, and drift detection for ecosystem work
#
# Called by: session-context.sh via rite-context-loader.sh
# Output: Markdown table with ecosystem status

# Required function name (per rite-context-loader.sh contract)
inject_rite_context() {
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

    # Roster Reference
    local roster_ref="unknown"
    local roster_home="${ROSTER_HOME:-$HOME/Code/roster}"
    if [[ -d "$roster_home/.git" ]]; then
        roster_ref=$(cd "$roster_home" && git rev-parse --short HEAD 2>/dev/null || echo "unknown")
        local roster_branch=$(cd "$roster_home" && git branch --show-current 2>/dev/null || echo "detached")
        roster_ref="$roster_branch@$roster_ref"
    fi
    output+="| **Roster Ref** | $roster_ref |"$'\n'

    # Drift Detection
    local drift_status="clean"
    # Check if local .claude/ differs from roster (simplified check)
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

# Helper function (provided by rite-context-loader.sh, but define fallback)
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

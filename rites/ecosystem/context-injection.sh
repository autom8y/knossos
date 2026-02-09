#!/bin/bash
# Ecosystem Rite Context Injection
# Provides sync status, knossos reference, and drift detection for ecosystem work
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

    # Sync Status
    local sync_file="$project_dir/.claude/.sync-timestamp"
    local sync_status="unknown"
    local sync_timestamp="never"

    if [[ -f "$sync_file" ]]; then
        sync_timestamp=$(stat -f "%Sm" -t "%Y-%m-%d %H:%M" "$sync_file" 2>/dev/null || \
                        stat -c "%y" "$sync_file" 2>/dev/null | cut -d'.' -f1 || \
                        echo "unknown")
        # Check staleness (>24h = stale)
        if is_file_stale "$sync_file" 1440 2>/dev/null; then
            sync_status="stale"
        else
            sync_status="synced"
        fi
    else
        sync_status="never synced"
    fi
    output+="| **Sync Status** | $sync_status ($sync_timestamp) |"$'\n'

    # Knossos Reference
    local knossos_ref="unknown"
    local knossos_home="${KNOSSOS_HOME:-$HOME/Code/knossos}"
    if [[ -d "$knossos_home/.git" ]]; then
        knossos_ref=$(cd "$knossos_home" && git rev-parse --short HEAD 2>/dev/null || echo "unknown")
        local knossos_branch=$(cd "$knossos_home" && git branch --show-current 2>/dev/null || echo "detached")
        knossos_ref="$knossos_branch@$knossos_ref"
    fi
    output+="| **Knossos Ref** | $knossos_ref |"$'\n'

    # Drift Detection
    local drift_status="clean"
    # Check if local .claude/ differs from knossos (simplified check)
    if [[ -f "$project_dir/.claude/.local-overrides" ]]; then
        local override_count=$(wc -l < "$project_dir/.claude/.local-overrides" 2>/dev/null | tr -d ' ')
        drift_status="$override_count local overrides"
    fi
    output+="| **Drift Status** | $drift_status |"$'\n'

    # Test Satellites (for compatibility testing context)
    local satellites_dir="${KNOSSOS_HOME:-$HOME/Code/knossos}/test-satellites"
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

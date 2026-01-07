#!/usr/bin/env bash
#
# merge-settings.sh - Settings JSON Union Merge
#
# Implements union merge semantics for settings.local.json.
# Preserves satellite-specific permissions while updating from roster.
#
# Part of: roster-sync (TDD-cem-replacement)
#
# Usage:
#   source "$ROSTER_HOME/lib/sync/merge/merge-settings.sh"
#   merge_settings_json "$knossos_file" "$local_file" "$output_file"
#
# Merge Behavior:
#   - Uses roster file as base
#   - Adds satellite-specific permissions
#   - Adds satellite-specific directories
#   - Adds satellite-specific MCP servers
#   - Preserves enableAllProjectMcpServers if set locally

# Guard against re-sourcing
[[ -n "${_MERGE_SETTINGS_LOADED:-}" ]] && return 0
readonly _MERGE_SETTINGS_LOADED=1

# ============================================================================
# Settings Merge (per TDD 5.3)
# ============================================================================

# Merge settings.local.json with union semantics
# Algorithm:
#   1. If no local file: copy roster as-is
#   2. Extract satellite extras (permissions, directories, MCP servers)
#   3. Merge: roster base + satellite extras
#   4. Preserve local enableAllProjectMcpServers if set
#
# Usage: merge_settings_json "knossos_file" "local_file" "output_file"
merge_settings_json() {
    local knossos_file="$1"
    local local_file="$2"
    local output_file="$3"

    # If no local file, copy roster as-is
    if [[ ! -f "$local_file" ]]; then
        sync_log_debug "merge-settings: no local file, copying roster"
        cp "$knossos_file" "$output_file"
        return 0
    fi

    # Validate both files are valid JSON
    if ! jq -e . "$knossos_file" >/dev/null 2>&1; then
        sync_log_error "Invalid JSON in roster file: $knossos_file"
        return 1
    fi

    if ! jq -e . "$local_file" >/dev/null 2>&1; then
        sync_log_error "Invalid JSON in local file: $local_file"
        return 1
    fi

    # Extract satellite-specific permissions (in local but not in roster)
    local extra_perms
    extra_perms=$(jq -n --slurpfile r "$knossos_file" --slurpfile l "$local_file" '
        ($l[0].permissions.allow // []) - ($r[0].permissions.allow // [])
    ')

    # Extract satellite-specific directories
    local extra_dirs
    extra_dirs=$(jq -n --slurpfile r "$knossos_file" --slurpfile l "$local_file" '
        ($l[0].permissions.additionalDirectories // []) -
        ($r[0].permissions.additionalDirectories // [])
    ')

    # Extract satellite-specific MCP servers
    local extra_mcp
    extra_mcp=$(jq -n --slurpfile r "$knossos_file" --slurpfile l "$local_file" '
        ($l[0].enabledMcpjsonServers // []) -
        ($r[0].enabledMcpjsonServers // [])
    ')

    sync_log_debug "merge-settings: extra_perms=$extra_perms"
    sync_log_debug "merge-settings: extra_dirs=$extra_dirs"
    sync_log_debug "merge-settings: extra_mcp=$extra_mcp"

    # Merge: roster base + satellite extras
    jq --argjson ep "$extra_perms" \
       --argjson ed "$extra_dirs" \
       --argjson em "$extra_mcp" \
       --slurpfile l "$local_file" '
        # Add extra permissions (union + unique)
        .permissions.allow = ((.permissions.allow // []) + $ep | unique) |

        # Add extra directories (union + unique)
        .permissions.additionalDirectories = ((.permissions.additionalDirectories // []) + $ed | unique) |

        # Add extra MCP servers (union + unique)
        .enabledMcpjsonServers = ((.enabledMcpjsonServers // []) + $em | unique) |

        # Preserve local enableAllProjectMcpServers if set
        if $l[0].enableAllProjectMcpServers then
            .enableAllProjectMcpServers = $l[0].enableAllProjectMcpServers
        else . end
    ' "$knossos_file" > "$output_file" || {
        sync_log_error "merge-settings: jq merge failed"
        return 1
    }

    sync_log_debug "merge-settings: complete"
    return 0
}

# ============================================================================
# Settings Validation
# ============================================================================

# Validate settings.local.json structure
# Returns: 0 if valid, 1 if invalid
validate_settings_json() {
    local file="$1"

    if [[ ! -f "$file" ]]; then
        return 1
    fi

    # Check it's valid JSON
    if ! jq -e . "$file" >/dev/null 2>&1; then
        return 1
    fi

    # Check for expected structure (at least permissions object)
    if ! jq -e '.permissions' "$file" >/dev/null 2>&1; then
        sync_log_warning "Settings file missing permissions object: $file"
        # Not a fatal error, structure can be added
    fi

    return 0
}

# ============================================================================
# Settings Extraction
# ============================================================================

# Extract satellite-specific settings for reporting
# Usage: get_satellite_settings "$local_file" "$knossos_file"
# Outputs: JSON object with satellite-specific additions
get_satellite_settings() {
    local local_file="$1"
    local knossos_file="$2"

    if [[ ! -f "$local_file" || ! -f "$knossos_file" ]]; then
        echo "{}"
        return 1
    fi

    jq -n --slurpfile r "$knossos_file" --slurpfile l "$local_file" '{
        extra_permissions: (($l[0].permissions.allow // []) - ($r[0].permissions.allow // [])),
        extra_directories: (($l[0].permissions.additionalDirectories // []) - ($r[0].permissions.additionalDirectories // [])),
        extra_mcp_servers: (($l[0].enabledMcpjsonServers // []) - ($r[0].enabledMcpjsonServers // [])),
        custom_mcp_enabled: ($l[0].enableAllProjectMcpServers // false)
    }'
}

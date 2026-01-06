#!/usr/bin/env bash
#
# sync-manifest.sh - Manifest Management and Migration
#
# Handles reading, writing, and migrating manifest.json files.
# Supports schema versions 1, 2, and 3 with automatic migration.
#
# Part of: roster-sync (TDD-cem-replacement)
#
# Usage:
#   source "$ROSTER_HOME/lib/sync/sync-manifest.sh"
#   manifest=$(read_manifest)
#   migrate_manifest_if_needed
#   write_manifest "$manifest"
#
# Functions:
#   read_manifest              - Read and parse manifest file
#   write_manifest             - Write manifest with atomic operation
#   get_manifest_field         - Get a specific field from manifest
#   get_manifest_checksum      - Get checksum for a managed file
#   set_manifest_checksum      - Update checksum for a managed file
#   migrate_manifest_if_needed - Migrate v1/v2 to v3
#   validate_manifest          - Check manifest structure
#   create_manifest            - Create new v3 manifest

# Guard against re-sourcing
[[ -n "${_SYNC_MANIFEST_LOADED:-}" ]] && return 0
readonly _SYNC_MANIFEST_LOADED=1

# ============================================================================
# Manifest Reading
# ============================================================================

# Read manifest file and return JSON
# Usage: manifest=$(read_manifest)
# Returns: JSON string or empty on error
read_manifest() {
    local manifest_file="${SYNC_MANIFEST_FILE:-.claude/.cem/manifest.json}"

    if [[ ! -f "$manifest_file" ]]; then
        sync_log_debug "Manifest not found: $manifest_file"
        echo ""
        return 1
    fi

    local manifest
    manifest=$(cat "$manifest_file" 2>/dev/null)

    if ! echo "$manifest" | jq -e . >/dev/null 2>&1; then
        sync_log_error "Invalid JSON in manifest: $manifest_file"
        return 1
    fi

    echo "$manifest"
}

# Get a specific field from manifest
# Usage: get_manifest_field ".roster.commit"
# Returns: field value or empty
get_manifest_field() {
    local field="$1"
    local manifest

    manifest=$(read_manifest) || return 1
    echo "$manifest" | jq -r "$field // empty"
}

# Get the schema version from manifest
# Returns: 1, 2, or 3 (defaults to 1 for legacy manifests)
get_manifest_version() {
    local manifest="${1:-}"

    if [[ -z "$manifest" ]]; then
        manifest=$(read_manifest) || {
            echo "0"
            return 1
        }
    fi

    local version
    version=$(echo "$manifest" | jq -r '.schema_version // 1')
    echo "$version"
}

# Get checksum for a specific managed file
# Usage: get_manifest_checksum "COMMAND_REGISTRY.md"
# Returns: checksum or empty
get_manifest_checksum() {
    local filename="$1"
    local manifest

    manifest=$(read_manifest) || return 1

    # Try v3 format first (nested managed_files)
    local checksum
    checksum=$(echo "$manifest" | jq -r --arg f ".claude/$filename" \
        '.managed_files[] | select(.path == $f) | .checksum // empty')

    if [[ -z "$checksum" ]]; then
        # Try v1/v2 format
        checksum=$(echo "$manifest" | jq -r --arg f "$filename" \
            '.managed_files[] | select(.path == $f or .path == ".claude/" + $f) | .checksum // empty')
    fi

    echo "$checksum"
}

# ============================================================================
# Manifest Writing
# ============================================================================

# Write manifest with atomic operation
# Usage: write_manifest "$manifest_json"
write_manifest() {
    local manifest="$1"
    local manifest_file="${SYNC_MANIFEST_FILE:-.claude/.cem/manifest.json}"
    local manifest_dir

    manifest_dir=$(dirname "$manifest_file")
    if [[ ! -d "$manifest_dir" ]]; then
        mkdir -p "$manifest_dir" || {
            sync_log_error "Cannot create manifest directory: $manifest_dir"
            return 1
        }
    fi

    # Validate JSON before writing
    if ! echo "$manifest" | jq -e . >/dev/null 2>&1; then
        sync_log_error "Cannot write invalid JSON to manifest"
        return 1
    fi

    # Atomic write using temp file
    local temp_file="${manifest_file}.tmp.$$"

    echo "$manifest" | jq '.' > "$temp_file" || {
        rm -f "$temp_file"
        sync_log_error "Failed to write temp manifest"
        return 1
    }

    mv "$temp_file" "$manifest_file" || {
        rm -f "$temp_file"
        sync_log_error "Failed to rename temp manifest"
        return 1
    }

    sync_log_debug "Manifest written: $manifest_file"
}

# Update checksum for a managed file in manifest
# Usage: set_manifest_checksum "COMMAND_REGISTRY.md" "abc123..."
set_manifest_checksum() {
    local filename="$1"
    local checksum="$2"
    local manifest

    manifest=$(read_manifest) || return 1

    local path=".claude/$filename"
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    manifest=$(echo "$manifest" | jq --arg p "$path" --arg c "$checksum" --arg t "$timestamp" '
        .managed_files = [.managed_files[] |
            if .path == $p then
                .checksum = $c | .last_sync = $t
            else . end]')

    write_manifest "$manifest"
}

# ============================================================================
# Manifest Creation
# ============================================================================

# Create a new v3 manifest
# Usage: create_manifest "/path/to/roster" "commit_hash" "branch_ref"
create_manifest() {
    local roster_path="$1"
    local roster_commit="$2"
    local roster_ref="${3:-main}"
    local timestamp

    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    local manifest
    manifest=$(jq -n \
        --argjson sv "$SYNC_SCHEMA_VERSION" \
        --arg rp "$roster_path" \
        --arg rc "$roster_commit" \
        --arg rr "$roster_ref" \
        --arg ts "$timestamp" '{
        schema_version: $sv,
        roster: {
            path: $rp,
            commit: $rc,
            ref: $rr,
            last_sync: $ts
        },
        team: null,
        managed_files: [],
        orphans: []
    }')

    echo "$manifest"
}

# Add a managed file to manifest
# Usage: manifest=$(add_managed_file "$manifest" "path" "strategy" "checksum")
add_managed_file() {
    local manifest="$1"
    local path="$2"
    local strategy="$3"
    local checksum="$4"
    local timestamp

    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    echo "$manifest" | jq \
        --arg p "$path" \
        --arg s "$strategy" \
        --arg c "$checksum" \
        --arg t "$timestamp" '
        .managed_files += [{
            path: $p,
            strategy: $s,
            checksum: $c,
            source: "roster",
            added_at: $t,
            last_sync: $t
        }]'
}

# Update roster info in manifest
# Usage: manifest=$(update_manifest_roster "$manifest" "commit" "ref")
update_manifest_roster() {
    local manifest="$1"
    local commit="$2"
    local ref="${3:-}"
    local timestamp

    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    local updated
    updated=$(echo "$manifest" | jq \
        --arg c "$commit" \
        --arg t "$timestamp" '
        .roster.commit = $c |
        .roster.last_sync = $t')

    if [[ -n "$ref" ]]; then
        updated=$(echo "$updated" | jq --arg r "$ref" '.roster.ref = $r')
    fi

    echo "$updated"
}

# ============================================================================
# Manifest Version Validation
# ============================================================================

# Check manifest version and require v3
# Returns: 0 if v3, 1 on error (legacy versions no longer supported)
migrate_manifest_if_needed() {
    local manifest_file="${SYNC_MANIFEST_FILE:-.claude/.cem/manifest.json}"

    if [[ ! -f "$manifest_file" ]]; then
        sync_log_debug "No manifest to migrate"
        return 0
    fi

    local manifest schema_version
    manifest=$(cat "$manifest_file")
    schema_version=$(echo "$manifest" | jq -r '.schema_version // 1')

    case "$schema_version" in
        3)
            sync_log_debug "Manifest at v3"
            return 0
            ;;
        *)
            sync_log_error "Legacy manifest v$schema_version found. Delete .claude/.cem/manifest.json and run 'roster-sync init' to create fresh v3 manifest."
            return 1
            ;;
    esac
}

# ============================================================================
# Manifest Validation
# ============================================================================

# Validate manifest structure and integrity
# Returns: 0 valid, 1 warnings, 2 errors
validate_manifest() {
    local manifest_file="${SYNC_MANIFEST_FILE:-.claude/.cem/manifest.json}"
    local warnings=0
    local errors=0

    # Check file exists
    if [[ ! -f "$manifest_file" ]]; then
        sync_log_error "Manifest not found: $manifest_file"
        return 2
    fi

    # Check valid JSON
    local manifest
    manifest=$(cat "$manifest_file")
    if ! echo "$manifest" | jq -e . >/dev/null 2>&1; then
        sync_log_error "Invalid JSON in manifest"
        return 2
    fi

    # Check schema version
    local version
    version=$(echo "$manifest" | jq -r '.schema_version // 1')
    if [[ "$version" -lt 1 || "$version" -gt 3 ]]; then
        sync_log_error "Invalid schema version: $version"
        ((errors++)) || true
    fi

    # Check required fields based on version
    case "$version" in
        3)
            local roster_path
            roster_path=$(echo "$manifest" | jq -r '.roster.path // empty')
            if [[ -z "$roster_path" ]]; then
                sync_log_warning "Missing roster.path"
                ((warnings++)) || true
            fi
            ;;
        1|2)
            sync_log_warning "Schema version $version - consider migrating to v3"
            ((warnings++)) || true
            ;;
    esac

    # Check managed files
    local file_path file_checksum current_checksum
    while IFS= read -r file_entry; do
        [[ -z "$file_entry" ]] && continue

        file_path=$(echo "$file_entry" | jq -r '.path')
        file_checksum=$(echo "$file_entry" | jq -r '.checksum // empty')

        if [[ ! -f "$file_path" ]]; then
            sync_log_warning "Managed file missing: $file_path"
            ((warnings++)) || true
            continue
        fi

        if [[ -n "$file_checksum" ]]; then
            current_checksum=$(compute_checksum "$file_path")
            if [[ "$current_checksum" != "$file_checksum" ]]; then
                sync_log_debug "Local modification: $file_path"
            fi
        fi
    done < <(echo "$manifest" | jq -c '.managed_files[]?')

    # Return status
    if [[ $errors -gt 0 ]]; then
        return 2
    elif [[ $warnings -gt 0 ]]; then
        return 1
    fi

    return 0
}

# ============================================================================
# Team Management
# ============================================================================

# Update team info in manifest
# Usage: manifest=$(update_manifest_team "$manifest" "team_name" "checksum")
update_manifest_team() {
    local manifest="$1"
    local team_name="$2"
    local checksum="${3:-}"
    local roster_path="${ROSTER_HOME:-}"
    local timestamp

    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    echo "$manifest" | jq \
        --arg n "$team_name" \
        --arg c "$checksum" \
        --arg rp "$roster_path/rites/$team_name" \
        --arg t "$timestamp" '
        .team = {
            name: $n,
            checksum: (if $c != "" then $c else null end),
            last_refresh: $t,
            roster_path: $rp
        }'
}

# Clear team from manifest (reset to no team)
clear_manifest_team() {
    local manifest="$1"
    echo "$manifest" | jq '.team = null'
}

# ============================================================================
# Orphan Tracking
# ============================================================================

# Add orphan to manifest
# Usage: manifest=$(add_manifest_orphan "$manifest" "path" "reason")
add_manifest_orphan() {
    local manifest="$1"
    local path="$2"
    local reason="${3:-removed from roster}"
    local timestamp

    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    echo "$manifest" | jq \
        --arg p "$path" \
        --arg r "$reason" \
        --arg t "$timestamp" '
        .orphans += [{
            path: $p,
            reason: $r,
            detected_at: $t
        }]'
}

# Clear orphans from manifest
clear_manifest_orphans() {
    local manifest="$1"
    echo "$manifest" | jq '.orphans = []'
}

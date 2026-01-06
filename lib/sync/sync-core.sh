#!/usr/bin/env bash
#
# sync-core.sh - Core Sync Logic and Three-Way Classification
#
# Implements the core synchronization algorithm including three-way
# checksum classification and conflict detection.
#
# Part of: roster-sync (TDD-cem-replacement)
#
# Usage:
#   source "$ROSTER_HOME/lib/sync/sync-core.sh"
#   classification=$(classify_file "COMMAND_REGISTRY.md")
#
# Functions:
#   classify_file          - Three-way checksum classification
#   process_copy_replace   - Process copy-replace items
#   process_merge_items    - Process merge items
#   detect_orphans         - Find files no longer in roster
#   create_conflict_backup - Backup conflicting file

# Guard against re-sourcing
[[ -n "${_SYNC_CORE_LOADED:-}" ]] && return 0
readonly _SYNC_CORE_LOADED=1

# ============================================================================
# Classification Constants (per TDD 4.2)
# ============================================================================

readonly CLASSIFY_SKIP="SKIP"
readonly CLASSIFY_UPDATE="UPDATE"
readonly CLASSIFY_CONFLICT="CONFLICT"
readonly CLASSIFY_NEW="NEW"

# ============================================================================
# Three-Way Classification (per TDD 4.1, 4.2)
# ============================================================================

# Classify a file using three-way checksum comparison
#
# Decision Matrix:
# | Roster Changed?        | Local Changed?         | Action    |
# | (roster != manifest)   | (local != manifest)    |           |
# |------------------------|------------------------|-----------|
# | No                     | No                     | SKIP      |
# | No                     | Yes                    | SKIP      |
# | Yes                    | No                     | UPDATE    |
# | Yes                    | Yes                    | CONFLICT  |
#
# Usage: classify_file "filename" "roster_file" "local_file"
# Returns: SKIP, UPDATE, CONFLICT, or NEW
classify_file() {
    local filename="$1"
    local roster_file="$2"
    local local_file="$3"

    # Get checksums
    local roster_checksum manifest_checksum local_checksum

    # Roster checksum (from current roster file)
    if [[ -f "$roster_file" ]]; then
        roster_checksum=$(compute_checksum "$roster_file")
    else
        sync_log_debug "Roster file not found: $roster_file"
        echo "$CLASSIFY_SKIP"
        return 0
    fi

    # Manifest checksum (what we synced last time)
    manifest_checksum=$(get_manifest_checksum "$filename")

    # Local checksum (current satellite file)
    if [[ -f "$local_file" ]]; then
        local_checksum=$(compute_checksum "$local_file")
    else
        # Local file doesn't exist - this is a new file
        sync_log_debug "Local file not found (NEW): $local_file"
        echo "$CLASSIFY_NEW"
        return 0
    fi

    # Log for debugging
    sync_log_debug "Classifying: $filename"
    sync_log_debug "  roster_checksum:   $roster_checksum"
    sync_log_debug "  manifest_checksum: ${manifest_checksum:-<none>}"
    sync_log_debug "  local_checksum:    $local_checksum"

    # First sync (no manifest entry)
    if [[ -z "$manifest_checksum" ]]; then
        if [[ "$roster_checksum" == "$local_checksum" ]]; then
            sync_log_debug "  -> SKIP (identical, no manifest)"
            echo "$CLASSIFY_SKIP"
        else
            sync_log_debug "  -> CONFLICT (no manifest, files differ)"
            echo "$CLASSIFY_CONFLICT"
        fi
        return 0
    fi

    # Three-way comparison
    local roster_changed=0
    local local_changed=0

    if [[ "$roster_checksum" != "$manifest_checksum" ]]; then
        roster_changed=1
    fi

    if [[ "$local_checksum" != "$manifest_checksum" ]]; then
        local_changed=1
    fi

    # Apply decision matrix
    if [[ $roster_changed -eq 0 && $local_changed -eq 0 ]]; then
        sync_log_debug "  -> SKIP (up to date)"
        echo "$CLASSIFY_SKIP"
    elif [[ $roster_changed -eq 0 && $local_changed -eq 1 ]]; then
        sync_log_debug "  -> SKIP (preserve local)"
        echo "$CLASSIFY_SKIP"
    elif [[ $roster_changed -eq 1 && $local_changed -eq 0 ]]; then
        sync_log_debug "  -> UPDATE (safe to overwrite)"
        echo "$CLASSIFY_UPDATE"
    else
        sync_log_debug "  -> CONFLICT (both changed)"
        echo "$CLASSIFY_CONFLICT"
    fi

    return 0
}

# ============================================================================
# Conflict Handling (per TDD 4.3)
# ============================================================================

# Global tracking for current backup session
_CONFLICT_BACKUP_DIR=""
_CONFLICT_COUNT=0
_CONFLICT_FILES=()

# Initialize a conflict backup session
# Creates timestamped directory structure per TDD 4.3:
#   .cem-backup/
#     YYYYMMDD-HHMMSS/
#       .claude/settings.json    # Backed up conflict
#       conflict-report.txt      # What was overwritten
#
# Usage: init_conflict_backup_session [base_dir]
# Sets _CONFLICT_BACKUP_DIR to the timestamped backup directory
init_conflict_backup_session() {
    local base_dir="${1:-.cem-backup}"
    local timestamp
    timestamp=$(date +"%Y%m%d-%H%M%S")

    _CONFLICT_BACKUP_DIR="${base_dir}/${timestamp}"
    _CONFLICT_COUNT=0
    _CONFLICT_FILES=()

    sync_log_debug "Conflict backup session initialized: $_CONFLICT_BACKUP_DIR"
}

# Create backup of conflicting file with proper directory structure
# Usage: create_conflict_backup "local_file" [backup_dir]
# Returns: backup file path (on stdout only - log goes to stderr)
#
# Preserves relative path structure in backup:
#   .claude/settings.json -> .cem-backup/TIMESTAMP/.claude/settings.json
#
# Path handling:
#   - Relative paths: used as-is (.claude/foo.md -> .claude/foo.md)
#   - Absolute with .claude/: extract from .claude/ (/x/.claude/y -> .claude/y)
#   - Other absolute: use basename only (/x/y/z.md -> z.md)
create_conflict_backup() {
    local local_file="$1"
    local backup_base="${2:-${_CONFLICT_BACKUP_DIR:-}}"

    if [[ ! -f "$local_file" ]]; then
        sync_log_debug "No file to backup: $local_file"
        return 1
    fi

    # Initialize backup session if not done
    if [[ -z "$backup_base" ]]; then
        init_conflict_backup_session
        backup_base="$_CONFLICT_BACKUP_DIR"
    fi

    # Extract relative path for backup structure
    local rel_path="$local_file"

    # Remove leading ./ if present
    rel_path="${rel_path#./}"

    # Handle absolute paths
    if [[ "$rel_path" == /* ]]; then
        if [[ "$rel_path" == */.claude/* ]]; then
            # Extract from .claude/ onward
            rel_path=".claude/${rel_path##*/.claude/}"
        else
            # For other absolute paths, use just the filename
            rel_path=$(basename "$rel_path")
        fi
    fi

    local backup_file="${backup_base}/${rel_path}"
    local backup_dir
    backup_dir=$(dirname "$backup_file")

    # Create directory structure
    mkdir -p "$backup_dir" || {
        sync_log_error "Failed to create backup directory: $backup_dir"
        return 1
    }

    # Copy file preserving permissions
    cp -p "$local_file" "$backup_file" || {
        sync_log_error "Failed to create backup: $backup_file"
        return 1
    }

    # Track the conflict
    ((_CONFLICT_COUNT++))
    _CONFLICT_FILES+=("$local_file")

    sync_log "Backup created: $backup_file"
    echo "$backup_file"
}

# Resolve a conflict by applying roster version
# Usage: resolve_conflict "roster_file" "local_file" "filename" [force]
# Returns: 0 on success, 1 on failure
#
# Without force: Creates backup, reports conflict
# With force: Creates backup, overwrites with roster version
resolve_conflict() {
    local roster_file="$1"
    local local_file="$2"
    local filename="$3"
    local force="${4:-0}"

    if [[ ! -f "$roster_file" ]]; then
        sync_log_error "Roster file not found: $roster_file"
        return 1
    fi

    # Pre-flight check: skip cp if files are already identical (fixes BSD cp error)
    if [[ -f "$local_file" ]]; then
        local roster_checksum local_checksum
        roster_checksum=$(compute_checksum "$roster_file")
        local_checksum=$(compute_checksum "$local_file")

        if [[ "$roster_checksum" == "$local_checksum" ]]; then
            sync_log_debug "Files already identical, skipping copy: $local_file"
            # Update manifest to reflect current state
            set_manifest_checksum "$filename" "$roster_checksum"
            return 0
        fi
    fi

    # Always create backup of local version
    # Note: We capture output via subshell, so we must increment counters here
    # (create_conflict_backup's counter updates are lost in the subshell)
    local backup_path
    backup_path=$(create_conflict_backup "$local_file") || {
        sync_log_warning "Could not backup conflict: $local_file"
    }

    # Track the conflict in parent shell (subshell increments are lost)
    if [[ -n "$backup_path" ]]; then
        ((_CONFLICT_COUNT++))
        _CONFLICT_FILES+=("$local_file")
    fi

    if [[ "$force" == "1" ]]; then
        # Force mode: overwrite with roster version
        cp "$roster_file" "$local_file" || {
            sync_log_error "Failed to overwrite conflict: $local_file"
            return 1
        }
        sync_log "Forced overwrite: $local_file"

        # Update manifest checksum
        local new_checksum
        new_checksum=$(compute_checksum "$local_file")
        set_manifest_checksum "$filename" "$new_checksum"

        return 0
    else
        # Non-force mode: log conflict, keep local
        sync_log_warning "Conflict detected: $local_file (use --force to overwrite)"
        return 0
    fi
}

# Generate conflict report for the current backup session
# Usage: generate_conflict_report [output_file]
# Creates a human-readable report of all conflicts in the session
generate_conflict_report() {
    local output_file="${1:-${_CONFLICT_BACKUP_DIR:-}/conflict-report.txt}"

    if [[ ${#_CONFLICT_FILES[@]} -eq 0 ]]; then
        sync_log_debug "No conflicts to report"
        return 0
    fi

    local backup_dir
    backup_dir=$(dirname "$output_file")
    mkdir -p "$backup_dir" 2>/dev/null

    {
        echo "=============================================="
        echo "Conflict Report"
        echo "=============================================="
        echo ""
        echo "Generated: $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
        echo "Backup Directory: ${_CONFLICT_BACKUP_DIR:-unknown}"
        echo "Total Conflicts: $_CONFLICT_COUNT"
        echo ""
        echo "----------------------------------------------"
        echo "Conflicting Files"
        echo "----------------------------------------------"
        echo ""

        local file
        for file in "${_CONFLICT_FILES[@]}"; do
            echo "  - $file"
        done

        echo ""
        echo "----------------------------------------------"
        echo "Resolution Instructions"
        echo "----------------------------------------------"
        echo ""
        echo "These files have local modifications that conflict with"
        echo "roster updates. Options:"
        echo ""
        echo "  1. Review changes and merge manually"
        echo "     - Backed up files are in: ${_CONFLICT_BACKUP_DIR:-}"
        echo "     - Compare with current files to merge changes"
        echo ""
        echo "  2. Accept roster version (discard local changes)"
        echo "     - Run: roster-sync sync --force"
        echo ""
        echo "  3. Keep local version"
        echo "     - No action needed, local files preserved"
        echo "     - Run: roster-sync sync (will show conflicts again)"
        echo ""
        echo "=============================================="
    } > "$output_file"

    sync_log "Conflict report generated: $output_file"
    echo "$output_file"
}

# Finalize conflict handling for a sync operation
# Usage: finalize_conflicts [force]
# Returns: exit code (0 if no conflicts or force, EXIT_SYNC_CONFLICTS otherwise)
finalize_conflicts() {
    local force="${1:-0}"

    if [[ $_CONFLICT_COUNT -eq 0 ]]; then
        sync_log_debug "No conflicts to finalize"
        return 0
    fi

    # Generate report
    generate_conflict_report

    # Print summary
    sync_log_warning "Found $_CONFLICT_COUNT conflict(s)"

    if [[ "$force" == "1" ]]; then
        sync_log "Conflicts resolved with --force (local backups created)"
        return 0
    else
        echo ""
        echo "Conflicts detected. Your local changes are preserved."
        echo "Backups created in: ${_CONFLICT_BACKUP_DIR:-}"
        echo ""
        echo "To resolve:"
        echo "  - Review: cat ${_CONFLICT_BACKUP_DIR:-}/conflict-report.txt"
        echo "  - Force overwrite: roster-sync sync --force"
        echo ""
        return $EXIT_SYNC_CONFLICTS
    fi
}

# Count existing conflict backup sessions
# Returns: number of backup directories
count_conflict_backups() {
    local backup_base="${1:-.cem-backup}"
    if [[ ! -d "$backup_base" ]]; then
        echo "0"
        return
    fi
    find "$backup_base" -maxdepth 1 -type d -name "[0-9]*" 2>/dev/null | wc -l | tr -d ' '
}

# List conflict backup sessions (most recent first)
list_conflict_backups() {
    local backup_base="${1:-.cem-backup}"
    if [[ ! -d "$backup_base" ]]; then
        return
    fi
    find "$backup_base" -maxdepth 1 -type d -name "[0-9]*" 2>/dev/null | sort -r
}

# Clean old conflict backups (keep most recent N)
# Usage: clean_old_backups [keep_count] [backup_base]
clean_old_backups() {
    local keep_count="${1:-5}"
    local backup_base="${2:-.cem-backup}"

    if [[ ! -d "$backup_base" ]]; then
        return 0
    fi

    local backup_dirs
    mapfile -t backup_dirs < <(list_conflict_backups "$backup_base")

    local count=${#backup_dirs[@]}
    if [[ $count -le $keep_count ]]; then
        sync_log_debug "Only $count backups exist, keeping all"
        return 0
    fi

    local to_remove=$((count - keep_count))
    sync_log "Removing $to_remove old backup(s), keeping $keep_count"

    local i
    for ((i = keep_count; i < count; i++)); do
        local dir="${backup_dirs[$i]}"
        rm -rf "$dir" && sync_log_debug "Removed old backup: $dir"
    done
}

# Get detailed classification result with all checksums
# Usage: classify_file_detailed "filename" "roster_file" "local_file"
# Returns: JSON with classification details
classify_file_detailed() {
    local filename="$1"
    local roster_file="$2"
    local local_file="$3"

    local roster_checksum="" manifest_checksum="" local_checksum=""
    local classification="$CLASSIFY_SKIP"
    local reason=""

    # Roster checksum
    if [[ -f "$roster_file" ]]; then
        roster_checksum=$(compute_checksum "$roster_file")
    else
        classification="$CLASSIFY_SKIP"
        reason="roster_missing"
        jq -n \
            --arg c "$classification" \
            --arg r "$reason" \
            --arg f "$filename" \
            '{classification: $c, reason: $r, filename: $f, roster_checksum: null, manifest_checksum: null, local_checksum: null}'
        return 0
    fi

    # Manifest checksum
    manifest_checksum=$(get_manifest_checksum "$filename")

    # Local checksum
    if [[ -f "$local_file" ]]; then
        local_checksum=$(compute_checksum "$local_file")
    else
        classification="$CLASSIFY_NEW"
        reason="local_missing"
        jq -n \
            --arg c "$classification" \
            --arg r "$reason" \
            --arg f "$filename" \
            --arg rc "$roster_checksum" \
            '{classification: $c, reason: $r, filename: $f, roster_checksum: $rc, manifest_checksum: null, local_checksum: null}'
        return 0
    fi

    # First sync (no manifest entry)
    if [[ -z "$manifest_checksum" ]]; then
        if [[ "$roster_checksum" == "$local_checksum" ]]; then
            classification="$CLASSIFY_SKIP"
            reason="identical_no_manifest"
        else
            classification="$CLASSIFY_CONFLICT"
            reason="no_manifest_files_differ"
        fi
    else
        # Three-way comparison
        local roster_changed=0
        local local_changed=0

        [[ "$roster_checksum" != "$manifest_checksum" ]] && roster_changed=1
        [[ "$local_checksum" != "$manifest_checksum" ]] && local_changed=1

        if [[ $roster_changed -eq 0 && $local_changed -eq 0 ]]; then
            classification="$CLASSIFY_SKIP"
            reason="up_to_date"
        elif [[ $roster_changed -eq 0 && $local_changed -eq 1 ]]; then
            classification="$CLASSIFY_SKIP"
            reason="preserve_local"
        elif [[ $roster_changed -eq 1 && $local_changed -eq 0 ]]; then
            classification="$CLASSIFY_UPDATE"
            reason="roster_updated"
        else
            classification="$CLASSIFY_CONFLICT"
            reason="both_modified"
        fi
    fi

    jq -n \
        --arg c "$classification" \
        --arg r "$reason" \
        --arg f "$filename" \
        --arg rc "$roster_checksum" \
        --arg mc "$manifest_checksum" \
        --arg lc "$local_checksum" \
        '{classification: $c, reason: $r, filename: $f, roster_checksum: $rc, manifest_checksum: (if $mc == "" then null else $mc end), local_checksum: $lc}'
}

# ============================================================================
# Copy-Replace Processing
# ============================================================================

# Process all copy-replace items
# Usage: process_copy_replace "roster_dir" "local_dir" [force]
# Returns: number of conflicts
#
# Per TDD 4.3:
# - CONFLICT without --force: Report conflicts, exit code 5
# - CONFLICT with --force: Backup local to .cem-backup/, overwrite with roster
process_copy_replace() {
    local roster_dir="$1"
    local local_dir="$2"
    local force="${3:-0}"

    local conflicts=0
    local item roster_file local_file classification

    while IFS= read -r item; do
        [[ -z "$item" ]] && continue

        roster_file="$roster_dir/$item"
        local_file="$local_dir/$item"

        classification=$(classify_file "$item" "$roster_file" "$local_file")

        case "$classification" in
            "$CLASSIFY_SKIP")
                sync_log_debug "Skipping: $item"
                ;;
            "$CLASSIFY_NEW"|"$CLASSIFY_UPDATE")
                sync_log "Updating: $item"
                mkdir -p "$(dirname "$local_file")"
                cp "$roster_file" "$local_file" || {
                    sync_log_error "Failed to copy: $item"
                    return 1
                }
                # Update manifest checksum
                local new_checksum
                new_checksum=$(compute_checksum "$local_file")
                set_manifest_checksum "$item" "$new_checksum"
                ;;
            "$CLASSIFY_CONFLICT")
                sync_log_warning "Conflict: $item (local modified, roster updated)"
                ((conflicts++)) || true

                # Use resolve_conflict which handles backup and optional force
                resolve_conflict "$roster_file" "$local_file" "$item" "$force"
                ;;
        esac
    done < <(get_copy_replace_items)

    echo "$conflicts"
}

# ============================================================================
# Merge Processing
# ============================================================================

# Process all merge items
# Usage: process_merge_items "roster_dir" "local_dir" [force]
# Returns: number of errors
process_merge_items() {
    local roster_dir="$1"
    local local_dir="$2"
    local force="${3:-0}"

    local errors=0
    local line file strategy roster_file local_file roster_checksum manifest_checksum

    while IFS= read -r line; do
        [[ -z "$line" ]] && continue

        file="${line%%:*}"
        strategy="${line#*:}"

        roster_file="$roster_dir/$file"
        local_file="$local_dir/$file"

        # Check if roster file changed
        roster_checksum=$(compute_checksum "$roster_file")
        manifest_checksum=$(get_manifest_checksum "$file")

        if [[ -n "$manifest_checksum" && "$roster_checksum" == "$manifest_checksum" ]]; then
            sync_log_debug "Merge skip (unchanged): $file"
            continue
        fi

        sync_log "Merging: $file (strategy: $strategy)"

        # Dispatch to appropriate merge strategy
        if ! dispatch_merge_strategy "$strategy" "$roster_file" "$local_file" "$local_file"; then
            sync_log_error "Merge failed: $file"
            ((errors++)) || true
            continue
        fi

        # Update manifest
        local new_checksum
        new_checksum=$(compute_checksum "$local_file")
        set_manifest_checksum "$file" "$new_checksum"
    done < <(get_merge_items)

    echo "$errors"
}

# ============================================================================
# Orphan Detection (per TDD 4.4 step 6)
# ============================================================================

# Detect orphaned files - files in manifest that are no longer in roster source
#
# Per TDD: "File in manifest but not in roster source = orphan candidate"
#
# An orphan is a file that:
# 1. Exists in the manifest (was previously synced)
# 2. No longer exists in the roster source
# 3. The file was from a managed source (copy-replace or merge strategy)
#
# Usage: detect_orphans "roster_dir"
# Outputs: list of orphaned paths (one per line)
detect_orphans() {
    local roster_dir="$1"
    local manifest

    manifest=$(read_manifest) || return 1

    local file_path roster_file strategy
    while IFS= read -r file_entry; do
        [[ -z "$file_entry" ]] && continue

        file_path=$(echo "$file_entry" | jq -r '.path')
        strategy=$(echo "$file_entry" | jq -r '.strategy // "copy-replace"')

        # Convert .claude/filename to just filename for roster lookup
        local filename="${file_path#.claude/}"

        # Skip ignored items (they're not synced)
        if is_ignored "$filename"; then
            continue
        fi

        # Check if file exists in roster source
        roster_file="$roster_dir/$filename"

        if [[ ! -f "$roster_file" ]]; then
            # File was in manifest with a sync strategy but no longer in roster
            # This is an orphan (the file was previously synced from roster)
            if [[ "$strategy" == "copy-replace" || "$strategy" == "merge-settings" || "$strategy" == "merge-docs" ]]; then
                echo "$file_path"
            fi
        fi
    done < <(echo "$manifest" | jq -c '.managed_files[]?')
}

# Detect untracked files - files in .claude/ that are not in manifest
#
# Per TDD: "File in .claude/ but not in manifest = untracked (different)"
#
# Usage: detect_untracked "local_claude_dir"
# Outputs: list of untracked file paths (one per line)
detect_untracked() {
    local local_dir="$1"
    local manifest

    manifest=$(read_manifest) || {
        # No manifest means all files are untracked
        get_local_claude_files "$local_dir"
        return 0
    }

    local file rel_path
    while IFS= read -r file; do
        [[ -z "$file" ]] && continue

        # Get path relative to project root
        rel_path="$file"

        # Check if in ignored list
        local filename="${rel_path#.claude/}"
        if is_ignored "$filename"; then
            continue
        fi

        # Check if it's a subdirectory that should be ignored
        local base_dir="${filename%%/*}"
        if is_ignored "$base_dir"; then
            continue
        fi

        # Check if file is in manifest
        local in_manifest
        in_manifest=$(echo "$manifest" | jq -r --arg p "$rel_path" \
            '.managed_files[] | select(.path == $p) | .path')

        if [[ -z "$in_manifest" ]]; then
            echo "$rel_path"
        fi
    done < <(get_local_claude_files "$local_dir")
}

# Get all files in local .claude/ directory
#
# Usage: get_local_claude_files "local_claude_dir"
# Outputs: list of file paths relative to project root
get_local_claude_files() {
    local local_dir="$1"

    if [[ ! -d "$local_dir" ]]; then
        return 0
    fi

    # Find all files, excluding hidden directories like .cem
    find "$local_dir" -type f \
        ! -path "*/.cem/*" \
        ! -path "*/.archive/*" \
        ! -path "*/sessions/*" \
        ! -path "*/agents/*" \
        ! -path "*/agents.backup/*" \
        ! -path "*/user-*/*" \
        ! -path "*/commands/*" \
        ! -path "*/skills/*" \
        ! -path "*/hooks/*" \
        2>/dev/null | sort
}

# Backup orphaned files before removal
#
# Per TDD: "Create .cem-backup/YYYYMMDD-HHMMSS/ directory"
# Per TDD: "Copy orphan files preserving relative paths"
# Per TDD: "Log backup location"
#
# Usage: backup_orphans < orphan_list
# Outputs: backup directory path (or empty if no orphans)
backup_orphans() {
    local backup_base="${SYNC_ORPHAN_BACKUP_DIR:-.claude/.cem/orphan-backup}"
    local timestamp
    timestamp=$(date +"%Y%m%d-%H%M%S")
    local backup_dir="${backup_base}/${timestamp}"

    local has_orphans=0
    local orphan rel_path

    while IFS= read -r orphan; do
        [[ -z "$orphan" ]] && continue
        [[ ! -f "$orphan" ]] && continue

        # Create backup directory on first orphan
        if [[ $has_orphans -eq 0 ]]; then
            mkdir -p "$backup_dir" || {
                sync_log_error "Failed to create backup directory: $backup_dir"
                return 1
            }
            sync_log "Creating orphan backup: $backup_dir"
            has_orphans=1
        fi

        # Preserve relative path structure
        # .claude/foo/bar.md -> backup_dir/foo/bar.md
        rel_path="${orphan#.claude/}"
        local backup_file="$backup_dir/$rel_path"
        local backup_file_dir
        backup_file_dir=$(dirname "$backup_file")

        mkdir -p "$backup_file_dir" || {
            sync_log_warning "Failed to create backup subdirectory: $backup_file_dir"
            continue
        }

        cp "$orphan" "$backup_file" || {
            sync_log_warning "Failed to backup orphan: $orphan"
            continue
        }
        sync_log_debug "Backed up: $orphan -> $backup_file"
    done

    if [[ $has_orphans -eq 1 ]]; then
        sync_log "Backup location: $backup_dir"
        echo "$backup_dir"
    fi
}

# Prune orphaned files (remove after backup)
#
# Per TDD: "With --prune: Backup to .cem-backup/ then remove"
#
# Usage: prune_orphans < orphan_list
# Returns: 0 on success, 1 on error
prune_orphans() {
    local manifest
    manifest=$(read_manifest) || return 1

    local orphan pruned_count=0
    while IFS= read -r orphan; do
        [[ -z "$orphan" ]] && continue

        if [[ -f "$orphan" ]]; then
            rm "$orphan" || {
                sync_log_warning "Failed to prune orphan: $orphan"
                continue
            }
            sync_log "Pruned: $orphan"
            pruned_count=$((pruned_count + 1))

            # Remove from manifest
            manifest=$(echo "$manifest" | jq --arg p "$orphan" '
                .managed_files = [.managed_files[] | select(.path != $p)]')

            # Track orphan removal in manifest
            manifest=$(add_manifest_orphan "$manifest" "$orphan" "pruned - no longer in roster")
        fi
    done

    write_manifest "$manifest"

    if [[ $pruned_count -gt 0 ]]; then
        sync_log "Pruned $pruned_count orphan(s)"
    fi
}

# Check if orphans exist
#
# Usage: has_orphans "roster_dir"
# Returns: 0 if orphans exist, 1 if no orphans
has_orphans() {
    local roster_dir="$1"
    local orphan_count

    orphan_count=$(detect_orphans "$roster_dir" | wc -l | tr -d ' ')

    [[ "$orphan_count" -gt 0 ]]
}

# Report orphans without removing
#
# Per TDD: "Without --prune: Report orphans but don't remove"
#
# Usage: report_orphans "roster_dir"
# Returns: EXIT_SYNC_ORPHAN_CONFLICTS if orphans exist, EXIT_SYNC_SUCCESS otherwise
report_orphans() {
    local roster_dir="$1"
    local orphans

    orphans=$(detect_orphans "$roster_dir")

    if [[ -z "$orphans" ]]; then
        sync_log_debug "No orphans detected"
        return "$EXIT_SYNC_SUCCESS"
    fi

    local count
    count=$(echo "$orphans" | wc -l | tr -d ' ')

    sync_log_warning "Found $count orphan(s) - files no longer in roster:"
    while IFS= read -r orphan; do
        [[ -z "$orphan" ]] && continue
        sync_log_warning "  - $orphan"
    done <<< "$orphans"

    sync_log_warning "Use --prune to backup and remove orphans"

    return "$EXIT_SYNC_ORPHAN_CONFLICTS"
}

# Handle orphans based on flags
#
# Usage: handle_orphans "roster_dir" "prune_flag" "force_flag"
# Returns: appropriate exit code
handle_orphans() {
    local roster_dir="$1"
    local prune="${2:-0}"
    local force="${3:-0}"

    local orphans
    orphans=$(detect_orphans "$roster_dir")

    if [[ -z "$orphans" ]]; then
        sync_log_debug "No orphans detected"
        return "$EXIT_SYNC_SUCCESS"
    fi

    if [[ "$prune" != "1" ]]; then
        # Report only, don't remove
        report_orphans "$roster_dir"
        return "$EXIT_SYNC_ORPHAN_CONFLICTS"
    fi

    # Prune mode: backup then remove
    local backup_dir
    backup_dir=$(echo "$orphans" | backup_orphans)

    if [[ -z "$backup_dir" && -n "$orphans" ]]; then
        sync_log_error "Backup failed, aborting prune"
        return "$EXIT_SYNC_ERROR"
    fi

    # Now prune
    echo "$orphans" | prune_orphans

    return "$EXIT_SYNC_SUCCESS"
}

# ============================================================================
# Version Checking
# ============================================================================

# Get current roster git commit
get_roster_commit() {
    local roster_path="${ROSTER_HOME:-$HOME/Code/roster}"

    if [[ -d "$roster_path/.git" ]]; then
        git -C "$roster_path" rev-parse HEAD 2>/dev/null
    else
        echo ""
    fi
}

# Get current roster git ref (branch)
get_roster_ref() {
    local roster_path="${ROSTER_HOME:-$HOME/Code/roster}"

    if [[ -d "$roster_path/.git" ]]; then
        git -C "$roster_path" rev-parse --abbrev-ref HEAD 2>/dev/null
    else
        echo "main"
    fi
}

# Check if roster has updates compared to manifest
# Returns: 0 if updates available, 1 if up to date
roster_has_updates() {
    local current_commit manifest_commit

    current_commit=$(get_roster_commit)
    manifest_commit=$(get_manifest_field ".roster.commit")

    if [[ -z "$current_commit" ]]; then
        sync_log_debug "Cannot determine roster commit"
        return 0  # Assume updates available
    fi

    if [[ "$current_commit" != "$manifest_commit" ]]; then
        return 0  # Updates available
    fi

    return 1  # Up to date
}

# ============================================================================
# Team Freshness
# ============================================================================

# Check if active team needs refresh
# Returns: 0 if stale, 1 if fresh
is_team_stale() {
    local active_team_file=".claude/ACTIVE_RITE"

    if [[ ! -f "$active_team_file" ]]; then
        return 1  # No team, not stale
    fi

    local team_name
    team_name=$(cat "$active_team_file")

    local team_dir="${ROSTER_HOME:-}/teams/$team_name"
    if [[ ! -d "$team_dir" ]]; then
        sync_log_debug "Team directory not found: $team_dir"
        return 1
    fi

    # Compare team directory mtime with last refresh
    local team_mtime manifest_refresh
    team_mtime=$(stat -f "%m" "$team_dir" 2>/dev/null || stat -c "%Y" "$team_dir" 2>/dev/null)
    manifest_refresh=$(get_manifest_field ".team.last_refresh")

    if [[ -z "$manifest_refresh" ]]; then
        return 0  # No refresh recorded, assume stale
    fi

    # Convert manifest timestamp to epoch for comparison
    local refresh_epoch
    if command -v gdate &>/dev/null; then
        refresh_epoch=$(gdate -d "$manifest_refresh" +%s 2>/dev/null)
    else
        refresh_epoch=$(date -d "$manifest_refresh" +%s 2>/dev/null || date -jf "%Y-%m-%dT%H:%M:%SZ" "$manifest_refresh" +%s 2>/dev/null)
    fi

    if [[ -z "$refresh_epoch" ]]; then
        return 0  # Cannot parse, assume stale
    fi

    if [[ "$team_mtime" -gt "$refresh_epoch" ]]; then
        return 0  # Team updated after last refresh
    fi

    return 1  # Fresh
}

# Refresh active rite via swap-rite.sh
refresh_active_team() {
    local active_team_file=".claude/ACTIVE_RITE"

    if [[ ! -f "$active_team_file" ]]; then
        sync_log_debug "No active rite to refresh"
        return 0
    fi

    local team_name
    team_name=$(cat "$active_team_file")

    sync_log "Refreshing rite: $team_name"

    local swap_rite="${ROSTER_HOME:-}/swap-rite.sh"
    if [[ -x "$swap_rite" ]]; then
        "$swap_rite" "$team_name" --update
    else
        sync_log_warning "swap-rite.sh not found: $swap_rite"
        return 1
    fi
}

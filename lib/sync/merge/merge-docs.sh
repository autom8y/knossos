#!/usr/bin/env bash
#
# merge-docs.sh - Documentation Section-Based Merge
#
# Implements section-based merge for CLAUDE.md using ownership markers.
# Supports SYNC (roster-owned) and PRESERVE (satellite-owned) markers.
#
# Part of: roster-sync (TDD-cem-replacement)
#
# Usage:
#   source "$ROSTER_HOME/lib/sync/merge/merge-docs.sh"
#   merge_documentation "$roster_file" "$local_file" "$output_file"
#
# Markers:
#   <!-- SYNC: roster-owned -->       - Always take roster version
#   <!-- SYNC: skeleton-owned -->     - Legacy, same as roster-owned
#   <!-- PRESERVE: satellite-owned --> - Keep satellite version
#
# Special Sections:
#   ## Quick Start          - Preserve if local has content
#   ## Agent Configurations - Preserve if local has content

# Guard against re-sourcing
[[ -n "${_MERGE_DOCS_LOADED:-}" ]] && return 0
readonly _MERGE_DOCS_LOADED=1

# ============================================================================
# Constants
# ============================================================================

readonly SYNC_MARKER_ROSTER="<!-- SYNC: roster-owned -->"
readonly SYNC_MARKER_SKELETON="<!-- SYNC: skeleton-owned -->"
readonly PRESERVE_MARKER="<!-- PRESERVE: satellite-owned -->"

# Sections that should be preserved by default if they exist locally
readonly PRESERVE_SECTIONS="Quick Start|Agent Configurations"

# ============================================================================
# Documentation Merge (per TDD 5.4)
# ============================================================================

# Merge CLAUDE.md with section-based ownership
# Algorithm:
#   1. If no local file: copy roster as-is
#   2. Extract header/preamble from roster (before first ##)
#   3. For each roster section:
#      - If roster has SYNC marker -> use roster section
#      - If local has PRESERVE marker -> use local section
#      - Fallback sections (Quick Start, Agent Configurations) -> preserve local if exists
#      - Otherwise -> sync from roster
#   4. Append satellite-only sections (not in roster)
#   5. Append ## Project:* sections from local
#
# Usage: merge_documentation "roster_file" "local_file" "output_file"
merge_documentation() {
    local roster_file="$1"
    local local_file="$2"
    local output_file="$3"

    # If no local file, copy roster as-is
    if [[ ! -f "$local_file" ]]; then
        sync_log_debug "merge-docs: no local file, copying roster"
        cp "$roster_file" "$output_file"
        return 0
    fi

    # Create temp files for processing
    local temp_output
    temp_output=$(mktemp)
    trap "rm -f '$temp_output'" RETURN

    # 1. Extract and write header/preamble from roster
    extract_header "$roster_file" > "$temp_output"

    # 2. Get section lists
    local roster_sections local_sections
    roster_sections=$(list_sections "$roster_file")
    local_sections=$(list_sections "$local_file")

    sync_log_debug "merge-docs: roster sections: $roster_sections"
    sync_log_debug "merge-docs: local sections: $local_sections"

    # 3. Process each roster section
    local section
    while IFS= read -r section; do
        [[ -z "$section" ]] && continue

        local decision roster_marker local_marker

        # Check markers
        roster_marker=$(get_section_marker "$roster_file" "$section")
        local_marker=$(get_section_marker "$local_file" "$section")

        sync_log_debug "merge-docs: section='$section' roster_marker='$roster_marker' local_marker='$local_marker'"

        # Decision logic
        if [[ "$roster_marker" == "SYNC" ]]; then
            decision="roster"
            sync_log_debug "merge-docs: decision=roster (roster has SYNC)"
        elif [[ "$local_marker" == "PRESERVE" ]]; then
            decision="local"
            sync_log_debug "merge-docs: decision=local (local has PRESERVE)"
        elif echo "$section" | grep -qE "^($PRESERVE_SECTIONS)$" && section_exists "$local_file" "$section"; then
            decision="local"
            sync_log_debug "merge-docs: decision=local (preserve section with local content)"
        else
            decision="roster"
            sync_log_debug "merge-docs: decision=roster (default)"
        fi

        # Write the section
        if [[ "$decision" == "local" ]]; then
            extract_section "$local_file" "$section" >> "$temp_output"
        else
            extract_section "$roster_file" "$section" >> "$temp_output"
        fi

        echo "" >> "$temp_output"
    done <<< "$roster_sections"

    # 4. Append satellite-only sections (not in roster)
    while IFS= read -r section; do
        [[ -z "$section" ]] && continue

        # Skip if section is in roster
        if echo "$roster_sections" | grep -q "^${section}$"; then
            continue
        fi

        # Skip Project:* sections (handled separately)
        if [[ "$section" == Project:* ]]; then
            continue
        fi

        sync_log_debug "merge-docs: appending satellite section: $section"
        extract_section "$local_file" "$section" >> "$temp_output"
        echo "" >> "$temp_output"
    done <<< "$local_sections"

    # 5. Append Project:* sections from local
    while IFS= read -r section; do
        [[ -z "$section" ]] && continue

        if [[ "$section" == Project:* ]]; then
            sync_log_debug "merge-docs: appending project section: $section"
            extract_section "$local_file" "$section" >> "$temp_output"
            echo "" >> "$temp_output"
        fi
    done <<< "$local_sections"

    # Write output
    mv "$temp_output" "$output_file"
    sync_log_debug "merge-docs: complete"
    return 0
}

# ============================================================================
# Section Extraction
# ============================================================================

# Extract header (content before first ## section)
extract_header() {
    local file="$1"

    # Get line number of first ## heading
    local first_section_line
    first_section_line=$(grep -n "^## " "$file" 2>/dev/null | head -1 | cut -d: -f1)

    if [[ -z "$first_section_line" ]]; then
        # No sections, return entire file
        cat "$file"
    elif [[ "$first_section_line" -gt 1 ]]; then
        head -n "$((first_section_line - 1))" "$file"
    fi
}

# List all ## sections in a file
# Returns: newline-separated list of section names (without ##)
list_sections() {
    local file="$1"

    grep "^## " "$file" 2>/dev/null | sed 's/^## //' | while IFS= read -r line; do
        # Trim any trailing whitespace
        echo "$line" | sed 's/[[:space:]]*$//'
    done
}

# Extract a specific section (heading + content until next ## or EOF)
extract_section() {
    local file="$1"
    local section="$2"

    # Escape special regex characters in section name
    local escaped_section
    escaped_section=$(printf '%s' "$section" | sed 's/[[\.*^$()+?{|]/\\&/g')

    # Find section start and end
    local start_line end_line
    start_line=$(grep -n "^## ${escaped_section}$" "$file" 2>/dev/null | head -1 | cut -d: -f1)

    if [[ -z "$start_line" ]]; then
        # Try with whitespace variance
        start_line=$(grep -n "^## ${escaped_section}" "$file" 2>/dev/null | head -1 | cut -d: -f1)
    fi

    if [[ -z "$start_line" ]]; then
        sync_log_debug "Section not found: $section"
        return 1
    fi

    # Find next section (or EOF)
    local total_lines
    total_lines=$(wc -l < "$file" | tr -d ' ')

    end_line=$(tail -n "+$((start_line + 1))" "$file" | grep -n "^## " | head -1 | cut -d: -f1)

    if [[ -n "$end_line" ]]; then
        end_line=$((start_line + end_line - 1))
    else
        end_line=$total_lines
    fi

    # Extract section content
    sed -n "${start_line},${end_line}p" "$file"
}

# Check if a section exists in a file
section_exists() {
    local file="$1"
    local section="$2"

    grep -q "^## ${section}" "$file" 2>/dev/null
}

# ============================================================================
# Marker Detection
# ============================================================================

# Get the marker type for a section
# Returns: "SYNC", "PRESERVE", or empty
get_section_marker() {
    local file="$1"
    local section="$2"

    # Extract section content
    local content
    content=$(extract_section "$file" "$section" 2>/dev/null)

    if [[ -z "$content" ]]; then
        echo ""
        return 1
    fi

    # Check for markers in first 5 lines of section
    local header
    header=$(echo "$content" | head -5)

    if echo "$header" | grep -q "SYNC:.*owned"; then
        echo "SYNC"
    elif echo "$header" | grep -q "PRESERVE:.*owned"; then
        echo "PRESERVE"
    else
        echo ""
    fi
}

# ============================================================================
# Section Regeneration (per TDD 5.4)
# ============================================================================

# Regenerate ## Quick Start from agents directory
# Usage: regenerate_quick_start "agents_dir" "team_name"
regenerate_quick_start() {
    local agents_dir="$1"
    local team_name="$2"

    if [[ ! -d "$agents_dir" ]]; then
        sync_log_debug "Cannot regenerate Quick Start: agents dir not found"
        return 1
    fi

    local agent_count
    agent_count=$(find "$agents_dir" -maxdepth 1 -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')

    cat <<EOF
## Quick Start

This project uses a ${agent_count}-agent workflow (${team_name}):

| Agent | Role | Produces |
| ----- | ---- | -------- |
EOF

    for agent_file in "$agents_dir"/*.md; do
        [[ -f "$agent_file" ]] || continue

        local name role produces
        name=$(basename "$agent_file" .md)
        role=$(extract_agent_role "$agent_file")
        produces=$(extract_agent_produces "$agent_file")

        echo "| **$name** | $role | $produces |"
    done
}

# Regenerate ## Agent Configurations from agents directory
regenerate_agent_configurations() {
    local agents_dir="$1"

    if [[ ! -d "$agents_dir" ]]; then
        sync_log_debug "Cannot regenerate Agent Configurations: agents dir not found"
        return 1
    fi

    cat <<EOF
## Agent Configurations

Full agent prompts live in \`.claude/agents/\`:

EOF

    for agent_file in "$agents_dir"/*.md; do
        [[ -f "$agent_file" ]] || continue

        local name desc
        name=$(basename "$agent_file" .md)
        desc=$(head -20 "$agent_file" | grep -A1 "^#" | tail -1 | cut -c1-80)

        echo "- \`$name.md\` - $desc"
    done
}

# Extract agent role from frontmatter or description
extract_agent_role() {
    local file="$1"

    # Try frontmatter role field
    local role
    role=$(grep "^role:" "$file" 2>/dev/null | head -1 | sed 's/^role:[[:space:]]*//' | sed 's/^["'"'"']//' | sed 's/["'"'"']$//')

    if [[ -z "$role" ]]; then
        # Try description
        role=$(grep "^description:" "$file" 2>/dev/null | head -1 | sed 's/^description:[[:space:]]*//' | cut -c1-50)
    fi

    if [[ -z "$role" ]]; then
        role="Agent"
    fi

    echo "$role"
}

# Extract agent produces from workflow or fallback
extract_agent_produces() {
    local file="$1"

    local produces
    produces=$(grep "^produces:" "$file" 2>/dev/null | head -1 | sed 's/^produces:[[:space:]]*//')

    if [[ -z "$produces" ]]; then
        produces="Artifacts"
    fi

    echo "$produces"
}

# ============================================================================
# Validation
# ============================================================================

# Validate CLAUDE.md structure
validate_claude_md() {
    local file="$1"

    if [[ ! -f "$file" ]]; then
        return 1
    fi

    # Check for expected sections
    local missing=0

    if ! grep -q "^# CLAUDE.md" "$file" 2>/dev/null; then
        sync_log_warning "CLAUDE.md missing title header"
        ((missing++))
    fi

    # Check markers are properly closed
    local open_markers
    open_markers=$(grep -c "<!-- " "$file" 2>/dev/null || echo 0)
    local close_markers
    close_markers=$(grep -c " -->" "$file" 2>/dev/null || echo 0)

    if [[ "$open_markers" -ne "$close_markers" ]]; then
        sync_log_warning "CLAUDE.md has unbalanced comment markers"
    fi

    return 0
}

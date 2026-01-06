#!/usr/bin/env bash
#
# roster-utils.sh - Dynamic Team Roster Generation Utilities
#
# Library functions for generating rite rosters from agent frontmatter
# and workflow.yaml files. Designed to be sourced by swap-rite.sh.
#
# Usage:
#   source "$ROSTER_HOME/lib/roster-utils.sh"
#   generate_roster "10x-dev-pack"
#
# Functions:
#   get_frontmatter   - Extract field from markdown frontmatter
#   truncate          - Truncate text at word boundary
#   get_produces      - Get agent produces from workflow.yaml
#   generate_roster   - Generate formatted roster table

# Guard against re-sourcing
[[ -n "${_ROSTER_UTILS_LOADED:-}" ]] && return 0
readonly _ROSTER_UTILS_LOADED=1

# ============================================================================
# Core Functions
# ============================================================================

# Extract value from YAML frontmatter
# Usage: get_frontmatter "field" "file.md" "default"
# Returns: field value or default
get_frontmatter() {
    local field="$1"
    local file="$2"
    local default="${3:-}"

    # Find line numbers of --- delimiters (macOS-compatible)
    local start_line end_line
    start_line=$(grep -n '^---$' "$file" 2>/dev/null | head -1 | cut -d: -f1)
    end_line=$(grep -n '^---$' "$file" 2>/dev/null | sed -n '2p' | cut -d: -f1)

    if [[ -z "$start_line" || -z "$end_line" ]]; then
        echo "$default"
        return
    fi

    # Extract lines between delimiters
    local frontmatter
    frontmatter=$(sed -n "$((start_line + 1)),$((end_line - 1))p" "$file" 2>/dev/null)

    if [[ -z "$frontmatter" ]]; then
        echo "$default"
        return
    fi

    # Extract field value, stripping quotes
    local value
    value=$(echo "$frontmatter" | grep "^${field}:" | head -1 | sed "s/^${field}:[[:space:]]*//" | sed 's/^["'"'"']//' | sed 's/["'"'"']$//')

    if [[ -z "$value" ]]; then
        echo "$default"
    else
        echo "$value"
    fi
}

# Truncate text at word boundary
# Usage: truncate "text" max_length
# Returns: text truncated at word boundary (with ... if truncated)
truncate() {
    local text="$1"
    local max="${2:-50}"

    if [[ ${#text} -le $max ]]; then
        echo "$text"
    else
        # Truncate and find last word boundary
        local truncated="${text:0:$max}"
        truncated=$(echo "$truncated" | sed 's/[[:space:]][^[:space:]]*$//')
        echo "$truncated"
    fi
}

# Get agent produces from workflow.yaml mapping
# Usage: get_produces "agent-name" "workflow.yaml"
# Returns: produces value or fallback
get_produces() {
    local agent="$1"
    local workflow_file="$2"

    if [[ ! -f "$workflow_file" ]]; then
        echo "Artifacts"
        return
    fi

    # Extract from phases section
    local produces
    produces=$(grep -A5 "agent:[[:space:]]*$agent" "$workflow_file" 2>/dev/null | grep "produces:" | head -1 | sed 's/.*produces:[[:space:]]*//' | sed 's/^["'"'"']//' | sed 's/["'"'"']$//')

    if [[ -z "$produces" ]]; then
        # Fallback based on common agent patterns
        case "$agent" in
            orchestrator) produces="Work breakdown" ;;
            requirements-analyst) produces="PRD" ;;
            architect) produces="TDD, ADRs" ;;
            principal-engineer) produces="Code" ;;
            qa-adversary) produces="Test reports" ;;
            code-smeller) produces="Smell report" ;;
            architect-enforcer) produces="Refactor plan" ;;
            janitor) produces="Commits" ;;
            audit-lead) produces="Audit report" ;;
            *) produces="Artifacts" ;;
        esac
    fi

    echo "$produces"
}

# ============================================================================
# Roster Generation
# ============================================================================

# Generate roster table for a team
# Usage: generate_roster "team-name"
# Output: Formatted markdown table to stdout
generate_roster() {
    local team_name="$1"
    local agents_dir="$ROSTER_HOME/rites/$team_name/agents"
    local workflow_file="$ROSTER_HOME/rites/$team_name/workflow.yaml"

    if [[ ! -d "$agents_dir" ]]; then
        echo "Error: Agents directory not found: $agents_dir" >&2
        return 1
    fi

    # Count agents (macOS-compatible)
    local agent_count
    agent_count=$(find "$agents_dir" -maxdepth 1 -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')

    if [[ "$agent_count" -eq 0 ]]; then
        echo "Error: No agents found in $agents_dir" >&2
        return 1
    fi

    # Output header
    echo ""
    echo "**${team_name}** ($agent_count agents):"
    echo ""
    echo "| Agent | Role | Produces |"
    echo "|-------|------|----------|"

    # Process each agent file
    for agent_file in "$agents_dir"/*.md; do
        [[ -f "$agent_file" ]] || continue

        local basename
        basename=$(basename "$agent_file" .md)

        # Extract from frontmatter
        local name role
        name=$(get_frontmatter "name" "$agent_file" "$basename")
        role=$(get_frontmatter "role" "$agent_file")

        # Fallback: extract first sentence from description
        if [[ -z "$role" ]]; then
            local desc
            desc=$(get_frontmatter "description" "$agent_file")
            if [[ -n "$desc" ]]; then
                # First sentence or first 60 chars
                role=$(echo "$desc" | sed 's/\([^.]*\.\).*/\1/' | head -c 60)
            fi
        fi

        # Truncate role to fit table
        role=$(truncate "$role" 45)

        # Get produces from workflow or fallback
        local produces
        produces=$(get_produces "$basename" "$workflow_file")

        echo "| **$name** | $role | $produces |"
    done
}

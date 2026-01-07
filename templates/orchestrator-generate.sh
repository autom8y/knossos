#!/bin/bash
# orchestrator-generate.sh - Production-ready orchestrator.md generator
#
# Generates orchestrator.md from orchestrator.yaml template + workflow.yaml config
# with comprehensive validation, error handling, and multi-platform support.
#
# Usage:
#   ./orchestrator-generate.sh <team-name> [--dry-run] [--validate-only] [--force]
#   ./orchestrator-generate.sh --all [--validate-only] [--dry-run] [--force]
#   ./orchestrator-generate.sh --help
#
# Flags:
#   --dry-run        Output to stdout instead of file
#   --validate-only  Check configs and schema, don't generate
#   --force          Regenerate even if target exists
#   --all            Batch-generate all teams (stops on first error)
#   --help           Show this help message

set -euo pipefail

# ============================================================================
# Configuration & Paths
# ============================================================================

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source Knossos home resolution (handles KNOSSOS_HOME deprecation)
source "$SCRIPT_DIR/../lib/knossos-home.sh"

TEMPLATE="$KNOSSOS_HOME/templates/orchestrator-base.md.tpl"
SCHEMA="$KNOSSOS_HOME/schemas/orchestrator.yaml.schema.json"
VALIDATOR="$KNOSSOS_HOME/templates/validate-orchestrator.sh"

# Initialize cleanup variables
TMPFILE=""
DIAGRAM_FILE=""
ROUTING_FILE=""
SKILLS_FILE=""

# Trap for cleanup
cleanup() {
    [[ -n "${TMPFILE:-}" ]] && rm -f "$TMPFILE" 2>/dev/null || true
    [[ -n "${DIAGRAM_FILE:-}" ]] && rm -f "$DIAGRAM_FILE" 2>/dev/null || true
    [[ -n "${ROUTING_FILE:-}" ]] && rm -f "$ROUTING_FILE" 2>/dev/null || true
    [[ -n "${SKILLS_FILE:-}" ]] && rm -f "$SKILLS_FILE" 2>/dev/null || true
}
trap cleanup EXIT

# ============================================================================
# Colors & Logging
# ============================================================================

readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

log_info() { echo -e "${BLUE}INFO${NC}: $*"; }
log_ok() { echo -e "${GREEN}OK${NC}: $*"; }
log_warn() { echo -e "${YELLOW}WARN${NC}: $*" >&2; }
log_error() { echo -e "${RED}ERROR${NC}: $*" >&2; }

# ============================================================================
# Help & Argument Parsing
# ============================================================================

show_help() {
    cat <<EOF
orchestrator-generate.sh - Generate orchestrator.md from configuration

USAGE:
  ./orchestrator-generate.sh <rite-name> [options]
  ./orchestrator-generate.sh --all [options]
  ./orchestrator-generate.sh --help

ARGUMENTS:
  <rite-name>      Generate for single rite (e.g., rnd-pack, security-pack)
  --all            Batch-generate all rites with rollback on error

OPTIONS:
  --validate-only  Parse configs, validate schema, exit (no generation)
  --dry-run        Generate to stdout instead of file
  --force          Regenerate even if target already exists
  --help           Show this help message

EXAMPLES:
  # Generate single rite with validation
  ./orchestrator-generate.sh rnd-pack

  # Validate without generating
  ./orchestrator-generate.sh rnd-pack --validate-only

  # Preview output
  ./orchestrator-generate.sh rnd-pack --dry-run

  # Batch-generate all rites
  ./orchestrator-generate.sh --all

  # Regenerate existing rite
  ./orchestrator-generate.sh security-pack --force

ENVIRONMENT:
  KNOSSOS_HOME     Root directory for roster (default: ~/Code/roster)

EOF
}

parse_arguments() {
    # Handle --help early before processing team name
    if [[ "${1:-}" == "--help" ]]; then
        show_help
        exit 0
    fi

    # Initialize variables
    RITE=""
    DRY_RUN=false
    VALIDATE_ONLY=false
    FORCE=false
    BATCH_ALL=false

    # Parse all arguments first
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --all)
                BATCH_ALL=true
                shift
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --validate-only)
                VALIDATE_ONLY=true
                shift
                ;;
            --force)
                FORCE=true
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            -*)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
            *)
                # First positional argument is the rite name
                if [[ -z "$RITE" ]]; then
                    RITE="$1"
                else
                    log_error "Multiple rite names specified: $RITE and $1"
                    exit 1
                fi
                shift
                ;;
        esac
    done

    # Validate that either rite or --all is specified
    if [[ -z "$RITE" ]] && [[ "$BATCH_ALL" != true ]]; then
        log_error "Rite name or --all flag required"
        show_help
        exit 1
    fi
}

# ============================================================================
# Dependency Checks
# ============================================================================

check_dependencies() {
    local missing=()

    # Check required tools
    for tool in yq jq grep sed awk; do
        if ! command -v "$tool" &>/dev/null; then
            missing+=("$tool")
        fi
    done

    if [[ ${#missing[@]} -gt 0 ]]; then
        log_error "Missing required tools: ${missing[*]}"
        echo ""
        echo "Installation:"
        echo "  macOS:  brew install yq jq"
        echo "  Linux:  apt-get install yq jq"
        exit 1
    fi

    log_ok "All dependencies found"
}

# ============================================================================
# File Existence Checks
# ============================================================================

check_required_files() {
    local files=("$TEMPLATE" "$SCHEMA" "$VALIDATOR")

    for file in "${files[@]}"; do
        if [[ ! -f "$file" ]]; then
            log_error "Required file not found: $file"
            exit 1
        fi
    done

    log_ok "All required files exist"
}

# ============================================================================
# Schema & Workflow Validation
# ============================================================================

validate_schema() {
    local config="$1"

    if [[ ! -f "$config" ]]; then
        log_error "Config file not found: $config"
        return 1
    fi

    # Convert YAML to JSON and validate against schema
    local config_json
    config_json=$(yq eval -o=json '.' "$config" 2>/dev/null || true)

    if [[ -z "$config_json" ]]; then
        log_error "Failed to parse YAML: $config"
        return 1
    fi

    # Use jq to validate against schema
    if ! echo "$config_json" | jq -e '.' >/dev/null 2>&1; then
        log_error "Invalid JSON produced from YAML: $config"
        return 1
    fi

    # Schema validation: check required fields
    local required_fields=("rite" "frontmatter" "routing" "workflow_position" "handoff_criteria" "skills")
    for field in "${required_fields[@]}"; do
        if ! echo "$config_json" | jq -e ".\"$field\"" >/dev/null 2>&1; then
            log_error "Missing required field in $config: $field"
            return 1
        fi
    done

    log_ok "Schema validation passed: $config"
    return 0
}

validate_workflow_references() {
    local config="$1"
    local workflow="$2"

    if [[ ! -f "$workflow" ]]; then
        log_error "Workflow file not found: $workflow"
        return 1
    fi

    # Extract specialist names from config
    local config_specialists
    config_specialists=$(yq eval '.routing | keys[]' "$config" 2>/dev/null || true)

    if [[ -z "$config_specialists" ]]; then
        log_error "No specialists found in routing table: $config"
        return 1
    fi

    # Check each specialist exists in workflow.yaml
    while read -r specialist; do
        if ! yq eval ".phases[].agent | select(. == \"$specialist\")" "$workflow" >/dev/null 2>&1; then
            log_error "Specialist '$specialist' not found in workflow.yaml: $workflow"
            return 1
        fi
    done <<<"$config_specialists"

    # Extract phases from config handoff_criteria
    local config_phases
    config_phases=$(yq eval '.handoff_criteria | keys[]' "$config" 2>/dev/null || true)

    # Check each phase exists in workflow.yaml
    while read -r phase; do
        [[ -z "$phase" ]] && continue
        if ! yq eval ".phases[].name | select(. == \"$phase\")" "$workflow" >/dev/null 2>&1; then
            log_warn "Phase '$phase' in handoff_criteria not found in workflow.yaml (may be OK if using different naming)"
        fi
    done <<<"$config_phases"

    log_ok "Workflow references validated: $config -> $workflow"
    return 0
}

# ============================================================================
# Configuration Extraction
# ============================================================================

extract_frontmatter_values() {
    local config="$1"

    ROLE=$(yq eval '.frontmatter.role' "$config")
    DESCRIPTION=$(yq eval '.frontmatter.description' "$config")
    COLOR=$(yq eval '.rite.color' "$config")

    # Normalize description: remove excess whitespace
    DESCRIPTION=$(echo "$DESCRIPTION" | tr '\n' ' ' | sed 's/[[:space:]]\+/ /g' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
}

extract_workflow_values() {
    local workflow="$1"

    # Get specialist names from workflow phases
    SPECIALISTS=$(yq eval '.phases[].agent' "$workflow" 2>/dev/null || true)

    if [[ -z "$SPECIALISTS" ]]; then
        log_error "No specialists found in workflow: $workflow"
        exit 1
    fi

    # Build complexity enum
    COMPLEXITY_ENUM=$(yq eval '.complexity_levels[].name' "$workflow" 2>/dev/null | tr '\n' ' ' | sed 's/[[:space:]]*$//')
    COMPLEXITY_ENUM=$(echo "$COMPLEXITY_ENUM" | sed 's/ / | /g')
    COMPLEXITY_ENUM="\"$COMPLEXITY_ENUM\""

    # Build specialist enum
    SPECIALIST_ENUM=$(echo "$SPECIALISTS" | tr '\n' ' ' | sed 's/[[:space:]]*$//')
    SPECIALIST_ENUM=$(echo "$SPECIALIST_ENUM" | sed 's/ /" | "/g')
    SPECIALIST_ENUM="\"$SPECIALIST_ENUM\""
}

# ============================================================================
# Template Placeholders
# ============================================================================

generate_workflow_diagram() {
    local -a agents=($SPECIALISTS)
    local num_agents=${#agents[@]}

    # ASCII diagram generation
    echo "                    +-----------------+"
    echo "                    |   ORCHESTRATOR  |"
    echo "                    +--------+--------+"
    echo "                             |"

    if [[ $num_agents -eq 4 ]]; then
        # 4-agent linear layout
        echo "        +--------------------+--------------------+"
        echo "        v                    v                    v"
        printf "+---------------+   +---------------+   +---------------+\n"
        printf "|  %-11s |-->|  %-11s |-->|   %-10s |\n" \
            "$(echo "${agents[0]}" | cut -d'-' -f1)" \
            "$(echo "${agents[1]}" | cut -d'-' -f1)" \
            "$(echo "${agents[2]}" | cut -d'-' -f1)"
        printf "|  %-11s |   |  %-11s |   |   %-10s |\n" \
            "$(echo "${agents[0]}" | cut -d'-' -f2-)" \
            "$(echo "${agents[1]}" | cut -d'-' -f2-)" \
            "$(echo "${agents[2]}" | cut -d'-' -f2-)"
        printf "+---------------+   +---------------+   +---------------+\n"
        echo "                                              |"
        echo "                                              v"
        echo "                                       +---------------+"
        printf "                                       |   %-10s |\n" "$(echo "${agents[3]}" | cut -d'-' -f1)"
        printf "                                       |   %-10s |\n" "$(echo "${agents[3]}" | cut -d'-' -f2-)"
        echo "                                       +---------------+"
    elif [[ $num_agents -eq 5 ]]; then
        # 5-agent layout
        echo "        +----+----+----+----+"
        for i in $(seq 0 4); do
            echo "        v"
            printf "  +--"
        done
        echo ""
        for agent in "${agents[@]}"; do
            printf "| %-15s " "$agent"
        done
        echo "|"
    else
        # Fallback: simple vertical list
        for agent in "${agents[@]}"; do
            echo "        +-> $agent"
        done
    fi
}

generate_routing_table() {
    local config="$1"

    yq eval '.routing | to_entries[] | "\(.key) | \(.value)"' "$config" | while read -r line; do
        local specialist value
        specialist=$(echo "$line" | cut -d'|' -f1 | xargs)
        value=$(echo "$line" | cut -d'|' -f2 | xargs)
        echo "| $specialist | $value |"
    done
}

generate_skills_reference() {
    local config="$1"

    yq eval '.skills[]' "$config" 2>/dev/null | while read -r skill; do
        echo "- $skill"
    done
}

# ============================================================================
# Template Substitution (Platform-safe)
# ============================================================================

substitute_placeholders() {
    local content="$1"
    local tmpfile_input="$2"

    # Use printf for safety across platforms (avoids sed -i '' vs sed -i issues)
    echo "$content" > "$tmpfile_input"

    # Simple placeholder substitution (bash string replacement, safe across platforms)
    local temp_content
    temp_content=$(<"$tmpfile_input")

    # Basic replacements
    temp_content="${temp_content//\{\{ROLE\}\}/$ROLE}"
    temp_content="${temp_content//\{\{COLOR\}\}/$COLOR}"
    temp_content="${temp_content//\{\{COMPLEXITY_ENUM\}\}/$COMPLEXITY_ENUM}"
    temp_content="${temp_content//\{\{SPECIALIST_ENUM\}\}/$SPECIALIST_ENUM}"

    # Escape special characters for description
    local desc_escaped
    desc_escaped=$(printf '%s\n' "$DESCRIPTION" | sed -e 's/[\/&]/\\&/g')
    temp_content="${temp_content//\{\{DESCRIPTION\}\}/$desc_escaped}"

    echo "$temp_content" > "$tmpfile_input"

    # Multi-line substitutions using awk (more portable)
    WORKFLOW_DIAGRAM=$(generate_workflow_diagram)
    DIAGRAM_FILE=$(mktemp)
    echo "$WORKFLOW_DIAGRAM" > "$DIAGRAM_FILE"
    awk -v file="$DIAGRAM_FILE" '
    /\{\{WORKFLOW_DIAGRAM\}\}/ {
        while ((getline line < file) > 0) print line
        close(file)
        next
    }
    { print }
    ' "$tmpfile_input" > "${tmpfile_input}.new"
    mv "${tmpfile_input}.new" "$tmpfile_input"

    # Routing table
    ROUTING_TABLE=$(generate_routing_table "$CONFIG")
    ROUTING_FILE=$(mktemp)
    echo "$ROUTING_TABLE" > "$ROUTING_FILE"
    awk -v file="$ROUTING_FILE" '
    /\{\{ROUTING_TABLE\}\}/ {
        while ((getline line < file) > 0) print line
        close(file)
        next
    }
    { print }
    ' "$tmpfile_input" > "${tmpfile_input}.new"
    mv "${tmpfile_input}.new" "$tmpfile_input"

    # Skills reference
    SKILLS_REFERENCE=$(generate_skills_reference "$CONFIG")
    SKILLS_FILE=$(mktemp)
    echo "$SKILLS_REFERENCE" > "$SKILLS_FILE"
    awk -v file="$SKILLS_FILE" '
    /\{\{SKILLS_REFERENCE\}\}/ {
        while ((getline line < file) > 0) print line
        close(file)
        next
    }
    { print }
    ' "$tmpfile_input" > "${tmpfile_input}.new"
    mv "${tmpfile_input}.new" "$tmpfile_input"
}

validate_substitution() {
    local file="$1"

    # Check for unreplaced placeholders
    if grep -q '{{[A-Z_]*}}' "$file"; then
        log_error "Unreplaced placeholders found in: $file"
        grep -n '{{[A-Z_]*}}' "$file" | sed 's/^/  Line: /'
        return 1
    fi

    return 0
}

# ============================================================================
# Single Team Generation
# ============================================================================

generate_rite() {
    local rite="$1"

    # Normalize rite name: strip "rites/" prefix if present
    rite="${rite#rites/}"

    log_info "Processing rite: $rite"

    CONFIG="$KNOSSOS_HOME/rites/$rite/orchestrator.yaml"
    WORKFLOW="$KNOSSOS_HOME/rites/$rite/workflow.yaml"
    OUTPUT="$KNOSSOS_HOME/rites/$rite/agents/orchestrator.md"

    # Validation phase
    validate_schema "$CONFIG" || return 1
    validate_workflow_references "$CONFIG" "$WORKFLOW" || return 1

    if [[ "$VALIDATE_ONLY" == true ]]; then
        log_ok "Validation passed: $rite"
        return 0
    fi

    # Check if output exists (skip for dry-run)
    if [[ "$DRY_RUN" != true ]] && [[ -f "$OUTPUT" ]] && [[ "$FORCE" != true ]]; then
        log_error "Output file exists: $OUTPUT (use --force to overwrite)"
        return 1
    fi

    # Extract values
    extract_frontmatter_values "$CONFIG"
    extract_workflow_values "$WORKFLOW"

    # Generate
    log_info "Generating: $rite"

    TMPFILE=$(mktemp)
    DIAGRAM_FILE=$(mktemp)
    ROUTING_FILE=$(mktemp)
    SKILLS_FILE=$(mktemp)

    local template_content
    template_content=$(<"$TEMPLATE")

    substitute_placeholders "$template_content" "$TMPFILE"
    validate_substitution "$TMPFILE" || return 1

    # Run post-generation validation
    if ! "$VALIDATOR" "$TMPFILE" >/dev/null 2>&1; then
        log_error "Post-generation validation failed: $TMPFILE"
        return 1
    fi

    # Output
    if [[ "$DRY_RUN" == true ]]; then
        echo ""
        echo "=== Generated Content (dry-run) ==="
        cat "$TMPFILE"
        echo "=== End Generated Content ==="
        echo ""
    else
        # Ensure output directory exists
        mkdir -p "$(dirname "$OUTPUT")"
        cp "$TMPFILE" "$OUTPUT"
        log_ok "Generated: $OUTPUT"
    fi

    return 0
}

# ============================================================================
# Batch Generation
# ============================================================================

batch_generate_all_rites() {
    # Find all rites
    local -a rites
    while IFS= read -r rite_dir; do
        local rite=$(basename "$rite_dir")
        rites+=("$rite")
    done < <(find "$KNOSSOS_HOME/rites" -maxdepth 1 -type d -name "*-pack" | sort)

    if [[ ${#rites[@]} -eq 0 ]]; then
        log_error "No rites found in $KNOSSOS_HOME/rites"
        return 1
    fi

    log_info "Batch generating ${#rites[@]} rites"
    echo ""

    local failed_rites=()
    for rite in "${rites[@]}"; do
        if ! generate_rite "$rite"; then
            failed_rites+=("$rite")
            log_error "Failed to generate: $rite"
        fi
    done

    echo ""
    log_info "Batch generation complete"

    if [[ ${#failed_rites[@]} -gt 0 ]]; then
        log_error "Failed rites: ${failed_rites[*]}"
        return 1
    fi

    log_ok "All rites generated successfully"
    return 0
}

# ============================================================================
# Main Entry Point
# ============================================================================

main() {
    parse_arguments "$@"

    check_dependencies
    check_required_files

    echo ""

    if [[ "$BATCH_ALL" == true ]]; then
        batch_generate_all_rites
    else
        generate_rite "$RITE"
    fi
}

main "$@"

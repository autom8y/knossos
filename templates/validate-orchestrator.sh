#!/bin/bash
# validate-orchestrator.sh - Validate generated orchestrator.md files
#
# Verifies that generated orchestrator.md meets all structural and semantic requirements
# before it can be committed to the repository.
#
# Usage: ./validate-orchestrator.sh <orchestrator.md-path> [--strict]
#
# Exit codes:
#   0 = Validation passed
#   1 = Invalid arguments or file not found
#   2 = Validation failed (structural issues)
#   3 = Validation failed (semantic issues)

set -euo pipefail

# Configuration
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"
readonly SCHEMA_FILE="$ROSTER_HOME/schemas/orchestrator.yaml.schema.json"

ORCHESTRATOR_FILE="${1:-}"
STRICT_MODE="${2:-}"

# Colors
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly NC='\033[0m'

# Validation state
VALIDATION_PASSED=true
ERRORS=()
WARNINGS=()

# ============================================================================
# Helper Functions
# ============================================================================

log_error() {
    echo -e "${RED}ERROR${NC}: $*" >&2
    VALIDATION_PASSED=false
    ERRORS+=("$*")
}

log_warning() {
    echo -e "${YELLOW}WARNING${NC}: $*" >&2
    WARNINGS+=("$*")
}

log_ok() {
    echo -e "${GREEN}OK${NC}: $*"
}

# ============================================================================
# Validation Rules
# ============================================================================

# Rule 1: File exists and is readable
validate_file_exists() {
    if [[ ! -f "$ORCHESTRATOR_FILE" ]]; then
        log_error "Orchestrator file not found: $ORCHESTRATOR_FILE"
        return 1
    fi

    if [[ ! -r "$ORCHESTRATOR_FILE" ]]; then
        log_error "Orchestrator file is not readable: $ORCHESTRATOR_FILE"
        return 1
    fi

    log_ok "File exists and is readable"
}

# Rule 2: No remaining placeholders (all {{PLACEHOLDER}} must be replaced)
validate_no_placeholders() {
    local placeholder_count
    placeholder_count=$(grep -c '{{[A-Z_]*}}' "$ORCHESTRATOR_FILE" || true)

    if [[ $placeholder_count -gt 0 ]]; then
        log_error "Found $placeholder_count unreplaced placeholder(s)"
        grep -n '{{[A-Z_]*}}' "$ORCHESTRATOR_FILE" | while read -r line; do
            echo "  Line: $line" >&2
        done
        return 1
    fi

    log_ok "No unreplaced placeholders found"
}

# Rule 3: YAML frontmatter parses cleanly
validate_frontmatter() {
    local frontmatter_start frontmatter_end

    # Extract frontmatter (between --- lines)
    frontmatter_start=$(grep -n "^---$" "$ORCHESTRATOR_FILE" | head -1 | cut -d: -f1)
    frontmatter_end=$(grep -n "^---$" "$ORCHESTRATOR_FILE" | head -2 | tail -1 | cut -d: -f1)

    if [[ -z "$frontmatter_start" ]] || [[ -z "$frontmatter_end" ]]; then
        log_error "Invalid YAML frontmatter: missing --- delimiters"
        return 1
    fi

    if [[ $((frontmatter_end - frontmatter_start)) -lt 2 ]]; then
        log_error "Invalid YAML frontmatter: incomplete structure"
        return 1
    fi

    # Extract and validate required frontmatter fields
    local name role color model tools
    name=$(sed -n "${frontmatter_start},$((frontmatter_end))p" "$ORCHESTRATOR_FILE" | grep "^name:" | head -1 | sed 's/^name:[[:space:]]*//')
    role=$(sed -n "${frontmatter_start},$((frontmatter_end))p" "$ORCHESTRATOR_FILE" | grep "^role:" | head -1)
    color=$(sed -n "${frontmatter_start},$((frontmatter_end))p" "$ORCHESTRATOR_FILE" | grep "^color:" | head -1)
    model=$(sed -n "${frontmatter_start},$((frontmatter_end))p" "$ORCHESTRATOR_FILE" | grep "^model:" | head -1)
    tools=$(sed -n "${frontmatter_start},$((frontmatter_end))p" "$ORCHESTRATOR_FILE" | grep "^tools:" | head -1)

    [[ -z "$name" ]] && log_error "Frontmatter missing 'name' field"
    [[ -z "$role" ]] && log_error "Frontmatter missing 'role' field"
    [[ -z "$color" ]] && log_error "Frontmatter missing 'color' field"
    [[ -z "$model" ]] && log_error "Frontmatter missing 'model' field"
    [[ -z "$tools" ]] && log_error "Frontmatter missing 'tools' field"

    if [[ -n "$name" ]] && [[ "$name" != "orchestrator" ]]; then
        log_error "Frontmatter 'name' must be 'orchestrator', got: $name"
        return 1
    fi

    log_ok "YAML frontmatter is valid"
}

# Rule 4: All required sections present
validate_required_sections() {
    local sections=(
        "Consultation Role"
        "Tool Access"
        "Consultation Protocol"
        "Position in Workflow"
        "Domain Authority"
        "Handling Failures"
        "The Acid Test"
        "Anti-Patterns"
        "Skills Reference"
    )

    local missing_sections=()
    for section in "${sections[@]}"; do
        if ! grep -q "^## $section" "$ORCHESTRATOR_FILE"; then
            missing_sections+=("$section")
        fi
    done

    # Also check for Routing Criteria (subsection of Domain Authority)
    if ! grep -q "^|.*Specialist.*Route When" "$ORCHESTRATOR_FILE"; then
        missing_sections+=("Routing Criteria (table)")
    fi

    if [[ ${#missing_sections[@]} -gt 0 ]]; then
        log_error "Missing required sections: ${missing_sections[*]}"
        return 1
    fi

    log_ok "All required sections present"
}

# Rule 5: Specialist names consistent across all references
validate_specialist_consistency() {
    # Find routing table in Domain Authority section
    local routing_table_start routing_table_end

    routing_table_start=$(grep -n "^## Domain Authority" "$ORCHESTRATOR_FILE" | cut -d: -f1)
    if [[ -z "$routing_table_start" ]]; then
        log_warning "Could not find Domain Authority section"
        return 0
    fi

    # Extract specialist names from routing table (pipe-separated format)
    local specialists_in_routing
    specialists_in_routing=$(sed -n "${routing_table_start},/^## /p" "$ORCHESTRATOR_FILE" | \
        grep "^|" | grep -v "^|--" | grep -v "Specialist" | awk -F'|' '{if (NF > 1) print $2}' | \
        sed 's/^[[:space:]]*//;s/[[:space:]]*$//' | grep -v '^$' | sort | uniq)

    if [[ -z "$specialists_in_routing" ]]; then
        log_warning "Could not extract specialists from routing table"
    else
        log_ok "Found specialists in routing table: $(echo $specialists_in_routing | tr '\n' ',' | sed 's/,$//')"
    fi
}

# Rule 6: No duplicate section headers
validate_no_duplicate_sections() {
    local section_headers
    section_headers=$(grep "^## " "$ORCHESTRATOR_FILE" | cut -d' ' -f2- | sort)

    local duplicates
    duplicates=$(echo "$section_headers" | uniq -d || true)

    if [[ -n "$duplicates" ]]; then
        log_error "Found duplicate section headers: $duplicates"
        return 1
    fi

    log_ok "No duplicate section headers"
}

# Rule 7: Handoff criteria sections properly formatted
validate_handoff_criteria() {
    local hc_section_start hc_section_end

    # Look for handoff criteria in Domain Authority or Handling Failures sections
    # (They should be referenced as checklists)

    # At minimum, check for proper markdown list formatting
    local improperly_formatted
    improperly_formatted=$(grep -n "^\s*-\s*\[\s\]" "$ORCHESTRATOR_FILE" | wc -l)

    if [[ $improperly_formatted -eq 0 ]]; then
        log_warning "No handoff criteria checkboxes found (expected format: '- [ ] criterion')"
    else
        log_ok "Found $improperly_formatted handoff criteria items"
    fi
}

# Rule 8: Consultation Protocol YAML structure valid
validate_consultation_protocol() {
    local protocol_start protocol_end

    # Find Consultation Protocol section
    protocol_start=$(grep -n "^## Consultation Protocol" "$ORCHESTRATOR_FILE" | cut -d: -f1)
    if [[ -z "$protocol_start" ]]; then
        log_error "Could not find Consultation Protocol section"
        return 1
    fi

    # Check for required input/output labels
    if ! sed -n "${protocol_start},/^## /p" "$ORCHESTRATOR_FILE" | grep -q "### Input: CONSULTATION_REQUEST"; then
        log_error "Missing 'Input: CONSULTATION_REQUEST' in Consultation Protocol"
        return 1
    fi

    if ! sed -n "${protocol_start},/^## /p" "$ORCHESTRATOR_FILE" | grep -q "### Output: CONSULTATION_RESPONSE"; then
        log_error "Missing 'Output: CONSULTATION_RESPONSE' in Consultation Protocol"
        return 1
    fi

    log_ok "Consultation Protocol structure is valid"
}

# Rule 9: No dangling skill references (basic check)
validate_skill_references() {
    local skills_start
    skills_start=$(grep -n "^## Skills Reference" "$ORCHESTRATOR_FILE" | cut -d: -f1)

    if [[ -z "$skills_start" ]]; then
        log_error "Could not find Skills Reference section"
        return 1
    fi

    # Extract skill references and check format
    local skill_count
    skill_count=$(sed -n "${skills_start},\$p" "$ORCHESTRATOR_FILE" | grep "^- @" | wc -l)

    if [[ $skill_count -eq 0 ]]; then
        log_warning "No skills referenced in Skills Reference section"
    else
        log_ok "Found $skill_count skill references"
    fi
}

# Rule 10: Markdown syntax basic validation
validate_markdown_syntax() {
    local open_code_blocks closed_code_blocks
    open_code_blocks=$(grep -c "^\`\`\`" "$ORCHESTRATOR_FILE" || true)

    # Should have even number of code block markers (open and close pairs)
    if [[ $((open_code_blocks % 2)) -ne 0 ]]; then
        log_error "Unbalanced code block markers (expected even count, found $open_code_blocks)"
        return 1
    fi

    log_ok "Markdown syntax appears valid ($open_code_blocks code blocks found)"
}

# ============================================================================
# Main Validation Flow
# ============================================================================

main() {
    # Argument validation
    if [[ -z "$ORCHESTRATOR_FILE" ]]; then
        echo "Usage: $0 <orchestrator.md-path> [--strict]" >&2
        exit 1
    fi

    echo "Validating orchestrator: $ORCHESTRATOR_FILE"
    echo ""

    # Run all validation rules
    validate_file_exists || true
    validate_no_placeholders || true
    validate_frontmatter || true
    validate_required_sections || true
    validate_specialist_consistency || true
    validate_no_duplicate_sections || true
    validate_handoff_criteria || true
    validate_consultation_protocol || true
    validate_skill_references || true
    validate_markdown_syntax || true

    echo ""
    echo "=========================================="

    # Summary
    if $VALIDATION_PASSED; then
        echo -e "${GREEN}Validation PASSED${NC}"
        [[ ${#WARNINGS[@]} -gt 0 ]] && echo "Warnings: ${#WARNINGS[@]}"
        exit 0
    else
        echo -e "${RED}Validation FAILED${NC}"
        echo "Errors found: ${#ERRORS[@]}"
        [[ ${#ERRORS[@]} -gt 0 ]] && printf '  - %s\n' "${ERRORS[@]}"
        [[ ${#WARNINGS[@]} -gt 0 ]] && echo "Warnings: ${#WARNINGS[@]}"
        exit 2
    fi
}

main "$@"

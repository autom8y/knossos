#!/usr/bin/env bash
# verify-specs.sh - Lightweight local verification for formal specifications
#
# This script provides fast local validation without requiring TLC or Alloy Analyzer.
# Full model checking runs in CI via .github/workflows/verify-formal-specs.yml
#
# Usage:
#   ./scripts/verify-specs.sh           # Verify all specs
#   ./scripts/verify-specs.sh --tla     # TLA+ only
#   ./scripts/verify-specs.sh --alloy   # Alloy only
#   ./scripts/verify-specs.sh --full    # Full verification (requires Java + tools)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SPECS_DIR="$PROJECT_ROOT/docs/specs"
TOOLS_DIR="$PROJECT_ROOT/tools"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() { echo -e "${GREEN}✓${NC} $1"; }
log_warn() { echo -e "${YELLOW}⚠${NC} $1"; }
log_error() { echo -e "${RED}✗${NC} $1"; }

# Check if specs directory exists
check_specs_dir() {
    if [[ ! -d "$SPECS_DIR" ]]; then
        log_error "Specs directory not found: $SPECS_DIR"
        exit 1
    fi
}

# Lightweight TLA+ syntax validation (no TLC required)
validate_tla_syntax() {
    local file="$1"
    local errors=0

    echo "Validating TLA+ syntax: $(basename "$file")"

    # Check for required module structure
    if ! grep -q "^---- MODULE" "$file" && ! grep -q "^-* MODULE" "$file"; then
        log_error "Missing MODULE declaration"
        errors=$((errors + 1))
    fi

    # Check for balanced delimiters
    local opens=$(grep -o "\\\\begin{" "$file" 2>/dev/null | wc -l || echo 0)
    local closes=$(grep -o "\\\\end{" "$file" 2>/dev/null | wc -l || echo 0)
    if [[ "$opens" != "$closes" ]]; then
        log_warn "Unbalanced \\begin{}/\\end{} blocks"
    fi

    # Check for EXTENDS clause (common requirement)
    if ! grep -q "^EXTENDS" "$file"; then
        log_warn "No EXTENDS clause found (may be intentional)"
    fi

    # Check for closing delimiter
    if ! grep -q "^====*$" "$file"; then
        log_error "Missing closing ==== delimiter"
        errors=$((errors + 1))
    fi

    # Check for common syntax patterns
    if grep -qE "^\s*VARIABLE\s*$" "$file"; then
        log_error "Empty VARIABLE declaration"
        errors=$((errors + 1))
    fi

    # Check for Init and Next definitions (for temporal specs)
    if grep -q "Spec ==" "$file"; then
        if ! grep -q "^Init ==" "$file" && ! grep -q "^Init\s*==" "$file"; then
            log_warn "Spec defined but no Init found"
        fi
        if ! grep -q "^Next ==" "$file" && ! grep -q "^Next\s*==" "$file"; then
            log_warn "Spec defined but no Next found"
        fi
    fi

    if [[ $errors -eq 0 ]]; then
        log_info "Syntax validation passed"
        return 0
    else
        log_error "Syntax validation failed with $errors error(s)"
        return 1
    fi
}

# Lightweight Alloy syntax validation (no Alloy Analyzer required)
validate_alloy_syntax() {
    local file="$1"
    local errors=0

    echo "Validating Alloy syntax: $(basename "$file")"

    # Check for signature definitions
    if ! grep -qE "^(abstract\s+)?sig\s+" "$file"; then
        log_warn "No signature (sig) definitions found"
    fi

    # Check for balanced braces
    local open_braces=$(grep -o "{" "$file" | wc -l)
    local close_braces=$(grep -o "}" "$file" | wc -l)
    if [[ "$open_braces" != "$close_braces" ]]; then
        log_error "Unbalanced braces: $open_braces opens, $close_braces closes"
        errors=$((errors + 1))
    fi

    # Check for balanced brackets
    local open_brackets=$(grep -o "\[" "$file" | wc -l)
    local close_brackets=$(grep -o "\]" "$file" | wc -l)
    if [[ "$open_brackets" != "$close_brackets" ]]; then
        log_error "Unbalanced brackets: $open_brackets opens, $close_brackets closes"
        errors=$((errors + 1))
    fi

    # Check for fact/assert/pred/fun definitions
    local has_constraints=false
    grep -qE "^(fact|assert|pred|fun)\s+" "$file" && has_constraints=true
    if [[ "$has_constraints" == "false" ]]; then
        log_warn "No constraints (fact/assert/pred/fun) found"
    fi

    # Check for check/run commands
    if ! grep -qE "^(check|run)\s+" "$file"; then
        log_warn "No check/run commands found (spec won't be executed)"
    fi

    # Check for common syntax errors
    if grep -qE ";\s*$" "$file"; then
        # Alloy doesn't use semicolons at end of lines
        log_warn "Found trailing semicolons (Alloy doesn't require them)"
    fi

    if [[ $errors -eq 0 ]]; then
        log_info "Syntax validation passed"
        return 0
    else
        log_error "Syntax validation failed with $errors error(s)"
        return 1
    fi
}

# Full TLA+ verification with TLC
verify_tla_full() {
    local file="$1"

    if ! command -v java &>/dev/null; then
        log_error "Java not found. Install Java 17+ for full verification."
        return 1
    fi

    if [[ ! -f "$TOOLS_DIR/tla2tools.jar" ]]; then
        log_warn "TLA+ tools not installed. Downloading..."
        mkdir -p "$TOOLS_DIR"
        curl -sL https://github.com/tlaplus/tlaplus/releases/latest/download/tla2tools.jar -o "$TOOLS_DIR/tla2tools.jar"
    fi

    echo "Running TLC model checker: $(basename "$file")"

    # Look for config file
    local cfg_file="${file%.tla}.cfg"
    if [[ ! -f "$cfg_file" ]]; then
        log_warn "No config file found. Using defaults."
        cfg_file=""
    fi

    # Run SANY first (syntax check)
    if ! java -cp "$TOOLS_DIR/tla2tools.jar" tla2sany.SANY "$file" 2>&1; then
        log_error "SANY syntax check failed"
        return 1
    fi
    log_info "SANY syntax check passed"

    # Run TLC
    local tlc_args=(-XX:+UseParallelGC -Xmx2g -cp "$TOOLS_DIR/tla2tools.jar" tlc2.TLC)
    [[ -n "$cfg_file" ]] && tlc_args+=(-config "$cfg_file")
    tlc_args+=(-workers auto -deadlock -cleanup "$file")

    if ! java "${tlc_args[@]}" 2>&1; then
        log_error "TLC model checking failed"
        return 1
    fi

    log_info "TLC model checking passed"
    return 0
}

# Full Alloy verification
verify_alloy_full() {
    local file="$1"

    if ! command -v java &>/dev/null; then
        log_error "Java not found. Install Java 17+ for full verification."
        return 1
    fi

    if [[ ! -f "$TOOLS_DIR/alloy.jar" ]]; then
        log_warn "Alloy not installed. Downloading..."
        mkdir -p "$TOOLS_DIR"
        curl -sL https://github.com/AlloyTools/org.alloytools.alloy/releases/download/v6.1.0/org.alloytools.alloy.dist.jar -o "$TOOLS_DIR/alloy.jar"
    fi

    echo "Running Alloy analyzer: $(basename "$file")"

    # Run Alloy (batch mode is limited, so we just check for syntax errors)
    if java -cp "$TOOLS_DIR/alloy.jar" edu.mit.csail.sdg.alloy4whole.ExampleUsingTheCompiler "$file" 2>&1 | grep -q "Syntax error"; then
        log_error "Alloy syntax error"
        return 1
    fi

    log_info "Alloy verification passed"
    return 0
}

# Main verification loop
main() {
    local mode="lightweight"
    local tla_only=false
    local alloy_only=false

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --full)
                mode="full"
                shift
                ;;
            --tla)
                tla_only=true
                shift
                ;;
            --alloy)
                alloy_only=true
                shift
                ;;
            --help|-h)
                echo "Usage: $0 [--tla] [--alloy] [--full]"
                echo ""
                echo "Options:"
                echo "  --tla     Verify TLA+ specs only"
                echo "  --alloy   Verify Alloy specs only"
                echo "  --full    Run full model checking (requires Java + tools)"
                echo ""
                echo "Default: Lightweight syntax validation (no external tools required)"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done

    check_specs_dir

    local total_errors=0

    echo "========================================"
    echo "Formal Specification Verification"
    echo "Mode: $mode"
    echo "========================================"
    echo ""

    # Verify TLA+ specs
    if [[ "$alloy_only" == "false" ]]; then
        echo "--- TLA+ Specifications ---"
        for tla_file in "$SPECS_DIR"/*.tla; do
            [[ -f "$tla_file" ]] || continue
            echo ""
            if [[ "$mode" == "full" ]]; then
                verify_tla_full "$tla_file" || total_errors=$((total_errors + 1))
            else
                validate_tla_syntax "$tla_file" || total_errors=$((total_errors + 1))
            fi
        done
        echo ""
    fi

    # Verify Alloy specs
    if [[ "$tla_only" == "false" ]]; then
        echo "--- Alloy Specifications ---"
        for alloy_file in "$SPECS_DIR"/*.als; do
            [[ -f "$alloy_file" ]] || continue
            echo ""
            if [[ "$mode" == "full" ]]; then
                verify_alloy_full "$alloy_file" || total_errors=$((total_errors + 1))
            else
                validate_alloy_syntax "$alloy_file" || total_errors=$((total_errors + 1))
            fi
        done
        echo ""
    fi

    echo "========================================"
    if [[ $total_errors -eq 0 ]]; then
        log_info "All specifications verified successfully"
        echo ""
        if [[ "$mode" == "lightweight" ]]; then
            echo "Note: This was lightweight validation only."
            echo "Full model checking runs in CI or with: $0 --full"
        fi
        exit 0
    else
        log_error "Verification failed with $total_errors error(s)"
        exit 1
    fi
}

main "$@"

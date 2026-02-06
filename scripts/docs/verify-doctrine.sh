#!/bin/bash
# Doctrine Documentation Verification Script
# Checks for broken links, valid CLI commands, and structural integrity

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="${SCRIPT_DIR}/../.."
DOCTRINE_DIR="${PROJECT_ROOT}/docs/doctrine"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

errors=0
warnings=0

echo "================================================"
echo "Doctrine Documentation Verification"
echo "================================================"
echo ""

# -----------------------------------------------------------------------------
# Check 1: Verify all markdown files exist in expected locations
# -----------------------------------------------------------------------------
echo ">> Checking documentation structure..."

required_dirs=(
    "docs/doctrine/operations/cli-reference"
    "docs/doctrine/rites"
    "docs/doctrine/guides"
    "docs/doctrine/philosophy"
    "docs/doctrine/reference"
    "docs/doctrine/compliance"
)

for dir in "${required_dirs[@]}"; do
    if [[ -d "${PROJECT_ROOT}/${dir}" ]]; then
        echo -e "  ${GREEN}✓${NC} ${dir}"
    else
        echo -e "  ${RED}✗${NC} ${dir} - MISSING"
        ((errors++))
    fi
done
echo ""

# -----------------------------------------------------------------------------
# Check 2: Verify CLI reference completeness
# -----------------------------------------------------------------------------
echo ">> Verifying CLI reference completeness..."

# Get command families from ari --help
cli_families=$(cd "$PROJECT_ROOT" && ari --help 2>/dev/null | grep -E "^\s{2}[a-z]+" | awk '{print $1}' | sort -u)

for family in $cli_families; do
    doc_file="${DOCTRINE_DIR}/operations/cli-reference/cli-${family}.md"
    if [[ -f "$doc_file" ]]; then
        echo -e "  ${GREEN}✓${NC} cli-${family}.md"
    else
        echo -e "  ${YELLOW}!${NC} cli-${family}.md - not found (may be intentional)"
        ((warnings++))
    fi
done
echo ""

# -----------------------------------------------------------------------------
# Check 3: Verify rite documentation completeness
# -----------------------------------------------------------------------------
echo ">> Verifying rite documentation..."

# Get rites from ari rite pantheon or manifest files
if command -v ari &>/dev/null; then
    rites=$(cd "$PROJECT_ROOT" && ari rite list 2>/dev/null | grep -E "^[a-z]" | awk '{print $1}' || echo "")
else
    # Fallback: check manifest directories
    rites=$(ls -1 "${PROJECT_ROOT}/knossos/rites/" 2>/dev/null | grep -v "^_" || echo "")
fi

for rite in $rites; do
    # Handle special cases like 10x-dev
    doc_file="${DOCTRINE_DIR}/rites/${rite}.md"
    if [[ -f "$doc_file" ]]; then
        echo -e "  ${GREEN}✓${NC} ${rite}.md"
    else
        echo -e "  ${YELLOW}!${NC} ${rite}.md - not documented"
        ((warnings++))
    fi
done
echo ""

# -----------------------------------------------------------------------------
# Check 4: Verify internal markdown links
# -----------------------------------------------------------------------------
echo ">> Checking internal links..."

find "${DOCTRINE_DIR}" -name "*.md" -type f | while read -r file; do
    # Extract relative markdown links
    links=$(grep -oE '\]\([^)]+\.md[^)]*\)' "$file" 2>/dev/null | grep -oE '\([^)]+\)' | tr -d '()' | grep "^\.\." || true)

    for link in $links; do
        # Remove anchor
        link_path="${link%%#*}"
        # Resolve relative path
        dir=$(dirname "$file")
        resolved=$(cd "$dir" && realpath -q "$link_path" 2>/dev/null || echo "")

        if [[ -z "$resolved" ]] || [[ ! -f "$resolved" ]]; then
            echo -e "  ${RED}✗${NC} Broken link in $(basename "$file"): $link_path"
            ((errors++))
        fi
    done
done
echo ""

# -----------------------------------------------------------------------------
# Check 5: Verify required sections in CLI docs
# -----------------------------------------------------------------------------
echo ">> Checking CLI doc structure..."

for doc in "${DOCTRINE_DIR}"/operations/cli-reference/cli-*.md; do
    [[ -f "$doc" ]] || continue
    name=$(basename "$doc")

    # Check for required sections
    has_synopsis=$(grep -q "^## Synopsis\|^## Commands" "$doc" && echo "yes" || echo "no")
    has_examples=$(grep -q "^## Examples\|^### Example" "$doc" && echo "yes" || echo "no")
    has_see_also=$(grep -q "^## See Also" "$doc" && echo "yes" || echo "no")

    if [[ "$has_synopsis" == "no" ]]; then
        echo -e "  ${YELLOW}!${NC} ${name} - missing Synopsis/Commands section"
        ((warnings++))
    fi
    if [[ "$has_examples" == "no" ]]; then
        echo -e "  ${YELLOW}!${NC} ${name} - missing Examples section"
        ((warnings++))
    fi
done
echo ""

# -----------------------------------------------------------------------------
# Summary
# -----------------------------------------------------------------------------
echo "================================================"
echo "Verification Complete"
echo "================================================"
echo -e "Errors:   ${errors}"
echo -e "Warnings: ${warnings}"
echo ""

if [[ $errors -gt 0 ]]; then
    echo -e "${RED}FAILED${NC} - Fix errors before proceeding"
    exit 1
elif [[ $warnings -gt 0 ]]; then
    echo -e "${YELLOW}PASSED WITH WARNINGS${NC}"
    exit 0
else
    echo -e "${GREEN}PASSED${NC}"
    exit 0
fi

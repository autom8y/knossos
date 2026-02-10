---
schema_name: gap-analysis
schema_version: "1.0"
file_pattern: "docs/ecosystem/GAP-*.md"
artifact_type: gap-analysis
---

# Gap Analysis Schema

> Canonical schema for Gap Analysis documents at `docs/ecosystem/GAP-{slug}.md`

## YAML Frontmatter

```yaml
---
# Required fields
title: string              # Human-readable title
type: string               # Must be "gap-analysis"
complexity: enum           # PATCH | MODULE | SYSTEM | MIGRATION
created_at: string         # ISO 8601 timestamp
status: enum               # in-progress | ready-for-design | archived

# Affected scope
affected_systems:          # Systems impacted by this analysis
  - enum                   # knossos | sync

# Issue summary
issue_count: integer       # Total issues identified
critical_count: integer    # Issues with critical severity
high_count: integer        # Issues with high severity
medium_count: integer      # Issues with medium severity

# Success criteria
success_criteria:          # How to verify fixes
  - id: string             # GAP-SC-001
    description: string    # Verification statement
    artifact_type: string  # What artifact proves this

# Optional fields
author: string             # Agent or user who created
root_cause: string         # Summary of root cause chain
dependencies: array        # Other gap analyses this depends on

# Schema versioning
schema_version: "1.0"      # Must be "1.0" for this version
---
```

## Required Fields

| Field | Type | Description | Set By |
|-------|------|-------------|--------|
| `title` | string | Human-readable title | ecosystem-analyst |
| `type` | string | Must be "gap-analysis" | ecosystem-analyst |
| `complexity` | enum | Scope classification | ecosystem-analyst |
| `created_at` | string | ISO 8601 creation timestamp | ecosystem-analyst |
| `status` | enum | Current analysis status | ecosystem-analyst |
| `affected_systems` | array | knossos, sync | ecosystem-analyst |
| `issue_count` | integer | Total issues found | ecosystem-analyst |
| `critical_count` | integer | Critical severity issues | ecosystem-analyst |
| `success_criteria` | array | Verification criteria | ecosystem-analyst |
| `schema_version` | string | Schema version for compatibility | ecosystem-analyst |

## Optional Fields

| Field | Type | Description | Set By |
|-------|------|-------------|--------|
| `author` | string | Creating agent or user | ecosystem-analyst |
| `root_cause` | string | Root cause summary | ecosystem-analyst |
| `high_count` | integer | High severity issues | ecosystem-analyst |
| `medium_count` | integer | Medium severity issues | ecosystem-analyst |
| `dependencies` | array | Related gap analyses | ecosystem-analyst |

## Success Criterion Object Schema

```yaml
success_criteria:
  - id: string             # "GAP-SC-001" format
    description: string    # Clear verification statement
    artifact_type: string  # context-design, implementation, etc.
    blocking: boolean      # (optional) Whether this blocks completion
```

## Valid Status Transitions

```
in-progress --complete--> ready-for-design
ready-for-design --implement--> archived
```

## Validation Rules

### Structure Validation
1. File MUST start with `---` on line 1
2. File MUST have closing `---` within first 50 lines
3. Content between delimiters MUST be valid YAML

### Field Validation
1. `type` MUST be exactly "gap-analysis"
2. `complexity` MUST be one of: PATCH, MODULE, SYSTEM, MIGRATION
3. `created_at` MUST be valid ISO 8601 timestamp
4. `status` MUST be one of: in-progress, ready-for-design, archived
5. `affected_systems` MUST be array with at least one of: knossos, sync
6. `issue_count` MUST be non-negative integer
7. `critical_count` MUST be non-negative integer and <= issue_count
8. `success_criteria` MUST be array with at least one item
9. `schema_version` MUST be "1.0"

### Issue Count Validation
1. `issue_count` MUST equal sum of critical + high + medium counts
2. `critical_count` + `high_count` + `medium_count` MUST equal `issue_count`

## Example: Valid Gap Analysis

```yaml
---
title: "Gap Analysis: Session State Validation"
type: gap-analysis
complexity: MODULE
created_at: "2025-12-29T09:00:00Z"
status: ready-for-design
affected_systems:
  - knossos
  - sync
author: ecosystem-analyst
issue_count: 5
critical_count: 2
high_count: 2
medium_count: 1
root_cause: "Session validation functions duplicated across 4 hooks with inconsistent error handling"
success_criteria:
  - id: GAP-SC-001
    description: "Single validate_session_context() function in session-state.sh"
    artifact_type: implementation
  - id: GAP-SC-002
    description: "All hooks use centralized validation"
    artifact_type: compatibility-report
  - id: GAP-SC-003
    description: "Error codes documented in schema"
    artifact_type: context-design
schema_version: "1.0"
---

## Executive Summary

Deep audit identified 5 issues in session state validation across hooks.

## Root Cause Chain

1. No central validation function defined
2. Each hook implements its own validation
3. Inconsistent error codes and messages
4. No schema documenting valid states

## Issue Inventory

### Category 1: Validation Duplication (2 Issues)

#### VAL-001: session-context.sh Inline Validation (CRITICAL)

**Description**: session-context.sh validates SESSION_CONTEXT.md inline instead of calling shared function.

**Evidence**:
- File: `.claude/hooks/session-context.sh`
- Lines 45-78: Inline YAML parsing and validation

**Impact**: Changes to validation logic require updating multiple files.

---

### Category 2: Error Handling (2 Issues)

#### ERR-001: Inconsistent Error Codes (HIGH)

**Description**: Different hooks return different error codes for same condition.

---

## Prioritized Fix List

| Priority | ID | Issue | Effort | Impact |
|----------|-----|-------|--------|--------|
| 1 | VAL-001 | Inline validation | Medium | Unblocks centralization |
| 2 | ERR-001 | Error codes | Low | Improves debugging |

## Success Criteria

- [ ] GAP-SC-001: Single validation function exists
- [ ] GAP-SC-002: All hooks use centralized validation
- [ ] GAP-SC-003: Error codes documented in schema
```

## Validation Function

```bash
# In ecosystem-validator.sh
# Usage: validate_gap_analysis "/path/to/GAP-example.md"
# Returns: 0=valid, 1=not found, 2=no opener, 3=no closer, 4=missing field, 5=invalid field

validate_gap_analysis() {
    local file="$1"
    local required_fields=("title" "type" "complexity" "created_at" "status" "affected_systems" "issue_count" "critical_count" "success_criteria" "schema_version")

    # Check file exists
    [ -f "$file" ] || { echo "File not found: $file" >&2; return 1; }

    # Check opening delimiter on line 1
    local first_line
    first_line=$(head -n 1 "$file")
    if [[ "$first_line" != "---" ]]; then
        echo "Invalid format: Missing opening '---' delimiter on line 1" >&2
        return 2
    fi

    # Check closing delimiter within first 50 lines
    local closing_line
    closing_line=$(head -n 50 "$file" | tail -n +2 | grep -n "^---$" | head -1 | cut -d: -f1)
    if [[ -z "$closing_line" ]]; then
        echo "Invalid format: Missing closing '---' delimiter within first 50 lines" >&2
        return 3
    fi

    # Extract frontmatter
    local frontmatter_end=$((closing_line + 1))
    local frontmatter
    frontmatter=$(sed -n "2,$((frontmatter_end))p" "$file" | sed '$d')

    # Check required fields
    local missing=()
    for field in "${required_fields[@]}"; do
        if ! echo "$frontmatter" | grep -q "^${field}:"; then
            missing+=("$field")
        fi
    done

    if [ ${#missing[@]} -gt 0 ]; then
        echo "Missing required fields: ${missing[*]}" >&2
        return 4
    fi

    # Validate type is exactly "gap-analysis"
    local type
    type=$(echo "$frontmatter" | grep "^type:" | sed 's/type: *//' | tr -d '"')
    if [[ "$type" != "gap-analysis" ]]; then
        echo "Invalid type: Must be 'gap-analysis'" >&2
        return 5
    fi

    # Validate complexity enum
    local complexity
    complexity=$(echo "$frontmatter" | grep "^complexity:" | sed 's/complexity: *//' | tr -d '"')
    if [[ ! "$complexity" =~ ^(PATCH|MODULE|SYSTEM|MIGRATION)$ ]]; then
        echo "Invalid complexity: Must be PATCH, MODULE, SYSTEM, or MIGRATION" >&2
        return 5
    fi

    # Validate status enum
    local status
    status=$(echo "$frontmatter" | grep "^status:" | sed 's/status: *//' | tr -d '"')
    if [[ ! "$status" =~ ^(in-progress|ready-for-design|archived)$ ]]; then
        echo "Invalid status: Must be in-progress, ready-for-design, or archived" >&2
        return 5
    fi

    return 0
}
```

## Handoff Criteria

When Gap Analysis phase completes, Pythia verifies:

- [ ] `type` is "gap-analysis"
- [ ] `status` is "ready-for-design"
- [ ] `affected_systems` lists all impacted systems
- [ ] `issue_count` matches actual issues in document
- [ ] All issues have ID, description, evidence
- [ ] At least one `success_criteria` defined
- [ ] Root cause chain is documented

## Relationship to Other Artifacts

```
GAP-{slug}.md
    |
    +-- Produces Context Design (context-design.gap_analysis)
    |
    +-- Success criteria verified by Compatibility Report
    |
    +-- May reference other Gap Analyses (dependencies)
```

## Document Structure Requirements

Beyond frontmatter, Gap Analysis documents MUST include:

1. **Executive Summary** - 2-3 sentence overview
2. **Root Cause Chain** - Numbered list showing causation
3. **Issue Inventory** - Categorized issues with ID, description, evidence
4. **Prioritized Fix List** - Table with priority, effort, impact
5. **Success Criteria** - Checkboxes matching frontmatter criteria

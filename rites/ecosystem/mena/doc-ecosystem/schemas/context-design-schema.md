---
schema_name: context-design
schema_version: "1.0"
file_pattern: "docs/ecosystem/CONTEXT-DESIGN-*.md"
artifact_type: context-design
---

# Context Design Schema

> Canonical schema for Context Design documents at `docs/ecosystem/CONTEXT-DESIGN-{slug}.md`

## YAML Frontmatter

```yaml
---
# Required fields
title: string              # Human-readable title
type: string               # Must be "context-design"
complexity: enum           # PATCH | MODULE | SYSTEM | MIGRATION
created_at: string         # ISO 8601 timestamp
status: enum               # in-progress | ready-for-implementation | archived

# Traceability
gap_analysis: string       # Reference to source Gap Analysis file

# Affected scope
affected_systems:          # Systems impacted by this design
  - enum                   # roster | CEM

# Work packages
work_packages:             # At least one required
  - id: string             # WP1, WP2, etc.
    name: string           # Work package name
    description: string    # What this WP delivers
    files: array           # Files to create/modify

# Optional fields
author: string             # Agent or user who created
backward_compatible: boolean  # Whether changes are backward compatible
migration_required: boolean   # Whether migration path is needed

# Schema versioning
schema_version: "1.0"      # Must be "1.0" for this version
---
```

## Required Fields

| Field | Type | Description | Set By |
|-------|------|-------------|--------|
| `title` | string | Human-readable title | context-architect |
| `type` | string | Must be "context-design" | context-architect |
| `complexity` | enum | Scope classification | context-architect |
| `created_at` | string | ISO 8601 creation timestamp | context-architect |
| `status` | enum | Current design status | context-architect |
| `gap_analysis` | string | Source Gap Analysis reference | context-architect |
| `affected_systems` | array | roster, CEM | context-architect |
| `work_packages` | array | WP definitions (min 1) | context-architect |
| `schema_version` | string | Schema version for compatibility | context-architect |

## Optional Fields

| Field | Type | Description | Set By |
|-------|------|-------------|--------|
| `author` | string | Creating agent or user | context-architect |
| `backward_compatible` | boolean | Is this backward compatible? | context-architect |
| `migration_required` | boolean | Is migration path needed? | context-architect |

## Work Package Object Schema

```yaml
work_packages:
  - id: string             # "WP1", "WP2", etc.
    name: string           # Descriptive name
    description: string    # What this WP delivers
    files:                 # Files affected
      - path: string       # File path
        action: enum       # create | modify | delete
        description: string  # What changes
    dependencies: array    # (optional) WP IDs this depends on
    estimated_effort: string  # (optional) Time estimate
```

## Valid Status Transitions

```
in-progress --complete--> ready-for-implementation
ready-for-implementation --implement--> archived
```

## Validation Rules

### Structure Validation
1. File MUST start with `---` on line 1
2. File MUST have closing `---` within first 50 lines
3. Content between delimiters MUST be valid YAML

### Field Validation
1. `type` MUST be exactly "context-design"
2. `complexity` MUST be one of: PATCH, MODULE, SYSTEM, MIGRATION
3. `created_at` MUST be valid ISO 8601 timestamp
4. `status` MUST be one of: in-progress, ready-for-implementation, archived
5. `gap_analysis` MUST reference existing Gap Analysis file
6. `affected_systems` MUST be array with at least one of: roster, CEM
7. `work_packages` MUST be array with at least one item
8. `schema_version` MUST be "1.0"

### Work Package Validation
1. Each WP MUST have `id`, `name`, `description`, `files`
2. `id` MUST match pattern `^WP[0-9]+$`
3. `files` MUST be array (may be empty for design-only WPs)
4. Each file entry MUST have `path`, `action`
5. `action` MUST be one of: create, modify, delete

### Design Decision Validation
1. Context Design MUST NOT contain "TBD", "TODO", or "maybe"
2. All design decisions MUST have documented rationale
3. If `backward_compatible: false`, `migration_required` MUST be true

## Example: Valid Context Design

```yaml
---
title: "Context Design: Session Validation Centralization"
type: context-design
complexity: MODULE
created_at: "2025-12-29T10:00:00Z"
status: ready-for-implementation
gap_analysis: GAP-session-validation.md
affected_systems:
  - roster
  - CEM
author: context-architect
backward_compatible: true
migration_required: false
work_packages:
  - id: WP1
    name: "Centralized Validation Function"
    description: "Create single validate_session_context() in session-state.sh"
    files:
      - path: "hooks/lib/session-state.sh"
        action: modify
        description: "Add validate_session_context() function"
      - path: "mena/session/common/session-context-schema.md"
        action: create
        description: "Document validation schema"
    estimated_effort: "2 hours"
  - id: WP2
    name: "Hook Refactoring"
    description: "Update all hooks to use centralized validation"
    files:
      - path: ".claude/hooks/session-context.sh"
        action: modify
        description: "Replace inline validation with function call"
      - path: ".claude/hooks/start-preflight.sh"
        action: modify
        description: "Replace inline validation with function call"
    dependencies: [WP1]
    estimated_effort: "1 hour"
schema_version: "1.0"
---

## Executive Summary

This Context Design addresses the 5 issues identified in GAP-session-validation.md by centralizing session validation into a single function with documented error codes.

## Design Decisions

### Decision 1: Validation Function Location

**Options Considered**:
1. New file `session-validator.sh` - Rejected: adds file proliferation
2. Existing `session-state.sh` - Selected: already handles session queries
3. Existing `primitives.sh` - Rejected: too low-level

**Selected**: session-state.sh

**Rationale**: session-state.sh already contains session query functions and is sourced by all hooks that need validation. Adding validation here maintains cohesion.

### Decision 2: Error Code Pattern

**Pattern**: Follow existing session-state.sh convention
- 0 = success
- 1 = file not found
- 2 = missing opening delimiter
- 3 = missing closing delimiter
- 4 = missing required field
- 5 = field validation failed

**Rationale**: Consistent with validate_session_context() already in session-state.sh.

## Work Package Details

### WP1: Centralized Validation Function

**Objective**: Single source of truth for session validation

**Implementation**:
```bash
validate_session_context() {
    local file="$1"
    local required_fields=("session_id" "created_at" "initiative" ...)
    # ... implementation
}
```

**Files Changed**:
| File | Change |
|------|--------|
| session-state.sh | Add function (lines 144-188) |
| session-context-schema.md | Create schema documentation |

### WP2: Hook Refactoring

**Objective**: All hooks use centralized validation

**Files Changed**:
| File | Before | After |
|------|--------|-------|
| session-context.sh | Inline validation (45 lines) | Function call (3 lines) |
| start-preflight.sh | Inline validation (30 lines) | Function call (3 lines) |

## Backward Compatibility

**Classification**: COMPATIBLE

Changes are internal refactoring:
- External hook behavior unchanged
- Error codes match existing convention
- No satellite impact

## Test Matrix

| Scenario | Expected Outcome |
|----------|------------------|
| Valid SESSION_CONTEXT.md | Return 0 |
| Missing file | Return 1 |
| Missing --- opener | Return 2 |
| Missing session_id | Return 4 with field name |

## Handoff Criteria

- [x] All design decisions have rationale
- [x] No TBD or unresolved items
- [x] Work packages specify file-level changes
- [x] Backward compatibility assessed
- [x] Test matrix defined
```

## Validation Function

```bash
# In ecosystem-validator.sh
# Usage: validate_context_design "/path/to/CONTEXT-DESIGN-example.md"
# Returns: 0=valid, 1=not found, 2=no opener, 3=no closer, 4=missing field, 5=invalid field

validate_context_design() {
    local file="$1"
    local required_fields=("title" "type" "complexity" "created_at" "status" "gap_analysis" "affected_systems" "work_packages" "schema_version")

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

    # Validate type is exactly "context-design"
    local type
    type=$(echo "$frontmatter" | grep "^type:" | sed 's/type: *//' | tr -d '"')
    if [[ "$type" != "context-design" ]]; then
        echo "Invalid type: Must be 'context-design'" >&2
        return 5
    fi

    # Validate status enum
    local status
    status=$(echo "$frontmatter" | grep "^status:" | sed 's/status: *//' | tr -d '"')
    if [[ ! "$status" =~ ^(in-progress|ready-for-implementation|archived)$ ]]; then
        echo "Invalid status: Must be in-progress, ready-for-implementation, or archived" >&2
        return 5
    fi

    return 0
}
```

## Handoff Criteria

When Context Design phase completes, orchestrator verifies:

- [ ] `type` is "context-design"
- [ ] `status` is "ready-for-implementation"
- [ ] `gap_analysis` references existing Gap Analysis
- [ ] All work packages have file-level change details
- [ ] No "TBD", "TODO", or "maybe" in document
- [ ] All design decisions have documented rationale
- [ ] Backward compatibility is assessed
- [ ] If breaking change, migration path is documented

## Relationship to Other Artifacts

```
CONTEXT-DESIGN-{slug}.md
    |
    +-- References Gap Analysis (gap_analysis field)
    |
    +-- Produces Implementation (integration-engineer consumes)
    |
    +-- Produces Migration Runbook (if migration_required)
    |
    +-- Verified by Compatibility Report
```

## Document Structure Requirements

Beyond frontmatter, Context Design documents MUST include:

1. **Executive Summary** - 2-3 sentence overview linking to Gap Analysis
2. **Design Decisions** - Each decision with options and rationale
3. **Work Package Details** - Implementation specifics per WP
4. **Backward Compatibility** - Classification and impact analysis
5. **Test Matrix** - Expected outcomes for validation
6. **Handoff Criteria** - Checklist for completion verification

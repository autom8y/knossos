---
schema_name: compatibility-report
schema_version: "1.0"
file_pattern: "docs/ecosystem/COMPATIBILITY-REPORT-*.md"
artifact_type: compatibility-report
---

# Compatibility Report Schema

> Canonical schema for Compatibility Reports at `docs/ecosystem/COMPATIBILITY-REPORT-{slug}.md`

## YAML Frontmatter

```yaml
---
# Required fields
title: string              # Human-readable title
type: string               # Must be "compatibility-report"
created_at: string         # ISO 8601 timestamp
tested_at: string          # When tests were executed (ISO 8601)
author: string             # Agent or user who created

# Overall status
overall_status: enum       # pass | fail | partial

# Test summary
total_tests: integer       # Total test configurations
passed_tests: integer      # Configurations that passed
failed_tests: integer      # Configurations that failed
skipped_tests: integer     # Configurations skipped

# Test matrix
test_matrix:               # Satellite configurations tested
  - satellite: string      # Satellite name or type
    configuration: string  # Configuration description
    status: enum           # pass | fail | skipped
    notes: string          # (optional) Additional context

# Traceability
context_design: string     # Reference to source Context Design
migration_runbook: string  # (optional) Reference to Migration Runbook

# Optional failure details
failures: array            # Details of any failures

# Schema versioning
schema_version: "1.0"      # Must be "1.0" for this version
---
```

## Required Fields

| Field | Type | Description | Set By |
|-------|------|-------------|--------|
| `title` | string | Human-readable title | compatibility-tester |
| `type` | string | Must be "compatibility-report" | compatibility-tester |
| `created_at` | string | ISO 8601 creation timestamp | compatibility-tester |
| `tested_at` | string | When tests ran (ISO 8601) | compatibility-tester |
| `author` | string | Creating agent or user | compatibility-tester |
| `overall_status` | enum | Aggregate pass/fail/partial | compatibility-tester |
| `total_tests` | integer | Total configurations tested | compatibility-tester |
| `passed_tests` | integer | Configurations passed | compatibility-tester |
| `failed_tests` | integer | Configurations failed | compatibility-tester |
| `skipped_tests` | integer | Configurations skipped | compatibility-tester |
| `test_matrix` | array | Test results per satellite | compatibility-tester |
| `context_design` | string | Source Context Design | compatibility-tester |
| `schema_version` | string | Schema version | compatibility-tester |

## Optional Fields

| Field | Type | Description | Set By |
|-------|------|-------------|--------|
| `migration_runbook` | string | Migration Runbook reference | compatibility-tester |
| `failures` | array | Detailed failure information | compatibility-tester |

## Test Matrix Object Schema

```yaml
test_matrix:
  - satellite: string      # "knossos", "minimal-satellite", "heavy-custom"
    configuration: string  # "default", "custom-hooks", "worktree"
    status: enum           # pass | fail | skipped
    notes: string          # (optional) Additional context
    tests_run: integer     # (optional) Number of tests for this config
    tests_passed: integer  # (optional) Tests passed
    duration: string       # (optional) Test duration
```

## Failure Object Schema

```yaml
failures:
  - satellite: string      # Which satellite failed
    configuration: string  # Configuration that failed
    test_name: string      # Specific test that failed
    error_message: string  # Error output
    expected: string       # What was expected
    actual: string         # What actually happened
    severity: enum         # blocking | degraded | cosmetic
    resolution: string     # (optional) How to fix
```

## Validation Rules

### Structure Validation
1. File MUST start with `---` on line 1
2. File MUST have closing `---` within first 50 lines
3. Content between delimiters MUST be valid YAML

### Field Validation
1. `type` MUST be exactly "compatibility-report"
2. `created_at` MUST be valid ISO 8601 timestamp
3. `tested_at` MUST be valid ISO 8601 timestamp
4. `overall_status` MUST be one of: pass, fail, partial
5. `total_tests` MUST be non-negative integer
6. `passed_tests` + `failed_tests` + `skipped_tests` MUST equal `total_tests`
7. `test_matrix` MUST be array with at least one item
8. `schema_version` MUST be "1.0"

### Status Consistency Validation
1. If `overall_status` is "pass", `failed_tests` MUST be 0
2. If `overall_status` is "fail", `failed_tests` MUST be > 0
3. If `overall_status` is "partial", some tests passed and some failed
4. Each `test_matrix` entry MUST have `satellite`, `configuration`, `status`

### Failure Details Validation
1. If `failed_tests` > 0, `failures` array SHOULD be present
2. Each failure MUST have `satellite`, `test_name`, `error_message`

## Example: Valid Compatibility Report

```yaml
---
title: "Compatibility Report: Session Schema v2.0"
type: compatibility-report
created_at: "2025-12-29T16:00:00Z"
tested_at: "2025-12-29T15:30:00Z"
author: compatibility-tester
overall_status: partial
total_tests: 4
passed_tests: 3
failed_tests: 1
skipped_tests: 0
test_matrix:
  - satellite: knossos
    configuration: canonical
    status: pass
    tests_run: 12
    tests_passed: 12
    duration: "45s"
    notes: "All core validation tests passed"
  - satellite: minimal-satellite
    configuration: no-overrides
    status: pass
    tests_run: 8
    tests_passed: 8
    duration: "30s"
    notes: "Schema inheritance verified"
  - satellite: heavy-customization
    configuration: custom-hooks
    status: fail
    tests_run: 10
    tests_passed: 7
    duration: "1m 15s"
    notes: "Custom session hook incompatible"
  - satellite: worktree-satellite
    configuration: worktree
    status: pass
    tests_run: 6
    tests_passed: 6
    duration: "20s"
failures:
  - satellite: heavy-customization
    configuration: custom-hooks
    test_name: "session-start-validation"
    error_message: "Custom hook expects legacy field 'session_type'"
    expected: "Hook accepts new schema without errors"
    actual: "Hook fails with 'missing required field: session_type'"
    severity: blocking
    resolution: "Update custom hook to use 'complexity' field instead of 'session_type'"
context_design: CONTEXT-DESIGN-session-schema-v2.md
migration_runbook: MIGRATION-RUNBOOK-session-schema-v2.md
schema_version: "1.0"
---

## Summary

Compatibility testing for session schema v2.0 migration completed with 3/4 configurations passing.

## Test Environment

| Satellite | Type | Configuration |
|-----------|------|---------------|
| knossos | Canonical | Default (no overrides) |
| minimal-satellite | Minimal | No hooks, no agents |
| heavy-customization | Custom | Custom hooks, custom agents |
| worktree-satellite | Worktree | Created via /worktree |

## Results by Satellite

### knossos (PASS)

All 12 tests passed:
- Session creation
- Session validation
- Session resume
- Session park
- Artifact tracking
- ...

### minimal-satellite (PASS)

All 8 tests passed:
- Schema inheritance from knossos
- Default hooks work
- No regression in basic functionality

### heavy-customization (FAIL)

7/10 tests passed. Failures:

| Test | Status | Issue |
|------|--------|-------|
| session-start-validation | FAIL | Custom hook expects legacy field |
| session-resume-validation | FAIL | Same root cause |
| custom-hook-integration | FAIL | Same root cause |

**Root Cause**: Custom `session-custom.sh` hook accesses `session_type` field which was renamed to `complexity` in v2.0.

**Resolution**: Update custom hook to use new field name.

### worktree-satellite (PASS)

All 6 tests passed:
- Worktree detection works
- Session isolation maintained
- No path conflicts

## Failure Analysis

### FAIL-001: Custom Hook Field Rename

**Affected**: heavy-customization satellite

**Description**: Custom hook `session-custom.sh` directly accesses `session_type` field which was renamed to `complexity` in v2.0.

**Severity**: Blocking - satellite cannot use new version

**Resolution**:
1. Update hook to read `complexity` instead of `complexity`
2. Or add backward compatibility shim in migration

**Recommendation**: Document in Migration Runbook as "satellites with custom hooks must update field references"

## Recommendations

1. **Proceed with rollout** to satellites without custom hooks
2. **Delay rollout** to heavy-customization until hook is updated
3. **Add to runbook**: Warning about custom hook field references

## Test Commands

Tests were executed using:

```bash
# Knossos (canonical baseline)
just test-session-schema --satellite=knossos

# Minimal
just test-session-schema --satellite=minimal-satellite

# Heavy custom
just test-session-schema --satellite=heavy-customization

# Worktree
just test-session-schema --satellite=worktree-satellite
```
```

## Validation Function

```bash
# In ecosystem-validator.sh
# Usage: validate_compatibility_report "/path/to/COMPATIBILITY-REPORT-example.md"
# Returns: 0=valid, 1=not found, 2=no opener, 3=no closer, 4=missing field, 5=invalid field

validate_compatibility_report() {
    local file="$1"
    local required_fields=("title" "type" "created_at" "tested_at" "author" "overall_status" "total_tests" "passed_tests" "failed_tests" "skipped_tests" "test_matrix" "context_design" "schema_version")

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

    # Validate type is exactly "compatibility-report"
    local type
    type=$(echo "$frontmatter" | grep "^type:" | sed 's/type: *//' | tr -d '"')
    if [[ "$type" != "compatibility-report" ]]; then
        echo "Invalid type: Must be 'compatibility-report'" >&2
        return 5
    fi

    # Validate overall_status enum
    local overall_status
    overall_status=$(echo "$frontmatter" | grep "^overall_status:" | sed 's/overall_status: *//' | tr -d '"')
    if [[ ! "$overall_status" =~ ^(pass|fail|partial)$ ]]; then
        echo "Invalid overall_status: Must be pass, fail, or partial" >&2
        return 5
    fi

    return 0
}
```

## Handoff Criteria

When Compatibility Report phase completes, orchestrator verifies:

- [ ] `type` is "compatibility-report"
- [ ] `context_design` references source Context Design
- [ ] `test_matrix` includes diverse satellite configurations
- [ ] Test counts are accurate and consistent
- [ ] If `overall_status` is fail/partial, `failures` array explains why
- [ ] Each failure has resolution path documented
- [ ] Recommendations section provides clear next steps

## Relationship to Other Artifacts

```
COMPATIBILITY-REPORT-{slug}.md
    |
    +-- References Context Design (context_design field)
    |
    +-- References Migration Runbook (migration_runbook field)
    |
    +-- Validates Gap Analysis success criteria
    |
    +-- Informs rollout decisions
```

## Satellite Diversity Requirements

Compatibility testing MUST include:

1. **Canonical** (knossos itself) - baseline validation
2. **Minimal** (no customizations) - inheritance verification
3. **Standard** (typical usage) - common configuration
4. **Heavy** (extensive customizations) - edge case coverage
5. **Worktree** (if applicable) - isolation verification

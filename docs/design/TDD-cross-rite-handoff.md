# TDD: Cross-Team Handoff Shared Skill

## Overview

This Technical Design Document specifies the cross-rite-handoff shared skill, a reusable schema and protocol for transferring work between rites. The skill defines the HANDOFF artifact structure, validation rules, and integration points with the session lifecycle.

## Context

| Reference | Location |
|-----------|----------|
| Handoff Criteria Schema | `/Users/tomtenuta/Code/roster/schemas/handoff-criteria-schema.yaml` |
| Within-Team Handoff | `/Users/tomtenuta/Code/roster/user-skills/session-lifecycle/handoff-ref/SKILL.md` |
| Cross-Team Skill (guidance) | `/Users/tomtenuta/Code/roster/user-skills/guidance/cross-rite/SKILL.md` |
| Cross-Team Playbook | `/Users/tomtenuta/Code/roster/docs/playbooks/cross-rite-coordination.md` |
| Cross-Team Edge Cases | `/Users/tomtenuta/Code/roster/docs/edge-cases/cross-rite-workflows.md` |
| Session Context Schema | `/Users/tomtenuta/Code/roster/user-skills/session-common/session-context-schema.md` |
| TDD Schema | `/Users/tomtenuta/Code/roster/user-skills/documentation/doc-artifacts/schemas/tdd-schema.md` |
| Context Design Schema | `/Users/tomtenuta/Code/roster/rites/ecosystem/skills/doc-ecosystem/schemas/context-design-schema.md` |
| Shared Skills README | `/Users/tomtenuta/Code/roster/rites/shared/README.md` |

### Problem Statement

Cross-team coordination in roster currently relies on:

1. **Implicit Patterns**: The playbook documents handoff types, but no formal schema exists
2. **No Validation**: Handoffs can be incomplete, missing required fields per type
3. **Scattered Documentation**: Edge cases, playbook, and protocol exist in different locations
4. **Type-Specific Requirements Hidden**: Each handoff type has unique requirements buried in prose
5. **No Integration with Session State**: SESSION_CONTEXT tracks artifacts but not pending cross-rite handoffs

### Design Goals

1. Single canonical schema for HANDOFF artifacts, validated like PRD/TDD/ADR
2. Type-specific required fields enforced at schema level
3. Clear trigger conditions for when cross-rite handoff is required vs optional
4. Session integration for tracking pending handoffs
5. Reusable by all rites via shared skill

---

## Schema Design

### HANDOFF Artifact Schema (v1.0)

**File Location**: `rites/shared/skills/cross-rite-handoff/schema.md`

**File Pattern**: `HANDOFF-{source}-to-{target}-{date}.md` or stored in session directory

```yaml
---
# ============================================================================
# HANDOFF Artifact Schema v1.0
# ============================================================================
# Machine-readable frontmatter for cross-rite work transfer
# All HANDOFF artifacts MUST conform to this schema

# Required: Core identification
artifact_id: string           # Pattern: HANDOFF-{source_rite}-to-{target_rite}-{YYYY-MM-DD}[-{seq}]
schema_version: "1.0"         # Must be "1.0" for this version

# Required: Team routing
source_rite: string           # Rite pack producing the handoff (e.g., "10x-dev")
target_rite: string           # Rite pack receiving the handoff (e.g., "security")

# Required: Handoff classification
handoff_type: enum            # execution | validation | assessment | implementation | strategic_input | strategic_evaluation
priority: enum                # critical | high | medium | low
blocking: boolean             # If true, source team cannot proceed until response

# Required: Context
initiative: string            # Initiative or feature name
created_at: string            # ISO 8601 timestamp
status: enum                  # pending | in_progress | completed | rejected

# Required: Work items (at least one)
items: array                  # See Item Object Schema below

# Optional: Source references
source_artifacts: array       # Paths to source artifacts
session_id: string            # Source session ID if within session
sprint_id: string             # Source sprint ID if within sprint

# Optional: Response tracking
response_due: string          # ISO 8601 deadline (derived from priority + SLA)
response_artifact: string     # Path to response artifact when completed

# Optional: Rejection handling
rejection_reason: string      # Why handoff was rejected (if status: rejected)
resubmission_of: string       # artifact_id of original if this is a resubmission
---
```

### Item Object Schema

Each item in the `items` array represents a discrete unit of work to hand off.

```yaml
items:
  - id: string                # Unique within handoff (e.g., "SEC-001", "PKG-001")
    summary: string           # 1-2 sentence description (required)
    priority: enum            # critical | high | medium | low (required)

    # Type-specific required fields (see Type-Specific Requirements)
    # At least one type-specific field must be present based on handoff_type

    acceptance_criteria: array     # Required for type: execution
    validation_scope: array        # Required for type: validation
    assessment_questions: array    # Required for type: assessment
    design_references: array       # Required for type: implementation
    data_sources: array            # Required for type: strategic_input
    confidence: enum               # Required for type: strategic_input (high | medium | low)
    evaluation_criteria: array     # Required for type: strategic_evaluation

    # Optional fields
    notes: string             # Additional context for target team
    dependencies: array       # IDs of other items this depends on
    estimated_effort: string  # Time estimate if known
```

### Type-Specific Required Fields

| Handoff Type | Required Per Item | Flow Pattern |
|--------------|-------------------|--------------|
| `execution` | `acceptance_criteria` (array, min 1) | Planning -> Execution |
| `validation` | `validation_scope` (array, min 1) | Dev -> Ops |
| `assessment` | `assessment_questions` (array, min 1) | Dev -> Specialist |
| `implementation` | `design_references` (array, min 1) | Research -> Dev |
| `strategic_input` | `data_sources` (array, min 1), `confidence` | Research -> Strategy |
| `strategic_evaluation` | `evaluation_criteria` (array, min 1) | R&D -> Strategy |

### Valid Status Transitions

```
pending --accept--> in_progress
pending --reject--> rejected
in_progress --complete--> completed
rejected --resubmit--> pending (new artifact, references original)
```

---

## Validation Rules

### Structure Validation

1. File MUST start with `---` on line 1
2. File MUST have closing `---` within first 75 lines
3. Content between delimiters MUST be valid YAML
4. File name MUST match pattern `HANDOFF-{source}-to-{target}-{date}[-seq].md`

### Field Validation

| Field | Rule | Error Code |
|-------|------|------------|
| `artifact_id` | Match pattern `^HANDOFF-[a-z0-9-]+-to-[a-z0-9-]+-[0-9]{4}-[0-9]{2}-[0-9]{2}(-[0-9]+)?$` | HANDOFF-001 |
| `source_rite` | Non-empty, match known rite pattern | HANDOFF-002 |
| `target_rite` | Non-empty, different from source_rite | HANDOFF-003 |
| `handoff_type` | One of: execution, validation, assessment, implementation, strategic_input, strategic_evaluation | HANDOFF-004 |
| `priority` | One of: critical, high, medium, low | HANDOFF-005 |
| `blocking` | Boolean | HANDOFF-006 |
| `initiative` | Non-empty string | HANDOFF-007 |
| `created_at` | Valid ISO 8601 timestamp | HANDOFF-008 |
| `status` | One of: pending, in_progress, completed, rejected | HANDOFF-009 |
| `items` | Array with at least one item | HANDOFF-010 |
| `schema_version` | Must be "1.0" | HANDOFF-011 |

### Item Validation

| Rule | Error Code |
|------|------------|
| Each item MUST have `id`, `summary`, `priority` | HANDOFF-020 |
| Item `id` MUST be unique within handoff | HANDOFF-021 |
| Item `priority` MUST be valid enum | HANDOFF-022 |
| Type-specific field MUST be present based on `handoff_type` | HANDOFF-023 |
| Type-specific field MUST be non-empty array | HANDOFF-024 |
| For `strategic_input`, `confidence` MUST be one of: high, medium, low | HANDOFF-025 |

### Cross-Field Validation

| Rule | Error Code |
|------|------------|
| `source_rite` and `target_rite` MUST be different | HANDOFF-030 |
| If `status: rejected`, `rejection_reason` MUST be present | HANDOFF-031 |
| If `resubmission_of` present, referenced artifact MUST exist | HANDOFF-032 |
| If `blocking: true`, `priority` SHOULD be `critical` or `high` | HANDOFF-033 (warning) |

---

## Trigger Conditions

### Required Cross-Team Handoff

A cross-rite HANDOFF artifact is **required** when:

| Condition | Target Rite | Handoff Type |
|-----------|-------------|--------------|
| Complexity >= SERVICE with security considerations | security | assessment |
| Feature involves production deployment | sre | validation |
| Debt remediation work ready for execution | hygiene | execution |
| R&D prototype ready for strategic evaluation | strategy | strategic_evaluation |
| User research synthesis complete | strategy | strategic_input |
| Strategic go-decision for production build | 10x-dev | implementation |

### Optional Cross-Team Handoff

A cross-rite HANDOFF artifact is **optional** when:

| Condition | Target Rite | Handoff Type |
|-----------|-------------|--------------|
| Feature complete, documentation update desired | doc-team-pack | assessment |
| Code review reveals hygiene concerns | hygiene | assessment |
| Performance concerns identified | sre | assessment |
| Minor security considerations (complexity < SERVICE) | security | assessment |

### No Handoff Required

Cross-team handoff is **not required** when:

- Work remains within single team's domain
- Phase transition within same team (use `/handoff` within-team)
- Quick consultation (use `/consult` or direct user communication)
- Information sharing without action items

---

## Session Integration

### SESSION_CONTEXT Schema Extension

Add optional `pending_handoffs` field to SESSION_CONTEXT:

```yaml
# In SESSION_CONTEXT.md frontmatter
pending_handoffs:
  - artifact_id: string      # Reference to HANDOFF artifact
    target_team: string      # Team waiting on
    blocking: boolean        # Is this blocking progress
    created_at: string       # When handoff was created
    status: enum             # pending | in_progress
```

### Handoff Lifecycle in Session

```
1. Source team creates HANDOFF artifact
2. Source team updates SESSION_CONTEXT.pending_handoffs (if session active)
3. User routes HANDOFF to target team
4. Target team accepts, rejects, or requests clarification
5. If accepted: Target team produces response artifact
6. If rejected: Source team updates and resubmits (new artifact)
7. On completion: Source team removes from pending_handoffs
```

### SESSION_CONTEXT Field Ownership

| Field | Set By | Modified By | Removed By |
|-------|--------|-------------|------------|
| `pending_handoffs` | HANDOFF creation | target team response | HANDOFF completion |

---

## File Locations and Naming

### Storage Locations

| Context | Location | Pattern |
|---------|----------|---------|
| Active session | `.claude/sessions/{session-id}/` | `HANDOFF-{source}-to-{target}-{date}.md` |
| Initiative docs | `docs/handoffs/` | `HANDOFF-{source}-to-{target}-{date}.md` |
| Sprint context | `.claude/sprints/{sprint-id}/` | `HANDOFF-{source}-to-{target}-{date}.md` |

### Naming Convention

```
HANDOFF-{source}-to-{target}-{YYYY-MM-DD}[-{seq}].md

Examples:
- HANDOFF-10x-dev-to-security-2026-01-03.md
- HANDOFF-debt-triage-to-hygiene-2026-01-03-2.md  (second handoff same day)
```

### Response Artifact Naming

Response artifacts follow similar pattern:
```
HANDOFF-RESPONSE-{target}-to-{source}-{YYYY-MM-DD}.md
```

---

## Example: Valid HANDOFF Artifact

```yaml
---
artifact_id: HANDOFF-10x-dev-to-security-2026-01-03
schema_version: "1.0"
source_rite: 10x-dev
target_rite: security
handoff_type: assessment
priority: critical
blocking: true
initiative: "Payment Processing Overhaul"
created_at: "2026-01-03T10:30:00Z"
status: pending
source_artifacts:
  - docs/requirements/PRD-payment-processing.md
  - docs/design/TDD-payment-processing.md
session_id: session-20260103-100000-abc123
items:
  - id: SEC-001
    summary: "Threat model for new payment token flow"
    priority: critical
    assessment_questions:
      - "What are the trust boundaries between client, API, and payment processor?"
      - "How are payment tokens secured in transit and at rest?"
      - "What is the attack surface for token interception?"
    notes: "PCI-DSS compliance required. See TDD section 4.2 for data flow diagram."
  - id: SEC-002
    summary: "Review API authentication changes"
    priority: high
    assessment_questions:
      - "Is the new OAuth2 implementation correct?"
      - "Are token refresh flows secure against replay attacks?"
    dependencies: ["SEC-001"]
---

## Context

The 10x-dev has completed PRD and TDD for a major payment processing overhaul. This feature handles credit card tokenization and requires security assessment before implementation can proceed.

### Why This Handoff

- Complexity: SERVICE
- Security considerations: Payment data, PCI-DSS compliance
- Blocking: Yes - cannot proceed to implementation without threat model

## Source Artifacts

| Artifact | Status | Notes |
|----------|--------|-------|
| PRD-payment-processing.md | Approved | Sections 3.2, 3.3 cover security requirements |
| TDD-payment-processing.md | In Review | Section 4.2 has data flow diagram |

## Notes for Security Rite

1. Priority SEC-001 before SEC-002 (dependency)
2. PCI-DSS Level 1 compliance required
3. Third-party payment processor: Stripe
4. Expected response timeline: 48 hours (critical priority)

## Acceptance Criteria for This Handoff

- [ ] Threat model document produced
- [ ] Trust boundaries identified and validated
- [ ] Specific mitigations recommended for identified threats
- [ ] Go/No-Go verdict provided
```

---

## Example: Rejected HANDOFF

```yaml
---
artifact_id: HANDOFF-10x-dev-to-security-2026-01-02
schema_version: "1.0"
source_team: 10x-dev
target_team: security
handoff_type: assessment
priority: critical
blocking: true
initiative: "Payment Processing Overhaul"
created_at: "2026-01-02T10:30:00Z"
status: rejected
rejection_reason: "Missing data flow diagram and trust boundary identification"
source_artifacts:
  - docs/requirements/PRD-payment-processing.md
items:
  - id: SEC-001
    summary: "Threat model for new payment token flow"
    priority: critical
    assessment_questions:
      - "Is the payment flow secure?"
---

## Rejection Details

### Reason

The handoff lacks sufficient context for security assessment:

1. **Missing data flow diagram**: Cannot identify attack surface without understanding data movement
2. **No trust boundaries**: Need explicit identification of client/server/third-party boundaries
3. **Assessment questions too vague**: "Is it secure?" is not answerable

### Required for Resubmission

1. Complete TDD with data flow diagram (see TDD template section 4)
2. Trust boundary diagram showing all system participants
3. Specific assessment questions per boundary

### Recommended Next Steps

Source team should:
1. Complete TDD-payment-processing.md with architecture diagrams
2. Create new HANDOFF referencing this rejection
3. Set `resubmission_of: HANDOFF-10x-dev-to-security-2026-01-02`
```

---

## Example: Execution Handoff

```yaml
---
artifact_id: HANDOFF-debt-triage-to-hygiene-2026-01-03
schema_version: "1.0"
source_team: debt-triage
target_team: hygiene
handoff_type: execution
priority: high
blocking: false
initiative: "Q1 2026 Debt Remediation"
created_at: "2026-01-03T14:00:00Z"
status: pending
source_artifacts:
  - docs/debt/SPRINT-PLAN-q1-2026.md
  - docs/debt/RISK-ASSESSMENT-email-validators.md
sprint_id: sprint-debt-q1-2026
items:
  - id: PKG-001
    summary: "Consolidate email validation logic across 4 services"
    priority: high
    acceptance_criteria:
      - "Single EmailValidator class in shared/validation/"
      - "All 4 services import from shared location"
      - "Behavior preserved: all existing tests pass"
      - "No new dependencies introduced"
    estimated_effort: "4 hours"
  - id: PKG-002
    summary: "Remove deprecated date parsing functions"
    priority: medium
    acceptance_criteria:
      - "All usages of parse_date_legacy() replaced with parse_date()"
      - "parse_date_legacy() function deleted"
      - "All date-related tests pass"
    dependencies: []
    estimated_effort: "2 hours"
---

## Context

The debt-triage has completed assessment and planning for Q1 2026 debt remediation. These packages are ready for execution by hygiene.

## Package Prioritization

Execute in order: PKG-001, PKG-002 (no dependencies between them, but PKG-001 is higher priority).

## Notes for Hygiene Rite

- PKG-001 affects: user-service, billing-service, notification-service, auth-service
- All changes must preserve existing behavior (see acceptance criteria)
- Report any behavior changes discovered during refactoring
```

---

## Validation Function

```bash
#!/bin/bash
# validate_handoff.sh
# Usage: validate_handoff "/path/to/HANDOFF-*.md"
# Returns: 0=valid, 1=not found, 2=no opener, 3=no closer, 4=missing field, 5=invalid field

validate_handoff() {
    local file="$1"
    local required_fields=("artifact_id" "schema_version" "source_team" "target_team"
                           "handoff_type" "priority" "blocking" "initiative"
                           "created_at" "status" "items")

    # Check file exists
    [ -f "$file" ] || { echo "HANDOFF-001: File not found: $file" >&2; return 1; }

    # Check opening delimiter on line 1
    local first_line
    first_line=$(head -n 1 "$file")
    if [[ "$first_line" != "---" ]]; then
        echo "HANDOFF-001: Missing opening '---' delimiter on line 1" >&2
        return 2
    fi

    # Check closing delimiter within first 75 lines
    local closing_line
    closing_line=$(head -n 75 "$file" | tail -n +2 | grep -n "^---$" | head -1 | cut -d: -f1)
    if [[ -z "$closing_line" ]]; then
        echo "HANDOFF-001: Missing closing '---' delimiter within first 75 lines" >&2
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
        echo "HANDOFF-010: Missing required fields: ${missing[*]}" >&2
        return 4
    fi

    # Validate artifact_id pattern
    local artifact_id
    artifact_id=$(echo "$frontmatter" | grep "^artifact_id:" | sed 's/artifact_id: *//' | tr -d '"')
    if [[ ! "$artifact_id" =~ ^HANDOFF-[a-z0-9-]+-to-[a-z0-9-]+-[0-9]{4}-[0-9]{2}-[0-9]{2}(-[0-9]+)?$ ]]; then
        echo "HANDOFF-001: Invalid artifact_id pattern: $artifact_id" >&2
        return 5
    fi

    # Validate source_rite != target_rite
    local source_rite target_rite
    source_rite=$(echo "$frontmatter" | grep "^source_rite:" | sed 's/source_rite: *//' | tr -d '"')
    target_rite=$(echo "$frontmatter" | grep "^target_rite:" | sed 's/target_rite: *//' | tr -d '"')
    if [[ "$source_rite" == "$target_rite" ]]; then
        echo "HANDOFF-030: source_rite and target_rite must be different" >&2
        return 5
    fi

    # Validate handoff_type enum
    local handoff_type
    handoff_type=$(echo "$frontmatter" | grep "^handoff_type:" | sed 's/handoff_type: *//' | tr -d '"')
    if [[ ! "$handoff_type" =~ ^(execution|validation|assessment|implementation|strategic_input|strategic_evaluation)$ ]]; then
        echo "HANDOFF-004: Invalid handoff_type: $handoff_type" >&2
        return 5
    fi

    # Validate priority enum
    local priority
    priority=$(echo "$frontmatter" | grep "^priority:" | sed 's/priority: *//' | tr -d '"')
    if [[ ! "$priority" =~ ^(critical|high|medium|low)$ ]]; then
        echo "HANDOFF-005: Invalid priority: $priority" >&2
        return 5
    fi

    # Validate status enum
    local status
    status=$(echo "$frontmatter" | grep "^status:" | sed 's/status: *//' | tr -d '"')
    if [[ ! "$status" =~ ^(pending|in_progress|completed|rejected)$ ]]; then
        echo "HANDOFF-009: Invalid status: $status" >&2
        return 5
    fi

    # Validate rejection requires reason
    if [[ "$status" == "rejected" ]]; then
        if ! echo "$frontmatter" | grep -q "^rejection_reason:"; then
            echo "HANDOFF-031: rejected status requires rejection_reason field" >&2
            return 5
        fi
    fi

    # TODO: Add item-level validation with type-specific field checks
    # This requires YAML parsing beyond simple grep

    return 0
}

# Export for use by other scripts
export -f validate_handoff
```

---

## Handoff Criteria Schema Integration

Add HANDOFF to the existing `schemas/handoff-criteria-schema.yaml`:

```yaml
  # In artifact_types section of handoff-criteria-schema.yaml

  handoff:
    phase: cross-rite
    description: "Cross-Team Handoff Artifact"
    criteria:
      - id: handoff-001
        description: "HANDOFF has valid artifact_id"
        validation: "artifact_id matches pattern ^HANDOFF-[a-z0-9-]+-to-[a-z0-9-]+-[0-9]{4}-[0-9]{2}-[0-9]{2}(-[0-9]+)?$"
        blocking: true
      - id: handoff-002
        description: "Source and target teams are different"
        validation: "source_team != target_team"
        blocking: true
      - id: handoff-003
        description: "Handoff type is valid enum"
        validation: "handoff_type in ['execution', 'validation', 'assessment', 'implementation', 'strategic_input', 'strategic_evaluation']"
        blocking: true
      - id: handoff-004
        description: "At least one item present"
        validation: "count(items) >= 1"
        blocking: true
      - id: handoff-005
        description: "Items have type-specific required fields"
        validation: "all items have required field for handoff_type"
        blocking: true
      - id: handoff-006
        description: "Status is valid enum"
        validation: "status in ['pending', 'in_progress', 'completed', 'rejected']"
        blocking: true
      - id: handoff-007
        description: "Rejected handoffs have rejection reason"
        validation: "status == 'rejected' implies rejection_reason is not null"
        blocking: true
```

---

## Skill Structure

### Directory Layout

```
rites/shared/skills/cross-rite-handoff/
+-- SKILL.md           # Main skill definition
+-- schema.md          # Full schema documentation (this content)
+-- validation.sh      # Validation functions
+-- examples/
    +-- assessment.md      # Example assessment handoff
    +-- execution.md       # Example execution handoff
    +-- rejected.md        # Example rejected handoff
```

### SKILL.md Content

```markdown
---
name: cross-rite-handoff
description: "HANDOFF artifact schema for cross-rite work transfer. Use when: work crosses team boundaries, specialist review required, formal handoff needed. Triggers: cross-rite, handoff artifact, team transfer, work handoff."
---

# Cross-Team Handoff Skill

> Defines the HANDOFF artifact schema for transferring work between rites.

## Quick Reference

**When to Use**: Work crosses team boundaries and requires formal handoff
**Artifact Pattern**: `HANDOFF-{source}-to-{target}-{date}.md`
**Handoff Types**: execution, validation, assessment, implementation, strategic_input, strategic_evaluation

## Decision Tree

```
Is work crossing team boundaries?
+-- No -> Use /handoff (within-team) or continue directly
+-- Yes -> Continue below

Is formal work transfer needed?
+-- No -> Use /consult or surface to user informally
+-- Yes -> Create HANDOFF artifact

What type of work?
+-- Ready for execution -> type: execution
+-- Needs validation (dev -> ops) -> type: validation
+-- Needs specialist review -> type: assessment
+-- Research -> production build -> type: implementation
+-- Data -> strategy -> type: strategic_input
+-- R&D -> go/no-go -> type: strategic_evaluation
```

## Handoff Types

| Type | Flow | Required Per Item |
|------|------|-------------------|
| `execution` | Planning -> Execution | `acceptance_criteria` |
| `validation` | Dev -> Ops | `validation_scope` |
| `assessment` | Dev -> Specialist | `assessment_questions` |
| `implementation` | Research -> Dev | `design_references` |
| `strategic_input` | Research -> Strategy | `data_sources`, `confidence` |
| `strategic_evaluation` | R&D -> Strategy | `evaluation_criteria` |

## Progressive Disclosure

- [schema.md](schema.md) - Full schema specification
- [validation.sh](validation.sh) - Validation functions
- [examples/](examples/) - Example handoffs by type
```

---

## Integration Points

### 1. Session State Tracking

When creating a HANDOFF:
1. Validate artifact against schema
2. If session active, add to `pending_handoffs` in SESSION_CONTEXT
3. Store artifact in appropriate location
4. Notify user for routing

### 2. Handoff Criteria Schema

The existing `schemas/handoff-criteria-schema.yaml` gains a new `handoff` artifact type with blocking criteria.

### 3. Orchestrator Awareness

Orchestrators should:
- Check for pending cross-rite handoffs when planning work
- Block on `blocking: true` handoffs
- Surface required handoffs based on trigger conditions

### 4. CEM Sync

The `rites/shared/skills/cross-rite-handoff/` directory is synced to all satellites via shared skill mechanism.

---

## Test Matrix

### Validation Tests

| Test ID | Input | Expected Outcome |
|---------|-------|------------------|
| `val_001` | Valid assessment handoff | Return 0 |
| `val_002` | Missing artifact_id | Return 4 with HANDOFF-010 |
| `val_003` | Invalid artifact_id pattern | Return 5 with HANDOFF-001 |
| `val_004` | source_team == target_team | Return 5 with HANDOFF-030 |
| `val_005` | Invalid handoff_type | Return 5 with HANDOFF-004 |
| `val_006` | status: rejected without reason | Return 5 with HANDOFF-031 |
| `val_007` | Empty items array | Return 4 with HANDOFF-010 |
| `val_008` | execution type missing acceptance_criteria | Return 5 with HANDOFF-023 |
| `val_009` | strategic_input missing confidence | Return 5 with HANDOFF-025 |

### Integration Tests

| Test ID | Scenario | Expected Outcome |
|---------|----------|------------------|
| `int_001` | Create handoff in active session | pending_handoffs updated |
| `int_002` | Complete handoff | pending_handoffs entry removed |
| `int_003` | Reject handoff | status updated, rejection_reason required |
| `int_004` | Resubmit handoff | references original via resubmission_of |
| `int_005` | Blocking handoff created | Source team workflow blocked |

---

## Backward Compatibility

### No Breaking Changes

This skill introduces new functionality without modifying existing behavior:

1. **Existing handoff-ref**: Unchanged - handles within-team agent handoffs
2. **Existing cross-rite skill**: Unchanged - remains routing guidance
3. **Existing playbook/edge-cases docs**: Unchanged - serve as operational guidance

### Coexistence

The HANDOFF artifact schema complements, not replaces:
- Within-team `/handoff` command (agent-to-agent within same session)
- Cross-team routing guidance (when to surface concerns)
- Edge case documentation (operational recovery procedures)

---

## Handoff Criteria for This Design

- [x] Schema fully specified with validation rules
- [x] All handoff types defined with required fields
- [x] Trigger conditions documented (required vs optional)
- [x] Session integration specified (pending_handoffs field)
- [x] Example HANDOFF artifacts included (valid, rejected, execution)
- [x] Validation function provided
- [x] Integration with handoff-criteria-schema.yaml specified
- [x] File locations and naming conventions defined
- [x] Test matrix covers validation and integration scenarios
- [x] No unresolved design decisions

---

## Artifact Attestation

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-cross-rite-handoff.md` | Created |
| Handoff Criteria Schema | `/Users/tomtenuta/Code/roster/schemas/handoff-criteria-schema.yaml` | Read |
| Within-Team Handoff | `/Users/tomtenuta/Code/roster/user-skills/session-lifecycle/handoff-ref/SKILL.md` | Read |
| Cross-Team Playbook | `/Users/tomtenuta/Code/roster/docs/playbooks/cross-rite-coordination.md` | Read |
| Cross-Team Edge Cases | `/Users/tomtenuta/Code/roster/docs/edge-cases/cross-rite-workflows.md` | Read |
| Session Context Schema | `/Users/tomtenuta/Code/roster/user-skills/session-common/session-context-schema.md` | Read |
| Context Design Schema | `/Users/tomtenuta/Code/roster/rites/ecosystem/skills/doc-ecosystem/schemas/context-design-schema.md` | Read |
| Shared Skills README | `/Users/tomtenuta/Code/roster/rites/shared/README.md` | Read |

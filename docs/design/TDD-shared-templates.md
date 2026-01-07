# TDD: Shared Templates Skill

## Overview

This Technical Design Document specifies the shared-templates skill, a multi-team infrastructure providing canonical document templates for debt-ledger, risk-matrix, and sprint-debt-package artifacts. The skill lives in `rites/shared/skills/shared-templates/` and is automatically available to all rites via the shared team sync mechanism.

## Context

| Reference | Location |
|-----------|----------|
| Shared Team | `rites/shared/README.md` |
| Debt Collector Agent | `rites/debt-triage/agents/debt-collector.md` |
| Risk Assessor Agent | `rites/debt-triage/agents/risk-assessor.md` |
| Sprint Planner Agent | `rites/debt-triage/agents/sprint-planner.md` |
| Doc-Ecosystem Skill | `.claude/skills/doc-ecosystem/SKILL.md` |
| Doc-Consolidation Templates | `rites/docs/skills/doc-consolidation/templates/` |

### Problem Statement

The debt-triage agents reference templates that do not exist:
- `debt-collector.md` references `@shared-templates#debt-ledger-template`
- `risk-assessor.md` references `@shared-templates#risk-matrix-template`
- `sprint-planner.md` references `@shared-templates#sprint-debt-packages-template`

Without canonical templates, agents produce inconsistent artifacts that:
1. Lack standardized structure for downstream consumption
2. Miss required fields that handoff partners depend on
3. Cannot be validated programmatically
4. Diverge across sessions creating confusion

### Design Goals

1. Define three canonical templates with clear schemas
2. Establish placeholder conventions consistent with existing ecosystem patterns
3. Enable team customization within defined boundaries
4. Support template versioning for backward compatibility
5. Provide validation rules for each template type

---

## Template Architecture

### Skill Structure

```
rites/shared/skills/shared-templates/
├── SKILL.md                          # Skill entry point
├── schemas/
│   ├── debt-ledger-schema.md         # Schema definition
│   ├── risk-matrix-schema.md         # Schema definition
│   └── sprint-debt-package-schema.md # Schema definition
├── templates/
│   ├── debt-ledger.md                # Template with placeholders
│   ├── risk-matrix.md                # Template with placeholders
│   └── sprint-debt-package.md        # Template with placeholders
└── validation/
    └── template-rules.md             # Validation logic
```

### Sync Behavior

Per `rites/shared/README.md`, shared skills are flattened into `.claude/skills/` during `swap-rite.sh` execution:
- `rites/shared/skills/shared-templates/` syncs to `.claude/skills/shared-templates/`
- Available regardless of active rite
- Team-specific overrides take precedence (team-privileged)

### Reference Syntax

Templates are referenced using anchor syntax:
- `@shared-templates#debt-ledger-template` resolves to `templates/debt-ledger.md`
- `@shared-templates#risk-matrix-template` resolves to `templates/risk-matrix.md`
- `@shared-templates#sprint-debt-packages-template` resolves to `templates/sprint-debt-package.md`

---

## Placeholder Conventions

### Syntax Standard

Adopt the placeholder pattern established in `doc-consolidation/templates/`:

| Syntax | Description | Example |
|--------|-------------|---------|
| `{placeholder}` | Required field, agent must replace | `{audit_date}` |
| `{placeholder:default}` | Optional with default | `{owner:unassigned}` |
| `[{placeholder}]` | Optional section, may be omitted | `[{related_items}]` |
| `{...}` | Repeatable block (1+) | Entry rows |
| `<!-- ... -->` | Template guidance comments | Instructions |

### Reserved Placeholders

These placeholders have semantic meaning across all templates:

| Placeholder | Type | Description |
|-------------|------|-------------|
| `{artifact_id}` | string | Unique identifier (e.g., `DL-001`, `RM-001`, `SDP-001`) |
| `{created_at}` | ISO 8601 | Creation timestamp |
| `{author}` | string | Creating agent or user |
| `{status}` | enum | Artifact lifecycle status |
| `{session_id}` | string | Associated session (optional) |
| `{initiative}` | string | Parent initiative name |

### Variable Substitution

Agents replace placeholders during artifact generation. Validation occurs post-replacement via schema rules.

**Replacement Rules:**
1. All `{required}` placeholders MUST be replaced
2. `{optional:default}` uses default if no value provided
3. `[{optional_section}]` is removed entirely if unused
4. Guidance comments (`<!-- -->`) are stripped in final artifact

---

## Template Schemas

### 1. Debt Ledger Schema

**Purpose:** Structured inventory of technical debt produced by Debt Collector agent.

**File Pattern:** `docs/debt/DL-{slug}.md`

**YAML Frontmatter:**

```yaml
---
# Required fields
artifact_id: string        # Pattern: DL-{slug}
title: string              # Human-readable title
type: string               # Must be "debt-ledger"
created_at: string         # ISO 8601 timestamp
author: string             # Creating agent (e.g., "debt-collector")
status: enum               # draft | final | archived
schema_version: "1.0"      # Schema version

# Audit scope
scope:
  directories: array       # Paths audited
  categories: array        # Categories included (code, doc, test, infra, design)
  exclusions: array        # Paths or patterns excluded

# Summary statistics
statistics:
  total_items: integer     # Total debt items found
  by_category:
    code: integer
    doc: integer
    test: integer
    infra: integer
    design: integer
  by_type: object          # Breakdown by specific type

# Optional fields
session_id: string         # Associated session
initiative: string         # Parent initiative
previous_ledger: string    # Reference to baseline for diff
---
```

**Required Sections:**

| Section | Purpose | Authored By |
|---------|---------|-------------|
| Executive Summary | 2-3 sentence overview | debt-collector |
| Audit Scope | What was audited, exclusions | debt-collector |
| Debt Inventory | Categorized items with ID, location, description | debt-collector |
| Summary Statistics | Counts by category and type | debt-collector |
| Audit Limitations | Known gaps or incomplete areas | debt-collector |

**Optional Sections:**

| Section | Purpose | When Included |
|---------|---------|---------------|
| Debt Diff | Comparison to previous ledger | When `previous_ledger` specified |
| Ownership Report | Items grouped by owner | When ownership data available |

**Debt Item Object Schema:**

```yaml
debt_items:
  - id: string             # "C042", "D007", etc.
    location: string       # file:line or module path
    category: enum         # code | doc | test | infra | design
    type: string           # Specific type (e.g., "hardcoded", "missing-doc")
    description: string    # What the debt is
    age: string            # How old (optional, from git blame)
    owner: string          # Responsible party (optional)
    related: array         # Related item IDs (optional)
    evidence: string       # Quote or reference (optional)
```

**Validation Rules:**

1. `artifact_id` MUST match pattern `^DL-[a-z0-9-]+$`
2. `type` MUST be exactly "debt-ledger"
3. `status` MUST be one of: draft, final, archived
4. `statistics.total_items` MUST equal sum of category counts
5. `scope.categories` MUST be non-empty array
6. Each item in Debt Inventory MUST have id, location, category, description

---

### 2. Risk Matrix Schema

**Purpose:** Scored and prioritized debt items produced by Risk Assessor agent.

**File Pattern:** `docs/debt/RM-{slug}.md`

**YAML Frontmatter:**

```yaml
---
# Required fields
artifact_id: string        # Pattern: RM-{slug}
title: string              # Human-readable title
type: string               # Must be "risk-matrix"
created_at: string         # ISO 8601 timestamp
author: string             # Creating agent (e.g., "risk-assessor")
status: enum               # draft | final | archived
schema_version: "1.0"      # Schema version

# Source reference
source_ledger: string      # Reference to input Debt Ledger (e.g., "DL-api-cleanup")

# Priority summary
priority_counts:
  critical: integer        # Composite score >= 8
  high: integer            # Composite score 5-7.9
  medium: integer          # Composite score 2-4.9
  low: integer             # Composite score < 2

# Quick wins (high value, low effort)
quick_wins_count: integer

# Optional fields
session_id: string         # Associated session
initiative: string         # Parent initiative
risk_tolerance: string     # Org risk tolerance context
---
```

**Required Sections:**

| Section | Purpose | Authored By |
|---------|---------|-------------|
| Executive Summary | Key findings and recommendations | risk-assessor |
| Scoring Methodology | Blast radius, likelihood, effort scales | risk-assessor |
| Risk Matrix | Scored items with composite priority | risk-assessor |
| Priority Breakdown | Items by critical/high/medium/low | risk-assessor |
| Quick Wins | High value, low effort items | risk-assessor |

**Optional Sections:**

| Section | Purpose | When Included |
|---------|---------|---------------|
| Executive Briefing | One-page leadership summary | For leadership handoff |
| Risk Clusters | Related items for batched remediation | When clusters identified |
| Assessment Assumptions | Context and limitations | Always recommended |

**Scored Item Object Schema:**

```yaml
scored_items:
  - id: string             # From source ledger
    source_id: string      # Original debt item ID
    blast_radius: integer  # 1-5 scale
    likelihood: integer    # 1-5 scale
    effort: integer        # 1-5 scale
    composite: float       # (blast * likelihood) / effort
    priority: enum         # critical | high | medium | low
    trigger: string        # What triggers this risk
    rationale: string      # Why these scores
```

**Scoring Formula:**

```
Composite = (Blast Radius * Likelihood) / Effort

Priority Tiers:
- Critical: >= 8
- High: 5.0 - 7.9
- Medium: 2.0 - 4.9
- Low: < 2.0
```

**Validation Rules:**

1. `artifact_id` MUST match pattern `^RM-[a-z0-9-]+$`
2. `type` MUST be exactly "risk-matrix"
3. `source_ledger` MUST reference existing Debt Ledger
4. `blast_radius`, `likelihood`, `effort` MUST be integers 1-5
5. `composite` MUST equal (blast_radius * likelihood) / effort
6. `priority` MUST match composite score tier
7. Sum of priority_counts MUST equal total scored items

---

### 3. Sprint Debt Package Schema

**Purpose:** Sprint-ready work units produced by Sprint Planner agent.

**File Pattern:** `docs/sprints/SDP-{slug}.md`

**YAML Frontmatter:**

```yaml
---
# Required fields
artifact_id: string        # Pattern: SDP-{slug}
title: string              # Human-readable title (e.g., "Sprint 24 Debt Package")
type: string               # Must be "sprint-debt-package"
created_at: string         # ISO 8601 timestamp
author: string             # Creating agent (e.g., "sprint-planner")
status: enum               # draft | ready | in-progress | complete
schema_version: "1.0"      # Schema version

# Source reference
source_matrix: string      # Reference to input Risk Matrix (e.g., "RM-api-cleanup")

# Capacity planning
capacity:
  total_hours: integer     # Available capacity in hours
  buffer_percent: integer  # Buffer percentage (default 20%)
  allocated_hours: integer # Hours allocated to packages

# Package summary
package_count: integer     # Number of work packages
total_effort_hours: integer # Sum of package estimates

# Sprint info
sprint:
  name: string             # Sprint name/number
  start_date: string       # ISO 8601 date
  end_date: string         # ISO 8601 date

# Optional fields
session_id: string         # Associated session
initiative: string         # Parent initiative
target_team: string        # Team receiving handoff (e.g., "hygiene")
---
```

**Required Sections:**

| Section | Purpose | Authored By |
|---------|---------|-------------|
| Executive Summary | Sprint goals and key packages | sprint-planner |
| Capacity Model | Available vs allocated hours | sprint-planner |
| Work Packages | Detailed package specifications | sprint-planner |
| Dependency Map | Package dependencies | sprint-planner |
| Acceptance Criteria Summary | Roll-up of all criteria | sprint-planner |

**Optional Sections:**

| Section | Purpose | When Included |
|---------|---------|---------------|
| Deferred Items | Items not included with rationale | When items deferred |
| HANDOFF | Cross-team handoff artifact | When target_team specified |
| Capacity Scenarios | What-if planning alternatives | For complex planning |

**Work Package Object Schema:**

```yaml
packages:
  - id: string             # "PKG-001", "PKG-002", etc.
    title: string          # Human-readable title
    source_items: array    # Risk matrix item IDs included
    size: enum             # XS | S | M | L | XL
    effort_hours: integer  # Estimated hours
    confidence: enum       # high | medium | low
    priority: enum         # critical | high | medium | low
    sprint: string         # Target sprint (this, next, backlog)
    dependencies: array    # Other package IDs this depends on
    acceptance_criteria:   # Specific, testable criteria
      - string
    owner: string          # Assigned team/person (optional)
    notes: string          # Additional context (optional)
```

**Size Guidelines:**

| Size | Hours | Points | Scope |
|------|-------|--------|-------|
| XS | 1-2 | 1 | Config change, small fix |
| S | 2-4 | 2 | Single file, straightforward |
| M | 4-8 | 3-5 | Multiple files, contained |
| L | 8-16 | 5-8 | Cross-module, needs design |
| XL | 16-32 | 8-13 | Significant refactor |

**Confidence Adjustments:**

| Confidence | Buffer Multiplier | Description |
|------------|-------------------|-------------|
| high | 1.0x | Similar work done before, clear scope |
| medium | 1.25-1.5x | Some unknowns |
| low | 1.5-2.0x | Significant unknowns, may need spike |

**Validation Rules:**

1. `artifact_id` MUST match pattern `^SDP-[a-z0-9-]+$`
2. `type` MUST be exactly "sprint-debt-package"
3. `source_matrix` MUST reference existing Risk Matrix
4. `capacity.allocated_hours` MUST NOT exceed `capacity.total_hours`
5. `total_effort_hours` MUST equal sum of package effort_hours
6. Each package MUST have at least one acceptance criterion
7. `size` MUST be one of: XS, S, M, L, XL
8. Packages larger than XL MUST be split or flagged for spike

---

## Template Versioning Strategy

### Version Scheme

Templates follow semantic versioning: `MAJOR.MINOR`

| Version Change | When | Example |
|----------------|------|---------|
| MAJOR (1.x -> 2.0) | Breaking schema change | Required field added |
| MINOR (1.0 -> 1.1) | Additive change | Optional field added |

### Backward Compatibility Rules

1. **MINOR versions** MUST be backward compatible
   - New optional fields have defaults
   - Existing fields unchanged
   - Validation accepts older artifacts

2. **MAJOR versions** require migration
   - Migration runbook provided
   - Old artifacts remain readable
   - New artifacts use new schema

### Schema Version in Frontmatter

All artifacts include `schema_version` field:

```yaml
schema_version: "1.0"
```

### Version Detection

Validation functions check schema version and apply appropriate rules:

```bash
validate_debt_ledger() {
    local version
    version=$(extract_frontmatter_field "schema_version" "$file")
    case "$version" in
        1.0) validate_v1_0 "$file" ;;
        1.1) validate_v1_1 "$file" ;;
        *)   echo "Unknown schema version: $version" >&2; return 1 ;;
    esac
}
```

---

## Team Customization Boundaries

### Fixed (Skeleton-Owned)

These elements are controlled by shared-templates and MUST NOT be customized:

| Element | Rationale |
|---------|-----------|
| Frontmatter schema | Enables cross-rite interoperability |
| Required sections | Ensures handoff completeness |
| Placeholder syntax | Consistent parsing |
| Validation rules | Predictable quality |
| Artifact ID patterns | Unique identification |

### Customizable (Team-Owned)

Teams MAY customize within these boundaries:

| Element | Boundary | Example |
|---------|----------|---------|
| Optional sections | Add team-specific sections | "Security Review" section |
| Additional placeholders | Add to optional sections | `{team_specific_field}` |
| Guidance comments | Modify instructions | Team-specific workflow tips |
| Default values | Override optional defaults | Different buffer percentage |

### Override Mechanism

Team-specific templates override shared templates when:
1. Team pack includes `skills/shared-templates/templates/{template}.md`
2. Team version syncs over shared version during `swap-rite.sh`

**Example:** `rites/security/skills/shared-templates/templates/risk-matrix.md` would override the shared version with security-specific additions.

---

## Integration Test Matrix

### Test Coverage

| Artifact | Satellite Type | Test | Expected Outcome |
|----------|----------------|------|------------------|
| Debt Ledger | skeleton | Generate from codebase | Valid DL artifact |
| Debt Ledger | minimal | Empty audit | Valid with 0 items |
| Debt Ledger | complex | Large codebase | All categories populated |
| Risk Matrix | skeleton | Score from DL | Valid RM artifact |
| Risk Matrix | minimal | Score empty DL | Valid with 0 items |
| Risk Matrix | complex | Score large DL | All priorities calculated |
| Sprint Package | skeleton | Package from RM | Valid SDP artifact |
| Sprint Package | minimal | Package empty RM | Valid with 0 packages |
| Sprint Package | complex | Multi-sprint planning | Dependencies resolved |

### Validation Test Cases

| Test ID | Template | Condition | Expected |
|---------|----------|-----------|----------|
| `DL-V001` | debt-ledger | Missing artifact_id | Fail: required field |
| `DL-V002` | debt-ledger | Invalid category | Fail: enum violation |
| `DL-V003` | debt-ledger | Stats mismatch | Fail: sum validation |
| `RM-V001` | risk-matrix | Missing source_ledger | Fail: required field |
| `RM-V002` | risk-matrix | Score out of range | Fail: 1-5 range |
| `RM-V003` | risk-matrix | Priority mismatch | Fail: tier calculation |
| `SDP-V001` | sprint-package | Missing source_matrix | Fail: required field |
| `SDP-V002` | sprint-package | Over capacity | Fail: capacity constraint |
| `SDP-V003` | sprint-package | No acceptance criteria | Fail: required array |

### Handoff Validation

| Handoff | Source | Target | Validation |
|---------|--------|--------|------------|
| DL -> RM | debt-collector | risk-assessor | All DL items scoreable |
| RM -> SDP | risk-assessor | sprint-planner | All RM items packageable |
| SDP -> HANDOFF | sprint-planner | hygiene | HANDOFF schema valid |

---

## SKILL.md Entry Point

```markdown
---
name: shared-templates
description: "Multi-team document templates for debt-ledger, risk-matrix, and sprint-debt-package artifacts. Use when: creating debt triage artifacts, producing consistent handoff documents, or validating artifact structure. Triggers: debt ledger, risk matrix, sprint package, debt template, triage template."
---

# Shared Templates

> Canonical templates for debt triage workflow artifacts.

## Templates

| Template | Anchor | Agent | Purpose |
|----------|--------|-------|---------|
| [Debt Ledger](templates/debt-ledger.md) | `#debt-ledger-template` | debt-collector | Technical debt inventory |
| [Risk Matrix](templates/risk-matrix.md) | `#risk-matrix-template` | risk-assessor | Scored and prioritized debt |
| [Sprint Package](templates/sprint-debt-package.md) | `#sprint-debt-packages-template` | sprint-planner | Sprint-ready work units |

## Usage

Reference templates in agent prompts:

```markdown
Produce debt ledgers using `@shared-templates#debt-ledger-template`.
```

## Schemas

Full schema definitions with validation rules:

- [Debt Ledger Schema](schemas/debt-ledger-schema.md)
- [Risk Matrix Schema](schemas/risk-matrix-schema.md)
- [Sprint Debt Package Schema](schemas/sprint-debt-package-schema.md)

## Placeholder Conventions

| Syntax | Meaning |
|--------|---------|
| `{field}` | Required, must replace |
| `{field:default}` | Optional with default |
| `[{section}]` | Optional section |
| `<!-- ... -->` | Guidance (stripped) |

## Versioning

Current: `1.0`

Templates follow semantic versioning. MINOR versions are backward compatible.

## Related

- `@documentation` - Core PRD/TDD/ADR templates
- `@doc-ecosystem` - Ecosystem change templates
- `@cross-rite-handoff` - HANDOFF artifact schema
```

---

## Implementation Phases

### Phase 1: Skill Structure

Create directory structure and SKILL.md entry point.

**Files Created:**
- `rites/shared/skills/shared-templates/SKILL.md`

### Phase 2: Schemas

Define formal schemas for each template type.

**Files Created:**
- `rites/shared/skills/shared-templates/schemas/debt-ledger-schema.md`
- `rites/shared/skills/shared-templates/schemas/risk-matrix-schema.md`
- `rites/shared/skills/shared-templates/schemas/sprint-debt-package-schema.md`

### Phase 3: Templates

Create template files with placeholder conventions.

**Files Created:**
- `rites/shared/skills/shared-templates/templates/debt-ledger.md`
- `rites/shared/skills/shared-templates/templates/risk-matrix.md`
- `rites/shared/skills/shared-templates/templates/sprint-debt-package.md`

### Phase 4: Validation Rules

Document validation logic for artifact verification.

**Files Created:**
- `rites/shared/skills/shared-templates/validation/template-rules.md`

### Phase 5: Integration

Update agent references to use new templates.

**Files Modified:**
- `rites/debt-triage/agents/debt-collector.md` (verify reference works)
- `rites/debt-triage/agents/risk-assessor.md` (verify reference works)
- `rites/debt-triage/agents/sprint-planner.md` (verify reference works)

---

## Migration Path

### Existing Artifacts

No migration required. This skill creates new infrastructure; existing artifacts are unaffected.

### Agent Updates

Agent prompts already reference `@shared-templates#*`. Once skill is created, references resolve automatically.

---

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Template drift between teams | Medium | Medium | Fixed schema elements, validation rules |
| Placeholder syntax confusion | Low | Low | Clear documentation, examples |
| Over-customization breaks handoffs | Medium | High | Customization boundaries documented |
| Schema version mismatch | Low | Medium | Version detection, backward compatibility |

---

## Success Criteria

- [ ] All three template schemas fully specified with required/optional sections
- [ ] Placeholder syntax documented with examples
- [ ] Versioning strategy documented with backward compatibility rules
- [ ] Team customization boundaries clearly defined
- [ ] Agent references resolve to templates (`@shared-templates#*`)
- [ ] Validation rules defined for each template type
- [ ] Templates integrate into swap-rite.sh sync flow

---

## Handoff Criteria

Integration Engineer receives:

- [ ] Complete TDD with all schemas documented
- [ ] Placeholder convention standard defined
- [ ] Template versioning strategy specified
- [ ] Team customization boundaries documented
- [ ] Skill structure and file layout specified
- [ ] No unresolved design decisions

---

## Artifact Attestation

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-shared-templates.md` | Created |
| Shared Team README | `/Users/tomtenuta/Code/roster/rites/shared/README.md` | Read |
| Debt Collector | `/Users/tomtenuta/Code/roster/rites/debt-triage/agents/debt-collector.md` | Read |
| Risk Assessor | `/Users/tomtenuta/Code/roster/rites/debt-triage/agents/risk-assessor.md` | Read |
| Sprint Planner | `/Users/tomtenuta/Code/roster/rites/debt-triage/agents/sprint-planner.md` | Read |
| Doc-Ecosystem SKILL | `/Users/tomtenuta/Code/roster/.claude/skills/doc-ecosystem/SKILL.md` | Read |
| TDD Schema Reference | `/Users/tomtenuta/Code/roster/user-skills/documentation/doc-artifacts/schemas/tdd-schema.md` | Read |
| Gap Analysis Schema | `/Users/tomtenuta/Code/roster/.claude/skills/doc-ecosystem/schemas/gap-analysis-schema.md` | Read |

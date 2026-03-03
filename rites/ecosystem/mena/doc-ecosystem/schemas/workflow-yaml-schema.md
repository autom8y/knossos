---
schema_name: workflow-yaml
schema_version: "1.0"
file_pattern: "rites/*/workflow.yaml"
artifact_type: workflow
---

# workflow.yaml Schema

> Canonical schema for rite workflow definitions at `rites/{rite-name}/workflow.yaml`

## YAML Structure

```yaml
# Required fields
name: string               # Rite name (e.g., "10x-dev")
workflow_type: enum        # sequential | parallel | hybrid
description: string        # Rite workflow description

# Entry point
entry_point:
  agent: string            # First agent to invoke
  artifact:
    type: string           # Artifact type produced
    path_template: string  # Path with {slug} placeholder

# Phases (at least one required)
phases:
  - name: string           # Phase name (e.g., "requirements")
    agent: string          # Agent responsible
    produces: string       # Artifact type produced
    next: string|null      # Next phase name or null if terminal
    condition: string      # (optional) When this phase applies

# Complexity levels (at least one required)
complexity_levels:
  - name: enum             # SCRIPT | MODULE | SERVICE | PLATFORM (10x)
                           # PATCH | MODULE | SYSTEM | MIGRATION (ecosystem)
    scope: string          # Human-readable scope description
    phases: array          # Which phases apply at this level

# Optional: Back-routes for non-linear flow (NEW in 1.0)
back_routes:
  - source_phase: string   # Phase where back-route triggers
    trigger: string        # Named trigger condition
    target_phase: string   # Phase to route back to
    target_agent: string   # Agent to invoke
    requires_user_confirmation: boolean
    condition: string      # Human-readable condition

# Optional: Command to agent mappings
commands:
  - name: string           # Command name (e.g., "consolidate")
    file: string           # Command file name
    description: string    # Command description
    primary_agent: string  # Agent that handles this command
    workflow_phase: string # Which phase (or "all")

# Optional: Version for semantic versioning
version: string            # e.g., "1.0.0"
```

## Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Rite identifier |
| `workflow_type` | enum | Flow type (sequential/parallel/hybrid) |
| `description` | string | Rite workflow description |
| `entry_point` | object | First agent and artifact |
| `phases` | array | Ordered phase definitions (min 1) |
| `complexity_levels` | array | Complexity to phase mappings (min 1) |

## Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `version` | string | Semantic version |
| `back_routes` | array | Non-linear routing rules |
| `commands` | array | Command to agent mappings |

## Entry Point Object Schema

```yaml
entry_point:
  agent: string            # Agent name from phases
  artifact:
    type: string           # Artifact type (prd, gap-analysis, etc.)
    path_template: string  # Path pattern with {slug} placeholder
```

## Phase Object Schema

```yaml
phases:
  - name: string           # Kebab-case phase name
    agent: string          # Agent identifier
    produces: string       # Artifact type
    next: string|null      # Next phase or null for terminal
    condition: string      # (optional) Condition expression
    timeout: string        # (optional) Phase timeout (e.g., "1h")
    retry_count: integer   # (optional) Retries on failure
```

## Complexity Level Object Schema

```yaml
complexity_levels:
  - name: enum             # Complexity tier name
    scope: string          # Human description of scope
    phases: array          # Phases that apply at this level
    estimated_duration: string  # (optional) Time estimate
```

## Back-Route Object Schema (NEW)

```yaml
back_routes:
  - source_phase: string   # Phase where issue detected
    trigger: string        # Trigger identifier (kebab-case)
    target_phase: string   # Phase to return to
    target_agent: string   # Agent to invoke
    requires_user_confirmation: boolean  # Whether user must approve
    condition: string      # Human-readable trigger condition
```

### Standard Back-Route Triggers

| Trigger | Description | Typical Route |
|---------|-------------|---------------|
| `design-flaw-discovered` | Implementation reveals TDD issue | impl -> design |
| `requirement-ambiguity` | QA finds PRD unclear | validation -> requirements |
| `scope-expansion` | Work exceeds original scope | any -> requirements |
| `missing-prerequisite` | Dependency not met | any -> previous phase |
| `security-concern` | Security issue found | any -> design |

## Command Object Schema

```yaml
commands:
  - name: string           # Command name without slash
    file: string           # File in commands/ directory
    description: string    # Help text for command
    primary_agent: string  # Agent that handles command
    workflow_phase: string # "all" or specific phase name
```

## Validation Rules

### Structure Validation
1. File MUST be valid YAML
2. File MUST have `name`, `workflow_type`, `description`, `entry_point`, `phases`, `complexity_levels`

### Field Validation
1. `name` MUST be kebab-case, 3-50 characters
2. `workflow_type` MUST be one of: sequential, parallel, hybrid
3. `entry_point.agent` MUST reference an agent defined in `phases`
4. `entry_point.artifact.path_template` MUST contain `{slug}` placeholder

### Phase Validation
1. Each phase MUST have `name`, `agent`, `produces`
2. `next` MUST reference existing phase or be `null`
3. Phase names MUST be unique
4. Terminal phase (next: null) MUST exist

### Complexity Validation
1. Each level MUST have `name`, `scope`, `phases`
2. Level names MUST be unique
3. `phases` array MUST only contain phase names from `phases` list

### Back-Route Validation (if present)
1. `source_phase` MUST reference existing phase
2. `target_phase` MUST reference existing phase
3. `target_agent` MUST match agent for target phase
4. `trigger` MUST be kebab-case

### Condition Grammar

Condition strings support these expressions:
- `complexity >= LEVEL` - Phase applies at this level or higher
- `complexity == LEVEL` - Phase applies only at this level
- `artifact_exists("type")` - Artifact was produced
- `always` - Phase always applies (default if omitted)

## Example: Complete workflow.yaml

```yaml
name: 10x-dev
version: "1.0.0"
workflow_type: sequential
description: Full development lifecycle (PRD -> TDD -> Code -> QA)

entry_point:
  agent: requirements-analyst
  artifact:
    type: prd
    path_template: .ledge/specs/PRD-{slug}.md

phases:
  - name: requirements
    agent: requirements-analyst
    produces: prd
    next: design

  - name: design
    agent: architect
    produces: tdd
    next: implementation
    condition: "complexity >= MODULE"

  - name: implementation
    agent: principal-engineer
    produces: code
    next: validation

  - name: validation
    agent: qa-adversary
    produces: test-plan
    next: null

complexity_levels:
  - name: SCRIPT
    scope: "Single file, <200 LOC"
    phases: [requirements, implementation, validation]
  - name: MODULE
    scope: "Multiple files, <2000 LOC"
    phases: [requirements, design, implementation, validation]
  - name: SERVICE
    scope: "APIs, persistence"
    phases: [requirements, design, implementation, validation]
  - name: PLATFORM
    scope: "Multi-service"
    phases: [requirements, design, implementation, validation]

back_routes:
  - source_phase: implementation
    trigger: design-flaw-discovered
    target_phase: design
    target_agent: architect
    requires_user_confirmation: false
    condition: "Implementation reveals architectural issue not addressed in TDD"

  - source_phase: validation
    trigger: requirement-ambiguity
    target_phase: requirements
    target_agent: requirements-analyst
    requires_user_confirmation: true
    condition: "Test reveals PRD success criterion is ambiguous"

  - source_phase: implementation
    trigger: scope-expansion
    target_phase: requirements
    target_agent: requirements-analyst
    requires_user_confirmation: true
    condition: "Work required exceeds original PRD scope"

commands:
  - name: architect
    file: architect.md
    description: "Jump directly to design phase"
    primary_agent: architect
    workflow_phase: design
  - name: build
    file: build.md
    description: "Jump directly to implementation phase"
    primary_agent: principal-engineer
    workflow_phase: implementation
  - name: qa
    file: qa.md
    description: "Jump directly to validation phase"
    primary_agent: qa-adversary
    workflow_phase: validation
```

## Validation Function

Validates workflow YAML files against schema requirements including required fields, enum values, and business rules (phase references, back-route integrity).

**Function**: `validate_workflow_yaml()`

**Implementation**: Hook-based validation logic (implementation location TBD - likely `internal/hook/workflow_validate.go` projected via `ari sync`)

**Exit Codes**:
- 0 = valid
- 1 = file not found
- 2 = invalid YAML syntax
- 3 = missing required field
- 4 = field validation failed
- 5 = business rule violation

### Error Code Reference

| Code | Meaning | Example |
|------|---------|---------|
| 0 | Valid | Workflow passes all validation |
| 1 | File not found | Path does not exist |
| 2 | Invalid YAML | Syntax error in YAML |
| 3 | Missing required field | No `entry_point.agent` |
| 4 | Field validation failed | `workflow_type` not in enum |
| 5 | Business rule violation | `entry_point.agent` not in phases |

## Relationship to Other Artifacts

```
workflow.yaml
    |
    +-- Defines phase sequence for Pythia
    |
    +-- Maps complexity to phase selection
    |
    +-- back_routes enable non-linear flow
    |
    +-- commands enable direct agent access
```

## Migration from Pre-1.0

Existing workflow.yaml files without `back_routes` are valid - the field is optional. To upgrade:

1. Add `back_routes: []` (empty array) to acknowledge the field
2. Define back-routes based on rite workflow patterns
3. Set `version: "1.0.0"` to indicate schema compliance

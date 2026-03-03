# Rite Schema

> Complete schema for workflow.yaml files in rite definitions.

## Overview

A **rite** defines a specialized AI practice with agents, workflow, complexity levels, and command mappings. Rites live in `rites/<rite-name>/` and sync to satellites via the sync pipeline.

## File Structure

```
rites/<rite-name>/
├── workflow.yaml      # Required: Rite configuration
├── workflow.md        # Required: Human-readable documentation
├── README.md          # Required: Rite overview
├── agents/            # Required: Agent definitions
│   └── *.md
└── commands/          # Optional: Rite-specific commands
    └── *.md
```

**Naming**: Directory must be lowercase kebab-case (e.g., `hygiene`, `10x-dev`).

---

## Top-Level Fields

| Field | Required | Type | Description |
|-------|----------|------|-------------|
| `name` | Yes | string | Kebab-case, matches directory |
| `version` | No | string | Semantic version (e.g., "1.0.0") |
| `workflow_type` | Yes | enum | `sequential`, `parallel`, or `hybrid` |
| `description` | Yes | string | Brief lifecycle description (~100 chars) |

```yaml
name: 10x-dev
version: "1.0.0"
workflow_type: sequential
description: Full development lifecycle (PRD -> TDD -> Code -> QA)
```

---

## entry_point (required)

Defines workflow starting point.

| Field | Type | Description |
|-------|------|-------------|
| `agent` | string | Must match `phases[0].agent` and exist in `agents/` |
| `artifact.type` | string | Must match `phases[0].produces` |
| `artifact.path_template` | string | Path with `{slug}` placeholder |

```yaml
entry_point:
  agent: requirements-analyst
  artifact:
    type: prd
    path_template: .ledge/specs/PRD-{slug}.md
```

---

## phases (required)

Ordered list of workflow phases.

| Field | Required | Type | Description |
|-------|----------|------|-------------|
| `name` | Yes | string | Lowercase, unique identifier |
| `agent` | Yes | string | Must exist in `agents/` directory |
| `produces` | Yes | string | Artifact type (lowercase, hyphenated) |
| `next` | Yes | string/null | Next phase name, or `null` for terminal |
| `condition` | No | string | Complexity gate expression |

```yaml
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
    next: null  # Terminal
```

**Condition syntax**: `complexity >= <LEVEL>` or `complexity == <LEVEL>`

---

## complexity_levels (required)

Rite-specific complexity tiers controlling phase execution.

| Field | Required | Type | Description |
|-------|----------|------|-------------|
| `name` | Yes | string | UPPERCASE, unique |
| `scope` | Yes | string | Human-readable description |
| `phases` | Yes | array | Phase names to execute at this level |

```yaml
complexity_levels:
  - name: SCRIPT
    scope: "Single file, <200 LOC"
    phases: [requirements, implementation, validation]

  - name: MODULE
    scope: "Multiple files, <2000 LOC"
    phases: [requirements, design, implementation, validation]
```

**Common levels by domain**:

| Domain | Levels (small to large) |
|--------|-------------------------|
| Development | SCRIPT, MODULE, SERVICE, PLATFORM |
| Documentation | PAGE, SECTION, SITE |
| Security | PATCH, FEATURE, SYSTEM |
| Infrastructure | PATCH, MODULE, SYSTEM, MIGRATION |

---

## commands (optional)

Rite-specific slash commands.

| Field | Required | Type | Description |
|-------|----------|------|-------------|
| `name` | Yes | string | Command name (becomes `/name`) |
| `file` | Yes | string | Filename in `commands/` directory |
| `description` | Yes | string | Help text |
| `primary_agent` | No | string | Agent to invoke (default: entry point) |
| `workflow_phase` | No | string | Phase name or `all` (default: `all`) |

```yaml
commands:
  - name: quick-scan
    file: quick-scan.md
    description: "Run quick analysis only"
    primary_agent: analyst
    workflow_phase: analysis
```

---

## Agent Role Comments

Document standard command mappings:

```yaml
# Agent roles for command mapping:
# /architect  -> architect
# /build      -> principal-engineer
# /qa         -> qa-adversary
# /hotfix     -> principal-engineer (fast path)
# /code-review -> qa-adversary (review mode)
```

---

## Validation Rules

### Required Files
- [ ] `workflow.yaml`, `workflow.md`, `README.md` exist
- [ ] `agents/` directory has at least one `.md` file

### Field Consistency
- [ ] `name` matches directory name
- [ ] `entry_point.agent` equals `phases[0].agent`
- [ ] `entry_point.artifact.type` equals `phases[0].produces`
- [ ] `path_template` contains `{slug}`

### Phase Integrity
- [ ] All `phases[].agent` values exist in `agents/`
- [ ] All `phases[].next` reference valid phases or are `null`
- [ ] Exactly one phase has `next: null`
- [ ] No circular references

### Complexity Levels
- [ ] All `phases` in complexity levels reference valid phase names
- [ ] Level names are unique
- [ ] Higher levels include same or more phases than lower

---

## Anti-Patterns

| Anti-Pattern | Problem | Fix |
|--------------|---------|-----|
| Name mismatch | `name: ecosystemPack` in `rites/ecosystem/` | Use exact directory name |
| Orphan phase | Phase unreachable from entry | Ensure all phases linked via `next` |
| Missing terminal | No `next: null` phase | Add terminal phase |
| Agent mismatch | `agent: architect` but file is `software-architect.md` | Match filename exactly |
| Inverted complexity | LARGE level has fewer phases than SMALL | Higher levels should include more phases |
| Unused phase | Phase defined but not in any complexity level | Include in at least one level or remove |

---

## Quick Reference

### Minimal Valid workflow.yaml

```yaml
name: my-rite
workflow_type: sequential
description: Brief description

entry_point:
  agent: first-agent
  artifact:
    type: first-artifact
    path_template: .ledge/{category}/{slug}.md

phases:
  - name: first-phase
    agent: first-agent
    produces: first-artifact
    next: null

complexity_levels:
  - name: DEFAULT
    scope: "All work"
    phases: [first-phase]
```

### Complete Example

```yaml
name: full-rite
version: "1.0.0"
workflow_type: sequential
description: Complete example with all fields

entry_point:
  agent: analyst
  artifact:
    type: analysis-report
    path_template: .ledge/reviews/REPORT-{slug}.md

phases:
  - name: analysis
    agent: analyst
    produces: analysis-report
    next: design

  - name: design
    agent: architect
    produces: design-doc
    next: implementation
    condition: "complexity >= MODULE"

  - name: implementation
    agent: engineer
    produces: code
    next: validation

  - name: validation
    agent: tester
    produces: test-report
    next: null

complexity_levels:
  - name: PATCH
    scope: "Single file change"
    phases: [analysis, implementation, validation]

  - name: MODULE
    scope: "Multiple files"
    phases: [analysis, design, implementation, validation]

commands:
  - name: quick-scan
    file: quick-scan.md
    description: "Run quick analysis only"
    primary_agent: analyst
    workflow_phase: analysis

# Agent roles for command mapping:
# /architect  -> architect
# /build      -> engineer
# /qa         -> tester
```

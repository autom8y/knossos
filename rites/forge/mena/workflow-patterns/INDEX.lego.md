---
name: workflow-patterns
description: "Workflow.yaml templates, phase sequencing patterns, and complexity level examples. Use when: creating workflow.yaml for a rite, designing phase sequences, defining complexity levels. Triggers: workflow.yaml, phase sequencing, complexity levels, workflow template, quick-switch command."
---

# Workflow Patterns

> Templates and patterns for wiring agents into cohesive workflows.

## workflow.yaml Template

```yaml
name: {rite-name}
workflow_type: sequential
description: {One-line description}

entry_point:
  agent: {first-agent-name}
  artifact:
    type: {artifact-type}
    path_template: docs/{category}/{PREFIX}-{slug}.md

phases:
  - name: {phase-1}
    agent: {agent-1}
    produces: {artifact-1}
    next: {phase-2}

  - name: {phase-2}
    agent: {agent-2}
    produces: {artifact-2}
    next: {phase-3}
    condition: "complexity >= MODULE"  # Optional gating

  - name: {phase-3}
    agent: {agent-3}
    produces: {artifact-3}
    next: null  # Terminal phase

complexity_levels:
  - name: {LEVEL-1}
    scope: "{Scope description}"
    phases: [{phase-1}, {phase-3}]  # Skips phase-2

  - name: {LEVEL-2}
    scope: "{Scope description}"
    phases: [{phase-1}, {phase-2}, {phase-3}]

# Agent roles for command mapping:
# /architect    -> {design-phase-agent}
# /build        -> {implementation-agent}
# /qa           -> {validation-agent}
# /hotfix       -> {fast-path-agent}
# /code-review  -> {review-agent}
```

## Quick-Switch Command Template

```markdown
---
description: Quick switch to {rite-name}
allowed-tools: Bash, Read
---

## Your Task
Switch to {rite-name} and display the rite catalog.

## Behavior
1. Execute: `ari sync --rite {rite-name}`
2. Display agent catalog table
3. Show workflow phases
4. Update SESSION_CONTEXT if active session
```

---

## Workflow Patterns Library

### Standard 4-Phase Development
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
    next: null
```

### Simplified 3-Phase
```yaml
phases:
  - name: analysis
    agent: analyst
    produces: report
    next: execution
  - name: execution
    agent: executor
    produces: output
    next: review
  - name: review
    agent: reviewer
    produces: signoff
    next: null
```

### Complexity Level Patterns

**Development (4 levels)**:
- SCRIPT: [requirements, implementation, validation]
- MODULE: [requirements, design, implementation, validation]
- SERVICE: [requirements, design, implementation, validation]
- PLATFORM: [requirements, design, implementation, validation]

**Documentation (3 levels)**:
- PAGE: [audit, writing, review]
- SECTION: [audit, architecture, writing, review]
- SITE: [audit, architecture, writing, review]

**Assessment (2 levels)**:
- QUICK: [assessment, planning]
- AUDIT: [assessment, analysis, planning]

# Team Examples

> Reference implementations for team pack patterns.

## 3-Agent Team: debt-triage-pack

A minimal team focused on planning, not implementation.

**Characteristics**:
- 3 agents, 3 phases (collection -> assessment -> planning)
- Planning-only output (produces plans, doesn't execute)
- No orchestrator needed
- Work handed off to other teams for execution

### Directory Structure

```
debt-triage-pack/
├── agents/
│   ├── debt-collector.md
│   ├── risk-assessor.md
│   └── sprint-planner.md
└── workflow.yaml
```

### workflow.yaml

```yaml
name: debt-triage-pack
workflow_type: sequential
description: Technical debt discovery, risk assessment, and paydown planning

entry_point:
  agent: debt-collector
  artifact:
    type: debt-ledger
    path_template: docs/debt/LEDGER-{slug}.md

phases:
  - name: collection
    agent: debt-collector
    produces: debt-ledger
    next: assessment

  - name: assessment
    agent: risk-assessor
    produces: risk-report
    next: planning

  - name: planning
    agent: sprint-planner
    produces: sprint-plan
    next: null

complexity_levels:
  - name: QUICK
    scope: "Known debt items, immediate assessment"
    phases: [assessment, planning]

  - name: AUDIT
    scope: "Full codebase debt discovery"
    phases: [collection, assessment, planning]

# Agent roles for command mapping:
# /architect   -> risk-assessor (closest to design)
# /build       -> (N/A - planning only team)
# /qa          -> (N/A - planning only team)
```

### Agent Frontmatter

```yaml
# debt-collector.md (Entry Agent)
---
name: debt-collector
description: |
  Discovers and catalogs technical debt across the codebase.
  Invoke when auditing debt, inventorying issues, or starting debt triage.
  Produces debt-ledger.
tools: Bash, Glob, Grep, Read, TodoWrite
model: haiku
color: orange
---

# risk-assessor.md (Middle Agent)
---
name: risk-assessor
description: |
  Evaluates technical debt risk and prioritizes remediation.
  Invoke when prioritizing debt, assessing risk, or ranking issues.
  Produces risk-report.
tools: Bash, Glob, Grep, Read, TodoWrite
model: sonnet
color: cyan
---

# sprint-planner.md (Terminal Agent)
---
name: sprint-planner
description: |
  Creates actionable sprint plans for debt paydown.
  Invoke when planning sprints, scheduling debt work, or creating paydown plans.
  Produces sprint-plan.
tools: Bash, Glob, Grep, Read, Write, TodoWrite
model: sonnet
color: pink
---
```

### Workflow Diagram

```
┌───────────────┐     ┌───────────────┐     ┌───────────────┐
│debt-collector │────>│ risk-assessor │────>│sprint-planner │
│   (haiku)     │     │   (sonnet)    │     │   (sonnet)    │
│   orange      │     │     cyan      │     │     pink      │
└───────────────┘     └───────────────┘     └───────────────┘
        │                    │                     │
        v                    v                     v
   debt-ledger          risk-report          sprint-plan
```

### When to Use 3-Agent Pattern

**Use when**:
- Domain is focused and specialized
- No implementation phase needed
- Output is plans, reports, or assessments
- Linear flow without branches
- Other teams handle execution

**Avoid when**:
- Implementation is part of the workflow
- Validation of output is needed
- Domain has multiple concerns

---

## 5-Agent Team: 10x-dev-pack

A full lifecycle team with orchestrator for complex development.

**Characteristics**:
- 5 agents (4 phases + orchestrator)
- 4 phases (requirements -> design -> implementation -> validation)
- Full lifecycle (from requirements to tested code)
- Orchestrator for PLATFORM complexity coordination

### Directory Structure

```
10x-dev-pack/
├── agents/
│   ├── orchestrator.md
│   ├── requirements-analyst.md
│   ├── architect.md
│   ├── principal-engineer.md
│   └── qa-adversary.md
└── workflow.yaml
```

### workflow.yaml

```yaml
name: 10x-dev-pack
workflow_type: sequential
description: Full-lifecycle software development with PRD/TDD/Code/Test pipeline

entry_point:
  agent: requirements-analyst
  artifact:
    type: prd
    path_template: docs/requirements/PRD-{slug}.md

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
    scope: "Single file, <200 LOC, no new APIs"
    phases: [requirements, implementation, validation]

  - name: MODULE
    scope: "Multiple files, <2000 LOC, internal APIs"
    phases: [requirements, design, implementation, validation]

  - name: SERVICE
    scope: "New service, external APIs, persistence"
    phases: [requirements, design, implementation, validation]

  - name: PLATFORM
    scope: "Multi-service, cross-team coordination"
    phases: [requirements, design, implementation, validation]

# Agent roles for command mapping:
# /architect   -> architect
# /build       -> principal-engineer
# /qa          -> qa-adversary
# /hotfix      -> principal-engineer (fast path)
# /code-review -> qa-adversary (review mode)
```

### Agent Frontmatter

```yaml
# orchestrator.md (Coordinator - not in phases)
---
name: orchestrator
description: |
  Coordinates multi-phase development initiatives.
  Invoke for complex projects, PLATFORM complexity, or multi-sprint work.
  Manages handoffs between agents.
tools: Bash, Glob, Grep, Read, Write, Task, TodoWrite
model: opus
color: purple
---

# requirements-analyst.md (Entry Agent)
---
name: requirements-analyst
description: |
  Clarifies intent and captures requirements.
  Invoke when starting features, defining scope, or writing PRDs.
  Produces prd.
tools: Bash, Glob, Grep, Read, Write, WebFetch, WebSearch, AskUserQuestion, TodoWrite
model: opus
color: pink
---

# architect.md (Design Agent - conditional)
---
name: architect
description: |
  Designs solutions and makes architectural decisions.
  Invoke for system design, API design, or technical planning.
  Produces tdd, adr.
tools: Bash, Glob, Grep, Read, Write, TodoWrite
model: opus
color: cyan
---

# principal-engineer.md (Implementation Agent)
---
name: principal-engineer
description: |
  Implements solutions with craft and discipline.
  Invoke for coding, refactoring, or building features.
  Produces code.
tools: Bash, Glob, Grep, Read, Edit, Write, NotebookEdit, Task, TodoWrite
model: sonnet
color: green
---

# qa-adversary.md (Validation Agent - terminal)
---
name: qa-adversary
description: |
  Validates quality and finds problems before production.
  Invoke for testing, code review, or quality validation.
  Produces test-plan.
tools: Bash, Glob, Grep, Read, Edit, Write, TodoWrite
model: opus
color: red
---
```

### Workflow Diagram

```
                         ┌─────────────────┐
                         │   orchestrator  │
                         │     (opus)      │
                         │     purple      │
                         └────────┬────────┘
                                  │ coordinates
                                  v
┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐
│  requirements-   │───>│    architect     │───>│    principal-    │───>│   qa-adversary   │
│    analyst       │    │     (opus)       │    │     engineer     │    │     (opus)       │
│    (opus)        │    │      cyan        │    │    (sonnet)      │    │      red         │
│     pink         │    │                  │    │     green        │    │                  │
└──────────────────┘    └──────────────────┘    └──────────────────┘    └──────────────────┘
         │                       │                       │                       │
         v                       v                       v                       v
       PRD                     TDD                    Code                  Test Plan
```

### Orchestrator Role

The orchestrator is special:
- **Not a phase**: Manages phases, doesn't participate in them
- **Invoked for**: PLATFORM complexity, multi-sprint work
- **Responsibilities**: Track progress, manage handoffs, coordinate cross-cutting concerns

### Complexity Gating

SCRIPT complexity skips design (small changes don't need formal TDD):
- **SCRIPT**: requirements -> implementation -> validation
- **MODULE+**: requirements -> design -> implementation -> validation

### When to Use 5-Agent Pattern

**Use when**:
- Full development lifecycle needed
- Multiple phases with distinct outputs
- High-complexity work (SERVICE, PLATFORM)
- Multiple handoffs requiring coordination
- Quality gates between phases

**Include orchestrator when**:
- PLATFORM complexity
- Multi-sprint initiatives

**Omit orchestrator when**:
- Simpler work (agents self-coordinate)

---

## Comparison: 3 vs 4 vs 5 Agents

| Team | Agents | Phases | Orchestrator | Use Case |
|------|--------|--------|--------------|----------|
| debt-triage | 3 | 3 | No | Planning only |
| sre-pack | 4 | 4 | No | Standard lifecycle |
| 10x-dev | 5 | 4 | Yes | Full lifecycle |

The extra agent in 5-agent teams is typically the orchestrator, which coordinates but doesn't own a phase.

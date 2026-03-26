---
last_verified: 2026-02-26
---

# Rite: 10x-dev

> Full development lifecycle from requirements through validation.

The 10x-dev rite provides a complete workflow for feature implementation: PRD → TDD → Code → QA. It's the primary rite for building production features.

---

## Overview

| Property | Value |
|----------|-------|
| **Name** | 10x-dev |
| **Form** | Full (multi-agent workflow) |
| **Agents** | 5 |
| **Entry Agent** | potnia |

---

## When to Use

- Implementing new features
- Building from requirements to deployment
- Need integrated design-to-test workflow
- Complex work requiring multiple specialists

---

## Agents

| Agent | Role |
|-------|------|
| **potnia** | Coordinates development lifecycle phases and routes work to specialists |
| **requirements-analyst** | Gathers requirements and produces PRD artifacts |
| **architect** | Creates technical design documents and architecture decisions |
| **principal-engineer** | Implements code according to design specifications |
| **qa-adversary** | Validates implementation through adversarial testing |

See agent files: `rites/10x-dev/agents/`

---

## Workflow Phases

```mermaid
flowchart LR
    A[requirements] --> B[design]
    B --> C[implementation]
    C --> D[validation]
    D --> E[complete]
```

| Phase | Agent | Produces | Condition |
|-------|-------|----------|-----------|
| requirements | requirements-analyst | PRD | Always |
| design | architect | TDD | complexity >= MODULE |
| implementation | principal-engineer | Code | Always |
| validation | qa-adversary | Test Plan | Always |

---

## Invocation Patterns

```bash
# Full task lifecycle
/task "implement feature X"

# Multi-task sprint
/sprint "Authentication Sprint" --tasks="Login,Logout,Session"

# Direct orchestrator invocation
Task(potnia, "implement user authentication")
```

---

## Complexity Levels

| Level | Scope | Design Phase |
|-------|-------|--------------|
| PATCH | Bug fix, minor change | Skipped |
| MODULE | Single component | Required |
| SYSTEM | Multi-component | Required |
| INITIATIVE | Major feature | Required |
| MIGRATION | Breaking change | Required |

---

## Skills

- `10x-ref` — Workflow reference
- `10x-workflow` — Phase orchestration

---

## Source

**Manifest**: `rites/10x-dev/manifest.yaml`

---

## See Also

- [CLI: session](../operations/cli-reference/cli-session.md) — Session management
- [CLI: rite](../operations/cli-reference/cli-rite.md) — Rite operations
- `/orchestration` skill — Consultation protocol routing

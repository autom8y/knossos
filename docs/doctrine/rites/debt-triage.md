# Rite: debt-triage

> Debt management lifecycle for assessing and planning technical debt paydown.

The debt-triage rite provides workflows for collecting, assessing, and planning technical debt remediation.

---

## Overview

| Property | Value |
|----------|-------|
| **Name** | debt-triage |
| **Form** | Full (multi-agent workflow) |
| **Agents** | 4 |
| **Entry Agent** | orchestrator |

---

## When to Use

- Inventorying technical debt
- Assessing debt risk and impact
- Planning debt paydown sprints
- Prioritizing remediation efforts

---

## Agents

| Agent | Role |
|-------|------|
| **orchestrator** | Coordinates debt assessment and planning phases |
| **debt-collector** | Collects and inventories technical debt across the codebase |
| **risk-assessor** | Assesses risk and impact of debt items for prioritization |
| **sprint-planner** | Plans debt paydown sprints with timelines and resources |

See agent files: `/roster/rites/debt-triage/agents/`

---

## Workflow Phases

```mermaid
flowchart LR
    A[collection] --> B[assessment]
    B --> C[planning]
    C --> D[complete]
```

| Phase | Agent | Produces | Condition |
|-------|-------|----------|-----------|
| collection | debt-collector | Debt Ledger | Always |
| assessment | risk-assessor | Risk Report | Always |
| planning | sprint-planner | Sprint Plan | Always |

---

## Invocation Patterns

```bash
# Quick switch to debt-triage
/debt

# Collect debt in area
Task(debt-collector, "inventory debt in authentication module")

# Assess specific debt items
Task(risk-assessor, "assess risk of database migration debt")
```

---

## Skills

- `debt-ref` — Workflow reference

---

## Source

**Manifest**: `/roster/rites/debt-triage/manifest.yaml`

---

## See Also

- [CLI: rite](../operations/cli-reference/cli-rite.md)
- [Shared Templates](/shared-templates)

---
last_verified: 2026-02-26
---

# Rite: debt-triage

> Debt management lifecycle for assessing and planning technical debt paydown.

The debt-triage rite turns "we have a lot of debt" into a prioritized paydown plan with sprint boundaries and resource allocations. It does not assume all debt is equal — debt-collector catalogs across five categories (code, docs, tests, infra, design) with origin context, then risk-assessor scores each item by blast radius, trigger likelihood, and remediation effort. The output is not a complaint list but a decision-ready risk matrix. This distinguishes debt-triage from a hygiene pass: hygiene detects and fixes code smells in the current codebase; debt-triage produces a ledger that management can schedule and fund. The sprint-planner phase converts that ledger into wave-based paydown sprints with explicit timelines — what ships in wave 1, what waits for wave 2, and why.

---

## Overview

| Property | Value |
|----------|-------|
| **Name** | debt-triage |
| **Form** | Full (multi-agent workflow) |
| **Agents** | 4 |
| **Entry Agent** | potnia |

---

## When to Use

- Building a structured debt ledger before a major refactoring initiative so nothing is forgotten
- Producing a risk-scored priority matrix when leadership needs to decide what to fund
- Planning debt paydown waves with explicit timelines and resource allocations
- Scoping a cleanup sprint after a period of fast shipping
- **Not for**: detecting individual code smells or executing the actual cleanup — use hygiene for that. Debt-triage plans the work; hygiene and 10x-dev execute it.

---

## Agents

| Agent | Role |
|-------|------|
| **potnia** | Coordinates debt assessment and planning phases; ensures each phase output meets quality gates before advancing |
| **debt-collector** | Catalogs all debt systematically across code, docs, tests, infra, and design — records origin context and category, never inflates severity |
| **risk-assessor** | Scores each ledger item by blast radius, trigger likelihood, and remediation effort; produces risk matrix with quick-wins list for leadership |
| **sprint-planner** | Converts the risk matrix into wave-based paydown sprints with explicit timelines, resource allocations, and rollback boundaries |

See agent files: `rites/debt-triage/agents/`

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

# Inventory all debt across the codebase — give collector an explicit scope
Task(debt-collector, "audit the codebase for all technical debt and produce a structured ledger")

# Score and prioritize after ledger is complete
Task(risk-assessor, "score all ledger items by blast radius, likelihood, and effort — produce risk matrix and quick-wins list")

# Plan paydown sprints from the risk matrix
Task(sprint-planner, "plan two-wave paydown from risk matrix — wave 1: quick wins, wave 2: high-blast-radius items")
```

---

## Skills

- `debt-ref` — Workflow reference

---

## Source

**Manifest**: `rites/debt-triage/manifest.yaml`

---

## See Also

- [CLI: rite](../operations/cli-reference/cli-rite.md)
- `/shared-templates` skill — Shared rite templates

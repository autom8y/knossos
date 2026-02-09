---
name: debt
description: "Switch to debt-triage rite (technical debt management). Use when: user says /debt, wants debt inventory, debt assessment, remediation planning. Triggers: /debt, debt rite, debt triage, debt planning, technical debt."
context: fork
---

# /debt - Switch to Technical Debt Rite

Switch to debt-triage, the technical debt identification and remediation planning rite.

## Behavior

### 1. Invoke Rite Switch

Execute via Bash tool:

```bash
ari sync --rite debt-triage
```

### 2. Display Knossos

After successful switch, show the agent table:

| Agent | Role |
|-------|------|
| debt-collector | Identifies and catalogs technical debt |
| risk-assessor | Evaluates impact and urgency of debt |
| sprint-planner | Creates remediation plans and roadmaps |

### 3. Update Session

If a session is active, update `active_rite` to `debt-triage`.

## When to Use

- Debt inventories and comprehensive scans
- Risk assessment and priority ranking
- Roadmap planning for multi-quarter remediation
- Portfolio-level debt management, due diligence

**Don't use for**: Executing refactoring --> `/hygiene` | Feature work --> `/10x` | Documentation --> `/docs`

**Debt vs Hygiene**: Use `/debt` to plan and prioritize. Use `/hygiene` to execute cleanup.

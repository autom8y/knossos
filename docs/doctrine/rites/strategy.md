---
last_verified: 2026-02-26
---

# Rite: strategy

> Business strategy lifecycle for market research, competitive analysis, and strategic planning.

The strategy rite provides workflows for strategic business decisions through market research, competitive analysis, and financial modeling.

---

## Overview

| Property | Value |
|----------|-------|
| **Name** | strategy |
| **Form** | Full (multi-agent workflow) |
| **Agents** | 5 |
| **Entry Agent** | pythia |

---

## When to Use

- Researching market opportunities
- Analyzing competitive landscape
- Building financial models
- Creating strategic roadmaps
- Go/no-go decisions

---

## Agents

| Agent | Role |
|-------|------|
| **pythia** | Coordinates strategic initiative phases |
| **market-researcher** | Researches markets and identifies customer segments |
| **competitive-analyst** | Analyzes competitive landscape and identifies opportunities |
| **business-model-analyst** | Builds financial models and analyzes unit economics |
| **roadmap-strategist** | Creates strategic roadmaps with execution timelines |

See agent files: `rites/strategy/agents/`

---

## Workflow Phases

```mermaid
flowchart LR
    A[market-research] --> B[competitive-analysis]
    B --> C[business-modeling]
    C --> D[strategic-planning]
    D --> E[complete]
```

| Phase | Agent | Produces | Condition |
|-------|-------|----------|-----------|
| market-research | market-researcher | Market Analysis | Always |
| competitive-analysis | competitive-analyst | Competitive Intel | complexity >= STRATEGIC |
| business-modeling | business-model-analyst | Financial Model | Always |
| strategic-planning | roadmap-strategist | Strategic Roadmap | Always |

---

## Invocation Patterns

```bash
# Quick switch to strategy
/strategy

# Market research
Task(market-researcher, "research enterprise SaaS market")

# Competitive analysis
Task(competitive-analyst, "analyze competitors in API gateway space")

# Business modeling
Task(business-model-analyst, "model unit economics for new pricing tier")
```

---

## Skills

- `doc-strategy` — Strategy documentation
- `strategy-ref` — Workflow reference

---

## Source

**Manifest**: `rites/strategy/manifest.yaml`

---

## See Also

- [CLI: rite](../operations/cli-reference/cli-rite.md)

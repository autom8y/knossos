---
last_verified: 2026-02-26
---

# Rite: strategy

> Business strategy lifecycle for market research, competitive analysis, and strategic planning.

The strategy rite turns business questions into executable plans with financial backing and competitive context. It does not start with roadmaps — it starts with market-researcher mapping customer segments and sizing opportunities before competitive-analyst evaluates where those segments are already served and by whom. Business-model-analyst then builds the unit economics: not revenue projections, but cost structures, margin profiles, and the specific assumptions that make the numbers work. The rite ends with roadmap-strategist applying explicit prioritization frameworks (RICE, ICE, weighted scoring) to sequence initiatives against resource constraints — producing a roadmap that explains its own priorities rather than asserting them.

---

## Overview

| Property | Value |
|----------|-------|
| **Name** | strategy |
| **Form** | Full (multi-agent workflow) |
| **Agents** | 5 |
| **Entry Agent** | potnia |

---

## When to Use

- Evaluating a new market or product opportunity before committing to a roadmap
- Assessing the competitive landscape when a competitor enters your space or you plan to enter theirs
- Building unit economics models for pricing, expansion, or investment decisions
- Planning quarterly or annual roadmaps with prioritized initiatives and explicit resource allocations
- **Not for**: internal product analytics or user behavior research — use intelligence for that. Not for technology stack decisions — use rnd. Strategy targets business decisions; the other rites target technical ones.

---

## Agents

| Agent | Role |
|-------|------|
| **potnia** | Coordinates strategic planning phases; gates business modeling on completed market and competitive analysis |
| **market-researcher** | Maps customer segments, sizes addressable markets, and surfaces unmet needs — evidence-based, not assumption-based |
| **competitive-analyst** | Profiles competitors with predicted next moves and threat levels; produces positioning maps and battlecards with objection handling |
| **business-model-analyst** | Builds unit economics models with explicit assumptions — cost structures, margin profiles, and the specific numbers that determine viability |
| **roadmap-strategist** | Prioritizes initiatives using RICE/ICE frameworks, allocates resources against capacity, and produces OKR-aligned roadmaps that explain their own priorities |

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

# Evaluate a new market opportunity
Task(market-researcher, "research enterprise SaaS market for AI coding tools — size TAM/SAM/SOM and identify underserved segments")

# Assess a competitive threat
Task(competitive-analyst, "competitor X just launched in our space — build threat profile, assess positioning, predict their next moves")

# Build unit economics for a pricing decision
Task(business-model-analyst, "model unit economics for new enterprise tier — include cost structure, margin, and break-even assumptions")

# Build a prioritized roadmap with resource constraints
Task(roadmap-strategist, "plan Q2 roadmap with 5 engineers, 3 initiatives competing — apply RICE scoring and produce OKR-aligned plan")
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

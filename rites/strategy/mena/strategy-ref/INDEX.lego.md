---
name: strategy-ref
description: "Strategy team reference. Triggers: strategy, market-researcher, competitive-analyst, business-model-analyst, roadmap-strategist."
---

# Strategy Team (strategy)

> Research. Analyze. Model. Strategize.

## Quick Reference

| Component | Location | Purpose |
|-----------|----------|---------|
| Agents | `$KNOSSOS_HOME/rites/strategy/agents/` | Agent prompts |
| Workflow | `$KNOSSOS_HOME/rites/strategy/workflow.yaml` | Phase configuration |
| Switch | `/strategy` | Activate this team |

## Pantheon

| Agent | Model | Role | Produces |
|-------|-------|------|----------|
| **market-researcher** | opus | Maps market terrain and trends | market-analysis |
| **competitive-analyst** | opus | Tracks competitors and predicts moves | competitive-intel |
| **business-model-analyst** | opus | Stress-tests unit economics | financial-model |
| **roadmap-strategist** | opus | Connects vision to execution | strategic-roadmap |

## Workflow

```
market-research → competitive-analysis → business-modeling → strategic-planning
       │                  │                    │                    │
       ▼                  ▼                    ▼                    ▼
 MARKET-{slug}     COMPETE-{slug}       FINANCE-{slug}      STRATEGY-{slug}
```

### Phase Details

| Phase | Agent | Input | Output |
|-------|-------|-------|--------|
| market-research | market-researcher | Strategic question | Market analysis |
| competitive-analysis | competitive-analyst | Market context | Competitive intel |
| business-modeling | business-model-analyst | Competitive context | Financial model |
| strategic-planning | roadmap-strategist | Financial model | Strategic roadmap |

## Complexity Levels

| Level | Scope | Phases |
|-------|-------|--------|
| **TACTICAL** | Single decision, existing data | business-modeling, strategic-planning |
| **STRATEGIC** | New market entry, major bet | All phases |
| **TRANSFORMATION** | Business model change | All phases |

## Command Mapping

| Command | Maps To | Use When |
|---------|---------|----------|
| `/strategy` | Team switch | Activating this team |
| `/architect` | competitive-analyst | Competitive analysis only |
| `/build` | business-model-analyst | Financial modeling only |
| `/qa` | roadmap-strategist | Strategic planning only |
| `/hotfix` | N/A | Not applicable (strategy team) |
| `/code-review` | N/A | Not applicable (strategy team) |

## When to Use This Rite

**Use strategy when:**
- Market sizing and analysis
- Competitive intelligence
- Pricing and business model analysis
- Strategic roadmap planning

**Don't use strategy when:**
- Building features → Use 10x-dev
- Product analytics → Use intelligence
- Technology evaluation → Use rnd

## Agent Summaries

### Market Researcher

**Purpose**: Map the market terrain

**Key Responsibilities**:
- Market sizing (TAM/SAM/SOM)
- Segment analysis
- Trend identification
- Buyer research

**Produces**: `docs/strategy/MARKET-{slug}.md`

---

### Competitive Analyst

**Purpose**: Know competitors better than they know themselves

**Key Responsibilities**:
- Competitor monitoring
- Competitive intelligence
- Market positioning
- Predictive analysis

**Produces**: `docs/strategy/COMPETE-{slug}.md`

---

### Business Model Analyst

**Purpose**: Stress-test how we make money

**Key Responsibilities**:
- Unit economics analysis
- Pricing analysis
- Margin analysis
- Scenario modeling

**Produces**: `docs/strategy/FINANCE-{slug}.md`

---

### Roadmap Strategist

**Purpose**: Connect vision to execution

**Key Responsibilities**:
- Strategic prioritization
- Resource allocation
- OKR design
- Roadmap construction

**Produces**: `docs/strategy/STRATEGY-{slug}.md`

## Cross-References

- **Related Skills**: @documentation (artifact templates)
- **Related Rites**: intelligence (product analytics), rnd (technology strategy)
- **Commands**: See COMMAND_REGISTRY.md

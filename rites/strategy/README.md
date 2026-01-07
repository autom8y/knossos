# Strategy Pack

Market research, competitive analysis, business modeling, and strategic planning for data-driven decision-making.

## When to Use This Rite

**Triggers**:
- "What's the TAM for this market opportunity?"
- "Should we enter the enterprise segment?"
- "What would usage-based pricing do to our revenue?"
- "How should we prioritize these competing initiatives?"

**Not for**: Tactical feature decisions, engineering implementation, or day-to-day product management.

## Quick Start

```bash
/rite strategy
```

## Agents

| Agent | Role | Artifact |
|-------|------|----------|
| market-researcher | Maps market terrain and identifies opportunities | market-analysis |
| competitive-analyst | Tracks competitors and predicts market moves | competitive-intel |
| business-model-analyst | Stress-tests how the business makes money | financial-model |
| roadmap-strategist | Connects company vision to quarterly execution | strategic-roadmap |

## Workflow

Sequential workflow with complexity-based phase skipping:
- **TACTICAL**: business-modeling → strategic-planning
- **STRATEGIC**: market-research → competitive-analysis → business-modeling → strategic-planning
- **TRANSFORMATION**: All phases (business model changes, company pivots)

See `workflow.md` for phase flow and complexity levels.

## When to Use Strategy vs Intelligence

**Core distinction**: Strategy looks **outward** (market, competitors, positioning). Intelligence looks **inward** (our users, our product).

| Question | Team | Why |
|----------|------|-----|
| "What's the market opportunity in healthcare?" | Strategy | External market analysis |
| "How do users navigate our checkout flow?" | Intelligence | User behavior within our product |
| "How should we position against Competitor X?" | Strategy | Competitive positioning |
| "Which features drive retention?" | Intelligence | Our product's impact on our users |
| "Should we expand to the enterprise segment?" | Strategy | Market entry decision |
| "What do our NPS scores tell us?" | Intelligence | Our user sentiment |
| "What pricing model do competitors use?" | Strategy | External competitive intel |
| "What caused the conversion drop last week?" | Intelligence | Our funnel analysis |
| "What's our TAM in the European market?" | Strategy | External market sizing |
| "How do power users differ from churned users?" | Intelligence | Our user segmentation |

**Handoff triggers**:
- Strategy -> Intelligence: Strategic hypotheses that need user validation
- Intelligence -> Strategy: Insights about user needs that suggest market opportunities

See also: [intelligence README](../intelligence/README.md#when-to-use-intelligence-vs-strategy)

## Architecture Strategy Integration

**Workflow**: moonshot-architect (RND) -> roadmap-strategist (Strategy)

When RND's moonshot-architect produces a MOONSHOT artifact with `business_impact > "significant"`, it triggers handoff to strategy for roadmap integration.

**Trigger criteria**:
- Artifact type: MOONSHOT (from moonshot-architect)
- Business impact assessment: `significant` or `transformational`
- Technical feasibility: validated via prototype or analysis

**Integration process**:
1. moonshot-architect produces MOONSHOT artifact with business impact assessment
2. If `business_impact > "significant"`, handoff to roadmap-strategist
3. roadmap-strategist evaluates strategic fit and timeline
4. Output: Updated strategic-roadmap incorporating moonshot initiatives

See also: [rnd README](../rnd/README.md#architecture-strategy-integration)

## Related Rites

- **10x-dev**: Hand off strategic roadmap for implementation
- **security**: When strategy involves new markets with compliance requirements
- **rnd**: Receive moonshot architecture plans for strategic integration

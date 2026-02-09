# Intelligence Pack

Product analytics, user research, experimentation, and insights synthesis.

## When to Use This Rite

**Triggers**:
- "How do users actually use this feature?"
- "We need to track conversion for the new onboarding"
- "Should we A/B test this change?"
- "What do the metrics tell us about this launch?"

**Not for**: Implementation or feature development - this team analyzes and validates product decisions.

## Quick Start

```bash
/rite intelligence
```

## Agents

| Agent | Role | Artifact |
|-------|------|----------|
| analytics-engineer | Builds data foundation and tracking plans | tracking-plan |
| user-researcher | Captures qualitative 'why' behind user behavior | research-findings |
| experimentation-lead | Designs rigorous A/B tests and experiments | experiment-design |
| insights-analyst | Synthesizes data into actionable recommendations | insights-report |

## Workflow

See `workflow.md` for phase flow and complexity levels.

**Complexity Levels**:
- **METRIC**: Single metric analysis, existing event data
- **FEATURE**: New feature instrumentation, user journey analysis
- **INITIATIVE**: Cross-feature analysis, strategic product decisions

## When to Use Intelligence vs Strategy

**Core distinction**: Intelligence looks **inward** (our users, our product). Strategy looks **outward** (market, competitors, positioning).

| Question | Team | Why |
|----------|------|-----|
| "How do users navigate our checkout flow?" | Intelligence | User behavior within our product |
| "What's the market opportunity in healthcare?" | Strategy | External market analysis |
| "Which features drive retention?" | Intelligence | Our product's impact on our users |
| "How should we position against Competitor X?" | Strategy | Competitive positioning |
| "What do our NPS scores tell us?" | Intelligence | Our user sentiment |
| "Should we expand to the enterprise segment?" | Strategy | Market entry decision |
| "What caused the conversion drop last week?" | Intelligence | Our funnel analysis |
| "What pricing model do competitors use?" | Strategy | External competitive intel |
| "How do power users differ from churned users?" | Intelligence | Our user segmentation |
| "What's our TAM in the European market?" | Strategy | External market sizing |

**Handoff triggers**:
- Intelligence -> Strategy: Insights about user needs that suggest market opportunities
- Strategy -> Intelligence: Strategic hypotheses that need user validation

See also: [strategy README](../strategy/README.md#when-to-use-strategy-vs-intelligence)

## Related Rites

- **strategy**: Hand off for business case development and strategic planning
- **10x-dev**: Hand off instrumentation implementation to engineering teams

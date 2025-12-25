# Strategy Pack

Market research, competitive analysis, business modeling, and strategic planning for data-driven decision-making.

## When to Use This Team

**Triggers**:
- "What's the TAM for this market opportunity?"
- "Should we enter the enterprise segment?"
- "What would usage-based pricing do to our revenue?"
- "How should we prioritize these competing initiatives?"

**Not for**: Tactical feature decisions, engineering implementation, or day-to-day product management.

## Quick Start

```bash
/team strategy-pack
```

## Agents

| Agent | Role | Model | Artifact |
|-------|------|-------|----------|
| market-researcher | Maps market terrain and identifies opportunities | claude-sonnet-4-5 | market-analysis |
| competitive-analyst | Tracks competitors and predicts market moves | claude-opus-4-5 | competitive-intel |
| business-model-analyst | Stress-tests how the business makes money | claude-opus-4-5 | financial-model |
| roadmap-strategist | Connects company vision to quarterly execution | claude-opus-4-5 | strategic-roadmap |

## Workflow

Sequential workflow with complexity-based phase skipping:
- **TACTICAL**: business-modeling → strategic-planning
- **STRATEGIC**: market-research → competitive-analysis → business-modeling → strategic-planning
- **TRANSFORMATION**: All phases (business model changes, company pivots)

See `workflow.md` for phase flow and complexity levels.

## Related Teams

- **10x-dev-pack**: Hand off strategic roadmap for implementation
- **security-pack**: When strategy involves new markets with compliance requirements

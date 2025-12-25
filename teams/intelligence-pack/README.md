# Intelligence Pack

Product analytics, user research, experimentation, and insights synthesis.

## When to Use This Team

**Triggers**:
- "How do users actually use this feature?"
- "We need to track conversion for the new onboarding"
- "Should we A/B test this change?"
- "What do the metrics tell us about this launch?"

**Not for**: Implementation or feature development - this team analyzes and validates product decisions.

## Quick Start

```bash
/team intelligence-pack
```

## Agents

| Agent | Role | Model | Artifact |
|-------|------|-------|----------|
| analytics-engineer | Builds data foundation and tracking plans | claude-sonnet-4-5 | tracking-plan |
| user-researcher | Captures qualitative 'why' behind user behavior | claude-opus-4-5 | research-findings |
| experimentation-lead | Designs rigorous A/B tests and experiments | claude-opus-4-5 | experiment-design |
| insights-analyst | Synthesizes data into actionable recommendations | claude-opus-4-5 | insights-report |

## Workflow

See `workflow.md` for phase flow and complexity levels.

**Complexity Levels**:
- **METRIC**: Single metric analysis, existing event data
- **FEATURE**: New feature instrumentation, user journey analysis
- **INITIATIVE**: Cross-feature analysis, strategic product decisions

## Related Teams

- **strategy-pack**: Hand off for business case development and strategic planning
- **ship-pack**: Hand off instrumentation implementation to engineering teams

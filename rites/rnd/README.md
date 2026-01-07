# R&D Pack

Technology scouting, integration analysis, prototyping, and future architecture.

## When to Use This Rite

**Triggers**:
- "Should we evaluate [new technology]?"
- "How would this integrate with our current stack?"
- "Can we build a quick proof-of-concept?"
- "What does our architecture look like in 2 years?"

**Not for**: Production feature development or immediate shipping - this team explores and validates future bets.

## Quick Start

```bash
/rite rnd
```

## Agents

| Agent | Role | Artifact |
|-------|------|----------|
| technology-scout | Watches technology horizon for opportunities | tech-assessment |
| integration-researcher | Maps how new tech integrates with existing systems | integration-map |
| prototype-engineer | Builds throwaway code that enables decisions | prototype |
| moonshot-architect | Designs systems for futures that haven't happened | moonshot-plan |
| tech-transfer | Bridges exploration to production with gap analysis | TRANSFER, HANDOFF |

## Workflow

See `workflow.md` for phase flow and complexity levels.

**Complexity Levels**:
- **SPIKE**: Quick feasibility check, single technology
- **EVALUATION**: Full technology evaluation with integration analysis
- **MOONSHOT**: Paradigm shift exploration, multi-year architecture

## When to Use RND vs 10x /spike

**RND Pack** is for exploring the unknown:
- Multi-session, learning-focused research
- Outcomes are knowledge and recommendations, not decisions
- "Can we even do this?" questions
- Technology scouting with unknown integration complexity
- Paradigm shifts and multi-year architecture exploration

**10x /spike** is for evaluating known options:
- Time-boxed (single session, typically hours)
- Outcome is a decision, not ongoing research
- "Which of these should we use?" questions
- Comparing specific technologies or approaches
- Quick feasibility checks with clear success criteria

### Decision Guide

| Scenario | Use |
|----------|-----|
| "Should we use React or Vue?" | 10x `/spike` |
| "Can we build an AI that understands legal documents?" | RND |
| "Which payment provider: Stripe or Square?" | 10x `/spike` |
| "How would quantum computing change our architecture?" | RND |
| "Is this library suitable for our caching layer?" | 10x `/spike` |
| "What would our system look like with event sourcing?" | RND |

**Rule of thumb**: If you can answer it in one focused session, use `/spike`. If you need to learn, experiment, and iterate across multiple sessions, use RND.

## Architecture Strategy Integration

**Workflow**: moonshot-architect (RND) -> roadmap-strategist (Strategy)

When moonshot-architect produces a MOONSHOT artifact with significant business implications, it triggers a handoff to strategy for roadmap integration.

**Trigger criteria**:
- Artifact type: MOONSHOT (from moonshot-architect)
- Business impact assessment: `significant` or `transformational`
- Technical feasibility: validated via prototype or analysis

**Handoff process**:
1. moonshot-architect completes MOONSHOT artifact with business impact assessment
2. If `business_impact > "significant"`, initiate handoff to strategy
3. roadmap-strategist receives MOONSHOT artifact and evaluates:
   - Strategic alignment with company vision
   - Timeline and resource implications
   - Market timing considerations
4. Output: Updated strategic-roadmap incorporating moonshot initiatives

**Example trigger**:
```
MOONSHOT artifact:
  title: "Event Sourcing Architecture Migration"
  business_impact: "transformational"
  feasibility: "validated"
  -> Triggers handoff to roadmap-strategist
```

See also: [strategy README](../strategy/README.md#architecture-strategy-integration)

## Related Rites

- **10x-dev**: Hand off validated prototypes for production implementation. RND produces TRANSFER artifacts (exploration findings, prototype specs, evaluation results) which the 10x-dev requirements-analyst consumes to create production PRDs.
- **strategy**: Hand off long-term architecture plans for strategic planning. Moonshots with `business_impact > "significant"` trigger roadmap-strategist integration.

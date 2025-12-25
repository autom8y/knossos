# R&D Pack

Technology scouting, integration analysis, prototyping, and future architecture.

## When to Use This Team

**Triggers**:
- "Should we evaluate [new technology]?"
- "How would this integrate with our current stack?"
- "Can we build a quick proof-of-concept?"
- "What does our architecture look like in 2 years?"

**Not for**: Production feature development or immediate shipping - this team explores and validates future bets.

## Quick Start

```bash
/team rnd-pack
```

## Agents

| Agent | Role | Model | Artifact |
|-------|------|-------|----------|
| technology-scout | Watches technology horizon for opportunities | claude-sonnet-4-5 | tech-assessment |
| integration-researcher | Maps how new tech integrates with existing systems | claude-sonnet-4-5 | integration-map |
| prototype-engineer | Builds throwaway code that enables decisions | claude-sonnet-4-5 | prototype |
| moonshot-architect | Designs systems for futures that haven't happened | claude-opus-4-5 | moonshot-plan |

## Workflow

See `workflow.md` for phase flow and complexity levels.

**Complexity Levels**:
- **SPIKE**: Quick feasibility check, single technology
- **EVALUATION**: Full technology evaluation with integration analysis
- **MOONSHOT**: Paradigm shift exploration, multi-year architecture

## Related Teams

- **ship-pack**: Hand off validated prototypes for production implementation
- **strategy-pack**: Hand off long-term architecture plans for strategic planning

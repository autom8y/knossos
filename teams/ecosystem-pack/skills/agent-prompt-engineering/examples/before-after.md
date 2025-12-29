# Before/After Transformation Examples

> Real agent prompt improvements with annotated changes

These examples demonstrate prompt optimization patterns from production agent audits. Each example shows a specific transformation with before/after comparison and key improvements.

## Examples

| Example | Focus Area | Score Improvement |
|---------|------------|-------------------|
| [Orchestrator Clarification](orchestrator-clarification.md) | Role identity, constraint placement | 3.7 -> 4.5 |
| [Handoff Criteria](handoff-criteria.md) | Objective vs subjective criteria | 3.5 -> 4.8 |
| [Token Efficiency](token-efficiency.md) | Skill extraction, deduplication | 3.0 -> 5.0 |
| [Example Quality](example-quality.md) | Workflow-specific examples | 3.0 -> 4.5 |

## Transformation Summary

| Agent | Before | After | Lines Saved | Key Change |
|-------|--------|-------|-------------|------------|
| Orchestrator | 291 lines | 185 lines | 106 (36%) | Front-loaded critical constraint |
| Requirements Analyst | 167 lines | 155 lines | 12 (7%) | Objective handoff criteria |
| All 5 agents | 125 lines | 5 lines | 120 (96%) | Shared file verification skill |
| Principal Engineer | 294 lines | 195 lines | 99 (34%) | Workflow-specific examples |

**Total optimization**: 337 lines saved across team while improving clarity.

## Applying These Patterns

When optimizing agent prompts:

1. **Check role identity**: Is it clear in first 2 sentences?
2. **Front-load constraints**: Critical rules appear before context
3. **Objectify criteria**: Replace subjective words with measurable conditions
4. **Extract shared content**: If 3+ agents share text, create a skill
5. **Contextualize examples**: Examples should show this agent's specific workflow

For the full rubric and validation checklist, see [scoring/rubric.md](../scoring/rubric.md) and [validation/checklist.md](../validation/checklist.md).

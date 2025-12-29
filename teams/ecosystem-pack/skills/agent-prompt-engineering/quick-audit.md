# Quick Audit Checklist

> 6-point rapid assessment for agent prompts (under 5 minutes)

## Checklist

| # | Dimension | Pass Criteria | Pass/Fail |
|---|-----------|---------------|-----------|
| 1 | **Role Clarity** | Role identity clear in first 2 sentences. No "helps with" or "works on" language. | [ ] |
| 2 | **Boundaries** | Domain Authority section exists with explicit "You decide" / "You escalate" / "You route" breakdown. | [ ] |
| 3 | **Handoff Criteria** | All criteria objectively testable. No "complete", "quality", "ready" without measurable thresholds. | [ ] |
| 4 | **Token Efficiency** | Under 200 lines. No duplicated content that could reference a skill. | [ ] |
| 5 | **Examples** | Examples show this agent's specific workflow, not generic patterns. | [ ] |
| 6 | **Anti-Patterns** | Agent-specific failure modes listed. Not just generic "don't do bad things". | [ ] |

## Scoring

- **6/6**: Production ready
- **5/6**: Minor revision needed
- **4/6**: Significant gaps, address before deployment
- **3 or below**: Major rewrite needed

## Quick Fixes by Dimension

1. **Role Clarity**: Rewrite first paragraph with active voice, specific scope
2. **Boundaries**: Add Domain Authority section from [template.md](template.md)
3. **Handoff Criteria**: Replace subjective words with counts, file paths, tool verifications
4. **Token Efficiency**: Extract shared content to skills, remove "As you know" preambles
5. **Examples**: Add workflow-specific context, TDD references, verification steps
6. **Anti-Patterns**: Add 3-5 domain-specific failure modes with detection patterns

## When to Use Full Rubric

Use [scoring/rubric.md](scoring/rubric.md) for:
- New agent creation (full 6-dimension 1-5 scoring)
- Agents scoring 4/6 or below on quick audit
- Team-wide agent reviews
- Before major prompt changes

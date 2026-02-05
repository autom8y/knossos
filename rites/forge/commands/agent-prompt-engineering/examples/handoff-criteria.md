# Requirements Analyst Handoff Criteria

> Part of [agent-prompt-engineering](../SKILL.md) skill examples

**Problem**: Subjective handoff criteria. Agent cycled indefinitely on "quality" judgment.

## Before (Score: 3.5/5)

```markdown
## Handoff Criteria

Ready for Architect when:
- [ ] Requirements are complete
- [ ] Quality meets standards
- [ ] Document is ready for review
- [ ] All necessary information included
- [ ] Work is finished
```

**Issues annotated**:
- "complete", "quality", "ready", "necessary", "finished" are all subjective
- Agent cannot objectively determine when to stop
- Different interpretations lead to premature or never-ending work

## After (Score: 4.8/5)

```markdown
## Handoff Criteria

Ready for Architect when:
- [ ] PRD contains 3+ acceptance criteria per user story
- [ ] All external dependencies listed with version constraints
- [ ] Rollback/recovery procedure documented
- [ ] No TODO or TBD markers remain in document
- [ ] User-facing text reviewed (no placeholder copy)
- [ ] PRD validated against PRD template (all sections present)
- [ ] File written to `.claude/artifacts/PRD-{id}.md` and verified via Read tool
```

## Key Improvements

- Each criterion is objectively testable
- Numeric thresholds where applicable (3+ criteria)
- File path explicit, verification method stated
- No subjective language

**Impact**: Agent now signals completion correctly. No more indefinite cycling.

---
name: wrap-ref
description: "Complete and finalize work session with validation and summary. Use when: all session goals achieved, work is ready for production, session scope exhausted, formal closure needed. Triggers: /wrap, finish session, complete session, finalize work, end session."
---

# /wrap - Complete and Finalize Session

> Validate quality gates, generate summary, archive session.

## Decision Tree

```
Finishing work?
├─ All goals achieved → /wrap
├─ Session parked → /resume first, then /wrap
├─ Want to pause (not finish) → /park
├─ Quality issues exist → Fix first, or /wrap --skip-checks
└─ Spike/prototype work → /wrap --skip-checks
```

## Usage

```bash
/wrap [--skip-checks] [--archive]
```

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `--skip-checks` | No | false | Skip quality gate validation |
| `--archive` | No | true | Archive SESSION_CONTEXT |

## Quality Gates

| Gate | Applies To | Checks |
|------|-----------|--------|
| PRD | All | File exists, sections complete, criteria testable |
| TDD | MODULE+ | Traces to PRD, ADRs exist, interfaces defined |
| Code | Implementation phase | Git clean, tests pass, types/lint clean |
| Validation | QA phase | Test plan exists, criteria validated, defects resolved |

## Quick Reference

**Pre-flight**: Session exists, not parked

**Actions**:
1. Run quality gates (unless --skip-checks)
2. Offer final QA review (if last_agent ≠ qa)
3. Generate session summary
4. Archive SESSION_CONTEXT
5. Update session index
6. Display confirmation

**Creates**:
- `/docs/sessions/SUMMARY-{session_id}.md`
- `/docs/sessions/INDEX.md` (if first wrap)

**Removes**:
- `.claude/sessions/{session_id}/` (moved to archive)

## Anti-Patterns

| Do NOT | Why | Instead |
|--------|-----|---------|
| Use --skip-checks habitually | Defeats quality gates | Only for spikes/emergencies |
| Wrap mid-implementation | Incomplete artifacts | Park if pausing, wrap when done |
| Ignore quality gate failures | Technical debt ships | Fix issues, then wrap |
| Wrap parked session | Park conflicts with completion | Resume first |

## Prerequisites

- Active session exists
- Quality gates passing (unless --skip-checks)

## Success Criteria

- Quality gates pass (or explicitly skipped)
- Session summary generated
- SESSION_CONTEXT archived
- Session index updated
- User receives next steps

## Related Commands

| Command | When to Use |
|---------|-------------|
| `/start` | Begin new session after wrap |
| `/park` | Pause instead of wrap |
| `/resume` | Required if currently parked |
| `/handoff` | Fix issues before wrap |

## Progressive Disclosure

- [behavior.md](behavior.md) - Full step-by-step sequence
- [quality-gates.md](quality-gates.md) - All gate definitions
- [examples.md](examples.md) - Usage scenarios
- [session-summary.md](session-summary.md) - Summary template
- [session-context-schema](../session-common/session-context-schema.md) - Field definitions

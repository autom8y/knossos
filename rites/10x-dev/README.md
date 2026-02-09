# 10x Dev Rite

Full-lifecycle software development from product requirements through quality validation.

## When to Use This Rite

**Triggers**:
- "Build a new feature end-to-end"
- "We need a PRD and technical design before starting"
- "Take this from requirements to tested, production-ready code"
- "Implement this API/service/module with proper testing"

**Not for**: Documentation work, infrastructure automation, one-off scripts without testing requirements

## Quick Start

```bash
/task build payment processing system with Stripe integration
# or invoke directly
/rite 10x-dev
```

## Agents

| Agent | Role | Artifact |
|-------|------|----------|
| requirements-analyst | Transform ambiguity into actionable specs | PRD |
| architect | Evaluate tradeoffs, produce design decisions | TDD, ADRs |
| principal-engineer | Build production-grade, tested code | Code, tests |
| qa-adversary | Break things before users do | Test plan, defects |

## Workflow

See `workflow.md` for phase flow and complexity levels.

**Complexity levels**:
- SCRIPT: Single file, <200 LOC (skips design phase)
- MODULE: Multiple files, <2000 LOC
- SERVICE: APIs, persistence
- PLATFORM: Multi-service

## When to Use 10x /spike vs RND

**10x /spike** is for evaluating known options:
- Time-boxed exploration (single session, typically hours)
- Outcome is a decision or recommendation
- Comparing specific technologies or approaches
- "Which of these options should we choose?"
- Clear success criteria and evaluation rubric

**RND Pack** is for exploring the unknown:
- Multi-session research with learning-focused outcomes
- No clear answer exists yet—you're discovering what's possible
- "Can we even do this?" questions
- Technology scouting and future architecture exploration
- Outcomes are knowledge, prototypes, and recommendations

### Decision Guide

| Scenario | Use |
|----------|-----|
| "Should we use React or Vue?" | `/spike` |
| "Can we build an AI that understands legal documents?" | RND |
| "Which payment provider: Stripe or Square?" | `/spike` |
| "How would quantum computing change our architecture?" | RND |
| "Is this library suitable for our caching layer?" | `/spike` |
| "What would our system look like with event sourcing?" | RND |

**Rule of thumb**: If you can answer it in one focused session with a decision at the end, use `/spike`. If you need to learn, experiment, and iterate across multiple sessions, switch to `rnd`.

## Related Rites

- **docs**: When implementation is complete and needs documentation
- **sre**: When deployment automation is required
- **rnd**: When exploration scope exceeds a single session

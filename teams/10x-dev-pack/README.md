# 10x Dev Pack

Full-lifecycle software development from product requirements through quality validation.

## When to Use This Team

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
/team 10x-dev-pack
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

## Related Teams

- **doc-team-pack**: When implementation is complete and needs documentation
- **infrastructure-pack**: When deployment automation is required

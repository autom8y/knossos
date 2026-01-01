---
name: hotfix-ref
description: "Rapid fix workflow for urgent production issues. Use when: production is broken, time-critical fix needed, emergency deployment required. Triggers: /hotfix, urgent fix, production issue, emergency fix, quick fix."
---

# /hotfix - Rapid Fix Workflow

> Execute rapid fixes for urgent production issues, skipping heavyweight documentation for speed.

## Decision Tree

```
Production issue?
├─ Production broken → /hotfix
├─ Time critical (< 1h) → /hotfix
├─ Urgent but can wait → /task
├─ Need to investigate → /spike
└─ New feature → /task or /sprint
```

## Usage

```bash
/hotfix "issue-description" [--severity=LEVEL]
```

### Parameters

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `issue-description` | Yes | - | What's broken and needs fixing |
| `--severity` | No | HIGH | CRITICAL \| HIGH \| MEDIUM |

## Quick Reference

**Pre-flight**:
- Production broken or severely degraded
- Users actively impacted
- Clear problem with focused solution
- Time-sensitive (< 60 min)

**Actions**:
1. Assess severity (CRITICAL/HIGH/MEDIUM)
2. Skip PRD (intentional for speed)
3. Minimal TDD (only if CRITICAL + architectural)
4. Diagnose → Fix → Test (Principal Engineer)
5. Fast QA validation
6. Ship with rollback plan

**Produces**:
- Fix implementation
- Minimal tests (critical paths)
- Rollback plan
- Inline documentation
- `/docs/design/HOTFIX-{slug}.md` (only if CRITICAL)

**Never Produces**:
- PRD (too slow)
- Full TDD (unless CRITICAL)
- Comprehensive test suite (defer to follow-up)

## Anti-Patterns

| Do NOT | Why | Instead |
|--------|-----|---------|
| Hotfix non-urgent issues | Wastes emergency process | Use `/task` |
| Hotfix without rollback plan | Risky deployment | Document rollback first |
| Skip follow-up task | Tech debt accumulates | Create `/task` for proper fix |
| Exceed time budget | Defeats purpose | Abort, escalate to `/task` |
| Hotfix unclear problems | Need investigation | Use `/spike` first |

## Prerequisites

- 10x-dev-pack active (minimum: Engineer + QA agents)
- Clear understanding of the issue
- Access to production logs/monitoring
- Time pressure (< 60 min fix target)

## Success Criteria

- Issue resolved in < 60 minutes
- Critical paths tested and working
- Rollback plan documented
- Code committed with clear message
- Follow-up task created (if needed)

## Related Commands

| Command | When to Use |
|---------|-------------|
| `/task` | Proper fix with full workflow (use for follow-up) |
| `/spike` | Research if root cause unclear |
| `/sprint` | Multi-fix coordination |
| `/pr` | Create PR after hotfix complete |

## Progressive Disclosure

- [behavior.md](behavior.md) - Full step-by-step sequence, time budgets, agent templates
- [examples.md](examples.md) - 3 severity scenarios, edge cases, commit message template
- [../shared-sections/time-boxing.md](../shared-sections/time-boxing.md) - Severity-based time limits
- [../shared-sections/agent-invocation.md](../shared-sections/agent-invocation.md) - Principal Engineer, QA delegation

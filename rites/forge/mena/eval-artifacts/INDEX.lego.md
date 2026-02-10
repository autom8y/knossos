---
name: eval-artifacts
description: "Validation checklists and eval report template for rite and agent evaluation. Use when: running agent completeness checks, validating workflow structure, preparing adversarial test prompts, writing eval reports. Triggers: eval checklist, agent completeness, workflow validity, adversarial prompts, eval report template, validation harness."
---

# Eval Artifacts

> Validation checklists, adversarial prompt banks, and eval report template for the Eval Specialist.

## Purpose

Provides structured validation harnesses for rite evaluation: agent completeness checks, workflow validity rules, adversarial prompt banks for stress testing, and the canonical eval-report.md template.

## Contents

| Resource | Purpose |
|----------|---------|
| [Eval Report Template](#eval-report-template) | Canonical eval-report.md structure |
| [Agent Completeness Checklist](#agent-completeness-checklist) | Per-agent structure and content checks |
| [Workflow Validity Checklist](#workflow-validity-checklist) | workflow.yaml schema and logic rules |
| [Adversarial Prompt Bank](#adversarial-prompt-bank) | Edge case, boundary, and error prompts |

## Eval Report Template

```markdown
# Eval Report: {rite-name}

**Date**: {timestamp}
**Status**: {PASS | FAIL | PASS WITH WARNINGS}

## Summary
- Total checks: {N}
- Passed: {N}
- Failed: {N}
- Warnings: {N}

## Structure Validation
- [ ] All agent files exist
- [ ] YAML frontmatter valid
- [ ] All 11 sections present
- [ ] workflow.yaml exists

## Schema Validation
- [ ] Frontmatter fields complete
- [ ] Description has triggers and examples
- [ ] Workflow has required fields
- [ ] Phase chain complete

## Logic Validation
- [ ] Single terminal phase
- [ ] All phases reachable
- [ ] Complexity levels valid
- [ ] Agent names match files

## Adversarial Testing
- [ ] Edge cases handled
- [ ] Boundary cases handled
- [ ] Error handling appropriate
- [ ] Handoffs trigger correctly

## Issues Found

### Blocking
{List of blocking issues or "None"}

### Warnings
{List of non-blocking issues or "None"}

## Recommendation
{SHIP | DO NOT SHIP | SHIP WITH CAVEATS}

{Explanation of recommendation}
```

## Agent Completeness Checklist

Per-agent validation (run for each agent in the rite):

```
[ ] name field present (kebab-case)
[ ] description is multi-line with:
    [ ] Role summary
    [ ] Trigger conditions
    [ ] What it produces
    [ ] <example> block
[ ] tools field present
[ ] model field present (opus/sonnet/haiku)
[ ] color field present

[ ] Section 1: Title and Overview (2-3 sentences)
[ ] Section 2: Core Responsibilities (4-6 bullets)
[ ] Section 3: Position in Workflow (ASCII diagram)
[ ] Section 4: Domain Authority (decide/escalate/route)
[ ] Section 5: How You Work (3-4 phases)
[ ] Section 6: What You Produce (artifact table)
[ ] Section 7: Handoff Criteria (checklist)
[ ] Section 8: The Acid Test (pivotal question)
[ ] Section 9: Skills Reference (cross-refs)
[ ] Section 10: Cross-Rite Notes (flags)
[ ] Section 11: Anti-Patterns (3-5 items)
```

## Workflow Validity Checklist

```
[ ] name matches rite directory
[ ] workflow_type is "sequential"
[ ] description present
[ ] entry_point.agent exists in agents/
[ ] entry_point.artifact has type and path_template
[ ] Each phase has: name, agent, produces, next
[ ] Exactly one phase has next: null
[ ] All phases reachable from entry
[ ] No orphan phases
[ ] complexity_levels present (2-4 levels)
[ ] Each level has: name, scope, phases
[ ] Phase references in levels are valid
```

## Adversarial Prompt Bank

### Edge Cases
- "Do everything at once"
- "I'm not sure what I want"
- "This is urgent, skip the planning"
- "Can you also handle {unrelated task}?"

### Boundary Cases
- Minimal input: just a rite name
- Maximum scope: platform-level complexity
- Conflicting requirements

### Error Handling
- Invalid rite name
- Missing required artifacts
- Circular dependencies requested

---
description: "Audit Report Template companion for templates skill."
---

# Audit Report Template

> Quality signoff for completed refactoring work.

```markdown
# Audit Report
**Initiative**: [Brief title of refactoring work]
**Audited**: [date]
**Auditor**: [agent/person name]

## Executive Summary
- Commits audited: [count]
- Test status: [passing/failing]
- Regressions found: [count]
- Quality verdict: [APPROVED|APPROVED_WITH_NOTES|REJECTED]

## Scope
**Repository**: [name]
**Branch**: [branch name]
**Commit range**: [start..end]
**Files changed**: [count]

## Audit Criteria
- [ ] All tests passing
- [ ] Each commit is atomic and revertible
- [ ] No behavior changes (refactoring only)
- [ ] Code quality improved (no new smells)
- [ ] Commit messages reference task IDs
- [ ] No regressions in performance or functionality

## Findings

### Commits Reviewed
| Commit | Task | Atomic | Tests | Notes |
|--------|------|--------|-------|-------|
| [hash] | [RF-001] | ✓ | ✓ | [any notes] |
| [hash] | [RF-002] | ✓ | ✓ | [any notes] |

### Regressions Detected
[List any regressions found, or "None" if clean]

**REG-001: [Description]**
- **Severity**: [Critical|High|Medium|Low]
- **Location**: [file:line]
- **Evidence**: [what broke or degraded]
- **Root cause**: [commit that introduced it]

### Quality Improvements
[Quantify improvements from refactoring]

**Before**:
- [Metric 1]: [value]
- [Metric 2]: [value]

**After**:
- [Metric 1]: [value]
- [Metric 2]: [value]

### Code Smells Remaining
[Any smells that weren't addressed or new ones introduced]

## Deviations from Plan
[Any cases where execution differed from refactoring plan]

**DEV-001: [Description]**
- **Justification**: [why it was necessary]
- **Impact**: [what changed]
- **Approval**: [architect-enforcer notified: yes/no]

## Test Coverage
- Total tests: [count]
- Passing: [count]
- Failing: [count]
- Skipped: [count]
- Coverage delta: [±X%]

## Performance Impact
[Any performance changes observed]

- Build time: [before] → [after]
- Test run time: [before] → [after]
- Binary size: [before] → [after]

## Rollback Assessment
[Verification that commits are independently revertible]

- Rollback points documented: [yes/no]
- Tested random commit revert: [yes/no]
- Revert left codebase valid: [yes/no]

## Verdict

**[APPROVED | APPROVED_WITH_NOTES | REJECTED]**

**Rationale**: [Why this verdict was reached]

**Conditions** (if approved with notes):
- [Condition 1]
- [Condition 2]

**Required fixes** (if rejected):
- [Fix 1]
- [Fix 2]

## Recommendations

### For This Initiative
- [Recommendation 1]
- [Recommendation 2]

### For Future Work
- [Pattern to continue]
- [Pattern to avoid]
- [Process improvement]

## Sign-off
**Audit Lead**: [name]
**Date**: [date]
**Next Steps**: [what happens next]
```

## Quality Gate

**Audit Report complete when:**
- All commits reviewed for atomicity
- Test suite run and status documented
- Any regressions clearly identified with severity
- Quality improvements quantified
- Verdict justified with clear rationale
- Rollback viability assessed

---
description: "Resilience Report Template companion for templates skill."
---

# Resilience Report Template

> Aggregate assessment of system resilience across multiple chaos experiments.

```markdown
# Resilience Report: [System/Service]

## Executive Summary
[One paragraph: Overall resilience posture, critical findings, top recommendations]

## Scope
- Services tested: [list]
- Time period: [dates]
- Environments: [dev/staging/prod]
- Experiment count: [number]

## Experiments Summary
| Experiment | Target | Result | Critical Findings |
|------------|--------|--------|-------------------|
| [name] | [service] | [PASS/FAIL] | [findings] |

## Resilience Scorecard
| Capability | Status | Evidence |
|------------|--------|----------|
| Database failover | [PASS/FAIL] | [experiment ref] |
| Circuit breakers | [PASS/FAIL] | [experiment ref] |
| Graceful degradation | [PASS/FAIL] | [experiment ref] |
| Auto-recovery | [PASS/FAIL] | [experiment ref] |
| Rollback procedures | [PASS/FAIL] | [experiment ref] |

## Critical Gaps
| Gap | Impact | Priority | Remediation |
|-----|--------|----------|-------------|
| [gap] | [impact] | [P1/P2/P3] | [fix] |

## Recommendations

### Immediate (This Week)
1. [Action]: [Expected improvement]

### Short-Term (This Month)
1. [Action]: [Expected improvement]

### Long-Term (This Quarter)
1. [Action]: [Expected improvement]

## Failure Mode Catalog
| Mode | Detection | Impact | Mitigation |
|------|-----------|--------|------------|
| [failure] | [how detected] | [blast radius] | [how to mitigate] |

## Next Steps
1. [Immediate action]
2. [Follow-up experiments]
3. [Remediation tracking]
```

## Quality Gate

**Resilience Report complete when:**
- All experiments referenced with outcomes
- Scorecard covers core resilience capabilities
- Gaps have priority and remediation plans
- Failure mode catalog populated from experiment evidence

---
description: "Chaos Experiment Template companion for templates skill."
---

# Chaos Experiment Template

> Pre-registration for chaos engineering tests with hypothesis, blast radius, and abort criteria.

```markdown
# Chaos Experiment: [Name]

## Metadata
- **Date**: [execution date]
- **Target**: [service/system]
- **Environment**: [dev/staging/prod]
- **Engineer**: [name]

## Hypothesis
**Given**: [steady state description]
**When**: [failure condition]
**Then**: [expected behavior]

## Steady State Definition
| Metric | Normal Range | Measurement |
|--------|--------------|-------------|
| Request rate | [range] | [source] |
| Error rate | [range] | [source] |
| Latency p99 | [range] | [source] |

## Experiment Design

### Failure Type
[Network / Process / Resource / Dependency]

### Injection Method
```
[How failure will be introduced - e.g., toxiproxy, tc, kill -9]
```

### Blast Radius
- **Scope**: [% of traffic / # of instances]
- **Duration**: [time]
- **Affected Users**: [estimate]

### Abort Criteria
- Error rate > [threshold]
- Latency p99 > [threshold]
- [Other conditions]

### Rollback Plan
```
[How to remove the failure condition]
```

## Execution Log
| Time | Action | Observation |
|------|--------|-------------|
| [time] | [action] | [what happened] |

## Results

### Outcome
**[PASS / PARTIAL / FAIL / ABORT]**

### Observations
[What actually happened vs. hypothesis]

### Gaps Discovered
1. [Gap]: [Impact] → [Recommendation]

### Evidence
[Links to dashboards, logs, screenshots]

## Action Items
| Action | Owner | Priority | Due |
|--------|-------|----------|-----|
| [action] | [name] | [P1/P2/P3] | [date] |

## Lessons Learned
[What did we learn that applies beyond this experiment?]
```

## Quality Gate

**Chaos Experiment complete when:**
- Hypothesis follows Given/When/Then format
- Steady state metrics have baselines
- Abort criteria defined before execution
- Rollback plan tested independently

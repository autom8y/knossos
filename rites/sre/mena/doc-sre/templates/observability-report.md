# Observability Report Template

> Audit current monitoring, logging, tracing, and alerting coverage for a system or service.

```markdown
# Observability Report: [System/Service]

## Executive Summary
[One paragraph: Current state, critical gaps, top recommendations]

## Scope
- Services analyzed: [list]
- Time period: [dates]
- Data sources: [metrics/logs/traces systems]

## Current State

### Metrics
| Service | Golden Signals | Custom Metrics | Gaps |
|---------|----------------|----------------|------|
| [name]  | [coverage %]   | [count]        | [list] |

### Logging
| Service | Structured | Correlation IDs | Retention |
|---------|------------|-----------------|-----------|
| [name]  | [yes/no]   | [yes/no]        | [days]    |

### Tracing
| Service | Instrumented | Sample Rate | Coverage |
|---------|--------------|-------------|----------|
| [name]  | [yes/no]     | [%]         | [%]      |

### Alerting
| Alert Category | Count | False Positive Rate | Actions |
|----------------|-------|---------------------|---------|
| Critical       | [n]   | [%]                 | [types] |

## Gap Analysis

### Critical Gaps (Must Fix)
1. [Gap]: [Impact] → [Recommendation]

### Important Gaps (Should Fix)
1. [Gap]: [Impact] → [Recommendation]

### Nice-to-Have Improvements
1. [Improvement]: [Benefit]

## Recommendations

### Quick Wins (< 1 week)
1. [Action]: [Expected outcome]

### Medium-Term (1-4 weeks)
1. [Action]: [Expected outcome]

### Long-Term (> 1 month)
1. [Action]: [Expected outcome]

## SLI/SLO Proposals

| Service | SLI | Current | Proposed SLO | Error Budget |
|---------|-----|---------|--------------|--------------|
| [name]  | [availability] | [%] | [%] | [hours/month] |

## Next Steps
1. [Immediate action]
2. [Follow-up]
```

## Quality Gate

**Observability Report complete when:**
- All four pillars assessed (metrics, logging, tracing, alerting)
- Gaps prioritized by severity
- SLI/SLO proposals include baselines and error budgets
- Recommendations have clear timelines

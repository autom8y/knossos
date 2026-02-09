# Insights Report Template

> Data-driven decision recommendation with segment analysis, alternative explanations, and ship/no-ship verdict.

```markdown
# INSIGHT-{slug}

## Executive Summary
{2-3 sentence summary: what we learned and what to do}

## Decision Recommendation
**Recommendation**: {Ship / Don't Ship / Iterate / More Testing}
**Confidence**: {High / Medium / Low}

## Background
{Context for this analysis}

## Key Findings

### Finding 1: {Headline}
**Impact**: {High/Medium/Low}
**Confidence**: {High/Medium/Low}

{Detailed explanation with data}

| Metric | Control | Treatment | Change | Significance |
|--------|---------|-----------|--------|--------------|
| {metric} | {value} | {value} | {+/-X%} | {p-value} |

**Interpretation**: {What this means}

### Finding 2: {Headline}
...

## Segment Analysis

### {Segment 1}
| Metric | Effect | vs Overall |
|--------|--------|------------|
| {metric} | {+/-X%} | {better/worse/same} |

**Interpretation**: {Why this segment differs}

### {Segment 2}
...

## Supporting Evidence

### Quantitative
{Data points supporting conclusions}

### Qualitative
{User research supporting conclusions}

## Alternative Explanations
{Hypotheses we considered and ruled out}

## Limitations
- {Limitation 1}
- {Limitation 2}

## Recommendations

### Primary Recommendation
{What to do, with rationale}

### If We Ship
1. {Follow-up action}
2. {Monitoring plan}

### If We Don't Ship
1. {Alternative approach}
2. {What to test next}

## Open Questions
{What we still don't know}

## Appendix
- Raw data
- Methodology details
- Statistical analysis
```

## Quality Gate

**Insights Report complete when:**
- Decision recommendation has confidence level
- Findings include statistical significance
- Alternative explanations considered and ruled out
- Both ship and no-ship paths documented

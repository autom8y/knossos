# Experiment Design Template

> A/B test pre-registration with hypothesis, sample size calculation, metrics, and stopping rules.

```markdown
# EXPERIMENT-{slug}

## Executive Summary
{What we're testing and why in 2-3 sentences}

## Hypothesis

### Belief
{The intuition or assumption we're testing}

### Null Hypothesis (H0)
{What we assume is true if treatment has no effect}

### Alternative Hypothesis (H1)
{What we expect if treatment works}

### Expected Effect Size
{Minimum detectable effect we care about}

## Experiment Design

### Type
{A/B test, multivariate, bandit, etc.}

### Variants
| Variant | Description | Traffic |
|---------|-------------|---------|
| Control | {Current experience} | {50%} |
| Treatment | {New experience} | {50%} |

### Randomization
- **Unit**: {User, session, device, etc.}
- **Stratification**: {Any stratification variables}

### Sample Size
- **MDE**: {X%} change in primary metric
- **Power**: {80/90}%
- **Significance**: {alpha = 0.05}
- **Required N per variant**: {calculated sample size}

### Duration
- **Minimum**: {X days based on sample size}
- **Recommended**: {X days including weekly cycles}
- **Maximum**: {X days before novelty effects concern}

## Metrics

### Primary Metric
- **Metric**: {Conversion rate, revenue, etc.}
- **Baseline**: {Current value}
- **Target**: {Expected improvement}
- **Success Threshold**: {Minimum to ship}

### Secondary Metrics
| Metric | Baseline | Expected Direction |
|--------|----------|-------------------|
| {metric} | {value} | {up/down/flat} |

### Guardrail Metrics
| Metric | Threshold | Action if Crossed |
|--------|-----------|-------------------|
| {metric} | {limit} | {stop/investigate} |

## Segments
{Key segments to analyze}
- {Segment 1}
- {Segment 2}

## Risks and Mitigations
| Risk | Mitigation |
|------|------------|
| {risk} | {mitigation} |

## Early Stopping Rules
- Stop for harm: {conditions}
- Stop for success: {conditions}
- Continue regardless: {conditions}

## Pre-Registration
This document serves as pre-registration. Analysis will follow this plan.

## Timeline
- **Design Complete**: {date}
- **Launch**: {date}
- **Minimum Run Time**: {date}
- **Analysis**: {date}

## Success Criteria
{When do we ship the treatment?}

## Post-Experiment
- [ ] Results document
- [ ] Decision recorded
- [ ] Learnings documented
```

## Quality Gate

**Experiment Design complete when:**
- Hypotheses stated in null/alternative form
- Sample size calculated with MDE, power, and significance
- Guardrail metrics defined with stopping rules
- Pre-registration locked before launch

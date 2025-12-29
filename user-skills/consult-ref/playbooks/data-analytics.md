# Playbook: Data Analytics & Experimentation

> From tracking instrumentation to A/B test insights

## When to Use

- Setting up analytics tracking for a feature
- Designing and running A/B tests
- Analyzing user behavior metrics
- Synthesizing experiment results into recommendations
- Establishing product instrumentation

## Prerequisites

- Clear hypothesis or question to answer
- Access to analytics infrastructure (or readiness to instrument)
- Stakeholder alignment on success metrics

## Command Sequence

### Phase 1: Initialize

```bash
/intelligence
```
**Expected output**: Team switched to intelligence-pack, roster displayed
**Decision point**: If simple metric lookup, may not need full workflow.

### Phase 2: Start Session

```bash
/start "Analytics goal" --complexity=METRIC
```
**Expected output**: Session created, context established
**Decision point**: Adjust complexity level:
- METRIC: Single metric or simple tracking
- FEATURE: Feature-level instrumentation and analysis
- INITIATIVE: Cross-functional analytics, major experiments

### Phase 3: Instrumentation

Analytics Engineer assesses tracking needs.

**Expected output**: tracking-plan artifact identifying events to capture
**Decision point**: Does existing tracking cover the need, or is new instrumentation required?

### Phase 4: Research (if qualitative needed)

```bash
/handoff user-researcher
```
**Expected output**: Research findings with qualitative insights
**Decision point**: Skip if purely quantitative analysis.

### Phase 5: Experimentation

```bash
/handoff experimentation-lead
```
**Expected output**: experiment-design with hypothesis, variants, sample size
**Decision point**: If no A/B test needed, skip to synthesis.

### Phase 6: Synthesis

```bash
/handoff insights-analyst
```
**Expected output**: insights-report with recommendations
**Decision point**: Review statistical significance before acting.

### Phase 7: Finalize

```bash
/wrap
```
**Expected output**: Session summary, decision recommendations

## Variations

- **METRIC complexity**: Skip experimentation, just analyze existing data
- **No instrumentation needed**: Skip Phase 3, use existing tracking
- **Qual + Quant**: Include user researcher in loop

## Success Criteria

- [ ] Tracking plan complete (if new instrumentation)
- [ ] Hypothesis tested with statistical rigor
- [ ] Insights synthesized into actionable recommendations
- [ ] Stakeholders can make data-driven decision

## Rollback

If experiment goes wrong:
```bash
/park                          # Preserve state
# Disable experiment variant
/continue                      # Reassess with control data
```

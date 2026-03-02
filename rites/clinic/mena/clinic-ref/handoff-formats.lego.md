---
name: clinic-ref-handoff-formats
description: "Cross-rite handoff artifact formats for clinic investigations. Use when: producing handoff-10x-dev.md, handoff-sre.md, or handoff-debt-triage.md after diagnosis. Triggers: handoff format, 10x-dev handoff, SRE handoff, debt triage handoff, treatment plan."
---

# Clinic: Cross-Rite Handoff Formats

Handoff artifacts are always recommendations. User decides whether to act.

## handoff-10x-dev.md

```markdown
# Handoff: 10x-dev
# Investigation: {slug}

## Root Cause Summary
{From diagnosis.md RC001/RC002 — not re-derived}

## Affected Files
- {specific file path}: {what to change}
- ...

## Fix Approach
{Recommended implementation strategy with rationale}

## Acceptance Criteria
{Specific, testable criteria for verifying the fix works}
- {e.g., checkout service returns 200 on /health after deploy}
- {e.g., error rate drops below 0.1% over 1 hour}

## Optional: Fix Ordering (compound failures)
{Which bug to fix first and why}

## Optional: Risk Assessment
{What could go wrong with the fix}

## Optional: Related Tests
{Existing tests to update or new tests to write}
```

## handoff-sre.md

```markdown
# Handoff: SRE
# Investigation: {slug}

## Signals Missing
{What monitoring or observability was absent that caused or prolonged this incident}

## Recommended Alerts
- Alert: {name}
  Condition: {specific trigger condition}
  Rationale: {would have caught this incident at {timestamp}}

## Recommended Dashboards
- Dashboard: {name}
  Panels: {what metrics/signals to show}
  Rationale: {would have aided debugging by showing X}

## Optional: SLO Impact
{How this incident affected service level objectives}

## Optional: Runbook Recommendation
{Incident response procedure for this failure class}
```

## handoff-debt-triage.md

```markdown
# Handoff: Debt Triage
# Investigation: {slug}

## Pattern Analysis
{How this issue manifests across the codebase — not just this instance}

## Scope of Problem
{How widespread the systemic issue is: N files, N services, estimated prevalence}

## Remediation Approach
{Long-term fix strategy, not just this instance}

## Optional: Debt Classification
{Type: architectural | dependency | test | design}

## Optional: Affected Components
{Broader list beyond the immediate investigation}

## Optional: Effort Estimate
{Rough sizing: S/M/L/XL}
```

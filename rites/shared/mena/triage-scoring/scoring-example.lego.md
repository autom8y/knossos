---
name: triage-scoring-example
description: "Worked scoring example for the Cassandra Protocol triage model. Use when: verifying your scoring arithmetic, understanding how zone overrides interact with thresholds, or calibrating scoring intuition against a concrete case. Triggers: scoring example, triage worked example, zone override example."
---

# Triage Scoring: Worked Example

## Standard Track Example

A complaint about a missing skill (`severity: medium`, `zone: behavior`, filed by 2 agents, 3 observations, no scar match, `effort_estimate: small`):

| Dimension | Raw Score | Weight | Contribution |
|-----------|-----------|--------|--------------|
| Severity | 45 | 25% | 11.25 |
| Recurrence | 65 | 20% | 13.00 |
| Zone Impact | 60 | 20% | 12.00 |
| Scar-Tissue Match | 20 | 15% | 3.00 |
| Effort-to-Impact | 75 | 10% | 7.50 |
| Source Diversity | 50 | 10% | 5.00 |
| **Total** | | | **51.75** |

Score 51.75 would normally auto-accept, but zone is `behavior` — zone override applies. Action: **human-review**.

## Quick-File Examples

A single `drift-detector` tool-fallback (`severity: low`, no zone/effort fields):

`(20 × 0.40) + (15 × 0.40) + (20 × 0.20) = 8 + 6 + 4 = 18` → **auto-reject**

A recurring tool-fallback (5+ observations, 2 filers):

`(20 × 0.40) + (90 × 0.40) + (50 × 0.20) = 8 + 36 + 10 = 54` → **auto-accept**

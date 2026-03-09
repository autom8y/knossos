---
name: thermia-ref
description: "Thermia rite methodology reference. Use when: coordinating consultation phases, applying the 6-gate decision framework, selecting cache patterns, understanding complexity levels, or routing back to earlier phases. Triggers: cache consultation, thermia orchestration, 6-gate framework, cache architecture, heat-mapper, complexity level, back-route, consultation resume."
---

# Thermia Methodology Reference

Cache architecture consultation lifecycle. Enforces "exhaust alternatives first."

```
assessment -> architecture -> specification -> validation
```

## Complexity

| Level | Phases |
|-------|--------|
| QUICK | assessment + validation (lite) |
| STANDARD / DEEP | all 4 phases |

## Phase Routing

| Specialist | Route When |
|------------|------------|
| heat-mapper | Always first |
| systems-thermodynamicist | Assessment complete, at least one CACHE verdict |
| capacity-engineer | Architecture complete (pattern + consistency + failure modes) |
| thermal-monitor | Specification complete, or assessment complete (QUICK) |

## Artifact Chain

All artifacts land in `.sos/wip/thermia/`:
`thermal-assessment.md` → `cache-architecture.md` → `capacity-specification.md` → `observability-plan.md`

## Cross-Rite

10x-dev (implementation), clinic (incidents), sre (monitoring), arch (architectural concerns)

## Companions

| Topic | File |
|-------|------|
| 6-gate framework (frequency, cost, staleness, UX, safety, scale) | `six-gate.md` |
| Cache patterns, consistency models, failure modes | `patterns.md` |
| Back-route protocols (assessment_gap, design_inconsistency) | `back-routes.md` |

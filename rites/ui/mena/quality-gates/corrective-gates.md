# Corrective Posture Quality Gates (QG-T1 through QG-T5)

Governing principle: "Eliminate imperfections before adding delight" (Lovin, asymmetric quality perception [MODERATE]).

Workflow: audit -> fix -> [impact at SYSTEM scope] -> validate

## Gate Definitions

| Gate ID | Dimension | Type | Phase | Criterion | Agent |
|---------|-----------|------|-------|-----------|-------|
| QG-T1 | D0: Accessibility | **Hard** | validate | WCAG 2.2 AA zero-tolerance across four layers (lint, axe-core, interaction testing, manual review protocol) | a11y-engineer |
| QG-T2 | D3: Consistency | Embedded | fix | Change follows existing patterns; no novel solutions unless existing patterns are demonstrably wrong | producing agent (self-check) |
| QG-T3 | D5: Edge state craft | Embedded | fix | Loading, empty, error, boundary states in modified component are intentionally handled | producing agent (self-check) |
| QG-T4 | D1: Micro-interactions | Advisory | validate | Animation timing and physics in modified interactions are preserved or improved | frontend-fanatic (if invoked) |
| QG-T5 | D2: Cognitive efficiency | Advisory | validate | No regression in interaction directness compared to pre-correction baseline | frontend-fanatic (if invoked) |

## Key Design Decisions

**QG-T1 (Hard gate)**: Only gate that blocks in corrective posture. All other gates are embedded or advisory. Corrective work is optimized for speed: audit, fix, validate accessibility, done.

**QG-T2, QG-T3 (Embedded)**: Self-check criteria by the producing agent (component-engineer or stylist). These are embedded in the phase exit criteria, not separate evaluation steps. Consistent with corrective posture's subtractive character -- fewer gates, smaller blast radius.

**QG-T4, QG-T5 (Advisory)**: Frontend-fanatic is NOT automatically invoked in corrective posture at COMPONENT scope. Optionally invokable at FEATURE and SYSTEM scope at potnia's discretion. Corrective work targets known imperfections; aesthetic evaluation is informative but not the purpose.

## Coverage Summary

| Dimension | Corrective Evaluation |
|-----------|----------------------|
| D0: Accessibility | Hard gate (a11y-engineer) |
| D1: Micro-interactions | Advisory (frontend-fanatic, discretionary) |
| D2: Cognitive efficiency | Advisory (frontend-fanatic, discretionary) |
| D3: Consistency | Embedded self-check |
| D4: Personality/Brand | Not evaluated (out of scope for corrective) |
| D5: Edge state craft | Embedded self-check |
| D6: Emotional design | Not evaluated (out of scope for corrective) |

Coverage: 5/7 dimensions. D4 and D6 are intentionally out of scope -- subtractive posture does not add personality or emotional design.

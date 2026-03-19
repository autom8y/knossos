# Generative Posture Quality Gates (QG-C1 through QG-C7)

Governing principle: "Design cannot be meaningfully critiqued in isolation from its functional context" (Lovin, holistic critique reframing [MODERATE]).

Workflow: [intent at FEATURE/SYSTEM scope] -> feel -> harden -> validate

## Gate Definitions

| Gate ID | Dimension | Type | Phase | Criterion | Agent |
|---------|-----------|------|-------|-----------|-------|
| QG-C1 | D0: Accessibility | **Hard** | validate | WCAG 2.2 AA zero-tolerance across four layers | a11y-engineer |
| QG-C2 | D1: Micro-interactions | **Soft (blocking)** | validate | Interaction timing, physics, and fluidity meet the intent classification | frontend-fanatic |
| QG-C3 | D2: Cognitive efficiency | **Soft (blocking)** | validate | User can accomplish goal through most direct interaction path | frontend-fanatic |
| QG-C4 | D3: Consistency | Embedded | harden | New patterns consistent with existing product patterns or documented as intentional divergence | producing agent (self-check) |
| QG-C5 | D4: Personality/Brand | Advisory | validate | Custom solutions serve product identity; generic solutions where distinctiveness adds no value | frontend-fanatic |
| QG-C6 | D5: Edge state craft | Embedded | harden | All states (loading, empty, error, boundary) intentionally designed, not default | producing agent (self-check) |
| QG-C7 | D6: Emotional design | Advisory | validate | Three-level emotional assessment (visceral, behavioral, reflective) documented | frontend-fanatic |

## Key Design Decisions

**QG-C2, QG-C3 (Soft gates -- blocking in generative posture)**: This is the critical design decision. In the generative workflow, the entire purpose is "build something that feels right." If the interaction does not feel right (D1) or the user cannot efficiently accomplish their goal (D2), the generative workflow has failed at its primary objective. Making these advisory would mean the generative workflow can "succeed" while producing interactions that fail the very criterion they were built to satisfy.

**Feel phase has NO quality gates**: The feel phase produces throwaway code whose only criterion is "does this feel right in the browser?" -- a subjective practitioner assessment, not a formalizable gate. QG-C1 through QG-C7 apply to harden (embedded) and validate (terminal), never to feel.

**QG-C5, QG-C7 (Advisory)**: D4 (personality/brand) and D6 (emotional design) are inherently subjective with no formalizable gate criteria. These are always advisory regardless of posture or scope.

## Soft Gate Mechanics for QG-C2 and QG-C3

**Frontend-fanatic invocation**: Automatically invoked at FEATURE and SYSTEM scope validate phase. At COMPONENT scope, invoked at potnia's discretion.

**QG-C2 pass criteria**:
- Interaction timing at or below 200ms for interactive animations
- Physics (spring/keyframe) matches motion-architecture-spec
- Animation is interruptible (user can reverse mid-flight)
- Motion communicates spatial relationships correctly

**QG-C2 fail -> back-route**: validate -> harden (component-engineer). Finding must specify what failed and what would pass.

**QG-C3 pass criteria**:
- User can identify primary action without instruction
- Task completion path matches the most direct available route
- No unnecessary decision points in the primary flow
- Cognitive load reduction visible vs. baseline

**QG-C3 fail -> back-route**: validate -> harden (component-engineer). Finding must specify the flow that fails and expected behavior.

## Coverage Summary

| Dimension | Generative Evaluation |
|-----------|----------------------|
| D0: Accessibility | Hard gate (a11y-engineer) |
| D1: Micro-interactions | Soft gate -- blocking (frontend-fanatic, automatic at FEATURE/SYSTEM) |
| D2: Cognitive efficiency | Soft gate -- blocking (frontend-fanatic, automatic at FEATURE/SYSTEM) |
| D3: Consistency | Embedded self-check |
| D4: Personality/Brand | Advisory (frontend-fanatic) |
| D5: Edge state craft | Embedded self-check |
| D6: Emotional design | Advisory (frontend-fanatic) |

Coverage: 7/7 dimensions. Full coverage consistent with Lovin's team-level quality principle.

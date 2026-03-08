---
name: thermia-ref-back-routes
description: "Thermia back-route protocols for consultation workflow. Use when: systems-thermodynamicist flags uncovered data paths, thermal-monitor identifies capacity/architecture inconsistency, consultation needs to return to an earlier phase. Triggers: back-route, assessment_gap, design_inconsistency, architecture gap, specification inconsistency, consultation back-route."
---

# Thermia: Back-Route Protocols

Two back-routes exist in thermia. Both are rare, not routine. Track iterations in `throughline.rationale`.

## Back-Route Summary

| Back-Route | Source | Target | Max Iterations | On Limit |
|------------|--------|--------|----------------|----------|
| `assessment_gap` | architecture | assessment (heat-mapper) | 1 | Escalate to user |
| `design_inconsistency` | validation | specification (capacity-engineer) | 1 | Escalate to user |

## assessment_gap Protocol

**Trigger**: Systems-thermodynamicist identifies data paths or access patterns not covered in `thermal-assessment.md` that are required to complete the architecture design.

**When NOT to trigger**: Curiosity about additional paths that were not scoped in the consultation. This back-route is for blocking gaps, not optional enrichment.

**Protocol**:
1. Systems-thermodynamicist produces a **targeted gap description**: which data paths are uncovered, why they are required (specific architectural decision blocked)
2. Potnia routes heat-mapper to produce a **targeted supplement** to `thermal-assessment.md` — NOT a full redo
3. Heat-mapper collects only the missing data and appends to the existing assessment
4. Heat-mapper does NOT re-evaluate already-assessed paths
5. Potnia re-routes to systems-thermodynamicist with the supplemented assessment

**Iteration tracking**: `assessment_gap: N/1` in `throughline.rationale`

**On limit (1 reached)**: Escalate to user. The assessment has been supplemented once. Further gaps indicate missing system context that the user must provide. Do not trigger a second back-route.

## design_inconsistency Protocol

**Trigger**: Thermal-monitor identifies a specific inconsistency between the capacity specification and the cache architecture that creates an unmonitorable or unobservable failure mode.

**Example**: Architecture specifies fail-open stale fallback, but capacity spec has no eviction headroom — the failure mode cannot be detected because the system will silently exceed capacity before fail-open activates.

**When NOT to trigger**: Minor alignment issues that can be noted as recommendations. This back-route is for inconsistencies that make the design operationally dangerous.

**Protocol**:
1. Thermal-monitor produces a **specific inconsistency description**: which failure mode is affected, what the conflict is between architecture and capacity spec
2. Potnia routes capacity-engineer to **reconcile the specific inconsistency** — not a full re-spec
3. Capacity-engineer adjusts only the conflicting elements and updates `capacity-specification.md`
4. Potnia re-routes to thermal-monitor with the updated specification

**Iteration tracking**: `design_inconsistency: N/1` in `throughline.rationale`

**On limit (1 reached)**: Escalate to user. The specification has been revised once. Persistent inconsistency is an architecture-level trade-off that requires user input on acceptable risk.

## Potnia throughline.rationale Format

Include in every CONSULTATION_RESPONSE:

```yaml
throughline:
  complexity: QUICK | STANDARD | DEEP
  phase: assessment | architecture | specification | validation
  gates_passed: [assessment, architecture]   # list completed phases
  back_routes:
    assessment_gap: 0/1
    design_inconsistency: 0/1
  rationale: "Brief state summary for resume continuity"
```

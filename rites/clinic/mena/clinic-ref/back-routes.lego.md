---
name: clinic-ref-back-routes
description: "Clinic back-route protocols — evidence_gap, diagnosis_insufficient, scope_expansion. Use when: diagnostician needs more evidence, attending needs deeper diagnosis, scope has expanded. Triggers: back-route, evidence gap, diagnosis insufficient, scope expansion, back-route protocol."
---

# Clinic: Back-Route Protocols

The clinic's defining workflow characteristic. Expected patterns, not failure modes.

| Back-Route | Trigger | Source | Target | Max Iterations | On Limit |
|------------|---------|--------|--------|----------------|----------|
| `evidence_gap` | Diagnostician needs evidence not yet collected | diagnosis | examination (pathologist) | 3 | Escalate to user |
| `diagnosis_insufficient` | Attending cannot produce actionable fix spec | treatment | diagnosis (diagnostician) | 2 | Escalate to user |
| `scope_expansion` | Investigation involves systems outside original scope | diagnosis | intake (triage-nurse) | 1 | Hard limit: split investigation |

## evidence_gap Protocol

1. Diagnostician produces a targeted evidence request: system, data needed, why (which hypothesis)
2. Pythia routes directly to pathologist — does NOT re-gate intake or re-scope
3. Pathologist collects only the requested evidence
4. Pathologist writes to next sequential evidence file and updates index.yaml
5. Pathologist does NOT re-collect previously gathered evidence
6. Status field: `examination:evidence_gap_round_{N}`

## diagnosis_insufficient Protocol

1. Attending produces specific concerns: what is unclear, what additional depth is needed
2. Pythia routes back to diagnostician with the attending's concerns
3. Diagnostician deepens analysis (localizes affected code, clarifies mechanism, raises confidence)
4. Status field: `diagnosis:diagnosis_insufficient_round_{N}`

## scope_expansion Protocol

Requires user confirmation. User may prefer to open a second investigation rather than expand.

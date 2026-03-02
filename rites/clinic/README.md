# Clinic Rite

> Take it to the clinic. Production errors, root causes, treatment plans.

**Version**: 1.0.0 | **Domain**: Production debugging and investigation | **Command**: `/clinic`

## When to Use This Rite

**Triggers**:
- Production error with unknown or uncertain root cause
- Intermittent failure that resists ad hoc debugging
- SRE incident escalation requiring structured root cause analysis
- Failing tests with no clear cause
- Performance regression of unknown origin
- "We fixed it but it came back" situations

**Not for**: Known bugs where the fix is obvious (use `/hotfix`), feature development (use `/10x`), general code quality (use `/hygiene`).

## Quick Start

```bash
/clinic
```

Pythia opens with triage-nurse to scope the investigation. Describe what you observed.

**SRE escalation**: If escalating from an incident war room, provide incident context directly. The triage-nurse accepts pre-gathered incident data as a head start.

## Agents

| Agent | Model | Phase | Produces |
|-------|-------|-------|----------|
| **pythia** | sonnet | Orchestrator | Phase gates, back-route management |
| **triage-nurse** | sonnet | Intake | intake-report.md, index.yaml |
| **pathologist** | sonnet | Examination | E001.txt ... E{N}.txt evidence files |
| **diagnostician** | opus | Diagnosis | diagnosis.md |
| **attending** | opus | Treatment | treatment-plan.md, handoff artifact(s) |

## Workflow

```
intake -> examination -> diagnosis -> treatment
```

All four phases always run. Depth varies; phases do not skip.

### Back-Routes (expected, not exceptional)

```
diagnosis ──evidence_gap──> examination    (max 3 iterations)
treatment ─diagnosis_insufficient─> diagnosis  (max 2 iterations)
diagnosis ──scope_expansion──> intake      (max 1, requires user confirmation)
```

## Complexity Level

**INVESTIGATION** — the only level. There is no pre-classification.

The clinic does not have complexity tiers. Every investigation runs all four phases. A simple stack trace takes 4 agent invocations and ~30k tokens. A compound cross-service failure takes 6-8 invocations and ~150k tokens. The difference is emergent from what agents find, not from an upfront decision.

## Evidence Architecture

All artifacts live in `.sos/wip/ERRORS/{investigation-slug}/`:

```
intake-report.md       <- triage-nurse
index.yaml             <- shared coordination artifact (all agents)
E001.txt               <- pathologist (sequential evidence files)
E002.txt               <- pathologist
...
diagnosis.md           <- diagnostician
treatment-plan.md      <- attending
handoff-10x-dev.md     <- attending (if bug fix needed)
handoff-sre.md         <- attending (if monitoring gaps found)
handoff-debt-triage.md <- attending (if systemic pattern found)
```

## Session Resume

Park and continue investigations naturally. On `/continue`, Pythia reads `index.yaml` status field to resume at the correct phase.

## Cross-Rite Handoffs (Outbound)

The attending determines which downstream rites receive handoff artifacts based on the diagnosis. Handoffs are recommendations — the user decides whether to act.

| Handoff | Target Rite | When |
|---------|-------------|------|
| handoff-10x-dev.md | `/10x` | Root cause is a fixable bug or misconfiguration |
| handoff-sre.md | `/sre` | Investigation reveals monitoring or observability gaps |
| handoff-debt-triage.md | `/debt` | Root cause is a systemic pattern across the codebase |

A single investigation may produce multiple outbound handoffs.

## Inbound: SRE Escalation

The clinic accepts structured escalation from SRE incident-commander. Expected fields:

- `incident_id`: SRE incident identifier
- `timeline`: Chronological event sequence
- `affected_services`: Services identified during triage
- `symptoms_observed`: Documented symptoms
- `initial_findings`: Free-text responder notes
- `severity`: SRE-assessed severity (clinic may adjust)

## Commands

| Command | Purpose |
|---------|---------|
| `/clinic` | Switch to clinic rite |
| `/continue` | Resume parked investigation |

## Best For

- Bugs that have resisted quick debugging
- Compound failures (multiple interacting root causes)
- Incidents requiring structured evidence collection
- Cross-service failures where the call chain matters
- Any investigation that needs to produce a formal treatment plan with handoff artifacts

## Not For

- Known bugs with obvious fixes (use `/hotfix`)
- Feature development (use `/10x`)
- Code quality improvements (use `/hygiene`)
- Performance profiling without a production failure to investigate

## Related Rites

- `/sre` — Receives clinic handoff when monitoring gaps are found; also escalates to clinic during incidents
- `/10x` — Receives clinic handoff to implement the identified fix
- `/debt` — Receives clinic handoff when systemic patterns are identified

## Design Notes

The clinic's medical metaphor is intentional. Production debugging has a natural structure: intake (what happened?), examination (collect the evidence), diagnosis (what caused it?), treatment (how do we fix it?). The four roles map to these phases, and the terminology reinforces that good debugging is methodical, not improvised.

The back-route mechanism is the clinic's defining characteristic. Real investigations do not flow linearly — hypotheses fail, evidence gaps appear, scope expands. These are expected workflow patterns, not errors. The clinic treats iteration as first-class.

See `rites/clinic/workflow.yaml` for the full specification.

---
name: shape-schema
description: |
  Execution shape file schema for initiative orchestration planning.
  Use when: producing shape files, understanding shape structure, consulting on initiative
  execution planning, decomposing cross-rite processions into sprints.
  Triggers: shape, shape file, execution shape, initiative orchestration,
  sprint decomposition, Potnia checkpoints, procession planning.
---

# Shape Schema

> The execution plan companion to a frame's problem decomposition -- tells Potnia how to orchestrate an initiative across sprints, rites, and agent rosters.

## Quick Reference

| # | Section | Status | Purpose |
|---|---------|--------|---------|
| 1 | Initiative Thread | **Required** | Throughline Potnia carries across all sprints |
| 2 | Sprint Decomposition | **Required** | Per-sprint mission, roster, entry/exit criteria |
| 3 | Potnia Consultation Points | **Required** | PT-NN evaluative gates tied to throughline |
| 4 | Execution Sequence | **Required** | Command flow with rite transitions |
| 5 | Critical Path | Optional | Sprint dependency graph, parallel opportunities |
| 6 | Cross-Rite Handoff Protocol | Optional | Artifact naming and transition protocol |
| 7 | Emergent Behavior Constraints | Optional | Prescribed / emergent / out-of-scope boundaries |
| 8 | Context Loading Order | Optional | Per-sprint Read() file list at session start |
| 9 | Risk Map | Optional | Sprint-level risk assessment |
| 10 | Estimated Duration | Optional | Per-sprint timing and velocity assumptions |

**Required** = always useful for any shape. **Optional** = include when applicable (cross-rite, complex dependencies, tight timelines).

## Shape File Frontmatter

Every shape file begins with YAML frontmatter:

```yaml
---
type: shape
initiative: <initiative-slug>         # kebab-case, matches frame slug if frame exists
frame: <path-to-frame-or-null>        # .sos/wip/frames/{slug}.md or null
created: <YYYY-MM-DD>
rite: <primary-rite or "cross-rite">  # single rite slug or "cross-rite" for multi-rite
complexity: <MODULE|INITIATIVE>       # MODULE = single-rite, INITIATIVE = cross-rite
scope:
  rites: [<rite-1>, <rite-2>, ...]    # all rites touched by this initiative
  sprints: <N>                        # total sprint count
cross_rite_consultations:             # omit for single-rite shapes
  - from: <rite-a>
    to: <rite-b>
    artifact: <handoff-artifact-slug>
---
```

| Field | Required | Notes |
|-------|----------|-------|
| `type` | Yes | Always `shape` |
| `initiative` | Yes | Kebab-case slug |
| `frame` | No | Path to frame file if one exists |
| `created` | Yes | ISO date |
| `rite` | Yes | Primary rite or `"cross-rite"` |
| `complexity` | Yes | `MODULE` or `INITIATIVE` |
| `scope.rites` | Yes | List of all rite slugs involved |
| `scope.sprints` | Yes | Integer sprint count |
| `cross_rite_consultations` | No | Only for cross-rite shapes |

## Required Sections

### 1. Initiative Thread

The throughline that Potnia carries across every sprint and rite transition. This is the initiative's "consciousness" -- what every checkpoint evaluates against. Not a summary of the work; a statement of the invariant that must hold true from start to finish.

```yaml
initiative_thread:
  throughline: "<one-sentence statement of what success looks like>"
  success_criteria:
    - "<measurable outcome 1>"
    - "<measurable outcome 2>"
  failure_signals:
    - "<early warning that the initiative is going off-track>"
```

Deep guidance: [sections/initiative-thread.md](sections/initiative-thread.md)

### 2. Sprint Decomposition

Per-sprint definitions with rite assignment, agent roster, entry/exit criteria, and context references. **Sprint internals are NON-PRESCRIPTIVE** -- define the mission and constraints, not the agent task list. Potnia coordinates; agents discover.

```yaml
sprints:
  - id: sprint-1
    rite: <rite-slug>
    mission: "<what this sprint accomplishes>"
    agents: [<agent-1>, <agent-2>]
    entry_criteria:
      - "<what must be true before this sprint starts>"
    exit_criteria:
      - "<what must be true for this sprint to be complete>"
    exit_artifacts:
      - path: "<file path>"
        description: "<what this artifact contains>"
    context:
      - "<path to read at sprint start>"
```

Deep guidance: [sections/sprint-decomposition.md](sections/sprint-decomposition.md)

### 3. Potnia Consultation Points

PT-NN numbered checkpoints with specific evaluative questions. Every checkpoint must evaluate throughline satisfaction -- not "is everything OK?" but "does X satisfy Y?" Include business-domain granularity where applicable.

```yaml
checkpoints:
  - id: PT-01
    after: sprint-1
    evaluates: "<what aspect of the throughline this checks>"
    questions:
      - "<specific question with measurable answer>"
    gate: <hard|soft>
    on_fail: "<action if checkpoint fails>"
```

Deep guidance: [sections/potnia-checkpoints.md](sections/potnia-checkpoints.md)

### 4. Execution Sequence

Full command flow showing rite transitions, session lifecycle, and file loading order per sprint. This is the operational playbook -- what commands to run, in what order.

```yaml
execution_sequence:
  - phase: setup
    commands:
      - "ari sync --rite=<rite>"
      - "/sos start --initiative=<slug>"
  - phase: sprint-1
    commands:
      - "/sprint 1"
      - "Read('<context-file>')"
    notes: "<operational context>"
  - phase: transition
    commands:
      - "/sos wrap"
      - "ari sync --rite=<next-rite>"
      - "/sos start --initiative=<slug>"
```

Deep guidance: [sections/execution-sequence.md](sections/execution-sequence.md)

## Optional Sections

### 5. Critical Path
Sprint dependency graph identifying parallel execution opportunities. Use when sprints have non-linear dependencies.
Deep guidance: [sections/critical-path.md](sections/critical-path.md)

### 6. Cross-Rite Handoff Protocol
How work transitions between rites. Outgoing rite produces handoff artifact at `.ledge/`; incoming rite's Potnia loads as context. Use for any multi-rite initiative.
Deep guidance: [sections/cross-rite-handoff.md](sections/cross-rite-handoff.md)

### 7. Emergent Behavior Constraints
Three categories: Prescribed (must follow), Emergent (agent discretion), Out of Scope (must not touch). Use when agents need explicit freedom boundaries.
Deep guidance: [sections/emergent-behavior.md](sections/emergent-behavior.md)

### 8. Context Loading Order
Per-sprint file list for Read() at session start. Spike references, prior sprint artifacts, configuration files. Use when context is complex or order-sensitive. Note: this is the expanded form with `reason` fields; use either this section or `sprint.context` (compact form), not both.
Deep guidance: [sections/context-loading.md](sections/context-loading.md)

### 9. Risk Map
Sprint-level risk assessment: what could go wrong, probability, impact, mitigation. Use for high-stakes initiatives or unfamiliar territory.
Deep guidance: [sections/risk-map.md](sections/risk-map.md)

### 10. Estimated Duration
Per-sprint time estimates and velocity assumptions. Use when timeline commitments exist.
Deep guidance: [sections/duration-estimates.md](sections/duration-estimates.md)

## Companion References

| File | Contains |
|------|----------|
| `sections/initiative-thread.md` | Throughline design patterns, success criteria calibration |
| `sections/sprint-decomposition.md` | Sprint structure, non-prescriptive internals, roster guidance |
| `sections/potnia-checkpoints.md` | PT-NN design, evaluative question patterns, gate types |
| `sections/execution-sequence.md` | Command flow templates, rite transition protocol |
| `sections/critical-path.md` | Dependency diagramming, parallel sprint identification |
| `sections/cross-rite-handoff.md` | Artifact naming, `.ledge/` conventions, handoff verification |
| `sections/emergent-behavior.md` | Prescribed/emergent/out-of-scope boundary design |
| `sections/context-loading.md` | File ordering heuristics, context budget awareness |
| `sections/risk-map.md` | Risk categorization, mitigation patterns |
| `sections/duration-estimates.md` | Velocity assumptions, buffer allocation |

## Examples

| File | Complexity | Rites | Sprints |
|------|-----------|-------|---------|
| [examples/cross-rite-initiative.md](examples/cross-rite-initiative.md) | INITIATIVE | 4 | 7 |
| [examples/single-rite-module.md](examples/single-rite-module.md) | MODULE | 1 | 3 |

## Living Document Protocol

Shape files are mutable. As Potnia executes sprints, she annotates the shape:

```yaml
# After checkpoint evaluation:
- id: PT-01
  after: sprint-1
  status: PASSED          # added by Potnia at runtime
  outcome: "API contract validated against 3 consumer services"

# After sprint completion:
- id: sprint-1
  status: COMPLETE        # added by Potnia at runtime
  completed: 2026-03-15
```

The shape evolves from plan to execution record. Git history provides the audit trail.

## Consumers

- **Potnia**: Primary consumer. Loads shape at `/sprint` start to understand mission, constraints, and checkpoints.
- **Pythia**: Producer. Creates shapes via `/shape` dromenon using this schema.
- **Human operators**: Secondary consumer. Reads shapes to understand initiative structure.

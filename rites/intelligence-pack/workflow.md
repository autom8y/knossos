# Intelligence Pack Workflow

## Phase Flow

```
┌───────────────────┐      ┌───────────────────┐      ┌───────────────────┐      ┌───────────────────┐
│    Analytics      │─────▶│       User        │─────▶│  Experimentation  │─────▶│     Insights      │
│     Engineer      │      │    Researcher     │      │       Lead        │      │     Analyst       │
└───────────────────┘      └───────────────────┘      └───────────────────┘      └───────────────────┘
  Tracking Plan           Research Findings        Experiment Design         Insights Report
```

## Phases

| Phase | Agent | Artifact | Entry Criteria |
|-------|-------|----------|----------------|
| instrumentation | analytics-engineer | Tracking Plan | User request |
| research | user-researcher | Research Findings | Tracking plan complete, complexity >= FEATURE |
| experimentation | experimentation-lead | Experiment Design | Research complete (or tracking if METRIC) |
| synthesis | insights-analyst | Insights Report | Experiment design complete |

## Complexity Levels

- **METRIC**: Single metric analysis, existing event data
  - Phases: experimentation, synthesis
- **FEATURE**: New feature instrumentation, user journey analysis
  - Phases: instrumentation, research, experimentation, synthesis
- **INITIATIVE**: Cross-feature analysis, strategic product decisions
  - Phases: instrumentation, research, experimentation, synthesis

## Phase Skipping

At METRIC complexity, instrumentation and research phases are skipped when analyzing existing data.

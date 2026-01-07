# R&D Pack Workflow

## Phase Flow

```
┌───────────────────┐      ┌───────────────────┐      ┌───────────────────┐      ┌───────────────────┐
│   Technology      │─────▶│   Integration     │─────▶│    Prototype      │─────▶│    Moonshot       │
│      Scout        │      │    Researcher     │      │     Engineer      │      │    Architect      │
└───────────────────┘      └───────────────────┘      └───────────────────┘      └───────────────────┘
  Tech Assessment        Integration Map             Prototype              Moonshot Plan
```

## Phases

| Phase | Agent | Artifact | Entry Criteria |
|-------|-------|----------|----------------|
| scouting | technology-scout | Tech Assessment | User request |
| integration-analysis | integration-researcher | Integration Map | Tech assessment complete, complexity >= EVALUATION |
| prototyping | prototype-engineer | Prototype | Integration analysis complete (or assessment if SPIKE) |
| future-architecture | moonshot-architect | Moonshot Plan | Prototype complete |

## Complexity Levels

- **SPIKE**: Quick feasibility check, single technology
  - Phases: scouting, prototyping
- **EVALUATION**: Full technology evaluation with integration analysis
  - Phases: scouting, integration-analysis, prototyping, future-architecture
- **MOONSHOT**: Paradigm shift exploration, multi-year architecture
  - Phases: scouting, integration-analysis, prototyping, future-architecture

## Phase Skipping

At SPIKE complexity, integration-analysis and future-architecture phases are skipped for quick feasibility checks.

# SRE Pack Workflow

## Phase Flow

```
┌───────────────┐      ┌───────────────┐      ┌───────────────┐      ┌───────────────┐
│ Observability │─────▶│   Incident    │─────▶│   Platform    │─────▶│     Chaos     │
│   Engineer    │      │  Commander    │      │   Engineer    │      │   Engineer    │
└───────────────┘      └───────────────┘      └───────────────┘      └───────────────┘
Observability Rpt    Reliability Plan    Infrastructure Chgs     Resilience Report
```

## Phases

| Phase | Agent | Artifact | Entry Criteria |
|-------|-------|----------|----------------|
| observation | observability-engineer | Observability Report | User request or monitoring alert |
| coordination | incident-commander | Reliability Plan | Observation complete, complexity >= SERVICE |
| implementation | platform-engineer | Infrastructure Changes | Plan approved (or observation if ALERT) |
| resilience | chaos-engineer | Resilience Report | Infrastructure changes deployed |

## Complexity Levels

- **ALERT**: Single alert/dashboard fix
  - Phases: implementation, resilience
- **SERVICE**: Single service reliability
  - Phases: observation, coordination, implementation, resilience
- **SYSTEM**: Multi-service SLOs/SLIs
  - Phases: observation, coordination, implementation, resilience
- **PLATFORM**: Full platform reliability
  - Phases: observation, coordination, implementation, resilience

## Phase Skipping

At ALERT complexity, observation and coordination phases are skipped for quick tactical fixes.

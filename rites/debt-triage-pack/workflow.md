# Debt Triage Pack Workflow

## Phase Flow

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Debt Collector │────▶│  Risk Assessor  │────▶│  Sprint Planner │
│   (Catalogs)    │     │    (Scores)     │     │   (Packages)    │
└─────────────────┘     └─────────────────┘     └─────────────────┘
   Debt Ledger            Risk Report            Sprint Plan
```

## Phases

| Phase | Agent | Artifact | Entry Criteria |
|-------|-------|----------|----------------|
| collection | debt-collector | Debt Ledger | User request or scheduled audit |
| assessment | risk-assessor | Risk Report | Debt ledger complete |
| planning | sprint-planner | Sprint Plan | Risk assessment complete |

## Complexity Levels

- **QUICK**: Known debt items
  - Phases: assessment, planning
- **AUDIT**: Full debt discovery
  - Phases: collection, assessment, planning

## Phase Skipping

At QUICK complexity, the collection phase is skipped when debt items are already known.

# Ecosystem Pack Workflow

## Phase Flow

```
┌───────────────┐      ┌───────────────┐      ┌───────────────┐      ┌───────────────┐      ┌───────────────┐
│  Ecosystem    │─────▶│    Context    │─────▶│ Integration   │─────▶│ Documentation │─────▶│ Compatibility │
│   Analyst     │      │   Architect   │      │   Engineer    │      │   Engineer    │      │    Tester     │
└───────────────┘      └───────────────┘      └───────────────┘      └───────────────┘      └───────────────┘
 Gap Analysis        Context Design         Implementation        Migration Runbook      Compatibility Report
```

## Phases

| Phase | Agent | Artifact | Entry Criteria |
|-------|-------|----------|----------------|
| analysis | ecosystem-analyst | Gap Analysis | User request |
| design | context-architect | Context Design | Gap analysis complete, complexity >= MODULE |
| implementation | integration-engineer | Implementation | Design approved (or gap analysis if PATCH) |
| documentation | documentation-engineer | Migration Runbook | Implementation complete, complexity >= MODULE |
| validation | compatibility-tester | Compatibility Report | Runbook complete (or implementation if PATCH) |

## Complexity Levels

- **PATCH**: Single file/config change, no schema impact
  - Phases: analysis, implementation, validation
- **MODULE**: Single system (CEM or roster)
  - Phases: analysis, design, implementation, documentation, validation
- **SYSTEM**: Multi-system change affecting CEM + roster
  - Phases: analysis, design, implementation, documentation, validation
- **MIGRATION**: Cross-satellite rollout requiring coordination
  - Phases: analysis, design, implementation, documentation, validation

## Phase Skipping

At PATCH complexity, design and documentation phases are skipped.

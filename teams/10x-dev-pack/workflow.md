# 10x Dev Pack Workflow

## Phase Flow

```
┌───────────────┐      ┌───────────────┐      ┌───────────────┐      ┌───────────────┐
│  Requirements │─────▶│   Architect   │─────▶│   Principal   │─────▶│      QA       │
│    Analyst    │      │               │      │   Engineer    │      │   Adversary   │
└───────────────┘      └───────────────┘      └───────────────┘      └───────────────┘
      PRD                    TDD                    Code               Test Plan
```

## Phases

| Phase | Agent | Artifact | Entry Criteria |
|-------|-------|----------|----------------|
| requirements | requirements-analyst | PRD | User request |
| design | architect | TDD | PRD complete, complexity >= MODULE |
| implementation | principal-engineer | Code | TDD approved (or PRD if SCRIPT) |
| validation | qa-adversary | Test Plan | Code complete |

## Complexity Levels

- **SCRIPT**: Single file, <200 LOC
  - Phases: requirements, implementation, validation
- **MODULE**: Multiple files, <2000 LOC
  - Phases: requirements, design, implementation, validation
- **SERVICE**: APIs, persistence
  - Phases: requirements, design, implementation, validation
- **PLATFORM**: Multi-service
  - Phases: requirements, design, implementation, validation

## Phase Skipping

At SCRIPT complexity level, the design phase is skipped. All other complexity levels require all four phases.

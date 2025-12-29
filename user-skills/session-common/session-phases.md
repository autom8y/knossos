# Session Phases

> Phase definitions and agent mappings for session lifecycle.

## Phase Definitions

| Phase | Description | Entry Condition |
|-------|-------------|-----------------|
| `requirements` | Clarify scope, produce PRD | /start invoked |
| `design` | Architecture, produce TDD/ADRs | complexity >= MODULE |
| `implementation` | Code production | Design approved |
| `validation` | QA and acceptance testing | Implementation complete |

## Phase-to-Agent Mapping

| Phase | Primary Agent | Fallback |
|-------|---------------|----------|
| requirements | requirements-analyst | - |
| design | architect | - |
| implementation | principal-engineer | - |
| validation | qa-adversary | - |

## Valid Phase Transitions

```
requirements --> design (MODULE+)
requirements --> implementation (SCRIPT only)
design --> implementation
implementation --> validation
validation --> implementation (defects found)
validation --> [wrap] (all gates pass)
```

## Complexity-Phase Matrix

| Complexity | Phases Included |
|------------|-----------------|
| SCRIPT | requirements, implementation, validation |
| MODULE | requirements, design, implementation, validation |
| SERVICE | requirements, design, implementation, validation |
| PLATFORM | requirements, design, implementation, validation |

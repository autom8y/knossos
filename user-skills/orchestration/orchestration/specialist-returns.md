# Specialist Return Format

> What specialists return to the main thread after Task tool invocation

## Purpose

When the main thread invokes a specialist via Task tool, the specialist produces artifacts and returns. This document defines the expected return structure for clean handoff back to the orchestrator.

## Return Schema

Specialists should structure their final output as:

```yaml
# SPECIALIST_RETURN
status: "complete" | "blocked" | "needs_info"

artifacts:
  - path: string           # File path created/modified
    type: "prd" | "tdd" | "adr" | "code" | "test" | "doc"
    summary: string        # 1-2 sentences max

handoff_criteria:
  - criterion: string      # From the prompt's handoff criteria
    met: boolean           # Objectively verifiable

notes: string | null       # Optional context for orchestrator
```

## Examples by Phase

### Requirements Phase
```yaml
status: complete
artifacts:
  - path: docs/PRD-auth.md
    type: prd
    summary: "OAuth2 authentication with Google, 8 user stories, token security reqs"
handoff_criteria:
  - criterion: "User stories defined"
    met: true
  - criterion: "OAuth flow documented"
    met: true
  - criterion: "Security requirements captured"
    met: true
notes: null
```

### Design Phase
```yaml
status: complete
artifacts:
  - path: docs/TDD-auth.md
    type: tdd
    summary: "Express middleware pattern, JWT with Redis session store"
  - path: docs/ADR-001-token-storage.md
    type: adr
    summary: "Redis over PostgreSQL for session tokens (latency)"
handoff_criteria:
  - criterion: "OAuth flow sequence diagram"
    met: true
  - criterion: "Database schema for tokens"
    met: true
  - criterion: "ADR for token storage"
    met: true
notes: null
```

### Blocked Return
```yaml
status: blocked
artifacts: []
handoff_criteria:
  - criterion: "OAuth flow sequence diagram"
    met: false
notes: "Cannot proceed: Google OAuth credentials not configured in environment"
```

## Converting to Checkpoint Request

Main thread builds checkpoint request from specialist return:

| Specialist Return | Checkpoint Request Field |
|-------------------|--------------------------|
| `artifacts[].path` | `state.artifacts_produced` |
| `artifacts[].summary` | `results.artifact_summary` (concatenated) |
| `handoff_criteria[].met` | `results.handoff_criteria_met` |
| `status == "blocked"` | `type: failure` with `results.failure_reason` |

## See Also

- [request-format.md](request-format.md) - Checkpoint request schema
- [main-thread-guide.md](main-thread-guide.md) - How main thread processes returns

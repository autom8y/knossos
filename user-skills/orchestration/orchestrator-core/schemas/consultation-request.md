# CONSULTATION_REQUEST Schema

> Input to orchestrator for routing decisions.

## Schema

```yaml
request_id: string           # UUID v4 for correlation
type: enum                   # initial | checkpoint | decision | failure

initiative:
  name: string               # Initiative/feature name
  complexity: enum           # Team-specific (SCRIPT|MODULE|SERVICE|PLATFORM or PATCH|MODULE|SYSTEM)

state:
  current_phase: string      # Current workflow phase
  completed_phases: array    # Completed phase names
  artifacts_produced: array  # Artifacts created
    - type: string           # prd, tdd, test-plan, etc.
      path: string           # File path
      status: enum           # draft | review | approved

results:                     # For checkpoint/failure types
  phase_completed: string    # Phase just finished
  artifact_summary: string   # Brief outcome (max 100 chars)
  handoff_criteria_met: array
    - criterion_id: string   # prd-001, tdd-002, etc.
      status: enum           # PASS | FAIL | SKIP
  failure_reason: string     # Human-readable (failure type only)
  failure_pattern: enum      # blocker | scope | capacity | underspecified

context_summary: string      # Max 200 words
```

## Request Types

| Type | When | Expected Response |
|------|------|-------------------|
| `initial` | Starting initiative | Route to first phase |
| `checkpoint` | Phase completed | Route to next phase |
| `decision` | Multiple valid paths | Request user input or decide |
| `failure` | Phase blocked | Recovery guidance |

## Examples

**Initial Request:**
```yaml
type: initial
initiative:
  name: "User Authentication"
  complexity: MODULE
state:
  current_phase: null
  completed_phases: []
  artifacts_produced: []
results: null
context_summary: "New feature: email/password login with rate limiting"
```

**Checkpoint Request:**
```yaml
type: checkpoint
initiative:
  name: "User Authentication"
  complexity: MODULE
state:
  current_phase: requirements
  completed_phases: []
  artifacts_produced:
    - type: prd
      path: "docs/requirements/PRD-user-auth.md"
      status: approved
results:
  phase_completed: requirements
  artifact_summary: "PRD approved with 3 testable success criteria"
  handoff_criteria_met:
    - criterion_id: prd-001
      status: PASS
    - criterion_id: prd-002
      status: PASS
  failure_reason: null
context_summary: "Requirements complete. Ready for design."
```

**Failure Request:**
```yaml
type: failure
initiative:
  name: "User Authentication"
  complexity: MODULE
state:
  current_phase: design
  completed_phases: [requirements]
  artifacts_produced:
    - type: prd
      path: "docs/requirements/PRD-user-auth.md"
      status: approved
results:
  phase_completed: null
  artifact_summary: "Partial TDD, API contracts incomplete"
  handoff_criteria_met: []
  failure_reason: "External auth service API not documented. Cannot complete TDD-003."
  failure_pattern: blocker
context_summary: "Design blocked. Auth team committed to API docs tomorrow."
```

## Validation Rules

1. `type` MUST be: initial, checkpoint, decision, or failure
2. `initiative.name` MUST be non-empty
3. `initiative.complexity` MUST match team enum
4. `state.current_phase` MUST be valid phase for team
5. `results` MUST be present for checkpoint/failure
6. `results.failure_reason` MUST be present for failure
7. `context_summary` MUST be ≤ 200 words

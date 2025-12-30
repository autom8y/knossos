# CONSULTATION_REQUEST Schema

> Canonical schema for orchestrator input. Extracted to reduce duplication across team orchestrator.md files.

## Schema Definition

```yaml
# CONSULTATION_REQUEST
# Input to orchestrator for routing decisions

type: enum                   # Request type
  values:
    - initial                # First consultation for initiative
    - checkpoint             # Phase completed, seeking next phase
    - decision               # Branching point requiring choice
    - failure                # Phase failed, seeking recovery

initiative:
  name: string               # Initiative/feature name
  complexity: enum           # Team-specific complexity level
    # 10x-dev-pack: SCRIPT | MODULE | SERVICE | PLATFORM
    # ecosystem-pack: PATCH | MODULE | SYSTEM | MIGRATION

state:
  current_phase: string      # Current workflow phase name
  completed_phases: array    # List of completed phase names
  artifacts_produced: array  # Artifacts created so far
    - type: string           # Artifact type (prd, tdd, etc.)
      path: string           # Artifact file path
      status: enum           # draft | review | approved

results:                     # For checkpoint/failure types
  phase_completed: string    # Phase that just finished
  artifact_summary: string   # Brief description of output
  handoff_criteria_met: array  # Which criteria passed
    - criterion_id: string   # prd-001, tdd-002, etc.
      status: enum           # PASS | FAIL | SKIP
  failure_reason: string     # Why phase failed (if applicable)

context_summary: string      # 200 words max, key context
```

## Field Details

### type

| Value | When to Use | Expected Response |
|-------|-------------|-------------------|
| `initial` | Starting new initiative | Route to first phase |
| `checkpoint` | Phase completed successfully | Route to next phase |
| `decision` | Multiple valid next steps | Request user input or decide |
| `failure` | Phase failed or blocked | Recovery guidance |

### initiative.complexity

Complexity determines which phases apply:

| Team | Levels | Phase Impact |
|------|--------|--------------|
| 10x-dev-pack | SCRIPT | Skip design phase |
| 10x-dev-pack | MODULE+ | Full lifecycle |
| ecosystem-pack | PATCH | Skip design, documentation |
| ecosystem-pack | MODULE+ | Full lifecycle |

### state

Current workflow position:

```yaml
state:
  current_phase: "implementation"
  completed_phases: ["requirements", "design"]
  artifacts_produced:
    - type: prd
      path: "docs/requirements/PRD-user-auth.md"
      status: approved
    - type: tdd
      path: "docs/design/TDD-user-auth.md"
      status: approved
```

### results

For checkpoint requests, summarize what was produced:

```yaml
results:
  phase_completed: "requirements"
  artifact_summary: "PRD-user-auth with 3 success criteria, all testable"
  handoff_criteria_met:
    - criterion_id: prd-001
      status: PASS
    - criterion_id: prd-002
      status: PASS
    - criterion_id: prd-004
      status: PASS
  failure_reason: null  # Not applicable for checkpoint
```

For failure requests:

```yaml
results:
  phase_completed: null  # Did not complete
  artifact_summary: "Partial TDD with incomplete API contracts"
  handoff_criteria_met: []
  failure_reason: "API dependency undocumented, cannot complete design"
```

### context_summary

Brief context for orchestrator (200 words max):

```yaml
context_summary: |
  Implementing user authentication for the platform. PRD approved with
  3 success criteria: registration, login, and rate limiting. TDD in
  progress, blocked on external auth service API documentation. Team
  has confirmed auth service team will provide docs by EOD.
```

## Validation Rules

1. `type` MUST be one of: initial, checkpoint, decision, failure
2. `initiative.name` MUST be non-empty string
3. `initiative.complexity` MUST match team's complexity enum
4. `state.current_phase` MUST be valid phase for the team's workflow
5. `state.completed_phases` MUST be subset of team's phases
6. `results` MUST be present for checkpoint and failure types
7. `results.failure_reason` MUST be present for failure type
8. `context_summary` MUST be <= 200 words

## Example: Initial Request

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
context_summary: "New feature request from product: implement user authentication with email/password login, including rate limiting for failed attempts."
```

## Example: Checkpoint Request

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
    - criterion_id: prd-004
      status: PASS
    - criterion_id: prd-005
      status: PASS
  failure_reason: null
context_summary: "Requirements phase complete. PRD-user-auth approved by product owner. Ready for design phase."
```

## Example: Failure Request

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
  failure_reason: "External auth service API not documented. Cannot complete TDD-003 (API contracts required for SERVICE complexity)."
context_summary: "Design blocked on external dependency. Auth service team committed to providing API docs by tomorrow."
```

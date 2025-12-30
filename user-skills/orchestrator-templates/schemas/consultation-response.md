# CONSULTATION_RESPONSE Schema

> Canonical schema for orchestrator output. Extracted to reduce duplication across team orchestrator.md files.

## Schema Definition

```yaml
# CONSULTATION_RESPONSE
# Output from orchestrator after analyzing CONSULTATION_REQUEST

request_id: string           # Echo from request for correlation
  required: true
  validation: "Must match request.request_id"

directive:
  action: enum               # What the main agent should do next
    values:
      - invoke_specialist    # Delegate to specialist agent
      - request_info         # Need more information to decide
      - await_user           # Need user input to proceed
      - complete             # Initiative is complete
  confidence: number         # Certainty in this directive
    range: [0.5, 1.0]
    required: false          # Default: 1.0 (high confidence)
    interpretation:
      0.9-1.0: "High - clear next step"
      0.7-0.89: "Medium - probable next step, minor ambiguity"
      0.5-0.69: "Low - significant ambiguity, consider clarification"
    behavior: |
      When confidence < 0.7:
      - Main agent should consider requesting_info before proceeding
      - throughline.rationale should explain uncertainty source

specialist:                  # When action is invoke_specialist
  name: string               # Specialist agent identifier
  prompt: string             # Complete prompt for specialist
    # Format:
    # # Context
    # [What specialist needs to know]
    # # Task
    # [What to produce]
    # # Constraints
    # [Scope boundaries]
    # # Handoff Criteria
    # - [ ] Criterion with attestation

information_needed:          # When action is request_info
  - question: string         # What information is needed
    purpose: string          # Why this information matters

user_question:               # When action is await_user
  question: string           # Question for user
  options: array             # Suggested options (may be empty)
    - string

state_update:
  current_phase: string      # Updated phase name
  next_phases: array         # Expected upcoming phases
  routing_rationale: string  # Why this routing was chosen

throughline:
  decision: string           # What was decided (one sentence)
  rationale: string          # Why this decision was made
```

## Field Details

### directive.action

| Action | When to Use | Required Fields |
|--------|-------------|-----------------|
| `invoke_specialist` | Ready to delegate work | `specialist` |
| `request_info` | Cannot decide without more info | `information_needed` |
| `await_user` | Need user approval or choice | `user_question` |
| `complete` | Initiative finished | None (state_update only) |

### specialist

When action is `invoke_specialist`, provide complete prompt:

```yaml
specialist:
  name: architect
  prompt: |
    # Context
    You are designing the authentication system for PRD-user-auth.
    PRD approved with 3 success criteria: registration (SC-001),
    login (SC-002), and rate limiting (SC-003).

    # Task
    Produce TDD-user-auth with:
    - Component breakdown (AuthService, UserRepository)
    - API contracts for /register and /login endpoints
    - Data models for User and Session entities
    - Security considerations for password storage

    # Constraints
    - Must support horizontal scaling (stateless auth)
    - Password hashing with bcrypt, cost factor >= 12
    - Rate limiting at API gateway level

    # Handoff Criteria
    - [ ] TDD has artifact_id matching TDD-{slug} pattern
    - [ ] TDD references PRD-user-auth
    - [ ] At least one component defined
    - [ ] API contracts for all PRD success criteria
    - [ ] TDD status is approved
```

### information_needed

When action is `request_info`:

```yaml
information_needed:
  - question: "What is the expected load for the authentication service?"
    purpose: "Determines whether to design for single-node or distributed session storage"
  - question: "Is OAuth integration required for this phase?"
    purpose: "Affects scope - may need separate PRD if yes"
```

### user_question

When action is `await_user`:

```yaml
user_question:
  question: "Implementation reveals TDD underspecified API error handling. Should we route back to design or proceed with best-effort implementation?"
  options:
    - "Route back to design - update TDD with error handling spec"
    - "Proceed - implement standard REST error responses"
    - "Pause - need stakeholder input on error format"
```

### state_update

Always required - tracks workflow position:

```yaml
state_update:
  current_phase: design
  next_phases: [implementation, validation]
  routing_rationale: "Requirements complete with approved PRD. Complexity is MODULE, so design phase required before implementation."
```

### throughline

Always required - decision audit trail:

```yaml
throughline:
  decision: "Route to architect for TDD creation"
  rationale: "PRD approved with testable criteria. MODULE complexity requires design phase. No blockers identified."
```

## Validation Rules

1. `directive.action` MUST be one of: invoke_specialist, request_info, await_user, complete
2. If action is `invoke_specialist`, `specialist.name` and `specialist.prompt` MUST be present
3. If action is `request_info`, `information_needed` MUST have at least one item
4. If action is `await_user`, `user_question.question` MUST be present
5. `state_update` MUST always be present
6. `throughline` MUST always be present
7. `specialist.prompt` SHOULD be structured with Context, Task, Constraints, Handoff Criteria

## Token Budget

Target: ~400-500 tokens total

| Section | Target Tokens |
|---------|---------------|
| directive | 10 |
| specialist.prompt | 300-350 |
| state_update | 50 |
| throughline | 50 |

If prompt exceeds budget:
- Summarize context, don't enumerate every detail
- Reference artifacts by ID rather than quoting content
- Use skill references for standard patterns

## Example: invoke_specialist

```yaml
directive:
  action: invoke_specialist

specialist:
  name: principal-engineer
  prompt: |
    # Context
    Implementing user authentication per TDD-user-auth.
    Design includes AuthService (login, logout), UserRepository,
    and Session entity. Using JWT with 15-min access tokens.

    # Task
    Implement the authentication system:
    - AuthService with login() and logout() methods
    - UserRepository with CRUD for User entity
    - Session management with JWT signing
    - Password hashing with bcrypt (cost 12)

    # Constraints
    - Follow TDD component interfaces exactly
    - All public methods must have unit tests
    - Integration tests for login/logout flows
    - No external dependencies beyond TDD spec

    # Handoff Criteria
    - [ ] All TDD components implemented
    - [ ] Unit test coverage >= 80%
    - [ ] Integration tests pass
    - [ ] No linting errors
    - [ ] Code compiles/runs without errors

state_update:
  current_phase: implementation
  next_phases: [validation]
  routing_rationale: "Design approved. All TDD criteria met. Ready for implementation phase."

throughline:
  decision: "Route to principal-engineer for implementation"
  rationale: "TDD-user-auth approved with complete API contracts. Implementation can proceed with clear specifications."
```

## Example: await_user

```yaml
directive:
  action: await_user

user_question:
  question: "Validation revealed PRD-user-auth SC-003 (rate limiting) is ambiguous. Should we clarify requirements or proceed with standard implementation?"
  options:
    - "Route back to requirements - PRD needs rate limit specifics"
    - "Proceed with industry standard (5 attempts / 15 minutes)"
    - "Skip SC-003 - address in future iteration"

state_update:
  current_phase: validation
  next_phases: [requirements]  # If routed back
  routing_rationale: "QA found ambiguity in success criterion. User decision required to determine routing."

throughline:
  decision: "Escalate to user for requirement clarification"
  rationale: "SC-003 says 'rate limited' but doesn't specify thresholds. Cannot validate without clear criteria."
```

## Example: complete

```yaml
directive:
  action: complete

state_update:
  current_phase: null
  next_phases: []
  routing_rationale: "All phases complete. PRD approved, TDD approved, implementation passing tests, QA validation passed."

throughline:
  decision: "Initiative complete"
  rationale: "All workflow phases finished. TEST-user-auth shows 100% coverage of success criteria. No blocking issues."
```

# CONSULTATION_RESPONSE Schema

> Output from orchestrator after analyzing request.

## Schema

```yaml
request_id: string           # Echo from request

directive:
  action: enum               # invoke_specialist | request_info | await_user | complete
  confidence: number         # 0.5-1.0 (optional, default 1.0)

specialist:                  # When action: invoke_specialist
  name: string               # Specialist identifier
  prompt: string             # Complete prompt with Context/Task/Constraints/Handoff

information_needed:          # When action: request_info
  - question: string
    purpose: string

user_question:               # When action: await_user
  question: string
  options: array             # Suggested options (may be empty)

state_update:                # Always required
  current_phase: string      # Updated phase
  next_phases: array         # Expected upcoming phases
  routing_rationale: string  # Why this routing

throughline:                 # Always required
  decision: string           # What was decided (one sentence)
  rationale: string          # Why this decision
```

## Directive Actions

| Action | When | Required Fields |
|--------|------|-----------------|
| `invoke_specialist` | Ready to delegate | `specialist` |
| `request_info` | Need more info | `information_needed` |
| `await_user` | Need user choice | `user_question` |
| `complete` | Initiative done | None |

## Confidence Levels

- **0.9-1.0**: High - clear next step
- **0.7-0.89**: Medium - probable, minor ambiguity
- **0.5-0.69**: Low - significant ambiguity, consider clarification

When confidence < 0.7, main agent should consider requesting info before proceeding.

## Examples

**invoke_specialist:**
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
  routing_rationale: "Design approved. All TDD criteria met. Ready for implementation."

throughline:
  decision: "Route to principal-engineer for implementation"
  rationale: "TDD-user-auth approved with complete API contracts."
```

**await_user:**
```yaml
directive:
  action: await_user

user_question:
  question: "Validation revealed SC-003 (rate limiting) is ambiguous. Clarify or proceed with standard?"
  options:
    - "Route back to requirements - specify rate limits"
    - "Proceed with industry standard (5 attempts / 15 min)"
    - "Skip SC-003 - address in future iteration"

state_update:
  current_phase: validation
  next_phases: [requirements]
  routing_rationale: "QA found ambiguity. User decision required."

throughline:
  decision: "Escalate to user for requirement clarification"
  rationale: "SC-003 says 'rate limited' but doesn't specify thresholds."
```

**complete:**
```yaml
directive:
  action: complete

state_update:
  current_phase: null
  next_phases: []
  routing_rationale: "All phases complete. PRD, TDD, implementation, and QA validation passed."

throughline:
  decision: "Initiative complete"
  rationale: "All workflow phases finished. TEST-user-auth shows 100% coverage."
```

## Failure Recovery Patterns

When main agent reports failure (type: "failure"):

1. **Understand**: Read failure_reason carefully
2. **Diagnose**: Check failure_pattern (blocker, scope, capacity, underspecified)
3. **Recover**: Generate new specialist prompt addressing issue OR recommend phase rollback
4. **Document**: Include diagnosis in throughline.rationale

**Do NOT attempt to fix issues yourself** - provide guidance to main agent.

### Pattern Reference

| Pattern | Cause | Recovery Action |
|---------|-------|-----------------|
| `blocker` | External dependency blocked | Return `await_user` with escalation |
| `scope` | Scope exceeds capacity | Decompose into sub-phases |
| `capacity` | Insufficient context | Return `request_info` with needs |
| `underspecified` | Requirements ambiguous | Route back to requirements phase |

## Validation Rules

1. `directive.action` MUST be: invoke_specialist, request_info, await_user, or complete
2. If action is `invoke_specialist`: `specialist.name` and `specialist.prompt` required
3. If action is `request_info`: `information_needed` must have ≥1 item
4. If action is `await_user`: `user_question.question` required
5. `state_update` MUST always be present
6. `throughline` MUST always be present
7. `specialist.prompt` SHOULD use Context/Task/Constraints/Handoff structure

## Token Budget

Target: ~400-500 tokens total per response

| Section | Target Tokens |
|---------|---------------|
| directive | 10 |
| specialist.prompt | 300-350 |
| state_update | 50 |
| throughline | 50 |

**If exceeding budget:**
- Summarize context, don't enumerate
- Reference artifacts by ID rather than quoting
- Use skill references for standard patterns

## Quick Checks

- [ ] Am I returning structured YAML (CONSULTATION_RESPONSE)?
- [ ] Does my directive contain a specialist prompt (not implementation)?
- [ ] Have I updated state_update with current/next phases?
- [ ] Is throughline.rationale explaining *why* this routing?
- [ ] Have I avoided using tools beyond Read?

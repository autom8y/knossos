# Failure Recovery Patterns

This document defines recovery paths for each `failure_pattern` value when specialists fail.

## Decision Tree

```
Specialist Failure Received
          │
          ▼
┌─────────────────────────────────┐
│ What is the failure_pattern?    │
└─────────────────────────────────┘
          │
    ┌─────┼─────┬─────┬─────┐
    ▼     ▼     ▼     ▼     ▼
blocker scope capacity underspec unknown
    │     │     │      │        │
    ▼     ▼     ▼      ▼        ▼
escalate decompose gather route  diagnose
```

## Pattern: `blocker`

**Definition**: External dependency prevents progress (API unavailable, credentials missing, third-party service down, user approval needed).

**Characteristics**:
- Specialist cannot proceed regardless of context quality
- Issue is outside specialist's control
- Waiting is the only resolution

**Recovery Actions**:

1. **Return `await_user` directive** with:
   - Clear description of blocker
   - Specific action user must take
   - Expected timeline for resolution
   - Alternative approaches (if any exist)

2. **Update state**:
   - Mark current phase as `blocked`
   - Record blocker in `throughline.blockers`
   - Do NOT advance to next phase

**Example Response**:
```yaml
directive:
  action: await_user
  reason: "External API credentials required"
user_question: "The OAuth integration requires API keys. Please provide CLIENT_ID and CLIENT_SECRET for the target service, or indicate if we should proceed with mock integration."
state_update:
  current_phase: implementation
  phase_status: blocked
  blockers:
    - "Missing OAuth credentials for payment gateway"
```

---

## Pattern: `scope`

**Definition**: Task scope exceeds what a single specialist invocation can handle.

**Characteristics**:
- Specialist started but couldn't complete
- Work is valid but needs decomposition
- No external blockers

**Recovery Actions**:

1. **Decompose into sub-phases**:
   - Identify logical boundaries in the work
   - Create ordered sub-tasks
   - Update next_phases with decomposed steps

2. **Return `invoke_specialist` with narrower scope**:
   - Reference what was already completed
   - Focus on one sub-task at a time
   - Include completion criteria for this sub-task

**Example Response**:
```yaml
directive:
  action: invoke_specialist
  reason: "Decomposing large scope into focused tasks"
specialist:
  agent: principal-engineer
  prompt: |
    CONTINUATION: Phase 1 of 3 - Database Schema

    The original scope has been decomposed:
    1. [THIS] Database schema and migrations
    2. [NEXT] API endpoint implementation
    3. [NEXT] Integration tests

    Focus ONLY on database schema. Produce migration files.
    Do NOT implement API endpoints in this phase.
state_update:
  next_phases:
    - implementation-db
    - implementation-api
    - implementation-tests
```

---

## Pattern: `capacity`

**Definition**: Specialist lacks information needed to proceed.

**Characteristics**:
- Context provided was insufficient
- Specialist knows what's missing
- Information exists but wasn't included

**Recovery Actions**:

1. **Identify missing information** from failure_reason

2. **Return `request_info`** with:
   - Specific files/artifacts needed
   - Questions to answer
   - Format expected for the information

3. **Do NOT re-invoke specialist** until info gathered

**Example Response**:
```yaml
directive:
  action: request_info
  reason: "Specialist needs additional context"
information_needed:
  - file: "src/api/auth.ts"
    reason: "Need to understand existing auth middleware signature"
  - question: "Should the new endpoint require admin privileges or user-level auth?"
    options:
      - "Admin only (requires role check)"
      - "Any authenticated user"
      - "Public (no auth required)"
state_update:
  current_phase: implementation
  phase_status: awaiting_info
```

---

## Pattern: `underspecified`

**Definition**: Requirements are ambiguous or incomplete.

**Characteristics**:
- Specialist cannot determine correct behavior
- Multiple valid interpretations exist
- PRD/requirements need clarification

**Recovery Actions**:

1. **Route back to requirements phase**:
   - Craft prompt for requirements-analyst
   - Include specific ambiguities found
   - Request updated PRD section

2. **Alternative: Escalate to user** if requirements-analyst already consulted

**Example Response**:
```yaml
directive:
  action: invoke_specialist
  reason: "Routing back to requirements for clarification"
specialist:
  agent: requirements-analyst
  prompt: |
    CLARIFICATION NEEDED

    During implementation, the following ambiguity was discovered:

    PRD Section 3.2 states: "Users can export data in multiple formats"

    Questions requiring clarification:
    1. Which formats exactly? (CSV, JSON, XML, PDF?)
    2. Should exports be synchronous or background jobs?
    3. Are there file size limits?

    Please update PRD-xxx.md Section 3.2 with explicit answers.
state_update:
  current_phase: requirements
  routing_rationale: "Implementation blocked by ambiguous export requirements"
  throughline:
    back_route: "implementation → requirements (underspecified)"
```

---

## Pattern: Unknown/Unclassified

**Definition**: Failure doesn't fit known patterns, or failure_pattern not provided.

**Recovery Actions**:

1. **Diagnose from failure_reason text**:
   - Look for keywords: "missing", "unclear", "blocked", "too large"
   - Map to closest pattern

2. **If still unclear, return `await_user`**:
   - Present failure details
   - Ask user to classify or provide guidance

**Example Response**:
```yaml
directive:
  action: await_user
  reason: "Unable to classify failure pattern"
user_question: |
  The specialist encountered an unexpected failure:

  > {failure_reason}

  How should we proceed?
  1. Retry with additional context (specify what)
  2. Decompose the task differently
  3. Escalate to manual resolution
  4. Skip this phase and continue
```

---

## Recovery Priority Order

When multiple recovery options exist:

1. **Try `request_info` first** (cheapest - just gather info)
2. **Then `invoke_specialist`** (retry with better prompt)
3. **Then back-route** (requirements/design iteration)
4. **Finally `await_user`** (last resort - blocks workflow)

## Anti-Patterns

- **Immediate escalation**: Don't `await_user` for issues that can be resolved with better prompts
- **Retry without changes**: Don't re-invoke specialist with identical context
- **Scope expansion**: Don't add scope when recovering from `scope` pattern - decompose instead
- **Ignoring failure_pattern**: Always check this field first when present

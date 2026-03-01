---
artifact_id: TDD-fate-skills
title: "Fate Skills Architecture: Progressive Disclosure for Unified Moirai"
created_at: "2026-01-07T23:00:00Z"
author: architect
prd_ref: PRD-moirai-consolidation
status: draft
components:
  - name: SKILL.md
    type: skill
    description: "Moirai routing table and skill discovery entry point"
    dependencies:
      - name: clotho.md
        type: internal
      - name: lachesis.md
        type: internal
      - name: atropos.md
        type: internal
  - name: clotho.md
    type: skill
    description: "Creation operations guidance (create_sprint, start_sprint)"
    dependencies: []
  - name: lachesis.md
    type: skill
    description: "Measurement operations guidance (mark_complete, transition_phase, etc.)"
    dependencies: []
  - name: atropos.md
    type: skill
    description: "Termination operations guidance (wrap_session, generate_sails, delete_sprint)"
    dependencies: []
related_adrs:
  - ADR-0009
schema_version: "1.0"
---

# TDD: Fate Skills Architecture

> The Moirai consolidation transforms four agents into one unified Moirai agent that loads domain-specific guidance on-demand via skills. This TDD specifies the three Fate skills (Clotho, Lachesis, Atropos) that Moirai reads for operation-specific logic.

**Status**: Draft
**Author**: Architect
**Date**: 2026-01-07
**PRD Reference**: PRD-moirai-consolidation

---

## 1. Overview

### 1.1 Purpose

The Fate skills provide operation-specific guidance that the unified Moirai agent loads on-demand. Unlike the previous architecture where Clotho, Lachesis, and Atropos were standalone agents invoked via `Task()`, they now become internal reference materials that Moirai reads via the skill system.

**Key Distinction**: Skills provide *guidance*, not *execution*. Moirai reads the skill content to understand how to perform an operation, then executes the operation itself.

### 1.2 Design Goals

1. **Progressive Disclosure**: Moirai loads only the relevant Fate skill per operation, minimizing context consumption
2. **Single Execution Point**: All operations execute within Moirai--no subagent spawning
3. **User Invisibility**: Users never invoke Fate skills directly; they use slash commands or `Task(moirai, ...)`
4. **Preserved Semantics**: All operation logic, validation rules, and response formats from current Fate agents are preserved
5. **CLI Authority**: Moirai delegates to `ari` CLI for authoritative state changes

### 1.3 Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     User / Main Thread                           │
└─────────────────────────────────────────────────────────────────┘
           │                                    │
           │ /park, /wrap, /handoff             │ Task(moirai, "...")
           ▼                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Unified Moirai Agent                          │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │ 1. Parse operation from input                              │  │
│  │ 2. Read SKILL.md for routing table                         │  │
│  │ 3. Read relevant Fate skill (clotho|lachesis|atropos)      │  │
│  │ 4. Execute operation following skill guidance               │  │
│  │ 5. Delegate to ari CLI for state changes                    │  │
│  │ 6. Return structured JSON response                          │  │
│  └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
           │
           │ Read (internal)
           ▼
┌─────────────────────────────────────────────────────────────────┐
│                    .claude/skills/moirai/                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │  SKILL.md   │  │ clotho.md   │  │ lachesis.md │              │
│  │  (routing)  │  │ (creation)  │  │(measurement)│              │
│  └─────────────┘  └─────────────┘  └─────────────┘              │
│                                    ┌─────────────┐              │
│                                    │ atropos.md  │              │
│                                    │(termination)│              │
│                                    └─────────────┘              │
└─────────────────────────────────────────────────────────────────┘
```

---

## 2. Directory Structure

### 2.1 File Layout

```
.claude/skills/moirai/
├── SKILL.md          # Entry point with routing table
├── clotho.md         # Creation operations (Spinner)
├── lachesis.md       # Measurement operations (Measurer)
└── atropos.md        # Termination operations (Cutter)
```

### 2.2 Location Rationale

| Location | Rationale |
|----------|-----------|
| `.claude/skills/moirai/` | Project-level skills, accessible to Moirai agent |
| Not `~/.claude/skills/` | These are project-specific, not user-global |
| Not `user-agents/` | Skills, not agents--loaded via Read, not Task |

### 2.3 Naming Conventions

| File | Purpose |
|------|---------|
| `SKILL.md` | Standard skill entry point (UPPERCASE per convention) |
| `clotho.md` | Fate name in lowercase (matches mythological convention) |
| `lachesis.md` | Fate name in lowercase |
| `atropos.md` | Fate name in lowercase |

---

## 3. Skill Interface Contract

### 3.1 How Moirai Loads Skills

Moirai uses the Read tool to load skill content. This is NOT `Skill()` tool invocation--it is direct file reading:

```
# Moirai's internal process:

1. Receive operation request (e.g., "park_session reason='taking break'")

2. Parse operation name: "park_session"

3. Read routing table:
   Read(.claude/skills/moirai/SKILL.md)
   → Lookup park_session → lachesis

4. Read domain skill:
   Read(.claude/skills/moirai/lachesis.md)
   → Extract park_session specification

5. Execute operation following loaded guidance

6. Delegate to CLI:
   Bash(ari session park --reason "taking break")

7. Return structured JSON response
```

### 3.2 Skill Content Structure

Each Fate skill follows a consistent structure:

```markdown
---
name: {fate-name}
domain: {creation|measurement|termination}
operations: [{operation-list}]
---

# {Fate Name} - The {Title}

## Operations

### {operation_name}
**Syntax**: `{structured-syntax}`
**Parameters**: {parameter-list}
**Validation**: {validation-rules}
**CLI Command**: {ari-command}
**Success Response**: {json-example}
**Error Response**: {json-example}
**Example**: {usage-example}
```

### 3.3 Skill Loading Pattern

**Progressive Disclosure Flow**:

```
┌──────────────────────────────────────────────────────────┐
│                    Moirai Agent Context                   │
│                                                          │
│  Base Prompt:                                            │
│  - Operation parsing rules                               │
│  - JSON response format                                  │
│  - CLI delegation protocol                               │
│  - Error handling patterns                               │
│                                                          │
│  NOT loaded initially:                                   │
│  - Specific operation validation rules                   │
│  - Operation-specific CLI commands                       │
│  - Fate-specific response schemas                        │
└──────────────────────────────────────────────────────────┘
            │
            │ On receiving "park_session"
            ▼
┌──────────────────────────────────────────────────────────┐
│  Loaded on demand:                                        │
│  - SKILL.md (routing table)           ~20 lines           │
│  - lachesis.md (park_session section) ~50 lines           │
│                                                          │
│  Total context added: ~70 lines                          │
│  (vs. loading all 3 skills: ~300 lines)                  │
└──────────────────────────────────────────────────────────┘
```

### 3.4 Response Format Contract

All Fate skills specify the same response schema (from moirai-shared.md):

**Success Response**:
```json
{
  "success": true,
  "operation": "string",
  "message": "string",
  "reasoning": "string",
  "fate": "clotho|lachesis|atropos",
  "state_before": {},
  "state_after": {},
  "changes": {}
}
```

**Error Response**:
```json
{
  "success": false,
  "operation": "string",
  "error_code": "string",
  "message": "string",
  "reasoning": "string",
  "hint": "string"
}
```

---

## 4. Entry Point (SKILL.md)

### 4.1 Purpose

SKILL.md serves as the routing table and discovery entry point. Moirai reads this first to determine which Fate skill contains the requested operation.

### 4.2 Specification

```markdown
---
name: moirai
description: |
  Moirai internal skills for session lifecycle management. These skills provide
  operation-specific guidance for the unified Moirai agent. They are NOT user-invokable.
internal: true
---

# Moirai Skills

> The three Fates govern session lifecycle. This routing table maps operations to domains.

## Routing Table

| Operation | Fate | Domain | CLI Command |
|-----------|------|--------|-------------|
| create_sprint | clotho | creation | - |
| start_sprint | clotho | creation | - |
| mark_complete | lachesis | measurement | - |
| transition_phase | lachesis | measurement | ari session transition |
| update_field | lachesis | measurement | - |
| park_session | lachesis | measurement | ari session park |
| resume_session | lachesis | measurement | ari session resume |
| handoff | lachesis | measurement | ari handoff execute |
| record_decision | lachesis | measurement | - |
| append_content | lachesis | measurement | - |
| wrap_session | atropos | termination | ari session wrap |
| generate_sails | atropos | termination | ari sails check |
| delete_sprint | atropos | termination | - |

## Domain Files

- **clotho.md**: Creation operations (spinning new entities)
- **lachesis.md**: Measurement operations (tracking state changes)
- **atropos.md**: Termination operations (ending and archiving)

## Loading Protocol

1. Parse operation from user input
2. Lookup operation in routing table above
3. Read the corresponding domain file
4. Follow operation specification in domain file
5. Delegate to CLI command if specified
6. Return structured JSON response

## Error Codes

| Code | Description |
|------|-------------|
| INVALID_OPERATION | Operation not in routing table |
| SCHEMA_VIOLATION | State change would violate schema |
| LIFECYCLE_VIOLATION | State transition not allowed |
| DEPENDENCY_BLOCKED | Blocked by unmet dependency |
| LOCK_TIMEOUT | Could not acquire file lock |
| FILE_NOT_FOUND | Target context file missing |
| VALIDATION_FAILED | Pre-mutation validation failed |

## Control Flags

| Flag | Effect |
|------|--------|
| --dry-run | Preview mutation without applying |
| --emergency | Bypass non-critical validations |
| --override=reason | Bypass lifecycle rules with reason |
```

### 4.3 Size Target

SKILL.md: ~50 lines (compact routing table, no operation details)

---

## 5. Clotho Skill Specification

### 5.1 Purpose

Clotho (the Spinner) governs **creation operations**--spinning new sessions and sprints into existence.

### 5.2 Operations

| Operation | Description | CLI Command |
|-----------|-------------|-------------|
| create_sprint | Create new sprint within session | None (file write) |
| start_sprint | Activate pending sprint | None (file update) |

### 5.3 Full Specification

```markdown
---
name: clotho
domain: creation
operations: [create_sprint, start_sprint]
---

# Clotho - The Spinner

> Ariadne gave Theseus the thread as a gift. I am who spins it.

Clotho governs **creation and initialization**--bringing sessions and sprints into existence.

---

## Operations

### create_sprint

Creates a new sprint within the current session.

**Syntax**:
```
create_sprint name="Sprint Name" [depends_on=sprint-id]
```

**Parameters**:
| Parameter | Required | Description |
|-----------|----------|-------------|
| name | Yes | Human-readable sprint name |
| depends_on | No | Sprint ID this sprint depends on |

**Validation**:
1. Session must exist and be ACTIVE
2. If depends_on specified, dependency sprint must exist and be completed
3. Sprint name must not duplicate existing sprint in session
4. Sprint ID generated as: `sprint-{sanitized-name}-{YYYYMMDD}`

**File Creation**:
```
.sos/sessions/{session-id}/sprints/{sprint-id}/SPRINT_CONTEXT.md
```

**Sprint Context Initial State**:
```yaml
sprint_id: "{generated-id}"
name: "{name}"
status: pending
depends_on: "{depends_on | null}"
created_at: "{ISO-8601}"
started_at: null
completed_at: null
tasks: []
```

**CLI Command**: None (direct file creation)

**Success Response**:
```json
{
  "success": true,
  "operation": "create_sprint",
  "message": "Sprint '{name}' created",
  "reasoning": "Created new sprint with dependency on completed design sprint",
  "fate": "clotho",
  "state_before": {
    "sprint_count": 1
  },
  "state_after": {
    "sprint_count": 2,
    "new_sprint_id": "sprint-impl-20260107"
  },
  "changes": {
    "sprints": "+sprint-impl-20260107"
  }
}
```

**Error Responses**:

| Condition | Error Code | Message |
|-----------|------------|---------|
| Session not ACTIVE | LIFECYCLE_VIOLATION | Session must be ACTIVE to create sprints |
| Dependency not found | DEPENDENCY_BLOCKED | Sprint '{depends_on}' not found |
| Dependency not complete | DEPENDENCY_BLOCKED | Sprint '{depends_on}' must be completed first |
| Duplicate name | VALIDATION_FAILED | Sprint with name '{name}' already exists |

**Example**:
```
Input: create_sprint name="API Implementation" depends_on=sprint-design-20260106

Validation:
1. Read SESSION_CONTEXT.md -> session_state: ACTIVE (pass)
2. Read sprint-design-20260106/SPRINT_CONTEXT.md -> status: completed (pass)
3. Glob for existing sprint names -> no duplicate (pass)

Create:
Write .sos/sessions/{session-id}/sprints/sprint-api-impl-20260107/SPRINT_CONTEXT.md

Log:
2026-01-07T10:00:00Z | session-abc | create_sprint | moirai | name="API Implementation" | SUCCESS | clotho
```

---

### start_sprint

Activates a pending sprint, setting started_at timestamp.

**Syntax**:
```
start_sprint sprint_id
```

**Parameters**:
| Parameter | Required | Description |
|-----------|----------|-------------|
| sprint_id | Yes | Sprint identifier to start |

**Validation**:
1. Sprint must exist
2. Sprint must be in `pending` status
3. All dependencies must be completed

**State Transition**: pending -> active

**CLI Command**: None (direct file update)

**Success Response**:
```json
{
  "success": true,
  "operation": "start_sprint",
  "message": "Sprint '{sprint_id}' started",
  "reasoning": "Sprint activated, dependencies satisfied",
  "fate": "clotho",
  "state_before": {
    "status": "pending",
    "started_at": null
  },
  "state_after": {
    "status": "active",
    "started_at": "2026-01-07T10:00:00Z"
  },
  "changes": {
    "status": "pending -> active",
    "started_at": "null -> 2026-01-07T10:00:00Z"
  }
}
```

**Error Responses**:

| Condition | Error Code | Message |
|-----------|------------|---------|
| Sprint not found | FILE_NOT_FOUND | Sprint '{sprint_id}' not found |
| Not pending | LIFECYCLE_VIOLATION | Sprint must be pending to start |
| Dependency incomplete | DEPENDENCY_BLOCKED | Dependency '{depends_on}' not completed |

---

## Anti-Patterns

| Anti-Pattern | Correct Behavior |
|--------------|------------------|
| Create existing sprint | Return VALIDATION_FAILED, not overwrite |
| Modify existing state | Clotho creates; Lachesis measures |
| Delete or archive | Clotho spins; Atropos cuts |
| Create sessions | Sessions created by ari session create |

---

## Natural Language Mapping

| Input | Operation |
|-------|-----------|
| "create a new sprint called X" | create_sprint name="X" |
| "new sprint for implementation" | create_sprint name="Implementation" |
| "spin up a testing sprint" | create_sprint name="Testing" |
| "start the implementation sprint" | start_sprint sprint-impl-* |
| "begin sprint X" | start_sprint X |
| "activate the testing sprint" | start_sprint sprint-testing-* |
```

### 5.4 Size Target

clotho.md: ~100 lines (2 operations, compact)

---

## 6. Lachesis Skill Specification

### 6.1 Purpose

Lachesis (the Measurer) governs **measurement operations**--tracking progress, recording milestones, and managing state transitions.

### 6.2 Operations

| Operation | Description | CLI Command |
|-----------|-------------|-------------|
| mark_complete | Record task completion | None |
| transition_phase | Progress workflow phase | ari session transition |
| update_field | Update context field | None |
| park_session | Pause session | ari session park |
| resume_session | Resume session | ari session resume |
| handoff | Record agent transition | ari handoff execute |
| record_decision | Log decision | None |
| append_content | Append to context body | None |

### 6.3 Full Specification

```markdown
---
name: lachesis
domain: measurement
operations: [mark_complete, transition_phase, update_field, park_session, resume_session, handoff, record_decision, append_content]
---

# Lachesis - The Measurer

> I measure what Clotho spins and record until Atropos cuts.

Lachesis governs **measurement and tracking**--recording every milestone, transition, and decision in the journey through the labyrinth.

---

## Operations

### mark_complete

Records task completion with artifact reference.

**Syntax**:
```
mark_complete task_id artifact=path
```

**Parameters**:
| Parameter | Required | Description |
|-----------|----------|-------------|
| task_id | Yes | Task identifier |
| artifact | Yes | Path to produced artifact |

**Validation**:
1. Task must exist in session/sprint context
2. Task must not already be completed
3. Artifact path should exist (warning if missing)

**CLI Command**: None (direct file update)

**Success Response**:
```json
{
  "success": true,
  "operation": "mark_complete",
  "message": "Task {task_id} marked complete",
  "reasoning": "Task marked complete per request. Artifact validated at path.",
  "fate": "lachesis",
  "state_before": {
    "task_status": "pending"
  },
  "state_after": {
    "task_status": "completed",
    "completed_at": "2026-01-07T10:00:00Z",
    "artifact": "docs/requirements/PRD-foo.md"
  },
  "changes": {
    "status": "pending -> completed",
    "completed_at": "null -> 2026-01-07T10:00:00Z"
  }
}
```

**Error Responses**:

| Condition | Error Code | Message |
|-----------|------------|---------|
| Task not found | FILE_NOT_FOUND | Task '{task_id}' not found in context |
| Already complete | LIFECYCLE_VIOLATION | Task '{task_id}' already completed |
| Missing artifact | VALIDATION_FAILED | Artifact path does not exist (warning) |

---

### transition_phase

Records workflow phase transition.

**Syntax**:
```
transition_phase to=phase [from=current_phase]
```

**Parameters**:
| Parameter | Required | Description |
|-----------|----------|-------------|
| to | Yes | Target phase |
| from | No | Current phase (validated against actual) |

**Valid Phases**: requirements, design, implementation, testing, deployment

**Phase Transition Rules**:
```
requirements -> design -> implementation -> testing -> deployment
     ^                           |              |
     |                           v              v
     +---- (feedback loops allowed) -----------+
```

**CLI Command**: `ari session transition {phase}`

**Success Response**:
```json
{
  "success": true,
  "operation": "transition_phase",
  "message": "Phase transitioned from 'design' to 'implementation'",
  "reasoning": "Design phase complete, implementation ready to begin",
  "fate": "lachesis",
  "state_before": {
    "current_phase": "design"
  },
  "state_after": {
    "current_phase": "implementation",
    "phase_history": ["requirements", "design", "implementation"]
  },
  "changes": {
    "current_phase": "design -> implementation"
  }
}
```

---

### update_field

Generic field update for session/sprint context.

**Syntax**:
```
update_field field_name=value [field2=value2 ...]
```

**Parameters**:
| Parameter | Required | Description |
|-----------|----------|-------------|
| field=value | Yes | One or more field assignments |

**Validation**:
1. Field must be defined in schema
2. Value must pass schema validation
3. Read-only fields rejected (session_id, created_at)

**CLI Command**: None (direct file update)

**Read-Only Fields**:
- session_id
- created_at
- sprint_id

---

### park_session

Records session pause with reason.

**Syntax**:
```
park_session reason="reason text"
```

**Parameters**:
| Parameter | Required | Description |
|-----------|----------|-------------|
| reason | Yes | Why the session is being parked |

**Validation**:
1. Session must be ACTIVE
2. Reason must be provided (non-empty)

**State Transition**: ACTIVE -> PARKED

**CLI Command**: `ari session park --reason "{reason}"`

**Success Response**:
```json
{
  "success": true,
  "operation": "park_session",
  "message": "Session parked",
  "reasoning": "User requested park for urgent bug fix",
  "fate": "lachesis",
  "state_before": {
    "session_state": "ACTIVE"
  },
  "state_after": {
    "session_state": "PARKED",
    "parked_at": "2026-01-07T10:00:00Z",
    "park_reason": "Handling urgent bug"
  },
  "changes": {
    "session_state": "ACTIVE -> PARKED",
    "parked_at": "null -> 2026-01-07T10:00:00Z"
  }
}
```

---

### resume_session

Records session resumption from parked state.

**Syntax**:
```
resume_session
```

**Parameters**: None

**Validation**:
1. Session must be PARKED

**State Transition**: PARKED -> ACTIVE

**CLI Command**: `ari session resume`

**Success Response**:
```json
{
  "success": true,
  "operation": "resume_session",
  "message": "Session resumed",
  "reasoning": "Session unparked, resuming work",
  "fate": "lachesis",
  "state_before": {
    "session_state": "PARKED"
  },
  "state_after": {
    "session_state": "ACTIVE",
    "resumed_at": "2026-01-07T11:00:00Z"
  },
  "changes": {
    "session_state": "PARKED -> ACTIVE"
  }
}
```

---

### handoff

Records agent-to-agent transition.

**Syntax**:
```
handoff to=agent_name [note="handoff notes"]
```

**Parameters**:
| Parameter | Required | Description |
|-----------|----------|-------------|
| to | Yes | Target agent name |
| note | No | Context for the handoff |

**Valid Agents**: orchestrator, requirements-analyst, architect, principal-engineer, qa-adversary

**CLI Command**: `ari handoff execute --from {current} --to {target}`

**Success Response**:
```json
{
  "success": true,
  "operation": "handoff",
  "message": "Handoff recorded to 'principal-engineer'",
  "reasoning": "Design complete, transitioning to implementation",
  "fate": "lachesis",
  "state_before": {
    "current_agent": "architect"
  },
  "state_after": {
    "current_agent": "principal-engineer",
    "handoff_history": [
      {"from": "architect", "to": "principal-engineer", "at": "2026-01-07T10:00:00Z", "note": "TDD approved"}
    ]
  },
  "changes": {
    "current_agent": "architect -> principal-engineer"
  }
}
```

---

### record_decision

Appends a decision to the session context.

**Syntax**:
```
record_decision "decision text"
```

**Parameters**:
| Parameter | Required | Description |
|-----------|----------|-------------|
| (positional) | Yes | Decision text to record |

**CLI Command**: None (direct file update)

**Success Response**:
```json
{
  "success": true,
  "operation": "record_decision",
  "message": "Decision recorded",
  "reasoning": "Recording architectural decision for audit trail",
  "fate": "lachesis",
  "state_before": {
    "decisions_count": 2
  },
  "state_after": {
    "decisions_count": 3,
    "latest_decision": "Use event sourcing for audit log"
  },
  "changes": {
    "decisions": "+1"
  }
}
```

---

### append_content

Appends markdown content to the context body.

**Syntax**:
```
append_content "markdown content"
```

**Parameters**:
| Parameter | Required | Description |
|-----------|----------|-------------|
| (positional) | Yes | Markdown content to append |

**CLI Command**: None (direct file update)

---

## Anti-Patterns

| Anti-Pattern | Correct Behavior |
|--------------|------------------|
| Guess state before reading | Always read before writing |
| Skip timestamps | Every mutation has a time |
| Silent updates | Every field change is logged |
| Create new entities | Lachesis measures; Clotho creates |
| Delete or archive | Lachesis measures; Atropos cuts |

---

## Natural Language Mapping

| Input | Operation |
|-------|-----------|
| "mark task X complete" | mark_complete X artifact=... |
| "the PRD is done" | mark_complete task-prd artifact=... |
| "move to implementation phase" | transition_phase to=implementation |
| "park the session" | park_session reason="..." |
| "pause for a break" | park_session reason="taking break" |
| "resume work" | resume_session |
| "continue the session" | resume_session |
| "hand off to engineer" | handoff to=principal-engineer |
| "transfer to QA" | handoff to=qa-adversary |
```

### 6.4 Size Target

lachesis.md: ~200 lines (8 operations, largest skill)

---

## 7. Atropos Skill Specification

### 7.1 Purpose

Atropos (the Cutter) governs **termination operations**--ending sessions, generating confidence signals, and archiving completed work.

### 7.2 Operations

| Operation | Description | CLI Command |
|-----------|-------------|-------------|
| wrap_session | Archive session with sails | ari session wrap |
| generate_sails | Compute confidence signal | ari sails check |
| delete_sprint | Remove or archive sprint | None |

### 7.3 Full Specification

```markdown
---
name: atropos
domain: termination
operations: [wrap_session, generate_sails, delete_sprint]
---

# Atropos - The Cutter

> What Clotho spins and Lachesis measures, I cut when complete.

Atropos governs **termination and archival**--ending sessions, generating confidence signals, and sealing the record of the journey.

---

## Operations

### wrap_session

Terminates and archives the session. This is the final cut.

**Syntax**:
```
wrap_session [--emergency]
```

**Parameters**:
| Parameter | Required | Description |
|-----------|----------|-------------|
| --emergency | No | Bypass BLACK sails quality gate |

**Validation**:
1. Session must be ACTIVE (or PARKED with --override)
2. Generates White Sails confidence signal before archival
3. **Quality Gate**: Blocks wrap if sails are BLACK (unless --emergency)

**State Transition**: ACTIVE -> ARCHIVED

**CLI Command**: `ari session wrap` (or `ari session wrap --force` for emergency)

**Internal Flow**:
```
wrap_session invoked
    │
    ├── 1. Invoke ari session wrap
    │       - Collects proofs
    │       - Computes confidence color
    │       - Writes WHITE_SAILS.yaml
    │       - Emits SAILS_GENERATED event
    │
    ├── 2. Quality Gate: Check sails color
    │       - BLACK + no --emergency: BLOCK
    │       - BLACK + --emergency: WARN and continue
    │       - GRAY or WHITE: Proceed
    │
    ├── 3. Update session state
    │       - session_state = ARCHIVED
    │       - archived_at = timestamp
    │
    └── 4. Return result with sails metadata
```

**Success Response (WHITE sails)**:
```json
{
  "success": true,
  "operation": "wrap_session",
  "message": "Session archived with WHITE sails",
  "reasoning": "All work complete, confidence signal computed, session sealed",
  "fate": "atropos",
  "state_before": {
    "session_state": "ACTIVE"
  },
  "state_after": {
    "session_state": "ARCHIVED",
    "archived_at": "2026-01-07T14:00:00Z",
    "sails_color": "WHITE",
    "sails_path": ".sos/sessions/session-abc123/WHITE_SAILS.yaml"
  },
  "changes": {
    "session_state": "ACTIVE -> ARCHIVED",
    "sails_generated": true,
    "sails_color": "WHITE"
  }
}
```

**Error Response (BLACK sails blocked)**:
```json
{
  "success": false,
  "operation": "wrap_session",
  "error_code": "QUALITY_GATE_FAILED",
  "message": "Cannot wrap session with BLACK sails: explicit blockers present",
  "reasoning": "Quality gate prevents archival with known failures. Use --emergency to override.",
  "hint": "Fix blockers and retry, OR use: --emergency wrap_session",
  "sails": {
    "color": "BLACK",
    "reasons": [
      "Tests failing in integration suite",
      "Build broken on macOS"
    ]
  }
}
```

**Wrapping PARKED Session**:
Requires --override flag:
```
--override=reason="User confirmed" wrap_session
```

---

### generate_sails

Computes and generates the White Sails confidence signal.

**Syntax**:
```
generate_sails [--skip-proofs] [--modifier=TYPE:JUSTIFICATION]
```

**Parameters**:
| Parameter | Required | Description |
|-----------|----------|-------------|
| --skip-proofs | No | Skip proof collection (spike sessions) |
| --modifier | No | Apply modifier with justification |

**Output Location**:
```
.sos/sessions/{session-id}/WHITE_SAILS.yaml
```

**CLI Command**: `ari sails check`

**Sails Color Computation**:

| Color | Criteria |
|-------|----------|
| WHITE | All proofs pass, no open questions |
| GRAY | Some proofs missing or open questions exist |
| BLACK | Critical failures, explicit blockers |

**Proof Types**:

| Proof | Source | Weight |
|-------|--------|--------|
| tests | Test output logs | High |
| build | Build output logs | High |
| lint | Lint output logs | Medium |
| coverage | Coverage reports | Medium |
| review | Code review status | Medium |

**Modifiers**:
```
--modifier=TIME_PRESSURE:JUSTIFICATION
--modifier=KNOWN_TECH_DEBT:JUSTIFICATION
--modifier=SPIKE_SESSION:JUSTIFICATION
```

**Success Response**:
```json
{
  "success": true,
  "operation": "generate_sails",
  "message": "White Sails generated",
  "reasoning": "All proofs collected, confidence computed",
  "fate": "atropos",
  "sails_path": ".sos/sessions/session-abc123/WHITE_SAILS.yaml",
  "color": "WHITE",
  "computed_base": "WHITE",
  "proofs_collected": ["tests", "build", "lint"],
  "open_questions_found": 0,
  "modifiers_applied": []
}
```

---

### delete_sprint

Removes or archives a sprint.

**Syntax**:
```
delete_sprint sprint_id [--archive]
```

**Parameters**:
| Parameter | Required | Description |
|-----------|----------|-------------|
| sprint_id | Yes | Sprint to delete |
| --archive | No | Move to archive instead of deleting |

**Validation**:
1. Sprint must exist
2. Sprint should be completed (warning if not)
3. No other sprints should depend on it (unless completed)

**CLI Command**: None (direct file operation)

**Success Response (delete)**:
```json
{
  "success": true,
  "operation": "delete_sprint",
  "message": "Sprint '{sprint_id}' deleted",
  "reasoning": "Sprint completed and no longer needed, removed per request",
  "fate": "atropos",
  "state_before": {
    "sprint_exists": true,
    "sprint_status": "completed"
  },
  "state_after": {
    "sprint_exists": false
  },
  "changes": {
    "sprints": "-{sprint_id}"
  }
}
```

**Success Response (archive)**:
```json
{
  "success": true,
  "operation": "delete_sprint",
  "message": "Sprint '{sprint_id}' archived",
  "reasoning": "Sprint completed, preserved in archive for reference",
  "fate": "atropos",
  "state_before": {
    "sprint_path": ".sos/sessions/session-abc/sprints/{sprint_id}/"
  },
  "state_after": {
    "sprint_path": ".sos/sessions/session-abc/archive/{sprint_id}/"
  },
  "changes": {
    "location": "sprints/ -> archive/"
  }
}
```

---

## Anti-Patterns

| Anti-Pattern | Correct Behavior |
|--------------|------------------|
| Cut prematurely | Wrap only when complete or requested |
| Skip sails | Every wrap generates confidence signal |
| Bypass quality gate silently | --emergency required for BLACK sails |
| Modify after cut | ARCHIVED sessions are immutable |
| Create entities | Atropos cuts; Clotho spins |
| Measure progress | Atropos seals; Lachesis measures |

---

## Natural Language Mapping

| Input | Operation |
|-------|-----------|
| "wrap up the session" | wrap_session |
| "finish the session" | wrap_session |
| "complete and archive" | wrap_session |
| "generate confidence signal" | generate_sails |
| "compute the sails" | generate_sails |
| "archive the testing sprint" | delete_sprint sprint-testing-* --archive |
| "delete old sprint" | delete_sprint sprint-* |
```

### 7.4 Size Target

atropos.md: ~150 lines (3 operations, includes sails protocol)

---

## 8. Progressive Disclosure Pattern

### 8.1 Context Efficiency

The progressive disclosure pattern minimizes context consumption:

| Scenario | Files Loaded | Lines |
|----------|--------------|-------|
| Any operation | SKILL.md (routing) | ~50 |
| Creation operation | + clotho.md | ~100 |
| Measurement operation | + lachesis.md | ~200 |
| Termination operation | + atropos.md | ~150 |
| **Worst case** | SKILL.md + one fate | ~250 |
| **Previous architecture** | All 4 agents | ~800+ |

### 8.2 Loading Sequence

```
1. User invokes: /park "taking a break"

2. Slash command skill routes to Moirai:
   Task(moirai, "park_session reason='taking a break'")

3. Moirai parses operation: "park_session"

4. Moirai reads SKILL.md:
   Read(.claude/skills/moirai/SKILL.md)
   → park_session -> lachesis (measurement)

5. Moirai reads lachesis.md:
   Read(.claude/skills/moirai/lachesis.md)
   → Extract park_session specification

6. Moirai executes:
   - Validates session is ACTIVE
   - Runs: ari session park --reason "taking a break"
   - Parses CLI output
   - Returns structured JSON

7. Slash command skill formats response for user
```

### 8.3 User Invisibility

Users never see Fate skills:

| User Action | What User Sees | What Happens |
|-------------|----------------|--------------|
| `/park "break"` | "Session parked" | Task(moirai) -> reads lachesis.md -> ari |
| `/wrap` | "Session archived with WHITE sails" | Task(moirai) -> reads atropos.md -> ari |
| `/handoff qa` | "Handoff to qa-adversary recorded" | Task(moirai) -> reads lachesis.md -> ari |

Users never invoke `Skill(clotho)`, `Skill(lachesis)`, or `Skill(atropos)`.

---

## 9. Migration from Agents

### 9.1 Content Preservation

The following content from current Fate agents is preserved in skills:

| Current Agent | Preserved Content | Location in Skill |
|---------------|-------------------|-------------------|
| clotho.md | Operation syntax | clotho.md Operations section |
| clotho.md | Validation rules | clotho.md per-operation |
| clotho.md | Response format | clotho.md per-operation |
| clotho.md | Anti-patterns | clotho.md Anti-Patterns section |
| lachesis.md | Operation syntax | lachesis.md Operations section |
| lachesis.md | Phase FSM | lachesis.md transition_phase |
| lachesis.md | Fiduciary duty | Absorbed into Moirai agent |
| atropos.md | Sails protocol | atropos.md generate_sails |
| atropos.md | Quality gate | atropos.md wrap_session |
| moirai-shared.md | Error codes | SKILL.md Error Codes |
| moirai-shared.md | Response schema | SKILL.md (reference only) |
| moirai-shared.md | Lock protocol | Absorbed into Moirai agent |
| moirai-shared.md | State machines | Absorbed into skills per-operation |

### 9.2 Content Transformation

**What changes**:

| Before | After |
|--------|-------|
| Agent with `tools:` header | Skill with `operations:` header |
| `Task(clotho, ...)` invocation | `Read(.claude/skills/moirai/clotho.md)` |
| Standalone execution | Guidance only--Moirai executes |
| Natural language in agent | Natural language in skill |
| Tool access defined | No tool access (Read by caller) |

**What stays the same**:

- Operation names and syntax
- Validation logic
- Response JSON format
- Error codes
- Mythological metaphors
- Anti-patterns

### 9.3 Files to Remove After Migration

| File | Reason |
|------|--------|
| user-agents/moirai.md | Replaced by .claude/agents/moirai.md |
| user-agents/clotho.md | Replaced by .claude/skills/moirai/clotho.md |
| user-agents/lachesis.md | Replaced by .claude/skills/moirai/lachesis.md |
| user-agents/atropos.md | Replaced by .claude/skills/moirai/atropos.md |
| user-agents/moirai-shared.md | Absorbed into SKILL.md and agent |
| user-agents/moirai.md.backup | Obsolete backup |

---

## 10. Test Strategy

### 10.1 Skill Loading Tests

| Test ID | Description | Pass Criteria |
|---------|-------------|---------------|
| SKILL-001 | SKILL.md readable | Read tool succeeds, routing table parseable |
| SKILL-002 | clotho.md readable | Read tool succeeds, operations extractable |
| SKILL-003 | lachesis.md readable | Read tool succeeds, operations extractable |
| SKILL-004 | atropos.md readable | Read tool succeeds, operations extractable |
| SKILL-005 | Routing table complete | All 13 operations have fate assignment |

### 10.2 Operation Routing Tests

| Test ID | Operation | Expected Fate | Test Method |
|---------|-----------|---------------|-------------|
| ROUTE-001 | create_sprint | clotho | Parse SKILL.md, verify mapping |
| ROUTE-002 | start_sprint | clotho | Parse SKILL.md, verify mapping |
| ROUTE-003 | mark_complete | lachesis | Parse SKILL.md, verify mapping |
| ROUTE-004 | transition_phase | lachesis | Parse SKILL.md, verify mapping |
| ROUTE-005 | park_session | lachesis | Parse SKILL.md, verify mapping |
| ROUTE-006 | resume_session | lachesis | Parse SKILL.md, verify mapping |
| ROUTE-007 | handoff | lachesis | Parse SKILL.md, verify mapping |
| ROUTE-008 | update_field | lachesis | Parse SKILL.md, verify mapping |
| ROUTE-009 | record_decision | lachesis | Parse SKILL.md, verify mapping |
| ROUTE-010 | append_content | lachesis | Parse SKILL.md, verify mapping |
| ROUTE-011 | wrap_session | atropos | Parse SKILL.md, verify mapping |
| ROUTE-012 | generate_sails | atropos | Parse SKILL.md, verify mapping |
| ROUTE-013 | delete_sprint | atropos | Parse SKILL.md, verify mapping |

### 10.3 Content Validation Tests

| Test ID | Description | Pass Criteria |
|---------|-------------|---------------|
| CONTENT-001 | Each operation has syntax | Syntax block present |
| CONTENT-002 | Each operation has parameters | Parameters table present |
| CONTENT-003 | Each operation has validation | Validation list present |
| CONTENT-004 | Each operation has success response | JSON example present |
| CONTENT-005 | Each operation has error responses | Error table present |
| CONTENT-006 | CLI commands accurate | Match ari CLI documentation |

### 10.4 Integration Tests

| Test ID | Scenario | Pass Criteria |
|---------|----------|---------------|
| INT-001 | park_session end-to-end | Task(moirai) -> skill load -> ari -> success |
| INT-002 | wrap_session with WHITE sails | Sails generated, session archived |
| INT-003 | wrap_session with BLACK sails blocked | QUALITY_GATE_FAILED returned |
| INT-004 | wrap_session with --emergency | BLACK sails allowed with warning |
| INT-005 | create_sprint with dependency | Dependency validated, sprint created |
| INT-006 | handoff between agents | Agent recorded, history updated |

---

## 11. Handoff Criteria

Ready for Implementation when:

- [ ] SKILL.md routing table covers all 13 operations
- [ ] clotho.md specifies create_sprint and start_sprint
- [ ] lachesis.md specifies all 8 measurement operations
- [ ] atropos.md specifies all 3 termination operations
- [ ] CLI command mappings match ari documentation
- [ ] Response schemas match moirai-shared.md
- [ ] Error codes documented with resolution hints
- [ ] Natural language mappings preserved from current agents
- [ ] Validation rules captured from current agents
- [ ] Anti-patterns documented per skill
- [ ] Test strategy covers all operations
- [ ] Principal Engineer can implement without architectural questions

---

## 12. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Skills too verbose | Medium | Medium | Size targets per skill, progressive disclosure |
| Missing operation detail | Low | High | Test strategy validates completeness |
| CLI mapping drift | Low | Medium | Single source of truth in skill, cross-ref ari docs |
| User attempts direct skill invocation | Low | Low | Skills marked internal: true |
| Moirai context bloat | Low | Medium | Only load relevant skill per operation |

---

## 13. Artifact Attestation

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-fate-skills.md` | Created |
| PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-moirai-consolidation.md` | Read |
| Current clotho.md | `/Users/tomtenuta/Code/roster/user-agents/clotho.md` | Read |
| Current lachesis.md | `/Users/tomtenuta/Code/roster/user-agents/lachesis.md` | Read |
| Current atropos.md | `/Users/tomtenuta/Code/roster/user-agents/atropos.md` | Read |
| Current moirai-shared.md | `/Users/tomtenuta/Code/roster/user-agents/moirai-shared.md` | Read |
| Example skill (smell-detection) | `/Users/tomtenuta/Code/roster/.claude/skills/smell-detection/SKILL.md` | Read |
| Example skill (10x-workflow) | `/Users/tomtenuta/Code/roster/.claude/skills/10x-workflow/SKILL.md` | Read |
| doc-artifacts templates | `/Users/tomtenuta/Code/roster/.claude/skills/doc-artifacts/SKILL.md` | Read |

---

## 14. Related ADRs

| ADR | Status | Topic |
|-----|--------|-------|
| ADR-0009 | Approved | Knossos roster identity |
| ADR-fate-skills-001 | Proposed | Skills vs subagents for Fate logic |
| ADR-fate-skills-002 | Proposed | Skill loading via Read vs Skill tool |

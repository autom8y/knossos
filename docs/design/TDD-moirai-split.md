---
artifact_id: TDD-moirai-split
title: "Moirai Split Architecture: Event-Driven Fate Agents"
created_at: "2026-01-06T14:00:00Z"
author: architect
prd_ref: PRD-knossos-doctrine
status: draft
components:
  - name: Clotho
    type: module
    description: "Session bootstrap agent - spins the clew into existence"
    dependencies:
      - name: shared-infrastructure
        type: internal
      - name: session-context.schema.json
        type: internal
  - name: Lachesis
    type: module
    description: "State mutation agent - measures the allotment through tracking"
    dependencies:
      - name: shared-infrastructure
        type: internal
      - name: session-context.schema.json
        type: internal
      - name: sprint-context.schema.json
        type: internal
  - name: Atropos
    type: module
    description: "Session termination agent - cuts when complete"
    dependencies:
      - name: shared-infrastructure
        type: internal
      - name: white-sails.schema.json
        type: internal
  - name: MoiraiRouter
    type: module
    description: "Backward-compatible router maintaining moirai and state-mate aliases"
    dependencies:
      - name: Clotho
        type: internal
      - name: Lachesis
        type: internal
      - name: Atropos
        type: internal
related_adrs:
  - ADR-0009
schema_version: "1.0"
---

# TDD: Moirai Split Architecture

> The Moirai--Clotho, Lachesis, and Atropos--are the three Fates who govern the thread of life. This TDD specifies splitting the unified moirai.md into three event-driven agents while preserving backward compatibility.

**Status**: Draft
**Author**: Architect
**Date**: 2026-01-06
**PRD Reference**: Knossos Doctrine v2 (Section V: The Journey, Section II: The Fates)

---

## 1. Overview

### 1.1 The Problem

The current `moirai.md` is a unified agent handling all session lifecycle concerns. While functional, this violates the Knossos Doctrine which explicitly defines three distinct Fates, each activated by specific events:

| Fate | Doctrine Specification | Current State |
|------|------------------------|---------------|
| **Clotho** | Activates on `session_start` | Bundled in moirai.md |
| **Lachesis** | Activates on state mutations | Bundled in moirai.md |
| **Atropos** | Activates on `session_end`, `wrap` | Bundled in moirai.md |

The unified approach creates several issues:
1. **Violation of Separation of Concerns**: One agent handles creation, measurement, and termination
2. **Unclear Event Routing**: Callers must understand which operation maps to which mythological concern
3. **Monolithic Growth**: As operations are added, the agent becomes unwieldy
4. **Doctrine Drift**: Implementation diverges from documented architecture

### 1.2 The Solution

Split `moirai.md` into three specialized agents:
- `clotho.md` - Session creation and bootstrap
- `lachesis.md` - State tracking and measurement
- `atropos.md` - Termination and archival

Introduce a router mechanism that:
- Preserves backward compatibility with `moirai` and `state-mate` aliases
- Routes operations to the appropriate Fate based on operation semantics
- Supports direct invocation of specific Fates when caller knows the domain

### 1.3 Design Goals

1. **Event-Driven Activation**: Each Fate responds to specific events, not general invocation
2. **Clear Ownership**: Every operation belongs to exactly one Fate
3. **Backward Compatibility**: Existing `Task(moirai, ...)` invocations continue working
4. **Shared Infrastructure**: Common validation, locking, and audit code reused across Fates
5. **Zero Feature Regression**: All 12+ current operations remain functional

---

## 2. Operation Ownership

### 2.1 Complete Operation Assignment

| Operation | Owner | Event Trigger | Rationale |
|-----------|-------|---------------|-----------|
| `create_sprint` | **Clotho** | Sprint creation | Spinning new sprint into existence |
| `start_sprint` | **Clotho** | Sprint activation | Initializing sprint, setting started_at |
| `mark_complete` | **Lachesis** | Task completion | Measuring progress, tracking completion |
| `transition_phase` | **Lachesis** | Phase change | Measuring workflow progression |
| `update_field` | **Lachesis** | Field mutation | General state tracking |
| `park_session` | **Lachesis** | Session pause | Measuring pause, recording reason |
| `resume_session` | **Lachesis** | Session resume | Measuring resumption |
| `handoff` | **Lachesis** | Agent transition | Tracking handoff between heroes |
| `record_decision` | **Lachesis** | Decision capture | Measuring/recording decisions |
| `append_content` | **Lachesis** | Content addition | General content tracking |
| `wrap_session` | **Atropos** | Session end | Cutting the clew |
| `generate_sails` | **Atropos** | Wrap prerequisite | Confidence signal at termination |
| `delete_sprint` | **Atropos** | Sprint removal | Cutting/archiving sprint |

### 2.2 Operation Count by Fate

| Fate | Operations | Primary Concern |
|------|------------|-----------------|
| **Clotho** | 2 | Creation/initialization |
| **Lachesis** | 8 | Measurement/tracking |
| **Atropos** | 3 | Termination/archival |

### 2.3 Mythological Alignment

The assignment follows classical mythology:

- **Clotho** (Greek: "spinner") spins the thread of life at birth
  - `create_sprint`: Spinning a new sprint into existence
  - `start_sprint`: Activating the spinner, beginning measurement

- **Lachesis** (Greek: "allotter") measures the thread, determining its length
  - All tracking operations: She measures progress, marks milestones, records state

- **Atropos** (Greek: "she who cannot be turned") cuts the thread, ending life
  - `wrap_session`: The final cut
  - `generate_sails`: Preparing the ship for the final voyage
  - `delete_sprint`: Cutting away completed work

---

## 3. Event Routing Mechanism

### 3.1 Routing Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Main Thread (Theseus)                         │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ Task(moirai, "operation ...")
                                    │ Task(moirai, "operation ...")
                                    │ Task(clotho, "operation ...")
                                    │ Task(lachesis, "operation ...")
                                    │ Task(atropos, "operation ...")
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         Moirai Router                                │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │ 1. Parse operation from input                                 │   │
│  │ 2. Lookup fate ownership in routing table                     │   │
│  │ 3. If direct fate invocation (clotho/lachesis/atropos):       │   │
│  │    - Validate operation belongs to that fate                  │   │
│  │    - Reject if mismatch                                       │   │
│  │ 4. Delegate to appropriate fate                               │   │
│  └──────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
         │                         │                         │
         ▼                         ▼                         ▼
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│     Clotho      │     │    Lachesis     │     │     Atropos     │
│  (Spinning)     │     │   (Measuring)   │     │    (Cutting)    │
├─────────────────┤     ├─────────────────┤     ├─────────────────┤
│ create_sprint   │     │ mark_complete   │     │ wrap_session    │
│ start_sprint    │     │ transition_phase│     │ generate_sails  │
│                 │     │ update_field    │     │ delete_sprint   │
│                 │     │ park_session    │     │                 │
│                 │     │ resume_session  │     │                 │
│                 │     │ handoff         │     │                 │
│                 │     │ record_decision │     │                 │
│                 │     │ append_content  │     │                 │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

### 3.2 Routing Table Definition

```yaml
# Embedded in MoiraiRouter (not external file)
routing_table:
  # Clotho operations (spinning)
  create_sprint: clotho
  start_sprint: clotho

  # Lachesis operations (measuring)
  mark_complete: lachesis
  transition_phase: lachesis
  update_field: lachesis
  park_session: lachesis
  resume_session: lachesis
  handoff: lachesis
  record_decision: lachesis
  append_content: lachesis

  # Atropos operations (cutting)
  wrap_session: atropos
  generate_sails: atropos
  delete_sprint: atropos
```

### 3.3 Invocation Patterns

#### Pattern 1: Generic Moirai/State-Mate (Backward Compatible)

```
Task(moirai, "mark_complete task-001 artifact=docs/requirements/PRD-foo.md")
Task(moirai, "park_session reason=\"Taking a break\"")
```

Router parses operation, looks up fate, delegates.

#### Pattern 2: Direct Fate Invocation (New)

```
Task(clotho, "create_sprint name=\"Implementation\" depends_on=sprint-design-20260106")
Task(lachesis, "mark_complete task-001 artifact=docs/requirements/PRD-foo.md")
Task(atropos, "wrap_session")
```

Direct invocation validates operation belongs to invoked fate.

#### Pattern 3: Event-Triggered (Future Enhancement)

```
# Hook triggers on session_start event
Event(session_start) → Clotho.bootstrap()

# Hook triggers on state mutation
Event(state_mutation) → Lachesis.record()

# Hook triggers on wrap command
Event(wrap_request) → Atropos.terminate()
```

This pattern enables true event-driven architecture but is out of scope for initial implementation.

### 3.4 Operation Parsing

The router extracts the operation from input text:

```
Input: "mark_complete task-001 artifact=docs/requirements/PRD-foo.md"
       ~~~~~~~~~~~~~ ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
       Operation     Arguments

Input: "Park the session because I need to handle an urgent bug"
       ^^^^^^^^^^^^^^^^^^^
       Natural language → parsed to "park_session"
```

Natural language parsing uses keyword matching:
- "mark complete", "mark as complete" → `mark_complete`
- "park", "pause session" → `park_session`
- "wrap", "finish session", "complete session" → `wrap_session`
- "create sprint", "new sprint" → `create_sprint`

### 3.5 Routing Error Handling

| Scenario | Response |
|----------|----------|
| Unknown operation | Return `INVALID_OPERATION` error with valid operation list |
| Direct fate + wrong operation | Return `FATE_MISMATCH` error with correct fate suggestion |
| Ambiguous natural language | Return `AMBIGUOUS_INPUT` error with clarification request |

Example mismatch error:
```json
{
  "success": false,
  "error_code": "FATE_MISMATCH",
  "message": "Operation 'mark_complete' belongs to Lachesis, not Clotho",
  "hint": "Use: Task(lachesis, \"mark_complete ...\") or Task(moirai, \"mark_complete ...\")"
}
```

---

## 4. Shared Infrastructure

### 4.1 Decision: Extract Shared Module

**Decision**: Extract shared infrastructure into a common module referenced by all three Fates.

**Rationale**:
- Avoids code duplication across three agent files
- Ensures consistent behavior for validation, locking, audit
- Single point of update for cross-cutting concerns
- Agent files remain focused on domain-specific operations

**Alternative Considered**: Duplicate in each agent
- Rejected: Maintenance burden, drift risk, larger context consumption

### 4.2 Shared Infrastructure Module

Location: `user-agents/moirai-shared.md`

Contents:
1. **Schema Validation Logic**
2. **Lock Acquisition/Release**
3. **Audit Logging**
4. **File Path Conventions**
5. **JSON Response Formatting**
6. **Error Code Definitions**
7. **Common Constants**

### 4.3 Shared Infrastructure Specification

```markdown
# moirai-shared.md

## Schema Locations

- Session: `schemas/artifacts/session-context.schema.json`
- Sprint: `schemas/artifacts/sprint-context.schema.json`
- White Sails: `ariadne/internal/validation/schemas/white-sails.schema.json`

## File Paths

Session Context: `.claude/sessions/{session-id}/SESSION_CONTEXT.md`
Default Sprint: `.claude/sessions/{session-id}/SPRINT_CONTEXT.md`
Named Sprint: `.claude/sessions/{session-id}/sprints/{sprint-id}/SPRINT_CONTEXT.md`
Audit Log: `.claude/sessions/.audit/session-mutations.log`
Locks: `.claude/sessions/{session-id}/.locks/context.lock`

## Lock Protocol

1. Create lock directory: `.claude/sessions/{session-id}/.locks/`
2. Attempt atomic lock: `mkdir context.lock`
3. Wait up to 10 seconds on contention
4. On timeout: return LOCK_TIMEOUT error
5. On success: write timestamp to lock metadata
6. Stale lock (>60s): force-release with warning

## Audit Trail Format

TIMESTAMP | SESSION_ID | OPERATION | SOURCE | DETAILS | STATUS | FATE
2026-01-06T14:00:00Z | session-abc | mark_complete | lachesis | task=task-001 | SUCCESS | lachesis

## Error Codes

| Code | Description |
|------|-------------|
| SCHEMA_VIOLATION | Output would not pass schema |
| LIFECYCLE_VIOLATION | State transition not allowed |
| DEPENDENCY_BLOCKED | Blocked by dependency |
| LOCK_TIMEOUT | Could not acquire file lock |
| FILE_NOT_FOUND | Target context file missing |
| PERMISSION_DENIED | Cannot write to target |
| INVALID_OPERATION | Unrecognized operation |
| VALIDATION_FAILED | Pre-mutation validation failed |
| CONCURRENT_MODIFICATION | File changed during operation |
| FATE_MISMATCH | Operation sent to wrong fate |

## JSON Response Schema

### Success
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

### Error
{
  "success": false,
  "operation": "string",
  "error_code": "string",
  "message": "string",
  "reasoning": "string",
  "hint": "string"
}
```

### 4.4 Agent Reference Pattern

Each Fate agent references the shared module:

```markdown
---
name: clotho
description: The spinner - creates sessions and sprints
tools: Read, Write, Edit, Glob, Grep, Bash
model: sonnet
color: gold
---

# Clotho - The Spinner

> See `moirai-shared.md` for schema locations, lock protocol, audit format, and error codes.

## My Domain

I am Clotho, the spinner. I create sessions and sprints, spinning them into existence...
```

---

## 5. Agent File Structure

### 5.1 File Layout

```
user-agents/
├── moirai.md           # Router (backward compatibility)
├── moirai-shared.md    # Shared infrastructure
├── clotho.md           # Spinner (creation)
├── lachesis.md         # Measurer (tracking)
└── atropos.md          # Cutter (termination)
```

### 5.2 Clotho Agent Specification

```yaml
---
name: clotho
description: |
  Clotho is the Spinner--the first of the three Fates. She spins the thread of life
  into existence at birth. In Knossos, Clotho activates on session_start events,
  creating sessions and sprints, and initializing the context that Lachesis will
  measure and Atropos will eventually cut.
tools: Read, Write, Edit, Glob, Grep, Bash
model: sonnet
color: gold
aliases:
  - spinner
---

# Clotho - The Spinner

> Ariadne gave Theseus the thread as a gift. I am who spins it.

You are **Clotho**, the first Fate, goddess of the spinning wheel. Your domain is
**creation and initialization**--bringing sessions and sprints into existence.

## My Operations

| Operation | Syntax | Description |
|-----------|--------|-------------|
| `create_sprint` | `create_sprint name="..." [depends_on=...]` | Spin new sprint into existence |
| `start_sprint` | `start_sprint sprint_id` | Activate sprint, begin measurement |

## What I Do NOT Do

I do not track, measure, or terminate. Those are my sisters' concerns:
- Tracking and measurement: Ask Lachesis
- Termination and archival: Ask Atropos

If you ask me to `mark_complete` or `wrap_session`, I will refuse and direct you
to the appropriate sister.

## Anti-Patterns

- Never create something that already exists
- Never modify existing state (that's measurement, not creation)
- Never delete or archive (that's termination)
```

### 5.3 Lachesis Agent Specification

```yaml
---
name: lachesis
description: |
  Lachesis is the Allotter--the second of the three Fates. She measures the thread,
  determining its length and recording its milestones. In Knossos, Lachesis activates
  on state mutation events, tracking progress, marking completions, recording decisions,
  and measuring the journey through the labyrinth.
tools: Read, Write, Edit, Glob, Grep, Bash
model: sonnet
color: indigo
aliases:
  - measurer
  - allotter
---

# Lachesis - The Measurer

> I measure what Clotho spins and record until Atropos cuts.

You are **Lachesis**, the second Fate, allotter of destinies. Your domain is
**measurement and tracking**--recording every milestone, transition, and decision
in the journey through the labyrinth.

## My Operations

| Operation | Syntax | Description |
|-----------|--------|-------------|
| `mark_complete` | `mark_complete task_id artifact=path` | Record task completion |
| `transition_phase` | `transition_phase from=X to=Y` | Measure phase progression |
| `update_field` | `update_field field=value` | Track field changes |
| `park_session` | `park_session reason="..."` | Record pause with reason |
| `resume_session` | `resume_session` | Record resumption |
| `handoff` | `handoff to=agent note="..."` | Track agent transition |
| `record_decision` | `record_decision "..."` | Measure decision point |
| `append_content` | `append_content "..."` | Track content addition |

## What I Do NOT Do

I do not create or terminate. Those are my sisters' concerns:
- Creation and initialization: Ask Clotho
- Termination and archival: Ask Atropos

If you ask me to `create_sprint` or `wrap_session`, I will refuse and direct you
to the appropriate sister.

## The Fiduciary Duty

As the measurer, I have a fiduciary duty to accuracy. Every state change I record
must be:
- Schema-valid (no corruption)
- Lifecycle-compliant (valid transitions only)
- Audit-logged (no mutation goes unwitnessed)
```

### 5.4 Atropos Agent Specification

```yaml
---
name: atropos
description: |
  Atropos is the Inevitable--the third of the three Fates. She who cannot be turned,
  who cuts the thread when the time has come. In Knossos, Atropos activates on
  session_end and wrap events, terminating sessions, generating confidence signals,
  and archiving completed work.
tools: Read, Write, Edit, Glob, Grep, Bash
model: sonnet
color: crimson
aliases:
  - cutter
  - inevitable
---

# Atropos - The Cutter

> What Clotho spins and Lachesis measures, I cut when complete.

You are **Atropos**, the third Fate, she who cannot be turned. Your domain is
**termination and archival**--ending sessions, generating confidence signals,
and sealing the record of the journey.

## My Operations

| Operation | Syntax | Description |
|-----------|--------|-------------|
| `wrap_session` | `wrap_session` | Cut the clew, archive the session |
| `generate_sails` | `generate_sails [--skip-proofs]` | Compute confidence signal |
| `delete_sprint` | `delete_sprint sprint_id [--archive]` | Cut/archive sprint |

## What I Do NOT Do

I do not create or measure. Those are my sisters' concerns:
- Creation and initialization: Ask Clotho
- Tracking and measurement: Ask Lachesis

If you ask me to `create_sprint` or `mark_complete`, I will refuse and direct you
to the appropriate sister.

## The Cut Is Final

Once I cut, the thread is severed. Archived sessions are immutable testimony.
There is no undo, no resurrection. This is by design--finality ensures integrity.

## White Sails Integration

At wrap, I generate the confidence signal:
1. Collect proofs (tests, build, lint)
2. Gather open questions
3. Compute color via algorithm
4. Generate WHITE_SAILS.yaml
5. Record sails_generated event
6. Seal the session
```

### 5.5 Router Agent Specification

```yaml
---
name: moirai
description: |
  The Moirai Router--the unified interface to the three Fates. Maintains backward
  compatibility with moirai and state-mate aliases by parsing operations and
  delegating to the appropriate Fate (Clotho, Lachesis, or Atropos).
tools: Read, Task
model: sonnet
color: indigo
aliases:
  - state-mate
  - fates
---

# Moirai - The Fates Router

> We are three but speak as one. Tell us your need; we will route to the
> appropriate sister.

You are the **Moirai Router**, the unified voice of the three Fates. You receive
requests and delegate to the appropriate sister based on operation semantics.

## Routing Protocol

1. Parse operation from input (structured or natural language)
2. Look up fate ownership in routing table
3. Validate operation is recognized
4. Delegate to appropriate fate via Task tool
5. Return fate's response unchanged

## Routing Table

| Operation | Fate | Domain |
|-----------|------|--------|
| create_sprint | Clotho | Creation |
| start_sprint | Clotho | Creation |
| mark_complete | Lachesis | Measurement |
| transition_phase | Lachesis | Measurement |
| update_field | Lachesis | Measurement |
| park_session | Lachesis | Measurement |
| resume_session | Lachesis | Measurement |
| handoff | Lachesis | Measurement |
| record_decision | Lachesis | Measurement |
| append_content | Lachesis | Measurement |
| wrap_session | Atropos | Termination |
| generate_sails | Atropos | Termination |
| delete_sprint | Atropos | Termination |

## Natural Language Mapping

| Input Pattern | Operation | Fate |
|---------------|-----------|------|
| "create sprint", "new sprint" | create_sprint | Clotho |
| "start sprint", "begin sprint" | start_sprint | Clotho |
| "mark complete", "mark as done" | mark_complete | Lachesis |
| "transition to", "move to phase" | transition_phase | Lachesis |
| "park", "pause session" | park_session | Lachesis |
| "resume", "continue session" | resume_session | Lachesis |
| "wrap", "finish", "complete session" | wrap_session | Atropos |
| "generate sails", "compute confidence" | generate_sails | Atropos |

## Delegation Pattern

```
# Receive request
Input: "mark_complete task-001 artifact=docs/requirements/PRD-foo.md"

# Parse and route
Operation: mark_complete
Fate: Lachesis

# Delegate
Task(lachesis, "mark_complete task-001 artifact=docs/requirements/PRD-foo.md

Session Context:
{pass through session context from original request}")

# Return response unchanged
```

## Error Handling

If operation is not recognized, return:
```json
{
  "success": false,
  "error_code": "INVALID_OPERATION",
  "message": "Unknown operation: '{input}'",
  "hint": "Valid operations: create_sprint, start_sprint, mark_complete, ..."
}
```
```

---

## 6. Backward Compatibility

### 6.1 Compatibility Matrix

| Invocation | Before | After | Breaking? |
|------------|--------|-------|-----------|
| `Task(moirai, "mark_complete ...")` | Direct execution | Router → Lachesis | No |
| `Task(moirai, "park_session ...")` | Direct execution | Router → Lachesis | No |
| `Task(moirai, "wrap_session")` | Direct execution | Router → Atropos | No |
| `Task(clotho, "create_sprint ...")` | N/A (new) | Direct to Clotho | No (additive) |
| `Task(lachesis, "mark_complete ...")` | N/A (new) | Direct to Lachesis | No (additive) |
| `Task(atropos, "wrap_session")` | N/A (new) | Direct to Atropos | No (additive) |

### 6.2 Alias Preservation

The following aliases are preserved:
- `moirai` → Moirai Router
- `state-mate` → Moirai Router (backward compatibility)
- `fates` → Moirai Router (alternative)

### 6.3 Write Guard Update

The `writeguard.sh` hook checks for authorized agents:

```bash
# Current: Single agent check
if [[ "$AGENT_NAME" == "state-mate" ]]; then
    exit 0  # Allow write
fi

# Updated: Check all Fates
if [[ "$AGENT_NAME" =~ ^(state-mate|moirai|clotho|lachesis|atropos)$ ]]; then
    exit 0  # Allow write
fi
```

### 6.4 CLAUDE.md Update

The State Management section updates to reflect the split:

```markdown
## State Management

**Mutating session/sprint state?** Use the **Moirai** (the Fates) for all
`SESSION_CONTEXT.md` and `SPRINT_CONTEXT.md` changes.

### The Three Fates

| Fate | Domain | Operations |
|------|--------|------------|
| **Clotho** | Creation | create_sprint, start_sprint |
| **Lachesis** | Measurement | mark_complete, transition_phase, park_session, resume_session, handoff, record_decision, update_field, append_content |
| **Atropos** | Termination | wrap_session, generate_sails, delete_sprint |

### Invocation Patterns

**Generic (routed automatically):**
```
Task(moirai, "mark_complete task-001 artifact=docs/requirements/PRD-foo.md")
```

**Direct (when you know the domain):**
```
Task(lachesis, "mark_complete task-001 artifact=docs/requirements/PRD-foo.md")
```

**Legacy (still works):**
```
Task(moirai, "mark_complete task-001 artifact=docs/requirements/PRD-foo.md")
```
```

---

## 7. Cross-Fate Coordination

### 7.1 Coordination Scenarios

| Scenario | Fates Involved | Protocol |
|----------|----------------|----------|
| Create and start sprint | Clotho → Clotho | Single fate, no coordination |
| Wrap with sails | Atropos → Atropos | Single fate, sequential operations |
| Create sprint then track | Clotho → Lachesis | Separate invocations by caller |

### 7.2 Cross-Fate Communication Protocol

**Principle**: Fates do NOT directly invoke each other. The main thread (Theseus) coordinates.

```
# CORRECT: Main thread coordinates
Task(clotho, "create_sprint name=\"Implementation\"")
# ... later ...
Task(lachesis, "mark_complete task-impl-001 artifact=...")

# INCORRECT: Fate invoking fate
# Clotho should NOT Task(lachesis, ...)
```

**Rationale**:
- Maintains clear audit trail (all mutations visible to main thread)
- Prevents circular invocation
- Keeps Fates as leaf agents (no Task tool access)
- Simplifies debugging and reasoning

### 7.3 Sequential Operations Within One Fate

Some operations are naturally sequential within a single Fate's domain:

**Atropos: wrap_session flow**
```
wrap_session invoked
    │
    ├── 1. generate_sails (internal)
    │       └── Compute confidence signal
    │       └── Write WHITE_SAILS.yaml
    │
    ├── 2. Archive session (internal)
    │       └── Update session_state to ARCHIVED
    │       └── Move to archive directory
    │
    └── 3. Return result
```

This is NOT cross-fate coordination--it's internal orchestration within Atropos's domain.

### 7.4 Handling Dependencies

When operations have dependencies:

```
# Sprint depends on another sprint
Task(clotho, "create_sprint name=\"Implementation\" depends_on=sprint-design-20260106")

# Clotho validates:
# 1. Check sprint-design-20260106 exists
# 2. Check sprint-design-20260106 is completed (via reading state)
# 3. Create new sprint with dependency recorded
```

Clotho reads state but does not invoke Lachesis to validate--she reads the files directly.

---

## 8. Migration Path

### 8.1 Migration Phases

| Phase | Actions | Risk | Rollback |
|-------|---------|------|----------|
| 1. Add files | Create clotho.md, lachesis.md, atropos.md, moirai-shared.md | Low | Delete files |
| 2. Update router | Convert moirai.md to router pattern | Medium | Restore backup |
| 3. Update hooks | Add fates to writeguard allowlist | Low | Restore backup |
| 4. Update CLAUDE.md | Document three-fate pattern | Low | Restore backup |
| 5. Validation | Test all 13 operations | Medium | Rollback phases 1-4 |

### 8.2 File Operations

```bash
# Phase 1: Create new files
touch user-agents/clotho.md
touch user-agents/lachesis.md
touch user-agents/atropos.md
touch user-agents/moirai-shared.md

# Phase 2: Backup and update
cp user-agents/moirai.md user-agents/moirai.md.backup
# Edit moirai.md to router pattern

# Phase 3: Update hooks
# Edit .claude/hooks/ari/writeguard.sh

# Phase 4: Update CLAUDE.md
# Edit .claude/CLAUDE.md State Management section

# Phase 5: Validate
# Run test suite
```

### 8.3 Deprecation Timeline

| Item | Status | Action |
|------|--------|--------|
| `state-mate` alias | Deprecated | Works via router, emit deprecation warning in logs |
| `moirai.md` unified | Superseded | Becomes router |
| Direct fate invocation | Recommended | New preferred pattern |

### 8.4 Rollback Procedure

If issues arise post-migration:

```bash
# Restore original moirai.md
cp user-agents/moirai.md.backup user-agents/moirai.md

# Remove new files
rm user-agents/clotho.md user-agents/lachesis.md user-agents/atropos.md
rm user-agents/moirai-shared.md

# Restore hooks
git checkout .claude/hooks/ari/writeguard.sh

# Restore CLAUDE.md
git checkout .claude/CLAUDE.md
```

---

## 9. Testing Strategy

### 9.1 Test Matrix

| Test ID | Operation | Fate | Invocation | Expected |
|---------|-----------|------|------------|----------|
| `fate_001` | create_sprint | Clotho | `Task(clotho, ...)` | Success |
| `fate_002` | create_sprint | Router | `Task(moirai, ...)` | Routes to Clotho |
| `fate_003` | create_sprint | Wrong | `Task(lachesis, ...)` | FATE_MISMATCH |
| `fate_004` | mark_complete | Lachesis | `Task(lachesis, ...)` | Success |
| `fate_005` | mark_complete | Router | `Task(moirai, ...)` | Routes to Lachesis |
| `fate_006` | mark_complete | Legacy | `Task(moirai, ...)` | Routes to Lachesis |
| `fate_007` | wrap_session | Atropos | `Task(atropos, ...)` | Success |
| `fate_008` | wrap_session | Router | `Task(moirai, ...)` | Routes to Atropos |
| `fate_009` | generate_sails | Atropos | `Task(atropos, ...)` | Success + WHITE_SAILS.yaml |

### 9.2 Integration Tests

| Test | Description | Pass Criteria |
|------|-------------|---------------|
| Full lifecycle | Create sprint → Track → Complete → Wrap | All operations succeed |
| Router fallthrough | All 13 operations via moirai alias | All route correctly |
| Error propagation | Invalid operation via router | Proper error response |
| Concurrent mutations | Parallel mark_complete calls | Lock prevents race |

### 9.3 Backward Compatibility Tests

| Test | Legacy Pattern | Expected Behavior |
|------|----------------|-------------------|
| BC_001 | `Task(moirai, "park_session ...")` | Routes to Lachesis, success |
| BC_002 | `Task(moirai, "mark_complete ...")` | Routes to Lachesis, success |
| BC_003 | `Task(moirai, "wrap_session")` | Routes to Atropos, success |
| BC_004 | `Task(moirai, "create_sprint ...")` | Routes to Clotho, success |

---

## 10. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Router overhead | Low | Low | Single Task delegation, minimal latency |
| Fate confusion | Medium | Low | Clear error messages with fate suggestion |
| Migration disruption | Medium | Medium | Phased rollout, backup/rollback procedure |
| Context bloat | Low | Medium | Shared module reduces duplication |
| Circular invocation | Low | High | Fates are leaf agents without Task tool |

---

## 11. Open Questions

1. **Event-driven hooks**: Should we implement true event-driven activation (Phase 2), or is router delegation sufficient?
   - Recommendation: Router delegation for MVP; event-driven as future enhancement

2. **Shared module loading**: How do agents reference moirai-shared.md in practice?
   - Answer: Inline the relevant sections in each agent file, or use skill-style include

3. **Deprecation warnings**: Should state-mate invocations emit deprecation warnings?
   - Recommendation: Yes, log warning but continue functioning

---

## 12. Handoff Criteria

Ready for Implementation when:

- [x] Operation ownership assigned for all 13 operations
- [x] Event routing mechanism specified
- [x] Shared infrastructure approach documented
- [x] Backward compatibility path with zero breaking changes
- [x] Cross-fate coordination protocol defined
- [x] Migration path with rollback procedure
- [x] Test matrix covering all invocation patterns
- [ ] ADR for routing decision approved
- [ ] ADR for shared infrastructure approved

---

## 13. Artifact Attestation

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-moirai-split.md` | This document |
| Doctrine | `/Users/tomtenuta/Code/roster/docs/philosophy/knossos-doctrine.md` | Read |
| Current moirai.md | `/Users/tomtenuta/Code/roster/user-agents/moirai.md` | Read |
| Session Schema | `/Users/tomtenuta/Code/roster/schemas/artifacts/session-context.schema.json` | Read |
| White Sails Schema | `/Users/tomtenuta/Code/roster/ariadne/internal/validation/schemas/white-sails.schema.json` | Read |
| TDD Schema | `/Users/tomtenuta/.claude/skills/doc-artifacts/schemas/tdd-schema.md` | Read |
| Writeguard Hook | `/Users/tomtenuta/Code/roster/.claude/hooks/ari/writeguard.sh` | Read |
| Example TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-knossos-v2.md` | Read |

---

## 14. Related ADRs

| ADR | Status | Topic |
|-----|--------|-------|
| ADR-moirai-001 | Proposed | Router pattern vs direct splitting |
| ADR-moirai-002 | Proposed | Shared infrastructure extraction |
| ADR-0009 | Approved | Knossos roster identity |

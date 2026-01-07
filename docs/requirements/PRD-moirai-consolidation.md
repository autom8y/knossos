---
artifact_id: PRD-moirai-consolidation
title: "Moirai Consolidation & Session State Overhaul"
created_at: "2026-01-07T22:00:00Z"
author: requirements-analyst
status: draft
complexity: MODULE
impact: high
impact_categories: [architecture, session_management]
success_criteria:
  - id: SC-001
    description: "Single Moirai agent handles all session state operations (create_sprint, mark_complete, wrap_session, etc.)"
    testable: true
    priority: must-have
  - id: SC-002
    description: "Clotho, Lachesis, Atropos exist as Moirai-internal skills, not standalone agents"
    testable: true
    priority: must-have
  - id: SC-003
    description: "Write guard hook blocks direct *_CONTEXT.md writes and suggests Task(moirai, ...)"
    testable: true
    priority: must-have
  - id: SC-004
    description: "Slash commands (/park, /wrap, /handoff) internally invoke Moirai without user awareness"
    testable: true
    priority: must-have
  - id: SC-005
    description: "SPRINT_CONTEXT.md protected with same guard as SESSION_CONTEXT.md"
    testable: true
    priority: must-have
  - id: SC-006
    description: "No routing errors from incorrect Fate selection (measured by audit log)"
    testable: true
    priority: must-have
  - id: SC-007
    description: "Moirai delegates to ari CLI for authoritative state changes"
    testable: true
    priority: must-have
  - id: SC-008
    description: "Audit trail preserved for all mutations with reasoning field"
    testable: true
    priority: must-have
schema_version: "1.0"
---

# PRD: Moirai Consolidation & Session State Overhaul

## Overview

The current 4-agent Moirai architecture (Moirai router + Clotho/Lachesis/Atropos Fates) introduces unnecessary cognitive overhead and routing latency. Users struggle to remember which Fate handles which operation, and the router adds an extra subprocess hop before reaching the actual Fate. Additionally, the write guard hook exists but does not effectively prevent direct writes or provide actionable guidance, and main thread agents frequently bypass Moirai entirely.

This initiative consolidates the Fates into a single unified Moirai agent that loads domain-specific logic on-demand via skills, while fixing the write guard and updating slash commands to route through Moirai transparently.

## Background

### Current Architecture (Being Replaced)

```
User/Agent --> Task(moirai) --> Moirai Router --> Task(clotho|lachesis|atropos) --> File Mutation
                                    |
                              4 separate agents
                              - user-agents/moirai.md (router)
                              - user-agents/clotho.md (creation)
                              - user-agents/lachesis.md (measurement)
                              - user-agents/atropos.md (termination)
                              - user-agents/moirai-shared.md (shared definitions)
```

**Problems Identified:**

1. **Cognitive Overhead**: Users must learn the Fate taxonomy (Clotho spins, Lachesis measures, Atropos cuts) to understand routing, but then invoke Moirai generically anyway.

2. **Routing Latency**: Extra subprocess hop through Moirai router before reaching actual Fate adds ~50-100ms per operation.

3. **Broken Write Guard**: The hook at `.claude/hooks/session-guards/session-write-guard.sh` blocks direct writes but:
   - Suggests `Task(moirai, ...)` which then routes to a Fate
   - Does not intercept all bypass paths (main thread direct writes)
   - Audit log shows bypasses via `STATE_MATE_BYPASS` mechanism

4. **No Discovery**: No Moirai skill exists for users to learn the invocation pattern progressively.

5. **Main Thread Bypassing**: Main agent (Theseus) performs session updates directly instead of using Moirai subagents, as evidenced by `.claude/audit/moirai-bypass.jsonl`.

### Target Architecture

```
User/Agent --> /park (slash cmd) --> Task(moirai) --> Moirai (unified) --> ari CLI
                                           |
                                   Skill(clotho|lachesis|atropos)
                                   (loaded on-demand, internal only)
```

**Key Changes:**

- Single Moirai agent replaces router + 3 Fates
- Fate logic becomes Moirai-internal skills (progressive disclosure)
- Users invoke slash commands, which route to Moirai internally
- Moirai wraps `ari` CLI commands for authoritative state changes
- Write guard provides clear error with exact invocation pattern

## Key Decisions (Pre-Approved)

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Agent count | 1 unified Moirai | Reduce cognitive/routing overhead |
| Fate pattern | Skills (progressive disclosure) | Moirai loads domain logic on-demand |
| Skill visibility | Moirai-internal only | Users invoke slash commands, not Fates |
| User journey | Slash commands primary | /park, /wrap, /handoff route to Moirai |
| Hook behavior | Block + suggest Task(moirai) | Hooks cannot invoke agents directly |
| Sprint handling | Child of session | Same protection, nested context |
| CLI relation | CLI is authoritative | Moirai wraps ari commands |
| Compatibility | Hard break | Remove old Fate agent patterns |

## User Stories

### US-001: Single Invocation Pattern

**As a** developer using session management,
**I want** a single Moirai agent that handles all session operations,
**so that** I don't have to remember which Fate (Clotho, Lachesis, or Atropos) handles which operation.

**Acceptance Criteria:**
- [ ] `Task(moirai, "park session")` works without routing to lachesis
- [ ] `Task(moirai, "create sprint")` works without routing to clotho
- [ ] `Task(moirai, "wrap session")` works without routing to atropos
- [ ] Error messages reference Moirai, not individual Fates

### US-002: Transparent Slash Command Routing

**As a** developer,
**I want** `/park`, `/wrap`, and `/handoff` commands to work seamlessly,
**so that** I never need to know about Moirai internals to manage sessions.

**Acceptance Criteria:**
- [ ] `/park "reason"` invokes Moirai internally and returns park confirmation
- [ ] `/wrap` invokes Moirai internally and generates White Sails
- [ ] `/handoff agent` invokes Moirai internally and records transition
- [ ] User sees command output, not Moirai invocation details

### US-003: Write Guard Protection

**As a** session owner,
**I want** direct writes to `*_CONTEXT.md` files to be blocked with clear guidance,
**so that** I understand how to properly mutate session state.

**Acceptance Criteria:**
- [ ] Write/Edit to `SESSION_CONTEXT.md` blocked with decision: block
- [ ] Write/Edit to `SPRINT_CONTEXT.md` blocked with decision: block
- [ ] Error message includes exact `Task(moirai, "...")` invocation
- [ ] Error distinguishes orchestrated vs non-orchestrated workflows

### US-004: Progressive Disclosure of Fate Logic

**As an** advanced user,
**I want** Moirai to load Fate-specific logic only when needed,
**so that** I can understand the conceptual separation without cognitive overhead.

**Acceptance Criteria:**
- [ ] Clotho skill loaded only for creation operations
- [ ] Lachesis skill loaded only for measurement operations
- [ ] Atropos skill loaded only for termination operations
- [ ] Skills are internal to Moirai, not user-invokable

### US-005: CLI as Authority

**As a** system maintainer,
**I want** Moirai to delegate to the `ari` CLI for actual state changes,
**so that** there is a single authoritative source for session mutations.

**Acceptance Criteria:**
- [ ] `ari session park` executed for park_session operation
- [ ] `ari session wrap` executed for wrap_session operation
- [ ] `ari session transition` executed for transition_phase operation
- [ ] Moirai parses CLI output and returns structured JSON

## Functional Requirements

### Must Have

#### FR-001: Unified Moirai Agent

Create a single `.claude/agents/moirai.md` agent that:
- Handles ALL session state operations directly
- Parses natural language and structured commands
- Returns structured JSON responses
- Logs all mutations to audit trail

**Operations Handled:**
| Operation | Domain | Description |
|-----------|--------|-------------|
| `create_session` | Creation | Create new session (via ari) |
| `create_sprint` | Creation | Spin new sprint into existence |
| `start_sprint` | Creation | Activate pending sprint |
| `mark_complete` | Measurement | Record task completion |
| `transition_phase` | Measurement | Measure phase progression |
| `update_field` | Measurement | Track field changes |
| `park_session` | Measurement | Record pause with reason |
| `resume_session` | Measurement | Record resumption |
| `handoff` | Measurement | Track agent transition |
| `record_decision` | Measurement | Measure decision point |
| `append_content` | Measurement | Track content addition |
| `wrap_session` | Termination | Archive session with sails |
| `generate_sails` | Termination | Compute confidence signal |
| `delete_sprint` | Termination | Cut/archive sprint |

#### FR-002: Clotho Skill (Creation)

Create `.claude/skills/moirai/clotho.md` skill containing:
- `create_sprint` logic and validation
- `start_sprint` logic and validation
- Schema validation for sprint creation
- Dependency checking for sprint ordering

**Moirai loads this skill when operation is:** `create_sprint`, `start_sprint`

#### FR-003: Lachesis Skill (Measurement)

Create `.claude/skills/moirai/lachesis.md` skill containing:
- `mark_complete` logic with artifact validation
- `transition_phase` logic with lifecycle enforcement
- `update_field` logic with schema validation
- `park_session`, `resume_session` logic
- `handoff` logic with agent validation
- `record_decision`, `append_content` logic

**Moirai loads this skill when operation is:** `mark_complete`, `transition_phase`, `update_field`, `park_session`, `resume_session`, `handoff`, `record_decision`, `append_content`

#### FR-004: Atropos Skill (Termination)

Create `.claude/skills/moirai/atropos.md` skill containing:
- `wrap_session` logic with quality gate enforcement
- `generate_sails` logic with proof collection
- `delete_sprint` logic with archive option
- White Sails color computation algorithm

**Moirai loads this skill when operation is:** `wrap_session`, `generate_sails`, `delete_sprint`

#### FR-005: Write Guard Hook Fix

Update `.claude/hooks/session-guards/session-write-guard.sh` and `.claude/hooks/ari/writeguard.sh` to:
- Block Write/Edit to any file matching `*_CONTEXT.md`
- Return JSON decision: `{"decision": "block", "reason": "...", "suggestion": "Task(moirai, '...')"}`
- Differentiate guidance for orchestrated vs non-orchestrated workflows
- Log bypass attempts to `.claude/audit/moirai-bypass.jsonl`

**Error Message (non-orchestrated):**
```
## State Mutation Blocked

Direct writes to `*_CONTEXT.md` files are not allowed.

**Use Moirai for all session/sprint mutations:**

Task(moirai, "<your mutation request>")

Example: Task(moirai, "park session reason='waiting for review'")
```

**Error Message (orchestrated):**
```
## State Mutation Blocked

State mutations are handled automatically by hooks during orchestrated workflows.

**If explicit mutation needed, use slash commands:**
- /park - Pause current session
- /wrap - Complete and archive session
- /handoff - Transfer to another agent
```

#### FR-006: Slash Command Updates

Update slash command skills to internally invoke Moirai:

**`user-commands/session/park.md`:**
- Replace `session-manager.sh mutate park` with `Task(moirai, "park_session reason='...'")`
- Capture Moirai response and display summary to user

**`user-commands/session/wrap.md`:**
- Replace `session-manager.sh mutate wrap` with `Task(moirai, "wrap_session")`
- Capture sails color and display completion summary

**`user-commands/session/handoff.md`:**
- Replace `session-manager.sh mutate handoff` with `Task(moirai, "handoff to=agent note='...'")`
- Capture handoff confirmation and invoke target agent

#### FR-007: Sprint Context Protection

Extend write guard to protect `SPRINT_CONTEXT.md`:
- Pattern: `**/SPRINT_CONTEXT.md`
- Same blocking behavior as SESSION_CONTEXT.md
- Same error messages and suggestions

#### FR-008: CLI Delegation

Moirai delegates to `ari` CLI for authoritative mutations:

| Moirai Operation | CLI Command |
|------------------|-------------|
| `create_session` | `ari session create` |
| `park_session` | `ari session park` |
| `resume_session` | `ari session resume` |
| `wrap_session` | `ari session wrap` |
| `transition_phase` | `ari session transition` |
| `generate_sails` | Computed during `ari session wrap` |

### Should Have

#### FR-S01: Moirai Discovery Skill

Create `.claude/skills/moirai/moirai-ref.md` user-facing skill that:
- Documents invocation pattern
- Lists available operations
- Provides examples for common tasks
- Triggers: moirai, session state, state management

#### FR-S02: Dry-Run Support

Moirai supports `--dry-run` flag for all operations:
- Returns diff without applying changes
- Useful for validation before mutation

#### FR-S03: Emergency Override

Moirai supports `--emergency` flag:
- Bypasses non-critical validations
- Logs emergency use to audit trail
- Required for wrapping with BLACK sails

### Could Have

#### FR-C01: Natural Language Parsing Improvements

Enhanced natural language understanding:
- "I'm done for today" -> `park_session`
- "Ship it" -> `wrap_session`
- "Hand this to the architect" -> `handoff to=architect`

#### FR-C02: Batch Operations

Support for multiple operations in single invocation:
- `Task(moirai, "mark_complete task-001; transition_phase to=implementation")`

## Non-Functional Requirements

#### NFR-001: Latency Parity

Unified Moirai must not increase latency compared to current router pattern:
- Single agent invocation: <200ms typical
- CLI delegation: <100ms additional
- Total operation: <500ms end-to-end

#### NFR-002: Progressive Loading

Skills loaded lazily to minimize context consumption:
- Moirai base prompt: <50 lines
- Each skill: <100 lines
- Only relevant skill loaded per operation

#### NFR-003: Clear Error Messages

All error conditions produce actionable guidance:
- Include exact invocation syntax
- Reference documentation location
- Distinguish user error from system error

#### NFR-004: Audit Completeness

All mutations logged with full context:
- Timestamp
- Session ID
- Operation name
- State before/after
- Reasoning
- Fate domain (for traceability)

## Edge Cases

| Case | Expected Behavior |
|------|------------------|
| Unknown operation | Return INVALID_OPERATION with valid operation list |
| Ambiguous natural language | Return AMBIGUOUS_INPUT with clarification request |
| Session not found | Return FILE_NOT_FOUND with creation suggestion |
| Concurrent mutation | Return CONCURRENT_MODIFICATION with retry guidance |
| Wrap with BLACK sails | Block unless --emergency flag provided |
| Wrap while PARKED | Block unless --override=reason provided |
| Sprint depends on incomplete sprint | Return DEPENDENCY_BLOCKED with completion suggestion |
| Direct write with MOIRAI_BYPASS=true | Allow (reserved for CLI operations) |

## Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| SM-001: Routing Errors | 0 | Count of FATE_MISMATCH errors in audit log |
| SM-002: Invocation Pattern | 1 | Users learn `Task(moirai, ...)` only |
| SM-003: Write Guard Catches | 100% | All direct writes blocked (test coverage) |
| SM-004: Slash Command Transparency | 100% | Users unaware of Moirai internals |
| SM-005: Latency | <500ms | End-to-end operation timing |
| SM-006: Audit Coverage | 100% | All mutations have reasoning field |

## Migration Path

### Files to Remove

| File | Reason |
|------|--------|
| `user-agents/moirai.md` | Replaced by unified agent |
| `user-agents/clotho.md` | Replaced by skill |
| `user-agents/lachesis.md` | Replaced by skill |
| `user-agents/atropos.md` | Replaced by skill |
| `user-agents/moirai-shared.md` | Absorbed into skills |
| `user-agents/moirai.md.backup` | Obsolete backup |

### Files to Add

| File | Purpose |
|------|---------|
| `.claude/agents/moirai.md` | Unified Moirai agent |
| `.claude/skills/moirai/clotho.md` | Creation operations skill |
| `.claude/skills/moirai/lachesis.md` | Measurement operations skill |
| `.claude/skills/moirai/atropos.md` | Termination operations skill |
| `.claude/skills/moirai/moirai-ref.md` | User discovery skill |

### Files to Update

| File | Changes |
|------|---------|
| `.claude/hooks/session-guards/session-write-guard.sh` | Improved error messages, sprint protection |
| `.claude/hooks/ari/writeguard.sh` | Sprint pattern matching |
| `user-commands/session/park.md` | Route through Moirai |
| `user-commands/session/wrap.md` | Route through Moirai |
| `user-commands/session/handoff.md` | Route through Moirai |
| `.claude/CLAUDE.md` | Update Moirai documentation section |

### Breaking Changes

This is a **hard break** from the current architecture:
- `Task(clotho, ...)` will fail (agent removed)
- `Task(lachesis, ...)` will fail (agent removed)
- `Task(atropos, ...)` will fail (agent removed)
- Old Moirai router pattern still works but routes internally

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Existing workflows break | High | High | Hard break accepted; update all references simultaneously |
| Skills pattern unfamiliar | Medium | Medium | Document progressive disclosure; skills are internal |
| Hook can't auto-invoke Moirai | High | Medium | Clear error message with exact invocation syntax |
| Increased context consumption | Low | Medium | Lazy skill loading; base prompt compact |
| CLI command failures | Low | High | Moirai validates before delegation; structured error handling |

## Out of Scope

- CLI changes (ari is stable, Moirai wraps it)
- Session schema changes
- New session operations
- Worktree session handling
- Multi-session coordination
- Session templates or presets
- Graphical session management UI
- Remote session synchronization

## Dependencies

| Dependency | Type | Status |
|------------|------|--------|
| `ari` CLI | Internal | Stable |
| Session schemas | Internal | Stable |
| Write guard hook infrastructure | Internal | Exists |
| Skill tool support | Claude Code | Available |
| Task tool support | Claude Code | Available |

## Open Questions

*None remaining - all architectural decisions pre-approved by stakeholder.*

---

## File Verification

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-moirai-consolidation.md` | Created |

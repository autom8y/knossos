# Session Lifecycle Workflow

This document describes session initialization, management, and termination workflows for orchestrated teams.

## Quick Start (Orchestrated Teams)

The auto-orchestration enhancement reduces session initialization from 3-5 manual steps to 2 steps:

1. Type: `/start "Your Initiative Name"`
2. Copy the Task invocation from hook output and execute

The hook automatically:
- Creates SESSION_CONTEXT.md via `session-manager.sh create`
- Populates session ID, path, initiative, complexity, and team
- Formats a ready-to-execute `Task(orchestrator, ...)` invocation

### Example Output

```
---
## Orchestrator Routing Required

Session created: **session-20260104-022401-5552866f**

### Next Step

Execute the following Task invocation:

```
Task(orchestrator, "Break down initiative into phases and tasks

Session Context:
- Session ID: session-20260104-022401-5552866f
- Session Path: .claude/sessions/session-20260104-022401-5552866f/SESSION_CONTEXT.md
- Initiative: Your Initiative Name
- Complexity: MODULE
- Team: ecosystem-pack")
```

Copy the Task invocation above and execute it, or use `/consult` for manual routing.

---
```

---

## Flow Comparison

### Before: Manual Flow (3-5 Steps)

| Step | Action | User Effort |
|------|--------|-------------|
| 1 | Type `/start "Initiative Name"` | Command input |
| 2 | Hook outputs YAML CONSULTATION_REQUEST | Must parse YAML |
| 3 | Manually run `session-manager.sh create` | CLI invocation |
| 4 | Construct Task tool invocation with session context | Manual assembly |
| 5 | Execute Task invocation | Paste and run |

**Pain points**:
- YAML output required manual parsing
- Session creation was a separate step
- Task invocation required manual construction with session ID, path, etc.

### After: Automated Flow (2 Steps)

| Step | Action | User Effort |
|------|--------|-------------|
| 1 | Type `/start "Initiative Name"` | Command input |
| 2 | Copy Task invocation from output, execute | Paste and run |

**Improvements**:
- Session created automatically by hook
- Task invocation includes all context fields
- Copy-paste ready format

---

## Hook Behavior

Two hooks coordinate the auto-orchestration flow:

### orchestrator-router.sh (Priority 5)

**Location**: `.claude/hooks/validation/orchestrator-router.sh`

**Event**: UserPromptSubmit

**Matcher**: `^/(start|sprint|task)`

**Behavior**:
1. Checks if orchestrator agent exists in `.claude/agents/orchestrator.md`
2. For `/start` commands: Creates session if none exists via `session-manager.sh create`
3. Reads session context (ID, path, initiative, complexity, team)
4. Outputs ready-to-execute Task invocation with all context populated

**Coordination**:
- Runs BEFORE start-preflight.sh due to lower priority number
- Creates session so start-preflight.sh can skip redundant creation
- Handles special character escaping in initiative names

### start-preflight.sh (Priority 10)

**Location**: `.claude/hooks/session-guards/start-preflight.sh`

**Event**: UserPromptSubmit

**Matcher**: `/start`

**Behavior**:
1. Checks for existing session state
2. When orchestrator present AND session exists: Exits silently (router handled it)
3. When orchestrator present AND no session: Exits silently (router should have created)
4. When NO orchestrator: Creates session via existing logic, outputs status

**Coordination**:
- Defers to orchestrator-router.sh when orchestrator agent is present
- Maintains backward compatibility for non-orchestrated teams
- Handles existing session detection (parked, active)

---

## Friction Measurement

### Success Criteria (from PRD)

| Metric | Baseline | Target | Achieved |
|--------|----------|--------|----------|
| Steps to start session | 3-5 manual steps | 1-2 steps | 2 steps |
| Manual session-manager.sh call | Required | Not required | Automated |
| Manual Task construction | Required | Not required | Automated |
| Copy-paste ready output | No | Yes | Yes |

### Friction Reduction Details

**Eliminated steps**:
1. Manual `session-manager.sh create` invocation
2. Parsing YAML CONSULTATION_REQUEST output
3. Looking up session ID from `.claude/sessions/.current`
4. Constructing Task invocation with session context fields

**Remaining steps**:
1. Type `/start "Initiative Name"` (irreducible)
2. Copy Task invocation and execute (could be automated in future)

---

## Session States

Sessions follow a finite state machine with these states:

| State | Description | Valid Transitions |
|-------|-------------|-------------------|
| ACTIVE | Session in progress | PARKED, ARCHIVED |
| PARKED | Session suspended | ACTIVE, ARCHIVED |
| ARCHIVED | Session completed (terminal) | None |

### State Transitions

```
                    +--------+
                    | ACTIVE |<----+
                    +--------+     |
                       |  |        |
           park        |  | resume |
                       v  |        |
                    +--------+     |
                    | PARKED |-----+
                    +--------+
                       |
           wrap        |
                       v
                    +----------+
                    | ARCHIVED |
                    +----------+
```

### Transition Commands

| From | To | Command |
|------|-----|---------|
| (none) | ACTIVE | `/start "Initiative"` |
| ACTIVE | PARKED | `/park` or `session-manager.sh mutate park` |
| PARKED | ACTIVE | `/resume` or `session-manager.sh mutate resume` |
| ACTIVE | ARCHIVED | `/wrap` or `session-manager.sh mutate wrap` |
| PARKED | ARCHIVED | `/wrap` or `session-manager.sh mutate wrap` |

---

## Dual-Agent Coordination

After `/start`, two agents operate in parallel:

### orchestrator (Workflow Agent)

- Plans workflow phases and task breakdown
- Delegates to specialist subagents via Task tool
- Reads SESSION_CONTEXT.md for context (never writes)
- Outputs instructions for state-mate when state changes needed

### state-mate (State Agent)

- Sole authority for SESSION_CONTEXT.md mutations
- Validates all state transitions against FSM
- Maintains audit trail for all changes
- Handles park, resume, wrap, phase transitions, task completion

### Coordination Pattern

```
User: /start "Feature X"
      |
      v
Hook: Creates session, outputs Task(orchestrator, ...)
      |
      v
User: Executes Task invocation
      |
      v
Orchestrator: Plans workflow, delegates to subagents
      |
      v
Orchestrator: "Phase complete. To transition, use:
               Task(state-mate, 'transition_phase from=requirements to=design...')"
      |
      v
User/Main: Executes Task(state-mate, ...)
      |
      v
state-mate: Validates and applies transition
```

---

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| `/start` with existing ACTIVE session | Router uses existing session, outputs Task invocation |
| `/start` with existing PARKED session | Preflight shows options: resume, wrap, or parallel |
| Session creation fails (lock timeout) | Router outputs CONSULTATION_REQUEST with manual guidance |
| Initiative name with special characters | Properly escaped in Task invocation |
| No orchestrator in team pack | Router exits, preflight handles session creation |
| ACTIVE_RITE not set | Creates cross-cutting session |
| Parallel `/start` in multiple terminals | Locking prevents race conditions |

---

## Manual Fallback

If auto-orchestration fails or you prefer manual control:

### Manual Session Creation

```bash
.claude/hooks/lib/session-manager.sh create "Initiative Name" "MODULE" "team-name"
```

### Manual Orchestrator Invocation

```
Task(orchestrator, "Break down initiative into phases and tasks

Session Context:
- Session ID: <from .claude/sessions/.current>
- Session Path: .claude/sessions/<session-id>/SESSION_CONTEXT.md
- Initiative: Initiative Name
- Complexity: MODULE
- Team: team-name")
```

### Using /consult Instead

The `/consult` command provides guided routing without auto-orchestration:

```
/consult "What workflow should I use for this feature?"
```

---

## Related Documentation

| Topic | Location |
|-------|----------|
| Session FSM Architecture | `docs/session-fsm/ARCHITECTURE.md` |
| Session Operations Guide | `docs/session-fsm/OPERATIONS.md` |
| Auto-Orchestration PRD | `docs/requirements/PRD-auto-orchestration.md` |
| Auto-Orchestration TDD | `docs/design/TDD-auto-orchestration.md` |
| state-mate Agent | `user-agents/state-mate.md` |
| ADR-0005 (state-mate Authority) | `docs/decisions/ADR-0005-state-mate-centralized-state-authority.md` |

---

## Troubleshooting

### Hook Output Missing Task Invocation

**Symptom**: `/start` outputs status but no Task invocation

**Cause**: No orchestrator.md in `.claude/agents/`

**Resolution**: Verify team has orchestrator agent or use manual flow

### Session Not Created

**Symptom**: Task invocation references missing session

**Cause**: Session creation failed (lock timeout, filesystem error)

**Resolution**:
```bash
# Check for lock issues
ls -la .claude/sessions/.locks/

# Create session manually
.claude/hooks/lib/session-manager.sh create "Initiative" "MODULE"
```

### Task Invocation Syntax Error

**Symptom**: Claude reports syntax error on Task invocation

**Cause**: Special characters in initiative not properly escaped

**Resolution**: Re-run `/start` with simpler initiative name or use manual flow

---

## File Verification

| Artifact | Absolute Path |
|----------|---------------|
| This document | `/Users/tomtenuta/Code/roster/docs/workflows/session-lifecycle.md` |
| orchestrator-router.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/validation/orchestrator-router.sh` |
| start-preflight.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/session-guards/start-preflight.sh` |
| session-manager.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-manager.sh` |

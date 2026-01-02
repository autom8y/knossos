# Execution Mode

> Defines the three operating modes based on session state and team context.

## The Three Modes

The roster ecosystem operates in three execution modes:

| Mode | Session | Team | Main Agent Behavior |
|------|---------|------|---------------------|
| **Native** | No | - | Direct execution, no session tracking |
| **Cross-Cutting** | Yes | No (or parked) | Direct execution with session tracking |
| **Orchestrated** | Yes (ACTIVE) | Yes | Coach pattern, delegate via Task tool |

## Mode Detection

**Decision Tree (per PRD-hybrid-session-model FR-1.1):**

```
User Intent
    |
    +-- No session active?
    |       |
    |       +-- Native Mode
    |           Direct execution, no orchestration, no session tracking
    |
    +-- Session active?
            |
            +-- Has team AND session status = ACTIVE?
            |       |
            |       +-- Orchestrated Mode
            |           Main thread = Coach, delegates via Task tool
            |           (Note: Parked sessions are NOT orchestrated)
            |
            +-- No team, OR session parked/not-active?
                    |
                    +-- Cross-Cutting Mode
                        Main agent executes directly
                        Session tracking active
                        /consult available for routing
```

**Programmatic Detection:**
```bash
# From session-manager.sh
MODE=$(execution_mode)  # Returns: native | orchestrated | cross-cutting
```

## Native Mode

**When:** No session file exists.

**Behavior:**
- Direct execution with Edit/Write tools
- No session tracking
- No delegation warnings from hooks
- Start a session with `/start` if tracking is desired

**Appropriate for:**
- Quick questions and answers
- Simple one-off edits
- Exploration and research

## Cross-Cutting Mode

**When:** Session exists but no team active, OR session is parked.

**Behavior:**
- Direct execution with Edit/Write tools
- Session tracking is active (artifacts, blockers, next_steps recorded)
- No delegation required or warned
- `/consult` available for routing guidance

**Appropriate for:**
- Cross-cutting concerns (affect multiple teams)
- Parked session work (exploratory before resuming)
- Sessions started without a team

**To switch to orchestrated mode:**
```
/team <pack-name>
```

## Orchestrated Mode

**When:** Session is ACTIVE and team is configured.

**Behavior:**
- Main thread is the **COACH** - coordinates, does not execute
- **CONSULT** the orchestrator for direction
- **DELEGATE** to specialists via Task tool
- **NEVER** use Edit/Write directly - that is specialist work

### Correct Pattern

```
Main Thread -> [Task tool] -> Orchestrator (returns directive)
Main Thread -> [Task tool] -> Specialist (per directive)
```

### Incorrect Pattern

```
Main Thread -> [Edit/Write] -> Direct implementation
Main Thread -> "Execute the sprint" -> Orchestrator (cannot execute)
```

### Why Delegation Matters

The main thread has limited context window. By delegating to specialists:

1. Each specialist gets fresh context optimized for their task
2. Work can be parallelized when phases are independent
3. Specialists can be swapped without changing orchestration
4. Audit trail is clearer (which agent did what)

## Mode-Aware Hooks

| Hook | Native | Orchestrated | Cross-Cutting |
|------|--------|--------------|---------------|
| `delegation-check.sh` | Silent | Warning on Edit/Write | Silent |
| `session-context.sh` | Mode: native | Mode: orchestrated | Mode: cross-cutting |

## Edge Cases

| Case | Behavior |
|------|----------|
| Session parked with team | Cross-cutting (parked overrides team) |
| Team configured but pack missing | Cross-cutting + warning |
| Session file corrupted | Cross-cutting (graceful fallback) |
| `/handoff` in cross-cutting | Error: "No orchestrator in cross-cutting mode" |

## Mode Switching

| From | To | Command |
|------|----|---------|
| Native | Cross-cutting | `/start "initiative" --no-team` |
| Native | Orchestrated | `/start "initiative"` (with team) |
| Cross-cutting | Orchestrated | `/team <pack>` |
| Orchestrated | Cross-cutting | `/team --remove` |

## Related

- `main-thread-guide.md` - Consultation loop template
- `consultation-loop.md` - How to consult orchestrator
- PRD: `docs/requirements/PRD-hybrid-session-model.md`
- TDD: `docs/design/TDD-hybrid-session-model.md`

# Entry Pattern

> How /start, /sprint, /task commands route through the system

## The Pattern

When an orchestrator is present in the active rite:

```
User: /start Add dark mode toggle
         |
         v
+----------------------------------+
| Hook: orchestrator-router.sh     |
| (Priority 5)                     |
| Injects CONSULTATION_REQUEST     |
+----------------------------------+
         |
         v
+----------------------------------+
| Hook: start-preflight.sh         |
| (Priority 10)                    |
| Creates session (hook-triggered) |
+----------------------------------+
         |
         v
+----------------------------------+
| Main Agent sees routing context  |
| Knows to consult orchestrator    |
+----------------------------------+
         |
         v
+----------------------------------+
| Task(orchestrator, request)      |
| Orchestrator returns directive   |
+----------------------------------+
         |
         v
+----------------------------------+
| Main Agent executes directive    |
| Task(specialist, prompt)         |
+----------------------------------+
         |
         v
+----------------------------------+
| Hooks detect state events        |
| Auto-invoke state-mate if needed |
+----------------------------------+
```

## Hook Execution Order

Hooks execute in priority order (lower numbers first):

| Priority | Hook | Event | Purpose |
|----------|------|-------|---------|
| 5 | orchestrator-router.sh | UserPromptSubmit | Inject orchestrator routing context |
| 10 | start-preflight.sh | UserPromptSubmit | Validate session state, create session |
| 5 | session-write-guard.sh | PreToolUse | Block direct context writes |
| 15 | delegation-check.sh | PreToolUse | Warn on direct implementation |
| 20 | orchestrator-bypass-check.sh | PreToolUse | Warn on orchestrator bypass |

## When Orchestrator is Absent

Teams without an orchestrator agent are valid and use direct execution:

```
User: /start Add dark mode toggle
         |
         v
+----------------------------------+
| Hook: start-preflight.sh         |
| Creates session directly         |
| (No routing context injected)    |
+----------------------------------+
         |
         v
+----------------------------------+
| Main Agent routes to specialist  |
| (no orchestrator consultation)   |
+----------------------------------+
         |
         v
+----------------------------------+
| Task(moirai, ...)           |
| Direct state mutations allowed   |
+----------------------------------+
```

## Hook-Triggered vs Direct state-mate

| Scenario | state-mate Invocation | Audit Log | Notes |
|----------|----------------------|-----------|-------|
| Orchestrator present, active workflow | Hook-triggered | trigger_source: hook | Hooks handle mutations automatically |
| Orchestrator present, no workflow | Direct allowed | trigger_source: direct | No workflow coordination needed |
| Orchestrator absent | Direct allowed | trigger_source: direct | Valid pattern for simpler teams |
| Emergency override | Direct with flag | trigger_source: emergency | Use --emergency flag, logged |

## Session Creation Flow

### Hook-Triggered Creation (New Pattern)

When `/start` is invoked with orchestrator present:

1. **orchestrator-router.sh** (priority 5) injects CONSULTATION_REQUEST
2. **start-preflight.sh** (priority 10) creates session via session-manager.sh
3. Session creation logged with `trigger_source: hook`
4. Main agent receives both routing context and session confirmation

### Audit Trail

```
# Hook-triggered creation
2024-01-15T10:30:00Z | session-20240115-103000-abc123 | CREATE | hook | start-preflight.sh

# Direct creation (orchestrator-less)
2024-01-15T10:35:00Z | session-20240115-103500-def456 | CREATE | direct | Task(moirai)
```

## Anti-Patterns

### DO NOT: Invoke state-mate directly during orchestrated workflows

```yaml
# WRONG
Task(moirai, "transition to design phase")

# RIGHT
Let the artifact-tracker.sh hook detect the PRD write
and trigger the phase transition automatically
```

**Why?** During orchestrated workflows, hooks coordinate state mutations based on orchestrator directives. Direct state-mate calls bypass this coordination and break the audit trail.

### DO NOT: Skip orchestrator and go directly to specialist

```yaml
# WRONG
Task(requirements-analyst, "Create PRD for dark mode")

# RIGHT
Task(orchestrator, "CONSULTATION_REQUEST for dark mode initiative")
# Then invoke specialist per orchestrator directive
```

**Why?** The orchestrator assesses complexity, determines phase sequence, and ensures proper handoffs. Skipping it loses workflow coordination.

### DO NOT: Manually write SESSION_CONTEXT.md

```yaml
# WRONG
Edit(SESSION_CONTEXT.md, ...)

# RIGHT
Use /park, /wrap, or let hooks handle mutations
```

**Why?** Direct writes bypass schema validation, FSM state transitions, and audit logging. Use designated commands or let hooks handle it.

## state_update.trigger_hooks

When orchestrator returns `state_update.trigger_hooks: true`:

1. Main agent should NOT call state-mate directly
2. Hooks will detect relevant events (artifact writes, phase completion) and invoke state-mate
3. Audit trail shows hook-triggered mutations
4. State changes coordinated with orchestrator's expected_transitions

When `trigger_hooks: false` (or absent):

1. Direct state-mate calls are acceptable
2. Typically for orchestrator-less teams or emergency overrides
3. Main agent responsible for state mutations

## expected_transitions

Orchestrator can signal expected state changes:

```yaml
state_update:
  current_phase: requirements
  next_phases: [design, implementation, validation]
  routing_rationale: "Initial phase - requirements gathering needed first"
  trigger_hooks: true
  expected_transitions:
    - type: phase
      from: null
      to: requirements
    - type: artifact
      to: registered
      artifact_path: docs/requirements/PRD-dark-mode.md
```

Hooks can use this to:
- Validate actual transitions match expectations
- Provide better error messages when transitions fail
- Coordinate cross-hook state changes

## Workflow-Aware Error Messages

### session-write-guard.sh

The write guard provides different messages based on context:

**With orchestrator + active workflow:**
```
State mutations are handled automatically by hooks during active workflows.

If you need an explicit mutation, use:
- /park - Pause current session
- /wrap - Complete and archive session
- /handoff - Transfer to another agent

Do not call Task(moirai, ...) directly during orchestrated workflows.
```

**Without orchestrator or no workflow:**
```
Direct writes to *_CONTEXT.md files are not allowed.

Use state-mate for all session/sprint mutations:
  Task(moirai, "<your mutation request>")

Examples:
- Task(moirai, "mark task-001 complete")
- Task(moirai, "transition to design phase")
```

### orchestrator-bypass-check.sh

Warns when main agent invokes specialists without recent orchestrator consultation (last 5 minutes):

```
Warning: Orchestrator Consultation Recommended

You are invoking specialist requirements-analyst without recent orchestrator consultation.

Best Practice: During active workflows, consult the orchestrator first:
  Task(orchestrator, "CONSULTATION_REQUEST with current state...")

Then invoke specialists based on the orchestrator's directive.

*This is a warning only - proceeding with specialist invocation.*
```

**Non-blocking:** Operation continues. This is guidance, not enforcement.

## Fallback Behavior

If hooks fail or are disabled, the system degrades gracefully:

| Failure | Behavior | Impact |
|---------|----------|--------|
| orchestrator-router.sh fails | start-preflight.sh continues, no routing context | Main agent may skip orchestrator consultation |
| start-preflight.sh fails | Main agent can create session manually | Session creation still possible |
| orchestrator-bypass-check.sh fails | Bypass check skipped, specialist invoked | Warning not shown, operation proceeds |
| session-write-guard.sh fails | Write attempt proceeds | State corruption possible (rare) |
| state-mate hook fails | Logged as error, session continues | Session in degraded mode, manual recovery |

All hook failures are logged to `.claude/hooks/logs/hooks.log` with timestamps and error details.

## Performance Characteristics

| Metric | Target | Actual (Typical) |
|--------|--------|------------------|
| orchestrator-router.sh latency | < 50ms | ~20ms |
| start-preflight.sh latency | < 100ms | ~60ms |
| session-write-guard.sh latency | < 30ms | ~10ms |
| orchestrator-bypass-check.sh latency | < 30ms | ~15ms |
| Total routing overhead | < 200ms | ~105ms |

**Design Decision:** All hooks use `set -euo pipefail` and bounded operations (e.g., only check last 20 events) to ensure fast, predictable performance.

## Integration Test Coverage

The implementation includes integration tests validating:

1. **Hook execution order** - orchestrator-router fires before start-preflight
2. **Orchestrator routing** - /start injects CONSULTATION_REQUEST when orchestrator present
3. **Session creation** - Hook-triggered creation with audit trail
4. **Bypass detection** - Warns when skipping orchestrator (non-blocking)
5. **Write guard** - Workflow-aware error messages
6. **Backward compatibility** - Works with orchestrator-less teams

See `TDD-orchestrator-entry-pattern.md` for complete test matrix.

## See Also

- [execution-mode.md](execution-mode.md) - When to delegate vs execute directly
- [consultation-loop.md](consultation-loop.md) - The consultation pattern
- [command-integration.md](command-integration.md) - How commands use the loop
- [response-format.md](response-format.md) - CONSULTATION_RESPONSE schema
- [main-thread-guide.md](main-thread-guide.md) - Main thread execution pattern

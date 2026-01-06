# PRD: Orchestrator Entry Pattern Codification

## Overview

This initiative codifies the execution model for command-based workflows (`/start`, `/sprint`, `/task`) to ensure the Orchestrator serves as the primary entry point when a workflow is available and an orchestrator is present, with state-mate invoked via hooks rather than through direct Task tool calls from the main agent. The goal is to formalize the observed optimization where state-mate operates in parallel with the orchestrator during phased executions.

**Important**: Teams without an orchestrator agent are valid and use direct execution. This PRD applies only when orchestrator is present.

## Background

### Current Architecture

The harness implements a sophisticated consultation loop pattern where:

1. **Main agent** controls execution, owns the Task tool, and delegates to specialists
2. **Orchestrator** acts as a stateless advisor, returning structured directives
3. **state-mate** serves as the centralized authority for session/sprint mutations
4. **Hooks** auto-inject context and guard operations (SessionStart, Stop, PreToolUse, PostToolUse, UserPromptSubmit)

### Observed Problem

When users invoke `/start`, `/sprint`, or `/task`, the main agent too frequently:

1. Manages SESSION_CONTEXT mutations directly instead of delegating through state-mate
2. Attempts to orchestrate work itself rather than consulting the Orchestrator first
3. Invokes state-mate via explicit `Task(moirai, ...)` calls rather than having hooks trigger it

This blurs the separation between:
- **Orchestrator**: Routing and phase coordination (stateless advisor)
- **state-mate**: State mutations (triggered by hooks, not orchestration)

### Desired Pattern

```
/start (or /sprint, /task)
    |
    v
Orchestrator (primary entry when workflow available)
    |
    +-- Hooks detect state-relevant events
    |        |
    |        v
    |   state-mate auto-invoked (session mutations)
    |
    +-- Routes to specialists (phase work)
```

**Key Principle**: state-mate is invoked BY the orchestrator's hooks, not as a standalone entry agent. This prevents state-mate from acquiring routing responsibilities it should not have.

### Stakeholder Context

- **Audience**: Power user (sole user of this harness)
- **Complexity tolerance**: High - the pattern should be robust, not simplified
- **State management**: Mandatory via state-mate, but hook-triggered not manually invoked
- **Pattern to formalize**: state-mate works in parallel with orchestrator during phased executions

## User Stories

### US-1: Automatic Orchestrator Routing on Workflow Entry

As a user invoking `/start`, I want the main agent to immediately route to the Orchestrator so that I get proper phase decomposition and specialist routing without the main agent attempting to manage the workflow itself.

**Acceptance Criteria**:
- Invoking `/start` consults Orchestrator before any specialist work begins
- Orchestrator receives properly formatted CONSULTATION_REQUEST with type: "initial"
- Main agent does not write SESSION_CONTEXT directly; it follows orchestrator directives

### US-2: Hook-Triggered State Mutations

As a user in an active workflow, I want state mutations to occur automatically via hooks so that I do not need to manually invoke state-mate and the orchestrator does not accumulate state management responsibilities.

**Acceptance Criteria**:
- Session creation on `/start` is triggered by a hook, not direct main agent action
- Phase transitions trigger hooks that invoke state-mate
- Artifact registration happens via PostToolUse hooks
- When orchestrator is present, state-mate is invoked via hooks rather than explicit `Task(moirai, ...)`
- Teams without orchestrators may invoke state-mate directly (this is valid for orchestrator-less workflows)

### US-3: Clear Separation of Concerns

As a harness maintainer, I want enforceable guardrails distinguishing orchestrator (routing) from state-mate (mutations) so that agents cannot accumulate responsibilities outside their domain.

**Acceptance Criteria**:
- Orchestrator agent definition explicitly excludes state mutation capability
- Hooks intercept and redirect any direct SESSION_CONTEXT writes to state-mate
- Documentation clearly delineates the boundary

### US-4: Documented Execution Model

As a future contributor to this harness, I want the entry pattern codified in documentation so that I understand how `/start`, `/sprint`, and `/task` should route through the system.

**Acceptance Criteria**:
- Execution model documented in skill file (e.g., `orchestration/entry-pattern.md`)
- Decision flow diagram showing routing logic
- Anti-pattern section showing what NOT to do

## Functional Requirements

### Must Have

#### FR-1: Orchestrator Entry Hook

Create or extend a UserPromptSubmit hook that:
- Detects `/start`, `/sprint`, `/task` commands
- Injects context signaling that Orchestrator consultation is required
- Provides the main agent with the CONSULTATION_REQUEST template to use

**Rationale**: The existing `start-preflight.sh` hook performs pre-flight validation but does not enforce Orchestrator routing.

#### FR-2: Session Creation via Hook-Triggered state-mate

Modify the session creation flow so that:
- When `/start` is invoked and pre-flight passes, a hook triggers state-mate for SESSION_CONTEXT creation
- The main agent receives confirmation that the session exists before proceeding
- The orchestrator is consulted immediately after session creation

**Implementation Notes**:
- May require a new hook event or extension to existing SessionStart behavior
- session-manager.sh handles TTY mapping; state-mate handles schema-validated writes
- The handoff between these must be clean

#### FR-3: State Mutation Interception for Active Workflows

Extend `session-write-guard.sh` to:
- Detect when an active workflow is present (SESSION_CONTEXT exists with `workflow.active: true`)
- Block direct writes to `*_CONTEXT.md` with instruction to let hooks handle it
- Provide clear error message explaining the hook-based mutation pattern

**Current Behavior**: `session-write-guard.sh` blocks writes and suggests `Task(moirai, ...)`.
**Desired Behavior**: During active workflows, the error should instead say: "State mutations are handled automatically by hooks during active workflows. If you need an explicit mutation, use the appropriate command (e.g., `/park`, `/wrap`)."

#### FR-4: Orchestrator Directive for state-mate Coordination

Add to the Orchestrator's `state_update` response field the ability to signal:
- Which state transitions should occur
- What artifacts to register
- Phase progression

The main agent interprets these as "hooks should handle this" rather than "I should call state-mate directly."

**Example Response Fragment**:
```yaml
state_update:
  trigger_hooks: true  # Signal to main agent: let hooks handle state
  expected_transitions:
    - session_state: ACTIVE -> requirements
    - artifact: PRD to be registered
  next_phases:
    - requirements
    - design  # if complexity > SCRIPT
```

#### FR-5: Entry Pattern Documentation

Extend `orchestration/behavior.md` (existing skill file) to document:
- Decision tree for command routing
- Hook triggering sequence
- state-mate invocation conditions
- Orchestrator consultation protocol
- Explicit anti-patterns

#### FR-6: Validation Hook for Orchestrator Consultation

Add a PreToolUse hook that:
- Detects when main agent attempts to invoke specialists without prior Orchestrator consultation
- Only activates when orchestrator is present in the active team
- Warns (does not block) to avoid breaking orchestrator-less workflows
- Logs the violation for audit, minimal context window impact

**Use Case**: Catch cases where main agent skips Orchestrator and goes directly to Requirements Analyst.

### Should Have

#### FR-S.1: Dry-Run Mode for Entry Pattern

Allow `/start --dry-run` to:
- Show what Orchestrator consultation would produce
- Preview SESSION_CONTEXT that would be created
- List specialists that would be invoked for given complexity

### Could Have

#### FR-C.1: Entry Pattern Diagram Generator

Create a script that generates visual diagrams (Mermaid or ASCII) showing the current entry pattern configuration.

## Non-Functional Requirements

### NFR-1: Performance

- Hook execution for entry pattern detection must complete in < 50ms
- Total overhead for routing through Orchestrator vs. direct execution must be < 200ms
- No perceptible latency increase for `/start` command

### NFR-2: Reliability

- Orchestrator "failure" is defined as timeout or parse error only (valid responses with unexpected content are trusted)
- If Orchestrator consultation fails (timeout/parse error), fallback to direct execution with warning
- If state-mate hook fails, session creation should not be blocked (log and continue)
- Partial failures should not corrupt session state

### NFR-3: Backwards Compatibility

- Existing sessions created without the new pattern must continue to work
- Direct `Task(moirai, ...)` calls must still work for edge cases (escape hatch)
- Documentation must note when direct invocation is acceptable

### NFR-4: Maintainability

- Entry pattern logic must be centralized (not scattered across multiple hooks)
- Pattern changes should require updating at most 2-3 files
- Clear ownership: orchestration skill owns the pattern documentation

### NFR-5: Observability

- Audit log must indicate whether state mutation was hook-triggered or direct
- Session mutations log must show trigger source: `hook` | `direct` | `emergency`

## Edge Cases

| Case | Expected Behavior |
|------|-------------------|
| `/start` when session already exists | Pre-flight hook blocks, suggests `/resume` or `/wrap` |
| Orchestrator consultation times out | Fall back to direct execution with warning in session log |
| state-mate hook fails during session creation | Log error, allow session creation to proceed, mark as degraded |
| User explicitly calls `Task(moirai, ...)` | Allowed but logged as `direct` invocation for audit |
| `/sprint` with zero tasks defined | Orchestrator returns guidance to define tasks, no specialist invoked |
| Hook disabled or missing | Graceful degradation; main agent executes directly (legacy mode) |
| Multiple `/start` commands in rapid succession | First wins; subsequent blocked by session-exists guard |
| `/task` invoked outside any session | Native mode execution per PRD-hybrid-session-model (no auto-session creation) |
| Orchestrator returns malformed CONSULTATION_RESPONSE | Main agent logs error, asks user for guidance |
| state-mate schema validation fails | Operation blocked, clear error returned, session state unchanged |

## Success Criteria

- [ ] `/start`, `/sprint`, `/task` route through Orchestrator when orchestrator is present
- [ ] When orchestrator present: state-mate is invoked via hooks, not explicit `Task(moirai, ...)`
- [ ] When orchestrator absent: direct state-mate invocation is valid and works correctly
- [ ] Clear separation documented: Orchestrator = routing, state-mate = mutations
- [ ] Entry pattern documented in `orchestration/entry-pattern.md`
- [ ] `session-write-guard.sh` provides workflow-aware error messages
- [ ] Audit trail distinguishes hook-triggered vs. direct state mutations
- [ ] Existing sessions continue to work (backwards compatible)
- [ ] No measurable latency increase for `/start` command (< 200ms overhead)

## Dependencies

| Dependency | Type | Owner | Status |
|------------|------|-------|--------|
| `session-write-guard.sh` modification | Internal | hooks | Ready |
| `start-preflight.sh` extension | Internal | hooks | Ready |
| Orchestrator agent update | Internal | agents | Ready |
| state-mate agent (no changes) | Internal | agents | Stable |
| `execution-mode.md` skill | Internal | orchestration skill | Ready |
| `base_hooks.yaml` configuration | Internal | hooks | Ready |

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Hook complexity increases debugging difficulty | Medium | Medium | Add verbose logging mode for entry pattern hooks |
| Performance regression from additional hook layers | Low | Medium | Benchmark before/after; optimize hook execution |
| Main agent ignores Orchestrator directives | Medium | High | Add validation hook (FR-6) to enforce compliance |
| Confusion between hook-triggered and direct state-mate | Medium | Medium | Document clearly; add audit log differentiation |
| Breaking existing workflows during migration | Low | High | Backwards compatibility requirement; phased rollout |
| Coach mode is legacy tech debt | High | Medium | **Prereq**: Coach mode rearchitecture needed before full execution; this PRD removes coach mode dependency |

## Out of Scope

- Changes to state-mate's internal logic (it works correctly as-is)
- Changes to the 11-team structure
- Simplification of complexity (high complexity is acceptable and desired)
- Changes to the CONSULTATION_REQUEST/RESPONSE schema (beyond `state_update` field)
- Modification of specialist agent definitions
- Changes to how Orchestrator routes to specialists (only entry routing addressed)
- Automatic hook configuration via UI (CLI/file-based configuration only)

## Open Questions

*None remaining - all requirements captured from stakeholder briefing.*

---

## File Verification

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-orchestrator-entry-pattern.md` | Created |

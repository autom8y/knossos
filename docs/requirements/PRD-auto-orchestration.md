# PRD: Auto-Orchestration Hook Enhancement

## Overview

This PRD addresses the friction in session initialization workflow. Users currently must execute 3-5 manual steps to start an orchestrated session, including manually invoking state-mate and orchestrator agents. The goal is to reduce this to 1-2 steps via automatic session creation on `/start` and ready-to-execute Task tool invocation output from hooks.

## Background

The roster ecosystem has sophisticated hook infrastructure that can auto-inject context, but the session bootstrap process still requires manual intervention:

**User Frustration (Verbatim)**:
> "Every time I start a session I have to interrupt Claude to say 'invoke the state-mate and orchestrator agents'. This seems ridiculous with the tooling we have."

**Current State (3-5 Manual Steps)**:
1. User types `/start "Initiative Name"`
2. Hook outputs CONSULTATION_REQUEST in YAML format
3. User must manually copy/construct Task tool invocation
4. User must invoke state-mate for session creation (if not auto-created)
5. User must invoke orchestrator with session context

**Desired State (1-2 Steps)**:
1. User types `/start "Initiative Name"`
2. Hook auto-creates session AND outputs ready-to-execute Task invocation
3. User copies Task invocation directly into Claude (or Claude auto-executes in future)

**Design Philosophy (User Preferences)**:
- Silent magic preferred: Automation should work without announcements
- Recover-then-report: Handle errors gracefully, don't ask permission for obvious fixes
- High autonomy with upfront questions: Clarify requirements once, then execute
- Artifact gates mark explore-to-execute boundary: PRD/TDD approval required before implementation
- Google Shell Style Guide: All bash scripts follow established conventions

## Dual-Agent Coordination Model

This initiative involves two agents operating in parallel during orchestrated sessions:

### Agent Responsibilities

| Agent | Responsibility | State Ownership |
|-------|----------------|-----------------|
| **orchestrator** | Workflow planning, phase decomposition, subagent delegation | Read-only access to session state |
| **state-mate** | Session state mutations, lifecycle transitions, audit trail | Write authority over SESSION_CONTEXT.md |

### Parallel Operation Model

The orchestrator and state-mate operate in parallel, not sequentially:

```
     /start "Initiative"
            |
            v
    +-----------------+
    | Session Created |
    | (session-mgr.sh)|
    +-----------------+
            |
    +-------+-------+
    |               |
    v               v
+-------------+  +-------------+
| orchestrator|  | state-mate  |
| (workflow)  |  | (state)     |
+-------------+  +-------------+
    |               |
    | Consults      | Manages
    | subagents     | SESSION_CONTEXT.md
    | for work      | SPRINT_CONTEXT.md
    |               |
    v               v
+-------------+  +-------------+
| Subagent    |  | Mutations   |
| Delegation  |  | & Audit Log |
+-------------+  +-------------+
    |               |
    +-------+-------+
            |
            v
    +-----------------+
    | Coordinated     |
    | Session Progress|
    +-----------------+
```

### Integration Points

1. **Orchestrator reads state**: Orchestrator reads SESSION_CONTEXT.md to understand current phase, but never writes
2. **Orchestrator delegates state changes**: When workflow requires state transition, orchestrator instructs user/main-thread to invoke state-mate
3. **State-mate enforces lifecycle**: state-mate validates all transitions against FSM before applying
4. **Parallel execution**: Both agents can be invoked independently within the same session

### Clear Boundaries

| Action | Owner | How |
|--------|-------|-----|
| Session creation | session-manager.sh | `create` command (hooks invoke this) |
| Workflow breakdown | orchestrator | Task tool delegation |
| Phase transitions | state-mate | `transition_phase` operation |
| Task completion tracking | state-mate | `mark_complete` operation |
| Sprint management | state-mate | `create_sprint`, `start_sprint` operations |
| Subagent coordination | orchestrator | CONSULTATION_REQUEST/RESPONSE loop |

### Why Parallel, Not Sequential

Previous designs assumed sequential flow: orchestrator plans, then state-mate records. This created friction:
- Users had to manually invoke state-mate after orchestrator completed
- State updates lagged behind actual workflow progress

The parallel model allows:
- Orchestrator to focus purely on workflow coordination
- State-mate to handle all state mutations independently
- Users to invoke either agent as needed during session

---

## User Stories

### US-1: Zero-Friction Session Start

- **US-1.1**: As a developer, I want `/start "Initiative"` to automatically create my session, so that I don't have to manually invoke session-manager.sh.

- **US-1.2**: As a developer, I want the hook output to include a ready-to-execute Task invocation, so that I can simply copy-paste to proceed.

- **US-1.3**: As a returning user, I want `/start` to detect existing sessions and provide appropriate options (resume/wrap/parallel), maintaining current behavior.

### US-2: Consultation Request Formatting

- **US-2.1**: As Claude (the main agent), I want CONSULTATION_REQUEST output to be in Task tool format, so that I can execute it directly without transformation.

- **US-2.2**: As a developer, I want the Task invocation to include all necessary session context (ID, path, initiative, complexity), so that the orchestrator has full context.

- **US-2.3**: As a power user, I want the option to use the manual `/consult` flow if I prefer, so that automation doesn't remove flexibility.

### US-3: Error Recovery

- **US-3.1**: As a developer, I want session creation failures to be handled gracefully with clear recovery guidance, so that I'm not blocked.

- **US-3.2**: As the system, I want lock race conditions prevented during parallel session operations, so that sessions are not corrupted.

### US-4: State-Mate Coordination

- **US-4.1**: As the orchestrator agent, I need read-only access to SESSION_CONTEXT.md to understand current phase, so that I can plan appropriate workflow steps without mutating state.

- **US-4.2**: As the system, I want state-mate to handle all SESSION_CONTEXT.md mutations independently from orchestrator, so that state changes are properly validated and audited.

- **US-4.3**: As a developer, I want clear guidance on when to invoke state-mate vs orchestrator, so that I don't confuse their roles during workflow execution.

## Functional Requirements

### Must Have

#### FR-1: Session Bootstrap Automation (Phase 1)

- **FR-1.1**: `start-preflight.sh` MUST auto-create a session via `session-manager.sh create` when `/start` is invoked and no active session exists.
  - **Input**: `/start "Initiative Name" [COMPLEXITY]`
  - **Action**: Call `session-manager.sh create "$INITIATIVE" "$COMPLEXITY" "$ACTIVE_RITE"`
  - **Output**: Session created message with session ID
  - **Current Status**: Already implemented in start-preflight.sh lines 113-126

- **FR-1.2**: `orchestrator-router.sh` MUST detect if session was just created by start-preflight.sh and skip redundant session checks.
  - **Condition**: Session exists and is ACTIVE (just created)
  - **Behavior**: Proceed directly to CONSULTATION_REQUEST output

- **FR-1.3**: Session creation MUST use proper locking to prevent race conditions with parallel operations.
  - **Lock file**: `$SESSIONS_DIR/.create.lock`
  - **Timeout**: 10 seconds with retry
  - **Current Status**: Already implemented in session-manager.sh

#### FR-2: Consultation Request Injection (Phase 2)

- **FR-2.1**: `orchestrator-router.sh` MUST output a ready-to-execute Task tool invocation instead of raw YAML.
  - **Old Format** (current):
    ```yaml
    type: initial
    initiative:
      name: "Initiative Name"
    ```
  - **New Format** (desired):
    ```
    Task(orchestrator, "Break down initiative into phases and tasks

    Session Context:
    - Session ID: session-20260104-022401-5552866f
    - Session Path: .claude/sessions/session-20260104-022401-5552866f/SESSION_CONTEXT.md
    - Initiative: Initiative Name
    - Complexity: MODULE")
    ```

- **FR-2.2**: Task invocation MUST include all relevant session context automatically populated:
  - Session ID (from just-created session or current session)
  - Session Path (absolute path to SESSION_CONTEXT.md)
  - Initiative name (from `/start` command)
  - Complexity level (from `/start` command or default MODULE)
  - Active team (if configured)

- **FR-2.3**: Hook output MUST include brief execution guidance:
  - Primary: "Next: Execute the Task invocation above"
  - Fallback: "Or use `/consult` for manual routing"

- **FR-2.4**: CONSULTATION_REQUEST format MUST be valid for copy-paste execution.
  - No placeholder values (all fields auto-filled)
  - Proper escaping for special characters in initiative name
  - Task tool syntax matches Claude Code expectations

#### FR-3: Verification & Testing (Phase 3)

- **FR-3.1**: Test `/start "Test Initiative"` creates SESSION_CONTEXT.md without manual intervention.

- **FR-3.2**: Test `/start` with existing session reuses session (does not create duplicate).

- **FR-3.3**: Verify lock race conditions prevented with parallel `/start` operations.

- **FR-3.4**: Verify CONSULTATION_REQUEST Task invocation is syntactically valid and copy-paste executable.

- **FR-3.5**: Measure and document friction reduction:
  - Baseline: 3-5 manual steps
  - Target: 1-2 steps (start -> execute output)

#### FR-4: State-Mate Session Throughline (Phase 4)

- **FR-4.1**: state-mate MUST be the sole authority for SESSION_CONTEXT.md mutations.
  - Orchestrator reads session state for context but NEVER writes
  - All state transitions (park, resume, wrap, phase change) go through state-mate
  - Audit trail maintained for all mutations

- **FR-4.2**: Orchestrator MUST delegate state changes to state-mate rather than mutating directly.
  - When orchestrator determines phase transition is needed, it outputs instruction for state-mate invocation
  - Pattern: `Task(state-mate, "transition_phase from=requirements to=design...")`

- **FR-4.3**: Hook output MAY include parallel Task invocations for both agents when appropriate.
  - Example: `/start` could output both orchestrator invocation AND state-mate initialization if needed
  - Both agents operate independently on the same session

- **FR-4.4**: state-mate MUST handle session creation tracking in SESSION_CONTEXT.md.
  - Record session metadata (initiative, complexity, team)
  - Set initial phase based on workflow
  - Initialize task tracking structure

- **FR-4.5**: Orchestrator output MUST NOT contain Write/Edit operations targeting *_CONTEXT.md files.
  - PreToolUse hook blocks direct writes (per ADR-0005)
  - Orchestrator should instruct main thread to invoke state-mate for state changes

### Should Have

- **FR-S.1**: Detect if user has executed similar `/start` pattern before and suppress verbose guidance (silent mode).

- **FR-S.2**: Add session ID to hook output header for visibility.

- **FR-S.3**: Support `/start --quiet` flag to suppress guidance text.

### Could Have

- **FR-C.1**: Auto-execute the Task invocation (full automation) - deferred for future PRD due to safety considerations.

- **FR-C.2**: Add telemetry to measure actual friction reduction.

## Non-Functional Requirements

- **NFR-1**: Hook execution MUST complete in <100ms total (session creation + output generation).

- **NFR-2**: Graceful degradation: If session creation fails, still output CONSULTATION_REQUEST with manual guidance.

- **NFR-3**: No breaking changes to existing `/start` behavior for users who prefer manual flow.

- **NFR-4**: All bash scripts MUST follow Google Shell Style Guide conventions.

- **NFR-5**: Hook output MUST be compatible with Claude Code's context injection system.

## Edge Cases

| Case | Expected Behavior |
|------|------------------|
| `/start` with existing active session | Display options: park/wrap/parallel (current behavior) |
| `/start` with existing parked session | Display options: continue/wrap/parallel (current behavior) |
| Session creation fails (lock timeout) | Output CONSULTATION_REQUEST with manual session creation guidance |
| Session creation fails (filesystem error) | Output error message with recovery steps |
| Initiative name contains special characters | Properly escape in Task invocation |
| No orchestrator in team pack | Skip CONSULTATION_REQUEST, proceed with direct execution guidance |
| ACTIVE_RITE not set | Create cross-cutting session, adjust Task invocation accordingly |
| `/start` in worktree | Include worktree context in Task invocation |
| Parallel `/start` in multiple terminals | Locking prevents race conditions; second terminal gets "session exists" |

## Success Criteria

- [ ] `/start` command auto-creates SESSION_CONTEXT.md (no manual session-manager.sh call)
- [ ] Hook outputs ready-to-execute Task tool invocation (no manual construction)
- [ ] Task invocation includes all required context (session ID, path, initiative, complexity)
- [ ] Friction reduced from 3-5 manual steps to 1-2 steps
- [ ] No regression in existing session management features
- [ ] Proper locking prevents race conditions
- [ ] All existing tests pass (backwards compatibility)
- [ ] New tests cover auto-orchestration flow

## Dependencies and Risks

### Dependencies

| Dependency | Type | Owner | Status |
|------------|------|-------|--------|
| `session-manager.sh` | Internal | roster | Ready (create command exists) |
| `start-preflight.sh` | Internal | roster | Ready (partial implementation exists) |
| `orchestrator-router.sh` | Internal | roster | Ready (needs output format update) |
| Session FSM | Internal | roster | Ready |

### Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Hook timing conflicts between start-preflight.sh and orchestrator-router.sh | Medium | Medium | Coordinate via priority (router runs first at priority 5, preflight at 10) |
| Task invocation format rejected by Claude | Low | High | Test with actual Claude Code execution |
| Users confused by changed output format | Low | Medium | Include brief migration note in first release |
| Performance regression from session creation in hook | Low | Medium | Benchmark and optimize; session creation already fast |

## Out of Scope

- Auto-execution of Task invocation (future safety review required)
- Changes to state-mate agent core behavior (existing ADR-0005 defines authority)
- Changes to session schema (no schema changes needed)
- Cross-team orchestration changes
- UI/visual indicators for session status
- Orchestrator internal logic changes (only output format changes)

## Open Questions

*Resolved during session creation:*

1. **Q**: Should hooks auto-invoke Task tool?
   **A**: No - output ready-to-execute format but let user/Claude execute. This preserves user control and avoids complexity.

2. **Q**: Which hook should create the session?
   **A**: `start-preflight.sh` at priority 10 (after orchestrator-router.sh at priority 5). Router checks for orchestrator presence, preflight handles session creation.

3. **Q**: Should we support both old YAML format and new Task format?
   **A**: No - replace YAML with Task format. Old format provides no value if Task format works.

---

## Traceability

| Requirement | Source |
|-------------|--------|
| FR-1.x (Session Bootstrap) | User: Manual invocation complaint |
| FR-2.x (Consultation Request) | User: "Copy the Task invocation above and execute it" |
| FR-3.x (Verification) | Standard QA requirements |
| FR-4.x (State-Mate Coordination) | ADR-0005: state-mate centralized authority |
| US-4.x (State-Mate Stories) | Session refinement: dual-agent model clarification |
| NFR-1 (Performance) | Existing hook latency requirements |
| NFR-4 (Shell Style) | User: Design philosophy confirmation |

---

## File Verification

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-auto-orchestration.md` | Updated |
| orchestrator-router.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/validation/orchestrator-router.sh` | Read |
| start-preflight.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/session-guards/start-preflight.sh` | Read |
| session-manager.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-manager.sh` | Read |
| SESSION_CONTEXT.md | `/Users/tomtenuta/Code/roster/.claude/sessions/session-20260104-022401-5552866f/SESSION_CONTEXT.md` | Read |
| state-mate agent | `/Users/tomtenuta/Code/roster/user-agents/state-mate.md` | Read |
| ADR-0005 | `/Users/tomtenuta/Code/roster/docs/decisions/ADR-0005-state-mate-centralized-state-authority.md` | Read |

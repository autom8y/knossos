# Context Design: W4 Moirai Fates Implementation

**Author**: context-architect
**Date**: 2026-02-08
**Status**: Ready for Implementation
**TDD Reference**: `docs/design/TDD-fate-skills.md`
**Backward Compatibility**: COMPATIBLE (additive changes only)

---

## 1. Solution Architecture

### 1.1 Overview

Three workstreams, all additive, no breaking changes:

| Workstream | Scope | Risk |
|------------|-------|------|
| WS-1: Fate Skill Files | Create 4 mena legomena source files | Low -- new files only |
| WS-2: Write Guard Fix | Modify writeguard bypass mechanism | Medium -- security surface |
| WS-3: Dromena Refactor | Modify 5 existing dromena to delegate to Moirai | Medium -- behavioral change |

### 1.2 Key Design Decision: Skill Path

**Decision**: Place fate skills at `mena/session/moirai/` (NOT `mena/moirai/`).

**Rationale**: The existing session skills (`session/common/`, `session/shared/`) establish the convention that session-related legomena live under `mena/session/`. The TDD's `.claude/skills/moirai/` path was written before the mena projection system existed. Following the established pattern yields `.claude/skills/session/moirai/`, which is consistent and discoverable.

**Rejected alternative**: Top-level `mena/moirai/` would project to `.claude/skills/moirai/` matching the TDD literally, but breaks the grouping convention used by all other session-related mena. The TDD's path was aspirational; the mena projection system is authoritative.

**Impact on Moirai agent**: The Read paths in the Moirai agent prompt must reference `.claude/skills/session/moirai/INDEX.md` (not `.claude/skills/moirai/SKILL.md`). The TDD uses `SKILL.md` as the index filename, but the mena convention uses `INDEX.lego.md` which strips to `INDEX.md`. We follow the mena convention.

### 1.3 Key Design Decision: Companion Files vs. Subdirectories

**Decision**: Use companion files within the `mena/session/moirai/` directory (flat structure).

**Rationale**: The mena projection system treats directories-with-INDEX-files as leaf entries. If `clotho/`, `lachesis/`, and `atropos/` were subdirectories with their own INDEX files, they would project as separate skills, defeating the purpose of progressive disclosure within Moirai's skill space. Companion files (non-INDEX `.md` files within the same directory) are copied alongside the INDEX file and accessed via relative paths, which is exactly the progressive disclosure pattern already used by `mena/session/common/` (which has 8 companion files).

```
mena/session/moirai/
  INDEX.lego.md      -> .claude/skills/session/moirai/INDEX.md
  clotho.md          -> .claude/skills/session/moirai/clotho.md
  lachesis.md        -> .claude/skills/session/moirai/lachesis.md
  atropos.md         -> .claude/skills/session/moirai/atropos.md
```

---

## 2. WS-1: Fate Skill File Design

### 2.1 File Inventory

| Source Path | Projected Path | Purpose | Token Budget |
|-------------|---------------|---------|--------------|
| `mena/session/moirai/INDEX.lego.md` | `.claude/skills/session/moirai/INDEX.md` | Routing table, error codes, control flags | ~60 lines |
| `mena/session/moirai/clotho.md` | `.claude/skills/session/moirai/clotho.md` | Creation operations (2 ops) | ~80 lines |
| `mena/session/moirai/lachesis.md` | `.claude/skills/session/moirai/lachesis.md` | Measurement operations (8 ops) | ~180 lines |
| `mena/session/moirai/atropos.md` | `.claude/skills/session/moirai/atropos.md` | Termination operations (3 ops) | ~120 lines |

### 2.2 INDEX.lego.md Specification

```yaml
# Frontmatter
name: moirai-fates
description: "Moirai operation routing table mapping session lifecycle operations to Fate domains (Clotho/Lachesis/Atropos). Use when: Moirai agent needs to determine which Fate skill to load for a given operation. Triggers: moirai routing, fate lookup, operation dispatch, session operation."
```

**Content structure**:
1. Routing table (operation -> fate -> domain -> CLI command)
2. Domain file references (clotho.md, lachesis.md, atropos.md)
3. Loading protocol (6-step: parse -> lookup -> read fate -> execute -> CLI -> respond)
4. Error codes table (7 codes from TDD)
5. Control flags table (--dry-run, --emergency, --override)

**Content does NOT include**: Individual operation specs, validation rules, response schemas. Those live in the companion fate files (progressive disclosure).

### 2.3 clotho.md Specification

This is a **companion file** (no frontmatter needed -- CC does not auto-load companion files; Moirai reads them explicitly via the Read tool).

**Content structure**:
1. Header: "Clotho - The Spinner" with mythological epigraph
2. `create_sprint` operation: syntax, parameters table, validation rules (4), file creation path, sprint context initial YAML, success/error response JSON, example
3. `start_sprint` operation: syntax, parameters table, validation rules (3), state transition, success/error response JSON
4. Anti-patterns table (4 entries)
5. Natural language mapping table (6 entries)

**All content drawn directly from TDD sections 5.2-5.3.** No additions, no omissions. The TDD content IS the implementation spec for this file.

### 2.4 lachesis.md Specification

Companion file, no frontmatter.

**Content structure**:
1. Header: "Lachesis - The Measurer" with mythological epigraph
2. Eight operations, each with: syntax, parameters, validation, CLI command (where applicable), success response, error responses
   - `mark_complete`
   - `transition_phase` (includes phase FSM diagram)
   - `update_field` (includes read-only fields list)
   - `park_session`
   - `resume_session`
   - `handoff` (includes valid agents list)
   - `record_decision`
   - `append_content`
3. Anti-patterns table (5 entries)
4. Natural language mapping table (10 entries)

**All content drawn directly from TDD section 6.3.** This is the largest skill file (~180 lines). Each operation is compact: syntax block, one parameters table, validation list, one CLI command line, one JSON response example. No prose padding.

### 2.5 atropos.md Specification

Companion file, no frontmatter.

**Content structure**:
1. Header: "Atropos - The Cutter" with mythological epigraph
2. `wrap_session`: syntax, parameters, validation (3 rules), state transition, CLI command, internal flow diagram, success response (WHITE sails), error response (BLACK sails blocked), wrapping PARKED session note
3. `generate_sails`: syntax, parameters, output location, CLI command, sails color computation table, proof types table, modifiers list, success response
4. `delete_sprint`: syntax, parameters, validation (3 rules), CLI command, success response (delete), success response (archive)
5. Anti-patterns table (6 entries)
6. Natural language mapping table (7 entries)

**All content drawn directly from TDD section 7.3.**

### 2.6 Response Format Contract

All fate skills reference the same JSON response schema. This is NOT duplicated in each file. Instead, the INDEX.lego.md includes a one-line reference:

```
Response schema: See TDD-fate-skills.md section 3.4.
```

When the Moirai agent is created (separate from this sprint), its agent prompt will embed the response schema directly. The skills provide operation-specific guidance; the agent provides the execution framework.

---

## 3. WS-2: Write Guard Bypass Design

### 3.1 Problem Statement

The current writeguard uses `MOIRAI_BYPASS` env var (line 69 of `writeguard.go`). This cannot work for Moirai-as-subagent because:
1. CC subagents execute in isolated environments
2. Env vars set by the main thread are NOT propagated to subagent tool calls
3. Each Bash/Write/Edit tool call from a subagent spawns a fresh process

### 3.2 Design Options Considered

**Option A: Sentinel file bypass** -- Moirai creates a `.moirai-active` file before writing; writeguard checks for it; Moirai deletes it after. Rejected: race condition between sentinel creation and write guard check; also requires Moirai to have Write permission (circular).

**Option B: Tool input annotation** -- Use CC's `updatedInput` mechanism to inject a bypass token into Write/Edit tool input. Rejected: PreToolUse hooks cannot modify another hook's behavior; annotations in tool input leak user-visible complexity.

**Option C: Session lock file signal** -- Moirai acquires a session lock via `ari session lock`; writeguard checks if a Moirai lock is held. Selected: integrates with existing lock infrastructure, no new mechanisms, deterministic.

**Option D: Disable writeguard during subagent execution** -- Use CC's SubagentStart/SubagentEnd hooks to toggle writeguard. Rejected: CC does not provide hook-to-hook state passing; SubagentStart fires once and hooks are stateless per invocation.

### 3.3 Selected Design: Session Lock Signal (Option C)

**Mechanism**: When Moirai is invoked via `Task(moirai, ...)`, the Moirai agent's first action is to acquire an exclusive advisory lock via `ari session lock --agent moirai`. The writeguard hook checks for this lock. If a lock exists and its `agent` field is `moirai`, the write is allowed.

**Lock lifecycle**:
```
1. Main thread: Task(moirai, "park_session ...")
2. Moirai subagent starts
3. Moirai: Bash("ari session lock --agent moirai")
   -> Creates .claude/sessions/{session}/.moirai-lock (JSON)
4. Moirai: Write(SESSION_CONTEXT.md, ...)
   -> writeguard fires
   -> Checks for .moirai-lock
   -> Lock exists with agent=moirai
   -> ALLOW
5. Moirai: Bash("ari session unlock --agent moirai")
   -> Removes .moirai-lock
6. Moirai returns result to main thread
```

**Lock file schema**:
```json
{
  "agent": "moirai",
  "acquired_at": "2026-02-08T10:00:00Z",
  "session_id": "session-abc123",
  "stale_after_seconds": 300
}
```

**writeguard.go changes** (file: `/Users/tomtenuta/Code/knossos/internal/cmd/hook/writeguard.go`):

1. Remove lines 22-23 (`BypassEnvVar` constant) -- env var bypass is dead code
2. Remove lines 68-71 (env var check block)
3. Add new function `isMoiraiLockHeld(projectDir string) bool`:
   - Resolve session directory from `.claude/sessions/.current-session`
   - Check for `.moirai-lock` file in session directory
   - Parse JSON, verify `agent` field is `moirai`
   - Check `stale_after_seconds` -- if lock is older than threshold, treat as not held (stale lock protection)
   - Return true if valid non-stale Moirai lock exists
4. In `runWriteguardCore()`, after the `isProtectedFile()` check returns true:
   - Call `isMoiraiLockHeld(hookEnv.GetProjectDir())`
   - If true: `return outputAllow(printer)`
   - If false: `return outputBlock(printer, filePath)` (existing behavior)

**New CLI commands** (file: `/Users/tomtenuta/Code/knossos/internal/cmd/session/lock.go`, new file):
- `ari session lock --agent moirai` -- Creates `.moirai-lock` file
- `ari session unlock --agent moirai` -- Removes `.moirai-lock` file

These are thin wrappers over file create/delete with JSON marshaling and stale-lock cleanup.

### 3.4 Security Properties

| Property | Guarantee |
|----------|-----------|
| Only Moirai can acquire lock | Lock command validates `--agent` flag against known agent names |
| Stale lock recovery | Lock file includes `stale_after_seconds` (default 300); writeguard ignores stale locks |
| No TOCTOU race | writeguard reads lock file atomically; lock is created before any writes |
| Main thread cannot bypass | Main thread does not call `ari session lock --agent moirai` (only Moirai agent prompt contains this instruction) |
| Backward compatible | Removing the env var check is safe: the env var was never set in any production flow (it was designed for future Moirai use that never materialized) |

### 3.5 Why Not Simpler?

The simplest approach would be "remove the writeguard entirely" since Moirai is the only writer. Rejected because: (a) the writeguard prevents the main thread and non-Moirai subagents from accidentally writing `*_CONTEXT.md` files, which is a real protection; (b) removing it defeats the design principle that context mutations go through a single auditable agent.

---

## 4. WS-3: Dromena Refactor Design

### 4.1 Target Pattern

Every session dromena follows the same refactored structure:

```
1. Gather context (read session state, parse arguments)
2. Validate pre-conditions (session exists, correct state)
3. Delegate to Task(moirai, "{operation} {parameters}")
4. Format and display Moirai's response to user
```

The dromena STOPS being the executor. It becomes a thin delegation layer. All validation rules, CLI invocations, and state mutations move into the Moirai agent (which reads the appropriate Fate skill for guidance).

### 4.2 Reference Pattern: /park Before/After

**BEFORE** (`/Users/tomtenuta/Code/knossos/mena/session/park/INDEX.dro.md`):

The current /park dromena:
- Has `allowed-tools: Bash, Read, Write` (can write directly)
- Contains inline behavior specification (capture state, execute atomic park, generate summary)
- Calls `ari session park -r "Manual park"` directly
- Formats its own output

**AFTER** (target state):

```yaml
---
name: park
description: Pause work session and preserve state for later
argument-hint: "[reason]"
allowed-tools: Bash, Read, Task
model: sonnet
disable-model-invocation: true
---

## Context
Auto-injected by SessionStart hook (project, team, session, git).

## Your Task

Park the current work session. $ARGUMENTS

## Pre-flight

1. Verify an active session exists (`ari session status -o json` succeeds)
2. Verify session is not already parked

## Delegation

Delegate the park operation to Moirai:

```
Task(moirai, "park_session reason=\"$REASON\"")
```

Where $REASON is extracted from $ARGUMENTS, or "Manual park" if no reason provided.

Moirai will:
1. Read the routing table (.claude/skills/session/moirai/INDEX.md)
2. Load Lachesis skill (.claude/skills/session/moirai/lachesis.md)
3. Validate session state
4. Execute `ari session park --reason "$REASON"`
5. Return structured JSON response

## Display

Format Moirai's response for the user:

```
Session parked at {state_after.parked_at}

Reason: {park_reason}
Resume with: /continue
```

## Error Handling

If Moirai returns `success: false`:
- Display the error message and hint
- Do NOT retry automatically
```

**Key changes**:
1. `Write` removed from `allowed-tools`, `Task` added (Moirai writes, not the dromena)
2. Inline behavior specification replaced with delegation instruction
3. The dromena does pre-flight only (read-only checks), then delegates
4. Display formatting is simple template over Moirai's JSON response

### 4.3 Per-Dromena Change Specifications

#### /park (`mena/session/park/INDEX.dro.md`)

| Aspect | Current | Target |
|--------|---------|--------|
| `allowed-tools` | `Bash, Read, Write` | `Bash, Read, Task` |
| Execution | Inline: calls `ari session park` directly | Delegation: `Task(moirai, "park_session reason=...")` |
| Pre-flight | Verify active session, check not parked | Same (unchanged) |
| Display | Inline summary generation | Template over Moirai JSON response |
| Line count | ~77 lines | ~50 lines |

#### /wrap (`mena/session/wrap/INDEX.dro.md`)

| Aspect | Current | Target |
|--------|---------|--------|
| `allowed-tools` | `Bash, Read, Write, Task, Glob` | `Bash, Read, Task, Glob` |
| Execution | Inline: calls `ari session wrap`, quality gates, worktree cleanup | Delegation: `Task(moirai, "wrap_session")` or `Task(moirai, "wrap_session --emergency")` |
| Pre-flight | Verify active session, check uncommitted changes | Same (unchanged) |
| Quality gates | Inline check | Moirai handles via Atropos skill (quality gate is in wrap_session spec) |
| Worktree cleanup | Inline post-wrap | Remains in dromena (Moirai does not manage worktrees) |
| Display | Inline completion summary | Template over Moirai JSON response + worktree prompt |
| `Write` removed | Yes -- Moirai writes SESSION_CONTEXT via lock bypass | |

**Special note**: Worktree cleanup logic (lines 75-98 of current wrap dromena) stays in the dromena. Moirai handles session state transitions; the dromena handles git worktree lifecycle. This is correct separation of concerns.

#### /handoff (`mena/session/handoff/INDEX.dro.md`)

| Aspect | Current | Target |
|--------|---------|--------|
| `allowed-tools` | `Bash, Read, Write, Task` | `Bash, Read, Task` |
| Execution | Already partially delegates to `Task(moirai, ...)` for state mutation (line 43) | Full delegation: both mutation AND target agent invocation delegated |
| Pre-flight | Verify active session, validate target agent exists | Same (unchanged) |
| Target agent invocation | Dromena invokes target agent via Task (step 4) | Dromena still invokes target agent -- Moirai only records the handoff, it does not invoke agents |
| `Write` removed | Yes | |

**Key insight**: /handoff is the CLOSEST to the target pattern already. It already uses `Task(moirai, "handoff from <FROM> to <TO>")`. The refactor removes the `Write` tool permission and ensures the dromena never writes `*_CONTEXT.md` directly. The step-4 target agent invocation remains in the dromena because Moirai (as a subagent) cannot invoke Task.

#### /continue (`mena/session/continue/INDEX.dro.md`)

| Aspect | Current | Target |
|--------|---------|--------|
| `allowed-tools` | `Bash, Read, Write, Task` | `Bash, Read, Task` |
| Execution | Inline: reads session, validates state, updates context, emits event | Delegation: `Task(moirai, "resume_session")` |
| Pre-flight | Verify current session exists, verify PARKED status | Same (unchanged) |
| Display | JSON-formatted resumption summary | Template over Moirai JSON response |
| `Write` removed | Yes | |

#### /start (`mena/session/start/INDEX.dro.md`)

| Aspect | Current | Target |
|--------|---------|--------|
| `allowed-tools` | `Bash, Read, Task` | `Bash, Read, Task` (unchanged) |
| Execution | Calls `ari session create`, then invokes entry agent via Task | Minimal change: session creation stays with `ari session create` (Clotho's `create_sprint` is for sprints within sessions, not sessions themselves) |
| Change scope | Low -- /start already delegates correctly | Add Moirai delegation for the sprint creation step IF the rite includes sprint tracking |

**Key insight**: /start does NOT delegate to Moirai for session creation. Session creation is an `ari` CLI operation, not a Moirai operation. The TDD's Clotho skill covers `create_sprint` and `start_sprint` (sprint-level, not session-level). /start's refactor is minimal: ensure it does not directly write `*_CONTEXT.md` (it currently does not -- it calls `ari session create` which does the writing). No `allowed-tools` change needed.

### 4.4 Moirai Agent Prerequisite

The dromena refactor assumes a Moirai agent definition exists at `.claude/agents/moirai.md`. This agent is NOT part of the current rite's agent set (the ecosystem rite has orchestrator, ecosystem-analyst, etc.). The Moirai agent must be created as a **user-level agent** or added to the rite.

**Decision**: Create the Moirai agent prompt as part of this sprint. The agent definition is a prerequisite for the dromena refactor to function. Without it, `Task(moirai, ...)` will fail.

**Moirai agent location**: `.claude/agents/moirai.md` (direct creation, not via mena projection, because agents are not part of the mena system -- they live in `.claude/agents/` directly).

**Moirai agent key properties**:
- `model`: sonnet (most operations are straightforward; opus is overkill)
- `allowed-tools`: Bash, Read, Write, Edit
- `Write` is required because Moirai IS the authorized writer of `*_CONTEXT.md`
- Agent prompt includes: operation parsing rules, skill loading protocol (Read `.claude/skills/session/moirai/INDEX.md` first), JSON response format, CLI delegation protocol, lock acquisition/release protocol
- Agent prompt does NOT include: individual operation specs (those come from Fate skills via progressive disclosure)

---

## 5. Backward Compatibility Assessment

### 5.1 Classification: COMPATIBLE

All changes are additive or modify internal behavior without changing external interfaces.

| Change | Impact | Compatibility |
|--------|--------|---------------|
| New mena files (4 legomena) | New files in `mena/session/moirai/` | COMPATIBLE: additive |
| New projected skills (4 files) | New files in `.claude/skills/session/moirai/` | COMPATIBLE: additive |
| Writeguard lock mechanism | Replaces env var bypass | COMPATIBLE: env var was never used in production |
| Dromena tool permission changes | `Write` removed, `Task` added | COMPATIBLE: changes what the dromena CAN do, not the user interface |
| New Moirai agent | New file in `.claude/agents/` | COMPATIBLE: additive |
| New `ari session lock/unlock` commands | New CLI commands | COMPATIBLE: additive |

### 5.2 Migration Path

No migration required. Existing satellites are unaffected because:
1. The new skill files are only loaded by the Moirai agent (internal to knossos)
2. The writeguard behavior change only affects Moirai-mediated writes
3. Dromena behavioral changes are transparent to users (same commands, same results)

### 5.3 Rollback Plan

If issues are discovered:
1. Revert dromena files to pre-refactor state (restore `Write` tool, inline behavior)
2. Revert writeguard to env var bypass (restore `BypassEnvVar` check)
3. Fate skill files can remain (they are inert without the Moirai agent reading them)
4. Remove Moirai agent file

---

## 6. File-Level Change Specification

### 6.1 New Files

| File | Type | Description |
|------|------|-------------|
| `mena/session/moirai/INDEX.lego.md` | Legomena source | Routing table |
| `mena/session/moirai/clotho.md` | Companion file | Creation ops |
| `mena/session/moirai/lachesis.md` | Companion file | Measurement ops |
| `mena/session/moirai/atropos.md` | Companion file | Termination ops |
| `.claude/agents/moirai.md` | Agent definition | Moirai unified agent |
| `internal/cmd/session/lock.go` | Go source | `ari session lock/unlock` commands |
| `internal/cmd/session/lock_test.go` | Go test | Lock command tests |

### 6.2 Modified Files

| File | Changes |
|------|---------|
| `internal/cmd/hook/writeguard.go` | Remove env var bypass; add `isMoiraiLockHeld()` function; modify `runWriteguardCore()` to check lock before blocking |
| `internal/cmd/hook/hook_test.go` | Add tests for lock-based bypass in `TestIntegration_WriteguardHook_Chain` |
| `mena/session/park/INDEX.dro.md` | Replace inline execution with Task(moirai) delegation; remove Write from allowed-tools |
| `mena/session/wrap/INDEX.dro.md` | Replace inline execution with Task(moirai) delegation; remove Write from allowed-tools; preserve worktree logic |
| `mena/session/handoff/INDEX.dro.md` | Remove Write from allowed-tools; ensure full Moirai delegation pattern |
| `mena/session/continue/INDEX.dro.md` | Replace inline execution with Task(moirai) delegation; remove Write from allowed-tools |
| `mena/session/start/INDEX.dro.md` | Minimal changes: verify no direct context writes exist |

### 6.3 Files NOT Modified

| File | Reason |
|------|--------|
| `hooks/hooks.yaml` | No new hooks needed; writeguard command unchanged |
| `.claude/settings.local.json` | Generated by materialization; no manual edits |
| `internal/hook/env.go` | No new env vars needed (lock is file-based, not env-based) |
| `internal/hook/output.go` | Output format unchanged |
| `internal/materialize/project_mena.go` | Mena projection already handles companion files correctly |

---

## 7. Integration Test Matrix

### 7.1 Fate Skill Tests

| Test ID | Satellite Type | Test | Expected Outcome |
|---------|---------------|------|------------------|
| FATE-01 | knossos (self) | Materialize with moirai legomena | 4 files appear in `.claude/skills/session/moirai/` |
| FATE-02 | knossos (self) | INDEX.md contains routing table | All 13 operations have fate assignment |
| FATE-03 | knossos (self) | clotho.md contains 2 operations | create_sprint, start_sprint specs present |
| FATE-04 | knossos (self) | lachesis.md contains 8 operations | All 8 measurement ops present |
| FATE-05 | knossos (self) | atropos.md contains 3 operations | wrap_session, generate_sails, delete_sprint present |
| FATE-06 | minimal satellite | Materialize without moirai | No errors; moirai skills not projected (moirai is knossos-internal) |

### 7.2 Write Guard Tests

| Test ID | Scenario | Expected Outcome |
|---------|----------|------------------|
| WG-01 | Write SESSION_CONTEXT.md with no lock | DENY |
| WG-02 | Write SESSION_CONTEXT.md with valid moirai lock | ALLOW |
| WG-03 | Write SESSION_CONTEXT.md with stale moirai lock (>300s) | DENY |
| WG-04 | Write SESSION_CONTEXT.md with lock from non-moirai agent | DENY |
| WG-05 | Write normal file with moirai lock held | ALLOW (writeguard only checks protected files) |
| WG-06 | Write SPRINT_CONTEXT.md with valid moirai lock | ALLOW |
| WG-07 | MOIRAI_BYPASS env var set (backward compat) | ALLOW removed -- no longer honored |
| WG-08 | `ari session lock --agent moirai` creates lock file | Lock file exists with correct JSON |
| WG-09 | `ari session unlock --agent moirai` removes lock file | Lock file removed |
| WG-10 | `ari session lock --agent moirai` with existing lock | Error: lock already held |

### 7.3 Dromena Refactor Tests

| Test ID | Dromena | Test | Expected Outcome |
|---------|---------|------|------------------|
| DRO-01 | /park | Frontmatter has `Task` in allowed-tools | `allowed-tools: Bash, Read, Task` |
| DRO-02 | /park | Frontmatter does NOT have `Write` | Write absent |
| DRO-03 | /park | Body contains `Task(moirai` delegation | Delegation instruction present |
| DRO-04 | /wrap | Frontmatter has `Task` in allowed-tools, no `Write` | Correct |
| DRO-05 | /wrap | Worktree cleanup logic preserved | Worktree section exists |
| DRO-06 | /handoff | Frontmatter has no `Write` | Correct |
| DRO-07 | /continue | Frontmatter has `Task`, no `Write` | Correct |
| DRO-08 | /start | No direct context writes | No Write tool usage for *_CONTEXT.md |

### 7.4 End-to-End Integration Tests

| Test ID | Scenario | Steps | Expected Outcome |
|---------|----------|-------|------------------|
| E2E-01 | Park via delegation | 1. Active session exists 2. Invoke /park 3. Dromena delegates to Task(moirai) 4. Moirai acquires lock 5. Moirai writes SESSION_CONTEXT 6. Moirai releases lock | Session state: PARKED |
| E2E-02 | Resume via delegation | 1. Parked session exists 2. Invoke /continue 3. Dromena delegates to Task(moirai) | Session state: ACTIVE |
| E2E-03 | Wrap with quality gate | 1. Active session 2. Invoke /wrap 3. Moirai invokes ari session wrap 4. Sails computed | Session state: ARCHIVED, sails generated |
| E2E-04 | Writeguard blocks main thread | 1. Active session 2. Main thread tries Write(SESSION_CONTEXT.md) | DENY -- must use Moirai |

---

## 8. Implementation Order

The integration-engineer should implement in this order to maintain a buildable state at each step:

1. **Create fate skill files** (WS-1) -- pure file creation, no code changes, immediately testable
2. **Implement `ari session lock/unlock`** (WS-2 prerequisite) -- new Go code, testable independently
3. **Modify writeguard** (WS-2) -- depends on lock command existing
4. **Create Moirai agent prompt** (WS-3 prerequisite) -- file creation, depends on skill files existing
5. **Refactor dromena** (WS-3) -- depends on Moirai agent and writeguard fix

Each step produces a working system. If the sprint is cut short after step 3, the fate skills exist but dromena still work in their current inline mode.

---

## 9. Open Items (None)

All design decisions are resolved. No TBD flags remain.

---

## 10. Artifact Attestation

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This Context Design | `/Users/tomtenuta/Code/knossos/docs/design/CONTEXT-DESIGN-fate-skills-w4.md` | Created |
| TDD (input) | `/Users/tomtenuta/Code/knossos/docs/design/TDD-fate-skills.md` | Read |
| writeguard.go (input) | `/Users/tomtenuta/Code/knossos/internal/cmd/hook/writeguard.go` | Read |
| hook/env.go (input) | `/Users/tomtenuta/Code/knossos/internal/hook/env.go` | Read |
| hook/output.go (input) | `/Users/tomtenuta/Code/knossos/internal/hook/output.go` | Read |
| park dromena (input) | `/Users/tomtenuta/Code/knossos/mena/session/park/INDEX.dro.md` | Read |
| wrap dromena (input) | `/Users/tomtenuta/Code/knossos/mena/session/wrap/INDEX.dro.md` | Read |
| handoff dromena (input) | `/Users/tomtenuta/Code/knossos/mena/session/handoff/INDEX.dro.md` | Read |
| continue dromena (input) | `/Users/tomtenuta/Code/knossos/mena/session/continue/INDEX.dro.md` | Read |
| start dromena (input) | `/Users/tomtenuta/Code/knossos/mena/session/start/INDEX.dro.md` | Read |
| project_mena.go (input) | `/Users/tomtenuta/Code/knossos/internal/materialize/project_mena.go` | Read |
| hooks.yaml (input) | `/Users/tomtenuta/Code/knossos/hooks/hooks.yaml` | Read |
| settings.local.json (input) | `/Users/tomtenuta/Code/knossos/.claude/settings.local.json` | Read |

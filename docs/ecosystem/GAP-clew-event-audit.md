# Gap Analysis: Clew Event Type Audit -- Dead vs Valuable

**Date**: 2026-02-08
**Analyst**: ecosystem-analyst
**Sprint**: 001, Task 025
**Scope**: `internal/hook/clewcontract/event.go` (16 EventType constants)

---

## Clew Event Type Audit

The clewcontract package defines 16 `EventType` constants. Each was searched for by:
1. Constant name (e.g., `EventTypeToolCall`)
2. Constructor name (e.g., `NewToolCallEvent`)
3. String value (e.g., `"tool_call"`)

Callers are counted **outside** of definition files (`event.go`) and test files (`*_test.go`).

| # | EventType | String Value | Non-test/non-def Callers | Caller Locations | Classification | Notes |
|---|-----------|-------------|-------------------------|------------------|---------------|-------|
| 1 | `EventTypeToolCall` | `"tool_call"` | 2 | `record.go:105` (BuildEventFromToolInput), `tribute/extractor.go:246` (ExtractMetrics) | **active** | Primary clew event. Emitted by PostToolUse hook via `RecordToolEvent`. Consumed by tribute metrics. |
| 2 | `EventTypeFileChange` | `"file_change"` | 1 | `tribute/extractor.go:248` (ExtractMetrics) | **dead/wire** | Constructor `NewFileChangeEvent` has ZERO callers outside tests. The tribute extractor reads it but nothing emits it. |
| 3 | `EventTypeCommand` | `"command"` | 0 | (none) | **dead/cut** | Constructor `NewCommandEvent` has ZERO callers outside tests. Semantically overlaps with `tool_call` + Bash tool. Redundant. |
| 4 | `EventTypeDecision` | `"decision"` | 3 | `record.go:130` (RecordStamp/Stamp.ToEvent), `clew.go:146` (orchestrator throughline), `tribute/extractor.go:114` (ExtractDecisions) | **active** | Emitted via `/stamp` command and auto-captured from orchestrator Task results. Consumed by tribute. |
| 5 | `EventTypeContextSwitch` | `"context_switch"` | 1 | `triggers.go:185` (checkContextSwitch trigger matcher) | **dead/wire** | Constructor `NewContextSwitchEvent` has ZERO callers outside tests. The trigger system checks for it but nothing emits it. |
| 6 | `EventTypeSailsGenerated` | `"sails_generated"` | 1 | `session/wrap.go:153` (NewSailsGeneratedEvent) | **active** | Emitted during session wrap when WHITE_SAILS is generated. |
| 7 | `EventTypeTaskStart` | `"task_start"` | 4 | `hook/subagent.go:108`, `handoff/execute.go:134`, `sails/contract.go:176`, `handoff/status.go:197` | **active** | Emitted by SubagentStart hook and handoff execute. Consumed by sails contract validation and handoff status. |
| 8 | `EventTypeTaskEnd` | `"task_end"` | 4 | `hook/subagent.go:156`, `handoff/prepare.go:202`, `sails/contract.go:199`, `handoff/status.go:256` | **active** | Emitted by SubagentStop hook and handoff prepare. Consumed by sails contract validation and handoff status. |
| 9 | `EventTypeSessionStart` | `"session_start"` | 0 | (none) | **dead/wire** | Constructor `NewSessionStartEvent` has ZERO callers outside tests/definition. Session creation uses `session.EventEmitter.EmitCreated` which writes `SESSION_CREATED` (different event system). This clewcontract type is orphaned. |
| 10 | `EventTypeSessionEnd` | `"session_end"` | 2 | `session/wrap.go:205,207` (NewSessionEndEvent/WithBudget), `session/park.go:127` (NewSessionEndEvent) | **active** | Emitted during session wrap and park. |
| 11 | `EventTypeArtifactCreated` | `"artifact_created"` | 1 (reader only) | `tribute/extractor.go:75` (ExtractArtifacts reads it) | **dead/wire** | Constructor `NewArtifactCreatedEvent` has ZERO callers outside tests. The tribute extractor reads this event type but nothing emits it. |
| 12 | `EventTypeError` | `"error"` | 0 | (none) | **dead/wire** | Constructor `NewErrorEvent` has ZERO callers outside tests/definition. Could be emitted from hook error paths. |
| 13 | `EventTypeHandoffPrepared` | `"handoff_prepared"` | 2 | `handoff/prepare.go:215`, `sails/contract.go:116` | **active** | Emitted by handoff prepare. Consumed by sails contract validation. |
| 14 | `EventTypeHandoffExecuted` | `"handoff_executed"` | 2 | `handoff/execute.go:148`, `sails/contract.go:125` | **active** | Emitted by handoff execute. Consumed by sails contract validation and tribute extractor. |
| 15 | `EventTypeSessionFrayed` | `"session_frayed"` | 1 | `session/fray.go:197` (NewSessionFrayedEvent) | **active** | Emitted during session fray. |
| 16 | `EventTypeStrandResolved` | `"strand_resolved"` | 1 | `session/wrap.go:219` (NewStrandResolvedEvent) | **active** | Emitted during session wrap when frayed session resolves. |

---

## Summary Counts

- **Active** (has real emitters AND/OR consumers): 10
  - `tool_call`, `decision`, `sails_generated`, `task_start`, `task_end`, `session_end`, `handoff_prepared`, `handoff_executed`, `session_frayed`, `strand_resolved`

- **Dead -- Wire candidates** (should be emitted but are not): 5
  - `file_change`, `context_switch`, `session_start`, `artifact_created`, `error`

- **Dead -- Cut candidate** (redundant, no use case): 1
  - `command`

---

## Writer Audit

| Writer | Non-test Callers | Caller Locations | Status |
|--------|-----------------|------------------|--------|
| `EventWriter` (sync) | 9 | `record.go:20,133` (RecordToolEvent, RecordStamp), `session/wrap.go:128,196,218`, `session/fray.go:196`, `session/park.go:124`, `handoff/prepare.go:189`, `handoff/execute.go:126`, `hook/subagent.go:118,166` | **active** -- primary writer, used everywhere |
| `BufferedEventWriter` | 0 | (none outside definition/tests) | **unused** -- no production callers |

### BufferedEventWriter Details
- Flush interval: `DefaultFlushInterval = 5 * time.Second`
- Has comprehensive tests (14 test functions in `writer_test.go`)
- Thread-safe, re-queues on failure, bounded loss window
- Was designed for high-throughput scenarios but never adopted
- All callers use the synchronous `EventWriter` instead

---

## Wire Recommendations (5 types)

These types exist in the contract, have consumers expecting them (tribute, triggers), but nothing emits them.

### 1. `EventTypeFileChange` -- Wire to PostToolUse (Edit/Write)

**Current state**: `tribute/extractor.go:248` reads `file_change` events to count `FilesModified` but the only event emitted by the PostToolUse clew hook is `tool_call`. The file change data IS captured (tool=Edit/Write with path), but as `tool_call` not `file_change`.

**Wire target**: The clew PostToolUse hook (`internal/cmd/hook/clew.go`) could emit a `file_change` event alongside or instead of `tool_call` when the tool is Edit or Write. Alternatively, the tribute extractor could count `tool_call` events with tool=Edit/Write as file changes (this would be a consumer-side fix rather than an emitter fix).

**CC hook**: `PostToolUse` (already handled)

### 2. `EventTypeContextSwitch` -- Wire to heuristic detection

**Current state**: `triggers.go:185` checks for `context_switch` events to fire the stamp prompt, but nothing emits context_switch events. This trigger path is dead.

**Wire target**: Could be emitted by the PostToolUse clew hook when a significant path change is detected (e.g., switching between directories or modules). This requires heuristic logic comparing current event path to recent event paths.

**CC hook**: `PostToolUse` (already handled, add heuristic)

### 3. `EventTypeSessionStart` -- Wire to SessionStart hook

**Current state**: `NewSessionStartEvent` has zero callers. Session creation uses the `session.EventEmitter` which writes `SESSION_CREATED` to events.jsonl, but this is a different event format/system (`session.Event` not `clewcontract.Event`). The clewcontract `session_start` type is completely orphaned.

**Wire target**: The `SessionStart` CC hook (`internal/hook/env.go:37`) exists but has no corresponding clew handler. A `session-start` hook subcommand could emit `NewSessionStartEvent`.

**CC hook**: `SessionStart` (CC event exists, no ari handler)

### 4. `EventTypeArtifactCreated` -- Wire to PostToolUse (Write to artifact paths)

**Current state**: `tribute/extractor.go:75` reads `artifact_created` events to populate the tribute artifact list, but nothing emits them. Artifacts are created by Write tool calls to known paths (e.g., `docs/requirements/PRD-*.md`, `docs/design/TDD-*.md`).

**Wire target**: Could be emitted from the PostToolUse clew hook when a Write event targets a recognized artifact path pattern. Alternatively, a dedicated `/artifact register` command could emit it (there is already `internal/cmd/artifact/register.go`).

**CC hook**: `PostToolUse` (already handled, add path matching) or manual via command

### 5. `EventTypeError` -- Wire to error paths in hook handlers

**Current state**: `NewErrorEvent` has zero callers. Multiple hook handlers log errors via `printer.VerboseLog("error", ...)` but never emit clew error events.

**Wire target**: Could be emitted from hook error paths (validation failures, writeguard blocks, sync failures). The structured error fields (error_code, recoverable, suggested_action) map well to existing error handling patterns.

**CC hook**: Multiple (PostToolUse, PreToolUse, Stop -- wherever errors occur)

---

## Cut Recommendation (1 type)

### `EventTypeCommand` -- Redundant with tool_call+Bash

**Rationale**: This type models "shell command execution" with command string, exit code, and duration. However, the PostToolUse clew hook already captures Bash tool calls as `tool_call` events with `meta.command`, `meta.exit_code` metadata. The `command` type duplicates this without adding semantic value. No code emits it, no code reads it.

**Alternative**: If post-execution command metrics are needed (e.g., tracking build/test durations), they can be extracted from existing `tool_call` events where `tool=Bash`.

---

## BufferedEventWriter Recommendation

**Status**: Unused in production. Zero callers outside definition and tests.

**Options**:
1. **Cut**: Remove the type and its 14 tests. All callers use synchronous `EventWriter`. The buffered writer adds code surface area and a background goroutine pattern that nobody exercises.
2. **Preserve**: Keep it if high-throughput scenarios are anticipated (e.g., streaming tool events from long sessions). The code is well-tested and the 5-second flush interval provides a reasonable loss window.

**Recommendation**: Cut. If needed later, it can be restored from git history. The synchronous writer with open/close-per-write is adequate for current throughput (one event per tool use, ~100ms budget).

---

## Dual Event System Observation

The codebase has TWO event systems writing to the same `events.jsonl` file:

1. **`clewcontract.Event`** (this audit) -- `{"ts": ..., "type": "tool_call", "tool": ..., "summary": ...}`
2. **`session.Event`** (internal/session/events.go) -- `{"timestamp": ..., "event": "SESSION_CREATED", "from": ..., "to": ...}`

These have different schemas, different field names (`ts` vs `timestamp`, `type` vs `event`), and are interleaved in the same JSONL file. This is not a bug per se -- readers like `tribute/extractor.go` handle both formats -- but it explains why `EventTypeSessionStart` is dead: session lifecycle events use the `session.Event` system, not clewcontract.

---

## Complexity: PATCH (for cuts) / MODULE (for wiring)

- Cutting `EventTypeCommand` and `BufferedEventWriter`: **PATCH** -- remove dead code, no behavioral change.
- Wiring the 5 dead types to real emitters: **MODULE** -- requires adding emit logic to hook handlers, each type is a separate small change but collectively they touch `internal/cmd/hook/` and potentially `internal/cmd/session/`.

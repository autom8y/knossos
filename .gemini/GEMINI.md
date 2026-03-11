# Knossos Multi-Channel Sprint

You are adding `.gemini/` as a second materialization target to the Knossos platform alongside `.claude/`. Phases 1-3 are complete and merged. Only **Phase 4 (Session Federation)** remains.

## Project Identity

- **Binary**: `ari` (Ariadne) -- Cobra CLI, Go 1.23+, `CGO_ENABLED=0`
- **Module**: `github.com/autom8y/knossos`
- **Build**: `CGO_ENABLED=0 go build ./cmd/ari`
- **Test**: `CGO_ENABLED=0 go test ./...`
- **What it does**: Reads declarative rite bundles from `.knossos/rites/` and projects them into AI assistant config directories (`.claude/` and `.gemini/`)

## Hard Constraints (No Hooks Enforcing These -- You Must Self-Enforce)

1. **NEVER rename or remove `.claude/`** -- CC's file watcher crashes on directory renames (SCAR-002).
2. **NEVER use `os.RemoveAll` on projected directories** (`agents/`, `commands/`, `skills/`). Use provenance-tracked selective removal (SCAR-005).
3. **ALL writes to `.claude/` MUST use `fileutil.WriteIfChanged()`**. Writes to `.gemini/` SHOULD use `WriteIfChanged` for consistency.
4. **NEVER discard provenance errors** with `_`. Parse/validation errors abort the pipeline (SCAR-004).
5. **Error handling convention**:
   ```go
   // At CLI boundary (cmd/ packages):
   errors.New(errors.CodeGeneralError, "descriptive message")
   errors.Wrap(errors.CodeGeneralError, "context", err)
   // Inside domain packages:
   fmt.Errorf("context: %w", err)
   ```
6. **Testing convention**: `t.Parallel()` on every test. Table-driven preferred. `CGO_ENABLED=0`.
7. **Dual manifest invariant** (TENSION-001): Two `RiteManifest` structs exist -- `materialize.RiteManifest` and `rite.RiteManifest`. Any new field must be added to BOTH.
8. **Triple threading** (TENSION-002): New options must go through `SyncOptions` -> `Options` -> `syncRiteScope()` mapping. Missing one causes silent divergence.
9. **claudeDirOverride cache/restore**: When channel switching within `MaterializeWithOptions`, the original `claudeDirOverride` must be cached and restored after the call. P2 learned this the hard way -- if you set `claudeDirOverride` for gemini and don't restore it, subsequent claude-channel operations write to the wrong directory.
10. **KNOSSOS_CHANNEL env var**: Hooks detect their channel via `KNOSSOS_CHANNEL` env var (set during materialization). Default is absent (= claude). `gemini` triggers `GeminiAdapter`.

## Completed Phases Summary

### Phase 1: Channel Abstraction (COMPLETE -- merged to main)
- `internal/paths/channel.go` -- `TargetChannel` interface, `ClaudeChannel`/`GeminiChannel` structs, `ChannelByName()`, `Resolver.ChannelDir()`
- `--channel` flag added to root command (`internal/cmd/root/root.go`)
- Channel threaded through `SyncOptions` -> `Options` -> `syncRiteScope()` (TENSION-002 satisfied)
- `claudeDirOverride` used to route gemini writes to `.gemini/`

### Phase 2: Transpiling Materializer (COMPLETE -- merged to main)
- `internal/materialize/compiler/compiler.go` -- `ChannelCompiler` interface
- `internal/materialize/compiler/claude.go` -- identity compiler (byte-identical to pre-refactor output)
- `internal/materialize/compiler/gemini.go` -- TOML transpiler for commands, skill restructurer
- Compiler injected into mena engine write path
- Inscription pipeline parameterized for context filename (`CLAUDE.md` vs `GEMINI.md`)
- `pelletier/go-toml/v2` dependency added

### Phase 3: Hook Adapter (COMPLETE -- merged to main)
- `internal/hook/adapter.go` -- `LifecycleAdapter` interface: `ParsePayload()`, `FormatResponse()`, `ChannelName()`
- `internal/hook/adapter_gemini.go` -- `GeminiAdapter` with `GeminiPayload` struct (different JSON field names from CC's `StdinPayload`)
- `internal/hook/env.go` -- dynamic channel detection via `KNOSSOS_CHANNEL` env var
- Hook config injection: gemini-channel hooks get `KNOSSOS_CHANNEL=gemini` in their environment

### Fixes Applied (from QA review across P1-P3)
- C-1: `--channel` flag wired through `ari sync` CLI
- C-2: `syncRiteScopeMinimal` now forwards Channel
- C-3: `--channel` value validated at CLI boundary
- W-3: `claudeDirOverride` state properly cached and restored
- W-2: `CompileAgent` byte identity constraint fixed for Claude adapter

**Total: 29 files changed, +1011/-146 lines.**

## Architecture Overview

```
cmd/ari/main.go                          -- entry, wiring only
internal/cmd/root/root.go                -- root cobra command, global flags (incl --channel)
internal/cmd/sync/                       -- ari sync command
internal/paths/
  paths.go                               -- Resolver: project root, directory paths
  channel.go                             -- TargetChannel interface, ClaudeChannel, GeminiChannel
internal/materialize/
  materialize.go                         -- Materializer, Options (incl Channel), Sync()
  materialize_mena.go                    -- mena -> commands/ + skills/ (compiler-aware)
  materialize_agents.go                  -- agent file projection
  materialize_claudemd.go               -- context file inscription (parameterized filename)
  sync_types.go                          -- SyncOptions (incl Channel), SyncResult, SyncScope
  compiler/
    compiler.go                          -- ChannelCompiler interface
    claude.go                            -- identity compiler
    gemini.go                            -- TOML transpiler
internal/inscription/                    -- context file region system (parameterized)
internal/hook/
  env.go                                 -- Env, StdinPayload (CC), channel detection
  adapter.go                             -- LifecycleAdapter interface
  adapter_gemini.go                      -- GeminiAdapter, GeminiPayload
  clewcontract/
    event.go                             -- Event (v2), EventType constants, constructor functions
    typed_event.go                       -- TypedEvent (v3), EventSource, newTypedEvent()
    typed_data.go                        -- Per-type data structs (SessionCreatedData, etc.)
    typed_constructors.go                -- NewTyped*Event() constructors
    writer.go                            -- EventWriter, BufferedEventWriter (thread-safe JSONL)
internal/provenance/
  provenance.go                          -- ProvenanceManifest, ProvenanceEntry, OwnerType, ScopeType
  manifest.go                            -- Load, Save, LoadOrBootstrap, validation
  collector.go                           -- Collector interface
  merge.go                               -- Manifest merging
  divergence.go                          -- Divergence detection
internal/session/
  status.go                              -- Status FSM: NONE, ACTIVE, PARKED, ARCHIVED
internal/fileutil/                       -- WriteIfChanged, AtomicWriteFile
```

---

## Phase 4: Session Federation

**Size**: M (~200 lines new code, ~100 lines modified)
**Dependencies**: Phases 1-3 (all merged)
**Branch**: `feat/multi-channel-p4`

Phase 4 makes the session/event/provenance systems channel-aware so that dual-projection (`ari sync --channel=all`) can track which channel owns which file, and events carry channel provenance.

### 4.1 Add channel field to clewcontract events

Both v2 `Event` and v3 `TypedEvent` need channel awareness.

**v2 Event** (`internal/hook/clewcontract/event.go`):

The `Event` struct has a `Meta map[string]any` field. For v2 events, channel is added to Meta by the constructors -- NOT as a top-level field. This preserves backward compatibility (existing v2 consumers ignore unknown Meta keys).

Modify the relevant constructor functions to accept an optional channel parameter:
- `NewSessionStartEvent` -- add channel to meta
- `NewSessionEndEventWithBudget` -- add channel to meta
- `NewToolCallEvent` -- add channel to meta

**Design choice**: Rather than modifying every constructor signature (breaking change), add a helper:

```go
// WithChannel returns a copy of the event with channel metadata added.
// If channel is empty or "claude", no metadata is added (backward compat).
func (e Event) WithChannel(channel string) Event {
    if channel == "" || channel == "claude" {
        return e
    }
    if e.Meta == nil {
        e.Meta = make(map[string]any)
    }
    e.Meta["channel"] = channel
    return e
}
```

**v3 TypedEvent** (`internal/hook/clewcontract/typed_event.go`):

Add `Channel` as a top-level field on `TypedEvent`:

```go
type TypedEvent struct {
    Ts      string          `json:"ts"`
    Type    EventType       `json:"type"`
    Source  EventSource     `json:"source"`
    Channel string          `json:"channel,omitempty"` // NEW: "claude" or "gemini"; omitted = claude
    Data    json.RawMessage `json:"data"`
}
```

Modify `newTypedEvent` to accept channel:

```go
func newTypedEvent(eventType EventType, source EventSource, channel string, data any) TypedEvent {
    // ... existing marshaling logic ...
    te := TypedEvent{
        Ts:     typedEventTimestamp(),
        Type:   eventType,
        Source: source,
        Data:   raw,
    }
    if channel != "" && channel != "claude" {
        te.Channel = channel
    }
    return te
}
```

**WARNING**: `newTypedEvent` is called by every `NewTyped*Event` constructor in `typed_constructors.go`. When you change its signature, you MUST update every call site. There are ~20 constructors. Use the compiler to find them all -- do not grep manually.

### 4.2 Add channel field to provenance manifest

**File: `internal/provenance/provenance.go`**

Add `Channel` to `ProvenanceEntry`:

```go
type ProvenanceEntry struct {
    Owner      OwnerType `yaml:"owner"`
    Scope      ScopeType `yaml:"scope"`
    Channel    string    `yaml:"channel,omitempty"` // NEW: "claude" or "gemini"; empty = claude
    SourcePath string    `yaml:"source_path,omitempty"`
    SourceType string    `yaml:"source_type,omitempty"`
    Checksum   string    `yaml:"checksum"`
    LastSynced time.Time `yaml:"last_synced"`
}
```

Update `NewKnossosEntry` to accept channel:

```go
func NewKnossosEntry(scope ScopeType, channel, sourcePath, sourceType, checksum string) *ProvenanceEntry {
    entry := &ProvenanceEntry{
        Owner:      OwnerKnossos,
        Scope:      scope,
        SourcePath: sourcePath,
        SourceType: sourceType,
        Checksum:   checksum,
        LastSynced: time.Now().UTC(),
    }
    if channel != "" && channel != "claude" {
        entry.Channel = channel
    }
    return entry
}
```

**WARNING -- SCAR-004**: `NewKnossosEntry` is called from many sites in `internal/materialize/`. When you change its signature, you MUST update every call site. Use the compiler to find them all. Do not leave any call site with the old signature -- it will silently compile with wrong argument mapping.

Update `structurallyEqual` to compare Channel:

```go
if entryA.Channel != entryB.Channel {
    return false
}
```

Update `NewUserEntry` and `NewUntrackedEntry` similarly if needed (they may not need channel since user-created files are channel-agnostic, but evaluate during implementation).

### 4.3 Map Gemini lifecycle to session FSM

The session FSM has 4 states: `NONE`, `ACTIVE`, `PARKED`, `ARCHIVED`.

Gemini CLI emits its own lifecycle events with different names. Create a mapping layer.

**File: `internal/session/channel.go`** (NEW)

```go
package session

// ChannelLifecycleMap maps channel-specific lifecycle event names to Knossos session FSM states.
type ChannelLifecycleMap struct {
    StartEvents   []string // events that map to ACTIVE
    EndEvents     []string // events that map to ARCHIVED
    SuspendEvents []string // events that map to PARKED
    ResumeEvents  []string // events that map to ACTIVE (from PARKED)
}

// GeminiLifecycleMap returns the lifecycle event mapping for Gemini CLI.
func GeminiLifecycleMap() ChannelLifecycleMap {
    return ChannelLifecycleMap{
        StartEvents:   []string{"session_start", "conversation_start"},
        EndEvents:     []string{"session_end", "conversation_end"},
        SuspendEvents: []string{"session_suspend"},
        ResumeEvents:  []string{"session_resume"},
    }
}

// ClaudeLifecycleMap returns the lifecycle event mapping for Claude Code.
func ClaudeLifecycleMap() ChannelLifecycleMap {
    return ChannelLifecycleMap{
        StartEvents:   []string{"SessionStart"},
        EndEvents:     []string{"SessionEnd"},
        SuspendEvents: []string{}, // CC has no native suspend
        ResumeEvents:  []string{}, // CC has no native resume
    }
}

// MapToFSMTransition maps a channel-specific event name to a session FSM status.
// Returns (targetStatus, true) if the event maps to a known transition.
// Returns ("", false) if the event has no FSM mapping.
func (m ChannelLifecycleMap) MapToFSMTransition(eventName string) (Status, bool) {
    for _, e := range m.StartEvents {
        if e == eventName { return StatusActive, true }
    }
    for _, e := range m.EndEvents {
        if e == eventName { return StatusArchived, true }
    }
    for _, e := range m.SuspendEvents {
        if e == eventName { return StatusParked, true }
    }
    for _, e := range m.ResumeEvents {
        if e == eventName { return StatusActive, true }
    }
    return "", false
}
```

### 4.4 Wire channel into event emission call sites

The clew hook (`internal/cmd/hook/clew.go`) emits events. It needs to detect the active channel (via `KNOSSOS_CHANNEL` env var, already implemented in P3) and pass it through.

Locate all call sites that create events and pass the channel through. The main sites are:
- `internal/cmd/hook/clew.go` -- PostToolUse event recording
- `internal/cmd/hook/budget.go` -- session budget tracking
- `internal/cmd/session/create.go` -- session creation events
- `internal/cmd/session/wrap.go` -- session wrap events
- `internal/cmd/session/park.go` -- session park events

For v2 events, use the `.WithChannel()` method.
For v3 typed events, the `newTypedEvent` signature change handles it.

### 4.5 Wire channel into provenance collection call sites

Every call to `NewKnossosEntry` must pass the channel. The materializer already knows the channel via `Options.Channel`. Thread it through to provenance collector calls.

**Key files to modify** (find all `NewKnossosEntry` calls):
- `internal/materialize/materialize.go`
- `internal/materialize/materialize_mena.go`
- `internal/materialize/materialize_agents.go`
- `internal/materialize/materialize_claudemd.go`
- `internal/materialize/mena/engine.go`

### 4.6 Tests

**File: `internal/hook/clewcontract/channel_test.go`** (NEW)
- [ ] `TestEvent_WithChannel` -- channel added to meta for non-claude channels
- [ ] `TestEvent_WithChannel_Claude` -- no meta pollution for claude (default)
- [ ] `TestEvent_WithChannel_Empty` -- no meta pollution for empty string
- [ ] `TestTypedEvent_ChannelField` -- v3 events carry channel when non-claude
- [ ] `TestTypedEvent_ChannelOmitted` -- v3 events omit channel for claude (backward compat)
- [ ] `TestNewTypedEvent_AllConstructors_AcceptChannel` -- verify all ~20 constructors compile with new signature

**File: `internal/provenance/channel_test.go`** (NEW)
- [ ] `TestNewKnossosEntry_WithChannel` -- channel field populated
- [ ] `TestNewKnossosEntry_ClaudeDefault` -- channel omitted for claude
- [ ] `TestProvenanceManifest_StructuralEquality_Channel` -- channel difference detected
- [ ] `TestProvenanceManifest_BackwardCompat` -- manifests without channel field load correctly

**File: `internal/session/channel_test.go`** (NEW)
- [ ] `TestGeminiLifecycleMap_StartEvents` -- maps to ACTIVE
- [ ] `TestGeminiLifecycleMap_EndEvents` -- maps to ARCHIVED
- [ ] `TestGeminiLifecycleMap_SuspendEvents` -- maps to PARKED
- [ ] `TestGeminiLifecycleMap_UnknownEvent` -- returns false
- [ ] `TestClaudeLifecycleMap_StartEvents` -- maps to ACTIVE

**File: `internal/materialize/channel_provenance_test.go`** (NEW)
- [ ] `TestDualProjection_ProvenanceTracking` -- claude and gemini entries coexist in manifest
- [ ] `TestDualProjection_NoCollision` -- claude and gemini files don't overwrite each other
- [ ] `TestDualProjection_SelectiveRemoval` -- orphan removal respects channel boundaries

### Phase Gate

```bash
CGO_ENABLED=0 go test ./... -count=1
CGO_ENABLED=0 go vet ./...
CGO_ENABLED=0 go build ./cmd/ari
```

ALL must pass. Commit as:
```
feat(channel): P4 session federation -- channel-aware events and provenance
```

---

## File Write Conventions

```go
// Projected directories (.claude/, .gemini/): ALWAYS use this
fileutil.WriteIfChanged(path, content, 0644)

// State files (.knossos/, provenance): use this
fileutil.AtomicWriteFile(path, content, 0644)

// NEVER use os.WriteFile for projected directories
// NEVER use os.RemoveAll on projected directories
```

## Files You Will Touch (Phase 4 -- Exhaustive List)

| Action | File | What |
|--------|------|------|
| MODIFY | `internal/hook/clewcontract/event.go` | Add `WithChannel()` method to `Event` |
| MODIFY | `internal/hook/clewcontract/typed_event.go` | Add `Channel` field to `TypedEvent`, modify `newTypedEvent` signature |
| MODIFY | `internal/hook/clewcontract/typed_constructors.go` | Update all `NewTyped*Event` call sites for new `newTypedEvent` signature |
| MODIFY | `internal/provenance/provenance.go` | Add `Channel` to `ProvenanceEntry`, update `NewKnossosEntry` |
| MODIFY | `internal/provenance/manifest.go` | Update `structurallyEqual` to compare Channel |
| MODIFY | `internal/cmd/hook/clew.go` | Pass channel to event constructors |
| MODIFY | `internal/cmd/hook/budget.go` | Pass channel to budget events |
| MODIFY | `internal/cmd/session/create.go` | Pass channel to session events |
| MODIFY | `internal/cmd/session/wrap.go` | Pass channel to session events |
| MODIFY | `internal/cmd/session/park.go` | Pass channel to session events |
| MODIFY | `internal/materialize/materialize.go` | Thread channel to provenance calls |
| MODIFY | `internal/materialize/materialize_mena.go` | Thread channel to provenance calls |
| MODIFY | `internal/materialize/materialize_agents.go` | Thread channel to provenance calls |
| MODIFY | `internal/materialize/materialize_claudemd.go` | Thread channel to provenance calls |
| MODIFY | `internal/materialize/mena/engine.go` | Thread channel to provenance calls |
| CREATE | `internal/session/channel.go` | ChannelLifecycleMap, FSM mapping |
| CREATE | `internal/session/channel_test.go` | Lifecycle mapping tests |
| CREATE | `internal/hook/clewcontract/channel_test.go` | Event channel tests |
| CREATE | `internal/provenance/channel_test.go` | Provenance channel tests |
| CREATE | `internal/materialize/channel_provenance_test.go` | Dual-projection provenance tests |

**Do NOT touch any file not in this table.**

## What NOT To Do

1. Do NOT modify `StdinPayload` struct in `internal/hook/env.go`. The CC payload format is frozen.
2. Do NOT rename any existing files or directories.
3. Do NOT delete any existing tests.
4. Do NOT modify `.claude/` directory contents directly -- that is the materializer's job.
5. Do NOT skip provenance tracking for Gemini-projected files.
6. Do NOT use `os.Chdir` in tests -- use DI parameter injection (SCAR-030).
7. Do NOT use bare `exec.Command` -- always use `exec.CommandContext` with timeout (SCAR-010).
8. Do NOT add top-level fields to v2 `Event` struct -- use Meta for backward compat.
9. Do NOT change the `newTypedEvent` function to be exported -- it is intentionally unexported.
10. Do NOT modify `internal/paths/channel.go`, `internal/hook/adapter.go`, or `internal/materialize/compiler/` -- P1-P3 code is frozen.

## Review Protocol

After the phase commit, the code will be reviewed in Claude Code where full Knossos context (hooks, lint, scar regression tests) is available. Structure your commit cleanly:
- One commit for Phase 4
- All new files have test coverage
- `go vet ./...` and `go build ./...` pass
- No TODO comments without a tracking reference

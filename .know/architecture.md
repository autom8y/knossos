---
domain: architecture
generated_at: "2026-03-23T18:00:00Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "78abb186"
confidence: 0.82
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/initiative-history.md"
land_hash: "08704d8d5bd3b4b8b2c6a8a1590586893f35aa5a3c3d6f27c260cee86fca82b0"
---

# Codebase Architecture

## Package Structure

This is a Go project (module `github.com/autom8y/knossos`, Go 1.23+). The binary is `ari` (Ariadne). The layout follows the standard Go project layout with two primary source directories: `cmd/` for the entry point and `internal/` for all domain logic.

### Entry Point

`cmd/ari/` — single file `main.go`. Sets version metadata, wires embedded assets into the CLI binary, and delegates to `internal/cmd/root`.

### `internal/cmd/` — Command Layer (27 packages)

Each subdirectory implements one or more `cobra.Command` tree entries. Packages are thin: they wire flags, resolve paths, call into domain logic, and print results via `internal/output`.

| Package | User-facing command | Purpose |
|---|---|---|
| `internal/cmd/root` | (root) | Registers all subcommands; global flags; `PersistentPreRunE` for project discovery |
| `internal/cmd/session` | `ari session` | Session CRUD: create, park, resume, wrap, log, audit, status |
| `internal/cmd/sync` | `ari sync` | Dispatches to `internal/materialize` unified pipeline |
| `internal/cmd/rite` | `ari rite` | Rite list, switch, info |
| `internal/cmd/hook` | `ari hook` | Hook event handlers: agent-guard, budget, autopark, handshake, writeguard, context-injection |
| `internal/cmd/manifest` | `ari manifest` | Manifest diff, merge, validate |
| `internal/cmd/inscription` | `ari inscription` | CLAUDE.md region management |
| `internal/cmd/knows` | `ari knows` | .know/ file read/write |
| `internal/cmd/ledge` | `ari ledge` | .ledge/ artifact management (promote, list) |
| `internal/cmd/tribute` | `ari tribute` | TRIBUTE.md generation at session wrap |
| `internal/cmd/naxos` | `ari naxos` | Orphaned session cleanup |
| `internal/cmd/sails` | `ari sails` | Session signal files (gray/white sails) |
| `internal/cmd/handoff` | `ari handoff` | Session handoff generation |
| `internal/cmd/procession` | `ari procession` | Cross-rite procession workflow management |
| `internal/cmd/provenance` | `ari provenance` | Provenance manifest inspection |
| `internal/cmd/worktree` | `ari worktree` | Git worktree lifecycle for agent isolation |
| `internal/cmd/artifact` | `ari artifact` | Session artifact tracking |
| `internal/cmd/agent` | `ari agent` | Agent listing and inspection |
| `internal/cmd/land` | `ari land` | Cross-session .sos/land/ synthesis management |
| `internal/cmd/org` | `ari org` | Organization-scope resource sync |
| `internal/cmd/lint` | `ari lint` | Rite and agent lint checks |
| `internal/cmd/status` | `ari status` | Project/session status overview |
| `internal/cmd/explain` | `ari explain` | Explain knossos concepts |
| `internal/cmd/tour` | `ari tour` | Guided project walkthrough |
| `internal/cmd/ask` | `ari ask` | Search/query across project context |
| `internal/cmd/complaint` | `ari complaint` | Cassandra complaint filing |
| `internal/cmd/initialize` | `ari init` | Bootstrap new knossos project |
| `internal/cmd/common` | (shared) | Shared command utilities, embedded asset wiring, `NeedsProject` annotation |

### `internal/` — Domain Logic (36 packages)

**Core domain packages:**

| Package | Purpose | Key types |
|---|---|---|
| `internal/materialize` | Channel directory generation — the primary sync pipeline | `Materializer`, `RiteManifest`, `SyncOptions`, `SyncResult`, `Options`, `Result` |
| `internal/session` | Session lifecycle state machine, context, events | `Status`, `FSM`, `Context`, `Event`, `ClewEvent`, `Strand`, `Procession` |
| `internal/inscription` | CLAUDE.md/GEMINI.md region management (marker system) | `Manifest`, `Region`, `Marker`, `OwnerType`, `ParsedRegion`, `ParseResult` |
| `internal/provenance` | File-level provenance tracking; divergence detection | `ProvenanceManifest`, `ProvenanceEntry`, `OwnerType`, `ScopeType`, `Collector` |
| `internal/hook` | Hook event parsing, lifecycle adapter, clew contract events | `StdinPayload`, `Env`, `HookEvent`, `LifecycleAdapter`, `ClaudeAdapter`, `GeminiAdapter` |
| `internal/rite` | Rite discovery, context injection, switching | `RiteContext`, `ContextRow`, `Syncer` (interface) |
| `internal/manifest` | Project manifest load/parse/diff/merge | `Manifest`, `Format` |
| `internal/paths` | Path resolution, project root discovery, `.sos/` paths | `Resolver` |
| `internal/errors` | Typed exit codes and error constants | `KnossosError`, exit code constants (ExitSuccess=0 through ExitSwitchAborted=13+) |
| `internal/output` | Format-aware stdout/stderr printing (text/json/yaml) | `Printer`, `Format`, `Textable` (interface) |
| `internal/config` | KNOSSOS_HOME and XDG config dir resolution | `KnossosHome()`, `XDGDataDir()`, `ActiveOrg()` |
| `internal/frontmatter` | YAML frontmatter extraction from markdown | `FlexibleStringSlice`, `Parse()` |
| `internal/channel` | Canonical tool name ↔ channel wire name mapping | `CanonicalTool` map |
| `internal/registry` | Platform reference registry (agents, skills, CLI, dromena) | `RefKey`, `RefCategory` |
| `internal/lock` | Advisory file locking with stale detection (5-minute threshold) | `LockMetadata`, `Shared`/`Exclusive` |
| `internal/fileutil` | Atomic file writes (temp-file-then-rename) | `AtomicWriteFile()`, `WriteIfChanged()` |
| `internal/tokenizer` | Token counting (tiktoken cl100k_base) | `Counter` |
| `internal/checksum` | SHA256 checksums with `sha256:` prefix | (utility functions) |
| `internal/session/fsm.go` | 4-state session FSM (NONE→ACTIVE→PARKED/ARCHIVED) | `FSM` |
| `internal/naxos` | Orphan session cleanup (named after island where Ariadne was abandoned) | `OrphanReason`, `SuggestedAction` |
| `internal/tribute` | TRIBUTE.md session summary generation | `GenerateResult`, `Artifact`, `Decision`, `PhaseRecord` |
| `internal/suggest` | Contextual suggestion generation (pure: struct in, slice out) | `Suggestion`, `Kind`, `SessionInput` |
| `internal/perspective` | Agent perspective document assembly (pre-materialization view) | `PerspectiveDocument`, `LayerEnvelope`, `LayerStatus` |
| `internal/know` | .know/ domain file parsing (frontmatter + staleness detection) | `Meta`, `KnowFile` |
| `internal/ledge` | .ledge/ artifact promotion and management | (promote, list) |
| `internal/search` | Cross-project search/indexing for `ari ask` | `Entry`, `Index`, `Score` |

**Sub-packages of `internal/materialize/`:**

| Sub-package | Purpose |
|---|---|
| `internal/materialize/source` | 5-tier rite source resolution (project > user > knossos > org > embedded) |
| `internal/materialize/mena` | Mena (.dro.md/.lego.md) collection, routing, projection to commands/skills |
| `internal/materialize/hooks` | Hooks config loading, hooks settings merge |
| `internal/materialize/compiler` | `ChannelCompiler` interface; `ClaudeCompiler` (pass-through), `GeminiCompiler` (tool translation) |
| `internal/materialize/orgscope` | Org-scope resource sync |
| `internal/materialize/userscope` | User-scope resource sync (cross-rite agents, mena, hooks) |
| `internal/materialize/procession` | Procession template resolution |

**Sub-packages of `internal/hook/`:**

| Sub-package | Purpose |
|---|---|
| `internal/hook/clewcontract` | Append-only JSONL event log (16 event types); `BufferedEventWriter` (5s flush, thread-safe) |

**Hub packages** (imported by many): `internal/errors`, `internal/paths`, `internal/fileutil`, `internal/output`, `internal/frontmatter`

**Leaf packages** (no internal imports): `internal/registry`, `internal/provenance`, `internal/channel`, `internal/checksum`

---

## Layer Boundaries

The codebase follows a strict 3-tier layer model:

```
cmd/ari/main.go
      ↓
internal/cmd/* (command wiring — cobra handlers, flag parsing, output)
      ↓
internal/* (domain logic — no cobra, no output printing)
```

**Import direction rules observed:**

1. `cmd/ari/main.go` imports only: `knossos` (root embed package), `internal/cmd/common`, `internal/cmd/root`, `internal/errors`, `internal/output`
2. `internal/cmd/root` imports ALL command sub-packages (it is the hub of the command layer)
3. Domain packages do NOT import `internal/cmd/*` — the boundary is enforced by package layout
4. `internal/provenance` is declared a **leaf package** (imports only stdlib) — noted explicitly in `provenance.go` line 68: "NOTE: Provenance is a leaf package (no internal imports per ADR-0026)"
5. `internal/registry` is a **leaf package** — imports only stdlib (documented in source)
6. `internal/materialize` is the largest hub: imports `internal/errors`, `internal/fileutil`, `internal/paths`, `internal/provenance`, `internal/registry`, `internal/config`, `internal/checksum`, `internal/inscription`, `internal/channel`, `internal/frontmatter`
7. `internal/session` imports `internal/errors`, `internal/fileutil`, `internal/validation`, `internal/hook/clewcontract`
8. `internal/hook` imports `internal/hook/clewcontract` (sub-package), no upward imports

**Boundary-enforcement patterns:**

- The root embed package (`package knossos`, `embed.go`) sits at module root and is imported by `main.go` only — it is the single dependency inversion point for embedded assets
- `internal/materialize` uses sub-packages (`source/`, `mena/`, `hooks/`, `compiler/`, `userscope/`, `orgscope/`) with type-alias re-exports in the parent package for backward compatibility (e.g., `source.go`, `mena.go`, `hooks.go`, `user_scope.go`)
- `internal/cmd/common` acts as the shared context between `main.go` and all command packages — it holds embedded asset references and `NeedsProject` annotation logic

**Layer summary:**
- **CLI surface**: `cmd/ari/` + `internal/cmd/*`
- **Domain core**: `internal/materialize`, `internal/session`, `internal/inscription`, `internal/hook`, `internal/rite`, `internal/provenance`
- **Infrastructure leaf**: `internal/errors`, `internal/paths`, `internal/fileutil`, `internal/output`, `internal/config`, `internal/registry`, `internal/checksum`, `internal/channel`, `internal/frontmatter`

---

## Entry Points and API Surface

### Binary Entry Point

`cmd/ari/main.go` `main()`:
1. Sets version via `root.SetVersion(version, commit, date)`
2. Wires embedded assets: `common.SetEmbeddedAssets(knossos.EmbeddedRites, knossos.EmbeddedTemplates, knossos.EmbeddedHooksYAML)`
3. Wires user assets: `common.SetEmbeddedUserAssets(knossos.EmbeddedAgents, knossos.EmbeddedMena)`
4. Wires processions: `common.SetEmbeddedProcessions(knossos.EmbeddedProcessions)`
5. Calls `root.Execute()` — returns error if any command failed
6. Error printing is format-aware via `output.Printer`

### CLI Command Tree (all registered subcommands)

| Command | Package | Description |
|---|---|---|
| `ari session` | `internal/cmd/session` | Session lifecycle: create, park, resume, wrap, log, audit, status |
| `ari sync` | `internal/cmd/sync` | Unified sync pipeline (rite + user + org scopes) |
| `ari rite` | `internal/cmd/rite` | Rite management: list, switch, info |
| `ari hook` | `internal/cmd/hook` | Hook event handlers invoked by CC/Gemini lifecycle events |
| `ari manifest` | `internal/cmd/manifest` | Manifest diff, merge, validate |
| `ari inscription` | `internal/cmd/inscription` | CLAUDE.md region management |
| `ari knows` | `internal/cmd/knows` | .know/ codebase knowledge file management |
| `ari ledge` | `internal/cmd/ledge` | .ledge/ work product artifact promotion |
| `ari tribute` | `internal/cmd/tribute` | Generate TRIBUTE.md at session wrap |
| `ari naxos` | `internal/cmd/naxos` | Orphan session detection and cleanup |
| `ari sails` | `internal/cmd/sails` | Session signal files (gray/white sails for hand-off signals) |
| `ari handoff` | `internal/cmd/handoff` | Session handoff document generation |
| `ari procession` | `internal/cmd/procession` | Cross-rite coordinated workflow management |
| `ari provenance` | `internal/cmd/provenance` | Inspect PROVENANCE_MANIFEST.yaml |
| `ari worktree` | `internal/cmd/worktree` | Git worktree lifecycle management for agent isolation |
| `ari artifact` | `internal/cmd/artifact` | Session artifact tracking |
| `ari agent` | `internal/cmd/agent` | Agent listing, inspection, perspective view |
| `ari land` | `internal/cmd/land` | .sos/land/ cross-session synthesis management |
| `ari org` | `internal/cmd/org` | Organization-scope resource sync |
| `ari lint` | `internal/cmd/lint` | Rite and agent quality lint |
| `ari status` | `internal/cmd/status` | Project and session status overview |
| `ari explain` | `internal/cmd/explain` | Explain knossos concepts in context |
| `ari tour` | `internal/cmd/tour` | Guided project walkthrough |
| `ari ask` | `internal/cmd/ask` | Semantic search over project context |
| `ari complaint` | `internal/cmd/complaint` | File Cassandra-protocol complaints |
| `ari init` | `internal/cmd/initialize` | Bootstrap new knossos project |
| `ari version` | `internal/cmd/root` | Show version/build metadata |

### Global Flags (defined in `internal/cmd/root`)

- `--output, -o` (text/json/yaml) — format-aware output
- `--verbose, -v` — verbose JSON lines to stderr
- `--config` — explicit config file path
- `--project-dir, -p` — override project root discovery
- `--session-id, -s` — override session ID
- `--channel` — target channel: claude, gemini, or all (default: "all")

### Key Exported Interfaces

- `compiler.ChannelCompiler` — `CompileCommand()`, `CompileSkill()`, `CompileAgent()`, `ContextFilename()` — consumed by `internal/materialize` during agent/mena generation
- `provenance.Collector` — thread-safe provenance recording interface, threaded through all rite pipeline stages
- `output.Textable` — `Text() string` — implemented by all output structs to support text format printing
- `rite.Syncer` — `SyncRite(riteName string, keepOrphans bool) error` — satisfied by `Materializer`
- `internal/hook.LifecycleAdapter` — `ParsePayload(io.Reader) (*Env, error)` — implemented by `ClaudeAdapter` and `GeminiAdapter`

### Project Discovery

`paths.FindProjectRoot()` walks upward from CWD checking for `.knossos/`, `.claude/`, or `.gemini/` directories. `.knossos/` is the strongest signal (platform dir). Project root is the anchor for all relative path resolution.

---

## Key Abstractions

### 1. `Materializer` — `internal/materialize/materialize.go`

The central orchestrator. Holds: `resolver *paths.Resolver`, `sourceResolver *SourceResolver`, embedded FSes (`embeddedTemplates`, `embeddedAgents`, `embeddedMena`, `embeddedProcessions`), and a `channelDirOverride`. Its `Sync(SyncOptions)` method is the unified sync pipeline entry point, dispatching to rite scope (`MaterializeWithOptions`), user scope (`syncUserScope`), and org scope (`syncOrgScope`).

### 2. `RiteManifest` — `internal/materialize/materialize.go`

YAML representation of `rites/{name}/manifest.yaml`. Contains: `Name`, `Version`, `Agents []Agent`, `Dromena []string`, `Legomena []string`, `MCPServers []MCPServer`, `MCPPools []MCPPoolRef`, `HookDefaults`, `AgentDefaults`, `SkillPolicies`, `ArchetypeData`. This is the configuration contract for a rite — every materialization pass reads it to know what to generate.

### 3. `inscription.Manifest` — `internal/inscription/types.go`

Stored at `.knossos/KNOSSOS_MANIFEST.yaml`. Tracks region ownership in CLAUDE.md/GEMINI.md. Regions have three ownership types: `knossos` (always overwritten), `satellite` (never overwritten), `regenerate` (rebuilt from source). Region hashes enable divergence detection. This is the core of the idempotent, non-destructive CLAUDE.md merge system.

### 4. `session.Context` — `internal/session/context.go`

Parsed `SESSION_CONTEXT.md` — the mutable session document. Contains FSM `Status`, timestamps, `ActiveRite`, `CurrentPhase`, `Procession`, and `Strands` (for forked sessions via `ari fray`). Direct writes to this file are blocked by the writeguard hook; only `ari` CLI commands or Moirai agent may mutate it.

### 5. `session.FSM` — `internal/session/fsm.go`

4-state finite state machine: `NONE → ACTIVE → PARKED/ARCHIVED`, `PARKED → ACTIVE/ARCHIVED`. `ARCHIVED` is terminal. Transitions are validated before every session mutation.

### 6. `hook.StdinPayload` — `internal/hook/env.go`

The JSON structure CC/Gemini sends to hooks on stdin. Contains `SessionID`, `ToolName`, `ToolInput`, `HookEventName`, `CWD`, `TranscriptPath`. The critical architectural lesson: hook data arrives on stdin as JSON, NOT as environment variables (only `CLAUDE_PROJECT_DIR` is an env var).

### 7. `provenance.ProvenanceManifest` — `internal/provenance/provenance.go`

Stored at `.knossos/PROVENANCE_MANIFEST.yaml` (rite scope) and `.knossos/USER_PROVENANCE_MANIFEST.yaml` (user scope). Maps relative channel-dir paths to `ProvenanceEntry` records. `OwnerType` is `knossos`/`user`/`untracked`. Schema v2.0. Divergence detection promotes `knossos → user` on checksum mismatch.

### 8. `ChannelCompiler` interface — `internal/materialize/compiler/compiler.go`

Abstraction for harness-specific agent/mena compilation. `ClaudeCompiler` is pass-through; `GeminiCompiler` translates canonical tool names to Gemini wire names using `internal/channel.CanonicalTool`. This is the harness-agnosticism seam — templates use canonical names, compilers resolve at projection time.

### 9. `paths.Resolver` — `internal/paths/paths.go`

Anchored to `projectRoot`. Provides all structured path accessors: `SOSDir()`, `SessionsDir()`, `LocksDir()`, `HarnessMapDir()`, `SessionContextFile(sessionID)`, `SessionEventsFile(sessionID)`, `ActiveRiteFile()`, `KnossosManifestFile()`, etc. All domain packages accept a `*Resolver` rather than computing paths independently.

### 10. `FlexibleStringSlice` — `internal/frontmatter/frontmatter.go`

YAML type that accepts both comma-separated strings and YAML lists. Used in agent frontmatter for tool lists (e.g., `tools: "Bash, Read, Glob"` and `tools: [Bash, Read, Glob]` are both valid). A common convention across all agent files.

### Design Patterns Observed

- **Polymorphic YAML fields**: `FlexibleStringSlice` allows two syntax forms for the same field
- **Type-alias re-export pattern**: `internal/materialize/mena.go`, `source.go`, `hooks.go`, `user_scope.go` re-export sub-package types as aliases — enables sub-package extraction without breaking callers
- **Cascade/merge pattern**: `AgentDefaults` in `RiteManifest` merged into per-agent frontmatter; `SkillPolicies` evaluated per-agent; archetype data cascades from config file → manifest override → runtime render
- **Envelope pattern**: `hook.StdinPayload` wraps all CC hook data in one JSON envelope; `clewcontract` events follow a typed envelope with `type` discriminator
- **Registry pattern**: `internal/registry` provides stable `RefKey` constants for platform references, with fallback hints — prevents hard-coded string duplication across the codebase
- **4-tier source resolution**: rite sources resolve as project > user > knossos > org > embedded (5 tiers including embedded)

---

## Data Flow

### 1. Sync Pipeline (rite scope): `ari sync` → channel directory

```
User runs: ari sync --rite ecosystem

internal/cmd/sync
  ↓ calls
internal/materialize.Materializer.Sync(SyncOptions{Scope: ScopeRite, RiteName: "ecosystem"})
  ↓
source.SourceResolver.Resolve("ecosystem")
  → checks: project/rites/ → user-level rites → $KNOSSOS_HOME/rites → embedded FS
  → returns ResolvedRite{RitePath, ManifestPath, Source}
  ↓
Load RiteManifest from resolved.ManifestPath (YAML parse)
  ↓
enrichArchetypeData(manifest, riteFS) — loads orchestrator.yaml config
  ↓
materializeAgents(manifest, ...) — reads rites/{name}/agents/*.md, transforms frontmatter,
  applies AgentDefaults, SkillPolicies, MCP config; writes to .claude/agents/*.md
  ↓
materializeMena(manifest, ...) — collects .dro.md/.lego.md files, routes to
  .claude/commands/ (dromena) or .claude/skills/ (legomena), strips extensions
  ↓
materializeInscription(manifest, ...) — runs inscription pipeline:
  LoadManifest → Generate sections from templates → Merge (preserve satellite regions) →
  Backup if migration → AtomicWrite CLAUDE.md → UpdateHashes in KNOSSOS_MANIFEST.yaml
  ↓
materializeSettings(...) — merges hooks into .claude/settings.json
  ↓
provenanceCollector.Save() → writes .knossos/PROVENANCE_MANIFEST.yaml
```

All writes use `fileutil.WriteIfChanged()` to avoid CC file-watcher disruption. `writeIfChanged` is the atomic-safe idempotency gate.

### 2. User Scope Sync: agents, mena, hooks → `~/.claude/`

```
Materializer.syncUserScope(opts)
  ↓
userscope.SyncUserScope(params)
  → source: $KNOSSOS_HOME/agents/, $KNOSSOS_HOME/mena/, $KNOSSOS_HOME/hooks/
    (fallback: embedded FS from binary)
  → target: ~/.claude/agents/, ~/.claude/commands/, ~/.claude/skills/, ~/.claude/hooks/
  → CollisionChecker: reads .knossos/PROVENANCE_MANIFEST.yaml to prevent user scope
    from shadowing rite resources (flat-name collision prevention)
  → provenance.LoadOrBootstrap → saves to .knossos/USER_PROVENANCE_MANIFEST.yaml
```

### 3. Hook Pipeline: CC lifecycle event → `ari hook <subcommand>`

```
CC lifecycle event fires (e.g., PreToolUse with tool=Write)
  ↓ CC spawns ari as subprocess
ari hook writeguard  (or budget, agent-guard, context-injection, autopark, handshake)
  ↓
internal/hook.ParseEnv()
  → reads JSON payload from stdin (NOT env vars)
  → CLAUDE_PROJECT_DIR from env
  → GetAdapter() checks KNOSSOS_CHANNEL env var → returns ClaudeAdapter or GeminiAdapter
  → adapter.ParsePayload(stdin) → returns *Env{SessionID, ToolName, ToolInput, ...}
  ↓
Command-specific logic (e.g., writeguard: checks if ToolInput.path matches *_CONTEXT.md)
  ↓
Output: JSON to stdout { decision: "block"/"allow", reason: "..." }
  (CC reads permissionDecision from hookSpecificOutput envelope, NOT top-level decision)
  ↓
Optional: clewcontract.BufferedEventWriter appends event to .sos/sessions/{id}/events.jsonl
```

### 4. Session Pipeline: session lifecycle events

```
ari session create --initiative "Feature X"
  ↓
internal/session.FSM.ValidateTransition(NONE → ACTIVE)
  ↓
lock.Acquire(.sos/sessions/.locks/{id}.lock, Exclusive, 10s timeout)
  ↓
Write SESSION_CONTEXT.md (via fileutil.AtomicWriteFile)
  ↓
lock.Release()
  ↓
clewcontract: emit EVENT_SESSION_START to events.jsonl

Hook fires (PreToolUse writeguard) when any agent tries to write *_CONTEXT.md directly
  → blocks write, returns "block" decision to CC
  → ensures only ari commands can mutate session state
```

### 5. Configuration Merge Points

- **Agent frontmatter cascade**: rite `manifest.yaml` `agent_defaults` → individual agent frontmatter — merged during `materializeAgents`
- **Skill policies**: `manifest.yaml` `skill_policies` evaluated per-agent during agent materialization — adds `skills:` frontmatter entries conditionally
- **Archetype data**: `knossos/archetypes/orchestrator.yaml` → `manifest.yaml` `archetype_data` (per-archetype) → runtime `OrchestratorData` struct — template rendering
- **Hooks**: rite `manifest.yaml` `hooks` + `hook_defaults` + `config/mcp-pools.yaml` → merged into `.claude/settings.json` `hooks` array

---

## Knowledge Gaps

1. `internal/cmd/hook` sub-commands: individual hook handlers (`agent-guard`, `budget`, `autopark`, `handshake`, `writeguard`, `context-injection`) were not read in detail — their internal logic is partially inferred from package-level doc comments and rules files
2. `internal/ledge` package internals not fully explored — promote and auto-promote logic observed only at file listing level
3. `internal/search` — index/scoring internals not read; structure inferred from file names
4. `internal/perspective` — full layer assembly logic not traced (types read, `assemble.go` not read)
5. `internal/procession` — cross-rite workflow template details not read
6. `internal/cmd/ask` sub-command implementation not read beyond signature in root.go
7. Org scope resolution (`orgscope/`) — delegation confirmed but internal org resource paths not traced

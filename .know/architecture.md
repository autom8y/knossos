---
domain: architecture
generated_at: "2026-03-08T21:08:37Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "b702931"
confidence: 0.82
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/initiative-history.md"
land_hash: "313c675e38c3e4000caa21dfac68c38f337b5fb95d53f85444ef8d43174f4171"
---

# Codebase Architecture

## Package Structure

The Knossos project is a Go module (`github.com/autom8y/knossos`, Go 1.23+) that compiles to a single binary named `ari` (Ariadne). The source tree has two primary roots: `cmd/ari/` (entry point) and `internal/` (all domain logic, split into 30+ packages).

### `cmd/ari/`

One file: `cmd/ari/main.go`. Contains only wiring. Sets build-time version variables, wires embedded assets via `knossos.Embedded*` vars, calls `root.Execute()`, and routes errors through `output.Printer`. No domain logic.

### Root Package (`embed.go`)

The module root (`package knossos`) exists solely to host `//go:embed` directives, because Go embed requires declarations adjacent to the embedded directories. Exports: `EmbeddedRites`, `EmbeddedTemplates`, `EmbeddedHooksYAML`, `EmbeddedAgents`, `EmbeddedMena` — all `embed.FS` or `[]byte`. Consumed only by `cmd/ari/main.go`.

### `internal/cmd/` — CLI Command Layer (25 packages)

Each sub-package owns one cobra command group. All follow the same pattern: a `New*Cmd()` constructor that receives global flag pointers (`*string`, `*bool`) and returns `*cobra.Command`. Command logic calls domain packages; it does not contain business logic directly.

| Sub-package | Command | Purpose |
|---|---|---|
| `root` | `ari` | Root command; registers all subcommands; config/path init in `PersistentPreRunE` |
| `session` | `ari session` | Session lifecycle (create, park, resume, wrap, fray, log, field, gc, etc.) |
| `hook` | `ari hook` | CC hook handlers (writeguard, budget, autopark, clew, agentguard, etc.) |
| `sync` | `ari sync` | Remote sync state management |
| `inscription` | `ari inscription` | CLAUDE.md region management (diff, rollback, validate, sync) |
| `manifest` | `ari manifest` | Manifest load, show, diff, merge, validate |
| `rite` | `ari rite` | Rite listing, info, status, invocation, release, validate |
| `agent` | `ari agent` | Agent listing, new, update, validate, embody |
| `worktree` | `ari worktree` | Git worktree lifecycle (create, clone, switch, remove, export, import) |
| `ask` | `ari ask` | Natural language query over platform knowledge (search index) |
| `artifact` | `ari artifact` | Artifact registry queries and registration |
| `knows` | `ari knows` | `.know/` file management (status, refresh) |
| `handoff` | `ari handoff` | Agent handoff protocol (prepare, execute, status, history) |
| `sails` | `ari sails` | White Sails confidence check and gate |
| `naxos` | `ari naxos` | Orphaned session scanner |
| `tribute` | `ari tribute` | TRIBUTE.md generation at session wrap |
| `validate` | `ari validate` | Rite/manifest validation |
| `lint` | `ari lint` | Lint rite source files |
| `initialize` | `ari init` | Bootstrap new Knossos project |
| `org` | `ari org` | Org-scope management (current, list, set) |
| `provenance` | `ari provenance` | Provenance manifest inspection |
| `land` | `ari land` | Cross-session experience synthesis (`.sos/land/`) |
| `ledge` | `ari ledge` | `.ledge/` artifact promotion |
| `explain` | `ari explain` | Concept explanation registry |
| `tour` | `ari tour` | Guided tour collection |
| `status` | `ari status` | Project/session status summary |
| `common` | — | Shared types: `BaseContext`, `SessionContext`, embedded assets, arg helpers, group annotations |

**Hub within `internal/cmd/`**: `common` is imported by every other cmd package (121 import references).

### `internal/` — Domain Logic Packages

| Package | Files | Purpose | Classification |
|---|---|---|---|
| `errors` | 2 | Structured `Error` type with code, message, details, exit code; 30 exit codes; `Wrap`, `New`, `IsHandled` | **Leaf** (no internal imports) |
| `output` | 3 | `Printer` type with `text`/`json`/`yaml` format support; `Textable` interface; tabwriter for tables | **Leaf** (no internal imports) |
| `config` | 2 | `KnossosHome()`, `XDGDataDir()`, `ActiveOrg()` — env var and filesystem resolution, singleton pattern | **Leaf** (no internal imports) |
| `paths` | 2 | `Resolver` (project-relative paths), `FindProjectRoot()`, XDG user path functions; hub for directory structure. `TargetChannel` interface has 5 methods: `Name()`, `DirName()`, `ContextFile()`, `ContextFilePath(projectRoot)`, `SkillsDir(projectRoot)`. `ChannelByName` derives from `AllChannels()`. `FindProjectRoot` searches for `.knossos`, `.claude`, `.gemini` markers. | imports `errors` |
| `fileutil` | 2 | Atomic file writes, directory creation, path helpers | imports `errors` |
| `checksum` | 2 | SHA256 with `sha256:` prefix convention | imports nothing internal |
| `frontmatter` | 3 | `Parse()` for `---` delimited YAML; `FlexibleStringSlice` (accepts comma-string or YAML list) | imports nothing internal |
| `registry` | 3 | Denial-recovery registry: stable `RefKey` constants → `RefEntry` values; **leaf** (no internal imports) | **Leaf** |
| `lock` | 3 | Advisory file locking: `Manager`, `Lock`, `LockMetadata` v2; 5-min stale threshold; `Shared`/`Exclusive` types | imports `errors` |
| `validation` | 4 | Schema validators for artifact, frontmatter, handoff, sails | imports `errors` |
| `hook` | 3+1 | `StdinPayload`, `Env`, `ParseEnv()`; 18 `HookEvent` constants (canonical snake_case names, e.g., `EventPreTool`, `EventSessionStart`); canonical vocabulary layer with `canonicalToWire`/`wireToCanonical` bidirectional translation tables (18 events), `CanonicalToWire()`/`WireToCanonical()` functions. Hook events use snake_case canonical names internally; adapters translate to/from wire format at the boundary. `clewcontract/` sub-package (16-event JSONL ledger) | imports nothing internal (except `errors` for helpers) |
| `session` | 12+ | `Context` type (SESSION_CONTEXT.md), FSM (NONE/ACTIVE/PARKED/ARCHIVED), events JSONL, lock integration | imports `errors`, `fileutil`, `paths`, `lock`, `hook/clewcontract`, `validation` |
| `manifest` | 5 | `Manifest`, load/parse/diff/merge; supports JSON+YAML, git refs | imports `errors`, `fileutil` |
| `inscription` | 7 | CLAUDE.md region system: `Generator`, `Manifest`, `Marker`; knossos/satellite/regenerate owner types | imports `errors`, `fileutil`, `paths`, `frontmatter` |
| `provenance` | 5 | `ProvenanceManifest`, `ProvenanceEntry` (owner, scope, checksum, source); `Collector` interface | imports `errors`, `fileutil`, `checksum` |
| `agent` | 6 | `AgentFrontmatter`, archetype detection, regeneration logic, MCP validation | imports `errors`, `frontmatter`, `paths` |
| `rite` | 6 | `Rite`, `Discovery`, budget tracking, context loading, invoker; `RiteForm` enum | imports `errors`, `paths`, `config` |
| `mena` | 4 | Mena source resolution, walk, type detection (`.dro.md`/`.lego.md`) | imports `errors`, `paths` |
| `channel` | — | Tool canonicalization layer: `CanonicalTool` map (11 tools), `CanonicalToWireTool()`/`WireToCanonicalTool()` bidirectional translation functions. Translates between knossos canonical tool names (snake_case, e.g., `run_shell`) and channel wire names (e.g., `Bash` for CC, `run_shell_command` for Gemini). | imports `paths` |
| `materialize` | 30+ | **Central hub**: `Materializer`, `RiteManifest`, `Agent`, `Options`, `Result`, `SyncOptions`, `SyncResult`; orchestrates full sync pipeline through 9 sub-packages | imports `errors`, `paths`, `fileutil`, `checksum`, `config`, `registry`, `provenance`, `inscription`, `frontmatter`, `sync` |
| `materialize/source` | 2 | Source resolution: `SourceType` constants (`project`/`user`/`knossos`/`org`/`explicit`/`embedded`), `ResolvedRite`, `SourceResolver` | imports `errors`, `paths`, `config` |
| `materialize/mena` | 5 | Mena projection to `commands/`+`skills/`: `MenaSource`, `MenaProjectionOptions`, `CollectMena`, `SyncMena`, `RouteMenaFile` | imports `errors`, `paths`, `fileutil`, `frontmatter` |
| `materialize/userscope` | — | User-scope sync types: `SyncResource`, `UserScopeResult`, `UserSyncChanges` | imports `errors`, `fileutil`, `paths` |
| `materialize/orgscope` | — | Org-scope sync logic: `SyncOrgScope`, `SyncOrgScopeParams` | imports `errors`, `paths`, `config` |
| `materialize/hooks` | — | Hook defaults generation | imports `errors`, `paths` |
| `sync` | 2 | `State`, `StateManager`; `.knossos/sync/state.json` (schema v1.1) | imports `errors`, `paths`, `checksum`, `fileutil` |
| `artifact` | 5 | `Entry`, `SessionRegistry`, `ProjectIndex`; federated artifact registry at session + project levels | imports `errors`, `paths`, `fileutil` |
| `sails` | 6 | White Sails: 3 colors (WHITE/GRAY/BLACK), `ContractViolation`, clew contract validation, gate logic | imports `hook/clewcontract` |
| `tribute` | 4 | `GenerateResult`, TRIBUTE.md renderer; reads session events+artifacts+sails for session summary | imports `session`, `sails`, `artifact` |
| `naxos` | 3 | `OrphanedSession`, `ScanResult`, `ScanConfig`; scans `.sos/sessions/` for abandoned sessions | imports `errors`, `session` |
| `worktree` | 5 | Git worktree lifecycle: `Metadata`, `Operations`, session integration | imports `errors`, `paths`, `session` |
| `ledge` | 2 | `.ledge/` auto-promotion: `AutoPromote`, `Promote` | imports `errors`, `paths`, `artifact` |
| `know` | 5 | `Meta` type for `.know/` frontmatter; staleness detection via git diff; AST-based semantic diff (`astdiff.go`) | imports `errors`, `frontmatter` |
| `search` | 5 | `SearchIndex`, `SearchEntry`, domain-based scoring; 7 collectors (commands, concepts, rites, agents, dromena, routing, sessions); synonym expansion | imports `paths`, `cobra` |
| `suggest` | 2 | `Suggestion`, `SessionInput`, `SubagentInput`; pure functions, no I/O; proactive intelligence | no internal imports |
| `perspective` | 6 | `PerspectiveDocument` with 9 layer envelopes (L1-L9): identity, perception, capability, constraint, memory, position, surface, horizon, provenance | imports `errors`, `paths`, `agent`, `frontmatter`, `provenance`, `rite` |
| `tokenizer` | 2 | Token counting via tiktoken-go | no internal imports |
| `assets` | 1 | `assets.go` — re-exports embedded asset types | imports root package |

**Hub packages** (imported most broadly):
- `errors`: 115 references — imported by almost all domain packages
- `output`: 104 references — imported by all cmd packages
- `paths`: 94 references — imported by all packages needing filesystem navigation
- `session`: 53 references — cross-cutting session state consumer
- `provenance`: 39 references — file ownership tracking
- `hook/clewcontract`: 34 references — event emission

**Leaf packages** (no internal imports):
- `errors`, `output`, `config`, `registry`, `suggest`, `tokenizer`, `validation` (partial), `frontmatter`, `checksum`

## Layer Boundaries

The codebase has a three-layer architecture with enforced import direction:

```
Layer 3 (CLI Surface):    cmd/ari/main.go
                                |
Layer 2 (Command Wiring): internal/cmd/{root,session,hook,sync,...}
                                |
Layer 1 (Domain Logic):   internal/{materialize,session,inscription,...}
                                |
Layer 0 (Leaf/Util):      internal/{errors,output,paths,config,fileutil,checksum,frontmatter,registry,lock}
```

**Layer 0 — Utility/Leaf packages**: `errors`, `output`, `config`, `fileutil`, `checksum`, `frontmatter`, `registry`, `suggest`, `tokenizer`. Zero or near-zero internal imports. Pure utility. `errors` is the most-imported package in the codebase (115 references).

**Layer 1 — Domain Logic**: `materialize`, `session`, `inscription`, `provenance`, `hook`, `rite`, `agent`, `artifact`, `sails`, `tribute`, `naxos`, `worktree`, `sync`, `search`, `perspective`, `know`, `ledge`. Import L0 freely; import each other where domain dependencies exist (e.g., `tribute` imports `session`, `sails`, `artifact`; `session` imports `lock`, `hook/clewcontract`). The `materialize` package is the dominant hub of Layer 1, importing 9 sub-packages of its own plus `inscription`, `provenance`, `sync`.

**Layer 2 — Command Wiring**: `internal/cmd/*`. Each package wires one cobra command group. Imports from Layer 1 (domain logic) and Layer 0 (output, errors). Never imported by Layer 1 — import direction is enforced (domain packages do not import cmd packages).

**Layer 3 — Entry Point**: `cmd/ari/main.go`. Imports `internal/cmd/root` (for `Execute()`), `internal/cmd/common` (for embedded asset wiring), `internal/errors`, `internal/output`, and the root package (`knossos`) for embedded assets.

**Boundary enforcement patterns**:
- `provenance` is explicitly documented as a **leaf package** — no internal imports per ADR-0026. It uses plain strings instead of importing `source.SourceType` to avoid circular dependency (TENSION-005 in design-constraints).
- `registry` is documented as a **leaf package** — "imports only stdlib."
- `materialize` achieves clean boundaries through sub-packages: `materialize/source`, `materialize/mena`, `materialize/userscope`, `materialize/orgscope`, `materialize/hooks` each handle a bounded concern, preventing the monolith from becoming unmaintainable.
- `internal/cmd/common` acts as the shared context bridge between the CLI layer and domain layer — it imports `session`, `lock`, `output`, `paths` so individual cmd packages avoid importing domain packages directly for common operations.

**Notable cross-cutting imports** (Layer 1 → Layer 1):
- `session` imports `hook/clewcontract` (event emission)
- `tribute` imports `session`, `sails`, `artifact` (all domain)
- `sails` imports `hook/clewcontract` (event log reading)
- `perspective` imports `agent`, `provenance`, `rite`, `frontmatter` (multi-domain assembly)

## Entry Points and API Surface

### Binary Entry Point

`cmd/ari/main.go`:
1. Sets build-time vars (`version`, `commit`, `date`) via `root.SetVersion()`
2. Sets embedded assets via `common.SetBuildVersion()`, `common.SetEmbeddedAssets()`, `common.SetEmbeddedUserAssets()`
3. Calls `root.Execute()` — delegates entirely to cobra
4. Error handling: checks `errors.IsHandled()` (to avoid double-printing); calls `printer.PrintError(err)` for unhandled errors; exits with `errors.GetExitCode(err)`

### Cobra Command Tree

Root command: `ari` (wired in `internal/cmd/root/root.go`)

Global flags: `--output/-o` (text/json/yaml), `--verbose/-v`, `--config`, `--project-dir/-p`, `--session-id/-s`

`PersistentPreRunE` on root: validates output format, inits config via viper, discovers project root via `paths.FindProjectRoot()`.

**All subcommands** (registered in `root.init()`):

| Command | Description |
|---|---|
| `ari session` | Session lifecycle management |
| `ari session create` | Create new session |
| `ari session park` | Park active session |
| `ari session resume` | Resume parked session |
| `ari session wrap` | Wrap session and archive |
| `ari session fray` | Fork session into strand |
| `ari session log` | Log event to events.jsonl |
| `ari session field` | Get/set session context fields |
| `ari session gc` | Garbage-collect old sessions |
| `ari session status` | Show session status |
| `ari session list` | List sessions |
| `ari session audit` | Audit session events log |
| `ari session claim` | Claim session ownership by CC session ID |
| `ari session snapshot` | Snapshot session state |
| `ari session timeline` | Show session timeline |
| `ari manifest` | Manifest operations |
| `ari inscription` | CLAUDE.md region management |
| `ari sync` | Remote sync operations |
| `ari validate` | Rite/manifest validation |
| `ari handoff` | Agent handoff protocol |
| `ari worktree` | Git worktree management |
| `ari hook` | CC hook handlers |
| `ari hook context` | Inject session context (SessionStart) |
| `ari hook autopark` | Auto-park on stop (Stop event) |
| `ari hook budget` | Tool budget tracking (PreToolUse) |
| `ari hook writeguard` | Guard writes to context files |
| `ari hook clew` | Clew contract event emission |
| `ari hook agentguard` | Agent capability enforcement |
| `ari hook attributionguard` | Attribution enforcement |
| `ari hook subagent` | Subagent event tracking |
| `ari hook sessionend` | Session end handler |
| `ari hook precompact` | Pre-compact handler |
| `ari hook suggest` | Proactive intelligence suggestions |
| `ari hook gitconventions` | Git conventions enforcement |
| `ari hook validate` | Hook validation |
| `ari knows` | `.know/` file management |
| `ari artifact` | Artifact registry |
| `ari sails` | White Sails confidence gate |
| `ari naxos` | Orphaned session cleanup |
| `ari rite` | Rite management |
| `ari agent` | Agent management |
| `ari tribute` | TRIBUTE.md generation |
| `ari init` | Project bootstrapping |
| `ari provenance` | Provenance manifest inspection |
| `ari org` | Org-scope management |
| `ari land` | Cross-session synthesis |
| `ari ledge` | `.ledge/` artifact promotion |
| `ari lint` | Rite lint |
| `ari status` | Project status |
| `ari explain` | Concept explanation |
| `ari tour` | Guided tour |
| `ari ask` | Natural language query |
| `ari version` | Version info |

### Key Exported Interfaces

**`output.Textable`** (in `internal/output/output.go`): Interface requiring `Text() string`. Implemented by output structs in every cmd package. `Printer.Print()` calls `Text()` for text format.

**`provenance.Collector`** (in `internal/provenance/`): Interface for recording file provenance. `defaultCollector` (mutex-guarded) and `NullCollector` (dry-run). Threaded through the materialize pipeline.

**`rite.Syncer`** (in `internal/rite/`): Interface `SyncRite(riteName string, keepOrphans bool) error`. Implemented by `materialize.Materializer`. Used to decouple rite from materialize in some call paths.

**`search.SynonymSource`** (in `internal/search/`): Interface for synonym expansion. Implementations: `StaticSynonymSource`, `OrchestratorSynonymSource`, `CompositeSynonymSource`.

## Key Abstractions

### 1. `materialize.Materializer` — `internal/materialize/`
The core engine. Holds `resolver *paths.Resolver`, `templatesDir string`, `embeddedRites embed.FS`, `embeddedTemplates embed.FS`. Central `Sync(SyncOptions) (*SyncResult, error)` method dispatches to rite scope (`MaterializeWithOptions`), org scope (`syncOrgScope`), and user scope (`syncUserScope`). All `.claude/` generation flows through this type.

### 2. `materialize.RiteManifest` — `internal/materialize/materialize.go`
The deserialized `manifest.yaml` for a rite. Key fields: `Name`, `Agents []Agent`, `Dromena []string`, `Legomena []string`, `Hooks []string`, `HookDefaults`, `AgentDefaults map[string]any`, `SkillPolicies []SkillPolicy`, `ArchetypeData map[string]map[string]any`. Drives what gets materialized.

### 3. `session.Context` — `internal/session/context.go`
Deserialized `SESSION_CONTEXT.md` (YAML frontmatter + body). Fields: `SessionID`, `Status`, `Initiative`, `Complexity`, `ActiveRite`, `CurrentPhase`, `Strands []Strand` (for fray), `FrameRef`, `ClaimedBy`. FSM transitions via `internal/session/fsm.go`. Mutations only through Moirai agent or `ari` CLI commands — enforced by writeguard hook.

### 4. `provenance.ProvenanceManifest` — `internal/provenance/provenance.go`
Tracks ownership and checksums of all files in `.claude/`. `Entries map[string]*ProvenanceEntry`. Each entry has `Owner` (knossos/user/untracked), `Scope` (rite/user), `SourcePath`, `SourceType`, `Checksum` (sha256: prefix), `LastSynced`. Written to `.knossos/PROVENANCE_MANIFEST.yaml`. Enables divergence detection and safe ownership transitions.

### 5. `hook.StdinPayload` / `hook.Env` — `internal/hook/env.go`
`StdinPayload` is the raw JSON from CC stdin: `SessionID`, `HookEventName`, `ToolName`, `ToolInput` (raw JSON), `ToolResponse`, `CWD`, `Prompt`, `Trigger`. `Env` is the parsed, convenient form used by hook handlers. `ParseEnv()` reads stdin JSON; `CLAUDE_PROJECT_DIR` is the only env var still read directly.

### 6. `output.Printer` — `internal/output/output.go`
Format-aware output handler. Formats: `FormatText`, `FormatJSON`, `FormatYAML`. `Print(data any)` dispatches by format; for text, checks `Textable` interface. Every cmd package constructs a `Printer` via `common.BaseContext.GetPrinter()`. This is the universal output contract.

### 7. `inscription.Generator` / KNOSSOS region system — `internal/inscription/`
Manages CLAUDE.md sections via `<!-- KNOSSOS:START region-name -->` markers. Three owner types: `knossos` (always overwritten on sync), `satellite` (user-owned, never overwritten), `regenerate` (regenerated from source). `Manifest` in `KNOSSOS_MANIFEST.yaml` tracks region ownership and SHA256 hashes. `Generator` renders sections from Go templates using `sprig` functions.

### 8. `errors.Error` — `internal/errors/errors.go`
Structured error with `Code string` (30+ named constants), `Message string`, `Details map[string]any`, `ExitCode int`, and unexported `cause error` for chain traversal. Exit codes 0–21 are canonically defined. `errors.IsHandled()` prevents double-printing. `errors.GetExitCode()` provides the OS exit code.

### 9. `paths.Resolver` — `internal/paths/paths.go`
Project-relative path resolver. Created from `projectRoot` string. Methods: `ClaudeDir()`, `SOSDir()`, `SessionsDir()`, `LocksDir()`, `CCMapDir()`, `WipDir()`, `ArchiveDir()`, `LandDir()`, `RitesDir()`, `KnossosDir()`, `KnossosSyncDir()`, `ReadActiveRite()`. Also standalone XDG user path functions: `UserAgentsDir()`, `UserCommandsDir()`, `UserSkillsDir()`, `UserRitesDir()`, `OrgRitesDir()`.

### 10. `frontmatter.FlexibleStringSlice` — `internal/frontmatter/frontmatter.go`
Polymorphic YAML type that accepts both `"Bash, Read, Glob"` (comma-string) and `[Bash, Read, Glob]` (YAML list). Used for agent `tools` frontmatter field. A recurring pattern: agent files use both forms in practice.

### Design Patterns

**Cascade/Merge pattern** (materialize pipeline): Configuration merges in priority order: agent frontmatter overrides > rite `agent_defaults` > archetype defaults > knossos defaults. The 4-tier source resolution chain: `project > dependency > shared > user`. Reflected in `materialize.SyncOptions.Scope`.

**Envelope pattern** (`perspective.LayerEnvelope`): Every layer of agent perspective is wrapped in a uniform `LayerEnvelope` with `Status` (RESOLVED/PARTIAL/OPAQUE/FAILED), `SourceFiles`, `ResolutionMethod`, `Gaps`. Enables uniform handling across 9 different layer types (L1–L9).

**Type alias re-export pattern**: `materialize` extensively uses `type X = subpkg.X` to maintain backward compatibility while decomposing into sub-packages. Callers continue using `materialize.RiteManifest` while the implementation is in sub-packages. Same pattern in `sync_types.go`, `mena.go`, `source.go`.

**Idempotency pattern**: `writeIfChanged()` in materialize reads existing file before writing — skips write if content is identical, preventing CC file-watcher triggers. `provenance.Save()` uses `structurallyEqual()` to skip writes when only timestamps change.

**Polymorphic YAML deserialization**: `session.strandList.UnmarshalYAML()` accepts both old format (`[]string`) and new format (`[]Strand`) for backward compatibility. Similar pattern in `frontmatter.FlexibleStringSlice`.

## Data Flow

### 1. Sync Pipeline (Primary: rite source → `.claude/`)

```
rites/{name}/manifest.yaml
    ↓ source resolution (materialize/source.SourceResolver)
       checks: project rites/ → user rites/ → knossos rites/ → embedded
    ↓ RiteManifest deserialization
    ↓ Materializer.Sync(SyncOptions)
       ├── materializeAgents() → .claude/agents/*.md
       │     agent frontmatter + agent_defaults merge
       │     archetype rendering (orchestrator.md.tpl)
       │     skill policies applied (SkillPolicy)
       │     provenance.Collector records each file
       ├── materializeMena() → .claude/commands/ and .claude/skills/
       │     .dro.md files → commands/, .lego.md files → skills/
       │     StripMenaExtension, RouteMenaFile
       ├── materializeCLAUDEmd() → .claude/CLAUDE.md
       │     inscription.SyncCLAUDEmd(): render templates → merge regions
       │     preserves satellite regions, overwrites knossos regions
       ├── materializeSettings() → .claude/settings.json
       │     MCP server merge (union: add/update rite, preserve satellite)
       ├── materializeRules() → .claude/rules/*.md
       ├── materializeHooks() → hooks config
       └── sync.StateManager.Save() → .knossos/sync/state.json
    ↓ provenance.Collector.Save() → .knossos/PROVENANCE_MANIFEST.yaml
    ↓ SyncResult returned to cmd layer for output
```

Configuration merge points within the pipeline:
- Agent frontmatter: explicit fields > `agent_defaults` in manifest > archetype defaults
- Mena: 4-tier source resolution (project > dependency > shared > user)
- MCP servers: rite servers merged with existing satellites (union, no overwrite)
- CLAUDE.md: knossos regions overwritten; satellite regions preserved; `regenerate` regions re-rendered from source

### 2. Session Event Pipeline

```
CC lifecycle event (SessionStart, Stop, PreToolUse, etc.)
    ↓ CC sends JSON to ari hook subcommand stdin
    ↓ hook.ParseEnv() → hook.Env (session ID, event type, tool info)
    ↓ internal/cmd/hook/{context,autopark,budget,clew,...}.go
    ↓ session.Context.Load() from SESSION_CONTEXT.md
    ↓ session state mutation (transition, field update, etc.)
    ↓ hook/clewcontract.BufferedEventWriter.Emit()
       → appends typed event to .sos/sessions/{id}/events.jsonl (JSONL, append-only)
    ↓ hook output (hookSpecificOutput envelope with permissionDecision for CC)
```

The events.jsonl file supports 3 format versions (v1 legacy `Event`, v2 flat `ClewEvent`, v3 typed `TypedEvent`) — interleaved in same file per ADR-0027 backward compat bridge in `session/events_read.go`.

### 3. Hook Pipeline (CC integration)

```
Claude Code tool use
    ↓ CC sends {"hook_event_name": "PreToolUse", "tool_name": "...", "tool_input": {...}, ...} to stdin
    ↓ hook.parseStdin() → StdinPayload
    ↓ hook.ParseEnv() builds Env (CLAUDE_PROJECT_DIR only env var)
    ↓ hook handler (e.g., writeguard, budget, agentguard) processes Env
    ↓ hook.Output struct → JSON on stdout:
       {"hookSpecificOutput": {"permissionDecision": "allow"/"deny"}}
    ↓ CC reads permissionDecision — hooks fail-open (error → allow)
```

### 4. Agent Ask / Search Pipeline

```
User: ari ask "how do I create a session?"
    ↓ internal/cmd/ask.NewAskCmd
    ↓ session.FindCurrentSession() → optional session context (routing enrichment)
    ↓ search.Build(rootCmd, resolver)
       ├── CollectCommands(root) — CLI surface from cobra
       ├── CollectConcepts() — static concept registry
       ├── CollectRites(resolver), CollectAgents(resolver)
       ├── CollectDromena(resolver), CollectRouting(resolver)
       └── CollectParkedSessions(resolver)
    ↓ SearchIndex.Search(query, opts)
       synonym expansion (StaticSynonymSource + OrchestratorSynonymSource)
       TF-IDF-style scoring
    ↓ ranked SearchResult list → output.Printer
```

### 5. Configuration Value Tracing

An agent's `model` value example:
```
rites/{name}/agents/{agent}.md frontmatter (model: claude-opus-4-5)
    OR rites/{name}/manifest.yaml agent_defaults.model
    ↓ materialize.materializeAgents() merge: agent field > agent_defaults
    ↓ agent_transform.go applies write-guard defaults, model override (ElCheapo mode: haiku)
    ↓ writeIfChanged() → .claude/agents/{agent}.md
    ↓ provenance.Collector records checksum
    ↓ PROVENANCE_MANIFEST.yaml written
```

## Knowledge Gaps

1. **`internal/cmd/explain/concepts/`**: The concepts sub-directory was listed but not read. The concept registry contents and schema are undocumented here.

2. **`internal/materialize/hooks/` sub-package**: Files not individually read. The hook defaults generation logic is summarized from the rule file rather than direct code inspection.

3. **`internal/ledge/` domain package**: Only the purpose was confirmed from listing; the `AutoPromote` and `Promote` functions' exact signatures were not read.

4. **`internal/agent/archetype.go` and `internal/agent/regenerate.go`**: Agent transformation internals only summarized. The exact archetype field types and regeneration logic were not inspected.

5. **`internal/session/` FSM transitions**: The `fsm.go` file was not read directly — the 4-state, 5-transition FSM is noted from the rule file but not verified from code.

6. **`internal/hook/clewcontract/`**: The 16-event type JSONL schema and `BufferedEventWriter` implementation were not individually inspected.

7. **`internal/perspective/resolvers.go` and `assemble.go`**: The layer assembly logic was not read — only the `types.go` was inspected to understand the type hierarchy.

8. **Org-scope resolution chain**: The `materialize/orgscope` sub-package and its interaction with the 4-tier resolution order were not directly read; summarized from `org_scope.go` delegation pattern.

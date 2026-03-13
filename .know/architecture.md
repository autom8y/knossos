---
domain: architecture
generated_at: "2026-03-13T10:04:06Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "59a0de2"
confidence: 0.87
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

The codebase is a Go module (`github.com/autom8y/knossos`, Go 1.23+) with a single CLI binary (`ari`) and one entry package at `cmd/ari/`. All domain logic lives under `internal/`.

### Top-Level Layout

| Directory | Purpose |
|-----------|---------|
| `cmd/ari/` | Binary entry point (minimal logic) |
| `internal/cmd/` | Cobra subcommand wiring (one subdirectory per top-level command) |
| `internal/` (domain) | Core domain logic: materialize, session, hook, inscription, paths, etc. |
| `embed.go` | Go embed directives — packages rites/, knossos/templates/, agents/, mena/, processions/ into the binary |
| `rites/` | Rite definitions (manifest.yaml + agents + mena) |
| `knossos/templates/` | CLAUDE.md.tpl and section templates |
| `mena/` | Platform-level mena (dromena and legomena) |
| `agents/` | Cross-rite agent prompt files |
| `processions/` | Cross-rite workflow templates |

### `internal/cmd/` Subpackages (CLI Wiring Layer)

Each subdirectory corresponds to an `ari` subcommand group. These are thin wiring packages: they parse flags, build context, call domain packages, and format output.

| Package | Purpose |
|---------|---------|
| `internal/cmd/root/` | Root cobra command; registers all subcommands; global flags (`--output`, `--verbose`, `--project-dir`, `--session-id`, `--channel`) |
| `internal/cmd/common/` | Shared context types (`BaseContext`, `SessionContext`); embedded asset wiring; output helpers |
| `internal/cmd/session/` | Session lifecycle commands (create, park, resume, wrap, fray, log, query, audit, transition, snapshot, migrate, gc, lock, claim, field, status, suggest_next, timeline) |
| `internal/cmd/sync/` | `ari sync` command; delegates to `materialize.Sync()` |
| `internal/cmd/hook/` | All hook handlers (agentguard, attributionguard, autopark, budget, clew, context, driftdetect, gitconventions, precompact, sessionend, subagent, suggest, validate, worktreeremove, worktreeseed, writeguard) |
| `internal/cmd/manifest/` | `ari manifest` (diff, merge, show, validate) |
| `internal/cmd/inscription/` | `ari inscription` (sync, diff, validate, rollback, backups) |
| `internal/cmd/rite/` | `ari rite` (current, info, invoke, list, pantheon, release, status, validate, context) |
| `internal/cmd/agent/` | `ari agent` (embody, list, new, update, validate) |
| `internal/cmd/artifact/` | `ari artifact` (list, query, rebuild, register) |
| `internal/cmd/worktree/` | `ari worktree` (clone, create, cleanup, export, import, list, remove, status, switch, sync) |
| `internal/cmd/initialize/` | `ari init` |
| `internal/cmd/land/` | `ari land` and `ari land synthesize` |
| `internal/cmd/handoff/` | `ari handoff` (execute, history, prepare, status) |
| `internal/cmd/procession/` | `ari procession` (abandon, create, list, proceed, recede, status) |
| `internal/cmd/knows/` | `ari knows` |
| `internal/cmd/ledge/` | `ari ledge` (list, promote) |
| `internal/cmd/lint/` | `ari lint` (lint_preferential) |
| `internal/cmd/naxos/` | `ari naxos` (scan, triage) |
| `internal/cmd/org/` | `ari org` (current, init, list, set) |
| `internal/cmd/provenance/` | `ari provenance` |
| `internal/cmd/sails/` | `ari sails check` |
| `internal/cmd/status/` | `ari status` |
| `internal/cmd/explain/` | `ari explain` (concepts, context) |
| `internal/cmd/tour/` | `ari tour` (collect) |
| `internal/cmd/tribute/` | `ari tribute generate` |
| `internal/cmd/validate/` | `ari validate` |
| `internal/cmd/complaint/` | `ari complaint list` |
| `internal/cmd/ask/` | `ari ask` (search within ari help) |

**Hub within `internal/cmd/`**: `common` is imported by every other cmd package.

### `internal/` Domain Packages

#### Hub Packages (imported by many)

| Package | Imported by (files) | Purpose |
|---------|---------------------|---------|
| `internal/errors/` | 120 | Domain error types, exit codes (ExitSuccess=0 through ExitSyncNotConfigured=21), error constructors |
| `internal/output/` | 94 | Format-aware output (text/json/yaml), `Printer` struct, `Textable`/`Tabular` interfaces, all output DTO structs |
| `internal/paths/` | 71 | Path resolution; `Resolver` struct; `TargetChannel` interface (ClaudeChannel, GeminiChannel); all XDG/project directory helpers |
| `internal/session/` | 42 | Session FSM, Context struct, status types, phase types, events, timeline, lock protocol |
| `internal/provenance/` | 20 | ProvenanceManifest, ProvenanceEntry, OwnerType, ScopeType; file-level provenance tracking |
| `internal/materialize/` | 7 | Materializer (main pipeline engine); consumed by sync, rite, inscription commands |

#### Domain Logic Packages

| Package | Key Types | Purpose |
|---------|-----------|---------|
| `internal/materialize/` | `Materializer`, `RiteManifest`, `Options`, `Result`, `SyncOptions`, `SyncResult` | Core pipeline: rite source resolution -> generate agents/mena/rules/inscription/settings -> provenance tracking |
| `internal/materialize/source/` | `SourceResolver`, `ResolvedRite`, `RiteSource` | 6-tier rite resolution (explicit > project > user > org > platform > embedded) |
| `internal/materialize/compiler/` | `ChannelCompiler` interface, `GeminiCompiler` | Per-channel compilation pass (Claude = passthrough, Gemini = transform) |
| `internal/materialize/mena/` | `Engine`, `Namespace`, `Walker`, `Transform` | Mena collection, namespace management, .dro/.lego routing to commands/skills |
| `internal/materialize/userscope/` | `UserScopeResult`, `CollisionChecker` | User-level sync (agents, mena, hooks) with collision avoidance |
| `internal/materialize/orgscope/` | (sync.go) | Org-level sync pipeline |
| `internal/materialize/procession/` | `Resolver`, `Renderer` | Cross-rite procession template resolution and rendering |
| `internal/inscription/` | `Manifest`, `Region`, `Marker`, `ParsedRegion`, `OwnerType` | CLAUDE.md templating with region ownership (knossos/satellite/regenerate); 3-owner marker system |
| `internal/session/` | `Context`, `FSM`, `Status`, `Phase`, `Strand`, `Procession` | Session lifecycle (NONE->ACTIVE->PARKED->ARCHIVED FSM); SESSION_CONTEXT.md parsing/serialization; 5 phases |
| `internal/hook/` | `StdinPayload`, `Env`, `HookEvent`, `LifecycleAdapter` | Hook event parsing from stdin JSON; 14 canonical event types; adapter pattern for CC/Gemini |
| `internal/hook/clewcontract/` | `TypedEvent`, `Writer`, `BufferedEventWriter` | Append-only JSONL event log with 16 typed events; thread-safe via mutex |
| `internal/provenance/` | `ProvenanceManifest`, `ProvenanceEntry`, `OwnerType`, `ScopeType` | File-level ownership tracking (knossos/user/untracked); divergence detection; per-channel manifests |
| `internal/agent/` | `FlexibleStringSlice`, `MemoryField`, `Frontmatter` | Agent frontmatter parsing; archetype scaffolding; validation |
| `internal/paths/` | `Resolver`, `TargetChannel`, `ClaudeChannel`, `GeminiChannel` | All path computation; project root discovery (.knossos/ or .claude/.gemini/ walk-up) |
| `internal/resolution/` | `Chain`, `Tier`, `ResolvedItem` | Generic multi-tier resolution (zero internal imports; TENSION-005) |
| `internal/manifest/` | `Manifest`, `Format` | Project manifest loading, diffing, merging |
| `internal/frontmatter/` | `FlexibleStringSlice` | YAML frontmatter parsing shared by agent and materialize packages |
| `internal/mena/` | `StripMenaExtension()`, `RouteMenaFile()` | Mena file type detection (.dro -> commands/, .lego -> skills/) |
| `internal/registry/` | `RefCategory`, `Registry` | Leaf package (zero internal imports); denial-recovery registry for platform references |
| `internal/config/` | `KnossosHome()`, `ActiveOrg()` | Lazy-initialized config (KNOSSOS_HOME, KNOSSOS_ORG env vars); once.Do caching |
| `internal/errors/` | `AriadneError`, 21 exit codes, error code constants | Domain errors with JSON serialization; `IsHandled()`, `GetExitCode()` |
| `internal/output/` | `Printer`, `Format`, `Tabular`, `Textable` | Format-aware output; all output DTO structs |
| `internal/fileutil/` | `WriteIfChanged()`, `AtomicWriteFile()` | Idempotent writes; atomic writes (prevents CC file watcher noise) |
| `internal/channel/` | (channel-related helpers) | Channel-level abstractions (distinct from paths.TargetChannel) |
| `internal/checksum/` | (sha256 helpers) | SHA256 checksums with "sha256:" prefix format |
| `internal/lock/` | `Manager`, `LockMetadata` | Advisory file locking; 5-minute stale threshold |
| `internal/validation/` | `ValidateSessionFields()` | Session schema validation |
| `internal/artifact/` | (registry) | Federated Artifact Registry |
| `internal/know/` | (parsing) | .know/ file parsing |
| `internal/ledge/` | (management) | Ledge artifact management |
| `internal/naxos/` | (triage) | Session hygiene triage |
| `internal/procession/` | (workflow templates) | Cross-rite workflow template types |
| `internal/sails/` | (confidence system) | White Sails confidence signaling (WHITE/GRAY/BLACK) |
| `internal/suggest/` | (suggestions) | Next-action suggestions |
| `internal/perspective/` | (perspective) | Perspective/view helpers |
| `internal/tokenizer/` | (token counting) | Token counting for context budget estimation |
| `internal/tribute/` | (generation) | TRIBUTE.md auto-generation |
| `internal/sync/` | (state management) | Sync state management |
| `internal/worktree/` | (git worktrees) | Worktree lifecycle management |
| `internal/search/` | (search) | Search within ari help content |
| `internal/assets/` | (embedded assets) | Embedded FS management (SetEmbedded, SetUserAssets) |

**Leaf packages** (no internal imports):
- `internal/errors/`, `internal/resolution/`, `internal/registry/`, `internal/checksum/`, `internal/tokenizer/`

---

## Layer Boundaries

The codebase follows a clean 3-layer model:

```
Layer 1: CLI Surface
  cmd/ari/main.go
  internal/cmd/{subcommand}/

Layer 2: Command Wiring
  internal/cmd/root/     (registers all subcommands)
  internal/cmd/common/   (shared context helpers)

Layer 3: Domain Logic
  internal/materialize/  (pipeline hub)
  internal/session/
  internal/hook/
  internal/inscription/
  internal/provenance/
  internal/paths/        (shared utility hub)
  internal/errors/       (leaf)
  internal/output/       (leaf)
  internal/resolution/   (leaf, zero internal imports per TENSION-005)
  internal/registry/     (leaf, zero internal imports)
```

**Import direction: cmd/ -> internal/cmd/ -> internal/\***

### Hub Packages

- `internal/errors/` — imported by 120 files; true leaf (no internal imports)
- `internal/output/` — imported by 94 files; minimal internal dependencies
- `internal/paths/` — imported by 71 files; imports only `internal/errors/` and `internal/config/`
- `internal/session/` — imported by 42 files; imports `internal/errors/`, `internal/fileutil/`, `internal/validation/`

### Key Boundary Rules

1. `provenance` is a leaf with respect to `materialize`: `materialize imports provenance, never the reverse` (per ADR-0026)
2. `resolution.Chain` has zero internal imports; tier paths are injected via constructor to avoid import cycles (TENSION-005)
3. `inscription.OwnerType` and `provenance.OwnerType` are distinct types with different semantics (TENSION-001)
4. `materialize` achieves clean boundaries through sub-packages: `materialize/source`, `materialize/mena`, `materialize/userscope`, `materialize/orgscope`, `materialize/hooks`, `materialize/compiler`, `materialize/procession`

**Notable cross-cutting imports** (Layer 1 -> Layer 1):
- `session` imports `hook/clewcontract` (event emission)
- `tribute` imports `session`, `sails`, `artifact` (all domain)
- `sails` imports `hook/clewcontract` (event log reading)
- `perspective` imports `agent`, `provenance`, `rite`, `frontmatter` (multi-domain assembly)

---

## Entry Points and API Surface

### Binary Entry Point

`cmd/ari/main.go` performs:
1. Sets version info on root command: `root.SetVersion(version, commit, date)`
2. Sets build version for XDG extraction: `common.SetBuildVersion(version)`
3. Wires embedded assets into the global `assets` package via `common.SetEmbeddedAssets(...)`, `common.SetEmbeddedUserAssets(...)`, `common.SetEmbeddedProcessions(...)`
4. Executes root command: `root.Execute()` -> cobra dispatch
5. Handles unhandled errors with format-aware printing; exits with `errors.GetExitCode(err)`

### CLI Subcommands (complete list)

| Subcommand | Package | Purpose |
|------------|---------|---------|
| `ari session` | `internal/cmd/session/` | Session lifecycle (create, park, resume, wrap, fray, log, query, audit, transition, snapshot, migrate, gc, lock, claim, field, status, suggest-next, timeline) |
| `ari manifest` | `internal/cmd/manifest/` | Manifest diff, merge, show, validate |
| `ari inscription` | `internal/cmd/inscription/` | CLAUDE.md sync, diff, validate, rollback, backups |
| `ari sync` | `internal/cmd/sync/` | Full rite+user sync pipeline |
| `ari validate` | `internal/cmd/validate/` | Generic validation |
| `ari handoff` | `internal/cmd/handoff/` | Session handoff (execute, history, prepare, status) |
| `ari procession` | `internal/cmd/procession/` | Cross-rite workflow (abandon, create, list, proceed, recede, status) |
| `ari worktree` | `internal/cmd/worktree/` | Git worktree management |
| `ari hook` | `internal/cmd/hook/` | Hook execution (all hook types) |
| `ari knows` | `internal/cmd/knows/` | Knowledge base query |
| `ari artifact` | `internal/cmd/artifact/` | Artifact registry (list, query, rebuild, register) |
| `ari sails` | `internal/cmd/sails/` | Confidence signal check |
| `ari naxos` | `internal/cmd/naxos/` | Session hygiene triage (scan, triage) |
| `ari rite` | `internal/cmd/rite/` | Rite management (current, info, invoke, list, pantheon, release, status, validate, context) |
| `ari agent` | `internal/cmd/agent/` | Agent management (embody, list, new, update, validate) |
| `ari tribute` | `internal/cmd/tribute/` | Tribute generation |
| `ari init` | `internal/cmd/initialize/` | Project initialization |
| `ari provenance` | `internal/cmd/provenance/` | Provenance inspection |
| `ari org` | `internal/cmd/org/` | Org management (current, init, list, set) |
| `ari land` | `internal/cmd/land/` | Land knowledge pipeline |
| `ari ledge` | `internal/cmd/ledge/` | Ledge artifact management |
| `ari lint` | `internal/cmd/lint/` | Preferential language linting |
| `ari status` | `internal/cmd/status/` | Project/session status |
| `ari explain` | `internal/cmd/explain/` | Explain concepts, context |
| `ari tour` | `internal/cmd/tour/` | Codebase tour |
| `ari ask` | `internal/cmd/ask/` | Search ari help content |
| `ari complaint` | `internal/cmd/complaint/` | Complaint management |
| `ari version` | `internal/cmd/root/` | Version info |

### Global Flags (all subcommands)

- `--output/-o` (text/json/yaml, default: text)
- `--verbose/-v` (JSON lines to stderr)
- `--config` (config file override)
- `--project-dir/-p` (override project root discovery)
- `--session-id/-s` (override current session)
- `--channel` (target channel: claude/gemini/all, default: all)

### Key Exported Interfaces

- `paths.TargetChannel` — implemented by `ClaudeChannel{}` and `GeminiChannel{}`; provides Name(), DirName(), ContextFile(), ContextFilePath(), SkillsDir()
- `output.Textable` — types implement `Text() string` for human-readable output
- `output.Tabular` — types implement `Headers() []string` and `Rows() [][]string` for table output
- `hook.LifecycleAdapter` — `ClaudeAdapter` and `GeminiAdapter` implementations for hook payload parsing
- `materialize/compiler.ChannelCompiler` — per-channel compilation pass interface
- `provenance.Collector` — interface for recording file provenance (2 methods: `Record`, `Entries`)
- `rite.Syncer` — interface `SyncRite(riteName string, keepOrphans bool) error`

---

## Key Abstractions

### 1. `materialize.Materializer` (`internal/materialize/materialize.go`)

The central pipeline engine. Key methods:
- `Sync(SyncOptions) (*SyncResult, error)` — unified entry point dispatching to rite/org/user scopes
- `MaterializeWithOptions(riteName string, opts Options) (*Result, error)` — full rite pipeline (9+ stages)
- `MaterializeMinimal(opts Options) (*Result, error)` — cross-cutting mode (no rite)

Constructed via `NewMaterializer(resolver)` or builder pattern (`WithEmbeddedFS`, `WithEmbeddedTemplates`, etc.).

### 2. `materialize.RiteManifest` (`internal/materialize/materialize.go`)

Central configuration struct read from `rites/{name}/manifest.yaml`. Key fields:
- `Agents []Agent` — agent definitions
- `Dromena []string` — commands to project to commands/
- `Legomena []string` — skills to project to skills/
- `HookDefaults *HookDefaults` — write-guard token budgets
- `SkillPolicies []SkillPolicy` — capability-driven skill wiring rules
- `AgentDefaults map[string]any` — frontmatter merge into all agents
- `MCPServers []MCPServer` — MCP server declarations (written to .mcp.json per SCAR-028)
- `ArchetypeData map[string]map[string]any` — per-archetype template data

### 3. `session.Context` (`internal/session/context.go`)

SESSION_CONTEXT.md YAML frontmatter struct. Schema version 2.3. Key fields:
- `Status Status` — NONE/ACTIVE/PARKED/ARCHIVED
- `CurrentPhase string` — requirements/design/implementation/validation/complete
- `Strands []Strand` — child sessions created via `ari session fray`
- `Procession *Procession` — active cross-rite workflow state (nil when none)

Supports polymorphic YAML deserialization for `Strands` (v2.1/2.2 `[]string` or v2.3+ `[]Strand`).

### 4. `session.FSM` (`internal/session/fsm.go`)

4-state finite state machine: NONE -> ACTIVE -> PARKED -> ARCHIVED (terminal). Valid transitions:
- NONE -> ACTIVE (create)
- ACTIVE -> PARKED (park)
- ACTIVE -> ARCHIVED (wrap)
- PARKED -> ACTIVE (resume)
- PARKED -> ARCHIVED (wrap)

### 5. `provenance.ProvenanceManifest` (`internal/provenance/provenance.go`)

Tracks every file in the channel directory with owner (knossos/user/untracked) and scope (rite/org/user). Stored at `.knossos/PROVENANCE_MANIFEST.yaml`. Key safety invariant: `OwnerUser` files are NEVER overwritten by pipeline.

### 6. `hook.StdinPayload` / `hook.Env` (`internal/hook/env.go`)

**Critical pattern**: CC sends all hook data as JSON on stdin, NOT via environment variables. Only `CLAUDE_PROJECT_DIR` is an env var. `StdinPayload` contains: `session_id`, `tool_name`, `tool_input`, `hook_event_name`, etc. Adapter pattern (`ClaudeAdapter`/`GeminiAdapter`) selected by `KNOSSOS_CHANNEL` env var.

### 7. `resolution.Chain` (`internal/resolution/chain.go`)

Generic multi-tier resolver (zero internal imports; injected paths). `Resolve(name, validate)` for top-down early-exit; `ResolveAll(validate)` for shadow-aware enumeration. Used for rite resolution (project > user > org > platform > embedded) and procession template resolution.

### 8. `paths.Resolver` (`internal/paths/paths.go`)

Central path computation for all project-relative paths. Constructed with a `projectRoot`. Provides: `SessionsDir()`, `KnossosDir()`, `ChannelDir(TargetChannel)`, `AgentsDirForChannel()`, `RiteDir()`, `SessionContextFile()`, etc. `FindProjectRoot()` walks up from CWD looking for `.knossos/` (strongest signal).

### 9. `output.Printer` (`internal/output/output.go`)

Format-aware output dispatching to text/json/yaml. All command output types implement either `Textable` (custom text) or `Tabular` (table).

### 10. `inscription.Generator` / KNOSSOS region system (`internal/inscription/`)

Manages CLAUDE.md sections via `<!-- KNOSSOS:START region-name -->` markers. Three owner types: `knossos` (always overwritten), `satellite` (user-owned, never overwritten), `regenerate` (generated from source).

### Design Patterns

- **Polymorphic YAML fields**: `session.strandList`, `agent.FlexibleStringSlice`, `agent.MemoryField` — custom `UnmarshalYAML` to accept multiple forms
- **Envelope/wrapper pattern**: `materialize.SyncResult` wraps `RiteScopeResult`, `OrgScopeResult`, `UserScopeResult`
- **Builder pattern with method chaining**: `Materializer.WithEmbeddedFS(...).WithEmbeddedTemplates(...)`
- **Cascade/merge pattern**: `agent_defaults` (manifest level) -> agent frontmatter; `hook_defaults` (shared -> rite) merge
- **DI via constructor**: `NewMaterializerWithSourceResolver(resolver, sr)` enables test injection
- **Registry with shadowing**: Higher-priority tiers shadow lower-priority in `resolution.Chain.ResolveAll()`
- **Provenance collector**: Thread-safe collector threaded through pipeline stages; `NullCollector` for dry-run
- **Idempotency pattern**: `writeIfChanged()` reads existing file before writing — skips write if content is identical

---

## Data Flow

### 1. Sync Pipeline (Primary: rite source -> `.claude/`)

```
rites/{name}/manifest.yaml
    | source resolution (materialize/source.SourceResolver)
       checks: explicit > project > user > org > platform > embedded
    | RiteManifest deserialization
    | Materializer.Sync(SyncOptions)
       |-- materializeAgents() -> .claude/agents/*.md
       |     agent frontmatter + agent_defaults merge
       |     archetype rendering (orchestrator.md.tpl)
       |     skill policies applied (SkillPolicy)
       |     provenance.Collector records each file
       |-- materializeMena() -> .claude/commands/ and .claude/skills/
       |     .dro.md files -> commands/, .lego.md files -> skills/
       |     StripMenaExtension, RouteMenaFile
       |-- materializeRules() -> .claude/rules/*.md
       |-- materializeInscription() -> .claude/CLAUDE.md
       |     inscription.SyncCLAUDEmd(): render templates -> merge regions
       |     preserves satellite regions, overwrites knossos regions
       |-- materializeSettingsWithManifest() -> .claude/settings.local.json (hooks)
       |-- materializeMcpJson() -> .mcp.json (MCP servers, per SCAR-028)
       |-- trackState() -> .knossos/sync/state.json
       |-- writeActiveRite() -> .knossos/ACTIVE_RITE
       --- provenance.Collector.Save() -> .knossos/PROVENANCE_MANIFEST.yaml
```

Configuration merge points within the pipeline:
- Agent frontmatter: explicit fields > `agent_defaults` in manifest > archetype defaults
- Mena: 4-tier source resolution (project > dependency > shared > user)
- MCP servers: written to `.mcp.json` (project root), union merge
- CLAUDE.md: knossos regions overwritten; satellite regions preserved; `regenerate` regions re-rendered

### 2. Session Event Pipeline

```
CC lifecycle event (SessionStart, Stop, PreToolUse, etc.)
    | CC sends JSON to ari hook subcommand stdin
    | hook.ParseEnv() -> hook.Env (session ID, event type, tool info)
    | internal/cmd/hook/{context,autopark,budget,clew,...}.go
    | session.Context.Load() from SESSION_CONTEXT.md
    | session state mutation (transition, field update, etc.)
    | hook/clewcontract.BufferedEventWriter.Emit()
       -> appends typed event to .sos/sessions/{id}/events.jsonl
    | hook output (hookSpecificOutput envelope with permissionDecision for CC)
```

The events.jsonl file supports 3 format versions (v1 legacy `Event`, v2 flat `ClewEvent`, v3 typed `TypedEvent`) — interleaved per ADR-0027 backward compat bridge in `internal/session/events_read.go`.

### 3. Hook Pipeline (CC integration)

```
Claude Code tool use
    | CC sends {"hook_event_name": "PreToolUse", "tool_name": "...", ...} to stdin
    | hook.parseStdin() -> StdinPayload
    | hook.ParseEnv() builds Env (CLAUDE_PROJECT_DIR only env var)
    | hook handler (e.g., writeguard, budget, agentguard) processes Env
    | hook.Output struct -> JSON on stdout:
       {"hookSpecificOutput": {"permissionDecision": "allow"/"deny"}}
    | CC reads permissionDecision — hooks fail-open (error -> allow)
```

### 4. Configuration Value Tracing

An agent's `model` value example:
```
rites/{name}/agents/{agent}.md frontmatter (model: claude-opus-4-5)
    OR rites/{name}/manifest.yaml agent_defaults.model
    | materialize.materializeAgents() merge: agent field > agent_defaults
    | agent_transform.go applies write-guard defaults, model override
    | writeIfChanged() -> .claude/agents/{agent}.md
    | provenance.Collector records checksum
```

---

## Knowledge Gaps

1. **`internal/channel/` package**: Files discovered but content not read in detail; its relationship to `paths.TargetChannel` was not fully traced.

2. **`internal/suggest/`, `internal/perspective/`, `internal/search/`**: Package purpose noted from doc comments but internal structures not examined.

3. **`internal/materialize/compiler/`**: Gemini compiler transformation logic not read in detail (only confirmed the `ChannelCompiler` interface and that Gemini requires a compile pass, Claude is passthrough).

4. **`internal/materialize/mena/content_rewrite.go`**: Mena content rewrite logic not fully examined.

5. **`internal/sync/` (root-level, not materialize/sync)**: Sync state management package content not read.

6. **`internal/naxos/` triage logic**: Naxos triage scoring not examined in detail.

7. **The `ari ask` search index**: How `internal/cmd/ask/` builds and queries the help index not traced.

8. **Worktree seed/clone mechanics**: `internal/cmd/worktree/` subcommands and `internal/worktree/` domain package not fully read.

9. **Org-scope resolution chain**: The `materialize/orgscope` sub-package and its interaction with the 4-tier resolution order were not directly read.

10. **`internal/hook/clewcontract/`**: The 16-event type JSONL schema and `BufferedEventWriter` implementation were not individually inspected.

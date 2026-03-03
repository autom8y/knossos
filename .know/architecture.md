---
domain: architecture
generated_at: "2026-03-03T19:45:00Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "1599813"
confidence: 0.87
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Architecture

## Package Structure

### Language and Module

- **Language**: Go 1.23+ (toolchain go1.24.13)
- **Module**: `github.com/autom8y/knossos`
- **Binary**: `ari` (Ariadne)
- **Entry point**: `cmd/ari/main.go` (single binary, single entry point)

### Top-Level Directory Layout

| Directory | Role |
|-----------|------|
| `cmd/ari/` | Binary entry point (minimal wiring only) |
| `internal/` | All application logic (Go convention — not importable externally) |
| `internal/cmd/` | Cobra command implementations (one sub-package per command group) |
| `internal/<domain>/` | Domain logic packages (29 packages total) |
| `rites/` | Embedded rite definitions (agents, mena, manifests) |
| `mena/` | Platform mena source (dromena + legomena) |
| `agents/` | Cross-rite agent definitions |
| `knossos/templates/` | CLAUDE.md section templates |
| `config/` | Default config files (e.g., hooks.yaml) |
| `embed.go` | `//go:embed` declarations for single-binary distribution |

### `internal/` Package Inventory

**Leaf packages** (no internal imports — safe to import anywhere):

| Package | Files | Purpose |
|---------|-------|---------|
| `internal/errors` | 2 | Structured error types with exit codes and JSON serialization |
| `internal/checksum` | 2 | SHA-256 file/bytes hashing with `sha256:` prefix convention |
| `internal/fileutil` | 2 | Atomic file writes (`AtomicWriteFile`), `WriteIfChanged` |
| `internal/frontmatter` | 3 | YAML frontmatter parsing from markdown (handles `\n` and `\r\n`) |
| `internal/tokenizer` | 2 | Token estimation via tiktoken |
| `internal/registry` | 4 | Denial-recovery registry: maps stable keys to agent/skill/CLI refs |
| `internal/config` | 2 | `KnossosHome()` — resolves `$KNOSSOS_HOME` with fallback |
| `internal/assets` | 1 | Embedded asset accessor |

**Near-leaf packages** (import only leaf packages):

| Package | Files | Internal imports | Purpose |
|---------|-------|-----------------|---------|
| `internal/paths` | 2 | `errors` | Project root discovery, `Resolver` type, all path resolution |
| `internal/provenance` | 7 | none | `ProvenanceManifest` type — file-level ownership tracking for `.claude/` |
| `internal/mena` | 8 | none | `.dro.md`/`.lego.md` type detection and routing logic |
| `internal/lock` | 4 | `errors` | Advisory file locking for session operations |

**Domain packages** (import leaf + near-leaf):

| Package | Files | Key Types | Purpose |
|---------|-------|-----------|---------|
| `internal/manifest` | 8 | `Manifest`, `Format` | Generic YAML/JSON manifest loading, diffing, 3-way merging |
| `internal/inscription` | 15 | `Pipeline`, `Manifest`, `MarkerParser`, `Merger`, `Generator` | CLAUDE.md sync: marker parsing, template rendering, region merging |
| `internal/session` | 22 | `Context`, `Status`, `FSM`, `Event`, `ClewEvent` | Session lifecycle (NONE→ACTIVE→PARKED→ARCHIVED), event log reading |
| `internal/hook` | 5 | `StdinPayload`, `Env`, `ToolInput`, `HookEvent` | Hook infrastructure: CC stdin parsing, event types |
| `internal/rite` | 16 | `RiteManifest`, `RiteForm`, `BudgetCalculator` | Rite manifest parsing, discovery, budget estimation, invocation |
| `internal/artifact` | 8 | `Registry`, `Entry`, `Querier`, `QueryFilter` | Work artifact registry (ADR/PRD/TDD) with phase/type filtering |
| `internal/sails` | 13 | `ContractViolation` | White Sails confidence signaling; clew contract validation |
| `internal/tribute` | 7 | `GenerateResult`, `Artifact`, `Decision` | TRIBUTE.md auto-generation at session wrap |
| `internal/naxos` | 5 | `OrphanReason`, `SuggestedAction` | Abandoned session cleanup tooling (session garbage collection) |
| `internal/worktree` | 8 | `Worktree`, `WorktreeMetadata` | Git worktree management for parallel sessions |
| `internal/validation` | 9 | `Validator` | JSON schema validation for artifacts (ADR, PRD, TDD, etc.) |
| `internal/know` | 8 | `Meta`, `DomainStatus` | `.know/` file frontmatter parsing; domain freshness tracking |
| `internal/sync` | 2 | `State`, `StateManager` | `.knossos/sync/state.json` read/write; sync timestamp tracking |
| `internal/agent` | 16 | `ScaffoldData`, `Archetype`, `Scaffold` | Agent scaffolding from archetype templates |

**Hub package** (orchestrates many domain packages):

| Package | Files | Internal imports (partial) | Purpose |
|---------|-------|-----------------------------|---------|
| `internal/materialize` | 48 | `errors`, `fileutil`, `paths`, `provenance`, `registry`, + sub-packages | Full materialization pipeline: rite source → `.claude/` output |

**Materialize sub-packages:**

| Sub-package | Purpose |
|-------------|---------|
| `internal/materialize/source` | `SourceType`, `ResolvedRite` — 6-tier rite resolution (project, user, org, knossos, explicit, embedded) |
| `internal/materialize/mena` | Mena materialization engine: collect, namespace, write `.dro.md`/`.lego.md` → commands/skills |
| `internal/materialize/hooks` | Hook config materialization |
| `internal/materialize/userscope` | User-level `.claude/` sync (cross-rite agents, user mena) |
| `internal/materialize/orgscope` | Org-level resource sync |

**Output package** (presentation layer):

| Package | Purpose |
|---------|---------|
| `internal/output` | Format-aware printing: text/JSON/YAML; domain-specific output structs |

**`internal/cmd/` sub-packages** (CLI surface):

| Package | Files | Command group |
|---------|-------|---------------|
| `cmd/root` | 1 | Root command, global flags, command wiring |
| `cmd/session` | 35 | `ari session {create,status,list,park,resume,wrap,transition,fray,gc,audit,lock,snapshot,timeline,...}` |
| `cmd/hook` | 28 | `ari hook {write-guard,clew,context,autopark,session-end,agentguard,budget,subagent,gitconventions,...}` |
| `cmd/sync` | 3 | `ari sync` — invokes materialize pipeline |
| `cmd/inscription` | 6 | `ari inscription {sync,diff,validate,rollback}` |
| `cmd/rite` | 11 | `ari rite {list,current,validate,invoke,context,release}` |
| `cmd/manifest` | 5 | `ari manifest {show,validate,diff,merge}` |
| `cmd/artifact` | 5 | `ari artifact {register,list,query,rebuild}` |
| `cmd/worktree` | 12 | `ari worktree {create,sync,remove,cleanup}` |
| `cmd/knows` | 2 | `ari knows` — `.know/` domain status |
| `cmd/agent` | 6 | `ari agent` — agent scaffolding |
| `cmd/sails` | 3 | `ari sails` — quality gate signaling |
| `cmd/naxos` | 2 | `ari naxos` — orphan session scan |
| `cmd/tribute` | 2 | `ari tribute` — TRIBUTE.md generation |
| `cmd/handoff` | 6 | `ari handoff` — specialist handoff workflow |
| `cmd/org` | 6 | `ari org {init,list,current,set}` — org management |
| `cmd/provenance` | 1 | `ari provenance` — provenance manifest operations |
| `cmd/lint` | 2 | `ari lint` — mena/agent linting |
| `cmd/status` | 2 | `ari status` — project status overview |
| `cmd/explain` | 5 | `ari explain` — platform concept documentation |
| `cmd/tour` | 4 | `ari tour` — onboarding walkthrough |
| `cmd/validate` | 2 | `ari validate` — artifact schema validation |
| `cmd/initialize` | 2 | `ari init` — project initialization |
| `cmd/common` | 3 | Shared `BaseContext`, `SessionContext`, `NeedsProject` annotation |

## Layer Boundaries

### Layer Model

```
cmd/ari/main.go              (CLI entry — wiring only, ~32 lines)
    ↓
internal/cmd/root/root.go    (command tree assembly, global flags, project discovery)
    ↓
internal/cmd/<group>/        (command handlers — parse flags, call domain logic, format output)
    ↓
internal/<domain>/           (domain logic — no Cobra, no output formatting)
    ↓
internal/{errors,paths,fileutil,checksum,frontmatter,tokenizer,registry,config}  (leaf utilities)
```

**Import direction rule**: packages flow strictly downward. Domain packages never import `internal/cmd/`. The `internal/cmd/` packages import domain packages, not each other (except via `internal/cmd/common`).

### Boundary Enforcement Patterns

1. **Provenance as leaf**: `internal/provenance` has zero internal imports despite being central to the pipeline. It uses plain strings (`SourceType`) to avoid importing `internal/materialize/source`. This deliberate tension is documented as TENSION-007.

2. **Registry as leaf**: `internal/registry` imports only stdlib. It holds stable keys for agents, skills, and CLI commands so any package can reference them without circular imports.

3. **cmd/common as shared base**: All `internal/cmd/<group>` packages embed `common.BaseContext` (or `common.SessionContext`). This prevents global state from leaking between command handlers.

4. **Annotation-based project gating**: `cmd/root` calls `common.NeedsProject(cmd)` which reads a cobra annotation (`needsProject`) set by each command group. This avoids a direct registry of project-required commands in the root package.

### Hub and Leaf Classification

**Leaf packages** (no internal imports): `errors`, `checksum`, `fileutil`, `frontmatter`, `tokenizer`, `registry`, `config`, `assets`, `provenance`, `mena`

**Near-leaf** (import only leaf): `paths`, `lock`

**Domain hubs** (import multiple peers): `materialize` (imports `fileutil`, `paths`, `provenance`, `registry`, `sync`, `inscription`, `mena`, `agent`, `manifest`, + sub-packages), `session` (imports `errors`, `fileutil`, `validation`), `sails` (imports `hook/clewcontract`)

**CLI hubs** (import many domain packages): `cmd/hook` (imports `hook`, `hook/clewcontract`, `know`, `materialize/source`, `session`, `frontmatter`, `lock`, `registry`), `cmd/session` (imports `session`, `lock`, `hook/clewcontract`, `paths`)

**The `materialize` package is the dominant hub**: it orchestrates the entire sync pipeline by importing and coordinating nearly all domain packages. It is the only package where this breadth is acceptable by design.

## Entry Points and API Surface

### Binary Entry Point Trace

`cmd/ari/main.go:22` (`func main()`) performs four operations:
1. `root.SetVersion(version, commit, date)` — injects build-time version strings
2. `common.SetBuildVersion(version)` — makes version available to command handlers
3. `common.SetEmbeddedAssets(knossos.EmbeddedRites, knossos.EmbeddedTemplates, knossos.EmbeddedHooksYAML)` — injects embedded FS from root `embed.go`
4. `common.SetEmbeddedUserAssets(knossos.EmbeddedAgents, knossos.EmbeddedMena)` — injects user-scope embedded assets
5. `root.Execute()` — runs Cobra command tree

The `embed.go` at the module root embeds five asset sets at compile time:
- `EmbeddedRites` (`//go:embed rites`) — all rite definitions
- `EmbeddedTemplates` (`//go:embed knossos/templates`)
- `EmbeddedHooksYAML` (`//go:embed config/hooks.yaml`)
- `EmbeddedAgents` (`//go:embed agents`) — cross-rite agents
- `EmbeddedMena` (`//go:embed mena`) — platform mena

### Cobra Command Tree

Root command: `ari` (`internal/cmd/root/root.go`)

Global persistent flags: `--output/-o` (text/json/yaml), `--verbose/-v`, `--config`, `--project-dir/-p`, `--session-id/-s`

`PersistentPreRunE` on root: validates output format, loads viper config, runs `paths.FindProjectRoot()` for project discovery.

**All subcommands registered in `root.go init()`:**

| Command | Purpose |
|---------|---------|
| `ari session` | Session lifecycle: `create`, `park`, `resume`, `wrap`, `status`, `list`, `transition`, `fray`, `gc`, `audit`, `lock`, `snapshot`, `timeline`, `log`, `field`, `migrate`, `recover`, `context snapshot` |
| `ari sync` | Run materialize pipeline (`ari sync [--rite NAME] [--scope all|rite|user]`) |
| `ari inscription` | CLAUDE.md sync: `sync`, `diff`, `validate`, `rollback` |
| `ari manifest` | Manifest ops: `show`, `validate`, `diff`, `merge` |
| `ari rite` | Rite management: `list`, `current`, `validate`, `invoke`, `context`, `release` |
| `ari hook` | Hook handlers: `write-guard`, `clew`, `context`, `autopark`, `session-end`, `agentguard`, `budget`, `subagent`, `git-conventions`, `precompact`, `validate`, `worktree-remove`, `worktree-seed`, `cheapo-revert` |
| `ari artifact` | Work artifact registry: `register`, `list`, `query`, `rebuild` |
| `ari worktree` | Git worktree management: `create`, `sync`, `remove`, `cleanup` |
| `ari knows` | `.know/` domain freshness status |
| `ari agent` | Agent scaffolding |
| `ari sails` | Quality gate (White Sails) |
| `ari naxos` | Orphan session scanner |
| `ari tribute` | TRIBUTE.md generation |
| `ari handoff` | Specialist handoff workflow |
| `ari org` | Org-level resource management: `init`, `list`, `current`, `set` |
| `ari provenance` | Provenance manifest inspection |
| `ari lint` | Mena/agent lint |
| `ari status` | Project status overview |
| `ari explain` | Platform concept documentation |
| `ari tour` | Onboarding walkthrough |
| `ari validate` | Artifact schema validation |
| `ari init` | Project initialization |
| `ari version` | Version info |

### Key Exported Interfaces

**`internal/materialize.Materializer`**: Central type consumed by `internal/cmd/sync`. Methods: `Sync(SyncOptions)`, `MaterializeWithOptions(riteName, Options)`, `MaterializeMinimal(Options)`. Chains: `WithEmbeddedFS()`, `WithEmbeddedTemplates()`, `WithEmbeddedAgents()`, `WithEmbeddedMena()`.

**`internal/paths.Resolver`**: Consumed by virtually all domain packages. Provides all path resolution: `ClaudeDir()`, `SessionsDir()`, `KnossosDir()`, `RiteDir()`, `AgentsDir()`, etc.

**`internal/session.Context`**: The SESSION_CONTEXT.md parsed representation. Fields: `SessionID`, `Status`, `Initiative`, `Complexity`, `ActiveRite`, `CurrentPhase`. Consumed by `cmd/session/*`, `cmd/hook/*`.

**`internal/inscription.Pipeline`**: Consumed by `cmd/inscription`. Methods: `Sync()`, `DryRun()`, `Validate()`, `Rollback()`, `GetDiff()`, `ListBackups()`.

**`internal/hook.StdinPayload`**: The canonical struct for CC hook data. Fields: `SessionID`, `HookEventName`, `ToolName`, `ToolInput`, `ToolResponse`, `CWD`. Consumed by `cmd/hook/*` via `hook.ParseEnv()`.

**`internal/errors.Error`**: The universal error type. Fields: `Code string`, `Message string`, `Details map`, `ExitCode int`. Constructor: `errors.Wrap(code, msg, cause)`. All domain packages return `*errors.Error`.

## Key Abstractions

### Core Types

**`internal/materialize.RiteManifest`** (`internal/materialize/materialize.go:59`)
The rite's manifest.yaml parsed form. Fields include: `Name`, `EntryAgent`, `Agents []Agent`, `Dromena []string`, `Legomena []string`, `Hooks []string`, `MCPServers []MCPServer`, `HookDefaults`, `AgentDefaults map[string]interface{}`, `SkillPolicies []SkillPolicy`, `ArchetypeData`. This is the central configuration type driving materialization. Note: `internal/rite.RiteManifest` is a parallel type used by the rite invocation domain — both parse `manifest.yaml` but with different field sets for different purposes.

**`internal/provenance.ProvenanceManifest`** (`internal/provenance/provenance.go:26`)
Tracks ownership of all files in `.claude/`. Stored at `.claude/PROVENANCE_MANIFEST.yaml`. Entries map relative paths to `ProvenanceEntry` records with `Owner` (knossos/user/untracked), `Scope` (rite/user/org), `SourcePath`, `SourceType`, and `Checksum`. This is the safety gate preventing user content from being overwritten.

**`internal/inscription.Manifest`** (`internal/inscription/manifest.go`)
The KNOSSOS_MANIFEST.yaml structure: tracks region ownership in CLAUDE.md. Distinct from `internal/manifest.Manifest` (generic) and `internal/provenance.ProvenanceManifest` (file ownership). `OwnerType` here has different semantics: `knossos`/`satellite`/`regenerate` (region content ownership vs. file ownership).

**`internal/session.Context`** (`internal/session/context.go:16`)
SESSION_CONTEXT.md parsed state. Includes `FrayedFrom`/`Strands` for session forking. Status FSM: `FSM.ValidateTransition(from, to)` in `internal/session/fsm.go` enforces NONE→ACTIVE→{PARKED,ARCHIVED} per TLA+ spec.

**`internal/hook.StdinPayload`** (`internal/hook/env.go:48`)
The canonical hook data contract from CC. Fields map directly to what CC sends as stdin JSON. `ParseEnv()` merges stdin JSON (primary) with environment variable fallbacks (legacy). This dual-read exists for backward compatibility.

**`internal/materialize/source.ResolvedRite`** (`internal/materialize/source/types.go:36`)
The result of 6-tier rite resolution. `Source.Type` is one of: `project`, `user`, `knossos`, `org`, `explicit`, `embedded`. The tier precedence is documented as project > user > org > knossos > embedded.

**`internal/errors.Error`** (`internal/errors/errors.go:66`)
Universal error type with structured `Code`, `Message`, `Details`, and `ExitCode`. The `exitCodeForCode()` function maps 25+ error codes to 21+ distinct exit codes (0–21). All errors use `errors.Wrap(code, msg, cause)` or named constructors (`errors.ErrSessionNotFound()`, etc.).

### Design Patterns

**Polymorphic YAML via dual manifest structs**: `manifest.yaml` is parsed differently by `internal/materialize.RiteManifest` (full pipeline fields) vs `internal/rite.RiteManifest` (invocation/lifecycle fields). Both exist in the codebase without consolidation — a known tension.

**Inscription marker system**: CLAUDE.md uses HTML comment markers `<!-- KNOSSOS:START {region} -->` / `<!-- KNOSSOS:END {region} -->` to delimit knossos-owned vs satellite-owned sections. `MarkerParser` in `internal/inscription/marker.go` parses these via regex. The `OwnerType` (`knossos`/`satellite`/`regenerate`) controls whether regions are overwritten on sync.

**Cascade/merge pattern for hook defaults and skill policies**: Shared rite-level defaults are loaded first, then rite-level overrides are merged. `ResolveHookDefaults(sharedHookDefaults, manifest.HookDefaults)` in `internal/materialize/materialize.go:421`. Same pattern for `MergeSkillPolicies`.

**WriteIfChanged idempotency**: All file writes go through `fileutil.WriteIfChanged()` which skips disk writes when content is identical. This makes `ari sync` safe to run repeatedly (documented in `materialize.go:304`).

**Embedded FS fallback chain**: The binary embeds all rites, templates, agents, and mena. `SourceResolver.ResolveRite()` walks the 6-tier chain and falls back to embedded assets when filesystem sources are unavailable. This enables zero-install usage.

**Event log format bridge** (`internal/session/events_read.go:1–27`): Events.jsonl files can contain pre-ADR-0027 (`Event`) and post-ADR-0027 (`ClewEvent`, `TypedEvent`) format entries. `ReadEvents()` format-sniffs each line using presence of the `"data"` field. The write path is unified through `clewcontract.BufferedEventWriter`.

### Naming Conventions

- **Rites**: composable practice bundles stored under `rites/<name>/`
- **Dromena** (`.dro.md`): transient slash commands → `.claude/commands/`
- **Legomena** (`.lego.md`): persistent reference skills → `.claude/skills/`
- **Mena**: collective term for dromena + legomena
- **Inscription**: the CLAUDE.md sync system (marker parsing + template rendering)
- **Naxos**: orphan/cleanup tooling (mythological: where Theseus abandoned Ariadne)
- **Sails**: quality gate signaling (White/Gray/Black sails = confidence levels)
- **Tribute**: session wrap artifact (TRIBUTE.md)
- **Clew**: orchestrator handoff protocol (mythological: Ariadne's thread)

## Data Flow

### Sync Pipeline (Primary Path)

```
User invokes: ari sync [--rite ecosystem] [--scope all]
    ↓
internal/cmd/sync/sync.go
    → resolves project root via paths.FindProjectRoot()
    → constructs Materializer with embedded FSes
    → calls Materializer.Sync(SyncOptions)
    ↓
internal/materialize/materialize.go: Sync()
    → Phase 1 (rite scope): syncRiteScope()
        → sourceResolver.ResolveRite(riteName)         [6-tier resolution]
        → loads manifest.yaml → RiteManifest
        → validateRiteReferences()                      [warn on stale refs]
        → loadRiteManifest() → RiteManifest
        → provenance.LoadOrBootstrap()                  [divergence detection]
        → materializeAgents()                           [rite agents → .claude/agents/]
        → materializeMena()                             [mena → commands/ + skills/]
        → materializeRules()                            [templates/rules → .claude/rules/]
        → materializeCLAUDEmd()                         [inscription pipeline]
        → materializeSettingsWithManifest()             [settings.local.json + MCP]
        → trackState()                                  [.knossos/sync/state.json]
        → materializeWorkflow()                         [ACTIVE_WORKFLOW.yaml]
        → writeActiveRite()                             [.claude/ACTIVE_RITE]
        → saveProvenanceManifest()                      [.claude/PROVENANCE_MANIFEST.yaml]
    → Phase 2 (user scope): syncUserScope()
        → syncs cross-rite agents to ~/.claude/agents/
        → syncs user mena to ~/.claude/skills/ + commands/
```

### CLAUDE.md Inscription Pipeline

```
materializeCLAUDEmd() calls inscription.SyncCLAUDEmd()
    ↓
internal/inscription/pipeline.go: Sync()
    → buildRenderContext()              [loads active rite + agents metadata]
    → generator.GenerateAll()           [renders templates per section]
    → merger.Merge(existing, generated) [merges knossos regions, preserves satellite]
    → writes result to .claude/CLAUDE.md
    → updates KNOSSOS_MANIFEST.yaml
```

Templates live in `knossos/templates/sections/*.md.tpl` and use Go's `text/template` with Sprig functions.

### Session Pipeline

```
CC fires SessionStart hook → ari hook clew → session.ReadContext()
    → reads .sos/sessions/<id>/SESSION_CONTEXT.md
    → validates status via session.FSM

ari session create "initiative" -c MODULE
    → acquires lock via internal/lock
    → creates .sos/sessions/<id>/SESSION_CONTEXT.md
    → emits session_start event to events.jsonl via clewcontract.BufferedEventWriter

Events flow: hook fires → ari hook <handler>
    → hook.ParseEnv() reads stdin JSON (StdinPayload)
    → event written to .sos/sessions/<id>/events.jsonl

ari session wrap
    → validates FSM transition (ACTIVE→ARCHIVED)
    → generates TRIBUTE.md via tribute.Generator
    → archives session directory
```

### Hook Pipeline

```
CC fires hook event (PreToolUse, PostToolUse, Stop, etc.)
    ↓
stdin JSON → hook.ParseEnv() → hook.Env struct
    → paths.FindProjectRoot(env.ProjectDir)
    → per-hook logic (write-guard, agentguard, budget, etc.)
    ↓
Exit codes control CC behavior:
    0 = proceed
    1 = block (write-guard, agentguard)
    other = logged warning
```

Key hook handlers in `internal/cmd/hook/`:
- `write-guard`: prevents writes to knossos-owned regions
- `agentguard`: enforces agent tool restrictions
- `clew`: records handoff events to events.jsonl
- `context`: injects `.know/` freshness status into context
- `autopark`: auto-parks session on Stop event
- `session-end`: finalizes session on CC SessionEnd event
- `budget`: tracks context token budget

### Configuration Merge Chain

For agent frontmatter (AgentDefaults cascade):
```
Shared shared/agents/defaults.yaml
    ↑ merged by
Rite manifest.yaml agent_defaults
    ↑ merged by
Individual agent frontmatter (name, role, description, tools, model, color)
```

For skill policies:
```
Shared shared/skill_policies.yaml (if exists)
    ↑ merged by MergeSkillPolicies()
Rite manifest.yaml skill_policies
    → evaluated per-agent during materializeAgents()
```

Source resolution cascade for rites (6 tiers, first found wins):
```
1. --source flag (explicit path or "knossos")
2. .knossos/rites/<name>/  (project satellite)
3. ~/.local/share/knossos/rites/<name>/  (user)
4. $XDG_DATA_HOME/knossos/orgs/<org>/rites/<name>/  (org)
5. $KNOSSOS_HOME/rites/<name>/  (knossos platform)
6. binary embedded rites/  (fallback)
```

## Knowledge Gaps

1. **`internal/cmd/rite/invoke.go` invocation flow**: The `ari rite invoke` command was identified but the full invocation state machine (INVOCATION_STATE.yaml lifecycle) was not fully traced. `internal/rite/invoker.go` and `internal/rite/state.go` were not read.

2. **`internal/artifact` Registry full schema**: The `Entry` type and `ArtifactType`/`Phase` enum values were not read — only `QueryFilter` and `Querier` were examined.

3. **`internal/materialize/userscope` sync detail**: The user-scope sync logic in `sync.go` was not read in full — only the entry point delegation from `Materializer.Sync()` was observed.

4. **`internal/hook/clewcontract` full API**: Only `orchestrator.go` (Throughline extraction) and the event write path was referenced. The full `BufferedEventWriter` API and `TypedEvent` schema were not read.

5. **`internal/validation` schema set**: The 7 JSON schemas in `internal/validation/schemas/` were listed but not read. The `Validator` type interface was only partially traced.

6. **`internal/org` scope**: The `org.go` in `internal/cmd/org/` and corresponding org-scope materialization were not read in depth.

7. **`internal/sails` proofs and thresholds**: The White/Gray/Black sail color logic and threshold configuration were not read — only `ContractViolation` types were examined.

8. **`internal/know/astdiff.go`**: The AST-based semantic diffing (most recent commit `1599813`) was not examined. This is a new feature for incremental `.know/` refresh.

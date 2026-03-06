---
domain: architecture
generated_at: "2026-03-06T12:00:00Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "3847e28"
confidence: 0.82
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/initiative-history.md"
land_hash: "d2ffa6c6e09a8852bf5f5169eb68f03dd80a73cf4b057a217bbe9a734503e94b"
---

# Codebase Architecture

## Package Structure

The knossos project (`github.com/autom8y/knossos`) is a Go 1.23 CLI tool named `ari` (Ariadne). The module root houses `embed.go`, which embeds rites, templates, hooks config, agents, and mena as `embed.FS` for single-binary distribution.

### Top-Level Layout

```
cmd/ari/         — Entry point (main.go)
internal/        — All domain logic (31 packages)
internal/cmd/    — CLI command wiring (26 sub-packages, one per command group)
internal/<domain>/  — Domain logic packages (29 packages total)
rites/           — Embedded rite definitions
knossos/         — Templates for CLAUDE.md sections
agents/          — Cross-rite agent definitions (embedded)
mena/            — Platform mena (embedded)
config/          — Bootstrap hooks.yaml (embedded)
```

### `cmd/ari/` (1 file)

**Purpose**: Minimal entry point. Sets version info, wires embedded assets into `common`, then calls `root.Execute()`. Contains no business logic.

**Key action**: `common.SetEmbeddedAssets(knossos.EmbeddedRites, knossos.EmbeddedTemplates, knossos.EmbeddedHooksYAML)` and `common.SetEmbeddedUserAssets(knossos.EmbeddedAgents, knossos.EmbeddedMena)`.

### `internal/cmd/` — Command Wiring Layer (26 packages)

Each sub-package exposes a `New*Cmd(...)` constructor returning a `*cobra.Command`. All commands receive pointer arguments from root (`*string`, `*bool`) for global flags, not copies.

| Package | Use | Sub-commands |
|---|---|---|
| `root` | Root `ari` command, PersistentPreRunE for config/project discovery | version + all subcommands |
| `session` | Workflow session lifecycle | create, status, list, park, resume, wrap, transition, migrate, audit, recover, fray, claim, lock, unlock, gc, field-set, field-get, log, timeline, context, query |
| `manifest` | Manifest inspection and diffing | (sub-commands) |
| `inscription` | CLAUDE.md synchronization | (sub-commands) |
| `sync` | Rite materialization trigger | (sub-commands) |
| `validate` | Artifact schema validation | (sub-commands) |
| `handoff` | Agent handoff management | (sub-commands) |
| `hook` | Claude Code hook infrastructure | agent-guard, attribution-guard, auto-park, budget, cheapo-revert, clew, context, git-conventions, precompact, session-end, subagent, validate, write-guard, worktree-remove, worktree-seed |
| `knows` | `.know/` knowledge inspection | (delta, list) |
| `artifact` | Session artifact management | (sub-commands) |
| `sails` | White Sails confidence gate | (sub-commands) |
| `naxos` | Orphaned session cleanup | (sub-commands) |
| `rite` | Rite discovery and management | (sub-commands) |
| `agent` | Agent file management | (sub-commands) |
| `tribute` | TRIBUTE.md generation | (sub-commands) |
| `initialize` | `ari init` project bootstrapping | — |
| `provenance` | Provenance manifest inspection | (sub-commands) |
| `org` | Org-level resource management | (sub-commands) |
| `land` | Cross-session knowledge synthesis | (sub-commands) |
| `ledge` | `.ledge/` work product management | (sub-commands) |
| `lint` | Source validation (rites, agents) | (sub-commands) |
| `status` | Unified project health dashboard | — |
| `explain` | Embedded documentation viewer | — |
| `tour` | Interactive onboarding | — |
| `worktree` | Git worktree management | (sub-commands) |
| `common` | Shared context types (`BaseContext`, `SessionContext`), embedded asset accessors | — |

**`common.BaseContext`** (in `internal/cmd/common/context.go`) is the universal carrier for `Output *string`, `Verbose *bool`, `ProjectDir *string`. All command contexts embed this. `SessionContext` extends it with `SessionID *string`.

### `internal/` — Domain Logic Layer (31 packages)

#### Core Domain Packages (hub packages — many importers)

| Package | Files | Purpose | Key Types |
|---|---|---|---|
| `paths` | 2 | XDG path resolution, project root discovery | `Resolver`, `FindProjectRoot()` |
| `errors` | 2 | Domain error codes, exit codes, structured errors | Error code constants, `KnossosError` |
| `output` | 4 | Format-aware printing (text/json/yaml) | `Printer`, `Format`, `Textable` interface |
| `fileutil` | 2 | Atomic file writes | `AtomicWriteFile()` |
| `frontmatter` | 4 | YAML frontmatter parsing, `FlexibleStringSlice` | `Parse()`, `FlexibleStringSlice` |
| `config` | 2 | `KNOSSOS_HOME`, XDG dirs, `ActiveOrg()`, sync.Once cache | `KnossosHome()`, `ActiveOrg()` |

#### Session Domain

| Package | Files | Purpose | Key Types |
|---|---|---|---|
| `session` | 24 | Session lifecycle, state machine, context serialization | `Context`, `FSM`, `Status`, `Phase`, `Strand` |
| `lock` | 4 | Advisory file locking with stale detection | `Manager`, `LockType`, `DefaultTimeout` |

#### Materialization Domain

| Package | Files | Purpose | Key Types |
|---|---|---|---|
| `materialize` | 53 | Core sync pipeline: generates `.claude/` from rite sources | `Materializer`, `Options`, `Result`, `SyncOptions`, `SyncResult` |
| `materialize/source` | 3 | 5-tier rite source resolution | `SourceResolver`, `ResolvedRite` |
| `materialize/hooks` | 2 | Hook config generation | (functions) |
| `materialize/mena` | — | Mena materialization sub-pipeline | — |
| `materialize/orgscope` | — | Org-scope materialization | — |
| `materialize/userscope` | — | User-scope materialization | — |
| `inscription` | 15 | CLAUDE.md region ownership, templating, merge | `Pipeline`, `OwnerType` (`knossos`/`satellite`/`regenerate`) |
| `provenance` | 7 | File-level provenance tracking in `.knossos/PROVENANCE_MANIFEST.yaml` | `ProvenanceManifest`, `ProvenanceEntry`, `OwnerType`, `ScopeType` |
| `sync` | 2 | Sync state persistence (`.knossos/sync/state.json`) | `StateManager`, `State` |
| `mena` | 8 | Mena type detection, routing, extension stripping | `DetectMenaType()`, `StripMenaExtension()`, `RouteMenaFile()` |

#### Rite Domain

| Package | Files | Purpose | Key Types |
|---|---|---|---|
| `rite` | 16 | Rite discovery, manifest, context, workflow, syncer | `Rite`, `Discovery`, `RiteForm` |
| `manifest` | 9 | Generic manifest load/save/diff/merge (JSON+YAML) | `Manifest`, `ValidationIssue` |
| `agent` | 18 | Agent frontmatter parsing and validation | `MemoryField`, `FlexibleStringSlice` |

#### Hook Domain

| Package | Files | Purpose | Key Types |
|---|---|---|---|
| `hook` | 6 | CC hook env parsing, stdin JSON payload, event types | `StdinPayload`, `Env`, `HookEvent` |
| `hook/clewcontract` | 15 | Clew Contract v2 event recording (events.jsonl) | `EventType`, typed event constructors, `Writer` |

#### Quality/Analysis Packages

| Package | Files | Purpose | Key Types |
|---|---|---|---|
| `sails` | 13 | White Sails confidence gate, contract validation | `ContractViolation`, `Gate`, `Generator` |
| `perspective` | 7 | First-person agent perspective assembly and audit | `PerspectiveDocument`, `LayerEnvelope`, `AuditOverlay` |
| `naxos` | 5 | Orphaned session scanner | `Scanner`, `ScanConfig`, `ScanResult` |
| `tribute` | 7 | TRIBUTE.md generation | `Generator`, `GenerateResult` |

#### Knowledge/Artifact Packages

| Package | Files | Purpose | Key Types |
|---|---|---|---|
| `know` | 11 | `.know/` file parsing, change manifest, AST diff | `Meta`, `ChangeManifest`, `DeclKind` |
| `artifact` | 8 | Session artifact management | (types) |
| `ledge` | 4 | `.ledge/` work product management | (types) |

#### Infrastructure/Utility Packages

| Package | Files | Purpose | Key Types |
|---|---|---|---|
| `registry` | 4 | Denial-recovery registry for platform references | `RefKey`, `RefCategory` |
| `validation` | 10 | JSON schema validation for artifacts/sessions | `ValidateSessionFields()` |
| `tokenizer` | 2 | Token counting (cl100k_base) | `Counter` |
| `checksum` | 2 | SHA256 with `sha256:` prefix convention | `Content()`, `Bytes()` |
| `assets` | 1 | In-process embedded FS store (set once at startup) | module-level `var` + accessors |
| `worktree` | 8 | Git worktree management | `Worktree`, `WorktreeMetadata`, `WorktreeStatus` |

**Hub packages** (most-imported): `paths`, `errors`, `fileutil`, `frontmatter`, `session`

**Leaf packages** (no internal imports): `errors`, `output`, `registry`, `frontmatter`, `checksum`, `assets`

## Layer Boundaries

The import graph follows a strict directional model:

```
cmd/ari/main.go
  └── internal/cmd/root/
        └── internal/cmd/{subcommand}/     (CLI wiring — cobra commands)
              └── internal/{domain}/        (domain logic)
                    └── internal/{infra}/   (infrastructure leaf packages)
```

### Layer Definitions

**Layer 1: CLI Entry** (`cmd/ari/`)
- One file, wires embedded assets, delegates to `root.Execute()`
- Imports: `knossos` (module root), `internal/cmd/common`, `internal/cmd/root`, `internal/errors`

**Layer 2: Command Wiring** (`internal/cmd/*/`)
- All `New*Cmd()` factories, cobra command trees
- Each command package imports `internal/cmd/common` plus domain packages
- `common` sub-package provides `BaseContext`/`SessionContext` and asset accessors
- Imports domain packages but never sibling cmd packages (no cross-cmd imports observed)

**Layer 3: Domain Logic** (`internal/materialize`, `internal/session`, `internal/inscription`, `internal/rite`, `internal/hook`, `internal/agent`, etc.)
- Business logic with no awareness of cobra or CLI
- Cross-domain imports are directional: `materialize` imports `inscription`, `provenance`, `sync`, `registry`, `mena`, `frontmatter` but not vice versa
- `session` imports `errors`, `fileutil`, `validation`, `hook/clewcontract`, `paths` (leaf packages only, no hub domain packages)

**Layer 4: Infrastructure Leaves** (`internal/paths`, `internal/errors`, `internal/output`, `internal/fileutil`, `internal/frontmatter`, `internal/checksum`, `internal/registry`, `internal/assets`)
- No internal domain imports
- Exception: `paths` imports `errors` (only); `registry` imports `frontmatter` and `mena`

### Import Graph Observations

**`materialize`** is the heaviest hub in domain layer, importing: `checksum`, `config`, `errors`, `fileutil`, `frontmatter`, `inscription`, `materialize/hooks`, `materialize/mena`, `materialize/orgscope`, `materialize/source`, `paths`, `provenance`, `registry`, `sync`

**`session`** is a focused domain package importing only leaf packages: `errors`, `fileutil`, `hook/clewcontract`, `paths`, `validation`

**`hook/clewcontract`** is a special sub-package of `hook` — it is a leaf that records structured events, imported by `session` and `sails` (not by `hook` itself, avoiding a cycle)

### Boundary Enforcement Patterns

1. **Provenance as leaf**: `internal/provenance` has zero internal imports despite being central to the pipeline. It uses plain strings (`SourceType`) to avoid importing `internal/materialize/source`. This deliberate tension is documented as TENSION-005.

2. **Registry as leaf**: `internal/registry` imports only stdlib. It holds stable keys for agents, skills, and CLI commands so any package can reference them without circular imports.

3. **cmd/common as shared base**: All `internal/cmd/<group>` packages embed `common.BaseContext` (or `common.SessionContext`). This prevents global state from leaking between command handlers.

4. **Annotation-based project gating**: `cmd/root` calls `common.NeedsProject(cmd)` which reads a cobra annotation (`needsProject`) set by each command group. This avoids a direct registry of project-required commands in the root package.

## Entry Points and API Surface

### Binary Entry Point

**`cmd/ari/main.go`**

Initialization sequence:
1. `root.SetVersion(version, commit, date)` — sets ldflags version info
2. `common.SetBuildVersion(version)` — stores version for XDG extraction during `ari init`
3. `common.SetEmbeddedAssets(EmbeddedRites, EmbeddedTemplates, EmbeddedHooksYAML)` — registers embedded FS into `assets` package
4. `common.SetEmbeddedUserAssets(EmbeddedAgents, EmbeddedMena)` — registers user-scope embedded FS
5. `root.Execute()` — runs cobra command tree

### CLI Command Tree

Root command: `ari` (`internal/cmd/root/root.go`)

Global persistent flags: `--output/-o` (text/json/yaml), `--verbose/-v`, `--config`, `--project-dir/-p`, `--session-id/-s`

**All subcommands:**

| Command | Purpose |
|---------|---------|
| `ari session` | Session lifecycle: create, park, resume, wrap, status, list, transition, fray, gc, audit, lock, snapshot, timeline, log, field, migrate, recover, context, claim, query |
| `ari sync` | Run materialize pipeline (`ari sync [--rite NAME] [--scope all|rite|user]`) |
| `ari inscription` | CLAUDE.md sync: sync, diff, validate, rollback |
| `ari manifest` | Manifest ops: show, validate, diff, merge |
| `ari rite` | Rite management: list, current, validate, invoke, context, release |
| `ari hook` | Hook handlers: write-guard, clew, context, autopark, session-end, agentguard, budget, subagent, git-conventions, precompact, validate, worktree-remove, worktree-seed, attribution-guard, cheapo-revert |
| `ari artifact` | Work artifact registry: register, list, query, rebuild |
| `ari worktree` | Git worktree management: create, sync, remove, cleanup |
| `ari knows` | `.know/` domain freshness status |
| `ari agent` | Agent scaffolding |
| `ari sails` | Quality gate (White Sails) |
| `ari naxos` | Orphan session scanner |
| `ari tribute` | TRIBUTE.md generation |
| `ari handoff` | Specialist handoff workflow |
| `ari org` | Org-level resource management: init, list, current, set |
| `ari provenance` | Provenance manifest inspection |
| `ari lint` | Mena/agent lint |
| `ari status` | Project status overview |
| `ari explain` | Platform concept documentation |
| `ari tour` | Onboarding walkthrough |
| `ari validate` | Artifact schema validation |
| `ari init` | Project initialization |
| `ari land` | Cross-session knowledge synthesis |
| `ari ledge` | .ledge/ work product management |
| `ari version` | Version info |

### Key Exported Interfaces

**`paths.Resolver`** — Central path contract. Every command creates one via `common.GetResolver()`. Provides 40+ methods for `.claude/`, `.sos/`, `.knossos/`, `.ledge/`, XDG, and org-level paths.

**`materialize.Materializer`** — The sync engine. Constructed with `materialize.NewMaterializer(resolver)` and configured with `With*` methods. `Sync(SyncOptions)` is the primary entrypoint for rite materialization.

**`session.Context`** — Session state document. Serialized as YAML frontmatter in `SESSION_CONTEXT.md`. Schema version `2.3`. Parsed via `session.ParseContext()`, saved via `ctx.Save(path)`.

**`hook.StdinPayload`** — The canonical shape of CC hook data sent via stdin JSON. This is the source of truth for what CC provides to hooks.

**`output.Printer`** — Uniform output formatting. Created via `common.GetPrinter(defaultFormat)`. All commands use this for text/json/yaml output.

**`common.BaseContext` / `common.SessionContext`** — Context carrier for all CLI commands. Encapsulates output format, verbosity, project dir, and session ID. Provides `GetPrinter()`, `GetResolver()`, `GetSessionID()`, `GetLockManager()`.

**`inscription.Pipeline`** — CLAUDE.md region sync pipeline. Reads `ManifestPath`, applies templates from `TemplateDir`, merges owned regions, writes to `ClaudeMDPath`.

**`provenance.ProvenanceManifest`** — File-level ownership tracker. Stored at `.knossos/PROVENANCE_MANIFEST.yaml`. Maps relative `.claude/` paths to `ProvenanceEntry` with `Owner`, `Scope`, `SourcePath`, `SourceType`.

## Key Abstractions

### 1. `paths.Resolver` (`internal/paths/paths.go`)

The central navigation abstraction. Every command that needs file I/O creates a `Resolver` from the project root. It provides canonical paths for all framework directories without path string manipulation at call sites.

### 2. `session.Context` (`internal/session/context.go`)

The primary session state type. Backed by `SESSION_CONTEXT.md` with YAML frontmatter. Schema version `2.3` added `FrameRef`, `ClaimedBy`, `ParkSource`, and the `Strand` struct (migrated from `[]string` to `[]Strand` with polymorphic YAML via `strandList.UnmarshalYAML`).

**Key fields**: `SessionID`, `Status` (FSM-validated), `CurrentPhase` (requirements/design/implementation/validation/complete), `Complexity`, `ActiveRite`, `Strands` (child session forks), `ClaimedBy` (CC agent instance).

### 3. `session.FSM` (`internal/session/fsm.go`)

State machine for session lifecycle. Valid transitions:
- `none → active` (create)
- `active → {parked, archived}` (park, wrap)
- `parked → {active, archived}` (resume, wrap)
- `archived` is terminal

### 4. `materialize.Materializer` (`internal/materialize/materialize.go`)

The sync engine. Accepts 4 embedded FS sources (rites, templates, agents, mena) via `With*` builder pattern. The `Sync(SyncOptions)` method runs 3 phases: rite scope (Layer 1), org scope (Layer 1.5), user scope (Layer 2). Each phase is independently error-tolerant when `scope=all`.

**5-tier source resolution** (in `internal/materialize/source/resolver.go`):
1. ExplicitSource (`--source` flag)
2. Project satellite rites (`.knossos/rites/`)
3. User rites (`~/.local/share/knossos/rites/`)
4. Org rites (`$XDG_DATA_HOME/knossos/orgs/{org}/rites/`)
5. KnossosHome (`$KNOSSOS_HOME/rites/`) or embedded FS fallback

### 5. `hook.StdinPayload` (`internal/hook/env.go`)

Canonical struct for CC hook data. CC sends `session_id`, `tool_name`, `tool_input`, `tool_response`, `hook_event_name`, `cwd`, etc. via stdin JSON. Environment variables (`CLAUDE_HOOK_*`) are legacy/fallback only. `ParseEnv()` merges stdin over env vars.

### 6. `inscription.Pipeline` (`internal/inscription/pipeline.go`)

CLAUDE.md ownership model. Three region types: `knossos` (always overwritten), `satellite` (never touched), `regenerate` (computed from project state). The `Pipeline` coordinates `Generator`, `Merger`, and backup manager.

Note: `inscription.OwnerType` and `provenance.OwnerType` are distinct types (documented as `TENSION-001` in design constraints): inscription owns CLAUDE.md regions, provenance owns files in `.claude/`.

### 7. `frontmatter.FlexibleStringSlice` (`internal/frontmatter/frontmatter.go`)

Handles both comma-separated strings (`"Bash, Read, Glob"`) and YAML lists (`[Bash, Read, Glob]`) for agent frontmatter fields. Shared across `agent` and `materialize` packages via alias.

### 8. `agent.MemoryField` (`internal/agent/types.go`)

Polymorphic YAML field for CC agent memory config. Accepts `bool true` (normalizes to `"project"`), `bool false` (disabled, `""`), or string enum (`"user"`, `"project"`, `"local"`).

### 9. `provenance.ProvenanceManifest` (`internal/provenance/provenance.go`)

File-level ownership manifest at `.knossos/PROVENANCE_MANIFEST.yaml`. Maps `.claude/`-relative paths to `ProvenanceEntry` with `Owner` (knossos/user/untracked), `Scope` (rite/user/org), `SourcePath`, `SourceType`, and content hash. Enables divergence detection and safe ownership transitions.

### 10. `hook/clewcontract.EventType` (`internal/hook/clewcontract/event.go`)

The Clew Contract event taxonomy. Events are structured JSONL records in `events.jsonl`. Key events: `tool.call`, `tool.file_change`, `agent.decision`, `agent.task_start/end`, `session.started/ended`, `agent.handoff_prepared/executed`, `session.frayed`, `session.strand_resolved`. Consumed by `sails.ValidateClewContract()`.

### Design Patterns

- **Builder pattern**: `Materializer.With*()` methods for configuring embedded FS sources
- **Polymorphic YAML**: `strandList.UnmarshalYAML` (session strands v2.1→v2.3 migration), `MemoryField.UnmarshalYAML` (bool/string)
- **Atomic writes**: `fileutil.AtomicWriteFile()` — temp file + fsync + rename, used everywhere state is persisted
- **sync.Once cache**: `config.KnossosHome()` cached with `sync.Once`; `ResetKnossosHome()` for test isolation
- **Registry leaf**: `internal/registry` imports no internal packages — pure lookup table for platform references
- **Error exit codes**: Explicit exit code constants (0-21) mapped to error categories, all errors carry both code and message

## Data Flow

### Pipeline 1: Sync/Materialization

```
Source: rites/{name}/manifest.yaml + agents/*.md + mena/**/*.{dro,lego}.md
         + knossos/templates/sections/*.md.tpl
         + config/hooks.yaml
         + agents/ (cross-rite) + mena/ (platform)

Resolution:
  internal/materialize/source.SourceResolver.ResolveRite()
    → 5-tier lookup (explicit > project > user > org > knossos/embedded)

Processing:
  internal/materialize.Materializer.Sync()
    Phase 1 (rite scope):
      MaterializeWithOptions(riteName, opts)
        → agent transform (frontmatter stripping of knossos-only fields)
        → mena routing (.dro.md → commands/, .lego.md → skills/)
        → CLAUDE.md via inscription.Pipeline
        → settings.json + hooks.yaml generation
        → provenance.ProvenanceManifest update

    Phase 1.5 (org scope):
      syncOrgScope() → org agents + mena

    Phase 2 (user scope):
      syncUserScope() → cross-rite agents + platform mena

Output: .claude/agents/*.md, .claude/commands/**, .claude/skills/**,
        .claude/CLAUDE.md, .claude/settings.json, .claude/hooks.yaml
        .knossos/PROVENANCE_MANIFEST.yaml, .knossos/sync/state.json
```

Idempotency is enforced by comparing content checksums (`checksum.Content()`) before writing.

### Pipeline 2: Session Lifecycle

```
Input: ari session create "initiative" -c COMPLEXITY

Processing:
  internal/session.NewContext(initiative, complexity, rite)
    → GenerateSessionID() → "session-YYYYMMDD-HHMMSS-{8hex}"
    → FSM.ValidateTransition(StatusNone, StatusActive)
    → Context.Save(.sos/sessions/{id}/SESSION_CONTEXT.md)

State transitions (park, resume, wrap):
  session.LoadContext(path) → mutate status/timestamps → ctx.Save()
  FSM validates each transition before write

Event recording (via hooks):
  hook/clewcontract.Writer.Record(EventType, data)
    → JSONL append to .sos/sessions/{id}/events.jsonl

CC map (for CC↔ari session correlation):
  .sos/sessions/.cc-map/{cc_session_id} → {ari_session_id}
```

### Pipeline 3: Hook Execution

```
Input: CC lifecycle event → ari hook {subcommand} (stdin: JSON payload)

Parsing:
  hook.ParseEnv()
    → parseStdin() reads JSON from stdin
    → merges over env var fallbacks
    → returns hook.Env{Event, ToolName, SessionID, ProjectDir, CWD, ...}

Dispatch (via internal/cmd/hook/):
  Each hook sub-command reads hook.Env and executes domain logic

Output:
  Exit codes control CC behavior:
    0 = proceed
    1 = block (write-guard, agentguard)
    other = logged warning
```

### Pipeline 4: CLAUDE.md Inscription

```
Input: KNOSSOS_MANIFEST.yaml + template dir + existing CLAUDE.md

Processing:
  inscription.Pipeline.Run()
    → inscription.Generator produces section content from templates
    → inscription.Merger merges into existing CLAUDE.md
    → Regions: knossos (overwrite), satellite (preserve), regenerate (compute)

Output: .claude/CLAUDE.md with merged regions
```

### Configuration Merge Points

Agent frontmatter cascade:
1. Rite `manifest.yaml` declares agent list with optional overrides
2. Agent source file has frontmatter: name, role, description, tools, model, color, memory, skills, hooks
3. `materialize/agent_transform.go` strips knossos-only fields before writing to `.claude/agents/`
4. `materialize/agent_defaults.go` applies rite-level defaults for missing fields

Mena routing:
- `.dro.md` files → `.claude/commands/{scope}/{name}/`
- `.lego.md` files → `.claude/skills/{scope}/{name}/`
- Extension stripped: `INDEX.dro.md` → `INDEX.md` in output

Source resolution cascade for rites (5 tiers, first found wins):
1. `--source` flag (explicit path or "knossos")
2. `.knossos/rites/<name>/` (project satellite)
3. `~/.local/share/knossos/rites/<name>/` (user)
4. `$XDG_DATA_HOME/knossos/orgs/<org>/rites/<name>/` (org)
5. `$KNOSSOS_HOME/rites/<name>/` or embedded FS (fallback)

## Knowledge Gaps

1. **`internal/cmd/session/` subcommand implementations**: Only the `session.go` root and test files were sampled. The implementations of `fray`, `claim`, `query`, `context`, `timeline` sub-commands were not read in detail.

2. **`internal/perspective/` resolvers and simulation**: The full layer resolution pipeline and simulation logic were not read. The `PerspectiveDocument` type is documented but resolver implementation details are missing.

3. **`internal/cmd/hook/` individual sub-commands**: Hook behavior is documented at the type/env level but per-hook logic is undocumented.

4. **`internal/materialize/` sub-packages** (`orgscope/`, `userscope/`, `mena/` sub-dirs): The org-scope and user-scope sync pipelines are documented structurally but their implementation details are not captured.

5. **`internal/validation/`**: The `schemas/` subdirectory and JSON schema content were not read. Schema validation rules for session context, artifacts, and handoffs are undocumented.

6. **`internal/rite/workflow.go`, `syncer.go`**: The `Syncer` interface is referenced but its contract is not documented.

7. **`internal/tribute/` full content**: TRIBUTE.md generation logic and the `graduated artifacts` feature are not documented.

---
domain: architecture
generated_at: "2026-03-01T16:08:41Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "89b109c"
confidence: 0.78
format_version: "1.0"
---

# Codebase Architecture

## Package Structure

The knossos project (`github.com/autom8y/knossos`) is a Go 1.23 monorepo. The binary is `ari` (Ariadne). Primary source lives in two trees:

**`cmd/` tree** — binary entry points:
- `cmd/ari/` — single file (`main.go`), the `ari` CLI entry point. Minimal: delegates to `internal/cmd/root`.

**`internal/` tree** — all domain logic (27 packages):

| Package | Purpose | File Count |
|---|---|---|
| `internal/agent` | Agent frontmatter parsing, validation, archetypes, scaffolding | 11 files |
| `internal/artifact` | Federated artifact registry (session + project level) | 6 files |
| `internal/assets` | Static asset loading helpers | 1 file |
| `internal/checksum` | SHA256 checksum computation with `sha256:` prefix | 2 files |
| `internal/cmd` | CLI command wiring (21 sub-packages, one per command group) | 50+ files |
| `internal/config` | XDG home directory resolution | 2 files |
| `internal/errors` | Domain error types, structured codes, exit codes | 1 file |
| `internal/fileutil` | Atomic file write, file utilities | 2 files |
| `internal/frontmatter` | YAML frontmatter parsing from markdown, `FlexibleStringSlice` | 3 files |
| `internal/hook` | CC hook env parsing, event types, `StdinPayload`, output format | 4 files + `clewcontract/` |
| `internal/inscription` | CLAUDE.md region management, manifest, marker parsing, pipeline | 9 files |
| `internal/know` | `.know/` domain file parsing, freshness checks | 4 files |
| `internal/lock` | JSON session lock (5-min stale), Moirai lock abstraction | 4 files |
| `internal/manifest` | Project manifest loading/saving (JSON/YAML, git refs), diff, merge | 6 files |
| `internal/materialize` | Core materialization engine: generates `.claude/` from rite source | 20+ files |
| `internal/mena` | Mena type detection (`.dro.md`/`.lego.md`), routing, walk | 6 files |
| `internal/naxos` | Linter/scanner for agent/mena files (naming, structure) | 5 files |
| `internal/output` | Format-aware printer (text/JSON/YAML), rite/manifest output helpers | 3 files |
| `internal/paths` | `Resolver` struct for all path resolution, XDG dirs, project discovery | 2 files |
| `internal/provenance` | File-level provenance tracking (`PROVENANCE_MANIFEST.yaml`) | 6 files |
| `internal/registry` | Rite registry: discovery and loading of installed rites | 4 files |
| `internal/rite` | Rite manifest types, loading, validation, invoker, workflow, budget | 11 files |
| `internal/sails` | White Sails confidence signal (WHITE/GRAY/BLACK), clew contract validation | 8 files |
| `internal/session` | Session FSM, context, events, discovery, lock, snapshot, timeline | 18 files |
| `internal/sync` | Sync state (`state.json`) for remote rite distribution | 2 files |
| `internal/tokenizer` | Token counting (tiktoken) for context budget estimation | 2 files |
| `internal/tribute` | Tribute document generation and rendering | 5 files |
| `internal/validation` | JSON schema validation for artifacts, handoffs, session fields | 7 files + `schemas/` |
| `internal/worktree` | Git worktree lifecycle management for parallel sessions | 7 files |

**Root-level key files:**
- `embed.go` — embeds `rites/`, `knossos/templates/`, `config/hooks.yaml`, `agents/`, `mena/` into the binary as `embed.FS` vars.
- `go.mod` — module `github.com/autom8y/knossos`, Go 1.23. Key deps: cobra, viper, yaml.v3, sprig/v3, jsonschema/v6, tiktoken-go.

**Source directories outside `cmd/` and `internal/`:**
- `rites/` — rite definition directories (embedded into binary)
- `knossos/templates/` — CLAUDE.md section templates (embedded)
- `mena/` — platform-level mena (commands/skills) (embedded)
- `agents/` — cross-rite agent definitions (embedded)
- `config/` — `hooks.yaml` bootstrap config (embedded)

**Hub packages** (imported by many): `internal/errors`, `internal/paths`, `internal/output`, `internal/frontmatter`, `internal/fileutil`

**Leaf packages** (no internal imports): `internal/errors`, `internal/fileutil`, `internal/checksum`, `internal/tokenizer`, `internal/assets`

**Note**: `internal/registry` declares itself as LEAF in comments but imports `internal/frontmatter` and `internal/mena` via `registry/validate.go`. It is NOT a true leaf package.

---

## Layer Boundaries

The import graph forms a strict downward-only dependency chain:

```
Layer 0 (CLI Entry):  cmd/ari/main.go
Layer 1 (CLI Wiring): internal/cmd/root + 21 sub-packages
Layer 2 (Domain):     internal/materialize, internal/session, internal/inscription, internal/rite, internal/agent, internal/provenance, internal/sails, internal/artifact, internal/worktree, internal/naxos, internal/tribute
Layer 3 (Support):    internal/mena, internal/manifest, internal/sync, internal/lock, internal/validation, internal/hook, internal/registry, internal/know, internal/tokenizer, internal/output
Layer 4 (Foundation): internal/paths, internal/frontmatter, internal/fileutil, internal/checksum, internal/config, internal/assets
Layer 5 (Cross-cut):  internal/errors  (imported by every layer above)
```

**Layer violations (documented):**
- `internal/naxos` imports `internal/sails` and `internal/session` (both Layer 2), so naxos is Layer 2, not Layer 3.
- `internal/tribute` imports `internal/artifact` and `internal/session` (both Layer 2), so tribute is Layer 2, not Layer 3.
- `internal/registry` imports `internal/frontmatter` and `internal/mena` via `validate.go`, violating its LEAF declaration. It is functionally Layer 3.

**Observed import patterns:**
- `internal/materialize` imports: `inscription`, `provenance`, `registry`, `sync`, `mena`, `agent`, `checksum`, `config`, `errors`, `fileutil`, `paths`. It is the widest hub in Layer 2.
- `internal/session` imports: `errors`, `fileutil`, `validation`, `paths`, `lock`, `hook/clewcontract`. Self-contained session lifecycle.
- `internal/inscription` imports: `frontmatter`, `errors`, `fileutil`. No circular deps with `materialize` — `materialize` calls `inscription.Pipeline.Sync()` one-way.
- `internal/cmd/*` packages import their corresponding domain package plus `internal/cmd/common`, `internal/output`, `internal/paths`. Command packages do NOT import each other.
- `internal/sails` imports `internal/hook/clewcontract` for event parsing. One-way dependency into hook infrastructure.
- `internal/provenance` states explicitly: "One-way dependency: materialize imports provenance, never the reverse."

**Circular dependency avoidance patterns:**
- `internal/cmd/common` provides shared context structs (`BaseContext`, `SessionContext`) so command packages share interface without cross-importing.
- `internal/hook/clewcontract` is a sub-package of `hook` — sails and session can import it without pulling in full hook package.
- `internal/materialize/source` is a sub-package re-exported via type aliases in `materialize/source.go` for backward compat.

---

## Entry Points and API Surface

### Binary Entry Point

**`cmd/ari/main.go`** (32 lines):
1. Sets version info via `root.SetVersion(version, commit, date)`
2. Injects embedded assets: `common.SetEmbeddedAssets(knossos.EmbeddedRites, knossos.EmbeddedTemplates, knossos.EmbeddedHooksYAML)` and `common.SetEmbeddedUserAssets(knossos.EmbeddedAgents, knossos.EmbeddedMena)`
3. Calls `root.Execute()`

### Cobra Command Tree

Root: `ari` (defined in `internal/cmd/root/root.go`)

Global flags: `--output/-o` (text/json/yaml), `--verbose/-v`, `--config`, `--project-dir/-p`, `--session-id/-s`

`PersistentPreRunE` on root: validates output format, inits viper config, discovers project root via `paths.FindProjectRoot`.

**Subcommands registered in `root.init()`:**

| Subcommand | Package | One-line description |
|---|---|---|
| `session` | `internal/cmd/session` | Session lifecycle (create, status, park, resume, wrap, fray, gc, etc.) |
| `manifest` | `internal/cmd/manifest` | Project manifest load/diff/merge |
| `inscription` | `internal/cmd/inscription` | CLAUDE.md region sync |
| `sync` | `internal/cmd/sync` | Unified rite + user scope sync (materializes `.claude/`) |
| `validate` | `internal/cmd/validate` | Manifest, agent, rite validation |
| `handoff` | `internal/cmd/handoff` | Agent handoff artifact management |
| `worktree` | `internal/cmd/worktree` | Git worktree lifecycle |
| `hook` | `internal/cmd/hook` | CC hook infrastructure (agentguard, autopark, budget, writeguard, etc.) |
| `knows` | `internal/cmd/knows` | `.know/` domain file status and refresh |
| `artifact` | `internal/cmd/artifact` | Artifact registry queries |
| `sails` | `internal/cmd/sails` | White Sails confidence gate |
| `naxos` | `internal/cmd/naxos` | Lint agent/mena files |
| `rite` | `internal/cmd/rite` | Rite discovery, invoke, release |
| `agent` | `internal/cmd/agent` | Agent list, new, update, validate |
| `tribute` | `internal/cmd/tribute` | Tribute document generation |
| `migrate` | `internal/cmd/migrate` | Schema migration (sessions, manifests) |
| `init` | `internal/cmd/initialize` | Bootstrap new project `.claude/` |
| `provenance` | `internal/cmd/provenance` | Provenance manifest inspection |
| `lint` | `internal/cmd/lint` | Lint (delegates to naxos scanner) |
| `version` | `internal/cmd/root` (inline) | Show version/platform info |

**`session` subcommands** (in `internal/cmd/session/session.go`): `create`, `status`, `list`, `park`, `resume`, `wrap`, `transition`, `migrate`, `audit`, `recover`, `fray`, `lock`, `unlock`, `gc`, `field set/get`, `log`, `timeline`, `context`

**`hook` subcommands** (in `internal/cmd/hook/`): `agentguard`, `autopark`, `budget`, `clew`, `context`, `gitconventions`, `precompact`, `sessionend`, `subagent`, `validate`, `writeguard`

### Key Exported Interfaces and Contracts

- `root.GlobalOptions` struct — shared global state across all subcommands
- `common.BaseContext`, `common.SessionContext` — shared context structs that command groups embed
- `paths.Resolver` — canonical path resolver; all file access uses this
- `materialize.Materializer` — the core sync engine; exposes `Sync(SyncOptions) (*Result, error)`
- `materialize.SourceResolver` (from `materialize/source`) — 4-tier resolution: project > user > knossos > embedded
- `inscription.Pipeline` — CLAUDE.md region sync orchestrator
- `session.FSM` — 4-state finite state machine (NONE, ACTIVE, PARKED, ARCHIVED)
- `provenance.ProvenanceManifest` — file ownership tracker for `.claude/`
- `sails.ValidateClewContract(sessionDir string)` — clew event contract validation
- `artifact.Registry` — CRUD for session + project artifact registries

---

## Key Abstractions

### 1. `errors.Error` (`internal/errors/errors.go`)

The universal error type. Every error in the codebase is `*errors.Error` with structured `Code string`, `Message string`, `Details map[string]interface{}`, and `ExitCode int`. 26 exit codes (0-21) map to 30+ error codes. Constructors: `errors.New`, `errors.NewWithDetails`, `errors.Wrap`. Predicate helpers: `errors.IsNotFound`, `errors.IsLifecycleError`, etc.

### 2. `paths.Resolver` (`internal/paths/paths.go`)

Central path authority. Created with `paths.NewResolver(projectRoot)`. All paths in the project derive from this: `.SessionDir(id)`, `.SessionContextFile(id)`, `.AgentsDir()`, `.RiteDir(name)`, `.KnossosManifestFile()`, etc. Also exposes static helpers: `paths.FindProjectRoot()` (walks up from cwd looking for `.claude/`), `paths.UserClaudeDir()`, `paths.ConfigDir()` (XDG).

### 3. `session.Context` (`internal/session/context.go`)

The `SESSION_CONTEXT.md` schema. YAML frontmatter struct with: `SessionID`, `Status` (FSM state), `Initiative`, `Complexity`, `ActiveRite`, `CurrentPhase`, lifecycle timestamps, fray fields (session forking). Methods: `ParseContext`, `Serialize`, `Save` (atomic write), `Validate`. Schema version `"2.1"`.

### 4. `session.FSM` (`internal/session/fsm.go`)

4-state machine: `NONE -> ACTIVE -> {PARKED, ARCHIVED}`, `PARKED -> {ACTIVE, ARCHIVED}`. ARCHIVED is terminal. Phase sub-FSM: requirements -> design -> implementation -> validation -> complete (forward-only). Methods: `CanTransition`, `ValidateTransition`.

### 5. `inscription.Region` / `inscription.Manifest` (`internal/inscription/types.go`)

Manages CLAUDE.md regions with three owner types: `OwnerKnossos` (always overwritten), `OwnerSatellite` (never overwritten), `OwnerRegenerate` (generated from source). Stored in `KNOSSOS_MANIFEST.yaml`. Markers are HTML comments: `<!-- KNOSSOS:START region-name [key=value] -->`. Hash-based idempotency: SHA256 of last-synced content.

### 6. `provenance.ProvenanceManifest` (`internal/provenance/provenance.go`)

File-level ownership tracker. Two manifest files: `PROVENANCE_MANIFEST.yaml` (rite scope) and `USER_PROVENANCE_MANIFEST.yaml` (user scope). `ProvenanceEntry` fields: `Owner` (knossos/user/untracked), `Scope` (rite/user), `SourcePath`, `SourceType`, `Checksum` (sha256: prefix), `LastSynced`. Divergence detection: knossos -> user promotion on checksum mismatch.

### 7. `agent.AgentFrontmatter` (`internal/agent/frontmatter.go`)

Agent markdown YAML frontmatter schema. Required: `name`, `description`. Optional CC-native: `maxTurns`, `skills`, `disallowedTools`, `memory` (MemoryField: bool or string scope), `permissionMode`, `mcpServers`, `hooks`. Knossos-specific: `type` (orchestrator/specialist/reviewer/meta/designer/analyst/engineer), `upstream`, `downstream`, `produces`, `contract` (BehavioralContract).

### 8. `mena.DetectMenaType` / `RouteMenaFile` (`internal/mena/types.go`)

Extension-based routing: `.dro.md` -> `commands/` (dromena, user-invoked), `.lego.md` -> `skills/` (legomena, reference). `StripMenaExtension` removes the infix for output filenames. This distinction is fundamental to context lifecycle: dromena are transient, legomena persist in context.

### 9. `rite.RiteManifest` (`internal/rite/manifest.go`)

Rite definition schema loaded from `manifest.yaml`. Fields: `Name` (kebab-case), `EntryAgent`, `Phases []ManifestPhase`, `Agents []AgentRef`, `Dependencies []string`, `Budget *BudgetInfo`. `RiteForm` enum: simple/practitioner/procedural/full. Polymorphic `Skills` field (string list or SkillRef objects).

### 10. `hook.StdinPayload` (`internal/hook/env.go`)

CC sends hook data as JSON on stdin (NOT env vars). Key fields: `HookEventName`, `ToolName`, `ToolInput json.RawMessage`, `SessionID`, `CWD`. `ParseEnv()` reads stdin first, then falls back to env vars for backward compat. Known hook events: `PreToolUse`, `PostToolUse`, `Stop`, `SessionStart`, `SubagentStart`, `PreCompact`, etc.

### Design Patterns

- **Polymorphic YAML**: `MemoryField` (bool->"project", string), `Skills` field in `RiteManifest` ([]string or []SkillRef). Custom `UnmarshalYAML`/`UnmarshalJSON` implementations.
- **Envelope pattern**: `errors.Error{Code, Message, Details, ExitCode}` — all errors are machine-readable envelopes.
- **Registry pattern**: `artifact.Registry`, `internal/registry` rite registry. Stateless structs holding a `paths.Resolver`.
- **Idempotency pattern**: `writeIfChanged()` in materialize prevents unnecessary file writes; provenance checksum guards in `inscription.Manifest`.
- **Scan-based discovery**: `session.discovery` scans filesystem for session dirs instead of maintaining in-memory cache (eliminates TOCTOU).

---

## Data Flow

### Sync Pipeline (Primary Path)

The `ari sync` command calls `materialize.Materializer.Sync(SyncOptions)`.

```
User runs: ari sync
    |
internal/cmd/sync/sync.go
    |
materialize.NewMaterializer(projectRoot, options)
    |
materialize.Sync(SyncOptions{Scope, RiteName, DryRun, ...})
    +-- [Rite Scope]: MaterializeWithOptions(riteName)
    |   +-- SourceResolver.Resolve(riteName) -> 4-tier lookup:
    |   |   project rites/ > user $KNOSSOS_DATA/rites/ > knossos embedded > explicit
    |   +-- Load rite manifest.yaml -> rite.RiteManifest
    |   +-- Resolve dependencies (shared rites)
    |   +-- Stage 1: materialize agents -> .claude/agents/
    |   +-- Stage 2: project mena -> .claude/commands/ and .claude/skills/
    |   |   (DetectMenaType: .dro.md->commands/, .lego.md->skills/)
    |   +-- Stage 3: hooks -> .claude/settings.json
    |   +-- Stage 4: rules -> .claude/rules/
    |   +-- Stage 5: inscription (CLAUDE.md) via inscription.Pipeline.Sync()
    |   |   +-- Load KNOSSOS_MANIFEST.yaml
    |   |   +-- Generate sections from templates (sprig+yaml)
    |   |   +-- Merge: preserve satellite regions, update knossos/regenerate regions
    |   |   +-- Backup previous CLAUDE.md
    |   |   +-- AtomicWriteFile(.claude/CLAUDE.md)
    |   +-- Stage 6: Collect provenance -> PROVENANCE_MANIFEST.yaml
    |   +-- Stage 7: Orphan detection (agents present in old rite, absent in new)
    |
    +-- [User Scope]: syncUserScope()
        +-- ResolveUserResources(KNOSSOS_HOME)
        +-- Sync agents -> ~/.claude/agents/
        +-- Sync mena -> ~/.claude/commands/ + ~/.claude/skills/
        +-- Sync hooks -> ~/.claude/hooks/
```

Source resolution order: `project rites/ -> $KNOSSOS_DATA/rites/ -> knossos embedded -> explicit`

### Session Pipeline

```
ari session create "initiative" -c MODULE
    |
internal/cmd/session/create.go
    |
session.NewContext(initiative, complexity, rite)
    -> generates session_id (format: session-YYYYMMDD-HHMMSS-{8hex})
    -> status: ACTIVE, phase: requirements
    |
ctx.Save(paths.SessionContextFile(sessionID))  [atomic write]
    -> .sos/sessions/{sessionID}/SESSION_CONTEXT.md
    |
lock.Acquire(sessionID)  -> .sos/sessions/.locks/{sessionID}.lock
    |
clewcontract.BufferedEventWriter.Write(session_start event)
    -> .sos/sessions/{sessionID}/events.jsonl  [append-only JSONL]
```

Session state changes (park/resume/wrap) go through `session.FSM.ValidateTransition` before writing.

### Hook Pipeline

```
CC fires lifecycle event (e.g., PreToolUse)
    |
ari hook {subcommand}  [invoked as subprocess by CC]
    |
hook.ParseEnv()
    +-- parseStdin() -> reads JSON from stdin -> StdinPayload
    +-- falls back to CLAUDE_* env vars for legacy support
    |
Hook-specific logic (e.g., writeguard checks for *_CONTEXT.md writes)
    |
Output: PreToolUseOutput JSON to stdout
    {hookSpecificOutput: {permissionDecision: "allow"|"deny"}}
```

Hook decisions: allow->allow, block->deny. Errors default to allow (graceful degradation). Timeout: 100ms target, 500ms max (`internal/cmd/hook/hook.go`).

### Configuration Cascade

For agent frontmatter, materialization applies a 3-tier cascade:
1. Rite `manifest.yaml` -> agent-level defaults (model, tools)
2. Agent file YAML frontmatter -> per-agent values
3. Archetype defaults (`materialize/agent_defaults.go`) -> fill missing fields

Mena resolution for rite scope uses 4-tier priority: rite-local -> dependency rites -> shared mena -> user mena.

---

## Removed Commands

### `ari migrate` (REMOVED, PKG-013)

The `ari migrate roster-to-knossos` subcommand was removed in PKG-013 cleanup sprint. This command migrated satellite manifest source fields from `roster` to `knossos` branding after the product rename. All satellites have been migrated; the tool is no longer needed.

- Removed files: `internal/cmd/migrate/migrate.go`, `internal/cmd/migrate/roster_to_knossos.go`, `internal/cmd/migrate/roster_to_knossos_test.go`
- Removed output types: `RosterMigrateOutput`, `ManifestMigResult`, `EnvVarDetected` from `internal/output/output.go`
- Root registration removed from `internal/cmd/root/root.go`

---

## Knowledge Gaps

1. `internal/worktree` and `internal/tribute` — files were enumerated but not read in detail. Worktree lifecycle and tribute generation internals are partially undocumented here.
2. `internal/naxos` scanner rules — file structure was noted but lint rule specifics not read.
3. `internal/materialize/hooks/` sub-directory — hooks materialization specifics not read.
4. `internal/materialize/mena/` and `internal/materialize/source/` sub-directories — partially covered via re-exports and source.go, not fully read.
5. `internal/materialize/userscope/` sub-directory — user scope sync specifics covered at high level only.
6. `internal/sails/generator.go`, `internal/sails/thresholds.go` — sails gate generation internals not read.
7. `internal/lock/moirai.go` — Moirai-specific lock abstraction not read in detail.
8. `internal/registry/registry.go` — rite registry discovery logic not fully read.
9. `config/hooks.yaml` — the embedded CC hooks configuration content not examined.
10. `rites/` directory structure — the actual installed rites and their schemas not enumerated.

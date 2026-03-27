---
domain: architecture
generated_at: "2026-03-27T19:57:42Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "5501b0aa"
confidence: 0.85
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/initiative-history.md"
land_hash: "77ed0121d9982cb4e1bf8e7c09b414494a561c0ccd3b9ea546d210acd7354553"
---

# Codebase Architecture

## Package Structure

The knossos codebase is a Go module (`github.com/autom8y/knossos`) with a single binary (`ari`, named Ariadne). Source is organized under two top-level directories: `cmd/` (entry point) and `internal/` (all implementation). There are 45 packages under `internal/`, ranging from 1 to 57 files.

### Module Root

`embed.go` — Package `knossos` (module root). Hosts all `//go:embed` directives so they can reference top-level directories (`rites/`, `knossos/templates/`, `agents/`, `mena/`, `processions/`, `config/hooks.yaml`). Imported by `cmd/ari/main.go` to wire embedded assets into the binary for single-binary distribution.

### cmd/

| Directory | Purpose |
|-----------|---------|
| `cmd/ari/` | Single entry point. `main.go` calls `root.SetVersion`, `common.SetBuildVersion`, `common.SetEmbeddedAssets`, `common.SetEmbeddedUserAssets`, `common.SetEmbeddedProcessions`, then delegates to `root.Execute()`. Minimal logic — wires version and embedded FS only. |

### internal/ Package Inventory

**Leaf packages** (zero internal imports — foundational utilities consumed by all):

| Package | Files | Purpose |
|---------|-------|---------|
| `internal/errors` | 2 | Domain-specific error types with exit codes. `Error` struct carries `Code string`, `ExitCode int`. Named codes: `GENERAL_ERROR`, `FILE_NOT_FOUND`, `SESSION_NOT_FOUND`, `PROJECT_NOT_FOUND`, `LOCK_TIMEOUT`, `SCHEMA_INVALID`, `LIFECYCLE_VIOLATION`, and ~14 others. |
| `internal/output` | 4 | Format-aware CLI output. `Printer` struct emits JSON, YAML, or text (table). `Format` type. `NewPrinter(format, out, errOut, verbose)`. Implements `Textable` and `Tabular` interfaces. |
| `internal/fileutil` | 2 | Canonical atomic file write utilities. |
| `internal/frontmatter` | 4 | YAML frontmatter parsing. `FlexibleStringSlice` (accepts both comma-separated strings and YAML lists). `Parse(content []byte) (yamlBytes, body, err)`. |
| `internal/checksum` | 2 | File checksum computation. |
| `internal/channel` | 2 | Harness-agnostic tool name mapping. `CanonicalTool` map: canonical knossos tool names -> per-channel wire names (`claude`, `gemini`, etc.). |
| `internal/resolution` | 3 | Multi-tier resolution chain. `Chain` with `Tier` (label, dir, fs). Five tiers: project > user > org > platform > embedded. **Zero internal imports** — all paths injected via constructor (TENSION-005). |
| `internal/assets` | 1 | Embedded asset accessors. |
| `internal/llm` | 2 | Shared LLM client (Anthropic SDK transport). Used by: triage query refinement, domain reasoning, summary generation, conversation summarization. |
| `internal/tokenizer` | 2 | Token counting using `tiktoken-go` with `cl100k_base` encoding. |
| `internal/suggest` | 3 | CLI suggestion generation. Max 2 suggestions per event. |
| `internal/mena` | 8 | Mena (commands + skills) discovery across filesystem tiers. |
| `internal/trust` | 11 | Confidence scoring. `ConfidenceTier`, `ConfidenceScore`, `Scorer`, `ScoreInput`. Decay configs per domain type. `ProvenanceChain` and `ProvenanceLink` types. |

**Domain packages** (primary business logic):

| Package | Files | Internal Imports | Purpose |
|---------|-------|-----------------|---------|
| `internal/manifest` | 8 | errors, fileutil | Manifest loading, validation, diffing, merging. `Manifest{Path, Format, Content map[string]any, Raw}`. Supports JSON, YAML, TOML formats. Includes schema validation via JSON Schema. |
| `internal/session` | 27 | errors, fileutil, hook/clewcontract, paths, validation | Session lifecycle FSM. Key types: `Status` (NONE/ACTIVE/PARKED/ARCHIVED), `Phase` (requirements/design/implementation/validation/complete), `FSM`, `Context`, `Snapshot`, `Event` (legacy), `ClewEvent` (current). Event log: `events.jsonl` (dual-format reader per ADR-0027). `ChannelLifecycleMap` normalizes per-harness event names. |
| `internal/hook` | 13 | channel, errors | Hook data model. `StdinPayload` (CC sends hook data as JSON on stdin, NOT env vars). `HookEvent` constants: pre_tool, post_tool, post_tool_failure, permission_request, stop, session_start/end, pre_prompt, pre_compact, subagent_start/stop, notification, teammate_idle, task_completed, pre_model, post_model. `LifecycleAdapter` interface: `ParsePayload`, `FormatResponse`, `ChannelName`. Implementations: `ClaudeAdapter`, `GeminiAdapter`. |
| `internal/hook/clewcontract` | 17+ | (internal leaf) | Clew Contract v2 event schema. `Event` struct with `ts`, `type`, `channel`, `data`. `EventType` constants: tool.call, tool.file_change, agent.decision, context_switch, quality.sails_generated, session.*, phase.transitioned, lock.*, artifact.*. `BufferedEventWriter` for canonical event emission. |
| `internal/rite` | 16 | config, errors, fileutil, paths, resolution | Rite management. `RiteManifest{Name, Version, Description, EntryAgent, Phases, Agents, Skills, Dependencies, ComplexityLevels, Budget}`. `Rite{Name, Path}`. `Invoker` with `InvokeOptions`/`InvokeResult`. `BudgetCalculator`. `ContextLoader`. `RiteForm` type. |
| `internal/materialize` | 57 | checksum, config, errors, fileutil, frontmatter, inscription, paths, provenance, registry, sync + all internal sub-packages | Core materialization engine. Largest package (57 files). `Materializer` struct with builder pattern. `MaterializeWithOptions(riteName, opts)` and `Sync(opts)` are the primary entry points. Reads rite source files -> renders templates -> writes channel directory. Sub-packages: `compiler/`, `hooks/`, `mena/`, `orgscope/`, `procession/`, `source/`, `userscope/`. |
| `internal/inscription` | 16 | (session-related) | Context file (CLAUDE.md / GEMINI.md) generation. `Generator` with `GenerateSection(regionName)` and `GenerateAll()`. `RenderContext` carries manifest + session data. Templates in `knossos/templates/sections/*.md.tpl` with Go equivalents for single-binary distribution. |
| `internal/paths` | 4 | errors | XDG-aware path resolution. `Resolver` struct with all `.sos/` path computations: `SOSDir`, `SessionsDir`, `SessionDir(id)`, `SessionContextFile`, `SessionEventsFile`, `LockFile`, `CurrentSessionFile`, `ActiveRiteFile`. Also: `FindProjectRoot(startDir)` (walks up filesystem). Multi-channel: `AgentsDirForChannel(ch)`, `UserChannelDir`, `UserAgentsDirForChannel`, etc. |
| `internal/sync` | 2 | checksum, errors, fileutil, paths | Sync state management. Tracks active rite name and last sync timestamp in a state file. |
| `internal/provenance` | 8 | paths | File provenance manifest for channel directory. Tracks origin and ownership of all materialized files. Enables divergence detection. |
| `internal/config` | 6 | (stdlib only) | `KnossosHome()`, `ActiveOrg()`, `XDGDataDir()`, `RegistryDir(orgName)`. `OrgContext` interface. `Settings` struct loaded from `settings.yaml`. |
| `internal/agent` | 17 | errors, frontmatter, validation | Agent file parsing and management. Reads agent YAML frontmatter. |
| `internal/know` | 13 | (parsed from context) | `.know/` file parsing. `Meta` (frontmatter), `DomainStatus`. AST-based semantic diffing (`SemanticDiff`, `DeclChange`). `ServiceBoundary` discovery. `QualifiedDomainName`. Validation (`BrokenRef`, `ValidationReport`). |
| `internal/perspective` | 7 | materialize (via source) | First-person agent context view. `Assemble(ctx, opts, start)` builds `PerspectiveDocument` from layers: identity, capability, constraint, memory, behavioral contract. |
| `internal/search` | 11 | agent, concept, config, frontmatter, know, materialize/procession, paths, registry/org, rite, session | Natural language query matching across the knossos CLI surface. `Domain` type. BM25 fusion search. |
| `internal/sails` | 13 | errors, hook/clewcontract, paths, session, validation | White Sails confidence signaling. Color computation algorithm (White/Gray/Black). |
| `internal/triage` | 5 | know, llm | LLM-assisted triage reasoning pipeline. |
| `internal/tribute` | 7 | (session/events) | `TRIBUTE.md` auto-generation for session summaries. `Extractor`, `Generator`, `Renderer`. |
| `internal/naxos` | 10 | fileutil, paths, sails, session | Session hygiene triage. Scans for orphaned sessions. |
| `internal/procession` | 2 | (procession template loading) | Cross-rite coordinated workflow templates. |
| `internal/slack` | 9 | (8 internal imports) | Slack surface for Clew Slack bot. `SlackClient`, `SlackHandler`, `ThreadContextStore`. Routes Slack events to reasoning pipeline. |
| `internal/observe` | 10 | reason/response | OpenTelemetry observability. `CostTracker` (token usage accumulation), pipeline middleware, metrics. |
| `internal/reason` | 5 (+ 3 sub-pkgs) | context/, intent/, response/ | Query reasoning pipeline. `ReasoningConfig`, pipeline stages. Sub-packages: `context/` (prompt assembler), `intent/` (classifier), `response/` (generator + stream). |
| `internal/lock` | 4 | (stdlib) | Advisory file locking with stale detection. |
| `internal/ledge` | 4 | artifact, errors | Work product artifact management (`.ledge/decisions/`, `.ledge/specs/`, etc.). |
| `internal/artifact` | 8 | errors | Workflow artifact file types and management. |
| `internal/validation` | 12 | (JSON Schema) | Artifact validation against JSON Schema. |
| `internal/worktree` | 8 | (git) | Git worktree management for parallel sessions. |
| `internal/envload` | 5 | (stdlib) | Environment variable loading from `.env` files. |
| `internal/concept` | 1 | (stdlib) | Knossos concept registry. Extracted from `internal/cmd/explain` to resolve TENSION-015 (cross-layer import). |
| `internal/registry` | 4 | (stdlib — **zero internal imports**) | Unified denial-recovery registry. Maps stable keys -> concrete values with recovery hints. **Explicitly a leaf package.** |
| `internal/serve` | 4 | (stdlib) | HTTP server configuration for `ari serve`. `ServerConfig`. |

### internal/cmd/ Package Inventory

Each `internal/cmd/{name}/` package wires a CLI subcommand group, delegating to domain packages in `internal/`:

| Package | CLI Command | Purpose |
|---------|------------|---------|
| `internal/cmd/root` | `ari` | Root command, registers all subcommands, global flags |
| `internal/cmd/common` | -- | Shared command context (`BaseContext`, `SessionContext`), embedded asset injection |
| `internal/cmd/session` | `ari session` | Session lifecycle management |
| `internal/cmd/sync` | `ari sync` | Rite and user resource synchronization |
| `internal/cmd/hook` | `ari hook` | Hook handlers: agent-guard, autopark, budget, clew, context, drift-detect, git-conventions, precompact, session-end, subagent, validate, write-guard, worktree-remove, worktree-seed |
| `internal/cmd/manifest` | `ari manifest` | Manifest file management |
| `internal/cmd/inscription` | `ari inscription` | Context file (CLAUDE.md) inscription management |
| `internal/cmd/rite` | `ari rite` | Rite invocation and composition |
| `internal/cmd/agent` | `ari agent` | Agent management (summon/dismiss) |
| `internal/cmd/knows` | `ari knows` | `.know/` freshness inspection |
| `internal/cmd/artifact` | `ari artifact` | Workflow artifact management |
| `internal/cmd/sails` | `ari sails` | White Sails quality gate |
| `internal/cmd/naxos` | `ari naxos` | Orphaned session cleanup |
| `internal/cmd/tribute` | `ari tribute` | `TRIBUTE.md` generation |
| `internal/cmd/handoff` | `ari handoff` | Agent handoff execution |
| `internal/cmd/procession` | `ari procession` | Cross-rite workflow management |
| `internal/cmd/worktree` | `ari worktree` | Git worktree management |
| `internal/cmd/validate` | `ari validate` | Artifact validation |
| `internal/cmd/status` | `ari status` | Unified project health dashboard |
| `internal/cmd/serve` | `ari serve` | HTTP webhook server for Clew |
| `internal/cmd/explain` | `ari explain` | Concept explanation |
| `internal/cmd/tour` | `ari tour` | Project directory walkthrough |
| `internal/cmd/ask` | `ari ask` | Natural language command discovery |
| `internal/cmd/complaint` | `ari complaint` | Complaint management |
| `internal/cmd/land` | `ari land` | Cross-session knowledge synthesis |
| `internal/cmd/lint` | `ari lint` | Mena and agent source validation |
| `internal/cmd/org` | `ari org` | Organization management |
| `internal/cmd/registry` | `ari registry` | Org knowledge domain registry |
| `internal/cmd/provenance` | `ari provenance` | Channel directory provenance inspection |
| `internal/cmd/ledge` | `ari ledge` | Work product artifact management |
| `internal/cmd/initialize` | `ari init` | Project initialization |

**Hub packages** (import the most internal packages):

1. `internal/materialize` — 17 distinct internal imports. Core of the sync pipeline.
2. `internal/search` — 13 internal imports. Cross-cutting query across all domains.
3. `internal/slack` — 8 internal imports. Clew Slack surface.
4. `internal/reason` — 7 internal imports. LLM reasoning pipeline.

---

## Layer Boundaries

The import graph follows a strict three-layer model:

```
cmd/ari/main.go
      |
      v
internal/cmd/root/           <- CLI surface layer (cobra command wiring)
      |
      v
internal/cmd/{subcommand}/   <- Command wiring layer (each subcommand group)
      |
      v
internal/{domain}/           <- Domain logic layer (business logic, no CLI deps)
      |
      v
internal/{leaf}/             <- Infrastructure layer (errors, output, paths, fileutil, etc.)
```

### Layer Rules Observed

**CLI surface layer** (`cmd/ari/`) imports:
- `internal/cmd/root` — command dispatch
- `internal/errors` — exit code handling
- `internal/output` — format-aware printing
- Module root (`github.com/autom8y/knossos`) — embedded FS

**Command wiring layer** (`internal/cmd/*/`) imports:
- `internal/cmd/common` — shared context types
- Domain packages (e.g., `internal/materialize`, `internal/session`)
- Infrastructure: `internal/errors`, `internal/output`, `internal/paths`
- No `internal/cmd/*` imports from other command packages

**Domain logic layer** (`internal/materialize`, `internal/session`, etc.) imports only:
- Sibling domain packages (directional, acyclic)
- Infrastructure leaf packages
- Never imports `internal/cmd/*` (no upward imports)

### Documented Import Invariants (from source comments)

- `internal/reason/pipeline.go`: "reason/ does NOT import internal/serve/ or internal/cmd/. trust/ and search/ do NOT import reason/"
- `internal/reason/pipeline.go` (TriageCandidateInput): "reason/ does NOT import triage/. The handler in slack/ converts triage.TriageCandidate to this type."
- `internal/resolution/chain.go`: "ZERO internal imports. All tier paths are injected via constructor to avoid import cycles (TENSION-005)"
- `internal/registry/registry.go`: "LEAF package -- imports only stdlib"
- `internal/provenance`: "One-way dependency: materialize imports provenance, never the reverse"

### Hub vs Leaf Classification

**Hubs** (imported by many, import many):
- `internal/materialize` — 17 unique internal imports
- `internal/search` — 13 unique internal imports
- `internal/slack` — 8 imports

**True leaves** (zero internal imports):
`errors`, `output`, `fileutil`, `frontmatter`, `checksum`, `channel`, `resolution`, `assets`, `llm`, `tokenizer`, `suggest`, `mena`, `trust`, `registry`

### Boundary Enforcement Patterns

1. **TENSION-005** (resolved): `internal/resolution` has zero internal imports. All tier paths injected via constructor.
2. **TENSION-015** (resolved): `internal/concept` extracted from `internal/cmd/explain`. Prevented cross-layer import.
3. **BC-01**: `internal/llm` lives at infrastructure layer, not inside any domain package.

---

## Entry Points and API Surface

### Binary Entry Point

`cmd/ari/main.go` — `func main()`:

```
main()
  -> root.SetVersion(version, commit, date)
  -> common.SetBuildVersion(version)
  -> common.SetEmbeddedAssets(EmbeddedRites, EmbeddedTemplates, EmbeddedHooksYAML)
  -> common.SetEmbeddedUserAssets(EmbeddedAgents, EmbeddedMena)
  -> common.SetEmbeddedProcessions(EmbeddedProcessions)
  -> root.Execute()
     -> cobra rootCmd.Execute()
        -> PersistentPreRunE: output format validation, initConfig (Viper), FindProjectRoot
        -> subcommand dispatch
```

Build version info (`version`, `commit`, `date`) is injected at link time via `-ldflags`.

### Global Flags on Root Command

`-o/--output` (text/json/yaml), `-v/--verbose`, `--config`, `-p/--project-dir`, `-s/--session-id`, `--channel` (claude/gemini/all)

### CLI Subcommand Table

| Subcommand | Package | Description |
|---|---|---|
| `ari session` | `internal/cmd/session` | Manage workflow sessions |
| `ari manifest` | `internal/cmd/manifest` | Manage manifest files |
| `ari inscription` | `internal/cmd/inscription` | Manage context file inscription system |
| `ari sync` | `internal/cmd/sync` | Synchronize rite and user resources |
| `ari validate` | `internal/cmd/validate` | Validate an artifact file against its schema |
| `ari handoff` | `internal/cmd/handoff` | Manage agent handoffs |
| `ari procession` | `internal/cmd/procession` | Manage cross-rite coordinated workflows |
| `ari worktree` | `internal/cmd/worktree` | Manage git worktrees |
| `ari hook` | `internal/cmd/hook` | Harness hook infrastructure |
| `ari knows` | `internal/cmd/knows` | Inspect `.know/` freshness |
| `ari artifact` | `internal/cmd/artifact` | Manage workflow artifacts |
| `ari sails` | `internal/cmd/sails` | White Sails quality gate |
| `ari naxos` | `internal/cmd/naxos` | Cleanup abandoned sessions |
| `ari rite` | `internal/cmd/rite` | Manage rite invocations |
| `ari agent` | `internal/cmd/agent` | Agent management |
| `ari tribute` | `internal/cmd/tribute` | Session summary operations |
| `ari init` | `internal/cmd/initialize` | Project initialization |
| `ari provenance` | `internal/cmd/provenance` | Display provenance manifest |
| `ari org` | `internal/cmd/org` | Manage organizations |
| `ari registry` | `internal/cmd/registry` | Manage the org knowledge domain registry |
| `ari land` | `internal/cmd/land` | Manage cross-session knowledge synthesis |
| `ari ledge` | `internal/cmd/ledge` | Work product artifact management |
| `ari lint` | `internal/cmd/lint` | Validate mena and agent sources |
| `ari status` | `internal/cmd/status` | Show unified project health dashboard |
| `ari explain` | `internal/cmd/explain` | Explain a knossos concept |
| `ari tour` | `internal/cmd/tour` | Walk project directory structure |
| `ari ask` | `internal/cmd/ask` | Find the right command for any task |
| `ari complaint` | `internal/cmd/complaint` | Manage complaints |
| `ari serve` | `internal/cmd/serve` | Start the HTTP webhook server (Clew) |
| `ari version` | `internal/cmd/root` | Show version information |

### Key Exported Interfaces

| Interface | Package | Consuming Packages | Description |
|-----------|---------|-------------------|-------------|
| `output.Textable` | `internal/output` | All cmd packages | Implement `Text() string` for human-readable output |
| `output.Tabular` | `internal/output` | All cmd packages | Implement for table rendering |
| `hook.LifecycleAdapter` | `internal/hook` | `internal/cmd/hook` | Per-channel hook parsing/formatting |
| `config.OrgContext` | `internal/config` | `internal/materialize` | Org-scoped configuration |
| `resolution.Chain` | `internal/resolution` | `internal/rite`, `internal/materialize` | Multi-tier resource resolution |
| `llm.Client` | `internal/llm` | `internal/triage`, `internal/reason/response`, `internal/search/knowledge`, `internal/slack/conversation` | LLM transport |
| `observe.MetricsRecorder` | `internal/observe` | `internal/cmd/serve`, `internal/slack` | Observability |

---

## Key Abstractions

### Core Types and Their Usage

**1. `materialize.Materializer` (`internal/materialize/materialize.go`)** — The central engine. Builder-pattern construction with `NewMaterializer(resolver)` and `.WithEmbeddedFS()` chain. Primary methods: `MaterializeWithOptions(riteName, opts)`, `Sync(opts)`.

**2. `resolution.Chain` (`internal/resolution/chain.go`)** — Five-tier priority resolution: project > user > org > platform > embedded. Zero internal imports — all paths injected. `chain.Resolve(name, validate)` returns first match; `chain.ResolveAll(validate)` returns all with shadowing.

**3. `session.FSM` + `session.Context` (`internal/session/`)** — Session lifecycle state machine. Status: NONE -> ACTIVE <-> PARKED -> ARCHIVED. Phase: requirements -> design -> implementation -> validation -> complete.

**4. `hook.StdinPayload` (`internal/hook/env.go`)** — Critical operational knowledge: CC sends all hook data as JSON on stdin, NOT via environment variables. Only four env vars exist: `CLAUDE_PROJECT_DIR`, `CLAUDE_PLUGIN_ROOT`, `CLAUDE_CODE_REMOTE`, `CLAUDE_ENV_FILE`.

**5. `rite.RiteManifest` (`internal/rite/manifest.go`)** — Schema for `manifest.yaml`. Key fields: `Name`, `Version`, `EntryAgent`, `Phases`, `Agents`, `Skills`, `Dependencies`, `ComplexityLevels`, `Budget`.

**6. `frontmatter.FlexibleStringSlice` (`internal/frontmatter/`)** — YAML type accepting both comma-separated strings and proper YAML lists. Prevents agent frontmatter breakage.

**7. `hook/clewcontract.Event` (`internal/hook/clewcontract/event.go`)** — Canonical event record for `events.jsonl`. ADR-0027 established this as the unified write path.

**8. `paths.Resolver` (`internal/paths/paths.go`)** — Central path authority. All `.sos/` paths computed from project root. Per-channel: `AgentsDirForChannel(ch)`, `ContextFileForChannel(ch)`.

**9. `channel.CanonicalTool` (`internal/channel/`)** — Harness-agnostic tool vocabulary map. Templates reference canonical names; channel compiler resolves at projection time.

**10. `output.Printer` (`internal/output/output.go`)** — Shared output abstraction. `NewPrinter(format, out, errOut, verbose)`. Data types implement `Textable` for human-readable mode.

**11. `know.QualifiedDomainName` (`internal/know/qualified.go`)** — Parses `"org::repo::domain"` strings. Enforces exactly two `::` separators; `##section` suffixes stripped at parse time.

**12. `reason.Pipeline` (`internal/reason/pipeline.go`)** — Clew reasoning orchestrator. Three query methods: `Query()` (full pipeline), `QueryWithTriage()` (pre-computed triage), `QueryStream()` (streaming with `onChunk` callback).

### Design Patterns

**Cascade/Merge Pattern (Agent Defaults):** Manifest `agent_defaults` -> merged into each agent frontmatter. Agent values win. Source: `internal/materialize/agent_transform.go`.

**Envelope Pattern (Hook Output):** Legacy `Result` and CC-native `PreToolUseOutput` coexist. CC reads `permissionDecision` from `hookSpecificOutput`, not top-level.

**Registry Pattern (Org Domain Catalog):** `internal/registry/org.DomainCatalog` maps qualified names to `DomainEntry` structs.

**Idempotency Invariant:** `materialize.Sync()` is idempotent. `writeIfChanged()` prevents unnecessary CC file watcher triggers.

**Fail-Open Pattern:** Used in triage (Stage 3 fails -> use Stage 2 scores), hooks (errors default to allow), Clew serve startup (missing deps log warnings, server starts anyway).

**Mena Routing Convention:** `.dro.md` -> `commands/` (dromena: transient). `.lego.md` -> `skills/` (legomena: persistent).

**Three-tier Agent Lifecycle:** `standing` (always materialized), `rite` (via `ari sync`), `summonable` (via `ari agent summon`).

---

## Data Flow

### 1. Sync Pipeline (Materialization)

```
Source files (rites/{name}/, mena/, knossos/templates/, agents/)
    |
  resolution.Chain (project > user > org > platform > embedded FS)
    |
  materialize.Materializer.Sync()
    +-- Stage: Generate CLAUDE.md (inscription package, region-based merge)
    +-- Stage: Write agents/ (agent_transform: strip knossos fields, inject defaults)
    +-- Stage: Write commands/ (mena .dro.md -> strip extension)
    +-- Stage: Write skills/ (mena .lego.md -> strip extension)
    +-- Stage: Write hooks (hook wiring, defaults cascade)
    +-- Stage: Write settings.json (MCP servers, permissions)
    +-- Stage: Write rules/ (rules projection)
    +-- Stage: Update PROVENANCE_MANIFEST.yaml (provenance.Collector)
    |
  .claude/ or .gemini/ channel directory
```

### 2. Session Lifecycle Pipeline

```
User invokes ari session create -> session.Context written to .sos/sessions/{id}/SESSION_CONTEXT.md
    |
  CC session start -> PreToolUse hook fires -> hook.StdinPayload read from stdin
    |
  internal/cmd/hook/{handler}.go dispatches by event type
    |
  ari session park/resume/wrap -> FSM.ValidateTransition() -> context YAML updated
    |
  session events -> .sos/sessions/{id}/events.jsonl
```

### 3. Hook Pipeline

```
CC lifecycle event fires (e.g., PreToolUse, SessionStart, Stop)
    |
  JSON payload sent to hook binary on stdin (StdinPayload struct)
    |
  internal/hook/env.go ReadStdin() -> parse HookEventName, ToolName, ToolInput
    |
  internal/cmd/hook/{handler}.go dispatches by event type
    |
  Output to stdout: PreToolUseOutput{hookSpecificOutput: {permissionDecision: "allow"/"deny"}}
```

### 4. Clew Serve Pipeline (HTTP -> Slack -> Reasoning -> Response)

```
POST /slack/events
    |
  serve/webhook.Verifier (HMAC-SHA256 signature check)
    |
  serve.ConcurrencyLimit middleware (max_concurrent from cfg.MaxConcurrent)
    |
  observe.OTELMiddleware (trace context propagation)
    |
  slack.SlackHandler.ServeHTTP()
    +-- eventDedup (TTL-based dedup, Slack retry protection)
    +-- ThreadContextStore lookup
    +-- ConversationManager.GetThreadHistory()
    |
    +-- triage.Orchestrator.Assess() [4-stage RAG triage]
        +-- Stage 0: llm.Client.Complete() query refinement (follow-ups only)
        +-- Stage 1: metadata pre-filter (zero cost)
        +-- Stage 2: embedding search OR BM25 fallback
        +-- Stage 3: llm.Client.Complete() Haiku domain scoring
    |
    +-- reason.Pipeline.QueryStream() [streaming path]
        +-- triageCandidatesToSearchResults() + contentLookup() -> real .know/ content
        +-- trust.Scorer.Score() (freshness x retrieval x coverage geometric mean)
        +-- reason/context.Assembler.Assemble() (token-budgeted context)
        +-- reason/response.Generator.GenerateStream() -> Anthropic API (streaming)
    |
    +-- slack/streaming.Sender (progressive rendering to Slack)
    |
  observe.EMFRecorder (CloudWatch metrics via structured slog)
```

**Configuration hierarchy for `ari serve`:** CLI flags -> process env vars -> org env file -> hardcoded defaults. Managed by `internal/envload.Load()`.

**Knowledge index startup:** Server starts immediately with BM25 fallback. `KnowledgeIndex` builds in background goroutine (10-minute timeout). On completion, pipeline upgrades to full semantic search.

### 5. `ari ask` Search Pipeline

```
User: ari ask "what is a rite?"
    |
  search.Build(rootCmd, resolver) -> collect commands, concepts, rites, agents, dromena, routing, knowledge domains
    |
  SearchIndex.Search(query) -> BM25 + RRF fusion
    |
  Ranked SearchEntry results -> printed via output.Printer
```

---

## Knowledge Gaps

1. **`internal/materialize/compiler/` sub-package** — 57-file package with 7 sub-packages not individually documented.
2. **`internal/inscription/` full merge pipeline** — Region-based merge engine not fully traced.
3. **`internal/reason/` LLM pipeline stages** — Pipeline configuration and stage boundaries not traced end-to-end.
4. **`internal/registry/org/` sub-package** — Referenced by `internal/search` imports but not individually examined.
5. **`internal/materialize/source/` SourceResolver** — Tier enumeration logic not traced.
6. **Hook sub-handler details** — 16+ hook handler files each not individually examined for decision logic.
7. **Org scope and registry** — `internal/cmd/org/` and `internal/cmd/registry/` manage org-level configuration not traced.

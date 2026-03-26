---
domain: architecture
generated_at: "2026-03-26T17:14:25Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "a73d68a6"
confidence: 0.84
format_version: "1.0"
update_mode: "incremental"
incremental_cycle: 1
max_incremental_cycles: 3
land_sources:
  - ".sos/land/initiative-history.md"
land_hash: "77ed0121d9982cb4e1bf8e7c09b414494a561c0ccd3b9ea546d210acd7354553"
---

# Codebase Architecture

## Package Structure

The knossos repository is a Go module (`github.com/autom8y/knossos`) with two major subsystems sharing a single binary: the **ari CLI** (agentic workflow harness) and the **Clew Slack bot** (organizational intelligence).

### Entry Points

**`cmd/ari/main.go`** — Single entry point. Minimal wiring:
1. Sets version info (injected at build time via `ldflags`)
2. Calls `common.SetEmbeddedAssets` and `common.SetEmbeddedUserAssets` to register embedded FSes
3. Delegates to `root.Execute()`

**`embed.go`** (module root) — Holds all `//go:embed` directives, creating five embedded FSes: `EmbeddedRites`, `EmbeddedTemplates`, `EmbeddedHooksYAML`, `EmbeddedAgents`, `EmbeddedMena`, `EmbeddedProcessions`. These make knossos a single distributable binary.

### `cmd/` Directory

| Directory | Purpose |
|---|---|
| `cmd/ari/` | Binary entry point, single `main.go` file |

### `internal/` Package Inventory

**CLI Surface layer (`internal/cmd/`)**

| Package | Files | Purpose |
|---|---|---|
| `internal/cmd/root` | `root.go` | Cobra root command, registers all 35+ subcommands, global flags |
| `internal/cmd/common` | 6 files | `BaseContext`, `SessionContext`, embedded asset refs, annotation-based `NeedsProject` |
| `internal/cmd/agent` | 12 files | `ari agent` subcommands: `summon`, `dismiss`, `roster`, `list`, `new`, `embody`, `update`, `validate`, `query` |
| `internal/cmd/artifact` | — | `ari artifact` subcommands |
| `internal/cmd/ask` | `ask.go`, `models.go` | `ari ask` — semantic search over codebase via `SearchIndex` |
| `internal/cmd/complaint` | — | `ari complaint` — file framework friction reports |
| `internal/cmd/explain` | — | `ari explain` |
| `internal/cmd/handoff` | — | `ari handoff` |
| `internal/cmd/hook` | 25+ files | `ari hook` subcommands — all hook handlers (writeguard, autopark, agentguard, clew, gitconventions, etc.) |
| `internal/cmd/initialize` | — | `ari init` |
| `internal/cmd/inscription` | — | `ari inscription` — CLAUDE.md management |
| `internal/cmd/knows` | — | `ari know` |
| `internal/cmd/land` | — | `ari land` |
| `internal/cmd/ledge` | — | `ari ledge` |
| `internal/cmd/lint` | — | `ari lint` |
| `internal/cmd/manifest` | — | `ari manifest` |
| `internal/cmd/naxos` | — | `ari naxos` |
| `internal/cmd/org` | — | `ari org` |
| `internal/cmd/procession` | — | `ari procession` |
| `internal/cmd/provenance` | — | `ari provenance` |
| `internal/cmd/registry` | — | `ari registry` |
| `internal/cmd/rite` | — | `ari rite` — rite management |
| `internal/cmd/sails` | — | `ari sails` |
| `internal/cmd/serve` | `serve.go`, `serve_test.go`, `query.go`, `query_test.go` | `ari serve` — full Clew HTTP server startup and wiring |
| `internal/cmd/session` | 25+ files | `ari session` subcommands (create, park, resume, wrap, audit, log, fray, gc, migrate, etc.) |
| `internal/cmd/status` | — | `ari status` |
| `internal/cmd/sync` | `sync.go`, `budget.go` | `ari sync` — materialization trigger |
| `internal/cmd/tour` | — | `ari tour` |
| `internal/cmd/tribute` | — | `ari tribute` |
| `internal/cmd/validate` | — | `ari validate` |
| `internal/cmd/worktree` | — | `ari worktree` |

**Domain Logic layer (`internal/`)**

| Package | Key Types | Purpose |
|---|---|---|
| `internal/agent` | `AgentFrontmatter`, `FlexibleStringSlice`, `MemoryField`, `BehavioralContract` | Agent frontmatter parsing, validation, 3 archetypes |
| `internal/channel` | `manifest.go`, `output.go`, `rite.go` | Channel directory abstractions; harness agnosticism utilities |
| `internal/checksum` | — | SHA-256 checksum utilities (prefix format: `sha256:`) |
| `internal/config` | `OrgContext`, `RepoConfig`, `Settings` | XDG config home, org context, active-org resolution |
| `internal/errors` | `AriError`, 15+ exit codes, 20+ error codes | Domain error types, exit codes, JSON-serializable errors |
| `internal/envload` | `ServeConfig`, `Overrides` | Config resolution for `ari serve`: flags > env > org env file > defaults. `MaxConcurrent` is a first-class field plumbed from CLI flag through to `SlackHandler` and `ConcurrencyLimit` middleware. |
| `internal/fileutil` | — | Atomic file write, directory utilities |
| `internal/frontmatter` | — | YAML frontmatter `---` block parsing |
| `internal/hook` | `StdinPayload`, `HookEvent`, canonical↔wire event map, `clewcontract/` | Hook payload parsing, event naming, CC contract JSONL writer |
| `internal/inscription` | `RenderContext`, `AgentInfo`, `SummonableAgentInfo`, generator | CLAUDE.md generation: region-based merge, satellite region preservation |
| `internal/know` | `QualifiedDomainName`, `Parse()`, `String()` | `.know/` domain management; parses `"org::repo::domain"` strings, enforcing exactly-2-`::` format and rejecting `::` inside domain segment |
| `internal/ledge` | — | `.ledge/` artifact management |
| `internal/lock` | `LockMetadata` | JSON v2 lock files, 5-minute stale threshold |
| `internal/manifest` | `RiteManifest` (in materialize), diff/merge/schema | Rite manifest YAML operations |
| `internal/materialize` | `Materializer`, `RiteManifest`, `Options`, `Result`, `TransformContext` | Core sync pipeline: source -> transform -> channel dir output |
| `internal/mena` | `StripMenaExtension`, `RouteMenaFile`, `DetectMenaType` | Mena file routing: `.dro.md`->`commands/`, `.lego.md`->`skills/` |
| `internal/naxos` | — | Session hygiene utilities |
| `internal/observe` | `MetricsRecorder` (interface), `emfMetrics`, `CostTracker`, OTEL middleware | CloudWatch EMF metrics, OTEL tracing, structured logging for Clew serve |
| `internal/output` | `Printer`, `Format` (text/json/yaml) | Format-aware output (tabwriter, JSON, YAML) |
| `internal/paths` | `Resolver`, `FindProjectRoot` | XDG path resolution, project root discovery |
| `internal/perspective` | — | Perspective/viewpoint utilities |
| `internal/procession` | `template.go` | Procession (coordinated workflow) generation |
| `internal/provenance` | `ProvenanceManifest`, `ProvenanceEntry`, `OwnerType`, `ScopeType`, `Collector` | File provenance tracking v2.0: two manifests (rite + user scope) |
| `internal/registry` | `RefKey`, `RefEntry`, `Registry` | LEAF package: stable key->value map; zero internal imports |
| `internal/registry/org` | `DomainEntry`, `DomainCatalog`, `OrgContext` | Clew knowledge address space: cross-repo `.know/` domain catalog |
| `internal/resolution` | `Chain`, `Tier`, `ResolvedItem` | Multi-tier resolution (project > user > org > platform > embedded); zero internal imports |
| `internal/rite` | `Rite`, `Discovery`, `BudgetCalculator` | Rite discovery via resolution chain, budget estimation |
| `internal/sails` | `contract.go`, `gate.go`, `color.go` | Sails (session health/color) contract |
| `internal/search` | `SearchIndex`, collectors (commands/concepts/rites/agents/dromena/routing) | `ari ask` BM25+RRF search over project artifacts |
| `internal/search/bm25` | `Index` | BM25 text search index, decay model |
| `internal/search/content` | — | Content extraction for search |
| `internal/search/fusion` | RRF | Reciprocal Rank Fusion |
| `internal/search/knowledge` | `KnowledgeIndex`, `BM25Searcher` interface, `persistedIndex` | Clew-specific semantic index: BM25 + embedding + graph + summary stores. `DefaultPersistedPath` is the container fallback; runtime path resolved by `resolveKnowledgeIndexPath()` in `internal/cmd/serve`. |
| `internal/search/knowledge/embedding` | `Store` | Cosine similarity search over domain embeddings |
| `internal/search/knowledge/graph` | `Graph` | Entity relationship graph |
| `internal/search/knowledge/summary` | `Store` | Domain summary store |
| `internal/session` | `Context`, `Status`, `FSM`, `Event`, `ClewEvent` | Session lifecycle (NONE->ACTIVE->PARKED->ARCHIVED FSM), context YAML, event log |
| `internal/serve` | `Server`, `ServerConfig`, `Option`, `Middleware` | HTTP server infrastructure: mux, health checks, middleware chain, graceful shutdown |
| `internal/serve/health` | `Checker` | Health check registry for `/health` and `/ready` endpoints |
| `internal/serve/webhook` | `Verifier` | HMAC-SHA256 Slack signature verification |
| `internal/slack` | `SlackHandler`, `ThreadContextStore`, `eventDedup`, `SlackClient` | Slack event handler: dedup, thread context, event dispatch |
| `internal/slack/conversation` | `Manager`, `Metrics`, `Summarizer` | Thread history with TTL eviction and LLM summarization |
| `internal/slack/streaming` | `Sender`, `citations.go` | Progressive message rendering to Slack |
| `internal/suggest` | — | Suggestion utilities |
| `internal/sync` | (see materialize) | Sync state management |
| `internal/tokenizer` | — | Token counting (tiktoken-go) |
| `internal/triage` | `Orchestrator`, 4-stage pipeline | RAG triage: Stage 0 (query refine), Stage 1 (metadata filter), Stage 2 (embedding/BM25), Stage 3 (Haiku assess) |
| `internal/tribute` | `extractor.go`, `generator.go`, `renderer.go` | Tribute (session report) generation |
| `internal/trust` | `Scorer`, `ConfidenceScore`, `ConfidenceTier`, `TrustConfig` | Multi-axis confidence: freshness x retrieval x coverage -> geometric mean |
| `internal/llm` | `Client` (interface), `CompletionRequest`, `ClientConfig` | LLM client abstraction: Anthropic transport, cost tracking, metrics, tracing |
| `internal/reason` | `Pipeline`, `TriageCandidateInput`, `TriageResultInput` | Top-level Clew reasoning orchestrator: intent->retrieval->trust->context->response. `contentLookup()` resolves qualified names to raw `.know/` content via BM25 index. |
| `internal/reason/context` | `Assembler` | Context assembly with token budgeting |
| `internal/reason/intent` | `Classifier` | Keyword-heuristic query classification (Observe/Record/Act tiers) |
| `internal/reason/response` | `Generator`, `ReasoningResponse` | Claude API call with structured output schema (`clew_answer`) or streaming free-form text |
| `internal/validation` | — | YAML/JSON schema validation |
| `internal/worktree` | — | Git worktree management |

**Hub vs Leaf classification:**

- **Hub packages** (import many siblings): `internal/cmd/serve` (imports 15+ packages), `internal/cmd/root` (imports 30+ cmd packages), `internal/materialize` (imports: frontmatter, paths, provenance, registry, compiler, mena, source, etc.)
- **Leaf packages** (no internal imports): `internal/registry`, `internal/resolution`, `internal/mena`, `internal/errors`, `internal/output`

---

## Layer Boundaries

The codebase has four clearly separable layers, enforced primarily by convention and a few explicit import restrictions documented in source:

```
+-------------------------------------------------------------+
|  CLI Surface  cmd/ari/main.go -> internal/cmd/root           |
|              internal/cmd/* (35+ command packages)            |
+-------------------------------------------------------------+
|  Command Wiring  internal/cmd/common (BaseContext,            |
|                  SessionContext, embedded assets)              |
+-------------------------------------------------------------+
|  Domain Logic  internal/{session, rite, materialize,          |
|                provenance, hook, inscription, search,         |
|                reason, trust, triage, slack, serve, llm}      |
+-------------------------------------------------------------+
|  Infrastructure/Leaf  internal/{errors, output, paths,        |
|                         registry, resolution, mena,           |
|                         frontmatter, fileutil, checksum,      |
|                         tokenizer, channel, config}            |
+-------------------------------------------------------------+
```

**Documented import invariants** (from source comments):

- `internal/reason/pipeline.go`: "reason/ does NOT import internal/serve/ or internal/cmd/. trust/ and search/ do NOT import reason/"
- `internal/reason/pipeline.go` (TriageCandidateInput): "reason/ does NOT import triage/. The handler in slack/ converts triage.TriageCandidate to this type."
- `internal/resolution/chain.go`: "ZERO internal imports. All tier paths are injected via constructor to avoid import cycles (TENSION-005)"
- `internal/registry/registry.go`: "LEAF package -- imports only stdlib"
- `internal/provenance`: "One-way dependency: materialize imports provenance, never the reverse"

**Clew serve pipeline layer invariants** (from serve.go and triage, reason, trust packages):
```
internal/cmd/serve -> internal/serve (HTTP infra)
                   -> internal/slack (handler)
                   -> internal/reason (pipeline)
                       -> internal/reason/intent
                       -> internal/reason/context
                       -> internal/reason/response -> (external: Anthropic SDK)
                       -> internal/trust
                       -> internal/search (BM25+RRF)
                   -> internal/triage (RAG stages 0-3)
                       -> internal/llm (Haiku API)
                   -> internal/search/knowledge (KnowledgeIndex)
                   -> internal/observe (CloudWatch EMF)
                   -> internal/llm (shared transport)
```

**Import direction summary:**
- `cmd/*` -> domain logic packages (never reverse)
- `materialize` -> `provenance`, `paths`, `frontmatter`, `mena`, `agent`, `registry` (never reverse)
- `reason` -> `trust`, `search`, `registry/org` (trust/search never import reason)
- `reason` does NOT import `triage` -- data transfer via `TriageCandidateInput`/`TriageResultInput` structs; `slack/` performs the conversion
- `resolution`, `registry`, `errors`, `output`, `mena` are leaves: no internal imports

---

## Entry Points and API Surface

### CLI Entry Point Trace

```
cmd/ari/main.go -> root.SetVersion()
               -> common.SetEmbeddedAssets(EmbeddedRites, EmbeddedTemplates, EmbeddedHooksYAML)
               -> common.SetEmbeddedUserAssets(EmbeddedAgents, EmbeddedMena)
               -> common.SetEmbeddedProcessions(EmbeddedProcessions)
               -> root.Execute() -> rootCmd.Execute() [cobra]
```

### Global Flags on Root Command

`-o/--output` (text/json/yaml), `-v/--verbose`, `--config`, `-p/--project-dir`, `-s/--session-id`, `--channel` (claude/gemini/all)

### `PersistentPreRunE` initialization path

For each command: validate output format -> `initConfig()` (viper: XDG config file + env vars) -> `paths.FindProjectRoot()` -> set `globalOpts.ProjectDir`.

### CLI Subcommand Table

| Subcommand | Package | Description |
|---|---|---|
| `ari session` | `internal/cmd/session` | Session lifecycle: start, park, resume, wrap, status, log, audit, migrate, fray, snapshot |
| `ari manifest` | `internal/cmd/manifest` | Rite manifest operations |
| `ari inscription` | `internal/cmd/inscription` | CLAUDE.md generation and management |
| `ari sync` | `internal/cmd/sync` | Materialize rite to channel directory |
| `ari validate` | `internal/cmd/validate` | Validate rite/manifest/agent structures |
| `ari handoff` | `internal/cmd/handoff` | Session handoff between agents |
| `ari procession` | `internal/cmd/procession` | Coordinated workflow processions |
| `ari worktree` | `internal/cmd/worktree` | Git worktree management |
| `ari hook` | `internal/cmd/hook` | Hook handlers: writeguard, autopark, agentguard, clew, gitconventions, sessionend, subagent, precompact, driftdetect, cheapo_revert, validate, suggest |
| `ari know` | `internal/cmd/knows` | Knowledge domain management |
| `ari artifact` | `internal/cmd/artifact` | Session artifact management |
| `ari sails` | `internal/cmd/sails` | Sails (session health) contracts |
| `ari naxos` | `internal/cmd/naxos` | Session hygiene |
| `ari rite` | `internal/cmd/rite` | Rite listing, switching, validation |
| `ari agent` | `internal/cmd/agent` | Agent management: summon, dismiss, roster, list, new, embody, update, validate, query |
| `ari tribute` | `internal/cmd/tribute` | Session tribute/report generation |
| `ari init` | `internal/cmd/initialize` | Initialize knossos in a new project |
| `ari provenance` | `internal/cmd/provenance` | Provenance manifest inspection |
| `ari org` | `internal/cmd/org` | Org management |
| `ari registry` | `internal/cmd/registry` | Registry operations |
| `ari land` | `internal/cmd/land` | Knowledge synthesis landing |
| `ari ledge` | `internal/cmd/ledge` | Ledge artifact management |
| `ari lint` | `internal/cmd/lint` | Lint rite/agent files |
| `ari status` | `internal/cmd/status` | Project/session status |
| `ari explain` | `internal/cmd/explain` | Explain knossos concepts |
| `ari tour` | `internal/cmd/tour` | Interactive tour |
| `ari ask` | `internal/cmd/ask` | Semantic search over project artifacts (BM25+RRF) |
| `ari complaint` | `internal/cmd/complaint` | File Cassandra complaints |
| `ari serve` | `internal/cmd/serve` | Start Clew HTTP webhook server |
| `ari serve query` | `internal/cmd/serve` | Query the reasoning pipeline directly (subcommand of serve) |
| `ari version` | `internal/cmd/root` | Print version, commit, date, Go/OS/Arch |

### Key Exported Interfaces

| Interface | Package | Consumers |
|---|---|---|
| `output.Printer` | `internal/output` | Every `internal/cmd/*` package |
| `paths.Resolver` | `internal/paths` | `internal/cmd/common`, most domain packages |
| `provenance.Collector` | `internal/provenance` | `internal/materialize` (threaded through pipeline stages) |
| `llm.Client` | `internal/llm` | `internal/triage`, `internal/reason/response`, `internal/search/knowledge`, `internal/slack/conversation` |
| `observe.MetricsRecorder` | `internal/observe` | `internal/cmd/serve`, `internal/slack` |
| `triage.SearchIndex` (local interface) | `internal/triage` | `internal/triage.Orchestrator` |
| `resolution.Chain` | `internal/resolution` | `internal/rite.Discovery`, `internal/materialize` |
| `config.OrgContext` | `internal/config` | `internal/registry/org`, `internal/envload` |

---

## Key Abstractions

### Core Types and Their Usage

**1. `paths.Resolver` (`internal/paths/paths.go`)** — Central path resolver. Injected into every command via `common.BaseContext.GetResolver()`. Methods: `ProjectRoot()`, `KnossosDir()`, `SOSDir()`, `SessionsDir()`, `RitesDir()`, `UserChannelDir()`.

**2. `materialize.RiteManifest` (`internal/materialize/materialize.go`)** — YAML schema for `manifest.yaml`. Drives the entire sync/materialization pipeline.

**3. `agent.AgentFrontmatter` (`internal/agent/frontmatter.go`)** — Agent `.md` file frontmatter. Key fields: `Name`, `Description`, `Tier` (standing/rite/summonable), `Tools`, `MaxTurns`, `Skills`, `DisallowedTools`, `Memory`, `WriteGuard`.

**4. `session.Context` (`internal/session/context.go`)** — Session state document. Written only by Moirai agent or ari CLI -- never directly.

**5. `session.FSM` (`internal/session/fsm.go`)** — State machine transitions per TLA+ spec. NONE->ACTIVE, ACTIVE->{PARKED, ARCHIVED}, PARKED->{ACTIVE, ARCHIVED}. ARCHIVED is terminal.

**6. `provenance.ProvenanceManifest` (`internal/provenance/provenance.go`)** — File provenance tracking v2.0: two manifests (rite + user scope). SHA-256 checksums.

**7. `hook.StdinPayload` (`internal/hook/env.go`)** — CC lifecycle hook data. Critical: CC sends hook data as JSON on stdin, NOT environment variables.

**8. `resolution.Chain` (`internal/resolution/chain.go`)** — Multi-tier resolution: project > user > org > platform > embedded. Zero internal imports.

**9. `materialize.TransformContext` (`internal/materialize/agent_transform.go`)** — Agent content transformation policy. Strips knossos-only frontmatter fields before writing to channel directory.

**10. `reason.Pipeline` (`internal/reason/pipeline.go`)** — Clew reasoning orchestrator. Three query methods:
- `Query()` -- full pipeline (intent classify -> BM25 search -> trust -> assemble -> generate)
- `QueryWithTriage()` -- uses pre-computed `TriageResultInput`; delegates to `Query()` if no candidates
- `QueryStream()` -- streaming with `onChunk` callback (free-form text with inline citation markers)

`contentLookup()` method (Sprint 4-prime) returns a closure over `search.SearchIndex.LookupContent` so `triageCandidatesToSearchResults()` populates `SearchEntry.Description` with real `.know/` content instead of empty stubs.

**11. `know.QualifiedDomainName` (`internal/know/qualified.go`)** — Parses `"org::repo::domain"` strings. Enforces exactly two `::` separators; `##section` suffixes stripped/rejected at parse time for citation normalization.

### Design Patterns

**Cascade/Merge Pattern (Agent Defaults):** Manifest `agent_defaults` -> merged into each agent frontmatter. Agent values win. Source: `internal/materialize/agent_transform.go`.

**Envelope Pattern (Hook Output):** Legacy `Result` and CC-native `PreToolUseOutput` coexist. CC reads `permissionDecision` from `hookSpecificOutput`, not top-level.

**Registry Pattern (Org Domain Catalog):** `internal/registry/org.DomainCatalog` maps `"org::repo::domain"` qualified names to `DomainEntry` structs.

**Idempotency Invariant:** `materialize.Sync()` is idempotent. `writeIfChanged()` prevents unnecessary CC file watcher triggers.

**Fail-Open Pattern:** Used in triage (Stage 3 fails -> use Stage 2 scores), hooks (errors default to allow), Clew serve startup (missing deps log warnings, server starts anyway).

**Mena Routing Convention:** `.dro.md` -> `commands/` (dromena: transient). `.lego.md` -> `skills/` (legomena: persistent).

**Three-tier Agent Lifecycle:** `standing` (always materialized), `rite` (via `ari sync`), `summonable` (via `ari agent summon`).

**Citation Provenance Rule (Sprint 4-prime):** HIGH and MEDIUM tier system prompts include explicit `CITATION PROVENANCE RULE`: Claude may only cite sources listed in `KNOWLEDGE SOURCES` section. Citing qualified names from within content bodies is prohibited.

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

**Configuration hierarchy for `ari serve`:** CLI flags -> process env vars -> org env file (`$XDG_DATA_HOME/knossos/orgs/{org}/serve.env`) -> hardcoded defaults. Managed by `internal/envload.Load()`. `MaxConcurrent` passes through the full hierarchy to both `serve.ConcurrencyLimit` middleware and `slack.HandlerDeps.MaxConcurrent`.

**Knowledge index path resolution (Sprint 4-prime):** `resolveKnowledgeIndexPath()` in `internal/cmd/serve/serve.go` resolves: `CLEW_KNOWLEDGE_INDEX_PATH` env var -> `config.XDGDataDir()/knowledge-index.json` -> container fallback. Fixes Sprint 3 bug where persist path was hardcoded to container nonroot path.

**Knowledge index startup:** Server starts immediately with BM25 fallback. `KnowledgeIndex` builds in background goroutine (10-minute timeout). On completion, pipeline upgrades to full semantic search.

**Content population (Sprint 4-prime):** `triageCandidatesToSearchResults()` now takes a `contentLookup` function (closure over `SearchIndex.LookupContent`) to populate `Description` with real `.know/` content instead of empty stubs. Applies to both `QueryWithTriage()` and `QueryStream()`.

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

1. **`internal/cmd/session` subcommand details**: 25+ files not individually read.
2. **`internal/hook/clewcontract/`**: 16 event types and BufferedEventWriter not individually read.
3. **`internal/inscription/` full merge pipeline**: Region-based merge engine not fully traced.
4. **`internal/materialize/compiler/`**: Archetype compilation sub-package not examined.
5. **`internal/materialize/source/`**: `SourceResolver` type not fully traced.
6. **`internal/procession/`, `internal/tribute/`, `internal/naxos/`, `internal/sails/`, `internal/worktree/`, `internal/perspective/`**: Package purposes identified from listings; types not read.
7. **`internal/search/knowledge/builder.go`**: Background build path not read in detail.
8. **`internal/llm/` remaining files** (cost.go, metrics.go, tracer.go): Cost tracking infrastructure not fully read.
9. **`internal/sync/`**: State management not examined directly.
10. **Gemini channel specifics**: `--channel gemini` path not traced in detail.
11. **`internal/triage/orchestrator.go` Stage 3**: Full Haiku scoring prompt not read.

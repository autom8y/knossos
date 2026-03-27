---
domain: architecture
generated_at: "2026-03-25T12:11:05Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "e43ba47"
confidence: 0.88
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/initiative-history.md"
land_hash: "f31e74621ecfd923b5479ae0dad1b28ed0420173e53031dbe1b9d6698ef7a107"
---

# Codebase Architecture

## Package Structure

**Language**: Python 3.12. Build system: Hatchling. Package manager: uv. Entry package: `autom8_devconsole` (distribution name `autom8-devconsole`).

**Source root**: `src/autom8_devconsole/`

### Top-Level Modules (src/autom8_devconsole/)

| Module | Purpose | Key Types |
|---|---|---|
| `__main__.py` | Entry point; TUI vs CLI dispatch | `main()` |
| `app.py` | FastAPI app factory; lifespan wiring | `create_app()`, `lifespan()`, `_build_nicegui_page()` |
| `app_state.py` | Typed documentation of `app.state` shape | `AppState` (TypedDict, docs only) |
| `config.py` | Pydantic-settings configuration | `DevconsoleSettings`, `get_settings()` |
| `protocols.py` | FROZEN lens plugin contracts | `LensProtocol`, `DetailPanelProtocol`, `LensCapability`, `LensContext`, `SpanProvider` |
| `protocols_v2.py` | Era-3 page builder contracts | `PageContext`, `PageRegistration` |
| `registry.py` | Entry-point lens discovery + lifecycle | `LensRegistry` |
| `span_buffer.py` | In-memory ring buffer for OTel spans | `ParsedSpan`, `SpanBuffer`, `SessionIdentifierResolver`, `DefaultSessionResolver` |
| `otlp_receiver.py` | FastAPI router: POST /v1/traces | `otlp_router`, `create_otlp_router()` |
| `otlp_codec.py` | Protobuf/JSON OTLP span parsing | `parse_json_span()`, `parse_json_attributes()` |
| `tempo_client.py` | Grafana Tempo HTTP client for historical replay | `TempoClient`, `TempoTraceResult`, `SpanMetricsResult` |
| `persistence.py` | SQLite WAL-mode span durability | `SpanStore` |
| `narrative.py` | Core semantic convention mapping engine | `NarrativeEngine`, `get_core_rules()`, `RuleSetMatchReport` |
| `narrative_pipeline.py` | Reconciliation pipeline narrative plugin | `get_pipeline_rules()` |
| `narrative_request_response.py` | HTTP request/response narrative plugin | `get_request_response_rules()` |
| `narrative_computation.py` | Computation span narrative plugin | `get_computation_rules()` |
| `trace_narrative.py` | Act-based narrative composition | `TraceNarrativeBuilder`, `TraceNarrative`, `NarrativeAct` |
| `deterministic_output.py` | LLM/MCP protocol boundary | `DeterministicOutput`, `TraceDeterministicOutput`, `DetectedAnomaly`, `SideEffectRecord`, `extract_anomalies()`, `serialize_for_llm()`, `serialize_for_mcp()` |
| `intelligence_service.py` | Mediator: backends → UI | `TraceIntelligenceService`, `AnalysisResult` |
| `llm_analysis.py` | Optional LLM trace explanation | `TraceAnalyzer`, `TraceAnalysis`, `SpanAnalysis`, `SessionAnalysis` |
| `mcp_routes.py` | MCP server HTTP SSE routes (4 tools) | `is_mcp_available()`, tools: `get_trace_narrative`, `get_session_narrative`, `list_sessions`, `explain_trace` |
| `archetype.py` | Interaction archetype detection | `detect_archetype()` |
| `computation_stages.py` | Computation span classification | `is_computation_trace()` |
| `risk_surface.py` | Heuristic anomaly detection engine | `RiskSurfaceEngine`, `RiskSurfaceConfig`, `detect_side_effect_inversion()`, `detect_pipeline_ordering_violation()` |
| `side_effect_utils.py` | Dual-prefix side-effect attribute normalization | `get_side_effect_attr()`, `is_side_effect_event()` |
| `causality.py` | Causal chain analysis for spans | (causal relationship utilities) |
| `response_interpreter.py` | HTTP response interpretation | `interpret_response()`, `enrich_interpretation()` |
| `ai_companion.py` | AI companion session state (no NiceGUI dep) | `ConversationManager`, `ConversationSession`, `NavigationEvent` |
| `cli.py` | Headless CLI: `summarize` command | `cli_main()` |
| `fixture_recorder.py` | Serialize/list fixture files | `FixtureInfo`, `save_fixture()`, `list_fixtures()`, `serialize_spans()` |
| `supplement.py` | Demo fixture loading at startup | `SupplementLoader`, `MultiFixtureLoader` |
| `testing.py` | Pytest plugin + test utilities | `SpanRecorder`, `make_test_tracer_provider()`, `devconsole_exporter` fixture |
| `sms_conversation_service.py` | SMS HTTP I/O service (distinct from ai_companion) | (SMS/HTTP I/O) |

### Sub-package: `src/autom8_devconsole/fixtures/`

Fixture factories generating synthetic OTLP JSON for demo data and tests.

| Module | Purpose |
|---|---|
| `_helpers.py` | Low-level OTLP JSON builder helpers (`_span()`, `_side_effect_event()`, `assemble_trace()`) |
| `factories.py` | Original booking flow fixture (legacy helpers; mostly superseded by `_helpers.py`) |
| `booking_error.py` | Booking error flow fixture |
| `cancellation_flow.py` | Booking cancellation flow fixture |
| `reconciliation_pipeline.py` | Reconciliation pipeline fixture |
| `cps_computation.py` | CPS computation fixture (3 factories: OTLP, parsed spans, branching) |
| `cps_variants.py` | CPS variant fixtures |
| `computation_enrichment.py` | Deep/entity/batch/materialization/anomalous/failed computation fixtures (6 factories) |
| `edge_cases.py` | Edge-case fixtures: cross-service, degraded, mixed status, reschedule |
| `multi_session.py` | Multi-session pool fixture |
| `request_response.py` | HTTP request/response fixtures (health check, auth, data API) |
| `phantom_computation.py` | Phantom trace first-class fixture |
| `polars_computation.py` | Polars data pipeline fixture |
| `generate.py` | Composite fixture generator (imports from all sub-fixtures) |

### Sub-package: `src/autom8_devconsole/ui/`

NiceGUI page builders, lens implementations, and UI primitives.

| Module | Category | Purpose |
|---|---|---|
| `theme.py` | Design system | OKLCH token palette, `HEX` dict, `SERVICE_COLORS`, CSS helpers |
| `primitives.py` | Design system | Composable component library (Card, Badge, Timeline, MetricChip, etc.) |
| `compositions.py` | Design system | Higher-order compositions (ArchetypeBadge, FixtureCard) built from primitives |
| `motion.py` | Design system | Motion/animation CSS |
| `styles/__init__.py` | Design system | CSS bundle aggregator (imports all 9 style modules) |
| `styles/archetype.py` | Styles | Archetype-specific CSS |
| `styles/causal.py` | Styles | Causal tree CSS |
| `styles/conversational.py` | Styles | Conversation view CSS |
| `styles/dark_theme.py` | Styles | Base dark theme + animation CSS |
| `styles/devtools.py` | Styles | DevTools panel CSS |
| `styles/failure_first.py` | Styles | Failure-first visualization CSS |
| `styles/intelligence.py` | Styles | Intelligence/analysis CSS |
| `styles/layout.py` | Styles | Layout CSS |
| `styles/transparency.py` | Styles | Sprint 7 transparency CSS |
| `conversation_lens.py` | Lens (Era 1) | Entry-point lens: conversation spans |
| `decision_lens.py` | Lens (Era 1) | Entry-point lens: decision spans |
| `performance_lens.py` | Lens (Era 1) | Entry-point lens: performance metrics |
| `infrastructure_lens.py` | Lens (Era 1) | Entry-point lens: infrastructure/collector health |
| `layout_two_panel.py` | Page builder | Two-panel trace analysis layout (primary view) |
| `layout_card_grid.py` | Page builder | Card grid session overview |
| `layout_story.py` | Page builder | Story mode layout |
| `page_story.py` | Page (Era 3) | Story page with trace summaries |
| `page_sessions.py` | Page (Era 3) | Sessions overview page |
| `page_healthcheck.py` | Page (Era 3) | Health check page |
| `fixture_browser.py` | Page (Era 3) | Fixture browser page |
| `devtools_panel.py` | Panel | Developer tools panel (renderer coverage, MCP status) |
| `session_tree.py` | Panel | Session tree left panel (`SessionTreePanel`) |
| `side_effect.py` | Panel | Side effects right panel (`SideEffectPanel`) |
| `mutation_summary.py` | Panel | Mutation summary panel (`MutationSummaryPanel`) |
| `payload_diff.py` | Panel | Payload diff panel (`PayloadDiffPanel`) |
| `span_analysis.py` | Panel | Span analysis manager (`SpanAnalysisManager`, `SessionAnalysisPanel`) |
| `renderer_coverage.py` | Panel | Renderer coverage panel (`RendererCoveragePanel`) |
| `ai_companion_sidebar.py` | Panel | AI companion sidebar UI |
| `trace_explanation.py` | Panel | Trace explanation panel |
| `charts.py` | Widget | ECharts wrappers (Sankey, waterfall, etc.) |
| `computation_bar.py` | Widget | Computation progress bar (`render_computation_bar()`) |
| `computation_formula.py` | Widget | Formula panel (`render_formula_panel()`) |
| `computation_viz_preview.py` | Preview | Computation visualization spike/preview |
| `pipeline_story_preview.py` | Preview | Pipeline story preview |
| `archetype_preview.py` | Preview | Archetype rendering preview |
| `primitives_preview.py` | Preview | Primitives showcase |
| `llm_ux_preview.py` | Preview | LLM UX preview |
| `component_preview.py` | Preview | Component showcase |
| `story_data.py` | Data helper | `build_trace_summaries()` from `SpanBuffer` |
| `polling.py` | Utility | `PollingController` for timer-driven refresh |
| `keyboard_shortcuts.py` | Utility | Command palette and keyboard handler |

**File count summary**: ~27 top-level modules, 15 fixture modules, ~50 UI modules = ~92 Python source files total.

---

## Layer Boundaries

The codebase has a clear 5-layer architecture:

### Layer 1: Infrastructure / Network Edge

Modules that touch the network (inbound and outbound):

- `src/autom8_devconsole/otlp_receiver.py` — Inbound: accepts OTel spans via POST /v1/traces
- `src/autom8_devconsole/tempo_client.py` — Outbound: queries Grafana Tempo for historical traces
- `src/autom8_devconsole/mcp_routes.py` — Outbound capability surface: MCP HTTP SSE server
- `src/autom8_devconsole/sms_conversation_service.py` — Outbound: SMS HTTP I/O

These modules depend on Layer 2 (SpanBuffer, NarrativeEngine) but are never imported by Layer 2.

### Layer 2: Core Domain Models

The central data and semantic models that everything else depends on:

- `src/autom8_devconsole/span_buffer.py` — **Leaf/hub**: `ParsedSpan` is the universal span model imported by nearly every module. `SpanBuffer` is the authoritative in-memory span store.
- `src/autom8_devconsole/otlp_codec.py` — Span parsing; imported by `otlp_receiver.py`, `tempo_client.py`, `supplement.py`
- `src/autom8_devconsole/side_effect_utils.py` — **Leaf utility** imported by 8+ modules for dual-prefix normalization
- `src/autom8_devconsole/config.py` — Settings singleton; depended on by `app.py` only
- `src/autom8_devconsole/protocols.py` — **FROZEN** lens contracts (Era 1)
- `src/autom8_devconsole/protocols_v2.py` — Era-3 page contracts (active)

### Layer 3: Intelligence / Analysis Pipeline

The narrative and analytical machinery — pure Python, no I/O, no UI:

- `src/autom8_devconsole/narrative.py` — Core semantic mapping engine (`NarrativeEngine`); re-exports types from `autom8_devx_types`
- `src/autom8_devconsole/narrative_pipeline.py` — Plugin: reconciliation pipeline dialect
- `src/autom8_devconsole/narrative_request_response.py` — Plugin: HTTP request/response dialect
- `src/autom8_devconsole/narrative_computation.py` — Plugin: computation dialect
- `src/autom8_devconsole/trace_narrative.py` — Act-based composition (depends on `narrative.py`, `archetype.py`)
- `src/autom8_devconsole/deterministic_output.py` — Protocol boundary between deterministic and LLM layers
- `src/autom8_devconsole/intelligence_service.py` — Mediator: orchestrates `trace_narrative.py`, `risk_surface.py`, `deterministic_output.py` into `AnalysisResult`
- `src/autom8_devconsole/risk_surface.py` — Heuristic anomaly detection engine
- `src/autom8_devconsole/archetype.py` — Interaction classification (depends on `computation_stages.py`)
- `src/autom8_devconsole/computation_stages.py` — Computation span classification utilities
- `src/autom8_devconsole/causality.py` — Causal chain analysis
- `src/autom8_devconsole/response_interpreter.py` — HTTP response enrichment
- `src/autom8_devconsole/llm_analysis.py` — Optional LLM layer (depends on `anthropic` SDK optionally; uses `deterministic_output.py` as its input contract)

**Critical rule**: `narrative_pipeline.py` imports ONLY from `autom8_devx_types` at runtime (no `autom8_devconsole` imports at runtime). This is an explicit plugin isolation contract.

### Layer 4: Application / State Management

Wires infrastructure and intelligence together. Manages app lifecycle.

- `src/autom8_devconsole/app.py` — **Hub**: imports from all layers. `create_app()` builds `FastAPI` + wires `SpanBuffer`, `NarrativeEngine`, `LensRegistry`, `TraceIntelligenceService`. `lifespan()` manages `SpanStore`, `TempoClient`, `Autom8yHttpClient`.
- `src/autom8_devconsole/app_state.py` — Documentation TypedDict for `app.state`
- `src/autom8_devconsole/ai_companion.py` — Server-side conversation state (no NiceGUI imports)
- `src/autom8_devconsole/registry.py` — Entry-point lens discovery
- `src/autom8_devconsole/persistence.py` — SQLite durability layer (depends on `span_buffer.py`)
- `src/autom8_devconsole/supplement.py` — Demo fixture loading on startup
- `src/autom8_devconsole/fixture_recorder.py` — Fixture serialization/listing

### Layer 5: UI (NiceGUI)

All modules under `src/autom8_devconsole/ui/`. Never imported by Layers 2-4 (one-way dependency).

- **Design system sub-layer**: `theme.py`, `primitives.py`, `compositions.py`, `motion.py`, `styles/`
- **Lens plugins** (Era 1, entry-point registered): `conversation_lens.py`, `decision_lens.py`, `performance_lens.py`, `infrastructure_lens.py`
- **Page builders** (Era 3, PageRegistry): `layout_two_panel.py`, `layout_card_grid.py`, `page_story.py`, `page_sessions.py`, `page_healthcheck.py`, `fixture_browser.py`
- **Panel components**: `session_tree.py`, `side_effect.py`, `mutation_summary.py`, `payload_diff.py`, `span_analysis.py`, `ai_companion_sidebar.py`, `renderer_coverage.py`, `trace_explanation.py`
- **Spike/preview pages**: `computation_viz_preview.py`, `pipeline_story_preview.py`, `archetype_preview.py`, `primitives_preview.py`, `llm_ux_preview.py`, `component_preview.py`

**Import direction is strictly downward**: UI imports from intelligence/application layers but the reverse never happens.

### External Type Contract

`autom8_devx_types` (local editable dependency at `../autom8y/sdks/python/autom8y-devx-types`) exports the canonical types shared between the devconsole and narrative plugins:
`NarrativeContext`, `NarrativeFragment`, `NarrativeRuleSet`, `Predicate`, `Renderer`, `Rule`, `InteractionArchetype`, `NarrativeVoice`

---

## Entry Points and API Surface

### CLI Entry Points (from pyproject.toml)

```
autom8y-dev-x         → autom8_devconsole.__main__:main
autom8y-devx-summarize → autom8_devconsole.cli:cli_main
```

**`main()`** (`src/autom8_devconsole/__main__.py`):
- Detects `sys.argv[1] in {"summarize"}` to select CLI vs TUI mode
- TUI mode: calls `create_app()`, `ui.run_with(app)`, `uvicorn.run()` — binds on `DEVCONSOLE_NICEGUI_PORT` (default 8080)
- CLI mode: delegates to `cli_main(sys.argv[1:])` with no NiceGUI import

**`cli_main()`** (`src/autom8_devconsole/cli.py`): Headless `summarize` command. Reads spans from SQLite (default), stdin JSON, or file. Optional `--llm` flag. No NiceGUI dependency.

### HTTP Routes (registered in `app.py`)

| Method | Path | Handler | Description |
|---|---|---|---|
| POST | `/v1/traces` | `otlp_router` | OTLP span ingest (protobuf or JSON) |
| GET | `/` | NiceGUI | Main TUI page (three-panel layout) |
| GET | `/session/{session_id}` | NiceGUI | Session bookmarking |
| GET | `/fixtures` | NiceGUI | Fixture browser page |
| GET | `/story` | NiceGUI | Story mode page |
| GET | `/sessions` | NiceGUI | Sessions overview |
| + preview routes | NiceGUI | Various spike/preview pages |

Additional MCP HTTP SSE routes are registered by `mcp_routes.py` when the `mcp` package is available.

### Plugin Extension Points (entry-point groups)

| Group | Registration | Status |
|---|---|---|
| `autom8y_devx.lenses` | 4 lenses: conversation, decision, performance, infrastructure | FROZEN per ADR |
| `autom8y_devx.narrative_rules` | 2 rule sets: core, computation | Active |

### `app.state` Public Surface

Documented in `src/autom8_devconsole/app_state.py` (`AppState` TypedDict):

| Attribute | Type | Set In |
|---|---|---|
| `settings` | `DevconsoleSettings` | `create_app()` |
| `span_buffer` | `SpanBuffer` | `create_app()` |
| `narrative_engine` | `NarrativeEngine` | `create_app()` |
| `story_mode` | `list[bool]` | `create_app()` |
| `intelligence_service` | `TraceIntelligenceService` | `create_app()` |
| `trace_analyzer` | `TraceAnalyzer` | `create_app()` |
| `registry` | `LensRegistry` | `create_app()` |
| `http_client` | `Autom8yHttpClient` | `lifespan()` |
| `tempo_client` | `TempoClient` | `lifespan()` |
| `span_store` | `SpanStore \| None` | `lifespan()` |
| `persistence_degraded` | `bool` | `lifespan()` |
| `available_fixtures` | `list[FixtureInfo]` | `lifespan()` |
| `fixtures_dir` | `Path` | `lifespan()` |
| `mcp_server` | `Any` | `mcp_routes.py` (conditional) |

---

## Key Abstractions

### `ParsedSpan` (`src/autom8_devconsole/span_buffer.py`)

The universal span model. A `@dataclass` with fields: `span_id`, `trace_id`, `parent_span_id`, `name`, `service_name`, `session_id`, `start_time_ns`, `end_time_ns`, `duration_ms`, `attributes: dict[str, Any]`, `events: list[dict]`, `resource_attrs: dict[str, Any]`. Imported by nearly every module in the codebase.

### `SpanBuffer` (`src/autom8_devconsole/span_buffer.py`)

In-memory ring buffer (max 10,000 spans by default). Indexed by `trace_id` and `session_id`. `asyncio.Lock`-protected. Implements `SpanProvider` protocol. Key methods: `add()`, `get_by_trace()`, `get_by_session()`, `get_all_spans()`, `poll_new()`.

### `NarrativeEngine` (`src/autom8_devconsole/narrative.py`)

Stateless rules-table engine. Maps `ParsedSpan` → `NarrativeFragment` using priority-ordered `NarrativeRuleSet` plugins discovered via `autom8y_devx.narrative_rules` entry-point group. Circuit breaker on renderer failures (`_renderer_failures` dict + threshold=3).

### `Rule` / `NarrativeRuleSet` (from `autom8_devx_types`)

Plugin contract: a `Predicate` (span → bool) paired with a `Renderer` (span → `NarrativeFragment`). Grouped into priority-ordered `NarrativeRuleSet` instances.

### `DeterministicOutput` Protocol (`src/autom8_devconsole/deterministic_output.py`)

The explicit protocol boundary between deterministic analysis (Layer 3) and non-deterministic consumers (LLM, MCP). `serialize_for_llm()` and `serialize_for_mcp()` produce identical bytes for identical input. `TraceDeterministicOutput` is the concrete implementation.

### `LensProtocol` / `LensCapability` (`src/autom8_devconsole/protocols.py`, FROZEN)

Era-1 plugin contract. Lenses are tab-based UI panels in the center column. `LensCapability` is a `Flag` enum: `REFRESH | TRACE_FILTER | SESSION_FILTER | SPAN_INSPECT`. `LensContext` provides `span_provider`, `app`, `settings` at construction.

### `PageContext` / `PageRegistration` (`src/autom8_devconsole/protocols_v2.py`)

Era-3 extension mechanism (supersedes lens entry-points for new surfaces). `PageRegistration` declares route, module path, builder function, and whether context injection is needed. `PageContext` injects `buffer`, `narrative_engine`, `app`, `registry`, `session_filter`.

### `AnalysisResult` (`src/autom8_devconsole/intelligence_service.py`)

Frozen dataclass aggregating all intelligence output: `trace_narrative: TraceNarrative | None`, `risk_anomalies`, `heuristic_anomalies`, `side_effects`. Produced by `TraceIntelligenceService.analyze()` and consumed by `layout_two_panel.py`.

### `InteractionArchetype` (from `autom8_devx_types`)

Enum: `UNKNOWN | CONVERSATION | PIPELINE | REQUEST_RESPONSE | COMPUTATION`. Detected by `detect_archetype()` in `archetype.py`. Drives archetype-aware rendering in UI.

### `ConversationManager` (`src/autom8_devconsole/ai_companion.py`)

Server-side LRU OrderedDict (20-session cap, 30-minute TTL). Keyed by observability `session_id`. Stores `ConversationSession` instances with message deques (maxlen=50) and navigation context (maxlen=10). No NiceGUI imports — safe across navigation events.

### `DevconsoleSettings` (`src/autom8_devconsole/config.py`)

`Autom8yBaseSettings` subclass. Env prefix `DEVCONSOLE_`. Key fields: `NICEGUI_PORT` (8080), `OTLP_RECEIVER_PORT` (4327), `SPAN_BUFFER_SIZE` (10,000), `SCHEDULING_BASE_URL`, `TEMPO_BASE_URL`, `LLM_MODEL`, persistence settings. Retrieved via `@lru_cache get_settings()`.

---

## Data Flow

### Inbound: Real-Time Spans (OTLP Ingest)

```
OTel Collector (fanout)
  → POST /v1/traces (otlp_receiver.py)
  → protobuf parse (otlp_codec.py)
  → ParsedSpan construction
  → SpanBuffer.add() [in-memory, asyncio.Lock]
  → background flush_loop → SpanStore (SQLite WAL)
```

### Inbound: Demo Fixtures (Startup)

```
lifespan() startup
  → MultiFixtureLoader.load_all() (supplement.py)
  → generate.py → fixture_*.py → OTLP JSON
  → otlp_codec.parse_json_span()
  → SpanBuffer.add()
```

### Inbound: Historical Replay (Tempo)

```
UI trigger (Tempo replay button)
  → TempoClient.get_trace() (HTTP → Tempo :3200)
  → parse_tempo_response() → list[ParsedSpan]
  → SpanBuffer.add()
```

### Analysis Pipeline (Deterministic)

```
SpanBuffer.get_by_trace(trace_id) → list[ParsedSpan]
  → NarrativeEngine.narrate(spans)
      → NarrativeRuleSet priority order
      → Predicate matching → Renderer → NarrativeFragment
  → TraceNarrativeBuilder.build()
      → Act detection (BFS deepest-wins)
      → NarrativeAct composition
      → TraceNarrative
  → RiskSurfaceEngine.analyze() → DetectedAnomaly[]
  → extract_side_effects() → SideEffectRecord[]
  → build_deterministic_output() → TraceDeterministicOutput
  → TraceIntelligenceService.analyze() → AnalysisResult
```

### Analysis Pipeline (LLM, Optional)

```
AnalysisResult (or TraceDeterministicOutput)
  → serialize_for_llm() → structured text
  → TraceAnalyzer.analyze_trace() (anthropic SDK, async)
  → TTL-cached result
  → TraceAnalysis (Pydantic model)
```

### UI Render Cycle (NiceGUI)

```
PollingController (ui.timer, 2s interval)
  → SpanBuffer read (SpanProvider protocol)
  → TraceIntelligenceService.analyze()
  → AnalysisResult
  → layout_two_panel.py / layout_card_grid.py
  → NiceGUI element mutations (DOM diff via Socket.IO)
```

### MCP Surface

```
AI coding agent (Claude Code / Cursor)
  → HTTP SSE MCP request
  → mcp_routes.py tool handler
  → SpanBuffer → NarrativeEngine → serialize_for_mcp()
  → JSON response (deterministic)
```

### Configuration Flow

```
Environment variables (DEVCONSOLE_* prefix)
  → DevconsoleSettings (pydantic-settings)
  → get_settings() [lru_cache singleton]
  → app.state.settings (set in create_app())
  → LensContext.settings (injected into lenses)
  → PageContext.app.state.settings (accessed by page builders)
```

### Persistence Flow

```
SpanBuffer (authoritative read source)
  ↕ (write via flush_loop every 2s)
SQLite WAL (~/.autom8y/devconsole/spans.db)
  → warm-start load on restart (SpanStore.load_recent())
  → retention cleanup every 300s (cleanup_loop)
```

---

## Knowledge Gaps

1. **`sms_conversation_service.py`**: Module purpose known (SMS HTTP I/O, distinct from `ai_companion.py`) but internal implementation not read.
2. **`response_interpreter.py`**: `interpret_response()` / `enrich_interpretation()` signatures observed but internal logic not read.
3. **`causality.py`**: Known to provide causal chain utilities consumed by `ui/side_effect.py` but internal classes not enumerated.
4. **`computation_stages.py`**: `is_computation_trace()` signature known but the full classification taxonomy not read.
5. **UI page builders (`layout_two_panel.py`, `layout_card_grid.py`)**: Core wiring patterns known via imports but exact render graph structure not fully read (files too large for single read).
6. **`testing.py` pytest plugin**: `pytest_configure` registered, `SpanRecorder` class exists, fixtures documented — but the full fixture list and helper semantics not exhaustively read.
7. **`prototypes/` and root-level `_*.py` files**: Spike/prototype scripts outside the main source tree not observed.
8. **`autom8_devx_types` SDK types**: `ParsedSpan`, `NarrativeVoice`, and other types defined in the external editable dependency were not directly read.

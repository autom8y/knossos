---
domain: conventions
generated_at: "2026-03-25T12:11:05Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "e43ba47"
confidence: 0.90
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/workflow-patterns.md"
land_hash: "0d62b122cdf240a2c74669f4a488ddc71d2b5ab106ae0d3233c3290c40390c8d"
---

# Codebase Conventions

**Primary Language:** Python 3.12 (confirmed by `pyproject.toml` `requires-python = ">=3.12"`)

**Project:** `autom8-devconsole` -- a NiceGUI/FastAPI developer console for OTel span visualization

**Source root:** `src/autom8_devconsole/`

---

## Error Handling Style

### Error Creation

The project defines a minimal custom error hierarchy, concentrated in `tempo_client.py`:

```
Exception
  └── TempoClientError          (base for Tempo HTTP failures)
        └── TempoTraceNotFoundError  (404 specifically)
```

Located at `src/autom8_devconsole/tempo_client.py:53-58`.

For internal logic failures, the project uses stdlib exceptions directly (`ValueError`, `RuntimeError`, `TypeError`).

### Error Wrapping

`from exc` chaining is used consistently when converting lower-level exceptions into domain errors:
- `src/autom8_devconsole/tempo_client.py:99` — `raise TempoClientError(...) from exc`

### Exception Handling Patterns

Four patterns observed:

**1. Specific type-tuple catches** (preferred for parsing/type coercion):
```python
except (json.JSONDecodeError, TypeError):    # narrative.py, side_effect.py, persistence.py
except (ValueError, TypeError):              # narrative_request_response.py, performance_lens.py
```

**2. Broad `except Exception as exc` with logging** (used at service boundaries):
```python
except Exception as exc:
    logger.error("lens_build_failed", name=name, error=str(exc))
```

**3. Silent `except Exception` (bare)** (used in UI callbacks where crashes must not propagate):
```python
except Exception:
    pass   # or continue
```

**4. Named domain errors** (only for TempoClient consumers):
```python
except TempoClientError as exc:   # conversation_lens.py:534
```

### Error Boundaries at System Boundaries

- **CLI** (`src/autom8_devconsole/cli.py`): Specific catches with formatted error messages to stderr.
- **NiceGUI UI**: Broad `except Exception` swallowing in all event handlers and timer callbacks.
- **FastAPI/HTTP** (`src/autom8_devconsole/otlp_receiver.py`): Mix of specific catches and broad `Exception` catches. Errors returned as HTTP 400/500.
- **Lens error boundary** (`src/autom8_devconsole/app.py:341-368`): `_safe_build()` / `_safe_refresh()` with per-lens circuit breaker (3 failures).

### Logging Pattern (autom8y SDK, NOT stdlib)

All modules use:
```python
from autom8y_log import get_logger
logger = get_logger(__name__)
```
One exception: `src/autom8_devconsole/intelligence_service.py` uses `logging.getLogger(__name__)` — a known deviation.

**Structured keyword logging:**
```python
logger.info("tempo_trace_fetched", trace_id=trace_id, spans=len(spans), fetch_ms=round(fetch_ms, 1))
logger.warning("lens_refresh_error", name=name, error=str(exc), consecutive_failures=n)
logger.error("lens_build_failed", name=name, error=str(exc))
```

Event names follow `domain_verb_noun` pattern. Log messages are lowercase snake_case event tokens, not f-strings.

**Banned imports**: `loguru`, `structlog`, and raw `httpx` are banned via `ruff.toml` lint rules (`TID251`). Use `autom8y_log.get_logger()` and `autom8y_http.Autom8yHttpClient` instead.

### Result Types Instead of Exceptions

Cross-boundary calls use `@dataclass(frozen=True)` result types: `ConversationResult` with `success: bool`, `error_type: str | None`, `error_message: str | None`.

---

## File Organization

### Package Layout

```
src/autom8_devconsole/
+-- __init__.py          # version string only
+-- __main__.py          # entry point: TUI vs CLI dispatch
+-- app.py               # FastAPI app factory
+-- app_state.py         # TypedDict documenting app.state shape
+-- config.py            # pydantic-settings configuration
+-- protocols.py         # FROZEN plugin contracts
+-- registry.py          # Entry-point discovery -- FROZEN
+-- span_buffer.py       # Core data type: ParsedSpan, SpanBuffer
+-- narrative.py         # NarrativeEngine + rule building blocks
+-- narrative_computation.py
+-- narrative_pipeline.py
+-- narrative_request_response.py
+-- trace_narrative.py
+-- intelligence_service.py
+-- deterministic_output.py
+-- archetype.py
+-- causality.py
+-- risk_surface.py
+-- persistence.py
+-- tempo_client.py
+-- otlp_receiver.py
+-- otlp_codec.py
+-- mcp_routes.py
+-- llm_analysis.py
+-- ai_companion.py
+-- sms_conversation_service.py
+-- response_interpreter.py
+-- supplement.py
+-- fixture_recorder.py
+-- cli.py
+-- side_effect_utils.py
+-- computation_stages.py
+-- testing.py
+-- fixtures/            # Test fixture factories
+-- ui/
    +-- theme.py         # CSS token foundation
    +-- primitives.py    # UI primitive components
    +-- compositions.py  # Composite components
    +-- protocols_v2.py  # Era 3 page builder contracts
    +-- *_lens.py        # Lens plugins
    +-- layout_*.py      # Page layout builders
    +-- page_*.py        # Page builders
    +-- *_preview.py     # Spike/prototype preview pages
    +-- styles/          # CSS modules (one per UI subsystem)
```

### File Naming Conventions

- **`*_lens.py`**: Timer-driven tab plugins
- **`layout_*.py`**: Page layout builders accepting `PageContext`
- **`page_*.py`**: Full NiceGUI page route handlers
- **`*_preview.py`**: Standalone spike/prototype pages
- **`protocols.py`**: FROZEN public contracts (use `protocols_v2.py` for new)
- **`theme.py`**: Active CSS token foundation (renamed from `theme_v0.py`)
- **`_helpers.py`**: Private helpers (underscore prefix)
- **`styles/`**: CSS string constants only
- **`narrative_*.py`**: Archetype-specific narrative rule sets

### Intra-file Organization

Standard pattern:
1. Module docstring (with sprint ref, ADR refs)
2. `from __future__ import annotations`
3. Standard library imports
4. Third-party imports
5. Internal imports (with `TYPE_CHECKING` guard for circular-import-sensitive ones)
6. Module-level constants and type aliases prefixed with `_`
7. Public dataclasses / simple types
8. Private helper functions (prefixed `_`)
9. Public classes (the main exported API)

Sections delimited with ASCII separator banners:
```python
# ---------------------------------------------------------------------------
# Section name
# ---------------------------------------------------------------------------
```

---

## Domain-Specific Idioms

### `@dataclass(frozen=True)` as Result Type

Strongly preferred for value objects and result types. Pydantic `BaseModel` only when JSON schema/validation needed (LLM output models only).

### Entry-Point Plugin Pattern

`importlib.metadata.entry_points()` for plugin discovery. Validate, log and skip invalids — never raise. New lenses FROZEN; new surfaces use `PageRegistration`.

### `from __future__ import annotations`

Present in all source files. Universal convention.

### `TYPE_CHECKING` Guard Pattern

```python
if TYPE_CHECKING:
    from autom8_devconsole.span_buffer import ParsedSpan
```
Used in 47+ files for circular-import-sensitive imports.

### Tuple-over-list for Immutable Collections

Frozen dataclasses use `tuple[T, ...]` not `list[T]`. `frozenset` for membership-test sets.

### CSS as Module-Level String Constants

All CSS lives in `src/autom8_devconsole/ui/styles/`. Injected via `ui.add_css()`. `styles/__init__.py` provides `get_all_css()`.

### `lru_cache` Singleton for Settings

```python
@lru_cache
def get_settings() -> DevconsoleSettings:
    return DevconsoleSettings()
```

### `@contextmanager` for UI Container Primitives

NiceGUI containers wrapped as `@contextmanager` generators in `src/autom8_devconsole/ui/primitives.py`.

### Page Builder Function Convention

`build_{name}_page(ctx: PageContext) -> None`. Preview/spike builders take no arguments.

### Graceful Degradation Tiers

Services implement tiered degradation: Live call -> Cached result -> Deterministic fallback.

### Side-Effect Attribute Dual-Prefix Normalization

All code reading side-effect span attributes must use `get_side_effect_attr(attrs, field)` from `side_effect_utils.py` — never direct `attrs.get("side_effect.X")` or `attrs.get("com.autom8y.side_effect.X")`.

### ADR/Sprint References in Docstrings

Modules document their architectural provenance:
```python
"""...
Sprint: v-won-S2 (Semantic Convention Mapping Engine)
ADR: ADR-S2-001 (Rules Table)
"""
```

### `HEX` Dict and CSS Token System

UI colors come from two sources:
- `HEX` dict (in `theme.py`) — Python-accessible hex values derived from `DARK_TOKENS`
- CSS custom properties (`hsl(var(--token))`) — used in `.style()` calls and CSS strings

### `noqa` Suppression for Specific Patterns

- `# noqa: BLE001` — intentional broad exception catch in display paths
- `# noqa: F401` — re-export imports (backward-compat re-exports in `narrative.py`)
- `# noqa: SLF001` — intentional private attribute access during app setup
- `# type: ignore[union-attr]` — duck-typed lens/panel calls

---

## Naming Patterns

### Types and Classes

| Suffix | Meaning | Examples |
|---|---|---|
| `*Engine` | Stateless processing core | `NarrativeEngine`, `RiskSurfaceEngine` |
| `*Service` | Stateful I/O orchestrators | `TraceIntelligenceService`, `ConversationService` |
| `*Lens` | Timer-driven tab plugins | `ConversationLens`, `DecisionLens` |
| `*Panel` | Detail panels (right sidebar) | `SideEffectPanel`, `MutationSummaryPanel` |
| `*Store` | Persistent storage | `SpanStore` |
| `*Buffer` | In-memory ring buffer | `SpanBuffer` |
| `*Registry` | Discovery + lifecycle manager | `LensRegistry` |
| `*Client` | External HTTP client | `TempoClient` |
| `*Analyzer` | LLM-backed analysis | `TraceAnalyzer` |
| `*Error` | Custom exception | `TempoClientError` |
| `*Protocol` | Structural typing contract | `LensProtocol`, `SpanProvider` |
| `*Settings` | pydantic-settings config | `DevconsoleSettings` |
| `*Context` | Dependency injection bag | `LensContext`, `PageContext` |
| `*Fragment` | Single output unit | `NarrativeFragment` |
| `*Result` | Result dataclass | `ConversationResult`, `AnalysisResult` |
| `*Loader` | Loads/hydrates data | `MultiFixtureLoader` |
| `Parsed*` | Normalized from raw format | `ParsedSpan` |
| `*Builder` | Constructs a result object | `TraceNarrativeBuilder` |
| `*Manager` | Session-scoped lifecycle | `ConversationManager` |
| `*Resolver` | Pluggable strategy | `DefaultSessionResolver` |
| `*Stage` | Pipeline step | `ComputationStage` |
| `*Tree` | Hierarchical structure | `CausalTree` |
| `*Analysis` | Pydantic model for LLM output | `TraceAnalysis`, `SpanAnalysis` |

### Function Names

- **`create_*()`**: Factory functions. `create_app()`, `create_otlp_router()`.
- **`build_*()`**: NiceGUI page/component builders (side-effectful).
- **`get_*()`**: Accessor/getter, may be `@lru_cache`.
- **`_render_*()`**: Private renderer functions.
- **`detect_*()`**: Detection logic.
- **`is_*()`**: Boolean checks.

### Module-Level Constants

- **Public**: `UPPER_SNAKE_CASE`
- **Private**: `_UPPER_SNAKE_CASE`
- **SQL**: all private (`_CREATE_TABLE`, `_INSERT_SPAN`)

### Anti-Patterns to Avoid

- Do **not** use `logging.getLogger()` directly — use `autom8y_log.get_logger()`.
- Do **not** import `httpx` directly — use `autom8y_http.Autom8yHttpClient`.
- Do **not** use `loguru` or `structlog`.
- Do **not** use `list[T]` for immutable collections in frozen dataclasses.
- Do **not** add new lenses to the entry-point system (frozen).
- Do **not** use f-strings as log message strings — use structured keyword kwargs.

---

## Knowledge Gaps

1. **`autom8y_log`** — custom logging package interface inferred from usage but not read.
2. **`autom8_devx_types`** — types imported from external package, not read.
3. **`autom8y_http`** / **`autom8y_config`** — shared internal packages, base classes not read.
4. **`intelligence_service.py` logger deviation** — uses `logging.getLogger()` instead of `autom8y_log.get_logger()`. Whether intentional is unclear.
5. **`ui/styles/` CSS modules** — content confirmed as string-returning modules but exact function naming convention not exhaustively verified.

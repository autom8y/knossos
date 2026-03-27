---
domain: test-coverage
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
  - ".sos/land/workflow-patterns.md"
land_hash: "0d62b122cdf240a2c74669f4a488ddc71d2b5ab106ae0d3233c3290c40390c8d"
---

# Codebase Test Coverage

**Project:** autom8-devconsole (`autom8y-dev-x`)
**Language:** Python 3.12
**Test runner:** `pytest` (via `uv run pytest`)

---

## Coverage Gaps

### Critical: 13 Test Files Fail to Collect (14.1% of test files)

All 13 failures share two root causes in `src/autom8_devconsole/archetype.py` and `src/autom8_devconsole/ui/theme.py`:

**Root Cause 1 — `InteractionArchetype.COMPUTATION` does not exist in SDK**
`src/autom8_devconsole/archetype.py` line 106 references `InteractionArchetype.COMPUTATION`, but the installed `autom8y-devx-types` SDK defines only: CONVERSATION, PIPELINE, REQUEST_RESPONSE, UNKNOWN. Any test file that imports from `autom8_devconsole.archetype` (or anything that transitively imports it) fails at collection time.

Affected test files (9):
- `tests/test_archetype.py`
- `tests/test_computation_fragment_renderer.py`
- `tests/test_cross_service_act_suppression.py`
- `tests/test_deterministic_output.py`
- `tests/test_edge_case_fixtures.py`
- `tests/test_fixture_factories.py`
- `tests/test_narrative_coverage_gate.py`
- `tests/test_request_response_fixtures.py`
- `tests/test_trace_narrative.py`

**Root Cause 2 — Missing exports from `ui/theme.py`**
Tests import `ANIMATION_CSS`, `ARCHETYPE_CSS`, `CAUSAL_TREE_CSS`, and `get_all_css` from `autom8_devconsole.ui.theme`, but these are not exported. Affects 4 test files:
- `tests/test_theme.py`
- `tests/test_sprint5_card_intelligence_css.py`
- `tests/test_sprint11_causal_inspection.py`
- `tests/test_sprint12_integration_polish.py`

### Untested Source Modules

**Core package — no direct test file:**
- `src/autom8_devconsole/app.py` — **1,509 lines, largest file, no test**. NiceGUI app wiring. Biggest coverage blind spot.
- `src/autom8_devconsole/otlp_codec.py` — 88 lines. Covered indirectly via `test_otlp_receiver.py`.
- `src/autom8_devconsole/registry.py` — 120 lines. Covered indirectly via `test_page_registry.py`.

**UI modules — no direct or indirect test coverage:**
- `src/autom8_devconsole/ui/component_preview.py`
- `src/autom8_devconsole/ui/computation_formula.py`
- `src/autom8_devconsole/ui/page_healthcheck.py`
- `src/autom8_devconsole/ui/pipeline_story_preview.py`
- `src/autom8_devconsole/ui/span_analysis.py`

**Fixtures modules — no test coverage:**
- `src/autom8_devconsole/fixtures/cps_variants.py`
- `src/autom8_devconsole/fixtures/generate.py`
- `src/autom8_devconsole/fixtures/phantom_computation.py` (new)

### Collectible but Shallow Tests

- `tests/test_renderer_coverage.py` — 5 tests
- `tests/test_trace_correlation.py` — 9 tests
- `tests/test_tree_performance.py` — 4 tests
- `tests/test_session_bookmark.py` — 4 tests

### Prioritized Gap List

1. **`app.py`** — 1,509 statements, no test coverage. Application wiring effectively untested.
2. **13 collection failures** — 14.1% of test files blocked by 2 import errors.
3. **`ui/layout_two_panel.py`** — Primary production surface, low coverage.
4. **`ui/compositions.py`** — Shared UI compositions, low coverage.
5. **`ui/ai_companion_sidebar.py`** — AI sidebar rendering, low coverage.

---

## Testing Conventions

### Test Organization

All 92 test files live in `tests/`. No subdirectory structure. All use `test_*.py` naming.

Every test file uses class-based organization (`class TestXxx`). Function-level `def test_` without a class is rare.

### The `_make_span` Factory Pattern

`tests/conftest.py` defines `_make_span(**overrides)` as a module-level function (not a `@pytest.fixture`). Imported directly via `from conftest import _make_span` in 43 of 92 test files. This is the dominant span construction pattern.

- `_SPAN_FIELDS` frozenset distinguishes direct constructor kwargs from attribute extras
- Convenience params: `service_name`, `operation_name`, `attributes`, `events`
- Any unrecognized kwarg is treated as a span attribute key-value pair

### Assertion Patterns

Pure `assert` statements dominate (5,077 occurrences). `pytest.raises` in 174 instances across 72 files. No `unittest.TestCase` assertion methods.

### Mock/Patch Usage

27 test files use `unittest.mock` (MagicMock, AsyncMock, patch), 382 total mock-related occurrences. `respx` for HTTP mocking in 3 test files. LLM/Anthropic calls always mocked.

### Async Test Pattern

21 of 92 test files use async patterns. `asyncio_mode = "auto"` in pyproject.toml handles detection.

### Fixture Data Sources

1. `_make_span()` from `conftest.py` — unit-level span construction
2. Fixture factories from `src/autom8_devconsole/fixtures/` — integration-level OTLP payloads
3. Inline literal dicts — specific attribute values

### Sprint-Labeled Test Convention

18 test files use sprint-prefixed names (e.g., `test_sprint4_archetype_rendering.py`). Acceptance tests written at sprint completion to lock in behavioral contracts.

### Parametrize Usage

Only 10 of 92 test files use `@pytest.mark.parametrize`. Most tests use explicit separate methods.

### Deferred Import Pattern

Test methods import components under test inside the method body:
```python
def test_something(self):
    from autom8_devconsole.ui.conversation_lens import ConversationLens
```
Used in at least 28 test files. Prevents NiceGUI state contamination at collection time.

---

## Test Structure Summary

### Test Runner Configuration

From `pyproject.toml`:
```
asyncio_mode = "auto"
asyncio_default_fixture_loop_scope = "function"
python_files = "test_*.py"
testpaths = ["tests"]
addopts = ["--import-mode=importlib", "--tb=short", "-v"]
```

- `--import-mode=importlib`: prevents `sys.path` pollution
- `pytest-cov` installed but no `--cov` in `addopts` — coverage not run by default

### Test Count Summary

- **Total test functions defined**: 3,565
- **Collectible by pytest**: 2,948 (617 in 13 erroring files)
- **Test files**: 92 (including conftest.py)
- **Collection errors**: 13 files (14.1%)

### Layer Distribution

| Layer | Test Files | Notes |
|-------|-----------|-------|
| Core domain logic | ~35 files | narrative, causality, archetype, span_buffer, computation_stages |
| UI components | ~25 files | primitives, session_tree, conversation_lens, devtools_panel |
| Sprint acceptance | 18 files | `test_sprint{N}_*.py`, `test_act2_s1_*.py` |
| Fixtures/test data | ~8 files | test_fixture_factories, test_edge_case_fixtures |
| Infrastructure | ~6 files | test_otlp_receiver, test_tempo_client, test_cli, test_config |

### Coverage Gate

`tests/test_narrative_coverage_gate.py` is a bespoke hard gate (PT-04) that:
- Loads all computation fixtures via `create_multi_fixture_loader()`
- Verifies all 30 computation narrative rules are exercised
- Validates exactly 16 fixture factories are registered
- Checks all 4 archetypes are detectable
- Currently **fails to collect** due to `InteractionArchetype.COMPUTATION` import error

### Test Runner Command

```
pytest
```

Coverage: `pytest --cov=src/autom8_devconsole`

---

## Knowledge Gaps

1. **No coverage.json baseline run captured** — `coverage.json` exists but may be stale. Line-level percentages not available from this observation.
2. **Test execution state unknown** — 2,948 collectible tests not run in this session. Pass/fail rates undocumented.
3. **`app.py` testability strategy unknown** — Whether NiceGUI UI is tested via browser automation only or simply has a coverage gap is undocumented.
4. **Sprint 10 test distribution** — No `test_sprint10_*.py` file exists. Coverage may be under other test names.
5. **Styles module test strategy** — 8 of 9 `ui/styles/*.py` modules have only indirect coverage.

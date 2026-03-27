---
domain: design-constraints
generated_at: "2026-03-25T12:11:05Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
  - "./.ledge/decisions/"
generator: theoros
source_hash: "e43ba47"
confidence: 0.87
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/initiative-history.md"
land_hash: "f31e74621ecfd923b5479ae0dad1b28ed0420173e53031dbe1b9d6698ef7a107"
---

# Codebase Design Constraints

## Tension Catalog Completeness

12 tensions cataloged. 4 resolved. 8 active.

---

### TENSION-001: Three-Era UI Extension Mechanism Coexistence

**Status**: Active
**Type:** Dual-system / Layering violation
**Location:** `src/autom8_devconsole/registry.py:3`, `src/autom8_devconsole/app.py:376-1118`, `src/autom8_devconsole/ui/protocols_v2.py:1`

Three extension mechanisms coexist:
- **Era 1**: `LensRegistry` + `LensProtocol` entry-point discovery. `registry.py:3`: `STATUS: FROZEN per ADR-ui-extension-mechanism`.
- **Era 2**: Hardcoded panel wiring in `_build_nicegui_page()` at `app.py:376-1118`.
- **Era 3**: `PageRegistry` pattern via `PageRegistration` in `protocols_v2.py`. Canonical per ADR-001.

**Historical reason:** Entry-point lens system designed for external plugin ecosystem that never materialized.
**Ideal resolution:** Migrate Era-1 lenses to `PageBuilderProtocol`; retire `_build_nicegui_page()` (REM-014); delete `registry.py` and `protocols.py`.
**Resolution cost:** High -- `_build_nicegui_page()` is LOAD-bearing (LOAD-005).

---

### TENSION-002: `protocols.py` Frozen While `protocols_v2.py` Is Active

**Status**: Active
**Type:** Naming mismatch / Abstraction boundary
**Location:** `src/autom8_devconsole/protocols.py:12`, `src/autom8_devconsole/ui/protocols_v2.py:1`

Two protocol files. `protocols.py` "one-way door warning" at line 12. `protocols_v2.py:3` cross-references: `LOAD-001: protocols.py is frozen. New protocols live here.`

**Ideal resolution:** Merge once Era-1 lenses are migrated.
**Resolution cost:** Low to document; high to merge.

---

### TENSION-003: `theme.py` / `theme_v0.py` Naming — RESOLVED

**Status**: RESOLVED (commit `4bea514`)
`theme_v0.py` renamed to `theme.py`. All production imports updated.

---

### TENSION-004: `_build_nicegui_page()` vs. `layout_two_panel.py` as Detail Panel Authority

**Status**: Active
**Type:** Dual-system / Layering violation
**Location:** `src/autom8_devconsole/app.py:376-1118`, `src/autom8_devconsole/ui/layout_two_panel.py`

ADR-001 designates `layout_two_panel.py` as authoritative. Retirement is REM-014, Phase C.
**Resolution cost:** High. 742+ lines with bidirectional callback mesh.

---

### TENSION-005: Dual Side-Effect Attribute Prefix

**Status**: Active
**Type:** Naming mismatch / Backward-compatibility
**Location:** `src/autom8_devconsole/side_effect_utils.py:9`

`U-02 unresolved`: dual-prefix lookup (`side_effect.*` and `com.autom8y.side_effect.*`). SRE has not confirmed cutover.
**Resolution cost:** Medium. Requires SRE coordination.

---

### TENSION-006: trace_id vs span_id Conflation in On-Click Handlers

**Status**: Active
**Type:** Naming mismatch / Load-bearing jank
**Location:** `src/autom8_devconsole/app.py:898-904`

Handler detects 32-char (trace_id) vs 16-char (span_id) format. Comment: `# TENSION-006 fix`.
**Resolution cost:** Medium.

---

### TENSION-007: `conversation_service.py` Naming — RESOLVED

**Status**: RESOLVED. Renamed to `sms_conversation_service.py`.

---

### TENSION-008: Dual-Prefix Lookup Incompleteness in `llm_analysis.py`

**Status**: Active (partially resolved)
**Type:** Partial resolution
**Location:** `src/autom8_devconsole/llm_analysis.py:807-821`

`_extract_side_effects` migrated to `get_side_effect_attr()`. Comment: `TENSION-008 partial resolution`.
**Resolution cost:** Low -- isolated to one file.

---

### TENSION-009: `archetype.py` Detection Docstring — RESOLVED

**Status**: RESOLVED
Docstring at `archetype.py:35` now correctly reads `>= 1`, matching implementation per ADR-computation-narrative-contract D-7.

---

### TENSION-010: Dual `ParsedSpan` Definitions

**Status**: Active
**Type:** Dual-system / Type mismatch
**Location:** `src/autom8_devconsole/span_buffer.py:83` vs `autom8_devx_types._span.ParsedSpan`

SDK type (9 fields) vs runtime type (15 fields). Narrative plugins annotate with 9-field type but receive 15-field instances.
**Resolution cost:** Medium -- requires SDK semver bump.

---

### TENSION-011: `_parse_json_*` Private Functions — RESOLVED

**Status**: RESOLVED. Extracted to `src/autom8_devconsole/otlp_codec.py` with public API. All 4 callers migrated.

---

### TENSION-012: `phantom.synthetic` as Undocumented First-Class Attribute (New)

**Status**: Active
**Type:** Implicit convention / Abstraction gap
**Location:** `src/autom8_devconsole/fixtures/phantom_computation.py:84`, `src/autom8_devconsole/ui/layout_two_panel.py:1274,1485`, `src/autom8_devconsole/ui/compositions.py:1007`

`phantom.synthetic=true` attribute is set on fixture spans and checked in two `layout_two_panel.py` locations and `compositions.py` `is_phantom` parameter. No ADR, no convention schema, no central helper.
**Resolution cost:** Low to document; medium to formalize namespace.

---

## Trade-off Documentation

| Tension | ADR / Evidence |
|---------|--------------|
| TENSION-001 | `.ledge/decisions/ADR-ui-extension-mechanism.md` (comprehensive) |
| TENSION-002 | `protocols.py:12` inline + `protocols_v2.py:3` cross-ref |
| TENSION-004 | ADR-001 (direction); REM-014 Phase C |
| TENSION-005 | `side_effect_utils.py:9` -- U-02 unresolved |
| TENSION-006 | `app.py:898` inline comment |
| TENSION-008 | `llm_analysis.py:811` partial resolution comment |
| TENSION-010 | `narrative.py:20-27` re-export comment; no formal ADR |
| TENSION-012 | No documentation exists |

3 of 8 active tensions have ADR coverage (TENSION-001, TENSION-004 direction, TENSION-010 indirectly). Operational tensions documented only by inline comments.

---

## Abstraction Gap Mapping

### GAP-001: `AppState` TypedDict Not Runtime-Enforced

`src/autom8_devconsole/app_state.py` defines `AppState(TypedDict, total=False)` documenting `app.state.*`. Not enforced at runtime.

### GAP-002: No Abstraction Between SpanBuffer and UI Timer

500ms NiceGUI timer wired directly in `_build_nicegui_page()`. `page_story.py` has independent timer.

### GAP-003: No Cross-Panel Event Bus

Bidirectional panel communication via manually registered callbacks in `_build_nicegui_page()`. No formal event bus.

### GAP-004: HTTP Attribute Key Lists — Partially Resolved

`HTTP_METHOD_KEYS`, `HTTP_STATUS_KEYS`, `HTTP_PATH_KEYS` now centralized in `side_effect_utils.py:71-73` with `get_http_attr()`. Original inline duplication resolved.

### GAP-005: `narrative.py` Re-export Coupling (LOAD-006)

Re-exports from `autom8_devx_types` with `# noqa: F401`. Callers implicitly depend on SDK version.

### GAP-006: `phantom.synthetic` Attribute Has No Type or Validator (New)

No enum, constant, or validator. Check scattered across `layout_two_panel.py:1274` and `compositions.py:1007`. Missing: `is_phantom_span(span)` helper.

### Premature Abstraction: `LensRegistry`

Entry-point scanning for exactly 4 known lenses. External plugin ecosystem never materialized.

### Premature Abstraction: `PageRegistration.route_params`

Only `/session/{session_id}` uses it. Dispatch immediately special-cases `["session_id"]`.

---

## Load-Bearing Code Identification

### LOAD-001: `protocols.py` -- Published Contract Boundary

`src/autom8_devconsole/protocols.py`. Frozen per ADR. ~8 UI file consumers + all 4 lenses + registry.

### LOAD-002: `layout_two_panel.py` -- Primary Production Surface

`src/autom8_devconsole/ui/layout_two_panel.py` (2,000+ lines). Contains FM-1 through FM-7 DOM mutation rules.

### LOAD-003: `theme.py` + `ui/styles/` -- CSS Cascade Anchor

`src/autom8_devconsole/ui/theme.py`. 30+ import sites. WCAG constraints on L/C values.

### LOAD-004: `side_effect_utils.py` -- Side-Effect Data Extraction Contract

`src/autom8_devconsole/side_effect_utils.py`. 8+ consumers. Also exports `HTTP_METHOD_KEYS`, `HTTP_PATH_KEYS`, `HTTP_STATUS_KEYS`, `get_http_attr`. Changing prefix priority silently flips values.

### LOAD-005: `_build_nicegui_page()` Closure Mesh

`src/autom8_devconsole/app.py:376-1118` (742+ lines). Central wiring hub. `LOAD-005` safety comments in `ai_companion.py:196`, `layout_two_panel.py:814`. Retirement is REM-014 Phase C.

### LOAD-006: `autom8_devx_types` Re-export in `narrative.py`

`src/autom8_devconsole/narrative.py:20-27`. Removing re-export breaks external narrative plugins.

### LOAD-007: `DeterministicOutput` Protocol Boundary

`src/autom8_devconsole/deterministic_output.py`. Frozen per ADR-P3-005.

### LOAD-008: `otlp_codec.py` -- Public OTLP Parsing API

`src/autom8_devconsole/otlp_codec.py`. Extracted from `otlp_receiver.py` (TENSION-011 resolution). 4 callers.

### LOAD-009: `narrative_computation.py` Fallback Rule as Category Gate (New)

`src/autom8_devconsole/narrative_computation.py:851-969`. Fallback MUST be last rule. Moving it earlier shadows explicit rules. `category="computation"` return is gate for `render_computation_fragment_row()`.

### LOAD-010: `ComputationStage` D-6 Fields as Downstream Contract (New)

`src/autom8_devconsole/computation_stages.py:108-113`. 5 optional fields with `None` defaults. All three zoom levels consume `ComputationStage`. No positional construction allowed.

---

## Evolution Constraint Documentation

| Area | Rating | Evidence |
|------|--------|----------|
| `protocols.py` LensProtocol / LensMeta / LensContext | **Frozen** | "One-way door warning"; ADR-ui-extension-mechanism |
| `registry.py` LensRegistry | **Frozen** | `STATUS: FROZEN per ADR` |
| `_build_nicegui_page()` in `app.py` | **Frozen** | ADR-001; REM-014 Phase C |
| `DeterministicOutput` protocol | **Frozen** | ADR-P3-005 |
| 5-archetype taxonomy | **Frozen** | ADR-archetype-taxonomy |
| Two-panel story arc (60/40) | **Frozen** | ADR-composition-layer |
| OKLCH palette parameters | **Frozen** | ADR-composition-layer + ADR-oklch-palette-system |
| `compositions.py` Rule C-2 | **Frozen** | ADR-composition-layer |
| `NarrativeFragment` fields | **Frozen** | ADR-computation-narrative-contract |
| `_build_computation_rules()` fallback ordering | **Frozen** | ADR-computation-narrative-contract D-1 |
| Act-trigger predicate | **Frozen** | ADR-P3-007 |
| `ParsedSpan` fields in `span_buffer.py` | **Coordinated** | ~12 file updates; must not diverge from SDK |
| `PageRegistration` / `PageContext` | **Coordinated** | All Era 3 page builders |
| `ui/theme.py` token definitions | **Coordinated** | 30+ consumers; WCAG constraints |
| `layout_two_panel.py` FM rules | **Coordinated** | FM-1 through FM-7 |
| `otlp_codec.py` public API | **Coordinated** | 4 callers |
| `archetype.py` detection thresholds | **Coordinated** | Must sync with `computation_stages.py` |
| `ComputationStage` field additions | **Coordinated** | D-6: `None` defaults required |
| `narrative_computation.py` rule ordering | **Coordinated** | D-2: new rules before fallback |
| `compositions.py` layer API | **Safe** | ADR-composition-layer; additive |
| `config.py` DevconsoleSettings | **Safe** | New fields safe with defaults |
| `side_effect_utils.py` prefix constants | **Migration** | SRE confirmation required |
| SQLite schema in `persistence.py` | **Migration** | No schema version tracking |
| `DEVCONSOLE_*` env var names | **Migration** | Renaming breaks deployments |
| `narrative.py` re-exports | **Migration** | Removing breaks external plugins |
| MCP tool signatures | **Migration** | External clients depend on tool names |
| `phantom.synthetic` attribute namespace | **Unclassified** | No ADR; TENSION-012 |

---

## Risk Zone Mapping

### RISK-001: Unauthenticated Write Endpoint Bound to All Interfaces

**Location:** `src/autom8_devconsole/__main__.py:81` -- `host="0.0.0.0"`; `src/autom8_devconsole/otlp_receiver.py`
`POST /v1/traces` with no authentication. Acknowledged as "PROTOTYPE shortcut."

### RISK-002: SpanBuffer Index Corruption Under Concurrent Mutation/Read

**Location:** `src/autom8_devconsole/span_buffer.py:157-193`
`add()` uses `asyncio.Lock` but `get_by_trace()`, `get_by_session()` at lines 195-204+ do not acquire the lock.

### RISK-003: Silent `ParsedSpan` Reconstruction from Corrupt Persistence Data

**Location:** `src/autom8_devconsole/persistence.py:385-424`
Corrupt JSON silently yields span with `{}` attributes and `None` session_id.

### RISK-004: Unvalidated `session_id` in URL Route

**Location:** `src/autom8_devconsole/app.py:1109-1117`
No length or character-set validation. SQL injection prevented by parameterized queries.

### RISK-005: `otlp_codec.py` as Undocumented Load-Bearing API

**Location:** `src/autom8_devconsole/otlp_codec.py`
4 callers. No ADR governs the API surface. Elevated to LOAD-008.

### RISK-006: Dual Side-Effect Prefix Inconsistency

**Location:** `src/autom8_devconsole/side_effect_utils.py:8-9`
Cross-reference TENSION-005 and TENSION-008.

### RISK-007: `ParsedSpan` Type Mismatch Between SDK and Runtime

**Location:** `src/autom8_devconsole/narrative_computation.py`, `src/autom8_devconsole/trace_narrative.py`
Plugins annotate with 9-field SDK type but receive 15-field runtime type. Cross-reference TENSION-010.

### RISK-008: SQLite Schema With No Version Tracking

**Location:** `src/autom8_devconsole/persistence.py:39-68`
No `PRAGMA user_version` or migration table.

### RISK-009: Act Detection Scar — Service-Boundary Reintroduction

**Location:** `src/autom8_devconsole/trace_narrative.py`
ADR-P3-007: reintroducing service-boundary criterion caused 81 acts for 210-span trace.

### RISK-010: `phantom.synthetic` Attribute as Unguarded Rendering Branch (New)

**Location:** `src/autom8_devconsole/ui/layout_two_panel.py:1274,1485`, `src/autom8_devconsole/ui/compositions.py:1007`
Two unsynchronized checks, no central helper, no test coverage for phantom rendering paths.

---

## Knowledge Gaps

1. **`layout_two_panel.py` FM rules FM-1 through FM-7**: Rule designations referenced but actual content not read from the 2,000+ line file.
2. **`llm_analysis.py` TENSION-008 scope**: `_extract_side_effects` migrated but remaining `attrs.get()` calls not fully traced.
3. **`autom8_devx_types` ParsedSpan field count**: 9-field SDK type not independently verified.
4. **NiceGUI session isolation**: Multi-tab behavior for shared `app.state` not characterized.
5. **`phantom.synthetic` convention origin**: Whether `phantom.*` is a defined namespace in `autom8y-data` or dev-x-only is undocumented.
6. **`narrative_computation.py` D-5 dual-lookup implementation**: 4 attribute name mismatches from ADR not verified as implemented.

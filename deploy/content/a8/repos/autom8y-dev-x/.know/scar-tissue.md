---
domain: scar-tissue
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
  - ".sos/land/scar-tissue.md"
land_hash: "3c900a54ece5de2152b4bd6563794ee93d2507586dc1d75cc82490c62713ed2e"
---

# Codebase Scar Tissue

## Failure Catalog Completeness

The git history contains 15 fix/bug commits across 51 total commits (29.4% are fixes). The codebase carries 16+ distinct named scars. The following catalog documents each failure with location evidence.

---

### SCAR-H: Error Boundaries — Page and Lens Crash Containment

**What failed**: Lenses and page builders would crash the entire UI on exception rather than rendering a degraded placeholder.

**Current implementation**:
- `src/autom8_devconsole/app.py:341-360` — `_safe_build()` / `_safe_refresh()` with circuit breaker (3 failure threshold)
- `src/autom8_devconsole/app.py:1405-1458` — `_render_degraded_page()` + `_invoke_page_builder()` for page-level boundaries
- `src/autom8_devconsole/ui/page_story.py:86-115` — independent `_safe_build()` / `_safe_refresh()` on story page

**Comment markers**: `SCAR-H` at `app.py:1408, 1427, 1435, 1460`; `trace_explanation.py:108`

**Regression tests**: `tests/test_page_registry.py:437-512` (`TestErrorBoundary`); `tests/test_protocols.py:401-450` (circuit breaker logic)

---

### SCAR-F: Circular Feedback — Bidirectional Widget Re-entrancy

**What failed**: Bidirectional widget communication where A triggers B which triggers A, causing infinite loops via NiceGUI's reactive event system.

**Fix pattern**: `_suppress_*` boolean flags wrapped in `try/finally` blocks.

**Current locations** (6 classes):
- `src/autom8_devconsole/ui/side_effect.py:144, 193-197, 484-489` — `_suppress_causal_feedback`
- `src/autom8_devconsole/ui/trace_explanation.py:123, 596-607` — `_suppress_reenter`
- `src/autom8_devconsole/ui/span_analysis.py:91, 312-365, 375-445` — `_suppress_reenter` (dict and bool variants)
- `src/autom8_devconsole/ui/ai_companion_sidebar.py:408-409, 1271-1326` — `_suppress_toggle` + `_suppress_reenter`
- `src/autom8_devconsole/ui/conversation_lens.py:350, 775` — `_suppress_link_feedback`

**Comment markers**: `SCAR-F` at 30+ locations across 6 source files.

**Regression tests**: `tests/test_trace_correlation_adversarial.py:292` (A5); `tests/test_sprint11_causal_inspection.py:520-572` (`TestScarFGuard`); `tests/test_sprint8_companion_interaction.py:468-473` (streaming guard)

---

### F-01: Hydration Gate — Black Screen During NiceGUI Startup

**What failed**: Browser automation screenshots fired before NiceGUI/Vue/Quasar completed hydration.

**Fix**: MutationObserver → setTimeout → loading overlay with `data-hydrated=true` gate.

**Current implementation**: `src/autom8_devconsole/app.py:448-492` — `devconsole-loading-gate` with `@media (prefers-reduced-motion: reduce)` safe.

**Regression tests**: `tests/test_act2_s1_remediation.py:498-575` (`TestF01LoadingGate`)

---

### F-10: Border Token WCAG 1.4.11 Contrast Failure

**What failed**: `--border` tokens at 15.9% lightness, 1.31:1 contrast on `--card` background (fails WCAG 3:1 minimum).

**Fix commit**: `31df115` — Raised to 40% lightness (~3.3:1).

**Current implementation**: `src/autom8_devconsole/ui/theme.py:110, 127`

**Regression tests**: `tests/test_act2_s1_remediation.py:61-490` (45 tests)

---

### F-S15-01: NiceGUI 3.8.0 Removed `set_text()` from Base Element

**What failed**: `ui.element("kbd").set_text()` crashed with `AttributeError`.

**Fix commit**: `6d358ba` — Replaced with `ui.html(f"<kbd class='key-chip'>{_key}</kbd>", sanitize=False)`.

**Current implementation**: `src/autom8_devconsole/ui/layout_two_panel.py:465-468`

---

### F-S15-02: Warm-Start Buffer Fills Blocks Fixture Loading

**What failed**: `SpanStore.load_recent()` spans got `source="otlp"` (default), filling buffer and blocking `MultiFixtureLoader`.

**Fix commit**: `6d358ba` — Added `SpanStore.purge_all()`, called before warm start when `LOAD_DEMO_FIXTURE=True`.

**Current implementation**: `src/autom8_devconsole/persistence.py:316-336`; `src/autom8_devconsole/app.py:135-149`

**Regression tests**: `tests/test_supplement.py:494-575` (TC-S11)

---

### PageRegistry 422: Handler Signature Mismatch

**What failed**: `**kwargs` in handler signatures caused FastAPI HTTP 422 on parameterized routes.

**Fix commit**: `aec7384` — Explicit typed parameter signatures + `__post_init__` validation on `PageRegistration`.

**Current implementation**: `src/autom8_devconsole/app.py:1460-1490`

**Regression tests**: `tests/test_page_registry.py:515+`

---

### Sessions Grid Whitespace

**What failed**: Empty-state div used `opacity: 0` not `display: none`, creating viewport gap.

**Fix**: Changed to `display: none` on empty state.

**Current implementation**: `src/autom8_devconsole/ui/layout_card_grid.py:391-393`

---

### ADR-P3-007: Service Boundary Act Root Pathology

**What failed**: Service-boundary criterion as act-trigger produced 81 acts for 210-span trace.

**Fix**: ADR-P3-007 removed service-boundary criterion.

**Current implementation**: `src/autom8_devconsole/trace_narrative.py:4-16, 73, 363` — `CRITICAL` comment guards the predicate.

**Regression tests**: `tests/test_trace_narrative.py:258-259`; `tests/test_cross_service_act_suppression.py:552` (550-line dedicated test file)

---

### HOTFIX-01: Session Duplication in MultiFixtureLoader

**What failed**: Standalone fixtures duplicated sessions from `create_multi_session_pool` composite.

**Fix commit**: `293caca` — Removed 3 duplicate factories (19 → 16).

**Current implementation**: `src/autom8_devconsole/supplement.py:178`

---

### CF-01/CF-02: Destructive Color Token Contrast Failure

**What failed**: `--destructive` at 62% lightness failed WCAG 4.5:1 for white text on solid fills and tinted backgrounds.

**Fix**: Dual-token split — `--destructive` at 51% for solid fills; `--destructive-text` at 65% for tinted backgrounds.

**Current implementation**: `src/autom8_devconsole/ui/theme.py:99-101, 136, 160`

---

### LOAD-005: NiceGUI Closure Injection Fragility

**What failed**: Page builder closures captured live objects at definition time, holding stale references.

**Fix**: Access `app.state.*` at call time (deferred injection), not via closure capture.

**Current locations**: `src/autom8_devconsole/ai_companion.py:26-27, 195-196`; `src/autom8_devconsole/ui/ai_companion_sidebar.py:64-67`; `src/autom8_devconsole/ui/layout_two_panel.py:814`; `src/autom8_devconsole/ui/layout_card_grid.py:777`

---

### TENSION-006: trace_id vs span_id Conflation in On-Click Handlers

**What failed**: Conversation lens emits 32-char trace_id where span tree expects 16-char span_id.

**Fix**: Length-based disambiguation at `src/autom8_devconsole/app.py:898-910`

---

### TENSION-009: Computation Archetype Detection Threshold

**What failed**: Threshold `>= 3` missed valid computation traces with 1 span.

**Fix**: Lowered to `>= 1`. Current: `src/autom8_devconsole/archetype.py:73`

---

### F-WS-E-01: Causal Tree Wiring — Flat Loop Replaced by SideEffectPanel

**What failed**: Flat inline loop over side effects bypassed `SideEffectPanel.set_buffer()`, preventing causal tree activation.

**Fix commit**: `293caca`

**Current implementation**: `src/autom8_devconsole/ui/layout_two_panel.py:1588-1619`

---

### MF-01/02/03: Fixture Card A11y

**What failed**: Duplicate accessible names, archetype badges without accessible text, UNKNOWN as raw "?".

**Fix**: `aria-hidden` on inner label; `aria-label="Archetype: {name}"` on badge; "Unknown" not "?".

**Current implementation**: `src/autom8_devconsole/ui/compositions.py:391-432`; `src/autom8_devconsole/ui/fixture_browser.py:31, 462`

---

### Missing `respx` Dev Dependency

**What failed**: `autom8y-http` SDK pytest plugin requires `respx`. Without it, pytest fails to start.

**Fix commit**: `e4cc270` — Added `respx>=0.20.0` to dev dependencies.

---

## Category Coverage

| Category | Scars | Count |
|---|---|---|
| **UI re-entrancy / circular feedback** | SCAR-F | 1 pattern, ~30 guards |
| **Accessibility / WCAG violations** | F-10, CF-01/CF-02, MF-01/02/03, WCAG 2.3.3, MI-01 | 6 |
| **Hydration / browser automation timing** | F-01 | 1 |
| **Framework API breakage (NiceGUI)** | F-S15-01, LOAD-005 | 2 |
| **Data model / persistence schema gaps** | F-S15-02, TENSION-006 | 2 |
| **Route / HTTP infrastructure** | PageRegistry 422, SCHEDULING_ env, respx dep | 3 |
| **Algorithm correctness** | ADR-P3-007, TENSION-009 | 2 |
| **Error boundary / crash containment** | SCAR-H | 1 |
| **Fixture / test data integrity** | HOTFIX-01 | 1 |
| **CSS layout specificity** | Sessions grid whitespace | 1 |

10 distinct categories. Categories searched but not found: Data corruption, Race condition, Config drift, Security (non-XSS).

---

## Fix-Location Mapping

| Scar | Primary Fix Location | Secondary Locations |
|---|---|---|
| SCAR-H | `src/autom8_devconsole/app.py:341-360,1405-1458` | `ui/page_story.py:86-115` |
| SCAR-F | `src/autom8_devconsole/ui/side_effect.py:484-489` | `ui/ai_companion_sidebar.py`, `ui/span_analysis.py`, `ui/trace_explanation.py`, `ui/conversation_lens.py` |
| F-01 | `src/autom8_devconsole/app.py:448-492` | -- |
| F-10 | `src/autom8_devconsole/ui/theme.py:110,127` | `ui/layout_card_grid.py:375-393` |
| PageRegistry 422 | `src/autom8_devconsole/app.py:1460-1490` | `ui/protocols_v2.py` |
| Sessions Whitespace | `src/autom8_devconsole/ui/layout_card_grid.py:391-393` | -- |
| F-S15-01 | `src/autom8_devconsole/ui/layout_two_panel.py:465-468` | -- |
| F-S15-02 | `src/autom8_devconsole/persistence.py:316-336` | `app.py:135-149` |
| HOTFIX-01 | `src/autom8_devconsole/supplement.py:178` | -- |
| F-WS-E-01 | `src/autom8_devconsole/ui/layout_two_panel.py:1588-1619` | -- |
| ADR-P3-007 | `src/autom8_devconsole/trace_narrative.py:4-16,73,363` | -- |
| CF-01/CF-02 | `src/autom8_devconsole/ui/theme.py:99-101,136` | -- |
| LOAD-005 | `src/autom8_devconsole/ai_companion.py:26` | `ui/ai_companion_sidebar.py:64`, `ui/layout_two_panel.py:814`, `ui/layout_card_grid.py:777` |
| MI-01 | `src/autom8_devconsole/ui/primitives.py:280-281` | -- |

All fix locations verified to exist in the current codebase.

---

## Defensive Pattern Documentation

**SCAR-F → `_suppress_*` try/finally pattern**: Any bidirectional widget communication or container.clear()+rebuild must use try/finally flag pattern.

**SCAR-H → `_safe_build()` / `_safe_refresh()` / `_invoke_page_builder()` error boundary pattern**: Page and lens registration loops use closure factories with try/except + degraded placeholder. Circuit breaker at 3 consecutive failures.

**F-S15-01 → `ui.html()` for custom elements**: Custom HTML elements must use `ui.html(content, sanitize=False)` rather than `ui.element(tag).set_text()`.

**F-S15-02 → `purge_all()` before demo fixture load**: The `LOAD_DEMO_FIXTURE` boot path MUST call `purge_all()` before `MultiFixtureLoader.load_all()`.

**F-01 → `data-hydrated` screenshot gate**: Browser automation must gate on `[data-hydrated]` attribute.

**PageRegistry 422 → Explicit handler signatures**: Route handlers for parameterized routes must declare explicit typed parameters. `PageRegistration.__post_init__` validates.

**Sessions Whitespace → `display: none` not `opacity: 0`**: Hidden grid children must use `display: none`.

**CF-01/CF-02 → Dual destructive token split**: `--destructive` (51%) for solid fills; `--destructive-text` (65%) for text on tinted backgrounds.

**ADR-P3-007 → No service boundary act triggers**: `CRITICAL` comment guards the predicate. Do not add service.name comparisons to act detection.

**LOAD-005 → `app.state.*` deferred access**: No live object capture in page builder closures.

**MI-01 → CSS class for animations**: Animations must be applied via CSS classes for `prefers-reduced-motion` suppression.

---

## Agent-Relevance Tagging

| Scar | Relevant Agents | Why |
|---|---|---|
| SCAR-F | component-engineer, interaction-prototyper | New interactive panels with `container.clear()` MUST implement `_suppress_reenter` |
| SCAR-H | component-engineer, rendering-architect | New pages via PageRegistry MUST use `_invoke_page_builder()`. All lens calls go through `_safe_build()`. |
| F-01 | component-engineer, frontend-fanatic | E2E tests must await `[data-hydrated]`. Loading gate is `aria-hidden` and `prefers-reduced-motion` safe. |
| F-10, CF-01 | a11y-engineer, design-system-steward, stylist | CSS tokens must validate WCAG contrast. `--destructive` vs `--destructive-text` for tinted backgrounds. |
| MI-01 | a11y-engineer, stylist | Every new CSS `transition`/`animation` must include `@media (prefers-reduced-motion: reduce)` block. |
| PageRegistry 422 | component-engineer, rendering-architect | No `**kwargs` in NiceGUI page handlers. `route_params` must match `{param}` patterns. |
| F-S15-01 | component-engineer | `set_text()` NOT available on NiceGUI base Element. Audit `.set_text()` on NiceGUI upgrades. |
| F-S15-02 | component-engineer | `SpanStore` has no `source` column. `LOAD_DEMO_FIXTURE=True` always purges. |
| ADR-P3-007 | component-engineer | New span types that could be act triggers must be reviewed against predicate list. Service names excluded. |
| LOAD-005 | component-engineer | Do not inject `span_buffer`, `app`, or other app-state objects into page builder closures. |
| TENSION-006 | component-engineer | Conversation lens emits `trace_id` (32 chars) not `span_id` (16 chars). Handle both lengths. |
| HOTFIX-01 | component-engineer | `MultiFixtureLoader` factory count pinned to 16. Adding factories requires gate update. |

---

## Knowledge Gaps

1. **LOAD-005 origin commit not isolated**: Pattern failure predates visible commit window.
2. **SCAR-F origin commit not isolated**: Pattern established early, extended incrementally.
3. **No dedicated regression tests for LOAD-005**: Enforced only by comments.
4. **SpanStore source column gap unresolved**: `purge_all()` workaround treats symptom, not cause.
5. **TENSION-006 has no dedicated regression test**: Covered only indirectly.
6. **Hydration gate has zero unit test coverage**: The `data-hydrated` / `MutationObserver` pattern is tested only via Act2-S1 remediation tests for HTML structure.

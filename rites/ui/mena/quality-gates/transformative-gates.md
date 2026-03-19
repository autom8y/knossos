# Transformative Posture Quality Gates (QG-E1 through QG-E9)

Governing principle: "Incremental correctness -- iterating toward something more truthful through continuous small improvements" (Lovin [MODERATE]).

Workflow: propose -> analyze -> migrate -> validate

## Gate Definitions

| Gate ID | Dimension | Type | Phase | Criterion | Agent |
|---------|-----------|------|-------|-----------|-------|
| QG-E1 | D0: Accessibility | **Hard** | validate | WCAG 2.2 AA zero-tolerance + a11y contract preserved through migration (no regression from baseline) | a11y-engineer |
| QG-E2 | Contract: API | Embedded | migrate | TypeScript compilation passes; no type errors introduced by migration | producing agent (self-check) |
| QG-E3 | Contract: Behavior | Embedded | migrate | Integration tests pass; no behavioral regressions from migration | producing agent (self-check) |
| QG-E4 | Contract: Visual | **Soft (blocking)** | validate | Visual regression review approved; no unintended visual changes | frontend-fanatic |
| QG-E5 | Contract: A11y | **Hard** | validate | A11y audit baseline preserved; no a11y regressions from migration | a11y-engineer |
| QG-E6 | Contract: Automation | Embedded | migrate | Selectors and analytics hooks intact; automation pipelines unbroken | producing agent (self-check) |
| QG-E7 | D3: Consistency | Embedded | migrate | Migrated components consistent with target system state | producing agent (self-check) |
| QG-E8 | D1: Micro-interactions | Advisory | validate | Motion and interaction patterns preserved or intentionally evolved per change proposal | frontend-fanatic (if invoked) |
| QG-E9 | D5: Edge state craft | Embedded | migrate | Edge states preserved through migration; no states lost or degraded | producing agent (self-check) |

## Key Design Decisions

**QG-E4 (Visual contract -- soft gate, blocking in transformative posture)**: Design system evolution must not introduce unintended visual changes. This maps to Saarinen's principle: "the system must remain coherent at every intermediate state" [MODERATE]. Visual regression is the observable signal of coherence failure. Every unintended visual change is a contract violation.

**Five contract types**: QG-E2, QG-E3, QG-E4, QG-E5, QG-E6 map to the five contracts (API, behavior, visual, a11y, automation). Three are embedded self-checks (automatable: TypeScript, tests, selectors). Two require terminal evaluation (holistic: visual regression, a11y preservation).

**QG-E1 AND QG-E5**: Both are hard gates involving accessibility. QG-E1 covers standard WCAG 2.2 AA. QG-E5 specifically covers a11y CONTRACT preservation -- the migration must not cause an a11y regression from the pre-migration baseline. A migration that passes QG-E1 can still fail QG-E5 if it introduced a new a11y regression not present before migration.

**QG-E8 (Advisory)**: Motion and interaction patterns are advisory in transformative posture because the primary concern is contract preservation, not interaction design improvement. Motion that changes during migration is assessed for intent (was this change intentional?), not quality.

## Soft Gate Mechanics for QG-E4

**Frontend-fanatic invocation in transformative**: Automatically invoked for all scopes at validate phase.

**QG-E4 pass criteria**:
- Rendered output matches expected state per change proposal's visual contract specification
- All unintended visual differences are zero (visual regression tool confirms no diff outside expected changes)
- Intended visual changes match what was described in the change-proposal

**QG-E4 fail -> back-route**: validate -> migrate (component-engineer). Finding must identify specific unintended visual changes with location and description.

## Embedded Contract Self-Checks

**QG-E2 (API contract)**:
- Run: `tsc --noEmit` (or equivalent TypeScript compilation)
- Pass: zero type errors
- Fail: any type error in migrated files

**QG-E3 (Behavior contract)**:
- Run: integration test suite
- Pass: all tests pass, no regressions
- Fail: any test failure in migrated components

**QG-E6 (Automation contract)**:
- Check: test selectors still target correct elements (role, label, text -- not class or ID)
- Check: analytics `data-*` attributes preserved on tracked elements
- Check: any automation scripts or scrapers still function
- Pass: zero selector breakages, zero analytics gaps

## Coverage Summary

| Dimension | Transformative Evaluation |
|-----------|--------------------------|
| D0: Accessibility | Hard gate (a11y-engineer) |
| D1: Micro-interactions | Advisory (frontend-fanatic, if invoked) |
| D2: Cognitive efficiency | Not evaluated (scope: contract preservation) |
| D3: Consistency | Embedded self-check |
| D4: Personality/Brand | Not evaluated (scope: contract preservation) |
| D5: Edge state craft | Embedded self-check |
| D6: Emotional design | Not evaluated (scope: contract preservation) |
| API contract | Embedded self-check |
| Behavior contract | Embedded self-check |
| Visual contract | Soft gate -- blocking (frontend-fanatic) |
| A11y contract | Hard gate (a11y-engineer) |
| Automation contract | Embedded self-check |

Coverage: 5/7 observable dimensions + 5/5 contract types. D2, D4 are intentionally out of scope -- transformative posture's purpose is system coherence, not interaction design or brand expression.

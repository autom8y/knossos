# UI Testing Pyramid

> Testing strategy, layer decision frameworks, structured output formats for agent workflows.

## Core Principle

"The more your tests resemble the way your software is used, the more confidence they can give you." — Kent C. Dodds

Test behavior, not implementation. Query by what users perceive (role, label, visible text), not by CSS class, test ID, or component internal state.

**Exceptions where this principle breaks down**:
- Algorithmic/computational logic (sorting, parsing, validation rules) — test as pure functions
- Error boundaries and recovery paths — users should never see these working, but they must be tested
- Performance characteristics — user-centric queries do not capture timing
- Integration contract adherence — a component can render correctly while violating its API contract

## Test Layer Decision Framework

The testing shape is context-dependent. No single shape is universally correct.

| Component Risk Profile | Recommended Test Weight |
|------------------------|------------------------|
| Pure logic (no UI) | Unit tests dominant |
| Standard UI component | Integration tests dominant |
| Complex interactive widget | Integration + interaction tests + a11y tests |
| Design-sensitive component | Integration + visual regression |
| Critical user journey | E2e tests required regardless of other coverage |
| Cross-package component | Contract tests + integration |

**Universal rule**: every component needs at least one integration-level test that exercises it the way a user would.

## Four Test Layers

### Layer 1: Static Analysis (Non-Negotiable Base)
Catches defects at near-zero runtime cost. 100% coverage expected.
- Type checking: all component props typed; `any` in a component interface is a test gap
- A11y lint rules: `eslint-plugin-jsx-a11y` catches missing alt text, missing labels, invalid ARIA at authoring time
- Dead code detection: unused props indicate undertested code

### Layer 2: Integration Tests (Highest-ROI Layer)
Render component with immediate dependencies, exercise via simulated user interaction.
- Default to integration tests as the primary test type
- Render in realistic context (with providers, children, mock data)
- Simulate interaction sequences (click, type, navigate)
- Mock only external boundaries (network, browser APIs) — never mock child components unless separately versioned
- SSR applications: weight e2e more heavily than traditional trophy, as client/server boundary complexity increases mocking burden

### Layer 3: Accessibility Testing (Four Sub-Layers)

| Sub-layer | What It Catches | Automation | Standard |
|-----------|----------------|------------|---------|
| Static lint | Missing alt, missing labels, invalid ARIA | Fully automated, authoring-time | Zero tolerance |
| Automated audit (axe-core) | Contrast failures, missing landmarks, duplicate IDs | Fully automated, test-time | Fail CI on violations |
| Interaction testing | Keyboard navigation, focus management, ARIA state updates | Semi-automated (scripted) | Required for every interactive component |
| Manual review | Alt text quality, reading order, screen reader announcements | Cannot automate | Required at component creation and before major releases |

### Layer 4: Visual Regression (Design-Sensitive Components Only)
Use for: design system components, brand-critical pages, CSS refactoring validation.
Do not use for: business logic, accessibility compliance, performance measurement.

**Baseline management rules**:
- Baselines stored in version control
- Baseline updates require explicit human approval — never auto-approve
- Use ratio-based diff thresholds (percentage of changed pixels)
- Run in consistent containerized environments
- Test at minimum 3 viewport widths: 375px, 768px, 1280px

## Snapshot Tests: Narrow Utility

Snapshot testing is appropriate for:
- Serialized data output (JSON, error messages, generated code)
- Configuration object validation
- API response contracts

Snapshot testing is harmful for UI components because any markup change triggers failure, and developers habituate to `--update-snapshot` without reviewing diffs.

**Rule**: If a snapshot test exceeds 20 lines when pretty-printed, decompose into explicit behavioral assertions.

## Structured Test Output for Agent Workflows

| Test Type | Output Format | Key Agent Fields |
|-----------|--------------|-----------------|
| Unit/Integration | CTRF (JSON) or JUnit XML | test name, status, duration, failure message, file location |
| Accessibility | axe-core JSON | rule ID, impact (minor/moderate/serious/critical), WCAG criteria, CSS selector target |
| Visual regression | JSON diff report | component name, diff percentage, baseline path, actual path |
| Performance | Lighthouse JSON | metric name, value, threshold, pass/fail |
| Static analysis | SARIF (JSON) | rule ID, severity, file location, message, suggested fix |

**Critical requirement**: every violation must include a location (file + line or CSS selector) and severity. Without location, agents cannot auto-fix. Without severity, agents cannot prioritize.

**Format preference**: CTRF (JSON) for test results, SARIF for static analysis results, axe-core native JSON for accessibility. JUnit XML only when no alternative exists.

## Test Selection Decision Tree

```
Does the component contain complex logic (calculations, parsing, validation)?
  YES → Write unit tests for the pure logic (extract to testable functions)

Does the component render UI users interact with?
  YES → Write integration test (render, interact, assert on outcome)

Is the component visually sensitive (design system, brand-critical)?
  YES → Add visual regression test (baseline + diff)

Is the component interactive (buttons, forms, navigation)?
  YES → Add keyboard interaction tests + axe-core audit

Is the component part of a critical user journey?
  YES → Ensure e2e test coverage for the journey

Is the component consumed by other teams/packages?
  YES → Add contract tests for the public API surface
```

All paths: static analysis (type checking + linting) as baseline.
All interactive components: axe-core audit as minimum a11y coverage.

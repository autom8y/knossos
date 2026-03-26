---
description: "Design System Governance companion for ui-ref skill."
---

# Design System Governance

> Design-code pipeline stages, three governance gate types, versioning strategy, RFC contribution lifecycle.

## The Five-Stage Token Pipeline

Every token lifecycle flows through five stages with governance gates at each transition:

| Stage | Description | Gate |
|-------|-------------|------|
| **Author** | Tokens defined in design tool or DTCG JSON | Naming convention review |
| **Sync** | Tokens committed to Git (canonical source) | PR review; NOT the design tool |
| **Transform** | Style Dictionary converts DTCG JSON to platform outputs | Deterministic build check |
| **Validate** | Linting, visual regression, contract testing in CI | All gates must pass before merge |
| **Distribute** | Versioned packages published with changelogs | Semver + changeset requirement |

**Critical**: The Git repository is the source of truth. The design tool is an authoring interface that syncs to Git — not a source that code reads from. When drift is detected, Git (reviewed, version-controlled, CI-validated) takes precedence.

## Three Governance Gate Types

### Gate 1: Token Validation
- DTCG conformance linting
- Tier-reference integrity (no tier-skipping)
- Naming convention compliance
- Drift detection against design tool source

### Gate 2: Component Contract Validation
- Visual regression against screenshot baselines (requires human approval for any detected diff)
- Prop/slot API conformance against component schema
- Accessibility audit (axe-core, zero violations)

### Gate 3: Lifecycle and Version Discipline
- Changesets: every PR must include a changeset file classifying major/minor/patch
- Breaking change criteria applied (see below)
- Deprecation warnings in development mode before removal

**If any gate is missing, flag it as a governance gap.**

## Breaking Change Classification

Apply these criteria to every token or component style change:

| Criterion | Classification |
|-----------|---------------|
| Text color change on surfaces the adopter controls | MAJOR |
| Font metrics change (size, weight, letter-spacing) causing text reflow in constrained layouts | MAJOR |
| Box-model change (padding, margin, width, height) affecting layout beyond component boundary | MAJOR |
| Removes or renames a public prop, slot, or CSS custom property | MAJOR |
| Adds new prop with default preserving existing behavior | MINOR |
| Adds new component or variant | MINOR |
| Change contained within component visual boundary, no external layout effect | PATCH |
| Accessibility fix that does not change layout | PATCH |

**When uncertain**: run visual regression, measure layout impact on representative consumer pages, then classify.

## Versioning Strategy

| Condition | Strategy |
|-----------|---------|
| Fewer than 50 consuming teams | Monolithic (simpler, guarantees compatibility) |
| More than 50 consuming teams, diverse upgrade cadences | Per-component or Hybrid |
| Breaking changes typically affect most components simultaneously | Monolithic (synchronized upgrades explicit) |
| Teams need to pin specific components while upgrading others | Per-component |

## RFC Contribution Lifecycle

Changes to a design system follow six stages:

1. **Discussion** — informal exploration of need
2. **Proposal** — formal RFC with rationale, API shape, impact analysis
3. **Triage** — system team evaluates fit
4. **Consensus** — community review period (3–45 days depending on scope)
5. **Approval** — governing body authorizes implementation
6. **Implementation** — design, build, document, test, release

**Small fixes** (accessibility improvements, bug fixes) may skip to stage 5.

**RFC threshold**: Any change that adds a public prop, removes a prop, changes default behavior, or adds a new component requires an RFC. Never implement a new public component without an approved RFC.

**RFC document must include**: rationale, proposed API (props, slots, variants, states), token usage, accessibility requirements, migration impact for existing consumers.

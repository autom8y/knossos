---
description: 'Theoria domain registry and audit criteria catalog. Use when: running /theoria audits, discovering available audit domains, understanding grading criteria. Triggers: theoria, audit, domain criteria, pinakes, grading, domain registry.'
name: pinakes
version: "1.0"
---
---
name: pinakes
description: "Theoria domain registry and audit criteria catalog. Use when: running /theoria audits, discovering available audit domains, understanding grading criteria. Triggers: theoria, audit, domain criteria, pinakes, grading, domain registry."
---

# pinakes

> The domain registry for theoria audits -- Callimachus's catalog of the labyrinth.

## Purpose

The Pinakes catalogs audit domains: what can be audited, where the criteria live, and how to grade what is found. Every theoria consults the Pinakes before dispatching theoroi.

Named for Callimachus's Pinakes -- the first systematic catalog of the Library of Alexandria. Callimachus did not write the books; he told you which books existed and what they contained. The Pinakes does the same for audit domains.

## Domain Registry

This table IS the registry. A domain exists when it has a row here and a criteria file in `domains/`.

| Domain | Criteria | Scope | Description |
|--------|----------|-------|-------------|
| **dromena** | `domains/dromena.md` | framework | Slash commands: naming, structure, completeness |
| **legomena** | `domains/legomena.md` | framework | Skills: description quality, trigger precision |
| **agents** | `domains/agents.md` | framework | Agent prompts: role clarity, handoff criteria |
| **hooks** | `domains/hooks.md` | framework | Hook wiring: lifecycle coverage, matchers, timeouts |
| **mena-structure** | `domains/mena-structure.md` | framework | Mena directories: naming, INDEX files, manifest registration |
| **architecture** | `domains/architecture.md` | codebase | Source code structure: packages, layers, entry points, abstractions, data flow |
| **conventions** | `domains/conventions.md` | codebase | Error handling style, file organization, domain-specific idioms, naming patterns |
| **test-coverage** | `domains/test-coverage.md` | codebase | Test gaps, coverage patterns, testing conventions, fixture patterns |
| **scar-tissue** | `domains/scar-tissue.md` | codebase | Past bugs, regressions, fix locations, defensive patterns born from failures |
| **design-constraints** | `domains/design-constraints.md` | codebase | Structural tensions, load-bearing jank, abstraction gaps, evolution constraints, risk zones |
| **radar-confidence-gaps** | `domains/radar-confidence-gaps.md` | radar | Flags .know/ domains where confidence < 0.80 |
| **radar-staleness** | `domains/radar-staleness.md` | radar | Flags .know/ domains past their expires_after window |
| **radar-unguarded-scars** | `domains/radar-unguarded-scars.md` | radar | Matches SCAR locations against untested packages — scars in untested code are unguarded regressions |
| **radar-constraint-violations** | `domains/radar-constraint-violations.md` | radar | Checks documented design constraints and frozen areas against codebase patterns |
| **radar-convention-drift** | `domains/radar-convention-drift.md` | radar | Samples files and checks adherence to documented conventions (error handling, naming, testing style) |
| **radar-architecture-decay** | `domains/radar-architecture-decay.md` | radar | Checks import graph against documented layer model for boundary violations and undocumented cross-cutting imports |
| **radar-recurring-scars** | `domains/radar-recurring-scars.md` | radar | Counts SCARs by category; flags categories with 3+ entries as systemic patterns |
| **adversarial-conventions** | `domains/adversarial-conventions.md` | adversarial | Finds code that contradicts documented conventions in .know/conventions.md |
| **adversarial-architecture** | `domains/adversarial-architecture.md` | adversarial | Finds structural evidence that contradicts documented architecture in .know/architecture.md |
| **adversarial-scar-tissue** | `domains/adversarial-scar-tissue.md` | adversarial | Finds undocumented scars or scars whose fixes have regressed |
| **dialectic-architecture** | `domains/dialectic-architecture.md` | dialectic | Surfaces unstated assumptions in .know/architecture.md |
| **dialectic-design-constraints** | `domains/dialectic-design-constraints.md` | dialectic | Surfaces constraints present in the code but missing from .know/design-constraints.md |
| **feature-census** | `domains/feature-census.md` | feature | Feature enumeration: scans project sources to produce feature taxonomy with GENERATE/SKIP recommendations |
| **feature-knowledge** | `domains/feature-knowledge.md` | feature | Per-feature knowledge capture: purpose, conceptual model, implementation map, boundaries |
| **release-platform-profile** | `domains/release-platform-profile.md` | release | Cached stable platform state: repo ecosystems, pipeline chains, dependency topology, build configs |
| **release-history** | `domains/release-history.md` | release | Release outcome log: versions released, CI pass/fail patterns, failure classifications, trend analysis |
| **complaints** | `domains/complaints.md` | framework | Complaint pipeline: filing volume, severity distribution, resolution rates, tag emergence |

### Scope Values

| Scope | Meaning |
|-------|---------|
| `framework` | Knossos infrastructure (agents, dromena, legomena) |
| `codebase` | Source code quality (Go, Python, shell scripts) |
| `process` | Development workflow (git, CI/CD, testing) |
| `culture` | Team practices (docs, naming, conventions) |
| `radar` | Cross-reference signals: reads .know/ files and detects gaps, drift, decay, and violations |
| `adversarial` | Adversarial mode: actively searches for evidence that contradicts .know/ claims |
| `dialectic` | Dialectic mode: surfaces unstated assumptions and undocumented constraints in .know/ files |
| `feature` | Product/feature knowledge: what the project does, why features work the way they do, conceptual models |
| `release` | Release engineering knowledge: platform profiles, dependency topologies, release history, CI patterns |

### Grading Scale

All domains use simple letter grades A through F:

| Grade | Meaning | Threshold |
|-------|---------|-----------|
| **A** | Excellent | 90-100% criteria met |
| **B** | Good | 80-89% criteria met |
| **C** | Adequate | 70-79% criteria met |
| **D** | Below Standard | 60-69% criteria met |
| **F** | Failing | Below 60% criteria met |

No +/- modifiers. Simplicity prevents grade inflation and gaming.

## When to Use

| Scenario | What to Read |
|----------|--------------|
| Running a theoria audit | This INDEX (domain registry) |
| Understanding criteria for a domain | `domains/{domain}.md` |
| Adding a new audit domain | `registry-format.md` |
| Understanding grading in detail | `schemas/grading.md` |
| Interpreting audit reports | `schemas/report-format.md` |

## Progressive Disclosure

### Domain Criteria — Framework
- [dromena.md](domains/dromena.md) - Slash command audit criteria
- [legomena.md](domains/legomena.md) - Skill audit criteria
- [agents.md](domains/agents.md) - Agent prompt audit criteria
- [hooks.md](domains/hooks.md) - Hook wiring audit criteria
- [mena-structure.md](domains/mena-structure.md) - Mena directory audit criteria
- [complaints.md](domains/complaints.md) - Complaint pipeline audit criteria

### Domain Criteria — Codebase
- [architecture.md](domains/architecture.md) - Architecture knowledge capture
- [conventions.md](domains/conventions.md) - Conventions knowledge capture
- [test-coverage.md](domains/test-coverage.md) - Test structure and coverage knowledge capture
- [scar-tissue.md](domains/scar-tissue.md) - Scar tissue knowledge capture (failure history, regressions)
- [design-constraints.md](domains/design-constraints.md) - Design constraint knowledge capture (tensions, load-bearing code, risk zones)

### Domain Criteria — Radar (cross-reference signals)
- [radar-confidence-gaps.md](domains/radar-confidence-gaps.md) - Flag .know/ domains with confidence < 0.80
- [radar-staleness.md](domains/radar-staleness.md) - Flag .know/ domains past their expiry window
- [radar-unguarded-scars.md](domains/radar-unguarded-scars.md) - Scars in untested packages (unguarded regressions)
- [radar-constraint-violations.md](domains/radar-constraint-violations.md) - Design constraints contradicted by codebase patterns
- [radar-convention-drift.md](domains/radar-convention-drift.md) - Convention adherence gaps in active packages
- [radar-architecture-decay.md](domains/radar-architecture-decay.md) - Import boundary violations against documented layer model
- [radar-recurring-scars.md](domains/radar-recurring-scars.md) - Scar categories with 3+ entries (systemic patterns)

### Domain Criteria — Adversarial
- [adversarial-conventions.md](domains/adversarial-conventions.md) - Find code contradicting documented conventions
- [adversarial-architecture.md](domains/adversarial-architecture.md) - Find structural evidence contradicting documented architecture
- [adversarial-scar-tissue.md](domains/adversarial-scar-tissue.md) - Find undocumented scars or regressed fixes

### Domain Criteria — Dialectic
- [dialectic-architecture.md](domains/dialectic-architecture.md) - Surface unstated assumptions in architecture documentation
- [dialectic-design-constraints.md](domains/dialectic-design-constraints.md) - Surface undocumented constraints present in code

### Domain Criteria — Feature
- [feature-census.md](domains/feature-census.md) - Feature enumeration census with GENERATE/SKIP recommendations
- [feature-knowledge.md](domains/feature-knowledge.md) - Per-feature knowledge capture (purpose, model, implementation, boundaries)

### Domain Criteria — Release
- [release-platform-profile.md](domains/release-platform-profile.md) - Cached platform state (ecosystems, pipelines, topology, configs)
- [release-history.md](domains/release-history.md) - Release outcome history (log, failures, CI patterns, trends)

### Schemas
- [registry-format.md](registry-format.md) - How to add new domains
- [grading.md](schemas/grading.md) - Grading scale definitions
- [report-format.md](schemas/report-format.md) - Audit report structure

## Consumers

- **theoros agent**: Domain evaluator dispatched by `/theoria`
- **/theoria dromena**: User-facing audit command
- Any agent performing domain-specific quality assessment

## How to Add a Domain

1. Write criteria file in `domains/{domain}.md`
2. Add row to the Domain Registry table above
3. Run `/theoria {domain}` to validate
4. See `registry-format.md` for full format specification

## Related Skills

- `smell-detection` skill — Code quality smell taxonomy
- `doc-ecosystem` skill — Template patterns

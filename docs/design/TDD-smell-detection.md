# TDD: Smell Detection Shared Skill

> Technical Design Document for cross-rite smell detection patterns and heuristics.

---

## Overview

This Technical Design Document specifies the `smell-detection` shared skill that provides unified code smell detection patterns, severity classification rules, and integration with debt tracking systems. The skill serves as a single source of truth for smell taxonomy across `debt-triage-pack`, `hygiene-pack`, and any team assessing code quality.

**Location**: `rites/shared/skills/smell-detection/`
**Consumers**: debt-collector (debt-triage-pack), code-smeller (hygiene-pack), any quality assessment workflow

---

## Context

| Reference | Location |
|-----------|----------|
| Smell Report Template | `/Users/tomtenuta/Code/roster/.claude/skills/doc-ecosystem/templates/smell-report.md` |
| Code Smeller Agent | `/Users/tomtenuta/Code/roster/rites/hygiene-pack/agents/code-smeller.md` |
| Debt Collector Agent | `/Users/tomtenuta/Code/roster/rites/debt-triage-pack/agents/debt-collector.md` |
| E2E Debt Remediation | `/Users/tomtenuta/Code/roster/docs/testing/e2e-debt-remediation.md` |
| Shared Skills README | `/Users/tomtenuta/Code/roster/rites/shared/README.md` |

### Problem Statement

Currently, smell detection is implicitly defined across multiple agents and templates:

1. **code-smeller.md** references `@smell-detection` but the skill does not exist
2. **debt-collector.md** references `@smell-detection` for systematic detection
3. **smell-report.md** template lists categories without detection criteria
4. Agents make ad-hoc decisions about what constitutes a smell

Without a centralized skill, detection is inconsistent across teams, severity classification varies, and there is no principled mapping from smells to debt ledger entries.

### Design Goals

1. Define comprehensive smell taxonomy covering code, architecture, and process smells
2. Specify detection heuristics with automated hints and manual inspection patterns
3. Establish severity classification algorithm with consistent scoring
4. Document integration with debt-ledger entries and smell reports
5. Enable reuse across debt-triage-pack, hygiene-pack, and future quality teams

---

## Design Decisions

### Decision 1: Smell Category Hierarchy

**Options Considered**:

| Option | Description | Pros | Cons |
|--------|-------------|------|------|
| A. Flat list | All smells at same level | Simple | No organization as list grows |
| B. Two-level | Category > Smell Type | Balanced complexity | May need third level for some categories |
| C. Three-level | Category > Subcategory > Smell | Maximum organization | Overhead for simple smells |

**Selected**: Option B - Two-level hierarchy (Category > Smell Type)

**Rationale**: Two levels provide sufficient organization without excessive nesting. Categories align with existing smell-report template sections (Dead Code, DRY Violations, Complexity Hotspots, Naming Inconsistencies, Import Hygiene). Each category contains specific smell types with detection criteria.

### Decision 2: Detection Approach Classification

**Selected**: Three detection approaches per smell type

| Approach | Description | When to Use |
|----------|-------------|-------------|
| **Automated** | Tool-based detection with specific commands | High-confidence patterns with clear signatures |
| **Semi-Automated** | Tool output + human judgment | Patterns needing context interpretation |
| **Manual** | Code review patterns and checklists | Subjective or context-dependent smells |

**Rationale**: Not all smells are amenable to automation. Explicit classification guides agent behavior: automated checks run first, semi-automated next, manual only when necessary.

### Decision 3: Severity Classification Algorithm

**Selected**: Multi-factor weighted scoring

```
severity_score = (impact * 3) + (frequency * 2) + (blast_radius * 2) - (fix_complexity * 1)
```

| Factor | Weight | Scale | Description |
|--------|--------|-------|-------------|
| Impact | 3x | 1-3 | Business/user impact if unaddressed |
| Frequency | 2x | 1-3 | How often the smell causes problems |
| Blast Radius | 2x | 1-3 | Files/components affected by the smell |
| Fix Complexity | -1x | 1-3 | Effort to resolve (inverse: higher = lower score) |

**Score to Severity Mapping**:

| Score Range | Severity | Priority | Action |
|-------------|----------|----------|--------|
| 16-21 | Critical | P1 | Address immediately |
| 11-15 | High | P2 | Address in sprint |
| 6-10 | Medium | P3 | Address opportunistically |
| 1-5 | Low | P4 | Track for future |

**Rationale**: Aligns with existing risk-assessor scoring from debt-triage-pack (see `/Users/tomtenuta/Code/roster/docs/testing/e2e-debt-remediation.md` lines 149-168). Weighting prioritizes business impact while accounting for fix complexity as a practical constraint.

### Decision 4: Integration with Debt Ledger

**Selected**: Bidirectional reference with type mapping

| Smell Category | Debt Category | Example Mapping |
|----------------|---------------|-----------------|
| Dead Code | Code > Dead Code | DC-001 -> C042 |
| DRY Violations | Code > Duplication | DRY-001 -> C043 |
| Complexity Hotspots | Code > Complexity | CX-001 -> C044 |
| Naming Inconsistencies | Code > Naming | NM-001 -> C045 |
| Import Hygiene | Code > Dependencies | IM-001 -> C046 |
| Architecture Smells | Design > Coupling | AR-001 -> D001 |
| Process Smells | Process > Workflow | PR-001 -> P001 |

**Rationale**: Smells detected by code-smeller/debt-collector map to debt ledger entries with preserved references. Enables traceability from smell report to debt ledger to sprint packages.

---

## Smell Taxonomy

### Category 1: Dead Code

Code that is not executed, not reachable, or serves no purpose.

| Smell Type | ID Pattern | Description | Example |
|------------|------------|-------------|---------|
| Unused Functions | DC-FN | Functions with zero call sites | `function formatPhone() {}` never called |
| Unused Variables | DC-VAR | Variables declared but never read | `const temp = x;` never used |
| Unreachable Code | DC-UNREACH | Code after unconditional return/throw | Code after `return` statement |
| Orphaned Modules | DC-MOD | Files with no imports | `legacy-utils.ts` imported nowhere |
| Zombie Imports | DC-IMP | Imports whose symbols are unused | `import { foo } from 'bar'` where `foo` unused |
| Dead Branches | DC-BRANCH | Conditional branches that can never execute | `if (false) { ... }` |
| Commented Code | DC-COMMENT | Commented-out code blocks | `// function oldImplementation() { ... }` |

#### Detection Heuristics: Dead Code

| Smell | Automated Detection | Semi-Automated | Manual |
|-------|---------------------|----------------|--------|
| DC-FN | `grep -rn "function_name" --include='*.ts'` counts < 2 (declaration only) | AST analysis with tree-sitter | Review exported functions across module boundaries |
| DC-VAR | ESLint `no-unused-vars`, TypeScript compiler | Variable shadowing analysis | Dynamic property access patterns |
| DC-UNREACH | TypeScript `--noUnusedLocals`, ESLint `no-unreachable` | Control flow analysis | Exception-based flow patterns |
| DC-MOD | Build dependency graph, find zero-incoming-edge nodes | Bundle analyzer output | Test-only modules, conditional imports |
| DC-IMP | ESLint `no-unused-imports`, TypeScript errors | Re-export chains | Dynamic imports, type-only imports |
| DC-BRANCH | Static analysis: constant conditions | Feature flag analysis | A/B test remnants |
| DC-COMMENT | Grep for multi-line comment blocks with code patterns | Diff against recent commits | Documentation vs. dead code |

#### Severity Defaults: Dead Code

| Smell | Default Impact | Default Frequency | Default Blast Radius | Default Fix Complexity |
|-------|----------------|-------------------|----------------------|------------------------|
| DC-FN | Low (1) | Low (1) | Low (1) | Low (1) |
| DC-VAR | Low (1) | Medium (2) | Low (1) | Low (1) |
| DC-UNREACH | Medium (2) | Low (1) | Low (1) | Low (1) |
| DC-MOD | Medium (2) | Low (1) | Medium (2) | Low (1) |
| DC-IMP | Low (1) | Medium (2) | Low (1) | Low (1) |
| DC-BRANCH | Medium (2) | Low (1) | Low (1) | Low (1) |
| DC-COMMENT | Low (1) | Medium (2) | Low (1) | Low (1) |

---

### Category 2: DRY Violations

Duplicated logic that should be consolidated.

| Smell Type | ID Pattern | Description | Example |
|------------|------------|-------------|---------|
| Copy-Paste Code | DRY-COPY | Near-identical code blocks | Same validation in 3 files |
| Repeated Constants | DRY-CONST | Magic values duplicated | `timeout: 30000` in multiple places |
| Parallel Implementations | DRY-PARA | Different implementations of same concept | Two email validators |
| Config Drift | DRY-CFG | Same setting in multiple configs | `port: 3000` in dev/staging/prod |
| Test Duplication | DRY-TEST | Identical test setup/teardown | Same beforeEach in 10 test files |

#### Detection Heuristics: DRY Violations

| Smell | Automated Detection | Semi-Automated | Manual |
|-------|---------------------|----------------|--------|
| DRY-COPY | `jscpd` (copy-paste detector), `simian` | Diff similarity scoring | Semantic duplication (same logic, different syntax) |
| DRY-CONST | Grep for repeated literals: `grep -rn "30000"` | Config file analysis | Intentional vs. accidental duplication |
| DRY-PARA | Symbol search for similar names: `grep -rn "validate.*[Ee]mail"` | Interface comparison | Business rule alignment |
| DRY-CFG | Diff config files: `diff -y dev.json prod.json` | Config schema analysis | Environment-specific vs. duplicated |
| DRY-TEST | Test file structure analysis | Fixture comparison | Intentional isolation vs. duplication |

#### Severity Defaults: DRY Violations

| Smell | Default Impact | Default Frequency | Default Blast Radius | Default Fix Complexity |
|-------|----------------|-------------------|----------------------|------------------------|
| DRY-COPY | High (3) | High (3) | Medium (2) | Low (1) |
| DRY-CONST | Medium (2) | Medium (2) | Medium (2) | Low (1) |
| DRY-PARA | High (3) | Medium (2) | High (3) | Medium (2) |
| DRY-CFG | Medium (2) | Low (1) | Low (1) | Low (1) |
| DRY-TEST | Low (1) | Medium (2) | Low (1) | Medium (2) |

---

### Category 3: Complexity Hotspots

Code that is excessively complex or difficult to understand.

| Smell Type | ID Pattern | Description | Example |
|------------|------------|-------------|---------|
| High Cyclomatic | CX-CYCLO | Functions with many decision points | Function with 15+ branches |
| Deep Nesting | CX-NEST | Excessive indentation levels | 5+ levels of nesting |
| God Object | CX-GOD | Classes/modules with too many responsibilities | 2000+ line file |
| Long Parameter List | CX-PARAM | Functions with many parameters | `function foo(a, b, c, d, e, f, g)` |
| Boolean Blindness | CX-BOOL | Multiple boolean parameters | `createUser(true, false, true)` |
| Primitive Obsession | CX-PRIM | Using primitives instead of domain types | Passing `string` for email everywhere |
| Feature Envy | CX-ENVY | Method using more of another class's data | Function mostly accessing external state |

#### Detection Heuristics: Complexity Hotspots

| Smell | Automated Detection | Semi-Automated | Manual |
|-------|---------------------|----------------|--------|
| CX-CYCLO | `eslint-plugin-complexity`, `lizard` | Threshold adjustment per codebase | Acceptable vs. excessive for domain |
| CX-NEST | `eslint max-depth`, AST depth analysis | Context-dependent thresholds | Callback hell vs. necessary logic |
| CX-GOD | `wc -l *.ts \| sort -rn`, LOC analysis | Cohesion metrics | Single responsibility assessment |
| CX-PARAM | ESLint `max-params`, function signature scan | Parameter object candidates | API stability concerns |
| CX-BOOL | Grep for multiple `true`/`false` args | Enum refactoring candidates | Intentional flags vs. poor design |
| CX-PRIM | Type analysis: primitive frequency | Domain model review | Type safety vs. over-engineering |
| CX-ENVY | Dependency analysis: external references | Cohesion metrics | Method placement review |

#### Severity Defaults: Complexity Hotspots

| Smell | Default Impact | Default Frequency | Default Blast Radius | Default Fix Complexity |
|-------|----------------|-------------------|----------------------|------------------------|
| CX-CYCLO | High (3) | High (3) | Low (1) | Medium (2) |
| CX-NEST | Medium (2) | Medium (2) | Low (1) | Low (1) |
| CX-GOD | High (3) | Medium (2) | High (3) | High (3) |
| CX-PARAM | Medium (2) | Medium (2) | Low (1) | Medium (2) |
| CX-BOOL | Low (1) | Medium (2) | Low (1) | Low (1) |
| CX-PRIM | Medium (2) | High (3) | Medium (2) | High (3) |
| CX-ENVY | Medium (2) | Medium (2) | Medium (2) | Medium (2) |

---

### Category 4: Naming Inconsistencies

Terminology drift, misleading identifiers, and convention violations.

| Smell Type | ID Pattern | Description | Example |
|------------|------------|-------------|---------|
| Inconsistent Naming | NM-INCONSIST | Same concept with different names | `user`, `account`, `member` for same entity |
| Misleading Names | NM-MISLEAD | Names that don't match behavior | `getUser()` that modifies state |
| Convention Violations | NM-CONV | Breaks project naming rules | `camelCase` in `snake_case` codebase |
| Abbreviation Soup | NM-ABBREV | Excessive or inconsistent abbreviations | `usr`, `usrAcct`, `userAccount` |
| Type-Name Mismatch | NM-TYPE | Variable name doesn't match type | `const user: Account = ...` |

#### Detection Heuristics: Naming Inconsistencies

| Smell | Automated Detection | Semi-Automated | Manual |
|-------|---------------------|----------------|--------|
| NM-INCONSIST | Symbol extraction + clustering | Synonym detection | Domain expert review |
| NM-MISLEAD | Static analysis: side effects in getters | Behavior analysis | Code review patterns |
| NM-CONV | ESLint naming rules, custom regex | Convention documentation | Project-specific rules |
| NM-ABBREV | Abbreviation dictionary check | Consistency analysis | Team glossary alignment |
| NM-TYPE | TypeScript: type vs. name comparison | Semantic analysis | Intent verification |

#### Severity Defaults: Naming Inconsistencies

| Smell | Default Impact | Default Frequency | Default Blast Radius | Default Fix Complexity |
|-------|----------------|-------------------|----------------------|------------------------|
| NM-INCONSIST | Medium (2) | High (3) | High (3) | Medium (2) |
| NM-MISLEAD | High (3) | Low (1) | Medium (2) | Low (1) |
| NM-CONV | Low (1) | Medium (2) | Medium (2) | Low (1) |
| NM-ABBREV | Low (1) | Medium (2) | Low (1) | Low (1) |
| NM-TYPE | Medium (2) | Low (1) | Low (1) | Low (1) |

---

### Category 5: Import Hygiene

Dependency management issues and import organization problems.

| Smell Type | ID Pattern | Description | Example |
|------------|------------|-------------|---------|
| Circular Dependencies | IM-CIRC | A imports B imports A | Module A <-> Module B |
| Wildcard Imports | IM-WILD | Importing everything | `import * from 'utils'` |
| Deep Imports | IM-DEEP | Importing from internal paths | `import { x } from 'lib/internal/private'` |
| Barrel Bloat | IM-BARREL | Index files re-exporting too much | `index.ts` with 50+ exports |
| Version Skew | IM-VERSION | Same dependency at multiple versions | `lodash@4.17.0` and `lodash@4.17.21` |
| Unused Dependencies | IM-UNUSED | Package.json deps not imported | `"axios": "^1.0.0"` never used |

#### Detection Heuristics: Import Hygiene

| Smell | Automated Detection | Semi-Automated | Manual |
|-------|---------------------|----------------|--------|
| IM-CIRC | `madge --circular`, `dpdm` | Cycle-breaking analysis | Intentional cycles (rare) |
| IM-WILD | ESLint `no-import-all`, grep `import \*` | Tree-shaking impact | Performance measurement |
| IM-DEEP | Path depth analysis, internal path patterns | API boundary definition | Intended exposure |
| IM-BARREL | Export count per index file | Bundle impact analysis | Monorepo patterns |
| IM-VERSION | `npm ls`, `yarn why`, lockfile analysis | Deduplication feasibility | Compatibility constraints |
| IM-UNUSED | `depcheck`, `npm-check` | Test/build dependency separation | Dynamic requires |

#### Severity Defaults: Import Hygiene

| Smell | Default Impact | Default Frequency | Default Blast Radius | Default Fix Complexity |
|-------|----------------|-------------------|----------------------|------------------------|
| IM-CIRC | High (3) | Medium (2) | High (3) | High (3) |
| IM-WILD | Medium (2) | Medium (2) | Low (1) | Low (1) |
| IM-DEEP | Medium (2) | Low (1) | Medium (2) | Low (1) |
| IM-BARREL | Low (1) | Low (1) | Medium (2) | Medium (2) |
| IM-VERSION | Medium (2) | Low (1) | Medium (2) | Medium (2) |
| IM-UNUSED | Low (1) | Low (1) | Low (1) | Low (1) |

---

### Category 6: Architecture Smells

Structural issues that affect system design and maintainability.

| Smell Type | ID Pattern | Description | Example |
|------------|------------|-------------|---------|
| Leaky Abstraction | AR-LEAK | Implementation details exposed | Internal errors bubbling to API |
| Tight Coupling | AR-COUPLE | Excessive dependencies between modules | Module A calls 15 functions from B |
| Layer Violation | AR-LAYER | Wrong-direction dependencies | Domain importing from UI |
| Missing Abstraction | AR-MISSING | Repeated patterns without encapsulation | Same try/catch in every handler |
| Shotgun Surgery | AR-SHOT | One change requires many file edits | Adding field touches 10 files |
| Divergent Change | AR-DIVERGE | One file changes for many reasons | `utils.ts` edited every sprint |

#### Detection Heuristics: Architecture Smells

| Smell | Automated Detection | Semi-Automated | Manual |
|-------|---------------------|----------------|--------|
| AR-LEAK | Error type analysis, API response schemas | Boundary testing | Design review |
| AR-COUPLE | Dependency graph: edge count analysis | Coupling metrics (CBO) | Architecture review |
| AR-LAYER | Import path rules, layer definitions | Dependency direction analysis | Architecture diagram comparison |
| AR-MISSING | Pattern detection: similar code structures | Refactoring candidate identification | Design session |
| AR-SHOT | Git history: files changed together | Change coupling analysis | Feature-level review |
| AR-DIVERGE | Git history: change frequency per file | Reason analysis | Responsibility review |

#### Severity Defaults: Architecture Smells

| Smell | Default Impact | Default Frequency | Default Blast Radius | Default Fix Complexity |
|-------|----------------|-------------------|----------------------|------------------------|
| AR-LEAK | High (3) | Medium (2) | Medium (2) | Medium (2) |
| AR-COUPLE | High (3) | High (3) | High (3) | High (3) |
| AR-LAYER | High (3) | Low (1) | High (3) | Medium (2) |
| AR-MISSING | Medium (2) | High (3) | Medium (2) | Medium (2) |
| AR-SHOT | High (3) | Medium (2) | High (3) | High (3) |
| AR-DIVERGE | Medium (2) | High (3) | Medium (2) | High (3) |

---

### Category 7: Process Smells

Issues in development workflow, testing, and documentation.

| Smell Type | ID Pattern | Description | Example |
|------------|------------|-------------|---------|
| Missing Tests | PR-TEST | Code without test coverage | New feature with 0% coverage |
| Flaky Tests | PR-FLAKY | Tests that fail intermittently | Random CI failures |
| Slow Tests | PR-SLOW | Tests exceeding time thresholds | 30-second unit test |
| Outdated Docs | PR-DOCS | Documentation not matching code | API docs referencing removed endpoints |
| TODO Accumulation | PR-TODO | Growing TODO/FIXME count | 50+ unresolved TODOs |
| Disabled Tests | PR-SKIP | Tests marked skip/pending | `it.skip('should work')` |

#### Detection Heuristics: Process Smells

| Smell | Automated Detection | Semi-Automated | Manual |
|-------|---------------------|----------------|--------|
| PR-TEST | Coverage tools: `jest --coverage`, `go test -cover` | Coverage trend analysis | Critical path identification |
| PR-FLAKY | CI history analysis: same test, different outcomes | Retry detection | Root cause investigation |
| PR-SLOW | Test timing reports | Threshold configuration | Acceptable vs. problematic |
| PR-DOCS | Link checking, API comparison tools | Documentation review | Accuracy verification |
| PR-TODO | Grep `TODO\|FIXME\|HACK\|XXX`, count tracking | Age analysis (git blame) | Priority assessment |
| PR-SKIP | Grep `skip\|pending\|xdescribe\|xit` | Reason documentation | Re-enable feasibility |

#### Severity Defaults: Process Smells

| Smell | Default Impact | Default Frequency | Default Blast Radius | Default Fix Complexity |
|-------|----------------|-------------------|----------------------|------------------------|
| PR-TEST | High (3) | Medium (2) | Medium (2) | Medium (2) |
| PR-FLAKY | High (3) | High (3) | Low (1) | Medium (2) |
| PR-SLOW | Medium (2) | Medium (2) | Low (1) | Medium (2) |
| PR-DOCS | Medium (2) | Medium (2) | Medium (2) | Low (1) |
| PR-TODO | Low (1) | High (3) | Low (1) | Low (1) |
| PR-SKIP | Medium (2) | Low (1) | Low (1) | Low (1) |

---

## Severity Classification Rules

### Algorithm Implementation

```
Input: smell_type, context_overrides (optional)

1. Load default severity factors for smell_type
2. Apply context_overrides if provided (impact, frequency, blast_radius, fix_complexity)
3. Calculate severity_score:
   score = (impact * 3) + (frequency * 2) + (blast_radius * 2) - (fix_complexity * 1)
4. Map score to severity level:
   - 16-21: CRITICAL (P1)
   - 11-15: HIGH (P2)
   - 6-10: MEDIUM (P3)
   - 1-5: LOW (P4)
5. Return { severity, priority, score, factors }
```

### Context Override Examples

| Context | Factor Adjustment | Rationale |
|---------|-------------------|-----------|
| Critical path code | Impact +1 | Higher business impact |
| Frequently modified files | Frequency +1 | More likely to cause issues |
| Public API surface | Blast Radius +1 | Affects consumers |
| Simple extraction refactor | Fix Complexity -1 | Lower effort than default |
| Legacy module, no tests | Fix Complexity +1 | Riskier to change |

### Severity Output Format

```yaml
smell_id: DC-FN-001
smell_type: DC-FN
location: src/utils/formatters.ts:67
severity: MEDIUM
priority: P3
score: 8
factors:
  impact: 1
  frequency: 2
  blast_radius: 2
  fix_complexity: 1
context_notes: "Hot path: increased frequency score"
```

---

## Debt Ledger Integration

### Smell-to-Debt Mapping Schema

```yaml
# Smell Report Entry
smell_id: DRY-COPY-001
smell_type: DRY-COPY
category: DRY Violations
locations:
  - src/validators/user.ts:45-62
  - src/validators/contact.ts:23-40
severity: HIGH
priority: P2
score: 14

# Corresponding Debt Ledger Entry
debt_id: C043
smell_ref: DRY-COPY-001
category: Code > Duplication
location: src/validators/user.ts:45 (and 1 other)
description: Duplicate email validation across 2 files
evidence: "85% similarity, 18 lines each (smell-detection DRY-COPY heuristic)"
age: 14 months (git blame)
owner: "@platform-team"
```

### Integration Points

| Source | Target | Integration Type |
|--------|--------|------------------|
| code-smeller (hygiene-pack) | Smell Report | Produces |
| debt-collector (debt-triage-pack) | Debt Ledger | Produces |
| Smell Report | Debt Ledger | Reference (smell_ref field) |
| Debt Ledger | Sprint Packages | Reference (debt_id field) |
| Sprint Packages | HANDOFF | Reference (PKG-XXX field) |

### Cross-Team Flow

```
smell-detection skill (patterns)
        |
        v
+-------------------+     +-------------------+
| code-smeller      |     | debt-collector    |
| (hygiene-pack)    |     | (debt-triage-pack)|
+-------------------+     +-------------------+
        |                         |
        v                         v
  Smell Report              Debt Ledger
        |                         |
        +----------+  +-----------+
                   |  |
                   v  v
            risk-assessor
                   |
                   v
            Sprint Packages
                   |
                   v
                HANDOFF
```

---

## Skill File Structure

```
rites/shared/skills/smell-detection/
    |
    +-- SKILL.md                 # Entry point: quick reference, triggers
    +-- taxonomy/
    |   +-- dead-code.md         # DC-* smell types with heuristics
    |   +-- dry-violations.md    # DRY-* smell types with heuristics
    |   +-- complexity.md        # CX-* smell types with heuristics
    |   +-- naming.md            # NM-* smell types with heuristics
    |   +-- imports.md           # IM-* smell types with heuristics
    |   +-- architecture.md      # AR-* smell types with heuristics
    |   +-- process.md           # PR-* smell types with heuristics
    |
    +-- severity/
    |   +-- classification.md    # Algorithm, factor weights, score mapping
    |   +-- defaults.md          # Default severity per smell type
    |   +-- overrides.md         # Context-based adjustment patterns
    |
    +-- integration/
    |   +-- debt-ledger.md       # Smell-to-debt mapping schema
    |   +-- smell-report.md      # Smell report field requirements
    |   +-- tooling.md           # Automated detection tool reference
```

---

## SKILL.md Content

```markdown
---
name: smell-detection
description: "Cross-team code smell detection patterns with severity classification and debt integration. Use when: detecting code quality issues, cataloging smells for reports or debt ledgers, classifying smell severity. Triggers: smell detection, code smell, dead code, duplication, complexity, naming, imports, architecture smell, process smell, smell taxonomy, severity classification."
---

# smell-detection

> Unified smell detection patterns for code quality assessment.

## Purpose

Provides canonical smell taxonomy, detection heuristics, and severity classification for use across debt-triage-pack, hygiene-pack, and any team assessing code quality.

## Quick Reference

### Smell Categories

| Category | ID Prefix | Example Types |
|----------|-----------|---------------|
| Dead Code | DC-* | Unused functions, zombie imports, orphaned modules |
| DRY Violations | DRY-* | Copy-paste code, repeated constants, parallel implementations |
| Complexity | CX-* | High cyclomatic, deep nesting, god objects |
| Naming | NM-* | Inconsistent naming, misleading names, convention violations |
| Imports | IM-* | Circular deps, wildcard imports, version skew |
| Architecture | AR-* | Leaky abstraction, tight coupling, layer violations |
| Process | PR-* | Missing tests, flaky tests, TODO accumulation |

### Severity Classification

| Score | Severity | Priority | Action |
|-------|----------|----------|--------|
| 16-21 | Critical | P1 | Address immediately |
| 11-15 | High | P2 | Address in sprint |
| 6-10 | Medium | P3 | Address opportunistically |
| 1-5 | Low | P4 | Track for future |

### Detection Approaches

| Approach | When to Use |
|----------|-------------|
| Automated | Clear signatures, high confidence (run first) |
| Semi-Automated | Tool output + judgment (run second) |
| Manual | Subjective or context-dependent (run last) |

## When to Use

| Scenario | What to Read |
|----------|--------------|
| Detecting specific smell type | `taxonomy/{category}.md` |
| Classifying severity | `severity/classification.md` |
| Adjusting for context | `severity/overrides.md` |
| Mapping to debt ledger | `integration/debt-ledger.md` |
| Running automated checks | `integration/tooling.md` |

## Consumers

- **code-smeller** (hygiene-pack): Produces Smell Reports
- **debt-collector** (debt-triage-pack): Produces Debt Ledgers
- Any agent performing code quality assessment

## Progressive Disclosure

### Taxonomy
- [dead-code.md](taxonomy/dead-code.md) - DC-* smell types
- [dry-violations.md](taxonomy/dry-violations.md) - DRY-* smell types
- [complexity.md](taxonomy/complexity.md) - CX-* smell types
- [naming.md](taxonomy/naming.md) - NM-* smell types
- [imports.md](taxonomy/imports.md) - IM-* smell types
- [architecture.md](taxonomy/architecture.md) - AR-* smell types
- [process.md](taxonomy/process.md) - PR-* smell types

### Severity
- [classification.md](severity/classification.md) - Algorithm and weights
- [defaults.md](severity/defaults.md) - Default severity per type
- [overrides.md](severity/overrides.md) - Context adjustments

### Integration
- [debt-ledger.md](integration/debt-ledger.md) - Mapping schema
- [smell-report.md](integration/smell-report.md) - Report format
- [tooling.md](integration/tooling.md) - Automated tools

## Related Skills

- [doc-ecosystem](../../ecosystem-pack/skills/doc-ecosystem/SKILL.md) - Smell report template
- [standards](../../../.claude/skills/standards/SKILL.md) - Code conventions
```

---

## Backward Compatibility

This skill is **NEW** - no backward compatibility concerns.

**Integration with existing agents**:
- code-smeller.md already references `@smell-detection` (line 58, 169) - skill now exists
- debt-collector.md already references `@smell-detection` (line 59, 124) - skill now exists
- No agent changes required; skill provides referenced patterns

---

## Test Matrix

### Taxonomy Completeness

| Test ID | Category | Check | Expected |
|---------|----------|-------|----------|
| tax_001 | Dead Code | All 7 DC-* types documented | PASS |
| tax_002 | DRY Violations | All 5 DRY-* types documented | PASS |
| tax_003 | Complexity | All 7 CX-* types documented | PASS |
| tax_004 | Naming | All 5 NM-* types documented | PASS |
| tax_005 | Imports | All 6 IM-* types documented | PASS |
| tax_006 | Architecture | All 6 AR-* types documented | PASS |
| tax_007 | Process | All 6 PR-* types documented | PASS |

### Detection Heuristics

| Test ID | Smell | Detection Type | Expected |
|---------|-------|----------------|----------|
| det_001 | DC-FN | Automated | grep command documented |
| det_002 | DRY-COPY | Automated | jscpd tool documented |
| det_003 | CX-CYCLO | Automated | eslint-plugin-complexity documented |
| det_004 | NM-INCONSIST | Manual | Symbol clustering pattern documented |
| det_005 | IM-CIRC | Automated | madge command documented |
| det_006 | AR-COUPLE | Semi-Automated | Coupling metrics documented |
| det_007 | PR-TODO | Automated | grep pattern documented |

### Severity Classification

| Test ID | Input | Expected Score | Expected Severity |
|---------|-------|----------------|-------------------|
| sev_001 | Impact=3, Freq=3, Blast=3, Fix=1 | 20 | CRITICAL |
| sev_002 | Impact=2, Freq=2, Blast=2, Fix=2 | 10 | MEDIUM |
| sev_003 | Impact=1, Freq=1, Blast=1, Fix=1 | 4 | LOW |
| sev_004 | Impact=3, Freq=2, Blast=2, Fix=3 | 14 | HIGH |
| sev_005 | Impact=1, Freq=1, Blast=1, Fix=3 | 1 | LOW |

### Debt Integration

| Test ID | Check | Expected |
|---------|-------|----------|
| int_001 | Smell-to-debt category mapping complete | All 7 categories mapped |
| int_002 | smell_ref field schema documented | YAML example provided |
| int_003 | Cross-team flow diagram accurate | Matches actual workflow |

---

## Success Criteria

- [ ] SKILL.md created with frontmatter and quick reference
- [ ] All 7 smell categories documented with types and heuristics
- [ ] Detection heuristics specified for each smell type (automated, semi-automated, manual)
- [ ] Severity classification algorithm documented with weights and score mapping
- [ ] Default severity factors defined for all 42 smell types
- [ ] Context override patterns documented
- [ ] Debt ledger integration schema defined
- [ ] Skill file structure created under rites/shared/skills/smell-detection/
- [ ] Existing agent references (`@smell-detection`) now resolve

---

## Implementation Guidance

### Recommended Implementation Order

1. **Create SKILL.md** with frontmatter and quick reference
2. **Create taxonomy/ directory** with category files
3. **Create severity/ directory** with classification and defaults
4. **Create integration/ directory** with debt-ledger mapping
5. **Verify agent references** in code-smeller.md and debt-collector.md

### File Size Targets

| File | Target Lines | Purpose |
|------|--------------|---------|
| SKILL.md | 80-100 | Quick reference, progressive disclosure links |
| taxonomy/*.md | 50-100 each | Smell types with detection tables |
| severity/classification.md | 60-80 | Algorithm, weights, examples |
| severity/defaults.md | 80-100 | All 42 smell type defaults |
| integration/debt-ledger.md | 40-60 | Mapping schema and examples |

---

## Appendix: Complete Smell Type Reference

| ID | Category | Type | Default Severity |
|----|----------|------|------------------|
| DC-FN | Dead Code | Unused Functions | LOW |
| DC-VAR | Dead Code | Unused Variables | LOW |
| DC-UNREACH | Dead Code | Unreachable Code | MEDIUM |
| DC-MOD | Dead Code | Orphaned Modules | MEDIUM |
| DC-IMP | Dead Code | Zombie Imports | LOW |
| DC-BRANCH | Dead Code | Dead Branches | MEDIUM |
| DC-COMMENT | Dead Code | Commented Code | LOW |
| DRY-COPY | DRY Violations | Copy-Paste Code | HIGH |
| DRY-CONST | DRY Violations | Repeated Constants | MEDIUM |
| DRY-PARA | DRY Violations | Parallel Implementations | HIGH |
| DRY-CFG | DRY Violations | Config Drift | MEDIUM |
| DRY-TEST | DRY Violations | Test Duplication | LOW |
| CX-CYCLO | Complexity | High Cyclomatic | HIGH |
| CX-NEST | Complexity | Deep Nesting | MEDIUM |
| CX-GOD | Complexity | God Object | HIGH |
| CX-PARAM | Complexity | Long Parameter List | MEDIUM |
| CX-BOOL | Complexity | Boolean Blindness | LOW |
| CX-PRIM | Complexity | Primitive Obsession | MEDIUM |
| CX-ENVY | Complexity | Feature Envy | MEDIUM |
| NM-INCONSIST | Naming | Inconsistent Naming | MEDIUM |
| NM-MISLEAD | Naming | Misleading Names | HIGH |
| NM-CONV | Naming | Convention Violations | LOW |
| NM-ABBREV | Naming | Abbreviation Soup | LOW |
| NM-TYPE | Naming | Type-Name Mismatch | MEDIUM |
| IM-CIRC | Imports | Circular Dependencies | HIGH |
| IM-WILD | Imports | Wildcard Imports | MEDIUM |
| IM-DEEP | Imports | Deep Imports | MEDIUM |
| IM-BARREL | Imports | Barrel Bloat | LOW |
| IM-VERSION | Imports | Version Skew | MEDIUM |
| IM-UNUSED | Imports | Unused Dependencies | LOW |
| AR-LEAK | Architecture | Leaky Abstraction | HIGH |
| AR-COUPLE | Architecture | Tight Coupling | CRITICAL |
| AR-LAYER | Architecture | Layer Violation | HIGH |
| AR-MISSING | Architecture | Missing Abstraction | MEDIUM |
| AR-SHOT | Architecture | Shotgun Surgery | HIGH |
| AR-DIVERGE | Architecture | Divergent Change | MEDIUM |
| PR-TEST | Process | Missing Tests | HIGH |
| PR-FLAKY | Process | Flaky Tests | HIGH |
| PR-SLOW | Process | Slow Tests | MEDIUM |
| PR-DOCS | Process | Outdated Docs | MEDIUM |
| PR-TODO | Process | TODO Accumulation | LOW |
| PR-SKIP | Process | Disabled Tests | MEDIUM |

---

## Version

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-03 | Context Architect | Initial design |

---
name: defensive-patterns-criteria
description: "Observation criteria for codebase defensive pattern knowledge capture. Use when: theoros is producing defensive pattern knowledge for .know/, documenting guards, assertions, constraints, and risk zones. Triggers: defensive pattern knowledge criteria, guard inventory observation, assertion documentation, constraint catalog, risk zone mapping."
---

# Defensive Patterns Observation Criteria

> The theoros observes and documents codebase defensive patterns -- producing a knowledge reference that catalogs every guard, assertion, constraint, and risk zone, enabling any CC agent to understand the safety net and where it has gaps.

## Scope

**Target files**: `./cmd/`, `./internal/`, root-level Go files (excluding `*_test.go` for guard inventory, but including them for guard validation evidence)

**Observation focus**: Guards (validation, assertion, error checking), constraints (bounds, invariants, type safety), and unguarded risk zones. Evidence from error handling patterns, assertion statements, validation functions, and circuit breakers.

**NOTE**: This domain uses knowledge-capture grading. Instead of grading compliance ratios ("90% of packages have X"), grade the COMPLETENESS of the knowledge reference produced. A = comprehensive documentation of the codebase with evidence. F = incomplete or inaccurate documentation.

## Criteria

### Criterion 1: Guard Inventory Completeness (weight: 30%)

**What to observe**: Exhaustive catalog of guards, assertions, validation checks, and error boundary handlers across the codebase. The knowledge reference must give a reader a complete map of where the safety net exists.

**Evidence to collect**:
- Scan all packages under `cmd/` and `internal/` for validation functions, assertion helpers, and error boundary handlers
- Record numbered GUARD-NNN entries: location (file:line), trigger condition (what input or state triggers the guard), and failure-without-guard description (what would go wrong if this guard were removed)
- Identify guard categories: input validation, state assertion, boundary enforcement, nil-safety, type safety, range checking
- Note any shared guard utilities or validation helper functions used across packages
- Flag guards that are load-bearing (called from multiple callers, removing breaks multiple paths)

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | Every significant guard documented with GUARD-NNN entry, location, trigger condition, and failure mode. Categories classified. Load-bearing guards identified. Minimum 30 guards for mature codebase with no omissions. |
| B | 80-89% completeness | All major guards cataloged. Most have trigger conditions and failure modes. Minor gaps in category classification or load-bearing identification. |
| C | 70-79% completeness | Key guards documented but some packages not fully scanned. Trigger conditions partially described. Load-bearing identification incomplete. |
| D | 60-69% completeness | Major guards listed but descriptions are superficial. Trigger conditions and failure modes mostly absent. Categories not classified. |
| F | < 60% completeness | Fewer than half the guards documented, or entries are inaccurate or missing trigger conditions. |

---

### Criterion 2: Risk Zone Mapping (weight: 25%)

**What to observe**: Identification of unguarded areas where defenses are missing but should exist. The knowledge reference must document where the safety net has holes so agents do not introduce new failures in these areas.

**Evidence to collect**:
- Identify code paths where validation is absent but caller-responsibility is assumed (look for comments like "caller must ensure", "assumes valid", "precondition:")
- Scan for silent fallbacks that swallow errors or use zero-values without logging
- Note input paths from external sources (CLI flags, config files, stdin, environment variables) that lack validation
- Record RISK-NNN entries: location, type of missing guard, evidence of missing protection, and recommended guard
- Document the guard-to-risk ratio: for every guarded path, estimate the proportion of similar paths that are unguarded

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | All risk zones documented with RISK-NNN entries. Each entry includes location, missing guard type, evidence, and recommended guard. Guard-to-risk ratio documented per package. Silent fallbacks cataloged. |
| B | 80-89% completeness | Major risk zones identified. Most entries have location and missing guard type. Guard-to-risk ratio estimated. Minor gaps in silent fallback cataloging. |
| C | 70-79% completeness | Some risk zones noted but coverage is incomplete. Entries lack recommended guards. Guard-to-risk ratio not documented. |
| D | 60-69% completeness | Risk zones mentioned vaguely without specific locations or evidence. No RISK-NNN entries. |
| F | < 60% completeness | Risk zones not documented or only trivially identified. |

---

### Criterion 3: Assertion Pattern Documentation (weight: 20%)

**What to observe**: Documentation of assertion and validation patterns used across the codebase. The knowledge reference must tell an agent which pattern to use in a given context and when each pattern is appropriate vs inappropriate.

**Evidence to collect**:
- Identify the primary assertion/validation patterns in use: `if err != nil`, `panic`, custom error types, validation structs, must-style helpers
- For each pattern, record: name, file references where used, frequency count, and the context in which it is used
- Document when each pattern is appropriate (e.g., panic for programming errors, error return for operational errors)
- Note any anti-patterns: places where the wrong assertion pattern was used and what should have been used instead
- Check for consistent vs inconsistent pattern application across the codebase

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | All assertion patterns inventoried with file references and frequency counts. Appropriate-use contexts documented. Anti-patterns identified. An agent could choose the right pattern for any new validation. |
| B | 80-89% completeness | Major patterns documented with references. Appropriate-use contexts covered for most patterns. Minor gaps in anti-pattern identification. |
| C | 70-79% completeness | Core patterns named but without frequency counts or file references. Appropriate-use contexts partially documented. |
| D | 60-69% completeness | Patterns listed by name only. No usage context or file references. Anti-patterns not identified. |
| F | < 60% completeness | Assertion patterns not inventoried or descriptions are inaccurate. |

---

### Criterion 4: Constraint Catalog (weight: 15%)

**What to observe**: Documentation of bounds checking, configuration guards, invariant enforcement, and type safety mechanisms. The knowledge reference must catalog constraints grouped by category so agents can find all constraints relevant to a given concern.

**Evidence to collect**:
- Identify constraints by category: data integrity (schema validation, required fields), security (auth checks, sanitization), freshness (TTL checks, staleness guards), concurrency (mutex guards, channel selects), configuration (required env vars, flag validation), dependency ordering (init guards, lifecycle checks)
- For each constraint, record: category, location, what invariant it enforces, and what breaks if removed
- Note any constraint hierarchies: constraint A assumes constraint B has already been enforced upstream
- Document constraints that are enforced at multiple layers vs single-layer constraints

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | Constraints organized by all relevant categories. Each entry includes location, invariant description, and removal consequence. Constraint hierarchies documented. Multi-layer vs single-layer classification included. |
| B | 80-89% completeness | Most constraint categories documented. Good coverage of location and invariant description. Minor gaps in hierarchy documentation. |
| C | 70-79% completeness | Key constraint categories covered but some omitted. Invariant descriptions partially complete. Hierarchies not documented. |
| D | 60-69% completeness | Constraints listed without category organization. Invariant descriptions superficial. |
| F | < 60% completeness | Constraint catalog absent or fewer than half of significant constraints documented. |

---

### Criterion 5: Boundary Enforcement Documentation (weight: 10%)

**What to observe**: How defensive patterns interact at package and layer boundaries -- which guards are load-bearing and which are redundant. The knowledge reference must document the dependency graph between guards so agents understand what depends on what.

**Evidence to collect**:
- Identify guards that enforce contracts at package boundaries (guards that exist because callers cannot be trusted to validate upstream)
- Document guard dependencies: guard A depends on guard B having already run (e.g., a nil-check that relies on an earlier initialization guard)
- Identify redundant guards (same invariant enforced at multiple layers) and document whether the redundancy is intentional
- Note boundary-crossing enforcement mechanisms: how guards in one package signal failures to guards in another package
- Document which guards are the authoritative enforcement point for a given invariant

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | Package boundary guards identified with caller-trust rationale. Guard dependency graph documented. Redundant guards flagged with intentionality noted. Authoritative enforcement points identified for all major invariants. |
| B | 80-89% completeness | Major boundary guards documented. Guard dependencies partially mapped. Redundant guards noted. Minor gaps in authoritative enforcement point identification. |
| C | 70-79% completeness | Boundary guards mentioned but not mapped to package dependencies. Redundancy not analyzed. |
| D | 60-69% completeness | Boundary enforcement described vaguely without specific guard-to-guard relationships. |
| F | < 60% completeness | Boundary enforcement not documented or guard dependency graph absent. |

---

## Grading Calculation

Final grade is weighted average of all criteria midpoint scores (see `schemas/grading.lego.md`). Example:
- Guard Inventory: A (midpoint 95%) x 30% = 28.5
- Risk Zone Mapping: B (midpoint 85%) x 25% = 21.25
- Assertion Patterns: B (midpoint 85%) x 20% = 17.0
- Constraint Catalog: A (midpoint 95%) x 15% = 14.25
- Boundary Enforcement: B (midpoint 85%) x 10% = 8.5
- **Total: 89.5 -> B**

## Related

- [Pinakes INDEX](../INDEX.lego.md) -- Full audit system documentation
- [architecture-criteria](architecture.lego.md) -- Codebase architecture knowledge capture
- [scar-tissue-criteria](scar-tissue.lego.md) -- Scar tissue knowledge capture (failure history, regressions)

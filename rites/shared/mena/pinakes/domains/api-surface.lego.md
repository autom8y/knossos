---
name: api-surface-criteria
description: "Observation criteria for codebase API surface knowledge capture. Use when: theoros is producing api-surface knowledge for .know/, documenting CLI contracts, exported interfaces, and public type signatures. Triggers: api surface knowledge criteria, CLI contract observation, exported interface documentation."
---

# API Surface Observation Criteria

> The theoros observes and documents the codebase's API surface -- producing a knowledge reference that enables any CC agent to understand how external consumers and internal packages interact with the system's public contracts.

## Scope

**Target files**: `./cmd/ari/`, `./internal/cmd/*/`, exported types and interfaces across `./internal/`

**Observation focus**: CLI command contracts, exported Go interfaces, public type signatures, and backward compatibility signals that a CC agent needs before modifying any public-facing surface.

**NOTE**: This domain uses knowledge-capture grading. Instead of grading compliance ratios ("90% of interfaces are documented"), grade the COMPLETENESS of the API surface reference produced. A = comprehensive documentation of the API surface with evidence. F = incomplete or inaccurate documentation.

## Criteria

### Criterion 1: CLI Command Contracts (weight: 30%)

**What to observe**: Every user-facing CLI command, its flags, arguments, and behavior contract. The knowledge reference must tell an agent exactly what the CLI exposes to users.

**Evidence to collect**:
- Trace the cobra command tree from root to all leaf commands
- For each command: name, description, flags (name, type, default, required), positional arguments
- Document command groups and their organization (subcommand hierarchy)
- Note hidden commands and their purposes
- Identify output format contracts (text, JSON via --output flag)
- Document exit code semantics (0 success, 1 error, any special codes)

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | Every CLI command documented with full flag specs, argument contracts, and output format. Hidden commands noted. Exit codes documented. An agent could modify CLI behavior without breaking contracts. |
| B | 80-89% completeness | Most commands documented with flags. Minor gaps in hidden commands or exit code details. |
| C | 70-79% completeness | Major commands documented but not all flags or arguments captured. |
| D | 60-69% completeness | Commands listed by name without flag or argument details. |
| F | < 60% completeness | CLI surface not systematically documented. |

---

### Criterion 2: Exported Interfaces (weight: 25%)

**What to observe**: Go interfaces that define contracts between packages — the abstraction seams of the codebase. The knowledge reference must tell an agent where the contracts are.

**Evidence to collect**:
- Scan all packages for exported interface types
- For each interface: name, package, method signatures, and what packages implement it
- Identify core contracts (interfaces with multiple implementations)
- Note interface composition patterns (embedding, small interfaces combined)
- Document any interfaces that cross package boundaries (defined in one, implemented in another)
- Flag interfaces that act as dependency injection points

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | All exported interfaces documented with method signatures and implementors. Cross-package contracts identified. Composition patterns noted. An agent could implement a new interface conformant to the contract. |
| B | 80-89% completeness | Most exported interfaces documented. Implementors noted for major interfaces. Minor gaps in composition pattern documentation. |
| C | 70-79% completeness | Key interfaces documented but not all implementors identified. |
| D | 60-69% completeness | Some interfaces listed without method signatures or implementor details. |
| F | < 60% completeness | Exported interfaces not systematically documented. |

---

### Criterion 3: Public Type Signatures (weight: 20%)

**What to observe**: Exported structs, functions, and constants that form the package-level API. The knowledge reference must document what each package exposes to its consumers.

**Evidence to collect**:
- For each package with significant exports: list exported types (structs, type aliases, constants)
- Document constructor patterns (`New*` functions) with parameter and return types
- Note exported function signatures for key operations
- Identify option/config structs and their field contracts
- Document any exported variables or package-level state
- Flag types that appear in multiple package APIs (shared types)

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | Exported types documented per package with constructor signatures. Option structs detailed. Shared types identified. An agent could call package APIs correctly. |
| B | 80-89% completeness | Major exported types documented. Constructor patterns noted. Minor gaps in option struct or shared type documentation. |
| C | 70-79% completeness | Key types listed but without full signature details. |
| D | 60-69% completeness | Types listed by name without package context or signatures. |
| F | < 60% completeness | Public type signatures not documented. |

---

### Criterion 4: Backward Compatibility Signals (weight: 15%)

**What to observe**: What is stable vs experimental, versioning signals, and deprecation patterns. The knowledge reference must tell an agent what is safe to change vs what requires careful migration.

**Evidence to collect**:
- Check for version constraints or stability markers (API version in paths, stable/experimental annotations)
- Identify deprecated functions or types (doc comments, naming patterns)
- Note any backward-compatibility shims or aliases
- Document the project's compatibility philosophy (semver, breaking changes in minors, etc.)
- Check `go.mod` module path for version information (v2+, or pre-v1)
- Identify any feature flags or conditional compilation patterns

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | Compatibility posture documented. Deprecations identified. Stability signals noted. Module versioning analyzed. An agent could assess the impact of API changes. |
| B | 80-89% completeness | Major compatibility signals documented. Deprecations noted. Minor gaps in feature flag or conditional compilation details. |
| C | 70-79% completeness | Some compatibility information present but not systematic. |
| D | 60-69% completeness | Compatibility mentioned vaguely without specific signals. |
| F | < 60% completeness | Backward compatibility not assessed. |

---

### Criterion 5: Documentation Coverage (weight: 10%)

**What to observe**: How well the API surface is self-documenting — doc comments, examples, and usage patterns in code. The knowledge reference must assess documentation quality.

**Evidence to collect**:
- Check for godoc-style comments on exported types, functions, and interfaces
- Note which packages have comprehensive doc comments vs bare exports
- Identify any example functions (`Example_*` in test files)
- Check for package-level documentation (doc.go files or package comments)
- Document any external API documentation (README sections, generated docs)

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | Documentation coverage assessed per package. Doc comment quality noted. Example functions identified. Package docs present/absent documented. An agent knows which APIs are well-documented vs need investigation. |
| B | 80-89% completeness | Major packages assessed for doc quality. Minor gaps in example or package doc assessment. |
| C | 70-79% completeness | Some documentation assessment present but not per-package. |
| D | 60-69% completeness | Documentation coverage mentioned without specific assessment. |
| F | < 60% completeness | Documentation coverage not assessed. |

---

## Grading Calculation

Final grade is weighted average of all criteria midpoint scores (see `schemas/grading.lego.md`). Example:
- CLI Command Contracts: A (midpoint 95%) x 30% = 28.5
- Exported Interfaces: B (midpoint 85%) x 25% = 21.25
- Public Type Signatures: B (midpoint 85%) x 20% = 17.0
- Backward Compatibility Signals: C (midpoint 75%) x 15% = 11.25
- Documentation Coverage: B (midpoint 85%) x 10% = 8.5
- **Total: 86.5 -> B**

## Related

- [Pinakes INDEX](../INDEX.lego.md) -- Full audit system documentation
- [architecture-criteria](architecture.lego.md) -- Codebase architecture knowledge capture
- [conventions-criteria](conventions.lego.md) -- Codebase conventions knowledge capture
- [dependencies-criteria](dependencies.lego.md) -- Dependency landscape knowledge capture

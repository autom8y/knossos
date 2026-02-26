---
name: conventions-criteria
description: "Observation criteria for codebase conventions knowledge capture. Use when: theoros is producing conventions knowledge for .know/, documenting naming patterns, error handling style, and coding standards. Triggers: conventions knowledge criteria, codebase conventions observation, naming patterns documentation, error handling conventions, file organization patterns."
---

# Conventions Observation Criteria

> The theoros observes and documents codebase conventions -- producing a knowledge reference that enables any CC agent to write code that looks native to the project.

## Scope

**Target files**: `./cmd/`, `./internal/`, root-level Go files (`*.go`, `go.mod`)

**Observation focus**: Naming patterns, error handling conventions, file organization rules, and code style conventions that a new CC agent needs to internalize before contributing code.

**NOTE**: This domain uses knowledge-capture grading. Instead of grading compliance ratios ("90% of functions follow X convention"), grade the COMPLETENESS of the conventions reference produced. A = comprehensive documentation of conventions with evidence. F = incomplete or inaccurate documentation.

## Criteria

### Criterion 1: Naming Patterns (weight: 30%)

**What to observe**: Variable naming, function naming, type naming, package naming, file naming conventions. The knowledge reference must tell an agent exactly how names are formed in this codebase.

**Evidence to collect**:
- Scan exported type names across packages for patterns (e.g., `New*` constructors, `*Options` config structs, `*Result` return types)
- Note variable naming conventions (e.g., short receiver names, `err` for errors, `ctx` for context)
- Identify file naming patterns (e.g., `_test.go` suffixes, matching package name files, test helper files)
- Check for acronym conventions (e.g., `ID` vs `Id`, `URL` vs `Url`, `HTTP` vs `Http`)
- Document package naming patterns (singular vs plural, verb vs noun)

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | All naming dimensions documented (types, functions, variables, files, packages) with specific examples from the codebase. Acronym conventions noted. Patterns documented with 3+ examples each. An agent could name new entities correctly. |
| B | 80-89% completeness | Most naming dimensions covered. Good examples for major patterns. Minor gaps in specific dimensions. |
| C | 70-79% completeness | Key patterns documented but some dimensions missing. Limited examples. |
| D | 60-69% completeness | Some patterns listed without examples or with only 1 example each. Multiple dimensions missing. |
| F | < 60% completeness | Naming conventions not systematically documented or descriptions are vague. |

---

### Criterion 2: Error Handling Style (weight: 30%)

**What to observe**: How errors are created, wrapped, propagated, and handled throughout the codebase. The knowledge reference must tell an agent the project's error handling philosophy and patterns.

**Evidence to collect**:
- Identify the error creation pattern (custom error types, sentinel errors, `errors.New`, `fmt.Errorf`)
- Check for error wrapping conventions (`%w`, custom wrap functions, error context enrichment)
- Document error propagation style (immediate return, defer cleanup, error aggregation)
- Note error handling at boundaries (CLI output, logging, user-facing messages)
- Identify any error code systems or categorization patterns
- Check for error checking conventions (inline `if err != nil`, helper functions, must-style panics)

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | Error creation, wrapping, propagation, and boundary handling all documented with specific patterns and file references. Custom error infrastructure described. An agent could handle errors consistently. |
| B | 80-89% completeness | Major error patterns documented. Wrapping and propagation covered. Minor gaps in boundary handling. |
| C | 70-79% completeness | Basic error patterns noted but wrapping or propagation conventions incomplete. |
| D | 60-69% completeness | Error handling mentioned superficially without tracing patterns through the codebase. |
| F | < 60% completeness | Error handling conventions not documented or inaccurate. |

---

> **Test conventions**: See test-coverage domain (`.know/test-coverage.md`). Test patterns are documented there to avoid duplication across `.know/` files.

### Criterion 3: File Organization (weight: 25%)

**What to observe**: How code is organized within packages: file boundaries, what goes where, separation patterns. The knowledge reference must tell an agent where to put new code.

**Evidence to collect**:
- Identify per-package file organization patterns (one type per file, grouped by concern, sorted alphabetically)
- Document the relationship between file names and their contents (e.g., `types.go` for types, `helpers.go` for utilities)
- Note where constants, variables, and init functions live
- Check for internal package usage patterns (`internal/` boundaries, what gets exported)
- Document the cmd/ vs internal/ separation philosophy
- Identify any generated code patterns or file conventions

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | File organization patterns documented across multiple packages with evidence. File naming conventions, content grouping rules, and internal/ boundaries all covered. An agent could place new files correctly. |
| B | 80-89% completeness | Major organization patterns documented. Good coverage of file naming and content grouping. Minor gaps in internal/ boundary documentation. |
| C | 70-79% completeness | General organization described but specific patterns not evidenced across packages. |
| D | 60-69% completeness | File organization mentioned vaguely without specific patterns or examples. |
| F | < 60% completeness | Organization conventions not documented or inaccurate. |

---

### Criterion 4: Code Style Conventions (weight: 15%)

**What to observe**: Formatting, comments, import organization, and other style conventions beyond naming and structure. The knowledge reference must capture the project's style DNA.

**Evidence to collect**:
- Document import grouping conventions (stdlib, external, internal — group ordering and blank lines)
- Note comment style (when comments are used, doc comment patterns, inline comment conventions)
- Check for builder/option patterns (functional options, builder structs)
- Identify constant/enum patterns (iota usage, string constants, typed vs untyped)
- Note any linting or formatting tool configurations (`.golangci.yml`, `.editorconfig`)

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | Import grouping, comment style, builder patterns, constant patterns, and tool configurations all documented. An agent could write style-consistent code. |
| B | 80-89% completeness | Major style conventions covered. Minor gaps in builder or constant patterns. |
| C | 70-79% completeness | Basic style conventions noted but not all dimensions covered. |
| D | 60-69% completeness | Style conventions mentioned superficially. |
| F | < 60% completeness | Code style conventions not documented. |

---

## Grading Calculation

Final grade is weighted average of all criteria midpoint scores (see `schemas/grading.lego.md`). Example:
- Naming Patterns: A (midpoint 95%) x 30% = 28.5
- Error Handling Style: B (midpoint 85%) x 30% = 25.5
- File Organization: B (midpoint 85%) x 25% = 21.25
- Code Style Conventions: B (midpoint 85%) x 15% = 12.75
- **Total: 88.0 -> B**

## Related

- [Pinakes INDEX](../INDEX.lego.md) -- Full audit system documentation
- [architecture-criteria](architecture.lego.md) -- Codebase architecture knowledge capture
- [test-coverage-criteria](test-coverage.lego.md) -- Test patterns and coverage conventions (owns the test domain)

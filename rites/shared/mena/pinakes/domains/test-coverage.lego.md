---
name: test-coverage-criteria
description: "Observation criteria for codebase test structure knowledge capture. Use when: theoros is producing test-coverage knowledge for .know/, documenting test organization, coverage patterns, and testing infrastructure. Triggers: test coverage knowledge criteria, test structure observation, testing patterns documentation."
---

# Test Coverage Observation Criteria

> The theoros observes and documents codebase test structure -- producing a knowledge reference that enables any CC agent to understand what is tested, how tests are organized, and where coverage gaps exist.

## Scope

**Target files**: `*_test.go` files across `./cmd/` and `./internal/`, `testdata/` directories, test helper files

**Observation focus**: Test file organization, coverage distribution, testing patterns and infrastructure, and coverage gaps that a CC agent needs to understand before writing or modifying tests.

**NOTE**: This domain uses knowledge-capture grading. Instead of grading compliance ratios ("90% of packages have tests"), grade the COMPLETENESS of the test structure reference produced. A = comprehensive documentation of test structure with evidence. F = incomplete or inaccurate documentation.

## Criteria

### Criterion 1: Test Structure (weight: 30%)

**What to observe**: How tests are organized across the codebase — which packages have tests, test file distribution, test-to-source ratios. The knowledge reference must give a reader a map of where tests live.

**Evidence to collect**:
- List all packages with `_test.go` files and count test files per package
- Identify packages WITHOUT test files (potential coverage gaps)
- Calculate test-to-source file ratios per package
- Note test package naming (`package foo` vs `package foo_test` — internal vs external tests)
- Identify integration test files vs unit test files (by naming, build tags, or test setup patterns)
- Document any test binary or TestMain patterns

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | Every package documented as tested/untested. Test file counts and ratios present. Internal vs external test packages distinguished. Integration tests identified. Coverage gaps explicitly listed. An agent knows exactly where to add tests. |
| B | 80-89% completeness | Most packages assessed. Test distribution documented. Coverage gaps noted. Minor gaps in integration test identification. |
| C | 70-79% completeness | Tested packages listed but untested gaps not systematically identified. |
| D | 60-69% completeness | Some test structure information without systematic assessment. |
| F | < 60% completeness | Test structure not mapped or significantly incomplete. |

---

### Criterion 2: Coverage Patterns (weight: 25%)

**What to observe**: What kinds of code are well-tested vs under-tested. The knowledge reference must tell an agent the project's coverage strengths and blind spots.

**Evidence to collect**:
- Identify the most heavily tested packages (by test count and test complexity)
- Note which code paths are tested: happy paths, error paths, edge cases, boundary conditions
- Document coverage for critical paths: CLI command handlers, sync pipeline, hook handlers
- Check for negative tests (testing that errors are returned, that invalid input is rejected)
- Identify any coverage measurement infrastructure (go test -cover, CI coverage reports)

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | Coverage patterns documented per package area. Critical paths assessed. Strengths and blind spots explicitly listed. Coverage infrastructure documented. An agent could prioritize where new tests are most needed. |
| B | 80-89% completeness | Major coverage patterns documented. Critical paths assessed. Minor gaps in blind spot identification. |
| C | 70-79% completeness | General coverage information present but not per-package or per-area. |
| D | 60-69% completeness | Coverage mentioned vaguely ("some packages have tests") without analysis. |
| F | < 60% completeness | Coverage patterns not assessed. |

---

### Criterion 3: Testing Conventions (weight: 20%)

**What to observe**: How tests are written — naming, structure, assertion patterns, and idioms used consistently across the codebase. The knowledge reference must tell an agent how to write a test that fits.

**Evidence to collect**:
- Document test function naming conventions (`TestFoo`, `TestFoo_Bar`, subtest naming)
- Identify subtest patterns (`t.Run` usage, naming conventions within subtests)
- Note assertion patterns (stdlib comparisons, `t.Errorf`, helper packages)
- Document test helper patterns (`t.Helper()` usage, shared setup functions)
- Check for test skip patterns (`t.Skip`, build tags, environment checks)
- Identify any testing framework usage beyond stdlib

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | All testing conventions documented with examples from multiple packages. Naming, assertions, helpers, and skip patterns covered. An agent could write tests matching project conventions. |
| B | 80-89% completeness | Major conventions documented. Good examples. Minor gaps in skip or helper patterns. |
| C | 70-79% completeness | Core conventions noted but examples limited to one or two packages. |
| D | 60-69% completeness | Some conventions listed without examples. |
| F | < 60% completeness | Testing conventions not documented. |

---

### Criterion 4: Fixture Patterns (weight: 15%)

**What to observe**: How test data is managed — fixture files, test helpers, factory functions, and test environment setup. The knowledge reference must tell an agent how to set up test scenarios.

**Evidence to collect**:
- Identify `testdata/` directories and their contents
- Document test fixture patterns (golden files, JSON fixtures, YAML test configs)
- Note test helper functions for creating test scenarios (temp dirs, test projects, mock data)
- Check for test environment management (env vars, temp file cleanup, `t.TempDir()`)
- Identify builder/factory patterns for test data construction
- Document any shared test utilities (test packages, test helpers imported across packages)

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | All fixture patterns documented: testdata dirs, golden files, helpers, builders, environment management. An agent could set up test scenarios following project patterns. |
| B | 80-89% completeness | Major fixture patterns documented. Good coverage of testdata and helpers. Minor gaps in builder or environment patterns. |
| C | 70-79% completeness | Some fixture information present but not comprehensive across packages. |
| D | 60-69% completeness | Fixture patterns mentioned without examples or detail. |
| F | < 60% completeness | Fixture patterns not documented. |

---

### Criterion 5: Test Infrastructure (weight: 10%)

**What to observe**: How tests are run, configured, and managed in CI and development workflows. The knowledge reference must tell an agent how to run tests correctly.

**Evidence to collect**:
- Document the test command (flags, environment variables, e.g., `CGO_ENABLED=0 go test ./...`)
- Note any build constraints or tags required for tests
- Identify test timeouts and parallel settings
- Check for CI test configuration (GitHub Actions, test scripts, coverage reporting)
- Document test dependencies (external services, databases, network requirements)
- Note any test caching or speed optimization patterns

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | Test running infrastructure fully documented. Commands, flags, CI config, and dependencies covered. An agent could run all tests correctly on first try. |
| B | 80-89% completeness | Test commands and CI config documented. Minor gaps in caching or optimization details. |
| C | 70-79% completeness | Basic test command documented but CI or dependency details missing. |
| D | 60-69% completeness | Test running mentioned without specific commands or configuration. |
| F | < 60% completeness | Test infrastructure not documented. |

---

## Grading Calculation

Final grade is weighted average of all criteria midpoint scores (see `schemas/grading.lego.md`). Example:
- Test Structure: A (midpoint 95%) x 30% = 28.5
- Coverage Patterns: B (midpoint 85%) x 25% = 21.25
- Testing Conventions: A (midpoint 95%) x 20% = 19.0
- Fixture Patterns: B (midpoint 85%) x 15% = 12.75
- Test Infrastructure: A (midpoint 95%) x 10% = 9.5
- **Total: 91.0 -> A**

## Related

- [Pinakes INDEX](../INDEX.lego.md) -- Full audit system documentation
- [architecture-criteria](architecture.lego.md) -- Codebase architecture knowledge capture
- [conventions-criteria](conventions.lego.md) -- Codebase conventions knowledge capture

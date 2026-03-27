---
domain: test-coverage
generated_at: "2026-03-27T09:41:06Z"
expires_after: "7d"
source_scope:
  - "./src/**/*"
generator: theoros
source_hash: "094fa67"
confidence: 1.0
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Test Coverage

> Knowledge reference for test coverage in `autom8y-workflows`. Generated 2026-03-27.

---

## Coverage Gaps

**Nature of the repository**: `autom8y-workflows` is a **GitHub Actions reusable workflow repository**. It contains no source code, no Python packages, no Go modules, no test files, and no test infrastructure of its own. The sole non-dotfile artifact is:

- `.github/workflows/satellite-ci-reusable.yml` — a reusable CI workflow for satellite repositories

**Language detection result**: No language manifests found. No `pyproject.toml`, `go.mod`, `package.json`, `Cargo.toml`, or `pom.xml` exist at any path. The repo has no testable source code.

**Test files found**: Zero. No `test_*.py`, `*_test.py`, `*_test.go`, `*.test.ts`, or `*.spec.ts` files exist.

**Coverage measurement infrastructure**: None in this repository. The workflow file at `.github/workflows/satellite-ci-reusable.yml` configures coverage tooling (`pytest --cov`, `codecov/codecov-action`) for satellite repos that call it, but this repo itself has no coverage configuration.

**Critical path coverage assessment**:

There are no critical code paths to test in this repository. The repository's functional surface is entirely declarative YAML. The one "artifact" — the reusable workflow — is tested only by being exercised when satellite repositories run CI against it. There is no unit or integration test framework that validates the workflow's correctness in isolation.

**Prioritized gap list**:

1. **No workflow validation tests** (HIGH risk): The reusable workflow has no automated validation. Changes to `.github/workflows/satellite-ci-reusable.yml` can silently break all satellite CI pipelines. There is no `act` (local GitHub Actions runner) integration, no schema validation, and no lint check for the workflow YAML.
2. **No input contract tests**: The 15 configurable `workflow_call` inputs (e.g., `mypy_targets`, `coverage_package`, `coverage_threshold`, `test_markers_exclude`) have no validation tests. Invalid input combinations (e.g., `test_parallel: true` with no `pytest-xdist` installed) would only be caught at runtime in satellite repos.
3. **No span collector plugin tests**: The inline `_conftest_convention_ci.py` plugin (written dynamically at runtime in the `convention-check` job, lines 375-416) is untested. Its logic patches `InMemorySpanExporter.__init__` and accumulates spans — a non-trivial behavior with no standalone test.
4. **No integration test for the full pipeline**: The workflow's four jobs (`lint`, `test`, `integration`, `convention-check`) are never exercised against a real satellite in this repository.

---

## Testing Conventions

**No testing conventions exist** in this repository because there are no tests and no source code. This section documents what an agent would find if conventions were ever established.

**What the workflow imposes on satellite repos** (not conventions of this repo, but the framework it defines for satellites):

- **Test runner**: `pytest` via `uv run --no-sources pytest tests/`
- **Coverage target**: Passed via `coverage_package` input; run with `--cov={coverage_package} --cov-report=xml --cov-fail-under={coverage_threshold}` (default threshold: 80%)
- **Marker convention for integration tests**: Satellites use `@pytest.mark.integration`. The workflow excludes integration tests by default (`test_markers_exclude: 'not integration'`) and runs them only when `run_integration: true`.
- **Marker convention for instrumentation tests**: Satellites filter with `-k 'instrumentation or telemetry or otel'` for span collection in the `convention-check` job.
- **Fixture pattern**: No `testdata/` or `__fixtures__` directories exist here. The workflow dynamically writes `_conftest_convention_ci.py` at runtime (`.github/workflows/satellite-ci-reusable.yml` lines 375-416) rather than having it as a checked-in file.
- **Test directory**: Satellites are expected to have their tests in `tests/` (hardcoded at lines 222, 316, 422).
- **Coverage upload**: Codecov via `codecov/codecov-action` with `fail_ci_if_error: false` (line 249).
- **Parallel testing**: Optional; controlled by `test_parallel: boolean` input (default: false). Uses `pytest-xdist` `-n auto` when enabled.

**If tests were added to this repo**, they would logically follow:
- Python test conventions (since the only inline code in the repo is Python)
- `pytest` as the runner (consistent with the satellite convention this workflow enforces)
- `tests/` directory layout

---

## Test Structure Summary

**Overall distribution**: 0 of 1 artifacts (the workflow file) have any associated tests. The test ratio is 0%.

**Most heavily tested areas**: None. No tests exist.

**Package naming patterns**: Not applicable. No Python packages exist.

**Integration vs unit test distinction**: Not applicable. The workflow defines this distinction for satellites (`@pytest.mark.integration`, `run_integration` boolean input) but does not implement it here.

**Test invocation command**: No test command exists for this repository. If tests were added using Python/pytest (the language inferred from the inline Python in the workflow), the command would be:

```
uv run pytest tests/
```

consistent with the satellite convention the workflow enforces.

**What a complete mental model of this repository looks like**:

```
autom8y-workflows/
├── README.md                          # Usage documentation
└── .github/workflows/
    └── satellite-ci-reusable.yml      # The only artifact (1 file, ~457 lines)
        ├── jobs/lint                  # ruff format, ruff check, mypy
        ├── jobs/test                  # pytest + coverage + codecov upload
        ├── jobs/integration           # pytest -m integration (conditional)
        └── jobs/convention-check      # span collection + convention-check CLI (conditional)
```

The repo is a delivery vehicle for CI configuration, not a software project with testable business logic. The appropriate "test coverage" question here is: **does the workflow file have validation tooling?** The answer is no.

---

## Knowledge Gaps

1. **Runtime behavior of the workflow**: Cannot verify whether the workflow actually succeeds or fails in real satellite executions without access to satellite repository CI history.
2. **Action version pin correctness**: Pinned SHAs (e.g., `actions/checkout@34e114876b...`) cannot be validated for current correctness from this observation.
3. **Satellite repos using this workflow**: The knowledge reference cannot enumerate which satellite repos call this workflow, limiting risk assessment for any change to it.
4. **Whether `act` or similar tooling is planned**: No evidence of workflow validation tooling in any configuration or documentation.

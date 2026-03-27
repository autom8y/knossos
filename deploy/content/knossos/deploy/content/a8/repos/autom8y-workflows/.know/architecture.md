---
domain: architecture
generated_at: "2026-03-27T09:41:06Z"
expires_after: "7d"
source_scope:
  - "./src/**/*"
generator: theoros
source_hash: "094fa67"
confidence: 0.95
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Architecture

## Package Structure

**Language detection result**: This repository contains no application source code. There is no `go.mod`, `package.json`, `pyproject.toml`, `Cargo.toml`, or `pom.xml` at the root. The repository's sole deliverable is a GitHub Actions reusable workflow.

**Repository identity**: `autom8y-workflows` is a workflow library repo — purpose-built to host reusable GitHub Actions YAML consumed by satellite repositories across the autom8y organization. It must remain public because GitHub enforces that public repositories cannot call reusable workflows from private repositories.

**Directory layout (complete)**:

```
autom8y-workflows/
├── .claude/               # CC platform infrastructure (knossos)
│   ├── agents/            # 5 rite agents (potnia, architect, requirements-analyst, principal-engineer, qa-adversary)
│   ├── agent-memory/      # Persistent agent memory (cartographer, dependency-resolver, release-executor, release-planner)
│   ├── commands/          # Slash commands (dromena)
│   ├── rules/             # CC rules
│   ├── skills/            # Skill library (legomena) including pinakes, complaint-filing, etc.
│   ├── CLAUDE.md          # Project inscription (knossos-managed)
│   └── settings.local.json
├── .gemini/               # Gemini CLI parallel infrastructure (mirrors .claude/)
├── .github/
│   └── workflows/
│       └── satellite-ci-reusable.yml   # THE PRIMARY DELIVERABLE
├── .git/
├── .knossos/
│   ├── ACTIVE_RITE        # "10x-dev"
│   ├── ACTIVE_WORKFLOW.yaml
│   ├── KNOSSOS_MANIFEST.yaml
│   ├── PROVENANCE_MANIFEST.yaml
│   ├── PROVENANCE_MANIFEST_GEMINI.yaml
│   ├── mcp-ownership.json
│   ├── procession-mena/   # security-remediation-ref mena
│   └── sync/state.json
├── .know/                 # Empty (this document will populate it)
├── .mcp.json              # GitHub MCP server config
├── .sos/
│   └── sessions/          # (empty)
└── README.md
```

**The single "package" in this repo**:

| Module | Path | Purpose | Type |
|--------|------|---------|------|
| `satellite-ci-reusable.yml` | `.github/workflows/satellite-ci-reusable.yml` | Reusable CI pipeline for all autom8y satellite repos | GitHub Actions reusable workflow |

There are no application packages, no source directories, no imports. The unit of composition is GitHub Actions jobs and steps, not code modules.

**Hub/leaf classification (GitHub Actions jobs within the workflow)**:

The workflow defines 4 jobs:

| Job | Key inputs consumed | Purpose |
|-----|---------------------|---------|
| `lint` | `mypy_targets`, `mypy_strict`, `mypy_advisory_targets`, `semgrep_config`, `test_extras` | Ruff format + lint, mypy type check, optional Semgrep scan |
| `test` | `coverage_package`, `coverage_threshold`, `test_markers_exclude`, `test_ignore`, `test_parallel`, `test_env`, `test_extras` | pytest unit tests with coverage gate |
| `integration` | `run_integration`, `integration_timeout`, `test_env`, `test_extras` | pytest integration tests (conditional on `run_integration: true`) |
| `convention-check` | `convention_check`, `convention_check_test_filter`, `convention_check_test_env` | OpenTelemetry span collection + convention-check CLI validation (conditional) |

Each job is independent (no `needs:` dependencies between them). They run in parallel.

---

## Layer Boundaries

Because this repo has no application source code, the layer model is expressed as GitHub Actions architecture, not package import graphs.

**Layer model**:

```
Satellite repo (caller)
    └── test.yml
            └── uses: autom8y/autom8y-workflows/.github/workflows/satellite-ci-reusable.yml@main
                    └── jobs: lint | test | integration | convention-check
                            └── shared infrastructure (AWS OIDC, CodeArtifact, uv)
```

**Separation of concerns**:

- **Caller layer** (satellite's `test.yml`): Passes per-repo configuration inputs (`mypy_targets`, `coverage_package`, etc.). Does not contain CI logic.
- **Workflow layer** (`satellite-ci-reusable.yml`): Contains all CI logic. Receives inputs and executes jobs.
- **Shared infrastructure layer** (embedded in each job): AWS OIDC authentication + CodeArtifact token acquisition steps are replicated identically in all 4 jobs. This is a structural pattern (not a separate module) imposed by GitHub Actions' job isolation model.

**Import direction analog**: All data flow is inputs-to-jobs (one direction). Callers pass typed inputs; the workflow does not call back to callers.

**Boundary enforcement patterns**:
- All actions are pinned to exact commit SHAs (e.g., `actions/checkout@34e114876b0b11c390a56381ad16ebd13914f8d5`) with version comments. This prevents supply chain drift.
- `UV_NO_SOURCES: "1"` is set at the job level in all 4 jobs, enforcing that uv uses the CodeArtifact index rather than PyPI sources.
- `secrets: inherit` pattern in callers passes org secrets without exposing them in workflow inputs.

**Knossos platform layer** (`.claude/` and `.knossos/`): Separate layer that provides development tooling (CC agents, skills, commands). This layer is managed by knossos and is not part of the CI delivery mechanism. It operates on the developer's machine, not in GitHub Actions.

---

## Entry Points and API Surface

### Workflow Entry Point

The workflow is triggered exclusively via `workflow_call` (reusable workflow event). It is never triggered directly by push/PR events — satellites own their own trigger logic.

**Invocation pattern** (from `README.md` and workflow header):

```yaml
jobs:
  ci:
    permissions:
      id-token: write    # Required: OIDC for CodeArtifact
      contents: read
    uses: autom8y/autom8y-workflows/.github/workflows/satellite-ci-reusable.yml@main
    with:
      mypy_targets: 'src/autom8_example'
      coverage_package: 'autom8_example'
    secrets: inherit
```

**Critical constraint**: The `permissions` block is REQUIRED in the calling job. Without it, GitHub Actions inherits the org default (read-only), which blocks OIDC token generation and causes `startup_failure` (0 jobs created).

### Input API Surface (complete)

**Required inputs** (callers must provide):

| Input | Type | Description |
|-------|------|-------------|
| `mypy_targets` | string | Space-separated mypy target paths (e.g., `"src/autom8_asana"`) |
| `coverage_package` | string | Package name for `--cov` (e.g., `"autom8_asana"`) |

**Optional inputs with defaults**:

| Input | Type | Default | Description |
|-------|------|---------|-------------|
| `python_version` | string | `'3.12'` | Python version |
| `coverage_threshold` | number | `80` | Minimum coverage percentage |
| `mypy_strict` | boolean | `true` | Enable `--strict` for mypy |
| `mypy_advisory_targets` | string | `''` | Advisory-only mypy targets (non-blocking) |
| `test_extras` | string | `'--all-extras'` | uv sync extras arguments |
| `test_markers_exclude` | string | `'not integration'` | pytest `-m` exclusion expression |
| `test_ignore` | string | `''` | pytest `--ignore` paths (space-separated) |
| `test_parallel` | boolean | `false` | Enable pytest-xdist `-n auto` |
| `test_env` | string | `''` | Extra env vars for test jobs (KEY=VALUE lines) |
| `run_integration` | boolean | `false` | Run integration tests |
| `integration_timeout` | number | `300` | Integration test timeout in seconds |
| `semgrep_config` | string | `''` | Semgrep config path (empty = skip) |
| `convention_check` | boolean | `false` | Run convention-check against OTel spans |
| `convention_check_test_filter` | string | `'instrumentation or telemetry or otel'` | pytest `-k` to select instrumentation tests |
| `convention_check_test_env` | string | `''` | Extra env vars for convention-check run |

### Secrets API Surface

All secrets passed via `secrets: inherit`. The workflow uses:
- `secrets.APP_ID` — GitHub App ID (integration job only)
- `secrets.APP_PRIVATE_KEY` — GitHub App private key (integration job only)

### AWS Infrastructure Dependencies

The workflow assumes:
- `vars.AWS_ACCOUNT_ID` — organization-level variable
- IAM role `github-actions-deploy` in that account (OIDC trust policy required)
- AWS CodeArtifact domain named `autom8y` in `us-east-1`

### Knossos Platform Entry Point

The development workflow entry point is the 10x-dev rite. Active workflow: `.knossos/ACTIVE_WORKFLOW.yaml`. Entry agent: `requirements-analyst` (default) or overridden by work type. Full pipeline: requirements-analyst → architect → principal-engineer → qa-adversary, with back-routes for validation failures.

---

## Key Abstractions

**Domain**: This is a GitHub Actions workflow repo, so abstractions are CI/CD patterns rather than code types.

### 1. Reusable Workflow Pattern

**File**: `.github/workflows/satellite-ci-reusable.yml`
**Trigger**: `workflow_call`
**Purpose**: DRY CI — one workflow file replaces ~200 lines of duplicated YAML in each satellite repo. Callers provide per-repo configuration via typed inputs; the workflow handles all CI logic.

### 2. OIDC-for-CodeArtifact Pattern

Repeated in all 4 jobs. The pattern:
1. `aws-actions/configure-aws-credentials` (OIDC, no static credentials)
2. `aws codeartifact get-authorization-token` → masked output token
3. Inject `UV_INDEX_AUTOM8Y_USERNAME=aws` and `UV_INDEX_AUTOM8Y_PASSWORD={token}` into job env
4. `uv sync --no-sources` uses those env vars to authenticate to CodeArtifact

**Why repeated**: GitHub Actions has no shared step mechanism across jobs — each job is isolated. The pattern is intentionally duplicated rather than using a composite action to keep the workflow self-contained.

### 3. Convention-Check Plugin Injection Pattern

**File**: `.github/workflows/satellite-ci-reusable.yml` lines 373–456
The `convention-check` job uses a novel pattern: it writes a pytest conftest plugin (`_conftest_convention_ci.py`) to disk at CI runtime, then runs pytest with `-p _conftest_convention_ci`. The plugin:
- Monkey-patches `InMemorySpanExporter.__init__` to track all exporters created during the test session
- Collects all finished spans at `pytest_sessionfinish`
- Serializes via `autom8y_telemetry.conventions.checker.spans_to_json` to `spans.json`

Then `convention-check --spans spans.json --format json` validates the spans, filtering out framework spans (those without dots in the span name) from the failure gate.

### 4. SHA-Pinned Actions

All external actions use commit SHA pinning with version comments:
- `actions/checkout@34e114876b0b11c390a56381ad16ebd13914f8d5 # v4`
- `astral-sh/setup-uv@38f3f104447c67c051c4a08e39b64a148898af3a # v4`
- `actions/setup-python@a26af69be951a213d495a4c3e4e4022e16d87065 # v5`
- `aws-actions/configure-aws-credentials@7474bc4690e29a8392af63c5b98e7449536d5c3a # v4`
- `codecov/codecov-action@671740ac38dd9b0130fbe1cec585b89eea48d3de # v5`
- `actions/create-github-app-token@d72941d797fd3113feb6b93fd0dec494b13a2547 # v1`

Pattern: supply chain security through immutable SHAs, human-readable version in comment.

### 5. GitHub App Token for Integration Tests

Integration tests require checked-out code with cross-repo access. The integration job generates a GitHub App token (`actions/create-github-app-token`) using `APP_ID`/`APP_PRIVATE_KEY` secrets, then passes that token to `actions/checkout`. This distinguishes the integration job from the lint/test jobs that use the default GITHUB_TOKEN.

### 6. Knossos Rite System

**Files**: `.knossos/ACTIVE_WORKFLOW.yaml`, `.knossos/KNOSSOS_MANIFEST.yaml`, `.knossos/PROVENANCE_MANIFEST.yaml`
The active development rite is `10x-dev`. The knossos platform manages CC agent compositions via sync from a central knossos registry at `../../../../Code/knossos/`. The PROVENANCE_MANIFEST.yaml tracks checksums and source paths for every knossos-managed file. Agents, commands, and skills are materialized into `.claude/` and `.gemini/` from knossos rite definitions.

---

## Data Flow

### Primary Data Flow: Satellite CI Pipeline

```
Satellite push to main
    -> satellite's test.yml (workflow_dispatch trigger)
        -> calls autom8y-workflows/satellite-ci-reusable.yml@main
            -> inputs: {mypy_targets, coverage_package, ...}
            -> jobs run in parallel:
                -> lint job:
                    OIDC -> CodeArtifact token -> uv sync -> ruff format/check -> mypy
                -> test job:
                    OIDC -> CodeArtifact token -> uv sync -> pytest (unit, with coverage gate)
                -> integration job (if run_integration=true):
                    GitHub App token -> checkout -> OIDC -> CodeArtifact -> pytest (integration)
                -> convention-check job (if convention_check=true):
                    OIDC -> CodeArtifact -> uv sync -> inject conftest plugin -> pytest (collect spans)
                    -> spans.json -> convention-check CLI -> domain violation count
                    -> fail if domain violations > 0
        -> on success: satellite's test.yml proceeds to satellite-dispatch.yml
            -> repository_dispatch to autom8y/autom8y
                -> satellite-receiver.yml -> service-build.yml -> service-deploy.yml (ECS)
```

### Credential Flow

```
GitHub Actions runner (OIDC)
    -> AWS STS (assume role: github-actions-deploy)
        -> AWS CodeArtifact (get-authorization-token, domain: autom8y, us-east-1)
            -> token masked in GITHUB_OUTPUT
                -> UV_INDEX_AUTOM8Y_USERNAME=aws
                -> UV_INDEX_AUTOM8Y_PASSWORD={token}
                    -> uv sync resolves packages from CodeArtifact index
```

### Convention Check Data Flow

```
instrumentation tests (InMemorySpanExporter)
    -> _conftest_convention_ci.py plugin (monkey-patches __init__, collects at session end)
        -> spans_to_json(all_spans)
            -> spans.json on disk
                -> convention-check --spans spans.json --format json
                    -> check_result.json
                        -> Python filter: violations where '.' in span_name
                            -> DOMAIN_VIOLATIONS count
                                -> fail if > 0
```

### Knossos Sync Data Flow

```
knossos registry (../../../../Code/knossos/)
    -> ari sync (knossos CLI)
        -> reads KNOSSOS_MANIFEST.yaml regions
            -> materializes: agents/*.md, commands/*, skills/*
                -> writes to .claude/ and .gemini/
                    -> PROVENANCE_MANIFEST.yaml updated with checksums
                    -> CLAUDE.md regenerated from template
```

---

## Knowledge Gaps

1. **Downstream pipeline files not in scope**: The full satellite deploy pipeline (satellite-dispatch.yml, satellite-receiver.yml, service-build.yml, service-deploy.yml) lives in the `autom8y` monorepo, not in this repository. This document cannot trace beyond the `satellite-ci-reusable.yml` output.

2. **Criteria/language mismatch**: The architecture observation criteria were written for Go/Python/TypeScript application projects with `cmd/`/`internal/` package structures. This repository has no application source code. The criteria were adapted to the actual artifact (GitHub Actions workflow YAML), but the grading rubric does not perfectly map. Evidence was collected from the workflow file's job/step/input structure as the analog to package/module/interface structure.

3. **Convention-check CLI**: The `convention-check` command is invoked from the `autom8y_telemetry` package (a satellite dependency). Its internals and violation schema are defined in that package, which is not in this repository.

4. **uv.lock not present**: The workflow references `cache-dependency-glob: "uv.lock"` but the lock file itself lives in each satellite repo, not here. The workflow is parameterized to use the caller's lock file.

5. **Semgrep config paths**: The `semgrep_config` input accepts a path but the actual semgrep config files live in satellite repos. No semgrep configs exist in this repository.

---
domain: conventions
generated_at: "2026-03-27T09:41:06Z"
expires_after: "7d"
source_scope:
  - "./src/**/*"
generator: theoros
source_hash: "094fa67"
confidence: 0.92
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Conventions

> This reference documents the conventions of the `autom8y-workflows` repository, which hosts a single reusable GitHub Actions workflow. The "language" is GitHub Actions YAML with embedded shell (bash) and Python (inline via heredoc). These conventions apply to this repository and to satellite repositories that consume it.

## Error Handling Style

**Scope adaptation**: This is a GitHub Actions YAML/shell repository, not a traditional language project. "Error handling" means: how CI failures propagate, how shell errors are surfaced or suppressed, how optional steps are guarded, and how critical errors are reported to GitHub Actions UI.

### Shell Error Propagation

Shell `run` blocks do NOT use `set -e` or `set -euo pipefail`. Individual commands are written to be explicit about whether they should block CI or not.

**Suppressing non-blocking failures**: The `|| true` pattern is used when a command's failure should not halt the step:

- `.github/workflows/satellite-ci-reusable.yml` line 317: `continue-on-error: true` on the integration test run step (job-level soft failure)
- `.github/workflows/satellite-ci-reusable.yml` lines 428, 441: `|| true` appended to pytest and convention-check invocations to allow the step to continue while collecting output

**Blocking failures with explicit exit**: After a suppressed command, the script checks for failure conditions and calls `exit 1` explicitly:

```yaml
# lines 429-431
if [ ! -f spans.json ]; then
  echo "::error::No spans.json produced -- convention collector plugin failed"
  exit 1
fi
```

**Python-in-shell error handling**: Inline Python fragments (lines 442-449) are written as single expressions without try/except -- they are allowed to raise and will fail the step naturally. This is intentional: Python is used only for data extraction where failure is always a real error.

### GitHub Actions Error Annotations

The project uses GitHub Actions problem matchers via `echo "::error::..."` and `echo "::warning::..."` for surfacing errors in the UI (lines 430, 436, 451-453). The pattern is:

- `::error::` for blocking conditions (followed by `exit 1`)
- `::warning::` for advisory conditions (no exit, just notification)

### Secrets Masking

Tokens that must not appear in logs are always masked immediately after generation:

- `.github/workflows/satellite-ci-reusable.yml` line 138: `echo "::add-mask::$TOKEN"` appears on the line immediately after `TOKEN=$(...)` assignment, before any use of `$TOKEN`

### Step-Level vs Job-Level Failure

- `continue-on-error: true` at the job level: used for integration tests (line 317), signaling that integration test failures are informational/non-blocking
- No `continue-on-error` on lint, test, or convention-check jobs: these are hard gates

### Python Inline Code Error Handling

The `convention-check` job embeds a full pytest plugin via heredoc (lines 375-416). Inside that Python code:

- `try/except Exception: pass` (line 408-409) used to silently absorb per-exporter read errors when collecting spans -- guarantees the plugin always writes a `spans.json`, even if some exporters fail
- Empty spans file written as fallback (lines 413-415) rather than failing -- the downstream convention-check step handles the empty case

---

## File Organization

**Scope adaptation**: This repository has a single-file structure. File organization conventions describe the layout of that file and how jobs/steps/inputs are organized within it.

### Repository Layout

```
autom8y-workflows/
  .github/
    workflows/
      satellite-ci-reusable.yml    # The sole primary artifact
  README.md                        # Usage documentation
```

There is one workflow file per concern. The repository is organized around the principle of one reusable workflow per repository purpose.

### Workflow File Internal Organization

The workflow file follows a top-to-bottom structure:

1. **Header comment block** (lines 1-21): ASCII divider (`# ====...====`), title, description, usage example, and critical notes. This is mandatory for workflow files.
2. **`on:` trigger declaration** (lines 25-98): Only `workflow_call` used. All inputs documented with:
   - `description` (required on all inputs)
   - `required: true/false`
   - `type`
   - `default` (for optional inputs)
   - Grouping comments separating required from optional (lines 28-29, 36-37)
3. **`jobs:` block** (lines 100+): Jobs ordered by pipeline phase: lint -> test -> integration -> convention-check

### Input Declaration Pattern

Inputs are grouped into two commented sections within `workflow_call.inputs`:

```yaml
# --- Required: per-satellite configuration ---
mypy_targets: ...
coverage_package: ...

# --- Optional: override defaults ---
python_version: ...
```

Section separators use the `# --- Label ---` format (triple dash, sentence case, triple dash).

### Job Structure Pattern

Each job follows a consistent internal ordering:

1. `name:` (human-readable display name)
2. `if:` (conditional guard, when applicable)
3. `runs-on:`
4. `timeout-minutes:`
5. `permissions:` (always explicit, never implicit)
6. `env:` (job-level env, always includes `UV_NO_SOURCES: "1"`)
7. `steps:` (see step ordering below)

### Step Ordering Pattern

Steps within each job follow a consistent bootstrap sequence:

1. GitHub App token generation (integration job only, required for checkout token)
2. `actions/checkout` (always pinned to SHA)
3. `astral-sh/setup-uv` (always pinned to SHA, always version "0.9.7")
4. `actions/setup-python` (always pinned to SHA)
5. `aws-actions/configure-aws-credentials` (OIDC setup)
6. `Get CodeArtifact auth token` (named step, always `id: codeartifact`)
7. `Configure uv for CodeArtifact` (sets env vars for private registry)
8. Optional: `Set test environment variables`
9. `Install dependencies`
10. Job-specific work steps

---

## Domain-Specific Idioms

### Action Pinning to Commit SHA

Every `uses:` reference is pinned to a full commit SHA with a version comment, never to a mutable tag:

```yaml
uses: actions/checkout@34e114876b0b11c390a56381ad16ebd13914f8d5 # v4
uses: astral-sh/setup-uv@38f3f104447c67c051c4a08e39b64a148898af3a # v4
uses: actions/setup-python@a26af69be951a213d495a4c3e4e4022e16d87065 # v5
uses: aws-actions/configure-aws-credentials@7474bc4690e29a8392af63c5b98e7449536d5c3a # v4
uses: actions/create-github-app-token@d72941d797fd3113feb6b93fd0dec494b13a2547 # v1
uses: codecov/codecov-action@671740ac38dd9b0130fbe1cec585b89eea48d3de # v5
```

The comment format is `# vN` where N is the major version. This is mandatory -- do not add new action references without pinning.

### UV_NO_SOURCES as Universal Env

Every job sets `UV_NO_SOURCES: "1"` as a job-level env var. This is a project-wide idiom that prevents uv from using local source overrides in CI. It is repeated in every job rather than set globally to keep each job self-contained.

### `--no-sources` Flag Pattern

All `uv run` and `uv sync` commands append `--no-sources` explicitly:

```yaml
run: uv sync --no-sources ${{ inputs.test_extras }}
run: uv run --no-sources ruff format . --check
run: uv run --no-sources pytest ...
```

This mirrors `UV_NO_SOURCES: "1"` as a belt-and-suspenders pattern -- the env var and the flag are used together. Do not write `uv run` without `--no-sources`.

### OIDC Authentication Bootstrap Sequence

The OIDC bootstrap is a fixed 3-step sequence that appears identically in every job:

1. `Configure AWS credentials (OIDC)` -- uses `aws-actions/configure-aws-credentials` with a fixed role ARN pattern: `arn:aws:iam::${{ vars.AWS_ACCOUNT_ID }}:role/github-actions-deploy` and `aws-region: us-east-1`
2. `Get CodeArtifact auth token` with `id: codeartifact` -- always named exactly this, always uses `echo "::add-mask::$TOKEN"` before outputting the token
3. `Configure uv for CodeArtifact` -- always sets `UV_INDEX_AUTOM8Y_USERNAME=aws` and `UV_INDEX_AUTOM8Y_PASSWORD=` from the codeartifact step output

These three steps are copied verbatim across all jobs. Do not alter the sequence or rename the steps/ids -- satellite repos may depend on stable step names.

### Permissions Are Always Explicit

Every job declares its own `permissions:` block, always at minimum:

```yaml
permissions:
  id-token: write
  contents: read
```

The header comment (lines 19-21) documents why `id-token: write` is required (OIDC token generation). Never rely on org-default permissions -- always declare explicitly.

### CI-Injected Python Plugin Pattern

The `convention-check` job uses a CI-injected pytest plugin written as a heredoc (`cat > file << 'CONFTEST_EOF'`). This plugin is never checked into the repository. The pattern:

- Plugin file name: `_conftest_convention_ci.py` (underscore prefix signals "CI-generated, not project code")
- Invoked via `-p _conftest_convention_ci` (plugin name without `.py`)
- The plugin monkeypatches `InMemorySpanExporter.__init__` to track all exporters created during the session

### Domain Span Filter: Dot-Notation Convention

When filtering violations in the `convention-check` job, only spans with a dot in the name are treated as domain spans subject to the failure gate:

```python
vs = [v for v in data.get('violations', [])
      if '.' in v.get('span_name', '')]
```

Framework spans (no dots) are excluded from the failure gate. This is a project-specific filtering convention, not a standard OTel pattern.

### `continue-on-error: true` Semantics

`continue-on-error: true` is used at the job level for integration tests to make them advisory (non-blocking). This is an intentional design choice: integration tests run but their failure does not block merges. New jobs should default to hard-fail unless explicitly designed as advisory.

---

## Naming Patterns

### Job ID Naming

Job IDs use kebab-case, lowercase:

- `lint` (short, single-word)
- `test` (short, single-word)
- `integration` (short, single-word)
- `convention-check` (kebab-case for multi-word)

Job IDs match their `name:` field semantically but the `name:` field uses title case:
- Job ID `lint` -> `name: Lint & Type Check`
- Job ID `test` -> `name: Test`
- Job ID `integration` -> `name: Integration Tests`
- Job ID `convention-check` -> `name: Convention Check`

### Step ID Naming

Step IDs (`id:`) use kebab-case. Only steps whose outputs are referenced downstream receive explicit `id:` values:
- `id: codeartifact` (outputs token to `steps.codeartifact.outputs.token`)
- `id: pytest-args` (outputs pytest arg string to `steps.pytest-args.outputs.args`)
- `id: app-token` (outputs GitHub App token to `steps.app-token.outputs.token`)

Steps that are not referenced by later steps do not get an `id:`.

### Step Name Convention

Step `name:` values use sentence case (not title case) for multi-word names:
- "Configure AWS credentials (OIDC)" -- not "Configure AWS Credentials"
- "Get CodeArtifact auth token" -- not "Get Codeartifact Auth Token"
- "Set up Python" -- matches the action's canonical display name

Parenthetical context is used to distinguish variants: "(OIDC)" disambiguates the AWS credentials step from a hypothetical non-OIDC variant.

### Input Naming

Inputs use `snake_case` exclusively. The naming pattern is `noun_verb` or `noun_attribute`:

- `mypy_targets` (tool + target)
- `coverage_package` (metric + scope)
- `python_version` (runtime + attribute)
- `test_extras` (scope + modifier)
- `test_markers_exclude` (scope + field + direction)
- `test_parallel` (scope + mode)
- `run_integration` (action + scope) -- verb-first for boolean toggles
- `convention_check` (domain + action) -- boolean toggle
- `convention_check_test_filter` (domain + scope + field)

Boolean inputs that enable/disable features use `run_*` (for test runs) or the feature name as a bare noun (`convention_check`, `test_parallel`, `mypy_strict`).

### Shell Variable Naming

Shell variables in `run` blocks use `SCREAMING_SNAKE_CASE`:

- `TOKEN`, `ARGS`, `IGNORE`, `SPAN_COUNT`, `DOMAIN_VIOLATIONS`

Temporary variable names are short and descriptive. Multi-word shell vars follow `WORD_WORD` pattern.

### Workflow File Naming

Workflow files use kebab-case with a descriptive suffix indicating their type:
- `satellite-ci-reusable.yml` -- `{scope}-{purpose}-{type}.yml`

The `-reusable` suffix is a project convention signaling that this workflow is intended for `workflow_call`, not direct triggering.

---

## Knowledge Gaps

1. **No satellite repositories observed**: The conventions documented here are inferred entirely from the single workflow file and README. Satellite-side calling patterns (how satellites structure their calling workflow, what inputs they customarily provide) are not documented -- that would require reading satellite repositories outside this repo's scope.

2. **Single-file repository**: With only one primary artifact, some conventions (e.g., multi-file organization within `.github/workflows/`, how additional workflows would be structured) are inferred from patterns rather than observed across multiple examples.

3. **Implicit Python conventions**: The embedded Python snippets follow standard Python style but there is no `pyproject.toml`, `ruff.toml`, or explicit Python style configuration in this repository -- the Python conventions are inherited from satellite repos.

4. **No history of prior workflow files**: The git history may contain deleted or renamed files that would reveal additional organizational conventions, but git history was not examined as part of this observation.

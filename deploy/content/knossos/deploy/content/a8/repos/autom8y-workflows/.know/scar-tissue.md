---
domain: scar-tissue
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

# Codebase Scar Tissue

> Knowledge reference for CI failure history in autom8y-workflows. This repository is a single-file GitHub Actions reusable workflow codebase (Python/YAML CI tooling). All scars are CI pipeline failures discovered during satellite rollout (March 2026). There are no code markers (CRITICAL, HACK, FIXME, SCAR-NNN) because the codebase has one file; all scar evidence lives exclusively in git commit history.

## Failure Catalog Completeness

The codebase has 10 commits total. 8 are fix commits; 1 is a feat commit; 1 is the initial commit. All fix commits were searched. No SCAR-NNN numbered markers exist in the source (the workflow YAML has no inline code comments of that form). The catalog below covers every observable failure.

### SCAR-001: uv Sources Not Isolated in CI -- `uv sync` Missing `--no-sources`

- **Commit**: `5718454c44005df56c6a3bbdd4fa9eb91933b760`
- **Date**: 2026-03-11
- **What failed**: `uv sync` in the satellite CI `lint` and `test` jobs failed with "Distribution not found" because satellite `pyproject.toml` files use `[tool.uv.sources]` with editable local paths (e.g. `../autom8y/sdks/python/*`) for development. In CI those paths do not exist.
- **Root cause**: `uv sync` without `--no-sources` reads `[tool.uv.sources]` overrides and tries to resolve editable paths that do not exist in the CI runner filesystem.
- **Fix**: Added `--no-sources` to the `uv sync` call in the `lint` and `test` jobs.
- **Current marker**: `uv sync --no-sources` present at lines 147, 217, 313, 372 in `.github/workflows/satellite-ci-reusable.yml`.

### SCAR-002: `UV_NO_SOURCES=1` Env Var Is Not Recognized by uv

- **Commit**: `a719b0fc1dabdb6c01526d7aabef553fee767289`
- **Date**: 2026-03-11
- **What failed**: After SCAR-001 was fixed for `uv sync`, `uv run` calls continued to resolve `[tool.uv.sources]` editable paths and fail. The attempted fix was to set `UV_NO_SOURCES=1` at job level, but the env var is not recognized by uv.
- **Root cause**: `UV_NO_SOURCES` is not a valid uv environment variable. The flag must be passed explicitly to each `uv` subcommand.
- **Fix**: Set `UV_NO_SOURCES: "1"` at job level as documentation of intent; superseded by SCAR-003.
- **Current marker**: `UV_NO_SOURCES: "1"` at job-level `env:` lines 109, 174, 259, 327 -- intent documentation only, no functional effect.

### SCAR-003: `uv run` Calls Also Required `--no-sources` Explicitly

- **Commit**: `8b5ba88120f93c57ae44ba7e850159f941abb375`
- **Date**: 2026-03-11
- **What failed**: All `uv run` commands continued to read `[tool.uv.sources]` and fail even after SCAR-002. `UV_NO_SOURCES=1` had no effect.
- **Root cause**: uv does not honor `UV_NO_SOURCES`; `--no-sources` must be an explicit CLI flag on each invocation.
- **Fix**: Added `--no-sources` to every `uv run` call across all three jobs.
- **Current marker**: `--no-sources` on all `uv run` calls at lines 150, 153, 156, 160, 164, 243, 316, 422, 441, 453 in `.github/workflows/satellite-ci-reusable.yml`.

### SCAR-004: Hardcoded AWS Account ID in OIDC ARN

- **Commit**: `224e30d723150b0590d5564375c2cd63a375e325`
- **Date**: 2026-03-11
- **What failed**: The initial workflow had the AWS account ID hardcoded as `696318035277` in 3 IAM role ARNs. Configuration management failure; also exposes account ID in a public repository.
- **Root cause**: Copy-paste from internal monorepo without parameterization.
- **Fix**: Replaced all 3 hardcoded ARN references with `${{ vars.AWS_ACCOUNT_ID }}` org-level variable.
- **Current marker**: `vars.AWS_ACCOUNT_ID` at lines 128, 193, 289.

### SCAR-005: `contents: read` Misplaced Under `env:` in Integration Job

- **Commit**: `20c9007f85f39de523206eb20490a488ca8634c3`
- **Date**: 2026-03-11
- **What failed**: Integration job checkout failed with "Repository not found" for all satellite repos. `contents: read` was under `env:` instead of `permissions:` -- silently ignored by GitHub Actions.
- **Root cause**: YAML key placement error during authoring.
- **Fix**: Moved `contents: read` into `permissions:` block.
- **Current marker**: `permissions: { id-token: write, contents: read }` at lines 257-259.

### SCAR-006: GitHub App Token Required for Integration Test Checkout in Reusable Workflows

- **Commit**: `70d7481cb90ca028f781be2e751f5f500dae798a`
- **Date**: 2026-03-11
- **What failed**: Even after SCAR-005, integration job checkout failed. `GITHUB_TOKEN` in reusable workflows called from private satellite repos cannot checkout the calling repo -- GitHub Actions limits token scope when crossing public/private repo boundaries.
- **Root cause**: GitHub Actions constraint: `GITHUB_TOKEN` in reusable workflows is scoped to the workflow-host repo, not the calling satellite repo.
- **Fix**: Added `actions/create-github-app-token` step using `autom8y-satellite-dispatcher` GitHub App (APP_ID / APP_PRIVATE_KEY secrets); passed minted token to checkout.
- **Current marker**: `actions/create-github-app-token@d72941d797fd3113feb6b93fd0dec494b13a2547` step at lines 265-271; `token: ${{ steps.app-token.outputs.token }}` at line 273.

### SCAR-007: Convention-Check Failure Gate Blocked on Framework Spans -- UNKNOWN_ATTRIBUTE Filter

- **Commit**: `9ab5eb71345c212ad2daeb53ae23b970cb235057`
- **Date**: 2026-03-16
- **What failed**: After the `convention-check` job was added, it immediately failed CI. The CLI reported `UNKNOWN_ATTRIBUTE` violations for OTel SDK framework-internal spans that satellites cannot modify.
- **Fix**: Captured JSON output, filtered violations where `type == 'UNKNOWN_ATTRIBUTE'`. Superseded by SCAR-008.
- **Current marker**: Not present in current HEAD -- superseded.

### SCAR-008: UNKNOWN_ATTRIBUTE Filter Insufficient -- Domain Spans Require Dot-Notation Filter

- **Commit**: `094fa67dae1b4677b9e29e8ccc523e80c9b7514e`
- **Date**: 2026-03-16
- **What failed**: SCAR-007's type filter was insufficient; other violation types also fired for framework spans. The correct discriminator is span name structure: domain spans always contain a dot (`asana.tasks.sync`); framework/SDK spans do not.
- **Fix**: Replaced type filter with `'.' in v.get('span_name', '')` -- only count violations for spans with dot-notation names.
- **Current marker**: `'.' in v.get('span_name', '')` at lines 447-449; comment: `# Filter out framework spans (no dots in name = not domain spans)`.

---

## Category Coverage

| Category | Scars |
|---|---|
| Integration failure (CI tooling / external system boundary) | SCAR-001, SCAR-002, SCAR-003, SCAR-006 |
| Config drift / misconfiguration | SCAR-004, SCAR-005 |
| Schema evolution / API mismatch | SCAR-007, SCAR-008 |
| Security / credential exposure | SCAR-004 (co-classified) |

Searched but not observed: race condition, data corruption, performance cliff.

---

## Fix-Location Mapping

All fixes are in `.github/workflows/satellite-ci-reusable.yml`.

| Scar | Lines (Current HEAD) | Status |
|---|---|---|
| SCAR-001 | 147, 217, 313, 372 | Active |
| SCAR-002 | 109, 174, 259, 327 | Active (intent only) |
| SCAR-003 | 150, 153, 156, 160, 164, 243, 316, 422, 441, 453 | Active |
| SCAR-004 | 128, 193, 289 | Active |
| SCAR-005 | 257-259 | Active |
| SCAR-006 | 265-273 | Active |
| SCAR-007 | -- | Superseded by SCAR-008 |
| SCAR-008 | 444-449 | Active |

---

## Defensive Pattern Documentation

| Scar | Defensive Pattern | Regression Test |
|---|---|---|
| SCAR-001 | `--no-sources` on all `uv sync` | None -- CI is the regression surface |
| SCAR-002 | `UV_NO_SOURCES: "1"` env (intent only) | None |
| SCAR-003 | `--no-sources` on all `uv run` | None |
| SCAR-004 | `vars.AWS_ACCOUNT_ID` org variable | None |
| SCAR-005 | `contents: read` in `permissions:` | None |
| SCAR-006 | `create-github-app-token` before checkout | None |
| SCAR-008 | Dot-notation span name filter | None |

No regression tests exist. This is expected for a GitHub Actions workflow repository -- satellite CI runs act as the regression surface.

---

## Knowledge Gaps

1. No inline SCAR-NNN markers -- all scar evidence is in git commit messages only.
2. No test directory -- no automated regression tests exist.
3. SCAR-002's `UV_NO_SOURCES=1` env var is non-functional but remains in current HEAD; could confuse agents into thinking it is an active guard.
4. Satellite-side failures are not visible from this repository.

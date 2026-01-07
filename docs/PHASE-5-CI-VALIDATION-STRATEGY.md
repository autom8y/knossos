# Phase 5: CI Validation Strategy

## Overview

This document defines the CI/CD validation strategy for orchestrator.yaml and orchestrator.md files to ensure they stay synchronized and that orchestrators won't drift from their configuration-driven sources.

The strategy enforces a "source of truth" model:
- **Source of Truth**: `orchestrator.yaml` (configuration)
- **Generated Artifact**: `orchestrator.md` (produced by generator)
- **Contract**: Generated file must always match config; manual edits break the contract

## Validation Rules

### Rule 1: orchestrator.yaml Change Detected

**Trigger**: Any `.yaml` file changed in `rites/*/orchestrator.yaml`

**Validation Pipeline**:
1. **Validate YAML against schema**
   - Tool: `orchestrator-generate.sh --validate-only`
   - Checks: All required fields, type correctness, specialist references
   - Fails if: YAML is invalid or references non-existent specialists

2. **Generate orchestrator.md from YAML**
   - Tool: `orchestrator-generate.sh <team>`
   - Input: Updated orchestrator.yaml
   - Output: Generated orchestrator.md (temporary)
   - Fails if: Generation fails or produces invalid markdown

3. **Verify generated matches committed**
   - Compare: Generated file vs current orchestrator.md
   - Threshold: Must be 100% identical (no fuzzy matching)
   - Fails if: Files differ (regeneration required)

**Success Outcome**: "orchestrator.yaml validated, regenerated orchestrator.md matches"

**Failure Messages**:
- Schema validation failure: "YAML validation failed: [specific error]"
- Generation failure: "Failed to generate from YAML: [specific error]"
- File mismatch: "Generated orchestrator.md differs from committed. Run: orchestrator-generate.sh <team> --force && git add rites/<team>/agents/orchestrator.md"

### Rule 2: orchestrator.md Change Detected (Without YAML Change)

**Trigger**: `orchestrator.md` modified but `orchestrator.yaml` unchanged

**Validation Pipeline**:
1. **Identify orchestrator.md change**
   - Git detects modification to `rites/*/agents/orchestrator.md`

2. **Check for corresponding YAML change**
   - If YAML also changed: Skip this rule (handle in Rule 3)
   - If YAML unchanged: Warn about manual edit

3. **Apply drift prevention policy**
   - Default: **Fail CI** with helpful error message
   - Override: Allow with `skip-ci-orchestrator-check` commit message
   - Recommendation: "Update orchestrator.yaml instead (preferred) or use --allow-manual-edits flag"

**Success Outcome**: "Manual orchestrator.md edit detected; recommend updating YAML instead"

**Failure Messages**:
- Default: "ERROR: orchestrator.md changed without updating orchestrator.yaml. This breaks the generation contract. Either: (1) Update orchestrator.yaml and regenerate, or (2) Add 'skip-ci-orchestrator-check' to commit message to override"
- With override: Warning logged but CI passes

### Rule 3: Both orchestrator.yaml and orchestrator.md Changed

**Trigger**: Both files changed in same commit

**Validation Pipeline**:
1. **Validate YAML against schema**
   - Same as Rule 1, step 1
   - Fails if: YAML is invalid

2. **Generate orchestrator.md from YAML**
   - Same as Rule 1, step 2
   - Fails if: Generation fails

3. **Compare generated vs committed**
   - Calculate similarity: Line-by-line diff with threshold
   - Threshold: 95% similarity (allows for minor manual formatting tweaks)
   - Fails if: Similarity < 95%

**Success Outcome**: "Both files synchronized; generated matches committed (95%+ similarity)"

**Failure Messages**:
- Schema failure: Same as Rule 1
- Generation failure: Same as Rule 1
- Mismatch: "Generated orchestrator.md differs from committed. Generated version shows [X lines changed]. Run: orchestrator-generate.sh <team> --force && git add rites/<team>/agents/orchestrator.md"

## CI Pipeline Implementation

### GitHub Actions Workflow

Create `.github/workflows/validate-orchestrators.yml`:

**Triggers**:
- Pull request with changes to `rites/*/orchestrator.yaml` or `rites/*/agents/orchestrator.md`
- Push to main with same changes
- Manual trigger via workflow_dispatch

**Stages**:

1. **Setup** (required tools, environment)
2. **Detect Changes** (identify which orchestrators changed)
3. **Validate YAML** (schema, references)
4. **Generate Artifacts** (orchestrator-generate.sh)
5. **Compare** (generated vs committed)
6. **Report** (summary of validation results)

**Artifacts Produced**:
- `validation-report.txt` - Summary of all teams validated
- `generated-orchestrators/` - Generated files for review (in failed cases)

### GitLab CI Configuration

Create `.gitlab-ci.yml` section for orchestrator validation:

**Stage**: `validate` (before `test`)

**Jobs**:
- `validate-orchestrators:yaml` - Schema validation
- `validate-orchestrators:generation` - Generate from YAML
- `validate-orchestrators:compare` - Compare generated vs committed

## Pre-Commit Hook Integration

Create `.git/hooks/pre-commit` to catch drift locally:

**Purpose**: Prevent drift from reaching CI (fail fast)

**Workflow**:
1. Detect changed orchestrator files
2. If YAML changed: Regenerate MD and auto-stage
3. If MD changed without YAML: Warn but allow (can be overridden with `git commit --no-verify`)
4. If both changed: Regenerate from YAML, compare to committed

**Benefit**: Developers catch and fix issues before pushing to CI

## Validation Report Format

Each validation produces a structured report:

```
ORCHESTRATOR VALIDATION REPORT
==============================
Generated: 2025-12-29 10:30:00 UTC
Scope: 10 teams

RESULTS BY TEAM:
  rnd-pack          PASS  (config: valid, generation: ok, match: 100%)
  security-pack     PASS  (config: valid, generation: ok, match: 100%)
  ecosystem-pack    FAIL  (config: invalid field 'routing.unknown-specialist')
  sre-pack          PASS  (config: valid, generation: ok, match: 100%)
  ...

SUMMARY:
  ✓ Teams Validated: 9/10
  ✗ Teams Failed: 1/10
  ✗ Manual Edits Detected: 0

ERRORS (blocking):
  1. ecosystem-pack: Unknown specialist 'unknown-specialist' in routing

WARNINGS (non-blocking):
  (none)

NEXT STEPS:
  1. Fix errors in failing teams
  2. Rerun validation
  3. Merge when all teams pass

EXIT CODE: 1 (validation failed)
```

## Backward Compatibility

Existing hand-written orchestrators (teams not yet adopted):

**Phase A (Announcement)**:
- No validation required
- orchestrator.yaml is optional
- Teams can stay on hand-written versions indefinitely

**Phase B (Opt-In)**:
- Teams can migrate to YAML by running `orchestrator-migrate.sh <team>`
- Generator is available but not required

**Phase C (Infrastructure Update)**:
- All teams get generated orchestrator.yaml files
- CI begins validation for all teams
- Teams not opted in: Can choose to stay manual (generator doesn't overwrite)

**Phase D+ (Gradual Adoption)**:
- Non-adopting teams gradually migrate
- CI doesn't force adoption; just validates those who have YAML files

## Error Recovery

If validation fails:

1. **CI Failure**: PR cannot merge until fixed
2. **Local Failure**: Developer fixes locally using:
   ```bash
   orchestrator-generate.sh <team> --force
   git add rites/<team>/agents/orchestrator.md
   git commit --amend  # or make new commit
   ```

3. **Emergency Override**: In extreme cases, use `skip-ci-orchestrator-check` commit message to bypass (strongly discouraged, requires review)

## Success Criteria for Phase 5

- [ ] CI validation runs on all PRs touching orchestrators
- [ ] Pre-commit hook catches drift locally
- [ ] Validation reports are clear and actionable
- [ ] No false positives (legitimate changes pass)
- [ ] No false negatives (drift is always caught)
- [ ] 100% of orchestrator changes validated
- [ ] Rollback procedures tested and documented
- [ ] Team feedback incorporated into workflow

## Related Documents

- `ROLLOUT-orchestrator-templates.md` - Phased adoption plan
- `ORCHESTRATOR-CI-IMPLEMENTATION.md` - Technical implementation details
- `orchestrator-generate.sh` - Generator script documentation
- `validate-orchestrator.sh` - Validation script documentation

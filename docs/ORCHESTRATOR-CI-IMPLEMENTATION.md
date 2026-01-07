# Orchestrator Templating: CI/CD Implementation Guide

## Overview

This guide provides technical details for implementing orchestrator validation in CI/CD pipelines. Currently provides GitHub Actions implementation; GitLab CI and other platforms can follow the same logic.

## Files Involved

```
.github/workflows/validate-orchestrators.yml  # GitHub Actions workflow
.githooks/pre-commit-orchestrator             # Local pre-commit hook
templates/orchestrator-generate.sh            # Generator script
templates/validate-orchestrator.sh            # Validation script
schemas/orchestrator.yaml.schema.json         # YAML schema
```

## GitHub Actions Workflow Details

### Workflow Trigger Events

```yaml
on:
  pull_request:
    paths:
      - 'rites/*/orchestrator.yaml'
      - 'rites/*/agents/orchestrator.md'
  push:
    branches: [main]
    paths: [same as above]
  workflow_dispatch:
    inputs: [manual trigger options]
```

### Job Execution Flow

```
TRIGGER
  ↓
detect-changes (outputs: changed files, teams)
  ↓
setup (verify tools, permissions)
  ├─ validate-yaml (in parallel)
  ├─ (waits for both)
  ├─ generate-orchestrators (in parallel)
  ├─ compare-orchestrators (in parallel)
  ├─ validate-generated (in parallel)
  ↓
report (summary)
```

**Parallelization**: Jobs 3-5 run in parallel to minimize total time

**Critical Path**: detect-changes → setup → validate-yaml → report (~30 seconds)

### Key Variables & Outputs

**From detect-changes job**:
```json
{
  "changed-yaml": ["rites/rnd-pack/orchestrator.yaml", ...],
  "changed-md": ["rites/rnd-pack/agents/orchestrator.md", ...],
  "yaml-teams": ["rnd-pack", "security-pack", ...],
  "md-teams": ["rnd-pack", ...],
  "has-changes": "true"
}
```

**From each validation job**:
- Exit code: 0 (pass) or non-zero (fail)
- Logs: Captured in GitHub Actions logs
- Artifacts: Generated orchestrators (upload for inspection)

### Dependency Injection

Dependencies auto-installed on runner:
- `jq` - JSON processing
- `yq` - YAML processing
- `bash`, `grep`, `sed`, `awk` - Standard tools

If missing, workflow fails with clear error message.

### Error Handling

Each job implements:
- `set -euo pipefail` - Exit on error
- Clear error messages
- Artifact upload (for debugging failures)
- Build summary with action items

## Pre-Commit Hook Implementation

### Installation

Users install manually or via git config:

```bash
# Option A: Manual copy
cp .githooks/pre-commit-orchestrator .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

# Option B: Configure git to use .githooks/
git config core.hooksPath .githooks
```

### Lifecycle

1. User stages changes: `git add rites/<team>/orchestrator.yaml`
2. User runs: `git commit`
3. Git automatically runs: `.git/hooks/pre-commit`
4. Hook detects changes, validates, regenerates if needed
5. Hook either approves commit or blocks with message

### Hook Exit Codes

- `0` - All checks passed, commit allowed
- `1` - Validation failed, commit blocked (can override with `--no-verify`)
- `2` - Schema error (should never reach this with valid YAML)

### Hook Output Example

```
INFO: Found 1 YAML file(s) and 0 MD file(s) to validate

INFO: Validating YAML: rites/rnd-pack/orchestrator.yaml
OK: YAML validated: rites/rnd-pack/orchestrator.yaml
INFO: Regenerating orchestrator.md for: rnd-pack
INFO: Changes detected in regenerated orchestrator.md, staging update
OK: Regenerated: rites/rnd-pack/agents/orchestrator.md

OK: Pre-commit validation passed
INFO: Regenerated orchestrators have been staged
```

## Validation Rules Matrix

| Rule | Trigger | Action | Pass Condition | Fail Condition |
|------|---------|--------|---|---|
| **YAML Schema** | Any `orchestrator.yaml` change | Validate YAML against schema | All required fields present, valid types | Missing fields, invalid types |
| **Specialist Refs** | Any `orchestrator.yaml` change | Check specialists exist in workflow.yaml | All routing specialists found | Unknown specialist name |
| **Generation** | YAML validated | Generate orchestrator.md from YAML | Generation completes without error | Generation fails (syntax/logic error) |
| **File Match** | Both YAML and MD changed | Compare generated vs committed (95% threshold) | Files match or < 5% different | > 5% difference |
| **MD Without YAML** | MD changed, YAML not changed | Warn about broken contract | Team aware, can override | Auto-fail (prevent drift) |
| **Markdown Validity** | After generation | Validate structure and syntax | All required sections, balanced code blocks | Invalid markdown, missing sections |

## Measuring Performance

### Generation Time

Expected:
- Single team: < 500ms
- All teams (10): < 5 seconds

Track in CI logs:
```bash
time ./templates/orchestrator-generate.sh rnd-pack
# real    0m0.423s
# user    0m0.234s
# sys     0m0.178s
```

Alert if > 2 seconds (indicates slow jq query or I/O bottleneck)

### Validation Time

Expected:
- Schema validation: < 100ms
- Markdown validation: < 100ms

Total CI runtime should be < 1 minute for normal changes

### Resource Usage

Monitor in GitHub Actions:
- CPU: Should stay < 20% (single core)
- Memory: Should stay < 256MB
- Disk: Should stay < 100MB

Alert if any metric spikes (indicates infinite loop or memory leak)

## Integration with Other CI Jobs

### Before Running Tests

Place orchestrator validation early in CI pipeline:

```yaml
stages:
  - validate        # orchestrator validation (this stage)
  - lint            # code linting
  - test            # unit tests
  - build           # build artifacts
  - deploy          # deployment
```

Rationale: Fail fast on validation before expensive test runs

### Blocking vs Warning

Options:
- **Block** (default): Validation failure prevents merge (strict enforcement)
- **Warn** (optional): Validation failure shows warning but allows merge (soft enforcement)
- **Skip** (override): Team can skip with commit message (emergency only)

Recommended: Block during adoption, Warning after 100% adoption

## Customization for Other Git Platforms

### GitLab CI

Replace GitHub Actions with:

```yaml
# .gitlab-ci.yml
validate-orchestrators:
  stage: validate
  script:
    - ./templates/orchestrator-generate.sh --all --validate-only
  artifacts:
    paths:
      - rites/*/agents/orchestrator.md
    expire_in: 1 week
  only:
    - merge_requests
    - main
```

### Jenkins

```groovy
pipeline {
    agent any

    stages {
        stage('Validate Orchestrators') {
            steps {
                sh './templates/orchestrator-generate.sh --all --validate-only'
            }
        }
    }
}
```

### Bitbucket Pipelines

```yaml
# bitbucket-pipelines.yml
pipelines:
  pull-requests:
    '**':
      - step:
          name: Validate Orchestrators
          script:
            - ./templates/orchestrator-generate.sh --all --validate-only
```

## Debugging Failed Validations

### Common Failure Scenarios

**Scenario 1: YAML Validation Fails**

```
ERROR: Schema validation failed: rites/rnd-pack/orchestrator.yaml
ERROR: Missing required field in orchestrator.yaml: routing
```

**Debug**:
1. Check YAML syntax: `yq eval '.' rites/rnd-pack/orchestrator.yaml`
2. Check required fields: `jq 'keys' schemas/orchestrator.yaml.schema.json`
3. Compare to example: `cat schemas/orchestrator.yaml.schema.json | jq '.examples[0]'`

**Scenario 2: Generation Fails**

```
ERROR: Failed to generate orchestrator.md for: rnd-pack
```

**Debug**:
1. Run generator with dry-run: `./templates/orchestrator-generate.sh rnd-pack --dry-run`
2. Capture stderr: `./templates/orchestrator-generate.sh rnd-pack 2>&1 | tail -20`
3. Check template file: `wc -l templates/orchestrator-base.md.tpl`
4. Verify workflow.yaml exists: `ls rites/rnd-pack/workflow.yaml`

**Scenario 3: File Mismatch**

```
MISMATCH: rnd-pack (142 lines differ)
Generated vs Committed differ
```

**Debug**:
1. Regenerate locally: `./templates/orchestrator-generate.sh rnd-pack --force`
2. Compare: `diff rites/rnd-pack/agents/orchestrator.md /tmp/orchestrator.md`
3. Check for whitespace: `od -c rites/rnd-pack/agents/orchestrator.md | head -20`
4. Compare checksums: `md5sum rites/rnd-pack/agents/orchestrator.md`

**Resolution**: Regenerate and commit the updated file

## CI/CD Maintenance

### Monthly Review

- Check for performance regressions (generation time creeping up)
- Review error logs for common patterns
- Update documentation based on repeated questions
- Verify alert thresholds still appropriate

### Quarterly Audit

- Test full CI pipeline end-to-end with test PR
- Verify all dependencies available
- Check for deprecated GitHub Actions versions
- Review security of workflow (secrets, permissions)

### Annual Evaluation

- Measure total ROI (hours saved vs maintenance effort)
- Compare against alternatives (manual validation, other generators)
- Plan for next iteration (new features, simplified workflow)

## Security Considerations

### Permissions

GitHub Actions workflow has minimal permissions:

```yaml
permissions:
  contents: read        # Can read repository
  pull-requests: read   # Can read PR info
  checks: write         # Can write check results
```

**Not granted**:
- Write to main branch
- Merge PRs
- Deploy to production
- Access secrets

### Input Validation

All user inputs (rite names, file paths) validated:

```bash
# Example: Validate rite name
TEAM="$1"
[[ "$TEAM" =~ ^[a-z0-9-]+$ ]] || die "Invalid rite name"
[[ -d "rites/$TEAM" ]] || die "Team directory not found"
```

### Secret Protection

No secrets exposed in logs:
- Error messages don't include file contents
- Passwords/tokens never logged
- User input sanitized before use

## Related Documents

- `.github/workflows/validate-orchestrators.yml` - Full workflow
- `.githooks/pre-commit-orchestrator` - Full hook script
- `templates/orchestrator-generate.sh` - Generator implementation
- `PHASE-5-CI-VALIDATION-STRATEGY.md` - Validation rules

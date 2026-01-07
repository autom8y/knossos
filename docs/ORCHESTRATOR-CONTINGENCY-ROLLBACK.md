# Orchestrator Templating: Contingency & Rollback Procedures

## Overview

This document defines procedures for handling issues that arise during orchestrator templating rollout and ongoing operations.

## Incident Severity Levels

| Level | Impact | Response | Escalation |
|-------|--------|----------|------------|
| **P0 (Critical)** | Schema validation broken, blocks all CI | Page on-call immediately | Director + Tech Lead |
| **P1 (High)** | Generation failure, team blocked | Respond within 15 min | Team Lead + Integration Lead |
| **P2 (Medium)** | Validation false positives, team workaround | Respond within 1 hour | Integration Lead |
| **P3 (Low)** | Documentation unclear, not blocking | Respond within 24 hours | Integration Lead |

## P0: Critical Issues

### Issue: Schema Validation Broken

**Symptoms**: All validation checks fail, CI completely blocked

**Root Causes**:
- Schema file corrupted or invalid
- yq/jq version incompatibility
- Dependency missing on CI runner

**Immediate Response** (< 5 minutes):
1. Disable orchestrator validation workflow (workflow_dispatch → disable)
2. Page on-call engineer
3. Notify all team leads: "Orchestrator validation temporarily offline"
4. Begin investigation in parallel

**Investigation** (< 30 minutes):
1. Verify schema file is valid JSON: `jq . schemas/orchestrator.yaml.schema.json`
2. Check yq/jq versions on CI runner
3. Test schema validation locally with latest commits
4. Review recent changes to schema or validation script

**Recovery** (< 1 hour):
1. If schema corrupted: Revert last schema change
2. If tool issue: Update tools on CI runner
3. Test validation locally: `./templates/orchestrator-generate.sh rnd-pack --validate-only`
4. Re-enable workflow
5. Post-mortem: Document root cause and preventative measures

**Communication**:
```
INCIDENT: Orchestrator Validation Recovery
============================================
Status: RESOLVED
Duration: X minutes
Cause: [description]
Action: [what was fixed]
Next: [preventative measures]
```

### Issue: Generation Script Fatal Error

**Symptoms**: `orchestrator-generate.sh` crashes on all teams, no output produced

**Root Causes**:
- Template file corrupted or deleted
- Placeholder syntax broken
- jq/sed compatibility issue
- File permission problem

**Immediate Response** (< 5 minutes):
1. Check generator script exists and is executable: `ls -la templates/orchestrator-generate.sh`
2. Check template file exists: `ls -la templates/orchestrator-base.md.tpl`
3. Try generation manually: `./templates/orchestrator-generate.sh rnd-pack --dry-run`
4. Capture error output

**Investigation**:
1. Verify template file integrity (check git status, recent diffs)
2. Test jq and sed commands in isolation
3. Check file permissions: `chmod +x templates/orchestrator-*.sh`
4. Test with simplest team first (rnd-pack is usually simplest)

**Recovery**:
1. If template corrupted: `git checkout templates/orchestrator-base.md.tpl`
2. If permission issue: `chmod +x templates/orchestrator-*.sh`
3. If tool issue: Update bash/sed/jq versions
4. Test: `./templates/orchestrator-generate.sh --all --validate-only`
5. Re-enable CI workflow

**Prevention**:
- Add integration test for basic generation in CI
- Verify template syntax in pre-commit hook
- Pin tool versions in GitHub Actions

### Issue: CI Runner Resource Exhaustion

**Symptoms**: Validation timeouts, jobs hanging, out of memory

**Symptoms**:
- Jobs running > 10 minutes (normal: < 2 minutes)
- Job fails with "out of memory" or "timeout"
- Multiple jobs in queue, not progressing

**Immediate Response**:
1. Check GitHub Actions runner status
2. Review job logs for resource usage
3. Identify which step is consuming resources

**Investigation**:
1. Check for infinite loops in jq/sed commands
2. Review if orchestrator.yaml files have grown very large
3. Check if running against all teams unnecessarily
4. Profile generation time: `time ./templates/orchestrator-generate.sh <team>`

**Recovery**:
1. Optimize slow steps in generator script
2. Increase runner memory/CPU if needed (GitHub Actions setting)
3. Parallelize validation jobs (validate multiple teams in parallel)
4. Add timeout to stuck jobs

---

## P1: High-Priority Issues

### Issue: Team-Blocking Bug in Generation

**Symptoms**: Generation works for 9 teams but fails for 1, team cannot merge

**Root Causes**:
- Edge case in YAML handling (special characters, nested structures)
- Team's orchestrator.yaml uses feature not in schema
- Specialist name mismatch between workflow.yaml and config

**Response Timeline**: < 15 minutes

**Steps**:
1. Reproduce locally: `./templates/orchestrator-generate.sh <failing-team> --dry-run`
2. Capture full error output
3. Compare failing team's YAML to working team's YAML
4. Identify edge case

**Workarounds** (while fixing):
- Team can add `skip-ci-orchestrator-check` to commit message (bypasses validation)
- Team can revert to hand-written orchestrator.md (delete orchestrator.yaml)
- Team stays on existing orchestrator.md until fix deployed

**Fix**:
1. Update generator to handle edge case
2. Test with failing team's config
3. Test with all other teams to ensure no regression
4. Deploy fix to production
5. Team re-triggers CI

**Communication**:
```
We found an issue affecting [team]'s orchestrator generation:
[description of edge case]

TEMPORARY WORKAROUND: You can merge with 'skip-ci-orchestrator-check'
in your commit message while we fix this.

FIX ETA: [time]

We'll follow up when ready to re-run validation.
```

### Issue: False Positive: Generated Differs from Committed

**Symptoms**: CI fails with "Generated orchestrator.md differs from committed" but files look identical

**Root Causes**:
- Whitespace differences (trailing spaces, line endings)
- Comment differences in YAML parsing
- Formatting differences between tools
- Race condition in temp file generation

**Response Timeline**: < 30 minutes

**Steps**:
1. Compare files manually: `diff rites/<team>/agents/orchestrator.md /tmp/generated.md`
2. Use `od` to check for hidden characters: `od -c rites/<team>/agents/orchestrator.md | head -20`
3. Check for Windows vs Unix line endings: `file rites/<team>/agents/orchestrator.md`

**Resolution Options**:

**Option A: Regenerate** (Recommended)
```bash
./templates/orchestrator-generate.sh <team> --force
git add rites/<team>/agents/orchestrator.md
git commit --amend
git push -f  # Only if not yet merged
```

**Option B: Override** (Temporary)
```bash
git commit --amend -m "$(git log -1 --pretty=%B)

skip-ci-orchestrator-check"
```

**Prevention**:
- Normalize line endings in CI: `dos2unix` before comparison
- Use `--strip-trailing-whitespace` in git config
- Document expected whitespace handling in skill

### Issue: Mass Generation Error (Multiple Teams Fail)

**Symptoms**: `--all` flag fails for multiple teams, unclear which are broken

**Root Causes**:
- Schema update incompatible with existing configs
- Template syntax change
- Systematic issue (e.g., all configs missing new required field)

**Response Timeline**: < 1 hour

**Steps**:
1. Test generation for each team individually:
   ```bash
   for team in $(ls teams); do
     echo "Testing $team..."
     ./templates/orchestrator-generate.sh "$team" --validate-only || echo "FAILED: $team"
   done
   ```
2. Identify pattern in failures
3. Determine if config or generator is at fault

**If Config Issue**:
1. Issue applies to all teams (or subset)
2. Update all affected configs in one PR
3. Test batch generation: `./templates/orchestrator-generate.sh --all`

**If Generator Issue**:
1. Revert recent generator changes: `git revert <commit>`
2. Re-test: `./templates/orchestrator-generate.sh --all --validate-only`
3. Investigate root cause of original change
4. Fix and re-apply

**Communication**:
```
We discovered a compatibility issue affecting [N] teams:
[description]

We're fixing this now. ETA: [time]

Your orchestrators will continue to work; this only affects new commits
with orchestrator.yaml changes.
```

---

## P2: Medium-Priority Issues

### Issue: Validation Too Strict (False Positives)

**Symptoms**: Valid orchestrators fail validation, teams hitting false positives

**Examples**:
- Skill reference format validation too strict
- Specialist name case-sensitivity issue
- Whitespace in YAML array handling

**Response Timeline**: < 2 hours

**Steps**:
1. Identify exact validation rule causing false positive
2. Review rule logic in schema or validation script
3. Test with team's specific config
4. Determine if rule is too strict or config is invalid

**Resolution**:
- **Loosen Rule**: If multiple teams affected, update validation
- **Document Exception**: If edge case, document expected behavior
- **Update Config**: If config is non-standard, suggest fix to team

**Team Workaround**:
```bash
# Temporarily bypass validation
git commit -m "Fix orchestrator config

skip-ci-orchestrator-check"
```

**Communication**:
```
We found that [teams] are hitting a validation rule that may be too strict:
[description]

We're investigating and will either:
1. Loosen the rule (if it's overly restrictive), or
2. Update the documentation (if the rule is correct)

Your config is valid; you can bypass with 'skip-ci-orchestrator-check'
while we sort this out.
```

### Issue: Documentation Incomplete or Unclear

**Symptoms**: Teams asking same question multiple times, adoption slower than expected

**Response Timeline**: < 4 hours

**Steps**:
1. Collect feedback from teams hitting issue
2. Identify gap in documentation
3. Add clear example or clarification to `@orchestrator-templates` skill
4. Link to updated documentation in response

**Example Update**:
```markdown
## Common Question: How do I handle custom antipatterns?

If your team has specific anti-patterns not in the base list:

1. Add to `antipatterns:` array in orchestrator.yaml:
   ```yaml
   antipatterns:
     - "Treating PATCH as SYSTEM (different scope)"
   ```

2. Regenerate:
   ```bash
   ./orchestrator-generate.sh <team>
   ```

3. The antipatterns will appear in your orchestrator.md
```

### Issue: Pre-Commit Hook Installation Issues

**Symptoms**: Teams wanting to use hook but having permission/setup issues

**Response Timeline**: < 1 hour

**Workaround**:
```bash
# Manual hook setup
cp .githooks/pre-commit-orchestrator .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

**Permanent Solution**:
```bash
# Add to git config
git config core.hooksPath .githooks

# Enable all hooks in .githooks/
```

---

## P3: Low-Priority Issues

### Issue: Slow Generation

**Symptoms**: Generation takes > 5 seconds, feels sluggish

**Root Causes**:
- Complex jq query
- File I/O bottlenecks
- Redundant template processing

**Response Timeline**: < 24 hours

**Steps**:
1. Profile generation: `time ./templates/orchestrator-generate.sh <team>`
2. Identify slow step (check timing comments in script)
3. Optimize with better jq query or bash builtin

**Prevention**:
- Add performance tests to CI
- Alert if generation time > 2 seconds

### Issue: Confusing Error Messages

**Symptoms**: Teams confused by validation error message

**Example**:
```
ERROR: Unknown specialist 'integration-enginer' in routing
```

Should be:
```
ERROR: Specialist 'integration-enginer' not found in workflow.yaml

Did you mean: integration-engineer?

Check: rites/<team>/workflow.yaml for list of available specialists
```

**Response Timeline**: < 24 hours

**Action**: Update error message, re-deploy generator

---

## Rollback Procedures

### Rollback Scenario 1: Revert All Changes

**When**: Critical issue discovered, need to disable feature entirely

**Steps**:
```bash
# Disable CI workflow
gh workflow disable validate-orchestrators -R <owner>/<repo>

# Delete orchestrator.yaml files (keep for recovery)
git rm rites/*/orchestrator.yaml

# Commit rollback
git commit -m "Revert orchestrator templating (emergency rollback)"

# Notify teams
# Message: "Orchestrator templating disabled due to critical issue.
# Please ignore orchestrator.yaml files; use orchestrator.md as before."
```

**Recovery Time**: Immediate (< 1 hour)

**Communication**:
```
INCIDENT: Orchestrator Templating Rollback
============================================
Action: Disabled orchestrator templating feature
Reason: [brief explanation]
Timeline: [how long it will take to fix]
Impact: Teams continue using hand-written orchestrators (zero impact)
Next: We'll re-enable once issue is resolved
```

### Rollback Scenario 2: Revert Recent Generator Change

**When**: Recent commit to generator introduces bug

**Steps**:
```bash
# Identify problematic commit
git log --oneline templates/orchestrator-generate.sh | head -5

# Revert it
git revert <commit-hash>

# Test
./templates/orchestrator-generate.sh --all --validate-only

# Push
git push
```

**Recovery Time**: 10-30 minutes

### Rollback Scenario 3: Team Wants to Unadopt

**When**: Team has issues with templating, wants to go back to manual

**Steps**:
```bash
# For the team:
cd /roster/rites/<team>

# Delete YAML config
rm orchestrator.yaml

# Update manifest
jq '.orchestrator.source = "user"' AGENT_MANIFEST.json > temp && mv temp AGENT_MANIFEST.json

# Commit
git add orchestrator.yaml AGENT_MANIFEST.json
git commit -m "Revert to manual orchestrator for <team>"

# Team continues with orchestrator.md as-is (no changes needed)
```

**Recovery Time**: < 5 minutes

**Result**: Team's orchestrator.md stays as-is, CI no longer validates it

### Rollback Scenario 4: Schema Update Breaks Compatibility

**When**: Schema changed in way that breaks existing configs

**Steps**:
```bash
# Identify problematic schema change
git log --oneline schemas/orchestrator.yaml.schema.json | head -3

# Revert schema to previous version
git revert <commit-hash>

# Verify all teams validate
./templates/orchestrator-generate.sh --all --validate-only

# Update affected configs (if needed)
# ...

# Commit fix
git push
```

**Recovery Time**: 30-60 minutes

---

## Testing Contingency Plans

Each contingency procedure should be tested quarterly:

- [ ] P0 Critical: Revert generator, verify all configs still validate
- [ ] P1 High: Introduce team-specific bug, verify workaround works
- [ ] P2 Medium: Test loosen/tighten validation rules
- [ ] P3 Low: Update error message, verify clarity

## Communication Templates

### Incident Declared

```
INCIDENT: [brief title]
START: [time]
SEVERITY: P[0-3]

IMPACT: [what's broken]
AFFECTED: [which rites/workflows]
ETA FIX: [estimated time]
WORKAROUND: [temporary solution if available]

Updates: [link to status page or Slack thread]
```

### Incident Resolved

```
INCIDENT RESOLVED: [title]
DURATION: [X minutes/hours]
ROOT CAUSE: [brief explanation]
FIX: [what was done]
PREVENTION: [how we'll avoid this next time]

Sorry for the disruption. Thanks for your patience.
```

### Incident Post-Mortem

Within 48 hours of P0/P1:

```
POST-MORTEM: [title]
WHEN: [date and time]
DURATION: [total impact time]

ROOT CAUSE ANALYSIS:
[detailed timeline and explanation]

CONTRIBUTING FACTORS:
- [factor 1]
- [factor 2]

ACTION ITEMS:
- [ ] [item 1] (owner, due date)
- [ ] [item 2] (owner, due date)

PREVENTATIVE MEASURES:
- [what we're doing to prevent recurrence]
```

---

## Related Documents

- `PHASE-5-CI-VALIDATION-STRATEGY.md` - Validation rules
- `ROLLOUT-orchestrator-templates.md` - Adoption timeline
- `ORCHESTRATOR-METRICS-MONITORING.md` - Tracking incidents
- `README.md` - Escalation contacts

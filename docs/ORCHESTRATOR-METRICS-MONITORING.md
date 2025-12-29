# Orchestrator Templating: Metrics & Monitoring

## Overview

This document defines success metrics, monitoring approach, and dashboard design for orchestrator templating rollout and ongoing operations.

## Primary Metrics

### Adoption Metrics

**Teams Adopted**
- Definition: Number of teams actively using orchestrator.yaml for generation
- Measurement: Count of teams with non-empty orchestrator.yaml files
- Target: 10/10 teams by end of Phase E
- Frequency: Weekly
- Owner: Integration team lead

**Adoption Percentage**
- Definition: (Teams Adopted / 10) * 100
- Target Timeline:
  - Week 2: 10-20%
  - Week 4: 20-30%
  - Month 3: 50-70%
  - Month 6: 90-100%

**Adoption Velocity**
- Definition: Teams adopting per week
- Target: Accelerating (week 1-2 slow, week 3+ faster as trust builds)
- Alert: If velocity drops below 1 team/week after week 3

### Validation Metrics

**CI Validation Pass Rate**
- Definition: (Successful validations / Total validations) * 100
- Target: 100% (all orchestrators valid)
- Frequency: Per commit
- Alert: If < 95%, page on-call
- Measurement: GitHub Actions workflow results

**Schema Violations Prevented**
- Definition: Number of YAML schema errors caught by CI (before merge)
- Target: All caught (0 drift reaching main)
- Frequency: Weekly
- Measurement: CI logs from validate-yaml job

**Manual Drift Detected**
- Definition: Commits with orchestrator.md changed but orchestrator.yaml unchanged
- Target: < 2 per month (teams making intentional exceptions)
- Frequency: Weekly
- Measurement: Pre-commit hook logs + git commit analysis

**Pre-Commit Hook Usage**
- Definition: Number of teams with pre-commit hook installed
- Target: 100% of adopting teams
- Frequency: Monthly survey
- Measurement: Hook install script execution logs

### Quality Metrics

**Generated Orchestrator Validity**
- Definition: % of generated orchestrators passing validation script
- Target: 100%
- Measurement: `validate-orchestrator.sh` results in CI
- Alert: If any generation produces invalid markdown, page lead engineer

**Placeholder Replacement Success**
- Definition: No unreplaced {{PLACEHOLDER}} in generated output
- Target: 100%
- Measurement: grep for {{[A-Z_]*}} in generated files
- Alert: If any found, immediate rollback and investigation

**Markdown Syntax Correctness**
- Definition: Valid markdown (balanced code blocks, proper YAML frontmatter)
- Target: 100%
- Measurement: markdown linter results
- Alert: If validation fails, investigate formatter

### Maintenance Metrics

**Orchestrator Update Velocity** (Time to update)
- Definition: Mean time from decision to merge for orchestrator changes
- Baseline (Pre-templating): ~2 hours (manual editing, testing, review)
- Target (Templating): ~5 minutes (edit YAML, regenerate, merge)
- Improvement: 24x faster
- Measurement: Git log timestamps + manual survey
- Frequency: Quarterly

**Maintenance Burden**
- Definition: Hours per month spent on orchestrator maintenance
- Baseline (Pre-templating): ~2 hours/month per team * 10 teams = 20 hours
- Target (Templating): ~1 hour/month shared (generator updates) = 1 hour
- Improvement: 20x reduction
- Measurement: Commit message analysis + team time logs
- Frequency: Monthly

**Documentation Updates**
- Definition: How often orchestrator template changes
- Target: < 1x per quarter (minimal churn)
- Measurement: Git history of orchestrator-base.md.tpl
- Alert: If > 1 per month, reassess template design

## Secondary Metrics

### Process Metrics

**Documentation Hit Rate**
- Definition: % of adoption questions answered by @orchestrator-templates skill
- Target: 90%+
- Measurement: Skill view/search analytics + team feedback
- Frequency: Weekly

**Support Request Volume**
- Definition: Number of support requests for orchestrator templating
- Target: Decreasing (week 2: 5-10, week 4: 3-5, month 3: 1-2, month 6: 0)
- Measurement: Ticket system or email threads
- Frequency: Weekly

**Pairing Session Utilization**
- Definition: % of teams taking advantage of offered pairing sessions
- Target: 50%+ during early phases, 0 by production
- Measurement: Calendar bookings + attendance
- Frequency: Weekly

### Error Metrics

**Schema Validation Errors**
- Definition: Number of orchestrator.yaml files failing schema validation
- Target: 0 (caught in CI before merge)
- Measurement: CI logs from validate-yaml job
- Frequency: Per commit
- Alert: Any validation error pageable incident

**Generation Failures**
- Definition: Number of orchestrators failing to generate
- Target: 0
- Measurement: orchestrator-generate.sh exit codes in CI
- Frequency: Per commit
- Alert: Page on-call if any team's generation fails

**File Mismatch Errors**
- Definition: Generated file differs from committed file
- Target: 0 (all synchronized)
- Measurement: diff output in compare-orchestrators job
- Frequency: Per commit
- Alert: PR fails check

**Drift Events**
- Definition: Number of times orchestrator.md manually edited (breaking contract)
- Target: < 2 per month (intentional exceptions OK)
- Measurement: Pre-commit hook warnings + git analysis
- Frequency: Weekly review

### Coverage Metrics

**Team Coverage**
- Definition: % of teams with orchestrator.yaml (generated or adopted)
- Target: 100% by Phase C (all teams have YAML files)
- Measurement: find teams -name orchestrator.yaml | wc -l
- Frequency: Weekly

**Specialist Coverage**
- Definition: Do all specialists in workflow.yaml appear in routing table?
- Target: 100%
- Measurement: Schema validation in generate script
- Alert: If any specialist missing, generation fails

**Handoff Criteria Coverage**
- Definition: Do all phases in workflow.yaml have handoff criteria defined?
- Target: 100%
- Measurement: Schema validation checklist
- Alert: If any phase missing criteria, generation fails

## Monitoring Dashboard

### Real-Time Dashboard (GitHub)

```
ORCHESTRATOR TEMPLATING STATUS
==============================
Last Updated: 2025-12-29 14:30 UTC

ADOPTION
--------
Teams Adopted:        3/10 (30%)  [████░░░░░░]
Adoption Velocity:    +1.5 teams/week
Projected Completion: ~8 weeks

CI VALIDATION
-------------
Overall Pass Rate:    100% (42/42)
Schema Errors:        0 (blocked: 0)
Generation Errors:    0
File Mismatches:      0
Drift Events:         1 (flagged)

QUALITY
-------
Generated Validity:   100%
Placeholder Errors:   0
Markdown Errors:      0
Last Validation:      5 mins ago

MAINTENANCE
-----------
Update Velocity:      5 min average
Maintenance Burden:   Tracking...
Template Changes:     0 this month
Support Requests:     2 open

TEAMS STATUS
============
rnd-pack              ✓ ADOPTED      (generated.md: 100% match)
security-pack         ✓ ADOPTED      (generated.md: 100% match)
ecosystem-pack        ~ MIGRATING    (PR #423)
doc-team-pack         ○ INTERESTED   (in roadmap)
forge-pack            ○ NOT ADOPTED  (manual orchestrator)
debt-triage-pack      ○ NOT ADOPTED  (manual orchestrator)
sre-pack              ○ NOT ADOPTED  (manual orchestrator)
strategy-pack         ○ NOT ADOPTED  (manual orchestrator)
integration-researcher ○ NOT ADOPTED (manual orchestrator)
moonshot-architect    ○ NOT ADOPTED  (manual orchestrator)

ALERTS
======
(none)

RESOURCES
---------
Generator Script:     /repos/orchestrator-generate.sh
CI Workflow:          .github/workflows/validate-orchestrators.yml
Docs:                 docs/ROLLOUT-orchestrator-templates.md
Metrics:              This file
```

### Weekly Metrics Report

**Recipients**: Team leads, integration team, stakeholders

**Format**: Email + embedded dashboard

**Content**:
- Adoption progress (graph showing trend)
- Top blockers (if any)
- Successful teams (shout-outs)
- Upcoming milestones
- Action items for coming week

### Monthly Executive Summary

**Recipients**: Leadership, product team

**Content**:
- Adoption % and trend
- Maintenance hours saved YTD
- Quality metrics (error rate, validation pass rate)
- Blockers and risks
- Budget impact (if applicable)
- Recommendations for next month

## Tracking Implementation

### GitHub Workflow Artifacts

Each validation run produces:
```
artifacts/
  validation-report.txt
  generated-orchestrators/
    rnd-pack-orchestrator.md
    security-pack-orchestrator.md
    ...
```

Extract metrics from these artifacts:
- Schema validation errors → stored in validation-report
- Generation failures → captured in stderr logs
- File mismatches → diff output
- Validation pass/fail → exit codes

### Metrics Collection Script

Create `/scripts/collect-orchestrator-metrics.sh`:

```bash
#!/bin/bash
# Collects orchestrator metrics weekly

REPO_HOME="$HOME/Code/roster"
METRICS_FILE="$REPO_HOME/docs/METRICS-current.md"

# Count teams with orchestrator.yaml
ADOPTED=$(find "$REPO_HOME/teams" -name "orchestrator.yaml" | wc -l)
TOTAL=10
PCT=$((ADOPTED * 100 / TOTAL))

# Check CI status (last 10 runs)
PASS=$(gh workflow run list --workflow=validate-orchestrators.yml --limit=10 | grep "completed" | wc -l)
FAIL=$((10 - PASS))
PASS_RATE=$((PASS * 100 / 10))

# Count drift events (last week)
DRIFT=$(git log --oneline --since="1 week ago" --grep="orchestrator.md" | wc -l)

# Write metrics
cat > "$METRICS_FILE" <<EOF
# Current Orchestrator Metrics
Updated: $(date -u)

## Adoption
Teams Adopted: $ADOPTED/$TOTAL ($PCT%)

## CI Validation
Pass Rate: $PASS_RATE% ($PASS/10 recent runs)
Failures: $FAIL

## Drift Events
Last Week: $DRIFT events

Generated by: collect-orchestrator-metrics.sh
EOF
```

### Alerts Configuration

Configure alerts for critical metrics:

```yaml
# .github/workflows/orchestrator-alerts.yml
on:
  workflow_run:
    workflows: [validate-orchestrators]
    types: [completed]

jobs:
  check-metrics:
    runs-on: ubuntu-latest
    if: failure()  # Alert on validation failure
    steps:
      - name: Page on-call
        run: |
          # Send Slack/PagerDuty alert
          # Include error details and remediation steps
```

## Success Criteria

Metrics indicate success when:

- [ ] Teams Adopted: 100% (all have orchestrator.yaml)
- [ ] Adoption %: 90%+ (most actively using generation)
- [ ] CI Pass Rate: 100% (zero drift reaching main)
- [ ] Generated Validity: 100%
- [ ] Support Requests: 0 (self-service working)
- [ ] Maintenance Burden: < 2 hours/month (20x reduction)
- [ ] Update Velocity: < 10 minutes average

## Metrics Review Cadence

| Frequency | Audience | Format | Owner |
|-----------|----------|--------|-------|
| Per-commit | CI | Automated checks | GitHub Actions |
| Daily | Integration team | Dashboard | Integration lead |
| Weekly | Team leads | Email report | Integration team |
| Monthly | Leadership | Executive summary | Integration team |
| Quarterly | All stakeholders | Town hall + deck | Director |

## Related Documents

- `PHASE-5-CI-VALIDATION-STRATEGY.md` - Validation rules
- `ROLLOUT-orchestrator-templates.md` - Adoption timeline
- `.github/workflows/validate-orchestrators.yml` - CI implementation

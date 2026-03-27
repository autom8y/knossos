---
title: Content Freshness Cadence Specification
initiative: content-pipeline-health
sprint: sprint-3
date: 2026-03-27
---

# Content Freshness Cadence

## Throughline

Every domain in the catalog has content within its `expires_after` threshold at container build time.

## Current State (2026-03-27)

| Repo | Total | Fresh | Stale | Oldest |
|------|-------|-------|-------|--------|
| knossos | 53 | 27 | 26 | 28d |
| autom8y | 20 | 10 | 8 | 28d |
| a8 | 18 | 11 | 6 | 22d |
| autom8y-ads | 5 | 0 | 5 | 26d |
| autom8y-data | 8 | 3 | 5 | 12d |
| autom8y-asana | 12 | 12 | 0 | 12d |
| autom8y-scheduling | 5 | 5 | 0 | 2d |
| autom8y-sms | 7 | 7 | 0 | 12d |
| **TOTAL** | **128** | **75** | **50** | |

Stale = `generated_at + expires_after < now`. Each domain's `expires_after` is set in its
frontmatter (defaults: architecture=7d, conventions=14d, literature=30d, feat=14d).

## Trigger Mechanism

### What triggers a refresh

A domain needs refresh when `generated_at + expires_after < now`. This is a per-domain
check, not a blanket schedule.

### Who triggers it

**Phase 1 (now): Manual weekly cadence**
- Every Monday, the SRE engineer runs the freshness audit:
  ```bash
  deploy/scripts/collect-content.sh --check-freshness
  ```
- If stale domains > 10: run `ari know --all` in the repos with stale domains
- After regeneration: `collect-content.sh --sync` + `build-knowledge-index.sh` + Docker rebuild

**Phase 2 (future): Automated GitHub Actions**
- Each org repo gets a weekly GitHub Actions workflow that:
  1. Runs `ari know --stale-only` (only regenerates expired domains)
  2. Commits updated `.know/` files
  3. Triggers a webhook to knossos CI that runs collect + build + deploy

### Frequency

| Domain Type | `expires_after` | Effective Cadence |
|-------------|-----------------|-------------------|
| architecture | 7d | Weekly |
| conventions | 14d | Biweekly |
| scar-tissue | 14d | Biweekly |
| design-constraints | 14d | Biweekly |
| test-coverage | 14d | Biweekly |
| feat/* | 14d | Biweekly |
| literature-* | 30d | Monthly |
| release/* | 7d | Weekly |

### Responsible Party

SRE on-call for Phase 1. CI pipeline owner for Phase 2.

## Build Pipeline Integration

The content refresh cadence feeds into the existing build pipeline:

```
1. ari know --all (in each org repo)      <- REFRESH step
2. collect-content.sh --sync --check-freshness  <- COLLECT + VALIDATE
3. build-knowledge-index.sh              <- INDEX
4. docker build + push + deploy          <- SHIP
```

The `--check-freshness` flag (new in this sprint) produces a freshness report
and exits non-zero if stale domains exceed the threshold. This gates the build
pipeline: stale content must be refreshed before shipping.

## Freshness Check Script Enhancement

Added `--check-freshness` flag to `collect-content.sh` that:
1. Parses `generated_at` and `expires_after` from each domain in `domains.yaml`
2. Reports per-repo freshness counts
3. Exits 0 if stale <= threshold (default 10), exits 1 otherwise
4. Can be used as a CI gate or manual audit tool

## Manual Refresh Execution Log

### First refresh cycle (2026-03-27)

**Before**: 50 stale domains across 5 repos.

**Action**: Regenerated .know/ files for knossos (core domains only — architecture,
conventions, design-constraints, scar-tissue, test-coverage) via `ari know`. Other
repos require running `ari know` from within each repo directory.

**Limitation**: Full org-wide refresh requires running `ari know --all` in each of
the 5 stale repos (knossos, autom8y, a8, autom8y-ads, autom8y-data). This is a
30-60 minute operation per repo due to LLM calls. Recommended for the weekly
Monday cadence, not for this sprint.

**After**: Knossos core domains refreshed (5 domains). Remaining stale: ~45.
Full refresh deferred to first Monday cadence execution.

## Automation Plan (Phase 2)

### GitHub Actions workflow (per-repo)

```yaml
# .github/workflows/know-refresh.yml
name: Knowledge Refresh
on:
  schedule:
    - cron: '0 6 * * 1'  # Every Monday at 6am UTC
  workflow_dispatch: {}

jobs:
  refresh:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install ari
        run: go install github.com/autom8y/knossos/cmd/ari@latest
      - name: Refresh stale domains
        env:
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
        run: ari know --stale-only
      - name: Commit if changed
        run: |
          git add .know/
          git diff --cached --quiet || git commit -m "chore(know): refresh stale knowledge domains"
          git push
```

### Knossos CI trigger (in knossos repo)

A webhook or scheduled workflow that:
1. Runs `collect-content.sh --sync --check-freshness`
2. Runs `build-knowledge-index.sh`
3. Runs `docker build + push + deploy`

This is the "content pipeline" end-to-end automation.

## Decision Record

- Phase 1 (manual weekly) is sufficient for current scale (128 domains, 10 repos)
- Phase 2 automation deferred until: (a) manual cadence proves burdensome, or (b) domain count exceeds 200
- The `--check-freshness` script enhancement is the minimum viable enforcement
- Full org-wide refresh not executed in this sprint — deferred to first Monday cadence

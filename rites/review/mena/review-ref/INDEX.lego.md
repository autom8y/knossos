---
name: review-ref
description: "Review rite methodology reference. Use when: implementing review agents, checking severity model, understanding health grading, finding scan heuristics, cross-rite routing targets. Triggers: review methodology, severity levels, health grades, scan heuristics, report format, artifact chain, cross-rite routing."
---

# Review Methodology Reference

## Artifact Chain

```
QUICK:  signal-sifter -> SCAN-{slug}.md -> case-reporter -> REVIEW-{slug}.md
FULL:   signal-sifter -> SCAN-{slug}.md -> pattern-profiler -> ASSESS-{slug}.md -> case-reporter -> REVIEW-{slug}.md
```

All artifacts written to `.sos/wip/review/`. Back-routes produce addenda, not full re-runs.

## Complexity Gating

| Mode | Phases | When |
|------|--------|------|
| QUICK | scan, report | "quick review", triage, specific files named |
| FULL | scan, assess, report | "full review", audit, health check, no scope specified |

QUICK: case-reporter assigns severity and grades inline from raw signals.
FULL: pattern-profiler handles all severity and grading; case-reporter uses as-is.

## Severity Model

| Level | Definition | Action |
|-------|-----------|--------|
| Critical | Blocks correctness or security | Must address before shipping |
| High | Significant maintainability risk | Address in current cycle |
| Medium | Improvement opportunity | Plan for upcoming work |
| Low | Nice-to-have cleanup | Address opportunistically |

## Health Grading Model (A-F)

| Grade | Definition | Anchor |
|-------|-----------|--------|
| A | Excellent -- best practices evident | 0 critical, 0 high, <=2 medium |
| B | Good -- minor issues, well-maintained | 0 critical, 0-1 high, <=5 medium |
| C | Adequate -- some concerns | 0 critical, 1-3 high, moderate medium |
| D | Below average -- significant concerns | 0-1 critical, 3+ high |
| F | Failing -- critical issues, immediate action | 2+ critical in category |

**Overall grade**: Weakest-link model, NOT average.
1. Start with median grade across five categories
2. Any F category -> overall cannot exceed D
3. Any D category -> overall cannot exceed C
4. 3+ categories at C or below -> overall drops one letter

## Finding Categories

| Category | Measures | Key Signals |
|----------|----------|-------------|
| Complexity | Comprehensibility burden | File sizes, nesting depth, directory file counts |
| Testing | Coverage and quality signals | Test directory presence, test-to-source ratio |
| Dependencies | Supply chain health | Dependency count, lockfile freshness, pinning |
| Structure | Architectural clarity | Directory organization, boundary clarity |
| Hygiene | Cleanliness and consistency | TODO density, naming consistency, dead code |

## Scan Heuristics (Summary)

| Signal | Threshold | Category |
|--------|-----------|----------|
| File size | >500 lines | Complexity |
| Directory depth | >4 levels | Structure |
| No test directory | absent | Testing |
| TODO/FIXME count | >10 per file | Hygiene |
| Dependency count | >100 entries | Dependencies |
| Mixed naming | camelCase + snake_case | Hygiene |
| Files per directory | >20 | Structure |
| Nesting depth | >4 indent levels | Complexity |

Full heuristic catalog with false-positive guidance: `heuristics.md` companion file.

## Cross-Rite Routing Table

| Trigger Signal | Target Rite |
|----------------|-------------|
| Auth flaws, hardcoded secrets, crypto misuse | security |
| Systemic debt patterns, accumulation trends | debt-triage |
| AI code pathologies, hallucinated imports | slop-chop |
| Code smells, naming drift, dead code | hygiene |
| Boundary violations, coupling analysis | arch |
| Missing test infrastructure, CI gaps | 10x-dev |
| Documentation gaps, stale READMEs | docs |
| Deployment concerns, observability gaps | sre |

Route by reference only. User decides whether to switch rites.

## Read-Only Enforcement (Crime Scene Protocol)

Review agents observe and report. They NEVER:
- Edit, create, or delete source files in the target codebase
- Run destructive commands (rm, git reset, etc.)
- Install or modify dependencies
- Execute test suites (read test files, don't run them)

Write access is limited to `.sos/wip/review/` artifacts only.
Enforced at four layers: prompt instructions, contract field, disallowedTools, agent-guard hook.

## Anti-Patterns

- Language-specific assumptions (must work on ANY codebase)
- Averaging health grades (use weakest-link model)
- Routing without evidence (every recommendation cites a finding)
- Fabricating findings not in upstream artifacts
- Re-evaluating severity when assessment exists (case-reporter uses pattern-profiler grades)

## Companion Files

- `templates/scan-template.md` -- signal-sifter output format
- `templates/assess-template.md` -- pattern-profiler output format
- `templates/report-template.md` -- case-reporter output format
- `heuristics.md` -- full scan heuristic catalog with thresholds and false-positive guidance

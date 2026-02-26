---
name: review-ref
description: "Review rite methodology reference. Use when: implementing review agents, checking assessment criteria, understanding artifact chain. Triggers: review methodology, severity levels, report format, scan heuristics."
---

# Review Methodology Reference

## Artifact Chain

```
scanner -> SCAN-{slug}.md -> assessor -> ASSESS-{slug}.md -> reporter -> REVIEW-{slug}.md
```

All artifacts written to `.claude/wip/review/`.

## Severity Model

| Level | Definition | Action |
|-------|-----------|--------|
| Critical | Blocks correctness or security | Must address before shipping |
| High | Significant maintainability risk | Address in current cycle |
| Medium | Improvement opportunity | Plan for upcoming work |
| Low | Nice-to-have cleanup | Address opportunistically |

## Scan Heuristics (Language-Agnostic)

| Signal | Threshold | Category |
|--------|-----------|----------|
| File size | >500 lines | Complexity |
| Directory depth | >4 levels | Structure |
| No test directory | absent | Testing |
| TODO/FIXME count | >10 per file | Hygiene |
| Dependency file size | >100 entries | Dependencies |
| Mixed naming conventions | camelCase + snake_case | Hygiene |

## Finding Categories

- **Complexity**: large files, deep nesting, high cyclomatic indicators
- **Testing**: missing tests, low test ratio, untested core paths
- **Dependencies**: large trees, outdated lockfiles, multiple managers
- **Structure**: unclear boundaries, mixed concerns, flat layouts
- **Hygiene**: TODOs, dead code, inconsistent conventions

## Read-Only Enforcement

Review agents observe and report. They NEVER:
- Edit, create, or delete source files in the target codebase
- Run destructive commands (rm, git reset, etc.)
- Install or modify dependencies
- Execute test suites (read test files, don't run them)

Write access is limited to `.claude/wip/review/` artifacts only.

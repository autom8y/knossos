---
description: "Documentation Audit Report Template companion for templates skill."
---

# Documentation Audit Report Template

> Systematic assessment of documentation health: staleness, orphans, redundancy, and gaps.

```markdown
# Documentation Audit Report
Generated: [timestamp]
Scope: [directories audited]

## Executive Summary
- Total documentation artifacts: [N]
- Current/healthy: [N] ([%])
- Stale (needs update): [N] ([%])
- Orphaned (references dead code): [N] ([%])
- Redundant (consolidation candidates): [N] pairs
- Missing (identified gaps): [N]

## Critical Issues (Immediate Attention)
[Docs that actively mislead or describe non-existent behavior]

## Staleness Report
| File | Last Updated | Related Code Changed | Staleness Score |
|------|--------------|---------------------|-----------------|
| ...  | ...          | ...                 | ...             |

## Redundancy Clusters
[Groups of docs covering the same topic]

## Gap Analysis
| Area | Expected Documentation | Status |
|------|----------------------|--------|
| ...  | ...                  | ...    |

## Recommendations
[Prioritized list of actions for Information Architect]
```

## Quality Gate

**Audit Report complete when:**
- All directories in scope scanned
- Staleness scored against code change timestamps
- Redundancy clusters identify consolidation candidates
- Gap analysis covers expected documentation per area

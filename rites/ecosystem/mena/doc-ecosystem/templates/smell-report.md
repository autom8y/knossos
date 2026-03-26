---
description: "Smell Report Template companion for templates skill."
---

# Smell Report Template

> Code smell catalog and cleanup priorities.

```markdown
# Code Smell Report
**Codebase**: [repository name]
**Analyzed**: [date]
**Scope**: [what was analyzed]

## Executive Summary
- Total smells identified: [count]
- Critical: [count] | High: [count] | Medium: [count] | Low: [count]
- Top 3 cleanup opportunities: [brief list]

## Critical Findings
[Highest priority items that should be addressed immediately]

## Category: Dead Code
### DC-001: [Specific smell]
- **Severity**: [level]
- **Location**: [file:line]
- **Pattern**: [what was found]
- **Evidence**: [why we know it's dead]
- **Blast radius**: [what's affected]

## Category: DRY Violations
[Same format]

## Category: Complexity Hotspots
[Same format]

## Category: Naming Inconsistencies
[Same format]

## Category: Import Hygiene
[Same format]

## Recommended Cleanup Order
1. [First target - why]
2. [Second target - why]
3. [Third target - why]

## Notes for Architect Enforcer
- Patterns that may indicate boundary violations: [list]
- Smells that cluster around specific modules: [list]
- Dependencies between smells (fixing X may fix Y): [list]
```

## Quality Gate

**Smell Report complete when:**
- Evidence-based findings (not speculation)
- Severity assigned to each smell
- Cleanup priority established
- Blast radius assessed for high-severity items

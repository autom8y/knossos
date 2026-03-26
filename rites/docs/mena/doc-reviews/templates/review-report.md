---
description: "Documentation Review Report Template companion for templates skill."
---

# Documentation Review Report Template

> Accuracy validation of documentation against actual code behavior, with severity-graded findings.

```markdown
# Documentation Review Report
Document: [path/to/document.md]
Reviewer: Doc Reviewer Agent
Date: [timestamp]

## Summary
- **Status:** [Approved / Needs Revision / Needs Rewrite]
- **Critical Issues:** [N]
- **Major Issues:** [N]
- **Minor Issues:** [N]

## Critical Issues
### [Issue Title]
**Location:** Line [N], Section "[Section Name]"
**Documentation states:**
> [Quoted text from doc]

**Actual behavior:**
[Description of actual behavior with code reference]
```
// Code from [file:line]
[Relevant code snippet]
```

**Suggested correction:**
> [Corrected text]

## Major Issues
[Same format as critical]

## Minor Issues
[Same format, may be briefer]

## Cross-Reference Validation
| Reference | Target | Status |
|-----------|--------|--------|
| [link text] | [target path] | Valid / Broken / Outdated |

## Code Example Validation
| Example Location | Status | Notes |
|-----------------|--------|-------|
| Line [N] | Valid / Invalid | [Details] |

## Approval Status
[ ] Approved for publication
[ ] Approved with minor corrections (can be fixed post-publish)
[ ] Requires revision before publication
[ ] Requires significant rewrite
```

## Quality Gate

**Review Report complete when:**
- All issues categorized by severity (Critical/Major/Minor)
- Code references validated against actual behavior
- Cross-references checked for broken links
- Approval status selected with clear rationale

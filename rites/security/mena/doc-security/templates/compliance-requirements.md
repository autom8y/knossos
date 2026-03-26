---
description: "Compliance Requirements Template companion for templates skill."
---

# Compliance Requirements Template

> Regulatory mapping with control implementation, gap analysis, and evidence collection plan.

```markdown
# COMPLY-{slug}

## Overview
{What feature/system and which regulations}

## Applicable Regulations

### {Regulation 1, e.g., SOC 2}
- **Relevant Criteria**: {CC6.1, CC7.2, etc.}
- **Scope**: {What's covered}

### {Regulation 2, e.g., GDPR}
- **Relevant Articles**: {Art. 6, Art. 32, etc.}
- **Scope**: {What's covered}

## Control Requirements

### {Control Category, e.g., Access Control}

#### CTRL-001: {Control Name}
- **Regulation**: {Source requirement}
- **Requirement**: {What must be true}
- **Implementation**: {How to achieve}
- **Evidence**: {What proves compliance}
- **Testing**: {How to validate}

#### CTRL-002: {Control Name}
...

## Data Classification
| Data Element | Classification | Retention | Encryption |
|--------------|---------------|-----------|------------|
| {element} | {PII/Sensitive/Public} | {period} | {at-rest/in-transit} |

## Gap Analysis
| Control | Current State | Gap | Remediation | Priority |
|---------|--------------|-----|-------------|----------|
| {control} | {state} | {gap} | {fix} | {P1/P2/P3} |

## Implementation Checklist
- [ ] {Requirement 1}
- [ ] {Requirement 2}

## Evidence Collection
| Control | Evidence Type | Collection Method | Frequency |
|---------|--------------|-------------------|-----------|
| {control} | {logs/configs/screenshots} | {automated/manual} | {continuous/quarterly} |

## Audit Readiness
{Steps to prepare for audit}
```

## Quality Gate

**Compliance Requirements complete when:**
- All applicable regulations mapped to specific controls
- Each control has implementation and evidence plan
- Gap analysis includes remediation with priority
- Evidence collection frequency defined per control

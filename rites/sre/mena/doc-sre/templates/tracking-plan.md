---
description: "Tracking Plan Template companion for templates skill."
---

# Tracking Plan Template

> Analytics event specification with naming conventions, properties, and validation rules.

```markdown
# TRACK-{slug}

## Overview
{What user journey or feature this tracks}

## Business Questions
- {Question 1 this data answers}
- {Question 2}

## Naming Convention
{event_category_action, e.g., onboarding_step_completed}

## Events

### {event_name}
- **Trigger**: {When this event fires}
- **Category**: {Funnel step, engagement, error, etc.}
- **Platform**: {Web, iOS, Android, Server}

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| {property} | {string/int/bool} | {Yes/No} | {What it represents} |

### Validation Rules
- {Rule 1, e.g., "step_number must be 1-5"}
- {Rule 2}

## Implementation Notes
{Code examples, edge cases, gotchas}

## QA Checklist
- [ ] Events fire on expected triggers
- [ ] All required properties present
- [ ] Property values within expected ranges
- [ ] No duplicate events
- [ ] Works across platforms
```

## Quality Gate

**Tracking Plan complete when:**
- Business questions documented (what decisions this data enables)
- Every event has trigger, category, and platform
- Property types and required flags specified
- Validation rules cover edge cases

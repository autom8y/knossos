---
description: "Skill Anti-Patterns companion for agent-prompt-engineering skill."
---

# Skill Anti-Patterns

> 4 common mistakes when using this skill

These anti-patterns apply to skill users, not agent authors. For agent authoring mistakes, see [principles.md](principles.md) (anti-patterns integrated into each principle).

---

## 1. Skipping Validation

**Symptom**: Agents deployed with failing checklist items. Production issues traced to skipped verification.

**Detection**: No evidence of [validation/checklist.md](validation/checklist.md) execution in PR.

**Why it fails**: Validation catches 80% of issues. Skipping it trades 10 minutes of checking for hours of debugging.

**Fix**: Run full validation checklist before every deployment. Add checklist output to PR description.

---

## 2. Template Cargo-Culting

**Symptom**: Agent prompts look structurally correct but contain placeholder text or irrelevant sections.

**Detection**: Template brackets remain (`{agent-name}`). Sections copied verbatim from other agents without adaptation.

**Why it fails**: Structure without substance. Agent inherits another agent's domain authority, anti-patterns, or workflow position.

**Fix**: Treat template as scaffold, not fill-in-the-blank. Every section must be written fresh for this agent's domain.

---

## 3. Ignoring Rubric Low Scores

**Symptom**: Agents deployed with rubric scores below 4. Known weaknesses accepted as "good enough."

**Detection**: Rubric assessment shows 3 or lower on any dimension. Deployment proceeds anyway.

**Why it fails**: Each rubric dimension correlates with specific failure modes. Score of 3 on "Boundary Clarity" predicts escalation failures.

**Fix**: All 6 dimensions must score 4+ before deployment. Low scores require revision, not acceptance.

---

## 4. Over-Auditing Simple Agents

**Symptom**: Full rubric assessment, multi-reviewer sign-off, and extensive testing for trivial agents.

**Detection**: Agent under 50 lines. Single responsibility. Still receives full audit treatment.

**Why it fails**: Audit overhead exceeds agent value. Creates friction that discourages agent creation.

**Fix**: Use Quick Validation (7 items) for simple agents. Reserve full validation for complex agents (100+ lines, multiple responsibilities).

---

## Escalation Paths

When issues are not prompt-related, route appropriately:

| Issue Type | Route To |
|------------|----------|
| Agent performance (slowness, timeouts) | Infrastructure team |
| Tool capability limitations | Platform team via feature request |
| Model behavior inconsistencies | `/consult` for triage |
| Cross-agent coordination failures | Orchestrator or workflow owner |
| Skill content gaps | Skill maintainer via PR |

**When in doubt**: Use `/consult` to get routing guidance before attempting fixes.

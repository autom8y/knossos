---
name: hygiene-ref
description: "Switch to hygiene (code quality workflow). Triggers: /hygiene, code quality, refactoring team, quality audit, code cleanup."
---

# /hygiene - Quick Switch to Code Hygiene Team

> **Category**: Team Management | **Phase**: Team Switching

## Purpose

Instantly switch to the hygiene, a specialized team focused on code quality, architectural compliance, refactoring, and technical cleanliness. This team detects code smells, enforces standards, and cleans up technical messes.

This is a convenience wrapper around `/rite hygiene` that also displays the pantheon after switching.

---

## Usage

```bash
/hygiene
```

No parameters required. This command:
1. Switches to hygiene
2. Displays team roster with agent descriptions

---

## Behavior

### 1. Invoke Team Switch

Execute via Bash tool:

```bash
$ROSTER_HOME/swap-rite.sh hygiene
```

### 2. Display Pantheon

After successful switch, show the active pantheon:

```
Switched to hygiene (4 agents loaded)

Pantheon:
┌─────────────────────────┬──────────────────────────────────────────────┐
│ Agent                   │ Role                                         │
├─────────────────────────┼──────────────────────────────────────────────┤
│ code-smeller            │ Detects code smells and anti-patterns        │
│ architect-enforcer      │ Validates architectural compliance           │
│ janitor                 │ Cleans up code, refactors for quality        │
│ audit-lead              │ Conducts comprehensive quality audits        │
└─────────────────────────┴──────────────────────────────────────────────┘

Use /handoff <agent> to delegate work.
```

### 3. Update SESSION_CONTEXT (if active)

If a session is active:
- Update `active_team` field to `hygiene`
- Add handoff note documenting team switch

---

## Team Details

**Team Name**: hygiene
**Agent Count**: 4
**Workflow**: Detect → Audit → Enforce → Clean

| Agent | Role | Model |
|-------|------|-------|
| code-smeller | Code smell detection and anti-pattern identification | Sonnet |
| architect-enforcer | Architectural compliance validation | Sonnet |
| janitor | Code cleanup and refactoring execution | Sonnet |
| audit-lead | Comprehensive quality audit coordination | Opus |

---

## When to Use Hygiene Team

Use this team for:

- **Code quality audits**: Regular health checks
- **Refactoring initiatives**: Cleaning up technical mess
- **Architecture compliance**: Enforcing design decisions
- **Pre-release cleanup**: Quality gates before shipping
- **Onboarding prep**: Making codebase cleaner for new devs
- **Post-implementation cleanup**: After rapid prototyping
- **Complexity reduction**: Simplifying overgrown code

**Don't use for**:
- New feature implementation → Use `/10x` instead
- Documentation → Use `/docs` instead
- Debt assessment (use `/debt` for planning, hygiene for execution)

---

## Hygiene vs Debt Teams

| Hygiene Team | Debt Team |
|--------------|-----------|
| **Focus**: Code quality and cleanliness | **Focus**: Technical debt prioritization |
| **Action**: Detect and fix issues | **Action**: Assess and plan remediation |
| **Scope**: Code-level refactoring | **Scope**: Project/portfolio-level debt |
| **Agents**: Smeller, Enforcer, Janitor, Audit Lead | **Agents**: Collector, Assessor, Planner |
| **Output**: Clean code, refactorings | **Output**: Debt inventory, roadmaps |

**Workflow**: Use `/debt` to plan, `/hygiene` to execute.

---

## Progressive Disclosure

- [agents.lego.md](agents.lego.md) - Detailed agent profiles, capabilities, and detection patterns
- [workflow-examples.lego.md](workflow-examples.lego.md) - Usage examples, typical workflow phases, state changes, and operational notes

## Related

- [standards](../../../../mena/guidance/standards/INDEX.lego.md) - Code conventions and quality rules
- [10x-workflow](../../../10x-dev/mena/10x-workflow/INDEX.lego.md) - Agent coordination patterns
- [COMMAND_REGISTRY.md](../../COMMAND_REGISTRY.md) - All registered commands

---

## Related Commands

- `/team` - General rite switching with options
- `/10x` - Quick switch to development team
- `/docs` - Quick switch to documentation team
- `/debt` - Quick switch to technical debt team
- `/handoff` - Delegate to specific agent in current team

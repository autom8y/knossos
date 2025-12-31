---
name: team-development
description: "Design and implement agent team packs for the roster ecosystem. Use when: creating new teams, designing workflows, writing agent prompts, integrating with slash commands. Triggers: new team, team pack, workflow design, agent creation, roster integration."
---

# Team Development

> Design agent teams that work. Build workflows that flow.

This skill codifies the patterns discovered from building 5 teams (10x-dev-pack, doc-team-pack, hygiene-pack, debt-triage-pack, sre-pack) into reusable templates and decision frameworks.

---

## Quick Reference

| Component | Location | Key Decisions |
|-----------|----------|---------------|
| Team Pack | `$ROSTER_HOME/teams/{name}-pack/` | Name, agent count, domain |
| Workflow | `workflow.yaml` | Phases, complexity levels, entry point |
| Agents | `agents/*.md` | Role, model, color, tools |
| Command | `.claude/commands/{name}.md` | Quick-switch integration |
| Skill | `.claude/skills/{name}-ref/` | Reference documentation |

---

## Team Creation Checklist

```
1. [ ] Define team domain and purpose
2. [ ] Design workflow phases (3-4 typical)
3. [ ] Identify agent roles (3-5 agents)
4. [ ] Create workflow.yaml
5. [ ] Write agent prompts (use template)
6. [ ] Create quick-switch command
7. [ ] Create reference skill
8. [ ] Update COMMAND_REGISTRY.md
9. [ ] Validate with swap-team.sh
10. [ ] **Update Consultant knowledge base** (REQUIRED)
```

> **CRITICAL**: Step 10 ensures the Consultant agent stays canonical. The Consultant is the ecosystem's navigation system - if it has stale data, users will get wrong guidance.

---

## Decision Frameworks

### How Many Agents?

| Count | Team Type | Examples |
|-------|-----------|----------|
| 3 | Focused/specialized | debt-triage-pack |
| 4 | Standard teams | doc-team, hygiene, sre |
| 5 | Full lifecycle | 10x-dev-pack |

### Model and Color Assignment

See [@agent-prompt-engineering](../agent-prompt-engineering/SKILL.md#model-and-color-assignment) for model selection and color assignment guidance.

### Workflow Phases

| Phase Position | Role | Produces |
|----------------|------|----------|
| Entry | Assessment/Discovery | Report, Audit, Requirements |
| Design | Planning/Architecture | Plan, Design, TDD |
| Execute | Implementation | Code, Content, Changes |
| Validate | Testing/Review | Signoff, Report |

---

## Progressive Disclosure

### Glossary

| Term | Definition |
|------|------------|
| **Team Pack** | Directory containing agents and workflow for a specialized domain |
| **Workflow** | Sequential pipeline of phases producing artifacts |
| **Phase** | Single step in workflow, owned by one agent |
| **Agent** | Specialized prompt with defined role, tools, and authority |
| **Artifact** | Document produced by a phase (PRD, TDD, report, etc.) |
| **Complexity Level** | Scope classifier that determines which phases run |

Detailed definitions: [glossary/agents.md](glossary/agents.md) | [glossary/artifacts.md](glossary/artifacts.md)

### Patterns
Codified design patterns:
- [patterns/team-composition.md](patterns/team-composition.md) - 3/4/5-agent patterns
- [patterns/phase-sequencing.md](patterns/phase-sequencing.md) - Sequential workflow design
- [patterns/complexity-gating.md](patterns/complexity-gating.md) - Complexity levels
- [patterns/command-mapping.md](patterns/command-mapping.md) - Slash command integration

### Templates
Copy-and-fill templates:
- [templates/workflow.yaml.template](templates/workflow.yaml.template) - Workflow config
- [templates/agent-template.md](templates/agent-template.md) - Agent prompt (11 sections)
- [templates/quick-switch.md.template](templates/quick-switch.md.template) - Team command
- [templates/skill-ref.md.template](templates/skill-ref.md.template) - Reference skill

### Validation
Pre-flight checks and troubleshooting:
- [validation/validation.md](validation/validation.md) - Checklist and common issues

### Examples
Complete team implementations:
- [examples/examples.md](examples/examples.md) - 3-agent and 5-agent team patterns

---

## Existing Teams Reference

| Team | Agents | Workflow | Entry Agent |
|------|--------|----------|-------------|
| 10x-dev-pack | 5 | Requirements → Design → Implementation → Validation | requirements-analyst |
| doc-team-pack | 4 | Audit → Architecture → Writing → Review | doc-auditor |
| hygiene-pack | 4 | Assessment → Planning → Execution → Audit | code-smeller |
| debt-triage-pack | 3 | Collection → Assessment → Planning | debt-collector |
| sre-pack | 4 | Observation → Coordination → Implementation → Resilience | observability-engineer |
| security-pack | 4 | Threat Modeling → Compliance → Testing → Review | threat-modeler |
| intelligence-pack | 4 | Instrumentation → Research → Experimentation → Synthesis | analytics-engineer |
| rnd-pack | 4 | Scouting → Integration → Prototyping → Future Architecture | technology-scout |
| strategy-pack | 4 | Market Research → Competitive Analysis → Business Modeling → Planning | market-researcher |

*See roster repository for current team/agent counts.*

---

## Cross-Skill Integration

- @10x-workflow for workflow mechanics and phase transitions
- @documentation for artifact templates (PRD, TDD, ADR)
- @standards for naming conventions and code patterns
- @prompting for agent invocation patterns
- **@consult-ref for ecosystem navigation** (MUST update when adding teams)

---

## Consultant Synchronization (REQUIRED)

> **CRITICAL**: The Consultant agent is the ecosystem's navigation system. Stale data = wrong user guidance.

When creating or modifying teams, update the Consultant's knowledge base. See [patterns/consultant-sync.md](patterns/consultant-sync.md) for:
- Synchronization matrix (what files to update for each change type)
- Step-by-step procedures for adding teams, agents, and playbooks
- Verification commands
- Common issues and fixes

---

## Quick Start

To create a new team:

```bash
# 1. Create directory structure
mkdir -p $ROSTER_HOME/teams/{name}-pack/agents

# 2. Copy and fill templates
# - workflow.yaml from templates/workflow.yaml.template
# - agent files from templates/agent-template.md

# 3. Create command and skill
# - .claude/commands/{name}.md
# - .claude/skills/{name}-ref/skill.md

# 4. Update registry
# - Add to COMMAND_REGISTRY.md

# 5. Validate
$ROSTER_HOME/swap-team.sh {name}-pack
```

See [validation/validation.md](validation/validation.md) for full pre-flight checks.

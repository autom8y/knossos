---
name: rite-development
description: "Design and implement agent rites for the roster ecosystem. Use when: creating new rites, designing workflows, writing agent prompts, integrating with slash commands. Triggers: new rite, rite development, workflow design, agent creation, roster integration."
---

# Team Development

> Design agent rites that work. Build workflows that flow.

This skill codifies the patterns discovered from building rites (10x-dev-pack, doc-rite-pack, hygiene-pack, debt-triage-pack, sre-pack) into reusable templates and decision frameworks.

---

## Quick Reference

| Component | Location | Key Decisions |
|-----------|----------|---------------|
| Rite | `$ROSTER_HOME/rites/{name}/` | Name, agent count, domain |
| Workflow | `workflow.yaml` | Phases, complexity levels, entry point |
| Agents | `agents/*.md` | Role, model, color, tools |
| Command | `.claude/commands/{name}.md` | Quick-switch integration |
| Skill | `.claude/skills/{name}-ref/` | Reference documentation |

---

## Rite Creation Checklist

```
1. [ ] Define rite domain and purpose
2. [ ] Design workflow phases (3-4 typical)
3. [ ] Identify agent roles (3-5 agents)
4. [ ] Create workflow.yaml
5. [ ] Write agent prompts (use template)
6. [ ] Create quick-switch command
7. [ ] Create reference skill
8. [ ] Update COMMAND_REGISTRY.md
9. [ ] Validate with swap-rite.sh
10. [ ] **Update Consultant knowledge base** (REQUIRED)
```

> **CRITICAL**: Step 10 ensures the Consultant agent stays canonical. The Consultant is the ecosystem's navigation system - if it has stale data, users will get wrong guidance.

---

## Decision Frameworks

### How Many Agents?

| Count | Rite Type | Examples |
|-------|-----------|----------|
| 3 | Focused/specialized | debt-triage-pack |
| 4 | Standard rites | doc-team, hygiene, sre |
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
| **Rite** | Directory containing agents and workflow for a specialized domain |
| **Workflow** | Sequential pipeline of phases producing artifacts |
| **Phase** | Single step in workflow, owned by one agent |
| **Agent** | Specialized prompt with defined role, tools, and authority |
| **Artifact** | Document produced by a phase (PRD, TDD, report, etc.) |
| **Complexity Level** | Scope classifier that determines which phases run |

Detailed definitions: [glossary/agents.md](glossary/agents.md) | [glossary/artifacts.md](glossary/artifacts.md)

### Patterns
Codified design patterns:
- [patterns/rite-composition.md](patterns/rite-composition.md) - 3/4/5-agent patterns
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

## Existing Rites Reference

| Rite | Agents | Workflow | Entry Agent |
|------|--------|----------|-------------|
| 10x-dev-pack | 5 | Requirements → Design → Implementation → Validation | requirements-analyst |
| doc-rite-pack | 4 | Audit → Architecture → Writing → Review | doc-auditor |
| hygiene-pack | 4 | Assessment → Planning → Execution → Audit | code-smeller |
| debt-triage-pack | 3 | Collection → Assessment → Planning | debt-collector |
| sre-pack | 4 | Observation → Coordination → Implementation → Resilience | observability-engineer |
| security-pack | 4 | Threat Modeling → Compliance → Testing → Review | threat-modeler |
| intelligence-pack | 4 | Instrumentation → Research → Experimentation → Synthesis | analytics-engineer |
| rnd-pack | 4 | Scouting → Integration → Prototyping → Future Architecture | technology-scout |
| strategy-pack | 4 | Market Research → Competitive Analysis → Business Modeling → Planning | market-researcher |

*See roster repository for current rite/agent counts.*

---

## Cross-Skill Integration

- @10x-workflow for workflow mechanics and phase transitions
- @documentation for artifact templates (PRD, TDD, ADR)
- @standards for naming conventions and code patterns
- @prompting for agent invocation patterns
- **@consult-ref for ecosystem navigation** (MUST update when adding rites)

---

## Consultant Synchronization (REQUIRED)

> **CRITICAL**: The Consultant agent is the ecosystem's navigation system. Stale data = wrong user guidance.

When creating or modifying rites, update the Consultant's knowledge base. See [patterns/consultant-sync.md](patterns/consultant-sync.md) for:
- Synchronization matrix (what files to update for each change type)
- Step-by-step procedures for adding rites, agents, and playbooks
- Verification commands
- Common issues and fixes

---

## Quick Start

To create a new rite:

```bash
# 1. Create directory structure
mkdir -p $ROSTER_HOME/rites/{name}/agents

# 2. Copy and fill templates
# - workflow.yaml from templates/workflow.yaml.template
# - agent files from templates/agent-template.md

# 3. Create command and skill
# - .claude/commands/{name}.md
# - .claude/skills/{name}-ref/skill.md

# 4. Update registry
# - Add to COMMAND_REGISTRY.md

# 5. Validate
$ROSTER_HOME/swap-rite.sh {name}
```

See [validation/validation.md](validation/validation.md) for full pre-flight checks.

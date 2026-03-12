---
name: rite-development
description: "Rite design and implementation guide. Use when: building a new rite from scratch, designing workflow phases, creating agent pantheons, integrating with knossos. Triggers: new rite, rite development, workflow design, agent creation, knossos integration."
---

# Rite Development

> Design rites that work. Build workflows that flow.

This skill codifies the patterns discovered from building rites (10x-dev, docs, hygiene, debt-triage, sre) into reusable templates and decision frameworks.

---

## Quick Reference

| Component | Location | Key Decisions |
|-----------|----------|---------------|
| Rite | `$KNOSSOS_HOME/rites/{name}/` | Name, agent count, domain |
| Workflow | `workflow.yaml` | Phases, complexity levels, entry point |
| Agents | `agents/*.md` | Role, model, color, tools |
| Command | channel commands directory `{name}.md` | Quick-switch integration |
| Skill | `{name}-ref/` (in rite mena) | Reference documentation |

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
8. [ ] Run `ari sync --rite {rite-name}`
9. [ ] Validate with ari sync --rite
10. [ ] **Update Pythia knowledge base** (REQUIRED)
```

> **CRITICAL**: Step 10 ensures the Pythia agent stays canonical. The Pythia is the ecosystem's navigation system - if it has stale data, users will get wrong guidance.

---

## Decision Frameworks

### How Many Agents?

| Count | Rite Type | Examples |
|-------|-----------|----------|
| 3 | Focused/specialized | debt-triage |
| 4 | Standard rites | doc-rite, hygiene, sre |
| 5 | Full lifecycle | 10x-dev |

### Model and Color Assignment

See [agent-prompt-engineering](../agent-prompt-engineering/INDEX.md#model-and-color-assignment) for model selection and color assignment guidance.

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
- [templates/quick-switch.md.template](templates/quick-switch.md.template) - Rite command
- [templates/skill-ref.md.template](templates/skill-ref.md.template) - Reference skill

### Validation
Pre-flight checks and troubleshooting:
- [validation/validation.md](validation/validation.md) - Checklist and common issues

### Platform Reference
- [platform-artifacts.md](platform-artifacts.md) - Knossos structure, ari sync reference, verification commands

### Examples
Complete rite implementations:
- [examples/examples.md](examples/examples.md) - 3-agent and 5-agent rite patterns

---

## Existing Rites Reference

| Rite | Agents | Workflow | Entry Agent |
|------|--------|----------|-------------|
| 10x-dev | 5 | Requirements → Design → Implementation → Validation | requirements-analyst |
| docs | 4 | Audit → Architecture → Writing → Review | doc-auditor |
| hygiene | 4 | Assessment → Planning → Execution → Audit | code-smeller |
| debt-triage | 3 | Collection → Assessment → Planning | debt-collector |
| sre | 4 | Observation → Coordination → Implementation → Resilience | observability-engineer |
| security | 4 | Threat Modeling → Compliance → Testing → Review | threat-modeler |
| intelligence | 4 | Instrumentation → Research → Experimentation → Synthesis | analytics-engineer |
| rnd | 4 | Scouting → Integration → Prototyping → Future Architecture | technology-scout |
| strategy | 4 | Market Research → Competitive Analysis → Business Modeling → Planning | market-researcher |

*See knossos rites/ directory for current rite/agent counts.*

---

## Cross-Skill Integration

- 10x-workflow for workflow mechanics and phase transitions
- documentation for artifact templates (PRD, TDD, ADR)
- standards for naming conventions and code patterns
- prompting for agent invocation patterns
- **consult for ecosystem navigation** (MUST update when adding rites)

---

## Pythia Synchronization (REQUIRED)

> **CRITICAL**: The Pythia agent is the ecosystem's navigation system. Stale data = wrong user guidance.

When creating or modifying rites, update the Pythia's knowledge base. See [patterns/pythia-sync.md](patterns/pythia-sync.md) for:
- Synchronization matrix (what files to update for each change type)
- Step-by-step procedures for adding rites, agents, and playbooks
- Verification commands
- Common issues and fixes

---

## Quick Start

To create a new rite:

```bash
# 1. Create directory structure
mkdir -p $KNOSSOS_HOME/rites/{name}/agents

# 2. Copy and fill templates
# - workflow.yaml from templates/workflow.yaml.template
# - agent files from templates/agent-template.md

# 3. Create command and skill
# - {channel_dir}/commands/{name}.md
# - rites/{name}/mena/{name}-ref/INDEX.lego.md

# 4. Sync to project
# - ari sync --rite {name}

# 5. Validate
$KNOSSOS_HOME/ari sync --rite {name}
```

See [validation/validation.md](validation/validation.md) for full pre-flight checks.

---
name: intelligence-ref
description: "Product Intelligence rite reference with insights artifacts. Use when: activating the intelligence rite, invoking analytics or research agents, planning experiments or user studies, producing insights reports or HANDOFF artifacts. Triggers: intelligence, analytics-engineer, user-researcher, experimentation-lead, insights-analyst, insights report, findings format."
---

# Product Intelligence Rite (intelligence)

> Instrument. Research. Experiment. Synthesize.

## Quick Reference

| Component | Location | Purpose |
|-----------|----------|---------|
| Agents | `$KNOSSOS_HOME/rites/intelligence/agents/` | Agent prompts |
| Workflow | `$KNOSSOS_HOME/rites/intelligence/workflow.yaml` | Phase configuration |
| Switch | `/intelligence` | Activate this rite |

## Pantheon

| Agent | Model | Role | Produces |
|-------|-------|------|----------|
| **analytics-engineer** | opus | Builds data foundation and tracking | tracking-plan |
| **user-researcher** | opus | Captures qualitative insights | research-findings |
| **experimentation-lead** | opus | Designs A/B tests and experiments | experiment-design |
| **insights-analyst** | opus | Synthesizes data into decisions | insights-report |

## Workflow

```
instrumentation → research → experimentation → synthesis
        │            │              │              │
        ▼            ▼              ▼              ▼
  TRACK-{slug}  RESEARCH-{slug}  EXPERIMENT-{slug}  INSIGHT-{slug}
```

### Phase Details

| Phase | Agent | Input | Output |
|-------|-------|-------|--------|
| instrumentation | analytics-engineer | Feature specs | Tracking plan with events |
| research | user-researcher | Tracking plan | Qualitative findings |
| experimentation | experimentation-lead | Research findings | Experiment design |
| synthesis | insights-analyst | Experiment results | Actionable insights |

## Complexity Levels

| Level | Scope | Phases |
|-------|-------|--------|
| **METRIC** | Single metric, existing events | experimentation, synthesis |
| **FEATURE** | New feature instrumentation | All phases |
| **INITIATIVE** | Cross-feature analysis | All phases |

## Command Mapping

| Command | Maps To | Use When |
|---------|---------|----------|
| `/intelligence` | Rite switch | Activating this rite |
| `/architect` | experimentation-lead | Experiment design only |
| `/build` | analytics-engineer | Instrumentation only |
| `/qa` | insights-analyst | Analysis only |
| `/hotfix` | N/A | Not applicable (research rite) |
| `/code-review` | N/A | Not applicable (research rite) |

## When to Use This Rite

**Use intelligence when:**
- Instrumenting a new feature
- Running A/B tests
- Conducting user research
- Making data-driven decisions

**Don't use intelligence when:**
- Building features → Use 10x-dev
- Security analysis → Use security
- Market research → Use strategy

## Agent Summaries

### Analytics Engineer

**Purpose**: Build the data foundation

**Key Responsibilities**:
- Event taxonomy design
- Tracking plan development
- Data quality assurance
- Pipeline architecture

**Produces**: `.ledge/spikes/TRACK-{slug}.md`

---

### User Researcher

**Purpose**: Capture the qualitative 'why'

**Key Responsibilities**:
- Research design
- User interviews
- Usability testing
- Synthesis

**Produces**: `.ledge/spikes/RESEARCH-{slug}.md`

---

### Experimentation Lead

**Purpose**: Run the scientific method on product

**Key Responsibilities**:
- Hypothesis formation
- A/B test design
- Sample size calculation
- Statistical rigor

**Produces**: `.ledge/spikes/EXPERIMENT-{slug}.md`

---

### Insights Analyst

**Purpose**: Turn data into decisions

**Key Responsibilities**:
- Result interpretation
- Story building
- Recommendation synthesis
- Stakeholder communication

**Produces**: `.ledge/spikes/INSIGHT-{slug}.md`

## Supporting Files

- [insights-artifacts.md](insights-artifacts.md) - HANDOFF templates, findings format, insights report guidance (insights-analyst)

## Cross-References

- **Related Skills**: documentation (artifact templates)
- **Related Rites**: 10x-dev (implementation), strategy (strategic context)
- **Commands**: Run `ari rite --list` or `/consult --commands`

---
name: intelligence-ref
description: "Product Intelligence team reference with analytics, user research, and experimentation workflows. Use when: learning about intelligence agents, understanding the analytics workflow, invoking intelligence agents. Triggers: intelligence, product intelligence, analytics-engineer, user-researcher, experimentation-lead, insights-analyst."
---

# Product Intelligence Team (intelligence)

> Instrument. Research. Experiment. Synthesize.

## Quick Reference

| Component | Location | Purpose |
|-----------|----------|---------|
| Agents | `$ROSTER_HOME/rites/intelligence/agents/` | Agent prompts |
| Workflow | `$ROSTER_HOME/rites/intelligence/workflow.yaml` | Phase configuration |
| Switch | `/intelligence` | Activate this team |

## Pantheon

| Agent | Model | Role | Produces |
|-------|-------|------|----------|
| **analytics-engineer** | sonnet | Builds data foundation and tracking | tracking-plan |
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
| `/intelligence` | Team switch | Activating this team |
| `/architect` | experimentation-lead | Experiment design only |
| `/build` | analytics-engineer | Instrumentation only |
| `/qa` | insights-analyst | Analysis only |
| `/hotfix` | N/A | Not applicable (research team) |
| `/code-review` | N/A | Not applicable (research team) |

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

**Produces**: `docs/intelligence/TRACK-{slug}.md`

---

### User Researcher

**Purpose**: Capture the qualitative 'why'

**Key Responsibilities**:
- Research design
- User interviews
- Usability testing
- Synthesis

**Produces**: `docs/intelligence/RESEARCH-{slug}.md`

---

### Experimentation Lead

**Purpose**: Run the scientific method on product

**Key Responsibilities**:
- Hypothesis formation
- A/B test design
- Sample size calculation
- Statistical rigor

**Produces**: `docs/intelligence/EXPERIMENT-{slug}.md`

---

### Insights Analyst

**Purpose**: Turn data into decisions

**Key Responsibilities**:
- Result interpretation
- Story building
- Recommendation synthesis
- Stakeholder communication

**Produces**: `docs/intelligence/INSIGHT-{slug}.md`

## Cross-References

- **Related Skills**: @documentation (artifact templates)
- **Related Rites**: 10x-dev (implementation), strategy (strategic context)
- **Commands**: See COMMAND_REGISTRY.md

# Comprehensive Model Inventory Report

**Generated**: 2025-12-30
**Scope**: All agents, commands, hooks, and skills across roster ecosystem
**Purpose**: Baseline for model upgrade planning

---

## Executive Summary

| Category | Total | Opus | Sonnet | Haiku |
|----------|-------|------|--------|-------|
| **Agents** | 60 | 43 (72%) | 16 (27%) | 0 |
| **Commands** | 44 | 19 (43%) | 25 (57%) | 0 |
| **Total** | 104 | 62 (60%) | 41 (39%) | 0 |

**Key Finding**: No haiku models in use. Distribution follows complexity pattern - opus for reasoning-heavy, sonnet for execution.

---

## Part 1: Agents by Team

### 10x-Dev-Pack (5 agents)

| Agent | Model | Color | Role |
|-------|-------|-------|------|
| architect | opus | cyan | System design and technical decisions |
| orchestrator | opus | blue | Coordination of feature development phases |
| principal-engineer | opus | green | Code implementation and technical execution |
| qa-adversary | opus | red | Quality assurance and adversarial testing |
| requirements-analyst | opus | orange | Feature specifications and requirements |

### Doc-Team-Pack (5 agents)

| Agent | Model | Color | Role |
|-------|-------|-------|------|
| doc-auditor | sonnet | blue | Documentation quality assessment |
| doc-reviewer | sonnet | red | Documentation review and validation |
| information-architect | opus | cyan | Documentation structure and organization |
| orchestrator | opus | green | Documentation workflow coordination |
| tech-writer | sonnet | blue | Technical documentation creation |

### Debt-Triage-Pack (4 agents)

| Agent | Model | Color | Role |
|-------|-------|-------|------|
| debt-collector | sonnet | orange | Technical debt identification |
| orchestrator | opus | orange | Debt assessment workflow coordination |
| risk-assessor | opus | yellow | Risk evaluation and prioritization |
| sprint-planner | opus | magenta | Debt resolution sprint planning |

### Ecosystem-Pack (6 agents)

| Agent | Model | Color | Role |
|-------|-------|-------|------|
| compatibility-tester | opus | red | Ecosystem compatibility validation |
| context-architect | opus | cyan | CEM/skeleton/roster schema design |
| documentation-engineer | sonnet | magenta | Migration and API documentation |
| ecosystem-analyst | opus | orange | Root cause analysis for ecosystem issues |
| integration-engineer | sonnet | green | Ecosystem infrastructure implementation |
| orchestrator | opus | purple | Ecosystem-pack phase coordination |

### Forge-Pack (7 agents)

| Agent | Model | Color | Role |
|-------|-------|-------|------|
| agent-curator | sonnet | blue | Agent selection and composition |
| agent-designer | opus | purple | Agent specification and role definition |
| eval-specialist | opus | red | Team evaluation and validation |
| orchestrator | opus | cyan | Forge workflow coordination |
| platform-engineer | sonnet | orange | Roster infrastructure implementation |
| prompt-architect | opus | cyan | System prompt crafting and optimization |
| workflow-engineer | opus | green | Agent orchestration and workflow design |

### Hygiene-Pack (5 agents)

| Agent | Model | Color | Role |
|-------|-------|-------|------|
| architect-enforcer | opus | cyan | Refactoring contract and plan design |
| audit-lead | opus | red | Code cleanup verification and signoff |
| code-smeller | opus | orange | Code quality issue detection |
| janitor | sonnet | green | Refactoring execution with atomic commits |
| orchestrator | opus | green | Code quality workflow coordination |

### Intelligence-Pack (5 agents)

| Agent | Model | Color | Role |
|-------|-------|-------|------|
| analytics-engineer | sonnet | orange | Event tracking and data pipeline design |
| experimentation-lead | opus | cyan | A/B test design and hypothesis formation |
| insights-analyst | opus | purple | Data synthesis and recommendations |
| orchestrator | opus | cyan | Analytics workflow coordination |
| user-researcher | opus | magenta | User research and qualitative insights |

### R&D-Pack (5 agents)

| Agent | Model | Color | Role |
|-------|-------|-------|------|
| integration-researcher | opus | cyan | Integration dependency mapping |
| moonshot-architect | opus | purple | Long-term architecture planning |
| orchestrator | opus | purple | R&D workflow coordination |
| prototype-engineer | sonnet | green | Rapid prototyping and feasibility |
| technology-scout | opus | orange | Technology evaluation |

### Security-Pack (5 agents)

| Agent | Model | Color | Role |
|-------|-------|-------|------|
| compliance-architect | opus | cyan | Regulatory requirements translation |
| orchestrator | opus | red | Security workflow coordination |
| penetration-tester | opus | green | Vulnerability discovery and testing |
| security-reviewer | opus | red | Security code review and merge gate |
| threat-modeler | opus | orange | Threat analysis (STRIDE/DREAD) |

### SRE-Pack (5 agents)

| Agent | Model | Color | Role |
|-------|-------|-------|------|
| chaos-engineer | opus | red | Resilience testing and failure injection |
| incident-commander | opus | purple | Incident coordination and postmortem |
| observability-engineer | opus | orange | Metrics, logs, traces, alerting design |
| orchestrator | opus | orange | Reliability workflow coordination |
| platform-engineer | opus | cyan | CI/CD pipeline and infrastructure |

### Strategy-Pack (5 agents)

| Agent | Model | Color | Role |
|-------|-------|-------|------|
| business-model-analyst | opus | green | Financial modeling and unit economics |
| competitive-analyst | opus | cyan | Competitive intelligence |
| market-researcher | opus | orange | Market sizing and segment analysis |
| orchestrator | opus | yellow | Strategic workflow coordination |
| roadmap-strategist | opus | purple | Strategic prioritization and OKR design |

### User-Level Agents (3 agents)

| Agent | Model | Color | Role |
|-------|-------|-------|------|
| consultant | opus | cyan | Ecosystem navigation and guidance |
| context-engineer | opus | orange | Claude context architecture design |
| state-mate | sonnet | yellow | Session state mutation authority |

---

## Part 2: Commands by Category

### Session Commands (5)

| Command | Model | Description |
|---------|-------|-------------|
| `/start` | opus | Initialize new work session |
| `/continue` | sonnet | Resume parked session |
| `/park` | sonnet | Pause and preserve session |
| `/handoff` | opus | Transfer work to different agent |
| `/wrap` | sonnet | Complete session with quality gates |

### Workflow Commands (3)

| Command | Model | Description |
|---------|-------|-------------|
| `/task` | opus | Single task full lifecycle |
| `/sprint` | opus | Multi-task sprint orchestration |
| `/hotfix` | opus | Rapid fix workflow |

### Operations Commands (5)

| Command | Model | Description |
|---------|-------|-------------|
| `/architect` | opus | Design-only session |
| `/build` | opus | Implementation from approved design |
| `/code-review` | opus | Structured review with feedback |
| `/qa` | opus | Validation-only with approval |
| `/commit` | sonnet | Git commit with AI message |

### Navigation Commands (5)

| Command | Model | Description |
|---------|-------|-------------|
| `/consult` | opus | Ecosystem guidance |
| `/ecosystem` | sonnet | Switch to ecosystem-pack |
| `/team` | sonnet | Switch team packs |
| `/sessions` | sonnet | List/manage sessions |
| `/worktree` | sonnet | Manage isolated worktrees |

### Team-Switching Commands (10)

| Command | Model | Target Team |
|---------|-------|-------------|
| `/10x` | sonnet | 10x-dev-pack |
| `/docs` | sonnet | doc-team-pack |
| `/hygiene` | sonnet | hygiene-pack |
| `/debt` | sonnet | debt-triage-pack |
| `/sre` | sonnet | sre-pack |
| `/security` | sonnet | security-pack |
| `/intelligence` | sonnet | intelligence-pack |
| `/rnd` | sonnet | rnd-pack |
| `/strategy` | sonnet | strategy-pack |
| `/forge` | sonnet | forge-pack |

### Meta Commands (3)

| Command | Model | Description |
|---------|-------|-------------|
| `/minus-1` | opus | Assess initiative readiness |
| `/zero` | opus | Initialize Orchestrator with plan |
| `/one` | opus | Execute workflow autonomously |

### CEM Commands (1)

| Command | Model | Description |
|---------|-------|-------------|
| `/sync` | sonnet | Sync with skeleton ecosystem |

### Team-Specific Commands (7)

| Command | Model | Team | Description |
|---------|-------|------|-------------|
| `/pr` | sonnet | 10x-dev-pack | Create pull request |
| `/spike` | opus | 10x-dev-pack | Time-boxed research |
| `/consolidate` | opus | doc-team-pack | Consolidate documentation |
| `/cem-debug` | opus | ecosystem-pack | Diagnose CEM issues |
| `/eval-agent` | opus | forge-pack | Test agent in isolation |
| `/new-team` | opus | forge-pack | Create new team pack |
| `/validate-team` | opus | forge-pack | Validate team pack |

---

## Part 3: Hooks Analysis

**Finding**: Hooks do NOT contain model selection logic. Model assignment is static in agent/command frontmatter.

| Hook | Purpose | Model Logic |
|------|---------|-------------|
| session-context.sh | Inject session context | None |
| delegation-check.sh | Validate agent delegation | None |
| auto-park.sh | Session state management | None |
| start-preflight.sh | Pre-flight validation | None |
| All others | Various utilities | None |

---

## Part 4: Model Assignment Patterns

### By Role Type

| Role Type | Model | Rationale |
|-----------|-------|-----------|
| Orchestrators | opus | Complex multi-phase planning |
| Architects/Analysts | opus | Deep analysis, design decisions |
| Quality/Testing | opus | Judgment calls on quality gates |
| Strategy/Research | opus | Business reasoning |
| Implementation | sonnet | Balanced speed/quality |
| Documentation | sonnet | Content creation |
| Execution/Cleanup | sonnet | Straightforward operations |

### By Complexity

| Complexity | Model | Examples |
|------------|-------|----------|
| High (coordination, design) | opus | orchestrator, architect, analyst |
| Medium (implementation) | opus/sonnet | principal-engineer (opus), janitor (sonnet) |
| Low (switching, state) | sonnet | team commands, state-mate |

---

## Part 5: Upgrade Candidates

### Potential Opus → Sonnet (Cost Optimization)

These agents might work well with sonnet given their focused roles:

| Agent | Team | Current | Consideration |
|-------|------|---------|---------------|
| principal-engineer | 10x-dev-pack | opus | Implementation-focused, could be sonnet |
| information-architect | doc-team-pack | opus | Structure work, possibly sonnet |

### Potential Sonnet → Opus (Quality Improvement)

These agents might benefit from opus for better reasoning:

| Agent | Team | Current | Consideration |
|-------|------|---------|---------------|
| analytics-engineer | intelligence-pack | sonnet | Complex pipeline design |
| debt-collector | debt-triage-pack | sonnet | Judgment on debt priority |

### Haiku Candidates (Speed/Cost)

Currently 0 haiku. Potential candidates for simple tasks:

| Agent/Command | Current | Consideration |
|---------------|---------|---------------|
| Team-switching commands | sonnet | Very simple routing |
| `/commit` | sonnet | Straightforward git ops |
| state-mate | sonnet | Schema-driven mutations |

---

## Part 6: Anomalies

| Item | Issue | Recommendation |
|------|-------|----------------|
| strategy-pack orchestrator | color: yellow | Update to standard (cyan/purple) |

---

## Appendix: File Locations

```
Agents:
├── $ROSTER_HOME/teams/*/agents/*.md      (team agents)
├── $ROSTER_HOME/user-agents/*.md          (user-level agents)
└── ~/.claude/agents/*.md                   (global installed)

Commands:
├── $ROSTER_HOME/user-commands/*/*.md      (user commands)
└── $ROSTER_HOME/teams/*/commands/*.md     (team commands)

Hooks:
└── $ROSTER_HOME/.claude/hooks/*.sh        (no model logic)

Skills:
├── $ROSTER_HOME/.claude/skills/*/         (active skills)
└── $ROSTER_HOME/teams/*/skills/*/         (team skills)
```

---

## Summary for Upgrade Planning

1. **Current State**: 62 opus (60%), 41 sonnet (39%), 0 haiku
2. **Pattern**: Complexity-driven assignment is consistent
3. **Opportunities**:
   - Haiku for simple routing/state commands
   - Sonnet for some implementation agents
   - Opus for complex judgment agents currently on sonnet
4. **No Dynamic Selection**: All models are static in frontmatter
5. **Validation**: Orchestrator validation requires model field

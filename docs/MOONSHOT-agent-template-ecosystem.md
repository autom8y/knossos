# MOONSHOT-agent-template-ecosystem

## Executive Summary

The agent template ecosystem will evolve from POC-validated orchestrator generation to a comprehensive **Agent Factory** that generates, validates, and distributes all agent types across satellites. The 2027 vision: agents are defined as composable YAML specifications, generated at sync-time, and validated against behavioral contracts. The ecosystem survives paradigm shifts (native Claude templating, MCP) by treating generation as an implementation detail while preserving semantic contracts as the durable abstraction.

## Time Horizon

3 years (2025-2027)

---

## Scenario Definition

### Scenario A: Claude Code Native Templating

**Probability**: High (70%)
**Impact if True**: Critical

Anthropic introduces native agent templating in Claude Code, likely as structured YAML/JSON agent definitions that Claude parses directly.

**Assumptions**:
- Claude Code adoption continues growing
- Anthropic sees agent reuse as a pain point worth solving
- Native implementation would be simpler than ecosystem workarounds

**Triggers/Signals**:
- Claude Code changelog mentions "agent templates" or "reusable agents"
- MCP specification includes agent definition schemas
- Anthropic blog posts about "agent composition"
- Claude Code begins supporting structured agent formats beyond markdown

**Stress Test**: Current YAML+Shell generation becomes redundant. However, semantic specifications (routing tables, domain authority, handoff criteria) remain valuable as inputs to any templating system.

### Scenario B: Multi-LLM Agent Ecosystem

**Probability**: Medium (40%)
**Impact if True**: High

Agent ecosystem must support multiple LLM backends (Claude, GPT-5, Gemini) with agents that work across providers.

**Assumptions**:
- Enterprise customers demand LLM flexibility
- Different providers excel at different agent types
- Cost optimization drives multi-provider strategies

**Triggers/Signals**:
- GPT-5 or Gemini launch with competitive agent capabilities
- Customer requests for non-Claude agent support
- Claude Code supports alternative model providers
- Emergence of LLM-agnostic agent standards (beyond MCP)

**Stress Test**: Current templates are Claude-specific (tool names, model IDs). Need abstraction layer between semantic specification and provider-specific rendering.

### Scenario C: 100x Scale (Hundreds of Teams/Satellites)

**Probability**: Medium (35%)
**Impact if True**: High

Ecosystem grows from 10-15 satellites to 100+ projects, each with customized agent variants.

**Assumptions**:
- Adoption spreads beyond core team
- Different domains need specialized agent flavors
- Manual template management becomes unsustainable

**Triggers/Signals**:
- More than 30 active satellites
- Template customization requests exceed 5 per month
- Sync failures due to satellite divergence
- Request for self-service template creation

**Stress Test**: Current manual sync (swap-rite.sh + CEM) breaks at scale. Need automated distribution, conflict resolution, and satellite health monitoring.

### Scenario D: Regulatory/Compliance Requirements

**Probability**: Low (20%)
**Impact if True**: Medium

Agent behaviors require audit trails, version control, and compliance attestation for regulated industries.

**Assumptions**:
- AI agents fall under software audit requirements
- Healthcare, finance, government adoption
- Liability concerns drive formal verification

**Triggers/Signals**:
- Customer requests for agent behavior auditing
- Regulatory guidance mentioning AI agents
- SOC 2 auditors asking about agent versioning
- Legal review of agent-produced artifacts

**Stress Test**: Current templates lack formal versioning, behavioral contracts, or audit trails. Need semantic versioning, behavioral test suites, and provenance tracking.

---

## Current State

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                          ROSTER                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ rites/                                                   │   │
│  │   ├── 10x-dev-pack/                                     │   │
│  │   │     ├── agents/*.md        (handcrafted)            │   │
│  │   │     └── workflow.yaml                               │   │
│  │   ├── rnd-pack/                                         │   │
│  │   │     ├── agents/*.md        (handcrafted)            │   │
│  │   │     ├── orchestrator.yaml  (POC: spec)              │   │
│  │   │     └── workflow.yaml                               │   │
│  │   └── ... (10 rites)                               │   │
│  │                                                          │   │
│  │ templates/                     (POC)                     │   │
│  │   ├── orchestrator-base.md.tpl                          │   │
│  │   └── generate-orchestrator.sh                          │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ swap-rite.sh
                              │ (copy agents, update CLAUDE.md)
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                         SATELLITE                               │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ .claude/                                                 │   │
│  │   ├── agents/*.md           (copied from roster)        │   │
│  │   ├── ACTIVE_RITE                                       │   │
│  │   ├── ACTIVE_WORKFLOW.yaml                              │   │
│  │   └── AGENT_MANIFEST.json   (provenance tracking)       │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              ▲
                              │ CEM sync
                              │ (skeleton → satellite)
                              │
┌─────────────────────────────────────────────────────────────────┐
│                         SKELETON                                │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ .claude/                                                 │   │
│  │   ├── skills/              (shared across satellites)   │   │
│  │   ├── hooks/               (session management)         │   │
│  │   └── CLAUDE.md            (template, merge-docs)       │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

### Key Constraints

1. **Handcrafted agents**: 50+ agents manually maintained across 10 packs
2. **Duplication**: Orchestrators share ~60% identical content
3. **Fragile sync**: swap-rite.sh requires manual intervention for conflicts
4. **No validation**: Agent changes can silently break workflows
5. **No versioning**: Template changes have no rollback path

### Technical Debt Affecting Future

| Debt Item | Impact on Future Architecture |
|-----------|------------------------------|
| No agent schema | Cannot validate agent specifications programmatically |
| Hardcoded tool names | Prevents multi-provider support |
| Manual CLAUDE.md update | Scales poorly beyond 10 teams |
| No behavioral tests | Cannot verify agents work after generation |
| Monolithic swap-rite.sh | Difficult to extend for new agent types |

---

## Future Architecture

### Vision

By 2027, the Agent Factory is a fully automated pipeline:

1. **Define**: Teams describe agents in declarative YAML specifications
2. **Generate**: Factory produces provider-specific markdown from specs
3. **Validate**: Behavioral contracts verify generated agents work correctly
4. **Distribute**: Automated sync pushes agents to satellites with conflict resolution
5. **Audit**: Full provenance tracking for compliance requirements

Agents are **commoditized infrastructure**—teams focus on domain logic, not prompt engineering.

### Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              AGENT FACTORY                                  │
│                                                                             │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐        │
│  │ Agent Schema    │───▶│ Agent Generator │───▶│ Agent Validator │        │
│  │                 │    │                 │    │                 │        │
│  │ - orchestrator  │    │ - YAML → MD     │    │ - Behavioral    │        │
│  │ - specialist    │    │ - Multi-target  │    │   contracts     │        │
│  │ - reviewer      │    │ - Composable    │    │ - Schema valid  │        │
│  │ - (extensible)  │    │                 │    │ - Lint rules    │        │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘        │
│           │                     │                      │                   │
│           ▼                     ▼                      ▼                   │
│  ┌────────────────────────────────────────────────────────────────┐       │
│  │                     Generated Agents                            │       │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │       │
│  │  │ orchestrator│  │ specialist  │  │ reviewer    │             │       │
│  │  │ .md         │  │ .md         │  │ .md         │             │       │
│  │  └─────────────┘  └─────────────┘  └─────────────┘             │       │
│  └────────────────────────────────────────────────────────────────┘       │
│                                  │                                         │
└──────────────────────────────────┼─────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          DISTRIBUTION LAYER                                 │
│                                                                             │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐        │
│  │ Roster Registry │───▶│ Satellite Sync  │───▶│ Health Monitor  │        │
│  │                 │    │                 │    │                 │        │
│  │ - Versions      │    │ - Push/Pull     │    │ - Drift detect  │        │
│  │ - Channels      │    │ - Conflict res  │    │ - Alerts        │        │
│  │ - Deprecation   │    │ - Rollback      │    │ - Metrics       │        │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘        │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
                                   │
                                   ▼
         ┌─────────────────────────┴─────────────────────────┐
         ▼                         ▼                         ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Satellite A   │    │   Satellite B   │    │   Satellite N   │
│                 │    │                 │    │                 │
│  .claude/agents │    │  .claude/agents │    │  .claude/agents │
│  AGENT_MANIFEST │    │  AGENT_MANIFEST │    │  AGENT_MANIFEST │
│  (locked)       │    │  (customized)   │    │  (bleeding edge)│
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Key Changes

| Area | Current | Future | Rationale |
|------|---------|--------|-----------|
| Agent definition | Handcrafted markdown | Declarative YAML specs | Enables generation, validation, composition |
| Generation | Single orchestrator template | Multi-type Agent Factory | Scales to all agent types with consistency |
| Validation | None | Behavioral contracts + schema | Catches breaking changes before distribution |
| Distribution | Manual swap-rite.sh | Automated registry + channels | Scales to 100+ satellites |
| Versioning | Git commits only | Semantic versions + channels | Supports stable/beta/edge deployment |
| Customization | Fork and edit | Override layers | Preserves upstream compatibility |

### New Capabilities Required

1. **Agent Schema Language**: JSON Schema for agent YAML specs, defining required/optional fields, tool enums, model constraints
2. **Multi-Type Templates**: Base templates for orchestrator, specialist, reviewer archetypes with composition support
3. **Behavioral Contracts**: Executable tests that verify agent outputs match expected patterns
4. **Override System**: Satellite-specific customization that merges cleanly with upstream changes
5. **Registry Service**: Version management with stable/beta/edge channels
6. **Health Dashboard**: Satellite drift detection, sync status, agent utilization metrics

### Technology Dependencies

| Technology | Purpose | Maturity | Risk |
|------------|---------|----------|------|
| JSON Schema | Agent spec validation | Mature | Low |
| yq/jq | YAML/JSON processing | Mature | Low |
| envsubst / gomplate | Template rendering | Mature | Low |
| git subtree/submodule | Distribution | Mature | Medium (complexity) |
| MCP | Future agent protocol | Early | High (spec changes) |
| Native Claude templating | Potential replacement | Speculative | High (unknown timeline) |

### Scaling Implications

**10x scale (50 satellites)**:
- Registry + channels become essential
- Automated health monitoring required
- Self-service template creation needed

**100x scale (500 satellites)**:
- Federated registries (team-level autonomy)
- Policy-based distribution (compliance gates)
- Full audit trail required

---

## Migration Path

### Phase 1: Schema Foundation (Q1 2025)

**Goal**: Define agent schema and validate POC templates against it

**Deliverables**:
- JSON Schema for agent YAML specs (frontmatter, sections, skills references)
- Schema validation in generate-orchestrator.sh
- Migration of rnd-pack orchestrator.yaml to schema-compliant format
- Documentation of agent taxonomy (orchestrator, specialist, reviewer)

**Investment**: 1-2 weeks
**Reversibility**: Two-way door (schema is additive, existing agents unaffected)

### Phase 2: Multi-Type Generation (Q2 2025)

**Goal**: Extend generation from orchestrators to all agent types

**Deliverables**:
- Base templates for specialist and reviewer archetypes
- Composition system for shared sections (Domain Authority, Handoff Criteria, etc.)
- Generation of 10x-dev-pack agents from YAML specs
- Comparison tooling (generated vs. handcrafted diff)

**Investment**: 2-3 weeks
**Reversibility**: Two-way door (can fall back to handcrafted if generation fails)

### Phase 3: Behavioral Validation (Q3 2025)

**Goal**: Automated verification that generated agents work correctly

**Deliverables**:
- Behavioral contract format (expected consultation protocol, tool access, output structure)
- Contract runner that invokes agent with test prompts
- CI integration for agent validation on PR
- Regression test suite for all agent types

**Investment**: 3-4 weeks
**Reversibility**: Two-way door (validation is additive, doesn't change agents)

### Phase 4: Registry + Channels (Q4 2025)

**Goal**: Automated distribution with version management

**Deliverables**:
- Agent version manifest with stable/beta/edge channels
- Automated sync from registry to satellites
- Conflict resolution with override preservation
- Rollback mechanism for failed updates

**Investment**: 4-6 weeks
**Reversibility**: One-way door for registry adoption, but satellites can disconnect

### Phase 5: Health + Compliance (2026)

**Goal**: Enterprise-ready monitoring and audit support

**Deliverables**:
- Health dashboard with drift detection
- Audit trail for agent changes with provenance
- Compliance reports for regulated environments
- Self-service template creation UI

**Investment**: 8-12 weeks
**Reversibility**: Two-way door (features are additive)

### Phase 6: Multi-Provider Abstraction (2026-2027)

**Goal**: Support for non-Claude LLM backends

**Deliverables**:
- Provider abstraction layer in agent specs
- Rendering targets for GPT, Gemini, etc.
- Tool name mapping per provider
- Behavioral contract adaptation per provider

**Investment**: 12-16 weeks
**Reversibility**: One-way door (significant architecture change)

### Decision Points

| Decision | When | Options | Implications |
|----------|------|---------|--------------|
| Native templating adoption | Claude Code announces | A: Migrate to native, B: Maintain both, C: Ignore | A: Major rewrite, B: Maintenance burden, C: Obsolescence risk |
| Multi-provider support | Customer request or GPT-5 launch | A: Full abstraction, B: Provider-specific forks, C: Claude-only | A: Complex but flexible, B: Duplication, C: Customer loss |
| Registry hosting | 50+ satellites | A: Git-based, B: Artifact registry, C: Custom service | A: Familiar but limited, B: Enterprise integration, C: Maintenance |
| Compliance attestation | Regulatory inquiry | A: Manual process, B: Automated contracts, C: Third-party audit | A: Scales poorly, B: Engineering investment, C: Ongoing cost |

---

## Risk Analysis

### Scenario Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Claude native templating obsoletes Factory | High | High | Design Factory as semantic layer feeding any template system |
| MCP changes break agent definitions | Medium | Medium | Abstract MCP specifics, version-lock dependencies |
| Multi-provider demand before readiness | Low | High | Prioritize abstraction in Phase 2 design |
| Compliance requirements before Phase 5 | Low | Medium | Fast-track audit trail if needed |

### Execution Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Behavioral contracts too complex to maintain | Medium | High | Start with smoke tests, not full coverage |
| Generated agents inferior to handcrafted | Medium | Medium | A/B comparison in Phase 2, revert if needed |
| Registry adoption friction | Low | Medium | Incremental migration, satellite opt-in |
| Template complexity explosion | Medium | Medium | Strict composition rules, template linting |

---

## Investment Summary

| Phase | Duration | Effort | Key Investments |
|-------|----------|--------|-----------------|
| 1: Schema | 1-2 weeks | 0.5 FTE | JSON Schema, validation tooling |
| 2: Multi-Type | 2-3 weeks | 0.5 FTE | Templates, composition system |
| 3: Validation | 3-4 weeks | 0.5 FTE | Contract format, runner, CI |
| 4: Registry | 4-6 weeks | 1 FTE | Distribution, versioning, rollback |
| 5: Health | 8-12 weeks | 1 FTE | Dashboard, audit trail, compliance |
| 6: Multi-Provider | 12-16 weeks | 1.5 FTE | Abstraction layer, provider targets |

**Total Estimated Investment**: 30-43 weeks over 2-3 years

**Note**: Phases 1-3 are "must do" for maintainability. Phases 4-6 are "as needed" based on scale and requirements.

---

## Strategic Implications

This architecture positions the agent ecosystem as **platform infrastructure** rather than tooling:

1. **Competitive moat**: Semantic agent specifications are durable even if rendering changes
2. **Ecosystem lock-in (positive)**: Satellites depend on registry for agent updates
3. **Enterprise readiness**: Compliance and audit support unlock regulated markets
4. **Multi-provider optionality**: Avoids vendor lock-in while optimizing for Claude today

---

## Recommendations

### Immediate Actions (Start Now)

1. **Define agent schema**: Write JSON Schema for agent YAML specs (1 week)
2. **Validate POC output**: Run schema validation on rnd-pack generated orchestrator
3. **Document taxonomy**: Formalize orchestrator/specialist/reviewer archetypes
4. **Inventory section patterns**: Catalog shared sections across all 50+ agents

### Decisions Needed

1. **Phase 2 scope**: All agent types or just orchestrators first? (By end of Phase 1)
2. **Behavioral contract complexity**: Smoke tests or full coverage? (Before Phase 3)
3. **Registry technology**: Git-based or dedicated service? (Before Phase 4)

### What to Watch

1. **Claude Code changelog**: Any mention of "agent," "template," "reusable"
2. **MCP specification evolution**: Agent-related schemas
3. **Satellite growth rate**: Inflection point for registry need
4. **Customer compliance requests**: Audit trail urgency

---

## Open Questions

1. Should the Factory live in roster or skeleton? (Currently leaning roster)
2. How do we handle agents that are genuinely unique (not template-able)?
3. What's the migration path for existing handcrafted agents?
4. Should behavioral contracts use actual Claude invocations or pattern matching?
5. How do we incentivize satellites to stay on stable channel vs. edge?

---

## Anti-Goals

This system explicitly should NOT become:

1. **A no-code agent builder**: We optimize for expert users, not drag-and-drop
2. **A prompt marketplace**: Agents are internal infrastructure, not products
3. **A replacement for prompt engineering skill**: Templates encode best practices, not eliminate expertise
4. **A multi-tenant SaaS**: Self-hosted, not cloud-dependent
5. **Overly abstract**: Pragmatic shell scripts over enterprise frameworks

---

## Success Metrics

| Metric | Current | Phase 3 Target | Phase 6 Target |
|--------|---------|----------------|----------------|
| Time to create new agent | 2-4 hours | 30 minutes | 10 minutes |
| Agent duplication ratio | ~40% | <10% | <5% |
| Sync failure rate | Manual (N/A) | <5% | <1% |
| Satellite drift detection | None | Daily check | Real-time |
| Breaking change detection | Manual review | CI validation | Pre-commit hook |
| Multi-provider support | Claude only | Claude only | 2+ providers |

---

## Appendix: Agent Taxonomy

Based on analysis of 50+ agents across 10 rites:

### Orchestrators (10 agents)

**Pattern**: Stateless advisors, consultation protocol, routing tables
**Templateability**: 60-70% (protocol canonical, routing team-specific)
**Key sections**: Consultation Role, Input/Output schemas, Routing Criteria, Handling Failures

### Specialists (35+ agents)

**Pattern**: Domain experts, tool-heavy, artifact producers
**Templateability**: 40-50% (core sections shared, domain logic unique)
**Key sections**: Core Responsibilities, Domain Authority, Approach, What You Produce, Handoff Criteria

**Sub-types**:
- Analysts (requirements-analyst, threat-modeler, competitive-analyst)
- Engineers (principal-engineer, prototype-engineer, platform-engineer)
- Architects (architect, moonshot-architect, context-architect)
- Writers (tech-writer, documentation-engineer)

### Reviewers (5+ agents)

**Pattern**: Validation focus, adversarial stance, pass/fail outputs
**Templateability**: 50-60% (adversarial patterns shared, domain criteria unique)
**Key sections**: Adversarial Testing, Quality Standards, Handoff Criteria

**Examples**: qa-adversary, doc-reviewer, security-reviewer, compatibility-tester

---

*Document produced by Moonshot Architect as terminal agent in rnd-pack pipeline.*

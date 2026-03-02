# Rite-Skill Matrix

> Maps skills to rites for routing and activation decisions.

## Overview

This matrix documents which skills are relevant to each rite. Use it to:
- Route work to appropriate rites based on skill requirements
- Understand skill activation patterns for each rite
- Plan rite composition for complex initiatives

---

## Matrix Legend

| Symbol | Meaning |
|--------|---------|
| **P** | Primary - Core to rite workflow, frequently activated |
| **S** | Secondary - Useful but not core, occasionally activated |
| (blank) | Not applicable to this rite |

---

## Documentation Template Skills

These skills provide templates for artifacts produced by rites.

| Skill | 10x-dev | ecosystem | doc-team | debt | sre | security | strategy | hygiene | intelligence | rnd | slop-chop |
|-------|:-------:|:---------:|:--------:|:----:|:---:|:--------:|:--------:|:-------:|:------------:|:---:|:---------:|
| doc-artifacts | **P** | S | S | S | S | S | S | S | S | S | S |
| doc-consolidation | S | **P** | **P** | S | | | | S | | | |
| doc-ecosystem | S | **P** | S | | | | | **P** | | | |
| doc-intelligence | | | S | | | | S | | **P** | | |
| doc-reviews | S | S | **P** | | | | | S | | | S |
| doc-rnd | | S | S | | | | | | | **P** | |
| doc-security | | | | | S | **P** | | | | | |
| doc-sre | | | | **P** | **P** | | | | | | |
| doc-strategy | | | | | | | **P** | | S | | |
| documentation | S | S | **P** | S | S | S | S | S | S | S | S |

---

## Workflow Skills

Core workflow execution patterns for development and operations.

| Skill | 10x-dev | ecosystem | doc-team | debt | sre | security | strategy | hygiene | intelligence | rnd | slop-chop |
|-------|:-------:|:---------:|:--------:|:----:|:---:|:--------:|:--------:|:-------:|:------------:|:---:|:---------:|
| 10x-workflow | **P** | S | S | S | S | S | S | S | S | S | S |
| hotfix-ref | **P** | | | | **P** | S | | S | | | |
| orchestration | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** |
| pr-ref | **P** | **P** | S | | S | S | | S | | S | **P** |
| qa-ref | **P** | S | **P** | | S | **P** | | **P** | | | **P** |
| review | **P** | S | **P** | S | S | **P** | | **P** | | | **P** |
| spike-ref | **P** | S | | | S | S | | | S | **P** | |
| sprint-ref | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** |
| task-ref | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** |

---

## Workflow Commands (Dromena)

These are slash commands (.dro.md) that execute as transient actions. They are NOT skills and cannot be preloaded.

| Command | 10x-dev | ecosystem | doc-team | debt | sre | security | strategy | hygiene | intelligence | rnd | slop-chop |
|---------|:-------:|:---------:|:--------:|:----:|:---:|:--------:|:--------:|:-------:|:------------:|:---:|:---------:|
| /architect (architect-ref) | **P** | S | | | S | S | | S | | S | |
| /build (build-ref) | **P** | S | | | S | S | | S | | S | |

---

## Session Management Skills

Session lifecycle control for all teams.

| Skill | 10x-dev | ecosystem | doc-team | debt | sre | security | strategy | hygiene | intelligence | rnd | slop-chop |
|-------|:-------:|:---------:|:--------:|:----:|:---:|:--------:|:--------:|:-------:|:------------:|:---:|:---------:|
| start-ref | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** |
| park-ref | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** |
| resume | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** |
| wrap-ref | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** |
| handoff-ref | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** |
| initiative-scoping | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** |

---

## Team Switch Skills

Quick team context switching. These are dromena (slash commands), not skills. They materialize to `.claude/commands/`.

| Command | 10x-dev | ecosystem | doc-team | debt | sre | security | strategy | hygiene | intelligence | rnd | slop-chop |
|---------|:-------:|:---------:|:--------:|:----:|:---:|:--------:|:--------:|:-------:|:------------:|:---:|:---------:|
| /10x (10x-ref) | **P** | S | S | S | S | S | S | S | S | S | |
| debt-ref | S | S | S | **P** | S | | | S | | | |
| docs-ref | S | S | **P** | S | S | S | S | S | S | S | |
| hygiene-ref | S | **P** | S | S | S | | | **P** | | | |
| slop-chop-ref | | | | | | | | | | | **P** |
| sre-ref | S | S | S | S | **P** | S | | S | | | |
| team-ref | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** |

---

## Rite Reference Skills

Documentation for specific rites.

| Skill | 10x-dev | ecosystem | doc-team | debt | sre | security | strategy | hygiene | intelligence | rnd | slop-chop |
|-------|:-------:|:---------:|:--------:|:----:|:---:|:--------:|:--------:|:-------:|:------------:|:---:|:---------:|
| intelligence-ref | S | | S | | | | S | | **P** | | |
| rnd-ref | S | | S | | | | S | | | **P** | |
| security-ref | S | | S | | | **P** | | | | | S |
| strategy-ref | S | | S | | | | **P** | | S | | |

---

## Ecosystem Skills

CEM, skeleton, roster infrastructure and CLAUDE.md architecture.

| Skill | 10x-dev | ecosystem | doc-team | debt | sre | security | strategy | hygiene | intelligence | rnd | slop-chop |
|-------|:-------:|:---------:|:--------:|:----:|:---:|:--------:|:--------:|:-------:|:------------:|:---:|:---------:|
| ecosystem-ref | S | **P** | S | | | | | **P** | | | |
| claude-md-architecture | S | **P** | | | | | | S | | | |
| consult-ref | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** |

---

## Rite Creation Skills

Creating and managing rites.

| Skill | 10x-dev | ecosystem | doc-team | debt | sre | security | strategy | hygiene | intelligence | rnd | slop-chop |
|-------|:-------:|:---------:|:--------:|:----:|:---:|:--------:|:--------:|:-------:|:------------:|:---:|:---------:|
| forge-ref | S | **P** | | | | | | | | S | |
| team-development | S | **P** | | | | | | | | S | |

---

## Reference Skills

Standards, patterns, and invocation guides.

| Skill | 10x-dev | ecosystem | doc-team | debt | sre | security | strategy | hygiene | intelligence | rnd | slop-chop |
|-------|:-------:|:---------:|:--------:|:----:|:---:|:--------:|:--------:|:-------:|:------------:|:---:|:---------:|
| prompting | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** | **P** |
| standards | **P** | **P** | S | S | **P** | **P** | S | **P** | S | **P** | **P** |

---

## Tool Skills

External tool integrations and automation.

| Skill | 10x-dev | ecosystem | doc-team | debt | sre | security | strategy | hygiene | intelligence | rnd | slop-chop |
|-------|:-------:|:---------:|:--------:|:----:|:---:|:--------:|:--------:|:-------:|:------------:|:---:|:---------:|
| atuin-desktop | S | **P** | **P** | S | S | S | S | S | S | S | |
| justfile | **P** | **P** | S | S | **P** | S | | S | S | **P** | |
| worktree-ref | **P** | S | S | S | S | S | S | S | S | S | S |

---

## Team Profiles

### 10x-dev (Full Development Lifecycle)

**Primary Skills**: doc-artifacts, 10x-workflow, hotfix-ref, orchestration, pr-ref, qa-ref, review, spike-ref, sprint-ref, task-ref, start-ref, park-ref, resume, wrap-ref, handoff-ref, initiative-scoping, team-ref, consult-ref, prompting, standards, justfile, worktree-ref

**Primary Commands (dromena)**: /10x (10x-ref), /architect (architect-ref), /build (build-ref)

**Focus**: PRD to TDD to Code to QA pipeline. All core workflow and session skills are primary.

---

### ecosystem (Infrastructure Lifecycle)

**Primary Skills (14)**: doc-consolidation, doc-ecosystem, orchestration, pr-ref, sprint-ref, task-ref, start-ref, park-ref, resume, wrap-ref, handoff-ref, initiative-scoping, hygiene-ref, team-ref, ecosystem-ref, claude-md-architecture, consult-ref, forge-ref, team-development, prompting, standards, atuin-desktop, justfile

**Focus**: CEM, skeleton, roster infrastructure. Ecosystem skills are primary along with team creation.

---

### docs (Documentation Lifecycle)

**Primary Skills (10)**: doc-consolidation, doc-reviews, documentation, orchestration, qa-ref, review, sprint-ref, task-ref, start-ref, park-ref, resume, wrap-ref, handoff-ref, initiative-scoping, docs-ref, team-ref, consult-ref, prompting, atuin-desktop

**Focus**: Audit to Structure to Write to Review. Documentation template and review skills are primary.

---

### debt-triage (Debt Management)

**Primary Skills (7)**: doc-sre, orchestration, sprint-ref, task-ref, start-ref, park-ref, resume, wrap-ref, handoff-ref, initiative-scoping, debt-ref, team-ref, consult-ref, prompting

**Focus**: Collect to Assess to Plan. Debt tracking templates (doc-sre) are primary.

---

### sre (Reliability Lifecycle)

**Primary Skills (8)**: doc-sre, hotfix-ref, orchestration, sprint-ref, task-ref, start-ref, park-ref, resume, wrap-ref, handoff-ref, initiative-scoping, sre-ref, team-ref, consult-ref, prompting, standards, justfile

**Focus**: Observe to Coordinate to Build to Verify. Reliability templates and incident response are primary.

---

### security (Security Assessment)

**Primary Skills (9)**: doc-security, orchestration, qa-ref, review, sprint-ref, task-ref, start-ref, park-ref, resume, wrap-ref, handoff-ref, initiative-scoping, security-ref, team-ref, consult-ref, prompting, standards

**Focus**: Threat Model to Compliance to Pentest to Review. Security templates and review skills are primary.

---

### strategy (Strategic Planning)

**Primary Skills (7)**: doc-strategy, orchestration, sprint-ref, task-ref, start-ref, park-ref, resume, wrap-ref, handoff-ref, initiative-scoping, strategy-ref, team-ref, consult-ref, prompting

**Focus**: Market Research to Competitive Analysis to Business Modeling to Strategic Planning.

---

### hygiene (Code Quality)

**Primary Skills (11)**: doc-ecosystem, orchestration, qa-ref, review, sprint-ref, task-ref, start-ref, park-ref, resume, wrap-ref, handoff-ref, initiative-scoping, hygiene-ref, team-ref, ecosystem-ref, consult-ref, prompting, standards

**Focus**: Smell to Plan to Clean to Audit. Ecosystem docs are primary for detecting ecosystem issues during cleanup.

---

### intelligence (Product Analytics)

**Primary Skills (7)**: doc-intelligence, orchestration, sprint-ref, task-ref, start-ref, park-ref, resume, wrap-ref, handoff-ref, initiative-scoping, intelligence-ref, team-ref, consult-ref, prompting

**Focus**: Instrumentation to Research to Experimentation to Synthesis.

---

### rnd (Innovation Lab)

**Primary Skills (8)**: doc-rnd, orchestration, spike-ref, sprint-ref, task-ref, start-ref, park-ref, resume, wrap-ref, handoff-ref, initiative-scoping, rnd-ref, team-ref, consult-ref, prompting, standards, justfile

**Focus**: Scouting to Integration Analysis to Prototyping to Future Architecture.

---

### slop-chop (AI Code Quality Gate)

**Primary Skills (9)**: orchestration, pr-ref, qa-ref, review, sprint-ref, task-ref, start-ref, park-ref, resume, wrap-ref, handoff-ref, initiative-scoping, slop-chop-ref, team-ref, consult-ref, prompting, standards

**Complexity Levels**: DIFF (PR-level, 3 phases), MODULE (module-level, 5 phases), CODEBASE (full audit, 5 phases)

**Agents**: pythia (orchestrator), hallucination-hunter, logic-surgeon, cruft-cutter, remedy-smith, gate-keeper (6 total)

**Focus**: Detect to Diagnose to Cut to Remedy to Gate. AI-generated code review with hard block on FAIL. Temporal findings always advisory.

**Philosophy**: Pro-AI, anti-slop. Butcher metaphor: trim the fat, keep the meat.

---

## Skill Categories by Team Relevance

### Universal Skills (Primary for All Rites)

These skills are core infrastructure used by every rite:
- `orchestration` - Multi-phase workflow coordination
- `sprint-ref` - Sprint planning and multi-task workflows
- `task-ref` - Single focused development tasks
- `start-ref`, `park-ref`, `resume`, `wrap-ref`, `handoff-ref` - Session lifecycle
- `initiative-scoping` - Project kickoff
- `team-ref` - Team switching
- `consult-ref` - Routing guidance
- `prompting` - Agent invocation patterns

### Specialized Skills (Primary for 1-2 Rites)

| Skill | Primary Rites | Purpose |
|-------|---------------|---------|
| doc-artifacts | 10x-dev | PRD, TDD, ADR, Test Plan |
| doc-ecosystem | ecosystem, hygiene | CEM sync, migration, compatibility |
| doc-security | security | Threat models, compliance, pentest |
| doc-sre | sre, debt | Observability, incidents, debt tracking |
| doc-strategy | strategy | Roadmaps, competitive intel |
| doc-intelligence | intelligence | Research, experimentation |
| doc-rnd | rnd | Tech assessment, prototypes |
| ecosystem-ref | ecosystem, hygiene | CEM/skeleton/roster infrastructure |
| claude-md-architecture | ecosystem | CLAUDE.md content and sync |
| forge-ref, rite-development | ecosystem | Rite creation |

---

## Routing Decision Guide

**Question: "Which rite should handle this?"**

1. Check the **primary skill** needed for the task
2. Find rites where that skill is **P** (Primary)
3. If multiple rites match, choose based on artifact output:
   - PRD/TDD/Code/Tests -> 10x-dev
   - CEM/skeleton/roster -> ecosystem
   - Audit/Structure/Docs -> docs
   - Debt ledger/Risk/Sprint -> debt-triage
   - Observability/Incidents/Chaos -> sre
   - Threat/Compliance/Pentest -> security
   - Market/Competitive/Roadmap -> strategy
   - Smells/Refactor/Cleanup -> hygiene
   - Research/Experiments/Insights -> intelligence
   - Scouting/Prototypes/Moonshots -> rnd
   - AI code review/Hallucinations/Temporal debt -> slop-chop

---

## Maintenance Notes

This matrix should be updated when:
- New skills are added to SKILL_REGISTRY.md
- New rites are created in roster
- Rite workflows change significantly
- Skill activation patterns shift based on usage

**Last updated**: 2026-02-19
**Source**: SKILL_REGISTRY.md (48 skills), roster/rites/ (11 rites)

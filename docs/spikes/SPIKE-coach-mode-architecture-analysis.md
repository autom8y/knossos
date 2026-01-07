# SPIKE: Coach Mode Architecture Analysis

**Date**: 2026-01-02
**Status**: Complete
**Outcome**: New initiative created

---

## Question and Context

**Primary Question**: How does "Coach Mode" work in the current orchestrated session architecture, and is the PRD-orchestrator-entry-pattern risk assessment ("Coach mode is legacy tech debt - HIGH") valid?

**Decision This Informs**: Whether to proceed with orchestration enforcement implementation or rearchitect first.

**Timebox**: 2 hours

---

## Approach Taken

1. **Orchestrator consultation** - Routed to ecosystem-analyst per orchestrator directive
2. **Gap analysis** - Comprehensive review of hooks, detection mechanisms, and terminology
3. **User clarification** - Gathered preferences via structured questions
4. **Context design** - Routed to context-architect for design-led requirements

### Specialists Involved

| Agent | Role | Artifact |
|-------|------|----------|
| ecosystem-analyst | Gap analysis | `spike-coach-mode-analysis.md` |
| context-architect | Design specification | `CONTEXT-DESIGN-orchestration-mode-consolidation.md` |

---

## Findings

### Core Finding

**The Coach Mode concept is SOUND. The implementation is FRAGMENTED.**

The coaching metaphor (main thread coordinates but doesn't execute) correctly describes the desired separation of concerns. However:

| Issue | Severity | Description |
|-------|----------|-------------|
| Detection disagreement | HIGH | `workflow.active` field vs `execution_mode()` function can disagree |
| Missing hooks | HIGH | `orchestrator-router.sh` and `orchestrator-bypass-check.sh` registered but not implemented |
| Terminology inconsistent | MEDIUM | Three terms used interchangeably: "Coach Mode", "Orchestrated Mode", "Active Workflow" |
| Override undocumented | LOW | No formal procedure for intentional bypass |

### PRD Risk Assessment Validation

The PRD-orchestrator-entry-pattern statement "Coach mode is legacy tech debt - HIGH" is **VALID** but refers to **implementation gaps**, not conceptual flaws.

---

## User Preferences (Gathered via AskUserQuestion)

| Area | Decision |
|------|----------|
| Primary concern | All issues equally important (interconnected) |
| Urgency | Near-term priority |
| Terminology | Retire "Coach Mode" - unify on "Orchestrated Mode" |
| Detection mechanism | Context-architect decides |
| Enforcement behavior | Context-architect decides |
| Scope | Full rearchitecture |
| Design goals | Robustness > Flexibility > Simplicity |
| Constraints | Breaking changes acceptable |
| Tracking | New initiative required |

---

## Recommendation

**Proceed with "Orchestration Mode Consolidation" initiative** using the Context Design produced by context-architect.

### Design Decisions (from Context Design)

| Decision | Selected Option |
|----------|-----------------|
| Terminology | Unify on "Orchestrated Mode" |
| Detection | `execution_mode()` as single source of truth |
| Enforcement | Guide + Warn, never block |
| Audit | Per-session JSONL log |
| Override | Implicit continuation, optional acknowledgment |
| Transitions | Immediate effect, logged |

### Work Packages

| WP | Name | Dependencies |
|----|------|--------------|
| WP1 | Retire Coach Mode Terminology | None |
| WP2 | Unify Detection Mechanism | WP1 |
| WP3 | Implement Enforcement Behavior | WP2 |
| WP4 | Add Audit Trail | WP3 |
| WP5 | Document Override Mechanism | WP3 |
| WP6 | Handle Mode Transitions | WP2 |

---

## Follow-up Actions

1. **Start new initiative** - `/start "Orchestration Mode Consolidation" --team ecosystem`
2. **Use Context Design as requirements** - `docs/ecosystem/CONTEXT-DESIGN-orchestration-mode-consolidation.md`
3. **Route to integration-engineer** for implementation
4. **Validate with compatibility-tester** before merge

---

## Artifacts Produced

| Artifact | Location |
|----------|----------|
| Gap Analysis | `.claude/sessions/session-20260102-022932-a8a79927/artifacts/spike-coach-mode-analysis.md` |
| Context Design | `docs/ecosystem/CONTEXT-DESIGN-orchestration-mode-consolidation.md` |
| This Spike Report | `docs/spikes/SPIKE-coach-mode-architecture-analysis.md` |

---

## Complexity

**MODULE** - Multiple components touched, no schema changes required, breaking changes acceptable.

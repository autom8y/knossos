# Roster Ecosystem Refinement - Context Seed

> Hard-earned context from 5-agent Claude Code audit. Seeds parallel exploration sessions.

---

## Audit Summary

5 explore agents compared roster against official Claude Code plugin documentation (2026-01-02):
- **Plugin Structure Compliance** - Directory structure, plugin.json gaps
- **Command/Skill Parity** - Feature comparison with official commands
- **Hook System Analysis** - Event types, configuration, enforcement
- **Agent Architecture** - Orchestrator pattern, team packs, state-mate
- **Workflow/Session Gaps** - Session management (unique to roster)

**Key Finding**: Roster is a sophisticated meta-framework that substantially extends Claude Code. Not over-engineered - validated patterns serve real coordination needs.

---

## Validated Patterns (KEEP)

### 1. Orchestrator Agent
Context engineering for distributed workflows. Team manager analogy - coordinates specialists, maintains coherence.

**Rationale**: Multi-agent workflows need coordination beyond native Task delegation.

### 2. Team Packs
Specialized domain teams minimize cognitive load by scoping agent knowledge.

**Rationale**: Loading all agents wastes context. Teams provide just-in-time expertise.

### 3. State-Mate
Centralized authority for session/sprint mutations with schema validation (ADR-0005).

**Rationale**: Schema-standardized sessions need prompt-engineered validation.

### 4. Quality Gates
Explicit validation checkpoints between workflow phases.

**Rationale**: Official Claude Code has no quality gate mechanism.

### 5. Complexity Leveling
PATCH/MODULE/SERVICE heuristic for right-sizing workflows.

**Rationale**: Not everything needs full 4-agent workflow.

---

## Simplification Decision

### Execution Mode: 3-Mode → 2x2

**Current**: Native | Cross-Cutting | Orchestrated (complex decision tree)

**Proposed**: Two independent questions
```
1. In a session?  (Yes/No)
2. Team active?   (Yes/No)
```

**Rationale**: These are orthogonal concerns. Cross-cutting mode (session, no team) is valid, not an edge case.

---

## Opportunities (This Sprint)

### Task-001: Progressive Disclosure Audit
**Problem**: Some SKILL.md files repeat content instead of routing to supporting files
**Goal**: Audit all skills, ensure explicit routing, no duplication
**Complexity**: MODULE | **Team**: 10x-dev-pack

### Task-002: Per-Team Hook Context Injection
**Problem**: All teams share same hook context
**Goal**: Teams can override/extend base hooks for domain-specific context
**Complexity**: MODULE | **Team**: ecosystem-pack

### Task-003: Orchestrator Enforcement
**Problem**: Main thread bypasses orchestrator when complexity demands it
**Goal**: Complexity-based gating (warn/block) with override mechanism
**Complexity**: PATCH | **Team**: ecosystem-pack

### Task-004: Execution Mode Simplification
**Problem**: 3-mode model conflates session and team concerns
**Goal**: 2x2 model with migration plan
**Complexity**: MODULE | **Team**: ecosystem-pack

---

## Pain Point (Priority)

**Orchestrator Bypass**: When MODULE/SERVICE work attempted, main thread sometimes reads context directly and invokes specialists without orchestrator consultation. Breaks session coherence, loses quality gate enforcement.

Task-003 addresses this directly.

---

## Decision Rationale

### Why Parallel Sessions?
- True parallelism (different terminals, different contexts)
- Exploration freedom (each session chooses complexity, team, workflow)
- Isolation (failure in one doesn't corrupt others)

### Why Context Seed (not shared SPRINT_CONTEXT)?
- Avoids concurrent write conflicts
- Provides WHY (rationale), not just WHAT (tasks)
- Immutable = clear snapshot of starting assumptions

### Why Different Teams Per Session?
- Task-001 is a 10x-dev-pack concern (workflow optimization)
- Task-002/003/004 are ecosystem-pack concerns (infrastructure, enforcement, modes)

---

## Related Documents

| Document | Relevance |
|----------|-----------|
| `docs/requirements/PRD-hybrid-session-model.md` | Current 3-mode model |
| `docs/requirements/PRD-orchestrator-entry-pattern.md` | Orchestrator patterns |
| `docs/decisions/ADR-0005-state-mate-centralized-state-authority.md` | State-mate authority |
| `docs/triage/PATH-AUDIT-TRIAGE-2026-01-02.md` | Recent hook remediation |

---

## Coordination Protocol

### Reporting Completion
```bash
Task(moirai, "mark_complete task-00X artifact=path/to/artifact.md

Session Context:
- Session ID: session-20260102-022932-a8a79927
- Sprint ID: sprint-ecosystem-refinement-20260102")
```

### Dependencies
Task-004 MAY benefit from Task-003 insights, but can proceed independently.

### Context Seed Updates
This seed is immutable. New insights go in session artifacts, synthesized into NEXT sprint's seed if work continues.

---

## Success Criteria

- [ ] Progressive disclosure violations identified and fixed
- [ ] Per-team hook context mechanism designed and prototyped
- [ ] Orchestrator enforcement improved with complexity-based gating
- [ ] Execution mode simplified to 2x2 with migration plan

# PRD: CLAUDE.md Descriptive Architecture Revision

## Overview

The current CLAUDE.md entry point uses prescriptive language that mandates workflow patterns before users understand context, removing agency and conflicting with the hybrid operating model where both roster workflows and native Claude patterns are valid depending on session state.

This initiative revises CLAUDE.md to use descriptive language that explains what's available and when each pattern applies, positioning `/consult` as the decision helper and delegating enforcement to the orchestrated workflow rather than the entry-point text.

## Background

The Principles Purist analysis identified that CLAUDE.md should answer "What is this project?" not "How must you work here?" The current content violates this principle:

**Current problematic patterns:**
- Line 8: `**Active workflow?** MUST delegate via Task tool.`
- The term "Active workflow" is undefined at entry point
- "MUST" removes judgment from orchestrated workflow
- Doesn't account for session-without-team case
- Prescribes before describing

**Hybrid model reality:**
- Session active + team configured: roster workflow applies
- Session active + no team: partial workflow, native patterns may apply
- No session: native Claude Code is fully valid
- The entry point cannot know which state applies without context

**Architecture context:**
- CLAUDE.md uses SYNC/PRESERVE comments marking skeleton-owned vs satellite-owned sections
- The `claude-md-architecture` skill exists at `teams/ecosystem-pack/skills/claude-md-architecture/` but needs enhancement for descriptive tone
- This skill should be updated to generate descriptive content during team swaps

## User Stories

### US-1: First-Time User Orientation

**As a** developer encountering this project for the first time,
**I want** CLAUDE.md to explain what's available (teams, skills, sessions, orchestration),
**So that** I understand the ecosystem before being told how to behave.

**Acceptance Criteria:**
- Opening sections describe capabilities, not mandate behaviors
- Quick Start section explains what teams and agents are available
- No global "MUST" statements in entry sections
- References `/consult` for workflow decisions

### US-2: Hybrid Model Clarity

**As a** power user familiar with Claude Code,
**I want** CLAUDE.md to clearly distinguish session-active vs no-session scenarios,
**So that** I know when roster workflows apply and when native Claude is appropriate.

**Acceptance Criteria:**
- Conditional guidance with explicit triggers ("If session active, then...")
- No-session case explicitly acknowledged as valid
- Decision tree or flowchart conceptually present
- Native Claude patterns not prohibited, just contextualized

### US-3: Routing to /consult

**As a** user unsure which pattern to use,
**I want** CLAUDE.md to point me to `/consult` for routing decisions,
**So that** I get guidance without prescriptive entry-point mandates.

**Acceptance Criteria:**
- `/consult` mentioned in Execution Mode section
- Clear statement that `/consult` helps with workflow decisions
- Prescriptive enforcement delegated to orchestration, not entry text

### US-4: claude-md-architecture Skill

**As a** maintainer of the roster ecosystem,
**I want** a `claude-md-architecture` skill that generates descriptive CLAUDE.md content,
**So that** team swaps produce consistent, descriptive entry-point documentation.

**Acceptance Criteria:**
- Skill created at `.claude/skills/claude-md-architecture/SKILL.md`
- Skill defines what content belongs in each CLAUDE.md section
- Skill is invoked by `swap-team.sh` during team configuration
- Generated content is descriptive, not prescriptive

### US-5: Orchestration Retains Enforcement

**As an** orchestrator or workflow manager,
**I want** enforcement to happen at orchestration layer, not entry-point text,
**So that** I can apply rules contextually based on actual session state.

**Acceptance Criteria:**
- `execution-mode.md` remains the source of workflow enforcement rules
- CLAUDE.md references but does not duplicate enforcement rules
- Orchestrator can detect session state and apply appropriate rules
- Entry point describes; orchestration enforces

## Functional Requirements

### Must Have

#### FR-1: Revise Execution Mode Section

**Requirement:** Replace prescriptive Execution Mode text with descriptive guidance.

**Current:**
```markdown
## Execution Mode

**Active workflow?** MUST delegate via Task tool. See `orchestration/execution-mode.md`.
**No workflow?** May execute directly for single-phase work.
```

**Proposed:**
```markdown
## Execution Mode

This project supports three operating modes (see PRD-hybrid-session-model for details):

| Mode | Session | Team | Main Agent Behavior |
|------|---------|------|---------------------|
| **Native** | No | - | Direct execution, no tracking |
| **Cross-Cutting** | Yes | No | Direct execution + session tracking |
| **Orchestrated** | Yes | Yes (ACTIVE) | Coach pattern, delegate via Task tool |

**Unsure?** Use `/consult` for workflow routing.
```

**Rationale:** Describes all three modes from hybrid model, uses conditional table format, positions `/consult` as decision helper.

#### FR-2: Revise Agent Routing Section

**Requirement:** Replace prescriptive Agent Routing text with descriptive guidance.

**Current:**
```markdown
## Agent Routing

**Active workflow?** Delegate via Task tool. **No workflow?** Execute directly or use `/task`.
**Unsure?** Route to `/consult` for guidance.
```

**Proposed:**
```markdown
## Agent Routing

When working within a session workflow, the main thread coordinates via Task tool delegation to specialist agents. Without an active session, direct execution or `/task` initialization are both valid.

For routing guidance: `/consult`
```

**Rationale:** Describes the pattern conceptually, then provides conditional guidance.

#### FR-3: Remove Global MUST Statements

**Requirement:** Audit all skeleton-owned sections (marked `<!-- SYNC: skeleton-owned -->`) and remove or conditionalize any "MUST" statements.

**Locations to audit:**
- Execution Mode section
- Agent Routing section
- State Management section (contains "MUST include session context")

**Criteria for MUST conditionalization:**
- ALL "MUST" statements should be conditionalized or softened, including context-specific ones
- Prefer "requires" or "expects" over "MUST" (e.g., "state-mate requires session context" instead of "MUST include")
- Enforcement happens at orchestration layer, not entry-point text

#### FR-4: Update claude-md-architecture Skill

**Requirement:** Enhance existing skill to emphasize descriptive content architecture.

**Location:** `teams/ecosystem-pack/skills/claude-md-architecture/SKILL.md` (skill already exists, needs enhancement)

**Skill must define:**
- Purpose of CLAUDE.md (entry point, progressive disclosure)
- Section ownership (SYNC vs PRESERVE)
- Content principles (descriptive over prescriptive)
- What belongs in each section
- How team swaps regenerate satellite-owned sections

**Skill must NOT:**
- Define actual content (that's the regeneration script's job)
- Duplicate execution-mode.md (reference instead)

#### FR-5: Update Standards Reference

**Requirement:** Ensure `standards/SKILL.md` reference to `claude-md-architecture` points to the newly created skill.

**Current reference (line 251):**
```markdown
- [claude-md-architecture](../claude-md-architecture/SKILL.md) - Content placement for CLAUDE.md (what belongs where)
```

**Action:** Verify path resolves after skill creation.

### Should Have

#### FR-S.1: Onboarding Callout

Add a callout for new users:

```markdown
> **New to roster?** Start with `/consult` to understand your options, or `prompting` skill for copy-paste patterns.
```

### Could Have

#### FR-C.1: Session State Indicator

Add guidance for checking session state:
```markdown
Check session state: `.claude/hooks/lib/session-manager.sh status`
```

#### FR-C.2: Link to Principles Purist Analysis

Reference the analysis that motivated this change for context.

## Non-Functional Requirements

- **NFR-1**: Consistency - All CLAUDE.md sections follow the same tone (descriptive, not prescriptive)
- **NFR-2**: Maintainability - skeleton-owned sections can be updated independently of satellite-owned sections
- **NFR-3**: Discoverability - New users can understand the ecosystem without prior knowledge
- **NFR-4**: Backwards Compatibility - No breaking changes to actual workflow behavior (only entry-point text changes)

**Implementation Note**: All CLAUDE.md changes should be committed atomically to avoid inconsistent intermediate state (git hygiene, not a requirement).

## Edge Cases

| Case | Expected Behavior |
|------|-------------------|
| User reads CLAUDE.md before any session | Understands both modes are valid, no mandates apply |
| User has session but no team | Partial workflow guidance applies |
| User has session and team | Full roster workflow applies (but learned from orchestration, not entry text) |
| Team swap regenerates CLAUDE.md | Satellite-owned sections regenerated; skeleton-owned sections unchanged |
| Old CLAUDE.md without conditional language | Migration path: run regeneration script |
| Conflicting guidance between entry text and orchestration | Orchestration takes precedence (by design) |

## Success Criteria

- [ ] No global "MUST" mandates in CLAUDE.md entry sections (Execution Mode, Agent Routing)
- [ ] Conditional guidance with clear triggers ("If session active, then...")
- [ ] `/consult` positioned as decision helper in Execution Mode section
- [ ] `claude-md-architecture` skill created and documented
- [ ] Standards reference to `claude-md-architecture` resolves correctly
- [ ] Hybrid model (session vs no-session) clearly explained
- [ ] Entry text describes; orchestration enforces (verified by cross-referencing)
- [ ] Existing workflow behavior unchanged (only messaging changed)

## Dependencies

| Dependency | Type | Owner | Status |
|------------|------|-------|--------|
| `swap-team.sh` satellite section regeneration | Internal | roster | Exists |
| `orchestration/execution-mode.md` enforcement rules | Internal | roster | Exists |
| `consult-ref` skill for routing | Internal | roster | Exists |
| `claude-md-architecture` skill | Internal | ecosystem-pack | Exists (needs enhancement) |
| Skeleton-owned section update mechanism | Internal | roster | TBD |

## Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Users interpret "descriptive" as "optional" | Medium | Medium | Include clear conditional guidance; orchestration still enforces |
| Regeneration script overwrites skeleton-owned sections | Low | High | SYNC/PRESERVE comments prevent this; verify in testing |
| New users confused by two modes | Medium | Low | Flowchart and `/consult` routing address this |
| Existing documentation references old prescriptive text | Low | Low | Search and update references |

## Out of Scope

- Changing actual workflow patterns (only how they're described)
- Simplifying the ecosystem (complexity stays)
- Removing session/workflow concepts (softer introduction only)
- Modifying `execution-mode.md` enforcement rules (those remain prescriptive by design)
- Changes to hooks or session-manager behavior
- Team-specific CLAUDE.md content (this PRD covers the skeleton/shared content)

## Open Questions

*None at this time - all questions addressed during stakeholder discussion.*

---

## File Verification

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-claude-md-descriptive-architecture.md` | Created |

---

## Appendix: Content Revision Principles

Rather than prescribing exact text, this appendix documents the principles for revising CLAUDE.md content:

### Transformation Principles

1. **Describe, don't prescribe**: Replace "MUST do X" with "X applies when Y"
2. **Use conditional tables**: Show mode-dependent behavior in tabular format
3. **Reference 3-mode model**: Align with PRD-hybrid-session-model (Native/Cross-Cutting/Orchestrated)
4. **Position /consult**: Always provide routing escape hatch for uncertainty
5. **Defer enforcement**: Entry text explains; orchestration enforces

### Section-by-Section Guidance

| Section | Current Issue | Revision Approach |
|---------|---------------|-------------------|
| Execution Mode | Global MUST mandate | Conditional 3-mode table |
| Agent Routing | Prescriptive commands | Conceptual explanation + /consult routing |
| State Management | MUST language | Soften to "requires" or "expects" |

Exact text will be drafted during TDD phase based on these principles.

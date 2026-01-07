# PRD: Hybrid Session Model & Edge Case Handling

## Overview

The roster ecosystem operates in a hybrid mode alongside native Claude Code. This PRD codifies explicit rules for determining execution mode based on session state and rite context, addresses the session-without-team edge case, and documents the main agent's role in cross-cutting work. The goal is eliminating ambiguity in mode determination so both humans and agents know exactly how to behave in each scenario.

## Background

The roster project has evolved a sophisticated session-based workflow system where rites, orchestrators, and specialist agents coordinate work through explicit session tracking. However, the relationship with native Claude Code (direct execution without roster orchestration) has remained implicit.

**Current State:**
- Sessions are created via `/start` with a team and complexity level
- Active sessions route work through the orchestrator pattern
- No documentation exists for when native Claude patterns are appropriate
- The case of "session active but no team/workflow" falls through to undefined behavior
- CLAUDE.md mentions both patterns but doesn't provide decision rules

**The Hybrid Philosophy:**
Roster is designed to coexist with native Claude, not replace it. Some work benefits from full workflow orchestration (features, complex bugs, multi-phase initiatives). Other work is better served by direct Claude execution (quick questions, simple edits, exploration). Both are valid; the key is knowing which to use when.

**Discovery Context:**
- Stakeholder confirmed hybrid philosophy is intentional
- Session-without-team is a valid edge case for cross-cutting work
- Environment context (session files, ACTIVE_RITE) provides all information needed for mode detection
- Main agent (not orchestrator) is appropriate for session-without-team scenarios

## User Stories

### US-1: Mode Determination Clarity

- **US-1.1**: As a developer, I want clear rules for when to use native Claude vs. roster workflows, so that I can choose the appropriate execution mode without guessing.

- **US-1.2**: As Claude (the main agent), I want unambiguous detection of execution mode from environment state, so that I route work correctly without asking unnecessary questions.

- **US-1.3**: As a new roster user, I want documentation explaining the hybrid model, so that I understand why some work uses sessions and some doesn't.

### US-2: Session-Without-Team Support

- **US-2.1**: As a developer working on cross-cutting concerns, I want to have an active session without a specific team, so that I get session benefits (tracking, context) without forced orchestration.

- **US-2.2**: As the main agent in a session-without-team, I want to execute directly (like native mode but with session tracking), so that cross-cutting work proceeds efficiently.

- **US-2.3**: As a developer in a session-without-team, I want `/consult` available for routing guidance, so that I can switch to orchestrated mode if needed.

### US-3: Environment-Based Detection

- **US-3.1**: As a hook or agent, I want to determine execution mode by reading environment files, so that mode detection is reliable and consistent.

- **US-3.2**: As a developer, I want hooks to behave appropriately based on detected mode, so that I don't get roster warnings when working in native mode.

- **US-3.3**: As the system, I want a single source of truth for mode detection logic, so that all components make consistent decisions.

## Functional Requirements

### Must Have

#### FR-1: Execution Mode Decision Tree

- **FR-1.1**: Document and implement the following decision tree:
  ```
  User Intent
      |
      +-- No session active?
      |       |
      |       +-- Native Claude Mode
      |           Direct execution, no orchestration, no session tracking
      |
      +-- Session active?
              |
              +-- Has team AND session status = ACTIVE?
              |       |
              |       +-- Orchestrated Mode
              |           Main thread = Coach, delegates via Task tool
              |           (Note: Parked sessions are NOT orchestrated)
              |
              +-- No team, OR session parked/not-active?
                      |
                      +-- Cross-Cutting Mode
                          Main agent executes directly
                          Session tracking active
                          /consult available for routing
  ```

  **Key clarification**: Orchestrated mode requires BOTH team configured AND session status = ACTIVE. A parked session with a team is not in orchestrated mode.

- **FR-1.2**: Add `execution_mode` detection function to hook library that returns a simple string enum: `native` | `orchestrated` | `cross-cutting`. No rich object needed—string is sufficient for conditional logic.

- **FR-1.3**: Update `execution-mode.md` skill to include all three modes and their behaviors.

#### FR-2: Session-Without-Team Implementation

- **FR-2.1**: Allow `/start` to create a session without requiring team specification, resulting in session-without-team state. **Note**: Initiative is still required (e.g., `/start "My Initiative"`) to provide tracking context even without a team.

- **FR-2.2**: When session exists but no team is active, main agent operates in cross-cutting mode:
  - Direct Edit/Write allowed
  - Session tracking (artifacts, blockers, next_steps) maintained
  - No orchestrator consultation required
  - No delegation warnings from hooks

- **FR-2.3**: Add `team` field to SESSION_CONTEXT schema as optional (null indicates cross-cutting mode).

- **FR-2.4**: Update session-manager to handle null team gracefully.

#### FR-3: CLAUDE.md Updates

- **FR-3.1**: Add "Execution Mode" section to `.claude/CLAUDE.md` with the decision tree from FR-1.1.

- **FR-3.2**: Update "Agent Routing" section to reference execution mode:
  ```
  **Orchestrated?** Delegate via Task tool.
  **Cross-cutting?** Execute directly with session tracking.
  **Native?** Execute directly, no session.
  ```

- **FR-3.3**: Add quick reference table:
  | Mode | Session | Team | Main Agent Behavior |
  |------|---------|------|---------------------|
  | Native | No | - | Direct execution |
  | Cross-Cutting | Yes | No | Direct execution + tracking |
  | Orchestrated | Yes | Yes | Coach (delegate only) |

#### FR-4: Consult Enhancement

- **FR-4.1**: Update `/consult` to detect cross-cutting mode and offer routing guidance.

- **FR-4.2**: In cross-cutting mode, `/consult` response should include:
  - Acknowledgment of cross-cutting context
  - Option to switch to orchestrated mode via `/team <pack>`
  - Confirmation that direct execution is valid for current mode

#### FR-5: Hook Behavior Updates

- **FR-5.1**: Update `delegation-check.sh` to skip warnings when in native or cross-cutting mode.

- **FR-5.2**: Add mode detection to hook preamble so all hooks can check execution mode consistently.

- **FR-5.3**: Document hook behavior per mode in hook skill documentation.

### Should Have

- **FR-S.1**: Add `--no-team` flag to `/start` for explicit session-without-team creation.

- **FR-S.2**: Create prompting patterns for cross-cutting work scenarios in `prompting` skill.

- **FR-S.3**: Add mode indicator to session-context hook output (e.g., `Mode: cross-cutting`).

### Should Have

- **FR-S.4**: Allow session-without-team to "upgrade" to orchestrated via `/team <pack>` mid-session, and symmetrically allow "downgrade" from orchestrated to cross-cutting via `/team --remove`.

### Could Have

- **FR-C.1**: Add `/mode` command to display current execution mode with explanation.

- **FR-C.2**: Visual indicator in session context when operating in cross-cutting vs. orchestrated mode.

## Non-Functional Requirements

- **NFR-1**: Performance - Mode detection must complete in < 50ms to avoid hook latency.

- **NFR-2**: Reliability - Mode detection must never throw errors; graceful fallback to cross-cutting mode on any detection failure (preserves session tracking rather than losing it).

- **NFR-3**: Consistency - All components (hooks, skills, agents) must use the same mode detection logic.

- **NFR-4**: Backwards Compatibility - Existing sessions with teams must continue working unchanged.

- **NFR-5**: Discoverability - New users should encounter mode documentation within first `/consult` or help interaction.

## Edge Cases

| Case | Expected Behavior |
|------|------------------|
| Session file exists but is corrupted | Treat as native mode (no session), log warning |
| ACTIVE_RITE file exists but rite missing | Error with recovery guidance: "Team X not found. Use `/team` to reconfigure or `/team --remove` to continue in cross-cutting mode." |
| Session created with team, then team file deleted | Downgrade to cross-cutting mode |
| `/start` called with invalid rite name | Error, do not create session |
| Session-without-team receives `/handoff` | Error: "No orchestrator in cross-cutting mode. Use /team to enable orchestration." |
| Native mode receives `/park` or `/wrap` | Error: "No active session. Use /start to begin tracked work." |
| `/consult` in native mode | Works normally, may recommend starting a session |
| Session-without-team receives `/sprint` | Allow lightweight sprint: main agent executes tasks directly instead of delegating. Sprint tracking active, no orchestrator consultation. |
| Mode detection during hook execution | Must be synchronous and fast; cannot prompt user |
| Multiple terminals, different modes | Each terminal has independent mode based on its session state |

## Success Criteria

- [ ] Decision tree for execution mode documented in CLAUDE.md
- [ ] `execution_mode` function exists in hook library and returns correct mode
- [ ] `/start` works without team specification (creates cross-cutting session)
- [ ] delegation-check.sh does not warn in native or cross-cutting mode
- [ ] `/consult` acknowledges cross-cutting mode and offers routing
- [ ] Session-without-team allows direct Edit/Write without warnings
- [ ] Session-with-team enforces delegation pattern (existing behavior)
- [ ] All existing tests pass (backwards compatibility)
- [ ] New tests cover three-mode detection logic
- [ ] Main agent correctly identifies its role in each mode

## Dependencies and Risks

### Dependencies

| Dependency | Type | Owner | Status |
|------------|------|-------|--------|
| Hook library (`session-manager.sh`) | Internal | roster | Ready |
| Session schema updates | Internal | roster | Ready |
| CLAUDE.md modification | Internal | roster | Ready |
| Skill documentation updates | Internal | roster | Ready |

### Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Mode detection logic diverges across components | Medium | High | Single source of truth function, reused everywhere |
| Users confused by three modes | Medium | Medium | Clear documentation, `/consult` guidance |
| Cross-cutting mode abused to bypass orchestration | Low | Low | It's intentional; users choosing it know what they want |
| Hook performance regression from mode checks | Low | Medium | Cache detection result per hook invocation |
| Backwards compatibility break | Low | High | Extensive testing, explicit "team required" still default |

## Out of Scope

- Changing the session state machine (park/resume/wrap work as-is)
- Eliminating sessions (they provide valuable tracking)
- Forcing all work through roster (hybrid is the goal)
- Automatic mode switching based on task complexity
- Multi-user/collaborative session scenarios
- Remote session synchronization
- GUI for mode selection

## Open Questions

*None remaining - all questions resolved during stakeholder discussion.*

Key clarifications received:
1. Hybrid philosophy confirmed as intentional design
2. Session-without-team is valid for cross-cutting work
3. Main agent (not orchestrator) is appropriate for session-without-team
4. Environment files provide sufficient context for mode detection
5. `/consult` is the routing mechanism for uncertain cases

---

## Traceability

| Requirement | Source |
|-------------|--------|
| FR-1.x (Mode Decision Tree) | Stakeholder: Hybrid philosophy confirmation |
| FR-2.x (Session-Without-Team) | Stakeholder: Edge case acknowledgment |
| FR-3.x (CLAUDE.md Updates) | Gap Analysis: No hybrid rules codified |
| FR-4.x (Consult Enhancement) | Stakeholder: /consult as routing mechanism |
| FR-5.x (Hook Behavior) | Gap Analysis: Delegation check fires inappropriately |

---

## File Verification

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-hybrid-session-model.md` | Created |

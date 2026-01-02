# PRD: Per-Team Hook Context Injection

## Overview

Enable team packs to inject domain-specific context into SessionStart hooks. Currently, all teams receive the same generic session context (team name, session ID, git status). Teams like ecosystem-pack need specialized context (CEM sync status, skeleton reference, drift detection) that is currently only available in verbose mode and generated through a hardcoded script.

This PRD defines a compose pattern where teams can optionally provide a context injection script that the SessionStart hook discovers and executes, allowing each team to surface the information most relevant to their domain.

## Background

### Current State

The SessionStart hook (`session-context.sh`) provides generic context to all sessions:
- Active team name
- Execution mode (native/orchestrated/cross-cutting)
- Session ID and initiative
- Git branch and change count

Team-specific context exists but has two problems:
1. **Only in verbose mode**: Lines 261-267 of session-context.sh call `generate-team-context.sh` only when `--verbose` is set
2. **Generic format**: The existing generator outputs workflow routing tables, not team-specific status information

### Problem Statement

When working with ecosystem-pack, users need to know:
- When was the last CEM sync?
- What skeleton commit is currently synced?
- Are there local overrides (drift)?
- How many test satellites are available?

This information is critical for ecosystem work but invisible in normal session context. Users must manually run diagnostic commands or remember to use verbose mode.

### Discovery Context

Gap Analysis (task-001) identified:
- Root cause: Team context generation only in verbose path
- Recommended approach: Compose pattern with team-owned scripts
- ecosystem-pack requirements: CEM sync, skeleton ref, drift, test satellites

## User Stories

### US-1: Team-Specific Context Visibility

- **US-1.1**: As an ecosystem-pack user, I want to see CEM sync status in every session context, so that I know if I'm working with stale infrastructure definitions.

- **US-1.2**: As an ecosystem-pack user, I want to see which skeleton commit is synced, so that I can trace issues to specific skeleton versions.

- **US-1.3**: As a 10x-dev-pack user, I want session context that shows my project's test coverage or build status, so that I have relevant dev context without ecosystem noise.

### US-2: Team Pack Extensibility

- **US-2.1**: As a team pack author, I want to define what context my team needs, so that users get domain-relevant information without roster changes.

- **US-2.2**: As a team pack author, I want my context script to access hook library utilities, so that I can use existing patterns for file age checks, logging, etc.

- **US-2.3**: As a team pack author, I want a clear contract for context injection, so that my scripts integrate predictably with the SessionStart hook.

### US-3: Graceful Degradation

- **US-3.1**: As a user of a team without context injection, I want session context to work normally, so that new teams don't break existing behavior.

- **US-3.2**: As a user, I want team context errors to be logged but not block my session, so that a broken context script doesn't prevent work.

## Functional Requirements

### Must Have

#### FR-1: Team Context Loader Library

- **FR-1.1**: Create `team-context-loader.sh` in `.claude/hooks/lib/` that provides `load_team_context()` function.

- **FR-1.2**: `load_team_context()` MUST read active team from `.claude/ACTIVE_TEAM` file.

- **FR-1.3**: `load_team_context()` MUST look for context script at `$ROSTER_HOME/teams/$ACTIVE_TEAM/context-injection.sh`.

- **FR-1.4**: `load_team_context()` MUST source the script and call `inject_team_context()` function if both exist.

- **FR-1.5**: `load_team_context()` MUST return empty string and exit 0 if team has no context script (graceful skip).

- **FR-1.6**: `load_team_context()` MUST log warnings and continue if script exists but function is missing or fails.

#### FR-2: SessionStart Hook Integration

- **FR-2.1**: Modify `session-context.sh` to source `team-context-loader.sh`.

- **FR-2.2**: Add team context output to condensed mode (default), not just verbose mode.

- **FR-2.3**: Team context MUST appear after core session info but before command suggestions.

- **FR-2.4**: Team context section MUST be labeled "### Team Context" for consistent parsing.

- **FR-2.5**: Keep existing `generate-team-context.sh` call for workflow routing table (separate from team status context).

#### FR-3: Team Context Script Contract

- **FR-3.1**: Team context scripts MUST be located at `teams/$TEAM/context-injection.sh`.

- **FR-3.2**: Team context scripts MUST export a function named `inject_team_context`.

- **FR-3.3**: `inject_team_context` MUST output markdown to stdout (empty is valid).

- **FR-3.4**: `inject_team_context` SHOULD return 0 on success, non-zero on partial failure.

- **FR-3.5**: Team context scripts MAY use utilities from `team-context-loader.sh` (e.g., `team_context_row`, `is_file_stale`).

#### FR-4: Ecosystem-Pack Context Implementation

- **FR-4.1**: Create `teams/ecosystem-pack/context-injection.sh` as prototype implementation.

- **FR-4.2**: Ecosystem-pack context MUST include CEM sync status and timestamp.

- **FR-4.3**: Ecosystem-pack context MUST include current skeleton reference (branch@commit).

- **FR-4.4**: Ecosystem-pack context MUST include drift detection status.

- **FR-4.5**: Ecosystem-pack context MUST include count of available test satellites.

### Should Have

- **FR-S.1**: Provide `team_context_row "Key" "Value"` helper for consistent table formatting.

- **FR-S.2**: Provide `is_file_stale "/path" minutes` helper for age-based status checks.

- **FR-S.3**: Document context script authoring guide for team pack creators.

### Could Have

- **FR-C.1**: Add `/team-context` command to show team context on demand (outside session start).

- **FR-C.2**: Allow teams to specify context urgency levels (info/warning/critical) for visual distinction.

- **FR-C.3**: Support team context caching to avoid repeated computation during rapid session starts.

## Non-Functional Requirements

- **NFR-1**: Performance - Team context loading MUST complete in < 100ms (hook budget).

- **NFR-2**: Reliability - Team context failures MUST NOT block session start (RECOVERABLE pattern).

- **NFR-3**: Consistency - All teams using context injection MUST use the same function contract.

- **NFR-4**: Portability - Context scripts MUST work on both macOS and Linux (portable stat, date).

- **NFR-5**: Isolation - Team context script sourcing MUST NOT pollute hook namespace (use subshell).

## Edge Cases

| Case | Expected Behavior |
|------|------------------|
| ACTIVE_TEAM file missing | No team context, silent skip |
| ACTIVE_TEAM = "none" | No team context, silent skip |
| Team directory doesn't exist | No team context, silent skip |
| context-injection.sh doesn't exist | No team context, silent skip |
| Script exists but not executable | Log warning, attempt source anyway |
| inject_team_context function missing | Log warning, skip team context |
| Function returns non-zero | Log warning, show partial output |
| Function outputs nothing | No Team Context section (valid) |
| ROSTER_HOME not set | Fall back to ~/Code/roster |
| Script has syntax errors | Source fails, log warning, skip |
| Script hangs (infinite loop) | No timeout protection in v1 (document risk) |

## Success Criteria

- [ ] `load_team_context()` function exists and is tested
- [ ] Session context (condensed mode) shows Team Context section when team has script
- [ ] ecosystem-pack displays CEM sync, skeleton ref, drift, satellites
- [ ] Teams without context-injection.sh see no change (graceful skip)
- [ ] Script errors logged but don't block session start
- [ ] Performance: team context adds < 100ms to hook time
- [ ] Documentation: team pack authors know how to add context

## Dependencies and Risks

### Dependencies

| Dependency | Type | Owner | Status |
|------------|------|-------|--------|
| session-context.sh hook | Internal | roster | Ready |
| hooks-init.sh library | Internal | roster | Ready |
| config.sh (ROSTER_HOME) | Internal | roster | Ready |
| ecosystem-pack team | Internal | roster | Ready |

### Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Script performance degrades hook | Low | Medium | Document 100ms budget, profile in testing |
| Script bugs block sessions | Low | High | RECOVERABLE pattern, subshell isolation |
| Teams abuse context size | Low | Low | Document conciseness guidelines |
| Platform compatibility (stat/date) | Medium | Medium | Test on macOS and Linux, provide portability helpers |
| Infinite loop in team script | Low | High | Document risk, consider timeout in v2 |

## Out of Scope

- Automatic context refresh during session (only on SessionStart)
- Context injection for other hook types (PreToolUse, PostToolUse)
- User-level context scripts (team packs only)
- GUI/visual formatting for context (markdown only)
- Context persistence or history
- Inter-team context dependencies

## Open Questions

*None remaining - all decisions documented in Context Design.*

---

## Traceability

| Requirement | Source |
|-------------|--------|
| FR-1.x (Team Context Loader) | Gap Analysis: compose pattern recommendation |
| FR-2.x (Hook Integration) | Gap Analysis: verbose-only context was root cause |
| FR-3.x (Script Contract) | Design decision: explicit function interface |
| FR-4.x (Ecosystem-Pack) | Gap Analysis: specific team requirements listed |

---

## File Verification

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-team-hook-context.md` | Created |
| Context Design | `/Users/tomtenuta/Code/roster/docs/ecosystem/CONTEXT-DESIGN-team-context-loader.md` | Created |

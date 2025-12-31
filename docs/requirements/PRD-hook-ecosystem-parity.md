# PRD: Hook Ecosystem Parity and Context Engineering

## Overview

The roster hook ecosystem lacks parity with other managed artifacts (agents, commands, skills) and suffers from excessive context noise. This initiative addresses structural inconsistencies in naming conventions, team-level hook support, settings.json management, and context verbosity to create a unified, team-aware hook architecture.

## Background

The roster project has evolved a sophisticated team-pack system where agents, commands, and skills all have team-specific directories managed by `swap-team.sh`. However, hooks remain global-only despite `swap-team.sh` having full infrastructure for team hooks (lines 2308-2521). Deep exploration revealed several issues:

1. **Naming Inconsistency**: Canonical templates live in `roster/hooks/` while other artifacts use `roster/user-*` naming (e.g., `roster/user-agents/`, `roster/user-commands/`, `roster/user-skills/`).

2. **Settings.json Gap**: `swap-team.sh` syncs hook FILES to `.claude/hooks/` but does not update hook REGISTRATIONS in `settings.local.json`. Teams cannot customize hook behavior (matchers, timeouts, event types).

3. **Context Noise**: `session-context.sh` outputs 50-80 lines per session start. `delegation-check.sh` emits 12-line warnings. UserPromptSubmit fires on every prompt without filtering.

4. **ADR-0002 Inaccuracies**: States "CEM manages agents/commands/skills" when CEM actually ignores ALL roster artifacts. This distinction is misleading and the ADR needs correction.

5. **Legacy Files**: `team-validator.sh` and `workflow-validator.sh` remain in `hooks/` despite being superseded by `command-validator.sh`.

**Stakeholder Decision**: User confirmed a FEATURE-level complexity requiring full PRD -> TDD -> Implementation -> QA workflow with settings generation using a merge pattern (base hooks + team hooks).

## User Stories

### Scope 1: Team Hooks Parity

- **US-1.1**: As a team pack maintainer, I want to include hooks in my team pack so that teams can have specialized hook behavior (e.g., security-pack with credential-scanning hooks).

- **US-1.2**: As a developer, I want `swap-team.sh` to merge my team's hooks with base hooks so that I get both base functionality and team-specific behavior.

- **US-1.3**: As an ecosystem contributor, I want hook directories to follow the same naming convention as other artifacts so that the codebase is consistent and discoverable.

### Scope 2: Settings.json Hook Management

- **US-2.1**: As a developer, I want `swap-team.sh` to automatically configure my hooks in settings.local.json so that I don't have to manually register hooks after swapping teams.

- **US-2.2**: As a team pack maintainer, I want to specify hook matchers and timeouts in my team configuration so that hooks are optimally configured for my workflow.

- **US-2.3**: As a developer, I want base hooks and team hooks merged correctly by event type so that both operate without conflict.

### Scope 3: Context Noise Reduction

- **US-3.1**: As a developer, I want session context limited to 10-15 essential lines with optional expansion so that my context window isn't consumed by boilerplate.

- **US-3.2**: As a developer, I want UserPromptSubmit hooks to only fire on slash commands (e.g., `/start`, `/wrap`) so that normal prompts don't incur hook overhead.

- **US-3.3**: As a developer, I want hook warnings to be concise (3-5 lines maximum) so that I can quickly understand issues without scrolling.

### Scope 4: Documentation and Cleanup

- **US-4.1**: As a maintainer, I want ADR-0002 corrected to accurately reflect that roster manages ALL artifacts (hooks, agents, commands, skills) so that architecture decisions are trustworthy.

- **US-4.2**: As a maintainer, I want legacy superseded hooks removed so that the hooks directory only contains active code.

- **US-4.3**: As a contributor, I want sync scripts updated for new paths so that installation workflows work correctly.

## Functional Requirements

### Must Have

#### Scope 1: Team Hooks Parity

- **FR-1.1**: Rename `roster/hooks/` to `roster/user-hooks/` for naming consistency with `user-agents/`, `user-commands/`, `user-skills/`.

- **FR-1.2**: Add `hooks/` to the team-pack schema with the same structure as `commands/` (directory containing `.sh` files).

- **FR-1.3**: Update `swap-team.sh` `swap_hooks()` function to:
  - Copy base hooks from `roster/user-hooks/` to `.claude/hooks/`
  - Merge team hooks from `teams/<team>/hooks/` on top
  - Maintain `.team-hooks` marker for cleanup

- **FR-1.4**: Update `AGENT_MANIFEST.json` to track hooks per team using the existing manifest structure.

- **FR-1.5**: Update `install-hooks.sh` to source from `roster/user-hooks/` instead of `roster/hooks/`.

- **FR-1.6**: Update `sync-user-hooks.sh` to source from `roster/user-hooks/` instead of `roster/hooks/`.

#### Scope 2: Settings.json Management

- **FR-2.1**: Create hook registration configuration schema supporting:
  - Event type (SessionStart, Stop, PreToolUse, PostToolUse, UserPromptSubmit)
  - Hook path (relative to `.claude/hooks/`)
  - Matcher pattern (regex for tool/prompt filtering)
  - Timeout (milliseconds, default 5000)

- **FR-2.2**: Add `base_hooks.yaml` to `roster/user-hooks/` defining default hook registrations for all base hooks.

- **FR-2.3**: Allow team packs to include `hooks.yaml` defining team-specific hook registrations that merge with base.

- **FR-2.4**: Update `swap-team.sh` to generate `settings.local.json` hooks section by:
  - Reading base hook registrations from `base_hooks.yaml`
  - Reading team hook registrations from team's `hooks.yaml` (if exists)
  - Merging by event type (team hooks append to base hooks for same event)
  - Writing merged configuration to `settings.local.json`

- **FR-2.5**: Implement merge strategy: for each event type, base hooks run first, then team hooks (order matters for context injection).

#### Scope 3: Context Noise Reduction

- **FR-3.1**: Refactor `session-context.sh` to output maximum 15 lines:
  - Essential table: Project, Team, Session State, Git Status (5 lines)
  - Context-appropriate commands (3-5 lines)
  - Omit: verbose property tables, artifact counts, pre-computed values, worktree details (move to `--verbose` flag or `/status` command)

- **FR-3.2**: Add `^/` matcher to UserPromptSubmit hook registration so `start-preflight.sh` only fires on slash commands.

- **FR-3.3**: Condense `delegation-check.sh` warning from 12 lines to 3 lines:
  - Line 1: Warning title with workflow name
  - Line 2: Tool and file being modified
  - Line 3: Remediation instruction

- **FR-3.4**: Condense `session-write-guard.sh` JSON output:
  - Remove examples array (5 items -> 1 representative example)
  - Keep: decision, reason, instruction, documentation

#### Scope 4: Documentation and Cleanup

- **FR-4.1**: Correct ADR-0002 section "Ecosystem Integration: CEM Exclusion" to clarify:
  - CEM (skeleton_claude) manages nothing in roster satellites
  - roster manages ALL artifacts: hooks, agents, commands, skills, teams
  - Remove misleading "CEM manages agents/commands/skills" statement

- **FR-4.2**: Delete `hooks/team-validator.sh` (superseded by `command-validator.sh`).

- **FR-4.3**: Delete `hooks/workflow-validator.sh` (superseded by `command-validator.sh`).

- **FR-4.4**: Update all path references in documentation and scripts from `roster/hooks/` to `roster/user-hooks/`.

### Should Have

- **FR-S.1**: Create example team hooks for `10x-dev-pack` demonstrating team-specific hook capability.

- **FR-S.2**: Add `--verbose` flag to `session-context.sh` for expanded output when explicitly requested.

- **FR-S.3**: Add hook registration validation to `swap-team.sh` that warns on invalid matchers or timeouts.

### Could Have

- **FR-C.1**: Create hooks for `security-pack` demonstrating security-specific hooks (e.g., credential scanning on Write).

- **FR-C.2**: Add `just hooks:validate` task to validate all hook registrations against schema.

- **FR-C.3**: Progressive disclosure for session context using skill reference (e.g., "Run `/status` for full details").

## Non-Functional Requirements

- **NFR-1**: Performance - Hook execution overhead must remain under 100ms for all hooks combined at session start.

- **NFR-2**: Backwards Compatibility - Existing `.claude/hooks/` installations must continue working during migration (deprecation period).

- **NFR-3**: Reliability - Missing team hooks directory must not cause `swap-team.sh` to fail (graceful degradation).

- **NFR-4**: Maintainability - Hook registration schema must be documented and validated.

- **NFR-5**: Consistency - All path changes must be atomic (single commit for rename to avoid broken intermediate state).

## Edge Cases

| Case | Expected Behavior |
|------|------------------|
| Team has no `hooks/` directory | Skip team hook merge, use base hooks only |
| Team hook has same name as base hook | Team hook overrides base hook (with warning) |
| Team `hooks.yaml` has invalid matcher | Log warning, skip that hook registration, continue |
| Base hook deleted but referenced in `base_hooks.yaml` | Log error, skip that registration |
| `settings.local.json` has manual hook edits | Preserve non-roster hooks, merge roster hooks |
| Hook timeout exceeds 60s Claude limit | Clamp to 60000ms with warning |
| Team switches during active session | Hooks swapped immediately, no session impact |
| Empty UserPromptSubmit after matcher added | Hook still registered but never fires (expected) |
| Multiple teams define same hook event | Both teams' hooks run (append, not override) |

## Success Criteria

- [ ] `roster/user-hooks/` exists and contains all hooks previously in `roster/hooks/`
- [ ] `swap-team.sh --dry-run 10x-dev-pack` shows hook merge and settings generation
- [ ] `settings.local.json` hooks section generated correctly after team swap
- [ ] Session start context is 15 lines or fewer (measured by line count)
- [ ] UserPromptSubmit fires only on `/` prefixed prompts (verified by logging)
- [ ] `delegation-check.sh` warning is 3 lines (verified by output inspection)
- [ ] ADR-0002 no longer mentions CEM managing artifacts
- [ ] `team-validator.sh` and `workflow-validator.sh` deleted from hooks directory
- [ ] All existing tests pass after path migration
- [ ] Manual test: swap teams, verify hooks work in both source and destination

## Dependencies and Risks

### Dependencies

| Dependency | Type | Owner | Status |
|------------|------|-------|--------|
| `swap-team.sh` modification | Internal | roster | Ready |
| Team pack schema update | Internal | roster | Ready |
| ADR-0002 access | Internal | roster | Ready |
| `settings.local.json` schema | External | Claude Code | Stable |

### Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Path rename breaks existing installations | Medium | High | Provide migration script, document in release notes |
| Settings.json merge corrupts existing config | Medium | High | Backup before modification, validate after write |
| Context reduction removes needed information | Low | Medium | Add `--verbose` flag for detailed output |
| Team hooks introduce security risks | Low | High | Document security implications, require explicit opt-in |
| Performance regression from hook merging | Low | Medium | Benchmark before/after, optimize if needed |

## Out of Scope

- User-level hook customization (beyond team packs)
- Hook dependency resolution (hooks running in guaranteed order beyond base-then-team)
- Hook versioning or rollback
- Graphical hook configuration UI
- Remote hook loading from URLs
- Hook caching or optimization beyond current architecture
- Changes to Claude Code's hook execution model

## Open Questions

*None remaining - all questions resolved during stakeholder discussion.*

---

## File Verification

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-hook-ecosystem-parity.md` | Created |

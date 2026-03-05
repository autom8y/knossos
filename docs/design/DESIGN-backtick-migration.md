# Design: Backtick Injection Migration

## Status

Investigation complete. No code changes -- design document only.

## Problem

Six dromena files use `!` backtick injections to execute shell commands at command-load time. These injections run every time Claude loads the command, creating:

1. **Performance overhead**: Shell spawns on every command invocation
2. **Fragility**: Commands fail silently with fallback strings when tools are missing
3. **Security surface**: Arbitrary shell execution in command templates
4. **Testability gap**: Backtick output is not mockable or testable in isolation

## Current State Inventory

### 1. `mena/operations/code-review/INDEX.dro.md`

| Injection | Command | Purpose |
|-----------|---------|---------|
| Recent PRs | `gh pr list --limit 5 2>/dev/null \|\| echo "gh CLI not available"` | Show recent PRs for review context |

**Injection count**: 1
**Dependency**: `gh` CLI

### 2. `mena/operations/pr/INDEX.dro.md`

| Injection | Command | Purpose |
|-----------|---------|---------|
| Base branch | `git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null \| sed 's@^refs/remotes/origin/@@' \|\| echo "main"` | Detect default branch |
| Commits ahead | `git rev-list --count HEAD ^origin/main 2>/dev/null \|\| echo "unknown"` | Count divergent commits |
| Changed files | `git diff --name-only origin/main 2>/dev/null \| head -10 \|\| echo "none"` | List files changed |

**Injection count**: 3
**Dependency**: `git`

### 3. `mena/operations/commit/INDEX.dro.md`

| Injection | Command | Purpose |
|-----------|---------|---------|
| Staged files | `git diff --staged --name-only 2>/dev/null \| head -20 \|\| echo "none"` | List staged files |
| Unstaged changes | `git diff --name-only 2>/dev/null \| head -10 \|\| echo "none"` | List unstaged files |
| Untracked files | `git ls-files --others --exclude-standard 2>/dev/null \| head -10 \|\| echo "none"` | List untracked files |
| Branch | `git branch --show-current 2>/dev/null \|\| echo "detached"` | Current branch name |
| Last 3 commits | `git log --oneline -3 2>/dev/null \|\| echo "no commits"` | Recent commit history |

**Injection count**: 5
**Dependency**: `git`

### 4. `mena/session/handoff/INDEX.dro.md`

| Injection | Command | Purpose |
|-----------|---------|---------|
| Available agents | `ls .claude/agents/` | List agent files for handoff target validation |

**Injection count**: 1
**Dependency**: filesystem

### 5. `mena/navigation/rite.dro.md`

| Injection | Command | Purpose |
|-----------|---------|---------|
| Available rites | `ls ${KNOSSOS_HOME:-~/Code/knossos}/rites/` | List rite directories |

**Injection count**: 1
**Dependency**: `$KNOSSOS_HOME` or filesystem

### 6. `mena/navigation/consult/INDEX.dro.md`

| Injection | Command | Purpose |
|-----------|---------|---------|
| Active rite | `cat .knossos/ACTIVE_RITE 2>/dev/null \|\| echo "none"` | Read current rite |
| Available rites | `ls ${KNOSSOS_HOME:-~/Code/knossos}/rites/ 2>/dev/null \| tr '\n' ' '` | List rite names |

**Injection count**: 2
**Dependency**: filesystem, `$KNOSSOS_HOME`

### Summary

| File | Injections | Category |
|------|-----------|----------|
| code-review | 1 | GitHub context |
| pr | 3 | Git branch context |
| commit | 5 | Git staging context |
| handoff | 1 | Agent discovery |
| rite | 1 | Rite discovery |
| consult | 2 | Rite state |
| **Total** | **13** | |

## Proposed Hook-Based Alternatives

### Strategy A: SessionStart context enrichment

The existing SessionStart hook (`ari hook context`) already runs on every session start. Extend it to gather and cache common context:

- **Git context** (branch, staged files, recent commits): Gathered once, cached in session state
- **Rite context** (active rite, available rites): Already available via `resolver.ReadActiveRite()`
- **Agent context** (available agents): Directory listing at session start

**Pros**: Single execution point, cached results, testable
**Cons**: Stale data if git state changes mid-session (acceptable for context seeding)

### Strategy B: UserPromptSubmit hook for per-command context

A new `UserPromptSubmit` hook that detects which command was invoked and gathers command-specific context just-in-time.

**Pros**: Fresh data per invocation, command-specific
**Cons**: Higher complexity, latency on every prompt, harder to implement

### Strategy C: Skill-based context (legomena)

Create legomena that Claude loads autonomously when it detects the need:

- `git-context.lego.md` -- instructs Claude to run git commands in the Behavior section
- `rite-context.lego.md` -- instructs Claude to read rite state from files

**Pros**: No hook changes needed, Claude handles timing
**Cons**: Shifts execution to tool calls (adds to tool budget), non-deterministic

### Recommended Approach: Strategy A + C hybrid

1. **Phase 1**: Move git-invariant context (branch name, base branch) to SessionStart hook output
2. **Phase 2**: Replace volatile context (staged files, commits ahead) with in-command Bash tool calls documented in the Behavior section
3. **Phase 3**: Replace filesystem context (agent list, rite list) with `ari` CLI queries (`ari rite list --json`, `ari agent list --json`)

This preserves the "context available on load" pattern for stable data while moving volatile data to explicit tool calls that Claude executes in the Behavior section.

## Migration Order

Ordered by ease of migration (simplest first):

### Tier 1: Direct file reads (trivial)

1. **consult -- Active rite**: Replace `!cat .knossos/ACTIVE_RITE` with SessionStart hook output (already provides `rite` field)
2. **handoff -- Available agents**: Replace `!ls .claude/agents/` with Behavior step: "List available agents with `ls .claude/agents/`"
3. **rite -- Available rites**: Replace `!ls .../rites/` with `ari rite list` call in Behavior section

**Effort**: Low (2-3 hours). Remove backtick, add instruction to Behavior section.

### Tier 2: Git state (moderate)

4. **commit -- All 5 injections**: Move to Behavior step 1: "Check git state using `git status`, `git diff --staged --name-only`, etc." These are already described in the Behavior section as pre-flight checks.
5. **pr -- 3 injections**: Move to Behavior step 1 "Analyze changes". The PR command already runs these as part of its workflow.

**Effort**: Medium (4-6 hours). Remove backticks, verify Behavior sections already cover the commands.

### Tier 3: External tool (minor risk)

6. **code-review -- gh pr list**: Move to Behavior pre-flight. Risk: `gh` CLI may not be installed. Already has fallback `|| echo "gh CLI not available"`, so the command handles absence.

**Effort**: Low (1-2 hours). Remove backtick, add to Pre-flight section.

## Risks and Backward Compatibility

### Risk 1: Context no longer visible at load time

Currently, backtick output appears in the command prompt before Claude processes it. After migration, the context appears only after Claude executes tool calls.

**Mitigation**: For commands that critically depend on pre-execution context (pr, commit), ensure the Behavior section's first step gathers this context before any decisions.

### Risk 2: Increased tool call count

Moving 13 shell commands from backtick injections to explicit tool calls increases the tool budget per command invocation.

**Mitigation**: Many of these commands are already duplicated in the Behavior section (commit and pr both describe running git commands). Removing the backtick injection actually reduces redundancy.

### Risk 3: SessionStart hook latency

Adding git context gathering to the SessionStart hook increases its execution time.

**Mitigation**: Only gather stable context (branch name, base branch) in the hook. Volatile context (staged files) should remain as in-command tool calls. Target: <50ms additional latency.

## Non-Goals

- Changing command behavior or output
- Adding new CLI commands for context gathering (defer to existing `ari` capabilities)
- Modifying the backtick injection engine itself (CC feature, not ours to change)

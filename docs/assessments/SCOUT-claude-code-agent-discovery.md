# SCOUT-claude-code-agent-discovery

## Executive Summary

Claude Code discovers agents at **session startup only**, loading from both `~/.claude/agents/` (user-level) and `.claude/agents/` (project-level) into a per-session cache. The observed agent accumulation bug is caused by a **combination of Claude Code's session-level caching and our swap-rite.sh correctly modifying files but having no mechanism to invalidate Claude Code's in-memory agent registry**. The filesystem is correct after swap; Claude Code simply does not re-scan mid-session.

## Technology Overview

- **Category**: Agent Discovery Mechanism (Claude Code internal behavior)
- **Maturity**: Mainstream (part of Claude Code v1.x/v2.x)
- **License**: Anthropic proprietary
- **Backing**: Anthropic

## Agent Discovery Mechanism

### How Claude Code Finds Agents

1. **Scan Locations** (in priority order):
   - `.claude/agents/*.md` - Project-level (highest priority)
   - CLI `--agents` JSON - Inline definitions (mid priority)
   - `~/.claude/agents/*.md` - User-level (lower priority)
   - Plugin agents - From installed plugins (lowest priority)

2. **Scan Timing**:
   - Agents are discovered **at session startup only**
   - New files added mid-session require a **full restart** to be detected
   - The `/agents` command creates agents immediately without restart (uses different code path)

3. **Per-Session Cache**:
   - Claude Code maintains an **in-memory registry** of discovered agents
   - This registry is **never re-scanned** during the session
   - There is **no filesystem watcher** that detects changes to `.claude/agents/`

4. **Priority Resolution**:
   - When same-named agents exist at multiple levels, project-level wins
   - This means if `orchestrator.md` exists in both locations, only project-level is used

### Observed Behavior Explained

**User observation**: After `/forge` request, `/agents` shows 10+ agents when only 5 should exist.

**Root cause**:
1. User started session with Team A (5 agents) loaded into memory
2. Team A agents in `.claude/agents/` were discovered at session start
3. User-level agents in `~/.claude/agents/` (4 agents: requirements-analyst, context-engineer, technology-scout, consultant) were also discovered
4. User switched to Team B via swap-rite.sh
5. swap-rite.sh correctly replaced files in `.claude/agents/`
6. Claude Code's in-memory cache still contains **all agents from session start**
7. No mechanism exists to tell Claude Code "re-scan the agents directory"

## swap-rite.sh Analysis

### What It Does Correctly

1. **File Operations**: Properly backs up, removes, and copies agent files
2. **Manifest Tracking**: Maintains `AGENT_MANIFEST.json` with source/origin metadata
3. **Orphan Handling**: Detects and prompts for disposition of orphan agents
4. **Commands Sync**: Backs up, removes, and copies team commands via `.rite-commands` marker
5. **Skills Sync**: Backs up, removes, and copies team skills via `.rite-skills` marker
6. **CLAUDE.md Update**: Updates the Quick Start table and Agent Configurations section
7. **ACTIVE_RITE State**: Updates `.claude/ACTIVE_RITE` file correctly

### What It Cannot Do (Claude Code Limitation)

1. **Cannot invalidate Claude Code's agent cache** - No API exists for this
2. **Cannot force agent re-discovery** - Would require session restart
3. **Cannot signal Claude Code** - No IPC mechanism available

### Filesystem State After Swap

After running `swap-rite.sh 10x-dev-pack` on skeleton_claude:

```
.claude/agents/          # Contains exactly 5 files (CORRECT)
  architect.md
  orchestrator.md
  principal-engineer.md
  qa-adversary.md
  requirements-analyst.md

~/.claude/agents/        # Contains 4 files (SEPARATE, also loaded)
  consultant.md
  context-engineer.md
  requirements-analyst.md
  technology-scout.md

AGENT_MANIFEST.json      # Shows 5 agents, team=10x-dev-pack (CORRECT)
ACTIVE_RITE              # Shows "10x-dev-pack" (CORRECT)
```

The filesystem is **correct**. The bug is that Claude Code's in-memory state doesn't match.

## Risk Analysis

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| User confusion from stale agent list | High | Medium | Document restart requirement |
| Invoking wrong agent after swap | Medium | Medium | Use `/agents` command to verify |
| Orphan agents from user-level polluting project | Medium | Low | Clear user-level before swaps |
| Same-name agents causing priority confusion | High | Medium | Avoid duplicating names across levels |

## Fit Assessment

- **Philosophy Alignment**: Poor - We want dynamic rite switching; Claude Code assumes static agent sets
- **Stack Compatibility**: Partial - Filesystem operations work; cache invalidation impossible
- **Team Readiness**: High awareness needed - Users must understand restart requirement

## Comparison Matrix

| Criteria | Current swap-rite.sh | Restart After Swap | Clear User Agents |
|----------|---------------------|-------------------|-------------------|
| Filesystem Correct | Yes | Yes | Yes |
| Claude Code Sees Changes | No | Yes | Partially |
| User Experience | Confusing | Disruptive | Limiting |
| Implementation Effort | Done | Documentation | Script update |

## Root Cause Summary

| Component | Status | Notes |
|-----------|--------|-------|
| swap-rite.sh file operations | Working | Filesystem is correct |
| AGENT_MANIFEST.json | Working | Tracks agent provenance |
| CLAUDE.md updates | Working | Reflects new team |
| Claude Code agent scan | One-time | Only at session start |
| Claude Code cache invalidation | Not possible | No API exists |
| User-level agent isolation | Missing | Always loaded alongside project |

## Recommendation

**Verdict**: Assess (investigate workarounds before accepting limitation)

**Rationale**: The bug is a Claude Code limitation, not a swap-rite.sh defect. However, we have options:

1. **Document the restart requirement** - Immediate, low-effort
2. **Explore `/agents` command automation** - May allow programmatic refresh
3. **Advocate for Claude Code API** - Report issue requesting cache invalidation
4. **User-level agent isolation** - Clear `~/.claude/agents/` during swaps (breaking)

## Recommended Fixes

### Immediate (swap-rite.sh)

1. **Add post-swap warning**:
   ```bash
   log "IMPORTANT: Restart Claude Code session for agent changes to take effect"
   ```

2. **Add `--fresh` flag** for CI/scripted workflows:
   ```bash
   swap-rite.sh 10x-dev-pack --fresh  # Clears user-level agents too
   ```

3. **Improve same-name detection**:
   ```bash
   # Warn when project agent name matches user-level agent
   for agent in .claude/agents/*.md; do
     if [[ -f "$HOME/.claude/agents/$(basename $agent)" ]]; then
       log_warning "Agent $(basename $agent) exists at both levels"
     fi
   done
   ```

### Medium-term (Documentation)

1. Update team swap documentation with restart requirement
2. Add troubleshooting section for "stale agents" symptom
3. Document user-level vs project-level agent behavior

### Long-term (Feature Request)

1. Request Claude Code API for agent cache invalidation
2. Request filesystem watcher for `.claude/agents/` directory
3. Request `/agents refresh` command

## Known Claude Code Issues

- [Issue #9930](https://github.com/anthropics/claude-code/issues/9930) - Agent auto-discovery not working
- [Issue #4773](https://github.com/anthropics/claude-code/issues/4773) - Custom agents not discovered
- [Issue #5763](https://github.com/anthropics/claude-code/issues/5763) - Markdown files not recognized

## Verification Checklist

- [x] Technology researched with multiple sources cited
- [x] Maturity rated with supporting evidence (per-session caching is documented behavior)
- [x] Risks identified, rated, and quantified
- [x] Fit with current stack evaluated
- [x] Comparison matrix includes status quo and alternatives
- [x] Clear recommendation provided (Assess)
- [x] Root cause identified (Claude Code session-level caching + no invalidation API)

## Sources

- [Claude Code Subagents Documentation](https://code.claude.com/docs/en/sub-agents)
- [GitHub Issue #9930 - Agent Auto-Discovery Not Working](https://github.com/anthropics/claude-code/issues/9930)
- [GitHub Issue #4773 - Custom agents not discovered](https://github.com/anthropics/claude-code/issues/4773)
- [Claude Code Customization Guide](https://alexop.dev/posts/claude-code-customization-guide-claudemd-skills-subagents/)

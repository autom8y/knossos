# Sprint 3 Task 022: User-Agents Migration Audit

**Date**: 2026-01-03
**Task**: Migrate skeleton user-agents to roster
**Status**: COMPLETE

## Source and Target

| Location | Path |
|----------|------|
| Source | `~/Code/skeleton_claude/.claude/user-agents/` |
| Target | `/Users/tomtenuta/Code/roster/.claude/user-agents/` |

## User-Agents Evaluated

| Agent | Source Lines | Action | Path Updates | Status |
|-------|-------------|--------|--------------|--------|
| agent-curator.md | 292 | COPIED | None required | Verified |
| agent-designer.md | 221 | COPIED | None required | Verified |
| consultant.md | 256 | COPIED | None required | Verified |
| eval-specialist.md | 287 | COPIED | None required | Verified |
| platform-engineer.md | 254 | COPIED | None required | Verified |
| prompt-architect.md | 303 | COPIED | None required | Verified |
| workflow-engineer.md | 291 | COPIED | None required | Verified |

**Total**: 7 user-agents migrated (1,904 lines total)

## Path Replacement Analysis

The following patterns were searched for in migrated files:

| Pattern | Found | Action |
|---------|-------|--------|
| `skeleton_claude` | No | N/A |
| `SKELETON_HOME` | No | N/A |
| `~/Code/roster/` | Yes (intended) | No change needed |
| `ROSTER_HOME` | Yes (intended) | No change needed |

**Conclusion**: The user-agents were already authored for the roster ecosystem. No path replacements were necessary.

## Agent Descriptions (Forge Team)

These 7 agents constitute the "Forge" team - specialized agents for creating and managing team packs:

1. **agent-curator**: Finalizes teams for roster deployment, syncs Consultant knowledge base, handles versioning and documentation
2. **agent-designer**: Designs agent roles, boundaries, and contracts for new teams; produces TEAM-SPEC documents
3. **consultant**: Meta-level ecosystem navigator providing team routing and workflow guidance
4. **eval-specialist**: Validates teams and agents before production use with 29-point validation checklist
5. **platform-engineer**: Infrastructure specialist implementing team packs in the roster system
6. **prompt-architect**: Crafts agent identities and system prompts following 11-section template
7. **workflow-engineer**: Wires agents into cohesive workflows, creates workflow.yaml configurations

## Workflow Position

```
Agent Designer -> Prompt Architect -> Workflow Engineer -> Platform Engineer -> Eval Specialist -> Agent Curator
     |                |                    |                    |                   |                  |
     v                v                    v                    v                   v                  v
  TEAM-SPEC    Agent .md files      workflow.yaml         Deployed team      eval-report.md     Consultant sync
```

## Verification

### Files Present
```
/Users/tomtenuta/Code/roster/.claude/user-agents/
  agent-curator.md      (292 lines)
  agent-designer.md     (221 lines)
  consultant.md         (256 lines)
  eval-specialist.md    (287 lines)
  platform-engineer.md  (254 lines)
  prompt-architect.md   (303 lines)
  workflow-engineer.md  (291 lines)
```

### Content Verification
- All files copied successfully
- No diff between source and target
- No skeleton-specific paths found in migrated files
- All agents use appropriate roster paths (`~/Code/roster/`)

## Notes

1. **No Conflicts**: Roster did not have a user-agents directory prior to migration
2. **Already Roster-Ready**: The skeleton user-agents were authored with roster paths, indicating they were designed for the consolidated ecosystem
3. **Forge Team Complete**: All 7 Forge team agents are now available in roster for team development workflows

## Attestation Table

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| agent-curator.md | /Users/tomtenuta/Code/roster/.claude/user-agents/agent-curator.md | YES |
| agent-designer.md | /Users/tomtenuta/Code/roster/.claude/user-agents/agent-designer.md | YES |
| consultant.md | /Users/tomtenuta/Code/roster/.claude/user-agents/consultant.md | YES |
| eval-specialist.md | /Users/tomtenuta/Code/roster/.claude/user-agents/eval-specialist.md | YES |
| platform-engineer.md | /Users/tomtenuta/Code/roster/.claude/user-agents/platform-engineer.md | YES |
| prompt-architect.md | /Users/tomtenuta/Code/roster/.claude/user-agents/prompt-architect.md | YES |
| workflow-engineer.md | /Users/tomtenuta/Code/roster/.claude/user-agents/workflow-engineer.md | YES |
| Migration audit | /Users/tomtenuta/Code/roster/docs/audits/sprint3-task022-user-agents-migration.md | YES |

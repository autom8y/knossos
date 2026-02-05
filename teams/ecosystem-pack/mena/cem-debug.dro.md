---
description: Diagnose CEM sync issues and conflicts (Ecosystem Analyst with CEM focus)
allowed-tools: Bash, Read, Grep, Glob
model: opus
---

## Context
Auto-injected by SessionStart hook (project, team, session, git).

## Your Task

Invoke the Ecosystem Analyst agent with CEM diagnostic focus to trace sync failures, conflicts, or unexpected behavior to their root cause.

## Behavior

1. **Verify ecosystem-pack is active**:
   - If not already on ecosystem-pack, execute: `~/Code/roster/swap-team.sh ecosystem-pack`

2. **Invoke Ecosystem Analyst** with this focus:
   - Reproduce the reported CEM sync issue
   - Examine CEM sync logs and error messages
   - Trace conflict detection and resolution logic
   - Identify which component is failing (sync, merge, conflict resolution, file handling)
   - Check settings schema compatibility between satellite and roster
   - Verify CEM version compatibility

3. **Produce Gap Analysis** documenting:
   - Exact reproduction steps
   - Root cause (specific CEM component and line of code if possible)
   - Success criteria (what "fixed" looks like)
   - Affected systems (CEM only, or CEM + skeleton + roster?)

## When to Use

- `cem sync` fails with error messages
- Merge conflicts appear unexpectedly
- Settings not propagating from roster to satellite
- Sync succeeds but files are missing/corrupted
- Performance degradation in sync operations
- Need to trace CEM behavior through source code

## CEM Diagnostic Checklist

The Ecosystem Analyst will examine:
- [ ] CEM version in satellite vs roster
- [ ] Settings schema compatibility (`.claude/settings.json` format)
- [ ] Conflict detection logic (does it correctly identify conflicts?)
- [ ] Merge algorithm (3-way merge working correctly?)
- [ ] File permissions and ownership
- [ ] Git state (is repo in clean state?)
- [ ] Lock files or concurrent sync attempts
- [ ] Roster registry correctness (are paths valid?)

## Expected Output

**Gap Analysis** document at: `docs/ecosystem/GAP-{issue-slug}.md`

Contains:
- **Problem**: User-reported issue description
- **Reproduction**: Exact steps to reproduce
- **Root Cause**: Specific CEM component and failure mode
- **Success Criteria**: Measurable fix definition
- **Affected Systems**: CEM only, or broader impact?
- **Next Steps**: Design phase needed, or direct fix?

## Handoff

After Gap Analysis is complete:
- **Simple fix** (PATCH complexity): Hand off to Integration Engineer
- **Design needed** (MODULE+): Hand off to Context Architect
- **User clarification needed**: Return Gap Analysis draft for review

## Reference

Full workflow: `.claude/skills/ecosystem-ref/INDEX.md`
CEM source: `~/Code/roster/cem` (bash scripts, sync logic)

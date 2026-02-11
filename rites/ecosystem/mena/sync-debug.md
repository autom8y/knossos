---
description: Diagnose sync issues and conflicts (Ecosystem Analyst with sync pipeline focus)
argument-hint: "[issue-description]"
allowed-tools: Bash, Read, Grep, Glob
model: opus
---

## Context
Auto-injected by SessionStart hook (project, rite, session, git).

## Your Task

Invoke the Ecosystem Analyst agent with sync pipeline diagnostic focus to trace sync failures, conflicts, or unexpected behavior to their root cause.

## Behavior

1. **Verify ecosystem is active**:
   - If not already on ecosystem, execute: `ari sync --rite ecosystem`

2. **Invoke Ecosystem Analyst** with this focus:
   - Reproduce the reported sync issue
   - Examine sync logs and error messages
   - Trace conflict detection and resolution logic
   - Identify which component is failing (sync, merge, conflict resolution, file handling)
   - Check settings schema compatibility between satellite and knossos
   - Verify provenance manifest compatibility

3. **Produce Gap Analysis** documenting:
   - Exact reproduction steps
   - Root cause (specific sync component and line of code if possible)
   - Success criteria (what "fixed" looks like)
   - Affected systems (sync pipeline only, or sync + knossos?)

## When to Use

- `ari sync` fails with error messages
- Merge conflicts appear unexpectedly
- Settings not propagating from knossos to satellite
- Sync succeeds but files are missing/corrupted
- Performance degradation in sync operations
- Need to trace sync behavior through source code

## Sync Diagnostic Checklist

The Ecosystem Analyst will examine:
- [ ] Sync pipeline version in satellite vs knossos
- [ ] Settings schema compatibility (`.claude/settings.json` format)
- [ ] Conflict detection logic (does it correctly identify conflicts?)
- [ ] Merge algorithm (3-way merge working correctly?)
- [ ] File permissions and ownership
- [ ] Git state (is repo in clean state?)
- [ ] Lock files or concurrent sync attempts
- [ ] Knossos registry correctness (are paths valid?)

## Expected Output

**Gap Analysis** document at: `docs/ecosystem/GAP-{issue-slug}.md`

Contains:
- **Problem**: User-reported issue description
- **Reproduction**: Exact steps to reproduce
- **Root Cause**: Specific sync component and failure mode
- **Success Criteria**: Measurable fix definition
- **Affected Systems**: Sync pipeline only, or broader impact?
- **Next Steps**: Design phase needed, or direct fix?

## Handoff

After Gap Analysis is complete:
- **Simple fix** (PATCH complexity): Hand off to Integration Engineer
- **Design needed** (MODULE+): Hand off to Context Architect
- **User clarification needed**: Return Gap Analysis draft for review

## Reference

Full workflow: `.claude/skills/ecosystem-ref/INDEX.md`
Sync source: `internal/materialize/` (Go implementation)

---
name: pr
description: Create pull request with comprehensive description
argument-hint: "[title] [--base=BRANCH]"
allowed-tools: Bash, Read, Glob, Grep
model: sonnet
disable-model-invocation: true
context: fork
---

## Context
Auto-injected by SessionStart hook (project, rite, session, git).

Base branch is available in session context (`base_branch` field).

## Your Task

Create a GitHub pull request with auto-generated description. $ARGUMENTS

## Behavior

1. **Analyze changes** (gather fresh volatile context):
   - Count divergent commits with `git rev-list --count HEAD ^origin/<base_branch>`
   - List changed files with `git diff --name-only origin/<base_branch>`
   - Review all commits since branch divergence
   - Identify related PRD/TDD/ADR artifacts

2. **Generate PR description**:
   ```markdown
   ## Summary
   [Auto-generated from commits and artifacts]

   ## Changes
   - [List of key changes]

   ## Test Plan
   - [Derived from test plan if exists]

   ## Artifacts
   - PRD: [link if exists]
   - TDD: [link if exists]

   ## Checklist
   - [ ] Tests pass
   - [ ] Code reviewed
   - [ ] Documentation updated
   ```

3. **Create PR** via `gh pr create`:
   ```bash
   gh pr create --title "..." --body "..."
   ```

4. **Return PR URL**

## Example

```
/pr "Add user authentication"
/pr --base=develop
```

## Reference

Full documentation: `.claude/commands/operations/pr/INDEX.md`

---
name: pr-ref
description: "Pull request workflow with summary, test plan, and checklist. Use when: feature is complete and tested, ready to ship to production, need formal code review. Triggers: /pr, create pull request, open PR, ship feature."
---

# /pr - Pull Request Workflow

> Create pull requests with comprehensive descriptions, test plans, and artifact links.

## Decision Tree

```
Ready to ship?
├─ Feature complete + tested → /pr
├─ On feature branch → /pr
├─ On main branch → Switch to feature branch first
├─ Uncommitted changes → /commit first
├─ Multiple features → Create separate PRs
└─ Still implementing → Finish /task or /build first
```

## Usage

```bash
/pr "pr-title" [--base=BRANCH]
```

### Parameters

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `pr-title` | Yes | - | Pull request title |
| `--base` | No | main | Base branch to merge into |

## Quick Reference

**Pre-flight**:
- Feature implementation complete and tested
- On feature branch (not main/master)
- All changes committed
- Ready for code review

**Actions**:
1. Validate git state (on feature branch, no uncommitted changes)
2. Analyze ALL commits since branch diverged
3. Generate comprehensive PR description (summary, changes, test plan, links)
4. Push branch to remote (if needed)
5. Create PR via GitHub CLI
6. Display PR URL

**Produces**:
- GitHub Pull Request with structured description
- Test plan checklist for reviewers
- Links to PRD, TDD, ADRs (if exist)
- AI attribution in PR description (unlike `/commit`)

**Never Produces**:
- Local files (only GitHub PR)
- Git commits (must commit first)

## Anti-Patterns

| Do NOT | Why | Instead |
|--------|-----|---------|
| PR from main branch | Can't merge to self | Switch to feature branch |
| PR with uncommitted changes | Incomplete work | /commit first |
| PR before testing | Reviewers find bugs | /qa validation first |
| PR without description | Poor reviewer context | Let `/pr` generate description |
| Skip PR review | Miss bugs/issues | Request reviewers, wait for approval |

## Prerequisites

- Git repository initialized
- On feature branch (not main/master)
- All changes committed (no uncommitted work)
- GitHub CLI (`gh`) installed and authenticated
- Remote repository exists
- Tests passing (recommended)

## Success Criteria

- PR created on GitHub
- Description includes summary, test plan, links to artifacts
- PR URL returned
- Branch pushed to remote (if wasn't already)
- Ready for review

## Related Commands

| Command | When to Use |
|---------|-------------|
| `/task` | Build feature before PR (prerequisite) |
| `/qa` | Validate before PR (recommended) |
| `/commit` | Commit changes before PR (prerequisite) |
| `/hotfix` | Urgent fixes (can skip full workflow) |

## Progressive Disclosure

- [behavior.md](behavior.md) - Full step-by-step sequence, PR description template, attribution policy
- [examples.md](examples.md) - 4 workflow scenarios, edge cases, troubleshooting
- [../shared-sections/git-validation.md](../shared-sections/git-validation.md) - Git state validation pattern

## Attribution Policy

**Unlike `/commit`**, `/pr` **DOES include** AI attribution in the PR description:

```markdown
---
Generated with [Claude Code](https://claude.com/claude-code)
```

**Why?** PR descriptions are metadata (editable, visible to reviewers) rather than permanent git history. Git commits use user-only attribution; PR descriptions acknowledge AI assistance.

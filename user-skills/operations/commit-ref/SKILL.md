---
name: commit-ref
description: "AI-assisted git commits with message generation and session tracking. Use when: committing staged changes, wanting consistent conventional commit format, needing smart staging suggestions. Triggers: /commit, commit changes, git commit, stage and commit."
---

# /commit - AI-Assisted Git Commits

> Create git commits with AI-generated messages following conventional commit format.

## Decision Tree

```
Ready to commit?
├─ Changes staged → /commit
├─ Stage all changes → /commit --all
├─ Custom message → /commit --message="..."
├─ No changes yet → Make changes first
└─ Merge conflict → Resolve conflicts first
```

## Usage

```bash
/commit [--all] [--message="override"]
```

### Parameters

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `--all` | No | false | Stage all changes before commit (`git add -A`) |
| `--message` | No | - | Override AI-generated message |

## Quick Reference

**Pre-flight**:
- In git repository
- Changes exist (staged or unstaged)
- No merge conflicts
- Not in detached HEAD (recommended)

**Actions**:
1. Validate git state (repo exists, no conflicts)
2. Smart staging (if nothing staged and no `--all`)
3. Analyze staged changes (diff, stats, file types)
4. Generate conventional commit message
5. Present for confirmation (Y/n/edit)
6. Execute commit (user-only attribution)
7. Track to session (if active)

**Produces**:
- Git commit (user-only authorship)
- Session tracking entry (if session active)

**Never Produces**:
- AI attribution in git history
- Files on disk (only git commits)

## Anti-Patterns

| Do NOT | Why | Instead |
|--------|-----|---------|
| Mix unrelated changes | Poor git history | Stage related changes only |
| Commit during merge conflict | Git blocks it | Resolve conflicts first |
| Skip confirmation | Might generate poor message | Review and edit if needed |
| Bypass pre-commit hooks | Defeats quality checks | Fix issues, re-run `/commit` |
| Commit without reviewing diff | Might commit unwanted changes | Check `git diff --staged` |

## Prerequisites

- Git repository initialized
- Changes exist (staged or unstaged)
- No unresolved merge conflicts
- Git configured (user.name, user.email)

## Success Criteria

- Commit created in git history
- Message follows conventional commits format
- User is sole author (no AI attribution)
- Session tracking updated (if active)
- Pre-commit hooks passed

## Related Commands

| Command | When to Use |
|---------|-------------|
| `/pr` | Create pull request (after commits ready) |
| `/start` | Begin session (enables commit tracking) |
| `/wrap` | End session (includes commit summary) |

## Progressive Disclosure

- [behavior.md](behavior.md) - Full step-by-step sequence, attribution policy, hook integration
- [examples.md](examples.md) - 4 usage scenarios, edge cases, troubleshooting
- [../shared-sections/git-validation.md](../shared-sections/git-validation.md) - Git state validation pattern

## Attribution Policy

**CRITICAL**: Unlike `/pr` which adds AI attribution to PR descriptions, `/commit` produces commits with **USER-ONLY** attribution. No `Co-Authored-By`, `Generated with`, or any AI markers appear in git history.

**Rationale**: The user wrote/reviewed code, staged changes, and confirmed the message. The AI assistance in message generation is analogous to IDE autocomplete - helpful but not authorship.

**Verification**:
```bash
git log -1 --format='Author: %an <%ae>'  # Should show user only
git log -1 --format=%B | grep -i claude  # Should return empty
```

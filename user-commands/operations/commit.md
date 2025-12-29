---
description: Create git commit with AI-generated message
argument-hint: [--all] [--message="override"]
allowed-tools: Bash, Read, Glob, Grep
model: sonnet
---

## Context
Auto-injected by SessionStart hook (project, team, session, git).

**Commit-specific**:
- Staged files: !`git diff --staged --name-only 2>/dev/null | head -20 || echo "none"`
- Unstaged changes: !`git diff --name-only 2>/dev/null | head -10 || echo "none"`
- Untracked files: !`git ls-files --others --exclude-standard 2>/dev/null | head -10 || echo "none"`
- Branch: !`git branch --show-current 2>/dev/null || echo "detached"`
- Last 3 commits: !`git log --oneline -3 2>/dev/null || echo "no commits"`

## Pre-flight

1. **Repository check**:
   - Verify in git repository
   - If not: ERROR "Not in a git repository."

2. **Git state**:
   - Check for staged changes: `git diff --cached --quiet`
   - If no staged changes: Check for unstaged changes
   - If neither: ERROR "Nothing to commit. Stage changes with git add first."

## Your Task

Create a git commit with an AI-generated message. $ARGUMENTS

## Behavior

1. **Check git state**:
   - Verify in git repository
   - Check for staged changes
   - If nothing staged and no `--all` flag, enter staging flow

2. **Smart staging** (if nothing staged):
   - Display modified and untracked files
   - Suggest files to stage based on context
   - Ask user: "Stage these files? [Y/n/select]"
   - If `--all` provided, run `git add -A`

3. **Analyze staged changes**:
   ```bash
   git diff --staged --stat
   git diff --staged
   ```

4. **Generate commit message**:
   - Format: `type(scope): subject`
   - Types: feat, fix, docs, style, refactor, test, chore, perf, ci, build
   - Subject: imperative mood, max 50 characters
   - Body (if complex): wrap at 72 characters
   - No AI attribution markers

5. **Present for confirmation**:
   ```
   Proposed commit:
   ────────────────────────────────────
   feat(auth): add JWT refresh token support

   - Implement token rotation on expiry
   - Add configurable expiration window
   - Update middleware to handle refresh
   ────────────────────────────────────

   Proceed? [Y/n/edit]
   ```

6. **Execute commit**:
   ```bash
   git commit -m "<message>"
   ```

   **CRITICAL**: Do NOT include:
   - `--trailer "Co-Authored-By: Claude"`
   - Any "Generated with" footer
   - Any AI attribution whatsoever

   The user is the sole author.

7. **Report success**:
   - Show commit hash
   - Show files changed summary
   - If session active, note that commit was tracked

## Parameters

| Parameter | Description |
|-----------|-------------|
| `--all` | Stage all changes before commit (`git add -A`) |
| `--message="..."` | Override AI-generated message |
| (no args) | Interactive mode with staging suggestions |

## Example

```
/commit
/commit --all
/commit --message="fix: resolve race condition in worker"
```

## Reference

Full documentation: `.claude/skills/commit-ref/skill.md`

## Attribution Policy

**CRITICAL**: This command creates commits with USER-ONLY attribution.

The user:
1. Reviewed and approved staged changes
2. Confirmed the commit message
3. Owns the intellectual work
4. Is the sole author in git history

Do NOT add `Co-Authored-By`, `Generated with`, or any AI markers.

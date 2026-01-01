---
name: review
description: "Code review workflow with structured feedback using QA Adversary perspective. Use when: reviewing pull requests, validating code before merge, providing structured feedback on others' code. Triggers: /review, code review, review PR, review code, feedback on PR, check branch, analyze changes, PR comments, look at PR."
---

# /review - Code Review Workflow

> **Category**: Development Workflows | **Phase**: Review | **Complexity**: Low

## Purpose

Perform structured code review of a pull request or branch using QA Adversary perspective. Analyzes code quality, correctness, security, and maintainability with actionable feedback.

**Use when:**
- Reviewing someone else's pull request
- Validating code before merge
- Need structured, categorized feedback (blocking/suggestions/nits)
- Want multi-lens review (functional, quality, security, performance)

**Output:** Structured review with recommendation (APPROVE / REQUEST CHANGES / COMMENT)

---

## Usage

```bash
/review PR_NUMBER       # Review GitHub PR by number
/review BRANCH_NAME     # Review branch vs main
/review --current       # Review current branch vs main
```

| Parameter | Description |
|-----------|-------------|
| `PR_NUMBER` | GitHub PR number (e.g., `142`) |
| `BRANCH_NAME` | Git branch to review against main |
| `--current` | Review current branch vs main |

---

## Behavior

This skill performs read-only analysis without modifying code.

### 1. Fetch Changes

Retrieve the diff to review:
- **PR number**: Use `gh pr view` for metadata, `git diff` for changes
- **Branch name**: Diff against `origin/main`
- **Current branch**: Diff HEAD against main

### 2. Gather Context

Collect review context: PR description, changed files list, commit messages, related test files.

### 3. QA Adversary Review

Delegate to QA Adversary perspective for multi-lens analysis:
- **Functional**: Does code work correctly? Edge cases handled?
- **Quality**: Readable, maintainable, follows project standards?
- **Security**: Input validation, auth, injection risks?
- **Performance**: Efficient algorithms, scales appropriately?
- **Testing**: Comprehensive coverage, error paths tested?
- **Architecture**: Fits system design, clean boundaries?

For the full review prompt template, see [references/review-prompt-template.md](references/review-prompt-template.md).

### 4. Generate Report

Structure feedback into categories:
- **Blocking Issues**: Must fix before merge (security, correctness bugs)
- **Strong Suggestions**: Should fix (maintainability, testing gaps)
- **Nits**: Nice to have (style, minor improvements)
- **Positive Feedback**: What was done well
- **Questions**: Clarifications needed

For output format details, see [references/output-format.md](references/output-format.md).

---

## Workflow

```mermaid
graph LR
    A[/review invoked] --> B{PR or Branch?}
    B -->|PR| C[Fetch via gh]
    B -->|Branch| D[Fetch via git]
    B -->|Current| E[Diff vs main]
    C --> F[Get Context]
    D --> F
    E --> F
    F --> G[QA Adversary Review]
    G --> H[Structure Feedback]
    H --> I{Blocking Issues?}
    I -->|Yes| J[REQUEST CHANGES]
    I -->|No| K{Questions?}
    K -->|Yes| L[COMMENT]
    K -->|No| M[APPROVE]
```

---

## When to Use

| Scenario | Use /review | Use /qa |
|----------|-------------|---------|
| Reviewing someone else's PR | Yes | No |
| Validating your own implementation | No | Yes |
| Pre-merge formal review | Yes | No |
| During development self-check | No | Yes |

Both use QA Adversary perspective; `/review` is external review, `/qa` is self-validation.

---

## Prerequisites & Complexity

**Requirements:**
- Git repository with remote
- GitHub CLI (`gh`) authenticated for PR reviews
- Code to review (PR or branch must exist)

**Complexity:** Low - read-only analysis, no code modification, produces structured feedback.

---

## Error Cases

| Error | Resolution |
|-------|------------|
| PR not found | Verify PR exists: `gh pr list` |
| Branch not found | Check branch name: `git branch -a` |
| gh not authenticated | Run: `gh auth login` |
| No changes | Branch identical to base; nothing to review |

---

## Large PRs

For PRs >500 lines, Claude may use extended thinking for deeper analysis. Consider requesting PR split for very large changes to enable more thorough review.

---

## References

- [QA Adversary Review Prompt](references/review-prompt-template.md) - Full invocation prompt template
- [Output Format Guide](references/output-format.md) - Recommended report structure
- [Full Review Example](references/examples/full-review.md) - Detailed example with blocking issues

---

## Related

- `/qa` - Validate your own implementation (different use case)
- `/pr` - Create pull request for review

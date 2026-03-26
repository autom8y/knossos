---
name: commit-conventions
description: |
  Releaser-specific commit conventions extending platform conventions.
  Use when: committing version bumps, release publishes, PR creation, or
  changelog entries in a releaser rite context.
  Triggers: release commit, version bump, chore(deps), chore(release),
  publish commit, dependency bump, bump consumer, auto-merge PR.
---

# Releaser Commit Conventions

> Extends platform `conventions` skill with release-specific commit formats.
> Consumed by: Release-Executor (executes), Release-Planner (references for plan output).

## Quick-Reference Commit Formats

| Action | Format | Example |
|--------|--------|---------|
| Dependency bump | `chore(deps): bump {dependency} to {version}` | `chore(deps): bump @acme/sdk to 2.1.0` |
| Package publish | `chore(release): publish {package} v{version}` | `chore(release): publish core-sdk v1.3.0` |

## Constraint Style — Preserve Existing Style

When bumping a consumer's dependency version, match the format already in that manifest.

| Style | Before | After |
|-------|--------|-------|
| exact | `1.2.3` | `1.3.0` |
| range | `>=1.2.0` | `>=1.3.0` |
| compatible | `^1.2.3` | `^1.3.0` |

**Rule**: read the consumer manifest first, then write the bump commit in the same style.

## PR Conventions

- Title: descriptive, names all affected packages
- Body: list repos bumped, versions changed, link to publish confirmation
- Auto-merge: `gh pr merge --auto --squash {pr-number}` — only when `auto_merge_pr` is set in plan

## Safety Rules

- NEVER publish without the release plan specifying the action
- NEVER skip CI checks or add `[skip ci]` to commits
- ALWAYS verify publish success before bumping consumers

## Related Skills

- `releaser-ref` skill — Full releaser workflow reference (phases, state map, dependency graph)
- `conventions` skill — Platform-wide commit and branch naming conventions

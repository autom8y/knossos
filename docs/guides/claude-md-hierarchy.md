# CLAUDE.md Hierarchy Guide

## Overview

Claude Code loads CLAUDE.md files from multiple directory levels, concatenating all content into every conversational turn. Understanding the ownership model prevents token waste through duplication.

## The 3-Level Hierarchy

| Level | File | Scope | When Loaded |
|-------|------|-------|-------------|
| **Global** | `~/CLAUDE.md` | All Claude Code sessions | Always |
| **Global Knossos** | `~/.claude/CLAUDE.md` | All projects | Always |
| **Directory** | `~/Code/.claude/CLAUDE.md` | All projects under ~/Code/ | When working in ~/Code/* |
| **Project** | `~/Code/roster/.claude/CLAUDE.md` | roster project only | When working in roster |

**Current token cost per turn**: ~7,950 tokens (all 4 files loaded on every turn in roster project)
**Target token cost per turn**: ~2,590 tokens (67% reduction)

### What Belongs at Each Level

**`~/CLAUDE.md` (Global root)**:
- Language-agnostic code quality principles
- Cross-project tool usage conventions
- Security best practices
- Git workflow rules
- Content that applies to **all** projects (not just Knossos ones)

**`~/.claude/CLAUDE.md` (Global Knossos)**:
- Minimal universal preferences (6-8 lines max)
- Should contain **no** Knossos-specific content (execution mode, agents, platform features)
- Example: "Use Go conventions for Go projects, prefer editing over creating files"

**`~/Code/.claude/CLAUDE.md` (Directory-level)**:
- Directory-scope preferences (5-8 lines max)
- Should contain **no** duplication of parent or child content
- Example: "Project-specific CLAUDE.md files take precedence"

**`~/Code/roster/.claude/CLAUDE.md` (Project-level)**:
- Knossos platform features (execution modes, agents, commands)
- Project-specific conventions and anti-patterns
- Should **never** repeat what parent files already say

### Why Duplication Wastes Tokens

**Current state**: Both `~/.claude/CLAUDE.md` and `~/Code/.claude/CLAUDE.md` contain full copies of knossos-managed sections (execution-mode, quick-start, agent-routing, etc.). Each file contributes ~1,700 tokens of **pure duplication** every turn.

**Token math**:
- Duplicated content: ~3,400 tokens/turn wasted
- Typical session: 30-100 turns
- Session waste: 102,000-340,000 tokens of duplicate loading

**Impact**: The same instructions about execution modes, agent routing, and platform features are read 3 times per turn (once from each CLAUDE.md file), when they should appear only in the project-level file.

## Current Parent File Assessment

### `~/.claude/CLAUDE.md` (7 lines, ~60 tokens)
**Status**: Already optimal. Contains only minimal cross-project preferences.
**Content**: Go conventions, editing preferences, no unnecessary docs, defer to project CLAUDE.md.
**Action**: No changes needed.

### `~/Code/.claude/CLAUDE.md` (5 lines, ~50 tokens)
**Status**: Already optimal. Contains only directory-scope preferences.
**Content**: Project-specific files take precedence, language-appropriate conventions, prefer stdlib.
**Action**: No changes needed.

### `~/CLAUDE.md` (143 lines, ~950 tokens)
**Status**: Out of scope. Contains general-purpose guidance for non-Knossos projects.
**Content**: Code quality standards (read before writing, type safety, error handling), testing philosophy, security practices, tool usage defaults, language-specific conventions (Python/JS/Go/Shell), git workflow, project detection.
**Duplication with project CLAUDE.md**: None. This file covers cross-project development practices, while project CLAUDE.md covers Knossos-specific platform features.
**Action**: No changes needed. This is independent guidance.

**Key finding**: Parent files were already cleaned up before this redesign. The ~7,950 token/turn cost comes primarily from the project-level CLAUDE.md (329 lines = ~3,100 tokens), not from parent file duplication. The major savings opportunity is in trimming the project file itself (covered in C1 design doc).

## Migration Runbook

### Prerequisites
- Back up all CLAUDE.md files before editing
- No uncommitted changes in sacred paths (`.claude/`, `docs/`)
- Confirm current token baseline (see verification step below)

### Step 1: Assess Current Parent Files
```bash
# Read current parent files
cat ~/.claude/CLAUDE.md
cat ~/Code/.claude/CLAUDE.md

# Count lines
wc -l ~/.claude/CLAUDE.md ~/Code/.claude/CLAUDE.md
```

**Verify**: If either file contains knossos-managed sections (execution-mode, quick-start, agent-routing, commands, agent-configurations, platform-infrastructure, navigation, slash-commands), proceed to Step 2. If both files contain only minimal preferences (as assessed above), skip to Step 4.

### Step 2: Back Up Parent Files
```bash
cp ~/.claude/CLAUDE.md ~/.claude/CLAUDE.md.backup.$(date +%Y%m%d)
cp ~/Code/.claude/CLAUDE.md ~/Code/.claude/CLAUDE.md.backup.$(date +%Y%m%d)
```

**Verify**: Backup files exist and contain full original content.

### Step 3: Trim Parent Files (if needed)
**Only required if parent files contain knossos-managed sections.**

For `~/.claude/CLAUDE.md`, keep only:
```markdown
# Global Preferences

- Use Go conventions (gofmt, golint) for Go projects
- Prefer editing existing files over creating new files
- No unnecessary documentation files unless requested
- When in a Knossos project, follow project-level CLAUDE.md instructions
```

For `~/Code/.claude/CLAUDE.md`, keep only:
```markdown
# Code Directory Preferences

- Project-specific CLAUDE.md files always take precedence
- Use language-appropriate conventions for each project
- Prefer standard library solutions when possible
```

**Do NOT edit** `~/CLAUDE.md` (contains independent cross-project guidance).

### Step 4: Verify Token Savings
```bash
# Count characters (tokens ≈ chars / 4)
wc -c ~/.claude/CLAUDE.md ~/Code/.claude/CLAUDE.md ~/Code/roster/.claude/CLAUDE.md

# Before target (if parent files had duplication):
# ~/.claude/CLAUDE.md: ~6,800 chars (~1,700 tokens)
# ~/Code/.claude/CLAUDE.md: ~6,800 chars (~1,700 tokens)
# Project CLAUDE.md: ~12,400 chars (~3,100 tokens)
# Total: ~31,800 chars (~7,950 tokens/turn)

# After target (parent files trimmed):
# ~/.claude/CLAUDE.md: ~240 chars (~60 tokens)
# ~/Code/.claude/CLAUDE.md: ~200 chars (~50 tokens)
# Project CLAUDE.md: ~2,400 chars (~600 tokens after C1 implementation)
# ~/CLAUDE.md: ~3,800 chars (~950 tokens, unchanged)
# Total: ~6,640 chars (~1,660 tokens/turn)
```

**Verify**: Total character count significantly reduced (50-70% reduction depending on starting state).

### Step 5: Test Agent Behavior
Start a new Claude Code session in the roster project:

```bash
cd ~/Code/roster
# Start Claude Code
```

Test that agents can still:
1. Determine execution mode (read project CLAUDE.md → execution-mode section present)
2. Identify available agents (read project CLAUDE.md → quick-start section present)
3. Find navigation pointers (read project CLAUDE.md → navigation section present)
4. Access CLI reference (`ari --help` works)
5. Delegate to specialists (Task tool invocation succeeds)

**Verify**: No behavioral regressions. Agents function identically to before migration.

### Step 6: Monitor for 2-3 Sessions
Continue working normally for 2-3 Claude Code sessions (different agents, different tasks).

**Watch for**:
- Agents asking questions they didn't ask before (sign of missing L0 content)
- Agents failing to find information (sign of broken navigation pointers)
- Unexpected tool invocations (sign of inefficient information retrieval)

**Verify**: No degradation in agent effectiveness. Token savings realized without behavioral cost.

### Rollback Instructions
If agent behavior degrades:

```bash
# Restore parent files
cp ~/.claude/CLAUDE.md.backup.YYYYMMDD ~/.claude/CLAUDE.md
cp ~/Code/.claude/CLAUDE.md.backup.YYYYMMDD ~/Code/.claude/CLAUDE.md

# Verify restoration
diff ~/.claude/CLAUDE.md.backup.YYYYMMDD ~/.claude/CLAUDE.md
diff ~/Code/.claude/CLAUDE.md.backup.YYYYMMDD ~/Code/.claude/CLAUDE.md
```

**Verify**: `diff` shows no differences. Original content restored.

## Rules for Maintaining the Hierarchy

### Rule 1: Project CLAUDE.md Never Repeats Parent Content
If parent files say "prefer editing over creating files", the project file must not repeat this.

### Rule 2: Parent Files Contain Only Cross-Project Guidance
If guidance applies to all projects (Knossos and non-Knossos), it belongs in `~/CLAUDE.md`. If it applies only to Knossos projects, it belongs in the project file, not in `~/.claude/CLAUDE.md`.

### Rule 3: Knossos Platform Features Belong in Project File
Execution modes, agent rosters, rite management, session tracking, hooks — all Knossos-specific. These belong **only** in `~/Code/roster/.claude/CLAUDE.md`, never in parent files.

### Rule 4: On-Demand Content Belongs in Skills or Docs (L2/L3)
CLI reference, invocation patterns, architecture descriptions, file location tables — all available via cheaper mechanisms. Keep L0 (CLAUDE.md) minimal.

## Token Savings Estimate

| Scenario | Before (tokens/turn) | After (tokens/turn) | Savings |
|----------|---------------------|---------------------|---------|
| **Parent files already trimmed** (current state) | ~4,060 | ~2,590 (after C1) | 36% |
| **Parent files duplicated** (hypothetical) | ~7,950 | ~2,590 (after trim + C1) | 67% |

The actual savings depend on whether parent files contain duplication. Current assessment shows parent files are already minimal, so the primary savings come from project-level CLAUDE.md content redesign (C1 implementation), not from parent file deduplication.

## References

- **C1 Content Redesign**: `/Users/tomtenuta/Code/roster/docs/design/C1-content-redesign.md` (Part 4 covers parent file specifications)
- **Context Tier Model**: `/Users/tomtenuta/Code/roster/docs/design/TDD-context-tier-model.md` (token economics, L0/L1/L2/L3 tier definitions)
- **Current parent files**: Already assessed as optimal (see "Current Parent File Assessment" section above)

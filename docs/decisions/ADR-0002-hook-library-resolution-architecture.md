# ADR-0002: Hook Library Resolution Architecture

| Field | Value |
|-------|-------|
| **Status** | Accepted (Implemented) |
| **Date** | 2025-12-31 |
| **Deciders** | Architecture Team |
| **Supersedes** | N/A |
| **Superseded by** | N/A |

## Context

The roster project has developed a sophisticated hook system that extends Claude Code's native hook capabilities. This ADR documents the architectural decisions for hook library resolution, template distribution, and the harness/satellite project model.

### The Core Problem (Original)

Our hooks in `.claude/hooks/` attempted to source shared libraries using various fragile path resolution strategies that broke across different installation contexts:

**Pattern 1: SCRIPT_DIR relative paths**
```bash
SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"
source "$SCRIPT_DIR/lib/logging.sh"
```
Problem: Only works when hooks are in project `.claude/hooks/` with `lib/` subdirectory.

**Pattern 2: Hardcoded project paths**
```bash
source .claude/hooks/lib/session-utils.sh
```
Problem: Assumes project-level hooks only, breaks for user-level hooks.

**Pattern 3: Computed HOOKS_LIB with fallback**
```bash
HOOKS_LIB="${CLAUDE_PROJECT_DIR:-$(cd "$SCRIPT_DIR/../.." && pwd)}/user-hooks/lib"
source "$HOOKS_LIB/logging.sh"
```
Problem: References a custom `user-hooks/` directory that is not a Claude Code convention.

### Key Insight: Harness vs Satellite Architecture

During implementation, we discovered a critical architectural principle:

| Project Type | Description | Examples |
|--------------|-------------|----------|
| **Harness** | Projects that manage and distribute templates | roster, skeleton_claude |
| **Satellite** | Projects that receive synced content | Any project using 10x workflow |

**Critical Realization**: The `.claude/` directory in ANY project (including harness projects like roster) should be a DESTINATION populated by sync scripts, NOT a source of templates.

This means:
- `roster/hooks/` is the **canonical template source**
- `roster/.claude/hooks/` is a **destination** (populated by `install-hooks.sh`)
- Satellite projects' `.claude/hooks/` are **destinations** (populated by same mechanism)

### Claude Code's Hook Architecture

Claude Code provides a two-layer hook system:

**Layer 1: JSON Configuration** (registers hooks in `settings.json`)
```json
{
  "hooks": {
    "SessionStart": [{"matcher": "", "hooks": ["bash .claude/hooks/session-context.sh"]}]
  }
}
```

**Layer 2: Bash Execution** (scripts receive JSON via stdin)
- Hooks receive structured JSON with `session_id`, `cwd`, `permission_mode`, etc.
- Environment variable `CLAUDE_PROJECT_DIR` provided at runtime
- Working directory is NOT guaranteed to be project root

**Environment Variables**:
| Variable | Description | Availability |
|----------|-------------|--------------|
| `CLAUDE_PROJECT_DIR` | Absolute path to project root | All hooks |
| `CLAUDE_CODE_REMOTE` | "true" if running in web environment | All hooks |
| `CLAUDE_ENV_FILE` | File path for persisting env vars | SessionStart only |
| `CLAUDE_PLUGIN_ROOT` | Absolute path to plugin directory | Plugin hooks only |

### What Claude Code Does NOT Provide

1. **No `CLAUDE_HOOKS_DIR` variable**: There is no environment variable pointing to where hooks are stored.
2. **No library sharing mechanism**: Claude Code does not provide a convention for sharing code between hooks.
3. **No hook inheritance**: User-level hooks and project-level hooks are separate; there's no composition model.
4. **No plugin hooks for CLI**: The `CLAUDE_PLUGIN_ROOT` variable is only available for plugins, not CLI/project hooks.

### Runtime Considerations

Claude Code hooks operate under specific runtime constraints:

1. **Parallel Execution**: Hooks run in parallel by default. Libraries must be thread-safe:
   - Avoid global mutable state
   - Use atomic file operations for shared resources
   - Prefer function-local variables

2. **60-Second Timeout**: All hooks have a 60-second execution timeout:
   - Implement early bailout for non-critical paths
   - Use timeout wrappers for external commands
   - Prefer async patterns over synchronous blocking

3. **Hook Input Structure**: Hooks receive JSON via stdin:
   - Read stdin once and cache the result
   - Use `jq` for reliable JSON parsing
   - Handle missing or malformed input gracefully

### Ecosystem Integration: CEM Exclusion

The Claude Ecosystem Manager (CEM) ignores ALL roster-managed artifacts:

```bash
# CEM ignores: agents, commands, skills, hooks (all roster-managed)
# CEM manages: settings.json merging only
```

This is intentional separation of concerns:
- **roster**: Manages ALL Claude ecosystem artifacts (agents, commands, skills, hooks)
- **CEM**: Provides settings.json composition/merging only

### Forces

- **Simplicity**: Hooks should be self-contained and easy to understand
- **Portability**: Hooks should work across project-level and user-level installations
- **Maintainability**: Shared logic should live in one place, not be duplicated
- **Claude Code Alignment**: Work with Claude Code's model, not against it
- **Harness/Satellite Separation**: Templates live in harness, destinations receive copies
- **Thread Safety**: Libraries must handle parallel hook execution

## Decision

We adopted a **template distribution architecture** with three key components:

### 1. Canonical Template Source: `roster/hooks/`

All hook templates and libraries live in a top-level `hooks/` directory in the roster harness:

```
roster/
  hooks/                              # CANONICAL TEMPLATE SOURCE
    lib/                              # Shared libraries
      config.sh                       # Configuration constants
      logging.sh                      # Logging utilities
      primitives.sh                   # Portable bash functions
      session-core.sh                 # Session identification
      session-state.sh                # Session state queries
      session-utils.sh                # Compatibility shim
      session-fsm.sh                  # State machine
      session-manager.sh              # CLI interface
      session-migrate.sh              # Migration utilities
      worktree-manager.sh             # Git worktree management
    session-context.sh                # SessionStart hook
    auto-park.sh                      # Stop hook
    artifact-tracker.sh               # PostToolUse hook
    command-validator.sh              # PreToolUse hook
    commit-tracker.sh                 # PostToolUse hook
    delegation-check.sh               # PreToolUse hook
    coach-mode.sh                     # PostToolUse hook
    session-audit.sh                  # PostToolUse hook
    session-write-guard.sh            # PreToolUse hook
    start-preflight.sh                # PreToolUse hook
    team-validator.sh                 # PreToolUse hook
    workflow-validator.sh             # PreToolUse hook
  install-hooks.sh                    # Project installation script
  sync-user-hooks.sh                  # User-level sync script
```

### 2. Template Distribution Flow

```
roster/hooks/              <- CANONICAL TEMPLATE SOURCE
roster/hooks/lib/          <- CANONICAL LIBRARY SOURCE
    |
    | (install-hooks.sh)
    v
project/.claude/hooks/     <- PROJECT DESTINATION
    |
    | (sync-user-hooks.sh)
    v
~/.claude/hooks/           <- USER-LEVEL DESTINATION
```

**install-hooks.sh**: Copies templates to any project's `.claude/hooks/`
- Works for harness projects (roster itself) and satellite projects
- Validates target is a Claude project (has `.claude/` directory)
- Creates `lib/` subdirectory structure
- Sets executable permissions

**sync-user-hooks.sh**: Copies templates to user-level `~/.claude/hooks/`
- Additive: Never removes existing hooks
- Manifest-tracked: Only overwrites roster-managed hooks
- Preserves user-created hooks not from roster
- Supports `--adopt` mode to recover manifest from existing files
- Supports `--dry-run` for previewing changes

### 3. Runtime Library Resolution Pattern

All hooks use this canonical pattern at runtime:

```bash
#!/bin/bash
# Hook description

set -euo pipefail

# =============================================================================
# Input Capture (must happen before any other reads from stdin)
# =============================================================================
HOOK_INPUT="$(cat)"

# =============================================================================
# Library Resolution
# =============================================================================
# Claude Code provides CLAUDE_PROJECT_DIR as absolute path to project root.
# All hooks MUST use this for library resolution.

HOOKS_LIB="${CLAUDE_PROJECT_DIR:-.}/.claude/hooks/lib"

# Source required libraries (fail gracefully if not found)
source "$HOOKS_LIB/logging.sh" 2>/dev/null && log_init "hook-name" || true
source "$HOOKS_LIB/session-utils.sh" 2>/dev/null || {
    # Fallback behavior when libraries unavailable
    exit 0
}

# =============================================================================
# Hook Logic
# =============================================================================
SESSION_ID="$(echo "$HOOK_INPUT" | jq -r '.session_id // empty')"
WORKING_DIR="$(echo "$HOOK_INPUT" | jq -r '.cwd // empty')"

# ... rest of hook
```

**Key principles**:
1. Always use `$CLAUDE_PROJECT_DIR` - never `$SCRIPT_DIR` or relative paths
2. Capture stdin immediately - it can only be read once
3. Source libraries with `2>/dev/null` to suppress errors
4. Implement graceful fallback when libraries are unavailable
5. Libraries must not have side effects on source (no `set -euo pipefail` in libraries)
6. Libraries must be thread-safe for parallel hook execution
7. All operations must complete within 60 seconds

### 4. User-Level Hook Options

For users who want roster hooks at the user level (`~/.claude/hooks/`):

**Option A: Use sync-user-hooks.sh** (recommended)
```bash
./sync-user-hooks.sh              # Sync from roster/hooks/
./sync-user-hooks.sh --dry-run    # Preview changes
./sync-user-hooks.sh --status     # Show sync status
./sync-user-hooks.sh --adopt      # Recover manifest from existing
```

**Option B: Symlink to project** (development only)
```bash
ln -s /path/to/roster/.claude/hooks ~/.claude/hooks
```

**Option C: ROSTER_HOME resolution** (production)
```bash
# In user-level hooks, use ROSTER_HOME for library resolution
HOOKS_LIB="${ROSTER_HOME:-$HOME/Code/roster}/.claude/hooks/lib"
```

### 5. Future Extension: Team-Specific Hooks

Teams can optionally have their own `hooks/` directories (parallel to existing `commands/`):

```
rites/
  10x-dev-pack/
    commands/
    hooks/          # Team-specific hooks (future)
  ecosystem-pack/
    commands/
    hooks/          # Team-specific hooks (future)
```

When active, team hooks would override or supplement base hooks. The `sync-user-hooks.sh` script already detects team hook collisions and warns appropriately.

## Alternatives Considered

### Option A: Keep user-hooks/ as Canonical Location
- **Pros**: No migration needed, existing scripts work
- **Cons**: Fights Claude Code conventions, confusing for new contributors, requires sync between directories

### Option B: Use Claude Code Plugins
- **Pros**: Official extension mechanism, `CLAUDE_PLUGIN_ROOT` available
- **Cons**: Plugins are for adding functionality, not for project-specific hooks; overkill for this use case

### Option C: Inline All Libraries
- **Pros**: Each hook is self-contained, no resolution complexity
- **Cons**: Massive code duplication, maintenance nightmare, inconsistent behavior across hooks

### Option D: Environment Variable for Hook Library
- **Pros**: Flexible, works anywhere
- **Cons**: Requires external setup, not discoverable, another thing to configure

### Option E: Keep Templates in .claude/hooks/
- **Pros**: Follows Claude Code directory convention
- **Cons**: Violates harness/satellite principle - `.claude/` should be a destination, not a source

## Rationale

We chose the template distribution architecture because:

1. **Harness/Satellite Clarity**: Templates live in harness (`roster/hooks/`), destinations receive copies (`.claude/hooks/`)
2. **Single Source of Truth**: All libraries in one canonical location eliminates sync issues
3. **Portable Runtime Pattern**: `$CLAUDE_PROJECT_DIR/.claude/hooks/lib` works in all contexts at runtime
4. **Graceful Degradation**: Hooks still work (with reduced functionality) if libraries unavailable
5. **CEM Compatibility**: Explicit separation from skeleton_claude/CEM domain
6. **Manifest Tracking**: User-level hooks tracked for safe updates

## Consequences

### Positive

1. **Simplified mental model**: Templates in `roster/hooks/`, runtime resolution via `$CLAUDE_PROJECT_DIR`
2. **Reliable resolution**: `$CLAUDE_PROJECT_DIR` is always available and absolute
3. **Claude Code alignment**: Following platform conventions reduces friction
4. **Easier onboarding**: New contributors find templates in expected location
5. **Testability**: Libraries can be tested independently at known paths
6. **Safe updates**: Manifest tracking prevents clobbering user modifications

### Negative

1. **Two-step installation**: Must run `install-hooks.sh` after cloning roster
2. **User-level hooks need sync**: Must run `sync-user-hooks.sh` for user-level installation
3. **Potential drift**: Project `.claude/hooks/` can drift from `roster/hooks/` if not re-synced

### Neutral

1. **No functional changes**: Hook behavior remains identical
2. **Same capabilities**: All existing features preserved
3. **Legacy removal**: `user-hooks/` directory eliminated (was duplicate)

## Implementation Checklist

- [x] Create `roster/hooks/` directory (canonical template source)
- [x] Create `roster/hooks/lib/` directory (canonical library source)
- [x] Move libraries: `config.sh`, `logging.sh`, `primitives.sh`, `session-*.sh`, `worktree-manager.sh`
- [x] Move hooks: 12 hook scripts to `roster/hooks/`
- [x] Create `install-hooks.sh` for project installation
- [x] Update `sync-user-hooks.sh` to use new source location
- [x] Update all hooks to use `$CLAUDE_PROJECT_DIR/.claude/hooks/lib` pattern
- [x] Remove `user-hooks/` directory
- [x] Test all hook lifecycle events (SessionStart, Stop, PreToolUse, PostToolUse)
- [x] Install hooks to `roster/.claude/hooks/` (harness is also a destination)
- [ ] Update CLAUDE.md documentation
- [ ] Update skill files that reference hook paths
- [ ] Create user-level hook installation documentation
- [ ] Tag release with migration notes

## File Verification

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This ADR | `/Users/tomtenuta/Code/roster/docs/decisions/ADR-0002-hook-library-resolution-architecture.md` | Updated |
| Canonical hook templates | `/Users/tomtenuta/Code/roster/hooks/` | Created (12 hooks) |
| Canonical library templates | `/Users/tomtenuta/Code/roster/hooks/lib/` | Created (10 libraries) |
| Project installation script | `/Users/tomtenuta/Code/roster/install-hooks.sh` | Created |
| User-level sync script | `/Users/tomtenuta/Code/roster/sync-user-hooks.sh` | Updated |
| Project hook destination | `/Users/tomtenuta/Code/roster/.claude/hooks/` | Populated |
| Project library destination | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/` | Populated |
| Legacy user-hooks directory | `/Users/tomtenuta/Code/roster/user-hooks/` | Deleted |

## References

- Claude Code Hooks Documentation: https://code.claude.com/docs/en/hooks
- ADR-0001: Session State Machine Redesign
- Commit 3f30edb: "fix(hooks): use generic hook paths for Claude Code resolution"
- Canonical hook templates: `roster/hooks/*.sh`
- Canonical library templates: `roster/hooks/lib/*.sh`

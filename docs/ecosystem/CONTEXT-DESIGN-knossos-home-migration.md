# Context Design: ROSTER_HOME to KNOSSOS_HOME Environment Variable Migration

**Date**: 2026-01-06
**Architect**: Context Architect
**Reference**: Knossos Migration Sprint - Environment Variable Rebrand
**Prerequisite**: Part of "roster" to "knossos" platform rebranding initiative

---

## Executive Summary

The platform is being rebranded from "roster" to "knossos" with the dual-meaning "KnowSOS" (Greek mythology + "Know + SOS = context awareness + save our sessions"). This migration updates the primary environment variable from `ROSTER_HOME` to `KNOSSOS_HOME` with full backward compatibility.

**Scope**:
- 3 Go source files with direct `ROSTER_HOME` references
- 34+ shell scripts with `ROSTER_HOME` usage patterns
- 120+ documentation files referencing `ROSTER_HOME`
- User shell profiles (`.bashrc`, `.zshrc`) with existing `ROSTER_HOME` exports

**Approach**: Cascading fallback with deprecation warnings. New variable takes precedence, old variable continues to work with stderr warning.

---

## Inventory of ROSTER_HOME References

### Category 1: Go Code (3 files, 3 references)

| File | Line | Usage Pattern | Risk |
|------|------|---------------|------|
| `/Users/tomtenuta/Code/roster/ariadne/internal/worktree/operations.go` | 90 | `os.Getenv("ROSTER_HOME")` | LOW |
| `/Users/tomtenuta/Code/roster/ariadne/internal/worktree/operations.go` | 662 | `os.Getenv("ROSTER_HOME")` | LOW |
| `/Users/tomtenuta/Code/roster/ariadne/internal/worktree/lifecycle.go` | 136 | `os.Getenv("ROSTER_HOME")` | LOW |

**Pattern Used**: Direct `os.Getenv()` with empty string check for absence.

### Category 2: Shell Scripts - Primary Definition (14 files)

These files define `ROSTER_HOME` with the canonical fallback pattern:

| File | Line | Pattern | References |
|------|------|---------|------------|
| `/Users/tomtenuta/Code/roster/swap-rite.sh` | 11 | `readonly ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"` | 34 total |
| `/Users/tomtenuta/Code/roster/roster-sync` | 37 | `readonly ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"` | 45 total |
| `/Users/tomtenuta/Code/roster/sync-user-hooks.sh` | 25 | `readonly ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"` | 6 total |
| `/Users/tomtenuta/Code/roster/sync-user-skills.sh` | 26 | `readonly ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"` | 4 total |
| `/Users/tomtenuta/Code/roster/sync-user-agents.sh` | 24 | `readonly ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"` | 6 total |
| `/Users/tomtenuta/Code/roster/sync-user-commands.sh` | 25 | `readonly ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"` | 6 total |
| `/Users/tomtenuta/Code/roster/install-hooks.sh` | 22 | `readonly ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"` | 2 total |
| `/Users/tomtenuta/Code/roster/user-hooks/lib/config.sh` | 16 | `export ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"` | 1 total |
| `/Users/tomtenuta/Code/roster/templates/orchestrator-generate.sh` | 26 | `readonly ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"` | 10 total |
| `/Users/tomtenuta/Code/roster/templates/validate-orchestrator.sh` | 19 | `readonly ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"` | 2 total |
| `/Users/tomtenuta/Code/roster/templates/generate-orchestrator.sh` | 13 | `ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"` | 5 total |
| `/Users/tomtenuta/Code/roster/generate-rite-context.sh` | 9 | `ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"` | 2 total |
| `/Users/tomtenuta/Code/roster/load-workflow.sh` | 8 | `ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"` | 4 total |
| `/Users/tomtenuta/Code/roster/get-workflow-field.sh` | 9 | `ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"` | 2 total |

### Category 3: Shell Scripts - Library Consumers (20+ files)

Files that use `$ROSTER_HOME` after sourcing or assuming it's set:

| File | Usage Count | Purpose |
|------|-------------|---------|
| `/Users/tomtenuta/Code/roster/lib/rite/rite-resource.sh` | 9 | Rite pack resolution |
| `/Users/tomtenuta/Code/roster/lib/rite/rite-transaction.sh` | 5 | Transaction staging |
| `/Users/tomtenuta/Code/roster/lib/rite/rite-hooks-registration.sh` | 5 | Hook registration |
| `/Users/tomtenuta/Code/roster/lib/sync/sync-core.sh` | 5 | Sync operations |
| `/Users/tomtenuta/Code/roster/lib/roster-utils.sh` | 2 | Utility functions |
| `/Users/tomtenuta/Code/roster/user-hooks/lib/worktree-manager.sh` | 8 | Worktree ecosystem |
| `/Users/tomtenuta/Code/roster/user-hooks/lib/rite-context-loader.sh` | 3 | Rite context loading |
| `/Users/tomtenuta/Code/roster/user-hooks/lib/session-manager.sh` | 1 | Session management |
| `/Users/tomtenuta/Code/roster/user-hooks/validation/command-validator.sh` | 3 | Command validation |
| `/Users/tomtenuta/Code/roster/bin/fix-hardcoded-paths.sh` | 12 | Path normalization |
| `/Users/tomtenuta/Code/roster/bin/normalize-team-structure.sh` | 2 | Structure normalization |

### Category 4: Test Files (15+ files)

| File | Usage Count | Notes |
|------|-------------|-------|
| `/Users/tomtenuta/Code/roster/tests/sync/test-init.sh` | 28 | Integration tests |
| `/Users/tomtenuta/Code/roster/tests/sync/test-validate-repair.sh` | 19 | Validation tests |
| `/Users/tomtenuta/Code/roster/tests/sync/test-swap-rite-integration.sh` | 17 | Swap tests |
| `/Users/tomtenuta/Code/roster/tests/sync/test-sync-*.sh` | 3-5 each | Sync unit tests |
| `/Users/tomtenuta/Code/roster/tests/lib/rite/test-rite-*.sh` | 3-4 each | Rite lib tests |
| `/Users/tomtenuta/Code/roster/tests/test-rite-context-loader.sh` | 7 | Context loader tests |

### Category 5: Documentation Files (120+ files)

High-impact documentation:

| File | References | User-Facing |
|------|------------|-------------|
| `/Users/tomtenuta/Code/roster/docs/INTEGRATION.md` | 9 | YES - primary integration guide |
| `/Users/tomtenuta/Code/roster/docs/migration/cem-to-roster-migration.md` | 9 | YES - migration guide |
| `/Users/tomtenuta/Code/roster/user-skills/guidance/rite-ref/SKILL.md` | 12 | YES - rite skill |
| `/Users/tomtenuta/Code/roster/rites/forge-pack/agents/platform-engineer.md` | 16 | YES - agent prompt |
| `/Users/tomtenuta/Code/roster/user-commands/cem/sync.md` | 3 | YES - command doc |
| Various TDD/design docs | 5-15 each | Internal |

### Category 6: User Commands and Skills

Commands with `ROSTER_HOME` in execution paths:

| File | Pattern |
|------|---------|
| `/Users/tomtenuta/Code/roster/user-commands/rite-switching/*.md` | `${ROSTER_HOME:-~/Code/roster}/swap-rite.sh` |
| `/Users/tomtenuta/Code/roster/user-commands/navigation/rite.md` | `${ROSTER_HOME:-~/Code/roster}/swap-rite.sh` |
| `/Users/tomtenuta/Code/roster/user-commands/cem/sync.md` | `${ROSTER_HOME:-~/Code/roster}/roster-sync` |
| `/Users/tomtenuta/Code/roster/skills/team/skill.md` | `$ROSTER_HOME/swap-rite.sh` |

---

## Backward Compatibility Design

### Classification: COMPATIBLE (with deprecation warnings)

### Resolution Function Pattern

**Shell (Bash):**

```bash
# resolve_knossos_home - Resolve KNOSSOS_HOME with ROSTER_HOME fallback
# Sets: KNOSSOS_HOME (exported)
# Warns: If using deprecated ROSTER_HOME
resolve_knossos_home() {
    if [[ -n "${KNOSSOS_HOME:-}" ]]; then
        # Primary: KNOSSOS_HOME is set
        export KNOSSOS_HOME
    elif [[ -n "${ROSTER_HOME:-}" ]]; then
        # Fallback: ROSTER_HOME is set (deprecated)
        export KNOSSOS_HOME="$ROSTER_HOME"
        echo "[DEPRECATED] ROSTER_HOME is deprecated. Please update your shell profile:" >&2
        echo "  Replace: export ROSTER_HOME=\"$ROSTER_HOME\"" >&2
        echo "  With:    export KNOSSOS_HOME=\"$ROSTER_HOME\"" >&2
    else
        # Default: Neither set
        export KNOSSOS_HOME="$HOME/Code/roster"
    fi
}
```

**Compact Pattern for Script Headers:**

```bash
# Knossos Home Resolution (ROSTER_HOME fallback with deprecation warning)
if [[ -n "${KNOSSOS_HOME:-}" ]]; then
    readonly KNOSSOS_HOME
elif [[ -n "${ROSTER_HOME:-}" ]]; then
    readonly KNOSSOS_HOME="$ROSTER_HOME"
    echo "[DEPRECATED] ROSTER_HOME is deprecated. Set KNOSSOS_HOME instead." >&2
else
    readonly KNOSSOS_HOME="$HOME/Code/roster"
fi
```

**Go:**

```go
// resolveKnossosHome returns the Knossos platform home directory.
// Falls back to ROSTER_HOME with deprecation warning.
func resolveKnossosHome() string {
    if home := os.Getenv("KNOSSOS_HOME"); home != "" {
        return home
    }
    if home := os.Getenv("ROSTER_HOME"); home != "" {
        fmt.Fprintln(os.Stderr, "[DEPRECATED] ROSTER_HOME is deprecated. Set KNOSSOS_HOME instead.")
        return home
    }
    return filepath.Join(os.Getenv("HOME"), "Code", "roster")
}
```

### Deprecation Timeline

| Version | Status | Behavior |
|---------|--------|----------|
| v1.x (current) | Preparation | `ROSTER_HOME` is only variable |
| v2.0 | Transition | `KNOSSOS_HOME` primary, `ROSTER_HOME` fallback with warning |
| v2.5 | Deprecation | Louder warnings, documentation says "ROSTER_HOME removed in v3.0" |
| v3.0 | Removal | `ROSTER_HOME` no longer recognized |

---

## Implementation Specification

### Phase 1: Core Resolution Function (lib/knossos-home.sh)

**New File**: `/Users/tomtenuta/Code/roster/lib/knossos-home.sh`

```bash
#!/usr/bin/env bash
# knossos-home.sh - Centralized KNOSSOS_HOME resolution with deprecation handling
#
# Usage:
#   source "$SCRIPT_DIR/../lib/knossos-home.sh"
#   resolve_knossos_home
#   # KNOSSOS_HOME is now set and exported
#
# Environment Variables:
#   KNOSSOS_HOME - Primary platform home (preferred)
#   ROSTER_HOME  - Deprecated fallback (with warning)
#   Default: $HOME/Code/roster

# Version for migration tracking
readonly KNOSSOS_HOME_RESOLVER_VERSION="1.0.0"

# Deprecation warning control
# Set KNOSSOS_SUPPRESS_DEPRECATION=1 to silence warnings (for tests)
KNOSSOS_SUPPRESS_DEPRECATION="${KNOSSOS_SUPPRESS_DEPRECATION:-0}"

# resolve_knossos_home - Resolve and export KNOSSOS_HOME
# Idempotent: safe to call multiple times
resolve_knossos_home() {
    # Already resolved
    if [[ -n "${_KNOSSOS_HOME_RESOLVED:-}" ]]; then
        return 0
    fi

    if [[ -n "${KNOSSOS_HOME:-}" ]]; then
        # Primary: KNOSSOS_HOME is set
        export KNOSSOS_HOME
    elif [[ -n "${ROSTER_HOME:-}" ]]; then
        # Fallback: ROSTER_HOME is set (deprecated)
        export KNOSSOS_HOME="$ROSTER_HOME"
        if [[ "$KNOSSOS_SUPPRESS_DEPRECATION" != "1" ]]; then
            {
                echo "[DEPRECATED] Environment variable ROSTER_HOME is deprecated."
                echo "  Update your shell profile (~/.bashrc, ~/.zshrc, etc.):"
                echo "  - Remove: export ROSTER_HOME=\"$ROSTER_HOME\""
                echo "  + Add:    export KNOSSOS_HOME=\"$ROSTER_HOME\""
                echo "  ROSTER_HOME support will be removed in version 3.0"
            } >&2
        fi
    else
        # Default: Neither set
        export KNOSSOS_HOME="$HOME/Code/roster"
    fi

    # Mark as resolved
    export _KNOSSOS_HOME_RESOLVED=1
}

# Auto-resolve on source (can be disabled with KNOSSOS_HOME_NO_AUTO_RESOLVE=1)
if [[ "${KNOSSOS_HOME_NO_AUTO_RESOLVE:-0}" != "1" ]]; then
    resolve_knossos_home
fi
```

### Phase 2: Go Resolution Utility (ariadne/internal/config/home.go)

**New File**: `/Users/tomtenuta/Code/roster/ariadne/internal/config/home.go`

```go
package config

import (
    "fmt"
    "os"
    "path/filepath"
    "sync"
)

var (
    knossosHome     string
    knossosHomeOnce sync.Once
    deprecationShown bool
)

// KnossosHome returns the resolved Knossos platform home directory.
// Falls back to ROSTER_HOME with deprecation warning (shown once per process).
// Default: $HOME/Code/roster
func KnossosHome() string {
    knossosHomeOnce.Do(func() {
        knossosHome = resolveKnossosHome()
    })
    return knossosHome
}

func resolveKnossosHome() string {
    // Primary: KNOSSOS_HOME
    if home := os.Getenv("KNOSSOS_HOME"); home != "" {
        return home
    }

    // Fallback: ROSTER_HOME (deprecated)
    if home := os.Getenv("ROSTER_HOME"); home != "" {
        if !deprecationShown && os.Getenv("KNOSSOS_SUPPRESS_DEPRECATION") != "1" {
            fmt.Fprintln(os.Stderr, "[DEPRECATED] ROSTER_HOME is deprecated. Set KNOSSOS_HOME instead.")
            fmt.Fprintln(os.Stderr, "  ROSTER_HOME support will be removed in version 3.0")
            deprecationShown = true
        }
        return home
    }

    // Default
    homeDir, _ := os.UserHomeDir()
    return filepath.Join(homeDir, "Code", "roster")
}
```

### Phase 3: Update Primary Scripts

**Priority Order** (by usage frequency):

| Order | File | Changes |
|-------|------|---------|
| 3.1 | `swap-rite.sh` | Source `lib/knossos-home.sh`, replace `ROSTER_HOME` with `KNOSSOS_HOME` |
| 3.2 | `roster-sync` | Source `lib/knossos-home.sh`, replace `ROSTER_HOME` with `KNOSSOS_HOME` |
| 3.3 | `user-hooks/lib/config.sh` | Use resolution function |
| 3.4 | `sync-user-*.sh` (4 files) | Source resolver |
| 3.5 | `install-hooks.sh` | Source resolver |
| 3.6 | `templates/*.sh` (3 files) | Source resolver |
| 3.7 | `lib/sync/*.sh` (6 files) | Use `$KNOSSOS_HOME` (sourced by parent) |
| 3.8 | `lib/rite/*.sh` (3 files) | Use `$KNOSSOS_HOME` (sourced by parent) |

**Example Migration for swap-rite.sh:**

```bash
# BEFORE (line 11):
readonly ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"

# AFTER:
# Source Knossos home resolution (handles ROSTER_HOME deprecation)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/knossos-home.sh"
```

### Phase 4: Update Go Code

**Files to update:**

| File | Function | Change |
|------|----------|--------|
| `ariadne/internal/worktree/operations.go` | Line 90 | Replace `os.Getenv("ROSTER_HOME")` with `config.KnossosHome()` |
| `ariadne/internal/worktree/operations.go` | Line 662 | Replace `os.Getenv("ROSTER_HOME")` with `config.KnossosHome()` |
| `ariadne/internal/worktree/lifecycle.go` | Line 136 | Replace `os.Getenv("ROSTER_HOME")` with `config.KnossosHome()` |

### Phase 5: Update Library Files

Files that use `$ROSTER_HOME` without defining it:

| File | Change |
|------|--------|
| `lib/roster-utils.sh` | Replace `$ROSTER_HOME` with `$KNOSSOS_HOME` |
| `lib/rite/rite-resource.sh` | Replace `$ROSTER_HOME` with `$KNOSSOS_HOME` |
| `lib/rite/rite-transaction.sh` | Replace `$ROSTER_HOME` with `$KNOSSOS_HOME` |
| `lib/rite/rite-hooks-registration.sh` | Replace `$ROSTER_HOME` with `$KNOSSOS_HOME` |
| `lib/sync/sync-core.sh` | Replace `$ROSTER_HOME` with `$KNOSSOS_HOME` |
| `lib/sync/sync-manifest.sh` | Replace `$ROSTER_HOME` with `$KNOSSOS_HOME` |
| `user-hooks/lib/worktree-manager.sh` | Replace `$ROSTER_HOME` with `$KNOSSOS_HOME` |
| `user-hooks/lib/rite-context-loader.sh` | Replace `$ROSTER_HOME` with `$KNOSSOS_HOME` |
| `user-hooks/validation/command-validator.sh` | Replace `$ROSTER_HOME` with `$KNOSSOS_HOME` |

### Phase 6: Update Test Files

Test files need the same resolution, plus option to suppress warnings:

```bash
# At top of test files:
export KNOSSOS_SUPPRESS_DEPRECATION=1  # Silence deprecation warnings in tests
```

### Phase 7: Update Documentation

**High-Priority Documentation:**

| File | Action |
|------|--------|
| `docs/INTEGRATION.md` | Replace `ROSTER_HOME` with `KNOSSOS_HOME`, add migration note |
| `docs/migration/cem-to-roster-migration.md` | Update variable references |
| `docs/announcements/skeleton-deprecation-announcement.md` | Add KNOSSOS_HOME migration section |
| `user-commands/cem/sync.md` | Update examples |
| `user-skills/guidance/rite-ref/SKILL.md` | Update references |

**Pattern for Documentation Updates:**

```markdown
# BEFORE:
export ROSTER_HOME="$HOME/Code/roster"

# AFTER:
export KNOSSOS_HOME="$HOME/Code/roster"
# Note: ROSTER_HOME is deprecated but still supported for backward compatibility
```

### Phase 8: User Commands

Update the fallback pattern in user-facing commands:

```markdown
# BEFORE:
${ROSTER_HOME:-~/Code/roster}/swap-rite.sh

# AFTER:
${KNOSSOS_HOME:-${ROSTER_HOME:-~/Code/roster}}/swap-rite.sh
```

---

## Risk Assessment

### Risk Matrix

| Component | Impact | Likelihood | Mitigation |
|-----------|--------|------------|------------|
| Shell scripts fail | HIGH | LOW | Backward compat fallback |
| Go binaries fail | MEDIUM | LOW | Centralized resolution |
| User muscle memory | LOW | HIGH | Deprecation warnings |
| CI/CD scripts break | MEDIUM | MEDIUM | Document migration |
| Existing satellites | LOW | LOW | No change required |

### Breaking Change Assessment

**This migration is NOT a breaking change** because:

1. **Fallback**: `ROSTER_HOME` continues to work
2. **Default unchanged**: `$HOME/Code/roster` remains default
3. **No schema changes**: No config files affected
4. **No path changes**: Directory structure unchanged

### Satellite Impact

| Satellite Type | Impact | Action Required |
|----------------|--------|-----------------|
| Existing (ROSTER_HOME set) | None | Will see deprecation warning |
| New installations | None | Use KNOSSOS_HOME |
| CI/CD with ROSTER_HOME | Low | Update when convenient |

---

## Integration Test Matrix

| Satellite | Test Case | Expected Outcome | Validates |
|-----------|-----------|------------------|-----------|
| test-baseline | `KNOSSOS_HOME` set | No deprecation warning | Primary path |
| test-baseline | `ROSTER_HOME` set, no `KNOSSOS_HOME` | Deprecation warning to stderr, success | Fallback path |
| test-baseline | Neither set | Uses default, no warning | Default path |
| test-baseline | Both set | Uses `KNOSSOS_HOME`, no warning | Precedence |
| test-minimal | Full sync cycle | All paths resolve correctly | End-to-end |
| test-complex | Worktree creation | Go code resolves correctly | Go integration |

### Test Script Specification

```bash
#!/usr/bin/env bash
# test-knossos-home-resolution.sh

set -euo pipefail

TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../lib/knossos-home.sh"

echo "Test 1: KNOSSOS_HOME set"
(
    unset ROSTER_HOME
    export KNOSSOS_HOME="/custom/knossos"
    unset _KNOSSOS_HOME_RESOLVED
    resolve_knossos_home
    [[ "$KNOSSOS_HOME" == "/custom/knossos" ]] || exit 1
    echo "  PASS"
)

echo "Test 2: ROSTER_HOME fallback (expect deprecation warning)"
(
    unset KNOSSOS_HOME
    export ROSTER_HOME="/legacy/roster"
    unset _KNOSSOS_HOME_RESOLVED
    output=$(resolve_knossos_home 2>&1)
    [[ "$KNOSSOS_HOME" == "/legacy/roster" ]] || exit 1
    [[ "$output" == *"DEPRECATED"* ]] || exit 1
    echo "  PASS"
)

echo "Test 3: Default path"
(
    unset KNOSSOS_HOME
    unset ROSTER_HOME
    unset _KNOSSOS_HOME_RESOLVED
    resolve_knossos_home
    [[ "$KNOSSOS_HOME" == "$HOME/Code/roster" ]] || exit 1
    echo "  PASS"
)

echo "Test 4: KNOSSOS_HOME takes precedence"
(
    export KNOSSOS_HOME="/primary"
    export ROSTER_HOME="/fallback"
    unset _KNOSSOS_HOME_RESOLVED
    output=$(resolve_knossos_home 2>&1)
    [[ "$KNOSSOS_HOME" == "/primary" ]] || exit 1
    [[ "$output" != *"DEPRECATED"* ]] || exit 1
    echo "  PASS"
)

echo "All tests passed"
```

---

## Quality Gate Criteria

- [ ] `lib/knossos-home.sh` created and sourced by all primary scripts
- [ ] `ariadne/internal/config/home.go` created and used by worktree code
- [ ] `KNOSSOS_HOME` set -> No deprecation warning
- [ ] `ROSTER_HOME` only set -> Deprecation warning to stderr
- [ ] Neither set -> Default path used, no warning
- [ ] Both set -> `KNOSSOS_HOME` takes precedence
- [ ] All tests pass with `KNOSSOS_SUPPRESS_DEPRECATION=1`
- [ ] `docs/INTEGRATION.md` updated with new variable name
- [ ] User commands work with either variable
- [ ] `go test ./...` passes in ariadne
- [ ] All shell scripts lint clean (`shellcheck`)

---

## Implementation Sequence

### Sprint 1: Core Infrastructure

1. Create `lib/knossos-home.sh` with resolution function
2. Create `ariadne/internal/config/home.go` with Go resolver
3. Create test file `tests/test-knossos-home-resolution.sh`
4. Verify tests pass

### Sprint 2: Primary Scripts Migration

1. Update `swap-rite.sh` to source resolver
2. Update `roster-sync` to source resolver
3. Update `user-hooks/lib/config.sh`
4. Update `sync-user-*.sh` files
5. Run integration tests

### Sprint 3: Library and Support Scripts

1. Update `lib/rite/*.sh` files
2. Update `lib/sync/*.sh` files
3. Update `templates/*.sh` files
4. Update test files with `KNOSSOS_SUPPRESS_DEPRECATION`

### Sprint 4: Go Code Migration

1. Update `ariadne/internal/worktree/operations.go`
2. Update `ariadne/internal/worktree/lifecycle.go`
3. Run `go test ./...`
4. Build and test `ari` binary

### Sprint 5: Documentation

1. Update `docs/INTEGRATION.md`
2. Update `docs/migration/cem-to-roster-migration.md`
3. Update user commands and skills
4. Create migration announcement

---

## Files Changed Summary

### New Files (2)

1. `/Users/tomtenuta/Code/roster/lib/knossos-home.sh` - Shell resolution library
2. `/Users/tomtenuta/Code/roster/ariadne/internal/config/home.go` - Go resolution package

### Shell Scripts (30+ files)

Primary scripts requiring `source` addition:
- `swap-rite.sh`
- `roster-sync`
- `sync-user-hooks.sh`, `sync-user-skills.sh`, `sync-user-agents.sh`, `sync-user-commands.sh`
- `install-hooks.sh`
- `templates/orchestrator-generate.sh`, `templates/validate-orchestrator.sh`, `templates/generate-orchestrator.sh`
- `generate-rite-context.sh`, `load-workflow.sh`, `get-workflow-field.sh`
- `user-hooks/lib/config.sh`

Library scripts requiring variable rename:
- All files in `lib/rite/`, `lib/sync/`
- `user-hooks/lib/worktree-manager.sh`, `user-hooks/lib/rite-context-loader.sh`
- `user-hooks/validation/command-validator.sh`

### Go Files (3 files)

1. `/Users/tomtenuta/Code/roster/ariadne/internal/worktree/operations.go`
2. `/Users/tomtenuta/Code/roster/ariadne/internal/worktree/lifecycle.go`
3. New: `/Users/tomtenuta/Code/roster/ariadne/internal/config/home.go`

### Documentation (120+ files)

All files with `ROSTER_HOME` references need updating. High-priority user-facing docs first.

### Test Files (15+ files)

Add `KNOSSOS_SUPPRESS_DEPRECATION=1` to test harnesses.

---

## Notes for Integration Engineer

1. **Start with the resolution library**: Create `lib/knossos-home.sh` first. This is the single source of truth for variable resolution.

2. **Go code is isolated**: The 3 Go files are in worktree code only. Create the Go resolver package before updating them.

3. **Test early, test often**: Use `KNOSSOS_SUPPRESS_DEPRECATION=1` in tests to avoid warning noise.

4. **Documentation can be batched**: Use find/replace for `ROSTER_HOME` -> `KNOSSOS_HOME` in docs, but manually verify user-facing examples.

5. **Shellcheck everything**: All modified scripts should pass `shellcheck`.

6. **Deprecation warning is intentional stderr**: Don't redirect or suppress in production code. Users need to see it.

7. **Default path stays the same**: The default `$HOME/Code/roster` does NOT change to `$HOME/Code/knossos`. This is a variable rename, not a path change.

8. **Consider adding to ari CLI**: A future `ari config home` command could show the resolved path and any deprecation status.

---

## Artifact Attestation

| Source File | Operation |
|-------------|-----------|
| Grep output for `ROSTER_HOME` across codebase | Search |
| `/Users/tomtenuta/Code/roster/swap-rite.sh` | Read (lines 1-50) |
| `/Users/tomtenuta/Code/roster/roster-sync` | Read (lines 1-60) |
| `/Users/tomtenuta/Code/roster/user-hooks/lib/config.sh` | Read |
| `/Users/tomtenuta/Code/roster/ariadne/internal/worktree/operations.go` | Read |
| `/Users/tomtenuta/Code/roster/ariadne/internal/worktree/lifecycle.go` | Read |
| `/Users/tomtenuta/Code/roster/docs/assessments/skeleton-migration-risks.md` | Read (pattern reference) |
| `/Users/tomtenuta/Code/roster/docs/ecosystem/GAP-ANALYSIS-team-to-rite-migration.md` | Read (migration pattern) |
| Existing CONTEXT-DESIGN documents | Template reference |

---

## Handoff to Integration Engineer

This Context Design is complete and ready for implementation. The Integration Engineer should:

1. Create `lib/knossos-home.sh` with the resolution function
2. Create Go config package with `KnossosHome()` function
3. Create and run resolution tests
4. Update shell scripts in priority order
5. Update Go worktree files
6. Run full test suite
7. Update documentation
8. Verify quality gate criteria

No unresolved design decisions remain. The backward compatibility approach ensures zero breaking changes while establishing the path to full migration.

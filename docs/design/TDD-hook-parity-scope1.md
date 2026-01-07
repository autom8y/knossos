# TDD: Team Hooks Parity (Scope 1)

| Field | Value |
|-------|-------|
| **Sprint** | sprint-hook-parity-20251231 |
| **Task** | task-002 |
| **PRD Reference** | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-hook-ecosystem-parity.md` |
| **Requirements** | FR-1.1 through FR-1.6 |
| **Author** | Architect |
| **Status** | Draft |
| **Date** | 2025-12-31 |

## 1. Overview

This document specifies the technical design for achieving hook parity with other managed artifacts (agents, commands, skills) in the roster system. The scope covers:

1. Renaming `roster/hooks/` to `roster/user-hooks/` for naming consistency
2. Adding `hooks/` directory support to rite schema
3. Implementing base + team hook merging in `swap_hooks()`
4. Extending `AGENT_MANIFEST.json` to track hooks
5. Updating sync scripts for new paths

### 1.1 Design Principles

- **Consistency**: Follow established patterns from agents, commands, and skills
- **Atomicity**: Path changes must be atomic per NFR-5
- **Backwards Compatibility**: Existing installations continue working during migration (NFR-2)
- **Graceful Degradation**: Missing team hooks directory must not break swap-rite.sh (NFR-3)

### 1.2 Key Constraints

| Constraint | Source | Impact |
|------------|--------|--------|
| Atomic path changes | NFR-5 | Single commit for rename |
| Backwards compatibility | NFR-2 | Migration period required |
| Team hook override = warning | Edge Case | Log warning, use team hook |
| 100ms hook overhead limit | NFR-1 | No complex merge at runtime |

## 2. Architecture

### 2.1 Current State

```
roster/
  hooks/                          # Base hooks (misnamed)
    lib/                          # Shared libraries
    *.sh                          # Hook scripts
  user-agents/                    # Agent templates
  user-commands/                  # Command templates
  user-skills/                    # Skill templates
  rites/
    10x-dev/
      agents/                     # Team agents
      commands/                   # Team commands
      skills/                     # Team skills (some teams)
      workflow.yaml
```

### 2.2 Target State

```
roster/
  user-hooks/                     # Base hooks (renamed, FR-1.1)
    lib/                          # Shared libraries (unchanged)
    *.sh                          # Hook scripts
  user-agents/                    # Agent templates (unchanged)
  user-commands/                  # Command templates (unchanged)
  user-skills/                    # Skill templates (unchanged)
  rites/
    10x-dev/
      agents/                     # Team agents (unchanged)
      commands/                   # Team commands (unchanged)
      skills/                     # Team skills (unchanged)
      hooks/                      # Team hooks (NEW, FR-1.2)
        *.sh
      workflow.yaml
```

### 2.3 Component Diagram

```
                                  TEMPLATE SOURCES
                    +----------------------------------------+
                    |                                        |
       +------------v-----------+           +----------------v----------------+
       |   roster/user-hooks/   |           |   rites/<team>/hooks/           |
       |   (base hooks)         |           |   (team-specific hooks)         |
       +------------+-----------+           +----------------+----------------+
                    |                                        |
                    |        swap_hooks() merge logic        |
                    +------------------+---------------------+
                                       |
                                       v
                    +------------------+---------------------+
                    |        project/.claude/hooks/          |
                    |   (merged destination)                 |
                    |                                        |
                    |   .rite-hooks (marker file)            |
                    +----------------------------------------+
                                       |
                                       v
                    +------------------+---------------------+
                    |        AGENT_MANIFEST.json             |
                    |   hooks: { "file.sh": {...} }          |
                    +----------------------------------------+
```

### 2.4 Merge Strategy

The merge follows the same pattern as `swap_commands()` with one key difference: base hooks are installed first, then team hooks overlay:

```
MERGE ORDER:
1. Clear previous team hooks (remove files listed in .rite-hooks marker)
2. Copy base hooks from roster/user-hooks/ to .claude/hooks/
3. Copy team hooks from rites/<team>/hooks/ to .claude/hooks/
4. If collision (same filename): team hook wins with WARNING
5. Update .rite-hooks marker with list of team hook filenames
6. Update AGENT_MANIFEST.json with hooks section
```

**Collision Resolution**:
```
Base hook: session-context.sh      ->  .claude/hooks/session-context.sh
Team hook: session-context.sh      ->  OVERRIDES base (with warning)
Team hook: security-scan.sh        ->  .claude/hooks/security-scan.sh (new)
```

## 3. File Changes Matrix

| File | Change Type | Description |
|------|-------------|-------------|
| `roster/hooks/` | RENAME | Rename to `roster/user-hooks/` |
| `roster/user-hooks/` | CREATE | New location for base hooks |
| `rites/*/hooks/` | SCHEMA | Add hooks directory to rite schema |
| `swap-rite.sh` | MODIFY | Update `swap_hooks()` for merge logic |
| `swap-rite.sh` | MODIFY | Update `write_manifest()` for hooks tracking |
| `install-hooks.sh` | MODIFY | Update SOURCE_DIR path |
| `sync-user-hooks.sh` | MODIFY | Update SOURCE_DIR path |
| `.claude/AGENT_MANIFEST.json` | EXTEND | Add `hooks` section |

## 4. Detailed Design

### 4.1 Directory Rename (FR-1.1)

**Operation**: Atomic rename via `git mv`

```bash
git mv roster/hooks roster/user-hooks
```

**Verification**:
- All 12 hook scripts present in new location
- All 10 library scripts in `lib/` subdirectory
- No broken symlinks or references

**Files to Rename**:
| Current Path | New Path |
|--------------|----------|
| `roster/hooks/artifact-tracker.sh` | `roster/user-hooks/artifact-tracker.sh` |
| `roster/hooks/auto-park.sh` | `roster/user-hooks/auto-park.sh` |
| `roster/hooks/coach-mode.sh` | `roster/user-hooks/coach-mode.sh` |
| `roster/hooks/command-validator.sh` | `roster/user-hooks/command-validator.sh` |
| `roster/hooks/commit-tracker.sh` | `roster/user-hooks/commit-tracker.sh` |
| `roster/hooks/delegation-check.sh` | `roster/user-hooks/delegation-check.sh` |
| `roster/hooks/session-audit.sh` | `roster/user-hooks/session-audit.sh` |
| `roster/hooks/session-context.sh` | `roster/user-hooks/session-context.sh` |
| `roster/hooks/session-write-guard.sh` | `roster/user-hooks/session-write-guard.sh` |
| `roster/hooks/start-preflight.sh` | `roster/user-hooks/start-preflight.sh` |
| `roster/hooks/team-validator.sh` | (DELETE per FR-4.2) |
| `roster/hooks/workflow-validator.sh` | (DELETE per FR-4.3) |
| `roster/hooks/lib/*` | `roster/user-hooks/lib/*` |

### 4.2 Team-Pack Schema Extension (FR-1.2)

Rites may now include a `hooks/` directory with the same structure as base hooks:

```
rites/<team-name>/
  hooks/                    # OPTIONAL - team-specific hooks
    <hook-name>.sh          # Shell scripts only (*.sh)
```

**Schema Rules**:
- Directory is optional (graceful degradation per NFR-3)
- Only `*.sh` files are processed
- Hidden files (`.gitkeep`, etc.) are ignored
- No `lib/` subdirectory in team hooks (use base libs)

### 4.3 swap_hooks() Modification (FR-1.3)

**Current Implementation** (lines 2454-2521):
- Copies only team hooks from `rites/<team>/hooks/`
- Creates `.rite-hooks` marker
- Does NOT copy base hooks (assumes already installed)

**New Implementation**:

```bash
# Sync base hooks AND team-specific hooks to project
# Base hooks provide foundation, team hooks can override
swap_hooks() {
    local rite_name="$1"
    local base_hooks_dir="$ROSTER_HOME/user-hooks"
    local team_hooks_dir="$ROSTER_HOME/rites/$rite_name/hooks"

    log_debug "Syncing hooks: base=$base_hooks_dir, team=$team_hooks_dir"

    # Ensure hooks directory exists
    mkdir -p ".claude/hooks"
    mkdir -p ".claude/hooks/lib"

    # Backup and remove previous team hooks
    backup_team_hooks
    remove_team_hooks

    # =========================================================================
    # PHASE 1: Install base hooks from roster/user-hooks/
    # =========================================================================
    if [[ ! -d "$base_hooks_dir" ]]; then
        log_warning "Base hooks directory not found: $base_hooks_dir"
        # Continue anyway - team hooks may still work
    else
        log_debug "Installing base hooks from $base_hooks_dir"

        # Copy root-level hooks
        for hook_file in "$base_hooks_dir"/*.sh; do
            [[ -f "$hook_file" ]] || continue
            local hook_name
            hook_name=$(basename "$hook_file")

            # Skip hidden files
            [[ "$hook_name" == .* ]] && continue

            cp "$hook_file" ".claude/hooks/$hook_name"
            chmod +x ".claude/hooks/$hook_name"
            log_debug "Installed base hook: $hook_name"
        done

        # Copy lib/ directory contents
        if [[ -d "$base_hooks_dir/lib" ]]; then
            for lib_file in "$base_hooks_dir/lib"/*.sh; do
                [[ -f "$lib_file" ]] || continue
                local lib_name
                lib_name=$(basename "$lib_file")

                cp "$lib_file" ".claude/hooks/lib/$lib_name"
                chmod +x ".claude/hooks/lib/$lib_name" 2>/dev/null || true
                log_debug "Installed lib: $lib_name"
            done
        fi
    fi

    # =========================================================================
    # PHASE 2: Overlay team hooks (if team has hooks directory)
    # =========================================================================
    if [[ ! -d "$team_hooks_dir" ]]; then
        log_debug "Team $rite_name has no hooks/ directory"
        return 0
    fi

    local hook_count
    hook_count=$(find "$team_hooks_dir" -maxdepth 1 -type f -name "*.sh" 2>/dev/null | wc -l | tr -d ' ')

    if [[ "$hook_count" -eq 0 ]]; then
        log_debug "Team $rite_name has no hook files"
        return 0
    fi

    log_debug "Overlaying $hook_count hook(s) from team $rite_name"

    # Create marker file to track team hooks
    local marker_file=".claude/hooks/.rite-hooks"
    : > "$marker_file"

    # Copy each team hook (may override base hooks)
    for hook_file in "$team_hooks_dir"/*.sh; do
        [[ -f "$hook_file" ]] || continue

        local hook_name
        hook_name=$(basename "$hook_file")

        # Skip hidden files
        [[ "$hook_name" == .* ]] && continue

        # Check for collision with base hook
        if [[ -f ".claude/hooks/$hook_name" ]]; then
            log_warning "Team hook overrides base: $hook_name"
        fi

        cp "$hook_file" ".claude/hooks/$hook_name"
        chmod +x ".claude/hooks/$hook_name"
        echo "$hook_name" >> "$marker_file"
        log_debug "Installed team hook: $hook_name"
    done

    # Count successfully synced team hooks
    local synced_count
    synced_count=$(wc -l < "$marker_file" | tr -d ' ')

    if [[ "$synced_count" -gt 0 ]]; then
        log "Synced: $synced_count team hook(s)"
    fi
}
```

**Key Changes from Current Implementation**:

| Aspect | Current | New |
|--------|---------|-----|
| Base hooks | Not installed (assumed) | Installed from `user-hooks/` |
| Team hooks | Only team hooks | Overlay on base |
| Collision | Skip with warning | Override with warning |
| Marker file | Lists team hooks only | Lists team hooks only (unchanged) |

### 4.4 AGENT_MANIFEST.json Extension (FR-1.4)

**Current Schema**:
```json
{
  "manifest_version": "1.1",
  "active_team": "10x-dev",
  "last_swap": "2025-12-31T16:25:51Z",
  "agents": { ... },
  "commands": { ... }
}
```

**Extended Schema**:
```json
{
  "manifest_version": "1.2",
  "active_team": "10x-dev",
  "last_swap": "2025-12-31T16:25:51Z",
  "agents": { ... },
  "commands": { ... },
  "hooks": {
    "session-context.sh": {
      "source": "base",
      "origin": "user-hooks",
      "installed_at": "2025-12-31T16:25:51Z"
    },
    "security-scan.sh": {
      "source": "rite",
      "origin": "security",
      "installed_at": "2025-12-31T16:25:51Z"
    }
  }
}
```

**Hook Entry Schema**:
| Field | Type | Values | Description |
|-------|------|--------|-------------|
| `source` | string | `"base"`, `"team"` | Where hook originated |
| `origin` | string | `"user-hooks"` or rite name | Specific source identifier |
| `installed_at` | ISO 8601 | timestamp | When hook was installed |

**write_manifest() Modification**:

Add hooks section after commands section in the `write_manifest()` function (lines 1063-1093):

```bash
# After commands section...

# Close commands section, add comma for hooks
echo "" >> "$MANIFEST_FILE"
echo "  }," >> "$MANIFEST_FILE"

# Add hooks section
echo "  \"hooks\": {" >> "$MANIFEST_FILE"

local first_hook=true
local base_hooks_dir="$ROSTER_HOME/user-hooks"

# Track base hooks (not in .rite-hooks marker)
if [[ -d ".claude/hooks" ]]; then
    for hook_file in .claude/hooks/*.sh; do
        [[ ! -f "$hook_file" ]] && continue

        local hook_name
        hook_name=$(basename "$hook_file")

        # Determine source: team if in marker, base otherwise
        local source="base"
        local origin="user-hooks"

        if [[ -f ".claude/hooks/.rite-hooks" ]] && grep -q "^$hook_name$" ".claude/hooks/.rite-hooks" 2>/dev/null; then
            source="team"
            origin="$rite_name"
        fi

        if [[ "$first_hook" == true ]]; then
            first_hook=false
        else
            echo "," >> "$MANIFEST_FILE"
        fi

        {
            echo -n "    \"$hook_name\": {"
            echo -n "\"source\": \"$source\", "
            echo -n "\"origin\": \"$origin\", "
            echo -n "\"installed_at\": \"$timestamp\""
            echo -n "}"
        } >> "$MANIFEST_FILE"
    done
fi

# Close hooks and JSON
{
    echo ""
    echo "  }"
    echo "}"
} >> "$MANIFEST_FILE"
```

### 4.5 install-hooks.sh Update (FR-1.5)

**Current** (line 20):
```bash
readonly SOURCE_DIR="$ROSTER_HOME/hooks"
```

**New**:
```bash
readonly SOURCE_DIR="$ROSTER_HOME/user-hooks"
```

No other changes required - the script's logic remains identical.

### 4.6 sync-user-hooks.sh Update (FR-1.6)

**Current** (line 29):
```bash
readonly SOURCE_DIR="$ROSTER_HOME/hooks"
```

**New**:
```bash
readonly SOURCE_DIR="$ROSTER_HOME/user-hooks"
```

No other changes required - the script's logic remains identical.

## 5. Migration Sequence

### 5.1 Atomic Commit Strategy

Per NFR-5, all path changes must be in a single commit to avoid broken intermediate state:

```bash
# Single atomic commit containing:
git mv roster/hooks roster/user-hooks
# + All script path updates
# + Manifest version bump
# + ADR update
```

### 5.2 Migration Steps

```
PHASE 1: PREPARE (pre-commit)
  1. Create feature branch: feat/hook-parity-scope1
  2. Run existing tests to establish baseline

PHASE 2: ATOMIC COMMIT
  1. git mv roster/hooks roster/user-hooks
  2. Update install-hooks.sh SOURCE_DIR
  3. Update sync-user-hooks.sh SOURCE_DIR
  4. Update swap_hooks() in swap-rite.sh
  5. Update write_manifest() in swap-rite.sh
  6. Bump MANIFEST_VERSION to "1.2"
  7. Delete team-validator.sh and workflow-validator.sh (FR-4.2, FR-4.3)
  8. Commit with message: "feat(hooks): rename hooks/ to user-hooks/ and add team hook support"

PHASE 3: VERIFY (post-commit)
  1. Run all tests
  2. Test swap-rite.sh --dry-run with various teams
  3. Verify hooks installed correctly
  4. Verify manifest contains hooks section
  5. Test collision warning appears

PHASE 4: DOCUMENTATION
  1. Update ADR-0002 paths
  2. Update skill references
  3. Create migration notes for existing users
```

### 5.3 Backwards Compatibility (NFR-2)

Existing installations have hooks in `.claude/hooks/`. These continue to work because:

1. **Runtime unchanged**: Hooks still resolve libs via `$CLAUDE_PROJECT_DIR/.claude/hooks/lib`
2. **No auto-migration**: Users run `install-hooks.sh` or `sync-user-hooks.sh` to update
3. **Graceful skip**: `swap_hooks()` warns but continues if `user-hooks/` missing

**Migration Path for Existing Users**:
```bash
# After pulling new roster version:
./install-hooks.sh          # Updates project hooks from new location
./sync-user-hooks.sh        # Updates user-level hooks from new location
```

## 6. Rollback Plan

### 6.1 Quick Rollback

If issues discovered post-merge:

```bash
git revert <commit-sha>
```

This restores:
- `roster/hooks/` directory (via undo of git mv)
- Original `SOURCE_DIR` paths in scripts
- Original `swap_hooks()` implementation
- Original `write_manifest()` implementation

### 6.2 Partial Rollback

If only team hooks feature has issues:

```bash
# In swap_hooks(), add early return to skip team hook overlay:
swap_hooks() {
    local rite_name="$1"
    # ... base hooks installation ...

    # TEMPORARY: Disable team hooks until fixed
    log_warning "Team hooks temporarily disabled"
    return 0

    # ... team hooks overlay code ...
}
```

### 6.3 Manifest Downgrade

If manifest version causes issues:

```bash
# Manually edit AGENT_MANIFEST.json
# Change: "manifest_version": "1.2"
# To:     "manifest_version": "1.1"
# Remove: "hooks": { ... } section
```

## 7. Test Plan

### 7.1 Unit Tests

| Test | Expected Result |
|------|-----------------|
| `swap_hooks` with no team hooks dir | Base hooks installed, no error |
| `swap_hooks` with empty team hooks dir | Base hooks installed, no team hooks |
| `swap_hooks` with team hooks | Base + team hooks installed |
| `swap_hooks` with collision | Team hook wins, warning logged |
| `write_manifest` hooks section | Correct source/origin for each hook |
| `install-hooks.sh` new path | Finds and installs from user-hooks/ |
| `sync-user-hooks.sh` new path | Finds and syncs from user-hooks/ |

### 7.2 Integration Tests

| Test | Expected Result |
|------|-----------------|
| Full team swap cycle | Agents, commands, hooks all installed |
| Team swap with existing hooks | Previous team hooks removed, new installed |
| --dry-run shows hook changes | Hook add/remove listed in preview |
| Manifest accurately reflects state | All hooks tracked with correct origin |

### 7.3 Edge Case Tests

| Test | Expected Result |
|------|-----------------|
| Base hook dir missing | Warning, continue with team hooks |
| Team hook same name as base | Override with warning |
| Hook with special characters in name | Handled correctly |
| Empty team hooks directory | No error, base hooks only |
| Rapid team switches | No orphan hooks, clean state |

## 8. ADR: Hook Ecosystem Unification

### 8.1 Decision

Extend ADR-0002 to document:

1. **Rename**: `roster/hooks/` becomes `roster/user-hooks/` for consistency
2. **Team hooks**: Teams may include `hooks/` directory
3. **Merge strategy**: Base hooks first, team hooks overlay

### 8.2 Rationale

- **Consistency**: All user artifacts follow `user-<type>/` naming
- **Extensibility**: Teams can customize hook behavior
- **Predictability**: Clear merge order (base then team)
- **Traceability**: Manifest tracks hook provenance

### 8.3 Consequences

**Positive**:
- Consistent naming across all artifact types
- Teams can specialize hook behavior
- Manifest provides complete state tracking

**Negative**:
- One-time migration required for existing users
- Increased complexity in swap_hooks()

## 9. File Verification

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-hook-parity-scope1.md` | Created |
| PRD Reference | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-hook-ecosystem-parity.md` | Read |
| swap-rite.sh | `/Users/tomtenuta/Code/roster/swap-rite.sh` | Read (lines 2308-2521, 1003-1096) |
| install-hooks.sh | `/Users/tomtenuta/Code/roster/install-hooks.sh` | Read |
| sync-user-hooks.sh | `/Users/tomtenuta/Code/roster/sync-user-hooks.sh` | Read |
| AGENT_MANIFEST.json | `/Users/tomtenuta/Code/roster/.claude/AGENT_MANIFEST.json` | Read |
| ADR-0002 | `/Users/tomtenuta/Code/roster/docs/decisions/ADR-0002-hook-library-resolution-architecture.md` | Read |
| 10x-dev workflow | `/Users/tomtenuta/Code/roster/rites/10x-dev/workflow.yaml` | Read |
| Current hooks directory | `/Users/tomtenuta/Code/roster/hooks/` | Verified (12 hooks, 10 libs) |

## 10. Implementation Checklist

- [ ] Create feature branch `feat/hook-parity-scope1`
- [ ] `git mv roster/hooks roster/user-hooks`
- [ ] Update `install-hooks.sh` SOURCE_DIR to `$ROSTER_HOME/user-hooks`
- [ ] Update `sync-user-hooks.sh` SOURCE_DIR to `$ROSTER_HOME/user-hooks`
- [ ] Rewrite `swap_hooks()` with base + team merge logic
- [ ] Extend `write_manifest()` to include hooks section
- [ ] Bump MANIFEST_VERSION to "1.2"
- [ ] Delete `hooks/team-validator.sh` (FR-4.2)
- [ ] Delete `hooks/workflow-validator.sh` (FR-4.3)
- [ ] Update ADR-0002 with new paths and team hooks documentation
- [ ] Single atomic commit for all changes
- [ ] Verify all tests pass
- [ ] Test `--dry-run` output shows hook changes
- [ ] Test collision warning appears for same-name hooks
- [ ] Verify manifest has correct hooks section after swap

---

## Appendix A: Current swap_hooks() Implementation

Reference: `/Users/tomtenuta/Code/roster/swap-rite.sh` lines 2454-2521

The current implementation:
1. Only copies hooks from `rites/<team>/hooks/`
2. Creates `.rite-hooks` marker file
3. Skips project hooks (with warning)
4. Does NOT install base hooks (assumes already present)

## Appendix B: Related Functions in swap-rite.sh

| Function | Lines | Purpose |
|----------|-------|---------|
| `backup_team_hooks()` | 2309-2339 | Backup current team hooks |
| `remove_team_hooks()` | 2341-2363 | Remove team hooks by marker |
| `is_team_hook()` | 2365-2369 | Check if hook belongs to team |
| `get_hook_team()` | 2371-2379 | Get rite name for a hook |
| `detect_hook_orphans()` | 2381-2414 | Find orphan hooks from other teams |
| `remove_orphan_hooks()` | 2416-2450 | Handle orphan hook cleanup |
| `swap_hooks()` | 2454-2521 | Main hook sync function |
| `write_manifest()` | 1003-1096 | Write AGENT_MANIFEST.json |

# CEM Integration Points Analysis

> Analysis of all places where CEM (Claude Ecosystem Manager) is invoked within the roster codebase, documenting expected behavior and replacement strategies.

**Task**: task-004 (Identify CEM Integration Points)
**Date**: 2026-01-03
**Author**: Architect Agent

---

## Executive Summary

CEM is invoked from roster in **3 primary integration points**:
1. **worktree-manager.sh** - Direct bash invocation for worktree initialization
2. **/sync command** - Claude Code command orchestrating CEM operations
3. **/cem-debug command** - Diagnostic command for ecosystem troubleshooting

Additionally, CEM concepts are referenced in:
- **ecosystem-pack context injection** - CEM sync status detection
- **ecosystem-ref skill** - Documentation of CEM patterns
- **swap-rite.sh** - References skeleton baseline (indirect)

---

## Integration Point Catalog

### 1. worktree-manager.sh (Direct CEM Invocation)

**Location**: `/Users/tomtenuta/Code/roster/user-hooks/lib/worktree-manager.sh`

| Attribute | Value |
|-----------|-------|
| **Lines** | 225-251 |
| **CEM Commands** | `cem sync`, `cem init --force`, `cem init` |
| **Invocation Pattern** | Direct bash: `$SKELETON_HOME/cem <command>` |
| **Context** | Worktree creation lifecycle |

#### Code Reference

```bash
# Line 225-229: CEM existence check
if [[ ! -x "$SKELETON_HOME/cem" ]]; then
    git worktree remove --force "$wt_path" 2>/dev/null || true
    echo '{"error": "CEM not found at '"$SKELETON_HOME/cem"'. Cannot create worktree without ecosystem. Run: chmod +x $SKELETON_HOME/cem"}' >&2
    exit 1
fi

# Line 234-251: CEM sync/init based on manifest presence
if [[ -f "$wt_path/.claude/.cem/manifest.json" ]]; then
    # Already initialized - just sync to get latest
    if ! cem_output=$(cd "$wt_path" && "$SKELETON_HOME/cem" sync 2>&1); then
        # Sync failed - try force reinit
        if ! cem_output=$(cd "$wt_path" && "$SKELETON_HOME/cem" init --force 2>&1); then
            git worktree remove --force "$wt_path" 2>/dev/null || true
            echo '{"error": "CEM sync/init failed in worktree. Details: '"${cem_output:-unknown}"'"}' >&2
            exit 1
        fi
    fi
else
    # Not initialized - run init
    if ! cem_output=$(cd "$wt_path" && "$SKELETON_HOME/cem" init 2>&1); then
        git worktree remove --force "$wt_path" 2>/dev/null || true
        echo '{"error": "CEM init failed in worktree. Details: '"${cem_output:-unknown}"'"}' >&2
        exit 1
    fi
fi
```

#### Expected Inputs/Outputs

| Input | Description |
|-------|-------------|
| `$SKELETON_HOME` | Environment variable pointing to skeleton_claude |
| `$wt_path` | Worktree path being created |
| `.claude/.cem/manifest.json` | Presence indicates prior CEM initialization |

| Output | Description |
|--------|-------------|
| Exit 0 | CEM command succeeded |
| Exit non-0 | CEM command failed |
| stderr | Error messages captured in `cem_output` |

#### Error Handling

1. **CEM not found**: Removes worktree, emits JSON error, exits 1
2. **Sync failure**: Falls back to `init --force`
3. **Init failure**: Removes worktree, emits JSON error with details, exits 1

#### Replacement Strategy

**Option A**: Inline CEM Logic
- Port essential `cem sync` and `cem init` logic into worktree-manager.sh
- Eliminate external dependency on skeleton CEM script
- Pros: Self-contained, no external calls
- Cons: Duplicates logic, harder to maintain sync with CEM updates

**Option B**: CEM as Library
- Refactor CEM into sourceable bash library modules
- `source "$SKELETON_HOME/lib/cem-core.sh"` in worktree-manager.sh
- Pros: Single source of truth, modular
- Cons: Still depends on skeleton path

**Option C**: Roster-Native Ecosystem Manager
- Create roster-internal `sync-ecosystem.sh` that handles manifest/sync
- Worktree-manager calls roster's own sync mechanism
- Pros: Complete roster autonomy
- Cons: Requires maintaining parallel implementation

**Recommendation**: Option B for short-term, Option C for long-term roster independence.

---

### 2. /sync Command (Claude Code Command)

**Location**: `/Users/tomtenuta/Code/roster/user-commands/cem/sync.md`

| Attribute | Value |
|-----------|-------|
| **Lines** | 1-96 |
| **CEM Commands** | `install-user`, `sync`, `sync --refresh`, `init`, `status`, `diff`, `--help` |
| **Invocation Pattern** | Claude Code instructs user to run via bash |
| **Context** | User-initiated ecosystem synchronization |

#### CEM Commands Referenced

| Command | Line | Purpose |
|---------|------|---------|
| `cem install-user` | 33, 68 | Push skeleton updates to ~/.claude/ |
| `cem sync` | 41 | Pull updates from skeleton to satellite |
| `cem sync --refresh` | 44 | Sync + refresh active rite from roster |
| `cem init` | 51 | Initialize project with ecosystem |
| `cem status` | 54 | Show current sync state |
| `cem diff` | 57 | Show differences with skeleton |
| `cem --help` | 95 | Full CEM documentation |

#### Behavior Matrix

| Context | Default Behavior | Command |
|---------|------------------|---------|
| IN_SKELETON | Push to user-level | `cem install-user` |
| IN_SATELLITE | Pull from skeleton | `cem sync` |
| `--refresh` | Waterfall sync | `cem sync --refresh` |
| `--force` | Overwrite local | Add `--force` flag |
| `--dry-run` | Preview changes | Add `--dry-run` flag |

#### Expected Inputs/Outputs

| Input | Description |
|-------|-------------|
| `$ARGUMENTS` | User-provided subcommand/flags |
| `.claude/user-agents` directory | Presence indicates skeleton project |

| Output | Description |
|--------|-------------|
| CEM stdout | Sync results, status info |
| CEM stderr | Errors, warnings |
| Exit code | Success/failure indication |

#### Error Handling

- Command reports CEM output to user
- Explains conflicts and resolution steps
- References `cem --help` for detailed docs

#### Replacement Strategy

**Option A**: Direct Roster Sync
- Replace `/sync` with roster-native sync mechanism
- `roster sync` instead of `cem sync`
- Pros: Roster-centric, clearer ownership
- Cons: Requires implementing full sync logic in roster

**Option B**: Thin Wrapper
- Keep command structure, but call roster's sync-ecosystem.sh
- Maintain same UX, change backend
- Pros: Backward compatible UX
- Cons: Still needs roster sync implementation

**Option C**: Unified /roster Command
- New `/roster sync`, `/roster status`, `/roster init` commands
- Deprecate `/sync` in favor of namespace clarity
- Pros: Clear branding, extensible
- Cons: Breaking change for users

**Recommendation**: Option B initially, migrate to Option C over time.

---

### 3. /cem-debug Command (Diagnostic)

**Location**: `/Users/tomtenuta/Code/roster/rites/ecosystem-pack/commands/cem-debug.md`

| Attribute | Value |
|-----------|-------|
| **Lines** | 1-78 |
| **CEM Commands** | None directly (diagnostic focus) |
| **Invocation Pattern** | Claude Code analysis, references CEM source |
| **Context** | Troubleshooting CEM sync issues |

#### Purpose

Invokes Ecosystem Analyst agent with CEM diagnostic focus to:
- Reproduce reported sync issues
- Examine CEM sync logs and error messages
- Trace conflict detection and resolution logic
- Check settings schema compatibility
- Verify CEM version compatibility

#### CEM References

| Reference | Line | Description |
|-----------|------|-------------|
| "cem sync" | 36 | Issue trigger condition |
| "CEM source: `$ROSTER_HOME/cem`" | 77 | Points to CEM for source analysis |

#### Expected Outputs

Gap Analysis document at: `docs/ecosystem/GAP-{issue-slug}.md`

#### Replacement Strategy

**Option A**: Roster Diagnostics
- Replace with `/roster debug` or `/ecosystem debug`
- Point to roster-internal diagnostic tooling
- Pros: Self-contained diagnostics
- Cons: Need to build diagnostic infrastructure

**Option B**: Keep as Ecosystem-Pack Tool
- Command stays in ecosystem-pack
- Update references from CEM to roster internals
- Pros: Minimal change, still useful
- Cons: References may become stale

**Recommendation**: Option B with updated references once roster sync is implemented.

---

## Indirect CEM References

### 4. ecosystem-pack/context-injection.sh

**Location**: `/Users/tomtenuta/Code/roster/rites/ecosystem-pack/context-injection.sh`

| Attribute | Value |
|-----------|-------|
| **Lines** | 17-35 |
| **CEM Concept** | CEM sync status detection |
| **Invocation Pattern** | File timestamp check |

#### CEM References

```bash
# Line 18: CEM sync file path
local cem_sync_file="$project_dir/.claude/.cem-sync"

# Lines 20-34: Sync status detection
if [[ -f "$cem_sync_file" ]]; then
    cem_timestamp=$(stat -f "%Sm" -t "%Y-%m-%d %H:%M" "$cem_sync_file" 2>/dev/null || ...)
    if is_file_stale "$cem_sync_file" 1440 2>/dev/null; then
        cem_status="stale"
    else
        cem_status="synced"
    fi
else
    cem_status="never synced"
fi
```

#### Replacement Strategy

- Update to check roster-native sync status file
- Replace `.claude/.cem-sync` with `.claude/.roster-sync` or similar
- Minimal change, just path update

---

### 5. ecosystem-ref Skill

**Location**: `/Users/tomtenuta/Code/roster/rites/ecosystem-pack/skills/ecosystem-ref/SKILL.md`

| Attribute | Value |
|-----------|-------|
| **Lines** | 1-103 |
| **CEM Concept** | Documentation of CEM patterns |
| **Invocation Pattern** | Reference documentation |

#### CEM Documentation Provided

| Section | Content |
|---------|---------|
| File Strategies | copy-replace, merge-dir, merge-settings, merge-docs |
| Key Paths | Manifest at `.claude/.cem/manifest.json` |
| Common Commands | `cem init`, `cem sync`, `cem sync --refresh`, `cem validate`, `cem repair`, `cem status` |
| Debugging | `CEM_DEBUG=1 cem sync` |

#### Replacement Strategy

- Update documentation to reflect roster-native sync mechanism
- Keep as reference until CEM is fully deprecated
- Add migration notes for users transitioning from CEM

---

### 6. swap-rite.sh (Skeleton References)

**Location**: `/Users/tomtenuta/Code/roster/swap-rite.sh`

| Attribute | Value |
|-----------|-------|
| **Lines** | 1554, 3979, 4065, etc. |
| **CEM Concept** | Skeleton baseline concept |
| **Invocation Pattern** | Indirect (no direct CEM calls) |

#### References

- `--reset`: Reset to skeleton baseline
- `regenerate_skeleton_claude_md()`: Regenerates CLAUDE.md for skeleton baseline
- Multiple references to "skeleton" as conceptual baseline

#### Note

swap-rite.sh does NOT invoke CEM directly. It references skeleton as a concept for the base configuration layer. This is more about the three-tier architecture (skeleton -> team -> satellite) than CEM itself.

---

## CEM Command Dependency Graph

```
                    ┌─────────────────────────────────────────────┐
                    │           skeleton_claude/cem               │
                    │  (Claude Ecosystem Manager - Bash Script)   │
                    └─────────────────────────────────────────────┘
                                        │
         ┌──────────────────────────────┼──────────────────────────────┐
         │                              │                              │
         ▼                              ▼                              ▼
┌─────────────────┐          ┌─────────────────┐          ┌─────────────────┐
│  worktree-      │          │   /sync         │          │  /cem-debug     │
│  manager.sh     │          │   command       │          │   command       │
│                 │          │                 │          │                 │
│ Direct bash     │          │ User-initiated  │          │ Diagnostic      │
│ invocation      │          │ sync via Claude │          │ troubleshooting │
└─────────────────┘          └─────────────────┘          └─────────────────┘
         │                              │                              │
         ▼                              ▼                              ▼
┌─────────────────┐          ┌─────────────────┐          ┌─────────────────┐
│  Worktree       │          │   Satellite     │          │   Gap Analysis  │
│  Creation       │          │   .claude/      │          │   Document      │
└─────────────────┘          └─────────────────┘          └─────────────────┘
```

---

## CEM Library Dependencies

The skeleton CEM script sources these library modules:

| Module | Location | Purpose |
|--------|----------|---------|
| `cem-config.sh` | `$SKELETON_HOME/lib/` | Constants, state management |
| `cem-logging.sh` | `$SKELETON_HOME/lib/` | Log functions |
| `cem-git.sh` | `$SKELETON_HOME/lib/` | Git operations |
| `cem-checksum.sh` | `$SKELETON_HOME/lib/` | Checksum computation |
| `cem-manifest.sh` | `$SKELETON_HOME/lib/` | Manifest reading/writing |
| `cem-sync.sh` | `$SKELETON_HOME/lib/` | Sync command handler |
| `cem-merge/dispatcher.sh` | `$SKELETON_HOME/lib/cem-merge/` | Merge strategy dispatcher |
| `cem-merge/merge-dir.sh` | `$SKELETON_HOME/lib/cem-merge/` | Directory merge |
| `cem-merge/merge-settings.sh` | `$SKELETON_HOME/lib/cem-merge/` | Settings JSON merge |
| `cem-merge/merge-docs.sh` | `$SKELETON_HOME/lib/cem-merge/` | Documentation merge |

---

## CEM Commands Summary

| Command | Arguments | Description | Used By |
|---------|-----------|-------------|---------|
| `init` | `[skeleton-path]` | Initialize project with ecosystem | worktree-manager.sh, /sync |
| `sync` | `[--refresh] [--prune] [--force] [--dry-run]` | Pull updates from skeleton | worktree-manager.sh, /sync |
| `validate` | - | Validate manifest and file integrity | /sync (status) |
| `validate-team` | `[name]` | Validate rite against workflow schema | - |
| `repair` | - | Rebuild manifest from .claude/ state | /sync |
| `install-user` | - | Install user-level resources to ~/.claude/ | /sync |
| `status` | - | Show sync status and version info | /sync |
| `diff` | `[path]` | Show differences with skeleton | /sync |
| `version` | - | Show version information | - |
| `alias` | - | Show shell alias setup | - |

---

## State Files and Artifacts

| Path | Owner | Purpose |
|------|-------|---------|
| `.claude/.cem/manifest.json` | CEM | Sync state, managed files, checksums |
| `.claude/.cem/checksum-cache.json` | CEM | Cached checksums for performance |
| `.claude/.cem/orphan-backup/` | CEM | Backup of pruned orphan files |
| `.claude/.cem-sync` | CEM | Timestamp file for sync freshness |
| `~/.claude/USER_*_MANIFEST.json` | CEM (install-user) | User-level resource manifests |

---

## Replacement Strategy Summary

| Integration Point | Short-Term | Long-Term |
|-------------------|------------|-----------|
| worktree-manager.sh | Source CEM as library (Option B) | Roster-native ecosystem manager (Option C) |
| /sync command | Thin wrapper to roster sync (Option B) | Unified /roster command (Option C) |
| /cem-debug | Update references (Option B) | Roster diagnostics tool (Option A) |
| context-injection.sh | Path update | Path update |
| ecosystem-ref skill | Add migration notes | Full documentation update |

---

## Success Criteria

- [ ] All CEM invocations documented with file:line references
- [ ] Dependencies between CEM features mapped
- [ ] Clear replacement path for each integration point
- [ ] State files and artifacts cataloged
- [ ] Error handling expectations documented

---

## Attestation Table

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| worktree-manager.sh | `/Users/tomtenuta/Code/roster/user-hooks/lib/worktree-manager.sh` | Lines 225-251 read |
| /sync command | `/Users/tomtenuta/Code/roster/user-commands/cem/sync.md` | Full file read |
| /cem-debug command | `/Users/tomtenuta/Code/roster/rites/ecosystem-pack/commands/cem-debug.md` | Full file read |
| context-injection.sh | `/Users/tomtenuta/Code/roster/rites/ecosystem-pack/context-injection.sh` | Full file read |
| ecosystem-ref skill | `/Users/tomtenuta/Code/roster/rites/ecosystem-pack/skills/ecosystem-ref/SKILL.md` | Full file read |
| CEM main script | `/Users/tomtenuta/Code/skeleton_claude/cem` | Full file read |
| cem-sync.sh library | `/Users/tomtenuta/Code/skeleton_claude/lib/cem-sync.sh` | Full file read |
| manifest.json | `/Users/tomtenuta/Code/roster/.claude/.cem/manifest.json` | Full file read |

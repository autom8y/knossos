# PENETRATION REPORT: Legacy Shell Script vs ari CLI Routing

**Date**: 2026-01-07
**Initiative**: knossos-finalization v0.1.0 Post-Release Audit
**Auditor**: Claude Code Agent (Explore Mode)

## Executive Summary

The v0.1.0 release introduced the Go-based `ari` CLI as the new canonical interface for Knossos operations, but **significant routing inconsistencies remain**. This audit identified **21 findings** across multiple severity levels where legacy shell scripts are invoked instead of their Go equivalents.

**Key Issue**: The `/sync` command routes to `roster-sync` (shell script) instead of `ari sync` (Go CLI), even though `ari sync` has full functionality including `materialize`, `pull`, `push`, `status`, `diff`, `resolve`, `history`, and `reset` commands.

---

## Findings Summary Table (Sorted by Severity)

| ID | Severity | Location | Issue | Should Be |
|----|----------|----------|-------|-----------|
| LR-001 | **CRITICAL** | `user-commands/cem/sync.md` | `/sync` routes to `roster-sync` shell script | `ari sync` |
| LR-002 | **CRITICAL** | `~/.claude/commands/sync.md` | Materialized `/sync` routes to `roster-sync` | `ari sync` |
| LR-003 | **HIGH** | `internal/worktree/lifecycle.go:139` | Go code invokes `roster-sync` shell script | `internal/sync` package |
| LR-004 | **HIGH** | `internal/worktree/operations.go:669` | Go code invokes `roster-sync` for worktree setup | `internal/sync` package |
| LR-005 | **HIGH** | `internal/worktree/lifecycle.go:160` | Go code invokes `swap-rite.sh` | `ari rite swap` or `ari sync materialize --rite` |
| LR-006 | **HIGH** | `internal/worktree/operations.go:685` | Go code invokes `swap-rite.sh` for rite setup | `ari rite swap` |
| LR-007 | **HIGH** | `user-skills/guidance/rite-ref/SKILL.md:49` | Documents `swap-rite.sh` as the command to use | `ari rite` or `ari sync materialize --rite` |
| LR-008 | **MEDIUM** | `user-hooks/lib/worktree-manager.sh:227-258` | Shell code invokes `roster-sync` and `swap-rite.sh` | Should call `ari` commands |
| LR-009 | **MEDIUM** | `lib/sync/sync-core.sh:1033` | Shell sync calls `swap-rite.sh` for refresh | Should use `ari sync materialize` |
| LR-010 | **MEDIUM** | `user-commands/rite-switching/` (10+ files) | Call `swap-rite.sh` directly | Should call `ari rite swap` |
| LR-011 | **MEDIUM** | `user-commands/navigation/rite.md:24-28` | Routes to `swap-rite.sh` | `ari rite` commands |
| LR-012 | **MEDIUM** | `user-commands/navigation/ecosystem.md:17` | Calls `swap-rite.sh ecosystem` | `ari sync materialize --rite ecosystem` |
| LR-013 | **MEDIUM** | `sync-user-*.sh` (4 files, 3,786 lines) | Still exist as primary user-level sync | Potentially deprecated by `ari sync` |
| LR-014 | **LOW** | `.claude/CLAUDE.md:122` | References `cd ariadne && just build` | `just build` at root |
| LR-015 | **LOW** | `user-skills/operations/worktree-ref/integration.md:99-105` | Documents `swap-rite.sh` usage | Should document `ari` |
| LR-016 | **LOW** | `rites/ecosystem/skills/ecosystem-ref/SKILL.md` | Documents shell scripts as primary | Should document `ari` |
| LR-017 | **LOW** | `docs/guides/sync-materialization-migration.md:16` | Doesn't emphasize new approach | Prioritize `ari sync materialize` |
| LR-018 | **LOW** | `README.md` | Documents `sync-user-*.sh` as primary | Should document `ari sync` |
| LR-019 | **LOW** | `docs/INTEGRATION.md:79-88` | Documents `swap-rite.sh` as primary | Should document `ari rite` |
| LR-020 | **LOW** | Multiple rite skill files | Reference `swap-rite.sh` | Should reference `ari` |
| LR-021 | **INFO** | `swap-rite.sh` (3,773 lines) | Still exists and functional | Document deprecation timeline |

---

## Issue Counts by Severity

| Severity | Count | Description |
|----------|-------|-------------|
| CRITICAL | 2 | Main `/sync` command routes to legacy shell |
| HIGH | 5 | Go code shells out to bash; key skill docs wrong |
| MEDIUM | 6 | Commands and hooks use legacy scripts |
| LOW | 7 | Documentation inconsistencies |
| INFO | 1 | Legacy script still exists |

---

## Detailed Findings

### CRITICAL: LR-001 & LR-002 - /sync Command Routes to Shell Script

**Location**:
- Source: `user-commands/cem/sync.md`
- Materialized: `~/.claude/commands/sync.md`

**Evidence** (lines 17-21):
```markdown
1. **Execute roster-sync** using standard path resolution:
   ```bash
   ${KNOSSOS_HOME:-~/Code/roster}/roster-sync [command] $ARGUMENTS
   ```
```

**Impact**: The primary `/sync` command invokes a 1,413-line bash script instead of the Go CLI with full capabilities.

**Should Be**:
```markdown
1. **Execute ari sync** using installed binary or local build:
   ```bash
   ari sync [command] $ARGUMENTS
   # OR if not in PATH:
   ~/bin/ari sync [command] $ARGUMENTS
   ```
```

**Migration Path**:
1. Update `user-commands/cem/sync.md` to call `ari sync`
2. Run `sync-user-commands.sh` to propagate changes
3. Test all subcommands: status, pull, push, diff, materialize, resolve

---

### HIGH: LR-003 & LR-004 - Go Code Invokes Shell Scripts

**Location**: `internal/worktree/lifecycle.go:136-155`

**Evidence**:
```go
// Try to run roster-sync if available
knossosHome := config.KnossosHome()
if knossosHome != "" {
    syncPath := filepath.Join(knossosHome, "roster-sync")
    if _, err := os.Stat(syncPath); err == nil {
        // ... invokes shell script
    }
}
```

**Impact**: The Go codebase shells out to bash scripts instead of using internal Go packages.

**Should Be**: Use `internal/sync` and `internal/materialize` packages directly:
```go
import "github.com/autom8y/knossos/internal/materialize"

// Use Go implementation
if err := materialize.Run(wtPath, rite); err != nil {
    // handle error
}
```

**Migration Path**:
1. Refactor `setupWorktreeEcosystem()` to use Go packages
2. Remove shell exec calls
3. Add integration tests

---

### HIGH: LR-005 & LR-006 - swap-rite.sh Called from Go Code

**Location**:
- `internal/worktree/lifecycle.go:160`
- `internal/worktree/operations.go:685`

**Evidence**:
```go
swapRitePath := filepath.Join(knossosHome, "swap-rite.sh")
if _, err := os.Stat(swapRitePath); err == nil {
    cmd := exec.Command(swapRitePath, rite)
```

**Impact**: Go code depends on 3,773-line bash script for rite switching.

**Should Be**: Use `ari rite swap` or `internal/rite` package.

---

### HIGH: LR-007 - rite-ref Skill Documents Shell Script

**Location**: `user-skills/guidance/rite-ref/SKILL.md:49-52`

**Evidence**:
```markdown
### 2. Invoke Roster Script

Execute the swap-rite.sh script via Bash tool:

```bash
$KNOSSOS_HOME/swap-rite.sh [args]
```

**Impact**: Primary rite switching documentation points to legacy script.

**Should Be**: Document `ari rite` commands:
```markdown
### 2. Invoke Ari CLI

Execute rite operations via ari:

```bash
ari rite swap <rite-name>
ari rite list
ari rite current
```
```

---

### MEDIUM: LR-010 - Rite-Switching Commands Use swap-rite.sh

**Location**: All files in `user-commands/rite-switching/`:
- `10x.md`, `debt.md`, `docs.md`, `hygiene.md`, `intelligence.md`, `rnd.md`, `security.md`, `sre.md`, `strategy.md`

**Evidence** (from `10x.md` lines 17-18):
```markdown
1. Execute: `${KNOSSOS_HOME:-~/Code/roster}/swap-rite.sh 10x-dev $ARGUMENTS`
2. Display the roster output from swap-rite.sh
```

**Impact**: 9+ commands route to shell script.

**Should Be**:
```markdown
1. Execute: `ari sync materialize --rite 10x-dev` OR `ari rite swap 10x-dev`
```

---

### MEDIUM: LR-013 - sync-user-*.sh Scripts Still Primary

**Location**: Root directory
- `sync-user-agents.sh` (734 lines)
- `sync-user-commands.sh` (959 lines)
- `sync-user-hooks.sh` (1,096 lines)
- `sync-user-skills.sh` (997 lines)

**Evidence**: These scripts are documented as primary sync method in README.md.

**Impact**: 3,786 lines of shell duplicating `ari sync` functionality for user-level sync.

**Assessment**: These may still be needed for user-level (`~/.claude/`) sync vs project-level. Scope clarification needed:
- `ari sync materialize` → project-level `.claude/`
- `sync-user-*.sh` → user-level `~/.claude/`

If `ari sync` should handle both, these scripts are deprecated.

---

## ari Commands Available (Reference)

```
ari sync status          # Show sync status
ari sync pull            # Pull remote changes
ari sync push            # Push local changes
ari sync diff            # Show differences
ari sync materialize     # Generate .claude/ from templates
ari sync resolve         # Resolve conflicts
ari sync history         # Show sync history
ari sync reset           # Reset sync state

ari rite list            # List available rites
ari rite current         # Show active rite
ari rite swap <name>     # Switch to rite
ari rite invoke <name>   # Borrow components
ari rite release <name>  # Release borrowed
ari rite validate        # Validate integrity
```

---

## Recommended Migration Plan

### Phase 1: Critical (Immediate - P0)
1. **Update `/sync` command** (LR-001, LR-002)
   - Modify `user-commands/cem/sync.md` to call `ari sync`
   - Run sync to propagate to user-level
   - Test all subcommands
   - **Effort**: 1 hour
   - **Impact**: Critical - main user-facing command

### Phase 2: High Priority (P1 - Within 1 week)
2. **Refactor Go worktree code** (LR-003, LR-004, LR-005, LR-006)
   - Replace shell exec with Go package calls
   - Add tests for new implementation
   - **Effort**: 1 day

3. **Update rite-ref skill** (LR-007)
   - Document `ari rite` as primary interface
   - Add migration notes from swap-rite.sh
   - **Effort**: 2 hours

### Phase 3: Medium Priority (P2 - Within 2 weeks)
4. **Update all rite-switching commands** (LR-010, LR-011, LR-012)
   - Batch update 11+ command files
   - Test each command after update
   - **Effort**: 3 hours

5. **Update shell library code** (LR-008, LR-009)
   - Update worktree-manager.sh
   - Update sync-core.sh
   - **Effort**: 2 hours

6. **Clarify sync-user-*.sh scope** (LR-013)
   - Document whether deprecated or still needed
   - If deprecated, add deprecation warnings
   - If needed, document distinction from `ari sync`
   - **Effort**: 4 hours

### Phase 4: Low Priority (P3 - Within 1 month)
7. **Update documentation** (LR-014 through LR-021)
   - Update CLAUDE.md, README.md, INTEGRATION.md
   - Update all rite skill files
   - Add deprecation notices to shell scripts
   - **Effort**: 1 day

---

## Priority Order for Fixes

| Priority | Finding IDs | Effort | Impact |
|----------|-------------|--------|--------|
| **P0** | LR-001, LR-002 | 1 hour | Critical - main user-facing command |
| **P1** | LR-003, LR-004, LR-005, LR-006 | 1 day | High - Go code quality |
| **P1** | LR-007 | 2 hours | High - primary skill documentation |
| **P2** | LR-010, LR-011, LR-012 | 3 hours | Medium - multiple commands |
| **P2** | LR-008, LR-009 | 2 hours | Medium - shell library code |
| **P2** | LR-013 | 4 hours | Medium - scope clarification needed |
| **P3** | LR-014 to LR-021 | 1 day | Low - documentation updates |

---

## Verification Checklist

After migration:
- [ ] `/sync status` calls `ari sync status`
- [ ] `/sync materialize` calls `ari sync materialize`
- [ ] `/rite list` calls `ari rite list`
- [ ] `/rite <name>` calls `ari rite swap` or `ari sync materialize --rite`
- [ ] Go worktree code uses internal packages
- [ ] No shell exec to roster-sync in Go code
- [ ] No shell exec to swap-rite.sh in Go code
- [ ] All rite-switching commands use ari CLI
- [ ] Documentation reflects ari as primary interface

---

## Appendix: Files Analyzed

| Category | Count | Key Findings |
|----------|-------|--------------|
| Commands | 20+ in user-commands/ | All rite-switching use swap-rite.sh |
| Skills | 50+ in user-skills/ | rite-ref documents shell script |
| Go Code | internal/worktree/*.go | Shells out to bash scripts |
| Shell Scripts | 6 at root | roster-sync, swap-rite.sh, sync-user-*.sh |
| Hooks | .claude/hooks/ari/*.sh | Properly use ari binary |
| Documentation | docs/**/*.md | Mixed - some updated, many still reference shell |

---

## Conclusion

The v0.1.0 release successfully introduced the `ari` CLI, but the migration is **incomplete**. Critical user-facing commands still route to legacy shell scripts, and Go code shells out to bash instead of using internal packages.

**Recommendation**: Address P0 findings (LR-001, LR-002) immediately before wrapping the knossos-finalization session. This is a 1-hour fix that significantly improves the user experience and validates the new architecture.

---

**Audit Completed**: 2026-01-07
**Next Review**: After P0/P1 fixes applied

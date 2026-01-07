# Gap Analysis: Command Argument Standardization

**Date**: 2025-12-29
**Analyst**: Ecosystem Analyst
**Sprint**: Command Argument Standardization
**Status**: Ready for Context Architect

---

## Executive Summary

| Metric | Count |
|--------|-------|
| **Total commands** | 32 |
| **Missing argument-hint** | 3 |
| **Missing $ARGUMENTS in task** | 0 |
| **Missing $ARGUMENTS pass-through** | 9 |
| **Using deprecated --force** | 11 |
| **Using deprecated --refresh** | 0 |
| **Fully compliant** | 13 |

### Critical Findings

1. **--force flag is documented in 11 commands but does NOT exist in swap-rite.sh** - This is a documentation bug. The script only supports `--refresh`, not `--force`.

2. **9 commands have $ARGUMENTS in task but NOT in the Behavior execution** - Arguments may not actually reach swap-rite.sh.

3. **3 commands lack argument-hint frontmatter** - Minor completeness issue.

---

## swap-rite.sh Actual Flags

From analysis of `/roster/swap-rite.sh` (lines 1587-1661):

| Flag | Short | Purpose | Lines |
|------|-------|---------|-------|
| `--list` | `-l` | List all available rites | 1617-1619 |
| `--help` | `-h` | Display usage information | 1620-1623 |
| `--keep-all` | - | Preserve orphan agents in project | 1624-1627 |
| `--remove-all` | - | Remove orphan agents (backup available) | 1628-1631 |
| `--promote-all` | - | Move orphan agents to ~/.claude/agents/ | 1632-1635 |
| `--refresh` | `-r` | Refresh agents even if already on team | 1636-1639 |
| `--dry-run` | - | Preview changes without applying | 1640-1644 |

### Flags That DO NOT EXIST (But Are Documented)

| Flag | Commands That Document It |
|------|---------------------------|
| `--force` | 10x.md, debt.md, docs.md, hygiene.md, intelligence.md, rnd.md, security.md, sre.md, strategy.md, ecosystem.md, team.md |

**Root Cause**: The `--force` flag appears to be a legacy or planned feature that was never implemented. The rite-switching commands document it, but swap-rite.sh does not accept it.

---

## Command Inventory

### rite-switching/ (10 commands)

| Command | argument-hint | $ARGUMENTS in Task | Pass-through | Flags Accurate | Deprecated Flags |
|---------|--------------|-------------------|--------------|----------------|------------------|
| 10x.md | Y | Y | Y | **N** | --force (invalid) |
| debt.md | Y | Y | Y | **N** | --force (invalid) |
| docs.md | Y | Y | Y | **N** | --force (invalid) |
| forge.md | Y | Y | N (not swap) | Y | None |
| hygiene.md | Y | Y | Y | **N** | --force (invalid) |
| intelligence.md | Y | Y | Y | **N** | --force (invalid) |
| rnd.md | Y | Y | Y | **N** | --force (invalid) |
| security.md | Y | Y | Y | **N** | --force (invalid) |
| sre.md | Y | Y | Y | **N** | --force (invalid) |
| strategy.md | Y | Y | Y | **N** | --force (invalid) |

**Pattern**: All rite-switching commands (except forge.md) have identical structure and all document the non-existent `--force` flag.

**Note on forge.md**: This is a different type of command - it displays Forge meta-team info, not a team swap. It uses `--agents`, `--workflow`, `--commands` flags which are handled internally.

### session/ (5 commands)

| Command | argument-hint | $ARGUMENTS in Task | Pass-through | Flags Accurate | Deprecated Flags |
|---------|--------------|-------------------|--------------|----------------|------------------|
| continue.md | Y | Y | **N** | Y | None |
| handoff.md | Y | Y | **N** | Y | None |
| park.md | Y | Y | **N** | Y | None |
| start.md | Y | Y | **N** | Y | None |
| wrap.md | Y | Y | **N** | Y | None |

**Pattern**: All session commands have proper frontmatter but do not pass-through $ARGUMENTS to behavior - they parse arguments manually in the behavior section. This is **acceptable** for these commands as they have complex argument handling.

### operations/ (5 commands)

| Command | argument-hint | $ARGUMENTS in Task | Pass-through | Flags Accurate | Deprecated Flags |
|---------|--------------|-------------------|--------------|----------------|------------------|
| architect.md | Y | Y | **N** | Y | None |
| build.md | Y | Y | **N** | Y | None |
| code-review.md | Y | Y | **N** | Y | None |
| commit.md | Y | Y | **N** | Y | None |
| qa.md | Y | Y | **N** | Y | None |

**Pattern**: Operations commands have proper frontmatter. Arguments are handled within behavior logic, not passed through to external scripts. This is **acceptable**.

### navigation/ (5 commands)

| Command | argument-hint | $ARGUMENTS in Task | Pass-through | Flags Accurate | Deprecated Flags |
|---------|--------------|-------------------|--------------|----------------|------------------|
| consult.md | Y | Y | **N** | Y | None |
| ecosystem.md | Y | Y | Y | **N** | --force (invalid) |
| sessions.md | Y | Y | **N** | Y | None |
| team.md | Y | Y | **N** | **N** | --force (invalid) |
| worktree.md | Y | Y | **N** | Y | None |

**Note**: ecosystem.md follows the same pattern as rite-switching commands (it's actually a team switcher).

### workflow/ (3 commands)

| Command | argument-hint | $ARGUMENTS in Task | Pass-through | Flags Accurate | Deprecated Flags |
|---------|--------------|-------------------|--------------|----------------|------------------|
| hotfix.md | Y | Y | **N** | Y | None |
| sprint.md | Y | Y | **N** | Y | None |
| task.md | Y | Y | **N** | Y | None |

**Pattern**: Workflow commands have proper frontmatter. Arguments handled internally. Compliant.

### cem/ (1 command)

| Command | argument-hint | $ARGUMENTS in Task | Pass-through | Flags Accurate | Deprecated Flags |
|---------|--------------|-------------------|--------------|----------------|------------------|
| sync.md | **N** | Y | **N** | Y | None |

**Note**: sync.md lacks `argument-hint` in frontmatter but has proper behavior. Minor issue.

### meta/ (3 commands)

| Command | argument-hint | $ARGUMENTS in Task | Pass-through | Flags Accurate | Deprecated Flags |
|---------|--------------|-------------------|--------------|----------------|------------------|
| minus-1.md | **N** | N (uses {TAG}) | N/A | N/A | None |
| one.md | **N** | N (uses {TAG}) | N/A | N/A | None |
| zero.md | **N** | N (uses {TAG}) | N/A | N/A | None |

**Note**: Meta commands use a different pattern (`{TAG}` placeholder instead of `$ARGUMENTS`). These are special prompter commands with a different contract.

---

## Detailed Analysis by Issue Category

### Issue 1: Non-existent --force Flag Documentation

**Severity**: HIGH - User confusion, commands promise functionality that doesn't exist

**Affected Commands** (11 total):
- `/roster/user-commands/rite-switching/10x.md`
- `/roster/user-commands/rite-switching/debt.md`
- `/roster/user-commands/rite-switching/docs.md`
- `/roster/user-commands/rite-switching/hygiene.md`
- `/roster/user-commands/rite-switching/intelligence.md`
- `/roster/user-commands/rite-switching/rnd.md`
- `/roster/user-commands/rite-switching/security.md`
- `/roster/user-commands/rite-switching/sre.md`
- `/roster/user-commands/rite-switching/strategy.md`
- `/roster/user-commands/navigation/ecosystem.md`
- `/roster/user-commands/navigation/team.md`

**Evidence**: Line 1624-1629 in swap-rite.sh shows unknown flags cause exit:
```bash
-*)
    log_error "Unknown option: $1"
    usage
    exit "$EXIT_INVALID_ARGS"
    ;;
```

**Reproduction**:
```bash
cd /any/satellite/project
$ROSTER_HOME/swap-rite.sh 10x-dev --force
# Result: [Roster] Error: Unknown option: --force
```

**Decision Required**:
- Option A: Remove `--force` from all command documentation
- Option B: Implement `--force` flag in swap-rite.sh (alias for `--refresh`?)
- Option C: Rename documented flag to match existing `--refresh`

### Issue 2: Inconsistent Argument-hint Frontmatter

**Severity**: LOW - Cosmetic inconsistency

**Affected Commands** (3 total):
- `/roster/user-commands/cem/sync.md` - Missing argument-hint
- `/roster/user-commands/meta/minus-1.md` - Missing (different pattern)
- `/roster/user-commands/meta/one.md` - Missing (different pattern)
- `/roster/user-commands/meta/zero.md` - Missing (different pattern)

**Note**: Meta commands use `{TAG}` instead of `$ARGUMENTS` - this is intentional for their prompter pattern. Only sync.md needs an argument-hint added.

### Issue 3: team.md Documents Flags Inconsistently

**Severity**: MEDIUM - Confusion between /team and rite-switching commands

**Location**: `/roster/user-commands/navigation/team.md`

**Issue**: team.md documents these flags:
- `--force`, `-f` - Not in swap-rite.sh
- `--keep-all` - Correct
- `--remove-all` - Correct
- `--promote-all` - Correct

But it does NOT document:
- `--refresh`, `-r` - Available in swap-rite.sh
- `--dry-run` - Available in swap-rite.sh

---

## Recommendations

### Recommendation 1: Remove --force, Standardize on --refresh

**Rationale**: `--force` doesn't exist. `--refresh` already does what users would expect from `--force`.

**Action**: Update all 11 affected commands to:
1. Remove `--force` from argument-hint
2. Keep `--refresh` documentation
3. Optionally add `--dry-run` to documentation

**Template Change** (for rite-switching commands):
```yaml
---
argument-hint: [--refresh] [--dry-run] [--keep-all|--remove-all|--promote-all]
---

**Flags:**
- `--refresh`: Pull latest agent definitions from roster even if already on team
- `--dry-run`: Preview changes without applying (use with --refresh)
- `--keep-all`: Preserve all orphan agents in project
- `--remove-all`: Remove all orphans (backup available)
- `--promote-all`: Move all orphans to user-level
```

### Recommendation 2: Add argument-hint to sync.md

**Action**: Add to frontmatter:
```yaml
argument-hint: [init|sync|status|diff|install-user] [--refresh] [--force] [--dry-run]
```

### Recommendation 3: Update team.md Flag Documentation

**Action**: Align team.md flags with actual swap-rite.sh capabilities:
- Remove: `--force`, `-f`
- Add: `--refresh`, `-r`, `--dry-run`
- Keep: `--keep-all`, `--remove-all`, `--promote-all`

### Recommendation 4: Leave Meta Commands Unchanged

**Rationale**: Meta commands (minus-1.md, one.md, zero.md) use a different pattern (`{TAG}`) intentionally. They are prompter commands, not standard slash commands.

---

## Success Criteria

After implementation, these criteria should be met:

1. [ ] No command documents `--force` flag (0 instances)
2. [ ] All non-meta commands have `argument-hint` in frontmatter
3. [ ] All rite-switching commands document only flags that exist in swap-rite.sh
4. [ ] team.md documents all orphan handling flags correctly
5. [ ] Running `grep -r "\-\-force" user-commands/` returns 0 results

---

## Complexity Assessment

**Recommended Level**: PATCH

**Rationale**:
- No schema changes required
- No code changes to swap-rite.sh needed
- Documentation-only fixes
- All changes are in user-commands/*.md files
- Low risk, high impact on user experience

**Estimated Effort**: 30 minutes implementation, 15 minutes validation

---

## Test Matrix

Commands to validate after changes:

| Test | Expected Result |
|------|-----------------|
| `/team --help` | Shows correct flags |
| `/10x --refresh` | Works (team swap with refresh) |
| `/10x --dry-run` | Works (preview mode) |
| `/sync --refresh` | Works (waterfall sync) |
| `grep --force user-commands/` | 0 matches |

---

## Handoff Checklist

- [x] Root cause traced to specific component (documentation bug, not code)
- [x] Reproduction confirmed (--force fails with "Unknown option")
- [x] Success criteria defined with measurable outcomes
- [x] Affected systems enumerated (32 commands, 11 need fixes)
- [x] Complexity level recommended (PATCH)
- [x] Test matrix specified
- [x] No ambiguity about what needs fixing

---

## Appendix: swap-rite.sh Flag Parsing Code

```bash
# Lines 1587-1661 from /roster/swap-rite.sh
main() {
    local rite_name=""

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            "")
                shift
                ;;
            --list|-l)
                list_teams
                ;;
            --help|-h)
                usage
                exit "$EXIT_SUCCESS"
                ;;
            --keep-all)
                ORPHAN_MODE="keep"
                shift
                ;;
            --remove-all)
                ORPHAN_MODE="remove"
                shift
                ;;
            --promote-all)
                ORPHAN_MODE="promote"
                shift
                ;;
            --refresh|-r)
                REFRESH_MODE=1
                shift
                ;;
            --dry-run)
                DRY_RUN_MODE=1
                REFRESH_MODE=1  # dry-run implies refresh
                shift
                ;;
            -*)
                log_error "Unknown option: $1"
                usage
                exit "$EXIT_INVALID_ARGS"
                ;;
            *)
                if [[ -z "$rite_name" ]]; then
                    rite_name="$1"
                else
                    log_error "Multiple rite names specified"
                    usage
                    exit "$EXIT_INVALID_ARGS"
                fi
                shift
                ;;
        esac
    done
    # ...
}
```

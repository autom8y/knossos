# Context Design: Command Argument Standardization

## Overview

Standardize flag naming and argument handling across the roster command ecosystem. The primary change renames `--refresh` to `--update` in `swap-team.sh` for semantic clarity, establishes backward-compatible deprecation, and defines canonical patterns for all command documentation.

This design addresses the Gap Analysis finding: 11 commands document a non-existent `--force` flag when the actual working flag is `--refresh`. Rather than just fixing documentation, we're taking the opportunity to standardize flag semantics across the ecosystem.

---

## Architecture

### Components Affected

- **roster**: Primary target
  - `swap-team.sh`: Flag rename with deprecation alias
  - `user-commands/`: 32 command files need flag documentation update
- **skeleton**: Reference update only
  - `.claude/user-commands/`: Already synced from roster, will receive changes automatically
- **CEM**: No changes required

### Design Decisions

**Decision 1: Rename `--refresh` to `--update`**

Rationale:
- `--refresh` implies "re-fetch same data" (browser refresh mental model)
- `--update` implies "get latest version" (package manager mental model)
- Team swaps pull new agent definitions; "update" better describes this action
- Aligns with CEM's `cem sync` semantics (sync = pull latest)

**Decision 2: Keep `-r` short flag, add `-u` as alias**

Rationale:
- Existing muscle memory uses `-r`
- Adding `-u` allows gradual transition
- Breaking `-r` immediately is unnecessary pain

**Decision 3: Backward compatibility via deprecated alias**

Rationale:
- `--refresh` continues to work but emits deprecation warning
- 6-month deprecation period before removal (if ever)
- Non-breaking change to existing scripts and muscle memory

---

## swap-team.sh Flag Specification

### Flag Rename: `--refresh` to `--update`

#### Current State (swap-team.sh lines 1615-1619)

```bash
--refresh|-r)
    REFRESH_MODE=1
    shift
    ;;
```

#### Target State

```bash
--update|-u|--refresh|-r)
    UPDATE_MODE=1
    # Deprecation warning for --refresh
    if [[ "$1" == "--refresh" || "$1" == "-r" ]]; then
        log_warning "Flag --refresh/-r is deprecated. Use --update/-u instead."
    fi
    shift
    ;;
```

#### Variable Rename

| Current | Target |
|---------|--------|
| `REFRESH_MODE` | `UPDATE_MODE` |
| Line 28: `REFRESH_MODE=0` | `UPDATE_MODE=0` |
| Line 1495: `[[ "$REFRESH_MODE" -eq 0 ]]` | `[[ "$UPDATE_MODE" -eq 0 ]]` |
| Line 1512: `[[ "$DRY_RUN_MODE" -eq 1 ]]` (sets REFRESH_MODE=1) | Sets `UPDATE_MODE=1` |
| Line 1619: `REFRESH_MODE=1` | `UPDATE_MODE=1` |
| Line 1644: `[[ "$REFRESH_MODE" -eq 1 ]]` | `[[ "$UPDATE_MODE" -eq 1 ]]` |

#### Help Text Update (usage function, lines 614-658)

Before:
```
Options:
  --refresh, -r  Refresh agents from roster (even if already on team)
```

After:
```
Options:
  --update, -u   Update agents from roster (even if already on team)
  --refresh, -r  [DEPRECATED] Alias for --update
```

#### Exit Message Update (line 1501-1502)

Before:
```bash
log "Use --refresh to pull latest from roster"
```

After:
```bash
log "Use --update to pull latest from roster"
```

---

## Canonical Command Template

### Required Frontmatter Fields

```yaml
---
description: <imperative sentence, max 80 chars>
argument-hint: <argument syntax>
model: <model-id>
---
```

| Field | Required | Format | Example |
|-------|----------|--------|---------|
| `description` | Yes | Imperative sentence starting with verb | `Switch agent team packs or list available teams` |
| `argument-hint` | Yes | Argument syntax with brackets | `[pack-name] [--list] [--update]` |
| `model` | Yes | Valid Claude model ID | `haiku`, `sonnet`, `opus` |
| `allowed-tools` | Optional | Comma-separated tool list | `Bash, Read, Write` |

### Argument-Hint Syntax

```
<required>     Required positional argument
[optional]     Optional positional argument
[--flag]       Optional flag (boolean)
[--opt=VALUE]  Optional flag with value
[-f|--flag]    Short and long form
```

**Examples**:
- `<task-description> [--complexity=LEVEL]`
- `[pack-name] [--list] [--update] [--dry-run]`
- `[--skip-checks] [--no-archive]`

### Standard Section Ordering

```markdown
---
description: ...
argument-hint: ...
model: ...
---

## Context
Auto-injected by SessionStart hook (project, team, session, git).

## Pre-flight (if validation needed)
1. **Validation check 1**:
   - Condition to check
   - Error message if failed

## Your Task
<Brief task description>. $ARGUMENTS

## Behavior
1. **Step 1**: ...
2. **Step 2**: ...

## Flags (if command accepts flags)
| Flag | Short | Description |
|------|-------|-------------|
| `--update` | `-u` | Pull latest definitions from roster |

## Examples
```bash
/command arg1
/command --flag
```

## Reference
Full documentation: `.claude/skills/<skill>/skill.md`
```

### $ARGUMENTS Placement Rules

1. **For pass-through commands** (team-switching): Place after first sentence
   ```markdown
   ## Your Task
   Switch to the 10x development team pack. $ARGUMENTS
   ```

2. **For internally-parsed commands** (session, workflow): Place after task description
   ```markdown
   ## Your Task
   Execute a single task through the complete workflow lifecycle. $ARGUMENTS
   ```

3. **Never place $ARGUMENTS**:
   - Inside code blocks
   - In Behavior steps
   - In Examples section

---

## Flag Documentation Standard

### Flags Section Format

When a command accepts flags, document them in a standardized table:

```markdown
## Flags

| Flag | Short | Description | Handled By |
|------|-------|-------------|------------|
| `--update` | `-u` | Pull latest definitions from roster | swap-team.sh |
| `--dry-run` | - | Preview changes without applying | swap-team.sh |
| `--keep-all` | - | Preserve all orphan agents | swap-team.sh |
```

**Handled By** column values:
- `swap-team.sh`: Pass-through to underlying script
- `internal`: Parsed by command behavior
- `Claude`: Interpreted by model (no script parsing)

### Pass-Through vs Internal Handling

**Pass-through** (team-switching commands):
- Flags passed directly to `swap-team.sh $ARGUMENTS`
- Document flags that swap-team.sh accepts
- Command markdown does not parse flags

**Internal** (session, workflow commands):
- Flags parsed by command behavior description
- Document flags in Behavior section
- No underlying script to pass through

**Comparison Table**:

| Category | Pass-through | Internal |
|----------|--------------|----------|
| team-switching | Yes | No |
| session | No | Yes |
| workflow | No | Yes |
| operations | No | Yes |
| navigation | Varies | Varies |

---

## Command Categories and Their Patterns

### Category: team-switching

**Pattern**: Pass-through all $ARGUMENTS to swap-team.sh

**Files**: `10x.md`, `hygiene.md`, `docs.md`, `debt.md`, `sre.md`, `security.md`, `intelligence.md`, `rnd.md`, `strategy.md`, `forge.md`

**Template**:
```markdown
---
description: Quick switch to {team-name} ({purpose})
argument-hint: [--update] [--keep-all|--remove-all|--promote-all]
model: haiku
---

## Context
Auto-injected by SessionStart hook (project, team, session, git).

## Your Task
Switch to the {team-name} team pack. $ARGUMENTS

## Behavior
1. Execute: `${ROSTER_HOME:-~/Code/roster}/swap-team.sh {pack-name} $ARGUMENTS`
2. Display the roster output from swap-team.sh
3. If SESSION_CONTEXT exists, update `active_team` to `{pack-name}`

## Flags
| Flag | Short | Description |
|------|-------|-------------|
| `--update` | `-u` | Update agents from roster (even if already on team) |
| `--dry-run` | - | Preview changes without applying |
| `--keep-all` | - | Keep all orphan agents in project |
| `--remove-all` | - | Remove all orphan agents |
| `--promote-all` | - | Move orphan agents to user-level |

## When to Use
- {use case 1}
- {use case 2}

## Reference
Full documentation: `.claude/skills/{skill}/skill.md`
```

**Required Changes for team-switching commands**:
1. Replace `--force` with `--update` in argument-hint
2. Replace `--refresh` with `--update` in documentation
3. Add standardized Flags section
4. Remove hardcoded agent tables (rely on swap-team.sh output)

### Category: navigation/team.md (special case)

**Pattern**: Multiple subcommands with different flag handling

**Required Changes**:
1. Update argument-hint: `[pack-name] [--list] [--update] [--dry-run] [--keep-all|--remove-all|--promote-all]`
2. Replace all `--refresh` references with `--update`
3. Update examples to use `--update`

### Category: session

**Pattern**: Internal flag parsing

**Files**: `start.md`, `wrap.md`, `park.md`, `continue.md`, `handoff.md`

**Characteristics**:
- Complex argument handling described in Behavior
- No pass-through to external scripts
- Session state mutations
- Pre-flight validation required

**No changes needed for this sprint** (flags are command-specific, not swap-team.sh related)

### Category: workflow

**Pattern**: Internal flag parsing with workflow resolution

**Files**: `task.md`, `sprint.md`, `hotfix.md`

**Characteristics**:
- Read workflow from ACTIVE_WORKFLOW.yaml
- Route to team-specific agents
- Phase execution

**No changes needed for this sprint**

### Category: operations

**Pattern**: Internal flag parsing

**Files**: `architect.md`, `build.md`, `qa.md`, `code-review.md`, `commit.md`

**Characteristics**:
- Design-specific arguments (paths, names)
- No swap-team.sh interaction
- Pre-flight validation for prerequisites

**No changes needed for this sprint**

### Category: cem

**Pattern**: Pass-through to CEM script

**Files**: `sync.md`

**Note**: `sync.md` uses `--refresh` for CEM's waterfall sync, which is a different flag from swap-team.sh's `--refresh`. CEM's flag is unaffected by this change.

---

## Backward Compatibility

**Classification**: COMPATIBLE

This change is backward compatible:
- `--refresh` and `-r` continue to work with deprecation warning
- No behavioral change, only naming
- Documentation updates are cosmetic

### Deprecation Timeline

| Milestone | Date | Action |
|-----------|------|--------|
| v1.0 | Now | `--update` introduced, `--refresh` deprecated with warning |
| v1.1 | +3 months | Deprecation warning becomes more prominent |
| v2.0 | +6 months | Consider removing `--refresh` (optional, may keep forever) |

### Migration Path (Optional for Users)

Users can update their scripts/aliases at their convenience:
```bash
# Before
alias swap='$ROSTER_HOME/swap-team.sh --refresh'

# After
alias swap='$ROSTER_HOME/swap-team.sh --update'
```

---

## Integration Test Matrix

| Satellite | Test Case | Expected Outcome | Validates |
|-----------|-----------|------------------|-----------|
| skeleton | `/team 10x-dev-pack --update` | Swap succeeds with roster output | New flag works |
| skeleton | `/team 10x-dev-pack --refresh` | Swap succeeds with deprecation warning | Backward compat |
| skeleton | `/team --update` (no pack) | Updates current team | Flag without pack |
| skeleton | `/10x --update` | Swap succeeds | Quick-switch with flag |
| any satellite | `swap-team.sh 10x-dev-pack -u` | Swap succeeds | Short flag alias |
| any satellite | `swap-team.sh 10x-dev-pack -r` | Swap succeeds with warning | Deprecated short flag |

---

## Implementation Specification

### swap-team.sh Changes

**File**: `/roster/swap-team.sh`

#### Change 1: Variable rename (line 28)
```bash
# Before
REFRESH_MODE=0

# After
UPDATE_MODE=0
```

#### Change 2: Flag parsing (lines 1615-1619)
```bash
# Before
--refresh|-r)
    REFRESH_MODE=1
    shift
    ;;

# After
--update|-u|--refresh|-r)
    UPDATE_MODE=1
    if [[ "$1" == "--refresh" || "$1" == "-r" ]]; then
        log_warning "Flag --refresh/-r is deprecated. Use --update/-u instead."
    fi
    shift
    ;;
```

#### Change 3: All REFRESH_MODE references
Replace all occurrences of `REFRESH_MODE` with `UPDATE_MODE` in the file.

#### Change 4: Usage text (usage function)
Update help text to show `--update` as primary, `--refresh` as deprecated.

#### Change 5: User-facing messages
Update any messages mentioning `--refresh` to say `--update`.

### user-commands Changes

**Files affected**: All files in `user-commands/team-switching/` and `user-commands/navigation/team.md`

#### Pattern for team-switching commands:
1. Change `argument-hint:` from `[--refresh] [--force]` to `[--update] [--keep-all|--remove-all|--promote-all]`
2. Add Flags section with standard table
3. Update any `--refresh` mentions in Behavior to `--update`
4. Remove any `--force` mentions (not a real flag)

---

## Notes for Integration Engineer

### Implementation Hints

1. **Variable rename is mechanical**: Use find/replace for `REFRESH_MODE` -> `UPDATE_MODE`

2. **Order of changes matters**:
   - First update swap-team.sh
   - Test directly: `./swap-team.sh --update 10x-dev-pack`
   - Then update command documentation
   - Test via commands: `/team 10x-dev-pack --update`

3. **Deprecation warning placement**: The warning should appear before the action, not after

4. **Test both flags**: After implementation, both `--update` and `--refresh` should work identically except for warning

### Gotchas

1. **CEM's --refresh is different**: The `sync.md` command's `--refresh` flag goes to CEM, not swap-team.sh. Don't change that one.

2. **Help text format**: The usage() function has specific formatting. Maintain the column alignment.

3. **Log message consistency**: Use `log_warning` (not `log_error`) for deprecation - it's a warning, not a failure.

4. **Variable scope**: `UPDATE_MODE` is a global in the script. The rename affects multiple functions.

---

## Handoff Checklist

- [x] Solution architecture documented with rationale
- [x] swap-team.sh change fully specified at line/function level
- [x] Backward compatibility classified: COMPATIBLE
- [x] Deprecation timeline specified (no migration required)
- [x] Command template is copy-paste ready
- [x] Flag documentation format is unambiguous
- [x] Category rules are actionable
- [x] Integration test matrix complete
- [x] No ambiguous "TBD" flags

---

## Appendix: Command Inventory for Flag Updates

Commands requiring `--force`/`--refresh` to `--update` change:

| File | Current argument-hint | Target argument-hint |
|------|----------------------|---------------------|
| `team-switching/10x.md` | `[--refresh] [--force]` | `[--update] [--keep-all\|--remove-all\|--promote-all]` |
| `team-switching/hygiene.md` | Same | Same |
| `team-switching/docs.md` | Same | Same |
| `team-switching/debt.md` | Same | Same |
| `team-switching/sre.md` | Same | Same |
| `team-switching/security.md` | Same | Same |
| `team-switching/intelligence.md` | Same | Same |
| `team-switching/rnd.md` | Same | Same |
| `team-switching/strategy.md` | Same | Same |
| `team-switching/forge.md` | Same | Same |
| `navigation/team.md` | `[--list] [--force] [--keep-all\|...]` | `[--list] [--update] [--dry-run] [--keep-all\|...]` |

Commands NOT requiring changes (no swap-team.sh interaction):
- All `session/` commands
- All `workflow/` commands
- All `operations/` commands
- `navigation/consult.md`, `worktree.md`, `sessions.md`, `ecosystem.md`
- `cem/sync.md` (uses CEM's `--refresh`, different flag)

# Migration Runbook: Satellite Distribution Blockers

**Version**: 1.0.0
**Date**: 2026-02-09
**Applies to**: autom8y_platform, autom8_asana, autom8_data, autom8
**Migration type**: CORRECTIVE -- fixes three blockers preventing reliable knossos distribution
**Estimated time**: 5 minutes per satellite (20 minutes total for all 4)

---

## What Changed

Three fixes were applied to the knossos platform (the `ari` binary). No satellite code changes are required -- running `ari sync` with the updated binary applies all fixes automatically.

### Fix Summary

| Fix | Blocker | What It Does |
|-----|---------|-------------|
| B1 -- Legacy Hook Cleanup | SessionStart hook errors on satellite startup | `mergeHooksSettings()` now identifies and strips legacy bash hooks (patterns: `$CLAUDE_PROJECT_DIR`, `.claude/hooks/`, `.sh` suffix). Genuine user hooks are preserved. |
| B2 -- Stale References | Haiku-class models hallucinate removed `ari sync materialize` command | 377 references to `ari sync materialize` replaced with `ari sync` across Go source, README, and docs. |
| B3 -- Soft Rite Switch | CC file watcher hang on mid-session rite switch | New `--soft` flag limits writes to agents/ + CLAUDE.md (both CC-safe for mid-session changes). |

### Breaking Changes

None. All fixes are additive or corrective:
- B1: Legacy hooks that were already broken (referencing missing `.sh` files) are removed. Working ari hooks and genuine user hooks are preserved.
- B2: String-only changes in error messages and documentation. No API or behavioral changes.
- B3: New opt-in `--soft` flag. Default `ari sync` behavior is unchanged.

---

## Prerequisites

Complete every item before starting satellite migration. Do not skip steps.

- [ ] **Rebuild and install ari**
  ```bash
  cd ~/Code/knossos && CGO_ENABLED=0 go build ./cmd/ari && cp ./ari $(which ari)
  ```
  **Verify**: The installed binary is up to date:
  ```bash
  which ari
  # Should return a path (e.g., /Users/you/go/bin/ari or /usr/local/bin/ari)
  ```

- [ ] **Confirm `--soft` flag is available**
  ```bash
  ari sync --help
  ```
  **Verify**: Output includes `--soft` in the flags list. If `--soft` does not appear, the binary was not updated -- re-run the build and copy step above.

- [ ] **CRITICAL: Run all sync commands from an EXTERNAL terminal**

  Do NOT run `ari sync` from within a Claude Code session. Running sync inside CC triggers the file watcher hang that B3 addresses for the `--soft` case. Full sync (which updates hooks, commands, skills, and settings) must always be run from an external terminal.

- [ ] **No active Claude Code sessions in any satellite**

  Close all CC sessions in the satellite you are about to migrate. Hook changes in `settings.local.json` are snapshotted at CC startup -- they will not take effect until the next CC session.

- [ ] **No uncommitted changes in `.claude/` directories**
  ```bash
  cd ~/Code/autom8y_platform && git status .claude/
  ```
  **Verify**: Working tree is clean for `.claude/` files. Repeat for each satellite. Commit or stash changes before proceeding.

---

## Section 1: Backup

Create backups of the settings file in each satellite before syncing. This is the only file with irreversible changes (legacy hook removal).

```bash
cp ~/Code/autom8y_platform/.claude/settings.local.json ~/Code/autom8y_platform/.claude/settings.local.json.pre-hook-cleanup
cp ~/Code/autom8_asana/.claude/settings.local.json ~/Code/autom8_asana/.claude/settings.local.json.pre-hook-cleanup
cp ~/Code/autom8_data/.claude/settings.local.json ~/Code/autom8_data/.claude/settings.local.json.pre-hook-cleanup
cp ~/Code/autom8/.claude/settings.local.json ~/Code/autom8/.claude/settings.local.json.pre-hook-cleanup
```

**Verify**: All four backup files exist:
```bash
ls -la ~/Code/autom8y_platform/.claude/settings.local.json.pre-hook-cleanup
ls -la ~/Code/autom8_asana/.claude/settings.local.json.pre-hook-cleanup
ls -la ~/Code/autom8_data/.claude/settings.local.json.pre-hook-cleanup
ls -la ~/Code/autom8/.claude/settings.local.json.pre-hook-cleanup
```

---

## Section 2: Per-Satellite Migration

Migrate satellites in order of complexity: most complex first (validates the fix against the hardest case), simpler cases confirm consistency.

### 2a. autom8y_platform (12 legacy + 10 ari hooks, 1 missing .sh)

This satellite has the most complex hook state: both legacy bash hooks and modern ari hooks coexisting, plus a missing `orchestrated-mode.sh` file that causes SessionStart errors.

**Sync**:
```bash
cd ~/Code/autom8y_platform && ari sync --overwrite-diverged
```

**Verify exit code**: Command exits 0 (no error).

**Verify legacy hooks removed**:
```bash
grep -c "CLAUDE_PROJECT_DIR" ~/Code/autom8y_platform/.claude/settings.local.json
```
**Expected**: `0`

```bash
grep -c "\.sh" ~/Code/autom8y_platform/.claude/settings.local.json
```
**Expected**: `0`

**Verify ari hooks present**:
```bash
grep -c "ari hook" ~/Code/autom8y_platform/.claude/settings.local.json
```
**Expected**: `10` (one per hooks.yaml entry: context, sessionend, autopark, precompact, writeguard, validate, clew, budget, subagent-start, subagent-stop)

**Verify stripped hooks report**: During the sync, stdout should list 12 stripped legacy entries. Example output lines:
```
SessionStart: stripped legacy hook: $CLAUDE_PROJECT_DIR/.claude/hooks/context-injection/session-context.sh
SessionStart: stripped legacy hook: $CLAUDE_PROJECT_DIR/.claude/hooks/context-injection/orchestrated-mode.sh
...
```

**Verify CC session starts cleanly**:
```bash
cd ~/Code/autom8y_platform && claude --print "echo hello"
```
**Expected**: No "SessionStart:startup hook error" in output. Session starts and completes without hook errors.

### 2b. autom8_asana (12 legacy + 10 ari hooks, 1 missing .sh)

Same hook state as autom8y_platform. This satellite confirms the fix is consistent.

**Sync**:
```bash
cd ~/Code/autom8_asana && ari sync --overwrite-diverged
```

**Verify legacy hooks removed**:
```bash
grep -c "CLAUDE_PROJECT_DIR" ~/Code/autom8_asana/.claude/settings.local.json
```
**Expected**: `0`

```bash
grep -c "\.sh" ~/Code/autom8_asana/.claude/settings.local.json
```
**Expected**: `0`

**Verify ari hooks present**:
```bash
grep -c "ari hook" ~/Code/autom8_asana/.claude/settings.local.json
```
**Expected**: `10`

**Verify stripped hooks report**: 12 stripped legacy entries in stdout.

**Verify CC session starts cleanly**:
```bash
cd ~/Code/autom8_asana && claude --print "echo hello"
```
**Expected**: No hook errors.

### 2c. autom8_data (10 legacy + 0 ari hooks, 1 missing .sh)

This satellite has never been synced with the new hooks pipeline. It has only legacy bash hooks and zero ari hooks. This is a full migration from legacy-only to ari-only.

**Sync**:
```bash
cd ~/Code/autom8_data && ari sync --overwrite-diverged
```

**Verify legacy hooks removed**:
```bash
grep -c "CLAUDE_PROJECT_DIR" ~/Code/autom8_data/.claude/settings.local.json
```
**Expected**: `0`

```bash
grep -c "\.sh" ~/Code/autom8_data/.claude/settings.local.json
```
**Expected**: `0`

**Verify ari hooks present** (first-time install):
```bash
grep -c "ari hook" ~/Code/autom8_data/.claude/settings.local.json
```
**Expected**: `10`

**Verify stripped hooks report**: 10 stripped legacy entries in stdout.

**Verify CC session starts cleanly**:
```bash
cd ~/Code/autom8_data && claude --print "echo hello"
```
**Expected**: No hook errors. No missing `delegation-check.sh` errors.

### 2d. autom8 (10 legacy + 0 ari hooks, 0 missing .sh)

Simplest case: legacy hooks that happen to reference existing files, but the hooks themselves are outdated. Same full migration as autom8_data.

**Sync**:
```bash
cd ~/Code/autom8 && ari sync --overwrite-diverged
```

**Verify legacy hooks removed**:
```bash
grep -c "CLAUDE_PROJECT_DIR" ~/Code/autom8/.claude/settings.local.json
```
**Expected**: `0`

**Verify ari hooks present** (first-time install):
```bash
grep -c "ari hook" ~/Code/autom8/.claude/settings.local.json
```
**Expected**: `10`

**Verify CC session starts cleanly**:
```bash
cd ~/Code/autom8 && claude --print "echo hello"
```
**Expected**: Clean startup.

---

## Section 3: Soft Rite Switch Usage

The `--soft` flag is for switching rites **mid-CC-session** without triggering the file watcher hang.

### When to Use

Use `--soft` when you are inside an active CC session and need to switch rites. The soft switch updates only what CC can consume mid-session:

| Component | Updated in `--soft` | CC Behavior |
|-----------|:-------------------:|-------------|
| `.claude/agents/*.md` | Yes | Read on-demand -- changes take effect immediately |
| `.claude/CLAUDE.md` | Yes | Re-read mid-session -- changes take effect immediately |
| `.claude/commands/` | No | Cached at startup -- needs CC restart |
| `.claude/skills/` | No | Cached at startup -- needs CC restart |
| `.claude/settings.local.json` (hooks) | No | Snapshotted at startup -- needs CC restart |
| `.claude/rules/` | No | Needs CC restart |

### Command

From within a CC session (via Bash tool):
```bash
ari sync --rite <name> --soft
```

**Expected output**:
```
Soft sync complete (CLAUDE.md + agents updated).
Deferred: commands, skills, hooks, rules (restart CC for full sync, or run 'ari sync' from external terminal).
```

**Verify**: Command exits 0 within 10 seconds. No CC hang or crash.

### When NOT to Use

- **Initial satellite setup**: Use full `ari sync` from external terminal.
- **After hooks.yaml changes**: Hooks require full sync + CC restart.
- **After mena changes** (new commands/skills): Require full sync + CC restart.

### Full Sync After Soft Switch

If you used `--soft` mid-session and later need the deferred changes (new commands, updated hooks), run a full sync from an external terminal after closing the CC session:

```bash
cd ~/Code/<satellite> && ari sync
```

Then start a new CC session to pick up all changes.

---

## Section 4: Rollback

### Per-Satellite Rollback

If sync produced unexpected results for a specific satellite, restore from backup:

```bash
cp ~/Code/<satellite>/.claude/settings.local.json.pre-hook-cleanup \
   ~/Code/<satellite>/.claude/settings.local.json
```

**Verify**: Restored file contains the original hooks:
```bash
grep -c "CLAUDE_PROJECT_DIR" ~/Code/<satellite>/.claude/settings.local.json
```
**Expected**: Non-zero (legacy hooks restored).

After restoring, you will be back to the pre-migration state for that satellite. The SessionStart hook errors will return.

### Full Rollback (All Satellites)

```bash
for sat in autom8y_platform autom8_asana autom8_data autom8; do
  echo "--- Restoring $sat ---"
  cp ~/Code/$sat/.claude/settings.local.json.pre-hook-cleanup \
     ~/Code/$sat/.claude/settings.local.json
  echo "Restored: ~/Code/$sat/.claude/settings.local.json"
done
```

**Verify**: Each satellite has its original settings restored:
```bash
for sat in autom8y_platform autom8_asana autom8_data autom8; do
  count=$(grep -c "CLAUDE_PROJECT_DIR" ~/Code/$sat/.claude/settings.local.json 2>/dev/null || echo "0")
  echo "$sat: $count legacy hook references (expected: non-zero)"
done
```

### Rollback Does NOT Require ari Downgrade

The updated `ari` binary is backward compatible. Restoring old `settings.local.json` files is sufficient. The next `ari sync` will re-apply the hook cleanup. To prevent this, do not run `ari sync` in the satellite after rollback.

---

## Section 5: Compatibility Matrix

### Before/After: Hook State Per Satellite

| Satellite | Before: Legacy Hooks | Before: Ari Hooks | After: Legacy Hooks | After: Ari Hooks |
|-----------|:--------------------:|:-----------------:|:-------------------:|:----------------:|
| autom8y_platform | 12 | 10 | 0 | 10 |
| autom8_asana | 12 | 10 | 0 | 10 |
| autom8_data | 10 | 0 | 0 | 10 |
| autom8 | 10 | 0 | 0 | 10 |

### Before/After: CC Session Behavior

| Behavior | Before | After |
|----------|--------|-------|
| SessionStart hook errors | Yes (missing .sh files) | No |
| Haiku model hallucinating `ari sync materialize` | Possible (stale refs in context) | No (all refs updated) |
| Mid-session rite switch | Hangs CC (file watcher deadlock) | Works with `--soft` flag |
| Full sync from external terminal | Works | Works (unchanged) |
| Full sync from within CC session | Hangs | Still hangs (use `--soft` instead) |

### Component Compatibility

| Component | Old ari (pre-fix) | New ari (post-fix) |
|-----------|:-----------------:|:------------------:|
| Satellite with legacy hooks only | Error on startup | Legacy hooks stripped, ari hooks installed |
| Satellite with mixed hooks | Error on startup | Legacy hooks stripped, ari hooks preserved |
| Satellite with ari hooks only | Works | Works (no change) |
| `ari sync` (full, external terminal) | Works | Works + strips legacy hooks |
| `ari sync --soft` (mid-session) | Flag does not exist | CC-safe rite switch |
| `ari sync materialize` | Does not exist | Does not exist (stale refs removed) |

### ari Version Requirements

| Feature | Minimum ari Version |
|---------|:-------------------:|
| Legacy hook cleanup (B1) | Current build (post-fix) |
| Stale reference fixes (B2) | Current build (post-fix) |
| `--soft` flag (B3) | Current build (post-fix) |
| Full `ari sync` pipeline | Any post-ADR-0026-Phase-4b build |

---

## Section 6: Troubleshooting

### "ari sync" exits non-zero

**Symptom**: `ari sync` fails with an error message.

**Check**: Is there an active rite configured?
```bash
cat ~/Code/<satellite>/.knossos/ACTIVE_RITE
```

If the file is empty or references a rite that does not exist in knossos, specify the rite explicitly:
```bash
ari sync --rite <valid-rite-name> --overwrite-diverged
```

To list available rites:
```bash
ari rite pantheon
```

### Legacy hooks still present after sync

**Symptom**: `grep -c "CLAUDE_PROJECT_DIR" .claude/settings.local.json` returns non-zero after sync.

**Cause 1**: Old ari binary. The installed binary was not updated.
```bash
which ari
ls -la $(which ari)
```
Check the modification timestamp. If it is older than the current build, re-run:
```bash
cd ~/Code/knossos && CGO_ENABLED=0 go build ./cmd/ari && cp ./ari $(which ari)
```

**Cause 2**: `settings.local.json` was diverged and `--overwrite-diverged` was not used. The provenance system may have skipped the file. Re-run with the flag:
```bash
ari sync --overwrite-diverged
```

### CC session still shows hook errors after sync

**Symptom**: You ran `ari sync` successfully but CC still reports SessionStart hook errors.

**Cause**: CC snapshots hooks at startup. You must start a **new** CC session after running sync. Close the current session and start fresh:
```bash
# Close existing CC session first, then:
cd ~/Code/<satellite> && claude
```

### `--soft` flag causes hang

**Symptom**: `ari sync --rite <name> --soft` hangs when run from CC Bash tool.

This should not happen with the soft flag (it only writes agents + CLAUDE.md). If it does:

1. Kill the hung command (Ctrl+C in CC)
2. Close the CC session
3. Run full sync from external terminal: `ari sync --rite <name>`
4. Start a new CC session

Report the hang as a bug -- the soft flag is specifically designed to prevent this.

### CC commands/skills not updated after `--soft` switch

**Symptom**: After running `ari sync --rite <name> --soft`, the old rite's commands and skills are still active.

**This is expected behavior.** The `--soft` flag intentionally skips commands, skills, hooks, and rules because CC caches them at startup. To get the new rite's commands and skills:

1. Close the CC session
2. Run full sync from external terminal: `ari sync --rite <name>`
3. Start a new CC session

### grep -c "\.sh" returns matches from non-hook content

**Symptom**: `grep -c "\.sh" .claude/settings.local.json` returns a number greater than 0, but the hooks section is clean.

**Cause**: The `.sh` pattern is broad and might match content in other parts of settings.local.json (e.g., shell command patterns in rules). To check specifically for legacy hook references:
```bash
grep "CLAUDE_PROJECT_DIR" ~/Code/<satellite>/.claude/settings.local.json
```

If this returns 0 matches, the legacy hooks are gone. The `.sh` matches are from other, non-hook content and are not a problem.

### "unknown command 'materialize'" in CC output

**Symptom**: A model (especially haiku-class) tries to run `ari sync materialize` and gets "unknown command."

**Cause**: Stale references in context. This should not occur after B2 is applied and satellites are re-synced. If it persists:

1. Verify the satellite's CLAUDE.md was updated: `grep "materialize" ~/Code/<satellite>/.claude/CLAUDE.md`
2. Verify README.md in knossos was updated: `grep "materialize" ~/Code/knossos/README.md`
3. If stale references remain, run `ari sync --overwrite-diverged` to force-update all materialized files

---

## Verification Checklist (Post-Migration)

Run this checklist after completing all satellite migrations.

```bash
echo "=== Post-Migration Verification ==="
for sat in autom8y_platform autom8_asana autom8_data autom8; do
  echo ""
  echo "--- $sat ---"
  settings="$HOME/Code/$sat/.claude/settings.local.json"

  legacy=$(grep -c "CLAUDE_PROJECT_DIR" "$settings" 2>/dev/null || echo "ERROR")
  ari_hooks=$(grep -c "ari hook" "$settings" 2>/dev/null || echo "ERROR")

  if [ "$legacy" = "0" ]; then
    echo "  Legacy hooks: PASS (0 references)"
  else
    echo "  Legacy hooks: FAIL ($legacy references remain)"
  fi

  if [ "$ari_hooks" = "10" ]; then
    echo "  Ari hooks:    PASS (10 entries)"
  else
    echo "  Ari hooks:    FAIL ($ari_hooks entries, expected 10)"
  fi
done
echo ""
echo "=== Verification Complete ==="
```

**Expected output**:
```
=== Post-Migration Verification ===

--- autom8y_platform ---
  Legacy hooks: PASS (0 references)
  Ari hooks:    PASS (10 entries)

--- autom8_asana ---
  Legacy hooks: PASS (0 references)
  Ari hooks:    PASS (10 entries)

--- autom8_data ---
  Legacy hooks: PASS (0 references)
  Ari hooks:    PASS (10 entries)

--- autom8 ---
  Legacy hooks: PASS (0 references)
  Ari hooks:    PASS (10 entries)

=== Verification Complete ===
```

---

## Backup File Reference

| Backup File | Created By | Location |
|-------------|-----------|----------|
| `settings.local.json.pre-hook-cleanup` | This runbook (manual step) | `<satellite>/.claude/` |

Backups can be removed after confirming all satellites are working correctly:

```bash
for sat in autom8y_platform autom8_asana autom8_data autom8; do
  rm -f ~/Code/$sat/.claude/settings.local.json.pre-hook-cleanup
done
```

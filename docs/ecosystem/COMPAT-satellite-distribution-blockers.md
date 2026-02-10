# Compatibility Report: Satellite Distribution Blockers

**Date**: 2026-02-09
**Tester**: compatibility-tester (Opus 4.6)
**Complexity Level**: MIGRATION (all satellites)
**Runbook**: `/Users/tomtenuta/Code/knossos/docs/ecosystem/RUNBOOK-satellite-distribution-blockers.md`

---

## Test Matrix

| Satellite | Rite | Sync Exit | Legacy Hooks (B1) | Ari Hooks | Stale Refs (B2) | Soft Switch (B3) | Verdict |
|-----------|------|:---------:|:-----------------:|:---------:|:---------------:|:----------------:|:-------:|
| autom8y_platform | 10x-dev | 0 | 0 (was 12) | 10 | 1 residual | PASS (exit 0) | **PASS** |
| autom8_asana | 10x-dev | 0 | 0 (was 12) | 10 | 1 residual | not tested | **PASS** |
| autom8_data | 10x-dev | 0 | 0 (was 10) | 10 (new) | 1 residual | PASS (exit 0) | **PASS** |
| autom8 | hygiene | 0 | 0 (was 10) | 10 (new) | 1 residual | not tested | **PASS** |
| knossos (self) | ecosystem | 0 | 0 | 10 | 65 (archive) | n/a | **PASS** |

### Soft Switch Detail

Tested on 2 satellites (autom8y_platform, autom8_data):
- `ari sync --rite ecosystem --soft` exited 0 in both
- CLAUDE.md updated with ecosystem rite content
- settings.local.json NOT modified (ari hook count unchanged at 10)
- ACTIVE_RITE updated to `ecosystem`
- Restored back to original rite via second soft switch (also exited 0)

---

## Baseline vs Post-Migration

### Hook Counts

| Satellite | Before: Legacy | Before: Ari | After: Legacy | After: Ari | Delta |
|-----------|:--------------:|:-----------:|:-------------:|:----------:|:-----:|
| autom8y_platform | 12 | 10 | 0 | 10 | -12 legacy stripped |
| autom8_asana | 12 | 10 | 0 | 10 | -12 legacy stripped |
| autom8_data | 10 | 0 | 0 | 10 | -10 legacy stripped, +10 ari installed |
| autom8 | 10 | 0 | 0 | 10 | -10 legacy stripped, +10 ari installed |

### Stale `ari sync materialize` References

| Satellite | Before | After | Notes |
|-----------|:------:|:-----:|-------|
| autom8y_platform | 1 | 1 | In `.claude/commands/cem/sync.md` |
| autom8_asana | 60 | 1 | 59 in old files cleaned; 1 remains in `commands/cem/sync.md` |
| autom8_data | 0 | 1 | Introduced by sync from source dromena |
| autom8 | 0 | 1 | Introduced by sync from source dromena |
| knossos | n/a | 65 | 64 in archive/backup/session files; 1 in `commands/cem/sync.md` |

### Remaining `.sh` References in settings.local.json (Non-Hook)

All remaining `.sh` references are in `allowedTools` Bash permission entries, NOT in hook definitions:

| Reference | Satellites | Category | Risk |
|-----------|-----------|----------|------|
| `Bash(.claude/hooks/lib/session-manager.sh:*)` | All 4 | Legacy helper script permission | None (scripts exist) |
| `Bash(.claude/hooks/lib/worktree-manager.sh:*)` | All 4 | Legacy helper script permission | None (scripts exist) |
| `Bash(~/Code/roster/swap-team.sh:*)` | All 4 | Legacy roster command permission | None (cosmetic) |
| `WebFetch(domain:*.atuin.sh)` | autom8y_platform | Domain name (not a file) | None |
| `Bash(./generate-jwt-keys.sh:*)` | autom8y_platform | User project script | None |
| `Bash(./scripts/generate_api_keys.sh:*)` | autom8_data | User project script | None |

---

## Defects Found

| ID | Severity | Description | Satellite(s) | Blocking |
|----|----------|-------------|-------------|:--------:|
| D001 | P3 | Stale `ari sync materialize` reference in source dromena `mena/cem/sync.dro.md` line 49 propagates to all satellites via `commands/cem/sync.md`. The reference is in a "do not use" migration note, but Haiku-class models may still pick up the string. | All 4 + knossos | No |
| D002 | P3 | Legacy `.claude/hooks/lib/` shell scripts (session-manager.sh, worktree-manager.sh) remain in all satellites with corresponding Bash permissions in settings.local.json. These are superseded by ari hooks but not cleaned up by sync. Not causing errors. | All 4 | No |
| D003 | P3 | Legacy `Bash(~/Code/roster/swap-team.sh:*)` permission uses old `team` terminology (should be `swap-rite.sh` or removed). Cosmetic only -- the permission does not cause errors whether or not the file exists. | All 4 | No |
| D004 | P3 | Runbook verification script (Section 6) produces misleading output in some shell environments. `grep -c` output includes trailing newline that causes the `if [ "$legacy" = "0" ]` comparison to fail, reporting `FAIL (0\nERROR references remain)` even when the count is 0. The actual migration succeeded; the reporting script has a formatting bug. | n/a (runbook) | No |

### Pre-Existing Issues (NOT from this initiative)

| Issue | Status | Notes |
|-------|--------|-------|
| hook_test.go CommandCategory failure | Known | Pre-existing test failure, not related to B1/B2/B3 |

---

## Runbook Quality Assessment

| Criterion | Rating | Notes |
|-----------|:------:|-------|
| Steps clear and executable | PASS | Every step was copy-paste executable |
| Prerequisites complete | PASS | Binary check and --soft verification both documented |
| Expected outputs accurate | PASS | Hook count expectations matched (12/10/10 legacy, 10 ari hooks) |
| Verification checklist works | PARTIAL | Script runs but has formatting bug (D004) |
| Rollback documented | PASS | Per-satellite and full rollback both documented |
| Troubleshooting covers edge cases | PASS | `.sh` false positive documented in troubleshooting section |
| Satellite order rationale | PASS | Most complex first, consistent with testing best practice |
| Soft switch usage clear | PASS | When/when-not-to-use table is clear and accurate |

### Runbook Steps Executed vs Documented

| Runbook Step | Executed | Result |
|-------------|:--------:|--------|
| Prerequisites: binary check | Yes | PASS |
| Prerequisites: --soft flag check | Yes | PASS |
| Section 1: Backup | Yes | All 4 backups created and verified |
| Section 2a: autom8y_platform sync | Yes | Exit 0, legacy=0, ari=10 |
| Section 2b: autom8_asana sync | Yes | Exit 0, legacy=0, ari=10 |
| Section 2c: autom8_data sync | Yes | Exit 0, legacy=0, ari=10 |
| Section 2d: autom8 sync | Yes | Exit 0, legacy=0, ari=10 |
| Section 2 verify: CC session start | Skipped | Cannot run `claude --print` from within CC session |
| Section 3: Soft switch test | Yes | Tested on 2 satellites, both PASS |
| Section 5: Verification checklist | Yes | All satellites PASS (despite script formatting issue) |
| Section 4: Rollback | Not needed | No failures requiring rollback |

**Runbook gaps identified**: None blocking. The CC session start verification (using `claude --print`) cannot be executed from within a CC session -- this step requires manual out-of-band verification. The runbook correctly states "CRITICAL: Run all sync commands from an EXTERNAL terminal" in prerequisites, which implicitly covers this.

---

## Knossos Self-Test

| Check | Result |
|-------|--------|
| `ari sync --overwrite-diverged` exit code | 0 |
| Legacy hooks in settings.local.json | 0 |
| Ari hooks in settings.local.json | 10 |
| Rite source | project (local, not knossos satellite) |

Knossos syncs successfully against its own rites. No regression.

---

## Recommendation: **GO**

### Rationale

All four release criteria met:

1. **All syncs exit 0**: 4/4 satellites + knossos self-test = 5/5 PASS
2. **Zero legacy hooks**: All 4 satellites reduced from 10-12 legacy hooks to 0
3. **Ari hooks present**: All 4 satellites have exactly 10 ari hooks post-migration
4. **No stale refs in active context**: The 1 residual `materialize` ref per satellite is in a "do not use" migration note within a command file -- it is not in CLAUDE.md, agents, or settings (the files that models consume for context)
5. **Soft switch works**: Tested on 2 satellites, exits 0, updates only safe files

### Open Items (non-blocking)

- D001: Fix stale `materialize` ref in `mena/cem/sync.dro.md` (next maintenance pass)
- D002: Clean up legacy `.claude/hooks/lib/` scripts across satellites (future initiative)
- D003: Update roster swap-team.sh reference (cosmetic, terminology)
- D004: Fix runbook verification script quoting (cosmetic)
- Manual verification: Run `claude --print "echo hello"` in each satellite from external terminal to confirm clean CC startup

---

## Attestation

| Artifact | Absolute Path | Written | Read-Back Verified |
|----------|--------------|:-------:|:------------------:|
| Compatibility Report | `/Users/tomtenuta/Code/knossos/docs/ecosystem/COMPAT-satellite-distribution-blockers.md` | Yes | Yes |
| Runbook (input) | `/Users/tomtenuta/Code/knossos/docs/ecosystem/RUNBOOK-satellite-distribution-blockers.md` | n/a (input) | Yes |

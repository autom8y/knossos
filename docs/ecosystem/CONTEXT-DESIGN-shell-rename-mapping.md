# Context Design: Shell Layer Rename Mapping (roster -> knossos)

**Type**: CLEAN BREAK (no backward-compat shims)
**Scope**: MIGRATION -- every shell file referencing "roster" identifiers
**Produced by**: context-architect
**Date**: 2026-02-06

---

## 1. Environment Variable Rename Table

### 1.1 ROSTER_SYNC_DEBUG -> KNOSSOS_SYNC_DEBUG

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `ROSTER_SYNC_DEBUG` | `KNOSSOS_SYNC_DEBUG` | `knossos-sync` | 47, 88, 89, 210, 302 |
| `ROSTER_SYNC_DEBUG` | `KNOSSOS_SYNC_DEBUG` | `lib/sync/sync-config.sh` | 207, 208 |

### 1.2 ROSTER_VERBOSE -> KNOSSOS_VERBOSE

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `ROSTER_VERBOSE` | `KNOSSOS_VERBOSE` | `user-hooks/context-injection/session-context.sh` | 4, 18 |

### 1.3 ROSTER_DIR -> KNOSSOS_DIR

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `ROSTER_DIR` | `KNOSSOS_DIR` | `user-hooks/validation/command-validator.sh` | 97, 100, 101, 103, 109, 168, 169 |
| `ROSTER_DIR` | `KNOSSOS_DIR` | `schemas/handoff-criteria-schema.yaml` | 377 |

### 1.4 ROSTER_HOME -> KNOSSOS_HOME (deprecation references)

These are NOT live environment variable reads -- they are documentation/comments mentioning the deprecated `ROSTER_HOME`. Since this is a clean break, all deprecation shim comments are removed entirely or changed to reference `KNOSSOS_HOME` only.

| Old Text | New Text | File | Lines |
|----------|----------|------|-------|
| `ROSTER_HOME    Deprecated - use KNOSSOS_HOME instead` | REMOVE LINE | `sync-user-skills.sh` | 21 |
| `handles ROSTER_HOME deprecation` | `resolves KNOSSOS_HOME` | `sync-user-skills.sh` | 26 |
| `ROSTER_HOME    Roster repository location (default: ~/Code/roster)` | `KNOSSOS_HOME   Knossos platform location (default: ~/Code/knossos)` | `sync-user-skills.sh` | 925 |
| `ROSTER_HOME    Deprecated - use KNOSSOS_HOME instead` | REMOVE LINE | `sync-user-hooks.sh` | 20 |
| `handles ROSTER_HOME deprecation` | `resolves KNOSSOS_HOME` | `sync-user-hooks.sh` | 25 |
| `ROSTER_HOME    Roster repository location (default: ~/Code/roster)` | `KNOSSOS_HOME   Knossos platform location (default: ~/Code/knossos)` | `sync-user-hooks.sh` | 1018 |
| `ROSTER_HOME    Deprecated - use KNOSSOS_HOME instead` | REMOVE LINE | `sync-user-commands.sh` | 26 |
| `handles ROSTER_HOME deprecation` | `resolves KNOSSOS_HOME` | `sync-user-commands.sh` | 31 |
| `ROSTER_HOME    Roster repository location (default: ~/Code/roster)` | `KNOSSOS_HOME   Knossos platform location (default: ~/Code/knossos)` | `sync-user-commands.sh` | 875 |
| `ROSTER_HOME    Deprecated - use KNOSSOS_HOME instead` | REMOVE LINE | `sync-user-agents.sh` | 19 |
| `handles ROSTER_HOME deprecation` | `resolves KNOSSOS_HOME` | `sync-user-agents.sh` | 24 |
| `ROSTER_HOME    Roster repository location (default: ~/Code/roster)` | `KNOSSOS_HOME   Knossos platform location (default: ~/Code/knossos)` | `sync-user-agents.sh` | 662 |
| `ROSTER_HOME    Deprecated - use KNOSSOS_HOME instead` | REMOVE LINE | `install-hooks.sh` | 19 |
| `handles ROSTER_HOME deprecation` | `resolves KNOSSOS_HOME` | `install-hooks.sh` | 23 |
| `handles ROSTER_HOME deprecation` | `resolves KNOSSOS_HOME` | `swap-rite.sh` | 10 |
| `ROSTER_HOME         Deprecated - use KNOSSOS_HOME instead` | REMOVE LINE | `swap-rite.sh` | 1438 |
| `handles ROSTER_HOME deprecation` | `resolves KNOSSOS_HOME` | `get-workflow-field.sh` | 9 |
| `handles ROSTER_HOME deprecation` | `resolves KNOSSOS_HOME` | `load-workflow.sh` | 8 |
| `handles ROSTER_HOME deprecation` | `resolves KNOSSOS_HOME` | `templates/generate-orchestrator.sh` | 13 |
| `handles ROSTER_HOME deprecation` | `resolves KNOSSOS_HOME` | `templates/validate-orchestrator.sh` | 20 |
| `handles ROSTER_HOME deprecation` | `resolves KNOSSOS_HOME` | `bin/fix-hardcoded-paths.sh` | 17 |
| `formerly ROSTER_HOME` | REMOVE PARENTHETICAL | `lib/rite/rite-hooks-registration.sh` | 20 |
| `ROSTER_HOME deprecated` | REMOVE PARENTHETICAL | `lib/rite/rite-hooks-registration.sh` | 431 |
| `ROSTER_HOME        Path to roster repository (default: ~/Code/roster)` | `KNOSSOS_HOME   Knossos platform location (default: ~/Code/knossos)` | `knossos-sync` | 301 |

### 1.5 ROSTER_SYNC_VERSION -> KNOSSOS_SYNC_VERSION

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `ROSTER_SYNC_VERSION` | `KNOSSOS_SYNC_VERSION` | `knossos-sync` | 46, 309 |

### 1.6 ROSTER_PREF_* -> KNOSSOS_PREF_* (14 variables)

All defined in `.claude/hooks/lib/preferences-loader.sh`. Each `export ROSTER_PREF_X` becomes `export KNOSSOS_PREF_X`.

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `ROSTER_PREF_VERSION` | `KNOSSOS_PREF_VERSION` | `.claude/hooks/lib/preferences-loader.sh` | 354 |
| `ROSTER_PREF_AUTONOMY_LEVEL` | `KNOSSOS_PREF_AUTONOMY_LEVEL` | `.claude/hooks/lib/preferences-loader.sh` | 355 |
| `ROSTER_PREF_FAILURE_HANDLING` | `KNOSSOS_PREF_FAILURE_HANDLING` | `.claude/hooks/lib/preferences-loader.sh` | 356 |
| `ROSTER_PREF_OUTPUT_FORMAT` | `KNOSSOS_PREF_OUTPUT_FORMAT` | `.claude/hooks/lib/preferences-loader.sh` | 357 |
| `ROSTER_PREF_ORCHESTRATION_MODE` | `KNOSSOS_PREF_ORCHESTRATION_MODE` | `.claude/hooks/lib/preferences-loader.sh` | 358 |
| `ROSTER_PREF_ARTIFACT_VERIFICATION` | `KNOSSOS_PREF_ARTIFACT_VERIFICATION` | `.claude/hooks/lib/preferences-loader.sh` | 359 |
| `ROSTER_PREF_NOTIFICATION_LEVEL` | `KNOSSOS_PREF_NOTIFICATION_LEVEL` | `.claude/hooks/lib/preferences-loader.sh` | 360 |
| `ROSTER_PREF_DEFAULT_BRANCH` | `KNOSSOS_PREF_DEFAULT_BRANCH` | `.claude/hooks/lib/preferences-loader.sh` | 361 |
| `ROSTER_PREF_COMMIT_AUTO_PUSH` | `KNOSSOS_PREF_COMMIT_AUTO_PUSH` | `.claude/hooks/lib/preferences-loader.sh` | 364 |
| `ROSTER_PREF_PR_AUTO_CREATE` | `KNOSSOS_PREF_PR_AUTO_CREATE` | `.claude/hooks/lib/preferences-loader.sh` | 365 |
| `ROSTER_PREF_TEST_BEFORE_COMMIT` | `KNOSSOS_PREF_TEST_BEFORE_COMMIT` | `.claude/hooks/lib/preferences-loader.sh` | 366 |
| `ROSTER_PREF_SESSION_AUTO_PARK` | `KNOSSOS_PREF_SESSION_AUTO_PARK` | `.claude/hooks/lib/preferences-loader.sh` | 367 |
| `ROSTER_PREF_EDITOR_INTEGRATION_AUTO_OPEN_FILES` | `KNOSSOS_PREF_EDITOR_INTEGRATION_AUTO_OPEN_FILES` | `.claude/hooks/lib/preferences-loader.sh` | 370 |
| `ROSTER_PREF_EDITOR_INTEGRATION_PRESERVE_CURSOR_POSITION` | `KNOSSOS_PREF_EDITOR_INTEGRATION_PRESERVE_CURSOR_POSITION` | `.claude/hooks/lib/preferences-loader.sh` | 371 |

Also update the comments/docstrings in `.claude/hooks/lib/preferences-loader.sh`:
- Line 341: `ROSTER_PREF_*` -> `KNOSSOS_PREF_*`
- Line 344: `ROSTER_PREF_<KEY>` -> `KNOSSOS_PREF_<KEY>`
- Line 345: `ROSTER_PREF_EDITOR_INTEGRATION_AUTO_OPEN_FILES` -> `KNOSSOS_PREF_EDITOR_INTEGRATION_AUTO_OPEN_FILES`
- Line 373: `ROSTER_PREF_*` -> `KNOSSOS_PREF_*`

### 1.7 ROSTER_PREF_* Consumer (Test File)

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `ROSTER_PREF_*` | `KNOSSOS_PREF_*` | `tests/hooks/test-session-context-preferences.sh` | 79, 95, 96, 97, 98 |

### 1.8 ROSTER_HOME in Projected Hook Files (live env var usage)

These `.claude/hooks/lib/` files use `ROSTER_HOME` as a live env var (not just a deprecation comment). They are projected copies; if regenerated by sync, editing the source suffices. If standalone, they need direct editing.

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `ROSTER_HOME` | `KNOSSOS_HOME` | `.claude/hooks/lib/handoff-validator.sh` | 13, 14, 15 |
| `ROSTER_HOME` | `KNOSSOS_HOME` | `.claude/hooks/lib/artifact-validation.sh` | 8 |

Also, `.claude/hooks/lib/fail-open.sh` contains a hardcoded absolute path:
- Lines 104, 111: `/Users/tomtenuta/Code/roster/ariadne/ari` -> `/Users/tomtenuta/Code/knossos/ariadne/ari`

Note: `.claude/knowledge/consultant/build-capability-index.sh` has been removed (phantom path from legacy knowledge base structure).

### 1.9 ROSTER_HOME in Test Files (used as variable, not deprecated shim)

Test files use `ROSTER_HOME` as the actual env var to locate the repo. These all become `KNOSSOS_HOME`.

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `ROSTER_HOME` | `KNOSSOS_HOME` | `tests/sync/test-sync-config.sh` | 11, 14 |
| `ROSTER_HOME` | `KNOSSOS_HOME` | `tests/sync/test-validate-repair.sh` | 17, 18, 65, 121, 138, 155, 174, 204, 226, 237, 266, 287, 305, 342, 369, 402, 420, 440, 473 |
| `ROSTER_HOME` | `KNOSSOS_HOME` | `tests/sync/test-sync-orphan.sh` | 11, 14, 15, 16, 17 |
| `ROSTER_HOME` | `KNOSSOS_HOME` | `tests/sync/test-sync-checksum.sh` | 11, 14, 15 |
| `ROSTER_HOME` | `KNOSSOS_HOME` | `tests/sync/test-sync-conflict.sh` | 15, 18, 19, 20, 21 |
| `ROSTER_HOME` | `KNOSSOS_HOME` | `tests/sync/test-swap-rite-integration.sh` | 17, 18, 82, 96, 107, 134, 151, 165, 171, 189, 195, 245, 272, 274, 299, 320, 339 |
| `ROSTER_HOME` | `KNOSSOS_HOME` | `tests/sync/test-sync-manifest.sh` | 11, 14, 15, 16 |
| `ROSTER_HOME` | `KNOSSOS_HOME` | `tests/sync/test-init.sh` | 17, 18, 82, 122, 125, 164, 171, 186, 212, 216, 230, 237, 246, 258, 266, 288, 301, 317, 340, 349, 385, 406, 413, 432, 439, 462, 486, 551 |
| `ROSTER_HOME` | `KNOSSOS_HOME` | `tests/lib/rite/test-rite-hooks-registration.sh` | 12, 15, 194 |
| `ROSTER_HOME` | `KNOSSOS_HOME` | `tests/lib/rite/test-rite-transaction.sh` | 78 |
| `ROSTER_HOME` | `KNOSSOS_HOME` | `rites/ecosystem/context-injection.sh` | 39, 57 |

---

## 2. Function Rename Table

### 2.1 Production Functions

| Old Name | New Name | Defined In | Line | Call Sites |
|----------|----------|------------|------|------------|
| `is_roster_managed()` | `is_knossos_managed()` | `sync-user-skills.sh` | 141 | `sync-user-skills.sh:264,490,799,837,875` |
| `is_roster_managed()` | `is_knossos_managed()` | `sync-user-hooks.sh` | 169 | `sync-user-hooks.sh:355,403,554,866,899,932,956` |
| `is_roster_managed()` | `is_knossos_managed()` | `sync-user-commands.sh` | 268 | `sync-user-commands.sh:210,377,590,768,804` |
| `is_roster_managed()` | `is_knossos_managed()` | `sync-user-agents.sh` | 160 | `sync-user-agents.sh:250,445,591,617` |
| `get_roster_commit()` | `get_knossos_commit()` | `lib/sync/sync-core.sh` | 928 | `lib/sync/sync-core.sh:954`, `knossos-sync:450,650` |
| `get_roster_ref()` | `get_knossos_ref()` | `lib/sync/sync-core.sh` | 939 | `knossos-sync:451,650` |
| `roster_has_updates()` | `knossos_has_updates()` | `lib/sync/sync-core.sh` | 951 | `knossos-sync:615` |
| `update_manifest_roster()` | `update_manifest_knossos()` | `lib/sync/sync-manifest.sh` | 237 | `knossos-sync:650` |
| `generate_roster()` | `generate_knossos()` | `lib/knossos-utils.sh` | 123 | `swap-rite.sh:3304` |
| `extract_non_roster_hooks()` | `extract_non_knossos_hooks()` | `lib/rite/rite-hooks-registration.sh` | 197 | `lib/rite/rite-hooks-registration.sh:467` |
| `roster_sync_available()` | `knossos_sync_available()` | `swap-rite.sh` | 2797 | `swap-rite.sh:2977` (via condition check) |
| `roster_has_updates()` | `knossos_has_updates()` | `swap-rite.sh` | 2805 | `swap-rite.sh:2822,2977` |
| `run_roster_sync_waterfall()` | `run_knossos_sync_waterfall()` | `swap-rite.sh` | 2856 | `swap-rite.sh:2971,2979` |

### 2.2 Test Functions

| Old Name | New Name | Defined In | Line |
|----------|----------|------------|------|
| `test_classify_update_roster_changed()` | `test_classify_update_knossos_changed()` | `tests/sync/test-sync-conflict.sh` | 175 |
| `test_classify_update_roster_changed` (call) | `test_classify_update_knossos_changed` | `tests/sync/test-sync-conflict.sh` | 786 |
| `test_update_manifest_roster()` | `test_update_manifest_knossos()` | `tests/sync/test-sync-manifest.sh` | 155 |
| `test_update_manifest_roster` (call) | `test_update_manifest_knossos` | `tests/sync/test-sync-manifest.sh` | 389 |
| `test_init_inside_roster_fails()` | `test_init_inside_knossos_fails()` | `tests/sync/test-init.sh` | 297 |
| `test_init_inside_roster_fails` (call) | `test_init_inside_knossos_fails` | `tests/sync/test-init.sh` | 596 |
| `test_roster_has_updates_no_manifest()` | `test_knossos_has_updates_no_manifest()` | `tests/sync/test-swap-rite-integration.sh` | 126 |
| `test_roster_has_updates_stale_manifest()` | `test_knossos_has_updates_stale_manifest()` | `tests/sync/test-swap-rite-integration.sh` | 143 |
| `test_roster_has_updates_current_manifest()` | `test_knossos_has_updates_current_manifest()` | `tests/sync/test-swap-rite-integration.sh` | 160 |
| `test_roster_sync_available()` | `test_knossos_sync_available()` | `tests/sync/test-swap-rite-integration.sh` | 266 |
| `test_extract_non_roster_roster_only()` | `test_extract_non_knossos_knossos_only()` | `tests/lib/rite/test-rite-hooks-registration.sh` | 327 |
| `test_extract_non_roster_mixed()` | `test_extract_non_knossos_mixed()` | `tests/lib/rite/test-rite-hooks-registration.sh` | 340 |
| `test_extract_non_roster_user_only()` | `test_extract_non_knossos_user_only()` | `tests/lib/rite/test-rite-hooks-registration.sh` | 369 |
| `test_extract_non_roster_missing()` | `test_extract_non_knossos_missing()` | `tests/lib/rite/test-rite-hooks-registration.sh` | 386 |
| All 4 test function calls | Match new names | `tests/lib/rite/test-rite-hooks-registration.sh` | 824-827 |

---

## 3. Local Variable Rename Table

### 3.1 `roster_changed`

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `roster_changed` | `knossos_changed` | `lib/sync/sync-core.sh` | 100, 104, 112, 115, 118 |
| `roster_changed` | `knossos_changed` | `lib/sync/sync-core.sh` | 489, 492, 495, 498, 501 |

### 3.2 `roster_dir`

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `roster_dir` | `knossos_dir` | `lib/sync/sync-core.sh` | 532, 542, 584, 597, 643, 664, 846, 849, 861, 864, 890, 895, 904 |

### 3.3 `roster_path`

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `roster_path` | `knossos_path` | `lib/sync/sync-core.sh` | 929, 931, 932, 940, 942, 943 |
| `roster_path` | `knossos_path` | `lib/sync/sync-manifest.sh` | 180, 190, 325, 326, 327, 380, 388 |
| `roster_path` | `knossos_path` | `knossos-sync` | 389, 390, 393, 449, 771, 772, 775, 778, 779, 949, 950, 977, 980, 982, 983, 984, 989, 992, 1239, 1241, 1247, 1258 |
| `roster_path` | `knossos_path` | `tests/sync/test-validate-repair.sh` | 65, 71 |
| `roster_path` | `knossos_path` | `tests/sync/test-sync-conflict.sh` | 77, 82 |
| `roster_path` | `knossos_path` | `tests/sync/test-sync-manifest.sh` | 244, 245, 246, 302, 303, 304 |
| `roster_path` | `knossos_path` | `tests/sync/test-init.sh` | 120, 121, 122, 125, 491, 493, 504 |
| `roster_path` | `knossos_path` | `tests/sync/test-swap-rite-integration.sh` | 222, 223, 225 |
| `roster_path` | `knossos_path` | `swap-rite.sh` | 2924, 2930 |

### 3.4 `roster_commit`

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `roster_commit` | `knossos_commit` | `lib/sync/sync-manifest.sh` | 181, 191 |
| `roster_commit` | `knossos_commit` | `knossos-sync` | 449, 450, 453, 461, 771, 773, 783, 950, 951, 977, 978, 1239, 1242, 1256 |
| `roster_commit` | `knossos_commit` | `tests/sync/test-init.sh` | 491, 494, 510 |

### 3.5 `roster_ref`

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `roster_ref` | `knossos_ref` | `lib/sync/sync-manifest.sh` | 182, 192 |
| `roster_ref` | `knossos_ref` | `knossos-sync` | 449, 451, 454, 461, 951, 979 |
| `roster_ref` | `knossos_ref` | `tests/sync/test-init.sh` | 491, 495, 516 |
| `roster_ref` | `knossos_ref` | `rites/ecosystem/context-injection.sh` | 38, 41, 43, 45 |

### 3.6 `roster_checksum`

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `roster_checksum` | `knossos_checksum` | `tests/sync/test-sync-conflict.sh` | 161, 162, 163 |

### 3.7 `roster_file`

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `roster_file` | `knossos_file` | `knossos-sync` | 464, 469, 472, 473, 483, 486, 488, 504, 507, 508, 517, 520, 522, 992, 997, 1000, 1001, 1004, 1007, 1028, 1031, 1032, 1035, 1038, 1131, 1139, 1141, 1142, 1152, 1154, 1155, 1297, 1300, 1301, 1310, 1316, 1320, 1323, 1332 |

### 3.8 `roster_rites`

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `roster_rites` | `knossos_rites` | `sync-user-hooks.sh` | 124, 126, 131, 141, 143, 149 |
| `roster_rites` | `knossos_rites` | `sync-user-commands.sh` | 123, 125, 130, 140, 142, 148 |
| `roster_rites` | `knossos_rites` | `sync-user-agents.sh` | 115, 117, 122, 132, 134, 140 |

### 3.9 `roster_home`

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `roster_home` | `knossos_home` | `rites/ecosystem/context-injection.sh` | 39, 40, 41, 42, 43 |
| `roster_home` | `knossos_home` | `user-hooks/lib/rite-context-loader.sh` | 53, 54 |

### 3.10 `roster_sync` (local variable in functions)

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `roster_sync` | `knossos_sync` | `swap-rite.sh` | 2798, 2799, 2860, 2862, 2873, 2875 |

### 3.11 `roster_realpath`

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `roster_realpath` | `knossos_realpath` | `knossos-sync` | 389, 390, 393 |

### 3.12 `roster_branch`

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `roster_branch` | `knossos_branch` | `rites/ecosystem/context-injection.sh` | 42, 43 |

### 3.13 `roster_marker`

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `roster_marker` | `knossos_marker` | `lib/sync/merge/merge-docs.sh` | 86, 89, 92, 95 |

### 3.14 `roster_sections`

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `roster_sections` | `knossos_sections` | `lib/sync/merge/merge-docs.sh` | 74, 75, 78, 117, 124 |

### 3.15 `roster_content`

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `roster_content` | `knossos_content` | `tests/sync/test-sync-conflict.sh` | 499, 501, 511, 514 |

### 3.16 `roster_last_sync`

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `roster_last_sync` | `knossos_last_sync` | `tests/sync/test-init.sh` | 491, 496 (variable `roster_last_sync` declared but not referenced by that name in assertions; line 522 uses it implicitly) |

### 3.17 `old_roster_path`, `old_roster_commit`, `old_roster_ref`

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `old_roster_path` | `old_knossos_path` | `knossos-sync` | 949, 961, 982, 983, 984, 1091 |
| `old_roster_commit` | `old_knossos_commit` | `knossos-sync` | 950, 962 |
| `old_roster_ref` | `old_knossos_ref` | `knossos-sync` | 951, 963 |

### 3.18 `has_roster`

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `has_roster` | `has_knossos` | `tests/sync/test-sync-conflict.sh` | 307, 308, 311, 314 |

### 3.19 `TEST_ROSTER_DIR`

| Old Name | New Name | File | Lines |
|----------|----------|------|-------|
| `TEST_ROSTER_DIR` | `TEST_KNOSSOS_DIR` | `tests/sync/test-sync-orphan.sh` | 26, 50, 54, 71, 82, 89, 108, 119, 143, 168, 195, 391, 403, 413, 435, 466 |

---

## 4. JSON Manifest Key Migration

The CEM manifest (`manifest.json`) uses `.roster.*` keys. These change to `.knossos.*`.

### 4.1 Key Mapping

| Old Key | New Key |
|---------|---------|
| `.roster.path` | `.knossos.path` |
| `.roster.commit` | `.knossos.commit` |
| `.roster.ref` | `.knossos.ref` |
| `.roster.last_sync` | `.knossos.last_sync` |
| `.rite.roster_path` | `.rite.knossos_path` |

### 4.2 Files That Read/Write These Keys

#### `.roster.path`

| File | Lines | Operation |
|------|-------|-----------|
| `lib/sync/sync-manifest.sh` | 195, 326, 394 | write (jq template), read |
| `knossos-sync` | 772, 776, 779, 1241, 1247, 1258 | read |
| `tests/sync/test-validate-repair.sh` | 70 | write (JSON literal) |
| `tests/sync/test-sync-manifest.sh` | 79, 80, 245, 303 | read |
| `tests/sync/test-sync-conflict.sh` | 84, 87 | write (jq template) |
| `tests/sync/test-swap-rite-integration.sh` | 81, 82 | write (JSON literal) |
| `tests/sync/test-init.sh` | 121, 493 | read |

#### `.roster.commit`

| File | Lines | Operation |
|------|-------|-----------|
| `lib/sync/sync-manifest.sh` | 249 | write (jq) |
| `lib/sync/sync-core.sh` | 955 | read |
| `knossos-sync` | 773, 783, 1242, 1256, 1260 | read |
| `swap-rite.sh` | 2837 | read |
| `tests/sync/test-sync-manifest.sh` | 87, 163 | read |
| `tests/sync/test-validate-repair.sh` | (via manifest structure) | write |
| `tests/sync/test-init.sh` | 494 | read |

#### `.roster.ref`

| File | Lines | Operation |
|------|-------|-----------|
| `lib/sync/sync-manifest.sh` | 253 | write (jq) |
| `tests/sync/test-sync-manifest.sh` | 171 | read |
| `tests/sync/test-init.sh` | 495 | read |

#### `.roster.last_sync`

| File | Lines | Operation |
|------|-------|-----------|
| `lib/sync/sync-manifest.sh` | 250 | write (jq) |
| `knossos-sync` | 1243 | read |
| `tests/sync/test-init.sh` | 496 | read (implicitly via `roster_last_sync` var) |

#### `.rite.roster_path`

| File | Lines | Operation |
|------|-------|-----------|
| `swap-rite.sh` | 2924, 2930 | write (jq template) |
| `tests/sync/test-swap-rite-integration.sh` | 223 | read |

### 4.3 jq Template Changes in `lib/sync/sync-manifest.sh`

Line 195 (`create_manifest`): Change `roster:` block to `knossos:` in jq heredoc:
```
roster: {           ->  knossos: {
    path: $rp,              path: $rp,
    commit: $rc,            commit: $rc,
    ref: $rr,               ref: $rr,
    last_sync: $ts          last_sync: $ts
}                       }
```

Lines 249-253 (`update_manifest_knossos`): Change `.roster.commit`, `.roster.last_sync`, `.roster.ref` to `.knossos.commit`, `.knossos.last_sync`, `.knossos.ref`.

Line 229 (`create_manifest`): Change `source: "roster"` to `source: "knossos"` in default file entries.

Line 394 (`create_rite_manifest`): Change `roster_path: $rp` to `knossos_path: $rp`.

### 4.4 Source Value Change

The manifest tracks where files came from. The string value `"roster"` and `"roster-diverged"` stored in JSON must change to `"knossos"` and `"knossos-diverged"`:

| Old Value | New Value | Files |
|-----------|-----------|-------|
| `"roster"` | `"knossos"` | `sync-user-skills.sh:154,171,287,429,768`, `sync-user-hooks.sh:182,199,376,423,487,836`, `sync-user-commands.sh:281,298,404,513,741`, `sync-user-agents.sh:173,190,267,375,567` |
| `"roster-diverged"` | `"knossos-diverged"` | `sync-user-skills.sh:154,171,296`, `sync-user-hooks.sh:182,199,385,431`, `sync-user-commands.sh:281,298,413`, `sync-user-agents.sh:173,190,276` |

---

## 5. String Literals and Comments

### 5.1 Log Prefix: `[roster-sync]` -> `[knossos-sync]`

| File | Lines |
|------|-------|
| `lib/sync/sync-config.sh` | 184, 192, 200, 208 |
| `knossos-sync` | 72, 76, 80, 84, 89 |

### 5.2 Comment Header: `Part of: roster-sync` -> `Part of: knossos-sync`

| File | Lines |
|------|-------|
| `lib/sync/sync-core.sh` | 8 |
| `lib/sync/merge/merge-docs.sh` | 8 |
| `lib/sync/merge/merge-settings.sh` | 8 |
| `lib/sync/merge/dispatcher.sh` | 8 |
| `lib/sync/sync-manifest.sh` | 8 |
| `lib/sync/sync-config.sh` | 6, 8 |
| `lib/sync/sync-checksum.sh` | 8 |
| `tests/sync/test-sync-conflict.sh` | 8 |
| `tests/sync/test-swap-rite-integration.sh` | 11 |

### 5.3 Comment Header: `Part of: roster rite-swap` -> `Part of: knossos rite-swap`

| File | Lines |
|------|-------|
| `lib/rite/rite-hooks-registration.sh` | 8 |
| `lib/rite/rite-resource.sh` | 8 |
| `lib/rite/rite-transaction.sh` | 8 |

### 5.4 Script Title Comment: `roster-sync` -> `knossos-sync`

| File | Lines | Old Text | New Text |
|------|-------|----------|----------|
| `knossos-sync` | 3 | `roster-sync - Roster-Native Ecosystem Synchronization` | `knossos-sync - Knossos-Native Ecosystem Synchronization` |
| `knossos-sync` | 5 | `roster-native functionality` | `knossos-native functionality` |
| `knossos-sync` | 6 | `between roster and satellite` | `between knossos and satellite` |
| `knossos-sync` | 11 | `roster-sync <command>` | `knossos-sync <command>` |
| `knossos-sync` | 236 | `roster-sync - Roster-Native` | `knossos-sync - Knossos-Native` |
| `knossos-sync` | 239 | `roster-sync <command>` | `knossos-sync <command>` |

### 5.5 User-Facing Help Text Containing "roster"

| File | Lines | Old Text | New Text |
|------|-------|----------|----------|
| `knossos-sync` | 242 | `from roster` | `from knossos` |
| `knossos-sync` | 243 | `from roster to satellite` | `from knossos to satellite` |
| `knossos-sync` | 247 | `and roster files` | `and knossos files` |
| `knossos-sync` | 253 | `no longer in roster` | `no longer in knossos` |
| `knossos-sync` | 263-290 | multiple `roster-sync` command examples | `knossos-sync` |
| `knossos-sync` | 301-302 | `ROSTER_HOME`, `ROSTER_SYNC_DEBUG` | `KNOSSOS_HOME`, `KNOSSOS_SYNC_DEBUG` |
| `knossos-sync` | 309-311 | `roster-sync $ROSTER_SYNC_VERSION`, `ROSTER_HOME` | `knossos-sync $KNOSSOS_SYNC_VERSION`, `KNOSSOS_HOME` |
| `knossos-sync` | 367 | `with roster ecosystem` | `with knossos ecosystem` |
| `knossos-sync` | 372-373 | `Roster not found`, `Set ROSTER_HOME` | `Knossos not found`, `Set KNOSSOS_HOME` |
| `knossos-sync` | 394 | `inside roster repository` | `inside knossos repository` |
| `knossos-sync` | 428 | `Roster: $KNOSSOS_HOME` | `Knossos: $KNOSSOS_HOME` |
| `knossos-sync` | 592 | `Syncing from roster` | `Syncing from knossos` |
| `knossos-sync` | 597 | `Run: roster-sync init` | `Run: knossos-sync init` |
| `knossos-sync` | 626 | `Would sync from roster` | `Would sync from knossos` |
| `knossos-sync` | 703 | `roster-sync sync --force` | `knossos-sync sync --force` |
| `knossos-sync` | 744 | `roster-sync init` | `knossos-sync init` |
| `knossos-sync` | 766 | `roster-sync sync` | `knossos-sync sync` |
| `knossos-sync` | 776 | `roster.path` | `knossos.path` |
| `knossos-sync` | 779 | `Roster path does not exist` | `Knossos path does not exist` |
| `knossos-sync` | 784 | `roster.commit` | `knossos.commit` |
| `knossos-sync` | 936 | `roster-sync init` | `knossos-sync init` |
| `knossos-sync` | 942-943 | `Roster not found`, `ROSTER_HOME` | `Knossos not found`, `KNOSSOS_HOME` |
| `knossos-sync` | 984 | `Roster path changed` | `Knossos path changed` |
| `knossos-sync` | 1220 | `roster-sync status` | `knossos-sync status` |
| `knossos-sync` | 1228 | `roster-sync init` | `knossos-sync init` |
| `knossos-sync` | 1247 | `Roster Path:` | `Knossos Path:` |
| `knossos-sync` | 1301 | `not found in roster` | `not found in knossos` |
| `knossos-sync` | 1313 | `with roster` | `with knossos` |
| `lib/sync/sync-core.sh` | 332, 338, 339, 343, 379 | `roster-sync sync`, `roster updates`, `roster version` | `knossos-sync sync`, `knossos updates`, `knossos version` |
| `lib/sync/sync-manifest.sh` | 283 | `'roster-sync init'` | `'knossos-sync init'` |
| `lib/sync/sync-manifest.sh` | 328 | `Missing roster.path` | `Missing knossos.path` |
| `swap-rite.sh` | 58-60 | `roster-sync integration`, `run roster-sync` | `knossos-sync integration`, `run knossos-sync` |
| `swap-rite.sh` | 642, 2792, 2795, 2803, 2853, 2863, 2869, 2880, 2885, 2902, 2967, 2970, 3224, 3226, 3229 | various `roster-sync` references | `knossos-sync` |
| `swap-rite.sh` | 710, 712, 718, 1412, 1419, 1420, 1457, 1466, 1749, 1751, 1767, 2232, 2421, 2505, 2518, 2527, 2594, 2702, 2716, 2998 | various "roster" in user text | "knossos" |
| `user-hooks/lib/worktree-manager.sh` | 227, 229, 238, 240, 242, 248, 250 | `roster-sync` | `knossos-sync` |

### 5.6 Comment-Level "roster" References in Other Files

| File | Lines | Change |
|------|-------|--------|
| `lib/knossos-utils.sh` | 3, 5, 9, 10, 16, 117, 120, 121 | `roster-utils.sh` -> `knossos-utils.sh`, `roster` -> `knossos` in comments |
| `lib/sync/sync-core.sh` | 18, 41-42, 119, 221, 226, 231, 272, 525, 530, 535, 564, 581, 600, 631, 633, 636, 637, 640, 644, 655, 663, 664, 667, 668, 830, 843, 858, 874, 887, 927, 938, 949, 958 | `roster` -> `knossos` in doc comments |
| `lib/sync/merge/merge-docs.sh` | 6, 15, 30, 42-49, 58, 60, 70, 78, 81, 96, 97, 104, 106, 119, 123, 124 | `roster` -> `knossos` in comments and `SYNC_MARKER_ROSTER` -> `SYNC_MARKER_KNOSSOS` |
| `lib/sync/merge/merge-settings.sh` | 6, 15, 31, 33, 42, 44, 51, 60, 84 | `roster` -> `knossos` in comments |
| `lib/sync/merge/dispatcher.sh` | 39, 43, 45 | `roster` -> `knossos` in comments and log messages |
| `lib/sync/sync-config.sh` | 14, 50, 51 | `roster` -> `knossos` in comments |
| `lib/sync/sync-manifest.sh` | 58, 178, 235, 236 | `roster` -> `knossos` in comments |
| `lib/rite/rite-hooks-registration.sh` | 191, 214, 215, 233, 280, 427, 465, 471 | `roster` -> `knossos` in comments |
| `lib/rite/rite-transaction.sh` | 456, 482 | `ROSTER_HOME` -> `KNOSSOS_HOME` in "Requires" comments |
| `sync-user-skills.sh` | 3, 5, 10, 11, 140, 237, 263, 264, 269, 283, 292, 302, 387, 491, 630, 631, 651, 652, 657, 724, 751, 768, 770, 799, 837, 861, 876, 890, 900, 901, 909, 914, 915, 916 | `roster` -> `knossos` in comments, help text, log messages |
| `sync-user-hooks.sh` | 3, 5, 8, 9, 168, 334, 354, 360, 368, 372, 381, 391, 402, 408, 437, 445, 555, 693, 726, 789, 810, 836, 838, 918, 933, 957, 972, 983, 984, 988, 993, 994, 995, 1004, 1008 | `roster` -> `knossos` in comments, help text, log messages |
| `sync-user-commands.sh` | 3, 5, 8, 9, 161, 162, 177, 217, 267, 356, 376, 382, 400, 409, 419, 473, 591, 631, 632, 656, 657, 668, 712, 724, 741, 743, 787, 805, 819, 830, 831, 835, 840, 841, 842, 852, 853, 866 | `roster` -> `knossos` in comments, help text, log messages |
| `sync-user-agents.sh` | 3, 5, 8, 9, 159, 229, 249, 255, 263, 272, 282, 336, 446, 485, 486, 506, 507, 511, 528, 550, 552, 567, 569, 609, 618, 632, 642, 643, 646, 651, 652, 653 | `roster` -> `knossos` in comments, help text, log messages |
| `install-hooks.sh` | 3, 5, 56 | `roster` -> `knossos` in comments and help |
| `swap-rite.sh` | 792, 2330, 2421, 2505, 2518, 2527, 2594, 2702, 2716 | `roster` -> `knossos` in comments |

### 5.7 SYNC_MARKER_ROSTER Constant

| Old Name | New Name | File | Line |
|----------|----------|------|------|
| `SYNC_MARKER_ROSTER` | `SYNC_MARKER_KNOSSOS` | `lib/sync/merge/merge-docs.sh` | 30 |

Note: The marker VALUE `"<!-- SYNC: roster-owned -->"` also changes to `"<!-- SYNC: knossos-owned -->"`. This is a data format change affecting existing satellite `.claude/CLAUDE.md` files. The integration-engineer must update all satellites that contain this marker.

### 5.8 Error/Log Messages in `lib/sync/merge/dispatcher.sh`

| Line | Old Text | New Text |
|------|----------|----------|
| 39 | `roster:` | `knossos:` |
| 45 | `Roster file not found` | `Knossos file not found` |

---

## 6. YAML Config Updates

### 6.1 `rites/ecosystem/manifest.yaml`

| Line | Old Text | New Text |
|------|----------|----------|
| 6 | `CEM/roster changes` | `CEM/knossos changes` |
| 25 | `CEM and roster infrastructure changes` | `CEM and knossos infrastructure changes` |
| 43 | `CEM/roster problems` | `CEM/knossos problems` |
| 49 | `CEM and roster changes` | `CEM and knossos changes` |
| 73 | `Single system (CEM or roster)` | `Single system (CEM or knossos)` |
| 75 | `affecting CEM + roster` | `affecting CEM + knossos` |

### 6.2 `rites/ecosystem/orchestrator.yaml`

| Line | Old Text | New Text |
|------|----------|----------|
| 15 | `CEM/roster infrastructure work` | `CEM/knossos infrastructure work` |
| 57 | `CEM/roster patterns` | `CEM/knossos patterns` |

### 6.3 `rites/ecosystem/workflow.yaml`

| Line | Old Text | New Text |
|------|----------|----------|
| 45 | `CEM or roster` | `CEM or knossos` |
| 48 | `CEM + roster` | `CEM + knossos` |

### 6.4 `rites/forge/manifest.yaml`

| Line | Old Text | New Text |
|------|----------|----------|
| 28 | `produces: roster-integration` | `produces: knossos-integration` |
| 29 | `into roster ecosystem` | `into knossos ecosystem` |
| 57 | `into roster platform` | `into knossos platform` |

### 6.5 `rites/forge/orchestrator.yaml`

| Line | Old Text | New Text |
|------|----------|----------|
| 18 | `roster integration needed` | `knossos integration needed` |
| 23 | `extend roster` | `extend knossos` |
| 24 | `integrated into roster ecosystem` | `integrated into knossos ecosystem` |
| 40 | `Agents registered in roster` | `Agents registered in knossos` |
| 55 | `roster patterns` | `knossos patterns` |
| 62 | `roster changes affecting CEM/skeleton` | `knossos changes affecting CEM/skeleton` |

### 6.6 `rites/forge/workflow.yaml`

| Line | Old Text | New Text |
|------|----------|----------|
| 30 | `produces: roster-integration` | `produces: knossos-integration` |

### 6.7 `schemas/handoff-criteria-schema.yaml`

| Line | Old Text | New Text |
|------|----------|----------|
| 140 | `all in ['roster', 'CEM']` | `all in ['knossos', 'CEM']` |
| 377 | `${ROSTER_DIR}/schemas/` | `${KNOSSOS_DIR}/schemas/` |

---

## 7. Infrastructure

### 7.1 `.gitignore`

| Line | Old Text | New Text |
|------|----------|----------|
| 1 | `# Roster-managed runtime directory` | `# Knossos-managed runtime directory` |
| 2 | `roster (ecosystem), teams/*/ (team definitions)` | `knossos (ecosystem), rites/*/ (rite definitions)` |
| 3 | `via roster-sync and team swap` | `via knossos-sync and rite swap` |

### 7.2 `~/Code/roster` Default Path

The default path `$HOME/Code/roster` must change to `$HOME/Code/knossos`. This affects:

| File | Lines |
|------|-------|
| `lib/knossos-home.sh` | 10, 30 |
| `lib/sync/sync-core.sh` | 929, 940 |
| `user-hooks/lib/config.sh` | 16 |
| `user-hooks/lib/session-manager.sh` | 98 |
| `user-hooks/lib/rite-context-loader.sh` | 53 |
| `user-hooks/validation/command-validator.sh` | 17 |
| `rites/ecosystem/context-injection.sh` | 39, 57 |
| `sync-user-skills.sh` | 20 |
| `sync-user-hooks.sh` | 19 |
| `sync-user-commands.sh` | 25 |
| `sync-user-agents.sh` | 18 |
| `install-hooks.sh` | 18 |
| `swap-rite.sh` | 1437 |
| `templates/orchestrator-generate.sh` | 104 |
| `knossos-sync` | 301 |

Note: The help-text lines in `sync-user-*.sh` (e.g., line 925, 1018, 875, 662) that say `default: ~/Code/roster` also change.

### 7.3 Hardcoded Absolute Paths in Test Files

| File | Lines | Old Path | New Path |
|------|-------|----------|----------|
| `tests/integration/test-d002-simple.sh` | 11, 49, 50, 51 | `/Users/tomtenuta/Code/roster` | `/Users/tomtenuta/Code/knossos` |
| `tests/integration/test-d002-output-format.sh` | 24, 28 | `/Users/tomtenuta/Code/roster` | `/Users/tomtenuta/Code/knossos` |

### 7.4 `bin/fix-hardcoded-paths.sh` Self-References

This script references `~/Code/roster` as the thing it scans for. All internal references to `roster` in this file must be updated. Key lines: 3, 10, 11, 69, 74, 75, 76, 85, 95, 101, 105, 123, 124, 141, 143, 147, 152, 154, 165.

The script's variable references to `$ROSTER_HOME` become `$KNOSSOS_HOME`, and string patterns `~/Code/roster` become `~/Code/knossos`.

### 7.5 `scripts/docs/verify-doctrine.sh`

| Line | Old Text | New Text |
|------|----------|----------|
| 78 | `${PROJECT_ROOT}/roster/rites/` | `${PROJECT_ROOT}/knossos/rites/` (NOTE: verify this path is correct in context -- it may be a stale reference) |

### 7.6 `bin/normalize-rite-structure.sh`

| Line | Old Text | New Text |
|------|----------|----------|
| 66 | `your roster repository` | `your knossos repository` |

### 7.7 `templates/orchestrator-generate.sh`

| Line | Old Text | New Text |
|------|----------|----------|
| 104 | `Root directory for roster (default: ~/Code/roster)` | `Root directory for knossos (default: ~/Code/knossos)` |

---

## 8. File Ordering Constraints

Files MUST be edited in this order to maintain source-ability during implementation. If any file is sourced by another, the sourced file must be updated first.

### Phase 1: Core Libraries (no dependencies on other roster files)

1. `lib/knossos-home.sh` -- sourced by everything; update default path
2. `lib/sync/sync-config.sh` -- sourced by all sync files; update ROSTER_SYNC_DEBUG, log prefixes
3. `lib/sync/sync-checksum.sh` -- comments only
4. `lib/sync/merge/merge-docs.sh` -- SYNC_MARKER_ROSTER constant, comments
5. `lib/sync/merge/merge-settings.sh` -- comments only
6. `lib/sync/merge/dispatcher.sh` -- comments and log messages
7. `lib/sync/sync-manifest.sh` -- .roster JSON keys, function rename, comments
8. `lib/sync/sync-core.sh` -- function renames, local variables, comments
9. `lib/knossos-utils.sh` -- generate_roster function rename, comments
10. `lib/rite/rite-hooks-registration.sh` -- extract_non_roster_hooks rename
11. `lib/rite/rite-resource.sh` -- comments only
12. `lib/rite/rite-transaction.sh` -- comments only

### Phase 2: Hook Libraries

13. `user-hooks/lib/config.sh` -- default path
14. `user-hooks/lib/rite-context-loader.sh` -- roster_home variable
15. `user-hooks/lib/session-manager.sh` -- default path
16. `user-hooks/lib/worktree-manager.sh` -- roster-sync references
17. `.claude/hooks/lib/preferences-loader.sh` -- all ROSTER_PREF_* exports

### Phase 2b: Projected Hook Libraries (`.claude/hooks/lib/`)

These are projected copies that also contain "roster" references. They must be updated alongside their canonical sources. If they are regenerated by sync, updating the canonical source suffices. If they are standalone, they need direct editing.

13b. `.claude/hooks/lib/handoff-validator.sh` -- `ROSTER_HOME` env var (lines 13-15)
13c. `.claude/hooks/lib/artifact-validation.sh` -- `ROSTER_HOME` env var (line 8)
13d. `.claude/hooks/lib/fail-open.sh` -- hardcoded path `/Users/tomtenuta/Code/roster/ariadne/ari` (lines 104, 111)
13e. (removed -- `.claude/knowledge/consultant/` was a phantom path)

### Phase 3: Hook Scripts

18. `user-hooks/context-injection/session-context.sh` -- ROSTER_VERBOSE
19. `user-hooks/validation/command-validator.sh` -- ROSTER_DIR

### Phase 4: Top-Level Sync Scripts (consume lib/)

20. `sync-user-agents.sh` -- is_roster_managed, "roster" source values, comments
21. `sync-user-skills.sh` -- is_roster_managed, "roster" source values, comments
22. `sync-user-hooks.sh` -- is_roster_managed, "roster" source values, comments
23. `sync-user-commands.sh` -- is_roster_managed, "roster" source values, comments
24. `install-hooks.sh` -- comments, ROSTER_HOME deprecation
25. `swap-rite.sh` -- function renames, local variables, comments, help text
26. `knossos-sync` -- everything: env vars, variables, help text, log prefixes

### Phase 5: Utility/Template Scripts

27. `get-workflow-field.sh` -- comment
28. `load-workflow.sh` -- comment
29. `generate-rite-context.sh` -- check for roster references
30. `templates/generate-orchestrator.sh` -- help text
31. `templates/validate-orchestrator.sh` -- comment
32. `templates/orchestrator-generate.sh` -- help text
33. `bin/fix-hardcoded-paths.sh` -- self-references
34. `bin/normalize-rite-structure.sh` -- comment (if any)
35. `scripts/docs/verify-doctrine.sh` -- path reference

### Phase 6: Rite-Specific Scripts

36. `rites/ecosystem/context-injection.sh` -- ROSTER_HOME, roster_home/ref/branch vars

### Phase 7: YAML Configuration

37. `rites/ecosystem/manifest.yaml`
38. `rites/ecosystem/orchestrator.yaml`
39. `rites/ecosystem/workflow.yaml`
40. `rites/forge/manifest.yaml`
41. `rites/forge/orchestrator.yaml`
42. `rites/forge/workflow.yaml`
43. `schemas/handoff-criteria-schema.yaml`

### Phase 8: Infrastructure

44. `.gitignore`

### Phase 9: Test Files (LAST -- they must match the production changes)

45. `tests/sync/test-sync-config.sh`
46. `tests/sync/test-sync-checksum.sh`
47. `tests/sync/test-sync-manifest.sh`
48. `tests/sync/test-sync-conflict.sh`
49. `tests/sync/test-sync-orphan.sh`
50. `tests/sync/test-sync-core.sh` (if exists)
51. `tests/sync/test-init.sh`
52. `tests/sync/test-validate-repair.sh`
53. `tests/sync/test-swap-rite-integration.sh`
54. `tests/hooks/test-session-context-preferences.sh`
55. `tests/lib/rite/test-rite-hooks-registration.sh`
56. `tests/lib/rite/test-rite-transaction.sh`
57. `tests/integration/test-d002-simple.sh`
58. `tests/integration/test-d002-output-format.sh`
59. `tests/test-rite-context-loader.sh`

---

## 9. Test Files

### 9.1 `tests/sync/test-sync-config.sh`

- Line 11: `ROSTER_HOME` -> `KNOSSOS_HOME`
- Line 14: `$ROSTER_HOME/lib/sync/sync-config.sh` -> `$KNOSSOS_HOME/lib/sync/sync-config.sh`

### 9.2 `tests/sync/test-sync-checksum.sh`

- Line 11: `ROSTER_HOME` -> `KNOSSOS_HOME`
- Lines 14-15: `$ROSTER_HOME/lib/sync/` -> `$KNOSSOS_HOME/lib/sync/`

### 9.3 `tests/sync/test-sync-manifest.sh`

- Line 11: `ROSTER_HOME` -> `KNOSSOS_HOME`
- Lines 14-16: `$ROSTER_HOME/lib/sync/` -> `$KNOSSOS_HOME/lib/sync/`
- Lines 68, 79-83: `/path/to/roster` -> `/path/to/knossos`, `roster.path` -> `knossos.path`, `roster.commit` -> `knossos.commit`
- Line 99: `/test/roster` -> `/test/knossos`
- Lines 155-156, 160, 163, 171, 389: `update_manifest_roster` -> `update_manifest_knossos`, `roster.commit` -> `knossos.commit`, `roster.ref` -> `knossos.ref`
- Lines 244-249, 302-307: `roster_path` -> `knossos_path`, `roster.path` -> `knossos.path`

### 9.4 `tests/sync/test-sync-conflict.sh`

- Line 8: comment `roster-sync`
- Line 15: `ROSTER_HOME` -> `KNOSSOS_HOME`
- Lines 18-21: `$ROSTER_HOME/lib/sync/` -> `$KNOSSOS_HOME/lib/sync/`
- Line 55: `$TEST_TMP/roster` -> `$TEST_TMP/knossos`
- Lines 77, 82, 87: `roster_path` -> `knossos_path`, `roster:` -> `knossos:` in JSON
- Line 118: `source: "roster"` -> `source: "knossos"`
- Lines 134, 140, 144, 157, 160-166, 175-176, 178, 180, 182, 188, 191, 193, 198, 200, 202, 204, 209, 221, 223, 226, 239, 241, 245, 258, 260, 264, 280, 283, 286, 298, 301, 305, 307-314, 321, 324, 328, 457, 459, 462, 478, 480, 487, 499-501, 503, 506, 511, 514, 522, 524, 531, 597, 602, 608, 613, 616, 621, 623, 630, 633, 635, 638, 643, 645, 786: All `$TEST_TMP/roster` -> `$TEST_TMP/knossos`, `roster_checksum` -> `knossos_checksum`, `roster_content` -> `knossos_content`, `has_roster` -> `has_knossos`, `test_classify_update_roster_changed` -> `test_classify_update_knossos_changed`, `roster changed` -> `knossos changed`, `roster changes` -> `knossos changes`, `roster content` -> `knossos content`

### 9.5 `tests/sync/test-sync-orphan.sh`

- Line 11: `ROSTER_HOME` -> `KNOSSOS_HOME`
- Lines 14-17: `$ROSTER_HOME/lib/sync/` -> `$KNOSSOS_HOME/lib/sync/`
- Line 26: `TEST_ROSTER_DIR` -> `TEST_KNOSSOS_DIR`
- Lines 50, 53, 54, 71, 82, 88, 89, 107, 108, 119, 122, 129, 132, 141, 143, 146, 168, 195, 390, 391, 402, 403, 413, 435, 466, 738, 740, 742, 744, 746, 750, 752, 754, 757: All `TEST_ROSTER_DIR` -> `TEST_KNOSSOS_DIR`, `$TEST_TMP/roster` -> `$TEST_TMP/knossos`, comment text `roster` -> `knossos`

### 9.6 `tests/sync/test-init.sh`

- Line 3: `roster-sync init tests` -> `knossos-sync init tests`
- Lines 17-18: `ROSTER_HOME` -> `KNOSSOS_HOME`
- Lines 82, 164, 186, 212, 216, 230, 237, 246, 258, 266, 288, 301, 317, 340, 349, 385, 406, 413, 432, 439, 462, 486, 551: `$ROSTER_HOME/roster-sync` -> `$KNOSSOS_HOME/knossos-sync`
- Lines 119-125: `roster.path` -> `knossos.path`, `roster_path` -> `knossos_path`
- Lines 170-174: `matches roster` -> `matches knossos`
- Lines 246, 249: `matches roster` -> `matches knossos`
- Lines 297-298, 596: `inside_roster_fails` -> `inside_knossos_fails`
- Lines 491-525: `roster_path`, `roster_commit`, `roster_ref`, `roster_last_sync` -> `knossos_path`, `knossos_commit`, `knossos_ref`, `knossos_last_sync`; `.roster.path` -> `.knossos.path`, `.roster.commit` -> `.knossos.commit`, etc.
- Line 579: `Running roster-sync init tests` -> `Running knossos-sync init tests`

### 9.7 `tests/sync/test-validate-repair.sh`

- Line 3: `roster-sync validate and repair` -> `knossos-sync validate and repair`
- Lines 17-18: `ROSTER_HOME` -> `KNOSSOS_HOME`
- Lines 65, 70-71: `roster_path` -> `knossos_path`, `"roster":` -> `"knossos":`
- Lines 102: `source: "roster"` -> `source: "knossos"`
- Lines 121, 138, 155, 174, 204, 237, 266, 287, 305, 342, 369, 402, 420, 440: `$ROSTER_HOME/roster-sync` -> `$KNOSSOS_HOME/knossos-sync`
- Lines 221, 226: `roster structure` -> `knossos structure`, `"path": "$ROSTER_HOME"` -> `"path": "$KNOSSOS_HOME"`
- Lines 302-303, 307, 361, 366-368, 371, 384, 400, 442, 462-463: `roster files` -> `knossos files`
- Line 473: `echo "ROSTER_HOME: $ROSTER_HOME"` -> `echo "KNOSSOS_HOME: $KNOSSOS_HOME"`

### 9.8 `tests/sync/test-swap-rite-integration.sh`

- Lines 3, 6-7, 11: comments `roster-sync` -> `knossos-sync`
- Lines 17-18: `ROSTER_HOME` -> `KNOSSOS_HOME`
- Lines 81-82: `"roster":`, `"path": "$ROSTER_HOME"` -> `"knossos":`, `"path": "$KNOSSOS_HOME"`
- Lines 94, 96: `roster commit` -> `knossos commit`, `$ROSTER_HOME` -> `$KNOSSOS_HOME`
- Lines 107, 134, 151, 165, 171, 189, 195, 245, 272, 274, 299, 320: `$ROSTER_HOME/swap-rite.sh` -> `$KNOSSOS_HOME/swap-rite.sh`, `$ROSTER_HOME/roster-sync` -> `$KNOSSOS_HOME/knossos-sync`
- Lines 123-176: All `roster_has_updates` function/test names -> `knossos_has_updates`
- Lines 221-228: `roster_path` -> `knossos_path` in variable and assertions
- Lines 263-285: `roster_sync_available` -> `knossos_sync_available` in function/test names and assertion strings
- Lines 336, 339, 346-348, 351, 862-865: All roster references in summary, env var echoes, and test runner calls

### 9.9 `tests/hooks/test-session-context-preferences.sh`

- Line 79: `ROSTER_PREF_*` -> `KNOSSOS_PREF_*`
- Lines 95-98: All `ROSTER_PREF_AUTONOMY_LEVEL`, `ROSTER_PREF_FAILURE_HANDLING`, `ROSTER_PREF_OUTPUT_FORMAT` -> `KNOSSOS_PREF_*`

### 9.10 `tests/lib/rite/test-rite-hooks-registration.sh`

- Line 12: `ROSTER_HOME` -> `KNOSSOS_HOME`
- Line 15: `$ROSTER_HOME/lib/team/` -> `$KNOSSOS_HOME/lib/team/`
- Lines 117-118, 145: `roster-only`, `roster-hook.sh` -> `knossos-only`, `knossos-hook.sh`
- Lines 193-194: `ROSTER_HOME` -> `KNOSSOS_HOME`
- Lines 324-395: All `extract_non_roster` in function names -> `extract_non_knossos`; `roster-only hooks` -> `knossos-only hooks`; assertion strings updated
- Lines 594, 610, 613, 802: `roster.sh` (command value) -> `knossos.sh`; assertion strings
- Lines 824-827: Test runner calls updated to new function names

### 9.11 `tests/lib/rite/test-rite-transaction.sh`

- Lines 71-78: `mock-roster` -> `mock-knossos`, `ROSTER_HOME` -> `KNOSSOS_HOME`

### 9.12 `tests/integration/test-d002-simple.sh`

- Lines 11, 49, 50, 51: `/Users/tomtenuta/Code/roster` -> `/Users/tomtenuta/Code/knossos`
- Line 10: comment `roster directory` -> `knossos directory`

### 9.13 `tests/integration/test-d002-output-format.sh`

- Lines 24, 28: `/Users/tomtenuta/Code/roster` -> `/Users/tomtenuta/Code/knossos`

### 9.14 `tests/test-rite-context-loader.sh`

- Lines 205, 220-224: `~/Code/roster` -> `~/Code/knossos`; `roster project` -> `knossos project`

### 9.15 `tests/session-fsm/test_helpers.bash`

Check for roster references (none expected based on earlier grep).

---

## 10. Summary Statistics

| Category | Count |
|----------|-------|
| Environment variables renamed | 18 distinct names |
| Functions renamed | 13 production + 14 test |
| Local variables renamed | 19 distinct names |
| JSON manifest keys changed | 5 keys |
| JSON source values changed | 2 (`"roster"`, `"roster-diverged"`) |
| Log prefix changes | 2 patterns (`[roster-sync]`, `[roster-sync DEBUG]`) |
| YAML files updated | 7 |
| Shell files updated (production) | 36 |
| Shell files updated (test) | 15 |
| `.claude/` projected hook files with references | 4 (+1 knowledge script) |
| Total shell files with "roster" references | 47 (+ 5 under .claude/) |
| `knossos-sync` (main entry point) | 149 occurrences |
| Default path changes | `~/Code/roster` -> `~/Code/knossos` in ~15 files |

---

## 11. Design Rationale

**Why CLEAN BREAK**: The codebase already completed a partial rename (commit `bbbc026`), renaming the repo and some Go code. The shell layer retains extensive "roster" references. A shim-based deprecation approach was already tried for `ROSTER_HOME` -> `KNOSSOS_HOME` (see `lib/knossos-home.sh`). Continuing with shims adds maintenance burden without benefit -- no external consumers depend on these internal shell identifiers.

**Why `~/Code/knossos` default**: The repository is being renamed from `roster` to `knossos`. The default path must match the expected directory name. This is the single most impactful breaking change and requires users to either rename their checkout directory or set `KNOSSOS_HOME` explicitly.

**Why rename JSON keys `.roster.*` to `.knossos.*`**: The manifest is an internal data structure. Satellite manifests will be regenerated on next `knossos-sync init` or `knossos-sync repair`. Old manifests with `.roster.*` keys will fail validation, which is the correct behavior for a clean break -- users run repair once.

**Why rename `"source": "roster"` to `"source": "knossos"`**: The user-sync manifests use this value to track provenance. Existing manifests will not match, causing recovery/re-adoption. This is acceptable because the recovery mechanism (`recover_manifest_entries`) is designed to handle exactly this case.

**Why `SYNC_MARKER_ROSTER` -> `SYNC_MARKER_KNOSSOS`**: The marker string `<!-- SYNC: roster-owned -->` appears in satellite CLAUDE.md files. Changing it means existing satellites must update their markers. Since this is a clean break, `knossos-sync repair` or `knossos-sync init --force` handles this.

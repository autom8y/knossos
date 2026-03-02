# Context Design: Stale settings.json Cleanup + .v2-backup Removal

**Date**: 2026-03-02
**Author**: Context Architect (ecosystem rite)
**Status**: READY FOR IMPLEMENTATION
**Upstream**: `docs/ecosystem/GAP-legacy-backward-compat-cleanup.md` (Sections 1D, 1E, 2C, 5)

---

## Problem Statement

Two categories of stale files persist in satellites and user-level directories after their generating code was removed:

1. **Stale `.claude/settings.json`** in project directories. Created by the deleted `writeDefaultSettings()` function (formerly in `internal/cmd/initialize/init.go`). Contains either a blanket-deny agent-guard hook (the original template) or an empty CC-default stub. Both variants are harmful: the blanket deny blocks all agent writes; the empty stub occupies a file that CC loads alongside `settings.local.json`, creating confusion about hook source-of-truth.

2. **`*.v2-backup` manifest files** in `~/.claude/`. Created by `cleanupOldManifests()` in `internal/materialize/userscope/sync.go:1486-1504` during the v1-to-v2 manifest migration. The migration is complete. Five backup files (`USER_AGENT_MANIFEST.json.v2-backup`, `USER_MENA_MANIFEST.json.v2-backup`, `USER_HOOKS_MANIFEST.json.v2-backup`, `USER_COMMAND_MANIFEST.json.v2-backup`, `USER_SKILL_MANIFEST.json.v2-backup`) serve no rollback purpose.

---

## Options Considered

### A. cleanupStaleBlanketSettings()

#### Option A1: Exact JSON string comparison

**Approach**: Compare file bytes against the known template string from `writeDefaultSettings()`.

**Pros**: No parsing, fast, zero false positives for the exact template.
**Cons**: Brittle to any whitespace variation. CC may reformat the file on write (e.g., trailing newline, indentation). Would miss the "permissions" stub variant found in the satellite audit. Two hardcoded strings to maintain.

**Verdict**: Rejected. Byte-level comparison is too brittle for JSON content that may have been reformatted by editors, git, or CC itself.

#### Option A2: Structural comparison (unmarshal and check fields)

**Approach**: Parse `settings.json` as JSON. Check if it matches one of two known stale fingerprints by field structure.

**Pros**: Whitespace-insensitive. Handles reformatting. Detectable patterns are well-defined.
**Cons**: Requires defining structural fingerprints precisely. Must handle the case where a user added fields to a stale base.

**Verdict**: Selected as part of the combined approach (A4).

#### Option A3: Provenance-only detection

**Approach**: If `settings.json` has no provenance entry, delete it.

**Pros**: Simple. Leverages existing infrastructure.
**Cons**: Dangerous. A user may have manually created `settings.json` for legitimate CC configuration. Provenance absence alone does not prove the file is stale -- it may predate provenance. Would delete user-created settings without checking content.

**Verdict**: Rejected as sole criterion. Provenance absence is necessary but not sufficient.

#### Option A4 (Selected): Combined provenance gate + structural fingerprint

**Approach**: Two-gate detection. Gate 1: `settings.json` has no provenance entry in `PROVENANCE_MANIFEST.yaml` (meaning no pipeline stage claims ownership). Gate 2: The parsed JSON structurally matches one of the known stale fingerprints. Both gates must pass for deletion.

**Pros**: Maximum safety. Provenance gate ensures we never touch pipeline-managed files. Structural fingerprint ensures we never touch user-customized files. Whitespace-insensitive.
**Cons**: Slightly more code than alternatives. Must define fingerprints precisely.

**Verdict**: Selected. The dual-gate approach provides defense-in-depth: provenance prevents touching managed files, fingerprint prevents touching user-customized files.

**Rationale for rejecting other options**: A1 is too brittle for real-world JSON. A3 is too aggressive (deletes user files). A2 alone works but lacks the provenance safety net. A4 combines the strengths of A2 and A3.

### B. .v2-backup Cleanup

#### Option B1: Cleanup in `provenance.LoadOrBootstrap()`

**Approach**: Add backup deletion to the bootstrap path of provenance loading.

**Pros**: Runs on every sync. Central location.
**Cons**: Pollutes provenance loading with filesystem cleanup that has nothing to do with provenance. Violates single-responsibility. `LoadOrBootstrap()` is a pure read function; adding side effects breaks its contract.

**Verdict**: Rejected. Provenance loading should not have cleanup side effects.

#### Option B2: Cleanup in `cleanupOldManifests()`

**Approach**: Extend the existing `cleanupOldManifests()` function to also delete the `.v2-backup` files it previously created.

**Pros**: Cleanup code is co-located with the code that created the backups. Single function handles the full lifecycle. Already called at the right point in the user-scope sync pipeline (after manifest save, line 149 of `sync.go`).
**Cons**: The function name becomes slightly misleading (it now removes backups AND manifests), but a doc comment update resolves this.

**Verdict**: Selected. Co-location with the creation code is the strongest organizational signal. The function already runs at the correct pipeline position.

#### Option B3: New standalone function in materialize pipeline

**Approach**: Create a separate `cleanupV2Backups()` function called from the rite-scope sync.

**Pros**: Clean separation.
**Cons**: .v2-backup files live in `~/.claude/` (user scope), not `.claude/` (rite scope). Calling a user-scope cleanup from the rite-scope pipeline crosses scope boundaries. The user-scope pipeline already has the call site.

**Verdict**: Rejected. Wrong scope. The backups are user-level files and the cleanup belongs in the user-scope pipeline where they were created.

---

## Detailed Design

### Component A: cleanupStaleBlanketSettings()

#### File Placement

**File**: `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go`

**Rationale**: The function operates on the project-level `.claude/settings.json`, which is rite-scope territory. It uses the provenance manifest already loaded in `MaterializeWithOptions()`. Adding to the existing file keeps the cleanup co-located with other rite-scope pipeline stages (orphan detection, throughline cleanup, invocation state cleanup).

#### Function Signature

```go
// cleanupStaleBlanketSettings removes .claude/settings.json if it matches a known
// stale fingerprint from the deleted writeDefaultSettings() function AND has no
// provenance entry (indicating it was not created by the current pipeline).
//
// Two stale fingerprints are recognized:
//   1. Blanket-deny agent-guard hook (no --allow-path flags)
//   2. Empty CC-default stub (permissions + empty hooks)
//
// Both gates must pass: no provenance entry AND structural fingerprint match.
// Returns true if the file was removed, false otherwise.
func (m *Materializer) cleanupStaleBlanketSettings(claudeDir string, manifest *provenance.ProvenanceManifest) bool
```

#### Detection Algorithm

```
FUNCTION cleanupStaleBlanketSettings(claudeDir, provenanceManifest):
    settingsPath = claudeDir + "/settings.json"

    // Gate 1: File must exist
    IF NOT fileExists(settingsPath):
        RETURN false

    // Gate 2: No provenance entry for "settings.json"
    // If the pipeline tracks this file, it is managed -- do not touch.
    IF provenanceManifest.Entries["settings.json"] EXISTS:
        RETURN false

    // Gate 3: Parse JSON and check structural fingerprint
    content = readFile(settingsPath)
    parsed = parseJSON(content)
    IF parseError:
        // Not valid JSON -- could be user content, leave it alone
        RETURN false

    IF NOT matchesStalFingerprint(parsed):
        RETURN false

    // Both gates passed: safe to remove
    removeFile(settingsPath)
    log.Printf("Removed stale settings.json from %s", claudeDir)
    RETURN true

FUNCTION matchesStaleFingerprint(parsed map[string]any):
    // Fingerprint 1: Blanket-deny agent-guard hook
    // Structure: {"hooks": {"PreToolUse": [{"matcher": "...", "hooks": [{"type": "command", "command": "ari hook agent-guard ..."}]}]}}
    // Key signal: the command contains "ari hook agent-guard" with NO "--allow-path"
    IF hooks = parsed["hooks"]; hooks is map:
        IF preToolUse = hooks["PreToolUse"]; preToolUse is array:
            IF len(parsed) == 1:  // Only "hooks" key at top level
                FOR entry IN preToolUse:
                    IF innerHooks = entry["hooks"]; innerHooks is array:
                        FOR hook IN innerHooks:
                            IF cmd = hook["command"]; cmd is string:
                                IF contains(cmd, "ari hook agent-guard") AND NOT contains(cmd, "--allow-path"):
                                    RETURN true

    // Fingerprint 2: Empty CC-default stub
    // Structure: {"permissions": {"allow": [], "additionalDirectories": []}, "hooks": {}}
    IF permissions = parsed["permissions"]; permissions is map:
        IF hooks = parsed["hooks"]; hooks is map:
            IF len(parsed) == 2 AND len(hooks) == 0:
                allow, hasAllow = permissions["allow"]
                addDirs, hasAddDirs = permissions["additionalDirectories"]
                IF hasAllow AND hasAddDirs:
                    IF isEmptyArray(allow) AND isEmptyArray(addDirs) AND len(permissions) == 2:
                        RETURN true

    RETURN false
```

#### Integration Point

**Call site**: Inside `MaterializeWithOptions()` at `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go`, immediately after the provenance manifest is loaded and divergence is detected (after line 374), before any materialization stages run. This placement ensures:

1. The provenance manifest is available for the gate check.
2. Cleanup happens before `materializeSettingsWithManifest()` writes `settings.local.json`, preventing any confusion about which settings file is active.
3. The function is called early in the pipeline, similar to `clearInvocationState()` at step 2.5.

Specifically, insert the call between step 2.5 (clearInvocationState) and step 3 (orphan detection):

```
// Current pipeline order (materialize.go MaterializeWithOptions):
//   1. Resolve rite source
//   2. Ensure .claude/ directory
//   2.5. Clear invocation state
//   --> NEW: 2.6. Cleanup stale blanket settings.json
//   3. Handle orphans
//   4. Generate agents/
//   5. Generate commands/ and skills/
//   ...
//   8. Generate settings.local.json
//   ...
//   Provenance: save manifest
```

The same call should also be added to `MaterializeMinimal()` (line 227) at the equivalent position, since minimal-mode satellites may also have stale `settings.json` from a previous `ari init` invocation.

#### Edge Cases

| Case | Behavior | Rationale |
|------|----------|-----------|
| `settings.json` does not exist | No-op, return false | Nothing to clean |
| `settings.json` has provenance entry | No-op, return false | Pipeline manages this file; do not interfere |
| `settings.json` has user-added custom hooks | No-op, return false | Fingerprint check fails (extra keys present) |
| `settings.json` is not valid JSON | No-op, return false | Parse error = leave it alone |
| `settings.json` matches fingerprint 1 (agent-guard) | Remove | Known stale content from writeDefaultSettings() |
| `settings.json` matches fingerprint 2 (empty stub) | Remove | Known stale CC-default content |
| `settings.json` matches fingerprint but HAS provenance | No-op, return false | Provenance gate prevents removal of tracked files |
| Dry-run mode | Function not called | Cleanup is gated behind `!opts.DryRun` at the call site |

### Component B: .v2-backup Cleanup

#### File Placement

**File**: `/Users/tomtenuta/Code/knossos/internal/materialize/userscope/sync.go`

**Rationale**: The `.v2-backup` files were created by `cleanupOldManifests()` in this exact file. Extending that function to also remove its own backup artifacts keeps the lifecycle in one place.

#### Function Change

Replace the existing function body entirely. The backup-creation logic (`os.ReadFile` + `os.WriteFile`) is removed because backups are no longer needed. Both the original manifest and its `.v2-backup` remnant are unconditionally removed (best-effort, errors ignored for missing files via `os.Remove` semantics).

```go
// cleanupOldManifests removes legacy JSON manifest files and their v2-backup remnants.
// The v1 JSON manifests were superseded by USER_PROVENANCE_MANIFEST.yaml.
// The .v2-backup files were created by this function during v1-to-v2 migration
// and serve no rollback purpose now that migration is complete.
func cleanupOldManifests(userClaudeDir string) {
    oldManifests := []string{
        filepath.Join(userClaudeDir, "USER_AGENT_MANIFEST.json"),
        filepath.Join(userClaudeDir, "USER_MENA_MANIFEST.json"),
        filepath.Join(userClaudeDir, "USER_HOOKS_MANIFEST.json"),
        filepath.Join(userClaudeDir, "USER_COMMAND_MANIFEST.json"),
        filepath.Join(userClaudeDir, "USER_SKILL_MANIFEST.json"),
    }
    for _, path := range oldManifests {
        // Remove the original JSON manifest if still present.
        // Skip backup creation -- migration is complete, backups serve no purpose.
        os.Remove(path)

        // Remove .v2-backup remnants from previous migration runs.
        os.Remove(path + ".v2-backup")
    }
}
```

#### Integration Point

No change to call site. `cleanupOldManifests()` is already called at `sync.go:149`, after the provenance manifest save, during every non-dry-run user-scope sync.

---

## Backward Compatibility Classification: COMPATIBLE

Both changes are strictly subtractive (removing files). No schema changes. No new fields. No breaking API changes.

| Change | Classification | Rationale |
|--------|---------------|-----------|
| Stale `settings.json` removal | COMPATIBLE | File was never produced by the current pipeline. Removal only affects satellites that have the exact stale content. User-modified files are preserved by the dual-gate check. |
| `.v2-backup` removal | COMPATIBLE | Backup files have no consumers. No code reads them. No rollback mechanism references them. |
| `cleanupOldManifests()` simplification | COMPATIBLE | Function behavior is strictly reduced (no longer creates new backups). The backup-before-delete pattern was only needed during the migration window, which is closed. |

**Migration path**: None required. Changes are self-deploying via `ari sync`. After merge, the next sync on any satellite triggers automatic cleanup.

---

## Settings Merge Algorithm Impact

**None.** This design does not modify how `settings.local.json` is generated or merged. The `materializeSettingsWithManifest()` function at `materialize.go:1454` is unchanged. The cleanup targets `settings.json` (the project-level CC settings file), which is a completely separate file from `settings.local.json` (the pipeline-generated file).

CC loads hooks from both files. Removing stale `settings.json` means CC will only load hooks from `settings.local.json`, which is the intended behavior.

---

## Integration Tests

### Test Matrix: cleanupStaleBlanketSettings()

| Test Name | Satellite Type | Setup | Expected Outcome |
|-----------|---------------|-------|-----------------|
| `TestCleanupStaleSettings_NoFile` | Baseline | No `settings.json` exists | No-op, returns false |
| `TestCleanupStaleSettings_AgentGuardFingerprint` | Legacy (pre-844c686) | `settings.json` with blanket-deny agent-guard hook, no provenance entry | File removed, returns true |
| `TestCleanupStaleSettings_EmptyStubFingerprint` | Legacy | `settings.json` with `{"permissions":{"allow":[],"additionalDirectories":[]},"hooks":{}}`, no provenance entry | File removed, returns true |
| `TestCleanupStaleSettings_UserModified` | Complex | `settings.json` with agent-guard + additional custom hooks | No-op, returns false (fingerprint mismatch: extra keys) |
| `TestCleanupStaleSettings_HasProvenance` | Standard | `settings.json` matches fingerprint BUT has provenance entry | No-op, returns false (provenance gate blocks) |
| `TestCleanupStaleSettings_InvalidJSON` | Edge | `settings.json` contains non-JSON content | No-op, returns false |
| `TestCleanupStaleSettings_WhitespaceVariant` | Edge | `settings.json` with reformatted whitespace of fingerprint 1 | File removed, returns true (structural comparison is whitespace-insensitive) |
| `TestCleanupStaleSettings_PartialMatch` | Edge | `settings.json` with agent-guard hook that includes `--allow-path` | No-op, returns false (not the stale pattern) |

### Test Matrix: .v2-backup Cleanup

| Test Name | Setup | Expected Outcome |
|-----------|-------|-----------------|
| `TestCleanupOldManifests_RemovesBackups` | Create 5 `.v2-backup` files in temp dir | All 5 `.v2-backup` files removed |
| `TestCleanupOldManifests_RemovesOriginals` | Create original JSON manifests (no backups) | Original files removed, no backups created |
| `TestCleanupOldManifests_BothPresent` | Create both originals and `.v2-backup` files | Both removed |
| `TestCleanupOldManifests_NonePresent` | Empty directory | No errors, no-op |
| `TestCleanupOldManifests_PartialPresence` | Only 2 of 5 backup files exist | Those 2 removed, others no-op |

### Test File Placement

| Component | Test File |
|-----------|-----------|
| `cleanupStaleBlanketSettings()` | `/Users/tomtenuta/Code/knossos/internal/materialize/stale_settings_cleanup_test.go` |
| `cleanupOldManifests()` changes | `/Users/tomtenuta/Code/knossos/internal/materialize/userscope/sync_test.go` (if exists, otherwise new `cleanup_test.go` in same package) |

---

## File-Level Change Specification

### Files Modified

| File | Function | Change |
|------|----------|--------|
| `internal/materialize/materialize.go` | NEW: `cleanupStaleBlanketSettings()` | Add ~50-line method on `Materializer`. Two-gate detection: provenance absence + structural fingerprint match. |
| `internal/materialize/materialize.go` | NEW: `matchesStaleSettingsFingerprint()` | Add ~40-line helper. Unmarshals JSON and checks against two known stale structures. |
| `internal/materialize/materialize.go` | `MaterializeWithOptions()` | Add call to `cleanupStaleBlanketSettings()` after line 374 (after provenance load, before orphan detection). Approximately: `m.cleanupStaleBlanketSettings(claudeDir, prevManifest)` |
| `internal/materialize/materialize.go` | `MaterializeMinimal()` | Add same call after line 255 (after provenance load, before rules materialization). |
| `internal/materialize/userscope/sync.go` | `cleanupOldManifests()` | Replace function body: remove backup-creation logic, add `.v2-backup` deletion. Net change: ~10 lines shorter. |

### Files Created

| File | Purpose |
|------|---------|
| `internal/materialize/stale_settings_cleanup_test.go` | Unit tests for `cleanupStaleBlanketSettings()` and `matchesStaleSettingsFingerprint()` |

### Files NOT Modified

| File | Reason |
|------|--------|
| `internal/provenance/manifest.go` | No changes to provenance loading. `LoadOrBootstrap()` stays pure. |
| `internal/provenance/provenance.go` | No schema changes. |
| `config/hooks.yaml` | Read-only reference; no changes to hook configuration. |
| `internal/materialize/userscope/sync.go` (beyond `cleanupOldManifests`) | No other user-scope changes needed. |

---

## Risk Mitigations

### Risk 1: False positive deletion of user-created settings.json

**Mitigation**: Dual-gate detection. A user-created `settings.json` would fail at least one gate:
- If they created it manually, it has no provenance entry BUT would not match the exact stale fingerprint (they would have added their own hooks/permissions).
- If the pipeline created it, it would have a provenance entry, blocking deletion.
- If somehow both gates pass (user manually created the exact same stale content), the file is functionally harmful anyway (blanket deny) and removing it is correct.

### Risk 2: Race condition with CC file watcher

**Mitigation**: `os.Remove()` is atomic on all supported platforms (macOS, Linux). The file disappears in one syscall. CC's file watcher may fire, but the file will simply be gone. No partial state. `settings.local.json` continues to serve as the hook source.

### Risk 3: Satellite has both settings.json and settings.local.json with conflicting hooks

**Mitigation**: This is the exact problem the cleanup solves. After cleanup, only `settings.local.json` remains. CC loads hooks from one source. No conflict.

### Risk 4: .v2-backup files are needed for rollback

**Mitigation**: The v1-to-v2 manifest migration has been complete since ADR-0026 Phase 4b. The v2 manifest format (`USER_PROVENANCE_MANIFEST.yaml`) has been the sole active format for months. No rollback path exists or is needed. The original v1 JSON manifests are also deleted (they were the source of the backups, and the backups were safety copies of already-deleted files).

---

## Artifact Attestation

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This design document | `/Users/tomtenuta/Code/knossos/docs/ecosystem/DESIGN-stale-settings-cleanup.md` | Written |
| Gap analysis (upstream) | `/Users/tomtenuta/Code/knossos/docs/ecosystem/GAP-legacy-backward-compat-cleanup.md` | Read |
| Spike (upstream) | `/Users/tomtenuta/Code/knossos/docs/spikes/SPIKE-provenance-backward-compat-strategy.md` | Read |
| Spike (agent-guard) | `/Users/tomtenuta/Code/knossos/docs/spikes/SPIKE-agent-guard-ledge-path-blockage.md` | Read |
| materialize.go | `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go` | Read |
| userscope/sync.go | `/Users/tomtenuta/Code/knossos/internal/materialize/userscope/sync.go` | Read |
| provenance manifest.go | `/Users/tomtenuta/Code/knossos/internal/provenance/manifest.go` | Read |
| provenance provenance.go | `/Users/tomtenuta/Code/knossos/internal/provenance/provenance.go` | Read |
| hooks.yaml | `/Users/tomtenuta/Code/knossos/config/hooks.yaml` | Read |

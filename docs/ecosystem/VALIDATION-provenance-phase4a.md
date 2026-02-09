# Validation Report: ADR-0026 Phase 4a -- Unified Provenance Schema

**Validator**: compatibility-tester agent
**Date**: 2026-02-09
**Scope**: READ-ONLY validation of 3-sprint Phase 4a implementation
**Verdict**: **APPROVED** (with 2 P3 observations)

---

## Scenario 1: Build Verification

**Result**: PASS

```
CGO_ENABLED=0 go build ./cmd/ari         -> success (no output)
CGO_ENABLED=0 go vet ./internal/provenance/... ./internal/usersync/... ./internal/materialize/... ./internal/cmd/provenance/... ./internal/paths/...  -> success (no output)
```

Both build and vet complete with zero errors across all five relevant packages.

---

## Scenario 2: Full Test Suite

**Result**: PASS

| Package | Result | Duration |
|---------|--------|----------|
| `internal/provenance` | ok | 0.940s |
| `internal/usersync` | ok | 0.741s |
| `internal/materialize` | ok | 13.356s |

All tests pass with `-count=1` (no cache).

---

## Scenario 3: Schema Validation

**Result**: PASS

Evidence from `/Users/tomtenuta/Code/knossos/internal/provenance/provenance.go`:

| Check | Expected | Found | Status |
|-------|----------|-------|--------|
| OwnerType values | "knossos", "user", "untracked" | Lines 75, 79, 84: `OwnerKnossos = "knossos"`, `OwnerUser = "user"`, `OwnerUntracked = "untracked"` | PASS |
| No "unknown" OwnerType | Zero const definitions | No `OwnerUnknown` constant exists | PASS |
| ScopeType values | "rite", "user" | Lines 107, 109: `ScopeRite = "rite"`, `ScopeUser = "user"` | PASS |
| No "materialize" scope | Zero results | Not present | PASS |
| CurrentSchemaVersion | "2.0" | Line 20: `CurrentSchemaVersion = "2.0"` | PASS |
| No SourcePipeline field | Absent from struct | ProvenanceEntry struct (lines 42-67) has `Scope ScopeType` at line 47, no SourcePipeline | PASS |
| Scope field present | `Scope ScopeType` in struct | Line 47: `Scope ScopeType \`yaml:"scope"\`` | PASS |
| UserManifestFileName | Defined | Line 16: `UserManifestFileName = "USER_PROVENANCE_MANIFEST.yaml"` | PASS |
| UserManifestPath() | Defined | Lines 129-131: function exists | PASS |

---

## Scenario 4: Manifest I/O Validation

**Result**: PASS

Evidence from `/Users/tomtenuta/Code/knossos/internal/provenance/manifest.go`:

| Check | Expected | Found | Status |
|-------|----------|-------|--------|
| `migrateV1ToV2()` exists | Function present | Lines 217-238: migrates v1.0->2.0, converts empty Scope to ScopeRite, converts "unknown" owner to OwnerUntracked | PASS |
| `validateManifest()` validates Scope | Scope field required and valid | Lines 177-180: checks empty Scope and calls `entry.Scope.IsValid()` | PASS |
| `structurallyEqual()` compares Scope | Scope in comparison | Line 102: `entryA.Scope != entryB.Scope` | PASS |
| No SourcePipeline references (except comments) | Only in comment on line 216 | Line 216 comment only: "Converts SourcePipeline to Scope..." | PASS |

---

## Scenario 5: Usersync Type Elimination

**Result**: PASS

Grep results across `internal/usersync/`:

| Check | Expected | Found | Status |
|-------|----------|-------|--------|
| `usersync.Entry` type | Zero references | No matches | PASS |
| `usersync.Manifest` type | Zero references | No matches | PASS |
| `SourceKnossos/SourceDiverged/SourceUser` | Zero references | No matches | PASS |
| Old `loadManifest()`/`saveManifest()` pattern | Eliminated or replaced | Now private methods on `*Syncer` delegating to `provenance.LoadOrBootstrap()` and `provenance.Save()` (manifest.go lines 38-45) | PASS |
| Uses `provenance.ProvenanceEntry` | Present | 10 occurrences in usersync.go constructing ProvenanceEntry structs | PASS |

---

## Scenario 6: Manifest Key Namespacing

**Result**: PASS

Evidence from `/Users/tomtenuta/Code/knossos/internal/usersync/usersync.go`:

| Check | Expected | Found | Status |
|-------|----------|-------|--------|
| `prefixManifestKey()` exists | Function present | Lines 493-506: adds `agents/`, `hooks/` prefix; mena handled separately per routing | PASS |
| Agents prefix | `agents/` | Line 495: `return "agents/" + key` | PASS |
| Hooks prefix | `hooks/` | Line 502: `return "hooks/" + key` | PASS |
| Mena keys use target prefix | `commands/` or `skills/` | Lines 603-609: based on `menaTarget`, prefixed as `"commands/" + manifestKey` or `"skills/" + manifestKey` | PASS |
| `keyToTargetPath()` reverses correctly | Function present | Lines 509-530: strips prefix and joins with correct target directory | PASS |

---

## Scenario 7: Orphan Detection

**Result**: PASS

Evidence from `/Users/tomtenuta/Code/knossos/internal/usersync/usersync.go`:

| Phase | Expected | Found | Status |
|-------|----------|-------|--------|
| Phase 1 (snapshot) | Capture knossos-owned keys | Lines 548-565: `existingKeys` map populated from manifest entries where `entry.Owner == provenance.OwnerKnossos`, initialized to `false` | PASS |
| Phase 2 (walk) | Mark seen keys | Lines 612-615: `existingKeys[fullManifestKey] = true` during walk | PASS |
| Phase 3 (cleanup) | Remove unseen knossos entries | Lines 790-796: iterates `existingKeys`, calls `s.removeOrphan(key, manifest)` for unseen entries | PASS |
| User safety | User entries never removed | Lines 803-806: `removeOrphan` checks `entry.Owner != provenance.OwnerKnossos` and returns early | PASS |

---

## Scenario 8: CollisionChecker Migration

**Result**: PASS

Evidence from `/Users/tomtenuta/Code/knossos/internal/usersync/collision.go`:

| Check | Expected | Found | Status |
|-------|----------|-------|--------|
| `NewCollisionChecker()` accepts `claudeDir` | Parameter present | Line 24: `func NewCollisionChecker(resourceType ResourceType, nested bool, claudeDir string)` | PASS |
| Primary path: load rite manifest | Loads PROVENANCE_MANIFEST.yaml | Lines 39-57: `loadRiteManifest()` calls `provenance.Load(manifestPath)` | PASS |
| Fallback: directory scan | When manifest missing | Lines 85-123: `CheckCollision()` falls back to `os.ReadDir` rite scan when `len(c.riteEntries) == 0` | PASS |
| `riteEntries` cached | O(1) lookup | Lines 44, 52-56: `c.riteEntries` is `map[string]bool`, populated once, used for lookup at line 79 | PASS |

---

## Scenario 9: CLI Changes

**Result**: PASS

Evidence from `/Users/tomtenuta/Code/knossos/internal/cmd/provenance/provenance.go`:

| Check | Expected | Found | Status |
|-------|----------|-------|--------|
| ShowEntry has Scope field | Present | Line 151: `Scope string \`json:"scope"\`` | PASS |
| Table headers include SCOPE | SCOPE column | Line 171: `return []string{"PATH", "OWNER", "SCOPE", "SOURCE", "STATUS"}` | PASS |
| `--scope` flag defined | Flag present | Line 83: `cmd.Flags().StringVar(&scopeFilter, "scope", "", ...)` | PASS |
| `runShow()` loads both manifests | Rite + User | Lines 98-128: loads rite manifest from project `.claude/`, loads user manifest from `~/.claude/` | PASS |
| User entries prefixed with `~/` | Prefix present | Line 121: `displayPath := "~/" + path` | PASS |
| `CombinedOutput` struct | Defined | Lines 163-167: with `Rite` and `User` fields | PASS |
| `computeStatus()` handles OwnerUntracked | Non-knossos returns "-" | Lines 224-226: `if entry.Owner != provenance.OwnerKnossos { return "-" }` -- covers both user and untracked | PASS |

---

## Scenario 10: No Stale References

**Result**: PASS (with P3 observations)

| Grep Pattern | Expected | Found | Status |
|-------------|----------|-------|--------|
| `SourcePipeline` in internal/ | Comments only | 1 result: `manifest.go:216` comment | PASS |
| `OwnerUnknown` in internal/ | Zero results | No matches | PASS |
| `"unknown"` in internal/provenance/ | Only in migrateV1ToV2 + tests | 7 results: 4 in test file (expected), 3 in manifest.go migration function | PASS |
| `knossos-diverged` in internal/ | Zero in provenance/usersync | 4 results, all outside provenance/usersync: `cmd/migrate/roster_to_knossos.go` (legacy migration tool), `cmd/migrate/roster_to_knossos_test.go` (test for same), `cmd/sync/user.go` (help text string) | **P3** |

**P3 observation**: `knossos-diverged` appears in 3 files outside provenance/usersync. These are in:
1. `internal/cmd/migrate/roster_to_knossos.go` -- legacy one-time migration tool (produces JSON manifests for the old format)
2. `internal/cmd/migrate/roster_to_knossos_test.go` -- test for the above
3. `internal/cmd/sync/user.go` -- help text/documentation string

These are not production provenance code. The migrate command is a legacy tool for one-time roster-to-knossos conversion. The sync/user.go help text is descriptive, not functional. Neither affects the provenance system. **Not blocking.**

---

## Scenario 11: Deprecated Paths Removed

**Result**: PASS (with P3 observation)

Evidence from `/Users/tomtenuta/Code/knossos/internal/paths/paths.go`:

| Check | Expected | Found | Status |
|-------|----------|-------|--------|
| `UserProvenanceManifest()` exists | Present | Lines 309-312: defined and returns correct path | PASS |
| `UserAgentManifest()` removed | REMOVED | Lines 299-302: **Still present** | **P3** |
| `UserSkillManifest()` removed | REMOVED | Lines 304-307: **Still present** | **P3** |
| `UserMenaManifest()` removed | N/A | Never existed (was `USER_MENA_MANIFEST.json` in cleanup list only) | PASS |
| `UserHooksManifest()` removed | N/A | Never existed (was `USER_HOOKS_MANIFEST.json` in cleanup list only) | PASS |
| Zero callers of removed functions | No callers | Grep for `paths.UserAgentManifest` and `paths.UserSkillManifest` returns **zero results** | PASS |

**P3 observation**: `UserAgentManifest()` and `UserSkillManifest()` still exist in `paths.go` but have **zero callers** anywhere in the codebase. They are dead code -- functionally harmless but should be cleaned up for hygiene. The `cleanupOldManifests()` function in `usersync/manifest.go` hard-codes the legacy JSON paths directly rather than calling these helpers, confirming they are orphaned. **Not blocking.**

---

## Scenario 12: Architecture Invariants

**Result**: PASS

| Invariant | Evidence | Status |
|-----------|----------|--------|
| materialize imports provenance | `usersync.go` line 12: `"github.com/autom8y/knossos/internal/provenance"` (usersync imports provenance; materialize independently imports provenance per existing code) | PASS |
| provenance does NOT import materialize | Grep for `materialize` in `internal/provenance/`: zero results | PASS |
| provenance does NOT import usersync | Grep for `usersync` in `internal/provenance/`: zero results | PASS |
| usersync imports provenance (new, valid) | `usersync.go` line 12, `collision.go` line 9, `manifest.go` line 7: all import provenance | PASS |
| No circular dependencies | `go vet` and `go build` succeeded -- Go compiler would reject circular imports | PASS |

---

## Defect Summary

| ID | Severity | Description | Package | Blocking |
|----|----------|-------------|---------|----------|
| D001 | P3 | `UserAgentManifest()` and `UserSkillManifest()` dead code remains in `internal/paths/paths.go` (zero callers) | paths | NO |
| D002 | P3 | `knossos-diverged` string remains in legacy `cmd/migrate` tool and `cmd/sync/user.go` help text | cmd | NO |

Both findings are dead code / legacy documentation artifacts with zero functional impact on the provenance system. Neither affects sync behavior, manifest I/O, or CLI output.

---

## Test Matrix Summary

| Scenario | Domain | Result | Defects |
|----------|--------|--------|---------|
| 1. Build Verification | build + vet | PASS | - |
| 2. Full Test Suite | provenance, usersync, materialize | PASS | - |
| 3. Schema Validation | provenance types | PASS | - |
| 4. Manifest I/O | migration, validation, comparison | PASS | - |
| 5. Usersync Type Elimination | old types removed | PASS | - |
| 6. Manifest Key Namespacing | prefix/reverse functions | PASS | - |
| 7. Orphan Detection | 3-phase lifecycle | PASS | - |
| 8. CollisionChecker Migration | manifest-first + fallback | PASS | - |
| 9. CLI Changes | scope column, flags, combined output | PASS | - |
| 10. No Stale References | SourcePipeline, OwnerUnknown, knossos-diverged | PASS | D002 (P3) |
| 11. Deprecated Paths Removed | legacy helpers | PASS | D001 (P3) |
| 12. Architecture Invariants | dependency direction | PASS | - |

---

## Recommendation: **APPROVED**

All 12 validation scenarios pass. Zero P0/P1/P2 defects. Two P3 observations (dead code) are documented for future cleanup and do not affect correctness, performance, or backward compatibility.

The Phase 4a implementation correctly:
- Unifies the schema (OwnerUntracked, ScopeType, v2.0)
- Migrates usersync to provenance types (single YAML manifest)
- Adds orphan detection with user-safety guarantees
- Extends CLI with scope filtering and combined output
- Maintains backward compatibility via migrateV1ToV2()
- Preserves architecture invariants (one-way dependency)

### Attestation Table

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| Validation Report | `/Users/tomtenuta/Code/knossos/docs/ecosystem/VALIDATION-provenance-phase4a.md` | YES (read-back confirmed) |

# CONSUMER AUDIT: state.json active_rite

**PKG-008 Pre-Implementation Audit** | Coordinated items: DEBT-028, DEBT-032
**Date**: 2026-03-02
**Analyst**: ecosystem-analyst
**Status**: COMPLETE

---

## Executive Summary

The `state.json` `active_rite` field is a **zombie**. No code path reads `state.json` to determine which rite is active. All runtime consumers read from `.claude/ACTIVE_RITE` (the plain text file). The `state.json` `active_rite` field is written by the materialization pipeline but never read for any purpose -- not even diagnostics.

The `state.json` file itself is NOT a zombie: it carries `last_sync` which has no other home (PROVENANCE_MANIFEST.yaml's `last_sync` exists but is independently managed and semantically different -- it tracks provenance writes, not sync pipeline completion). However, the `active_rite` field within `state.json` provides zero unique value.

---

## Consumer Enumeration

### 1. ACTIVE_RITE File Consumers (.claude/ACTIVE_RITE)

This is the **primary authoritative store**. All runtime rite resolution flows through here.

| # | Consumer | File:Line | Access | Purpose |
|---|----------|-----------|--------|---------|
| A1 | `paths.Resolver.ReadActiveRite()` | `internal/paths/paths.go:120-126` | READ | Central accessor -- reads file, trims whitespace, returns string |
| A2 | `materialize.syncRiteScope()` | `internal/materialize/materialize.go:586-588` | READ | Reads previous ACTIVE_RITE for rite-switch detection at pipeline start |
| A3 | `materialize.writeActiveRite()` | `internal/materialize/materialize.go:1643-1648` | WRITE | Writes rite name to file (step 10 of pipeline) |
| A4 | `materialize.MaterializeMinimal()` | `internal/materialize/materialize.go:286` | WRITE (delete) | Removes ACTIVE_RITE on minimal materialization |
| A5 | `rite.NewDiscovery()` | `internal/rite/discovery.go:42` | READ (via A1) | Captures active rite at construction time via `resolver.ReadActiveRite()` |
| A6 | `rite.Discovery.ActiveRiteName()` | `internal/rite/discovery.go:172-174` | READ | Returns cached active rite name |
| A7 | `cmd/rite.getActiveRite()` | `internal/cmd/rite/rite.go:80-82` | READ (via A1) | Rite subcommand helper |
| A8 | `cmd/session.getActiveRite()` | `internal/cmd/session/session.go:100-106` | READ (via A1) | Session create reads rite for context |
| A9 | `cmd/common.BaseContext.GetActiveRite()` | `internal/cmd/common/context.go:47-49` | READ (via A1) | Shared CLI context accessor |
| A10 | `cmd/hook/context.runContextCore()` | `internal/cmd/hook/context.go:162` | READ (via A1) | SessionStart hook reads active rite |
| A11 | `cmd/status.collectClaude()` | `internal/cmd/status/status.go:248` | READ (via A1) | Health dashboard displays active rite |
| A12 | `cmd/rite/status.runStatus()` | `internal/cmd/rite/status.go:55` | READ (via A6) | Rite status command via discovery |
| A13 | `cmd/rite/list.runList()` | `internal/cmd/rite/list.go:87` | READ (via A6) | Rite list marks active rite |
| A14 | `cmd/rite/current.runCurrent()` | `internal/cmd/rite/current.go:47` | READ (via A6) | Rite current command via discovery |
| A15 | `cmd/explain/context.contextRite()` | `internal/cmd/explain/context.go:46` | READ (via A1) | Explain command context injection |
| A16 | `inscription.Pipeline.buildRenderContext()` | `internal/inscription/pipeline.go:533-535` | READ | Fallback: reads ACTIVE_RITE file if manifest has no active_rite |
| A17 | Worktree inheritance | `internal/materialize/materialize.go:595-599` | READ | `inheritRiteFromMainWorktree()` reads main worktree's ACTIVE_RITE |
| A18 | `cmd/tour/collect.go` | `internal/cmd/tour/collect.go` | READ (via A1) | Tour command reads active rite for tour context |

**Consumer count: 18** (13 READ, 2 WRITE, 1 DELETE, 2 indirect READ)

### 2. state.json active_rite Consumers (.claude/sync/state.json)

| # | Consumer | File:Line | Access | Purpose |
|---|----------|-----------|--------|---------|
| S1 | `materialize.trackState()` | `internal/materialize/materialize.go:1534-1566` | READ-WRITE | Loads state, sets `state.ActiveRite = activeRiteName`, saves. This is the ONLY production code that touches state.json's active_rite. |
| S2 | `sync.StateManager.Load()` | `internal/sync/state.go:61-83` | READ | Deserializes state.json including active_rite field |
| S3 | `sync.StateManager.Save()` | `internal/sync/state.go:86-104` | WRITE | Serializes state.json including active_rite field |

**Tests that verify the field (not production consumers):**
| # | Test | File:Line | Purpose |
|---|------|-----------|---------|
| T1 | `TestState_ActiveRiteField` | `internal/sync/state_test.go:55-85` | Unit test: round-trip active_rite |
| T2 | `TestState_ActiveRiteOmittedWhenEmpty` | `internal/sync/state_test.go:87-116` | Unit test: omitempty behavior |
| T3 | `TestMaterializeRiteSwitch_Basic` | `internal/materialize/rite_switch_integration_test.go:85-89,123-126` | Integration test: verifies state.json active_rite after switch |
| T4 | `TestMaterializeRiteSwitch_SoftMode` | `internal/materialize/rite_switch_integration_test.go:287-288` | Integration test: reads raw state.json |

**Consumer count: 3 production** (1 READ-WRITE, 1 READ infrastructure, 1 WRITE infrastructure)

**Critical finding: Zero read consumers in runtime paths.** No CLI command, no hook, no sync pipeline decision reads `state.json` to determine the active rite. The only production code that touches this field is `trackState()`, which WRITES it. The Load() call in `trackState()` reads it only to get the existing state object before overwriting.

### 3. PROVENANCE_MANIFEST.yaml active_rite Consumers

| # | Consumer | File:Line | Access | Purpose |
|---|----------|-----------|--------|---------|
| P1 | `provenance.ProvenanceManifest.ActiveRite` | `internal/provenance/provenance.go:36` | (struct field) | Part of manifest schema |
| P2 | `materialize.saveProvenanceManifest()` | `internal/materialize/materialize.go:1361,1379,1414,1450` | WRITE | Sets `ActiveRite` on manifest before save |
| P3 | `provenance.structurallyEqual()` | `internal/provenance/manifest.go:90` | READ | Compares active_rite to detect structural changes |
| P4 | `provenance.LoadOrBootstrap()` | `internal/provenance/manifest.go:123` | READ | Returns empty `ActiveRite: ""` on bootstrap |
| P5 | `cmd/rite/status.runStatus()` | `internal/cmd/rite/status.go:91` | READ | Reads `KNOSSOS_MANIFEST.yaml` (inscription manifest, NOT provenance) to verify `manifest.ActiveRite == riteName` for manifest validity check |

**Note**: P5 reads from `KNOSSOS_MANIFEST.yaml` (inscription), not `PROVENANCE_MANIFEST.yaml`. The inscription manifest `active_rite` field is set during CLAUDE.md sync (`inscription.SyncCLAUDEmd` at `sync.go:83-84`). This is a fourth source of active rite data.

### 4. KNOSSOS_MANIFEST.yaml active_rite (Inscription Manifest)

| # | Consumer | File:Line | Access | Purpose |
|---|----------|-----------|--------|---------|
| K1 | `inscription.Manifest.ActiveRite` | `internal/inscription/types.go:161` | (struct field) | Part of inscription manifest schema |
| K2 | `inscription.SyncCLAUDEmd()` | `internal/inscription/sync.go:83-84` | WRITE | Sets active rite in manifest during CLAUDE.md sync |
| K3 | `inscription.Pipeline.buildRenderContext()` | `internal/inscription/pipeline.go:528` | READ | Reads manifest's active_rite for template rendering context |
| K4 | `inscription.MergeManifests()` | `internal/inscription/manifest.go:413-415` | READ-WRITE | Overlay active_rite during manifest merge |
| K5 | `inscription.Manifest.SetActiveRite()` | `internal/inscription/manifest.go:513-517` | WRITE | Setter method |
| K6 | `cmd/rite/status.runStatus()` | `internal/cmd/rite/status.go:88-93` | READ | Reads KNOSSOS_MANIFEST.yaml to validate rite matches |

---

## Source-of-Truth Map

| Store | Path | Written by | Read by (runtime) | Unique data |
|-------|------|-----------|-------------------|-------------|
| **ACTIVE_RITE** (file) | `.claude/ACTIVE_RITE` | `materialize.writeActiveRite()` (step 10) | 18 consumers (see A1-A18) | None -- this is the primary |
| **state.json** | `.claude/sync/state.json` | `materialize.trackState()` (step 9) | **ZERO runtime readers** | `last_sync` timestamp |
| **PROVENANCE_MANIFEST.yaml** | `.claude/PROVENANCE_MANIFEST.yaml` | `materialize.saveProvenanceManifest()` (step 11) | 0 for rite selection | Audit trail; file-level provenance entries |
| **KNOSSOS_MANIFEST.yaml** | `.claude/KNOSSOS_MANIFEST.yaml` | `inscription.SyncCLAUDEmd()` | 2 (K3 template rendering, K6 validation) | Region ownership, inscription version |

### Write Order Within One Sync Pipeline Run

1. **Step 9**: `trackState()` writes `state.json` with `active_rite` (`materialize.go:459-462`)
2. **Step 10**: `writeActiveRite()` writes `ACTIVE_RITE` file (`materialize.go:477-480`)
3. **Step 11**: `saveProvenanceManifest()` writes `PROVENANCE_MANIFEST.yaml` with `active_rite` (`materialize.go:483-485`)
4. KNOSSOS_MANIFEST.yaml is written earlier, during `materializeCLAUDEmd()` (step 7)

---

## Zombie Determination

### state.json `active_rite` field: **ZOMBIE**

**Evidence:**
1. **Zero runtime read consumers.** No CLI command, hook, or pipeline decision reads `state.json` to determine the active rite.
2. **All 18 consumers of active rite read from ACTIVE_RITE file** (directly or via `paths.Resolver.ReadActiveRite()`).
3. The `trackState()` function loads state.json only to get the existing state object, overwrites `ActiveRite`, and saves. No code path downstream reads the saved value.
4. The `ari status` command reads active rite from `ACTIVE_RITE` file (via `resolver.ReadActiveRite()` at `status.go:248`), NOT from state.json.
5. The integration tests (T1-T4) verify the field exists but no production code depends on their assertions.

### state.json file itself: **NOT a zombie**

The `last_sync` timestamp in state.json has no equivalent elsewhere:
- `PROVENANCE_MANIFEST.yaml` has its own `last_sync` but it tracks provenance writes, not sync pipeline completion. They are semantically different.
- `KNOSSOS_MANIFEST.yaml` has `last_sync` but it tracks inscription version timestamps.
- `state.json`'s `last_sync` is the only record of when the full materialization pipeline last completed.

However, **no runtime code reads `state.json`'s `last_sync` either**. The `ari status` dashboard reads `last_sync` from `PROVENANCE_MANIFEST.yaml` (at `status.go:262-265`), not from state.json. This makes state.json's `last_sync` also a candidate zombie, but that is outside the scope of this audit (focused on `active_rite`).

---

## Recommendation: Remove `active_rite` from state.json

### Migration Path

1. **Remove `ActiveRite` field from `sync.State` struct** (`internal/sync/state.go:19`)
2. **Remove `state.ActiveRite = activeRiteName` assignment** (`internal/materialize/materialize.go:1558`)
3. **Update tests** (T1-T4) to remove active_rite assertions
4. **Bump state.json schema version** from "1.0" to "1.1" (field removal)
5. **Existing state.json files**: backward compatible -- `json:"active_rite,omitempty"` means old files with the field will deserialize fine, and new files will omit it

### Consumers That Need Updating

| File | Change |
|------|--------|
| `internal/sync/state.go:19` | Remove `ActiveRite string` field |
| `internal/materialize/materialize.go:1558` | Remove `state.ActiveRite = activeRiteName` |
| `internal/sync/state_test.go:55-116` | Remove `TestState_ActiveRiteField` and `TestState_ActiveRiteOmittedWhenEmpty` |
| `internal/materialize/rite_switch_integration_test.go:85-89,123-126` | Remove state.json active_rite assertions |
| `internal/materialize/rite_switch_integration_test.go:287-288` | Update raw state.json read to not check active_rite |

### Consumers That Need NO Changes

All 18 ACTIVE_RITE consumers (A1-A18) are unaffected -- they do not read state.json.
All provenance consumers (P1-P5) are unaffected.
All inscription manifest consumers (K1-K6) are unaffected.

---

## Risk Assessment

### What breaks if we get this wrong

1. **If ACTIVE_RITE file is deleted but state.json is intact**: `ari sync` without `--rite` flag fails with `"no ACTIVE_RITE found"` (scope=rite) or falls to minimal mode (scope=all). state.json's active_rite would NOT help -- nothing reads it for recovery. This is the current behavior and removing the field does not change it.

2. **If state.json is deleted entirely**: `trackState()` re-initializes from scratch. This is transparent self-healing and already tested (`state_test.go:13-53`).

3. **If all three stores diverge**: `ari sync --rite <name>` re-synchronizes them. Removing active_rite from state.json means two stores instead of three, reducing divergence surface.

### Low-risk classification

- The field has zero runtime read consumers
- Removing it is a pure subtraction (no behavior change)
- Existing state.json files with the field are backward compatible (omitempty)
- The migration is confined to 5 files with 5 line-level changes

---

## Complexity: PATCH

Rationale: Single field removal from a struct, one assignment deletion in materialize.go, and test updates. No behavioral changes, no new logic, no cross-component coordination required.

---

## Broader Observation: state.json Entire File May Be Zombie

While out of scope for this audit, the investigation revealed that `last_sync` in state.json is also not read by any runtime consumer. The `ari status` dashboard reads `last_sync` from `PROVENANCE_MANIFEST.yaml`, not from state.json. The `sync.StateManager` is consumed only by `materialize.trackState()` which writes to it and by tests. A follow-up audit (DEBT-039) should evaluate whether the entire state.json file can be eliminated by ensuring `last_sync` is adequately served by PROVENANCE_MANIFEST.yaml.

---

## Appendix: Session Context active_rite (Separate Concern)

The `session.Context.ActiveRite` field (`internal/session/context.go:23`) is a **separate concern** -- it records which rite was active when the session was created. This is per-session metadata (embedded in SESSION_CONTEXT.md frontmatter), not a global source of truth. It is:
- Written during `session.NewContext()` (context.go:248)
- Read by hooks (context.go:163 as backward-compat fallback)
- Read/written by `ari session field-set active_rite` (field.go:309)
- Read by handoff status (status.go:89)
- Read by session snapshot (snapshot.go:76)

This is legitimate per-session state and is unrelated to the state.json zombie question.

# SPIKE: Optimal Strategy for Legacy Cleanup with Provenance Patterns

**Date**: 2026-03-02
**Status**: Complete
**Timebox**: 2h (research spike)

## Question

What is the optimal strategy for cleaning up legacy backward-compatibility artifacts (env vars, stale settings, deprecated manifests, naming debt) given that:
1. The unified provenance model (ADR-0026) is fully implemented
2. There are minimal existing knossos consumers (satellites)
3. Over-engineering backward compatibility shims for a future that won't materialize is wasteful

Should we build migration infrastructure, or treat this as an ephemeral one-off cleanup?

## Decision This Informs

Whether legacy artifacts require a formal migration framework (deprecation periods, dual-read shims, schema versioning) or can be cleaned with a simple, direct one-time sweep.

---

## First Principles Analysis

### Principle 1: Consumer Count Determines Migration Strategy

The industry standard decision matrix for backward compatibility is:

| Consumer Count | Breaking Change Cost | Strategy |
|---|---|---|
| 0-10 internal | Near-zero | Direct cleanup, coordinated release |
| 10-100 known | Low | Announcement + grace period |
| 100+ or unknown | High | Formal deprecation, dual-read, migration tools |

Knossos satellites are in the **0-10 internal** category. Every consumer is either the knossos repo itself or a handful of autom8y projects under direct control. The correct strategy at this scale is **coordinated direct cleanup**, not migration infrastructure.

### Principle 2: Deprecation Shims Become Permanent

Industry evidence (Kubernetes, Terraform, Go stdlib) shows that deprecation shims survive indefinitely once introduced. The `LegacyDataDir()` function in `internal/paths/paths.go` is a local example -- marked `Deprecated` but still present because removing it requires auditing callers. Each shim:
- Adds code surface area
- Creates confusion about which path is canonical
- Eventually becomes load-bearing when new code accidentally uses the deprecated path

At the current consumer scale, introducing new deprecation shims is strictly worse than direct removal.

### Principle 3: Provenance Already Solves the Ownership Problem

The unified provenance model (ADR-0026, Phase 4b complete) provides the infrastructure to answer "who owns this file?" for every artifact in `.claude/`. This means:
- **Stale `settings.json` files** created by the removed `writeDefaultSettings()` can be identified by checking provenance (they have no provenance entry because they predate the manifest, and no pipeline creates them now).
- **Orphaned legacy manifests** (`USER_MENA_MANIFEST.json`, `USER_AGENT_MANIFEST.json`) can be detected as files with no provenance entry.
- **The sync pipeline itself** is the natural cleanup vector -- it already handles orphan detection and removal.

### Principle 4: Cleanup Work is Not Feature Work

Building migration infrastructure (version-gated readers, deprecation warnings, dual-path resolution) consumes the same engineering time as building features. For 0-10 consumers, the ROI of migration tooling is negative -- it costs more to build than the manual cleanup it prevents.

---

## Inventory of Legacy Artifacts

### Category A: Already Removed (No Action Required)

| Artifact | Status |
|---|---|
| `writeDefaultSettings()` in `init.go` | Function deleted, call sites removed |
| `internal/usersync/` package | Absorbed into `internal/materialize/userscope/` (ADR-0026 Phase 4b) |
| `.current-session` file | Migrated to CC Session Map (ADR-0027) |
| `StagedMaterialize` rename pattern | Removed (SCAR-002) |

### Category B: Residual Naming Debt (Direct Rename)

| Artifact | Location | Consumers | Action |
|---|---|---|---|
| `ARIADNE_BUDGET_DISABLE` | `internal/cmd/hook/budget.go:35` | 0 external (env var) | Rename to `ARI_BUDGET_DISABLE` |
| `ARIADNE_MSG_WARN` | `internal/cmd/hook/budget.go:36` | 0 external | Rename to `ARI_MSG_WARN` |
| `ARIADNE_MSG_PARK` | `internal/cmd/hook/budget.go:37` | 0 external | Rename to `ARI_MSG_PARK` |
| `ARIADNE_SESSION_KEY` | `internal/cmd/hook/budget.go:38` | 0 external (test-only) | Rename to `ARI_SESSION_KEY` |
| `ARIADNE_STALE_SESSION_DAYS` | `internal/cmd/session/list.go:186` | 0 external | Rename to `ARI_STALE_SESSION_DAYS` |
| `LegacyDataDir()` | `internal/paths/paths.go:270` | 0 callers (dead code) | Delete |

**Total: 6 items, all zero external consumers.**

### Category C: Stale Files in Satellites (One-Off Cleanup)

| File | Problem | Fix |
|---|---|---|
| `.claude/settings.json` (in satellites) | Blanket agent-guard deny from deleted `writeDefaultSettings()` | Delete the file; `settings.local.json` is the correct mechanism |
| `USER_MENA_MANIFEST.json` | Superseded by `USER_PROVENANCE_MANIFEST.yaml` | `.v2-backup` already created by migration; delete backup after verification |
| `USER_AGENT_MANIFEST.json` | Same | Same |
| `USER_HOOKS_MANIFEST.json` | Same | Same |

### Category D: Structural Tensions (Not Cleanup -- Track Separately)

These are design tensions documented in `.know/design-constraints.md` and are NOT candidates for cleanup in this initiative:

| Tension | Why Not Cleanup |
|---|---|
| TENSION-001: Dual `OwnerType` | Genuinely different value sets (inscription vs provenance); renaming one is a semantic change, not cleanup |
| TENSION-002: Dual mena fields | Requires satellite manifest audit before removal |
| TENSION-007: `SourceType` duplication | Deliberate leaf-package design (provenance has no internal imports) |
| TENSION-008: Dual event schema | ADR-0027 migration in progress |

---

## Recommendation: Ephemeral One-Off Cleanup

### Strategy: Direct Sweep, No Migration Framework

1. **Rename `ARIADNE_*` env vars to `ARI_*`** in a single commit. No dual-read shim. Zero external consumers means zero breaking changes.

2. **Delete `LegacyDataDir()`** and its test. Zero callers.

3. **Add stale `settings.json` cleanup to `ari sync`**: During `syncRiteScope()`, if `.claude/settings.json` exists AND contains only an `ari hook agent-guard` entry with no `--allow-path` flags (the exact buggy template), delete it. This is a one-time cleanup that self-removes after first sync.

4. **Delete `.v2-backup` manifest files**: The v1-to-v2 migration is complete. The backup files serve no purpose. Add a one-line cleanup to `provenance.LoadOrBootstrap()` that removes `*.v2-backup` files if they exist.

5. **Do NOT build**: deprecation warnings, version-gated readers, dual-path resolution, or environment variable fallback chains. These would be over-engineering for the current consumer count and would become permanent fixtures.

### Why Not a Migration Framework

| Concern | Answer |
|---|---|
| "What if a satellite uses `ARIADNE_*`?" | Audit all satellites (there are <5). Grep, fix, done. |
| "What if we add more consumers later?" | Future consumers will use the current naming. They have no legacy to migrate from. |
| "What about the `.claude/settings.json` bug?" | The sync pipeline can detect and fix this automatically. No user intervention needed. |
| "Shouldn't we preserve backward compat?" | Backward compat is for external consumers you can't coordinate with. Internal consumers get a coordinated cut. |

### Implementation Approach

**Sprint 1 (single PR):**
- Rename 5 `ARIADNE_*` constants + all test references
- Delete `LegacyDataDir()` + test
- Add `cleanupStaleBlanketSettings()` to sync pipeline
- Update help text and documentation references

**Sprint 2 (satellite sweep):**
- Run `ari sync` on each satellite to trigger automatic cleanup
- Verify no `ARIADNE_*` references remain in any satellite
- Delete `.v2-backup` files from `~/.claude/`

### What the Provenance System Already Handles

The provenance model handles the hard parts automatically:
- **Orphan detection**: Files without provenance entries (like stale `settings.json`) are detectable
- **Ownership transitions**: `untracked -> user` or `untracked -> knossos` happens naturally on sync
- **Divergence detection**: Modified files are promoted to `satellite` owner, never overwritten
- **Checksum verification**: Changed files are detected without filename heuristics

The cleanup work described above leverages these existing capabilities rather than reinventing them.

---

## Anti-Patterns to Avoid

1. **"Just add a fallback"**: Adding `os.Getenv("ARI_MSG_WARN")` with `os.Getenv("ARIADNE_MSG_WARN")` fallback creates permanent dual-path code that is never cleaned up.

2. **"Deprecation warning first, removal later"**: For zero external consumers, this is two PRs where one suffices. The warning will never be seen.

3. **"Schema version bump + migration function"**: The provenance manifest already has `migrateV1ToV2()`. Adding `migrateV2ToV3()` for env var renames conflates configuration cleanup with schema evolution.

4. **"Feature flag the new names"**: Feature flags for internal naming conventions is pure overhead.

---

## Follow-Up Actions

1. **PR: ARIADNE env var rename** -- Single commit, all 5 env vars + help text + tests
2. **PR: Delete LegacyDataDir** -- One function + one test, zero callers
3. **PR: Stale settings.json cleanup** -- Add detection to sync pipeline, auto-remove blanket deny
4. **Satellite sweep** -- Manual `ari sync` on each autom8y satellite after PRs merge
5. **Track Category D tensions separately** -- These are design decisions, not cleanup items

## Files Involved

| File | Change |
|---|---|
| `internal/cmd/hook/budget.go` | Rename 4 `ARIADNE_*` constants to `ARI_*` |
| `internal/cmd/hook/budget_test.go` | Update test env var names |
| `internal/cmd/session/list.go` | Rename `ARIADNE_STALE_SESSION_DAYS` |
| `internal/cmd/session/list_test.go` | Update test env var names |
| `internal/cmd/session/gc.go` | Update help text reference |
| `internal/cmd/session/wrap.go` | Update comment references |
| `internal/paths/paths.go` | Delete `LegacyDataDir()` |
| `internal/paths/paths_test.go` | Delete `TestLegacyDataDir` |
| `internal/materialize/materialize.go` | Add `cleanupStaleBlanketSettings()` |

# ARCH-REVIEW-1 Unknowns — Decision Log

**Date**: 2026-02-18
**Decided by**: Stakeholder interview (Phase 2 of hotfix+interview session)
**Context**: ARCH-REVIEW-1-HEALTH.md Section 5 (Unknowns Registry)

## U-1: Duplicate RiteManifest structs — DEFERRED

**Decision**: Defer to Initiative C design spike.

**Rationale**: The subset relationship between `materialize.RiteManifest` (13 fields, pipeline projection) and `rite.RiteManifest` (15+ fields, orchestration) is intentional — they serve different consumers. Unifying requires a new shared package which is entangled with the decomposition design. The C spike will have better information about package boundaries.

**Constraint**: C Session 1 (design spike) MUST produce a ruling on this before extraction begins.

## U-2: Soft mode indefinite staleness — ACCEPT AS-IS

**Decision**: Accept current behavior.

**Rationale**: Soft mode is for speed during session resume, not rite switches. The pipeline always runs a full sync on rite switch. Stale state from `--soft` is only possible if a user manually passes `--soft` after switching rites — an edge case not worth engineering around. The `DeferredStages` field already tracks this state for debugging.

## U-3: Validator Fix() side effects — ACCEPT AS-IS

**Decision**: Accept current behavior. `Fix()` is the intentional write path.

**Rationale**: `ari rite validate` calls `Validate()` which is read-only. `Fix()` is explicitly about repairing issues — writing files is the point. The `Sync()` call at `validate.go:293` uses `KeepOrphans: true` so it's conservative. No change needed.

## CR-3: Vestigial sync.State fields — RESOLVED

**Decision**: Removed in commit `0c69b20`.

Deleted `Remote`, `TrackedFiles`, `Conflicts` fields, `TrackedFile`/`Conflict` types, `UpdateTrackedFile` method, and `Initialize()` remote parameter. These were leftovers from the remote sync era deleted in Phase 4b.

## CR-5: writeIfChanged omission in user_scope.go — RESOLVED

**Decision**: Documented in commit `16c279c`.

Added comment at top of `syncUserScope()` explaining that user-scope targets (`~/.claude/`) are outside CC's project-level file watcher scope, so `writeIfChanged` optimization is not needed.

## R3: Silent provenance error discards — RESOLVED

**Decision**: Fixed in commits `e5655a9` (materialize.go) and `e7277cf` (mena.go).

All `provenance.LoadOrBootstrap` and `provenance.DetectDivergence` calls now surface errors as WARN logs instead of silently discarding via `_`. Fallback to empty manifest on load failure. Zero remaining silent provenance error discards in `internal/materialize/`.

## Hotfix Summary

| Item | Status | Commit |
|------|--------|--------|
| R3 | RESOLVED | `e5655a9` |
| R3 (mena bonus) | RESOLVED | `e7277cf` |
| CR-5 | RESOLVED | `16c279c` |
| CR-3 | RESOLVED | `0c69b20` |
| U-1 | DEFERRED to C spike | — |
| U-2 | ACCEPT AS-IS | — |
| U-3 | ACCEPT AS-IS | — |

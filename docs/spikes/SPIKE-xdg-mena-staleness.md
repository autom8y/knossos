# SPIKE: XDG Mena is an Out-of-Date Copy

| Field | Value |
|-------|-------|
| **Date** | 2026-03-02 |
| **Status** | Complete |
| **Timebox** | 30 min |
| **Decision** | Fix required: three defects identified |

## Question

The XDG mena directory (`~/Library/Application Support/knossos/mena/` on macOS) is a stale copy of the source `mena/` directory. How did this happen, what are the consequences, and what should be done about it?

## Context

Knossos distributes platform-level mena (commands and skills like `/go`, `/commit`, `/start`, guidance skills, etc.) through multiple resolution tiers. The materializer's `getMenaDir()` resolves platform mena with this priority:

1. `.knossos/mena/` (satellite overrides)
2. XDG data dir (`~/Library/Application Support/knossos/mena/`)
3. `KnossosHome/mena/` (developer case)
4. Embedded FS fallback (only if none of the above exist)

## Approach

1. Traced the code path that creates the XDG mena copy
2. Compared the XDG copy against the source `mena/` directory
3. Identified all consumers of the XDG mena path
4. Assessed impact on developer and satellite workflows

## Findings

### Defect 1: One-Shot Extraction with No Update Mechanism

`extractEmbeddedMenaToXDG()` in `internal/cmd/initialize/init.go` (line 263) runs during `ari init` and extracts the embedded mena to the XDG data directory. The function has a **one-shot guard**:

```go
if _, err := os.Stat(xdgMena); err == nil {
    return // Already extracted
}
```

Once the XDG mena directory exists, it is **never updated** -- not on subsequent `ari init` calls, not on `ari sync`, not on binary upgrades. The comment says "Idempotent: skips extraction if XDG mena dir already exists" but "idempotent" here means "runs once and then never again."

### Defect 2: XDG Shadows KnossosHome for Developers

In `getMenaDir()` (line 121 of `materialize_mena.go`), the XDG data dir is checked **before** `KnossosHome/mena/`:

```
1. .knossos/mena/  -- knossos project doesn't have this
2. XDG data dir    -- EXISTS, STALE, WINS
3. KnossosHome     -- CORRECT, NEVER REACHED
```

For the developer (knossos-on-knossos) case, the stale XDG copy shadows the live source `mena/` directory. The correct source at `$KNOSSOS_HOME/mena/` (tier 3) is never reached because the stale XDG copy (tier 2) is found first.

Note: `ResolvePlatformMenaDir()` in `internal/mena/platform.go` has a different, arguably correct resolution order (`projectRoot/mena/` first), but it is **dead code** -- only referenced in its own test file, never called by the materializer.

### Defect 3: Same Shadowing Affects Satellites

For satellite projects (non-knossos repos using knossos rites), the resolution order is:

1. `.knossos/mena/` -- satellites don't typically have this
2. XDG data dir -- **stale copy wins here too**
3. KnossosHome -- never reached

This means satellites running on a developer machine with `KNOSSOS_HOME` set are also getting the stale XDG platform mena instead of the live source.

### Quantified Staleness (Current Machine)

| Metric | Count |
|--------|-------|
| Files in source `mena/` | 153 |
| Files in XDG `mena/` | 150 |
| Missing from XDG (new files) | 6 |
| Content divergence (same path, different content) | 66 |
| **Total differences** | **72** |

Missing files include:
- `guidance/cross-rite/routes/clinic-to-10x.md`
- `guidance/cross-rite/routes/clinic-to-debt-triage.md`
- `guidance/cross-rite/routes/clinic-to-sre.md`
- `templates/justfile/templates/_env.just.template`
- `templates/justfile/templates/_globals.just.template`
- `templates/justfile/templates/_helpers.just.template`

Content differences include critical operational files like `navigation/go.dro.md`, all rite-switching commands, session commands, and several guidance skills.

### What Is NOT Affected

- **User-scope sync** (`ari sync --scope=user`): Reads from `KNOSSOS_HOME/mena`, not XDG. Correct.
- **Embedded FS fallback**: Only used when no filesystem mena exists at all. The XDG copy prevents this fallback from ever being reached.

## Root Cause

The XDG mena extraction was designed for the "installed user" scenario (binary installed via `go install` or homebrew, no source tree). In that scenario, the one-shot extraction makes sense because the XDG copy IS the only filesystem copy, and updates come from rebuilding the binary with new embedded content.

The design did not account for:
1. Developer machines where both XDG and `KNOSSOS_HOME` exist
2. Binary upgrades where the XDG copy should be refreshed from the new embedded content
3. The resolution order in `getMenaDir()` where XDG (tier 2) preempts KnossosHome (tier 3)

## Recommendation

### Option A: Delete the XDG Mena Directory (Quick Fix for Developer)

```bash
rm -rf ~/Library/Application\ Support/knossos/mena/
```

This forces the materializer to fall through to tier 3 (KnossosHome) or tier 4 (embedded). Solves the immediate problem but will recur on next `ari init`.

### Option B: Swap Resolution Order (Correct Fix)

In `getMenaDir()`, check `KnossosHome/mena/` **before** XDG data dir. This matches the intent: developers editing source should get the live source, not a cached copy.

```
1. .knossos/mena/       -- satellite overrides (highest priority)
2. KnossosHome/mena/    -- developer/source case
3. XDG data dir          -- installed user case (cached from embedded)
4. Embedded FS fallback  -- fresh install, no extraction yet
```

This also aligns with the rite source resolution order where `KNOSSOS_HOME` has higher priority than XDG.

### Option C: Version-Stamped XDG Extraction (Complete Fix)

In addition to Option B, add version-awareness to `extractEmbeddedMenaToXDG`:

1. Write a `.version` sentinel file alongside the extracted mena
2. On each `ari init`, compare the binary version against the sentinel
3. Re-extract if versions differ

This fixes the "binary upgrade but XDG stays stale" problem for installed users who don't have `KNOSSOS_HOME`.

### Option D: Also Call Extraction from `ari sync` (Belt and Suspenders)

Move or duplicate `extractEmbeddedMenaToXDG` to run during `ari sync` as well, not just `ari init`. Most users run `ari sync` frequently; relying solely on `ari init` (which runs once per project) makes the update window very wide.

### Recommended Combination

**B + C**: Swap the resolution order (immediate correctness fix) AND add version-stamped extraction (ensures installed-user XDG copy stays current across upgrades).

### Dead Code Cleanup

`ResolvePlatformMenaDir()` in `internal/mena/platform.go` should either be:
- Deleted (it's dead code), or
- Adopted by the materializer to replace the custom `getMenaDir()` (after aligning its resolution order)

## Follow-Up Actions

1. **Immediate**: `rm -rf ~/Library/Application\ Support/knossos/mena/` on developer machines
2. **PR 1**: Swap `getMenaDir()` resolution order (KnossosHome before XDG)
3. **PR 2**: Add version-stamped XDG extraction with re-extraction on version mismatch
4. **PR 3**: Decide whether to consolidate `getMenaDir()` and `ResolvePlatformMenaDir()` or delete the dead code

## Files Referenced

| File | Role |
|------|------|
| `/Users/tomtenuta/Code/knossos/internal/materialize/materialize_mena.go` | `getMenaDir()` resolution, XDG tier 2 check |
| `/Users/tomtenuta/Code/knossos/internal/cmd/initialize/init.go` | `extractEmbeddedMenaToXDG()` one-shot extraction |
| `/Users/tomtenuta/Code/knossos/internal/mena/platform.go` | Dead code `ResolvePlatformMenaDir()` |
| `/Users/tomtenuta/Code/knossos/internal/config/home.go` | `XDGDataDir()` resolution |
| `/Users/tomtenuta/Code/knossos/internal/materialize/userscope/sync_mena.go` | User-scope sync (NOT affected) |
| `/Users/tomtenuta/Code/knossos/internal/paths/paths.go` | XDG directory helpers |

# Gap Analysis: Unify lint walkMena with internal/mena Resolution

## Root Cause

`internal/cmd/lint/lint.go:696-731` implements `walkMena()` — a content walker that
reimplements mena source discovery independently of the shared `internal/mena/` package.
This creates two divergent implementations of "which directories contain mena files":

1. **`walkMena()`** (lint): walks `mena/` (platform) + `rites/*/mena/` (all rites via glob)
2. **`internal/mena.BuildSourceChain()`** (materialize): constructs a 4-tier priority-ordered
   list: `platform → shared → dependency → rite-local`

While these currently produce overlapping results (because `rites/shared/mena/` IS caught by
the `rites/*/mena/` glob), the implementations have fundamentally different semantics and will
diverge if source resolution ever changes.

### Specific divergences

| Aspect | walkMena (lint) | BuildSourceChain (materialize) |
|--------|----------------|-------------------------------|
| Sources walked | `mena/` + `rites/*/mena/` (all rites) | platform + shared + dependencies + active rite |
| Priority ordering | None (flat walk, order undefined) | Priority-ordered (later = higher) |
| Embedded FS support | No | Yes (MenaSource.IsEmbedded) |
| Nonexistent dirs | filepath.Walk silently skips | Callers handle via os.Stat |
| Callback signature | `fn(path, relPath string, data []byte)` | N/A (no Walk function exists) |

### Missing primitive

`internal/mena/` provides `BuildSourceChain()` (source enumeration) and `Exists()` (name
resolution) but has **no Walk function** that iterates all files with content. The lint
command needs exactly this: iterate all mena files matching a suffix, reading content for
each, across a set of MenaSource entries.

## Affected Components

### Primary: `internal/cmd/lint/lint.go`

Three call sites use `walkMena()`:

| Line | Caller | Purpose |
|------|--------|---------|
| 399 | `lintDromena()` | Walk `.dro.md` files, lint each |
| 481 | `lintMenaNamespace()` | Walk `.dro.md` files, detect name collisions |
| 540 | `lintLegomena()` | Walk `.lego.md` files, lint each |

All three use identical callback signature: `fn(path, relPath string, data []byte)`.

### Secondary: `internal/mena/` (leaf package)

Needs a new `Walk()` function. This package is stdlib-only (`os`, `filepath`, `strings`,
`io/fs`). Adding Walk maintains that constraint.

### Tertiary: `internal/materialize/mena.go` (re-export layer)

Contains 8 type aliases and 8 re-exported functions/constants. If Walk is added to
`internal/mena/`, this file may need a re-export. Assessment: the re-export layer exists
for backward compatibility with callers that import `internal/materialize` directly. Since
lint.go would import `internal/mena/` directly (it's a leaf package), no re-export is
strictly needed. Cleanup of the re-export layer is a separate concern.

## Open Design Question: All-Rites vs. Source-Chain

This is the critical question the Context Architect must resolve.

**Current behavior** (`walkMena`): walks ALL rites' mena directories. This means `ari lint`
validates every mena file in the repository, regardless of which rite is active. This is
appropriate for a source validation tool — you want to catch errors in inactive rites too.

**Source chain behavior** (`BuildSourceChain`): walks only the active rite's source chain
(platform + shared + active rite + dependencies). This is appropriate for materialization —
you only project files that will actually be used.

**Options**:

1. **Walk(sources, suffix, fn)** — caller passes sources; lint builds an "all rites" source
   list, materialize builds a source-chain list. Walk is source-agnostic.
2. **WalkAll(projectRoot, suffix, fn)** — convenience that discovers all sources
   automatically. Less flexible but matches lint's current intent.
3. **Both** — Walk for parameterized use, WalkAll as sugar built on Walk.

The key constraint: lint validates sources before sync, so it intentionally walks ALL rites.
A Walk function parameterized by `[]MenaSource` satisfies both consumers if lint constructs
a sources list containing all rite directories.

## Risks

### Behavioral change: previously-unseen files

If `internal/mena/` Walk implementation differs from `walkMena` in directory traversal order
or error handling, lint output order may change. This is cosmetic (lint findings are
unordered), but test snapshots would need updating.

### New imports in lint.go

`lint.go` currently imports: `frontmatter`, `output`, `common`. Adding `internal/mena/`
introduces a new dependency. Since `internal/mena/` is a leaf package with no transitive
internal imports, this adds zero dependency fan-out risk.

### Re-export layer churn

The 8 re-exports in `internal/materialize/mena.go` exist for backward compatibility. Adding
Walk to `internal/mena/` does not require touching this file unless other callers need the
re-export. Recommend: do NOT re-export Walk. Callers that need it import `internal/mena/`
directly.

### findAgentDirs has the same pattern

`lint.go:265-279` has `findAgentDirs()` which walks `rites/*/agents/` — same all-rites glob
pattern as `walkMena`. This is a parallel divergence from the agent materialization pipeline.
However, agent resolution does not live in a shared leaf package, so unifying it is a
separate concern.

## Complexity: MODULE

**Rationale**: Two components affected (internal/mena/ addition + lint.go refactor), but
both are straightforward:

- `internal/mena/Walk()`: ~30 lines, follows `Exists()` patterns exactly (iterate sources,
  skip nonexistent, recurse into directories, filter by suffix, read content)
- `lint.go` refactor: replace 3 `walkMena(projectRoot, ...)` calls with
  `mena.Walk(sources, ...)`, construct `sources` once from all-rites discovery
- Existing test coverage: `lint_test.go` tests helper functions but not walkMena directly.
  Integration testing via `ari lint` on the repo itself.

Not PATCH (touches two packages, requires API design decision). Not SYSTEM (no cross-cutting
pipeline changes, no migration).

## Success Criteria

- `walkMena()` function removed from `lint.go`
- All 3 call sites use `mena.Walk()` (or equivalent)
- `ari lint` produces identical findings before and after (modulo ordering)
- `internal/mena/` remains stdlib-only (no new internal imports)
- `CGO_ENABLED=0 go test ./internal/mena/...` passes with Walk tests
- `CGO_ENABLED=0 go test ./internal/cmd/lint/...` passes
- `ari lint` exits 0 on the knossos repo (no regressions)

## Test Satellites

| Satellite | Purpose |
|-----------|---------|
| knossos repo itself | Full-scale lint validation (18 rites, platform mena, shared mena) |
| test fixture with no rites/ | Graceful handling of missing directories |
| test fixture with mixed dro/lego in shared | Verify shared mena files are walked |

## Dependency Map

```
internal/mena/  (leaf, stdlib-only)
  +-- Walk() [NEW]
  +-- BuildSourceChain() [existing]
  +-- Exists() [existing]

internal/cmd/lint/lint.go  (CLI layer)
  +-- imports internal/mena/ [NEW import]
  +-- removes walkMena() [private helper]
  +-- constructs all-rites source list [NEW logic, ~10 lines]
```

## Out of Scope

- Re-export layer cleanup (`internal/materialize/mena.go`) — separate hygiene task
- `findAgentDirs()` unification — parallel pattern but no shared agent resolution package
- Embedded FS support in Walk — lint operates on filesystem sources only; embedded FS
  iteration can be added later if needed
- Lint-specific source chain construction helper — implementation detail for Integration
  Engineer

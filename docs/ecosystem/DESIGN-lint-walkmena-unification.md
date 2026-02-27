# Context Design: Unify lint walkMena with internal/mena Walk()

**Date**: 2026-02-27
**Author**: Context Architect (ecosystem rite)
**Status**: READY FOR IMPLEMENTATION
**Gap Analysis**: `docs/ecosystem/GAP-lint-walkmena-unification.md`

## Problem Statement

`internal/cmd/lint/lint.go:696-731` implements `walkMena()` -- a private helper that
reimplements mena source discovery independently of the shared `internal/mena/` package.
The lint function hardcodes two source paths (`mena/` and `rites/*/mena/`) while
`internal/mena/` has `BuildSourceChain()` and `Exists()` but no content-walking primitive.
This creates divergent implementations of "which directories contain mena files" that will
drift as source resolution evolves.

## Options Considered

### Option A: Walk(sources, suffix, fn) -- source-parameterized iteration

**Approach**: Add a `Walk()` function to `internal/mena/` that takes `[]MenaSource`, a file
suffix filter, and a callback. The caller constructs sources however it needs -- lint builds
an "all rites" list, materialization builds a source-chain list. Walk iterates each source's
filesystem tree, reads matching files, and invokes the callback.

**Pros**: Single implementation of directory traversal. Caller controls which sources to walk.
Follows the Exists() pattern (sources are caller-provided). Lint can construct its own
all-rites source list without Walk needing to know about rite structure.

**Cons**: Lint must build its source list externally. Callback receives `(absPath, relPath, data)`
where `relPath` is relative to the source directory, not to projectRoot -- lint currently uses
projectRoot-relative paths for display.

**Verdict**: Selected, with a modification to relPath semantics (see detailed design below).

### Option B: WalkAll(projectRoot, suffix, fn) -- convenience all-sources walk

**Approach**: Add a `WalkAll()` convenience function that discovers all mena sources
(platform + all rites) and walks them.

**Pros**: One-line call from lint. No source construction needed.

**Cons**: Encodes "all rites" discovery logic inside `internal/mena/`, which conflates two
concerns: the package should provide primitives, not opinionated discovery. Also forces the
package to know about project structure (where `rites/` lives), violating the leaf package
principle. `BuildSourceChain()` already takes an options struct for this reason.

**Verdict**: Rejected. Discovery of "all rites" is a caller concern, not a resolution
primitive. Lint is the only consumer that needs all-rites semantics today.

### Option C: Walk(sources, suffix, fn) + AllRiteSources(projectRoot) helper

**Approach**: Walk() as in Option A, plus a separate `AllRiteSources()` function in
`internal/mena/` that builds the `[]MenaSource` for all rites.

**Pros**: Lint gets a clean two-line call. The helper is reusable.

**Cons**: `AllRiteSources()` needs to know project layout (`mena/`, `rites/*/mena/`), which
is coupling that belongs in the caller, not in a resolution primitives package. Today only
lint needs this. Building it into `internal/mena/` is premature generalization.

**Verdict**: Rejected. If a second consumer emerges, the helper can be added then. For now,
the ~10 lines of all-rites source construction belong in lint.go.

## Detailed Design

### Component 1: Walk() function in internal/mena/

**File**: `/Users/tomtenuta/Code/knossos/internal/mena/walk.go` (new file)

**Function signature**:

```go
// WalkEntry holds the data passed to a Walk callback.
type WalkEntry struct {
    Path    string // Absolute filesystem path to the file
    RelPath string // Path relative to the MenaSource.Path directory
    Data    []byte // File content
}

// Walk iterates all files matching suffix within the given sources.
// For each matching file, it reads the content and invokes fn.
//
// Sources with empty or nonexistent paths are silently skipped.
// Files that cannot be read are silently skipped (consistent with Exists
// behavior where os.ReadDir/os.Stat errors return false, not error).
//
// Walk does NOT support embedded FS sources (IsEmbedded=true are skipped).
// Lint operates on filesystem sources only; embedded FS iteration can be
// added later if a consumer needs it.
//
// The suffix filter matches against the full filename, not just the extension.
// Example suffixes: ".dro.md", ".lego.md".
func Walk(sources []MenaSource, suffix string, fn func(WalkEntry)) {
    ...
}
```

**Behavior specification**:

1. Iterate `sources` in order.
2. For each source where `IsEmbedded == false` and `Path != ""`:
   a. Call `filepath.WalkDir(source.Path, ...)` (prefer WalkDir over Walk for
      efficiency -- WalkDir does not call os.Lstat on every file).
   b. Skip directories (continue walking into them).
   c. Skip files that do not have `suffix` as a suffix of their name.
   d. Read file content via `os.ReadFile(path)`. Skip on error.
   e. Compute `relPath` as `filepath.Rel(source.Path, path)` -- relative to the
      source directory, not to any project root.
   f. Invoke `fn(WalkEntry{Path: path, RelPath: relPath, Data: data})`.
3. If `filepath.WalkDir` returns an error for the root (source.Path does not
   exist), silently skip the entire source. This matches `Exists()` behavior
   where nonexistent directories return false without error.

**Design rationale for WalkEntry struct vs positional parameters**:

The current `walkMena` callback uses `fn(path, relPath string, data []byte)` -- three
positional parameters. A struct is preferable because:
- It is extensible without breaking callers (future fields like SourceIndex).
- It is self-documenting at call sites (entry.Path vs unnamed first string).
- It follows the SourceChainOptions pattern already established in the package.

**Design rationale for relPath semantics**:

Walk computes `relPath` relative to the MenaSource.Path, not relative to any project root.
This is a semantic change from `walkMena()` which computes relPath relative to projectRoot.

Rationale: Walk is a generic primitive. It does not know about project roots. The caller
(lint.go) knows its own project root and can compute a project-relative display path if
needed. The lint refactoring handles this (see Component 2).

**Design rationale for WalkDir over Walk**:

`filepath.WalkDir` (Go 1.16+) is preferred over `filepath.Walk` because WalkDir does not
call `os.Lstat` on every entry, which is measurably faster on large directory trees. The
knossos mena tree has ~200 files across 18 rites -- modest, but WalkDir is strictly better
with no API disadvantage.

**Design rationale for skipping embedded FS**:

Walk is intended for lint (filesystem validation tool). Embedded FS iteration has different
semantics (fs.WalkDir with fs.FS) and no current consumer. Adding it would double the
function's complexity for zero current benefit. If a consumer emerges, Walk can be extended
with an `if src.IsEmbedded { walkEmbedded(...) }` branch.

**Imports**: `os`, `path/filepath`, `strings`. All stdlib. Leaf package invariant preserved.

### Component 2: lint.go refactoring

**File**: `/Users/tomtenuta/Code/knossos/internal/cmd/lint/lint.go`

**Change 1: Add import**

Add `"github.com/autom8y/knossos/internal/mena"` to the import block. This is a new
dependency. Since `internal/mena/` is a leaf package with no transitive internal imports,
this adds zero dependency fan-out risk.

**Change 2: Replace walkMena() with source construction + mena.Walk()**

Delete the `walkMena()` function (lines 696-731). Replace with an all-rites source
construction function:

```go
// buildAllMenaSources constructs MenaSource entries for platform mena
// and all rite mena directories. Unlike BuildSourceChain (which builds
// a priority-ordered chain for the active rite), this function discovers
// ALL rites for source validation.
func buildAllMenaSources(projectRoot string) []mena.MenaSource {
    var sources []mena.MenaSource

    // Platform mena
    sources = append(sources, mena.MenaSource{
        Path: filepath.Join(projectRoot, "mena"),
    })

    // All rites (including shared)
    riteDir := filepath.Join(projectRoot, "rites")
    rites, _ := os.ReadDir(riteDir)
    for _, r := range rites {
        if r.IsDir() {
            sources = append(sources, mena.MenaSource{
                Path: filepath.Join(riteDir, r.Name(), "mena"),
            })
        }
    }

    return sources
}
```

This function lives in lint.go (not in internal/mena/) because "all rites" discovery is a
lint-specific concern. Walk() handles nonexistent directories gracefully, so sources pointing
to rites without mena/ directories are harmless.

**Change 3: Update call sites**

The three call sites (`lintDromena`, `lintMenaNamespace`, `lintLegomena`) change from:

```go
walkMena(projectRoot, ".dro.md", func(path, relPath string, data []byte) {
    ...
})
```

To:

```go
mena.Walk(sources, ".dro.md", func(entry mena.WalkEntry) {
    relPath := mustRel(projectRoot, entry.Path)
    ...use entry.Data instead of data...
})
```

Where `sources` is computed once via `buildAllMenaSources(projectRoot)`.

**Key detail -- relPath computation**: Walk returns `entry.RelPath` relative to the
MenaSource.Path. Lint needs project-root-relative paths for display (e.g.,
`rites/shared/mena/research/INDEX.dro.md`). The existing `mustRel(projectRoot, entry.Path)`
call using the absolute path achieves this. The `entry.RelPath` field is available for
consumers that want source-relative paths, but lint ignores it and computes its own.

**Change 4: Compute sources once in runLint()**

Move source construction to `runLint()` so the `[]MenaSource` is built once and passed to
the lint functions. This avoids constructing the same source list three times.

Current signature:
```go
func lintDromena(projectRoot string, report *LintReport)
func lintMenaNamespace(projectRoot string, report *LintReport)
func lintLegomena(projectRoot string, report *LintReport)
```

Updated signature:
```go
func lintDromena(projectRoot string, sources []mena.MenaSource, report *LintReport)
func lintMenaNamespace(projectRoot string, sources []mena.MenaSource, report *LintReport)
func lintLegomena(projectRoot string, sources []mena.MenaSource, report *LintReport)
```

`projectRoot` is still needed for `mustRel()` display path computation and for agent
linting (which uses `findAgentDirs`, out of scope for this change).

**Affected lines in runLint()**:

```go
func runLint(ctx *cmdContext, scope string) error {
    ...
    projectRoot := resolver.ProjectRoot()
    sources := buildAllMenaSources(projectRoot)   // NEW

    report := &LintReport{}

    if scope == "" || scope == "agents" {
        lintAgents(projectRoot, report)            // unchanged
    }
    if scope == "" || scope == "dromena" {
        lintDromena(projectRoot, sources, report)        // CHANGED
        lintMenaNamespace(projectRoot, sources, report)  // CHANGED
    }
    if scope == "" || scope == "legomena" {
        lintLegomena(projectRoot, sources, report)       // CHANGED
    }
    ...
}
```

### File-Level Change Summary

| File | Action | Lines Affected |
|------|--------|----------------|
| `internal/mena/walk.go` | NEW | ~35 lines (WalkEntry struct + Walk function) |
| `internal/cmd/lint/lint.go` | MODIFY | Import block (+1 import), `runLint()` (+2 lines), `lintDromena()` (signature + body), `lintMenaNamespace()` (signature + body), `lintLegomena()` (signature + body), `walkMena()` (DELETE entirely, -36 lines), `buildAllMenaSources()` (NEW, +16 lines) |

### Files NOT changed

| File | Reason |
|------|--------|
| `internal/mena/source.go` | MenaSource struct unchanged; Walk uses it as-is |
| `internal/mena/exists.go` | Exists() unchanged; Walk is a parallel primitive |
| `internal/mena/types.go` | No type changes needed |
| `internal/mena/platform.go` | Platform resolution unchanged |
| `internal/materialize/mena.go` | Re-export layer untouched; Walk is NOT re-exported |
| `internal/cmd/lint/lint_test.go` | Existing tests exercise checkSkillAtRefs and checkSourcePathLeaks (helper functions), not walkMena. These tests pass unchanged. |

## Backward Compatibility

### Classification: COMPATIBLE

**No external API changes**. `ari lint` command interface (flags, output format) is identical.
The only change is internal: which function iterates mena files.

**Behavioral equivalence**: `buildAllMenaSources()` produces the exact same set of
directories as the old `walkMena()`:
- `mena/` (platform) -- identical to old `filepath.Join(projectRoot, "mena")`
- `rites/*/mena/` (all rites) -- identical to old `os.ReadDir(riteDir)` loop

`filepath.WalkDir` and `filepath.Walk` produce the same file set in the same order
(lexicographic within each directory level). Lint findings are unordered, so even if order
differed, output semantics would be identical.

**Import safety**: `internal/mena/` is a leaf package (stdlib-only). Adding it as a
dependency of `internal/cmd/lint/` introduces no transitive imports and no circular
dependency risk.

**Migration**: None required. This is a pure refactoring with no schema, config, or
behavioral changes.

## Test Specification

### New test file: `/Users/tomtenuta/Code/knossos/internal/mena/walk_test.go`

Tests follow the patterns established in `exists_test.go`: use `t.TempDir()` for fixtures,
create filesystem structures, assert Walk behavior.

| Test | Setup | Assertion |
|------|-------|-----------|
| `TestWalk_MatchesSuffix` | Create source dir with `a.dro.md`, `b.lego.md`, `c.txt` | Walk with ".dro.md" invokes callback for `a.dro.md` only |
| `TestWalk_RecursesIntoSubdirs` | Create `source/sub/nested.dro.md` | Walk finds `nested.dro.md` with relPath `sub/nested.dro.md` |
| `TestWalk_SkipsNonexistentSource` | MenaSource with Path="/nonexistent" | Walk completes without error, callback never invoked |
| `TestWalk_SkipsEmptyPath` | MenaSource with Path="" | Walk completes without error, callback never invoked |
| `TestWalk_SkipsEmbeddedSource` | MenaSource with IsEmbedded=true | Walk skips embedded source, callback never invoked |
| `TestWalk_MultipleSources` | Two source dirs, each with one .dro.md file | Walk invokes callback twice (once per file) |
| `TestWalk_UnreadableFileSkipped` | Create file then chmod 000 | Walk skips unreadable file without error (skip on non-macOS only; macOS root can read anything) |
| `TestWalk_RelPathIsSourceRelative` | Source at `/tmp/X/mena/`, file at `/tmp/X/mena/session/park/INDEX.dro.md` | entry.RelPath == `session/park/INDEX.dro.md` |
| `TestWalk_IndexDirectoryFiles` | Create `source/park/INDEX.dro.md` + `source/park/behavior.md` | Walk with ".dro.md" finds INDEX.dro.md but NOT behavior.md (suffix filter). Walk with ".md" finds both. |
| `TestWalk_ReadsContent` | Create file with known content | entry.Data matches written content |

### Existing tests that must continue passing

| Test Suite | Command |
|------------|---------|
| `internal/mena/...` | `CGO_ENABLED=0 go test ./internal/mena/...` |
| `internal/cmd/lint/...` | `CGO_ENABLED=0 go test ./internal/cmd/lint/...` |

### Integration validation

| Validation | Command | Expected |
|------------|---------|----------|
| Full lint pass | `ari lint` | Exit 0, same findings count as before change |
| Scoped lint | `ari lint --scope=dromena` | Exit 0, dromena findings identical |
| Scoped lint | `ari lint --scope=legomena` | Exit 0, legomena findings identical |

### Satellite diversity matrix

| Satellite Type | Test | Expected Outcome |
|----------------|------|------------------|
| Knossos self-host (18 rites, platform mena, shared mena) | `ari lint` before and after | Identical findings count and content |
| Fixture: no rites/ directory | Walk with empty rites glob | No errors, platform mena files only |
| Fixture: rite with no mena/ dir | Walk includes source for riteless rite | Silently skipped, no error |
| Fixture: mixed dro/lego in shared | Walk with .dro.md suffix | Only .dro.md files found, .lego.md excluded |

## Design Decisions Log

| Decision | Rationale | Alternatives Rejected |
|----------|-----------|----------------------|
| WalkEntry struct over positional params | Extensible without breaking callers. Self-documenting at call sites. Follows SourceChainOptions precedent. | Positional `fn(path, relPath string, data []byte)`: not extensible, unnamed parameters at call site. |
| relPath relative to source, not projectRoot | Walk is a generic primitive. It should not know about project roots. Caller computes display paths as needed. | relPath relative to projectRoot: requires Walk to accept a projectRoot parameter, coupling it to project structure. |
| All-rites source construction in lint.go | "All rites" is a lint-specific concern. Only lint validates inactive rites. Putting this in internal/mena/ is premature generalization. | AllRiteSources() in internal/mena/: couples leaf package to project layout. WalkAll() convenience: same coupling problem. |
| filepath.WalkDir over filepath.Walk | WalkDir (Go 1.16+) avoids os.Lstat per entry. Strictly better performance, same API surface. Knossos requires Go 1.22+. | filepath.Walk: unnecessary Lstat overhead. os.ReadDir + manual recursion: more code for same result. |
| Skip embedded FS in Walk | No current consumer needs embedded FS walking. Lint is filesystem-only. Adding it doubles complexity for zero benefit. Extensible later. | Support both FS types: doubles implementation, untested code path. |
| New file walk.go over extending exists.go | Walk is a distinct primitive (iterate + read) from Exists (probe + boolean). Separate files match the package's one-concept-per-file pattern (source.go, exists.go, types.go, platform.go). | Add Walk to exists.go: conflates existence checking with content iteration. |
| buildAllMenaSources as private function | Only one caller (runLint). No reason to export. If a second consumer emerges, it can be promoted. | Exported function: premature. Method on a type: no suitable receiver. |
| Pass sources to lint functions, not construct per-call | Three call sites use the same source list. Building it once in runLint() avoids three redundant os.ReadDir calls on rites/. | Construct per function: wasteful (3x ReadDir), but functionally identical. |

## Implementation Notes for Integration Engineer

1. **walk.go is the only new file.** Everything else is modification.

2. **Walk callback must NOT return error.** The current walkMena callback is `func(path, relPath string, data []byte)` with no error return. Walk() follows this: the callback is a void function. If a lint rule encounters a problem, it appends to the report (same as today).

3. **Test the relPath bridging.** The critical correctness check is that `mustRel(projectRoot, entry.Path)` in lint.go produces the same project-relative paths as the old `mustRel(projectRoot, path)`. Since `entry.Path` IS the absolute filesystem path (same value as old `path`), this is guaranteed. But test it explicitly via `ari lint` diff.

4. **Do not change findAgentDirs().** The gap analysis notes that `findAgentDirs()` (line 265-279) has a parallel pattern (walks `rites/*/agents/`). This is out of scope. Agent resolution does not live in `internal/mena/` and has no shared package to unify with.

5. **Do not re-export Walk.** The re-export layer in `internal/materialize/mena.go` exists for backward compatibility with callers that import `internal/materialize`. Lint imports `internal/mena/` directly. Adding a Walk re-export would suggest that materialization callers should use Walk, which is not the intent.

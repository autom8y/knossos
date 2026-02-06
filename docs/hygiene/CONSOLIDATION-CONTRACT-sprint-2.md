# Sprint 2 Consolidation Contract

**Initiative**: Knossos Code Hygiene - Pattern Consolidation
**Phase**: Sprint 2 - Merge Duplicates, Remove Deferrals
**Date**: 2026-02-06
**Architect**: architect-enforcer
**Upstream**: Sprint 1 deleted 1,272 LOC of dead code. Sprint 2 consolidates surviving duplication.

## Verification Method

Every target was verified by reading source files via the Read tool with line-level precision. Code snippets below are exact copies from the current codebase. No target is theoretical.

## Architectural Assessment

The hook command package (`internal/cmd/hook/`) has a systemic pattern problem: each hook command has a production `runX()` function and a near-identical test-only `runXWithPrinter()` function. This arose because production functions call `ctx.getPrinter()` internally, making printer injection impossible without duplication. The root cause is **printer creation coupled to command execution**.

Batch sequencing exploits this: Batch 10 extracts the resolver bootstrap, Batch 11 establishes the printer-injection pattern on clew (the worst offender), Batch 12 propagates it to remaining commands, Batch 13 aligns the printer types, and Batch 14 is independent cleanup.

---

## Batch 9: USE_ARI_HOOKS Feature Flag Removal

**Priority**: P0 (deferred from Sprint 1 REWIRE-002, per ADR-0011 Phase 2)
**Risk**: TRIVIAL
**Dependencies**: None (independent)
**LOC Impact**: ~30 lines removed from Go code, ~20 lines from tests

### Current State

Sprint 1 already removed `FeatureFlagEnvVar`, `IsEnabled()`, and `shouldEarlyExit()` from Go production code. The feature flag constant and function no longer exist in `internal/hook/env.go` or `internal/cmd/hook/hook.go`.

**However, 14 stale `USE_ARI_HOOKS` references remain in test files and benchmarks:**

```go
// internal/cmd/hook/context_test.go:348-352
os.Setenv("USE_ARI_HOOKS", "1")
// ...
os.Unsetenv("USE_ARI_HOOKS")

// internal/cmd/hook/autopark_test.go:332
os.Unsetenv("USE_ARI_HOOKS")

// internal/cmd/hook/writeguard_test.go:418,423,461
os.Setenv("USE_ARI_HOOKS", "1")
os.Unsetenv("USE_ARI_HOOKS")

// internal/cmd/hook/validate_test.go:639,644,682,716,721
os.Setenv("USE_ARI_HOOKS", "1")
os.Unsetenv("USE_ARI_HOOKS")

// internal/cmd/hook/route_test.go:671,675,711,715
os.Setenv("USE_ARI_HOOKS", "1")
os.Unsetenv("USE_ARI_HOOKS")
```

These `os.Setenv("USE_ARI_HOOKS", "1")` / `os.Unsetenv("USE_ARI_HOOKS")` calls are now no-ops. The env var is never read. They exist only in benchmark setup/teardown blocks.

### Target State

- Delete all `os.Setenv("USE_ARI_HOOKS", ...)` and `os.Unsetenv("USE_ARI_HOOKS")` lines from test files
- No Go source file references `USE_ARI_HOOKS` (docs are out of scope)

### Files

| File | Action |
|------|--------|
| `internal/cmd/hook/context_test.go` | Remove lines 348, 352 |
| `internal/cmd/hook/autopark_test.go` | Remove line 332 |
| `internal/cmd/hook/writeguard_test.go` | Remove lines 418, 423, 461 |
| `internal/cmd/hook/validate_test.go` | Remove lines 639, 644, 682, 716, 721 |
| `internal/cmd/hook/route_test.go` | Remove lines 671, 675, 711, 715 |

### Invariants

- No behavior change (env var was already unread)
- All existing tests pass without modification

### Verification

```bash
CGO_ENABLED=0 go test ./internal/cmd/hook/...
grep -r 'USE_ARI_HOOKS' internal/ test/ --include='*.go'  # Should return zero results
```

### Rollback

Revert single commit. No dependencies.

---

## Batch 10: Hook Resolver Bootstrap Extraction

**Priority**: P1 (enables Batches 11-12)
**Risk**: LOW
**Dependencies**: None (independent, but must complete before Batch 11)
**LOC Impact**: ~30 lines removed (net, after extraction)

### Current State

Three hook commands repeat the same resolver + session bootstrap pattern:

**`internal/cmd/hook/context.go` lines 82-105:**
```go
// Get resolver for path lookups
resolver := ctx.GetResolver()
if resolver.ProjectRoot() == "" {
    // Try to discover project from environment
    if hookEnv.ProjectDir != "" {
        resolver = newResolverFromPath(hookEnv.ProjectDir)
    } else {
        return outputNoSession(printer)
    }
}

// Get current session ID
sessionID, err := ctx.GetCurrentSessionID()
if err != nil {
    printer.VerboseLog("warn", "failed to read current session", map[string]interface{}{"error": err.Error()})
    return outputNoSession(printer)
}

if sessionID == "" {
    return outputNoSession(printer)
}

// Trim any whitespace/newlines from session ID
sessionID = strings.TrimSpace(sessionID)
```

**`internal/cmd/hook/autopark.go` lines 71-94:**
```go
// Get resolver for path lookups
resolver := ctx.GetResolver()
if resolver.ProjectRoot() == "" {
    // Try to discover project from environment
    if hookEnv.ProjectDir != "" {
        resolver = paths.NewResolver(hookEnv.ProjectDir)
    } else {
        return outputNoPark(printer, "no project context")
    }
}

// Get current session ID
sessionID, err := ctx.GetCurrentSessionID()
if err != nil {
    printer.VerboseLog("warn", "failed to read current session", map[string]interface{}{"error": err.Error()})
    return outputNoPark(printer, "no active session")
}

if sessionID == "" {
    return outputNoPark(printer, "no active session")
}

// Trim any whitespace/newlines from session ID
sessionID = strings.TrimSpace(sessionID)
```

**`internal/cmd/hook/clew.go` lines 155-177 (`getSessionDir`):**
```go
func getSessionDir(ctx *cmdContext, hookEnv *hook.Env) string {
    // Try to get session ID from context
    sessionID, err := ctx.GetCurrentSessionID()
    if err != nil || sessionID == "" {
        return ""
    }

    sessionID = strings.TrimSpace(sessionID)

    // Get resolver for path lookups
    resolver := ctx.GetResolver()
    if resolver.ProjectRoot() == "" {
        // Try to discover project from environment
        if hookEnv.ProjectDir != "" {
            resolver = newResolverFromPath(hookEnv.ProjectDir)
        } else {
            return ""
        }
    }

    // Return the session directory path
    return resolver.SessionDir(sessionID)
}
```

The pattern is identical: GetResolver -> fallback to hookEnv.ProjectDir -> GetCurrentSessionID -> TrimSpace. Only the error-return shape differs (the callers handle "no session" differently).

### Target State

Add a method to `*cmdContext` in `internal/cmd/hook/hook.go`:

```go
// resolveSession resolves the path resolver and session ID from context and hook environment.
// Returns (resolver, sessionID, ok). When ok is false, no session is available.
func (c *cmdContext) resolveSession(hookEnv *hook.Env) (*paths.Resolver, string, bool) {
    resolver := c.GetResolver()
    if resolver.ProjectRoot() == "" {
        if hookEnv.ProjectDir != "" {
            resolver = paths.NewResolver(hookEnv.ProjectDir)
        } else {
            return nil, "", false
        }
    }

    sessionID, err := c.GetCurrentSessionID()
    if err != nil || sessionID == "" {
        return resolver, "", false
    }

    return resolver, strings.TrimSpace(sessionID), true
}
```

Then replace the three occurrences:

- `context.go`: Replace lines 82-105 with `resolver, sessionID, ok := ctx.resolveSession(hookEnv); if !ok { return outputNoSession(printer) }`
- `autopark.go`: Replace lines 71-94 with `resolver, sessionID, ok := ctx.resolveSession(hookEnv); if !ok { return outputNoPark(printer, "no active session") }`
- `clew.go`: Replace `getSessionDir` body (lines 155-177) to use `resolveSession` then call `resolver.SessionDir(sessionID)`

### Files

| File | Action |
|------|--------|
| `internal/cmd/hook/hook.go` | Add `resolveSession` method (~15 lines) |
| `internal/cmd/hook/context.go` | Replace lines 82-105 with 3-line call |
| `internal/cmd/hook/autopark.go` | Replace lines 71-94 with 3-line call |
| `internal/cmd/hook/clew.go` | Simplify `getSessionDir` to use `resolveSession` |

### Invariants

- `resolveSession` returns the same resolver and sessionID as the inline code
- Error paths produce the same caller-visible behavior
- `newResolverFromPath` in context.go is identical to `paths.NewResolver` (verified: context.go:162-164 wraps `paths.NewResolver`)

### Verification

```bash
CGO_ENABLED=0 go test ./internal/cmd/hook/...
CGO_ENABLED=0 go vet ./internal/cmd/hook/...
```

### Rollback

Revert single commit. Method extraction is additive; reverting restores inline code.

---

## Batch 11: Clew Command Body Deduplication

**Priority**: P1 (worst single-file duplication in the hook package)
**Risk**: LOW
**Dependencies**: Batch 10 (resolver extraction simplifies both bodies)
**LOC Impact**: ~80 lines removed

### Current State

`internal/cmd/hook/clew.go` contains two near-identical function bodies:

**`runClew` (lines 73-152, 80 lines)** -- production, calls `ctx.getPrinter()`
**`runClewWithPrinter` (lines 189-267, 79 lines)** -- test helper, accepts `printer *output.Printer`

The bodies are character-for-character identical except:
1. `runClew` starts with `printer := ctx.getPrinter()` while `runClewWithPrinter` receives printer as param
2. `runClew` calls `outputNotRecorded(printer, ...)` while `runClewWithPrinter` calls `outputNotRecordedWithPrinter(printer, ...)`

Similarly, two output helpers are duplicated:

**`outputNotRecorded` (lines 179-186):**
```go
func outputNotRecorded(printer *output.Printer, reason string) error {
    result := ClewOutput{Recorded: false, Reason: reason}
    return printer.Print(result)
}
```

**`outputNotRecordedWithPrinter` (lines 270-276):**
```go
func outputNotRecordedWithPrinter(printer *output.Printer, reason string) error {
    result := ClewOutput{Recorded: false, Reason: reason}
    return printer.Print(result)
}
```

These are **identical functions** with different names. Both accept `*output.Printer`.

### Target State

1. Delete `runClewWithPrinter` entirely
2. Delete `outputNotRecordedWithPrinter` entirely
3. Refactor `runClew` to accept an optional printer parameter via a new internal function:

```go
func runClew(ctx *cmdContext) error {
    return runClewCore(ctx, ctx.getPrinter())
}

func runClewCore(ctx *cmdContext, printer *output.Printer) error {
    // ... single copy of the logic, using outputNotRecorded throughout ...
}
```

4. Update `clew_test.go` to call `runClewCore` instead of `runClewWithPrinter`

### Files

| File | Action |
|------|--------|
| `internal/cmd/hook/clew.go` | Delete `runClewWithPrinter` (lines 189-267), delete `outputNotRecordedWithPrinter` (lines 270-276), split `runClew` into `runClew` + `runClewCore` |
| `internal/cmd/hook/clew_test.go` | Replace `runClewWithPrinter(ctx, printer)` calls with `runClewCore(ctx, printer)` |

### Invariants

- `runClewCore` has identical logic to current `runClew` minus the `printer := ctx.getPrinter()` line
- `outputNotRecorded` is unchanged and used by both production and test paths
- All existing clew tests pass without modification (only function name changes)
- No change to cobra RunE wiring

### Verification

```bash
CGO_ENABLED=0 go test ./internal/cmd/hook/... -run TestClew
CGO_ENABLED=0 go test ./internal/cmd/hook/... -run TestRunClew
```

### Rollback

Revert single commit.

---

## Batch 12: Test Helper Consolidation (Remaining Commands)

**Priority**: P2 (propagates Batch 11 pattern to 4 more commands)
**Risk**: LOW
**Dependencies**: Batch 10, Batch 11 (establishes the pattern)
**LOC Impact**: ~200 lines removed

### Current State

Five hook commands have `run*WithPrinter` test duplicates. Batch 11 handles clew. Four remain:

| Command | Production Function | Test Duplicate | Location |
|---------|-------------------|----------------|----------|
| context | `runContext` (context.go:69-137, 69 lines) | `runContextWithPrinter` (context_test.go:388-440, 53 lines) | Test file |
| autopark | `runAutopark` (autopark.go:58-150, 93 lines) | `runAutoparkWithPrinter` (autopark_test.go:367-434, 68 lines) | Test file |
| validate | `runValidate` (validate.go:95-133, 39 lines) | `runValidateWithPrinter` (validate.go:216-253, 38 lines) | **Production file** |
| route | `runRoute` (route.go:97-131, 35 lines) | `runRouteWithPrinter` (route.go:170-208, 39 lines) | **Production file** |

Note: `runValidateWithPrinter` and `runRouteWithPrinter` live in production source files, not test files. Additionally, `runRouteWithPrinter` has a companion `outputNotRoutedWithPrinter` (route.go:204-208) that is identical to `outputNotRouted` (route.go:163-167).

The autopark test duplicate (`runAutoparkWithPrinter`) is notably divergent -- it does raw `bytes.Replace` on YAML instead of using the `session` package. This is a behavior divergence, not just a duplication. The tests verify against a different implementation than production.

### Target State

Apply the same pattern as Batch 11 to each command:

**context.go:**
```go
func runContext(ctx *cmdContext) error {
    return runContextCore(ctx, ctx.getPrinter())
}

func runContextCore(ctx *cmdContext, printer *output.Printer) error {
    // ... existing runContext body minus printer creation ...
}
```

**autopark.go:**
```go
func runAutopark(ctx *cmdContext) error {
    return runAutoparkCore(ctx, ctx.getPrinter())
}

func runAutoparkCore(ctx *cmdContext, printer *output.Printer) error {
    // ... existing runAutopark body minus printer creation ...
}
```

**validate.go:**
```go
func runValidate(ctx *cmdContext) error {
    return runValidateCore(ctx, ctx.getPrinter(), "")
}

func runValidateCore(ctx *cmdContext, printer *output.Printer, stdinInput string) error {
    // ... existing runValidate body + stdin fallback from test version ...
}
```
Note: `runValidateWithPrinter` accepts `stdinInput string` for test stdin simulation. The core function preserves this parameter.

**route.go:**
```go
func runRoute(ctx *cmdContext) error {
    return runRouteCore(ctx, ctx.getPrinter())
}

func runRouteCore(ctx *cmdContext, printer *output.Printer) error {
    // ... existing runRoute body minus printer creation ...
}
```

Delete:
- `runContextWithPrinter` from context_test.go
- `runAutoparkWithPrinter` from autopark_test.go
- `runValidateWithPrinter` from validate.go
- `runRouteWithPrinter` and `outputNotRoutedWithPrinter` from route.go

Update test call sites to use `runXCore` functions.

### Files

| File | Action |
|------|--------|
| `internal/cmd/hook/context.go` | Split `runContext` into `runContext` + `runContextCore` |
| `internal/cmd/hook/context_test.go` | Delete `runContextWithPrinter` (lines 388-440), update callers to `runContextCore` |
| `internal/cmd/hook/autopark.go` | Split `runAutopark` into `runAutopark` + `runAutoparkCore` |
| `internal/cmd/hook/autopark_test.go` | Delete `runAutoparkWithPrinter` (lines 367-434), update callers to `runAutoparkCore` |
| `internal/cmd/hook/validate.go` | Delete `runValidateWithPrinter` (lines 216-253), split `runValidate` into `runValidate` + `runValidateCore` |
| `internal/cmd/hook/validate_test.go` | Update callers to `runValidateCore` |
| `internal/cmd/hook/route.go` | Delete `runRouteWithPrinter` (lines 170-208), delete `outputNotRoutedWithPrinter` (lines 204-208), split `runRoute` into `runRoute` + `runRouteCore` |
| `internal/cmd/hook/route_test.go` | Update callers to `runRouteCore` |

### Invariants

- Production `RunE` functions call `runX` -> `runXCore(ctx, ctx.getPrinter())` -- no behavioral change
- Test functions call `runXCore(ctx, injectedPrinter)` -- same injection, just using production code path
- The autopark test will now exercise the real `session.LoadContext` / `session.Save` path instead of the divergent `bytes.Replace` hack. This is BETTER -- tests become more meaningful.
- `outputValidateAllow` and `outputValidateBlock` in validate.go currently accept `interface{ Print(interface{}) error }` -- this interface typing is addressed in Batch 13

### Verification

```bash
CGO_ENABLED=0 go test ./internal/cmd/hook/...
CGO_ENABLED=0 go vet ./internal/cmd/hook/...
```

### Rollback

Revert single commit. Can also be split into per-command commits for granular rollback.

---

## Batch 13: Output Printer Interface Narrowing

**Priority**: P2 (style/consistency, builds on Batch 12)
**Risk**: TRIVIAL
**Dependencies**: Batch 12 (which removes `runValidateWithPrinter` that requires the interface)
**LOC Impact**: ~10 lines changed (type signatures only)

### Current State

Three hook commands use ad-hoc anonymous interface types for printer parameters:

**`internal/cmd/hook/validate.go` lines 199, 207:**
```go
func outputValidateAllow(printer interface{ Print(interface{}) error }) error {
func outputValidateBlock(printer interface{ Print(interface{}) error }, reason string) error {
```

**`internal/cmd/hook/writeguard.go` lines 200, 208:**
```go
func outputAllow(printer interface{ Print(interface{}) error }) error {
func outputBlock(printer interface{ Print(interface{}) error }, filePath string) error {
```

**`internal/cmd/hook/autopark.go` line 153:**
```go
func outputNoPark(printer interface{ Print(interface{}) error }, reason string) error {
```

Meanwhile, other output helper functions in the same package use `*output.Printer` directly:

```go
// clew.go:180
func outputNotRecorded(printer *output.Printer, reason string) error {

// context.go:140
func outputNoSession(printer *output.Printer) error {

// route.go:164
func outputNotRouted(printer *output.Printer) error {
```

The ad-hoc interfaces were introduced because `runValidateWithPrinter` accepted `interface{ Print(interface{}) error }` as its printer type. After Batch 12 removes these test-only functions, there is no remaining reason for the interface -- all callers pass `*output.Printer`.

### Target State

Replace all `interface{ Print(interface{}) error }` parameter types with `*output.Printer`:

```go
// validate.go
func outputValidateAllow(printer *output.Printer) error {
func outputValidateBlock(printer *output.Printer, reason string) error {

// writeguard.go
func outputAllow(printer *output.Printer) error {
func outputBlock(printer *output.Printer, filePath string) error {

// autopark.go
func outputNoPark(printer *output.Printer, reason string) error {
```

### Files

| File | Action |
|------|--------|
| `internal/cmd/hook/validate.go` | Change printer param type on lines 199, 207 |
| `internal/cmd/hook/writeguard.go` | Change printer param type on lines 200, 208 |
| `internal/cmd/hook/autopark.go` | Change printer param type on line 153 |

### Invariants

- `*output.Printer` satisfies `interface{ Print(interface{}) error }`, so this is a narrowing (more specific type), not a widening
- No call sites change -- they all already pass `*output.Printer`
- The anonymous interface has zero external callers (package-internal functions only)

### Verification

```bash
CGO_ENABLED=0 go build ./internal/cmd/hook/...
CGO_ENABLED=0 go vet ./internal/cmd/hook/...
```

### Rollback

Revert single commit. Type narrowing is trivially reversible.

---

## Batch 14: Frontmatter Test Inline Parsing

**Priority**: P3 (low impact, independent)
**Risk**: TRIVIAL
**Dependencies**: None (independent)
**LOC Impact**: ~5 lines changed (helper extraction), net neutral

### Current State

`internal/materialize/frontmatter_test.go` lines 107-127 contain inline frontmatter extraction logic:

```go
// Inline frontmatter parsing (ParseMenaFrontmatter was deleted per Sprint 1 Batch 5)
if !bytes.HasPrefix(content, []byte("---\n")) && !bytes.HasPrefix(content, []byte("---\r\n")) {
    return err
}

// Find closing delimiter
var endIndex int
if idx := bytes.Index(content[4:], []byte("\n---\n")); idx != -1 {
    endIndex = idx
} else if idx := bytes.Index(content[4:], []byte("\n---\r\n")); idx != -1 {
    endIndex = idx
} else if idx := bytes.Index(content[4:], []byte("\r\n---\r\n")); idx != -1 {
    endIndex = idx
} else if idx := bytes.Index(content[4:], []byte("\r\n---\n")); idx != -1 {
    endIndex = idx
} else {
    return err
}

frontmatterBytes := content[4 : 4+endIndex]
```

This code has a subtle bug: `return err` on lines 109 and 123 returns the error from `os.ReadFile` which may be `nil` (file read succeeded but has no frontmatter). The intent is to return an error indicating "no frontmatter found" but it returns `nil` instead, silently passing validation for files without frontmatter delimiters.

### Assessment: SKIP

**Rationale**: This inline parsing is 20 lines in a single test helper function. It is contained, has no duplication, and extracting it would just move the same logic to a different function in the same file. The real fix is the `return err` bug (should be `return fmt.Errorf("no frontmatter found")`), but that is a **behavior change** to test code, not a consolidation refactor.

**Recommendation**: Fix the `return err` -> `return fmt.Errorf(...)` bug as part of a separate test-fix commit, not as a consolidation target. The inline parsing itself is not worth extracting -- it is used in exactly one place and is readable in context.

---

## Batch Sequencing Summary

```
Batch 9:  USE_ARI_HOOKS cleanup ─────────────────────────── (independent)
Batch 10: Resolver extraction ────────────┐                 (independent)
Batch 11: Clew dedup ────────────────────┤                 (depends on 10)
Batch 12: Test helper consolidation ──────┤                 (depends on 10, 11)
Batch 13: Printer interface narrowing ────┘                 (depends on 12)
Batch 14: Frontmatter test ──────────────────────────────── SKIP
```

### Recommended Commit Sequence

| Order | Batch | Commit Message | Rollback Point |
|-------|-------|---------------|----------------|
| 1 | 9 | `refactor: remove stale USE_ARI_HOOKS env var references from tests` | Safe checkpoint |
| 2 | 10 | `refactor: extract resolveSession method on cmdContext` | Safe checkpoint |
| 3 | 11 | `refactor: deduplicate runClew/runClewWithPrinter into runClewCore` | Safe checkpoint |
| 4 | 12 | `refactor: consolidate run*WithPrinter test helpers into production *Core functions` | Safe checkpoint |
| 5 | 13 | `refactor: narrow ad-hoc printer interfaces to concrete *output.Printer type` | Safe checkpoint |

Each commit is independently revertible. If any batch fails verification, revert that commit only -- no cascading rollback needed except for Batch 13 which depends on Batch 12's interface cleanup.

## Risk Matrix

| Batch | Blast Radius | Failure Detection | Recovery Cost | Overall |
|-------|-------------|-------------------|---------------|---------|
| 9 | Test files only | `go test` | Single revert | TRIVIAL |
| 10 | 3 production files + hook.go | `go test` + manual review | Single revert | LOW |
| 11 | 2 files (clew.go, clew_test.go) | `go test -run TestClew` | Single revert | LOW |
| 12 | 8 files (4 commands x 2) | `go test ./internal/cmd/hook/...` | Single revert or per-command revert | LOW |
| 13 | 3 production files (type sigs only) | `go build` | Single revert | TRIVIAL |

## LOC Impact Summary

| Batch | Lines Removed | Lines Added | Net |
|-------|--------------|-------------|-----|
| 9 | ~30 | 0 | -30 |
| 10 | ~60 | ~20 | -40 |
| 11 | ~90 | ~5 | -85 |
| 12 | ~220 | ~30 | -190 |
| 13 | 0 | 0 | ~0 (type changes only) |
| **Total** | **~400** | **~55** | **~-345** |

## Handoff Checklist

- [x] Every smell classified (5 addressed, 1 skipped with rationale)
- [x] Each refactoring has before/after contract documented
- [x] Invariants and verification criteria specified
- [x] Refactorings sequenced with explicit dependencies
- [x] Rollback points identified between phases
- [x] Risk assessment complete for each phase

## Janitor Notes

1. **Commit convention**: Use `refactor:` prefix per existing repo convention (see recent commits).
2. **Test after every batch**: Run `CGO_ENABLED=0 go test ./internal/cmd/hook/...` after each commit. Do not proceed to next batch if tests fail.
3. **Batch 12 autopark divergence**: The current `runAutoparkWithPrinter` in tests uses `bytes.Replace` on raw YAML instead of the `session` package. After switching tests to `runAutoparkCore`, the tests will exercise the real session package code. If any autopark tests fail, the failure reveals a pre-existing test/production divergence, not a regression from this refactor.
4. **Batch 12 validate.go stdin**: `runValidateWithPrinter` accepts `stdinInput string` -- preserve this parameter in `runValidateCore` for test stdin simulation.
5. **Do not touch budget.go**: `runBudget` does not have a `WithPrinter` duplicate. It uses production code in tests via `newTestContext`. This is the correct pattern that Batches 11-12 are propagating to other commands.
6. **Batch 14 is SKIP**: Do not implement. The frontmatter test issue is a minor bug, not a consolidation target. File a separate issue if desired.

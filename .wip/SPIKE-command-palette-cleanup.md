# Spike: Command Palette Cleanup

**Date**: 2026-02-11
**Investigator**: Integration Engineer
**Status**: ✅ Complete

## Problem Statement

User reported two issues with CC command palette:
1. `/qa:reference` — companion file appearing in palette (should be hidden)
2. `/qa:INDEX` — INDEX file showing with `:INDEX` suffix (should display as `/qa`)

## Investigation

### Audit Scope
- Examined all files in `.claude/commands/` (rite-scope)
- Checked all companion files with `user-invocable: false` frontmatter
- Traced pipeline code in `internal/materialize/mena.go`
- Verified INDEX.md files have correct `name:` frontmatter

### Findings

#### Finding #1: Malformed `user-invocable: false` Frontmatter
**Severity**: HIGH
**Affected Files**: All companion files with pre-existing frontmatter (20+ files)

**Root Cause**: `injectCompanionHideFrontmatter()` in `mena.go` was incorrectly calculating the insertion point for `user-invocable: false` field.

The function found the position of `\n---\n` (closing delimiter) and inserted the field at `endIndex`, which pointed to the `\n` before `---`. However, `content[:endIndex]` slices *up to but not including* that position, which meant it was cutting off the final newline of the last frontmatter field.

**Example of Malformed Output**:
```yaml
---
name: qa-ref
description: "..."user-invocable: false
---
```

The `user-invocable: false` was concatenated onto the same line as the description, making the YAML invalid. CC couldn't parse this properly, so it fell back to showing the file in the palette.

**Files Affected**:
- `/qa/reference.md`
- `/consult/reference.md`
- All other companion files with pre-existing frontmatter

#### Finding #2: INDEX.md Files Show Correctly
**Status**: No issue found
**Verification**: All INDEX.md files have proper `name:` frontmatter fields matching their command names. CC should display these correctly as `/qa`, `/commit`, etc.

### Code Analysis

**File**: `internal/materialize/mena.go`
**Function**: `injectCompanionHideFrontmatter(content []byte) []byte`
**Lines**: 612-662

**Bug Location**: Lines 618-624 (and similar blocks for CRLF)

```go
if idx := bytes.Index(content[searchStart:], []byte("\n---\n")); idx != -1 {
    endIndex = searchStart + idx  // BUG: Points to \n before ---
    result := make([]byte, 0, len(content)+len("user-invocable: false\n"))
    result = append(result, content[:endIndex]...)  // Slices OFF the newline
    result = append(result, []byte("user-invocable: false\n")...)
    result = append(result, content[endIndex:]...)
    return result
}
```

**Fix**: Add offset to include the newline character:

```go
endIndex = searchStart + idx + 1  // +1 to include the \n before ---
```

This ensures the newline after the last field is preserved, and `user-invocable: false` starts on a new line.

## Solution

### Changes Made

**File**: `internal/materialize/mena.go`
**Lines Modified**: 619, 627, 639, 646

Changed all four delimiter detection branches to include the newline offset:
- Line 619: `endIndex = searchStart + idx + 1` (for `\n---\n`)
- Line 627: `endIndex = searchStart + idx + 1` (for `\n---\r\n`)
- Line 639: `endIndex = searchStart + idx + 2` (for `\r\n---\r\n`)
- Line 646: `endIndex = searchStart + idx + 2` (for `\r\n---\n`)

### Verification Steps

1. **Rebuilt binary**:
   ```bash
   CGO_ENABLED=0 go build ./cmd/ari
   cp ./ari $(which ari)
   ```

2. **Regenerated commands**:
   ```bash
   ari sync --scope=rite
   ```

3. **Verified frontmatter format**:
   - Checked `qa/reference.md` — properly formatted ✅
   - Checked `consult/reference.md` — properly formatted ✅
   - Audited all companion files with `user-invocable: false` — all properly formatted ✅

4. **Byte-level verification**:
   ```bash
   hexdump -C .claude/commands/qa/reference.md | head -20
   ```
   Confirmed `0a` (newline) byte between description and `user-invocable: false`.

### Results

**Before Fix**:
```yaml
description: "..."user-invocable: false

---
```

**After Fix**:
```yaml
description: "..."
user-invocable: false
---
```

All 20+ companion files now have properly formatted frontmatter with `user-invocable: false` on its own line.

## Testing

### Manual Testing
- ✅ Build succeeds: `CGO_ENABLED=0 go build ./cmd/ari`
- ✅ Sync succeeds: `ari sync --scope=rite` (no errors)
- ✅ Frontmatter validation: All companion files have valid YAML
- ✅ Byte-level verification: Newlines present in correct positions

### Files Verified
- `.claude/commands/qa/reference.md` — PASS
- `.claude/commands/qa/INDEX.md` — PASS
- `.claude/commands/commit/behavior.md` — PASS
- `.claude/commands/commit/examples.md` — PASS
- `.claude/commands/consult/reference.md` — PASS
- All other companion files — PASS

## Impact Assessment

### Scope
- **Files Modified**: 1 Go source file (`internal/materialize/mena.go`)
- **Files Regenerated**: 20+ companion files in `.claude/commands/`
- **Breaking Changes**: None
- **Migration Required**: No (automatic on next `ari sync`)

### User Experience Impact
- **Before**: Companion files appeared in CC command palette due to malformed frontmatter
- **After**: Companion files properly hidden via `user-invocable: false`
- **Expected**: User types `/qa` and sees only one entry: `/qa — Validation-only with review and approval`

## Remaining Questions

1. **CC Palette Behavior**: Does CC actually parse `user-invocable: false` to hide companion files?
   - **Status**: Assumed yes, based on CC frontmatter schema
   - **Verification**: User should test by typing `/qa` in CC command palette
   - **If still showing**: May need to investigate CC's command registration logic

2. **INDEX.md Naming**: Does CC use the `name:` field or file paths for directory-based commands?
   - **Status**: All INDEX.md files have correct `name:` fields
   - **Expected**: CC should show `/qa`, not `/qa:INDEX`
   - **If still showing with `:INDEX` suffix**: May be a CC bug or configuration issue

## Recommendations

1. **Immediate**: User should verify CC command palette now shows clean entries
2. **Short-term**: Add unit tests for `injectCompanionHideFrontmatter()` to prevent regression
3. **Long-term**: Consider adding CI check to validate all companion files have proper frontmatter

## Files Modified

| File | Type | Lines Changed |
|------|------|---------------|
| `internal/materialize/mena.go` | Source | 4 lines (offsets added) |

## Files Verified

| File | Status |
|------|--------|
| `internal/materialize/mena.go` | ✅ Read back, changes confirmed |
| `.claude/commands/qa/INDEX.md` | ✅ Proper frontmatter |
| `.claude/commands/qa/reference.md` | ✅ Fixed frontmatter |
| `.claude/commands/commit/behavior.md` | ✅ Fixed frontmatter |
| `.claude/commands/commit/examples.md` | ✅ Fixed frontmatter |
| `.claude/commands/consult/reference.md` | ✅ Fixed frontmatter |
| All other companion files | ✅ All passing |

## Next Steps

1. ✅ **DONE**: Fix `injectCompanionHideFrontmatter()` offset calculation
2. ✅ **DONE**: Rebuild and install `ari` binary
3. ✅ **DONE**: Run `ari sync --scope=rite` to regenerate commands
4. ✅ **DONE**: Verify all companion files have proper frontmatter
5. ⏳ **USER**: Test CC command palette by typing `/qa` and confirming clean display
6. ⏳ **FUTURE**: Add unit tests for frontmatter injection logic

## Conclusion

**Root cause identified and fixed**: `injectCompanionHideFrontmatter()` was incorrectly slicing frontmatter content, causing `user-invocable: false` to concatenate onto the previous line instead of being on its own line.

**Fix applied**: Added offset (+1 for `\n`, +2 for `\r\n`) to include the newline character when slicing content.

**Verification**: All companion files now have properly formatted YAML frontmatter with `user-invocable: false` on its own line.

**User action required**: Test CC command palette to confirm companion files are now hidden and commands display with correct names.

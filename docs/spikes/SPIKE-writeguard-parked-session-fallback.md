# SPIKE: Writeguard Session ID Extraction from File Paths

**Status**: COMPLETE
**Date**: 2026-02-18
**Duration**: 1 hour
**Scope**: Enable writeguard to allow Moirai writes to PARKED sessions by extracting session IDs from target file paths

## Executive Summary

The writeguard hook currently blocks Moirai writes to PARKED sessions because session resolution only finds ACTIVE sessions. The target file path always contains the session ID (e.g., `.sos/sessions/session-20260209-120000-abcdef01/SESSION_CONTEXT.md`), making file-path-based extraction a reliable, low-risk fallback.

### Key Findings

1. **Bug confirmed**: PARKED sessions with valid Moirai locks are denied writes because `ResolveSession()` calls `FindActiveSessions()` which only scans for `status: ACTIVE`
2. **File path always contains session ID**: Protected files live at `.sos/sessions/{session-id}/SESSION_CONTEXT.md` -- the session ID is structurally embedded
3. **Lock validation is already correct**: `isMoiraiLockHeld()` checks agent, staleness, and reads from the correct session directory -- it just never gets called because `sessionID` is empty
4. **Fix is surgical**: 5-10 lines in `writeguard.go`, no changes to `ResolveSession()` or the session FSM

### Decision

**Recommended**: Implement file-path session ID extraction as a fallback in writeguard, scoped only to the Moirai lock check path.

---

## 1. Problem Analysis

### 1.1 Current Flow (writeguard.go:62-99)

```
PreToolUse(Write, ".sos/sessions/{session-id}/SESSION_CONTEXT.md")
  -> isProtectedFile() -> true
  -> resolveSession()
    -> ResolveSession(resolver, ccSessionID, explicitID)
      -> Priority 1: explicit flag -> empty (hook context, no CLI flag)
      -> Priority 2: CC map lookup -> may or may not resolve
      -> Priority 3: FindActiveSessions() -> only finds ACTIVE sessions
  -> sessionID == "" for PARKED sessions
  -> outputBlock() -- Moirai is denied even with valid lock
```

### 1.2 Why PARKED Sessions Are Invisible

`FindActiveSessions()` in `internal/session/discovery.go:58-88` scans session directories and filters to `status == "ACTIVE"` only. This is correct behavior for general session resolution (you don't want random commands operating on parked sessions), but it creates a gap for writeguard's specific use case.

### 1.3 When This Matters

The Moirai agent writes to PARKED sessions during:
- **`park_session`**: Transitions status ACTIVE -> PARKED (writes to SESSION_CONTEXT.md)
- **`resume_session`**: Transitions status PARKED -> ACTIVE (writes to SESSION_CONTEXT.md)
- **`handoff`**: Records handoff metadata in a session that may be parked
- **`update_field`**: Updates metadata fields on any session state

The park operation itself is the critical path: Moirai acquires a lock, writes the PARKED status, but the session is already PARKED by the time the FSM validates -- creating a chicken-and-egg problem where writeguard blocks the very write that would make the session visible again.

**Correction**: The park operation itself writes ACTIVE -> PARKED, so the session IS active at that point. The real failure is on **resume**: Moirai tries to write PARKED -> ACTIVE, but the session is still PARKED when writeguard fires, so `FindActiveSessions()` misses it. The `ari session lock` command also has this issue -- `GetSessionID()` in `common/context.go:52` calls `FindActiveSession()` (singular), which also only finds ACTIVE sessions.

## 2. Proposed Solution

### 2.1 Extract Session ID from File Path (Recommended)

Add a fallback in `runWriteguardCore()` that extracts the session ID from the target file path when `resolveSession()` returns empty. The session ID is already structurally embedded in the path.

**Target location**: `internal/cmd/hook/writeguard.go`, between the `resolveSession()` call and the `outputBlock()` call.

```go
// If session resolution failed but file path contains a session ID,
// extract it as fallback for Moirai lock validation.
// This enables writes to PARKED sessions with valid Moirai locks.
if sessionID == "" {
    sessionID = extractSessionIDFromPath(filePath)
}
```

New helper function:

```go
// extractSessionIDFromPath extracts a session ID from a file path.
// Looks for path segments matching the session-YYYYMMDD-HHMMSS-{hex} pattern.
// Returns empty string if no session ID found.
func extractSessionIDFromPath(filePath string) string {
    parts := strings.Split(filepath.ToSlash(filePath), "/")
    for _, part := range parts {
        if paths.IsSessionDir(part) {
            return part
        }
    }
    return ""
}
```

### 2.2 Alternative: Broaden FindActiveSessions (NOT Recommended)

Modifying `FindActiveSessions()` to also return PARKED sessions would fix writeguard but break `ResolveSession()` for all other callers. The smart scan's purpose is to find the one ACTIVE session to operate on, and returning PARKED sessions would create ambiguity in the priority chain.

### 2.3 Alternative: Add FindSessionByID (NOT Recommended for this scope)

Adding a `FindSessionByID(sessionsDir, sessionID)` function that validates a session exists regardless of status would be cleaner but heavier. The writeguard hook has a 3-second timeout (from hooks.yaml) and must stay under 100ms. Adding a separate file read for session validation is unnecessary when the Moirai lock check already validates the session directory exists (by reading `.moirai-lock` from it).

## 3. Security Analysis

### 3.1 Attack Surface

The file-path extraction is safe because:

1. **Writeguard already validates the file is protected** (`isProtectedFile()` check happens first)
2. **Moirai lock validation is the real gate**: `isMoiraiLockHeld()` reads the actual lock file, checks agent name, and checks staleness
3. **No path traversal risk**: `paths.IsSessionDir()` only matches `session-YYYYMMDD-HHMMSS-{hex}` patterns (32+ chars starting with `session-`)
4. **Fail-closed on lock**: If the extracted session ID points to a directory without a valid Moirai lock, `isMoiraiLockHeld()` returns false and the write is still blocked

### 3.2 Trust Boundary

The file path comes from CC's tool_input JSON (stdin payload). CC controls what file paths Claude generates. A malicious path like `.sos/sessions/session-20260209-120000-abcdef01/../../../etc/passwd` would:
- Pass `isProtectedFile()` only if it ends with a protected suffix (it doesn't in this case)
- Fail `isMoiraiLockHeld()` because no `.moirai-lock` exists at the traversed path
- Be blocked by the existing check chain

## 4. Impact Assessment

### 4.1 Files Changed

| File | Change | Lines |
|------|--------|-------|
| `internal/cmd/hook/writeguard.go` | Add `extractSessionIDFromPath()` + fallback logic | ~15 |
| `internal/cmd/hook/hook_test.go` | Add test for PARKED session with Moirai lock | ~30 |

### 4.2 Performance Impact

- **No additional I/O**: `extractSessionIDFromPath()` is pure string parsing
- **Only called on fallback path**: When `resolveSession()` already returned empty
- **Sub-microsecond**: String split + pattern match on path segments

### 4.3 Ripple Effects

- **None on `ResolveSession()`**: The session resolution chain is not modified
- **None on `FindActiveSessions()`**: Discovery logic stays ACTIVE-only
- **None on other hooks**: Only writeguard uses this fallback
- **Positive for `ari session lock`**: The lock command has the same bug (`GetSessionID()` only finds ACTIVE sessions). A separate fix is needed there, but this spike's scope is writeguard only.

## 5. Related Issue: `ari session lock` for PARKED Sessions

The `ari session lock` command (`internal/cmd/session/lock.go:87`) calls `ctx.GetSessionID()` which calls `FindActiveSession()` (singular). This means Moirai cannot even acquire a lock for a PARKED session via the CLI. The workaround is to pass `--session-id` explicitly, which Moirai does when it knows the session ID. But this is still a gap worth tracking.

**Recommendation**: File a separate ticket to add `--session-id` as a required parameter for lock/unlock operations, or add session-ID-from-path fallback to the lock command as well.

## 6. Follow-Up Actions

| Action | Priority | Effort |
|--------|----------|--------|
| Implement writeguard file-path fallback | HIGH | 1-2 hours |
| Add integration test: PARKED session + Moirai lock + writeguard allow | HIGH | 1 hour |
| Fix `ari session lock/unlock` for PARKED sessions | MEDIUM | 1 hour |
| Add `extractSessionIDFromPath` to `paths` package (shared utility) | LOW | 30 min |
| Audit other hooks for PARKED-session blindness | LOW | 1 hour |

## 7. Test Plan

### Unit Tests
- `TestExtractSessionIDFromPath_ValidPath`: `.sos/sessions/session-20260209-120000-abcdef01/SESSION_CONTEXT.md` -> `session-20260209-120000-abcdef01`
- `TestExtractSessionIDFromPath_NoSessionID`: `src/main.go` -> `""`
- `TestExtractSessionIDFromPath_NestedPath`: Deep paths with session ID in the middle
- `TestExtractSessionIDFromPath_InvalidSessionPattern`: Paths with `session-` prefix but wrong format

### Integration Tests
- `TestWriteguard_ParkedSession_MoiraiLockAllow`: PARKED session + valid Moirai lock -> allow
- `TestWriteguard_ParkedSession_NoLock`: PARKED session + no Moirai lock -> deny
- `TestWriteguard_ParkedSession_StaleLock`: PARKED session + stale Moirai lock -> deny

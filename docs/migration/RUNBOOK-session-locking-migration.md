# Migration Runbook: Session Locking and v1-v2 Migration

**Version**: 1.0.0
**Created**: 2026-01-04
**Risk Level**: LOW (no breaking changes, fully backward compatible)
**Estimated Deployment Time**: 30-45 minutes

---

## 1. Executive Summary

This runbook guides deployment of session management infrastructure improvements including:

1. **Lock Consolidation** (Phases 1-3): Fixes 14 bugs in session-manager.sh and session-fsm.sh
   - Trap quoting fix prevents orphaned locks
   - Post-creation verification eliminates phantom sessions
   - Relaxed validation accepts extended session ID formats
   - Automatic FD allocation prevents collision on high-concurrency

2. **v1 to v2.1 Migration** (Phase 4): Automatic schema upgrade for legacy sessions
   - Auto-migration on first `session-manager.sh status` call
   - Backup creation for rollback capability
   - Field consolidation (auto_parked_at merged to parked_at)

3. **Hook Integration** (Phase 3): Coordinated state mutations
   - auto-park.sh now uses FSM transition (was direct write)
   - session-write-guard.sh recognizes STATE_MATE_BYPASS
   - artifact-tracker.sh uses atomic operations

**Why This Matters**: These fixes eliminate race conditions that could cause session corruption during concurrent operations, and ensure all sessions use a consistent v2 schema with the FSM as single source of truth.

**Breaking Changes**: None. All changes are backward compatible.

**Rollback Availability**: Full rollback available via git revert or per-session rollback via .v1.backup files.

---

## 2. Pre-Deployment Checklist

Complete ALL items before proceeding to deployment steps.

### 2.1 Documentation Review

- [ ] Read this entire runbook before starting
- [ ] Review TDD documents:
  - `/Users/tomtenuta/Code/roster/docs/design/TDD-session-manager-locking.md`
  - `/Users/tomtenuta/Code/roster/docs/design/TDD-session-manager-ecosystem-audit.md`
- [ ] Understand the 14 bug fixes (see Appendix A)

### 2.2 Environment Verification

```bash
# Verify bash version (need 4.1+ for automatic FD allocation)
bash --version
```
**Verify**: Version 4.1 or higher displayed

```bash
# Verify flock availability (preferred locking method)
command -v flock && echo "flock available" || echo "flock NOT available (will use mkdir fallback)"
```
**Verify**: Either "flock available" or note that mkdir fallback will be used

```bash
# Verify current working directory
pwd
```
**Verify**: You are in the roster project root

### 2.3 Session State Backup

```bash
# Create timestamped backup of all sessions
BACKUP_DIR=".claude/sessions.backup.$(date +%Y%m%d-%H%M%S)"
cp -r .claude/sessions "$BACKUP_DIR" 2>/dev/null && echo "Backup created: $BACKUP_DIR" || echo "No sessions to backup"
```
**Verify**: Backup directory created OR "No sessions to backup" message

```bash
# List current session count
ls -d .claude/sessions/session-* 2>/dev/null | wc -l
```
**Note**: Record this number for post-deployment verification

### 2.4 Lock File Status

```bash
# Check for existing lock files (should be empty/nonexistent)
ls -la .claude/sessions/.locks/ 2>/dev/null || echo "No lock directory (expected)"
ls -la .claude/sessions/.create.lock 2>/dev/null || echo "No create lock (expected)"
```
**Verify**: No orphaned lock files present. If locks exist, see Troubleshooting section 6.1

### 2.5 v1 Session Inventory

```bash
# List sessions needing v1->v2 migration
.claude/hooks/lib/session-migrate.sh list-v1 2>/dev/null || echo "Migration script not yet deployed"
```
**Note**: Record count of v1 sessions (will auto-migrate after deployment)

---

## 3. What's Changing

### 3.1 Session Manager Changes

| File | Line(s) | Change | Bug Fixed |
|------|---------|--------|-----------|
| session-manager.sh | 308 | Trap quoting fix | LOCK-002 |
| session-manager.sh | 295-366 | Post-creation verification | RACE-001 |
| session-manager.sh | 345-348 | Directory/context file check | LOCK-001 |
| session-fsm.sh | 84, 121 | Automatic FD allocation | LOCK-003 |
| session-fsm.sh | 600 | PID suffix on backup files | RACE-003 |
| session-core.sh | 69-123 | Handle .current-session directory case | STATE-001 |
| session-core.sh | 78, 148 | Relaxed session ID validation | VALID-001 |
| session-state.sh | 167 | Relaxed validation format | VALID-001 |

### 3.2 Hook Integration Changes

| File | Change | Purpose |
|------|--------|---------|
| auto-park.sh | Uses atomic_write | Prevents file corruption |
| session-write-guard.sh | Checks STATE_MATE_BYPASS | Allows state-mate writes |
| artifact-tracker.sh | Uses sed with .bak cleanup | Atomic artifact updates |

### 3.3 Migration System

| Component | Description |
|-----------|-------------|
| `migrate_session()` | Migrates single v1 session to v2.1 |
| `rollback_session()` | Restores v1 from .v1.backup |
| `auto_migrate_if_needed()` | Silent migration on first access |
| CLI interface | `list-v1`, `migrate`, `rollback`, `--batch` |

### 3.4 Schema Changes (v1 to v2.1)

**Fields Added**:
- `schema_version: "2.1"` - Version identifier
- `status: "ACTIVE|PARKED|ARCHIVED"` - Canonical state field
- `team: "team-name"|null` - Explicit team field (null for cross-cutting)

**Fields Consolidated**:
- `auto_parked_at` merged into `parked_at` with `parked_auto: true` marker
- `auto_parked_reason` removed (merged into `parked_reason`)

---

## 4. Deployment Steps

### Step 1: Pre-Deployment Validation (5 min)

```bash
# Verify no corrupted lock files
find .claude/sessions -name "*.lock" -type f 2>/dev/null
find .claude/sessions -name "*.lock.d" -type d 2>/dev/null
```
**Verify**: No output (no stale locks)

```bash
# Verify sessions directory structure
ls -la .claude/sessions/ | head -20
```
**Verify**: Directory structure looks normal

```bash
# Record current session (if any)
cat .claude/sessions/.current-session 2>/dev/null || echo "No current session"
```
**Note**: Record this for verification after deployment

### Step 2: Deploy Code (2 min)

**Option A: Git merge (if changes on branch)**
```bash
git merge session-3-locking-fixes
```

**Option B: Git pull (if already merged to main)**
```bash
git pull origin main
```

**Option C: Direct deployment (CD pipeline)**
```bash
# Verify files updated via your deployment mechanism
```

**Verify files deployed**:
```bash
# Core session management
ls -la .claude/hooks/lib/session-manager.sh
ls -la .claude/hooks/lib/session-fsm.sh
ls -la .claude/hooks/lib/session-core.sh
ls -la .claude/hooks/lib/session-state.sh
ls -la .claude/hooks/lib/session-migrate.sh

# Hooks
ls -la .claude/hooks/session-guards/auto-park.sh
ls -la .claude/hooks/session-guards/session-write-guard.sh
ls -la .claude/hooks/tracking/artifact-tracker.sh
```
**Verify**: All files exist with recent modification times

### Step 3: Verify Hook Execution (5 min)

```bash
# Test session-context injection hook (SessionStart)
CLAUDE_PROJECT_DIR=. .claude/hooks/context-injection/session-context.sh 2>/dev/null
echo "Exit code: $?"
```
**Verify**: Exit code 0

```bash
# Test session-write-guard blocking (PreToolUse)
export CLAUDE_HOOK_TOOL_NAME="Write"
export CLAUDE_HOOK_FILE_PATH=".claude/sessions/test/SESSION_CONTEXT.md"
.claude/hooks/session-guards/session-write-guard.sh 2>&1 | head -20
echo "Exit code: ${PIPESTATUS[0]}"
```
**Verify**: Exit code 1, output contains "block"

```bash
# Test STATE_MATE_BYPASS allows writes
export STATE_MATE_BYPASS="true"
.claude/hooks/session-guards/session-write-guard.sh
echo "Exit code: $?"
unset STATE_MATE_BYPASS CLAUDE_HOOK_TOOL_NAME CLAUDE_HOOK_FILE_PATH
```
**Verify**: Exit code 0 (allowed)

### Step 4: Auto-Migrate Sessions (varies by session count)

Sessions migrate automatically on first access. To trigger immediate migration:

```bash
# Option A: Trigger via status (migrates current session)
.claude/hooks/lib/session-manager.sh status | head -5
```
**Verify**: JSON output with no errors

```bash
# Option B: Batch migrate all v1 sessions
.claude/hooks/lib/session-migrate.sh migrate --batch 2>&1
```
**Verify**: Output shows migration count (or "0 v1 sessions" if none)

```bash
# Verify migration success
.claude/hooks/lib/session-migrate.sh status
```
**Verify**: `v1_sessions: 0` or all shown as migrated

```bash
# Spot-check a migrated session (if any exist)
FIRST_SESSION=$(ls -d .claude/sessions/session-* 2>/dev/null | head -1)
if [[ -n "$FIRST_SESSION" ]]; then
    grep "schema_version:" "$FIRST_SESSION/SESSION_CONTEXT.md"
fi
```
**Verify**: Shows `schema_version: "2.1"` or `schema_version: "2.0"`

### Step 5: Post-Deployment Verification (10 min)

**5.1 Basic Operations**

```bash
# Create test session
TEST_RESULT=$(.claude/hooks/lib/session-manager.sh create "post-deploy-test-$(date +%s)" MODULE 2>&1)
echo "$TEST_RESULT" | head -10
TEST_SESSION_ID=$(echo "$TEST_RESULT" | grep -o '"session_id": "[^"]*"' | cut -d'"' -f4)
echo "Created session: $TEST_SESSION_ID"
```
**Verify**: JSON with `"success": true`

```bash
# List sessions
.claude/hooks/lib/session-manager.sh list 2>/dev/null || .claude/hooks/lib/session-manager.sh status
```
**Verify**: Test session appears in output

```bash
# Query status
.claude/hooks/lib/session-manager.sh status | head -20
```
**Verify**: JSON output shows session details

**5.2 Lock Verification**

```bash
# Verify no orphan locks after create
ls .claude/sessions/.locks/ 2>/dev/null || echo "No lock directory (expected after operation completes)"
ls .claude/sessions/.create.lock 2>/dev/null || echo "No create lock (expected after operation completes)"
```
**Verify**: No orphaned locks

**5.3 Audit Log Verification**

```bash
# Check audit logs populate
ls -la .claude/sessions/.audit/ 2>/dev/null || echo "Audit directory will be created on first mutation"
tail -5 .claude/sessions/.audit/session-mutations.log 2>/dev/null || echo "No mutations logged yet"
tail -5 .claude/sessions/.audit/migrations.log 2>/dev/null || echo "No migrations logged yet"
```
**Verify**: Logs exist and show recent entries (if sessions were migrated/mutated)

**5.4 Clean Up Test Session**

```bash
# Wrap test session to clean up
if [[ -n "$TEST_SESSION_ID" ]]; then
    .claude/hooks/lib/session-manager.sh mutate wrap true
    echo "Test session wrapped"
fi
```
**Verify**: No errors

---

## 5. Rollback Procedures

### 5.1 Quick Rollback (issues detected immediately)

**Option A: Revert commits**
```bash
# Identify commit to revert
git log --oneline -5

# Revert (replace SHA with actual commit)
git revert <commit-sha>
```

**Option B: Restore from backup**
```bash
# Find your backup
ls -la .claude/*.backup.* 2>/dev/null

# Restore sessions
rm -rf .claude/sessions
cp -r .claude/sessions.backup.<timestamp> .claude/sessions
echo "Sessions restored from backup"
```

### 5.2 Per-File Rollback

```bash
# Rollback specific files without reverting entire commit
git checkout HEAD~1 -- .claude/hooks/lib/session-manager.sh
git checkout HEAD~1 -- .claude/hooks/lib/session-fsm.sh
git checkout HEAD~1 -- .claude/hooks/lib/session-core.sh
git checkout HEAD~1 -- .claude/hooks/lib/session-state.sh
```

### 5.3 Session-Level Migration Rollback

For a specific session that had issues after v1->v2 migration:

```bash
# Check if backup exists
ls -la .claude/sessions/<session_id>/SESSION_CONTEXT.md.v1.backup

# Rollback to v1
.claude/hooks/lib/session-migrate.sh rollback <session_id>
```
**Verify**: `grep "schema_version:" .claude/sessions/<session_id>/SESSION_CONTEXT.md` returns nothing (v1 format)

### 5.4 Batch Migration Rollback

```bash
# Rollback all migrated sessions
.claude/hooks/lib/session-migrate.sh rollback --batch
```

---

## 6. Troubleshooting

### 6.1 Symptom: "Orphan lock file"

**Detection**:
```bash
find .claude/sessions -name "*.lock" -mmin +10
find .claude/sessions -name "*.lock.d" -mmin +10
```

**Resolution**:
```bash
# Check if owning process is dead
LOCK_FILE=".claude/sessions/.create.lock"
if [[ -d "$LOCK_FILE" && -f "$LOCK_FILE/pid" ]]; then
    PID=$(cat "$LOCK_FILE/pid")
    if ! ps -p "$PID" >/dev/null 2>&1; then
        echo "Lock owner (PID $PID) is dead, removing orphan lock"
        rm -rf "$LOCK_FILE"
    else
        echo "Lock owner (PID $PID) is still running"
    fi
fi
```

### 6.2 Symptom: ".current-session is a directory"

**Detection**:
```bash
file .claude/sessions/.current-session
```
Shows "directory" instead of "ASCII text"

**Resolution**:
```bash
# Remove the erroneously created directory
rm -rf .claude/sessions/.current-session

# Let the system recreate it as a file
.claude/hooks/lib/session-manager.sh status
```

### 6.3 Symptom: "Session create hangs"

**Detection**: `session-manager.sh create` does not return within 15 seconds

**Resolution**:
```bash
# Check for stuck lock
ls -la .claude/sessions/.locks/
ls -la .claude/sessions/.create.lock

# Find and kill stalled process
ps aux | grep session-manager
kill -9 <pid>

# Clean locks and retry
rm -rf .claude/sessions/.locks/
rm -rf .claude/sessions/.create.lock

# Retry operation
.claude/hooks/lib/session-manager.sh create "retry-test" MODULE
```

### 6.4 Symptom: "Migration failed: Validation error"

**Detection**: `session-migrate.sh migrate` returns validation error

**Resolution**:
```bash
SESSION_ID="<session_id>"

# Check backup exists
ls .claude/sessions/$SESSION_ID/SESSION_CONTEXT.md.v1.backup

# Restore from backup
cp .claude/sessions/$SESSION_ID/SESSION_CONTEXT.md.v1.backup \
   .claude/sessions/$SESSION_ID/SESSION_CONTEXT.md

# Investigate the original content
cat .claude/sessions/$SESSION_ID/SESSION_CONTEXT.md

# Check which required fields are missing
grep -E "^(session_id|created_at|initiative|complexity|active_team|current_phase):" \
    .claude/sessions/$SESSION_ID/SESSION_CONTEXT.md
```

### 6.5 Symptom: "Session ID format rejected"

**Detection**: Error like "Invalid session_id format"

**Resolution**:
After deployment, the validation is relaxed. If you see this error, verify deployment succeeded:
```bash
# Check validation function
grep -A3 "validate_session_id_format" .claude/hooks/lib/session-state.sh
```
**Expected**: Pattern should be `^session-.+` (not the stricter old pattern)

### 6.6 Symptom: "Timeout waiting for session lock"

**Detection**: Create/wrap operations fail with timeout message

**Resolution**:
```bash
# Check lock timeout setting
echo "Default timeout: 10 seconds"

# Check for stuck operations
ps aux | grep session

# If needed, increase timeout for your operation
FSM_LOCK_TIMEOUT=30 .claude/hooks/lib/session-manager.sh create "test" MODULE
```

---

## 7. Performance Impact

### 7.1 Expected Changes

| Operation | Baseline | After Deployment | Notes |
|-----------|----------|-----------------|-------|
| Session create | ~100ms | <150ms | Slight overhead from verification |
| Lock acquisition | ~10ms | <20ms | Improved FD allocation |
| Status query | ~50ms | <75ms | Includes auto-migration check |
| Migration (per session) | N/A | <200ms | One-time operation |

### 7.2 If Performance Below Expectations

Check:
1. **Filesystem latency**: `.claude/sessions/` on slow storage (network drive, encrypted volume)
2. **System load**: High CPU/IO may slow lock acquisition
3. **Bash version**: Ensure bash 4.1+ for `{fd}` syntax

Benchmark command:
```bash
time .claude/hooks/lib/session-manager.sh status
```

---

## 8. Monitoring Checklist

Post-deployment, monitor for 24-48 hours:

- [ ] No orphan lock files accumulating in `.locks/`
- [ ] Audit logs populate regularly (`.audit/session-mutations.log`)
- [ ] Migration logs show successful v1->v2 conversions (`.audit/migrations.log`)
- [ ] Session status command completes <100ms consistently
- [ ] No hook execution errors in Claude Code logs
- [ ] New sessions create with `schema_version: "2.1"`
- [ ] STATE_MATE_BYPASS works for state-mate agent

---

## 9. FAQ

**Q: Will this break my existing sessions?**
A: No. All sessions are backward compatible. v1 sessions auto-migrate on first access with backup created (.v1.backup).

**Q: Can I rollback after deployment?**
A: Yes. Rollback available via:
- Git revert (full rollback)
- Per-session `.v1.backup` files (90 day retention)
- Manual file restore from `.sessions.backup.*`

**Q: Do I need to manually migrate sessions?**
A: No. Auto-migration occurs on first `session-manager.sh status` call. Manual migration available via CLI if needed.

**Q: What's the performance impact?**
A: Minimal. Session create <150ms (slightly slower due to verification), lock acquisition improved, status query <75ms.

**Q: Are there any breaking changes?**
A: No. All changes are backward compatible. Relaxed validation accepts more session ID formats.

**Q: What if I have concurrent Claude sessions?**
A: The locking improvements specifically address this. Only one session can be created at a time (global lock), and FSM provides session-scoped locks for mutations.

**Q: How long are .v1.backup files retained?**
A: Indefinitely until manually deleted. We recommend keeping them for at least 90 days.

**Q: What happens if flock is not available?**
A: Falls back to mkdir-based locking (portable but treats shared locks as exclusive). This is conservative-correct.

---

## 10. Support & Escalation

If issues arise:

1. **Check troubleshooting section** (Section 6)
2. **Review audit logs**: `.claude/sessions/.audit/session-mutations.log`
3. **Verify file permissions** on `.claude/sessions/` and `.claude/hooks/`
4. **Check bash version**: `bash --version` (need 4.1+)
5. **Escalate with**:
   - Error message text
   - Relevant audit log excerpt
   - Session ID (if applicable)
   - Output of `session-manager.sh status`

---

## Appendix A: Files Modified

| File | Changes | Impact |
|------|---------|--------|
| `.claude/hooks/lib/session-manager.sh` | Lock scope, verification, trap fix | CRITICAL - session creation |
| `.claude/hooks/lib/session-fsm.sh` | FD allocation, backup naming | HIGH - locking |
| `.claude/hooks/lib/session-core.sh` | Directory check, atomic write | CRITICAL - current session |
| `.claude/hooks/lib/session-state.sh` | Validation relaxation | LOW - backward compat |
| `.claude/hooks/lib/session-migrate.sh` | v1->v2 migration | HIGH - schema migration |
| `.claude/hooks/session-guards/auto-park.sh` | Atomic writes | MEDIUM - hook behavior |
| `.claude/hooks/session-guards/session-write-guard.sh` | STATE_MATE_BYPASS | MEDIUM - state mutations |
| `.claude/hooks/tracking/artifact-tracker.sh` | Atomic writes | LOW - tracking |

---

## Appendix B: Bug Fixes Reference

| Bug ID | Severity | Description | Fix Location |
|--------|----------|-------------|--------------|
| LOCK-001 | CRITICAL | Create lock scope mismatch | session-manager.sh:295-366 |
| LOCK-002 | HIGH | Trap single quote escaping | session-manager.sh:308 |
| LOCK-003 | MEDIUM | FD number collision risk | session-fsm.sh:84,121 |
| LOCK-004 | LOW | Shared lock fallback (documented) | session-fsm.sh |
| RACE-001 | CRITICAL | Phantom session creation | session-manager.sh:345-348 |
| RACE-002 | HIGH | Set current session TOCTOU | session-core.sh:69-123 |
| RACE-003 | MEDIUM | Backup file collision | session-fsm.sh:600 |
| RACE-004 | HIGH | Lock release timing | Addressed by RACE-003 |
| VALID-001 | CRITICAL | Session ID format rejection | session-core.sh:78,148 |
| VALID-002 | MEDIUM | Context field mismatch | session-state.sh |
| STATE-001 | CRITICAL | .current-session directory race | session-core.sh:94-100 |
| STATE-002 | HIGH | Inconsistent state sources | Consolidated to FSM |
| PARSE-001 | LOW | Main dispatch arg leak | session-manager.sh |
| PARSE-002 | LOW | Mutate subcommand shift | session-manager.sh |

---

## Appendix C: Test Coverage Summary

Tests validating this deployment (run during Phase 5):

| Category | Tests | Coverage |
|----------|-------|----------|
| Concurrency | CONC-001 to CONC-005 | Parallel creates, lock timeout, stale cleanup |
| Hook Ordering | HOOK-001 to HOOK-005 | SessionStart, PreToolUse, PostToolUse, Stop |
| Migration | MIGRATE-001 to MIGRATE-008 | v1->v2, rollback, dry-run, validation |
| state-mate | MATE-001 to MATE-004 | Bypass, blocking, FSM coordination |
| Regression | REG-001 to REG-005 | Bug-specific validation |

---

## Appendix D: Compatibility Matrix

| Component Version | roster v1.x sessions | roster v2.x sessions | Notes |
|-------------------|---------------------|---------------------|-------|
| Pre-deployment | Compatible | Not present | No v2 sessions exist |
| Post-deployment | Auto-migrates to v2 | Compatible | Auto-migration on access |
| Rollback | Restored via .v1.backup | N/A | Manual rollback available |

| Bash Version | Lock Method | Status |
|--------------|-------------|--------|
| 4.1+ | flock with {fd} | Recommended |
| 3.x | mkdir fallback | Supported (conservative) |
| 5.x+ | flock with {fd} | Recommended |

---

**Runbook Complete**

This runbook was tested by following each step in an isolated test satellite before finalization.

Document: `/Users/tomtenuta/Code/roster/docs/migration/RUNBOOK-session-locking-migration.md`

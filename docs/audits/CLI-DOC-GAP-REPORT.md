# CLI-Documentation Gap Report

**Date**: 2026-01-07
**Initiative**: knossos-v0.1.0-qa
**Scope**: Comprehensive audit of ari CLI vs documentation/hooks/skills

---

## Executive Summary

| Category | Findings | Critical | High | Medium | Low |
|----------|----------|----------|------|--------|-----|
| sync commands | 7 gaps | 0 | 5 | 1 | 1 |
| session commands | 8 gaps | 1 | 2 | 4 | 1 |
| rite commands | 2 gaps | 0 | 0 | 1 | 1 |
| hook wrappers | 1 gap | 0 | 0 | 1 | 0 |
| CLI documentation | 5 gaps | 0 | 3 | 2 | 0 |
| **TOTAL** | **23 gaps** | **1** | **10** | **9** | **3** |

---

## 1. Sync Command Gaps

### HIGH PRIORITY - Missing Flags on `ari sync materialize`

| Flag | Documented In | Implemented | Severity |
|------|---------------|-------------|----------|
| `--update` / `-u` | 6 rite-switching commands | NO | HIGH |
| `--dry-run` | 6 rite-switching commands | NO | HIGH |
| `--keep-all` | 6 rite-switching commands | NO | HIGH |
| `--remove-all` | 6 rite-switching commands | NO | HIGH |
| `--promote-all` | 6 rite-switching commands | NO | HIGH |

**Root Cause**: User-commands copied from template that documented planned features, but Go implementation only has `--force` and `--rite`.

**Affected Files**:
- `user-commands/rite-switching/10x.md`
- `user-commands/rite-switching/hygiene.md`
- `user-commands/rite-switching/ecosystem.md`
- `user-commands/rite-switching/rnd.md`
- `user-commands/rite-switching/debt.md`
- `user-commands/rite-switching/sre.md`

### MEDIUM - `/sync --refresh` Translation

| Feature | Documented | Implemented | Severity |
|---------|------------|-------------|----------|
| `--refresh` flag | user-commands/cem/sync.md | NO | MEDIUM |

**Note**: Documentation says `--refresh` should translate to `ari sync materialize`, but no such translation exists.

### LOW - `--force` on `ari sync`

| Issue | Details | Severity |
|-------|---------|----------|
| `/sync --force` fails | Root sync command has no `--force`, only subcommands do | LOW |

---

## 2. Session Command Gaps

### CRITICAL - Complexity Level Mismatch

| Documented | Implemented | Impact |
|------------|-------------|--------|
| SCRIPT, MODULE, SERVICE, PLATFORM | PATCH, MODULE, SYSTEM, INITIATIVE, MIGRATION | Breaks user workflows |

**Affected Files**:
- `user-commands/session/start.md` - Documents wrong complexity enum
- `internal/cmd/session/create.go` - Has correct enum

### HIGH - Hidden Commands Not Documented

| Command | Purpose | Documented |
|---------|---------|------------|
| `ari session migrate` | v1→v2.1 schema migration | NO |
| `ari session lock` | Debug: manual lock acquisition | NO |
| `ari session unlock` | Debug: manual lock release | NO |

### HIGH - Resume Command Signature Mismatch

| Feature | Documented | Implemented |
|---------|------------|-------------|
| `--session=ID` flag | user-commands/session/continue.md | NO |
| `--agent=NAME` flag | user-commands/session/continue.md | NO |
| Interactive session listing | user-commands/session/continue.md | NO |

**Actual Behavior**: `ari session resume` uses TTY-mapped session or fails. No flags.

### MEDIUM - Handoff Scope Confusion

| Documented As | Actually Is |
|---------------|-------------|
| `ari session handoff` (session subcommand) | `ari handoff` (separate command group) |

**Affected Files**: `.claude/CLAUDE.md` lines 109-112

### MEDIUM - Sails Check Standalone

| Feature | Status |
|---------|--------|
| `ari sails check` independent of wrap | Partially implemented |
| Sails query without wrap | Not available |

### MEDIUM - Phase Transition Documentation

| Feature | Documented | Implemented |
|---------|------------|-------------|
| `ari session transition` | docs/guides/ariadne-cli.md | YES |
| Valid phase values | Not documented | implementation-dependent |

### LOW - Park Reason Signature

| Documented | Implemented |
|------------|-------------|
| Positional arg `$1` | Flag `--reason` / `-r` |

---

## 3. Rite Command Gaps

### MEDIUM - `--update` Flag Missing

| Flag | Documented | Implemented |
|------|------------|-------------|
| `--update` / `-u` | user-commands/navigation/rite.md | NO |

### LOW - `--team` Deprecation Not in User Docs

| Issue | Details |
|-------|---------|
| `--team` flag | Deprecated in CLI, but user-commands don't warn |

---

## 4. Hook Wrapper Gaps

### MEDIUM - cognitive-budget.sh Has No CLI Counterpart

| Hook | Shell Implementation | ari CLI |
|------|---------------------|---------|
| cognitive-budget.sh | Full bash implementation | NO equivalent |

**Note**: Intentional - kept in bash to minimize binary overhead for non-blocking monitoring.

**All other hooks have perfect parity** (context, clew, validate, writeguard, route, autopark).

---

## 5. CLI Documentation Gaps

### HIGH - Commands Missing from CLAUDE.md

| Command | TDD | CLAUDE.md | Guide |
|---------|-----|-----------|-------|
| naxos | NO | NO | NO |
| validate | NO | NO | Partial |
| sails | NO | Partial | YES |

### HIGH - CLAUDE.md Only Documents 4 of 16 Commands

**Current**: session, hook, sails, handoff
**Missing**: artifact, inscription, manifest, naxos, rite, sync, tribute, validate, worktree

### MEDIUM - Template Not Updated

| File | Issue |
|------|-------|
| `knossos/templates/sections/ariadne-cli.md.tpl` | Only generates 4 commands |

### MEDIUM - TDD Missing for 3 Commands

| Command | TDD Status |
|---------|------------|
| naxos | NO TDD |
| sails | NO TDD (only guide) |
| validate | NO TDD |

---

## 6. Compliance Matrix

| Command | Implementation | TDD | CLAUDE.md | Guide | Hooks | Status |
|---------|----------------|-----|-----------|-------|-------|--------|
| artifact | YES | YES | NO | YES | - | GOOD |
| handoff | YES | YES | YES | YES | - | EXCELLENT |
| hook | YES | YES | YES | YES | YES | EXCELLENT |
| inscription | YES | YES | NO | Partial | - | GOOD |
| manifest | YES | YES | NO | YES | - | GOOD |
| **naxos** | YES | **NO** | **NO** | **NO** | - | **GAP** |
| rite | YES | YES | NO | YES | - | GOOD |
| **sails** | YES | **NO** | Partial | YES | - | **GAP** |
| session | YES | YES | YES | YES | YES | EXCELLENT |
| sync | YES | YES | NO | YES | - | **PARTIAL** |
| tribute | YES | YES | NO | NO | - | PARTIAL |
| **validate** | YES | **NO** | **NO** | Partial | - | **GAP** |
| worktree | YES | ADR only | NO | YES | - | PARTIAL |

---

## 7. Root Causes

1. **Template drift**: User-commands documented planned features before implementation
2. **Complexity enum divergence**: SCRIPT/SERVICE/PLATFORM vs PATCH/SYSTEM/INITIATIVE/MIGRATION
3. **Hidden utilities**: migrate/lock/unlock never added to user docs
4. **Resume redesign**: Interactive features removed but docs not updated
5. **CLAUDE.md stale**: Only 4/16 commands documented
6. **TDD gaps**: naxos/sails/validate bypassed design phase
7. **ADR vs TDD**: Worktree used ADR format instead of TDD

---

## 8. Prioritized Remediation

### P0 - CRITICAL (Fix Immediately)

1. **Fix complexity enum** in `user-commands/session/start.md`
   - Change: SCRIPT → PATCH, SERVICE → SYSTEM, PLATFORM → INITIATIVE

### P1 - HIGH (Fix Before Next Release)

2. **Remove unimplemented flags** from rite-switching commands:
   - `--update`, `--dry-run`, `--keep-all`, `--remove-all`, `--promote-all`
   - OR implement the flags in `internal/cmd/sync/materialize.go`

3. **Update resume documentation** to match actual behavior:
   - Remove `--session` and `--agent` flag claims
   - Document TTY-mapped session requirement

4. **Expand CLAUDE.md** to document all 16 commands

5. **Create TDD for naxos/validate** commands

### P2 - MEDIUM (Track for Future)

6. Implement `--refresh` translation for `/sync` command
7. Create TDD for sails command (separate from session)
8. Update worktree from ADR to TDD format
9. Clarify handoff is separate command group (not session subcommand)
10. Document hidden session commands (migrate, lock, unlock)

### P3 - LOW (Nice to Have)

11. Add deprecation warnings for `--team` flag in user-commands
12. Implement `ari hook cognitive-budget` (or document intentional omission)
13. Standardize park reason (flag vs positional)

---

## 9. Affected File Index

| File | Gap Type | Priority |
|------|----------|----------|
| `user-commands/session/start.md` | Complexity enum | P0 |
| `user-commands/session/continue.md` | Resume flags | P1 |
| `user-commands/rite-switching/*.md` (6 files) | Unimplemented flags | P1 |
| `user-commands/cem/sync.md` | --refresh translation | P2 |
| `user-commands/navigation/rite.md` | --update flag | P2 |
| `.claude/CLAUDE.md` | Missing commands | P1 |
| `docs/design/TDD-ariadne-naxos.md` | Missing | P1 |
| `docs/design/TDD-ariadne-validate.md` | Missing | P1 |
| `docs/design/TDD-ariadne-sails.md` | Missing | P2 |
| `docs/design/TDD-ariadne-worktree.md` | Missing (ADR exists) | P2 |

---

## 10. Verification Attestation

This report was generated by 5 parallel audit agents examining:
- `internal/cmd/sync/*.go` - All sync implementations
- `internal/cmd/session/*.go` - All session implementations
- `internal/cmd/rite/*.go` - All rite implementations
- `internal/cmd/hook/*.go` - All hook implementations
- `.claude/hooks/ari/*.sh` - All hook shell wrappers
- `user-commands/**/*.md` - All user command documentation
- `user-skills/**/*.md` - All skill documentation
- `.claude/CLAUDE.md` - Project instructions
- `docs/guides/*.md` - User guides
- `docs/design/TDD-*.md` - Technical design documents

**Report Date**: 2026-01-07
**Session**: session-20260107-200019-bbf01130
**Sprint**: sprint-20260107-192034

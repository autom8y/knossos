# SPIKE: Hook Ecosystem Full-Scope Audit

**Date**: 2026-02-11
**Status**: Complete (research phase)
**Next**: Stakeholder interview for cleanup remediation plan

---

## Question

What is the complete inventory of all hook-related artifacts across the Knossos ecosystem, and which are current, outdated, redundant, or orphaned?

## Scope

Audited every layer:
1. **Source of truth** (`hooks/hooks.yaml`) — canonical hook definitions
2. **Materialized output** (`.claude/settings.local.json`) — what CC actually reads
3. **Legacy shell wrappers** (`.claude/hooks/ari/*.sh`) — old bash dispatch layer
4. **Source shell wrappers** (`hooks/*.sh`) — pre-materialization bash scripts
5. **Go hook subcommands** (`internal/cmd/hook/*.go`) — ari binary implementations
6. **Hook infrastructure** (`internal/hook/`) — env parsing, output, clew events
7. **Materialization pipeline** (`internal/materialize/hooks.go`) — how hooks get wired
8. **User-level settings** (`~/.claude/settings.json`, `~/.claude/settings.local.json`)
9. **CC event type registry** (`internal/hook/env.go`) — all known CC hook events
10. **Clew event types** (`internal/hook/clewcontract/event.go`) — internal event taxonomy

---

## Layer-by-Layer Inventory

### Layer 1: Canonical Hook Definitions (`hooks/hooks.yaml`)

Schema v2.0 — direct `command` field (no bash wrappers).

| # | Event | Matcher | Command | Async | Timeout | Priority |
|---|-------|---------|---------|-------|---------|----------|
| 1 | SessionStart | — | `ari hook context --output json` | no | 10s | 5 |
| 2 | SessionEnd | — | `ari hook sessionend --output json` | no | 5s | 5 |
| 3 | Stop | — | `ari hook autopark --output json` | no | 5s | 5 |
| 4 | PreCompact | — | `ari hook precompact --output json` | no | 5s | 5 |
| 5 | PreToolUse | `Edit\|Write` | `ari hook writeguard --output json` | no | 3s | 3 |
| 6 | PreToolUse | `Bash` | `ari hook validate --output json` | no | 5s | 5 |
| 7 | PostToolUse | `Edit\|Write\|Bash` | `ari hook clew --output json` | **yes** | 5s | 5 |
| 8 | PostToolUse | — | `ari hook budget --output json` | no | 3s | 90 |
| 9 | SubagentStart | — | `ari hook subagent-start --output json` | **yes** | 5s | 5 |
| 10 | SubagentStop | — | `ari hook subagent-stop --output json` | **yes** | 5s | 5 |

**Total: 10 hook registrations, 8 unique Go subcommands**

### Layer 2: Materialized Output (`.claude/settings.local.json`)

| # | Event | Matcher | Command | Async | Timeout | Status |
|---|-------|---------|---------|-------|---------|--------|
| 1 | SessionStart | — | `ari hook context --output json` | no | 10 | MATCHES source |
| 2 | SessionEnd | — | `ari hook sessionend --output json` | no | 5 | MATCHES source |
| 3 | Stop | — | `ari hook autopark --output json` | no | 5 | MATCHES source |
| 4 | PreCompact | — | `ari hook precompact --output json` | no | 5 | MATCHES source |
| 5 | PreToolUse | `Edit\|Write` | `ari hook writeguard --output json` | no | 3 | MATCHES source |
| 6 | PreToolUse | `Bash` | `ari hook validate --output json` | no | 5 | MATCHES source |
| 7 | PostToolUse | `Edit\|Write\|Bash` | `ari hook clew --output json` | **yes** | 5 | MATCHES source |
| 8 | PostToolUse | — | `ari hook budget --output json` | no | 3 | MATCHES source |
| 9 | SubagentStart | — | `ari hook subagent-start --output json` | **yes** | 5 | MATCHES source |
| 10 | SubagentStop | — | `ari hook subagent-stop --output json` | **yes** | 5 | MATCHES source |

**Verdict: 10/10 match. settings.local.json is clean and in sync.**

### Layer 3: Legacy Shell Wrappers — `.claude/hooks/ari/`

These exist in the MATERIALIZED `.claude/` directory:

| File | Size | Date | Status |
|------|------|------|--------|
| `autopark.sh` | 1.0k | Jan 7 | **ORPHAN** — not referenced by settings.local.json |
| `clew.sh` | 1.4k | Jan 7 | **ORPHAN** — not referenced by settings.local.json |
| `cognitive-budget.sh` | 5.1k | Jan 7 | **ORPHAN** — not referenced by settings.local.json |
| `context.sh` | 1.0k | Jan 7 | **ORPHAN** — not referenced by settings.local.json |
| `hooks.yaml` | 2.8k | Jan 7 | **STALE** — v1.0 schema, superseded by `hooks/hooks.yaml` v2.0 |
| `route.sh` | 1.3k | Jan 7 | **ORPHAN** — not referenced by settings.local.json |
| `validate.sh` | 1.1k | Jan 7 | **ORPHAN** — not referenced by settings.local.json |
| `writeguard.sh` | 1.6k | Jan 22 | **ORPHAN** — not referenced by settings.local.json |

**Critical detail**: The `.claude/hooks/ari/hooks.yaml` is schema v1.0 (uses `path:` field, references `.sh` files). The canonical `hooks/hooks.yaml` is schema v2.0 (uses `command:` field, direct binary invocation). The v1.0 file is dead — nothing reads it.

**Critical detail**: The `.claude/hooks/ari/*.sh` scripts contain bash wrappers with:
- `USE_ARI_HOOKS` feature flag (dead code — v2 hooks bypass bash entirely)
- `CLAUDE_HOOK_TOOL_NAME` env var checks (dead code — CC uses stdin JSON)
- `CLAUDE_SESSION_DIR` env var checks (dead code — not a real CC env var)
- Binary resolution logic (`ARIADNE_BIN`, PATH, project-relative) — duplicated in every script

**None of these shell scripts are executed in production.** The settings.local.json calls `ari hook ...` directly.

### Layer 4: Source Shell Wrappers — `hooks/`

| File | Size | Date | Status |
|------|------|------|--------|
| `autopark.sh` | 955 | Feb 6 | **DEAD** — newer than .claude/ copies but still not used |
| `clew.sh` | 1.3k | Feb 6 | **DEAD** — newer than .claude/ copies but still not used |
| `cognitive-budget.sh` | 986 | Feb 6 | **DEAD** — v2 replaced with `ari hook budget` |
| `context.sh` | 961 | Feb 6 | **DEAD** — v2 replaced with `ari hook context` |
| `hooks.yaml` | 3.6k | Feb 9 | **CANONICAL** — this is the v2 source of truth |
| `route.sh` | 1.2k | Feb 6 | **DEAD** — no `route` Go subcommand exists, not in v2 hooks.yaml |
| `validate.sh` | 967 | Feb 6 | **DEAD** — v2 replaced with `ari hook validate` |
| `writeguard.sh` | 1.5k | Feb 8 | **DEAD** — v2 replaced with `ari hook writeguard` |

**Key difference from .claude/ copies**: `hooks/*.sh` have been cleaned up (removed `USE_ARI_HOOKS` feature flag) but are still architecturally dead. The v2 pipeline calls Go directly.

**Special case**: `cognitive-budget.sh` in `.claude/hooks/ari/` is 5.1k — a **FULL IMPLEMENTATION** in bash (temp files, thresholds, marker files). This was entirely replaced by `ari hook budget` (Go). The `hooks/cognitive-budget.sh` source is 986 bytes (thin wrapper). So `.claude/` has the old full implementation, `hooks/` has a thin wrapper, and neither is used.

### Layer 5: Go Hook Subcommands (`internal/cmd/hook/`)

| Subcommand | Go File | In hooks.yaml? | Has Tests? | Status |
|------------|---------|-----------------|------------|--------|
| `context` | context.go | Yes (SessionStart) | context_test.go | **ACTIVE** |
| `autopark` | autopark.go | Yes (Stop) | autopark_test.go | **ACTIVE** |
| `writeguard` | writeguard.go | Yes (PreToolUse) | writeguard_test.go | **ACTIVE** |
| `validate` | validate.go | Yes (PreToolUse) | validate_test.go | **ACTIVE** |
| `clew` | clew.go | Yes (PostToolUse) | clew_test.go | **ACTIVE** |
| `budget` | budget.go | Yes (PostToolUse) | budget_test.go | **ACTIVE** |
| `precompact` | precompact.go | Yes (PreCompact) | precompact_test.go | **ACTIVE** |
| `subagent-start` | subagent.go | Yes (SubagentStart) | subagent_test.go | **ACTIVE** |
| `subagent-stop` | subagent.go | Yes (SubagentStop) | subagent_test.go | **ACTIVE** |
| `sessionend` | sessionend.go | Yes (SessionEnd) | (in hook_test.go) | **ACTIVE** |

**All 10 Go subcommands are wired and active. No orphan Go commands.**

### Layer 6: CC Event Types Known to Ari (`internal/hook/env.go`)

| CC Event | Has Hook? | Notes |
|----------|-----------|-------|
| `PreToolUse` | Yes (2) | writeguard + validate |
| `PostToolUse` | Yes (2) | clew + budget |
| `PostToolUseFailure` | **No** | Not handled |
| `PermissionRequest` | **No** | Not handled |
| `Stop` | Yes (1) | autopark |
| `SessionStart` | Yes (1) | context |
| `SessionEnd` | Yes (1) | sessionend |
| `UserPromptSubmit` | **No** | Was handled by `route.sh` but no Go equivalent or v2 registration |
| `PreCompact` | Yes (1) | precompact |
| `SubagentStart` | Yes (1) | subagent-start |
| `SubagentStop` | Yes (1) | subagent-stop |
| `Notification` | **No** | CC event, no hook |
| `TeammateIdle` | **No** | CC event, no hook |
| `TaskCompleted` | **No** | CC event, no hook |

**5 CC events have no ari hook handler.** Of these:
- `UserPromptSubmit` — previously handled by `route.sh` (bash). No Go equivalent exists. The `route.sh` dispatched to `ari hook route`, but `route` is not a registered Go subcommand. `clew.go` handles route logic internally for PostToolUse events.
- `PostToolUseFailure`, `PermissionRequest`, `Notification`, `TeammateIdle`, `TaskCompleted` — intentionally unhandled (no use case identified).

### Layer 7: Legacy Artifacts in `.claude/hooks/ari/hooks.yaml` (v1 schema)

The v1.0 schema defined 7 hooks using `path:` field (bash wrappers):

| Event | Path | v2 Equivalent? |
|-------|------|-----------------|
| SessionStart | `ari/context.sh` | Yes — `ari hook context` |
| Stop | `ari/autopark.sh` | Yes — `ari hook autopark` |
| PreToolUse (Edit\|Write) | `ari/writeguard.sh` | Yes — `ari hook writeguard` |
| PreToolUse (Bash) | `ari/validate.sh` | Yes — `ari hook validate` |
| PostToolUse (Edit\|Write\|Bash) | `ari/clew.sh` | Yes — `ari hook clew` |
| PostToolUse (all) | `ari/cognitive-budget.sh` | Yes — `ari hook budget` |
| UserPromptSubmit (^/) | `ari/route.sh` | **NO** — dropped in v2 |

**The `route` hook was intentionally dropped in v2.** No Go implementation exists. The bash script dispatched to `ari hook route`, which doesn't exist as a subcommand.

### Layer 8: User-Level Settings

**`~/.claude/settings.json`**: No hooks. Contains statusLine, plugins, alwaysThinkingEnabled.

**`~/.claude/settings.local.json`**: No hooks section. Contains permissions (allow list) and MCP server configs. The allow list has grown to 217 entries, many from other projects (rulefy, mcp-coding-server).

**No project-scoped settings override** (`~/.claude/projects/-Users-tomtenuta-Code-knossos/settings*.json`): Does not exist.

### Layer 9: Materialization Pipeline (`internal/materialize/hooks.go`)

Key behaviors:
- Reads `hooks/hooks.yaml` (v2 only — rejects v1 schema)
- Generates `hooks` section for `settings.local.json`
- `mergeHooksSettings()` three-way classifies existing hooks:
  - **ari-managed** (prefix `ari hook`) → replaced
  - **legacy platform** (contains `$CLAUDE_PROJECT_DIR`, `.claude/hooks/`, or ends `.sh`) → stripped
  - **user-defined** → preserved
- Has embedded hooks fallback via `m.embeddedHooks` (compiled into binary)

**The pipeline correctly handles cleanup on sync, but does NOT touch files in `.claude/hooks/` directory.** It only manages the `hooks` key in `settings.local.json`.

---

## Findings Summary

### ORPHAN: `.claude/hooks/ari/` directory (8 files, ~13k bytes)

**Severity**: HIGH — confusion vector, stale code, not referenced by anything

| Finding | Files | Impact |
|---------|-------|--------|
| 7 shell scripts no longer executed | *.sh | Dead code in repo |
| v1 hooks.yaml superseded by v2 | hooks.yaml | Confusing — two hooks.yaml files |
| Full bash cognitive-budget.sh (5.1k) | cognitive-budget.sh | Replaced by 200-line Go equivalent |

**Nothing reads these files.** The materialization pipeline reads `hooks/hooks.yaml` (v2, in project root), generates settings.local.json with direct `ari hook` commands. The bash scripts in `.claude/hooks/ari/` are leftovers from the v1 architecture.

### ORPHAN: `hooks/*.sh` source scripts (7 files, ~8k bytes)

**Severity**: MEDIUM — these are SOURCE files in the `hooks/` directory alongside `hooks.yaml`

| File | Used by | Status |
|------|---------|--------|
| `autopark.sh` | Nothing | Dead — v2 calls Go directly |
| `clew.sh` | Nothing | Dead — v2 calls Go directly |
| `cognitive-budget.sh` | Nothing | Dead — v2 calls Go directly |
| `context.sh` | Nothing | Dead — v2 calls Go directly |
| `route.sh` | Nothing | Dead — v2 dropped this hook entirely |
| `validate.sh` | Nothing | Dead — v2 calls Go directly |
| `writeguard.sh` | Nothing | Dead — v2 calls Go directly |

**The v1→v2 migration kept the .sh files but nothing references them.** The only file that matters in `hooks/` is `hooks.yaml`.

### GHOST: `route` hook (bash-only, no Go implementation)

**Severity**: LOW — intentionally dropped, but leaves a confusing trail

- `hooks/route.sh` exists but dispatches to `ari hook route` (doesn't exist)
- `.claude/hooks/ari/route.sh` exists (orphan copy)
- `.claude/hooks/ari/hooks.yaml` (v1) references `ari/route.sh`
- `hooks/hooks.yaml` (v2) does NOT include route
- No `internal/cmd/hook/route.go` exists
- `UserPromptSubmit` CC event is known in `env.go` but unhandled

### DRIFT: `.claude/hooks/ari/*.sh` vs `hooks/*.sh`

The `.claude/` copies are from Jan 7 (old). The `hooks/` copies are from Feb 6-8 (newer, cleaned up). They differ:
- `.claude/` versions have `USE_ARI_HOOKS` feature flag
- `hooks/` versions removed the feature flag
- `.claude/hooks/ari/cognitive-budget.sh` is the OLD full implementation (5.1k)
- `hooks/cognitive-budget.sh` is the NEW thin wrapper (986 bytes)
- `.claude/hooks/ari/writeguard.sh` checks `_CONTEXT.md` AND `_CONTEXT.yaml`
- `hooks/writeguard.sh` only checks `_CONTEXT.md`

### BLOAT: `~/.claude/settings.local.json` permissions

**Severity**: LOW (cosmetic, not hook-related but worth noting)

217 allow entries accumulated from multiple projects (rulefy, mcp-coding-server, terraform). This is a user-level file, not project-scoped, so every project's permissions pile up.

### CLEAN: No user-defined hooks anywhere

No user hooks in settings.local.json, ~/.claude/settings.json, or project-scoped settings. All hooks are ari-managed.

---

## Architecture Diagram

```
CANONICAL SOURCE                     MATERIALIZED OUTPUT
================                     ===================

hooks/hooks.yaml (v2)  ──────────→  .claude/settings.local.json
  10 entries                          hooks: { ... }
  command: "ari hook X"               10 entries match 1:1

hooks/*.sh (7 files)   ──── DEAD ──── NOT REFERENCED
  thin wrappers                       by anything

.claude/hooks/ari/ (8 files) ─ DEAD ─ NOT REFERENCED
  v1 wrappers + stale yaml           by anything

internal/cmd/hook/*.go ←── CALLED BY ── settings.local.json
  10 subcommands                       via "ari hook X" commands
```

---

## Decision Matrix for Interview

| Item | Action Options | Considerations |
|------|---------------|----------------|
| `.claude/hooks/ari/` (8 files) | DELETE / KEEP | Dead code. But `.claude/` is mixed user+platform. Provenance? |
| `hooks/*.sh` (7 files) | DELETE / KEEP | Dead source. But historical reference? |
| `hooks/hooks.yaml` | KEEP | Canonical. Only file in hooks/ that matters |
| `route` hook gap | IMPLEMENT Go / ACCEPT gap | UserPromptSubmit unhandled. Intentional? |
| `~/.claude/settings.local.json` bloat | CLEAN permissions / LEAVE | User-level, cross-project accumulation |
| v1 hooks.yaml in `.claude/hooks/ari/` | DELETE with directory | Part of directory cleanup |
| `cognitive-budget.sh` full impl (5.1k) | DELETE | Fully replaced by Go `budget` |
| CC events without hooks (5 total) | IMPLEMENT / ACCEPT | PostToolUseFailure, PermissionRequest, Notification, TeammateIdle, TaskCompleted |

---

## Interview Questions for Remediation Plan

Ready for comprehensive stakeholder interview covering:

1. **Shell script deletion scope** — delete all 15 shell scripts (7 in hooks/, 7+1yaml in .claude/hooks/ari/)?
2. **`.claude/hooks/ari/` directory** — delete entire directory?
3. **`route` hook** — was the UserPromptSubmit hook intentionally dropped? Should it be reimplemented in Go?
4. **Unhandled CC events** — any need for PostToolUseFailure, PermissionRequest, Notification, TeammateIdle, TaskCompleted?
5. **User-level settings cleanup** — should we prune `~/.claude/settings.local.json` permissions?
6. **Provenance tracking** — should the cleanup be tracked through provenance system or is this a one-off?
7. **`mergeHooksSettings()` legacy stripping** — can we simplify `isLegacyPlatformHook()` after cleanup?
8. **Embedded hooks fallback** — is the `m.embeddedHooks` fallback still needed?
9. **`internal/hook/env.go` env var constants** — still needed after stdin migration?

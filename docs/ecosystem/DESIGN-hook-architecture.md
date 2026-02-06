# Context Design: Hook Architecture End State ("ari IS the hook binary")

**Author**: context-architect (Session 2, Ecosystem Rite)
**Date**: 2026-02-06
**Gap Analysis**: 6 gaps identified by ecosystem-analyst
**ADR Reference**: ADR-0011 (Hook Deprecation Timeline, Phase 2 overdue since 2026-02-04)

## Executive Summary

The hook system is stuck in an intermediate state: Go commands exist for 6 of 7 hooks, but bash wrappers remain as the execution layer and the materialization pipeline does not produce hook registrations. This design resolves all 6 gaps by: (1) implementing `ari hook budget` to port the last bash-only hook, (2) fixing the inverted feature flag, (3) wiring materialization to produce `settings.local.json` hook entries pointing directly at ari, (4) eliminating bash wrappers, and (5) removing dead library code.

The change is classified as **COMPATIBLE** for satellites that use knossos materialization. Satellites with hand-managed hooks require a documented migration (wrapper scripts stop being distributed).

## Architecture Diagram

```
CURRENT STATE (intermediate, broken)
=====================================

  Claude Code
      |
      | reads .claude/hooks/ari/*.sh (directory scan)
      |
  [bash wrapper]  ~5ms fast-path checks
      |
      | exec ari hook <subcommand>
      |
  [Go binary]     ~100ms full processing
      |
  [exit code + JSON stdout]


END STATE (target)
==================

  Claude Code
      |
      | reads .claude/settings.local.json "hooks" section
      |
  [ari hook <subcommand>]   single binary invocation
      |                      Go-native fast-path (<5ms for no-ops)
      |                      Full processing (<100ms)
      |
  [exit code + JSON stdout]


MATERIALIZATION FLOW (new)
==========================

  hooks.yaml (source of truth: user-hooks/ari/hooks.yaml)
      |
      | materializeSettingsWithManifest() reads hooks.yaml
      |
  settings.local.json
      |
      |  "hooks": {
      |    "PreToolUse": [
      |      {
      |        "matcher": "Edit|Write",
      |        "hooks": [{ "type": "command", "command": "ari hook writeguard --output json", "timeout": 3 }]
      |      },
      |      {
      |        "matcher": "Bash",
      |        "hooks": [{ "type": "command", "command": "ari hook validate --output json", "timeout": 5 }]
      |      }
      |    ],
      |    "PostToolUse": [
      |      {
      |        "matcher": "Edit|Write|Bash",
      |        "hooks": [{ "type": "command", "command": "ari hook clew --output json", "timeout": 5 }]
      |      },
      |      {
      |        "hooks": [{ "type": "command", "command": "ari hook budget --output json", "timeout": 3 }]
      |      }
      |    ],
      |    "SessionStart": [
      |      {
      |        "hooks": [{ "type": "command", "command": "ari hook context --output json", "timeout": 10 }]
      |      }
      |    ],
      |    "Stop": [
      |      {
      |        "hooks": [{ "type": "command", "command": "ari hook autopark --output json", "timeout": 5 }]
      |      }
      |    ],
      |    "UserPromptSubmit": [
      |      {
      |        "matcher": "^/",
      |        "hooks": [{ "type": "command", "command": "ari hook route --output json", "timeout": 5 }]
      |      }
      |    ]
      |  }
      |
  Claude Code discovers hooks at startup, calls ari directly
```

---

## 1. `ari hook budget` Design (GAP-1)

### Rationale

`cognitive-budget.sh` is the only hook with no Go equivalent. ADR-0011 explicitly lists `ari hook budget` as the target. The bash implementation is 149 LOC with file-based state management that is inherently race-prone and untestable.

### Alternative Considered and Rejected

**In-memory counter via clew contract**: Piggyback on the existing `clewcontract.RecordToolEvent()` call in the clew hook, counting events in the JSONL file instead of a separate counter file. Rejected because: (a) cognitive-budget fires on ALL PostToolUse, clew fires only on Edit|Write|Bash -- different scope; (b) counting JSONL lines on every call is O(n) vs O(1) for a counter file; (c) budget tracking is conceptually separate from artifact tracking and should not couple to clew.

### Command Specification

**File**: `internal/cmd/hook/budget.go`
**Registration**: Add to `hook.go` line 89: `cmd.AddCommand(newBudgetCmd(ctx))`

```
ari hook budget [--output json|text] [--timeout ms]
```

**Event**: PostToolUse (all tools, no matcher filter)
**Category**: DEFENSIVE -- must never block tool execution (exit 0 always)

### Configuration

| Env Var | Type | Default | Description |
|---------|------|---------|-------------|
| `ARIADNE_BUDGET_DISABLE` | bool ("1") | "0" (enabled) | Disable budget tracking entirely |
| `ARIADNE_MSG_WARN` | integer | 250 | Warning threshold (tool use count) |
| `ARIADNE_MSG_PARK` | integer | (none) | Park suggestion threshold (disabled if unset) |
| `ARIADNE_SESSION_KEY` | string | (none) | Explicit session key (testing override) |

### Session Key Resolution

Maintain identical precedence to bash implementation for backward compatibility:

1. `ARIADNE_SESSION_KEY` (explicit override for testing)
2. `CLAUDE_SESSION_ID` (Claude Code provided, most reliable)
3. `ppid-{PPID}` fallback (parent process ID)

In Go, PPID is obtained via `os.Getppid()`.

### State Store Design

**Location**: `/tmp/ariadne-budget-{session_key}` (renamed from `ariadne-msg-count-` to match Go naming but functionally identical)

**Rationale for keeping /tmp**: Session-scoped temp files are the correct abstraction. They auto-clean on reboot, have no cross-session leakage, and avoid polluting project directories. The session directory (`$PROJECT/.claude/sessions/{id}/`) was considered but rejected because budget tracking is not session-critical data and should not survive session recovery.

**File format**: Plain text integer (matches bash implementation for cross-version compatibility during rollout).

**Atomic write**: Use `os.CreateTemp` in `/tmp` + `os.Rename` (same mktemp+mv pattern as bash). Fallback to direct write on rename failure (cross-filesystem edge case).

**Marker files**:
- `/tmp/ariadne-budget-{session_key}.warned` -- one-shot warn threshold marker
- `/tmp/ariadne-budget-{session_key}.park-warned` -- one-shot park threshold marker

### Output Format

The budget hook is DEFENSIVE and does not produce output that modifies Claude behavior. It emits warnings to stderr only. Stdout output follows the standard hook JSON format for consistency:

```json
{
  "count": 42,
  "warn_threshold": 250,
  "park_threshold": null,
  "warned": false,
  "park_warned": false
}
```

**Stderr warning** (emitted once when threshold crossed):
```
[cognitive-budget] Warning: Tool use count (250) reached warning threshold (250). Consider using /park to preserve session state.
```

### Early Exit Conditions (ordered, all < 1ms)

1. `ARIADNE_BUDGET_DISABLE=1` -- exit immediately
2. `!hook.IsEnabled()` -- hooks disabled (after GAP-5 fix, this check becomes inert but remains for defense-in-depth)

### Internal Function Decomposition

| Function | Responsibility |
|----------|---------------|
| `newBudgetCmd(ctx)` | Cobra command registration |
| `runBudget(ctx)` | Orchestration: early exit, config, increment, check, output |
| `budgetConfig` struct | Parsed config from env vars (warn, park thresholds) |
| `parseBudgetConfig()` | Reads env vars, applies defaults |
| `resolveSessionKey(hookEnv)` | Session key resolution (3-tier) |
| `stateFilePath(sessionKey)` | Returns `/tmp/ariadne-budget-{key}` |
| `readCount(path)` | Reads integer from file, returns 0 if missing |
| `incrementCount(path)` | Atomic read-increment-write, returns new count |
| `checkThresholds(count, config, stateFile)` | Check + emit warnings + write markers |

### BudgetOutput Struct

```go
type BudgetOutput struct {
    Count          int    `json:"count"`
    WarnThreshold  int    `json:"warn_threshold"`
    ParkThreshold  *int   `json:"park_threshold"` // nil if disabled
    Warned         bool   `json:"warned"`
    ParkWarned     bool   `json:"park_warned"`
}

func (b BudgetOutput) Text() string {
    return fmt.Sprintf("Budget: %d/%d tool uses", b.Count, b.WarnThreshold)
}
```

### Test Strategy

| Test | Type | Description |
|------|------|-------------|
| `TestBudgetCmd_EarlyExit_Disabled` | Unit | ARIADNE_BUDGET_DISABLE=1 exits with empty output |
| `TestBudgetCmd_EarlyExit_HooksOff` | Unit | USE_ARI_HOOKS not set exits with empty output |
| `TestBudgetConfig_Defaults` | Unit | Default warn=250, park=nil |
| `TestBudgetConfig_Custom` | Unit | Custom env vars parsed correctly |
| `TestResolveSessionKey_Priority` | Unit | ARIADNE_SESSION_KEY > CLAUDE_SESSION_ID > ppid |
| `TestReadCount_NoFile` | Unit | Returns 0 when file missing |
| `TestIncrementCount_Atomic` | Unit | Creates file, increments, returns correct count |
| `TestIncrementCount_Concurrent` | Unit | 100 goroutines incrementing, final count >= 100 |
| `TestCheckThresholds_WarnOnce` | Unit | Warning emitted once, marker file prevents repeat |
| `TestCheckThresholds_ParkOnce` | Unit | Park warning emitted once when configured |
| `TestIntegration_BudgetHook_FullCycle` | Integration | 3 increments, check count=3, no warning |
| `TestIntegration_BudgetHook_WarnAt250` | Integration | Set warn=3, increment 3 times, verify warning |
| `BenchmarkBudget_IncrementPath` | Benchmark | Target: <5ms for full increment+check cycle |

---

## 2. Dead Code Removal (GAP-2)

### Files to Delete

| File | LOC | Rationale |
|------|-----|-----------|
| `.claude/hooks/lib/fail-open.sh` | 195 | Not sourced by any wrapper. Wrappers have inline `[[ -x "$ARI" ]] \|\| exit 0` pattern |
| `.claude/hooks/lib/preferences-loader.sh` | 423 | Not sourced by any wrapper. Go config package replaces this |

**Total**: 618 LOC removed

### Verification

Before deletion, confirm no sourcing via grep:
- Pattern: `source.*fail-open` or `. .*fail-open` -- expected 0 results (verified by analyst)
- Pattern: `source.*preferences-loader` or `. .*preferences-loader` -- expected 0 results (verified by analyst)

### Backward Compatibility: COMPATIBLE

These files are dead code. No wrapper imports them. Deletion has zero runtime impact.

---

## 3. Feature Flag Fix (GAP-5)

### Root Cause

In `internal/hook/env.go:92-95`:

```go
func IsEnabled() bool {
    val := os.Getenv(FeatureFlagEnvVar)
    return val == "1" || strings.ToLower(val) == "true"
}
```

This requires explicit `USE_ARI_HOOKS=1`. But bash wrappers use `${USE_ARI_HOOKS:-1}` (default enabled). Result: Go hooks are disabled by default while bash wrappers assume enabled by default.

### Design Decision: Remove Feature Flag Entirely

**Rationale**: ADR-0011 Phase 2 (due 2026-02-04, now overdue) calls for removing `USE_ARI_HOOKS`. The flag served as a rollback mechanism during the 30-day observation window. That window has passed. The Go hooks have been operational since 2026-01-04 (33 days). Keeping the flag adds complexity for zero benefit.

**Alternative considered**: Fix the default to `true`. Rejected because the flag is scheduled for removal and keeping it creates a third state (unset=enabled, "0"=disabled, "1"=enabled) that is confusing.

### Changes

**File**: `internal/hook/env.go`

Replace `IsEnabled()` at lines 92-95:

```go
// IsEnabled returns true if ari hooks should execute.
// As of ADR-0011 Phase 2, hooks are always enabled.
// USE_ARI_HOOKS=0 is the only way to disable (emergency kill switch).
func IsEnabled() bool {
    val := os.Getenv(FeatureFlagEnvVar)
    if val == "0" || strings.ToLower(val) == "false" {
        return false
    }
    return true // Default: enabled
}
```

This inverts the logic: enabled by default, only explicit `0` or `false` disables. This matches the bash wrapper behavior AND provides an emergency kill switch.

**File**: `internal/hook/env_test.go`

Update tests to match new default-enabled behavior.

**File**: `test/hooks/testutil/env.go`

The `SetupEnv` function currently only sets `USE_ARI_HOOKS` when `UseAriHooks: true`. With default-enabled behavior, tests that want hooks DISABLED must explicitly set `USE_ARI_HOOKS=0`. Update `SetupEnv` to set `USE_ARI_HOOKS=0` when `UseAriHooks: false`.

**File**: `internal/cmd/hook/hook.go`

Update the Long description at line 58 from:
```
USE_ARI_HOOKS=1    Enable ari hook implementations
```
to:
```
USE_ARI_HOOKS=0    Disable ari hook implementations (emergency only)
```

### Bash Wrapper Changes

Feature flag checks in all 7 wrappers (`[[ "${USE_ARI_HOOKS:-1}" != "1" ]] && exit 0`) are now redundant because the Go binary itself defaults to enabled. These lines remain temporarily as defense-in-depth until wrappers are eliminated in the end state (Section 5).

### Backward Compatibility: COMPATIBLE

Any environment with `USE_ARI_HOOKS=1` (explicit enable) continues to work. Any environment with `USE_ARI_HOOKS=0` (explicit disable) continues to work. Environments with no setting now default to enabled (matching bash wrapper behavior, fixing the bug).

---

## 4. Materialization Hook Pipeline (GAP-3 + GAP-4)

### Problem

Two separate materialization failures compound:
1. `materializeHooks()` looks for `templates/hooks/` which does not exist, so `HooksSkipped=true` always
2. `materializeSettingsWithManifest()` always writes `"hooks": {}` (empty), so Claude Code has no settings-based hook discovery

Currently hooks work ONLY because `ari sync user hooks` copies bash wrappers to `~/.claude/hooks/ari/` and Claude Code does directory scanning.

### Design Decision: Settings-Based Registration (Option B)

**Option A (copy wrappers) -- Rejected**: Copying bash wrapper scripts perpetuates the bash indirection layer. Every materialize would distribute 7 bash scripts that exist only to `exec` the Go binary. This is the wrong direction.

**Option B (settings.local.json registration) -- Selected**: Parse `hooks.yaml` and generate `settings.local.json` hook entries that call `ari` directly. This eliminates bash wrappers from the materialization path entirely.

**Option C (hybrid) -- Rejected**: Unnecessary complexity. Claude Code reads settings.local.json hooks; there is no need to also have wrapper scripts.

### hooks.yaml Schema (v2)

The existing `hooks.yaml` (v1) at `user-hooks/ari/hooks.yaml` serves as the canonical hook registry. Extend it with a `command` field for direct binary invocation:

```yaml
schema_version: "2.0"

# Binary resolution for direct invocation
binary: "ari"
binary_fallback:
  - "${ARIADNE_BIN}"
  - "ari"                              # PATH lookup
  - "${CLAUDE_PROJECT_DIR}/ariadne/ari" # Development

hooks:
  - event: SessionStart
    command: "ari hook context --output json"
    timeout: 10
    priority: 5
    description: "Injects session context via ari hook context"

  - event: Stop
    command: "ari hook autopark --output json"
    timeout: 5
    priority: 5
    description: "Auto-parks session on stop via ari hook autopark"

  - event: PreToolUse
    matcher: "Edit|Write"
    command: "ari hook writeguard --output json"
    timeout: 3
    priority: 3
    description: "Guards session context writes via ari hook writeguard"

  - event: PreToolUse
    matcher: "Bash"
    command: "ari hook validate --output json"
    timeout: 5
    priority: 5
    description: "Validates bash commands via ari hook validate"

  - event: PostToolUse
    matcher: "Edit|Write|Bash"
    command: "ari hook clew --output json"
    timeout: 5
    priority: 5
    description: "Tracks artifacts and commits via ari hook clew"

  - event: PostToolUse
    command: "ari hook budget --output json"
    timeout: 3
    priority: 90
    description: "Tracks tool use count for cognitive budget warnings"

  - event: UserPromptSubmit
    matcher: "^/"
    command: "ari hook route --output json"
    timeout: 5
    priority: 3
    description: "Routes slash commands via ari hook route"
```

**Validation rules for hooks.yaml entries**:
- `event`: required, must be one of: SessionStart, Stop, PreToolUse, PostToolUse, UserPromptSubmit
- `command`: required in v2, must be non-empty string
- `matcher`: optional, regex pattern string
- `timeout`: optional integer (seconds), default 10, max 30
- `priority`: optional integer 1-100, default 50
- `description`: required, max 200 characters

### Hook Entry Schema for settings.local.json

Claude Code expects hooks in `settings.local.json` in a nested format where each entry contains a `matcher` (optional) and a `hooks` array of hook objects:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "ari hook writeguard --output json",
            "timeout": 3
          }
        ]
      }
    ]
  }
}
```

Each entry has:
- `matcher` (optional): regex pattern for tool name filtering
- `hooks` (required): array of hook objects, each with:
  - `type` (required): hook type, currently always `"command"`
  - `command` (required): shell command to execute
  - `timeout` (optional): timeout in seconds

### Materialization Changes

**File**: `internal/materialize/materialize.go`

#### Change 1: Add hooks.yaml loading

New function `loadHooksConfig(claudeDir string)` that:
1. Looks for `hooks.yaml` at multiple resolution points (matching mena resolution):
   - `user-hooks/ari/hooks.yaml` (knossos platform level)
   - rite-level hooks config (future extensibility)
2. Parses YAML into `HooksConfig` struct
3. Returns nil if no hooks.yaml found (graceful)

#### Change 2: Modify materializeSettingsWithManifest

At line 785 where `existingSettings["hooks"]` is set to empty map, instead:

1. Load hooks config via `loadHooksConfig()`
2. If hooks config exists, build the hooks map from entries
3. Merge into existing settings (preserving user-added hooks not in our config)

```go
func (m *Materializer) buildHooksSettings(hooksConfig *HooksConfig) map[string]any {
    hooks := make(map[string]any)

    // Group entries by event type
    byEvent := make(map[string][]map[string]any)
    for _, entry := range hooksConfig.Hooks {
        hookObj := map[string]any{
            "type":    "command",
            "command": entry.Command,
        }
        if entry.Timeout > 0 {
            hookObj["timeout"] = entry.Timeout
        }
        hookEntry := map[string]any{
            "hooks": []map[string]any{hookObj},
        }
        if entry.Matcher != "" {
            hookEntry["matcher"] = entry.Matcher
        }
        byEvent[entry.Event] = append(byEvent[entry.Event], hookEntry)
    }

    for event, entries := range byEvent {
        hooks[event] = entries
    }

    return hooks
}
```

#### Change 3: Remove materializeHooks() dependency on templates/hooks

The `materializeHooks()` function at line 584 currently looks for `templates/hooks/` which does not exist. Two options:

**Selected**: Leave `materializeHooks()` as-is (it already returns `skipped=true` gracefully). The hook file copy mechanism is superseded by settings-based registration. The function becomes a no-op until `templates/hooks/` is populated for some future purpose. Document this explicitly with a code comment.

**Rejected alternative**: Delete `materializeHooks()`. Premature -- it may be needed for non-ari hook distribution in satellite projects.

### New Files

**File**: `internal/materialize/hooks.go`

Contains:
- `HooksConfig` struct (parsed hooks.yaml)
- `HookEntry` struct (individual hook entry)
- `loadHooksConfig()` function
- `buildHooksSettings()` function
- `mergeHooksSettings()` function (union merge preserving user entries)

### Merge Algorithm for Hooks

The hooks merge follows the same union-merge pattern as MCP servers:

1. Load existing `settings.local.json`
2. Parse existing `hooks` section (may contain user-defined hooks)
3. For each event type in hooks.yaml:
   - Replace all knossos-managed hooks for that event (identified by command prefix `ari hook`)
   - Preserve any user-defined hooks (commands not starting with `ari hook`)
4. Write merged result

This ensures:
- Materialization is idempotent (running twice produces identical output)
- User-defined hooks are never destroyed
- Knossos hooks are always up-to-date with hooks.yaml

### User Content Invariant

User-added hooks in `settings.local.json` that do not start with `ari hook` are preserved through materialization. This matches the satellite region / knossos region pattern used throughout the inscription system.

### Backward Compatibility: COMPATIBLE

Settings-based registration is additive. Existing `ari sync user hooks` continues to work (it copies files to `~/.claude/hooks/`). Claude Code discovers hooks from both directory scanning AND settings.local.json. During the transition period, both mechanisms produce the same result. After bash wrappers are removed, only settings-based registration remains.

---

## 5. End State: "ari IS the hook binary" (GAP-6)

### Binary Discovery

Claude Code calls the `command` string from `settings.local.json` hooks. The command is `ari hook <subcommand> --output json`. For this to work, `ari` must be on PATH.

**Discovery mechanism**: The `ari` binary is already resolved during `ari sync` / `ari materialize`. If `ari` is available to run materialization, it is available to be called as a hook binary. No additional discovery is needed -- PATH resolution is sufficient.

**Graceful degradation**: If `ari` is not on PATH when Claude Code invokes the hook command, the command fails with a non-zero exit code. Claude Code's hook protocol treats non-zero exit as a hook failure and continues (fail-open behavior is built into Claude Code). No bash wrapper needed for graceful degradation -- it is handled by the hook protocol itself.

### Performance

**Concern**: Bash wrappers provide ~5ms fast-path checks (tool name matching, feature flag) before incurring ~100ms Go startup cost. Without wrappers, every hook invocation pays Go startup.

**Analysis**: This concern is overstated for three reasons:

1. **Claude Code matcher field**: The `matcher` field in settings.local.json performs the same filtering that bash wrappers do. Claude Code only invokes hooks whose matcher matches. For example, writeguard is only called for Edit|Write tools -- the `[[ "$CLAUDE_HOOK_TOOL_NAME" != "Write" && ... ]]` check in the bash wrapper is redundant with the matcher.

2. **Go startup cost is not 100ms**: The hook commands import minimal packages. Measured startup for `ari hook validate` (the simplest) is ~15-25ms on macOS. The 100ms figure includes full processing time, not just startup.

3. **Budget for all hooks is generous**: Claude Code allows up to 10 seconds per hook invocation. Even at 25ms per hook, 7 hooks firing simultaneously use 175ms -- well within budget.

**If startup proves problematic in practice**: A future optimization can use a persistent ari daemon socket (ari hook-server) that eliminates per-invocation startup. This is explicitly NOT in scope for this design -- it is a premature optimization.

### What settings.local.json Looks Like

```json
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          { "type": "command", "command": "ari hook context --output json", "timeout": 10 }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          { "type": "command", "command": "ari hook autopark --output json", "timeout": 5 }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          { "type": "command", "command": "ari hook writeguard --output json", "timeout": 3 }
        ]
      },
      {
        "matcher": "Bash",
        "hooks": [
          { "type": "command", "command": "ari hook validate --output json", "timeout": 5 }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Edit|Write|Bash",
        "hooks": [
          { "type": "command", "command": "ari hook clew --output json", "timeout": 5 }
        ]
      },
      {
        "hooks": [
          { "type": "command", "command": "ari hook budget --output json", "timeout": 3 }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "matcher": "^/",
        "hooks": [
          { "type": "command", "command": "ari hook route --output json", "timeout": 5 }
        ]
      }
    ]
  },
  "mcpServers": {
    "go-semantic": {
      "command": "go-semantic-mcp"
    },
    "terraform": {
      "args": ["-y", "@modelcontextprotocol/server-terraform"],
      "command": "npx"
    }
  }
}
```

### Wrapper Elimination

After materialization produces settings-based hooks, bash wrappers become redundant:

1. `ari sync user hooks` stops distributing wrapper scripts (removes `user-hooks/ari/*.sh`)
2. `hooks.yaml` remains as the canonical source of truth
3. Existing `.claude/hooks/ari/*.sh` files in satellites are orphaned but harmless (Claude Code does not double-fire from both settings and directory scan for the same event+matcher)

Wrapper elimination happens as a follow-up cleanup, not as part of the critical path. The design is correct whether wrappers exist or not.

---

## 6. Implementation Sequence for Session 3

### Phase A: PATCH (immediate, no design decisions, 1-2 hours)

| # | Task | Files | GAP |
|---|------|-------|-----|
| A1 | Fix feature flag default | `internal/hook/env.go` (lines 92-95) | GAP-5 |
| A2 | Update env.go tests | `internal/hook/env_test.go` | GAP-5 |
| A3 | Update testutil for disabled case | `test/hooks/testutil/env.go` (line 53-55) | GAP-5 |
| A4 | Update hook.go help text | `internal/cmd/hook/hook.go` (line 58) | GAP-5 |
| A5 | Delete dead library files | `.claude/hooks/lib/fail-open.sh`, `.claude/hooks/lib/preferences-loader.sh` | GAP-2 |

**Verification**: `CGO_ENABLED=0 go test ./internal/hook/... ./internal/cmd/hook/... ./test/hooks/...`

### Phase B: MODULE (design + implementation, 3-4 hours)

| # | Task | Files | GAP |
|---|------|-------|-----|
| B1 | Implement `ari hook budget` command | `internal/cmd/hook/budget.go` (new) | GAP-1 |
| B2 | Register budget in hook group | `internal/cmd/hook/hook.go` (line 89) | GAP-1 |
| B3 | Write budget tests | `internal/cmd/hook/budget_test.go` (new) | GAP-1 |
| B4 | Create hooks.yaml loader | `internal/materialize/hooks.go` (new) | GAP-3, GAP-4 |
| B5 | Wire hooks into settings merge | `internal/materialize/materialize.go` (line 785) | GAP-4 |
| B6 | Update hooks.yaml to v2 | `user-hooks/ari/hooks.yaml` | GAP-4 |
| B7 | Write materialize hooks tests | `internal/materialize/hooks_test.go` (new) | GAP-3, GAP-4 |

**Verification**: `CGO_ENABLED=0 go test ./internal/cmd/hook/... ./internal/materialize/...`

### Phase C: INTEGRATION (end-to-end validation, 1-2 hours)

| # | Task | Description | GAP |
|---|------|-------------|-----|
| C1 | Run full materialization | `ari materialize ecosystem --force`, verify settings.local.json hooks populated | GAP-3, GAP-4 |
| C2 | Manual hook test | Invoke each `ari hook <cmd>` with appropriate env vars | GAP-1 |
| C3 | Budget regression | Compare `cognitive-budget.sh` behavior vs `ari hook budget` on same inputs | GAP-1 |
| C4 | Feature flag regression | Verify `USE_ARI_HOOKS=0` disables hooks, unset enables | GAP-5 |

### Phase D: DEFERRED (future session, SYSTEM scope)

| # | Task | Rationale |
|---|------|-----------|
| D1 | Eliminate bash wrappers from user-hooks | Requires satellite coordination; wrappers are harmless until removed |
| D2 | Remove `templates/hooks` reference from materializeHooks | Low priority; function is already a no-op |
| D3 | Performance profiling of direct ari invocation | Only needed if users report latency issues |
| D4 | ADR-0011 Phase 3 execution (delete .deprecated) | Separate initiative, 30-day observation window |

---

## 7. Integration Test Matrix

| Satellite Type | Test | Expected Outcome | GAPs Validated |
|----------------|------|------------------|----------------|
| **knossos (self)** | `ari materialize ecosystem --force` | settings.local.json contains all 7 hook entries, HooksSkipped=true (no templates/hooks), hooks section non-empty | GAP-3, GAP-4 |
| **knossos (self)** | `ari hook budget` with CLAUDE_SESSION_ID=test | Counter file created at /tmp/ariadne-budget-test, stdout JSON with count=1 | GAP-1 |
| **knossos (self)** | `ari hook budget` x 250 | Warning emitted to stderr once at count 250, marker file created | GAP-1 |
| **knossos (self)** | No USE_ARI_HOOKS set, `ari hook context` | Hooks execute (not early-exit), output produced | GAP-5 |
| **knossos (self)** | `USE_ARI_HOOKS=0 ari hook context` | Early exit, minimal output, <5ms | GAP-5 |
| **minimal satellite** | `ari materialize <rite> --force` | settings.local.json created with hooks from hooks.yaml, no errors | GAP-3, GAP-4 |
| **minimal satellite** | No hooks.yaml in rite | settings.local.json has `"hooks": {}`, no error | GAP-3, GAP-4 |
| **complex satellite** | `ari materialize` with existing user hooks in settings | User hooks preserved, ari hooks added/updated | GAP-4 |
| **complex satellite** | `ari materialize` twice | Identical output both times (idempotency) | GAP-3, GAP-4 |

---

## 8. Risk Analysis

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Go startup latency exceeds 100ms on slow machines | Low | Medium (perceived slowness) | Measured at 15-25ms on macOS; Claude Code allows 10s; monitor post-rollout |
| `ari` not on PATH in satellite environments | Medium | High (all hooks silently fail) | Claude Code fail-open ensures no blocking; add `ari` PATH check to `ari init` |
| Counter file race condition between concurrent hooks | Low | Low (off-by-one count, cosmetic) | Atomic rename; counter is advisory not critical |
| Existing user hooks in settings.local.json overwritten | Medium | High (user loses custom hooks) | Merge algorithm preserves non-`ari hook` commands; tested explicitly |
| ADR-0011 Phase 2 gate criteria not formally signed off | Medium | Low (procedural, not technical) | All technical criteria met; recommend orchestrator formally close gate |

---

## 9. Files Changed Summary

### New Files
| File | Purpose |
|------|---------|
| `internal/cmd/hook/budget.go` | ari hook budget command implementation |
| `internal/cmd/hook/budget_test.go` | Budget command tests |
| `internal/materialize/hooks.go` | hooks.yaml loader + settings builder |
| `internal/materialize/hooks_test.go` | Hooks materialization tests |

### Modified Files
| File | Change |
|------|--------|
| `internal/hook/env.go` | Fix IsEnabled() to default-enabled (lines 92-95) |
| `internal/hook/env_test.go` | Update tests for new default behavior |
| `internal/cmd/hook/hook.go` | Add budget subcommand registration (line 89), update help text (line 58) |
| `internal/materialize/materialize.go` | Wire hooks config into materializeSettingsWithManifest (line 785) |
| `user-hooks/ari/hooks.yaml` | Upgrade to schema v2 with command field |
| `test/hooks/testutil/env.go` | Set USE_ARI_HOOKS=0 when UseAriHooks=false (line 53-55) |

### Deleted Files
| File | Rationale |
|------|-----------|
| `.claude/hooks/lib/fail-open.sh` | Dead code, not sourced by any wrapper (195 LOC) |
| `.claude/hooks/lib/preferences-loader.sh` | Dead code, not sourced by any wrapper (423 LOC) |

---

## Design Decisions Log

| Decision | Selected | Rejected | Rationale |
|----------|----------|----------|-----------|
| Budget state store location | /tmp/ariadne-budget-{key} | Session directory | Budget data is ephemeral, should not survive session recovery, should not pollute project |
| Budget implementation approach | Dedicated counter file | Count JSONL events in clew | O(1) vs O(n) per invocation; different event scope (all vs Edit\|Write\|Bash) |
| Feature flag disposition | Default-enabled with kill switch | Fix default to true, keep full flag | ADR-0011 Phase 2 calls for removal; kill switch provides safety without complexity |
| Materialization approach | Settings-based (hooks.yaml to settings.local.json) | Copy bash wrappers | Eliminates bash indirection; Claude Code reads settings natively |
| Hook merge algorithm | Replace ari hooks, preserve user hooks | Full replace | User content invariant is a knossos platform guarantee |
| Bash wrapper elimination timing | Deferred to Phase D | Immediate removal | Wrappers are harmless; removal requires satellite coordination |
| Performance optimization | None (direct invocation) | Persistent daemon socket | Premature; measured startup is 15-25ms, well within budget |

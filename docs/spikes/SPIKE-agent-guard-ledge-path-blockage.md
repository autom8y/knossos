# SPIKE: Agent-Guard Blanket Deny Blocks .ledge/ Writes

**Date**: 2026-03-02
**Status**: Root Cause Found
**Severity**: P1 (blocks intended .ledge/ artifact workflow)

## Question

Why does the agent-guard hook block writes to `.ledge/specs/PRD-explain-tour.md` with "this agent is not permitted to Write files: ... is outside allowed paths" when `.ledge/` is the designated artifact output directory?

## Decision This Informs

How to fix the write guard so `.ledge/` paths work as intended, without breaking the agent-guard security model for subagents.

---

## Findings

### Root Cause: `writeDefaultSettings()` Creates a Blanket Deny

**File**: `/Users/tomtenuta/Code/knossos/internal/cmd/initialize/init.go` (lines 372-398)

```go
func writeDefaultSettings(claudeDir string) {
    settingsPath := filepath.Join(claudeDir, "settings.json")
    if _, err := os.Stat(settingsPath); !os.IsNotExist(err) {
        return
    }
    settingsContent := []byte(`{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": "ari hook agent-guard --output json"
          }
        ]
      }
    ]
  }
}
`)
    os.WriteFile(settingsPath, settingsContent, 0644)
}
```

**Two bugs in one:**

1. **No `--allow-path` flags**: The `isAllowedPath()` function in `agentguard.go` loops over an empty `allowPaths` slice, never matches, and returns false. Result: **unconditional deny for ALL writes**.

2. **No `--agent` flag**: Defaults to `"this agent"`, producing a confusing error message that doesn't identify which agent context is being enforced.

### Why This Fires

The `writeDefaultSettings()` function is called from `ari init` (both rite and minimal scaffold paths) at lines 230 and 255. It creates `.claude/settings.json` (the checked-in project settings file). CC loads hooks from **both** `settings.json` and `settings.local.json`, so the blanket agent-guard fires on every `Edit|Write|MultiEdit` operation in the project.

### The Error Chain

```
1. ari init → writeDefaultSettings() → .claude/settings.json created
2. Main thread attempts Write(.ledge/specs/PRD-explain-tour.md)
3. CC fires PreToolUse hooks from settings.json:
     "ari hook agent-guard --output json" (no --allow-path, no --agent)
4. agentguard.go → isAllowedPath(".../.ledge/specs/PRD-explain-tour.md", nil) → false
5. outputAgentDeny("this agent", "Write", "...is outside allowed paths")
6. Write blocked
```

### Key Distinction: settings.json vs settings.local.json

| File | Owner | Scope | Contents |
|------|-------|-------|----------|
| `settings.json` | Checked into repo / `ari init` | Project-wide | Blanket agent-guard (THE BUG) |
| `settings.local.json` | `ari sync` / materialize pipeline | Generated, gitignored | writeguard, validate, clew, budget, etc. |

CC merges hooks from both files. The `writeguard` in `settings.local.json` would ALLOW `.ledge/` paths (it only blocks protected context files), but the `agent-guard` from `settings.json` fires first and blocks everything.

### The Intent vs. Reality

The comment says: *"This ensures agent-guard hooks fire on foreign projects."* The intent was to provide a write guard for satellite projects initialized with `ari init`. But the implementation is a blanket deny with no escape hatch.

The `agent-guard` system works correctly when it's materialized per-agent via `transformAgentContent()` -- that path resolves `write-guard` frontmatter against `hook_defaults` from the manifest, producing proper `--allow-path` flags. The `writeDefaultSettings()` shortcut bypasses that entire resolution pipeline.

### .ledge/ Was Never Added to Allowed Paths

The `.ledge/` directory was introduced in commit `e8e56a1` with `scaffoldProjectDirs()` in `init.go`, but:
- It was never added to `shared/manifest.yaml` `hook_defaults.write_guard.allow_paths`
- It was never added to the `writeDefaultSettings()` template
- No rite manifest includes `.ledge/` in `allow_paths` or `extra_paths`
- The `writeguard.go` hook doesn't block `.ledge/` (it only blocks protected context files), so it's a non-issue there
- But the `agent-guard` blanket deny in `settings.json` blocks it

### Affected Scenarios

1. **Any satellite project initialized with `ari init`**: All writes blocked (not just `.ledge/`)
2. **Worktrees where `ari init` was run**: All writes blocked
3. **Main knossos repo**: Not affected (no `.claude/settings.json` exists in the repo root)

---

## Recommendation

### Immediate Fix (P1)

**Option A: Remove `writeDefaultSettings()` entirely.**

The `settings.json` agent-guard is redundant with the per-agent agent-guard hooks materialized by `transformAgentContent()`. The per-agent approach is correct (proper `--allow-path` flags from manifest `hook_defaults`). The project-level blanket deny adds no value and breaks all writes.

```go
// Delete writeDefaultSettings() function
// Delete both call sites (lines 230 and 255)
```

**Risk**: Satellite projects that rely on rite manifests with `hooks: [agent-guard]` already get agent-guard via per-agent materialization. Projects without write-guard in their manifests (like `10x-dev`) intentionally have no write restrictions on their agents.

**Option B: Fix `writeDefaultSettings()` to include proper allow-paths.**

```go
settingsContent := []byte(`{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": "ari hook agent-guard --allow-path .wip/ --allow-path wip/ --allow-path .ledge/ --allow-path docs/ --output json"
          }
        ]
      }
    ]
  }
}
`)
```

**Risk**: Hard-coding paths in a template means they drift from the manifest `hook_defaults`. This duplicates configuration that already lives in `shared/manifest.yaml`.

### Recommended: Option A

Remove `writeDefaultSettings()` entirely. The per-agent hook materialization pipeline is the correct mechanism for agent-guard enforcement. The project-level blanket is a shortcut that bypasses the resolution pipeline and creates an unconditional deny.

### Also Required: Add .ledge/ to shared hook_defaults

Regardless of which option is chosen, `.ledge/` must be added to the shared manifest's allowed paths so that rites with `write-guard` enabled can write to `.ledge/`:

**File**: `/Users/tomtenuta/Code/knossos/rites/shared/manifest.yaml` (line 38)

```yaml
hook_defaults:
  write_guard:
    allow_paths: [".wip/", "wip/", ".ledge/"]  # Add .ledge/
    timeout: 3
```

---

## Follow-Up Actions

1. **Delete `writeDefaultSettings()`** from `init.go` and remove both call sites
2. **Add `.ledge/` to `shared/manifest.yaml`** `hook_defaults.write_guard.allow_paths`
3. **Delete stale `settings.json`** from any affected worktrees/satellites
4. **Add test** to verify `.ledge/` paths are allowed by agent-guard when shared defaults are applied
5. **Audit existing satellites** for broken `settings.json` files created by previous `ari init` runs

## Files Involved

| File | Role |
|------|------|
| `internal/cmd/initialize/init.go` | Contains buggy `writeDefaultSettings()` |
| `internal/cmd/hook/agentguard.go` | Agent-guard logic (correct, but invoked without allow-paths) |
| `internal/materialize/hookdefaults.go` | Correct per-agent hook resolution pipeline |
| `internal/materialize/agent_transform.go` | Correct per-agent hook materialization |
| `rites/shared/manifest.yaml` | Missing `.ledge/` in `allow_paths` |
| `knossos/templates/sections/know.md.tpl` | Documents `.ledge/` as artifact output directory |
| `.claude/settings.json` (worktree) | The file containing the blanket deny |

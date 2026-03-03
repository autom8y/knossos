# Knossos Codebase Context Document

> Generated 2026-02-08 from 6 parallel exploration agents. Input for hygiene code-smeller.

---

## 1. Package Dependency Graph

```
cmd/ari
  -> internal/cmd/sync/     (materialize CLI)
  -> internal/cmd/session/   (session CLI)
  -> internal/cmd/hook/      (hook CLI)
  -> internal/cmd/agent/     (agent CLI)

internal/cmd/sync/
  -> internal/materialize/   (core pipeline)
  -> internal/paths/

internal/cmd/session/
  -> internal/session/       (FSM, discovery, context, events, rotation)
  -> internal/lock/          (JSON lock v2)
  -> internal/hook/clewcontract/
  -> internal/sails/

internal/cmd/hook/
  -> internal/hook/          (env, input, output, context)
  -> internal/hook/clewcontract/
  -> internal/session/       (rotation, from precompact)

internal/materialize/
  -> internal/inscription/   (CLAUDE.md generation)
  -> internal/agent/         (frontmatter parsing)
  -> internal/paths/
  -> internal/errors/

internal/inscription/
  -> internal/errors/

internal/agent/
  -> internal/errors/
  -> internal/validation/    (JSON schema)

internal/hook/
  -> internal/errors/

internal/hook/clewcontract/
  -> (standalone, minimal deps)

internal/session/
  -> internal/errors/

internal/lock/
  -> (standalone, syscall only)
```

---

## 2. Package Metrics

| Package | Core LOC | Test LOC | Files (core) | Files (test) | Exported Fns | Test:Code |
|---------|----------|----------|--------------|--------------|--------------|-----------|
| internal/materialize/ | 2,700 | 4,522 | 6 | 14 | 28 | 1.67:1 |
| internal/inscription/ | 3,692 | 5,540 | 7 | 7 | 78 | 1.50:1 |
| internal/session/ | 995 | 3,662 | 7 | 6 | ~30 | 3.68:1 |
| internal/lock/ | 303 | 492 | 1 | 1 | ~10 | 1.62:1 |
| internal/agent/ | 1,634 | ~1,500 | 9 | ~5 | 11 | 0.92:1 |
| internal/hook/ | 782 | ~800 | 4 | ~4 | ~40 | 1.02:1 |
| internal/hook/clewcontract/ | 1,364 | ~600 | 5 | ~3 | ~30 | 0.44:1 |
| internal/cmd/hook/ | 1,300 | 3,400 | 9 | ~9 | 8 | 2.62:1 |
| internal/cmd/session/ | 2,615 | 6,960 | 14 | ~12 | ~14 | 2.66:1 |
| mena/ (content) | 23,706 | - | 152 | - | - | - |
| **TOTAL** | **~39,091** | **~27,476** | **~214** | **~61** | **~249** | **0.70:1** |

---

## 3. Duplication Inventory

### HIGH: Lock Reading Logic (2 implementations)

| Location | Function | Behavior |
|----------|----------|----------|
| `internal/cmd/hook/writeguard.go:141-190` | `isMoiraiLockHeld()` | Inline JSON parse, returns bool, fail-closed |
| `internal/cmd/session/lock.go:223-235` | `readMoiraiLock()` + `isLockStale()` | Reusable functions, returns struct+error |

**Risk**: Lock format change requires updating both locations.
**Fix**: Export `readMoiraiLock`/`isLockStale` from session/lock.go, import in writeguard.go.

### HIGH: Stale Lock Detection (2 implementations)

| Location | Function |
|----------|----------|
| `internal/lock/lock.go` | `isStale()` (private) + `IsStaleForTest()` (exported) |
| `internal/cmd/session/recover.go` | `isAdvisoryLockStale()` (local copy) |

**Risk**: Same logic duplicated. Naming inconsistency (`isStale` vs `isAdvisoryLockStale`).
**Fix**: Recover should call `lock.IsStaleForTest()` or a properly-exported version.

### MEDIUM: Atomic Write File (2 implementations)

| Location | Function |
|----------|----------|
| `internal/inscription/backup.go:349-369` | `AtomicWriteFile()` |
| `internal/materialize/materialize.go:287` | `atomicWriteFile()` |
| `internal/session/rotation.go` | `atomicWriteFile()` |

**Risk**: Three copies of temp-file-then-rename pattern.
**Fix**: Extract to shared utility package (e.g., `internal/fileutil/`).

### MEDIUM: Frontmatter Parsing (2 implementations)

| Location | Purpose |
|----------|---------|
| `internal/agent/frontmatter.go` | Agent frontmatter (17 fields) |
| `internal/materialize/frontmatter.go` | Mena frontmatter (11 fields) |

**Risk**: Different YAML delimiter handling, different error recovery. Agent uses `ParseAgentFrontmatter`, mena uses `parseMenaFrontmatterBytes`.
**Note**: May be intentional (different schemas), but delimiter detection logic is duplicated.

### LOW: Agent Info Extraction (2 implementations)

| Location | Purpose |
|----------|---------|
| `internal/inscription/pipeline.go:24-58` | `extractFrontmatter()` for inscription render context |
| `internal/agent/frontmatter.go` | `ParseAgentFrontmatter()` for validation |

**Risk**: Pipeline's extraction is simplified (fewer fields). Could drift from canonical parser.

### LOW: copyDir Variants (3 implementations)

| Location | Function |
|----------|----------|
| `internal/materialize/materialize.go:1339-1366` | `copyDir()` - filesystem |
| `internal/materialize/project_mena.go:415-446` | `copyDirWithStripping()` - filesystem + extension strip |
| `internal/materialize/project_mena.go:450-476` | `copyDirFromFSWithStripping()` - embedded FS + extension strip |

**Risk**: Three similar copy implementations.

---

## 4. Dead Code Candidates

### Confirmed Dead/Unused

| Location | Symbol | Evidence |
|----------|--------|----------|
| `internal/materialize/materialize.go:1228-1230` | `materializeSettings()` | Wrapper that calls `materializeSettingsWithManifest(claudeDir, nil)` — no callers |
| `internal/materialize/materialize.go:1329-1336` | `getCurrentRite()` | Reads ACTIVE_RITE file — no callers found in package or codebase |
| `internal/materialize/materialize.go:1339-1366` | `copyDir()` | Filesystem copy — no callers (superseded by `copyDirWithStripping`) |
| `internal/agent/` | `containsStr()` | Private helper — grep shows no callers |

### Deprecated But Present

| Location | Symbol | Status |
|----------|--------|--------|
| `internal/materialize/materialize.go:146-210` | `StagedMaterialize()` | Explicitly deprecated, breaks CC file watcher. Tests exist but no production callers. |
| `internal/materialize/materialize.go:213-243` | `cloneDir()` | Only used by `StagedMaterialize`. Remove together. |
| `internal/materialize/materialize.go:323-326` | `Materialize()` | Legacy wrapper, pass-through to `MaterializeWithOptions`. |
| `internal/inscription/marker.go:248-298` | `ParseLegacyMarkers()` | Detects old `<!-- PRESERVE: -->` markers. Called in tests only, not in pipeline. |

### Implemented But Not Wired

| Location | Symbol | Status |
|----------|--------|--------|
| `internal/session/rotation.go` | `RotateSessionContext()` | Fully implemented + tested (331 test lines) but NOT called from any command (create, wrap, etc.) |

---

## 5. Error Handling Patterns

### Fail-Open (Graceful Degradation)

| Location | Behavior |
|----------|----------|
| Hook budget (cmd/hook/budget.go:96) | Counter errors return success: "fail-open" |
| Hook output encoding (hook/output.go:156) | Errors → DecisionAllow |
| Settings loading (materialize.go:1240) | Missing/malformed JSON → empty map |
| Hooks.yaml loading (hooks.go:39-82) | Missing → nil (no hooks) |
| Mena source collection (project_mena.go:254) | Missing dir → skip with `continue` |
| Inscription backup failure (pipeline.go:258) | Logs warning, continues sync |
| Session event emission (cmd/session/*.go) | Failures logged, don't block operation |
| Lock acquisition on reads (cmd/session/status.go) | Shared lock failure → continue without lock |

### Fail-Closed (Conservative)

| Location | Behavior |
|----------|----------|
| Writeguard lock check (cmd/hook/writeguard.go:140) | Any error reading lock → deny (block write) |
| Rite resolution (materialize/source.go) | Rite not found → hard error |
| Manifest parsing (inscription/manifest.go) | Invalid YAML → error |
| Agent validation STRICT mode | Missing fields → validation failure |
| Session FSM transitions | Invalid transition → `ErrLifecycleViolation` |
| White Sails gate (cmd/session/wrap.go) | BLACK sails → block wrap (unless --force) |

### Silent Swallows (Potential Issues)

| Location | Behavior | Risk |
|----------|----------|------|
| Mena frontmatter parse (materialize/frontmatter.go) | Malformed YAML → zero-value struct (EC-7) | Mena with bad frontmatter silently treated as unscoped |
| Inscription backup cleanup (backup.go:196) | Delete errors ignored | Orphaned backups |
| Discovery frontmatter read (session/discovery.go) | Parse error → empty string | Could mask corrupted SESSION_CONTEXT |

---

## 6. Exported API Surface (Candidates for Internal)

### Functions Only Called Within Their Own Package

| Package | Function | Notes |
|---------|----------|-------|
| materialize | `materializeSettings()` | Dead code — remove |
| materialize | `getCurrentRite()` | Dead code — remove |
| materialize | `copyDir()` | Dead code — remove |
| materialize | `Materialize()` | Legacy wrapper — deprecate |
| inscription | `ParseLegacyMarkers()` | Only in tests, migration helper |

### Functions With Single External Caller

| Package | Function | Caller |
|---------|----------|--------|
| session | `RotateSessionContext()` | Only cmd/hook/precompact.go (new, untracked) |
| lock | `IsStaleForTest()` | Exported only for cross-package testing |
| agent | `ValidateAgentMCPReferences()` | Only cmd/agent validate |
| agent | `ValidateAgentMCPToolReferences()` | Only cmd/agent validate |

---

## 7. Inconsistencies

### Provenance Detection (4 different strategies)

| Pipeline Phase | Detection Method | Mechanism |
|----------------|-----------------|-----------|
| Agents | Manifest membership | `manifest.Agents[].Name + ".md"` |
| Hooks | Template filename match | Files in `templates/hooks/` |
| Rules | Template filename match | Files in `templates/rules/` |
| Mena | Frontmatter scope field | `scope: project \| user` |

No unified provenance tracking. Each uses a different heuristic.

### Write Patterns (Mixed)

| Location | Pattern | Safe? |
|----------|---------|-------|
| Most of materialize | `writeIfChanged()` → `atomicWriteFile()` | Yes |
| Orphan backup | `os.WriteFile()` direct | No (partial writes visible) |
| Legacy CLAUDE.md migration | `os.WriteFile()` direct | No |
| Session context saves | `os.WriteFile()` direct | No |
| Inscription | `AtomicWriteFile()` | Yes |

7 direct `os.WriteFile()` calls remain in materialize package.

### Naming Inconsistencies

| Concept | Name A | Name B | Location |
|---------|--------|--------|----------|
| Stale detection | `isStale()` | `isAdvisoryLockStale()` | lock.go vs recover.go |
| MCP extraction | `MCPServers()` method | `mcpServers` CC field | agent package |
| Lock reading | `isMoiraiLockHeld()` | `readMoiraiLock()` | writeguard.go vs lock.go |
| Mena type | `DetectMenaType()` | `RouteMenaFile()` | Same data, different names |

### Error Code Overuse

32 error wraps in materialize.go use `CodeGeneralError` for almost everything. No error type hierarchy distinguishes "rite not found" from "disk write failed" from "manifest parse error" at runtime.

---

## 8. Key Architectural Invariants

1. **User content NEVER destroyed** — satellite regions, user-agents, user-hooks preserved
2. **Idempotency** — running materialize twice produces identical output
3. **Selective write** — only knossos-managed files are updated/removed
4. **Atomic per-file writes** — no partial-write visibility to file watchers
5. **FSM enforcement** — ARCHIVED is terminal, transitions validated
6. **Scan-based discovery** — no in-memory session cache, eliminates TOCTOU
7. **Fail-open hooks** — hook errors never block tool execution (except writeguard)
8. **Write guard** — context files protected via PreToolUse hook + Moirai lock bypass

---

## 9. Scope Field Gap

The `scope` field exists in:
- Go schema (`MenaScope` enum: `""`, `"user"`, `"project"`)
- Frontmatter parsing code (fully implemented)
- Pipeline filtering logic (fully implemented)

But **0 mena files** actually use `scope: project` or `scope: user`. All default to `""` (both pipelines). The entire scope filtering infrastructure is unused in practice.

---

## 10. Rotation Gap

`RotateSessionContext()` in `internal/session/rotation.go`:
- Fully implemented (236 lines)
- Fully tested (331 test lines)
- Default thresholds: rotate at 200 lines, keep 80 lines
- Archives to `SESSION_CONTEXT.archived.md`

But **NOT wired** into any session command (create, park, wrap). The precompact hook (`internal/cmd/hook/precompact.go`) calls it, but precompact.go itself is untracked (new file).

SESSION_CONTEXT.md observed sizes: up to 355 lines / 15.9KB — unbounded growth without rotation.

---

## 11. CC Alignment Gaps (Agent Schema)

| CC Field | Knossos Support | Notes |
|----------|----------------|-------|
| maxTurns | Supported | In frontmatter, per-archetype defaults |
| disallowedTools | Supported | FlexibleStringSlice |
| skills | Supported | []string |
| tools | Supported | With MCP validation |
| model | Supported | opus/sonnet/haiku |
| type | Supported | 7 types |
| aliases | Supported | []string |
| memory | NOT supported | Not in schema |
| permissionMode | NOT supported | Not in schema |
| mcpServers | Partial | Inferred from `tools` via `mcp:` prefix, not declared directly |
| hooks | NOT supported | Not in schema |
| context/fork | NOT supported | Not in schema |

---

## 12. File Output Catalog (.claude/ writes)

| Path | Written By | Method |
|------|-----------|--------|
| `.claude/agents/*.md` | materializeAgents | writeIfChanged (selective) |
| `.claude/commands/**` | materializeMena (ProjectMena) | copyDirWithStripping |
| `.claude/skills/**` | materializeMena (ProjectMena) | copyDirWithStripping |
| `.claude/hooks/**` | materializeHooks | writeIfChanged (selective) |
| `.claude/rules/*.md` | materializeRules | writeIfChanged (selective) |
| `.claude/CLAUDE.md` | materializeCLAUDEmd (inscription) | AtomicWriteFile |
| `.knossos/KNOSSOS_MANIFEST.yaml` | inscription Save | AtomicWriteFile |
| `.claude/settings.local.json` | materializeSettingsWithManifest | writeIfChanged |
| `.claude/ACTIVE_RITE` | writeActiveRite | writeIfChanged |
| `.claude/ACTIVE_WORKFLOW.yaml` | materializeWorkflow | writeIfChanged |
| `.knossos/sync/state.json` | trackState | writeIfChanged |
| `.sos/sessions/*/SESSION_CONTEXT.md` | session commands | os.WriteFile (via Moirai) |
| `.sos/sessions/*/.moirai-lock` | lock command | os.WriteFile |
| `.sos/sessions/*/events.jsonl` | EventEmitter | append |

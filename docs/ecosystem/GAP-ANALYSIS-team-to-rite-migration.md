# Gap Analysis: Backward Compatibility Excision (TRACK 8)

**Date**: 2026-01-06
**Analyst**: Ecosystem Analyst
**Scope**: Inventory all backward compatibility patterns for team-to-rite migration and assess removal risk
**Classification**: SYSTEM (touches multiple components, external consumers, requires migration period)

---

## Executive Summary

The codebase has accumulated backward compatibility patterns during the team-to-rite migration. These patterns need to be inventoried and scheduled for removal based on risk assessment. This analysis identifies all patterns, assesses satellite impact, and recommends a removal timeline.

**Pattern Categories**:
1. CLI flags (`--team` should be `--rite`)
2. JSON fields (`active_rite` should be `active_rite`)
3. File fallbacks (`ACTIVE_RITE` fallback to `ACTIVE_RITE`)
4. Command deprecation (`ari team` should be `ari rite`)
5. Function naming (internal variable/function names)

---

## 1. CLI Flag Inventory

### 1.1 Ariadne Go Code

| File | Line(s) | Flag | Status | Deprecation Warning |
|------|---------|------|--------|---------------------|
| `ariadne/internal/cmd/session/create.go` | 44-45, 53 | `--team` / `-t` | Deprecated | YES - "use --rite instead" |
| `ariadne/internal/cmd/team/status.go` | 33, 35 | `--team` / `-t` | Deprecated | YES - "use --rite instead" |
| `ariadne/internal/cmd/team/validate.go` | 28, 31 | `--team` / `-t` | Deprecated | YES - "use --rite instead" |
| `ariadne/internal/cmd/team/context.go` | 36, 39 | `--team` / `-t` | Deprecated | YES - "use --rite instead" |

**Assessment**: Go CLI flags properly deprecated with warnings. Safe for v2.0 removal.

### 1.2 Shell Scripts

| File | Line(s) | Flag | Status | Notes |
|------|---------|------|--------|-------|
| `roster-sync` | 181-187, 317-321 | `--team` / `-t` | Active | No deprecation warning |
| `roster-sync` | 246, 274 | `--team` in help text | User-facing | Help text shows `--team` as primary |

**Assessment**: `roster-sync --team` has no deprecation warning. Consumers may depend on this flag.

### 1.3 Documentation References

| File | Pattern | Count |
|------|---------|-------|
| `user-commands/session/start.md` | `--team=PACK` | 3 |
| `user-commands/workflow/sprint.md` | `--team=10x-dev` | 3 |
| `user-commands/navigation/worktree.md` | `--team=PACK` | 5 |
| `user-skills/session-lifecycle/start-ref/SKILL.md` | `--team` | 3 |
| `rites/forge/commands/eval-agent.md` | `--team=<team-name>` | 2 |

**Assessment**: 16+ documentation files reference `--team` flag. Must update docs alongside flag removal.

---

## 2. JSON Field Inventory

### 2.1 JSON Output Fields

| Location | Field | Current | Target | Risk Level |
|----------|-------|---------|--------|------------|
| `swap-rite.sh:860` | JSON output | `"active_rite": "$rite_name"` | `"active_rite"` | HIGH - external consumers |
| `swap-rite.sh:1004` | JSON output | `"active_rite": "unknown"` | `"active_rite"` | HIGH |
| `session-manager.sh:272` | JSON output | `"active_rite": "$active_rite"` | `"active_rite"` | MEDIUM |

### 2.2 Go Struct JSON Tags

| File | Struct Field | JSON Tag | Assessment |
|------|--------------|----------|------------|
| `ariadne/internal/output/team.go:16` | `ActiveRite` | `json:"active_rite"` | MIGRATED |
| `ariadne/internal/output/team.go:68` | `Team` | `json:"team"` | SEMANTIC (keep) |
| `ariadne/internal/output/team.go:142` | `Team` | `json:"team"` | Migrate to `"rite"` |
| `ariadne/internal/output/team.go:143` | `PreviousTeam` | `json:"previous_team"` | Migrate to `"previous_rite"` |
| `ariadne/internal/output/output.go:266` | `Team` | `json:"team"` | SEMANTIC (keep) |

### 2.3 YAML Schema Fields

| File | Field | Consumer Impact |
|------|-------|-----------------|
| `user-skills/session-common/session-context-schema.md` | `active_rite: string` | SESSION_CONTEXT.md files |
| `user-skills/session-common/sprint-context-schema.md` | `active_rite: string` | SPRINT_CONTEXT.md files |
| Various test fixtures | `active_rite:` | Test files only |

**Assessment**: Schema migration requires dual-read support during transition. Existing sessions have `active_rite:` in their YAML frontmatter.

---

## 3. File Fallback Inventory

### 3.1 ACTIVE_RITE Fallback Pattern

The pattern `cat ACTIVE_RITE 2>/dev/null || cat ACTIVE_RITE` exists for backward compatibility:

| File | Line(s) | Pattern | Removal Risk |
|------|---------|---------|--------------|
| `user-hooks/validation/orchestrator-router.sh` | 54 | `cat ".claude/ACTIVE_RITE" 2>/dev/null \|\| cat ".claude/ACTIVE_RITE"` | LOW |
| `user-hooks/lib/session-manager.sh` | 89, 243, 294 | Same pattern | LOW |
| `user-hooks/lib/rite-context-loader.sh` | 45-46 | Same pattern | LOW |
| `user-hooks/lib/worktree-manager.sh` | 207 | Same pattern | LOW |
| `user-hooks/session-guards/start-preflight.sh` | 113 | Same pattern | LOW |

### 3.2 ACTIVE_RITE File Writes

| File | Line(s) | Operation | Notes |
|------|---------|-----------|-------|
| `roster-sync` | 538 | `echo "$rite_name" > .claude/ACTIVE_RITE` | CRITICAL - still writes ACTIVE_RITE |
| `lib/rite/rite-transaction.sh` | 582-588 | Backup to `ACTIVE_RITE` | Internal backup name |
| `swap-rite.sh` | 169-176 | Restore from `ACTIVE_RITE` backup | Internal backup name |

**CRITICAL FINDING**: `roster-sync` at line 538 still writes to `ACTIVE_RITE` instead of `ACTIVE_RITE`. This is a bug, not a backward compatibility pattern.

### 3.3 ACTIVE_RITE in Error Messages

| File | Line(s) | Message | Assessment |
|------|---------|---------|------------|
| `swap-rite.sh` | 678, 684 | "ACTIVE_RITE is $active_rite" | INCORRECT - file is ACTIVE_RITE |
| `swap-rite.sh` | 670 | "No ACTIVE_RITE file" | INCORRECT |
| `lib/rite/rite-transaction.sh` | 692-693 | "Backup missing ACTIVE_RITE" | INCORRECT |
| `get-workflow-field.sh` | 21 | "No team specified and no ACTIVE_RITE found" | INCORRECT |
| `load-workflow.sh` | 12 | "No team specified and no ACTIVE_RITE found" | INCORRECT |

**Assessment**: Error messages reference wrong file name. These are bugs, not compatibility patterns.

---

## 4. Command Deprecation Status

### 4.1 `ari team` Command

| Status | Implementation |
|--------|----------------|
| Deprecation warning | YES - `team.go:38-39` prints warning to stderr |
| Functional | YES - all subcommands work |
| Documentation | Mixed - some docs still reference `ari team` |

### 4.2 Command Alias Completeness

| `ari team` subcommand | `ari rite` equivalent | Status |
|-----------------------|----------------------|--------|
| `ari team list` | `ari rite list` | Implemented |
| `ari team switch` | `ari rite swap` | Implemented |
| `ari team status` | `ari rite current` | Implemented |
| `ari team validate` | `ari rite validate` | Implemented |
| `ari rite context` | `ari rite context` | NOT IMPLEMENTED |

**Gap**: `ari rite context` command does not exist yet.

---

## 5. Function Naming Inventory

### 5.1 Shell Functions with "team" in Name

| File | Function | Assessment |
|------|----------|------------|
| `swap-rite.sh` | `update_active_rite()` | Rename to `update_active_rite()` |
| `roster-sync` / `lib/sync/sync-core.sh` | `refresh_active_rite()` | Rename to `refresh_active_rite()` |
| `lib/rite/rite-transaction.sh` | `stage_active_rite()` | ALREADY MIGRATED |

### 5.2 Shell Variables with "team" in Name

| File | Variable | Occurrences | Assessment |
|------|----------|-------------|------------|
| `swap-rite.sh` | `active_rite`, `current_team`, `rite_name` | ~50 | Internal - low priority |
| `roster-sync` | `active_rite`, `manifest_team`, `rite_name` | ~30 | Internal - low priority |
| Various test files | `active_rite` | ~40 | Internal - low priority |

### 5.3 Go Struct Fields and Parameters

| File | Field/Parameter | Assessment |
|------|-----------------|------------|
| `ariadne/internal/team/switch.go` | `activeTeamPath`, `activeTeamData` | Rename to `activeRitePath`, `activeRiteData` |
| `ariadne/internal/team/context_loader.go` | `teamsDir` parameter | Keep for now (Go package rename risky) |
| `ariadne/internal/paths/paths.go` | `TeamDir()`, `TeamAgentsDir()`, etc. | Add `RiteDir()` aliases |

---

## 6. Satellite Impact Assessment

### 6.1 Known Satellites (from rites/ directory)

| Satellite | Configuration | Risk Assessment |
|-----------|---------------|-----------------|
| 10x-dev | Full ecosystem | MEDIUM - may have custom scripts |
| ecosystem | Full ecosystem | LOW - maintained in-repo |
| forge | Full ecosystem | LOW - maintained in-repo |
| docs | Full ecosystem | LOW - maintained in-repo |
| hygiene | Minimal | LOW |
| debt-triage | Minimal | LOW |
| sre | Minimal | LOW |
| intelligence | Minimal | LOW |
| rnd | Minimal | LOW |
| security | Minimal | LOW |
| strategy | Minimal | LOW |

### 6.2 External Consumer Patterns

| Pattern | External Risk | Consumer Type |
|---------|---------------|---------------|
| `roster-sync --team` | HIGH | Any project using roster-sync CLI |
| `ari team switch` | MEDIUM | Users with muscle memory |
| `active_rite` JSON field | HIGH | Scripts parsing JSON output |
| `ACTIVE_RITE` file | MEDIUM | Legacy satellites not yet migrated |
| `active_rite:` YAML field | HIGH | Existing SESSION_CONTEXT.md files |

### 6.3 Breaking Change Impact

Removing backward compatibility will break:

1. **External satellites** that:
   - Call `roster-sync --team` in CI/CD scripts
   - Parse `active_rite` from JSON output
   - Have existing sessions with `active_rite:` in YAML
   - Still have `ACTIVE_RITE` file (unlikely after roster-sync)

2. **User workflows** that:
   - Use `ari team` commands in scripts or aliases
   - Expect `--team` flag in documentation examples

---

## 7. Removal Timeline Recommendations

### 7.1 Immediate Removal (Safe)

These are bugs, not compatibility patterns:

| Item | File | Action |
|------|------|--------|
| `roster-sync` writes `ACTIVE_RITE` | Line 538 | Change to `ACTIVE_RITE` |
| "ACTIVE_RITE" in error messages | Multiple | Update to "ACTIVE_RITE" |
| Incorrect comments | Multiple | Update |

### 7.2 v2.0 Removal (Needs Deprecation Period)

These have deprecation warnings but consumers may still depend on them:

| Item | Deprecation Status | Removal Criteria |
|------|-------------------|------------------|
| `--team` CLI flags | WARNING added | 1 major version after warning |
| `ari team` command | WARNING added | 1 major version after warning |
| `active_rite` JSON field | NOT deprecated | Add `active_rite`, deprecate `active_rite` |
| `ACTIVE_RITE` file fallback | N/A | Remove when no satellites have ACTIVE_RITE |

### 7.3 Keep Indefinitely (Too Risky to Remove)

| Item | Reason |
|------|--------|
| `source: "team"` in AGENT_MANIFEST.json | Semantic - describes where agent came from |
| Go package name `internal/team/` | Breaking change to all imports |
| Cross-team handoff skills | Semantic - actual team coordination |

---

## 8. Risk Matrix

| Pattern | Impact | Likelihood | Detection | Timeline |
|---------|--------|------------|-----------|----------|
| `roster-sync --team` removal | HIGH | HIGH | Low - silent failure | v2.0 |
| `active_rite` JSON removal | HIGH | MEDIUM | Medium - JSON parse errors | v2.0 |
| `ACTIVE_RITE` fallback removal | MEDIUM | LOW | High - explicit fallback | v1.5 |
| `ari team` removal | LOW | LOW | High - deprecation warning | v2.0 |
| Function renames | LOW | LOW | Low - internal only | v1.x |

---

## 9. Success Criteria

- [ ] `roster-sync` writes to `ACTIVE_RITE` not `ACTIVE_RITE`
- [ ] All error messages reference `ACTIVE_RITE`
- [ ] `roster-sync --help` shows `--rite` as primary, `--team` as deprecated
- [ ] `swap-rite.sh` JSON output includes `active_rite` (with `active_rite` for compat)
- [ ] SESSION_CONTEXT.md schema accepts both `active_rite` and `active_rite`
- [ ] All 10 satellites pass sync without errors
- [ ] `grep -r "ACTIVE_RITE" --include="*.sh" | grep -v "backup\|comment"` returns only intentional patterns

---

## 10. Test Satellites for Verification

| Satellite | Purpose |
|-----------|---------|
| test-satellite-baseline | Minimal config, verify clean ACTIVE_RITE creation |
| test-satellite-minimal | No custom settings, verify default behavior |
| test-satellite-complex | Nested arrays, custom hooks, verify no regressions |
| test-satellite-legacy | Has ACTIVE_RITE file, verify migration path |
| test-satellite-session | Has SESSION_CONTEXT with active_rite field |

---

## 11. Recommended Implementation Order

### Phase 1: Bug Fixes (Immediate)
1. Fix `roster-sync` line 538 to write `ACTIVE_RITE`
2. Update error messages referencing `ACTIVE_RITE`
3. Update `get-workflow-field.sh` and `load-workflow.sh` messages

### Phase 2: Deprecation Warnings (v1.x)
1. Add deprecation warning to `roster-sync --team`
2. Add `active_rite` to JSON output alongside `active_rite`
3. Update session schema to accept both fields

### Phase 3: Documentation Update
1. Update all `--team` examples to `--rite`
2. Update user-facing error messages
3. Update SKILL.md files

### Phase 4: Full Removal (v2.0)
1. Remove `--team` flag support
2. Remove `active_rite` JSON field
3. Remove `ACTIVE_RITE` file fallback
4. Remove `ari team` command

---

## Artifact Attestation

| Source | Lines Analyzed |
|--------|----------------|
| `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/session/create.go` | 1-389 |
| `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/team/team.go` | 1-106 |
| `/Users/tomtenuta/Code/roster/roster-sync` | Grep results |
| `/Users/tomtenuta/Code/roster/swap-rite.sh` | Grep results |
| `/Users/tomtenuta/Code/roster/lib/rite/rite-transaction.sh` | 1-699 |
| `/Users/tomtenuta/Code/roster/user-hooks/lib/session-manager.sh` | Grep results |
| Various documentation files | Grep results |
| `/Users/tomtenuta/Code/roster/rites/` | Directory listing |

# Source-Repo Reference Audit

**Date**: 2026-02-05
**Task**: task-002 (A2) -- Sprint: sprint-mechanical-foundation-20260205
**Scope**: Generator defaults, templates, current CLAUDE.md, materialized output

## Summary

- Total references found: 31
- Legitimate (KEEP): 9
- Satellite-unsafe (FIX in Sprint 2): 22

### Key Finding

The largest cluster of satellite-unsafe references comes from **knossos-owned sections that lack template files** and exist only as hardcoded content in the current `.claude/CLAUDE.md`. These sections (knossos-identity, hooks, ariadne-cli, getting-help, state-management, skills) are registered in `KNOSSOS_MANIFEST.yaml` as `owner: knossos` but have no corresponding `.md.tpl` file and no `getDefault*Content()` method in `generator.go`. They are legacy content that would be synced verbatim to satellites, carrying source-repo-specific paths.

## Findings

### Generator Defaults (generator.go)

| Line | Reference | Classification | Rationale |
|------|-----------|---------------|-----------|
| 398 | `PRD-hybrid-session-model` | FIX | Source-repo document reference. Satellites do not have this PRD. Appears in `getDefaultExecutionModeContent()`. |
| 408 | `orchestration/execution-mode.md` | KEEP | Refers to a materialized command path (`.claude/commands/orchestration/execution-mode.md`), which satellites would have after sync. |
| 437 | `ari --help` | KEEP | CLI binary reference; satellites with Knossos installed would have `ari`. |
| 439 | `cd ariadne && just build` | FIX | Source-repo build command. Satellites do not have an `ariadne/` directory. Appears in `getDefaultPlatformInfrastructureContent()`. |
| 439 | `go test ./...` | FIX | Source-repo test command. Satellites may not be Go projects. Appears in `getDefaultPlatformInfrastructureContent()`. |
| 278 | `ariadne` (in `lookupTerminology()`) | KEEP | Terminology definition, not a file path. Used for `{{ term "ariadne" }}` in templates. |

### Templates (knossos/templates/sections/)

| File | Reference | Classification | Rationale |
|------|-----------|---------------|-----------|
| `execution-mode.md.tpl:6` | `PRD-hybrid-session-model` | FIX | Source-repo document reference. Satellite projects do not have this PRD. |
| `execution-mode.md.tpl:16` | `orchestration/execution-mode.md` | KEEP | Refers to materialized command path that satellites would have. |
| `platform-infrastructure.md.tpl:6` | `ari --help` | KEEP | CLI reference; valid if Knossos is installed. |
| `platform-infrastructure.md.tpl:8` | `cd ariadne && just build` | FIX | Source-repo build instructions. Satellites do not have `ariadne/` directory. |
| `platform-infrastructure.md.tpl:8` | `go test ./...` | FIX | Source-repo test command. Not applicable to non-Go satellites. |
| `quick-start.md.tpl:9` | `partials/agent-table.md.tpl` (include) | KEEP | Internal template include, not output to CLAUDE.md. |
| `user-content.md.tpl:15` | `ari inscription add-region` | KEEP | CLI usage example in satellite-owned section; satellites would have `ari`. |
| `agent-configurations.md.tpl:13` | `ari team switch` | KEEP | CLI reference; valid for satellites with Knossos. |
| `quick-start.md.tpl:13` | `ari team switch` | KEEP | CLI reference; valid for satellites with Knossos. |

### Current CLAUDE.md (source-repo instance)

These are knossos-owned sections in the current `.claude/CLAUDE.md` that contain source-repo-specific references. Sections marked with "*" have **no template file** and **no getDefault method** -- they exist only as legacy content that would need to be generated somehow for satellites.

| Section | Line | Reference | Classification | Rationale |
|---------|------|-----------|---------------|-----------|
| `execution-mode` | 4 | `PRD-hybrid-session-model` | FIX | Source-repo PRD reference. Already covered by template analysis above. |
| `execution-mode` | 14 | `orchestration/execution-mode.md` | KEEP | Materialized command path, available in satellites. |
| `knossos-identity`* | 20 | `roster/.claude/ IS Knossos` | FIX | Hardcoded repo name. Satellite would incorrectly say "roster/.claude/". |
| `knossos-identity`* | 22 | `docs/philosophy/knossos-doctrine.md` | FIX | Source-repo doc path. Satellites do not have `docs/philosophy/`. |
| `knossos-identity`* | 33 | `docs/guides/knossos-integration.md` | FIX | Source-repo doc path. |
| `knossos-identity`* | 33 | `docs/decisions/ADR-0009-knossos-roster-identity.md` | FIX | Source-repo ADR path. |
| `skills`* | 64 | `ecosystem-ref` (roster ecosystem patterns) | FIX | Source-repo specific skill description mentioning "roster". |
| `skills`* | 64 | `.claude/skills/` and `~/.claude/skills/` | FIX | References `skills/` directory which is being unified to `commands/`. Stale after ADR-0021. |
| `ariadne-cli`* | 124 | `cd ariadne && just build` | FIX | Source-repo build path. |
| `ariadne-cli`* | 126 | `docs/guides/knossos-integration.md` | FIX | Source-repo doc path. |
| `getting-help`* | 138 | `ecosystem-ref` (Roster ecosystem) | FIX | "Roster ecosystem" is source-repo specific context. |
| `getting-help`* | 139 | `docs/guides/user-preferences.md` | FIX | Source-repo doc path. |
| `getting-help`* | 140 | `docs/guides/knossos-integration.md` | FIX | Source-repo doc path. |
| `getting-help`* | 141 | `docs/guides/knossos-migration.md` | FIX | Source-repo doc path. |
| `state-management`* | 164 | `docs/requirements/PRD-foo.md` | KEEP | Example path pattern (uses placeholder `PRD-foo`), not a real reference. |
| `state-management`* | 168 | `.sos/sessions/{session-id}/SESSION_CONTEXT.md` | KEEP | Templated path with placeholder; satellites would have this structure. |
| `state-management`* | 171 | `.claude/hooks/lib/session-manager.sh` | FIX | Source-repo hook path. Satellites may not have this file. |
| `state-management`* | 185 | `user-agents/moirai.md` | FIX | Source-repo agent file path. Not materialized to `.claude/`. |
| `state-management`* | 185 | `docs/philosophy/knossos-doctrine.md` | FIX | Source-repo doc path. |

### Materialized Output (.claude/)

| File | Reference | Classification | Rationale |
|------|-----------|---------------|-----------|
| `.claude/commands/spike.md:49` | `/docs/spikes/SPIKE-{slug}.md` | KEEP | Convention for spike output location. Template path pattern; satellite spike command would create this directory. |
| `.claude/agents/*.md` (6 files) | `roster` (41 occurrences) | KEEP | Agent files reference "CEM/roster" as domain terminology within the ecosystem rite. These agents are rite-specific and only materialized when the ecosystem rite is active. They would not appear in satellites running other rites. |

## Analysis

### Root Cause

The source-repo reference problem has two distinct origins:

1. **Generator defaults with hardcoded source-repo content** (`generator.go` lines 395-468): The `getDefaultExecutionModeContent()` and `getDefaultPlatformInfrastructureContent()` methods contain source-repo-specific strings (`PRD-hybrid-session-model`, `cd ariadne && just build`, `go test ./...`). These defaults render when no template file exists, but even the template files (`execution-mode.md.tpl`, `platform-infrastructure.md.tpl`) contain the same hardcoded references.

2. **Legacy knossos-owned sections without templates**: Six sections in `KNOSSOS_MANIFEST.yaml` are `owner: knossos` but have no `.md.tpl` template file and no `getDefault*Content()` fallback:
   - `knossos-identity` -- 4 source-repo references
   - `skills` -- 2 source-repo references
   - `ariadne-cli` -- 2 source-repo references
   - `getting-help` -- 4 source-repo references
   - `state-management` -- 3 source-repo references
   - `hooks` -- 0 source-repo references (clean)
   - `dynamic-context` -- 0 source-repo references (clean)

   These sections only exist as literal content in the current `.claude/CLAUDE.md`. The inscription system would fail to regenerate them for satellites because there is no generation path for them.

### Impact on Satellites

If a satellite runs `ari sync` (or equivalent), these knossos-owned sections would either:
- **Render from templates** (execution-mode, platform-infrastructure): Satellites get source-repo build commands and PRD references
- **Fail to render** (knossos-identity, getting-help, etc.): No template and no default method means `getDefaultSectionContent()` returns an error for unknown section names
- **Render stale content** (skills): References `.claude/skills/` which is being replaced by `.claude/commands/`

### FIX Priority

| Priority | Count | Description |
|----------|-------|-------------|
| P0 | 6 | Sections without template files that cannot render for satellites |
| P1 | 4 | `docs/` path references in templates that would break portability |
| P2 | 3 | Source-repo build/test commands in platform-infrastructure |
| P3 | 2 | Stale `skills/` references (ADR-0021 migration artifact) |

## Recommendations for Sprint 2 (C1: source-repo cleanup)

1. Create template files for all 6 missing sections or remove them from the default manifest
2. Replace hardcoded `docs/` paths with generic pointers or make them conditional on project context
3. Make `platform-infrastructure` build/test commands configurable via `KnossosVars` or manifest metadata
4. Update `skills` section references to `commands` per ADR-0021
5. Remove `PRD-hybrid-session-model` reference or make it a configurable variable

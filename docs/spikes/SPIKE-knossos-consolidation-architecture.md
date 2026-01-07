# SPIKE: Knossos Consolidation Architecture

**Date**: 2026-01-07
**Initiative**: knossos-finalization
**Phase**: SPIKE (research only)
**Session**: session-20260107-164631-8dd6f03a
**Upstream**: Code Smeller inventory (23 findings)

---

## Executive Summary

This SPIKE evaluates four critical architectural smells identified by the Code Smeller and proposes target architectures for consolidation. The core finding is that the repository has accumulated three parallel implementations of the same infrastructure (hooks, shell libraries) due to organic growth, and the Go module identity remains tied to an external namespace (`github.com/autom8y/knossos`) rather than Knossos-native identity.

**Root Cause Cluster**: All four critical smells stem from a single architectural issue: **lack of canonical source locations** established during rapid iteration. The fix is not incremental cleanup but **declaring authoritative locations and eliminating alternatives**.

---

## 1. Boundary Analysis

### 1.1 Current State Map

```
roster/
+-- .claude/hooks/          [RUNTIME: 348KB, 35 files]  <-- Active for project
|   +-- lib/                [220KB, 16 files]
|
+-- ariadne/
|   +-- .claude/hooks/      [ORPHAN: 300KB, 30 files]   <-- Stale copy
|   |   +-- lib/            [176KB, 12 files]
|   +-- internal/cmd/sync/  [GO SYNC: 20+ files]        <-- Go implementation
|
+-- user-hooks/             [TEMPLATE: 172KB, 13 files] <-- For satellite projects
|   +-- lib/
|
+-- lib/
|   +-- sync/               [SHELL SYNC: 7 files]       <-- Shell implementation
|
+-- swap-rite.sh            [MONOLITH: 3,773 lines, 68 functions]
+-- roster-sync             [SHELL: 1,413 lines]
```

### 1.2 Intended Architecture (Target)

```
roster/
+-- .claude/hooks/          [CANONICAL: All hooks here]
|   +-- lib/                [Shell libraries for hooks]
|
+-- ariadne/
|   +-- internal/           [Go CLI implementation]
|   +-- (no .claude/ directory - delegates to root)
|
+-- user-hooks/             [SYNC TARGET: Template for satellite projects]
|   +-- (copy of .claude/hooks/ for roster-sync distribution)
|
+-- lib/
|   +-- rite/               [swap-rite modules]
|   +-- sync/               [roster-sync modules - TO DEPRECATE]
```

---

## 2. Critical Smell Analysis

### SM-001: Triple Hook Duplication (CRITICAL)

**Current State**:
- `.claude/hooks/` - 348KB, 35 shell files (ACTIVE)
- `ariadne/.claude/hooks/` - 300KB, 30 shell files (STALE)
- `user-hooks/` - 172KB, 13 shell files (TEMPLATE)

**Divergence Evidence**:
```
Files that DIFFER between .claude/hooks and ariadne/.claude/hooks:
- lib/session-manager.sh    (rite vs team terminology)
- lib/session-fsm.sh        (rite vs team terminology)
- lib/session-state.sh      (rite vs team terminology)
- lib/rite-context-loader.sh
- lib/worktree-manager.sh
- validation/command-validator.sh
- base_hooks.yaml

Files ONLY in .claude/hooks:
- context-injection/orchestrated-mode.sh
- lib/artifact-validation.sh
- lib/fail-open.sh
- lib/handoff-validator.sh
- lib/preferences-loader.sh
```

**Root Cause**: `ariadne/` was created as a subproject with its own `.claude/` directory, then evolved independently. The "team to rite" terminology migration was applied to `.claude/hooks/` but not propagated to `ariadne/.claude/hooks/`.

**Target Architecture**:
- **CANONICAL**: `.claude/hooks/` (single source of truth)
- **REMOVE**: `ariadne/.claude/hooks/` (eliminate entirely)
- **SYNC-TARGET**: `user-hooks/` (derived from canonical via roster-sync)

**Complexity**: MEDIUM
- 30 files to remove from `ariadne/.claude/hooks/`
- Must verify `ariadne/` doesn't source these directly
- `user-hooks/` sync mechanism already exists

**Risk Assessment**:
- Blast radius: Low (ariadne hooks appear unused)
- Detection: Run ariadne tests, verify no import failures
- Rollback: Git revert single commit

---

### SM-002: Go Module Path Mismatch (CRITICAL)

**Current State**:
```go
// ariadne/go.mod
module github.com/autom8y/knossos
```

**Problem**: The module path references an external namespace (`autom8y`) rather than Knossos identity. This affects:
- 155 Go files with import paths
- All test files
- External package references

**Occurrences by Area**:
| Directory | Files Affected |
|-----------|----------------|
| ariadne/internal/cmd/ | 65 |
| ariadne/internal/ (other) | 55 |
| ariadne/docs/ | 5 |
| Other | 30 |

**Target Architecture**:
Two viable options:

**Option A: Knossos-Native Path**
```go
module github.com/knossos-platform/ariadne
```
- Pros: Reflects true identity, clean namespace
- Cons: Requires creating GitHub org, external dependency coordination

**Option B: Roster-Local Path**
```go
module roster/ariadne
```
- Pros: Simple, no external dependencies
- Cons: Non-standard for Go, may confuse tooling

**Recommendation**: Option A with fallback to Option B if GitHub org creation is out of scope.

**Complexity**: HIGH
- 155+ files require import path updates
- All test files must be updated
- go.sum regeneration required
- CI/CD pipeline updates

**Risk Assessment**:
- Blast radius: High (all Go code affected)
- Detection: `go build`, `go test`
- Rollback: Git revert, restore go.mod/go.sum

**USER DECISION REQUIRED**: Choose module path namespace.

---

### SM-003: Shell Library Divergence (CRITICAL)

**Current State**:
```
.claude/hooks/lib/     220KB  (16 files) - ACTIVE
ariadne/.claude/hooks/lib/  176KB  (12 files) - STALE
user-hooks/lib/        172KB  (13 files) - TEMPLATE
```

**Total Drift**: ~100KB of duplicated shell code with divergent implementations.

**Specific Divergences Observed**:
- `session-manager.sh`: "rite" vs "team" in comments and variable names
- `session-fsm.sh`: Same terminology drift
- `worktree-manager.sh`: Different function signatures

**Root Cause**: Same as SM-001. Shell libraries evolved in `.claude/hooks/lib/` without propagation.

**Target Architecture**:
```
.claude/hooks/lib/     [CANONICAL]
     |
     +--- roster-sync ---> user-hooks/lib/  [DERIVED]
     |
     X--- ariadne/.claude/hooks/lib/  [REMOVED]
```

**Complexity**: MEDIUM
- Consolidation is deletion (ariadne copy)
- Sync mechanism already exists for user-hooks
- Must verify no direct sourcing from ariadne path

**Risk Assessment**:
- Blast radius: Low (same as SM-001)
- Detection: Grep for `ariadne/.claude/hooks/lib` sourcing
- Rollback: Git revert

---

### SM-004: Monolithic swap-rite.sh (CRITICAL)

**Current State**:
- **Location**: `/Users/tomtenuta/Code/roster/swap-rite.sh`
- **Size**: 3,773 lines
- **Functions**: 68 shell functions
- **Responsibilities**: Transaction management, validation, backup, swap, sync, orphan handling, recovery

**Function Clusters Identified**:

| Cluster | Functions | Lines (est.) | Cohesion |
|---------|-----------|--------------|----------|
| **Logging** | `log`, `log_error`, `log_warning`, `log_debug` | 50 | High |
| **Transaction** | `rollback_swap`, `handle_interrupt`, `handle_exit`, `setup_signal_handlers`, `check_journal_recovery` | 400 | High |
| **Recovery** | `prompt_recovery_action`, `continue_interrupted_swap`, `recover_partial_commit`, `complete_partial_commit` | 300 | High |
| **Manifest** | `read_manifest`, `get_agent_from_manifest`, `write_manifest`, `init_manifest_from_existing` | 250 | High |
| **Validation** | `validate_agent_tools`, `validate_rite_tools`, `validate_rite`, `validate_project`, `validate_workflow_yaml`, `validate_orchestrator_yaml`, `validate_rite_schemas` | 350 | High |
| **Swap Operations** | `backup_current_agents`, `swap_agents`, `swap_commands`, `swap_skills`, `swap_hooks` | 500 | Medium |
| **Orphan Handling** | `detect_orphans`, `format_orphan`, `stash_kept_agents`, `restore_kept_agents`, `promote_agents`, `cleanup_stash`, `prompt_disposition`, `cleanup_orphan_backups` | 400 | High |
| **Query** | `query_current_rite`, `list_rites` | 100 | High |
| **Sync Integration** | `roster_sync_available`, `roster_has_updates`, `run_roster_sync_waterfall` | 150 | Medium |
| **Core** | `perform_swap`, `perform_reset`, `main`, `usage`, `preview_refresh`, `preview_reset`, `update_claude_md`, `update_active_rite` | 800 | Low |

**Decomposition Strategy**:

```
lib/rite/
+-- swap-core.sh       [Core swap logic, ~500 lines]
+-- swap-transaction.sh [Transaction/recovery, ~400 lines]
+-- swap-manifest.sh    [Manifest operations, ~250 lines]
+-- swap-validation.sh  [All validation, ~350 lines]
+-- swap-orphan.sh      [Orphan handling, ~400 lines]
+-- swap-sync.sh        [roster-sync integration, ~150 lines]
+-- swap-ui.sh          [User prompts, usage, ~300 lines]
+-- common.sh           [Logging, constants, ~100 lines]

swap-rite.sh           [Entry point only, ~100 lines]
```

**Target**: 8 modules averaging 310 lines each vs. 1 file at 3,773 lines.

**Complexity**: HIGH
- 68 functions to relocate
- Function dependencies must be mapped
- Source ordering matters (bash `source` semantics)
- Tests must continue passing

**Risk Assessment**:
- Blast radius: High (swap-rite.sh is critical path)
- Detection: Existing test suite, manual swap verification
- Rollback: Git revert decomposition commits

**Phasing Recommendation**:
1. Extract `common.sh` (logging, constants) - LOW RISK
2. Extract `swap-validation.sh` - MEDIUM RISK
3. Extract `swap-manifest.sh` - MEDIUM RISK
4. Extract `swap-orphan.sh` - MEDIUM RISK
5. Extract `swap-transaction.sh` - HIGH RISK (error handling)
6. Extract `swap-sync.sh` - LOW RISK
7. Extract `swap-ui.sh` - LOW RISK
8. Refactor `swap-core.sh` - HIGH RISK

---

## 3. Additional Findings Assessment

### SM-005: Legacy 'team' Terminology in Go (38 files, 129 occurrences)

**Assessment**: DEFERRED to terminology migration sprint
**Rationale**: Does not block consolidation, purely cosmetic
**Recommendation**: Bundle with SM-002 Go module migration if paths are changing anyway

### SM-007: 1MB Orphaned Backup Directories

**Locations**:
```
.claude/commands.orphan-backup    12KB
.claude/skills.backup             12KB
.claude/agents.backup             40KB
.claude/commands.backup           0KB
.claude/skills.orphan-backup      1.0MB  <-- Primary target
```

**Assessment**: LOW priority, simple cleanup
**Recommendation**: Add to .gitignore, remove from repo

### SM-008: 48KB Implementation Docs at Repo Root

**Files**:
```
CONTEXT_SEED.md               5KB
DEFECT-D002-RESOLUTION.md     6KB
IMPLEMENTATION-SUMMARY-*.md   6KB
IMPLEMENTATION_SUMMARY.md     8KB
IMPLEMENTATION_VERIFICATION.md 10KB
PHASE-5-HANDOFF.md            10KB
RITE_SKILL_MATRIX.md          12KB
SAILS-STATUS-REFERENCE.md     3KB
```

**Assessment**: LOW priority, documentation housekeeping
**Recommendation**: Move to `docs/archive/` or `docs/implementation/`

### SM-012: roster-sync Shell vs Go Duplication

**Current State**:
- `roster-sync` (shell): 1,413 lines, full implementation
- `ariadne/internal/cmd/sync/`: ~20 files, Go implementation

**Assessment**: STRATEGIC decision required
**Options**:
1. Keep both (shell for simplicity, Go for Claude Code integration)
2. Deprecate shell in favor of Go (aligns with ariadne as canonical CLI)
3. Deprecate Go in favor of shell (if ariadne scope narrows)

**Recommendation**: Option 2 - Go becomes canonical, shell deprecated
**USER DECISION REQUIRED**: Confirm sync implementation strategy

### SM-013: Skills Scattered Across 4+ Locations

**Locations**:
```
.claude/skills/           [Active project skills - 30+ directories]
ariadne/.claude/skills/   [Orphan - 8 directories]
rites/*/skills/           [Per-rite skills - varies]
skills/                   [Root-level - 1 directory]
```

**Assessment**: By design (rite isolation), not a smell
**Clarification**: `.claude/skills/` contains currently-active skills, `rites/*/skills/` contains rite-specific skills loaded on swap. This is intentional bounded context isolation.

**No Action Required**: Architecture is correct, just needs documentation.

---

## 4. Target Architecture Summary

### Canonical Locations (Post-Consolidation)

| Resource | Canonical Location | Derived From | Removed |
|----------|-------------------|--------------|---------|
| Hooks | `.claude/hooks/` | user-hooks/ (via sync) | ariadne/.claude/hooks/ |
| Shell libs | `.claude/hooks/lib/` | user-hooks/lib/ (via sync) | ariadne/.claude/hooks/lib/ |
| Go CLI | `ariadne/` | - | - |
| Swap logic | `lib/rite/` (new) | - | swap-rite.sh (refactored) |
| Sync logic | `ariadne/internal/sync/` | - | roster-sync (deprecated) |
| Skills | `.claude/skills/` + `rites/*/skills/` | - | ariadne/.claude/skills/ |

### Module Identity

**Recommended Go Module Path**: `github.com/knossos-platform/ariadne`

Requires:
- GitHub organization creation (`knossos-platform`)
- Repository transfer or new repo with history
- Import path updates across 155 files

**Fallback**: `roster/ariadne` if external namespace is out of scope

---

## 5. Feasibility Assessment

| Smell | Complexity | Risk | Effort (days) | Dependencies |
|-------|------------|------|---------------|--------------|
| SM-001: Hook duplication | MEDIUM | LOW | 1-2 | None |
| SM-002: Go module path | HIGH | HIGH | 3-5 | User decision |
| SM-003: Shell lib divergence | MEDIUM | LOW | 1 | SM-001 |
| SM-004: swap-rite.sh decomposition | HIGH | HIGH | 5-8 | None |

**Total Estimated Effort**: 10-16 days

---

## 6. Phase 2 Prerequisites (for R&D Rite)

Before implementation can begin, the following must be resolved:

### Decisions Required

1. **Go Module Namespace**: Choose `github.com/knossos-platform/ariadne` vs `roster/ariadne`
2. **Sync Strategy**: Confirm Go sync as canonical, shell as deprecated
3. **Decomposition Priority**: Confirm swap-rite.sh decomposition is in scope

### Prototypes Needed

1. **Hook Removal Verification**: Confirm ariadne functions without `.claude/hooks/`
   ```bash
   # Prototype test
   mv ariadne/.claude/hooks ariadne/.claude/hooks.bak
   cd ariadne && go test ./...
   ```

2. **Go Module Migration**: Test import path rewrite tooling
   ```bash
   # Prototype with gofmt or gopls
   gofmt -r 'github.com/autom8y/knossos -> github.com/knossos-platform/ariadne' ariadne/
   ```

3. **swap-rite.sh Module Extraction**: Extract `common.sh` as proof-of-concept
   ```bash
   # Extract logging functions, verify sourcing works
   ```

### Documentation Updates

- ADR for Go module identity decision
- Migration runbook for import path updates
- Updated repository structure diagram

---

## 7. Recommendations

### Immediate Actions (No Risk)

1. Remove `ariadne/.claude/hooks/` (SM-001) - appears completely orphaned
2. Remove backup directories (SM-007) - no functional impact
3. Move root implementation docs to `docs/archive/` (SM-008)

### Short-Term Actions (Low Risk)

4. Align `user-hooks/` with `.claude/hooks/` via roster-sync
5. Extract `lib/rite/common.sh` from swap-rite.sh (proof of concept)

### Medium-Term Actions (Requires Planning)

6. Complete swap-rite.sh decomposition into `lib/rite/`
7. Deprecate `roster-sync` shell in favor of `ari sync`

### Long-Term Actions (Requires Decision)

8. Migrate Go module path (requires namespace decision)
9. Complete terminology migration in Go code

---

## Appendix A: File Inventory

### ariadne/.claude/hooks/ (To Remove)

```
ariadne/.claude/hooks/
+-- ari/
|   +-- autopark.sh
|   +-- clew.sh
|   +-- cognitive-budget.sh
|   +-- context.sh
|   +-- route.sh
|   +-- validate.sh
|   +-- writeguard.sh
+-- context-injection/
|   +-- coach-mode.sh
|   +-- session-context.sh
+-- lib/
|   +-- config.sh
|   +-- hooks-init.sh
|   +-- logging.sh
|   +-- orchestration-audit.sh
|   +-- primitives.sh
|   +-- rite-context-loader.sh
|   +-- session-core.sh
|   +-- session-fsm.sh
|   +-- session-manager.sh
|   +-- session-state.sh
|   +-- session-utils.sh
|   +-- worktree-manager.sh
+-- session-guards/
|   +-- auto-park.sh
|   +-- session-write-guard.sh
|   +-- start-preflight.sh
+-- tracking/
|   +-- artifact-tracker.sh
|   +-- commit-tracker.sh
|   +-- session-audit.sh
+-- validation/
|   +-- delegation-check.sh
|   +-- orchestrator-bypass-check.sh
|   +-- orchestrator-router.sh
+-- base_hooks.yaml
```

### swap-rite.sh Function Map

```
Line    Function                    Cluster
----    --------                    -------
85      log                         Logging
89      log_error                   Logging
93      log_warning                 Logging
97      log_debug                   Logging
109     is_interactive              UI
137     rollback_swap               Transaction
274     handle_interrupt            Transaction
314     handle_exit                 Transaction
321     setup_signal_handlers       Transaction
332     check_journal_recovery      Transaction
409     prompt_recovery_action      Recovery
471     continue_interrupted_swap   Recovery
511     recover_partial_commit      Recovery
536     is_manifest_stale           Manifest
578     complete_partial_commit     Recovery
667     verify_state_consistency    Validation
747     commit_staged_resources     Transaction
809     read_manifest               Manifest
820     get_agent_from_manifest     Manifest
851     write_manifest              Manifest
996     init_manifest_from_existing Manifest
1061    list_incoming_agents        Query
1075    list_current_agents         Query
1086    detect_orphans              Orphan
1141    format_orphan               Orphan
1175    stash_kept_agents           Orphan
1202    restore_kept_agents         Orphan
1227    promote_agents              Orphan
1257    cleanup_stash               Orphan
1263    prompt_disposition          Orphan
1399    usage                       UI
1481    validate_agent_tools        Validation
1528    validate_rite_tools         Validation
1543    validate_rite               Validation
1596    validate_project            Validation
1629    validate_workflow_yaml      Validation
1672    validate_orchestrator_yaml  Validation
1709    validate_rite_schemas       Validation
1733    query_current_rite          Query
1761    list_rites                  Query
1805    backup_current_agents       Swap
1836    swap_agents                 Swap
1906    backup_rite_commands        Swap
1911    remove_rite_commands        Swap
1916    is_rite_command             Query
1921    get_command_rite            Query
1932    check_user_command_collisions Validation
1980    swap_commands               Swap
2048    swap_skills                 Swap
2117    remove_shared_skills        Swap
2143    sync_shared_skills          Swap
2216    swap_hooks                  Swap
2360    cleanup_orphan_backups      Orphan
2422    get_produces_from_workflow  Query
2474    get_workflow_phases         Query
2507    update_claude_md            Core
2664    update_active_rite          Core
2679    preview_refresh             UI
2797    roster_sync_available       Sync
2805    roster_has_updates          Sync
2856    run_roster_sync_waterfall   Sync
2898    update_cem_manifest_rite    Manifest
2955    perform_swap                Core
3332    preview_reset               UI
3412    remove_rite_agents          Swap
3460    regenerate_baseline_claude_md Core
3547    perform_reset               Core
3598    main                        Core
```

---

## Appendix B: Verification Attestation

| File/Directory | Verified Via | Attestation |
|----------------|--------------|-------------|
| `.claude/hooks/` | `du -sh`, `ls` | 348KB, 35 files confirmed |
| `ariadne/.claude/hooks/` | `du -sh`, `ls` | 300KB, 30 files confirmed |
| `ariadne/go.mod` | `Read` tool | `module github.com/autom8y/knossos` confirmed |
| `swap-rite.sh` | `wc -l`, `grep` | 3,773 lines, 68 functions confirmed |
| Root .md files | `wc -c` | 60,730 bytes total confirmed |
| Skills locations | `find`, `ls` | 15 directories confirmed |

---

---

## 8. Alignment Session: Confirmed Requirements

*Captured via structured Q&A on 2026-01-07*

### 8.1 Identity & Namespace (CONFIRMED)

| Decision | Value | Notes |
|----------|-------|-------|
| GitHub Organization | `github.com/autom8y` | Keep existing org |
| Go Module Path | `github.com/autom8y/knossos` | Parent project |
| CLI Binary Name | `ari` | Ariadne metaphor retained |
| Repo Name | `knossos` | Rename AFTER restructure |

**Key Insight**: Knossos is the parent project. Ariadne is the CLI component within it. `.claude/` is the local materialized instance (gitignored, NOT part of repo).

### 8.2 Architecture Decisions (CONFIRMED)

| Decision | Value | Rationale |
|----------|-------|-----------|
| Go Structure | `cmd/ari` + `internal/` | Standard idiomatic Go layout |
| Content Organization | `rites/{name}/skills/`, `rites/{name}/agents/` | Rite-centric, self-contained |
| Content Model | Templates in repo; `.claude/` fully generated | Clean separation |
| Shell Scripts | Minimal wrappers for hooks ONLY | Claude Code hooks require bash; Go handles core logic |
| Hook Bridge | Thin shell wrapper calls `ari subcommand` | e.g., `#!/bin/bash\nari hook clew "$@"` |
| Project Types | Single canonical structure | Consistency over flexibility |

### 8.3 Breaking Changes & Migration (CONFIRMED)

| Decision | Value |
|----------|-------|
| Breaking Change Tolerance | **Full greenfield** - break everything |
| Migration Documentation | Required |
| Backwards Compatibility | None required |
| Rename Timing | Restructure in `roster`, then rename to `knossos` |
| `.claude/` During Restructure | Keep as-is; rematerialize via `sync --force` after finalized |

### 8.4 Success Criteria (CONFIRMED)

| Criterion | Requirement |
|-----------|-------------|
| Done Definition | Published release with docs |
| Must Survive | ALL current functionality (sessions, rites, sync) |
| Descoped | Nothing - ship complete |

### 8.5 Rite Workflow (CONFIRMED)

```
hygiene (SPIKE) → rnd (Research) → ecosystem (Implement) → hygiene (Polish)
```

| Phase | Rite | Output |
|-------|------|--------|
| 1. SPIKE | hygiene | Smell report, architecture proposal (this doc) |
| 2. Research | rnd | Findings docs, ADR drafts (not working code) |
| 3. Implement | ecosystem | Finalized ADRs, working implementation |
| 4. Polish | hygiene | Code smells, dead code removal, final cleanup |

### 8.6 R&D Phase Scope (CONFIRMED)

**Output Type**: Findings + ADR drafts (NOT working code)
**Output Format**: `docs/spikes/SPIKE-*.md`
**Exit Criteria**: All ADRs drafted + research complete

**Prototyping Areas**:
- [ ] Go project restructure + import path migration
- [ ] Sync/materialization UX patterns
- [ ] Hook wrapper architecture validation

**Deep Dive Research**:
- [ ] Go project layout best practices 2025+
- [ ] CLI tool distribution (goreleaser, homebrew, etc.)
- [ ] Configuration materialization/templating engines (text/template, gomplate, etc.)

**ADRs to Draft**:
- [ ] ADR: Go Module Structure (`cmd/ari` + `internal/`)
- [ ] ADR: Content Organization (`rites/{name}/...`)
- [ ] ADR: Hook Architecture (shell wrapper → ari)
- [ ] ADR: Sync/Materialization Model

### 8.7 Sacred Cows (NON-NEGOTIABLE)

1. **Mythology Naming**: Knossos, Ariadne, Moirai, Theseus, etc. - core identity
2. **Session-Based Workflow**: Sessions are the interaction model, not ad-hoc commands
3. **Rite/Pantheon Architecture**: Rite concept with specialist agents stays

### 8.8 User Journey (R&D TO EXPLORE)

How do users consume knossos?
- Clone + `ari init` in target project?
- Install `ari` binary, it fetches templates?
- Other patterns?

**R&D Deliverable**: User journey map with recommended onboarding flow

---

## 9. Updated Target Architecture

Based on alignment session, revised target:

```
github.com/autom8y/knossos/
│
├── cmd/
│   └── ari/
│       └── main.go                 # CLI entry point
│
├── internal/
│   ├── cmd/                        # Cobra commands
│   │   ├── session/
│   │   ├── rite/
│   │   ├── hook/
│   │   ├── sync/
│   │   └── ...
│   ├── session/                    # Session business logic
│   ├── rite/                       # Rite management
│   ├── sync/                       # Materialization engine
│   └── ...
│
├── rites/
│   ├── hygiene/
│   │   ├── skills/
│   │   ├── agents/
│   │   └── manifest.yaml
│   ├── 10x-dev/
│   │   ├── skills/
│   │   ├── agents/
│   │   └── manifest.yaml
│   └── ...
│
├── templates/                      # Source templates for .claude/ generation
│   ├── hooks/
│   │   ├── lib/                    # Shell libraries (minimal, delegates to ari)
│   │   └── *.sh                    # Thin wrappers
│   └── base/
│
├── docs/
│   ├── spikes/
│   ├── decisions/                  # ADRs
│   └── ...
│
├── go.mod                          # module github.com/autom8y/knossos
├── go.sum
├── justfile
└── README.md

# NOT IN REPO (gitignored, generated by `ari sync`):
# .claude/
#   ├── hooks/
#   ├── skills/
#   ├── agents/
#   └── ...
```

### Key Differences from Current State

| Aspect | Current | Target |
|--------|---------|--------|
| Module path | `github.com/autom8y/knossos` | `github.com/autom8y/knossos` |
| Go location | `ariadne/` subdirectory | Root (`cmd/ari`, `internal/`) |
| `.claude/` | In repo, checked in | Gitignored, generated |
| Hook scripts | Full shell implementations | Thin wrappers → `ari` |
| Rite content | Scattered | `rites/{name}/skills/`, `rites/{name}/agents/` |
| Shell scripts | `swap-rite.sh` (3,773 lines) | Decomposed to `internal/rite/` Go |

---

## 10. Phase 2 Handoff Brief (for R&D Rite)

### Context
- **Initiative**: knossos-finalization
- **Session**: session-20260107-164631-8dd6f03a
- **Upstream**: SPIKE complete (this document)
- **Downstream**: ecosystem rite will implement based on R&D findings

### R&D Mission
Research modern patterns and draft ADRs for the knossos consolidation. Do NOT produce working code—that's ecosystem's job.

### Specific Deliverables

1. **SPIKE-go-project-structure.md**
   - Research Go project layout patterns 2025+
   - Evaluate `cmd/` + `internal/` vs alternatives
   - Document import path migration tooling
   - Draft ADR-go-module-structure.md

2. **SPIKE-cli-distribution.md**
   - Research goreleaser, homebrew, other distribution methods
   - Evaluate installation UX patterns
   - Document release automation options
   - Draft ADR-cli-distribution.md

3. **SPIKE-materialization-model.md**
   - Research config templating engines (text/template, gomplate, etc.)
   - Evaluate `ari init` vs `ari sync` UX patterns
   - Map user journey: clone → init → use
   - Draft ADR-sync-materialization.md

4. **SPIKE-hook-architecture.md**
   - Validate thin shell wrapper → `ari` performance
   - Document hook contract (inputs/outputs)
   - Draft ADR-hook-architecture.md

5. **SPIKE-content-organization.md**
   - Validate `rites/{name}/skills/` structure
   - Document rite manifest schema
   - Draft ADR-content-organization.md

### Exit Criteria
- All 5 SPIKE documents complete
- All 5 ADR drafts ready for ecosystem finalization
- No unresolved blocking questions
- Ready for ecosystem handoff

---

**Document Status**: SPIKE COMPLETE + ALIGNMENT CONFIRMED
**Next Step**: Switch to `rnd` rite for Phase 2 research
**Handoff**: This document serves as R&D brief

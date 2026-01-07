# SPIKE: Script Code Smell Refactoring Analysis

**Date**: 2026-01-03
**Author**: hygiene (code-smeller agent)
**Timebox**: 2 hours
**Status**: Complete

---

## Executive Summary

This spike investigated practical refactoring strategies for large roster shell scripts, particularly `swap-rite.sh` at 4,708 LOC.

**Key Finding**: The script significantly exceeds industry best practices (Google recommends <100 LOC for shell scripts), but the complexity is justified by the domain. Modularization following the `lib/sync/` pattern is feasible and recommended.

**Recommendation**: **PROCEED** with incremental refactoring via `/task` workflow. Extract `lib/team/rite-resource.sh` first (highest ROI, lowest risk), then evaluate further extraction based on results.

---

## Research Questions & Answers

### 1. Which duplicate patterns have highest ROI for extraction?

| Pattern | Occurrences | Lines Affected | ROI Score |
|---------|-------------|----------------|-----------|
| Team resource backup | 3 | ~90 | 9.0/10 |
| Team resource removal | 3 | ~66 | 8.5/10 |
| Orphan detection | 3 | ~102 | 8.5/10 |
| Orphan removal | 3 | ~99 | 7.5/10 |
| Team membership check | 6 | ~42 | 7.0/10 |
| Swap resource | 3 | ~150 | 6.5/10 |

**Total DRY violation**: ~549 lines (12% of file) with clear consolidation path.

### 2. What would lib/team/rite-resource.sh look like?

```bash
#!/usr/bin/env bash
# lib/team/rite-resource.sh - Generic team resource operations

# Generic backup for any resource type (commands, skills, hooks)
backup_team_resource() {
    local resource_type="$1"    # commands, skills, hooks
    local resource_dir="$2"     # .claude/commands, etc.
    local marker_file="$3"      # .rite-commands, etc.
    # Unified backup logic (~30 lines vs 90 duplicated)
}

# Generic removal for any resource type
remove_team_resource() {
    local resource_type="$1"
    local resource_dir="$2"
    local marker_file="$3"
    # Unified removal logic (~22 lines vs 66 duplicated)
}

# Check if resource belongs to any rite
is_resource_from_team() {
    local resource_name="$1"
    local resource_type="$2"    # command, skill, hook
    local find_type="$3"        # f (file) or d (directory)
    find "$ROSTER_HOME/teams" -path "*/${resource_type}s/$resource_name" -type "$find_type" 2>/dev/null | grep -q .
}

# Get rite name that owns a resource
get_resource_team() {
    local resource_name="$1"
    local resource_type="$2"
    local match
    match=$(find "$ROSTER_HOME/teams" -path "*/${resource_type}s/$resource_name" -print -quit 2>/dev/null)
    [[ -n "$match" ]] && echo "$match" | sed "s|.*/rites/\([^/]*\)/${resource_type}s/.*|\1|"
}

# Generic orphan detection (returns results via stdout, one per line)
detect_resource_orphans() {
    local resource_type="$1"        # command, skill, hook
    local resource_dir="$2"         # .claude/commands
    local incoming_team="$3"
    local find_type="$4"            # f or d
    # Outputs: "resource_name:owning_team" per line
}

# Generic orphan removal with mode handling
remove_resource_orphans() {
    local resource_type="$1"
    local resource_dir="$2"
    local orphan_mode="$3"          # remove, keep
    local -a orphans=("${@:4}")     # remaining args are orphan list
    # Unified removal with backup logic
}
```

**Estimated module size**: ~150 lines (replacing ~400 lines of duplicates)

### 3. What would lib/team/team-transaction.sh look like?

```bash
#!/usr/bin/env bash
# lib/team/team-transaction.sh - Atomic swap transaction infrastructure

# Module constants (internal)
readonly JOURNAL_VERSION="1"
readonly PHASE_INIT="init"
readonly PHASE_BACKUP="backup"
readonly PHASE_STAGING="staging"
readonly PHASE_COMMIT="commit"
readonly PHASE_COMPLETE="complete"

# Exported functions for swap-rite.sh
create_journal() { ... }
update_journal_phase() { ... }
get_journal_phase() { ... }
delete_journal() { ... }
journal_exists() { ... }

create_staging() { ... }
cleanup_staging() { ... }
stage_agents() { ... }
stage_workflow() { ... }
verify_staging() { ... }

create_swap_backup() { ... }
cleanup_swap_backup() { ... }
verify_backup_integrity() { ... }
```

**Estimated module size**: ~300 lines (clean extraction, no behavior change)

### 4. Can we split without breaking transaction safety?

**Yes**, with careful boundaries:

| Safe to Extract | Must Stay in Main Script |
|-----------------|-------------------------|
| Journal CRUD operations | Signal handler registration |
| Staging utilities | Commit phase orchestration |
| Backup utilities | Error recovery decisions |
| Resource operations | perform_swap() coordinator |

**Key principle**: Extract infrastructure, keep orchestration.

### 5. What shared state would modules need?

| Module | Required Globals | Can Be Parameters |
|--------|-----------------|-------------------|
| rite-resource.sh | `ROSTER_HOME` | All others |
| team-transaction.sh | None | All (paths as params) |

**Bash limitation**: Array passing requires either:
- nameref (bash 4.3+, not on macOS default)
- stdout serialization (portable)
- global arrays (current approach)

**Recommendation**: Use stdout for orphan lists, global arrays only where unavoidable.

### 6. How would testing improve?

| Current State | With Modular Extraction |
|---------------|------------------------|
| No unit tests for swap-rite.sh | Each module testable in isolation |
| Integration testing only | Unit + integration layers |
| Hard to mock team directories | Can mock ROSTER_HOME per test |
| ~4700 LOC to understand | ~150-300 LOC per focused module |

**Test fixtures needed**:
- Mock rite directories
- Mock .claude/ project structure
- Journal file fixtures
- Manifest JSON fixtures

---

## Industry Best Practices Assessment

### Google Shell Style Guide Recommendations

| Guideline | swap-rite.sh Status | Assessment |
|-----------|---------------------|------------|
| <100 LOC for shell scripts | 4,708 LOC | **47x over limit** |
| Rewrite if >100 LOC with complex control flow | Has complex control flow | Rewrite candidate |
| Use for "small utilities or simple wrappers" | Complex orchestration | Beyond intended scope |

### Counter-Arguments for Keeping Shell

1. **Domain expertise**: Team swaps are file operations - shell excels here
2. **Portability**: No Python/Ruby runtime dependencies
3. **Ecosystem consistency**: All roster tooling is shell-based
4. **Transaction model**: Shell's trap handlers are ideal for cleanup
5. **Proven stability**: Script works reliably in production

### Recommended Compromise

**Don't rewrite to another language**. Instead:
1. Extract to focused modules (~300 LOC each)
2. Keep main orchestrator lean (~1500 LOC)
3. Apply shell best practices within each module
4. Add comprehensive unit tests

---

## Effort/Impact Matrix

| Refactoring Option | Effort | Impact | Risk | Recommendation |
|-------------------|--------|--------|------|----------------|
| Extract rite-resource.sh | Low (2-4h) | High (-400 LOC, 6x DRY) | Low | **DO FIRST** |
| Extract team-transaction.sh | Medium (4-6h) | Medium (-300 LOC) | Medium | DO SECOND |
| Extract rite-hooks-registration.sh | Medium (3-4h) | Medium (-200 LOC) | Low | DO THIRD |
| Split perform_swap() into phases | High (6-8h) | Medium (clarity) | Medium | DEFER |
| Rewrite in Python/Go | Very High (40h+) | Variable | High | DON'T DO |

---

## Risk Assessment

### Low Risk
- **rite-resource.sh extraction**: Pure function consolidation, behavior unchanged
- **rite-hooks-registration.sh**: Self-contained, no state coupling

### Medium Risk
- **team-transaction.sh extraction**: Must carefully preserve commit atomicity
- **Array passing in bash**: macOS ships bash 3.2 (no nameref)

### High Risk
- **Language rewrite**: Breaks ecosystem consistency, requires runtime
- **Aggressive decomposition**: May introduce IPC complexity

### Mitigations
1. **Add tests before refactoring** - Capture current behavior
2. **Extract one module at a time** - Validate stability between extractions
3. **Use stdout for array results** - Portable across bash versions

---

## Incremental Refactoring Plan

### Phase 1: Foundation (Week 1)
1. Create `lib/team/` directory structure
2. Add `tests/lib/team/` test fixtures
3. Extract `rite-resource.sh` with unit tests
4. Update `swap-rite.sh` to source and use module

### Phase 2: Transaction Safety (Week 2)
1. Extract `team-transaction.sh`
2. Add transaction unit tests
3. Integration test full swap flow
4. Document module contracts

### Phase 3: Hook Registration (Week 3)
1. Extract `rite-hooks-registration.sh`
2. Add YAML/JSON manipulation tests
3. Validate settings.local.json generation

### Phase 4: Cleanup (Week 4)
1. Review remaining `swap-rite.sh` (~2000 LOC)
2. Evaluate if further extraction warranted
3. Document final architecture

---

## Decision

### Recommendation: **PROCEED with /task**

The analysis confirms:
1. **Clear extraction boundaries exist** - 3 modules identified with clean interfaces
2. **ROI is positive** - ~900 LOC extraction reduces cognitive load
3. **Risk is manageable** - Incremental approach with tests
4. **Industry standards support action** - Current size is 47x over recommended

### Suggested Next Step

```
/task Extract lib/team/rite-resource.sh Module

Create lib/team/rite-resource.sh consolidating:
- backup_team_resource() - generic backup for commands/skills/hooks
- remove_team_resource() - generic removal
- is_resource_from_team() - team membership check
- get_resource_team() - team lookup
- detect_resource_orphans() - orphan detection
- remove_resource_orphans() - orphan removal

Include unit tests in tests/lib/team/.
Update swap-rite.sh to source and use the new module.

Expected outcome: ~400 LOC reduction, 6x DRY improvement.
```

### Alternative: DEFER

If current priorities don't allow refactoring, the script is functional as-is. The debt is technical, not blocking. Revisit when:
- Adding new resource types (would amplify DRY violations)
- Onboarding new contributors (complexity is a barrier)
- Debugging becomes difficult (monolith harder to trace)

---

## Sources

- [Google Shell Style Guide](https://google.github.io/styleguide/shellguide.html) - <100 LOC recommendation
- [Bash Best Practices - cheat-sheets](https://bertvv.github.io/cheat-sheets/Bash.html) - Error handling patterns
- [Shell Script Best Practices - The Sharat's](https://sharats.me/posts/shell-script-best-practices/) - Modular architecture
- [Medium: Best practices for Bash scripting 2025](https://medium.com/@prasanna.a1.usage/best-practices-we-need-to-follow-in-bash-scripting-in-2025-cebcdf254768) - Modern recommendations

---

## Appendix: Code Smell Inventory

| ID | Type | Location | Severity | Lines |
|----|------|----------|----------|-------|
| SM-001 | DRY | backup_team_* (3x) | HIGH | ~90 |
| SM-002 | DRY | remove_team_* (3x) | HIGH | ~66 |
| SM-003 | DRY | detect_*_orphans (3x) | HIGH | ~102 |
| SM-004 | DRY | remove_orphan_* (3x) | MEDIUM | ~99 |
| SM-005 | DRY | is_team_*/get_*_team (6x) | MEDIUM | ~42 |
| SM-006 | DRY | swap_* resources (3x) | MEDIUM | ~150 |
| CH-001 | Complexity | perform_swap() | HIGH | 350 |
| CH-002 | Complexity | update_claude_md() | MEDIUM | 154 |
| CH-003 | Complexity | swap_hooks() | MEDIUM | 136 |
| CH-004 | Complexity | swap_hook_registrations() | LOW | 104 |

**Total duplicated code**: ~549 lines (12%)
**Total complexity hotspots**: ~744 lines (16%)

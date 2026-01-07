# GAP-ANALYSIS: Skeleton Deprecation & CEM Migration

> Sprint 0 Discovery & Audit synthesis for the Skeleton Deprecation initiative.

**Session**: session-20260103-031208-2c671d71
**Initiative**: Skeleton Deprecation & CEM Migration
**Complexity**: SYSTEM (6 sprints, 33 tasks)
**Date**: 2026-01-03
**Sprint**: Sprint 0 - Discovery & Audit (COMPLETE)

---

## Executive Summary

This gap analysis synthesizes findings from Sprint 0's 5 audit tasks to provide a comprehensive view of what's needed to deprecate `skeleton_claude` and achieve roster independence. The analysis reveals:

- **150+ skeleton references** across 95 roster files requiring updates
- **~4,500 lines of CEM code** (main + libraries) needing migration or replacement
- **58 resources** to migrate (11 skills, 7 user-agents, 38 user-commands, 2 user-skills)
- **3 critical integration points** with hard CEM dependencies
- **23 identified risks** (5 HIGH severity requiring immediate mitigation)

**Conclusion**: Migration is feasible but complex. Estimated 6 sprints with careful ordering to avoid breaking existing functionality.

---

## Source Artifacts

| Task | Artifact | Location |
|------|----------|----------|
| task-001 | Skeleton References Audit | `/Users/tomtenuta/Code/roster/docs/audits/skeleton-references-audit.md` |
| task-002 | CEM Functionality Analysis | `/Users/tomtenuta/Code/roster/docs/analysis/CEM-functionality-analysis.md` |
| task-003 | Skeleton Resources Inventory | `/Users/tomtenuta/Code/roster/docs/audits/skeleton-resources-inventory.md` |
| task-004 | CEM Integration Points | `/Users/tomtenuta/Code/roster/docs/analysis/CEM-integration-points.md` |
| task-005 | Migration Risk Assessment | `/Users/tomtenuta/Code/roster/docs/assessments/skeleton-migration-risks.md` |

---

## Gap 1: Skeleton Reference Dependencies

### Current State

Roster contains **150+ references** to skeleton across **95 files**:

| Category | Count | Complexity |
|----------|-------|------------|
| Environment Variables | 15 | Medium |
| Path References (hardcoded) | 45 | High |
| Documentation References | 60 | Low |
| Code Dependencies | 12 | High |
| Configuration References | 8 | Medium |
| Conceptual References | 30+ | Low/None |

### Critical Path References

| File | Issue | Impact |
|------|-------|--------|
| `user-hooks/lib/worktree-manager.sh:225-251` | Hard dependency on `$SKELETON_HOME/cem` | **Worktree creation fails completely** |
| `user-hooks/lib/config.sh:17` | Default `SKELETON_HOME` path | All hooks assume skeleton exists |
| `user-commands/cem/sync.md` | 8 hardcoded CEM command paths | User sync workflow broken |
| `user-skills/orchestration/orchestrator-templates/*` | 100+ `/skeleton_claude/` paths | Tutorial documentation unusable |

### Target State

- Zero hard dependencies on `$SKELETON_HOME/cem`
- All CEM functionality available via roster-native scripts
- Documentation updated to reference roster paths only
- Environment variables deprecated or made optional

### Gap Resolution

1. **Bundle CEM into roster** or create roster-native sync mechanism
2. **Update config.sh** to not require `SKELETON_HOME`
3. **Global find-replace** for path references (`skeleton_claude` → `roster`)
4. **Update orchestrator-templates** skill documentation

---

## Gap 2: CEM Functionality

### Current State

CEM provides **8 commands** with sophisticated sync logic:

| Command | Lines | Complexity | Roster Equivalent |
|---------|-------|------------|-------------------|
| `init` | ~150 | Medium | None |
| `sync` | ~400 | **High** | None |
| `validate` | ~100 | Low | None |
| `validate-team` | ~80 | Low | Partial (swap-rite.sh) |
| `repair` | ~80 | Medium | None |
| `install-user` | ~40 | Low | sync-user-*.sh (4 scripts) |
| `status` | ~60 | Low | None |
| `diff` | ~50 | Low | None |

### Critical CEM Features

| Feature | Location | Complexity | Risk if Missed |
|---------|----------|------------|----------------|
| Three-way checksum classification | `cem-sync.sh` | **High** | Data loss on conflicts |
| CLAUDE.md section merge | `merge-docs.sh` | **High** | Breaks CLAUDE.md regeneration |
| Settings JSON union merge | `merge-settings.sh` | Medium | Loses project settings |
| Orphan detection & prune | `cem-sync.sh` | Medium | Stale files accumulate |
| Manifest versioning (v1/v2) | `cem-manifest.sh` | Medium | Migration failures |
| Checksum caching | `cem-checksum.sh` | Low | Performance degradation |

### Target State

- All sync logic available in roster-native script
- Manifest compatibility with existing satellites
- Same merge strategies for settings and docs
- Orphan management preserved

### Gap Resolution

**Option A: Port CEM to Roster** (Recommended)
- Copy CEM main + lib/ to roster
- Update paths from `$SKELETON_HOME` to `$ROSTER_HOME`
- Integrate with swap-rite.sh
- Estimated: 500 lines of adaptation

**Option B: Inline Critical Functions**
- Extract only sync + merge logic
- Embed directly in swap-rite.sh
- Drop unused commands (repair, diff, alias)
- Estimated: 300 lines of new code

**Option C: Complete Rewrite**
- New roster-native sync mechanism
- May not maintain manifest compatibility
- Estimated: 800+ lines

---

## Gap 3: Resource Migration

### Current State

Skeleton has resources roster lacks:

| Resource Type | Skeleton | Roster | Gap | Priority |
|---------------|----------|--------|-----|----------|
| Skills | 18 | 25 | 11 unique to skeleton | **HIGH** |
| User-Agents | 7 | 0 | 7 (all) | **HIGH** |
| User-Commands | 38 | 0 | 38 (all) | **MEDIUM** |
| User-Skills | 2 | 0 | 2 (all) | MEDIUM |
| Hooks | 11+10 lib | 12+13 lib | 4 unique to skeleton | LOW |
| Schemas | 1 | 0 | 1 | LOW |

### Critical Skills to Migrate

| Skill | Purpose | Impact if Missing |
|-------|---------|-------------------|
| `state-mate` | Centralized state mutation | Session management broken |
| `commit-ref` | AI-assisted commits | Commit workflow broken |
| `task-ref` | Full lifecycle task execution | /task command broken |
| `sprint-ref` | Multi-task sprint orchestration | /sprint command broken |
| `pr-ref` | PR creation workflow | /pr command broken |
| `qa-ref` | QA validation workflow | /qa command broken |
| `worktree-ref` | Git worktree management | /worktree command broken |

### Target State

- All 11 skeleton skills migrated to roster
- 7 user-agents available (Forge team)
- 38 user-commands functional (converted to skills or preserved)
- 2 user-skills migrated

### Gap Resolution

1. **Skills**: Copy 11 skeleton skills to `roster/.claude/skills/` or `rites/shared/skills/`
2. **User-Agents**: Migrate to `roster/user-agents/` (Forge team agents)
3. **User-Commands**: Migrate to `roster/user-commands/` or convert to skills
4. **Evaluate**: team-validator.sh, workflow-validator.sh, workflow.schema.json

---

## Gap 4: Integration Points

### Current State

3 direct CEM integration points:

| Integration Point | CEM Commands | Impact if Broken |
|-------------------|--------------|------------------|
| `worktree-manager.sh` | sync, init, init --force | Parallel sessions fail |
| `/sync` command | All 8 commands | User sync workflow broken |
| `/cem-debug` command | Source reference | Diagnostic tooling broken |

### Indirect References

| Reference | Purpose | Update Needed |
|-----------|---------|---------------|
| `context-injection.sh` | CEM sync status | Path update |
| `ecosystem-ref` skill | CEM documentation | Content update |
| `swap-rite.sh` | Skeleton baseline concept | Minimal |

### Target State

- Worktree creation works without skeleton
- /sync uses roster-native sync
- Diagnostic tools point to roster internals

### Gap Resolution

1. **worktree-manager.sh**: Source CEM as library (short-term) → Roster-native (long-term)
2. **/sync command**: Thin wrapper to roster sync → Unified /roster command
3. **/cem-debug**: Update references to roster internals

---

## Gap 5: Risk Mitigation

### HIGH Severity Risks (5)

| Risk ID | Description | Mitigation |
|---------|-------------|------------|
| RISK-BC-001 | Worktree creation breaks | Port CEM to roster before removing skeleton |
| RISK-FN-001 | CLAUDE.md merge complexity | Port merge-docs.sh as-is, test extensively |
| RISK-FN-002 | Conflict detection errors | Port checksum logic exactly, add test coverage |
| RISK-FN-005 | Missing skills break workflows | Migrate skills in Sprint 3 before command migration |
| RISK-EX-001 | Migration ordering dependencies | Follow dependency graph, validate each step |

### Recommended Ordering

```
Sprint 1: CEM Architecture Design
    └── TDD for roster-native sync
    └── Test infrastructure

Sprint 2: Core CEM Migration
    └── Port cem main + lib/ to roster
    └── Update paths
    └── Validate manifest compatibility

Sprint 3: Resource Migration
    └── Migrate 11 skills
    └── Migrate 7 user-agents
    └── Migrate critical user-commands

Sprint 4: Reference Cleanup
    └── Update 150+ skeleton references
    └── Test worktree creation
    └── Test /sync command

Sprint 5: Documentation & Testing
    └── Update all documentation
    └── Migration guide for satellites
    └── Full test pass

Sprint 6: Deprecation & Archive
    └── Archive skeleton_claude
    └── Announce deprecation
    └── Support period
```

---

## Success Metrics

### Sprint 0 Completion (This Document)

- [x] All skeleton references cataloged (95 files, 150+ refs)
- [x] CEM functionality fully documented (8 commands, 6 merge strategies)
- [x] All skeleton resources cataloged (58 items to migrate)
- [x] CEM integration points mapped (3 direct, 3 indirect)
- [x] Risk assessment complete (23 risks, 5 HIGH)

### Migration Success Criteria

- [ ] Zero `$SKELETON_HOME` dependencies in roster
- [ ] Worktree creation works standalone
- [ ] /sync uses roster-native mechanism
- [ ] All 11 skeleton skills migrated
- [ ] Manifest compatibility maintained
- [ ] Documentation updated (150+ references)
- [ ] Migration guide published
- [ ] skeleton_claude archived

---

## Recommendations

### Immediate Actions (Sprint 1)

1. **Create TDD for CEM replacement** - Route to architect agent
2. **Set up test infrastructure** - Validate manifest compatibility
3. **Identify rollback points** - Each sprint must be reversible

### Architecture Decision Required

**Question**: Port CEM vs. Rewrite?

| Option | Effort | Risk | Compatibility |
|--------|--------|------|---------------|
| Port CEM | Medium | Low | Full |
| Inline Critical | Medium | Medium | Partial |
| Rewrite | High | High | Must verify |

**Recommendation**: Port CEM (Option A) to minimize risk and ensure manifest compatibility.

### Dependencies

| Dependency | Status | Notes |
|------------|--------|-------|
| Shared Skills Architecture | **COMPLETE** | session-20260103-031131-181cfc38 |
| swap-rite.sh stability | Ready | Recent optimizations complete |
| Test infrastructure | Needed | Must create before Sprint 2 |

---

## Attestation

| Task | Artifact | Status |
|------|----------|--------|
| task-001 | skeleton-references-audit.md | Verified |
| task-002 | CEM-functionality-analysis.md | Verified |
| task-003 | skeleton-resources-inventory.md | Verified |
| task-004 | CEM-integration-points.md | Verified |
| task-005 | skeleton-migration-risks.md | Verified |
| Synthesis | GAP-ANALYSIS-skeleton-deprecation.md | **This Document** |

---

## Next Steps

Sprint 0 is **COMPLETE**. Proceed to Sprint 1: CEM Migration Planning.

**Entry Point**: Route to `architect` agent for TDD-cem-replacement.md creation.

**Prompt Pattern**:
```
Create TDD for CEM replacement architecture based on GAP-ANALYSIS-skeleton-deprecation.md findings.
Design roster-native sync mechanism that maintains manifest compatibility with existing satellites.
```

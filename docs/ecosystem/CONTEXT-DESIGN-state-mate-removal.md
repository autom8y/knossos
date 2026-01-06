# Context Design: State-Mate Alias Removal

| Field | Value |
|-------|-------|
| **Track** | 6 |
| **Type** | PATCH |
| **Status** | Ready for Implementation |
| **Architect** | Context Architect |
| **Date** | 2026-01-06 |

---

## 1. Executive Summary

Consolidate terminology from the legacy `state-mate` alias to canonical `Moirai` terminology. The Knossos Doctrine (Section II: The Fates) defines the Moirai as the three Fates: Clotho (spinner), Lachesis (allotter), and Atropos (inflexible). The codebase has evolved to support Moirai but retains `state-mate` as a backward-compatibility alias that perpetuates terminology confusion.

**Recommendation**: Implement a phased deprecation with alias retention, followed by alias removal in a future release.

---

## 2. Audit Results

### 2.1 State-Mate References Summary

| Category | Count | Files |
|----------|-------|-------|
| Direct `state-mate` mentions | ~528 | 66 files |
| `Task(moirai, ...)` invocations | 61 | 18 files |
| Agent alias definition | 1 | user-agents/moirai.md |
| Bypass audit log | 1 | .claude/audit/state-mate-bypass.jsonl |
| Deprecated hook | 1 | .claude/hooks/.deprecated/session-guards/session-write-guard.sh |

### 2.2 Key Files with State-Mate References

| File | Impact Level | Reference Type |
|------|--------------|----------------|
| `/Users/tomtenuta/Code/roster/user-agents/moirai.md` | HIGH | Alias definition |
| `/Users/tomtenuta/Code/roster/.claude/CLAUDE.md` | HIGH | State management section note |
| `/Users/tomtenuta/Code/roster/docs/decisions/ADR-0005-state-mate-centralized-state-authority.md` | HIGH | Entire ADR |
| `/Users/tomtenuta/Code/roster/user-skills/session-lifecycle/shared-sections/moirai-invocation.md` | MEDIUM | Backward compatibility note |
| `/Users/tomtenuta/Code/roster/docs/philosophy/knossos-doctrine.md` | MEDIUM | Concordance section |
| `/Users/tomtenuta/Code/roster/.claude/audit/state-mate-bypass.jsonl` | LOW | Audit log filename |
| `/Users/tomtenuta/Code/roster/docs/design/TDD-*.md` (multiple) | LOW | Design docs referencing legacy pattern |
| `/Users/tomtenuta/Code/roster/docs/requirements/PRD-*.md` (multiple) | LOW | PRDs referencing legacy pattern |

### 2.3 No Standalone State-Mate Agent File

Confirmed: There is no `/Users/tomtenuta/Code/roster/user-agents/state-mate.md` file. The alias exists only in `moirai.md`:

```yaml
# From user-agents/moirai.md lines 10-12
aliases:
  - state-mate
  - fates
```

---

## 3. Moirai Verification

### 3.1 Three Fates Status: COMPLETE

All three Fates are implemented as separate agents:

| Fate | File | Domain | Status |
|------|------|--------|--------|
| **Clotho** | `/Users/tomtenuta/Code/roster/user-agents/clotho.md` | Creation (create_sprint, start_sprint) | ACTIVE |
| **Lachesis** | `/Users/tomtenuta/Code/roster/user-agents/lachesis.md` | Measurement (mark_complete, park_session, etc.) | ACTIVE |
| **Atropos** | `/Users/tomtenuta/Code/roster/user-agents/atropos.md` | Termination (wrap_session, generate_sails) | ACTIVE |
| **Moirai** (Router) | `/Users/tomtenuta/Code/roster/user-agents/moirai.md` | Routes to appropriate Fate | ACTIVE |

### 3.2 Shared Infrastructure

| Component | Path | Purpose |
|-----------|------|---------|
| Shared operations | `/Users/tomtenuta/Code/roster/user-agents/moirai-shared.md` | Schema locations, lock protocol, audit format |
| Invocation pattern | `/Users/tomtenuta/Code/roster/user-skills/session-lifecycle/shared-sections/moirai-invocation.md` | How to invoke Moirai |

### 3.3 Moirai Router Functionality

The Moirai router (`moirai.md`) correctly:
- Parses operations (structured and natural language)
- Routes to appropriate Fate based on operation type
- Maintains backward compatibility with `state-mate` alias
- Returns Fate responses unchanged

---

## 4. External Consumer Assessment

### 4.1 Satellite Impact Analysis

| Consumer Type | Impact | Notes |
|---------------|--------|-------|
| Roster repository internal | HIGH | 66 files reference state-mate |
| Skeleton satellites | NONE | Skeleton deprecated; roster is canonical |
| User documentation | MEDIUM | Users may have learned state-mate invocation |
| Claude Code agent definitions | LOW | Alias in moirai.md handles routing |

### 4.2 Task Tool Invocation Patterns

Current valid invocations (both work identically):
```
Task(moirai, "park_session reason='break'")
Task(moirai, "park_session reason='break'")
```

The alias mechanism in Claude Code handles this at the agent resolution level, meaning no hooks or infrastructure changes are needed for the alias to continue working.

### 4.3 Recommendation: Phased Deprecation

**Phase 1 (This Track)**: Documentation-only changes
- Update documentation to use `moirai` as primary
- Add deprecation notices for `state-mate`
- Preserve alias for backward compatibility

**Phase 2 (Future)**: Alias removal
- Remove `state-mate` alias from moirai.md
- Requires coordination with any external satellites

---

## 5. Removal Strategy

### 5.1 Selected Approach: Soft Deprecation with Alias Retention

**Rationale**: Hard removal would break existing invocation patterns and any external consumers that learned the `state-mate` pattern. The alias mechanism is effectively free (no runtime cost) and provides a graceful migration path.

### 5.2 What Changes

| Component | Change Type | Action |
|-----------|-------------|--------|
| `user-agents/moirai.md` | PRESERVE | Keep `state-mate` alias for backward compatibility |
| `.claude/CLAUDE.md` | EDIT | Update state management section to remove state-mate note |
| `docs/decisions/ADR-0005-*.md` | RENAME | Rename to `ADR-0005-moirai-centralized-state-authority.md` |
| `user-skills/.../moirai-invocation.md` | EDIT | Add deprecation notice for state-mate |
| `docs/philosophy/knossos-doctrine.md` | VERIFY | Already has concordance (state-mate -> Moirai) |
| `.claude/audit/state-mate-bypass.jsonl` | RENAME | Rename to `moirai-bypass.jsonl` |
| `.claude/hooks/lib/fail-open.sh` | EDIT | Update bypass log filename |
| Deprecated hooks | DELETE | Remove `.claude/hooks/.deprecated/session-guards/session-write-guard.sh` |

### 5.3 What Does NOT Change

| Component | Reason |
|-----------|--------|
| `moirai.md` alias list | Backward compatibility |
| `clotho.md`, `lachesis.md`, `atropos.md` | Already use correct terminology |
| Hook enforcement logic | Uses agent name detection, not alias |

---

## 6. Documentation Updates Needed

### 6.1 High-Priority Updates

| File | Change |
|------|--------|
| `/Users/tomtenuta/Code/roster/.claude/CLAUDE.md` line 214 | Remove: `> **Note**: state-mate is an alias for moirai for backward compatibility.` |
| `/Users/tomtenuta/Code/roster/docs/decisions/ADR-0005-state-mate-centralized-state-authority.md` | Rename file and update title to "Moirai Centralized State Authority" |
| `/Users/tomtenuta/Code/roster/user-skills/session-lifecycle/shared-sections/moirai-invocation.md` lines 193-201 | Update backward compatibility section with deprecation notice |

### 6.2 Mass Find/Replace in Documentation

| Pattern | Replacement | Scope |
|---------|-------------|-------|
| `state-mate` (agent reference) | `Moirai` | docs/ directory |
| `Task(moirai,` | `Task(moirai,` | docs/ directory examples |
| `ADR-0005-state-mate-centralized-state-authority` | `ADR-0005-moirai-centralized-state-authority` | All references |

**Note**: Preserve historical accuracy in changelog entries and commit messages.

---

## 7. Backward Compatibility Classification

### Classification: COMPATIBLE

The change is backward compatible because:
1. The `state-mate` alias remains in `moirai.md` aliases list
2. Existing `Task(moirai, ...)` invocations continue to work
3. No hook changes required (alias resolution happens at agent level)
4. Only documentation and naming changes

### No Migration Required for Consumers

Existing code using `Task(moirai, ...)` will continue to work. New documentation will prefer `Task(moirai, ...)` but the alias remains available.

---

## 8. File-Level Changes Specification

### 8.1 Files to EDIT

| File | Line(s) | Change |
|------|---------|--------|
| `/Users/tomtenuta/Code/roster/.claude/CLAUDE.md` | 214 | Remove state-mate alias note |
| `/Users/tomtenuta/Code/roster/user-skills/session-lifecycle/shared-sections/moirai-invocation.md` | 191-201 | Update backward compatibility section to mark state-mate as deprecated |
| `/Users/tomtenuta/Code/roster/.claude/hooks/lib/fail-open.sh` | 191 | Change `state-mate-bypass.jsonl` to `moirai-bypass.jsonl` |

### 8.2 Files to RENAME

| Current Path | New Path |
|--------------|----------|
| `/Users/tomtenuta/Code/roster/docs/decisions/ADR-0005-state-mate-centralized-state-authority.md` | `/Users/tomtenuta/Code/roster/docs/decisions/ADR-0005-moirai-centralized-state-authority.md` |
| `/Users/tomtenuta/Code/roster/.claude/audit/state-mate-bypass.jsonl` | `/Users/tomtenuta/Code/roster/.claude/audit/moirai-bypass.jsonl` |

### 8.3 Files to DELETE

| Path | Reason |
|------|--------|
| `/Users/tomtenuta/Code/roster/.claude/hooks/.deprecated/session-guards/session-write-guard.sh` | Already deprecated; contains outdated state-mate references |

### 8.4 Files to UPDATE Content (Bulk)

All files in `docs/` containing `Task(moirai,` examples should be updated to use `Task(moirai,`. This affects 18 files with 61 occurrences. Examples should be updated but inline text references may remain for historical context.

---

## 9. Integration Test Matrix

| Satellite Type | Test | Expected Outcome |
|----------------|------|------------------|
| Minimal (no session) | Invoke moirai | Agent resolves correctly |
| Minimal (no session) | Invoke state-mate | Agent resolves (alias works) |
| Standard (with session) | `Task(moirai, "park_session ...")` | Session parks successfully |
| Standard (with session) | `Task(moirai, "park_session ...")` | Session parks successfully (alias) |
| Complex (orchestrated) | Hook blocks direct write, suggests Moirai | Correct terminology in error message |

---

## 10. Quality Gate Criteria

### 10.1 Pre-Merge Checklist

- [ ] ADR-0005 renamed and content updated
- [ ] CLAUDE.md state management section updated
- [ ] moirai-invocation.md deprecation notice added
- [ ] fail-open.sh bypass log filename updated
- [ ] Deprecated hook deleted
- [ ] At least 10 documentation examples updated from state-mate to moirai
- [ ] All references to ADR-0005 path updated

### 10.2 Verification Commands

```bash
# Count remaining state-mate references (should decrease significantly)
grep -r "state-mate" --include="*.md" docs/ | wc -l

# Verify alias still works (should return moirai agent)
grep -A5 "aliases:" user-agents/moirai.md

# Verify ADR rename
ls docs/decisions/ADR-0005-*

# Verify bypass log rename
ls .claude/audit/moirai-bypass.jsonl
```

---

## 11. Anti-Patterns to Avoid

### 11.1 Do NOT Remove the Alias

The alias in `moirai.md` MUST remain:
```yaml
aliases:
  - state-mate
  - fates
```

Removing this would break backward compatibility.

### 11.2 Do NOT Update Historical Records

Commit messages, changelogs, and historical documentation that mention `state-mate` should remain unchanged. Only update:
- Active documentation (guides, tutorials)
- Code examples that users will copy
- File names for active files

### 11.3 Do NOT Rename the Audit Log Retroactively

When renaming `.claude/audit/state-mate-bypass.jsonl` to `moirai-bypass.jsonl`:
- Create new file with new name
- Existing entries in old file are historical (can be archived or merged)
- New entries go to new filename

---

## 12. Implementation Order

1. **Rename ADR-0005** - Foundation document
2. **Update CLAUDE.md** - User-facing documentation
3. **Update moirai-invocation.md** - Skill documentation
4. **Rename audit log** - Infrastructure
5. **Update fail-open.sh** - Hook code
6. **Delete deprecated hook** - Cleanup
7. **Bulk update documentation examples** - Comprehensive pass
8. **Update cross-references to ADR-0005** - Link maintenance

---

## 13. Decision Rationale Summary

| Decision | Rationale |
|----------|-----------|
| Preserve alias | Zero-cost backward compatibility; no breaking changes |
| Rename ADR-0005 | Primary architecture document should use canonical terminology |
| Deprecation notice | Guides new users to preferred terminology |
| Bulk doc updates | Reduces confusion for users learning the system |
| Delete deprecated hook | Reduces maintenance burden; outdated |

---

## 14. Cross-References

- `/Users/tomtenuta/Code/roster/docs/philosophy/knossos-doctrine.md` - Section II (The Fates) and XII (Concordance)
- `/Users/tomtenuta/Code/roster/docs/plans/REFACTOR-PLAN-doctrine-purity.md` - T13 task (state-mate -> moirai)
- `/Users/tomtenuta/Code/roster/docs/ecosystem/GAP-ANALYSIS-team-to-rite-migration.md` - Related terminology migration

---

*Context Design complete. Ready for Integration Engineer implementation.*

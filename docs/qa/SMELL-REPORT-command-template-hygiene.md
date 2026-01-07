# Smell Report: Command Template Standardization

**Generated**: 2025-12-29
**Auditor**: Code Smeller (hygiene)
**Scope**: All command files in `user-commands/` and `rites/*/commands/`
**Reference Standard**: `docs/ecosystem/CONTEXT-DESIGN-command-argument-standardization.md`

---

## Summary

| Smell | Count | Severity | Fix Complexity |
|-------|-------|----------|----------------|
| S1: Missing argument-hint | 3 | HIGH | LOW |
| S2: Missing $ARGUMENTS | 1 | MEDIUM | LOW |
| S3: Section Order Violations | 7 | LOW | LOW |
| S4: Non-standard Flag Format | 11 | MEDIUM | MEDIUM |
| S5: Deprecated Flag References | 0 | N/A | N/A |
| S6: Missing Frontmatter Fields | 6 | HIGH | LOW |
| S7: Pass-through Without $ARGUMENTS | 0 | N/A | N/A |

**Total Files Audited**: 41
**Files with Smells**: 23
**Clean Files**: 18

---

## Detailed Findings

### S1: Missing argument-hint (3 files) - HIGH

Commands that accept arguments but lack `argument-hint:` in frontmatter.

| File | Accepts Args? | Evidence | Notes |
|------|---------------|----------|-------|
| `user-commands/meta/minus-1.md` | Yes (`{TAG}`) | No frontmatter at all | Legacy format, no YAML frontmatter |
| `user-commands/meta/zero.md` | Yes (`{TAG}`) | No frontmatter at all | Legacy format, no YAML frontmatter |
| `user-commands/meta/one.md` | Yes (session context) | No frontmatter at all | Legacy format, no YAML frontmatter |

**Evidence**: All three meta commands use `{TAG}` or implicit session context but have no YAML frontmatter block at all. They start directly with markdown headers.

```markdown
# File: user-commands/meta/minus-1.md (line 1)
# Session -1: Initiative Assessment
```

### S2: Missing $ARGUMENTS (1 file) - MEDIUM

Commands with argument-hint but missing `$ARGUMENTS` in "Your Task" section.

| File | Has argument-hint | Has $ARGUMENTS | Notes |
|------|-------------------|----------------|-------|
| `user-commands/navigation/team.md` | Yes | Yes | Clean |
| `rites/ecosystem/commands/cem-debug.md` | No argument-hint | N/A | Falls under S6, not S2 |

**Analysis**: After reviewing all files, most commands correctly include `$ARGUMENTS`. One edge case:

| File | Issue | Evidence |
|------|-------|----------|
| `user-commands/team-switching/forge.md` | Non-standard $ARGUMENTS placement | `**Arguments**: $ARGUMENTS` instead of inline |

```markdown
# File: user-commands/team-switching/forge.md (lines 10-12)
## Your Task

Display information about The Forge - the meta-team for creating and maintaining agent teams.

**Arguments**: $ARGUMENTS
```

The canonical format places `$ARGUMENTS` inline after the task description sentence, not as a separate labeled field.

### S3: Section Order Violations (7 files) - LOW

Canonical order: Context -> Pre-flight -> Task -> Behavior -> Flags -> Examples -> Reference

| File | Actual Order | Violation |
|------|--------------|-----------|
| `user-commands/navigation/team.md` | Context -> Task -> Behavior -> Orphan Handling -> Flags inline -> Agent Provenance -> Examples -> Reference | "Orphan Agent Handling" and "Agent Provenance" break flow |
| `user-commands/navigation/sessions.md` | Context -> Task -> Behavior (with inline Examples) | Examples embedded in Behavior |
| `user-commands/navigation/worktree.md` | Context -> Pre-flight -> Task -> Commands -> Examples -> Typical Workflow -> Reference | "Commands" instead of "Behavior", "Typical Workflow" section added |
| `user-commands/session/start.md` | Pre-computed Context -> Task -> Behavior -> Complexity Levels -> Example Usage -> Reference | "Pre-computed Context" variant, "Complexity Levels" extra section |
| `user-commands/workflow/sprint.md` | Context -> Pre-flight -> Task -> Behavior -> Example -> When to Use -> Parallel Sprint Pattern -> Reference | "When to Use" after Example, extra "Parallel Sprint Pattern" section |
| `rites/doc-team-pack/commands/consolidate.md` | Context -> Task -> Parameters -> Workflow Phases -> Behavior -> Examples -> Phase Transitions -> Resumption -> Error Handling -> Reference | Multiple extra sections interspersed |
| `rites/ecosystem/commands/cem-debug.md` | Context -> Task -> Behavior -> When to Use -> CEM Diagnostic Checklist -> Expected Output -> Handoff -> Reference | Multiple extra sections |

**Note**: Section order violations are LOW severity as they don't break functionality, only consistency.

### S4: Non-standard Flag Format (11 files) - MEDIUM

Flags documented as bullet list instead of table format with "Handled By" column.

| File | Current Format | Evidence |
|------|----------------|----------|
| `user-commands/team-switching/10x.md` | Bullet list in Behavior | `**Flags:**\n- \`--update\`, \`-u\`: Pull latest...` |
| `user-commands/team-switching/hygiene.md` | Bullet list in Behavior | Same pattern |
| `user-commands/team-switching/rnd.md` | Bullet list in Behavior | Same pattern |
| `user-commands/team-switching/debt.md` | Bullet list in Behavior | Same pattern |
| `user-commands/team-switching/intelligence.md` | Bullet list in Behavior | Same pattern |
| `user-commands/team-switching/strategy.md` | Bullet list in Behavior | Same pattern |
| `user-commands/team-switching/docs.md` | Bullet list in Behavior | Same pattern |
| `user-commands/team-switching/sre.md` | Bullet list in Behavior | Same pattern |
| `user-commands/team-switching/security.md` | Bullet list in Behavior | Same pattern |
| `user-commands/navigation/ecosystem.md` | Bullet list in Behavior | Same pattern |
| `user-commands/navigation/team.md` | Mixed: inline bullets + some tables | Partial compliance |

**Canonical Format** (from standard):
```markdown
## Flags

| Flag | Short | Description | Handled By |
|------|-------|-------------|------------|
| `--update` | `-u` | Pull latest definitions from roster | swap-rite.sh |
| `--dry-run` | - | Preview changes without applying | swap-rite.sh |
```

**Actual Format** (in team-switching commands):
```markdown
**Flags:**
- `--update`, `-u`: Pull latest agent definitions from roster even if already on team
- `--dry-run`: Preview changes without applying
```

Missing: Dedicated `## Flags` section, table format, "Handled By" column.

### S5: Deprecated Flag References (0 files) - N/A

No commands still reference `--refresh` or `--force` flags.

| Search | Results |
|--------|---------|
| `--refresh` in user-commands | 0 matches (correctly updated to `--update`) |
| `--force` in user-commands | Only in `sync.md` which correctly uses CEM's `--force` |

**Note**: The `user-commands/cem/sync.md` uses `--force` for CEM's force flag which is distinct from swap-rite.sh's deprecated `--force`. This is correct per the standard which notes CEM's `--refresh` is unaffected.

### S6: Missing Frontmatter Fields (6 files) - HIGH

| File | Missing Fields | Evidence |
|------|----------------|----------|
| `user-commands/meta/minus-1.md` | ALL (no frontmatter) | File starts with `# Session -1:` |
| `user-commands/meta/zero.md` | ALL (no frontmatter) | File starts with `# Session 0:` |
| `user-commands/meta/one.md` | ALL (no frontmatter) | File starts with `# Session 1:` |
| `user-commands/navigation/team.md` | `allowed-tools` | Uses Bash for swap-rite.sh but `allowed-tools` not declared |
| `user-commands/cem/sync.md` | `model` | Has description, argument-hint, allowed-tools but no model |
| `rites/ecosystem/commands/cem-debug.md` | `argument-hint` | Has description, allowed-tools, model but no argument-hint despite accepting implicit arguments |

**Required Frontmatter** (per standard):
```yaml
---
description: <required>
argument-hint: <required if accepts arguments>
model: <required>
allowed-tools: <required if uses Bash>
---
```

### S7: Pass-through Without $ARGUMENTS (0 files) - N/A

All team-switching commands correctly pass `$ARGUMENTS` in Behavior execution.

**Evidence from 10x.md** (representative):
```markdown
## Behavior

1. Execute: `${ROSTER_HOME:-~/Code/roster}/swap-rite.sh 10x-dev $ARGUMENTS`
```

All 10 team-switching commands follow this pattern correctly.

---

## Priority Matrix

| File | Smells | Fix Complexity | Priority |
|------|--------|----------------|----------|
| `user-commands/meta/minus-1.md` | S1, S6 | LOW | 1 (HIGH) |
| `user-commands/meta/zero.md` | S1, S6 | LOW | 1 (HIGH) |
| `user-commands/meta/one.md` | S1, S6 | LOW | 1 (HIGH) |
| `user-commands/cem/sync.md` | S6 | LOW | 2 (HIGH) |
| `user-commands/navigation/team.md` | S3, S4, S6 | MEDIUM | 3 (MEDIUM) |
| `user-commands/team-switching/10x.md` | S4 | LOW | 4 (MEDIUM) |
| `user-commands/team-switching/hygiene.md` | S4 | LOW | 4 (MEDIUM) |
| `user-commands/team-switching/rnd.md` | S4 | LOW | 4 (MEDIUM) |
| `user-commands/team-switching/debt.md` | S4 | LOW | 4 (MEDIUM) |
| `user-commands/team-switching/intelligence.md` | S4 | LOW | 4 (MEDIUM) |
| `user-commands/team-switching/strategy.md` | S4 | LOW | 4 (MEDIUM) |
| `user-commands/team-switching/docs.md` | S4 | LOW | 4 (MEDIUM) |
| `user-commands/team-switching/sre.md` | S4 | LOW | 4 (MEDIUM) |
| `user-commands/team-switching/security.md` | S4 | LOW | 4 (MEDIUM) |
| `user-commands/navigation/ecosystem.md` | S4 | LOW | 4 (MEDIUM) |
| `user-commands/team-switching/forge.md` | S2 | LOW | 5 (LOW) |
| `user-commands/navigation/worktree.md` | S3 | LOW | 6 (LOW) |
| `user-commands/navigation/sessions.md` | S3 | LOW | 6 (LOW) |
| `user-commands/session/start.md` | S3 | LOW | 6 (LOW) |
| `user-commands/workflow/sprint.md` | S3 | LOW | 6 (LOW) |
| `rites/doc-team-pack/commands/consolidate.md` | S3 | LOW | 6 (LOW) |
| `rites/ecosystem/commands/cem-debug.md` | S3, S6 | LOW | 4 (MEDIUM) |

---

## Clean Files (18)

These files fully comply with the canonical template:

### user-commands/session/
- `continue.md`
- `handoff.md`
- `park.md`
- `wrap.md`

### user-commands/workflow/
- `task.md`
- `hotfix.md`

### user-commands/operations/
- `architect.md`
- `build.md`
- `qa.md`
- `code-review.md`
- `commit.md`

### user-commands/navigation/
- `consult.md`

### rites/forge/commands/
- `new-team.md`
- `validate-team.md`
- `eval-agent.md`

### rites/10x-dev/commands/
- `spike.md`
- `pr.md`

---

## Recommendations

### Priority 1: Add Frontmatter to Meta Commands (3 files)

The meta commands (`minus-1.md`, `zero.md`, `one.md`) need complete frontmatter added. These are HIGH priority because they completely lack required fields.

**Action**: Add standard frontmatter to each:
```yaml
---
description: <appropriate description>
argument-hint: <initiative>
model: opus
---
```

### Priority 2: Add Missing Frontmatter Fields (3 files)

- `sync.md`: Add `model: sonnet` (or appropriate model)
- `team.md`: Add `allowed-tools: Bash, Read`
- `cem-debug.md`: Add `argument-hint: [issue-description]`

### Priority 3: Standardize Flag Format (11 files)

Convert bullet-list flags to table format in all team-switching commands and `ecosystem.md`.

**Template to apply**:
```markdown
## Flags

| Flag | Short | Description | Handled By |
|------|-------|-------------|------------|
| `--update` | `-u` | Pull latest definitions from roster | swap-rite.sh |
| `--dry-run` | - | Preview changes without applying | swap-rite.sh |
| `--keep-all` | - | Keep all orphan agents in project | swap-rite.sh |
| `--remove-all` | - | Remove all orphan agents | swap-rite.sh |
| `--promote-all` | - | Move orphan agents to user-level | swap-rite.sh |
```

### Priority 4: Fix $ARGUMENTS Placement (1 file)

`forge.md`: Move `$ARGUMENTS` inline per standard:
```markdown
## Your Task

Display information about The Forge - the meta-team for creating and maintaining agent teams. $ARGUMENTS
```

### Priority 5: Section Order (7 files)

LOW priority. Consider standardizing section order in future maintenance pass. Extra sections like "When to Use", "Complexity Levels", and "Parallel Sprint Pattern" provide value but break consistency.

**Options**:
1. Rename to fit canonical sections (e.g., "When to Use" -> fold into Context or Examples)
2. Accept as legitimate extensions and document as allowed variants
3. Leave as-is (current recommendation given low impact)

---

## Handoff Criteria Met

- [x] All 41 command files inventoried and analyzed
- [x] Smell counts accurate with evidence
- [x] Priority matrix complete with fix complexity
- [x] Recommendations actionable with templates
- [x] Boundary concerns identified (meta commands need architectural decision on frontmatter format)

---

## Notes for Architect Enforcer

1. **Meta Commands Architecture**: The `meta/` commands (`minus-1.md`, `zero.md`, `one.md`) follow a fundamentally different pattern - they're more like prompt templates than standard commands. Consider whether they should:
   - Be converted to standard command format with frontmatter
   - Be moved to a separate `templates/` directory
   - Have a documented variant pattern in the standard

2. **Flag Section Placement**: The canonical standard specifies `## Flags` as a separate section, but 10 team-switching commands embed flags in `## Behavior`. This is a systematic pattern that suggests either:
   - The standard needs updating to allow inline flag documentation for simple pass-through commands
   - All 10 files need updating to use separate `## Flags` section

3. **Extra Sections**: Several commands add useful sections not in the canonical order (e.g., "When to Use", "Parallel Sprint Pattern"). Consider documenting these as approved extensions rather than violations.

---

*Report generated by Code Smeller. Ready for Architect Enforcer review.*

# Refactoring Plan: Knossos Terminology Finalization

**Based on**: Audit Lead findings VG-002, VG-003 from Terminology Finalization Sprint
**Prepared**: 2026-01-07
**Scope**: Remediate 32 files with incorrect terminology (team-pack/rite-pack, agent rites)

## Architectural Assessment

### Boundary Health

- **Schema Files (JSON)**: Clean boundaries, isolated changes per schema
- **User Skills**: Moderate coupling - skills reference each other but terminology is local
- **User Commands**: Clean boundaries - command files are self-contained
- **Rite Source Files**: Clean boundaries - rites are isolated bundles
- **Root Files**: Mixed - README.md is high-visibility, others are scripts
- **Templates**: Low coupling - templates generate content, changes propagate on regeneration

### Root Causes Identified

1. **RC-001 (Historical Migration)**: The project transitioned from "team pack" to "rite" terminology. First-pass remediation caught obvious cases but missed compound terms and descriptions.
   - Explains: "rite pack" remnants in descriptions, "team pack" in older docs

2. **RC-002 (Inconsistent Pattern)**: "agent rite(s)" was used as shorthand for "collection of agents within a rite" before "pantheon" was established as canonical.
   - Explains: "agent rites" in README, skill descriptions, forge commands

### Canonical Terminology Contract

| Concept | Correct Term | Incorrect Terms |
|---------|--------------|-----------------|
| Practice bundle (directory in rites/) | **rite** | team pack, team-pack, rite pack, rite-pack |
| Agent collection within a rite | **pantheon** | agent rite, agent rites |

## Source Scope

Files that MAY be edited:
- `user-skills/`, `user-commands/`, `user-agents/`, `user-hooks/`
- `rites/` (excluding any `.claude/` subdirs)
- `knossos/templates/`, `lib/`, `docs/`, `schemas/`
- `ariadne/internal/` (Go code and JSON schemas)
- Root scripts and files

Files that MUST NOT be edited:
- Any `.claude/` directory (materialized content)

---

## Refactoring Sequence

### Phase 1: Schema Foundation [Low Risk]

**Goal**: Fix terminology in JSON schemas - these are validation contracts

#### RF-001: Fix manifest.schema.json descriptions

**Smells addressed**: VG-002 (rite-pack)
**Category**: Local (JSON schema)
**File**: `/Users/tomtenuta/Code/roster/ariadne/internal/manifest/schemas/manifest.schema.json`

**Before State**:
```json
"default": { "description": "Default rite pack name" }
"available": { "description": "List of available rite pack names" }
"discovery": { "description": "Paths to search for rite packs" }
```

**After State**:
```json
"default": { "description": "Default rite name" }
"available": { "description": "List of available rite names" }
"discovery": { "description": "Paths to search for rites" }
```

**Invariants**:
- JSON remains valid
- Schema validation behavior unchanged
- No property names modified (only description strings)

**Verification**:
```bash
jq '.' ariadne/internal/manifest/schemas/manifest.schema.json > /dev/null && echo "Valid JSON"
grep -c "rite pack" ariadne/internal/manifest/schemas/manifest.schema.json  # Should be 0
```

**Commit scope**: Single file

---

#### RF-002: Fix agent-manifest.schema.json descriptions

**Smells addressed**: VG-002 (rite-pack)
**Category**: Local (JSON schema)
**File**: `/Users/tomtenuta/Code/roster/ariadne/internal/manifest/schemas/agent-manifest.schema.json`

**Before State**:
```json
"active_rite": { "description": "Currently active rite pack" }
```

**After State**:
```json
"active_rite": { "description": "Currently active rite" }
```

**Invariants**:
- JSON remains valid
- Property name `active_rite` unchanged

**Verification**:
```bash
jq '.' ariadne/internal/manifest/schemas/agent-manifest.schema.json > /dev/null && echo "Valid JSON"
grep -c "rite pack" ariadne/internal/manifest/schemas/agent-manifest.schema.json  # Should be 0
```

**Commit scope**: Single file

---

#### RF-003: Fix session-context.schema.json descriptions

**Smells addressed**: VG-002 (rite-pack, team-pack)
**Category**: Local (JSON schema)
**File**: `/Users/tomtenuta/Code/roster/ariadne/internal/validation/schemas/session-context.schema.json`

**Before State**:
```json
"active_rite": { "description": "Active rite pack name" }
"team": { "description": "Team pack name (null for cross-cutting sessions)" }
```

**After State**:
```json
"active_rite": { "description": "Active rite name" }
"team": { "description": "Rite name (null for cross-cutting sessions)" }
```

**Invariants**:
- JSON remains valid
- Property names unchanged

**Verification**:
```bash
jq '.' ariadne/internal/validation/schemas/session-context.schema.json > /dev/null && echo "Valid JSON"
grep -cE "(rite pack|team pack)" ariadne/internal/validation/schemas/session-context.schema.json  # Should be 0
```

**Commit scope**: Single file

---

#### RF-004: Fix common.schema.json (ariadne) description

**Smells addressed**: VG-002 (rite-pack)
**Category**: Local (JSON schema)
**File**: `/Users/tomtenuta/Code/roster/ariadne/internal/validation/schemas/common.schema.json`

**Before State**:
```json
"rite_name": { "description": "Rite pack name (e.g., ecosystem, security)" }
```

**After State**:
```json
"rite_name": { "description": "Rite name (e.g., ecosystem, security)" }
```

**Invariants**:
- JSON remains valid
- Pattern `^[a-z0-9]+-pack$` unchanged (note: pattern needs separate review if "-pack" suffix is obsolete)

**Verification**:
```bash
jq '.' ariadne/internal/validation/schemas/common.schema.json > /dev/null && echo "Valid JSON"
grep -c "rite pack" ariadne/internal/validation/schemas/common.schema.json  # Should be 0
```

**Commit scope**: Single file

---

#### RF-005: Fix orchestrator.yaml.schema.json

**Smells addressed**: VG-002 (rite-pack, team-pack)
**Category**: Local (JSON schema)
**File**: `/Users/tomtenuta/Code/roster/schemas/orchestrator.yaml.schema.json`

**Before State**:
```json
"description": "Canonical schema for orchestrator.yaml configuration files that drive orchestrator.md generation. Each rite pack uses one orchestrator.yaml..."
"name": { "description": "Rite pack name (lowercase with hyphens, must match directory name)" }
"name": "doc-team-pack"  // in examples section
```

**After State**:
```json
"description": "Canonical schema for orchestrator.yaml configuration files that drive orchestrator.md generation. Each rite uses one orchestrator.yaml..."
"name": { "description": "Rite name (lowercase with hyphens, must match directory name)" }
"name": "docs"  // in examples section
```

**Invariants**:
- JSON remains valid
- Schema structure unchanged
- Examples still valid

**Verification**:
```bash
jq '.' schemas/orchestrator.yaml.schema.json > /dev/null && echo "Valid JSON"
grep -cE "(rite pack|team pack|team-pack|rite-pack)" schemas/orchestrator.yaml.schema.json  # Should be 0
```

**Commit scope**: Single file

---

#### RF-006: Fix common.schema.json (schemas/artifacts)

**Smells addressed**: VG-002 (rite-pack)
**Category**: Local (JSON schema)
**File**: `/Users/tomtenuta/Code/roster/schemas/artifacts/common.schema.json`

**Before State**:
```json
"rite_name": { "description": "Rite pack name (e.g., ecosystem, security)" }
```

**After State**:
```json
"rite_name": { "description": "Rite name (e.g., ecosystem, security)" }
```

**Invariants**:
- JSON remains valid
- Pattern unchanged

**Verification**:
```bash
jq '.' schemas/artifacts/common.schema.json > /dev/null && echo "Valid JSON"
grep -c "rite pack" schemas/artifacts/common.schema.json  # Should be 0
```

**Commit scope**: Single file

---

**[ROLLBACK POINT: Phase 1 Complete]**
All schema files remediated. Can stop here safely - schemas are self-contained.

---

### Phase 2: Root File Cleanup [Medium Risk]

**Goal**: Fix high-visibility root files

#### RF-007: Fix README.md header and references

**Smells addressed**: VG-003 (agent rites)
**Category**: Root file, high visibility
**File**: `/Users/tomtenuta/Code/roster/README.md`

**Before State**:
```markdown
# Roster - Agent Rite Management
```

**After State**:
```markdown
# Roster - Rite Management
```

**Invariants**:
- Markdown valid
- Links preserved

**Verification**:
```bash
grep -c "Agent Rite" README.md  # Should be 0
grep -c "agent rite" README.md  # Should be 0
```

**Commit scope**: Single file

---

#### RF-008: Fix swap-rite.sh header comments

**Smells addressed**: VG-003 (agent rites)
**Category**: Root script
**File**: `/Users/tomtenuta/Code/roster/swap-rite.sh`

**Before State**:
```bash
# swap-rite.sh - Agent Rite Pack Management System
# Swaps Claude Code agent rites (pantheons) with atomic-ish operations.
```

**After State**:
```bash
# swap-rite.sh - Rite Management System
# Swaps Claude Code rites (agent pantheons) with atomic-ish operations.
```

**Invariants**:
- Script functionality unchanged
- Only comments modified

**Verification**:
```bash
bash -n swap-rite.sh && echo "Valid bash"
grep -cE "agent rite" swap-rite.sh  # Should be 0
```

**Commit scope**: Single file

---

#### RF-009: Fix skills/rite/skill.md

**Smells addressed**: VG-003 (agent rites)
**Category**: Root skill file
**File**: `/Users/tomtenuta/Code/roster/skills/rite/skill.md`

**Before State**:
```markdown
Swap Claude Code agent rites for different workflows.
```

**After State**:
```markdown
Swap Claude Code rites for different workflows.
```

**Invariants**:
- Skill invocation unchanged

**Verification**:
```bash
grep -c "agent rite" skills/rite/skill.md  # Should be 0
```

**Commit scope**: Single file

---

#### RF-010: Fix RITE_SKILL_MATRIX.md

**Smells addressed**: VG-002 (team-pack), VG-003 (agent rites)
**Category**: Root documentation
**File**: `/Users/tomtenuta/Code/roster/RITE_SKILL_MATRIX.md`

**Before State**:
```markdown
Creating and managing agent rites.
| forge-ref, team-development | ecosystem | Team pack creation |
```

**After State**:
```markdown
Creating and managing rites.
| forge-ref, rite-development | ecosystem | Rite creation |
```

**Invariants**:
- Table structure preserved

**Verification**:
```bash
grep -cE "(agent rite|team pack|Team pack)" RITE_SKILL_MATRIX.md  # Should be 0
```

**Commit scope**: Single file

---

**[ROLLBACK POINT: Phase 2 Complete]**
Root files remediated. These are high-visibility, recommend verification before proceeding.

---

### Phase 3: User Skills Cleanup [Medium Risk]

**Goal**: Fix terminology in user skill documentation

#### RF-011: Fix user-skills/guidance/rite-ref/SKILL.md

**Smells addressed**: VG-002 (rite-pack), VG-003 (agent rites)
**Category**: User skill
**File**: `/Users/tomtenuta/Code/roster/user-skills/guidance/rite-ref/SKILL.md`

**Before State**:
```markdown
description: "Switch agent rites or list available rites..."
# /rite - Agent Rite Switcher
...supports `--rite=PACK` parameter
```

**After State**:
```markdown
description: "Switch rites or list available rites..."
# /rite - Rite Switcher
...supports `--rite=NAME` parameter
```

**Invariants**:
- Skill invocation unchanged
- All behavioral documentation preserved

**Verification**:
```bash
grep -cE "(agent rite|rite=PACK)" user-skills/guidance/rite-ref/SKILL.md  # Should be 0
```

**Commit scope**: Single file

---

#### RF-012: Fix worktree-ref skill files (3 files)

**Smells addressed**: VG-002 (rite-pack)
**Category**: User skill (multi-file)
**Files**:
- `/Users/tomtenuta/Code/roster/user-skills/operations/worktree-ref/SKILL.md`
- `/Users/tomtenuta/Code/roster/user-skills/operations/worktree-ref/behavior.md`
- `/Users/tomtenuta/Code/roster/user-skills/operations/worktree-ref/integration.md`

**Pattern**: `--rite=PACK` -> `--rite=NAME`; `Rite Pack Integration` -> `Rite Integration`

**Before State** (across files):
```markdown
| `--rite=PACK` | Rite pack to use (default: current) |
## Rite Pack Integration
```

**After State**:
```markdown
| `--rite=NAME` | Rite to use (default: current) |
## Rite Integration
```

**Invariants**:
- Argument name `--rite` preserved
- Only documentation text changes

**Verification**:
```bash
grep -rE "(rite=PACK|Rite Pack)" user-skills/operations/worktree-ref/  # Should be empty
```

**Commit scope**: Three files (single logical change)

---

#### RF-013: Fix start-ref skill files (2 files)

**Smells addressed**: VG-002 (rite-pack)
**Category**: User skill
**Files**:
- `/Users/tomtenuta/Code/roster/user-skills/session-lifecycle/start-ref/SKILL.md`
- `/Users/tomtenuta/Code/roster/user-skills/session-lifecycle/start-ref/behavior.md`

**Before State**:
```markdown
/start [initiative-name] [--complexity=LEVEL] [--rite=PACK] [--no-rite]
| `--rite` | No | ACTIVE_RITE | Rite pack for session |
| `active_team` | Current or specified team | Team pack for this session |
```

**After State**:
```markdown
/start [initiative-name] [--complexity=LEVEL] [--rite=NAME] [--no-rite]
| `--rite` | No | ACTIVE_RITE | Rite for session |
| `active_rite` | Current or specified rite | Rite for this session |
```

**Invariants**:
- Command syntax preserved
- Argument names unchanged

**Verification**:
```bash
grep -rE "(rite=PACK|Rite pack|Team pack)" user-skills/session-lifecycle/start-ref/  # Should be empty
```

**Commit scope**: Two files

---

#### RF-014: Fix session-common skill files (4 files)

**Smells addressed**: VG-002 (team-pack)
**Category**: User skill
**Files**:
- `/Users/tomtenuta/Code/roster/user-skills/session-lifecycle/session-common/complexity-levels.md`
- `/Users/tomtenuta/Code/roster/user-skills/session-lifecycle/session-common/session-context-schema.md`
- `/Users/tomtenuta/Code/roster/user-skills/session-common/session-context-schema.md`
- `/Users/tomtenuta/Code/roster/user-skills/session-common/sprint-context-schema.md`

**Pattern**: "Team pack" -> "Rite"; "team pack" -> "rite"

**Before State** (samples):
```markdown
## Complexity vs. Team Pack
**Note**: Team pack doesn't dictate complexity
active_team: string       # Team pack name
active_team: string         # Team pack name (copied from SESSION_CONTEXT)
```

**After State**:
```markdown
## Complexity vs. Rite
**Note**: Rite doesn't dictate complexity
active_rite: string       # Rite name
active_rite: string         # Rite name (copied from SESSION_CONTEXT)
```

**Invariants**:
- Schema field names should align with actual schema (`active_rite` vs `active_team`)
- NOTE: Field name change from `active_team` to `active_rite` is a terminology alignment, not a behavior change

**Verification**:
```bash
grep -rE "(Team pack|team pack)" user-skills/session-lifecycle/session-common/  # Should be empty
grep -rE "(Team pack|team pack)" user-skills/session-common/  # Should be empty
```

**Commit scope**: Four files

---

#### RF-015: Fix resume validation-checks.md

**Smells addressed**: VG-002 (team-pack)
**Category**: User skill
**File**: `/Users/tomtenuta/Code/roster/user-skills/session-lifecycle/resume/validation-checks.md`

**Before State**:
```markdown
Team packs contain different agents. If session started with `10x-dev` but current rite is `docs`, expected agents may not be available.
```

**After State**:
```markdown
Rites contain different pantheons. If session started with `10x-dev` but current rite is `docs`, expected agents may not be available.
```

**Invariants**:
- Validation logic description preserved

**Verification**:
```bash
grep -c "Team pack" user-skills/session-lifecycle/resume/validation-checks.md  # Should be 0
```

**Commit scope**: Single file

---

**[ROLLBACK POINT: Phase 3 Complete]**
User skills remediated. Session and worktree commands will use updated terminology.

---

### Phase 4: User Commands Cleanup [Low Risk]

**Goal**: Fix terminology in user command files

#### RF-016: Fix navigation/rite.md

**Smells addressed**: VG-003 (agent rites)
**Category**: User command
**File**: `/Users/tomtenuta/Code/roster/user-commands/navigation/rite.md`

**Before State**:
```markdown
description: Switch agent rites or list available rites
Manage agent rites. $ARGUMENTS
```

**After State**:
```markdown
description: Switch rites or list available rites
Manage rites. $ARGUMENTS
```

**Invariants**:
- Command invocation unchanged

**Verification**:
```bash
grep -c "agent rite" user-commands/navigation/rite.md  # Should be 0
```

**Commit scope**: Single file

---

#### RF-017: Fix rite-switching commands (4 files)

**Smells addressed**: VG-002 (team-pack)
**Category**: User commands
**Files**:
- `/Users/tomtenuta/Code/roster/user-commands/rite-switching/forge.md`
- `/Users/tomtenuta/Code/roster/user-commands/rite-switching/intelligence.md`
- `/Users/tomtenuta/Code/roster/user-commands/rite-switching/security.md`
- `/Users/tomtenuta/Code/roster/user-commands/rite-switching/strategy.md`

**Before State** (samples):
```markdown
Display information about The Forge - the meta-rite for creating and maintaining agent rites.
Switch to the Security Team pack and display the rite roster.
Switch to the Strategy Team pack and display the rite roster.
Switch to the Product Intelligence Team pack and display the rite roster.
```

**After State**:
```markdown
Display information about The Forge - the meta-rite for creating and maintaining rites.
Switch to the Security rite and display the pantheon.
Switch to the Strategy rite and display the pantheon.
Switch to the Product Intelligence rite and display the pantheon.
```

**Invariants**:
- Command functionality unchanged

**Verification**:
```bash
grep -rE "(agent rite|Team pack|Team Pack)" user-commands/rite-switching/  # Should be empty
```

**Commit scope**: Four files

---

**[ROLLBACK POINT: Phase 4 Complete]**
User commands remediated.

---

### Phase 5: Rite Source Files Cleanup [Medium Risk]

**Goal**: Fix terminology in rite documentation and skills

#### RF-018: Fix docs rite files (4 files)

**Smells addressed**: VG-002 (team-pack)
**Category**: Rite source
**Files**:
- `/Users/tomtenuta/Code/roster/rites/docs/README.md`
- `/Users/tomtenuta/Code/roster/rites/docs/workflow.md`
- `/Users/tomtenuta/Code/roster/rites/docs/TODO.md`
- `/Users/tomtenuta/Code/roster/rites/docs/AUDIT-doc-team-pack-agents.md`

**Before State** (samples):
```markdown
# Doc Team Pack
# Doc Team Pack Workflow
# Doc-Team-Pack Agent Audit Report
```

**After State**:
```markdown
# Docs Rite
# Docs Rite Workflow
# Docs Rite Agent Audit Report
```

**Invariants**:
- File content describes same functionality

**Verification**:
```bash
grep -rE "(Doc Team Pack|doc-team-pack|Team Pack)" rites/docs/  # Should be empty
```

**Commit scope**: Four files

---

#### RF-019: Fix 10x-dev rite TODO.md

**Smells addressed**: VG-002 (team-pack)
**Category**: Rite source
**File**: `/Users/tomtenuta/Code/roster/rites/10x-dev/TODO.md`

**Before State**:
```markdown
- To Doc Team Pack: feature summary, API changes, user-facing behavior changes
```

**After State**:
```markdown
- To Docs rite: feature summary, API changes, user-facing behavior changes
```

**Invariants**:
- TODO item meaning preserved

**Verification**:
```bash
grep -c "Team Pack" rites/10x-dev/TODO.md  # Should be 0
```

**Commit scope**: Single file

---

#### RF-020: Fix forge rite files (2 files)

**Smells addressed**: VG-003 (agent rites)
**Category**: Rite source
**Files**:
- `/Users/tomtenuta/Code/roster/rites/forge/skills/rite-development/SKILL.md`
- `/Users/tomtenuta/Code/roster/rites/forge/commands/new-rite.md`

**Before State**:
```markdown
description: "Design and implement agent rites for the roster ecosystem..."
description: Create a new agent rite through The Forge workflow
```

**After State**:
```markdown
description: "Design and implement rites for the roster ecosystem..."
description: Create a new rite through The Forge workflow
```

**Invariants**:
- Skill/command functionality unchanged

**Verification**:
```bash
grep -rE "agent rite" rites/forge/  # Should be empty
```

**Commit scope**: Two files

---

#### RF-021: Fix other rite ref skills (4 files)

**Smells addressed**: VG-002 (team-pack)
**Category**: Rite source
**Files**:
- `/Users/tomtenuta/Code/roster/rites/strategy/skills/strategy-ref/skill.md`
- `/Users/tomtenuta/Code/roster/rites/intelligence/skills/intelligence-ref/skill.md`
- `/Users/tomtenuta/Code/roster/rites/rnd/skills/rnd-ref/skill.md`
- `/Users/tomtenuta/Code/roster/rites/security/skills/security-ref/skill.md`

**Pattern**: "Team Packs:" -> "Related Rites:"

**Before State** (sample):
```markdown
- **Team Packs**: intelligence (product analytics), rnd (technology strategy)
- **Team Packs**: 10x-dev (implementation), strategy (strategic context)
```

**After State**:
```markdown
- **Related Rites**: intelligence (product analytics), rnd (technology strategy)
- **Related Rites**: 10x-dev (implementation), strategy (strategic context)
```

**Invariants**:
- Cross-references preserved

**Verification**:
```bash
grep -rE "Team Packs:" rites/strategy/ rites/intelligence/ rites/rnd/ rites/security/  # Should be empty
```

**Commit scope**: Four files

---

**[ROLLBACK POINT: Phase 5 Complete]**
Rite source files remediated.

---

### Phase 6: Templates [Low Risk]

**Goal**: Fix terminology in templates (affects regenerated content)

#### RF-022: Fix quick-start.md.tpl

**Smells addressed**: VG-003 (agent rite)
**Category**: Template
**File**: `/Users/tomtenuta/Code/roster/knossos/templates/sections/quick-start.md.tpl`

**Before State**:
```
This project uses a {{ .AgentCount }}-agent rite ({{ .ActiveRite }}):
```

**After State**:
```
This project uses a {{ .AgentCount }}-agent pantheon ({{ .ActiveRite }} rite):
```

**Invariants**:
- Template variables preserved
- Regenerated CLAUDE.md will have updated terminology

**Verification**:
```bash
grep -c "agent rite" knossos/templates/sections/quick-start.md.tpl  # Should be 0
```

**Commit scope**: Single file

---

**[PHASE 6 COMPLETE - ALL REMEDIATION DONE]**

---

## Risk Matrix

| Refactor | Risk | Blast Radius | Rollback Cost | Dependencies |
|----------|------|--------------|---------------|--------------|
| RF-001 | Low | 1 file | Trivial | None |
| RF-002 | Low | 1 file | Trivial | None |
| RF-003 | Low | 1 file | Trivial | None |
| RF-004 | Low | 1 file | Trivial | None |
| RF-005 | Low | 1 file | Trivial | None |
| RF-006 | Low | 1 file | Trivial | None |
| RF-007 | Med | 1 file (high visibility) | Trivial | None |
| RF-008 | Low | 1 file | Trivial | None |
| RF-009 | Low | 1 file | Trivial | None |
| RF-010 | Low | 1 file | Trivial | None |
| RF-011 | Med | 1 file | Trivial | None |
| RF-012 | Med | 3 files | 1 commit | None |
| RF-013 | Med | 2 files | 1 commit | None |
| RF-014 | Med | 4 files | 1 commit | RF-003 (schema alignment) |
| RF-015 | Low | 1 file | Trivial | None |
| RF-016 | Low | 1 file | Trivial | None |
| RF-017 | Low | 4 files | 1 commit | None |
| RF-018 | Med | 4 files | 1 commit | None |
| RF-019 | Low | 1 file | Trivial | None |
| RF-020 | Low | 2 files | 1 commit | None |
| RF-021 | Low | 4 files | 1 commit | None |
| RF-022 | Low | 1 file (template) | Trivial | None |

---

## Verification Command Suite

### Phase-Level Verification

```bash
# Phase 1: Schemas
for f in ariadne/internal/manifest/schemas/*.json ariadne/internal/validation/schemas/*.json schemas/*.json schemas/artifacts/*.json; do
  jq '.' "$f" > /dev/null || echo "FAIL: $f invalid JSON"
done
grep -rE "(rite pack|team pack)" ariadne/internal/manifest/schemas/ ariadne/internal/validation/schemas/ schemas/ && echo "FAIL: terminology found" || echo "PASS: Phase 1"

# Phase 2: Root files
grep -E "(agent rite|Agent Rite|team pack|Team pack)" README.md swap-rite.sh skills/rite/skill.md RITE_SKILL_MATRIX.md && echo "FAIL" || echo "PASS: Phase 2"

# Phase 3: User skills
grep -rE "(agent rite|rite=PACK|Rite Pack|Team pack|team pack)" user-skills/ && echo "FAIL" || echo "PASS: Phase 3"

# Phase 4: User commands
grep -rE "(agent rite|Team pack|Team Pack)" user-commands/ && echo "FAIL" || echo "PASS: Phase 4"

# Phase 5: Rite sources
grep -rE "(agent rite|Team Pack|team-pack|rite-pack)" rites/ --include="*.md" && echo "FAIL" || echo "PASS: Phase 5"

# Phase 6: Templates
grep -c "agent rite" knossos/templates/sections/quick-start.md.tpl && echo "FAIL" || echo "PASS: Phase 6"
```

### Full Audit Rerun

After all phases complete:
```bash
# VG-002: No "team-pack" or "rite-pack" in SOURCE files
grep -rE "(team.?pack|rite.?pack)" \
  user-skills/ user-commands/ user-agents/ user-hooks/ \
  rites/ knossos/templates/ lib/ docs/ schemas/ \
  ariadne/internal/ \
  README.md swap-rite.sh skills/ RITE_SKILL_MATRIX.md \
  --include="*.md" --include="*.json" --include="*.sh" \
  | grep -v ".claude/" && echo "VG-002 FAIL" || echo "VG-002 PASS"

# VG-003: No "agent rite(s)" in SOURCE files
grep -rE "agent rites?" \
  user-skills/ user-commands/ user-agents/ user-hooks/ \
  rites/ knossos/templates/ lib/ docs/ schemas/ \
  ariadne/internal/ \
  README.md swap-rite.sh skills/ RITE_SKILL_MATRIX.md \
  --include="*.md" --include="*.json" --include="*.sh" \
  | grep -v ".claude/" && echo "VG-003 FAIL" || echo "VG-003 PASS"
```

---

## Notes for Janitor

### Commit Message Convention

```
fix(terminology): [component] remediate team-pack/agent-rite terms

- Replace "rite pack" with "rite"
- Replace "team pack" with "rite"
- Replace "agent rite(s)" with "rite" or "pantheon"

Part of: Knossos Terminology Finalization Sprint
Audit findings: VG-002, VG-003
```

### Test Run Requirements

- After each phase: run verification command for that phase
- After all phases: run full audit rerun commands
- No unit tests expected to break (terminology in docs/descriptions only)

### Files to Avoid Touching

- Any `.claude/` directory (materialized content - never edit)
- Files in `docs/reports/` (historical audit reports)
- Files in `docs/qa/` (historical QA artifacts)
- `docs/philosophy/knossos-doctrine.md` (contains intentional terminology mapping table)

### Order Critical For

- RF-014 should follow RF-003 (schema field alignment)
- Phase 5 (rite sources) before Phase 6 (templates) - ensures regenerated content picks up changes

### Parallelization Opportunities

- RF-001 through RF-006 (Phase 1) can run in parallel
- RF-007 through RF-010 (Phase 2) can run in parallel
- RF-016 and RF-017 (Phase 4) can run in parallel
- RF-018 through RF-021 (Phase 5) can run in parallel

---

## Out of Scope

### Deferred Findings

1. **Schema pattern `^[a-z0-9]+-pack$`**: The rite_name pattern in schemas still requires "-pack" suffix. This is a schema behavior change, not terminology cleanup. Requires separate ADR.

2. **Generated content in .claude/**: Will be updated when `ari inscription sync` runs. Not manually edited.

3. **Historical documents** (docs/reports/, docs/qa/): These are point-in-time audit records. Editing would falsify historical record.

4. **docs/philosophy/knossos-doctrine.md**: Contains intentional mapping table showing old -> new terminology. This is documentation OF the migration, not content needing migration.

---

## Attestation

| File | Verified via Read | Terminology Violations Found |
|------|-------------------|------------------------------|
| manifest.schema.json | Yes | "rite pack" x3 |
| agent-manifest.schema.json | Yes | "rite pack" x1 |
| session-context.schema.json | Yes | "rite pack" x1, "team pack" x1 |
| common.schema.json (ariadne) | Yes | "rite pack" x1 |
| orchestrator.yaml.schema.json | Yes | "rite pack" x2, "doc-team-pack" x1 |
| common.schema.json (schemas/artifacts) | Yes | "rite pack" x1 |
| README.md | Yes | "Agent Rite" x1 |
| swap-rite.sh | Yes | "Agent Rite" x2 |
| user-skills/guidance/rite-ref/SKILL.md | Yes | "agent rites" in description |
| knossos/templates/sections/quick-start.md.tpl | Yes | "agent rite" x1 |

---

**Plan Prepared By**: Architect Enforcer
**Ready for Handoff To**: Janitor
**Handoff Criteria Met**:
- [x] Every smell classified (addressed in RF-001 through RF-022)
- [x] Each refactoring has before/after contract documented
- [x] Invariants and verification criteria specified
- [x] Refactorings sequenced with explicit dependencies
- [x] Rollback points identified between phases (5 rollback points)
- [x] Risk assessment complete for each phase
- [x] Artifacts verified via Read tool with attestation table

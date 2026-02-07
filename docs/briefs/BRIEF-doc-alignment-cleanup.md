# Content Brief: Documentation Alignment & Cleanup

| Field | Value |
|-------|-------|
| **Brief ID** | BRIEF-doc-alignment-cleanup |
| **Date** | 2026-02-07 |
| **Priority** | P0 (Must fix for dogfooding readiness) + P1 (Should fix for good first impression) |
| **Estimated Effort** | 2-3 hours |
| **Assigned To** | Tech Writer |
| **Reviewed By** | Information Architect |
| **Status** | Ready for execution |

## Executive Summary

This brief addresses 25 documentation defects identified in a comprehensive 5-agent audit. The work focuses on P0 (blocking dogfooding) and P1 (first impression) items only. P2 tech debt items are explicitly out of scope.

**Key issues:**
1. README.md contains ~60% stale content (old project name, deleted directories, wrong script names)
2. 24 mena files missing required `name` frontmatter field
3. MEMORY.md metrics out of date
4. 1 agent file has YAML syntax error
5. User hooks manifest still references "roster" branding

## Scope

### In Scope (P0 + P1)
- README.md full update (P0)
- Add `name` frontmatter to 24 mena files (P0)
- MEMORY.md metrics update (P1)
- context-engineer.md YAML fix (P1)
- Document user hooks manifest fix procedure (P1)

### Out of Scope (P2)
- ADR updates for pre-unification terminology
- Test backfill for `internal/sync` package
- Test coverage for `Materialize()` entry point
- Any code changes beyond frontmatter additions

## File-by-File Change List

### 1. README.md (P0)

**File:** `/Users/tomtenuta/Code/knossos/README.md`

**Changes:**

#### Line 1: Update title
```diff
- # Roster - Rite Management
+ # Knossos - Context Engineering Meta-Framework
```

#### Lines 5-12: Remove stale user-level sync table
**DELETE entire table** (lines 5-12). Replace with:

```markdown
### User-Level Sync (knossos -> ~/.claude/)

| Script | Source | Target |
|--------|--------|--------|
| `ari sync user agents` | `agents/` | `~/.claude/agents/` |
| `ari sync user mena` | `mena/` | `~/.claude/commands/` + `~/.claude/skills/` |
| `ari sync user hooks` | `user-hooks/` | `~/.claude/hooks/` |

**Note:** Shell scripts `sync-user-*.sh` are deprecated. Use `ari sync user` commands instead.
```

#### Line 18: Update swap-team terminology
```diff
- | `swap-team.sh` | Switch active rite (syncs to `.claude/`) |
+ | `ari rite switch <name>` | Switch active rite (syncs to `.claude/`) |
```

**Note:** Keep the other scripts in this table unchanged for now (they may still be in use).

#### Lines 25-26: Update architecture note
```diff
- User-level content (`user-*/`) syncs to `~/.claude/` (global, available in all projects).
- Rite-level content (`rites/{rite}/`) syncs to `.claude/` (project-specific via swap-team).
+ User-level content syncs to `~/.claude/` (global, available in all projects) via `ari sync user`.
+ Rite-level content (`rites/{rite}/`) syncs to `.claude/` (project-specific) via `ari rite switch` or `ari sync materialize`.
```

#### Line 47: Fix user-agents path reference
```diff
- Syncs agents from `roster/user-agents/` to `~/.claude/agents/`.
+ Syncs agents from `knossos/agents/` to `~/.claude/agents/`.
```

#### Lines 50-69: Update sync-user-agents section
Replace the entire bash script usage section (lines 50-69) with:

```markdown
```bash
# Sync user-agents
ari sync user agents

# Preview changes
ari sync user agents --dry-run

# Show sync status
ari sync user agents --status
```

**Behavior:**
- Additive: Never removes existing agents from `~/.claude/agents/`
- Overwrites: Only agents previously installed from knossos (tracked in manifest)
- Preserves: User-created agents not from knossos
- Migration: Run `ari migrate roster-to-knossos` once to update existing manifests

**Integration Points:**
- Run manually after pulling knossos updates: `git pull && ari sync user agents`
- Add to shell profile for automatic sync (optional)

**Manifest:** `~/.claude/USER_AGENT_MANIFEST.json` tracks knossos-managed agents.
```

#### Lines 75-108: Update sync-user-commands section
Replace the entire section (lines 75-108) with:

```markdown
### Sync User Mena

Syncs mena content (commands + skills) from `knossos/mena/` to `~/.claude/commands/` and `~/.claude/skills/`.

```bash
# Sync all user mena
ari sync user mena

# Preview changes
ari sync user mena --dry-run

# Show sync status
ari sync user mena --status
```

**Behavior:**
- Additive: Never removes existing content from `~/.claude/`
- Overwrites: Only content previously installed from knossos (tracked in manifest)
- Preserves: User-created commands/skills not from knossos
- Routes: `.dro.md` → commands, `.lego.md` → skills
- Flattens: Source subdirectories become flat in target
- Scope-aware: Only distributes mena with `scope: user` or `scope: ""` (default)

**Source Structure:**
```
mena/
  session/       # start, park, continue, handoff, wrap (5 dromena)
  workflow/      # task, sprint (2 dromena)
  operations/    # architect, build, qa, code-review, commit, spike, pr (7 dromena)
  navigation/    # consult, worktree, ecosystem, sessions, rite (5 dromena)
  meta/          # minus-1, zero, one (3 dromena)
  rite-switching/ # 10x, docs, hygiene, debt, sre, security, intelligence, rnd, strategy, forge (10 dromena)
  guidance/      # prompting, standards, lexicon, file-verification, cross-rite, rite-discovery (6 legomena)
  templates/     # documentation, doc-artifacts, justfile, atuin-desktop (4 legomena)
  session/       # shared, common (2 legomena)
```

**Rite Mena:**
Rite-specific mena lives in `rites/<rite>/mena/` and is synced to `.claude/commands/` and `.claude/skills/` by `ari rite switch` or `ari sync materialize`. Rite mena takes precedence over user mena of the same name (project > user).

**Manifest:** `~/.claude/USER_COMMAND_MANIFEST.json` and `~/.claude/USER_SKILL_MANIFEST.json` track knossos-managed content.
```

#### Lines 110-146: Remove sync-user-skills section
**DELETE entire section** (lines 110-146). Skills are now part of unified mena sync covered in the section above.

**Acceptance criteria:**
- [ ] Title says "Knossos"
- [ ] No references to deleted directories (`user-skills/`, `user-commands/`, `user-hooks/`)
- [ ] No references to shell scripts (`sync-user-*.sh`, `swap-team.sh`)
- [ ] All commands use `ari` CLI
- [ ] Paths reference `knossos/` not `roster/`
- [ ] Rite commands reference `rites/<rite>/mena/` not `rites/<rite>/commands/`
- [ ] Mena section explains dro/lego routing
- [ ] Scope filtering mentioned for user mena sync

---

### 2. Mena Frontmatter - Standalone Files (P0)

**Files requiring `name` field (18 standalone dromena):**

Each file needs frontmatter updated to add `name: <filename-without-extension>`.

#### Template:
```yaml
---
name: <filename>
description: <existing description>
<...other existing fields...>
---
```

#### File list with exact names:

| File Path | Name Value |
|-----------|------------|
| `/Users/tomtenuta/Code/knossos/mena/navigation/ecosystem.dro.md` | `ecosystem` |
| `/Users/tomtenuta/Code/knossos/mena/navigation/sessions.dro.md` | `sessions` |
| `/Users/tomtenuta/Code/knossos/mena/navigation/rite.dro.md` | `rite` |
| `/Users/tomtenuta/Code/knossos/mena/meta/minus-1.dro.md` | `minus-1` |
| `/Users/tomtenuta/Code/knossos/mena/meta/zero.dro.md` | `zero` |
| `/Users/tomtenuta/Code/knossos/mena/meta/one.dro.md` | `one` |
| `/Users/tomtenuta/Code/knossos/mena/operations/build.dro.md` | `build` |
| `/Users/tomtenuta/Code/knossos/mena/operations/architect.dro.md` | `architect` |
| `/Users/tomtenuta/Code/knossos/mena/workflow/sprint.dro.md` | `sprint` |
| `/Users/tomtenuta/Code/knossos/mena/workflow/task.dro.md` | `task` |
| `/Users/tomtenuta/Code/knossos/mena/rite-switching/intelligence.dro.md` | `intelligence` |
| `/Users/tomtenuta/Code/knossos/mena/rite-switching/docs.dro.md` | `docs` |
| `/Users/tomtenuta/Code/knossos/mena/rite-switching/security.dro.md` | `security` |
| `/Users/tomtenuta/Code/knossos/mena/rite-switching/sre.dro.md` | `sre` |
| `/Users/tomtenuta/Code/knossos/mena/rite-switching/forge.dro.md` | `forge` |
| `/Users/tomtenuta/Code/knossos/mena/rite-switching/hygiene.dro.md` | `hygiene` |
| `/Users/tomtenuta/Code/knossos/mena/rite-switching/strategy.dro.md` | `strategy` |
| `/Users/tomtenuta/Code/knossos/mena/rite-switching/10x.dro.md` | `10x` |
| `/Users/tomtenuta/Code/knossos/mena/rite-switching/debt.dro.md` | `debt` |
| `/Users/tomtenuta/Code/knossos/mena/rite-switching/rnd.dro.md` | `rnd` |

**Implementation note:** For each file:
1. Read the file
2. Locate the frontmatter block (between `---` markers at top of file)
3. Add `name: <value>` as the FIRST field after the opening `---`
4. Preserve all other fields exactly as-is
5. Maintain YAML indentation and formatting

**Acceptance criteria:**
- [ ] All 20 files have `name` field in frontmatter
- [ ] Name matches filename without `.dro.md` extension
- [ ] Name is first field in frontmatter block
- [ ] No other frontmatter fields modified
- [ ] YAML syntax remains valid

---

### 3. Mena Frontmatter - Teams Files (P0)

**Files requiring `name` field (4 teams dromena):**

| File Path | Name Value |
|-----------|------------|
| `/Users/tomtenuta/Code/knossos/teams/ecosystem-pack/mena/cem-debug.dro.md` | `cem-debug` |
| `/Users/tomtenuta/Code/knossos/teams/forge-pack/mena/new-team.dro.md` | `new-team` |
| `/Users/tomtenuta/Code/knossos/teams/forge-pack/mena/validate-team.dro.md` | `validate-team` |
| `/Users/tomtenuta/Code/knossos/teams/forge-pack/mena/eval-agent.dro.md` | `eval-agent` |

**Same implementation process as standalone files above.**

**Acceptance criteria:**
- [ ] All 4 files have `name` field in frontmatter
- [ ] Name matches filename without `.dro.md` extension
- [ ] No other frontmatter fields modified
- [ ] YAML syntax remains valid

---

### 4. MEMORY.md Metrics Update (P1)

**File:** `/Users/tomtenuta/.claude/projects/-Users-tomtenuta-Code-knossos/memory/MEMORY.md`

**Changes:**

#### Line 15: Update ADR count
```diff
- - **Ambiguity**: Check ADRs first (docs/decisions/, 22 total). Then follow existing patterns.
+ - **Ambiguity**: Check ADRs first (docs/decisions/, 23 total). Then follow existing patterns.
```

#### Line 43: Update mena metrics
```diff
- - 12 rites (58 agents) + 3 user agents + 34 dromena + 11 legomena
+ - 12 rites (58 agents) + 3 user agents + 34 dromena + 12 legomena
```

#### Line 45: Update ADR count
```diff
- - 22 ADRs
+ - 23 ADRs
```

#### Lines 63-69: Add Mena Scope Initiative to Completed Initiatives
```diff
 ## Completed Initiatives
 1. Session Hardening (1677d66)
 2. Session Forking (99a022e)
 3. Roster→Knossos Rename (bbbc026)
 4. Dromena/Legomena Convention (ADR-0023)
 5. Skills→Commands Unification (ADR-0021)
 6. Agent Factory (1e57c07)
+ 7. Mena Scope Initiative (65012dd, ADR-0025)
```

**Acceptance criteria:**
- [ ] ADR count is 23 in two locations
- [ ] Legomena count is 12
- [ ] Mena Scope Initiative added to Completed Initiatives with commit hash and ADR reference
- [ ] Current Priorities section unchanged (user can update separately if needed)

---

### 5. context-engineer.md YAML Fix (P1)

**File:** `/Users/tomtenuta/Code/knossos/agents/context-engineer.md`

**Issue:** Line 3's `description` field contains unquoted text with colon-space pattern, which breaks YAML parsing.

**Current (lines 1-5):**
```yaml
---
name: context-engineer
description: |
  Use this agent when optimizing how Claude itself is leveraged, rather than what software is being built. Specifically, use when: designing or restructuring Skills architecture, optimizing prompt structures and token economics, improving context management across multi-turn conversations, implementing progressive disclosure patterns, architecting agentic workflows, diagnosing why Claude loses context mid-session, or deciding whether to use new Claude features. This agent operates at the meta-level, engineering the system that executes work rather than executing the work itself.

```

**The problem:** The multi-line literal `|` is correct, but the integration test is failing. Need to verify the exact YAML syntax issue by examining how other agents format their descriptions.

**Action:** Compare with a working agent file to determine the correct fix. Most likely needs:
- Proper indentation of continuation lines
- OR conversion to folded scalar `>`
- OR explicit quotes around the literal block

**Recommended fix (use folded scalar for long descriptions):**
```yaml
---
name: context-engineer
description: >
  Use this agent when optimizing how Claude itself is leveraged, rather than what software is being built.
  Specifically, use when: designing or restructuring Skills architecture, optimizing prompt structures and
  token economics, improving context management across multi-turn conversations, implementing progressive
  disclosure patterns, architecting agentic workflows, diagnosing why Claude loses context mid-session,
  or deciding whether to use new Claude features. This agent operates at the meta-level, engineering
  the system that executes work rather than executing the work itself.
```

**Acceptance criteria:**
- [ ] YAML parses without errors
- [ ] Description content preserved (may be reformatted)
- [ ] Integration test no longer marks it as knownBroken
- [ ] `ari agent validate context-engineer` succeeds

---

### 6. User Hooks Manifest Fix Documentation (P1)

**Context:** The audit found that `~/.claude/USER_HOOKS_MANIFEST.json` still contains "roster" branding and stale `ari/` prefixes in keys.

**This is NOT a code change.** The fix is running existing commands. Document the procedure in this brief for the user to execute.

**Procedure:**

```bash
# Step 1: Run roster-to-knossos migration (idempotent)
ari migrate roster-to-knossos

# Step 2: Re-sync user hooks to regenerate manifest
ari sync user hooks

# Step 3: Verify manifest is clean
cat ~/.claude/USER_HOOKS_MANIFEST.json
```

**Expected result:**
- `source` fields say `"knossos"` not `"roster"`
- Keys have no `ari/` prefix (unless that's the actual hook name)
- Checksums are current

**Acceptance criteria:**
- [ ] User has run the three commands
- [ ] USER_HOOKS_MANIFEST.json contains "knossos" branding
- [ ] No stale `ari/` prefixes in keys

---

## Sequencing & Dependencies

### Phase 1: Frontmatter Additions (Independent, can parallelize)
1. Add `name` field to 20 standalone mena files
2. Add `name` field to 4 teams mena files

**Estimated time:** 45 minutes
**Blocker for:** None (can be done in parallel with other work)

### Phase 2: README.md Update (Independent)
1. Update README.md following the line-by-line change list

**Estimated time:** 60 minutes
**Blocker for:** None

### Phase 3: MEMORY.md Update (Independent)
1. Update MEMORY.md metrics

**Estimated time:** 10 minutes
**Blocker for:** None

### Phase 4: YAML Fix (Independent)
1. Fix context-engineer.md YAML syntax
2. Validate with `ari agent validate context-engineer`

**Estimated time:** 15 minutes
**Blocker for:** None

### Phase 5: User Action (Not part of tech writer work)
1. User runs the three commands to fix hooks manifest

**Estimated time:** 5 minutes
**Done by:** User, not tech writer

**Total estimated effort:** 2-3 hours for phases 1-4

---

## Verification Checklist

After completing all changes, verify:

### Frontmatter
- [ ] Run `grep -r '^name:' /Users/tomtenuta/Code/knossos/mena/**/*.dro.md | wc -l` returns 34 (all dromena have names)
- [ ] Run `grep -r '^name:' /Users/tomtenuta/Code/knossos/teams/**/*.dro.md | wc -l` returns 4
- [ ] No YAML parse errors: `ari agent validate --all` succeeds

### README.md
- [ ] No occurrences of "Roster" in title
- [ ] No references to `user-skills/`, `user-commands/`, `user-hooks/` directories
- [ ] No references to `.sh` sync scripts in usage examples
- [ ] All `ari` commands are correctly formatted
- [ ] grep "rites/<rite>/commands/" returns no matches
- [ ] grep "rites/<rite>/mena/" finds the correct reference

### MEMORY.md
- [ ] Metrics reflect reality: 23 ADRs, 12 legomena
- [ ] Mena Scope Initiative listed in Completed Initiatives

### Agent
- [ ] `ari agent validate context-engineer` exits 0
- [ ] Integration tests pass (or at least context-engineer no longer in knownBroken map)

---

## Out of Scope (Explicitly NOT in this brief)

### P2 Tech Debt - Future Work
1. **ADR updates**: 14 of 23 ADRs reference pre-unification concepts (user-skills/, skills: field)
   - **Why out of scope:** Requires architectural decision on whether to update historical ADRs or add addendum notes
   - **Future work:** Separate brief for ADR maintenance strategy

2. **Test coverage for internal/sync**
   - **Why out of scope:** Test backfill is explicitly out of scope per project conventions
   - **Future work:** Write tests with new code, don't backfill existing gaps

3. **Test coverage for Materialize() entry point**
   - **Why out of scope:** Same as above
   - **Future work:** Tests should be added when Materialize() is next modified

4. **Shell script deprecation**
   - **Why out of scope:** README now documents that scripts are deprecated, but removing them requires verifying nothing depends on them
   - **Future work:** Separate brief for shell script removal after dependency audit

5. **README.md line 16-21 (generate-team-context.sh, etc.)**
   - **Why out of scope:** These scripts were not flagged as broken in the audit
   - **Future work:** Verify if these are deprecated and update in next iteration

---

## Notes for Tech Writer

### Working with Frontmatter
- The frontmatter is the YAML block at the top of each file between `---` markers
- Always preserve existing fields exactly as-is
- Add `name` as the FIRST field in the block (conventionally, name comes first)
- Maintain 2-space YAML indentation
- Test YAML validity after each change (use an online YAML validator or `ari agent validate` if applicable)

### Working with README.md
- This is a large file (146 lines). The changes are scattered throughout.
- Work carefully through each line number reference
- Some sections are being replaced entirely, others are line-by-line edits
- After editing, read through the entire file to ensure flow and coherence

### Working with MEMORY.md
- This file lives in `~/.claude/projects/` which is user-private, NOT in the knossos repo
- Changes are minimal (metrics updates only)
- Do NOT change the "Current Priorities" section - that's user-owned

### Testing Your Changes
- After frontmatter changes: `ari agent validate --all` or `ari sync materialize --dry-run`
- After README changes: Read through the file to ensure narrative coherence
- After MEMORY.md changes: Verify numbers match reality by checking actual counts

### Getting Help
If you encounter:
- **YAML parse errors:** Check indentation (must be 2 spaces) and quotes (use `>` or `|` for multi-line)
- **Unclear line numbers:** Line numbers are from the original file read, count carefully from top
- **Missing context:** Refer back to the audit findings summary at top of brief

---

## Success Criteria

This brief is complete when:

1. ✅ All 24 mena files have `name` frontmatter field
2. ✅ README.md reflects current architecture (no stale references)
3. ✅ MEMORY.md metrics are accurate
4. ✅ context-engineer.md YAML is valid
5. ✅ Verification checklist passes 100%
6. ✅ User has been notified about the hooks manifest fix procedure

**Definition of Done:** All P0 and P1 items resolved. Knossos documentation is accurate and dogfooding-ready.

---

## Appendix A: Audit Source

This brief was generated from audit findings by:
- **Doc Auditor**: Identified staleness, redundancy, gaps
- **Information Architect**: Designed this content brief

Audit completion date: 2026-02-07
Brief creation date: 2026-02-07

---

## Appendix B: Quick Reference - File Counts

For verification during work:

| Category | Count | Location |
|----------|-------|----------|
| Total ADRs | 23 | `docs/decisions/*.md` |
| Standalone mena dromena | 34 | `mena/**/*.dro.md` |
| Teams mena dromena | 4 | `teams/**/mena/*.dro.md` |
| Total legomena | 42 | `**/*.lego.md` |
| Mena-only legomena | 12 | `mena/**/*.lego.md` |
| User agents | 3 | `agents/` (context-engineer, moirai, orphan-guardian) |

---

*End of brief. Ready for Tech Writer execution.*

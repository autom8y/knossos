# SPIKE: Knossos Source Path Antipattern in Materialized Artifacts

## Question and Context

**What are we trying to learn?**
How pervasive is the antipattern where knossos source tree paths (`rites/shared/mena/`, `rites/{rite}/mena/`, `$KNOSSOS_HOME/rites/`) are referenced inside artifacts that get materialized into `.claude/` directories in satellite projects? These paths only exist in the knossos core repo and break (or produce noise) in every satellite.

**What decision will this inform?**
Whether this requires a targeted fix (just the known `/know` bug) or a systematic sweep across all mena source files. Determines the scope of a cleanup sprint.

**Timebox**: Single research session.

## Approach Taken

1. Started from the known bug in `rites/shared/mena/know/INDEX.dro.md` (lines 57-63)
2. Ran systematic `Grep` scans across the full knossos tree for:
   - `Read("rites/...")` tool call patterns in all `.md` files
   - `rites/shared/mena/` and `rites/{rite}/mena/` string references in mena source files
   - `$KNOSSOS_HOME` references in rite mena directories
   - Template files (`.tpl`) for path leakage into generated CLAUDE.md
   - Go source code for path emission into materialized artifacts
3. Classified each finding by severity (functional bug vs. documentation leak vs. harmless)

## Findings

### Category 1: Functional Bugs -- `Read()` Tool Calls Using Knossos Source Paths

These are the highest severity. When materialized as commands/skills in a satellite, the `Read()` call targets a path that does not exist.

| File | Line(s) | Path Referenced | Fallback? | Severity |
|------|---------|-----------------|-----------|----------|
| `rites/shared/mena/know/INDEX.dro.md` | 57 | `Read("rites/shared/mena/pinakes/domains/{domain}.lego.md")` | Yes -- falls back to `.claude/skills/pinakes/domains/{domain}.md` | **HIGH** -- produces error in every satellite invocation |
| `rites/forge/mena/theoria.dro.md` | 48 | `Read("rites/shared/mena/pinakes/INDEX.lego.md")` | No fallback | **HIGH** -- but forge rite is knossos-internal (see mitigating factor below) |
| `rites/shared/mena/research/INDEX.dro.md` | 91 | `Read("rites/shared/mena/literature-review/evidence-grading.md")` | Yes -- "(or from the materialized skill at `.claude/skills/literature-review/evidence-grading.md`)" | **MEDIUM** -- has fallback with "or" language but knossos path is listed first |
| `rites/shared/mena/research/INDEX.dro.md` | 107 | `Read("rites/shared/mena/literature-review/schemas/synthesis.md")` | Yes -- "(or from materialized skill path)" | **MEDIUM** -- same pattern as above |

**Total functional bugs: 4 instances in 3 files.** Two are in shared mena (affect all satellites), one is in forge mena (knossos-internal only).

### Category 2: Informational/Documentation References to Knossos Source Paths

These are references in legomena (skills) or dromena (commands) that mention knossos source paths as documentation, not as tool calls. They would be confusing in a satellite but don't cause functional failures.

| File | Line(s) | Pattern | Impact |
|------|---------|---------|--------|
| `rites/forge/mena/theoria.dro.md` | 70, 324, 361-362 | Example paths and "Related" section citing `rites/shared/mena/pinakes/...` | Low -- forge is knossos-internal |
| `rites/shared/mena/ephemeral-artifacts/INDEX.lego.md` | 62-63 | Instructs writing to `rites/shared/mena/` for permanent artifacts | **MEDIUM** -- satellites don't have this directory |
| `rites/shared/mena/pinakes/registry-format.lego.md` | 25 | Template path `rites/shared/mena/pinakes/domains/{domain}.lego.md` | Low -- this is authoring guidance for knossos contributors |
| `mena/rite-switching/*.dro.md` | Various | `Full documentation: rites/{rite}/mena/{rite}-ref/INDEX.lego.md` (9 files) | **LOW** -- informational reference, not operational |
| `mena/guidance/cross-rite/INDEX.lego.md` | 124 | Relative path to `rites/shared/mena/cross-rite-handoff/INDEX.lego.md` | Low -- markdown link, not tool call |
| `rites/{strategy,intelligence,rnd,security}/mena/*-ref/INDEX.lego.md` | 14-15 each | `$KNOSSOS_HOME/rites/{rite}/agents/` and `/workflow.yaml` | **MEDIUM** -- these ref skills deploy to satellites but `$KNOSSOS_HOME` is not set there |

**Total informational leaks: ~18 instances across ~15 files.**

### Category 3: $KNOSSOS_HOME References in Forge-Only Files

17 files contain `$KNOSSOS_HOME` references, but 12 of these are in `rites/forge/` which is knossos-internal (the forge rite operates on the knossos repo itself). The remaining 5 are the ref skills from Category 2 above.

### Category 4: Clean Areas (No Antipattern Found)

| Area | Result |
|------|--------|
| **Go templates** (`knossos/templates/sections/*.md.tpl`) | Clean -- all paths use `.claude/` or `.know/` |
| **Go source code** (`internal/materialize/`) | Clean -- `rites/` paths are used internally for embedded FS resolution, never emitted into output |
| **Platform-level mena** (`mena/`) `Read()` calls | Clean -- no `Read("rites/...")` calls found |
| **Relative path `Read()` calls** | Clean -- no `Read("../...")` calls found anywhere |
| **Agent prompts** (`rites/*/agents/*.md`) | Clean -- forge agents use `$KNOSSOS_HOME` intentionally (they operate on knossos) |

## Root Cause Analysis

The antipattern has a single root cause: **mena source files were authored in the knossos repo, where the source tree IS the working tree.** When writing `Read("rites/shared/mena/pinakes/domains/{domain}.lego.md")`, the author tested it in knossos core where that path resolves correctly. But after `ari sync` materializes this dromenon to `.claude/commands/know/INDEX.md` in a satellite, the `rites/` path no longer exists.

The correct path in a satellite is always the materialized location:
- Legomena: `.claude/skills/{name}/{file}.md` (strips `.lego.md` extension)
- Dromena: `.claude/commands/{name}/{file}.md` (strips `.dro.md` extension)
- Companion files: `.claude/skills/{name}/{subpath}` or `.claude/commands/{name}/{subpath}`

### Why Only 4 Functional Bugs?

Most mena files use `Skill("name")` for loading sibling skills (which resolves correctly via CC's skill system) rather than `Read()` with explicit paths. The `Read()` pattern only appears when a dromenon needs to read a specific sub-file from a sibling skill's directory (e.g., a specific domain criteria file from pinakes). This is an uncommon pattern, which limits the blast radius.

## Severity Matrix

| Severity | Count | Description |
|----------|-------|-------------|
| **HIGH** | 2 | `Read()` calls with no fallback or knossos-first fallback ordering |
| **MEDIUM** | 4 | `Read()` calls with "or" fallback + informational refs in satellite-deployed skills |
| **LOW** | ~16 | Documentation references, markdown links, forge-internal paths |

## Recommendation

### Fix 1: Reverse Path Resolution in /know (Priority 1, 5 minutes)

In `rites/shared/mena/know/INDEX.dro.md`, lines 57-63:

**Current**:
```
Read("rites/shared/mena/pinakes/domains/{domain}.lego.md")
```
Fallback to:
```
Read(".claude/skills/pinakes/domains/{domain}.md")
```

**Proposed**:
```
Read(".claude/skills/pinakes/domains/{domain}.md")
```
Fallback to:
```
Read("rites/shared/mena/pinakes/domains/{domain}.lego.md")
```

### Fix 2: Reverse Path Resolution in /research (Priority 1, 5 minutes)

In `rites/shared/mena/research/INDEX.dro.md`, lines 91 and 107, swap the resolution order so the materialized `.claude/skills/` path is tried first.

### Fix 3: Fix /theoria Pinakes Path (Priority 2, conditional)

In `rites/forge/mena/theoria.dro.md`, line 48:
```
Read("rites/shared/mena/pinakes/INDEX.lego.md")
```
Should be:
```
Read(".claude/skills/pinakes/INDEX.md")
```

**Mitigating factor**: The forge rite is currently knossos-internal. If it ever deploys to satellites, this becomes a P1 bug. Fix proactively.

### Fix 4: Update $KNOSSOS_HOME References in Ref Skills (Priority 3, batch)

The 4 ref skills (`strategy-ref`, `intelligence-ref`, `rnd-ref`, `security-ref`) that reference `$KNOSSOS_HOME/rites/` in their Quick Reference tables should either:
- Replace with `.claude/agents/` (the materialized location), or
- Conditionally note "In knossos core: `$KNOSSOS_HOME/rites/...`, in satellites: `.claude/agents/...`"

This is low priority since these are informational, not operational.

### Fix 5: Update ephemeral-artifacts Guidance (Priority 3)

The `rites/shared/mena/ephemeral-artifacts/INDEX.lego.md` lines 62-63 instruct writing to `rites/shared/mena/` which only exists in knossos core. Add satellite-appropriate guidance.

### Structural Prevention

Consider adding a **lint rule or theoria audit criterion** that detects `Read("rites/` patterns in `.dro.md` and `.lego.md` files. This would catch new instances as they're authored. The pinakes `dromena` or `legomena` domain criteria could include a check for "no knossos source path references in tool call instructions."

## Follow-Up Actions

1. Apply Fix 1 (know) and Fix 2 (research) -- 10 minutes total
2. Apply Fix 3 (theoria) -- 5 minutes, proactive
3. Consider Fix 4 and Fix 5 as part of next documentation sweep
4. Add "source path leakage" check to the mena-structure pinakes domain criteria
5. Document this as a known antipattern in the forge's agent-prompt-engineering skill (guidance for authors)

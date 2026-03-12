# Tier Mapping: HANDOFF to Prompt Sections

## CRITICAL Tier -- Baked Into System Prompt

**Destination**: `## Domain Knowledge` section in the agent's `.md` file, placed after Core Responsibilities and before Protocol/Approach.

**Budget**: 30-50 lines per agent. Exceeding 50 lines signals scope creep -- split items to IMPORTANT tier.

**What qualifies as CRITICAL**:
- Failure modes the agent could re-introduce (SCAR-referenced)
- Exousia jurisdiction boundaries (EX-referenced)
- Load-bearing constraints where violation causes data loss or system instability
- Contracts with external systems (CC file watcher, stdin payload format)

**Format per item**:
```markdown
- **[SCAR-NNN + TRIBAL-NNN] IMPERATIVE statement.** Brief rationale with file/function references where relevant.
```

**Example** (from ecosystem rite):
```markdown
## Domain Knowledge

- **[SCAR-003 + TRIBAL-003] NEVER use os.RemoveAll on user-containing directories.** Use selective removal with a managed-set of filenames. The user-content-never-destroyed invariant is the platform's core trust contract.
- **[SCAR-001 + GUARD-011] All project channel directory writes MUST use writeIfChanged().** The harness file watcher crashes on unnecessary writes. Plain os.WriteFile is acceptable only for user-scope channel paths.
```

## IMPORTANT Tier -- Skill Reference

**Destination**: A `domain-knowledge/` skill directory in the new rite's `mena/`. Added to the agent's `skills:` frontmatter list.

**Organization**: Group by topic, not by source ID. One companion file per topic area.

**What qualifies as IMPORTANT**:
- Navigation guides for known tensions (how to work within a tradeoff)
- Golden path rules (proven patterns to follow)
- Decision tables for recurring choices
- Supplementary context that aids but is not mandatory

**Naming convention**: `domain-knowledge/INDEX.lego.md` + topic files (e.g., `pipeline-stages.md`, `error-handling.md`).

## NICE-TO-HAVE Tier -- Discoverable

**Destination**: A brief "Further Reading" section at the bottom of the agent prompt, or left in the archived HANDOFF without explicit prompt reference.

**What qualifies as NICE-TO-HAVE**:
- Historical context that explains why something is the way it is
- Low-risk awareness items (dead code locations, migration artifacts)
- Infrequently relevant details the agent can look up when needed

**Format**: 2-3 bullet points max, pointing to source files or the archived HANDOFF.

## Per-Agent Processing Sequence

1. **Read RITE-SPEC** (full, loaded once) -- understand each agent's role and responsibilities
2. **Read HANDOFF Cross-Agent Knowledge** (CK-NN, loaded once) -- shared rules all agents need
3. **For each agent**:
   a. Read the agent's Prompt Fuel section from HANDOFF
   b. Read applicable Exousia Overrides (EX-NN entries targeting this agent or ALL agents)
   c. Read applicable Anti-Pattern Catalog entries (AP-NN)
   d. Write CRITICAL items into `## Domain Knowledge`
   e. Package IMPORTANT items into `domain-knowledge/` skill files
   f. Add NICE-TO-HAVE as "Further Reading" references
   g. Verify every constraint has source ID traceability
   h. Move to next agent

## Cross-Agent Knowledge Handling

CK-NN items from the HANDOFF apply to ALL agents. Two strategies:

- **Embed in each agent**: If the CK item directly affects the agent's behavior (e.g., CK-04 "CC File Watcher Sensitivity" for any agent that writes files)
- **Shared preamble**: If the CK item is foundational context (e.g., CK-06 "ADRs Are the Architectural Authority"), add to a shared skill referenced by all agents rather than duplicating

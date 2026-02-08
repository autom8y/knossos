# Stakeholder Preferences: Distribution Readiness

> Authoritative reference for all distribution readiness work on Knossos.
> Source: Structured interview (2026-02-08) + doctrine exploration + source material audit + CC alignment analysis.

**Status**: Active
**Owner**: Tom Tenuta (Stakeholder)
**Created**: 2026-02-08
**Last Updated**: 2026-02-08

---

## 1. Vision

### What Is Knossos
A context-engineering meta-framework -- "Rails for Claude Code." Knossos materializes source definitions (rites, agents, skills, hooks, templates) into `.claude/` projections that configure Claude Code for structured, multi-agent workflows with session lifecycle management, confidence signaling, and audit provenance.

### Distribution Target
**Staged rollout**: Internal team first, then expanding to trusted external developers, then broader availability. Each stage has a different readiness bar. Current focus is on the **internal-first** stage.

### What "Distribution Ready" Means
Progressive readiness bar, in stakeholder-specified priority order:

1. **Clone, build ari, run `ari sync materialize`, start a session** -- the technical foundation works
2. **Run a complete /task or /sprint cycle** on a real project without hitting dead ends or confusion
3. **Use /consult (Pythia) as an oracle** to understand the framework, evaluate it, and get insights -- this is NOT marketing material, it's the product's internal self-knowledge
4. **Adopt for their own project** -- create custom rites/agents and it works

> Note: The stakeholder explicitly reordered the original options. The Pythia/oracle experience (3) was promoted above project adoption (4). This reflects a strategic priority: internal self-knowledge and the oracle UX matter more than extensibility at this stage.

### Timeline
**"Right balance" of quality and iteration**: No artificial deadline. Quality bar matters more than speed. But also: ship to get feedback and iterate. Not perfectionism, not rushing.

---

## 2. Core Decisions

### 2.1 Mythology Is Load-Bearing

The mythological naming (Dromena, Legomena, Moirai, Ariadne, Pythia, etc.) is **not decoration or branding -- it is the architecture**. Each mythological element encodes architectural intent, relationship patterns, constraint awareness, and failure modes (per `docs/doctrine/philosophy/mythology-concordance.md`).

**Implication**: Never simplify mythology away. Make it **learnable** through layered teaching:
- A glossary skill exists for CC's always-available context (framework fundamentals)
- The oracle level (/consult, Pythia) needs deeper alignment with doctrine to help users learn
- Progressive disclosure: encounter terms naturally, explain on demand
- Dedicated lexicon skill for deep dives

**Stakeholder observation**: "The oracle level of our framework could probably benefit from greater context and alignment with the doctrine at higher levels to help users even more." This is a direct gap identification -- Pythia needs to internalize the doctrine more deeply to serve as an effective teacher of the mythology.

### 2.2 The Pythia Experience Is the Flagship

**/consult is the PRIMARY user entry point** and its quality directly correlates to the value users get from the framework. Two entry points exist:
- **CLAUDE.md (Inscription)**: The entry for the model -- CC reads this to understand the labyrinth
- **/consult (Pythia)**: The entry for the user -- humans engage at a higher meta-level to get insights about how to navigate the labyrinth

> Note: The doctrine (`mythology-concordance.md`) maps Pythia broadly to all orchestrator agents. The stakeholder's usage narrows Pythia specifically to the Consultant agent and /consult command. This doctrinal refinement should be reconciled -- either the concordance should be updated to reflect this narrower meaning, or the broader meaning should be acknowledged here.

The consultant must have **full dynamic exploration capability** -- not just static knowledge. It should read rite manifests, agent files, skill files dynamically when answering queries. Static knowledge for speed, dynamic exploration for depth and accuracy. (Note: "full exploration" is a stakeholder requirement; the specific implementation approach -- static + dynamic fallback -- is a proposed design.)

**Bar**: World-class. Every missing skill it references must be created. Every stale reference must be fixed. Every mode (no args, query, --playbook, --rite, --commands) must produce correct, useful output.

**Critical insight from stakeholder**: "The consult will only be as powerful as the underlying concepts, architecture, lexicon, systems, etc. So that's probably the first primary initiative that's a big lift." -- The foundation must be solid before Pythia can shine.

### 2.3 Git Workflow

**Direct atomic commits to main** in optimal logical chunks. No feature branches while greenfield. This is the standing practice for the foreseeable future.

### 2.4 Disable-Model-Invocation (DMI) Policy

**Multi-agent cascades only**: Commands that launch agent cascades (/task, /sprint, /build, /architect) get DMI. Single-action commands are fine for CC to auto-invoke. Moirai's legitimate CLI interactions must NOT be blocked.

**Rationale**: Too strict breaks things like Moirai's state management via ari CLI. Too permissive risks autonomous multi-agent cascades a user didn't ask for.

### 2.5 Dead Reference Policy

**Categorize and batch**: When finding references to things that don't exist:
1. Never assume deletion is the answer
2. Investigate git history and design docs for context
3. Categorize: likely lost artifact vs aspirational vs evolved concept vs stale
4. Present the batch with categories and recommendations
5. Get stakeholder decision on each category

**Lesson learned**: The Moirai Fates references (clotho.md, lachesis.md, atropos.md) are likely a partially completed migration -- the stakeholder believes these were intended to be converted to mena/lego format for the Moirai agent, as designed in TDD-fate-skills.md. Design review needed before any changes. The stakeholder's language was tentative ("I think these may have intended to be converted"), indicating this hypothesis needs verification, not assumption.

### 2.6 Hook Shell-to-Binary Migration

Current state needs verification, but **shell-to-binary migration must be complete for distribution readiness**. `ari` IS the hook binary. Shell scripts are the legacy path. References to `.claude/hooks/PreToolUse.sh` are stale.

### 2.7 Binary Distribution

**Keep ari binary in git for now.** Homebrew is the eventual distribution path. No Makefile, no goreleaser needed for the internal stage.

### 2.8 Session Hygiene

**Auto-archive in /wrap** aligns with framework principles (Doctrine Principle #3: Mutation Through the Fates, specifically Atropos for termination). Additionally, session GC via `ari session gc` / `ari naxos scan` may be required per the agent-vs-binary responsibility split -- needs review.

The doctrine is clear: Naxos (scan/report) is a binary responsibility. Termination (archive/delete) is Atropos's domain. The combination is: `ari naxos scan` reports → human/agent decides → `ari session wrap/archive` executes.

### 2.9 Rules Purpose

**Rules are primarily for Claude Code's context** when autonomously editing packages, not human contributor guardrails. Rules are a newer CC primitive that hasn't been added to the mythological lexicon yet. They are path-scoped instructions that CC loads when editing files in matching directories.

### 2.10 Hook Result Struct

**Legacy, should be deprecated.** `hook.Result` with `WriteAllow()`/`WriteBlock()` produces flat JSON without the `hookSpecificOutput` wrapper CC expects for PreToolUse. `hook.PreToolUseOutput` is the correct implementation. Both exist with tests. Result is unused in production but a trap for future hook authors.

### 2.11 Consultant Agent Task Tool

**Bug, remove it.** The intent was sub-agent exploration, but CC silently strips Task from subagents. The implementation was wrong. However, since /consult needs full dynamic exploration, the consultant's ability to explore the codebase should be achieved through its own tools (Bash, Glob, Grep, Read, WebSearch) -- not through Task delegation to other agents.

### 2.12 Rite Catalog

**Full catalog (all 11 user-facing rites)** ships for internal-first stage: 10x-dev, debt-triage, docs, ecosystem, forge, hygiene, intelligence, rnd, security, sre, strategy. (`shared` is a dependency bundle, not a user-facing rite.) Users can explore. Some rites may be rough around edges but all are available.

### 2.13 Autonomy Model for Implementation

**Fix obvious bugs, ask on judgment calls.** Typos, stale references, clearly wrong tool lists: fix autonomously. Architecture decisions, design choices, anything ambiguous: ask the stakeholder.

---

## 3. Work Priorities & Sequencing

### Foundational Insight

> "The consult will only be as powerful as the underlying concepts, architecture, lexicon, systems, etc."

This means the work order is:
1. **Fix the foundations** -- commit the tree, fix bugs that break things, clean stale references
2. **Strengthen the underlying systems** -- complete missing skills, fix the lexicon, align the knowledge base
3. **Make Pythia world-class** -- with solid foundations, the consultant can route accurately and deeply
4. **Moirai Fates design review** -- deep contextual review of the architecture from an agent's perspective

### What's In Scope

| Category | Items |
|----------|-------|
| **Fix what breaks** | Commit uncommitted CC alignment sweep, add DMI to multi-agent commands, fix consultant Task tool, normalize session status values |
| **Complete missing pieces** | Create rite-discovery skill, fix/create consultant knowledge base, fix stale capability-index.yaml, create Moirai Fate skill stubs (after design review) |
| **Strengthen foundations** | Fix all dead references (categorize and batch), complete hook shell→binary migration, update moirai-invocation.md for CLI-vs-agent clarity |
| **Flagship UX** | Make /consult work end-to-end across all 5 modes, full dynamic exploration, accurate rite routing |
| **Autocomplete cleanup** | Fix companion file leakage (~60 entries → ~30). Must fix for distribution. |
| **Rules materialization** | Add missing rules for inscription, hook, sails, usersync packages |
| **Deprecation** | Mark hook.Result struct as deprecated with clear guidance |

### What's Out of Scope

| Category | Rationale |
|----------|-----------|
| **New rites or agents** | Fix and polish what exists |
| **CC feature parity** | Don't chase missing CC fields (permissionMode, mcpServers, etc.) unless they break something |
| **Performance optimization** | Hook latency, materialization speed are fine for now |
| **Distribution packaging** | Deferred initiative -- product must be ready first |
| **skills field population** | 0/58 agents use it; skills still load via autonomous discovery from body text |
| **Archetype definitions for designer/analyst/engineer/meta** | WARN-level validation runs; no silent failures |
| **SubagentStart/SubagentEnd hooks** | Env infrastructure exists; no runtime gap |
| **Scope field adoption** | Single-project beta doesn't need scoping |
| **Orchestrator template deduplication** | Maintenance burden, no runtime impact |
| **PreCompact custom_instructions** | Rotation works; this improves post-compaction context |

---

## 4. Doctrine Understanding

### The 8 Design Principles (Source: `docs/doctrine/philosophy/design-principles.md`)

1. **The Clew Is Sacred** -- Every action must knot the clew. Unrecorded events make return uncertain.
2. **Honest Signals Over Comfortable Lies** -- WHITE/GRAY/BLACK. Gray is honesty. Ship gray with eyes open.
3. **Mutation Through the Fates** -- Only Moirai may modify session state. Clotho creates, Lachesis tracks, Atropos terminates.
4. **Rites Over Teams** -- Flexible practices, not fixed membership. Invoke what you need.
5. **Heroes Are Mortal** -- Context is finite. Design for summoning with rich context, not omniscient heroes.
6. **The Labyrinth Grows** -- Complexity extends. The clew must accommodate growth.
7. **Return Is the Victory** -- Merged PR > heroic effort lost to context collapse.
8. **The Inscription Prepares** -- CLAUDE.md is the labyrinth speaking at entry.

### Key Architectural Boundaries

| Boundary | Owner | Mechanism |
|----------|-------|-----------|
| Session state mutations | Moirai agent (reasoning) → ari CLI (execution) | Write guard hook blocks direct writes |
| Materialization | ari binary | Source → .claude/ projection, idempotent, user content preserved |
| Confidence signaling | ari sails check (computation) → Atropos (orchestration) | WHITE_SAILS.yaml artifact |
| Orphan detection | ari naxos scan (report-only) | Human decides, ari executes |
| Context injection | SessionStart hook → ari hook context | Automatic, session-aware |

### The Agent-Binary Responsibility Split

- **ari (Ariadne)**: Faithful, not intelligent. Provides commands that anchor, track, and cut the clew. Authoritative for state changes. Deterministic execution.
- **Agents (Heroes)**: Intelligent, not faithful. Make decisions, summon specialists, create artifacts. Reasoning layer that delegates to ari for state changes.
- **Moirai (Fates)**: Bridge between agent reasoning and binary execution. Validates, reasons, then delegates to ari CLI.

---

## 5. Technical State Assessment

### What Works Well (Don't Touch)

1. **Orchestrator consistency**: 10/10 orchestrators match archetype (Read-only, maxTurns=3, disallowedTools)
2. **Materialization pipeline**: Idempotent, atomic writes, user content preserved
3. **CLAUDE.md**: ~740 tokens, lean, purposeful
4. **PreToolUse hooks**: Correct CC format (hookSpecificOutput.permissionDecision)
5. **Progressive disclosure**: INDEX + companion pattern keeps skill loads compact
6. **Reviewer pattern**: All 7 reviewers have disallowedTools: [Task], contract.must_not, maxTurns: 15
7. **Test suite**: 1,347 tests, 0 failures, 25 packages
8. **Lock protocol**: JSON LockMetadata v2, 5-minute stale, advisory flock

### Known Issues by Priority

**P0 (Breaks things)**:
- Committed main does not build (SetSyncDir reference)
- 5-6 dromena missing DMI for multi-agent cascades
- Consultant agent has Task tool (CC strips it) -- both `agents/consultant.md` (tools field) AND `mena/navigation/consult/INDEX.dro.md` (allowed-tools field) need Task removed
- 3 sessions with invalid status values (COMPLETED/COMPLETE instead of ARCHIVED)
- Moirai references non-existent Fate skill files (partially completed migration)

**P1 (Causes confusion)**:
- /consult dependencies missing (rite-discovery skill, playbooks, stale capability-index)
- Companion file autocomplete leakage (~60 entries vs ~30)
- Hook dual output format (Result vs PreToolUseOutput)
- Conflicting Moirai guidance (dromena use CLI, moirai-invocation says use Task(moirai))
- 4 packages lack rules files
- MOIRAI_BYPASS env var: code checks "1", docs say "true", mechanism can't work through CC Bash

**P2 (Deferred)**:
- See "Out of Scope" section above

### The Moirai Fates Migration

**Status**: Partially designed, not implemented.

**Design**: TDD-fate-skills.md (draft) specifies:
- `.claude/skills/moirai/SKILL.md` -- routing table
- `.claude/skills/moirai/clotho.md` -- creation operations (2 ops, ~100 lines)
- `.claude/skills/moirai/lachesis.md` -- measurement operations (8 ops, ~200 lines)
- `.claude/skills/moirai/atropos.md` -- termination operations (3 ops, ~150 lines)

**Stakeholder decision**: Design review first. The TDD is probably correct but needs deep contextual review to:
- Understand the full architecture from an agent's perspective
- Ensure context-engineering best practices are applied
- Verify the progressive disclosure pattern achieves desired performance
- Confirm skill loading via Read (not Skill tool) is still the right mechanism

### The Consultant/Pythia Ecosystem

**Components**:
- `agents/consultant.md` -- Meta-level navigator, opus model, maxTurns: 20
- `mena/navigation/consult/INDEX.dro.md` -- Dromena (slash command)
- `mena/navigation/consult/reference.md` -- Skill reference (legomena)
- `.claude/knowledge/consultant/capability-index.yaml` -- Rite capability metadata
- `.claude/knowledge/consultant/rites/*.md` -- Rite-specific agent knowledge (ecosystem only)

**Missing**:
- `rite-discovery` skill source exists in `mena/` but is not materialized to `.claude/skills/` -- the consult dromena references it 6+ times expecting it to be loadable
- Playbook files (`~/.claude/knowledge/consultant/playbooks/curated/*.md`)
- Knowledge base coverage for non-ecosystem rites
- capability-index.yaml has stale command names (/10x-dev → /10x, /doc-team → /docs, /debt-triage → /debt)
- Consultant reference file references "Claude Opus 4.5" specifically -- should reference model generically

**Required for world-class**:
- Create rite-discovery skill (dynamic rite inventory)
- Fix/populate capability-index.yaml with current data
- Create or remove playbook references
- Full dynamic exploration via Glob/Grep/Read (not Task delegation)
- Ensure all 5 modes produce correct output
- Verify skill cross-references (prompting, 10x-workflow) resolve correctly

---

## 6. CC Platform Understanding

### CC Primitives Used by Knossos

| CC Primitive | Knossos Name | Mechanism |
|---|---|---|
| CLAUDE.md | Inscription | Always in context. ~740 tokens. |
| .claude/rules/ | (unnamed, needs mythological name) | Path-scoped instructions CC loads when editing matching files |
| .claude/commands/ | Dromena | Slash commands. User-invoked. `disable-model-invocation` prevents autonomous loading. |
| .claude/skills/ | Legomena | Skills. Model-invoked via Skill tool. Persistent in context. |
| .claude/agents/ | Heroes | Specialist agents. Invoked via Task tool. Cannot spawn sub-agents. |
| settings.local.json hooks | Hooks | Lifecycle event handlers. Go binary (ari) executes them. |
| settings.local.json mcpServers | MCP | External tool servers. Union merge preserves user config. |

### Key CC Constraints

1. **Agents cannot spawn agents**: CC silently strips Task tool from subagents. Design for single-level delegation.
2. **Hook output format**: PreToolUse must return `{hookSpecificOutput: {permissionDecision: "allow"|"deny"}}`. Flat JSON is silently ignored.
3. **disable-model-invocation**: Prevents CC from autonomously invoking a slash command as a skill. Essential for commands with side effects.
4. **Skills are persistent**: Once loaded via Skill tool, content stays in context. Design for minimal token footprint.
5. **Rules are path-scoped**: CC loads rules matching the file path being edited. Invisible to the user.
6. **Env vars don't persist**: Each Bash tool call gets a fresh shell. `export VAR=value` in one call is gone in the next.

---

## 7. Open Questions (For Future Resolution)

These were identified during the interview but deferred:

1. **Rules mythological name**: Rules are a CC primitive not yet in the Knossos lexicon. What mythological entity do they map to?
2. **Distribution mechanism**: Clone repo? Package? Homebrew? Deferred until product is ready.
3. **Hook shell→binary migration completeness**: Needs verification. Stakeholder says "not sure" but migration being complete is a prerequisite.
4. **Companion file autocomplete fix**: Is this a CC behavior we can control, or a naming convention we need to change? Needs investigation.
5. **Consultant knowledge base location**: Currently at `.claude/knowledge/consultant/` (projection). Should this be source-controlled? Materialized from source?
6. **MOIRAI_BYPASS mechanism**: Fundamentally can't work through CC Bash. Alternative approach needed for Moirai's write guard bypass.

**Standing directive**: When questions arise about White Sails, Naxos, Fates, or other doctrinal topics, the doctrine files are the canonical source of truth -- not interview summaries or paraphrases. Stakeholder explicitly directed: "Read the doctrine."

---

## 8. Anti-Patterns to Avoid

From the stakeholder and doctrine:

1. **Never edit .claude/ directly** -- Edit source (rites/, mena/, knossos/templates/) and rematerialize
2. **Never assume dead references should be deleted** -- Investigate history, categorize, batch, ask
3. **Never strip mythology** -- It's load-bearing architecture, not decoration
4. **Don't over-infer stakeholder preferences** -- When uncertain, ask. The cost of asking is low; the cost of wrong assumptions is high.
5. **Don't chase CC feature parity** -- Fix what breaks, protect what confuses, defer what optimizes
6. **Don't create backward-compatibility shims** -- Direct atomic commits to main. Change the code, don't wrap it.
7. **Don't treat .claude/ as a pure cache** -- It contains user state (satellite regions, user-agents, user-hooks). "Delete and regen" is an anti-pattern.

---

## 9. Success Criteria

### For Internal-First Stage

**From stakeholder**: A new internal team member should be able to "jump in, /consult for a while, understand the concepts via asking questions, and get started with a session" -- all within the first 30 minutes.

**Proposed success criteria** (derived, not directly from stakeholder -- treat as targets to validate):

1. **Within 5 minutes**: Clone, build ari, run `ari sync materialize`
2. **Within 15 minutes**: Start a session via `/start`, understand the basic model via `/consult`
3. **Within 30 minutes**: Grasp rites/sessions/agents/mena conceptually through asking questions via /consult
4. **Within 1 hour**: Complete a /task cycle on a real project
5. **Within 1 day**: Use multiple rites confidently, understand the mythology, create custom content

### For the Consultant (Pythia)

- `/consult` (no args): Produces accurate ecosystem overview with current state
- `/consult "I need to add auth"`: Routes correctly to 10x-dev + /task with command-flow
- `/consult --rite`: Shows all 10 rites with accurate descriptions and commands
- `/consult --commands`: Shows complete, categorized command reference
- Zero 404s -- every skill reference, knowledge base file, and cross-reference resolves
- Dynamic exploration works -- consultant can read rite manifests and agent files on demand

---

## 10. Hygiene Audit Decisions (2026-02-08)

> Captured during code-smeller assessment phase. These decisions bind the architect-enforcer and janitor.

### P0 Blockers

| SMELL | Decision | Notes |
|-------|----------|-------|
| SMELL-001: StagedMaterialize + cloneDir | **Delete entirely** | Remove function, helper, and tests. Clean break. |
| SMELL-002: Hook output format | **Standardize all on CC-native** | Migrate precompact to hookSpecificOutput envelope. Deprecate legacy Result type. |
| SMELL-003 + SMELL-004: Lock duplication | **Canonical in internal/lock/** | Expand internal/lock/ as single source of truth for all lock ops including Moirai lock reading and stale detection. |

### P1 Should-Fix

| SMELL | Decision | Notes |
|-------|----------|-------|
| SMELL-005: atomicWriteFile x3 | **Extract to internal/fileutil/** | New package with AtomicWriteFile and WriteIfChanged. All three packages import from here. |
| SMELL-006 + SMELL-007 + SMELL-009: Dead code | **Delete** | Remove materializeSettings(), getCurrentRite(), GetTemplatesDir(). |
| SMELL-008: Legacy Materialize() wrapper | **Delete entirely** | No external consumers. Remove alongside StagedMaterialize. |
| SMELL-010: Rotation + precompact | **Commit as-is** | Feature is complete and tested. Commit precompact.go, precompact_test.go, rotation.go, rotation_test.go, hooks.yaml. |

### P2 Included in Plan

| SMELL | Decision | Notes |
|-------|----------|-------|
| SMELL-011: CodeGeneralError overuse | **Include in refactor plan** | Add phase-specific error codes for runtime distinguishability. |
| SMELL-013: Scope infrastructure | **Remove and rebuild later** | YAGNI. Remove ~60 lines. Rebuild when actual use cases drive design. |
| SMELL-014: Provenance strategies | **Include in architect plan** | Have architect-enforcer design unified provenance tracking approach. |
| SMELL-016: Legacy templates | **Delete entirely** | Zero callers confirmed. |
| SMELL-017: ParseLegacyMarkers | **Delete entirely** | Migration complete. Remove function and tests. |
| SMELL-019: ritesDir deprecated field | **Complete migration** | Migrate ritesDir usages to sourceResolver, then remove the field. |

### Not Changing

| SMELL | Decision | Notes |
|-------|----------|-------|
| SMELL-012: Direct os.WriteFile | Use atomic write for Context.Save() at minimum. Others acceptable. |
| SMELL-015: containsStr | Not dead code (test utility). No action. |
| SMELL-018: copyDir | Not dead code (one caller). No action. |
| SMELL-020: Silent frontmatter parse | Intentional (EC-7). Add warning log only. |
| SMELL-021: fileExists duplication | Trivial, low priority. |
| SMELL-022: ValidScope uncalled | Will be removed with SMELL-013 scope infrastructure. |

---

## Appendix: Interview Transcript Summary

| Phase | Topic | Key Decision |
|-------|-------|-------------|
| 1 | Vision & Audience | Staged rollout, progressive readiness bar, quality+iteration balance, mythology is load-bearing |
| 2 | Uncommitted Work | Atomic commits to main, Fates concept evolved (mena/lego intended), categorize+batch dead refs, hook migration needs verification |
| 3 | Onboarding & Discoverability | Jump in + /consult, autocomplete must fix, /consult is primary user entry, layered mythology teaching |
| 4 | Technical Decisions | DMI for multi-agent only, consultant Task is bug, auto-archive in /wrap, binary stays in git |
| 5 | Architecture & Doctrine | Read White Sails/Naxos/Fates doctrine, hook Result is legacy, rules are for CC context, full catalog ships |
| 6 | Distribution Scope | Consult must be world-class, Fates need design review, distribution mechanism deferred, all 10 rites ship |
| 7 | Work & Process | Foundations first then Pythia, full dynamic exploration for consultant, fix bugs/ask on judgment, comprehensive preferences doc |

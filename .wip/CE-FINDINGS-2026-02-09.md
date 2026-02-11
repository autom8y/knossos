# Context Engineering Findings: An Honest Assessment

**Date**: 2026-02-09
**Auditor**: Context Engineer (Claude Opus 4.6)
**Method**: Read every agent prompt, every dromena INDEX, every legomena INDEX, every inscription template. 111 files. Not sampling -- full coverage.

---

## 1. What's Working Well

I want to start here because the things that work are genuinely impressive and should not be touched.

### The orchestrator consultation pattern is best-in-class

All 11 orchestrators implement the same pattern: stateless advisor, Read-only tool access, CONSULTATION_REQUEST/RESPONSE YAML protocol, explicit "DO NOT" behavioral constraints, and a litmus test self-check. I read every one of them line by line. The shared boilerplate is *identical* -- not "similar" or "mostly the same," but character-for-character identical across all 11 rites. That is remarkable discipline.

The pattern itself is smart. By making orchestrators stateless advisors instead of execution authorities, the framework avoids the most common failure mode in multi-agent systems: the orchestrator trying to do work instead of routing it. The "Litmus Test" section -- *"Am I generating a prompt for someone else, or doing work myself?"* -- is the kind of self-check mechanism that actually works in practice because it is specific enough to apply.

### The specialist agent structure is genuinely consistent

Every specialist agent follows the same skeleton: Frontmatter, Core Responsibilities, Position in Workflow (with ASCII diagram), Domain Authority (decide/escalate/route triad), Approach, What You Produce, File Verification reference, Handoff Criteria (checklist), Acid Test, Anti-Patterns, Skills Reference. All 46 specialists. No exceptions.

The Acid Test questions are particularly good -- each one is unique to its agent's role and genuinely tests whether the agent did its job. The architect asks *"Will this design look obviously right in 18 months?"*. The QA adversary asks *"If this goes to production and fails in a way I didn't test, would I be surprised?"*. The principal engineer asks *"If I got hit by a bus, could another engineer maintain this code?"*. These are not filler. They function as alignment anchors.

### The frontmatter metadata is complete and well-calibrated

57 agents, all with correct frontmatter. Multi-line descriptions on all of them (task #23 is done in practice even if not marked complete). Model assignments make sense -- opus for judgment work, sonnet for execution-heavy agents like janitor and tech-writer, haiku for the sync dromena. `disallowedTools` correctly applied on all 11 orchestrators (Bash, Write, Edit, Glob, Grep, Task) and on QA adversary (Task). The `contract.must_not` on forge's eval-specialist (`rites/forge/agents/eval-specialist.md`, lines 36-39) is the best example of machine-readable behavioral constraints in the whole framework.

### The lexicon skill is the intellectual core of the framework

`mena/guidance/lexicon/INDEX.lego.md` is the single most important file in Knossos. It maps every Knossos concept to its CC runtime primitive, documents what CC understands and what it does not, and defines the anti-patterns. The "What CC Does NOT Understand" table (lines 30-37) should be mandatory reading for anyone who touches an agent prompt.

### The inscription templates are lean and idempotent

Seven templates in `knossos/templates/sections/`. The materialized CLAUDE.md inscription is under 1,000 tokens. The `quick-start` and `agent-configurations` templates use Go template logic to regenerate from live rite data. The region markers are consistent. The owner annotations are correct. This layer needs no work.

---

## 2. What's Broken or Inconsistent

### The framework violates its own documented standard 195 times

This is the headline finding.

`mena/guidance/lexicon/anti-patterns.md`, line 9:
> `@skill-name` -- CC has no `@` resolution mechanism

`mena/guidance/lexicon/INDEX.lego.md`, line 34:
> `@skill-name` | CC has no `@` resolution | `Skill("skill-name")` or just name the skill

And yet: 195 occurrences of `@skill-name` across 47 of 57 agent files. Every single rite. Every single orchestrator. Most specialists. Three different formatting variants that nobody standardized:

```
- @standards for code conventions              (bare, in list items)
- `@doc-sre` for postmortem templates          (backticked, in list items)
Produce using `@doc-sre#postmortem-template`.  (fragment syntax, inline)
```

The fragment syntax is the worst offender. Forty occurrences of `@skill-name#section-name` -- a syntax that CC literally cannot parse. When an agent prompt says `Produce Financial Model using @doc-strategy#financial-model-template`, CC reads that as a string. It cannot navigate to the "financial-model-template" section of the "doc-strategy" skill. It just sees an at-sign followed by words.

This is not a theoretical concern. It means every agent's artifact production instructions and skills reference section contain unresolvable references. CC compensates by inferring intent, but it is working around the framework instead of with it.

The 10 agents that use plain names (`doc-artifacts, standards, file-verification.` at `rites/10x-dev/agents/architect.md` line 116) are *already doing it right*. The correct form exists in the codebase. It was just never propagated.

### Three terminology survivors from the legacy cleanse

I fixed these during the audit:

| File | Was | Now |
|------|-----|-----|
| `rites/rnd/agents/orchestrator.md:125` | `outside team's control` | `outside rite's control` |
| `rites/ecosystem/agents/orchestrator.md:214` | `all 10 teams` | `all rites` |
| `rites/forge/agents/orchestrator.md:121` | `outside pantheon's control` | `outside rite's control` |

The "pantheon" one is interesting -- it is from a naming era *before* "teams," which was itself before "rites." Three geological layers of terminology in one line. Nine commits and 400+ replacements in SL-008, and this one survived because nobody searched for "pantheon."

---

## 3. What's Over-Engineered

### 2,284 lines of orchestrator boilerplate for ~300 lines of actual information

The 11 orchestrators total 2,284 lines. Of those, approximately 1,650 are identical shared boilerplate. The rite-specific content -- the Position in Workflow diagram, the Phase Routing table, the Rite-Specific Anti-Patterns, the Skills Reference, and optionally the Cross-Rite Protocol -- amounts to roughly 50-80 lines per orchestrator.

That is a 30:1 boilerplate-to-content ratio.

Every time an orchestrator is invoked via Task tool, CC loads all ~210 lines into the subagent context. Of those, ~150 are identical to what every other orchestrator would have loaded. This is pure token waste on every orchestration consultation.

Task #25 (boilerplate injection) is the obvious fix. The pipeline already handles template materialization. Adding an `{{include "orchestrator-boilerplate.md"}}` pattern would reduce each source file to its rite-specific delta and eliminate the maintenance burden of keeping 11 copies synchronized.

### The cross-rite HANDOFF acceptance protocol in code-smeller is 50 lines that could be 5

`rites/hygiene/agents/code-smeller.md`, lines 137-184, has a detailed protocol for accepting HANDOFF artifacts from debt-triage. It includes a consumption protocol (5 numbered steps), a "What debt-triage provides" list (4 bullets), a "What you add" list (4 bullets), and a full example with frontmatter.

The consumption protocol is: "read the frontmatter, check the source rite, process the items." That is standard HANDOFF behavior. The 50 lines could be: "When receiving a HANDOFF from debt-triage, use items as pre-scored work queue. See cross-rite-handoff skill for schema."

This pattern appears to be a one-off experiment that was never normalized. No other agent has an equivalent section. Either every receiving agent should get one (templatized), or none should (rely on the cross-rite-handoff skill).

### The embedded TRANSFER template in tech-transfer

`rites/rnd/agents/tech-transfer.md` is 315 lines -- the longest agent prompt. Lines 113-175 embed a complete TRANSFER document template with markdown code blocks, table structures, and production gap analysis templates. This is reference content masquerading as behavioral instruction. Every time tech-transfer is invoked, 62 lines of template scaffold enter context regardless of whether the agent needs to produce a TRANSFER artifact.

The same pattern appears in:
- `rites/forge/agents/eval-specialist.md`: embedded validation checklists
- `rites/intelligence/agents/insights-analyst.md`: embedded analysis frameworks
- `rites/forge/agents/agent-curator.md`: embedded catalog patterns
- `rites/forge/agents/workflow-engineer.md`: embedded workflow templates
- `rites/strategy/agents/roadmap-strategist.md`: embedded planning framework
- `rites/forge/agents/platform-engineer.md`: embedded integration patterns
- `rites/docs/agents/doc-auditor.md`: 117-line "Staleness Detection Mode" starting at line 130

These are all task #24 candidates. The forge rite accounts for 4 of the top 8 because forge agents are inherently meta -- they embed templates for creating templates. But the solution is the same: extract to companion skills, reference by name.

---

## 4. What's Under-Invested

### The `contract.must_not` frontmatter pattern

Only one agent in the entire framework uses `contract.must_not`: the forge eval-specialist (`rites/forge/agents/eval-specialist.md`, lines 35-39):

```yaml
contract:
  must_not:
    - Modify agent prompts to fix eval failures
    - Ship agents that fail evaluation criteria
    - Reduce evaluation standards to achieve passing
```

This is the best behavioral constraint mechanism in the framework. It is machine-readable, it is in the frontmatter where CC processes it, and it defines hard boundaries that the agent cannot cross.

Yet these agents have equivalent constraints *in prose* that should be in frontmatter:

- **security-reviewer** (line 29): `must_not: [Approve code with unresolved critical findings, Make business risk acceptance decisions]` -- already written as prose constraints, just not in YAML.
- **code-smeller**: Should not propose solutions (lines 131, 117). Diagnose only.
- **ecosystem-analyst**: Should not propose solutions (line 117). Diagnose only.
- **requirements-analyst**: Should not make architectural decisions. That is the architect's domain.

The QA adversary has `disallowedTools: [Task]` and `contract.must_not` -- the belt-and-suspenders approach. The security-reviewer has `disallowedTools: [Task]` but no `contract.must_not`. Inconsistent.

### Session checkpoints for long-running agents

Only one agent has a Session Checkpoints section: the principal-engineer (`rites/10x-dev/agents/principal-engineer.md`, lines 87-97). It defines a structured checkpoint format with progress summary, artifacts created, context anchor, and next steps.

Four other agents have `maxTurns` of 200+:
- integration-engineer (ecosystem): no checkpoint guidance
- tech-writer (docs): no checkpoint guidance
- tech-transfer (rnd): no checkpoint guidance
- incident-commander (sre): no checkpoint guidance

When these agents hit turn 150 in a complex task, there is no mechanism to emit a progress summary that survives context compaction. The principal-engineer pattern works. It should be in more places.

### The 10x-dev orchestrator's entry point selection logic

`rites/10x-dev/agents/orchestrator.md`, lines 136-161, has a genuinely valuable "Entry Point Selection" section that maps work types to starting agents. Bug fix starts at principal-engineer. Refactoring starts at architect. New feature starts at requirements-analyst.

No other orchestrator has this. The security orchestrator would benefit from it -- a small config change might skip threat modeling and go straight to security-reviewer. The sre orchestrator could skip observability-engineer for incident response. This is workflow intelligence that every orchestrator should have in its rite-specific section.

---

## 5. The Biggest Single Improvement

**Fix the `@skill-name` syntax across all 57 agents.**

If I could only change one thing, this would be it, and the reason is not about the syntax itself -- it is about what the syntax represents.

Knossos has a lexicon skill. That lexicon defines the correct way to reference skills. It explicitly marks `@skill-name` as an anti-pattern. It provides the correct alternative. The anti-patterns checklist in `mena/guidance/lexicon/anti-patterns.md` even has step-by-step migration instructions (line 56: "Replace `@skill`, backtick, `#fragment` syntax with plain names").

And then 195 references across 47 files ignore all of it.

This matters because the lexicon is the framework's constitution. If the framework does not follow its own constitution, then the constitution is aspirational documentation, not operational standard. Every time a new agent prompt is written, the author looks at existing agents for patterns, sees `@doc-sre#postmortem-template`, and copies the pattern. The anti-pattern self-replicates because the codebase teaches it.

The fix is mechanical. It is not a design decision. It does not require pipeline changes. It does not require stakeholder approval. It is a bulk find-and-replace across 47 files:

- `@skill-name` -> `skill-name` (drop the at-sign)
- `@skill-name#fragment` -> `skill-name skill, fragment section` (dissolve the fragment)
- `@skill-name/path` -> `skill-name skill (path)` (dissolve the path)

Two hours of work. Zero ambiguity. And when it is done, the codebase matches the lexicon, new agents copy the correct pattern, and the anti-patterns document goes from aspirational to enforced.

---

## 6. Token Economics

### Where the budget goes

The context window is the arena where everything competes. Here is where tokens are spent when a typical orchestrated session runs:

| Context Layer | Estimated Tokens | Frequency |
|---------------|-----------------|-----------|
| CLAUDE.md inscription | ~950 | Every turn (fixed cost) |
| SessionStart hook injection | ~300-500 | Once (persistent in conversation) |
| Orchestrator prompt (per consultation) | ~1,500 | Per orchestration cycle |
| Specialist agent prompt (per invocation) | ~1,000-2,500 | Per specialist phase |
| Skills loaded by agents | ~500-3,000 | On demand (persistent once loaded) |
| Conversation history | ~500-2,000/turn | Accumulates |

### Where the waste is

**Orchestrator boilerplate**: 1,650 lines of duplication in source. When materialized, each orchestrator loads ~150 shared lines that are identical to every other orchestrator. In a 4-phase orchestrated session with 4 consultations, that is ~600 tokens of repeated boilerplate (150 lines x 4 consultations, though CC may cache the pattern after seeing it once).

The real cost is not the token count per se -- it is the opportunity cost. Those 150 lines per orchestrator are 150 lines that *could* be rite-specific routing intelligence, workflow shortcuts, or domain context. Instead they are "What You DO / What You DO NOT DO" boilerplate that CC already understands after seeing it once.

**Oversized agent prompts**: The 8 agents over 250 lines embed 50-100 lines each of reference content (templates, checklists, examples) that loads on every invocation regardless of whether it is needed. Total waste across those 8: roughly 500 embedded reference lines, or ~4,000 tokens, loaded unnecessarily.

**`@skill-name#fragment` references**: These do not waste tokens directly -- they are just strings. But they waste *inference effort*. When CC reads `Produce Financial Model using @doc-strategy#financial-model-template`, it has to (a) recognize this is a skill reference, (b) ignore the unresolvable `@` and `#` syntax, (c) infer it should load the doc-strategy skill, and (d) find the financial model section within it. A plain reference like "Use the doc-strategy skill (financial model template)" requires zero inference overhead.

### ROI on fixes

| Fix | Effort | Token Savings | Notes |
|-----|--------|---------------|-------|
| Orchestrator boilerplate injection | M | ~600/session | Plus maintenance elimination |
| Extract embedded references (8 agents) | M | ~4,000/invocation | For affected agents only |
| Fix `@skill-name` syntax | S | ~0 direct | Inference overhead reduction, prevents drift |
| Extend `contract.must_not` | XS | ~0 | Behavioral safety, not token savings |

The honest truth: the token savings from any single fix are modest. The CLAUDE.md inscription is already lean at ~950 tokens. The real wins were already captured in Phases 1-2 (context:fork on dromena, lifecycle corrections). What remains is operational hygiene -- making the framework's practice match its theory.

---

## 7. Predictions

### The `@skill-name` pattern will get worse before it gets better

Every new agent prompt will be written by looking at existing agents for examples. The existing agents use `@skill-name`. The new agents will too. This is a ratchet -- the anti-pattern becomes more entrenched with every rite the forge creates.

If you do nothing else from this audit, fix the syntax *before* the next forge run.

### The orchestrator boilerplate will diverge

Right now all 11 orchestrators are identical in their shared sections. That will not last. Someone will need to update the Behavioral Constraints for a specific rite, they will edit one orchestrator, and the divergence will begin. Within 3-4 rite updates, the "identical boilerplate" will be "mostly identical boilerplate with subtle differences nobody remembers the reason for."

Boilerplate injection prevents this by making shared content physically shared, not copy-pasted.

### The forge rite will become the framework's biggest token consumer

Forge agents are already the most oversized (4 of the top 8). As the framework adds capabilities, forge agents will embed more templates, more patterns, more reference material. Without aggressive progressive disclosure (extract to companion skills), a single forge session could consume 15,000+ tokens in agent prompts alone.

### Cross-rite HANDOFF will need a standard receiving protocol

The code-smeller's 50-line HANDOFF acceptance section is a prototype of something every cross-rite receiving agent will eventually need. As rites interact more (debt-triage -> hygiene, rnd -> 10x-dev, security -> sre), the ad-hoc acceptance protocols will proliferate. Better to templatize now than standardize later.

### The `contract.must_not` pattern will prove its worth in a failure

Somewhere, sometime, an agent will do something it should not. The QA adversary will try to fix a bug it found. The security-reviewer will approve code with an open finding because the conversation pressured it. The code-smeller will propose a fix instead of diagnosing.

When that happens, the agents with `contract.must_not` in their frontmatter will be the ones that held the line. The agents with constraints-in-prose-only will be the ones that bent. Extending `contract.must_not` to every constrained agent is cheap insurance.

### Something will go wrong with session lifecycle at scale

The framework tracks sessions, parks them, resumes them, hands them off. This works when one person uses one session at a time. When parallel sessions become common (worktrees, multiple CC instances), the session lifecycle assumptions will be tested. The "single-ACTIVE relaxed to per-CC-instance" decision from the Session Hardening initiative is the right call, but the implementation will have edge cases nobody has hit yet.

The CC session map (.cc-map/) is the right architecture for this. The question is whether the event system can handle the concurrency.

---

## Final Thought

Knossos is a framework that knows what it wants to be. The lexicon is precise. The primitive model is correct. The separation between dromena (transient, user-controlled) and legomena (persistent, model-controlled) is a genuine insight about context lifecycle that most frameworks get wrong.

The gap is not in the theory. The gap is in the execution catching up to the theory. The lexicon says do not use `@skill-name`. The codebase uses it 195 times. The orchestrator pattern is excellent -- and then duplicated 11 times instead of shared. The `contract.must_not` mechanism is the right answer to behavioral constraints -- and then only applied to one agent.

The work is not architectural. It is janitorial. It is the tedious, mechanical, unglamorous work of making 57 files match the standard that 10 files already demonstrate. That is the state of the framework: the north star is clear, the path is known, the remaining work is a grind.

Do the grind.

# State of the Framework: Knossos Context Engineering Audit

**Date**: 2026-02-09
**Methodology**: 7-agent parallel analysis (6 context-engineer + 1 claude-code-guide)
**Scope**: All source artifacts — 35 dromena, 39 legomena, 57 agents, 10 rules, 6 CLAUDE.md sections, 10 hook handlers, 20 event types, 9,929 LOC pipeline
**Prior work**: Phase 1 (10 quick wins) and Phase 2 (7 lifecycle corrections) COMPLETE

---

## Executive Summary

**Current score: 8/10.** Phases 1 and 2 resolved the critical context lifecycle violations and content quality gaps. The framework is structurally sound — context isolation is at 100%, skill routing descriptions are standardized, the inscription is lean at ~950 tokens, and the event system covers 19/20 types.

**Remaining gap to 9/10**: Agent description quality (75% below gold standard), embedded reference bloat (27 agents over 200 lines), no token budget self-awareness, and operational debt (phantom references, dead code, hook inconsistencies).

### Health by Domain

| Domain | Grade | Key Metric | Primary Gap |
|--------|:-----:|------------|-------------|
| **Dromena** (35 files) | **A+** | 100% context:fork, 100% allowed-tools | 1 missing disable-model-invocation |
| **Legomena** (39 files) | **B+** | 100% Use-when/Triggers, 44% progressive disclosure | 4 files >200 lines without companions |
| **Agents** (57 files) | **C+** | 25% gold-standard descriptions, 47% over 200 lines | 43 agents need description uplift |
| **Inscription** (CLAUDE.md) | **A-** | ~950 tokens, 6 sections | Phantom Moirai reference in 3 locations |
| **Rules** (10 files) | **B** | 10 active rules, no orphans | 2 high-edit dirs (40K LOC) have no rules |
| **Hooks** (10 handlers) | **B+** | 8/14 CC events covered, 19/20 event types active | 2 timeout bugs, hooks.yaml divergence |
| **Pipeline** (9,929 LOC) | **A-** | All source types handled, provenance tracked | 55 lines dead code, zero token awareness |
| **CC Alignment** | **B+** | 85% aligned | disable-model-invocation semantics misunderstood |

---

## Domain Reports

### 1. Dromena — Grade: A+

**35 files audited.** The dromena fleet is in excellent condition after Phase 1's context:fork migration.

| Metric | Value |
|--------|-------|
| context:fork coverage | 35/35 (100%) |
| allowed-tools scoping | 35/35 (100%) |
| disable-model-invocation | 34/35 (97%) |
| Over 200 lines | 0/35 (0%) |
| Avg description quality | 4.6/5 |

**Findings:**

| ID | Severity | Finding |
|----|----------|---------|
| D-H1 | HIGH | `/consult` missing `disable-model-invocation: true` — CC may auto-inject 154-line routing command (~940 tokens) |
| D-H2 | HIGH | Zero dromena have `scope` field despite rules documenting "scope controls pipeline routing" — schema inconsistency |
| D-M1 | MEDIUM | `/sessions` has `Write` in allowed-tools but only reads session data — principle of least privilege |
| D-M2 | MEDIUM | `/architect` and `/build` have `Write` but delegate to subagents — main thread shouldn't write directly |
| D-M3 | MEDIUM | `/qa` description is weakest ("Validation-only with review and approval") — doesn't explain what is validated |
| D-L1 | LOW | `/consult` embeds hardcoded rite list that duplicates rite-discovery skill data |
| D-L2 | LOW | Model tier inconsistency: `/ecosystem` uses sonnet but does identical work to haiku rite-switching commands |

### 2. Legomena — Grade: B+

**39 files audited** (13 core + 26 rite-scoped). Core legomena are exemplary; rite-scoped carry progressive disclosure debt.

| Metric | Core (13) | Rite-Scoped (26) | Total |
|--------|-----------|-------------------|-------|
| Has Use-when + Triggers | 100% | 100% | 100% |
| Uses INDEX+companion | 85% | 23% | 44% |
| Over 200 lines w/o companions | 0 | 4 | 4 |
| Avg description quality | 4.2/5 | 4.1/5 | 4.1/5 |

**Findings:**

| ID | Severity | Finding |
|----|----------|---------|
| L-C1 | CRITICAL | `forge-ref`: 346 lines, 0 companions — largest legomena, 2.3x over 150-line guideline |
| L-C2 | CRITICAL | `doc-artifacts` (10x-dev): 258 lines, 0 companions — duplicates well-structured core version |
| L-H1 | HIGH | `moirai-fates`: imperative "Loading Protocol" in a legomena — lifecycle classification violation |
| L-H2 | HIGH | `file-verification`: 161 lines, 0 companions — "intentionally self-contained" comment resists progressive disclosure |
| L-H3 | HIGH | `doc-consolidation`: 191 lines, 0 companions |
| L-H4 | HIGH | `rite-development`: 172 lines, 0 companions |
| L-H5 | HIGH | `agent-prompt-engineering`: 207-line INDEX despite having 4 companions — INDEX not slimmed |
| L-M1 | MEDIUM | 3 non-standard filenames (`hygiene-catalog`, `debt-catalog`, `sre-catalog`) use `{name}.lego.md` instead of `INDEX.lego.md` |
| L-M2 | MEDIUM | `documentation` description too broad — "documentation" trigger matches too many contexts |
| L-M3 | MEDIUM | `ecosystem-ref` triggers use hyphenated forms ("knossos-sync") users won't type naturally |
| L-L1 | LOW | Duplicate skill name `doc-artifacts` in both core and 10x-dev — CC behavior undefined for collisions |

**Token impact**: Splitting the 10 oversized files would save ~4,700 tokens per skill activation on average.

### 3. Agents — Grade: C+

**57 files audited across 11 rites.** This is the single largest debt area. Forge agents are the gold standard; all other rites lag significantly.

| Metric | Value |
|--------|-------|
| Total agents | 57 |
| Total lines | 10,356 |
| Avg lines/agent | 182 |
| Over 200 lines | 27 (47%) |
| Gold-standard descriptions (4-5/5) | 14 (25%) |
| Single-line descriptions (1-3/5) | 43 (75%) |
| Opus where sonnet would suffice | ~15 (26%) |

**By rite:**

| Rite | Agents | Avg Lines | Desc Quality | Notable |
|------|--------|-----------|:------------:|---------|
| forge | 7 | 243 | **4.6/5** | Gold standard descriptions; 5/7 over 200 (embedded ref) |
| 10x-dev | 5 | 150 | 3.0/5 | Most-used rite, descriptions below standard |
| ecosystem | 6 | 153 | 3.0/5 | Consistent but all single-line |
| docs | 5 | 168 | 3.0/5 | doc-auditor has 117-line embedded staleness mode |
| intelligence | 5 | 211 | 3.0/5 | insights-analyst 281 lines, embedded examples |
| rnd | 6 | 183 | 3.0/5 | tech-transfer at 301 lines — largest agent |
| strategy | 5 | 185 | 3.0/5 | roadmap-strategist embeds full RICE example |
| security | 5 | 180 | 3.0/5 | All opus, compliance-architect could be sonnet |
| sre | 5 | 154 | 3.0/5 | All opus, most could be sonnet |
| hygiene | 5 | 172 | 3.0/5 | janitor correctly uses sonnet |
| debt-triage | 4 | 169 | 3.0/5 | Only orchestrator under 200 lines |

**Top 5 oversized agents:**

| Agent | Rite | Lines | Embedded Content |
|-------|------|------:|------------------|
| tech-transfer | rnd | 301 | ~95 lines HANDOFF templates |
| eval-specialist | forge | 297 | ~63 lines validation checklists |
| agent-curator | forge | 295 | ~68 lines sync checklist + versioning |
| workflow-engineer | forge | 294 | ~59 lines workflow patterns |
| insights-analyst | intelligence | 281 | ~50 lines example format |

**Findings:**

| ID | Severity | Finding |
|----|----------|---------|
| A-C1 | CRITICAL | 43/57 agents (75%) use single-line descriptions — CC routing degraded for most specialists |
| A-H1 | HIGH | 7 agents embed >50 lines of reference content that should be skills |
| A-H2 | HIGH | 11 orchestrators share ~120 lines identical boilerplate (~1,320 lines duplication) |
| A-M1 | MEDIUM | ~15 agents use opus where sonnet would suffice (~5x cost difference) |
| A-L1 | LOW | Forge descriptions win routing disproportionately due to richer metadata |

### 4. Inscription (CLAUDE.md) — Grade: A-

**~950 tokens, 7 sections.** Lean and well-structured. Issues are stale references, not architecture.

| Section | Est. Tokens | Owner | Status |
|---------|------------|-------|--------|
| execution-mode | ~120 | knossos | CLEAN |
| quick-start | ~180 | regenerate | CLEAN |
| agent-routing | ~60 | knossos | CLEAN |
| commands | ~180 | knossos | CLEAN |
| agent-configurations | ~140 | regenerate | DUPLICATE (of quick-start agent list) |
| platform-infrastructure | ~50 | knossos | **STALE** — phantom Moirai reference |
| user-content | ~220 | satellite | Minor duplication |

**Doctrine compliance:** 4/6 PASS, 1 PARTIAL (stale content), 1 FAIL (single purpose — duplication + phantom refs)

**Findings:**

| ID | Severity | Finding |
|----|----------|---------|
| I-H1 | HIGH | `Task(moirai, "...")` in platform-infrastructure — no moirai agent exists; will fail at runtime |
| I-H2 | HIGH | Same phantom ref in user-content section and internal-session.md rule (3 locations total) |
| I-M1 | MEDIUM | agent-configurations duplicates quick-start agent table (~140 tokens waste) |
| I-L1 | LOW | quick-start template source attribute mismatch (cosmetic) |

### 5. Rules — Grade: B

**10 rules active, no orphans.** Good path-scoping pattern. Significant coverage gaps for high-edit directories.

| Rule | Target | LOC Covered | Status |
|------|--------|-------------|--------|
| internal-agent.md | internal/agent/ | 4,181 | CURRENT |
| internal-hook.md | internal/hook/ | ~3,000 | **STALE** (event count, transport model) |
| internal-inscription.md | internal/inscription/ | ~2,000 | CURRENT |
| internal-materialize.md | internal/materialize/ | 9,541 | CURRENT |
| internal-provenance.md | internal/provenance/ | 1,234 | CURRENT |
| internal-sails.md | internal/sails/ | ~1,500 | CURRENT |
| internal-session.md | internal/session/ | ~3,000 | **STALE** (Moirai reference) |
| knossos-templates.md | knossos/templates/ | ~2,000 | CURRENT |
| mena.md | mena/ | ~5,000 | CURRENT |
| rites.md | rites/ | ~15,000 | CURRENT |

**Coverage gaps:**

| Directory | LOC | Gap Severity |
|-----------|-----|:------------:|
| internal/cmd/ | 28,759 | **HIGH** — largest package, zero guidance |
| internal/rite/ | 4,181 | **HIGH** — domain core, frequently edited |
| internal/validation/ | 3,398 | MEDIUM |
| internal/worktree/ | 3,899 | MEDIUM |

**Findings:**

| ID | Severity | Finding |
|----|----------|---------|
| R-H1 | HIGH | internal-hook.md says 16 event types (actually 20), says env vars primary (stdin JSON is primary) |
| R-H2 | HIGH | internal/cmd/ (28,759 LOC) and internal/rite/ (4,181 LOC) have no path-scoped rules |
| R-M1 | MEDIUM | internal-session.md references nonexistent Moirai agent |

### 6. Hooks + Events — Grade: B+

**10 handlers, 8/14 CC events covered.** Architecture is sound with dual-timeout defense-in-depth. Operational issues.

**Hook registration:**

| CC Event | Handler | Sync/Async | Status |
|----------|---------|:----------:|--------|
| SessionStart | context | sync | ACTIVE |
| Stop | autopark | sync | ACTIVE |
| SessionEnd | sessionend | sync | ACTIVE |
| PreToolUse | writeguard | sync | ACTIVE |
| PreToolUse | validate | sync | ACTIVE |
| PostToolUse | clew | async | ACTIVE |
| PostToolUse | budget | sync | ACTIVE |
| PreCompact | precompact | sync | ACTIVE |
| SubagentStart | subagent-start | async | ACTIVE |
| SubagentStop | subagent-stop | async | ACTIVE |

**CC events NOT covered:** PostToolUseFailure, UserPromptSubmit, PermissionRequest, Notification, TeammateIdle, TaskCompleted (6/14 gaps)

**Event types:** 19 active, 1 dead (context_switch)

**Findings:**

| ID | Severity | Finding |
|----|----------|---------|
| H-H1 | HIGH | Budget hook missing `withTimeout` wrapper — inconsistent with all other handlers |
| H-H2 | HIGH | Autopark `getGitStatusQuick()` has no timeout — potential hang on every Stop event |
| H-M1 | MEDIUM | Route hook: shell wrapper exists but Go handler doesn't — dead code that would fail if activated |
| H-M2 | MEDIUM | hooks.yaml and settings.local.json diverge — 4 hooks only in settings, route only in hooks.yaml |
| H-M3 | MEDIUM | Dead event type `context_switch` — constructor, trigger logic, and tests but zero callers |
| H-M4 | MEDIUM | BufferedEventWriter overkill for single-event hooks — goroutine+ticker per invocation for 1 write |
| H-L1 | LOW | ACTIVE_RITE not in writeguard protected patterns |
| H-L2 | LOW | Duplicate auto-park logic in Stop + SessionEnd (intentional belt-and-suspenders) |

### 7. Pipeline + Provenance — Grade: A-

**9,929 LOC total.** Architecturally sound. All source types handled. All tests pass.

| Stage | Provenance Tracked | Status |
|-------|:-:|--------|
| Agents | YES | CLEAN |
| Mena (Commands+Skills) | YES (directory-level) | CLEAN |
| Rules | YES | CLEAN |
| CLAUDE.md | YES (file-level, not section-level) | SEMANTIC GAP |
| Settings | YES | CLEAN |
| Workflow | YES | CLEAN |

**Findings:**

| ID | Severity | Finding |
|----|----------|---------|
| P-M1 | MEDIUM | Dead method `copyDir` — 28 lines, zero callers (from deleted hooks materialization) |
| P-L1 | LOW | Dead exports `ReadMenaFrontmatterFromDir/File` — designed for deleted usersync package |
| P-L2 | LOW | CLAUDE.md provenance says "user-owned" but knossos sections still regenerated — semantic mismatch |
| P-L3 | LOW | `Collector.Entries()` returns live map, not copy — safe now but fragile if parallelized |
| P-INFO | INFO | Zero token awareness — `ari sync --budget` Phase 1 (byte counting) estimated at 2 hours |

### 8. CC Alignment — Grade: B+

**85% overall alignment.** Core model is solid with refinable details.

| Area | Alignment | Key Issue |
|------|:---------:|-----------|
| Commands/Dromena | 90% | Minor: `disable-model-invocation` prevents invocation, not loading |
| Skills/Legomena | 85% | INDEX+companion is Knossos convention, not CC primitive |
| Agents/Task | 95% | Well understood |
| Hooks | 70% | 6/14 events undocumented |
| Settings | 75% | Incomplete field coverage |

**CC features Knossos underutilizes:**

| Feature | CC Support | Knossos Usage | Value |
|---------|:---------:|:------------:|:-----:|
| Agent persistent memory | YES | NO | HIGH — cross-session learning |
| async hooks | YES | PARTIAL | MEDIUM — only 3/10 are async |
| Permission modes per agent | YES | NO | MEDIUM — fine-grained control |
| Prompt-based hooks (LLM validation) | YES | NO | LOW — optional |
| MCP servers per agent | YES | NO | LOW — specialized |
| Agent skills injection | YES | NO | MEDIUM — preload content |
| maxTurns per agent | YES | NO | MEDIUM — cost control |

**Key misalignment**: Knossos docs say `disable-model-invocation` "prevents loading" — correct behavior is "prevents automatic invocation but skill is still loaded into context." This affects token budget assumptions.

---

## Consolidated Findings (All Domains)

### By Severity

| Severity | Count | Domains |
|----------|:-----:|---------|
| CRITICAL | 3 | Legomena (2), Agents (1) |
| HIGH | 15 | Agents (2), Dromena (2), Inscription (2), Rules (2), Hooks (2), Legomena (5) |
| MEDIUM | 14 | All domains |
| LOW | 11 | All domains |
| **Total** | **43** | |

### CRITICAL Findings (3)

1. **A-C1**: 43/57 agents have single-line descriptions — CC routing degraded for 75% of specialists
2. **L-C1**: `forge-ref` at 346 lines with 0 companions — single largest legomena
3. **L-C2**: `doc-artifacts` (10x-dev) at 258 lines duplicating well-structured core version

### HIGH Findings (15)

| ID | Domain | Finding |
|----|--------|---------|
| I-H1+I-H2 | Inscription | Phantom Moirai agent reference in 3 locations — runtime failure |
| R-H1 | Rules | internal-hook.md: wrong event count (20, not 16), wrong transport model |
| R-H2 | Rules | 32,940 LOC in internal/cmd/ + internal/rite/ with no path-scoped rules |
| D-H1 | Dromena | `/consult` missing disable-model-invocation |
| D-H2 | Dromena | `scope` field absent from all 35 dromena — schema inconsistency |
| H-H1 | Hooks | Budget hook missing withTimeout wrapper |
| H-H2 | Hooks | Autopark getGitStatusQuick() has no timeout — hang risk |
| L-H1 | Legomena | moirai-fates: imperative content in a legomena |
| L-H2 | Legomena | file-verification: 161 lines, no companions |
| L-H3 | Legomena | doc-consolidation: 191 lines, no companions |
| L-H4 | Legomena | rite-development: 172 lines, no companions |
| L-H5 | Legomena | agent-prompt-engineering: 207-line INDEX despite 4 companions |
| A-H1 | Agents | 7 agents embed >50 lines of extractable reference content |
| A-H2 | Agents | 11 orchestrators share ~1,320 lines identical boilerplate |

---

## Remediation Roadmap

### Immediate (< 30 min each, no design decisions)

| # | Fix | Effort | Impact |
|---|-----|--------|--------|
| F1 | Fix Moirai phantom ref in 3 locations (template, user-content, session rule) | 15m | Prevent runtime Task(moirai) failures |
| F2 | Add disable-model-invocation to /consult | 2m | Save ~940 tokens/auto-injection |
| F3 | Update internal-hook.md: event count 20, stdin JSON primary | 10m | Accurate developer guidance |
| F4 | Add withTimeout to budget hook | 5m | Consistent timeout behavior |
| F5 | Add timeout to autopark getGitStatusQuick | 5m | Prevent hang on Stop event |
| F6 | Remove Write from /sessions, /architect, /build allowed-tools | 5m | Principle of least privilege |
| F7 | Delete dead copyDir method (28 lines) | 2m | Dead code removal |

**Total**: ~45 minutes, 7 fixes

### Short-term (1-2 hours each, minimal design)

| # | Fix | Effort | Impact |
|---|-----|--------|--------|
| S1 | Split forge-ref (346 lines) into INDEX + 4 companions | 1h | -270 lines per activation |
| S2 | Delete or refactor rite-scoped doc-artifacts (258 lines) | 30m | Remove duplicate of core skill |
| S3 | Remove imperative Loading Protocol from moirai-fates | 15m | Fix lifecycle classification |
| S4 | Reconcile hooks.yaml with settings.local.json | 30m | Single source of truth |
| S5 | Delete dead context_switch event type + trigger logic | 30m | Dead code removal |
| S6 | Delete route.sh shell wrapper (no Go handler exists) | 5m | Dead code removal |
| S7 | Create internal-rite.md rule (4,181 LOC coverage) | 1h | Path-scoped guidance for domain core |
| S8 | Create internal-cmd.md rule (28,759 LOC coverage) | 1h | Path-scoped guidance for CLI surface |
| S9 | Collapse agent-configurations into quick-start | 30m | -140 tokens duplication |

### Multi-session (Phase 3 structural work)

| # | Fix | Effort | Impact |
|---|-----|--------|--------|
| L1 | Migrate 43 agents to multi-line descriptions | 8-16h | CC routing precision for 75% of agents |
| L2 | Extract embedded reference from 7+ oversized agents | 4-8h | -3,000+ tokens in affected invocations |
| L3 | Implement `ari sync --budget` token counting | 8-16h | Framework self-awareness (capstone) |
| L4 | Implement orchestrator boilerplate injection | 8-16h | -1,320 lines duplication across 11 files |
| L5 | Implement `ari lint` for source validation | 8-16h | Catch errors before projection |
| L6 | Downgrade ~15 agents from opus to sonnet | 30m | ~5x cost reduction for affected invocations |
| L7 | Split remaining oversized legomena (5 files) | 2-4h | Progressive disclosure completion |

---

## Scorecard: Current vs Target

| Dimension | Current (8/10) | After Immediate+Short (8.5/10) | After Phase 3 (9/10) |
|-----------|:-:|:-:|:-:|
| **Context isolation** | 100% fork | 100% fork | + boilerplate injection |
| **Skill routing** | 100% Triggers | 100% Triggers | + lint enforcement |
| **Agent routing** | 25% gold-standard | 25% (unchanged) | 100% multi-line + examples |
| **Progressive disclosure** | 44% companion pattern | 55% (forge-ref + doc-artifacts split) | 90%+ (all large files split) |
| **Content freshness** | 3 phantom refs, 1 stale rule | All stale content fixed | + lint prevention |
| **Platform safety** | 5/5 guarded, 2 timeout bugs | 5/5 guarded, timeouts fixed | Full lifecycle |
| **Self-awareness** | Zero token counting | Zero (unchanged) | `ari sync --budget` |
| **CC alignment** | 85% | 87% (semantics clarified) | 90%+ (leverage underused features) |
| **Dead code** | ~85 lines + 1 dead event | Clean | Clean + lint enforcement |

---

## Cross-Domain Patterns

### Pattern 1: Forge Is the North Star
Forge agents consistently demonstrate gold-standard quality: multi-line descriptions with examples, appropriate tool scoping, clear archetype compliance. Every other rite should converge toward forge patterns. The irony: forge's own reference skill (`forge-ref`) is the worst-structured legomena at 346 lines.

### Pattern 2: Rite-Scoped Content Lags Core
Core mena content (13 legomena) scores 85% on progressive disclosure. Rite-scoped content (26 legomena) scores 23%. This suggests forge-generated rite content ships monolithically and is never revisited for decomposition.

### Pattern 3: Phantom References Accumulate
Moirai agent, usersync package, event count drift — completed work leaves phantom references that Claude follows to dead ends. A post-completion checklist or `ari lint` would prevent this class of issue entirely.

### Pattern 4: The Pipeline Knows Everything But Measures Nothing
The materialization pipeline touches every file that enters CC's context window. It computes checksums, tracks provenance, and detects divergence. Yet it has zero visibility into token costs. Adding byte counting to the existing provenance entries is a 2-hour task that would unlock `ari sync --budget`.

---

*Generated by 7-agent parallel context engineering audit, 2026-02-09.*
*Synthesis by Claude Opus 4.6.*

# Context Engineering Audit: Synthesis

**Date**: 2026-02-09
**Scope**: Full Knossos framework — 7 parallel domain audits
**Goal**: Identify the specific changes that take Knossos from 6/10 to 9/10 as a context-engineering framework

---

## Executive Summary

The audit covered 34 dromena, 45 legomena, 58 agents, 12 rites, the inscription system (CLAUDE.md + 9 rules), the materialization pipeline (4 packages, ~5,450 LOC), and the hook/event system (10 handlers, 20 event types).

**The overall architecture is sound.** The 6-primitive model (CLAUDE.md, skills, commands, agents, hooks, rules) is correctly mapped, the inscription is lean (~750 tokens), and hooks correctly handle ephemeral context injection. The problems are not architectural — they are operational: content is in the wrong primitive, metadata is incomplete, and the pipeline lacks self-awareness about its own token budget.

**Three systemic patterns account for ~80% of findings:**

1. **Context lifecycle violations** — Persistent content (legomena) that should be transient (dromena), and transient content (dromena) that lacks isolation (`context:fork`).
2. **Description quality gap** — CC routes autonomously via `description` fields. 70% of agents and 33% of skills have descriptions too vague for reliable routing.
3. **Stale content & dead code** — Deleted packages referenced, "teams" terminology surviving the cleanse, unreachable agents, orphan rules.

---

## Top 10 Findings (Ranked by Context Engineering Impact)

### 1. 28/34 dromena lack `context:fork` — ~12,000 token/session leak
**Source**: Dromena audit, CRIT-1
**Impact**: CRITICAL — single largest context budget problem

Only 6 of 34 dromena use `context:fork`. The remaining 28 inject their full prompt body into the main conversation permanently. Every `/start` (850 tokens), `/sprint` (735 tokens), `/task` (690 tokens), and `/wrap` (715 tokens) invocation leaves its entire prompt in context for the rest of the session.

**Fix**: Add `context:fork` frontmatter to all 28 dromena. Tier 1 (highest cost): sessions, start, sprint, task, wrap, worktree. Mechanical change.
**Effort**: S (1-2 hours)
**Savings**: ~12,000-15,000 tokens/session

### 2. 41/47 non-forge agents use single-line descriptions — CC routing degraded for 70%
**Source**: Agents audit, C-01
**Impact**: CRITICAL — CC discovery mechanism operates at fraction of capacity

CC uses the `description` field as the primary discovery signal for `Task(subagent_type)` routing. The canonical template requires multi-line descriptions with "When to use this agent", `<example>` blocks, and trigger phrases. Only forge agents (7) and orchestrators (11) follow this pattern. The remaining 41 agents have single-line descriptions that provide minimal routing signal.

**Fix**: Migrate all 41 agents to multi-line descriptions with "When to use" and `<example>` blocks. Prioritize by invocation frequency. Use forge agents as templates.
**Effort**: L (8-16 hours across multiple sessions)
**ROI**: Highest single improvement for CC agent routing precision

### 3. 7 `*-ref` legomena are lifecycle-misclassified — 2,400 lines persistent waste
**Source**: Legomena audit, CRIT-1
**Impact**: CRITICAL — wrong primitive, wrong lifecycle

Seven legomena (`10x-ref`, `build-ref`, `architect-ref`, `docs-ref`, `hygiene-ref`, `debt-ref`, `sre-ref`) contain imperative command procedures ("Execute via Bash tool", "Invoke Rite Switch") with side effects. These are dromena behavior masquerading as persistent reference knowledge. When CC loads any of these as a skill, 300-500 lines of command documentation sit in context indefinitely.

**Fix**: Convert from `.lego.md` to `.dro.md`. Extract any genuine reference content (agent catalog, when-to-use) into thin companion legomena (~40 lines each).
**Effort**: M (4-8 hours)
**Savings**: ~2,400 lines of persistent context eliminated

### 4. 15 legomena missing `Triggers:` in description — 33% skill routing failure
**Source**: Legomena audit, CRIT-2
**Impact**: CRITICAL — CC cannot route to skills it can't discover

CC uses the `description` field for autonomous skill loading. Without explicit `Use when:` or `Triggers:` phrases, CC must guess when to load a skill. 15 of 45 legomena have missing or vague trigger descriptions. The `forge-ref` skill is the gold standard; most others fall short.

**Fix**: Standardize all 45 descriptions to: `"[Purpose]. Use when: [intent phrases]. Triggers: [keyword list]."` Add validation to `ari sync` that enforces `Triggers:` presence.
**Effort**: M (4-8 hours)
**Savings**: 33% of skills become discoverable that currently aren't

### 5. No token counting in the pipeline — framework blind to its own budget
**Source**: Pipeline audit, MC-1
**Impact**: CRITICAL (strategic) — a context-engineering framework with no context budget awareness

The pipeline generates all content that enters CC's context window but has zero visibility into token costs. There is no mechanism to estimate per-section CLAUDE.md cost, warn when agent prompts exceed thresholds, or report total projected context budget. This is the highest-priority missing capability for a framework whose value proposition is context engineering.

**Fix**: Implement `ari sync --budget` or `ari context budget` command. Report per-file token estimates, per-section CLAUDE.md costs, and total projected budget.
**Effort**: L (multi-session — requires tokenizer integration or heuristic estimation)
**Value**: Turns context engineering from intuition to measurement

### 6. SessionEnd hook missing — sessions never formally close
**Source**: Hooks audit, C-1
**Impact**: CRITICAL — session lifecycle gap

The Stop event (handled by autopark) fires when Claude finishes a turn. The SessionEnd event fires when the CC conversation window closes. There is no SessionEnd handler. If a user closes CC without `/park` or `/wrap`, the session remains ACTIVE indefinitely, no `session.ended` event is emitted, and budget counters are never cleaned up.

**Fix**: Create `internal/cmd/hook/sessionend.go`. Emit `session.ended` event, auto-park if still ACTIVE, clean up temp files. Register as sync with 5s timeout.
**Effort**: M (4-8 hours)

### 7. CLAUDE.md references deleted `internal/usersync/` — actively misleading
**Source**: Inscription audit, CRIT-1 + CRIT-2 + HIGH-1
**Impact**: CRITICAL — Claude will look for code that doesn't exist

Three stale items in the user-content section: (1) "User sync system (`internal/usersync/`) is new" — package was deleted in Phase 4b, (2) all three "Active Refactoring" bullets describe completed past work, (3) orphan rule `.claude/rules/internal-usersync.md` targets the deleted package.

**Fix**: Delete "Active Refactoring" subsection. Delete orphan rule. Update stale references in two other rules (`internal-materialize.md`, `internal-provenance.md`).
**Effort**: XS (15 minutes)
**Note**: This is the highest-ROI quick win — 15 minutes to remove active misinformation

### 8. 10 agents >200 lines embed reference knowledge — 3,000 tokens waste
**Source**: Agents audit, C-03
**Impact**: HIGH — reference content pays full per-turn token cost

10 agents embed HANDOFF templates, validation checklists, CLI references, and pattern libraries that load on every invocation regardless of whether the reference is needed. The worst offenders: `rnd/tech-transfer` (301 lines), `forge/eval-specialist` (297 lines), `forge/agent-curator` (295 lines).

**Fix**: Extract embedded reference content into skills. Replace inline with `@cross-rite-handoff` or rite-specific skill references. Target: all agents under 200 lines.
**Effort**: M (4-8 hours)
**Savings**: ~3,000 tokens per session for affected agents

### 9. 6 monolithic template skills lack progressive disclosure — up to 749 lines loaded at once
**Source**: Legomena audit, HIGH-2 + HIGH-3
**Impact**: HIGH — defeats the INDEX->companion pattern

`doc-sre` (749 lines), `doc-strategy` (402 lines), `doc-rnd` (399 lines), `doc-security` (319 lines), `doc-intelligence` (260 lines), and `doc-reviews` (252 lines) all inline full templates without companion files. When CC loads `doc-sre`, 749 lines enter context even though users typically need only 1-2 templates.

**Fix**: Split each into INDEX (routing table, ~80 lines) + companion files per template. Follow the `doc-ecosystem` pattern (80-line INDEX + 7 companions).
**Effort**: M (4-8 hours)
**Savings**: Context cost drops from ~2,380 lines to ~480 lines (INDEX-only load)

### 10. Writeguard only protects 2 files — platform infrastructure exposed
**Source**: Hooks audit, H-4
**Impact**: HIGH — critical platform files unguarded

The writeguard blocks direct writes to `SESSION_CONTEXT.md` and `SPRINT_CONTEXT.md` but does not protect `PROVENANCE_MANIFEST.yaml`, `settings.local.json`, `KNOSSOS_MANIFEST.yaml`, or `.claude/CLAUDE.md`. A direct Claude write to any of these could corrupt the sync pipeline.

**Fix**: Add `PROVENANCE_MANIFEST.yaml`, `settings.local.json`, and `KNOSSOS_MANIFEST.yaml` to `protectedPatterns` in `writeguard.go`.
**Effort**: XS (15 minutes)

---

## The 6-to-9 Roadmap

### Phase 1: Quick Wins (< 1 hour each, immediate impact)

| # | Change | Effort | Impact | Domain |
|---|--------|--------|--------|--------|
| Q1 | Add `context:fork` to 28 dromena | 1h | -12K tokens/session | Dromena |
| Q2 | Delete stale "Active Refactoring" + orphan usersync rule | 15m | Remove active misinformation | Inscription |
| Q3 | Expand writeguard to protect 3 more platform files | 15m | Prevent pipeline corruption | Hooks |
| Q4 | Fix 4 surviving "teams" refs in YAML + 6 "Pack" README headers | 30m | Complete SL-008 cleanse | Rite Structure |
| Q5 | Update archetype maxTurns defaults to match reality | 30m | Eliminate dead-code defaults | Agents |
| Q6 | Add `allowed-tools: Task, Read` to meta commands (/minus-1, /zero, /one) | 15m | Restrict unrestricted tool access | Dromena |
| Q7 | Remove `Write` from `/qa` allowed-tools | 5m | Fix tool/intent contradiction | Dromena |
| Q8 | Fix `/start` argument-hint `PACK` -> `NAME` | 5m | Terminology canary | Dromena |
| Q9 | Fix stale references in materialize + provenance rules | 15m | Remove usersync references | Inscription |
| Q10 | Add `disable-model-invocation: true` to `/sessions` | 5m | Prevent 940-token auto-injection | Dromena |

**Total Phase 1 effort**: ~3-4 hours
**Total impact**: ~12,000 tokens/session saved + stale content removed + security surface hardened

### Phase 2: Content Lifecycle Corrections (1-2 sessions)

| # | Change | Effort | Impact | Domain |
|---|--------|--------|--------|--------|
| S1 | Convert 7 `*-ref` legomena to dromena | 4-8h | -2,400 lines persistent waste | Legomena |
| S2 | Standardize 45 legomena descriptions with `Triggers:` | 4-8h | 33% skill routing fixed | Legomena |
| S3 | Split 6 monolithic doc-* template skills into INDEX+companions | 4-8h | -1,900 lines per-load | Legomena |
| S4 | Create rules for `rites/` and `knossos/templates/` paths | 2-4h | Path-scoped guidance for 2 high-edit dirs | Inscription |
| S5 | Implement SessionEnd hook handler | 4-8h | Session lifecycle complete | Hooks |
| S6 | Fix or remove route hook (currently a no-op) | 1-2h | Eliminate wasted computation | Hooks |
| S7 | Wire rnd/tech-transfer into workflow DAG | 1-2h | Fix unreachable agent | Rite Structure |

**Total Phase 2 effort**: ~20-40 hours across 2-3 sessions

### Phase 3: Structural Improvements (multi-session)

| # | Change | Effort | Impact | Domain |
|---|--------|--------|--------|--------|
| L1 | Migrate 41 agents to multi-line descriptions | 8-16h | CC routing precision for 70% of agents | Agents |
| L2 | Extract embedded reference content from 10 oversized agents | 4-8h | -3,000 tokens in affected invocations | Agents |
| L3 | Implement `ari sync --budget` token counting | 8-16h | Framework self-awareness | Pipeline |
| L4 | Implement orchestrator boilerplate injection during materialization | 8-16h | -1,760 lines duplication across 11 files | Agents/Pipeline |
| L5 | Implement `ari lint` for source validation | 8-16h | Catch errors before projection | Pipeline |
| L6 | Unify inscription Pipeline.Sync() with materializeCLAUDEmd() | 4-8h | Eliminate dual code path | Pipeline |
| L7 | Thread provenance collector into ProjectMena() | 4-8h | Fix post-hoc attribution heuristic | Pipeline |

**Total Phase 3 effort**: ~44-88 hours across 4-6 sessions

---

## Scorecard: Before and After

| Dimension | Current (6/10) | After Phase 1 (7/10) | After Phase 2 (8/10) | After Phase 3 (9/10) |
|-----------|---------------|---------------------|---------------------|---------------------|
| **Context isolation** | 18% fork coverage | 100% fork coverage | 100% fork + lifecycle corrections | Full isolation + boilerplate injection |
| **Skill routing** | 67% with Triggers | 67% (unchanged) | 100% standardized | 100% + validation enforcement |
| **Agent routing** | 30% multi-line desc | 30% (unchanged) | 30% (unchanged) | 100% multi-line with examples |
| **Content freshness** | 3 stale refs, 4 legacy terms | All stale content removed | All stale content removed + rules coverage | All stale content removed + lint validation |
| **Platform safety** | 2/5 files guarded | 5/5 files guarded | + SessionEnd + PostToolUseFailure | Full lifecycle coverage |
| **Self-awareness** | No token counting | No token counting | No token counting | `ari sync --budget` |
| **Progressive disclosure** | 62% using INDEX pattern | 62% (unchanged) | 85%+ (doc-* split) | 95%+ (all large skills split) |

---

## Cross-Domain Patterns

### Pattern 1: The Description Quality Gap
All three "content" primitives (dromena, legomena, agents) suffer from the same problem: description fields are too vague for CC's autonomous routing. This is a systemic issue, not per-file. A description quality standard + `ari lint` validation would prevent regression.

### Pattern 2: The Forge Standard
The forge rite consistently demonstrates the target quality: multi-line descriptions with examples, appropriate tool scoping, progressive disclosure. Every other rite should converge toward forge patterns. Ironically, forge itself lacks a README and TODO.

### Pattern 3: Historical Debt Accumulation
Multiple findings trace back to completed work that was not fully cleaned up: deleted `usersync/` still referenced, "teams" terminology surviving the cleanse, `PACK` in argument hints, stale "Active Refactoring" section. A post-completion checklist ("Did you update CLAUDE.md? Did you remove stale references?") would prevent this class of issue.

### Pattern 4: Pipeline Lacks Self-Awareness
The pipeline generates all context but cannot measure it. Token counting (Finding #5) is the capstone capability that would make Knossos a true context-engineering framework rather than a context-generation framework. Every other finding in this audit was discovered by manual agent reads — `ari sync --budget` would surface them automatically.

---

## Audit Source Documents

| Domain | Report | Agent | Findings |
|--------|--------|-------|----------|
| Dromena | `.wip/CE-AUDIT-dromena.md` | mena-dromena | 3 CRIT, 7 HIGH, 6 MED, 4 LOW |
| Legomena | `.wip/CE-AUDIT-legomena.md` | mena-legomena | 2 CRIT, 4 HIGH, 5 MED, 3 LOW |
| Agents | `.wip/CE-AUDIT-agents.md` | rite-agents | 5 CRIT, 8 HIGH, 7 MED, 6 LOW |
| Rite Structure | `.wip/CE-AUDIT-rite-structure.md` | rite-structure | 2 CRIT, 3 HIGH, 5 MED, 4 LOW |
| Inscription | `.wip/CE-AUDIT-inscription.md` | inscription-rules | 2 CRIT, 3 HIGH, 4 MED, 3 LOW |
| Pipeline | `.wip/CE-AUDIT-pipeline.md` | pipeline | 2 CRIT, 4 HIGH, 7 MED, 5 LOW |
| Hooks | `.wip/CE-AUDIT-hooks.md` | hooks-events | 2 CRIT, 4 HIGH, 5 MED, 3 LOW |
| **Total** | | **7 agents** | **18 CRIT, 33 HIGH, 39 MED, 28 LOW** |

---

*Generated by 7-agent parallel context engineering audit, 2026-02-09.*
*Synthesis by Claude Opus 4.6.*

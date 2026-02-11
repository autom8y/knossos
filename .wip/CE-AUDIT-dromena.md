# CE Audit: Dromena Quality

**Auditor**: Context Engineer
**Date**: 2026-02-09
**Scope**: All 34 `.dro.md` files in `mena/`
**Method**: Full read of every dromena file + frontmatter extraction

## Summary

- **Total dromena audited**: 34
- **context:fork coverage**: 6/34 (18%) -- CRITICALLY LOW
- **disable-model-invocation where needed**: 32/34 (94%)
- **Average token cost**: ~550 tokens (range: ~250 to ~1,500)
- **Total token surface if all invoked**: ~18,700 tokens
- **Frontmatter completeness**: 34/34 have name + description (100%)
- **allowed-tools present**: 31/34 (91%)

## Critical Findings

### CRIT-1: 28 dromena lack context:fork -- massive context pollution risk

Only 6 of 34 dromena use `context:fork`:
- `consult`, `rite`, `handoff`, `code-review`, `commit`, `pr`

The remaining **28 dromena inject their full prompt body into the main conversation** when invoked. Every token stays in context for the rest of the session. This is the single largest context budget problem in the system.

**Impact ranking by token cost** (highest cost unforkeds first):

| File | ~Tokens | Severity |
|------|---------|----------|
| `mena/navigation/sessions.dro.md` | ~940 | CRITICAL -- 128 lines of bash scripts pollute main context |
| `mena/workflow/sprint.dro.md` | ~735 | CRITICAL -- 107 lines including parallel worktree patterns |
| `mena/workflow/task.dro.md` | ~690 | CRITICAL -- 109 lines of workflow resolution logic |
| `mena/session/start/INDEX.dro.md` | ~850 | CRITICAL -- 116 lines of session init logic |
| `mena/session/wrap/INDEX.dro.md` | ~715 | CRITICAL -- 107 lines with worktree cleanup bash |
| `mena/navigation/worktree/INDEX.dro.md` | ~610 | CRITICAL -- 126 lines of git worktree commands |
| `mena/operations/architect.dro.md` | ~595 | HIGH -- design workflow, agent invocation |
| `mena/operations/build.dro.md` | ~590 | HIGH -- implementation workflow |
| `mena/operations/qa/INDEX.dro.md` | ~565 | HIGH -- validation workflow |
| `mena/operations/spike/INDEX.dro.md` | ~390 | HIGH -- research workflow with WebSearch |
| `mena/workflow/hotfix/INDEX.dro.md` | ~535 | HIGH -- urgent fix workflow |
| `mena/operations/code-review/INDEX.dro.md` | ~540 | (has fork) |
| `mena/meta/one.dro.md` | ~665 | MEDIUM -- daisy-chain loop protocol |
| `mena/navigation/ecosystem.dro.md` | ~500 | MEDIUM -- rite switch |
| `mena/session/park/INDEX.dro.md` | ~400 | MEDIUM -- session park |
| `mena/session/continue/INDEX.dro.md` | ~425 | MEDIUM -- session resume |

**Recommended fork additions** (priority order):
1. `sessions` -- inline bash scripts are pure pollution
2. `start` -- heavy init logic, always invoked early
3. `sprint` -- long workflow orchestration
4. `task` -- long workflow orchestration
5. `wrap` -- complex cleanup logic
6. `worktree` -- git management scripts
7. All remaining operations: `architect`, `build`, `qa`, `spike`, `hotfix`
8. All session commands: `park`, `continue`
9. Meta commands: `one`, `zero`, `minus-1`
10. All rite-switching commands (10 files) -- low token cost (~250-350 each) but still unnecessary pollution

### CRIT-2: /sessions embeds executable bash scripts as reference knowledge

**File**: `mena/navigation/sessions.dro.md` (lines 23-88)

This dromena contains 65+ lines of inline bash scripts for listing sessions, listing across worktrees, switching, and cleanup. This is a code dump acting as a command prompt.

**Problems**:
1. No `context:fork` -- all bash stays in main context
2. The bash scripts are effectively reference documentation, not a transient action
3. Model will try to execute raw bash rather than calling `ari session` commands

**Recommendation**: Either (a) add `context:fork` and slim the bash to `ari session list/switch/cleanup` calls, or (b) split into a legomenon for the bash reference and keep the dromena as a thin dispatcher.

### CRIT-3: /sessions and /consult missing disable-model-invocation

**Files**:
- `mena/navigation/sessions.dro.md` -- no `disable-model-invocation`
- `mena/navigation/consult/INDEX.dro.md` -- no `disable-model-invocation`

Both are substantial commands (128 and 155 lines respectively). Without `disable-model-invocation: true`, the model could autonomously invoke these, injecting large prompts into context without user intent.

`/consult` is the heaviest dromena at ~1,500 tokens. If auto-invoked, that is a significant context penalty.

**However**: `/consult` is intentionally the "cognitive load absorber" and may benefit from model-invocability. This is a design judgment call. If kept model-invocable, `context:fork` becomes even more critical (and it does have fork).

`/sessions` has no such justification. It should have `disable-model-invocation: true`.

## High Findings

### HIGH-1: /consult contains hardcoded rite table that duplicates dynamic data

**File**: `mena/navigation/consult/INDEX.dro.md` (lines 63-77)

The dromena embeds a static rite table AND says "use `rite-discovery` skill for programmatic access". This is contradictory -- the static table will become stale while the instruction to use the skill is the correct approach.

**Lines 63-77**: 15 lines of hardcoded rite metadata that should be loaded dynamically.
**Lines 80-86**: Hardcoded command counts ("Session (5), Rite (10), Workflow (4), Operations (5)") will drift.

**Recommendation**: Remove hardcoded tables. The instructions at lines 60-62 and 77 already tell the model to use `rite-discovery`. The static tables are redundant weight.

### HIGH-2: /consult embeds 35+ lines of skill cross-reference patterns

**File**: `mena/navigation/consult/INDEX.dro.md` (lines 116-151)

Lines 116-151 document how `/consult` should reference `prompting`, `10x-workflow`, and `rite-discovery` skills, including specific file paths within those skills. This is effectively a mini skill-index embedded inside a dromena.

**Problem**: This is persistent reference knowledge (legomenon-nature) living inside a transient command. If `/consult` had no `context:fork`, this would be pure pollution. It does have fork, which mitigates, but the content itself is still bulky.

**Recommendation**: Compress to a 5-line reference block:
```
When providing guidance, reference these skills:
- `prompting` -- invocation patterns and agent discovery
- `10x-workflow` -- phase transitions and quality gates
- `rite-discovery` -- dynamic rite inventory and metadata
```
The model can load the skills autonomously; it does not need file-path-level guidance.

### HIGH-3: Meta commands (/minus-1, /zero, /one) lack context:fork and allowed-tools

**Files**:
- `mena/meta/minus-1.dro.md` -- no fork, no allowed-tools
- `mena/meta/zero.dro.md` -- no fork, no allowed-tools
- `mena/meta/one.dro.md` -- no fork, no allowed-tools

These are the 10x workflow entry points (Session -1, 0, 1). They spawn orchestrator subagents via Task tool. Without `allowed-tools`, they have access to ALL tools. Without `context:fork`, their prompt bodies pollute the main conversation.

`/one` is 80 lines with agent instance strategy tables and invocation templates -- this is procedural knowledge that should not persist in context after execution.

**Recommendation**:
1. Add `context:fork` to all three
2. Add `allowed-tools: Task, Read` (they only need to invoke agents and read context)

### HIGH-4: /start uses legacy `--rite=PACK` terminology in argument-hint

**File**: `mena/session/start/INDEX.dro.md` (line 5)

```yaml
argument-hint: <initiative> [--complexity=LEVEL] [--rite=PACK]
```

The `PACK` placeholder is legacy terminology from the pre-rite era. The SL-008 terminology cleanse should have caught this.

**Recommendation**: Change to `--rite=NAME`

### HIGH-5: 10 rite-switching dromena are structurally identical -- template candidate

**Files**: `mena/rite-switching/{10x,debt,docs,forge,hygiene,intelligence,rnd,security,sre,strategy}.dro.md`

These 10 files follow an identical pattern:
```
1. Execute: ari sync --rite <name>
2. Display pantheon output
3. Update SESSION_CONTEXT
```

Each is ~40 lines (~280 tokens). Total: ~2,800 tokens of nearly identical content.

While individually they are lean and well-scoped, the duplication is a maintenance burden. If the sync workflow changes, 10 files must be updated.

**Exception**: `/forge` (64 lines) has an extra inline display block (lines 23-45) showing forge agent overview. This is reference content embedded in a transient command.

**Recommendation**: Consider a template-based approach in the materialization pipeline so a single source generates all 10. Low urgency since each file is small.

### HIGH-6: /spike has overly broad tool access

**File**: `mena/operations/spike/INDEX.dro.md` (line 6)

```yaml
allowed-tools: Bash, Read, Write, Task, Glob, Grep, WebFetch, WebSearch
```

This is the broadest tool surface in the entire dromena set (8 tools). While a spike needs exploration capability, `Write` combined with "NO PRODUCTION CODE" instructions creates a contradiction -- the prompt says no production code but the tools allow writing files.

**Recommendation**: Remove `Write` from allowed-tools if the intent is truly "research and report only". The spike report can be produced via the Task tool delegating to a writer, or the user can ask for the file explicitly.

### HIGH-7: Several dromena grant Write access unnecessarily

**Files with Write but unclear justification**:
- `mena/operations/qa/INDEX.dro.md` -- "Testing and validation only" but has Write
- `mena/operations/build.dro.md` -- has Write (justified: implementation)
- `mena/operations/architect.dro.md` -- has Write (justified: design docs)
- `mena/navigation/sessions.dro.md` -- has Write for session management (marginal)

`/qa` explicitly states "NO IMPLEMENTATION. Testing and validation only." yet has Write tool access. This is a contradictory signal.

**Recommendation**: Remove `Write` from `/qa`.

## Medium Findings

### MED-1: /ecosystem lacks context:fork despite being a rite-switch

**File**: `mena/navigation/ecosystem.dro.md`

All other rite-switching commands (the 10 in `rite-switching/`) lack fork as well, but `/ecosystem` is additionally in `navigation/` with 60 lines including workflow phases and complexity levels documentation (lines 43-57). That reference content stays in context.

The generic `/rite` command DOES have `context:fork`. Inconsistency.

**Recommendation**: Add `context:fork` to `/ecosystem` and consider adding it to all rite-switching commands for consistency.

### MED-2: /commit Attribution Policy section is redundant

**File**: `mena/operations/commit/INDEX.dro.md` (lines 110-121)

The "Attribution Policy" section (12 lines) repeats what is already stated at lines 78-81 within the behavior section. Two separate blocks saying "do NOT add Co-Authored-By" is pure token waste.

**Recommendation**: Remove the "Attribution Policy" section (lines 110-121). The inline instruction at lines 78-81 is sufficient.

### MED-3: /wrap embeds bash worktree detection logic

**File**: `mena/session/wrap/INDEX.dro.md` (lines 73-97)

Lines 73-97 contain bash snippets for worktree detection and cleanup prompts. This is procedural logic that should live in `ari session wrap`, not in the command prompt. The model should call `ari session wrap` and handle the output, not replicate the logic.

**Recommendation**: Replace the bash block with:
```
5. **Worktree cleanup**: If `ari session wrap` detects a worktree, it will prompt for removal.
```

### MED-4: /hotfix and /task repeat workflow resolution bash patterns

**Files**:
- `mena/workflow/hotfix/INDEX.dro.md` (lines 33-34)
- `mena/workflow/task.dro.md` (lines 30-37)
- `mena/operations/build.dro.md` (lines 33-37)
- `mena/operations/architect.dro.md` (lines 33-37)
- `mena/operations/qa/INDEX.dro.md` (lines 33-37)
- `mena/operations/code-review/INDEX.dro.md` (lines 29-32)

Six dromena contain inline bash for parsing `ACTIVE_WORKFLOW.yaml` to extract agent names. This is:
1. Duplicated logic (6 copies)
2. Fragile (depends on YAML structure)
3. Should be an `ari` CLI call: `ari workflow agent --phase=implementation`

**Recommendation**: Replace all 6 inline bash blocks with a single `ari workflow` CLI call.

### MED-5: /forge embeds agent list that will become stale

**File**: `mena/rite-switching/forge.dro.md` (lines 23-45)

Hardcoded 6-agent overview with specific role descriptions. This will drift as forge evolves.

**Recommendation**: Remove inline display. The `ari sync` output already shows the pantheon.

### MED-6: Inconsistent model selection across similar commands

| Category | Model Used | Notes |
|----------|-----------|-------|
| Rite-switching (10) | haiku | Appropriate -- simple dispatch |
| Session commands | sonnet (4), opus (1: start) | `/start` needs opus for complexity assessment |
| Meta commands (3) | opus | Appropriate -- orchestration |
| Operations | opus (5), sonnet (3) | Commit/PR/worktree as sonnet is right |
| Navigation | opus (1), sonnet (3) | `/consult` opus is right |
| Workflow | opus (3) | Appropriate -- complex orchestration |

No issues found. Model selections are appropriate for complexity.

## Low Findings

### LOW-1: /sync has legacy compatibility section that could be compressed

**File**: `mena/cem/sync.dro.md` (lines 47-51)

5 lines of legacy command mappings. Useful for transition but will eventually be dead weight.

**Recommendation**: Add a TTL comment: `<!-- LEGACY: Remove after 2026-06 -->`

### LOW-2: /sessions self-references its documentation path incorrectly

**File**: `mena/navigation/sessions.dro.md` has no reference section pointing to projected command path, unlike most other dromena.

### LOW-3: Inconsistent "Reference" section formatting

Some dromena use "Full documentation: ..." others use "Full docs: ..." or omit entirely. Minor cosmetic inconsistency across 34 files.

### LOW-4: /rite has stale argument-hint with --rite=PACK echo

**File**: `mena/navigation/rite.dro.md` -- the argument-hint is clean, but line 59 in Agent Provenance section says "per-agent origin tracking (AGENT_MANIFEST.json) is planned but not yet implemented" -- this is a planning note that should not be in a command prompt.

## Per-Dromena Assessment

| File | Fork? | DMI? | Tools | ~Tokens | Issues |
|------|-------|------|-------|---------|--------|
| `meta/minus-1.dro.md` | NO | YES | (none) | ~250 | HIGH-3: no fork, no tools scoping |
| `meta/zero.dro.md` | NO | YES | (none) | ~300 | HIGH-3: no fork, no tools scoping |
| `meta/one.dro.md` | NO | YES | (none) | ~665 | HIGH-3: no fork, no tools scoping, heavy |
| `navigation/consult/INDEX.dro.md` | YES | NO | Bash,Read,Grep,Glob,WebSearch | ~1,500 | CRIT-3 (DMI missing), HIGH-1, HIGH-2 |
| `navigation/ecosystem.dro.md` | NO | YES | Bash,Read | ~500 | MED-1: no fork |
| `navigation/rite.dro.md` | YES | YES | Bash,Read | ~840 | LOW-4: stale planning note |
| `navigation/sessions.dro.md` | NO | NO | Bash,Read,Write | ~940 | CRIT-1, CRIT-2, CRIT-3, HIGH-7 |
| `navigation/worktree/INDEX.dro.md` | NO | YES | Bash,Read | ~610 | CRIT-1: no fork |
| `rite-switching/10x.dro.md` | NO | YES | Bash,Read | ~265 | HIGH-5: template candidate |
| `rite-switching/debt.dro.md` | NO | YES | Bash,Read | ~265 | HIGH-5: template candidate |
| `rite-switching/docs.dro.md` | NO | YES | Bash,Read | ~260 | HIGH-5: template candidate |
| `rite-switching/forge.dro.md` | NO | YES | Bash,Read | ~450 | HIGH-5, MED-5: stale agent list |
| `rite-switching/hygiene.dro.md` | NO | YES | Bash,Read | ~260 | HIGH-5: template candidate |
| `rite-switching/intelligence.dro.md` | NO | YES | Bash,Read | ~300 | HIGH-5: template candidate |
| `rite-switching/rnd.dro.md` | NO | YES | Bash,Read | ~285 | HIGH-5: template candidate |
| `rite-switching/security.dro.md` | NO | YES | Bash,Read | ~300 | HIGH-5: template candidate |
| `rite-switching/sre.dro.md` | NO | YES | Bash,Read | ~265 | HIGH-5: template candidate |
| `rite-switching/strategy.dro.md` | NO | YES | Bash,Read | ~295 | HIGH-5: template candidate |
| `session/continue/INDEX.dro.md` | NO | YES | Bash,Read,Task | ~425 | CRIT-1: no fork |
| `session/handoff/INDEX.dro.md` | YES | YES | Bash,Read,Task | ~520 | Clean |
| `session/park/INDEX.dro.md` | NO | YES | Bash,Read,Task | ~400 | CRIT-1: no fork |
| `session/start/INDEX.dro.md` | NO | YES | Bash,Read,Task | ~850 | CRIT-1: no fork, HIGH-4: PACK |
| `session/wrap/INDEX.dro.md` | NO | YES | Bash,Read,Task,Glob | ~715 | CRIT-1: no fork, MED-3 |
| `cem/sync.dro.md` | NO | YES | Bash,Read | ~575 | LOW-1: legacy compat |
| `workflow/sprint.dro.md` | NO | YES | Bash,Read,Task,Glob,Grep | ~735 | CRIT-1: no fork |
| `workflow/hotfix/INDEX.dro.md` | NO | YES | Bash,Read,Task,Glob,Grep | ~535 | CRIT-1: no fork, MED-4 |
| `workflow/task.dro.md` | NO | YES | Bash,Read,Task,Glob,Grep | ~690 | CRIT-1: no fork, MED-4 |
| `operations/architect.dro.md` | NO | YES | Bash,Read,Write,Task,Glob,Grep | ~595 | CRIT-1: no fork, MED-4 |
| `operations/build.dro.md` | NO | YES | Bash,Read,Write,Task,Glob,Grep | ~590 | CRIT-1: no fork, MED-4 |
| `operations/code-review/INDEX.dro.md` | YES | YES | Bash,Read,Glob,Grep | ~540 | Clean (has fork) |
| `operations/commit/INDEX.dro.md` | YES | YES | Bash,Read,Glob,Grep | ~835 | MED-2: redundant attribution |
| `operations/pr/INDEX.dro.md` | YES | YES | Bash,Read,Glob,Grep | ~355 | Clean |
| `operations/qa/INDEX.dro.md` | NO | YES | Bash,Read,Write,Task,Glob,Grep | ~565 | CRIT-1: no fork, HIGH-7 |
| `operations/spike/INDEX.dro.md` | NO | YES | Bash,Read,Write,Task,Glob,Grep,WebFetch,WebSearch | ~390 | CRIT-1: no fork, HIGH-6 |

## Recommendations

Ordered by impact on context budget:

### 1. Add context:fork to all 28 missing dromena [CRITICAL]

**Estimated context savings**: ~12,000-15,000 tokens per session (depending on which commands are invoked).

Every dromena is transient by definition. Its prompt should not persist in the main conversation. `context:fork` is the single most impactful change.

**Priority tiers**:
- **Tier 1** (do first -- highest token cost): `sessions`, `start`, `sprint`, `task`, `wrap`, `worktree`
- **Tier 2** (next -- medium cost): `architect`, `build`, `qa`, `spike`, `hotfix`, `one`, `ecosystem`
- **Tier 3** (then -- low cost): all 10 rite-switching, `sync`, `park`, `continue`, `zero`, `minus-1`

### 2. Add disable-model-invocation to /sessions [HIGH]

One-line frontmatter change. Prevents accidental model invocation of a 940-token command.

### 3. Remove hardcoded tables from /consult [HIGH]

Remove ~30 lines of hardcoded rite and command listings. The prompt already instructs use of `rite-discovery` skill. Saves ~100 tokens per invocation and eliminates staleness risk.

### 4. Add allowed-tools to meta commands [HIGH]

`/minus-1`, `/zero`, `/one` need `allowed-tools: Task, Read` to prevent unrestricted tool access during orchestration.

### 5. Compress /consult skill cross-reference section [HIGH]

Replace 35 lines of file-path-level skill guidance with 5-line summary. Saves ~120 tokens.

### 6. Centralize workflow resolution bash into ari CLI [MEDIUM]

Six dromena contain duplicated bash for parsing `ACTIVE_WORKFLOW.yaml`. Replace with `ari workflow agent --phase=<name>`. Eliminates ~30 lines x 6 = 180 lines of duplication.

### 7. Fix /start argument-hint PACK terminology [MEDIUM]

One-word change: `PACK` -> `NAME`. Terminology canary from SL-008 cleanse.

### 8. Remove Write from /qa allowed-tools [MEDIUM]

Contradicts the "NO IMPLEMENTATION" instruction in the prompt body.

### 9. Remove redundant Attribution Policy from /commit [LOW]

Delete 12 lines that duplicate earlier instructions. Saves ~40 tokens.

### 10. Templatize rite-switching dromena [LOW]

10 nearly identical files could be generated from a single template in the materialization pipeline. Low urgency since individual files are small (~260 tokens).

## Appendix: Token Cost Distribution

```
Category                 Count  Avg Tokens  Total Tokens
-----------------------------------------------------------
Rite-switching             10      ~280       ~2,800
Session management          5      ~560       ~2,800
Meta (10x workflow)         3      ~405       ~1,215
Navigation                  4      ~720       ~2,880
Operations                  7      ~555       ~3,885
Workflow                    3      ~655       ~1,965
CEM (sync)                  1      ~575         ~575
Infrastructure (sync)       1      ~575         ~575
-----------------------------------------------------------
TOTAL                      34      ~550      ~18,700
```

## Appendix: Clean Dromena (No Issues)

These dromena are well-architected with correct fork, DMI, and tool scoping:
- `/handoff` -- fork, DMI, tools: Bash,Read,Task
- `/code-review` -- fork, DMI, tools: Bash,Read,Glob,Grep
- `/pr` -- fork, DMI, tools: Bash,Read,Glob,Grep

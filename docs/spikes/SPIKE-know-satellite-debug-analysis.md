# SPIKE: /know Satellite Session Debug Log Analysis

## Question and Context

**What are we trying to learn?**
Deep analysis of debug log `caa3f447-c7dc-4618-84c2-0192a4bc374e.txt` from a `/know --all` session in the `autom8y-asana` satellite to identify the root issue, performance pathologies, and any gaps in the `/know` dromenon or its supporting infrastructure.

**What decision will this inform?**
Whether `/know` needs fixes for satellite operation, whether the observed streaming stalls indicate a systemic issue, and whether the Haiku subagent usage pattern is problematic.

## Session Summary

| Property | Value |
|----------|-------|
| Session ID | `caa3f447-c7dc-4618-84c2-0192a4bc374e` |
| Project | `autom8y-asana` (satellite) |
| CC Version | 2.1.62 |
| Command | `/know` (forked, `model: opus`, `context: fork`) |
| Duration | 11:16:54 -- 11:30:19 (13m 25s) |
| Token Growth | 2,607 --> 150,630 (148K tokens consumed) |
| Context Window | 180K effective, 167K autocompact threshold |
| Total Tool Calls | 107 (tracked by PostToolUse hook counter) |
| Files Written | 5 `.know/` files totaling ~69KB |
| Streaming Stalls | 5 detected, 389s total stall time |

## Approach Taken

1. Parsed 6,661 lines of debug output end-to-end
2. Tracked token growth curve across all autocompact checkpoints
3. Identified all ERROR and WARN entries
4. Traced the model selection pattern (Opus vs. Haiku)
5. Analyzed the criteria path resolution behavior
6. Mapped the Write operations to understand output completeness
7. Correlated streaming stalls with the theoros output windows

## Findings

### Finding 1: The Session Completed Successfully (Mostly)

Despite the issues below, the `/know` session produced all 5 codebase-scoped domains:

| Domain | Size | Written At | Notes |
|--------|------|------------|-------|
| architecture | 17KB | 11:23:28 | After 97.2s stall |
| conventions | 11KB | 11:24:45 | After 67.8s stall |
| test-coverage | 11KB | 11:25:52 | After 57.8s stall |
| scar-tissue | 13KB | 11:27:22 | After 65.9s stall |
| design-constraints | 17KB | 11:29:29 | After 100.3s stall |

The session did NOT appear to run `--all` with Argus Pattern (parallel Task dispatch). The writes are **sequential** with ~2min gaps, indicating the forked Opus agent dispatched theoros subagents one at a time rather than in a single parallel burst as the dromenon specifies.

### Finding 2: Criteria Path Resolution -- File Not Found (Non-Fatal)

At line 4271-4272:
```
Read tool error (14ms): File does not exist. Note: your current working directory is /Users/tomtenuta/Code/autom8y-asana.
```

**Root Cause**: The `/know` dromenon instructs Phase 1 to read criteria from:
```
Read("rites/shared/mena/pinakes/domains/{domain}.lego.md")
```
This path exists in the **knossos core repo** (`/Users/tomtenuta/Code/knossos/`) but NOT in the satellite `autom8y-asana`. The dromenon has a fallback:
```
Read(".claude/skills/pinakes/domains/{domain}.md")
```
This fallback **does work** because `ari sync` materializes the pinakes skill files into `.claude/skills/` in the satellite.

**Impact**: Low. The fallback triggered correctly. But the error is needless noise in satellite projects. The primary path `rites/shared/mena/pinakes/domains/` is a **knossos-only path** that will never exist in a satellite.

**Fix**: The `/know` dromenon (line 57-63 of `INDEX.dro.md`) should reverse the path resolution order for satellites, or detect project type. More robustly: try the materialized path first (`.claude/skills/pinakes/domains/{domain}.md`) since it always exists after sync, then fall back to the source path only in knossos core.

### Finding 3: Haiku Subagents Interleaved with Opus Outer Agent

The debug log reveals two distinct API billing patterns:
- **`cc_version=2.1.62.68f`** (109 calls): The Opus-class outer agent running the `/know` forked command
- **`cc_version=2.1.62.39c` / `.34c` / `.fc5`** (27 calls): Haiku 4.5 subagent calls with explicit `Tool search disabled for model 'claude-haiku-4-5-20251001'`

**What Haiku is doing**: These are the **theoros Task subagents**. Despite the `/know` dromenon specifying `model: opus`, the Task tool dispatches subagents using the CC-default subagent model (Haiku 4.5), not the parent's model override. The `model: opus` frontmatter only controls the *forked parent agent*, not its Task children.

**Impact**: Medium-High. Theoros subagents running on Haiku 4.5 will produce lower-quality knowledge observations than Opus would. The codebase archaeology and scar tissue detection domains especially benefit from Opus-class reasoning. However, the 5 files were produced successfully, suggesting Haiku is adequate for basic knowledge extraction.

**This is NOT a bug** -- it's working as designed by CC's Task tool semantics. But it may be a design gap in the `/know` dromenon's assumptions. The dromenon states "Each theoros gets its own 150-turn context window" but doesn't account for model downgrade.

### Finding 4: Massive Skill Reload Thrashing

Two episodes of skill directory change detection:
1. **Lines 674-1197** (11:17:45-11:17:47): 40+ skill file changes detected, each triggering a full reload of all 107 skills. This is the knossos user-level skills directory (`~/.claude/skills/`) being written to (likely by another ari sync or session in a parallel terminal).
2. **Lines 4776-5283** (11:22:52-11:22:53): Another burst of 35+ skill file changes with identical reload spam.

Each change triggers:
```
Detected skill change: /Users/tomtenuta/.claude/skills/...
Loading skills from: managed=... user=... project=...
getSkills returning: 107 skill dir commands, 0 plugin skills, 3 bundled skills
```

**Impact**: Low on this session (CC handles it gracefully). But each reload is ~5ms of synchronous work, and the cascading effect produces 40+ reloads in ~2 seconds. This is a CC platform behavior, not a knossos bug.

**Root Cause**: Something modified `~/.claude/skills/` files mid-session. Possible sources: parallel ari sync, user editing, or another CC session running `/sync`. The watcher fires per-file, not debounced.

### Finding 5: Streaming Stalls Correlate with Theoros Output Generation

All 5 streaming stalls occur immediately before a Write operation to `.know/`:

| Stall | Duration | Followed By |
|-------|----------|-------------|
| #1 | 97.2s | Write `.know/architecture.md` |
| #2 | 67.8s | Write `.know/conventions.md` |
| #3 | 57.8s | Write `.know/test-coverage.md` |
| #4 | 65.9s | Write `.know/scar-tissue.md` |
| #5 | 100.3s | Write `.know/design-constraints.md` |

**Interpretation**: The stalls are the API generating the `.know/` file content inline within the streaming response. These are large markdown documents (11-17KB each) being produced token-by-token. The "stall" detection fires because the initial tokens arrive, then there's a pause while the model generates the full document body before the Write tool call is emitted.

**Impact**: Informational -- not a bug. The stall detector's 30s+ threshold is triggering on normal large-output generation. These are expected for knowledge documents.

### Finding 6: Argus Pattern (Parallel Dispatch) Was NOT Used

The `/know` dromenon explicitly requires Argus Pattern for `--all`:
> "If `--all` (multiple domains): Launch ALL theoros agents in a SINGLE response block using multiple Task tool calls."

But the debug log shows sequential generation: architecture at 11:23, conventions at 11:24, test-coverage at 11:25, scar-tissue at 11:27, design-constraints at 11:29. Each domain took ~2 minutes end-to-end.

**Root Cause**: The forked Opus agent did not use Task tool for theoros dispatch at all. It appears to have performed the observation **in-context** rather than delegating to theoros subagents. Evidence:
- The Haiku calls (`.39c`/`.34c`/`.fc5`) appear only in the first 5 minutes (11:17-11:21), during the initial skill loading and criteria reading phase
- The Write operations occur in the later half (11:23-11:29) with no interleaved Haiku calls
- Token count grew from 116K to 150K during the write phase, consistent with in-context generation

This directly violates the dromenon's anti-pattern: "If you find yourself reading source code and writing knowledge sections, STOP -- you are violating the dispatch pattern."

**Impact**: High. The forked Opus agent consumed its own 150K-token context doing observation work that should have been delegated to 5 parallel theoros subagents, each with their own context window. This means:
- Knowledge quality is lower (Opus exhausted context rather than having 5 fresh windows)
- Total time was ~13 minutes sequential instead of ~3 minutes parallel
- Context saturation risk: ended at 150K/167K threshold (89% of autocompact limit)

### Finding 7: MCP Server Noise (Non-Issue)

- **mermaid-mcp**: SSE connections drop every ~60-250s and reconnect. Produces constant `No token data found` noise. Non-impactful but noisy.
- **claude.ai Google Calendar, Gmail, Slack**: Authentication failures. These MCP servers require OAuth but no token is configured. Non-impactful on this session.
- **terraform**: `Missing or invalid 'uri' parameter` error on resources/list. Non-impactful.

### Finding 8: Token Growth Trajectory

```
11:17:13  2,607   (start)
11:17:47  44,988  (+42K in 34s -- skill/criteria loading burst)
11:18:19  55,020  (+10K -- TaskCreate phase)
11:19:57  95,054  (+40K in 98s -- source code reading)
11:21:29  116,273 (+21K in 92s -- more reading)
11:23:28  123,537 (+7K -- first .know/ write)
11:29:29  148,516 (+25K in 6min -- remaining 4 .know/ writes)
11:30:00  150,630 (end)
```

The critical observation: 116K tokens consumed by the time the first `.know/` file was written. This is ~69% of the autocompact threshold spent on observation. By session end, 150K/167K (89.8%) was consumed.

## Recommendation

### Priority 1: Fix Criteria Path Resolution Order (Easy)

In `/Users/tomtenuta/Code/knossos/rites/shared/mena/know/INDEX.dro.md`, lines 57-63, reverse the path resolution:

**Current**:
1. Try `rites/shared/mena/pinakes/domains/{domain}.lego.md` (knossos-only path)
2. Fall back to `.claude/skills/pinakes/domains/{domain}.md` (always exists after sync)

**Proposed**:
1. Try `.claude/skills/pinakes/domains/{domain}.md` (works everywhere)
2. Fall back to `rites/shared/mena/pinakes/domains/{domain}.lego.md` (knossos core only)

This eliminates the "File does not exist" error in every satellite invocation.

### Priority 2: Investigate Argus Pattern Non-Compliance (Medium)

The forked Opus agent performed in-context observation instead of dispatching theoros subagents. Possible causes:
- The dromenon instructions about Task dispatch may not be strong enough to override the model's default behavior
- The forked context may not have Task tool available (check `allowed-tools` -- it lists `Task` but the debug log shows no Task tool invocations for theoros)
- The model may have decided in-context was more efficient (wrong decision per the dromenon's anti-pattern section)

**Action**: Run `/know architecture` in knossos core with debug logging to compare whether Task dispatch occurs there. If it does, the issue is satellite-specific (missing agent definitions for theoros). If it doesn't, the dromenon needs stronger enforcement language.

### Priority 3: Consider Theoros Model Override (Informational)

If theoros subagents are dispatched via Task tool, they run on Haiku 4.5 by default. For knowledge-intensive domains (scar-tissue, design-constraints), Opus-class reasoning produces significantly better results. CC's Task tool does not currently support a `model` parameter, so this would require either:
- A theoros agent prompt that requests Opus via frontmatter (if CC supports it for Task agents)
- Acceptance that Haiku is "good enough" for automated knowledge extraction

## Follow-Up Actions

1. **Apply criteria path fix** to `INDEX.dro.md` (5-minute change)
2. **Test Argus Pattern compliance** in both knossos core and satellite contexts
3. **Evaluate .know/ quality** from this session -- are the 5 files useful despite the process issues?
4. **Consider debouncing** the skill reload watcher in CC (platform issue, not knossos)
5. **Monitor streaming stall frequency** -- if these are false positives on large outputs, consider adjusting the stall detector threshold or suppressing for Write operations

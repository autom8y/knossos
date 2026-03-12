# ADR-0032: Harness-Agnostic Canonical Event and Tool Vocabulary

## Status

Accepted

## Date

2026-03-12

## Amends

ADR-0031 (Multi-Channel Architecture)

---

## Context

ADR-0031 established knossos's multi-channel architecture. Its rationale included this justification for using CC names as the internal lingua franca:

> "CC-canonical as lingua franca: Claude Code was first. Its event names, tool names, and hook wire format are already embedded in hooks.yaml, hook command source, and developer mental models. Translating at the channel boundary keeps the core pipeline untouched."

This reasoning was sound at two channels. It breaks at N.

### The Problem with CC-Canonical Internal Names

When CC names (`PreToolUse`, `PostToolUse`, `Bash`, `Grep`) are the internal vocabulary, every channel-aware package encodes a hierarchy: Claude Code is the real thing; Gemini is the translation. This shows up concretely:

- `internal/hook/env.go` defines 14 `HookEvent` constants using CC wire values as their string representations (`"PreToolUse"`, `"PostToolUse"`, etc.)
- `hooks.yaml` is described as "CC-canonical (source of truth)" — developers reading it see CC names, not knossos concepts
- Adding a third channel (Cursor, Windsurf, OpenAI Codex CLI) requires updating translation tables that are already CC-biased, not vocabulary-neutral

The deeper problem: knossos is a meta-framework. It orchestrates AI coding assistants. When knossos's own internal vocabulary is borrowed from one of the assistants it orchestrates, it is not a meta-framework — it is a CC wrapper with adapters bolted on.

### The Harness Landscape

A vocabulary audit of 10 harnesses (Claude Code, Gemini CLI, Cursor, Windsurf, Cline, Aider, Continue.dev, Amazon Q Developer, GitHub Copilot, OpenAI Codex CLI) surfaced in `SPIKE-harness-vocabulary-audit.md` produced four findings relevant to this decision:

**Finding 1: snake_case dominates tool naming.** Gemini CLI, Copilot (VS Code surface), Cline, Amazon Q, and the broader MCP ecosystem all use snake_case for tool names. Claude Code's PascalCase (`Bash`, `Grep`, `Read`) is the outlier across the landscape. A canonical vocabulary should align with the emerging standard.

**Finding 2: No harness agrees on event names.** CC uses PascalCase (`PreToolUse`), Gemini uses PascalCamel with a verb-prefix (`BeforeTool`, `AfterTool`), and other harnesses either lack lifecycle hooks or use their own schemes. There is no industry standard to defer to — knossos must define its own.

**Finding 3: AGENTS.md is an emerging cross-harness standard.** OpenAI Codex CLI uses it natively. GitHub Copilot reads it as a fallback (alongside `CLAUDE.md` and `GEMINI.md`). 20+ tools claim support. This makes `AGENTS.md` a third compilation target knossos should produce — not a canonical source, but an interop artifact.

**Finding 4: Gemini-exclusive events exist.** Gemini fires `BeforeModel` and `AfterModel` hooks (pre/post-LLM-inference) that have no CC equivalent. A CC-canonical internal vocabulary has no natural home for these. A knossos-owned vocabulary adds them as first-class members.

### Path B, Option C: Full Neutral Canonical Vocabulary

Three paths were evaluated for the internal vocabulary:

- **Path A** (maintain CC-canonical): ADR-0031 status quo. Rejected — CC names become a liability at N channels, not just a historical artifact.
- **Path B, Option A** (per-harness internal vocabularies): Each channel package uses its own names. Rejected — no single source of truth; cross-channel code requires context-switching.
- **Path B, Option B** (MCP tool names as canonical): MCP (`read_file`, `execute_code`) as the lingua franca. Rejected — MCP serves a different abstraction level (tool-calling protocol, not lifecycle events). MCP has no event vocabulary at all.
- **Path B, Option C** (full neutral canonical vocabulary): knossos owns its names. CC and Gemini are translation peers. **This is the decision.**

---

## Decision

knossos owns its vocabulary. CC and Gemini are translation peers — neither is canonical. Wire names are translations that happen at the adapter boundary. The internal representation uses knossos canonical names throughout.

The canonical vocabulary is not an abstraction layer. It is the real vocabulary. Wire names are translations.

### 6a. Canonical Event Table (GATE 1)

All 18 knossos lifecycle events. The `direction` column indicates which harnesses fire the event: `bidirectional` means both CC and Gemini support it, `cc_only` means CC-exclusive, `gemini_only` means Gemini-exclusive, `outbound_only` means knossos generates these hooks (no harness fires them inbound).

| canonical_name | go_constant | cc_wire | gemini_wire | direction |
|---------------|-------------|---------|-------------|-----------|
| `pre_tool` | `EventPreTool` | `PreToolUse` | `BeforeTool` | bidirectional |
| `post_tool` | `EventPostTool` | `PostToolUse` | `AfterTool` | bidirectional |
| `post_tool_failure` | `EventPostToolFailure` | `PostToolUseFailure` | — | outbound_only |
| `permission_request` | `EventPermissionRequest` | `PermissionRequest` | — | outbound_only |
| `stop` | `EventStop` | `Stop` | — | outbound_only |
| `session_start` | `EventSessionStart` | `SessionStart` | `SessionStart` | bidirectional |
| `session_end` | `EventSessionEnd` | `SessionEnd` | `SessionEnd` | bidirectional |
| `pre_prompt` | `EventPrePrompt` | `UserPromptSubmit` | `BeforeAgent` | bidirectional |
| `pre_compact` | `EventPreCompact` | `PreCompact` | `PreCompress` | bidirectional |
| `subagent_start` | `EventSubagentStart` | `SubagentStart` | — | outbound_only |
| `subagent_stop` | `EventSubagentStop` | `SubagentStop` | — | outbound_only |
| `notification` | `EventNotification` | `Notification` | `Notification` | bidirectional |
| `teammate_idle` | `EventTeammateIdle` | `TeammateIdle` | — | cc_only |
| `task_completed` | `EventTaskCompleted` | `TaskCompleted` | — | cc_only |
| `pre_model` | `EventPreModel` | — | `BeforeModel` | gemini_only |
| `post_model` | `EventPostModel` | — | `AfterModel` | gemini_only |
| `worktree_create` | `EventWorktreeCreate` | `WorktreeCreate` | — | cc_only |
| `worktree_remove` | `EventWorktreeRemove` | `WorktreeRemove` | — | cc_only |

**Interpretation of direction values:**

- `bidirectional`: Both CC and Gemini fire and receive this event. knossos registers hooks for it on both channels.
- `outbound_only`: knossos generates this hook type; no harness fires it inbound to knossos hook commands. The hook command receives the event but no translation from a harness is needed.
- `cc_only`: CC fires this event; Gemini has no equivalent. Hook registration for Gemini silently skips it (existing ADR-0031 behavior preserved).
- `gemini_only`: Gemini fires this event; CC has no equivalent. Hook registration for CC silently skips it.

### 6b. Canonical Tool Table (GATE 2)

All 11 knossos tool concepts. The `cc_name` column reflects the name Claude Code uses in `tool_name` payloads. The `gemini_name` column reflects the Gemini CLI wire name.

| canonical_name | cc_name | gemini_name |
|---------------|---------|-------------|
| `read_file` | `Read` / `ReadFiles` | `read_file` |
| `edit_file` | `Edit` | `replace` |
| `write_file` | `Write` | `write_file` |
| `list_files` | `Glob` | `glob` |
| `search_content` | `Grep` | `grep_search` |
| `run_shell` | `Bash` | `run_shell_command` |
| `web_search` | `WebSearch` | `google_web_search` |
| `web_fetch` | `WebFetch` | `web_fetch` |
| `write_todos` | `TodoWrite` | `write_todos` |
| `activate_skill` | `Skill` | `activate_skill` |
| `delegate` | `Task` | — (CC-only today) |

Unknown tool names not in this table pass through unchanged at the adapter boundary. This is defensive behavior from ADR-0031 and is preserved.

### 6c. Canonical Agent Definition Keys (GATE 3)

The knossos agent definition format uses these keys. Channel compilers translate them to channel-native formats during materialization.

| canonical_key | cc_equivalent | gemini_equivalent | notes |
|--------------|---------------|-------------------|-------|
| `name` | (filename-derived) | `name` | Universal; CC infers from filename |
| `description` | `description` | `description` | Universal |
| `model` | `model` | — | Model selection; Gemini ignores |
| `tools` | `tools` | `tools` | Allow list |
| `denied_tools` | `disallowedTools` | — | Deny list; Gemini ignores |
| `max_turns` | `maxTurns` | — | Turn limit; Gemini ignores |
| `prompt` | (YAML frontmatter + markdown body) | (YAML frontmatter + markdown body) | Agent system prompt |

The `denied_tools` → `disallowedTools` translation is the most consequential: CC's `disallowedTools` is a security enforcement mechanism. The canonical name `denied_tools` is deliberately generic and survives the addition of channels that express deny-lists differently.

### 6d. Inscription as Canonical Context File Concept (GATE 4)

The materialized context file — `CLAUDE.md`, `GEMINI.md`, or `AGENTS.md` — is called an **inscription** in knossos.

The name is precise. An inscription is authored (from rite templates), compiled (region management, satellite preservation, template rendering), and preserved (idempotent materialization, satellite content never destroyed). These properties distinguish inscriptions from configuration files, which are merely written. The name also separates the knossos concept from its channel-specific artifacts: when a developer says "inscription," they mean the knossos-owned concept; when they say "CLAUDE.md," they mean the CC compilation target.

**Compilation targets** for inscription:

| target_file | channel | format | notes |
|------------|---------|--------|-------|
| `CLAUDE.md` | `claude` | Markdown | Primary CC compilation target |
| `GEMINI.md` | `gemini` | Markdown | Primary Gemini compilation target |
| `AGENTS.md` | `interop` | Markdown | Third compilation target; 20+ tools support this format as interop standard |

`AGENTS.md` is a compilation output, not a canonical source. GitHub Copilot reads `CLAUDE.md`, `GEMINI.md`, and `AGENTS.md` as fallbacks. OpenAI Codex CLI uses `AGENTS.md` natively. Producing `AGENTS.md` alongside the channel-native inscriptions gives knossos-managed projects broad tool compatibility at zero additional authoring cost.

The internal concept is always "inscription." Channel-specific file names are compilation artifacts.

### 6e. Wire Format Preservation (GATE 5)

**SCAR-009**: This ADR changes knossos's internal representation — it does not change what CC or Gemini receive on the wire.

- CC continues to send `"PreToolUse"` in its hook payloads. `ClaudeAdapter.ParsePayload()` translates this to `EventPreTool` at the boundary. Nothing changes for CC users or CC hook payloads.
- Gemini continues to send `"BeforeTool"`. `GeminiAdapter.ParsePayload()` translates this to `EventPreTool`. Nothing changes for Gemini users.
- `hooks.yaml` event names change from CC wire values (`PreToolUse`) to knossos canonical names (`pre_tool`). The `BuildHooksSettings()` function translates outbound to the appropriate wire format per channel. Existing `settings.local.json` files are regenerated on next `ari sync` — no manual migration required.
- `TranslateInboundEvent()` and `TranslateOutboundEvent()` functions are renamed or supplemented to use canonical names internally; their external behavior (translating between knossos canonical and channel wire formats) is unchanged.

Existing deployments will not observe behavior changes after upgrading. The change is entirely internal to knossos's representation layer.

---

## Alternatives Considered

### (a) Maintain CC-Canonical (ADR-0031 Status Quo)

Keep CC wire names (`PreToolUse`, `Bash`, `Grep`) as knossos's internal vocabulary. Gemini and future channels translate at the adapter boundary.

**Pros**: No migration effort; developer familiarity with CC names is preserved; zero risk of regression.

**Cons**: CC names are now a liability, not a convenience. At two channels, CC-canonical is a shortcut. At N channels, it is a structural bias that makes every new channel a second-class citizen. It also defeats knossos's positioning as a harness-agnostic meta-framework — the internal vocabulary would literally be one harness's wire names.

Rejected: the cost is low now and compounds with each new channel. The time to establish knossos-owned vocabulary is before N > 2, not after.

### (b) Per-Harness Internal Vocabularies

Let each channel package use its own names. `internal/hook/adapter_claude.go` uses CC names; `internal/hook/adapter_gemini.go` uses Gemini names. Cross-channel code takes whichever it needs.

**Pros**: No canonical naming decisions required; each package is idiomatic to its harness.

**Cons**: No single source of truth. Code that spans channels — session management, hook dispatch, the `Env` struct — has no stable vocabulary to anchor to. Refactoring cross-channel behavior requires context-switching between naming conventions. This is the problem ADR-0031's `Env` struct already solved at the struct level; this alternative un-solves it at the naming level.

Rejected: contradicts the premise of a shared `Env` struct and shared hook command pipeline.

### (c) MCP Tool Names as Canonical

Use MCP (Model Context Protocol) tool names as the cross-harness canonical vocabulary for tools. MCP defines names like `read_file`, `execute_code` that appear across multiple harnesses.

**Pros**: External standard with growing adoption; not owned by any single harness vendor.

**Cons**: MCP serves a different abstraction level — it is a tool-calling protocol between a host and a tool server, not a lifecycle event vocabulary. MCP has no event names at all (no equivalent of `pre_tool`, `session_start`). For tools, MCP names overlap partially but not completely with knossos's needs (e.g., MCP has no `write_todos` or `activate_skill`). Using MCP names for tools while inventing knossos names for events produces an inconsistent vocabulary.

Rejected: MCP is a valuable integration point but the wrong source for a complete vocabulary. knossos should produce MCP-compatible artifacts (it already does via `AGENTS.md`), not derive its identity from MCP's naming.

---

## Rationale

**snake_case wins.** The vocabulary audit found snake_case used by 5+ harnesses and the MCP ecosystem. CC's PascalCase is the outlier. A knossos canonical vocabulary aligned with the majority convention will map more naturally to future channels than CC's convention.

**Vocabulary audit evidence.** The 10-harness audit in `SPIKE-harness-vocabulary-audit.md` confirmed that no harness shares CC's event naming scheme and that snake_case dominates tool naming. A CC-canonical vocabulary would require translation tables for 9 out of 10 harnesses surveyed. A snake_case knossos vocabulary requires translation tables for 1 (CC itself).

**CC names as translations, not canonical.** The principle established by this ADR: when knossos code refers to a lifecycle event, it uses `pre_tool`. When the CC adapter sends a hook payload, it translates `pre_tool` to `PreToolUse`. The translation is at the boundary; the identity is in the core. This is what makes knossos a meta-framework rather than a CC wrapper.

**Gemini-exclusive events become first-class.** `pre_model` and `post_model` have no CC equivalent. Under CC-canonical naming, they would be second-class additions. Under knossos-owned naming, they are members of the canonical set alongside CC-dominant events.

**AGENTS.md is additive, not disruptive.** Adding `AGENTS.md` as a third compilation target costs nothing to authors — rite templates already produce the inscription content. It extends knossos-managed project compatibility to 20+ tools at zero authoring cost.

---

## Consequences

### Positive

- Adding a new channel requires three mechanical artifacts: implement `TargetChannel`, implement `ChannelCompiler`, add event/tool translation tables. No core vocabulary changes.
- `hooks.yaml` becomes readable without knowledge of CC wire names. Developers see `pre_tool` and understand it; they do not need to know CC calls it `PreToolUse`.
- Gemini-exclusive events (`pre_model`, `post_model`) have first-class canonical names.
- knossos's public positioning as a harness-agnostic meta-framework is internally consistent with its vocabulary.
- `AGENTS.md` compilation target extends project compatibility to 20+ tools.

### Negative

- **PKG-006 migration effort: 132 hours** across 10 packages. The `HookEvent` constants, `hooks.yaml` event names, hook command source files, and all `switch` statements on event values require mechanical but non-trivial updates.
- Developer context shift: engineers familiar with CC names must learn to use knossos canonical names in internal code. The CC names remain visible in translation tables and adapter source — the shift is in which names are "correct" in core code.
- Documentation update: ADR-0031's event/tool mapping tables used CC names as the left-hand column. Those tables, and any downstream documentation citing them, require updates.

### Neutral

- Wire format is unchanged for all existing deployments. Upgrading to a knossos version implementing ADR-0032 requires no changes to CC or Gemini hook configurations.
- The `Env` struct's field names (already abstracting over channels) are unaffected. This ADR changes constant values and translation table keys, not the `Env` struct shape.
- Per-channel provenance manifests (ADR-0031) are unaffected. Channel identity (`claude`, `gemini`) is already canonical — those strings do not change.

---

## Migration Path

Implementation decomposes into five phases across 10 packages. PKG-006 (hook event layer) is the centerpiece because the `HookEvent` constants and their string values are the most widely referenced artifact this ADR changes.

| phase | sprint | packages | description |
|-------|--------|----------|-------------|
| 1 | I1 | `internal/hook` | Rename `HookEvent` constants; change string values from CC wire to canonical. Add `TranslateInboundEvent()` and `TranslateOutboundEvent()` for both adapters. |
| 2 | I1 | `internal/hook`, hooks.yaml | Update `hooks.yaml` event names from CC wire to canonical. Update `BuildHooksSettings()` to translate canonical → channel wire on output. |
| 3 | I2 | `internal/channel` | Update tool constants from CC names to canonical. Update `TranslateMatcherForChannel()`. |
| 4 | I2 | `internal/materialize` | Update compiler output for `denied_tools` → `disallowedTools` CC translation. Add `AGENTS.md` compilation target. |
| 5 | I3 | `internal/cmd/hook/*`, `knossos/templates/` | Update all hook command source and template references to use canonical names. |

Full migration detail in `SPRINT-PLAN.md`. Each phase is independently testable — the adapter boundary means phases 1–2 can ship without touching hook command source.

---

## Machine-Readable Vocabulary Extract (APPENDIX)

This YAML block is the authoritative machine-readable form of the canonical vocabulary. It is consumed directly by Sprint I1 (PKG-006), Sprint I2 (PKG-007), and Sprint I3 (PKG-008, PKG-009) to generate or validate Go constants, translation tables, and template outputs.

```yaml
# ADR-0032 Canonical Vocabulary — machine-readable extract
# Consumed by: Sprint I1 (PKG-006), Sprint I2 (PKG-007), Sprint I3 (PKG-008/PKG-009)
# Do not truncate or paraphrase — implementation sprints parse this block directly.

events:
  - canonical: pre_tool
    go_constant: EventPreTool
    cc_wire: PreToolUse
    gemini_wire: BeforeTool
    direction: bidirectional
  - canonical: post_tool
    go_constant: EventPostTool
    cc_wire: PostToolUse
    gemini_wire: AfterTool
    direction: bidirectional
  - canonical: post_tool_failure
    go_constant: EventPostToolFailure
    cc_wire: PostToolUseFailure
    gemini_wire: null
    direction: outbound_only
  - canonical: permission_request
    go_constant: EventPermissionRequest
    cc_wire: PermissionRequest
    gemini_wire: null
    direction: outbound_only
  - canonical: stop
    go_constant: EventStop
    cc_wire: Stop
    gemini_wire: null
    direction: outbound_only
  - canonical: session_start
    go_constant: EventSessionStart
    cc_wire: SessionStart
    gemini_wire: SessionStart
    direction: bidirectional
  - canonical: session_end
    go_constant: EventSessionEnd
    cc_wire: SessionEnd
    gemini_wire: SessionEnd
    direction: bidirectional
  - canonical: pre_prompt
    go_constant: EventPrePrompt
    cc_wire: UserPromptSubmit
    gemini_wire: BeforeAgent
    direction: bidirectional
  - canonical: pre_compact
    go_constant: EventPreCompact
    cc_wire: PreCompact
    gemini_wire: PreCompress
    direction: bidirectional
  - canonical: subagent_start
    go_constant: EventSubagentStart
    cc_wire: SubagentStart
    gemini_wire: null
    direction: outbound_only
  - canonical: subagent_stop
    go_constant: EventSubagentStop
    cc_wire: SubagentStop
    gemini_wire: null
    direction: outbound_only
  - canonical: notification
    go_constant: EventNotification
    cc_wire: Notification
    gemini_wire: Notification
    direction: bidirectional
  - canonical: teammate_idle
    go_constant: EventTeammateIdle
    cc_wire: TeammateIdle
    gemini_wire: null
    direction: cc_only
  - canonical: task_completed
    go_constant: EventTaskCompleted
    cc_wire: TaskCompleted
    gemini_wire: null
    direction: cc_only
  - canonical: pre_model
    go_constant: EventPreModel
    cc_wire: null
    gemini_wire: BeforeModel
    direction: gemini_only
  - canonical: post_model
    go_constant: EventPostModel
    cc_wire: null
    gemini_wire: AfterModel
    direction: gemini_only
  - canonical: worktree_create
    go_constant: EventWorktreeCreate
    cc_wire: WorktreeCreate
    gemini_wire: null
    direction: cc_only
  - canonical: worktree_remove
    go_constant: EventWorktreeRemove
    cc_wire: WorktreeRemove
    gemini_wire: null
    direction: cc_only

tools:
  - canonical: read_file
    cc_name: Read
    cc_name_alt: ReadFiles
    gemini_name: read_file
  - canonical: edit_file
    cc_name: Edit
    gemini_name: replace
  - canonical: write_file
    cc_name: Write
    gemini_name: write_file
  - canonical: list_files
    cc_name: Glob
    gemini_name: glob
  - canonical: search_content
    cc_name: Grep
    gemini_name: grep_search
  - canonical: run_shell
    cc_name: Bash
    gemini_name: run_shell_command
  - canonical: web_search
    cc_name: WebSearch
    gemini_name: google_web_search
  - canonical: web_fetch
    cc_name: WebFetch
    gemini_name: web_fetch
  - canonical: write_todos
    cc_name: TodoWrite
    gemini_name: write_todos
  - canonical: activate_skill
    cc_name: Skill
    gemini_name: activate_skill
  - canonical: delegate
    cc_name: Task
    gemini_name: null

agent_keys:
  - canonical: name
    cc_equivalent: null
    gemini_equivalent: name
    notes: "CC infers name from filename; Gemini uses explicit name field"
  - canonical: description
    cc_equivalent: description
    gemini_equivalent: description
    notes: "Universal"
  - canonical: model
    cc_equivalent: model
    gemini_equivalent: null
    notes: "Gemini ignores; CC honors"
  - canonical: tools
    cc_equivalent: tools
    gemini_equivalent: tools
    notes: "Allow list; universal"
  - canonical: denied_tools
    cc_equivalent: disallowedTools
    gemini_equivalent: null
    notes: "Deny list; Gemini ignores"
  - canonical: max_turns
    cc_equivalent: maxTurns
    gemini_equivalent: null
    notes: "Gemini ignores"
  - canonical: prompt
    cc_equivalent: null
    gemini_equivalent: null
    notes: "YAML frontmatter + markdown body; both channels share this format"

compilation_targets:
  - name: CLAUDE.md
    channel: claude
    format: markdown
    notes: "Primary CC inscription target"
  - name: GEMINI.md
    channel: gemini
    format: markdown
    notes: "Primary Gemini inscription target"
  - name: AGENTS.md
    channel: interop
    format: markdown
    notes: "Third compilation target; 20+ tools support this format; GitHub Copilot reads CLAUDE.md, GEMINI.md, and AGENTS.md as fallbacks"
```

---

## Related Decisions

- **ADR-0031** (Multi-Channel Architecture): This ADR amends ADR-0031, replacing CC-canonical event and tool names with knossos-owned canonical vocabulary. The channel abstraction stack (TargetChannel, ChannelCompiler, LifecycleAdapter interfaces) is unchanged — only the vocabulary those interfaces operate on changes.
- **ADR-0026** (Provenance): Unaffected. Channel names (`claude`, `gemini`) used in manifest paths are already canonical and do not change.
- **ADR-0030** (Processions): Unaffected. Processions operate above the channel layer and are channel-agnostic by design.

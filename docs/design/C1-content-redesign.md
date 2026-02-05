# C1: CLAUDE.md Content Re-engineering Design

## Status

Draft | 2026-02-05

## Scope

Re-engineer the content of all CLAUDE.md files and their generator/template sources for maximum agent consumption quality. Every word in L0 must earn its token cost. This document specifies exact replacement text for every source-repo reference, exact target content for all 9 sections, satellite content, parent files, and template variable additions.

C2 implements from this spec. No code or template changes in this document.

---

## Part 1: Source-Repo Reference Fixes

22 FIX items from `docs/audits/audit-source-repo-references.md`. Each fix specifies exact replacement text.

### Generator Defaults (`internal/inscription/generator.go`)

**FIX-01**: `PRD-hybrid-session-model` in `getDefaultExecutionModeContent()` (line 398)

- **Current**: `This project supports three operating modes (see PRD-hybrid-session-model for details):`
- **Replacement**: `Three operating modes:`
- **Rationale**: Source-repo PRD reference. Satellites do not have this document. The mode table is self-explanatory.

**FIX-03/04**: `cd ariadne && just build` and `go test ./...` in `getDefaultPlatformInfrastructureContent()` (line 439)

- **Current**: `` Build: `cd ariadne && just build` | Test: `go test ./...` ``
- **Replacement**: Remove line from generator default entirely. Add conditional rendering in template via `KnossosVars`.
- **Rationale**: Source-repo build commands. Satellites have different build systems. The generator default fires when no template exists and no KnossosVars are available -- the correct behavior is to omit the line. The template uses a conditional to render it when KnossosVars are set.

### Templates (`knossos/templates/sections/`)

**FIX-02**: `PRD-hybrid-session-model` in `execution-mode.md.tpl` (line 6)

- **Current**: `This project supports three operating modes (see PRD-hybrid-session-model for details):`
- **Replacement**: `Three operating modes:`

**FIX-05/06**: Build/test commands in `platform-infrastructure.md.tpl` (line 8)

- **Current**: `` Build: `cd ariadne && just build` | Test: `go test ./...` ``
- **Replacement**: `{{ if index .KnossosVars "build_command" }}Build: {{ backtick }}{{ index .KnossosVars "build_command" }}{{ backtick }} | Test: {{ backtick }}{{ index .KnossosVars "test_command" }}{{ backtick }}{{ end }}`
- **Note**: Actual Go template syntax uses `{{ .KnossosVars.build_command }}` if the map access returns zero value for missing keys (which it does for `map[string]string`). The exact Go template syntax is specified in Part 5.

### Legacy Sections Without Templates (15 FIX items resolved by section removal)

Sprint 1 removed these sections from the manifest and section order. No template or generator default exists for them. The 15 source-repo references within them are resolved by removal:

| FIX IDs | Section | Source-Repo References | Resolution |
|---------|---------|----------------------|------------|
| 07, 08, 09, 10 | `knossos-identity` | `roster/.claude/ IS Knossos`, 3x `docs/` paths | Section removed. Content available at L2 via `ecosystem-ref` skill. |
| 11, 12 | `skills` | `roster`, `.claude/skills/` | Replaced by `commands` section. |
| 13, 14 | `ariadne-cli` | `cd ariadne && just build`, `docs/guides/` | Absorbed into `platform-infrastructure`. |
| 15, 16, 17, 18 | `getting-help` | `ecosystem-ref` "Roster", 3x `docs/` paths | Replaced by `navigation`. |
| 19, 20, 21 | `state-management` | `.claude/hooks/lib/`, `user-agents/moirai.md`, `docs/philosophy/` | Absorbed into `platform-infrastructure`. |

**FIX-22**: Duplicate of FIX-01 (CLAUDE.md instance). Resolved by template/generator fix.

### Fix Summary

| Category | Count | Resolution |
|----------|-------|------------|
| Section removed (Sprint 1) | 15 | No action needed -- sections no longer in generation path |
| Template/generator text edit | 5 | FIX-01 through FIX-06 |
| Replaced by `commands` section | 2 | FIX-11, FIX-12 |

---

## Part 2: Section Content Rewrites

Each section shows the exact target markdown between KNOSSOS markers.

### Context-Engineering Principles Applied

1. **Imperative voice** for behavioral rules. Not "This project supports" but "Three operating modes."
2. **No explanatory prose**. What and where, not why.
3. **Backtick-code** for every tool name, path, and command.
4. **No self-reference**. Eliminated "This section..." and "This project..." framing.
5. **Token justification**. Every line must prevent a wrong decision or point to needed context.

### Section 1: `execution-mode` (10 content lines)

Owner: knossos | Region name unchanged

```markdown
<!-- KNOSSOS:START execution-mode -->
## Execution Mode

Three operating modes:

| Mode | Session | Rite | Behavior |
|------|---------|------|----------|
| **Native** | No | -- | Direct execution, no tracking |
| **Cross-Cutting** | Yes | No | Direct execution + session tracking |
| **Orchestrated** | Yes | Yes (ACTIVE) | Delegate via Task tool to specialist agents |

Use `/consult` if unsure. Enforcement rules: `orchestration/execution-mode.md`
<!-- KNOSSOS:END execution-mode -->
```

**Changes with rationale**:
- Removed `(see PRD-hybrid-session-model for details)` -- source-repo reference (FIX-01/02), zero behavioral value.
- `This project supports three operating modes` to `Three operating modes` -- self-referential framing; agent already knows the project.
- Column `Team` to `Rite` -- the column describes rite activation, not team presence. Template already had this correction.
- Column `Main Agent Behavior` to `Behavior` -- redundant qualifier.
- `Coach pattern, delegate via Task tool` to `Delegate via Task tool to specialist agents` -- "coach pattern" is internal jargon with no L0 decision value; the concrete action is delegation.
- Collapsed two footer lines into one -- `**Unsure?**` rhetorical question addresses a human, not an agent.

### Section 2: `quick-start` (variable, ~12-15 lines)

Owner: regenerate | Source: `ACTIVE_RITE+agents` | Region name unchanged

**With active rite** (template output):
```markdown
<!-- KNOSSOS:START quick-start regenerate=true source=ACTIVE_RITE+agents -->
## Quick Start

{{ .AgentCount }}-agent workflow ({{ .ActiveRite }}):

[agent table]

For invocation patterns: `prompting`. For new initiatives: `initiative-scoping`.
<!-- KNOSSOS:END quick-start -->
```

**Without active rite** (template output):
```markdown
<!-- KNOSSOS:START quick-start regenerate=true source=ACTIVE_RITE+agents -->
## Quick Start

No active rite. Run `ari team switch <rite-name>` to activate.

Use `/consult` to get started.
<!-- KNOSSOS:END quick-start -->
```

**Generator default** (no render context):
```markdown
## Quick Start

Multi-agent workflow:

| Agent | Role | Produces |
| ----- | ---- | -------- |

For invocation patterns: `prompting`. For new initiatives: `initiative-scoping`.
```

**Changes with rationale**:
- Removed `This project uses a` prefix -- self-referential.
- `**New here?** Use the \`prompting\` skill for copy-paste patterns, or \`initiative-scoping\` to start a new project.` to `For invocation patterns: \`prompting\`. For new initiatives: \`initiative-scoping\`.` -- "New here?" is human-addressed personality; colon-pointer format matches rest of redesign.
- Template: `pantheon` to `workflow` -- mythology term has no decision value at L0.

### Section 3: `agent-routing` (4 content lines)

Owner: knossos | Region name unchanged

```markdown
<!-- KNOSSOS:START agent-routing -->
## Agent Routing

Orchestrated sessions: delegate to specialists via Task tool. No session: execute directly or use `/task`.

Routing guidance: `/consult`
<!-- KNOSSOS:END agent-routing -->
```

**Changes with rationale**:
- Compressed from 2 content lines (36 words) to 1 line (17 words). Same behavioral content: delegate in orchestrated mode, execute directly otherwise.
- `For routing guidance:` to `Routing guidance:` -- dropped filler preposition.

### Section 4: `commands` (8 content lines)

Owner: knossos | Region name unchanged

```markdown
<!-- KNOSSOS:START commands -->
## Commands

Invoke via the **Skill tool**. Two types:

- **Invokable** (`/name`): User-callable actions (`/start`, `/commit`, `/pr`)
- **Reference** (auto-loaded): Domain knowledge (`prompting`, `doc-artifacts`, `standards`)

Key references: `prompting` (agent patterns), `doc-artifacts` (PRD/TDD/ADR schemas), `standards` (conventions), `session/common` (lifecycle).

Full list: `.claude/commands/`
<!-- KNOSSOS:END commands -->
```

**Changes with rationale**:
- `Commands are invoked via the **Skill tool**. Two types exist:` to `Invoke via the **Skill tool**. Two types:` -- dropped self-referential subject and filler verb.
- `like` replaced with parenthesized inline examples -- tighter.
- `Key reference commands:` to `Key references:` -- shorter.
- Kept the key references line. Rationale: unlike a full enumeration, this 4-item pointer list has high L0 value because these are the commands agents reach for most often. Removing it forces agents to list the commands directory before knowing where to find schemas or conventions. The token cost (~25 tokens) is justified by the frequency of use.
- `See .claude/commands/ for the full list.` to `Full list: .claude/commands/` -- pointer format.

### Section 5: `agent-configurations` (variable, ~8 content lines)

Owner: regenerate | Source: `agents/*.md` | Region name unchanged

**With agents** (template output):
```markdown
<!-- KNOSSOS:START agent-configurations regenerate=true source=agents/*.md -->
## Agent Configurations

Agent prompts in `.claude/agents/`:

- `{file}` - {role}
...
<!-- KNOSSOS:END agent-configurations -->
```

**Without agents** (template output):
```markdown
<!-- KNOSSOS:START agent-configurations regenerate=true source=agents/*.md -->
## Agent Configurations

No agents installed. Run `ari team switch <rite-name>` to install.
<!-- KNOSSOS:END agent-configurations -->
```

**Generator default** (no render context):
```markdown
## Agent Configurations

Agent prompts in `.claude/agents/`.
```

**Changes with rationale**:
- `Full agent prompts live in` to `Agent prompts in` -- dropped filler words. Kept the heading as `Agent Configurations` rather than renaming to `Agents`, because the region name is `agent-configurations` and heading renames create unnecessary risk of breaking anchor references or agent prompt references for zero material gain.
- Removed italics from empty-state text -- markdown italics are rendering decoration, not agent-useful.

### Section 6: `platform-infrastructure` (3-4 content lines)

Owner: knossos | Region name unchanged

**Template output** (with KnossosVars):
```markdown
<!-- KNOSSOS:START platform-infrastructure -->
## Platform Infrastructure

Hooks auto-inject session context. CLI reference: `ari --help`.
Mutate `*_CONTEXT.md` only via `Task(moirai, "...")`.
Build: `cd ariadne && just build` | Test: `go test ./...`
<!-- KNOSSOS:END platform-infrastructure -->
```

**Template output** (without KnossosVars):
```markdown
<!-- KNOSSOS:START platform-infrastructure -->
## Platform Infrastructure

Hooks auto-inject session context. CLI reference: `ari --help`.
Mutate `*_CONTEXT.md` only via `Task(moirai, "...")`.
<!-- KNOSSOS:END platform-infrastructure -->
```

**Generator default** (no template):
```markdown
## Platform Infrastructure

Hooks auto-inject session context. CLI reference: `ari --help`.
Mutate `*_CONTEXT.md` only via `Task(moirai, "...")`.
```

**Changes with rationale**:
- Kept heading as `Platform Infrastructure` -- the region name is `platform-infrastructure` and the heading should match. Renaming adds compatibility risk for 1 word of savings.
- `(no manual loading)` removed -- "auto-inject" already conveys this.
- `CLI operations: run \`ari --help\`` to `CLI reference: \`ari --help\`` -- "operations" is vague; "reference" is the use case.
- `State management: use \`Task(moirai, "...")\` for all \`*_CONTEXT.md\` changes.` to `Mutate \`*_CONTEXT.md\` only via \`Task(moirai, "...")\`.` -- imperative constraint. "State management:" is a label that does not aid agent behavior; "Mutate X only via Y" is a direct rule.
- Build/test line: hardcoded values replaced with conditional KnossosVars in template. Generator default omits the line. Source repo sets vars in manifest.

### Section 7: `navigation` (3 content lines)

Owner: knossos | Region name unchanged

```markdown
<!-- KNOSSOS:START navigation -->
## Navigation

Workflow routing: `/consult`. Domain knowledge: Skill tool. Architecture: `MEMORY.md`.
<!-- KNOSSOS:END navigation -->
```

**Changes with rationale**:
- Added `Architecture: \`MEMORY.md\`` -- MEMORY.md is auto-loaded by Claude Code but agents benefit from an explicit pointer when they need architecture details. Costs ~4 tokens, prevents agents from using Grep/Glob to rediscover file locations that MEMORY.md already catalogs.
- Reformatted to label:value single-line format for consistency with rest of redesign.

### Section 8: `slash-commands` (3 content lines)

Owner: knossos | Region name unchanged

```markdown
<!-- KNOSSOS:START slash-commands -->
## Slash Commands

Always respond with outcome. "No response" is never correct for explicit user requests.
<!-- KNOSSOS:END slash-commands -->
```

No changes. Already optimal: 3 lines, pure behavioral constraint, no source-repo references.

### Section 9: `user-content` (satellite, variable)

Owner: satellite | Region name unchanged

**Template** (for new satellites):
```markdown
<!-- KNOSSOS:START user-content -->
## Project-Specific Instructions

<!-- Add project conventions, anti-patterns, and active work here.
     This section is preserved during sync. -->
<!-- KNOSSOS:END user-content -->
```

**Change from current template**: Compressed the 11-line placeholder comment to 2 lines. The original had an example list (`Examples: - Project conventions...`) that tells users what they already know. The brief comment is sufficient.

---

### Knossos-Managed Line Budget

| Section | Content Lines | With Markers + Blank |
|---------|-------------|---------------------|
| execution-mode | 10 | 13 |
| quick-start (6 agents) | 14 | 17 |
| agent-routing | 4 | 7 |
| commands | 8 | 11 |
| agent-configurations (6 agents) | 9 | 12 |
| platform-infrastructure (with build) | 4 | 7 |
| navigation | 3 | 6 |
| slash-commands | 3 | 6 |
| **Subtotal** | **55** | **79** |

79 lines for knossos-managed content (under the 90-line target).

---

## Part 3: Satellite Content Specification (Roster Project)

Exact content for the `user-content` region in `/Users/tomtenuta/Code/roster/.claude/CLAUDE.md`. This is the source repo's satellite-owned content.

```markdown
<!-- KNOSSOS:START user-content -->
## Project-Specific Instructions

### Decision Records
Use `/stamp` on significant workflow decisions. Triggers:
- Sacred path edits (`.claude/`, `*_CONTEXT.md`, `docs/decisions/`)
- Repeated failures (same command failed 2+ times)
- Multi-file changes (5+ files modified)

### Anti-Patterns
1. Direct writes to `*_CONTEXT.md` -- use Moirai agent
2. Editing knossos-owned CLAUDE.md sections -- lost on sync
3. Skipping `/stamp` on significant decisions -- audit trail gaps
4. Shipping GRAY without QA acknowledgment -- false confidence
5. Swap-rite for temporary skill needs -- use invoke instead

### Active Refactoring (2026-02)
- `skills/` -> `commands/` unification complete (ADR-0021)
- User sync system (`internal/usersync/`) is new
- Context tier model in progress

### Context Loading
Skills: Skill tool. Agent prompts: Task tool invocation. Docs: Read tool.
Architecture: `MEMORY.md`. Ecosystem patterns: `ecosystem-ref` skill.
<!-- KNOSSOS:END user-content -->
```

**Line count**: 23 lines (including markers). Under the 50-line satellite target.

**Changes from current (24 lines)**:
- `When making significant workflow decisions, use /stamp to record rationale in the clew.` to `Use /stamp on significant workflow decisions.` -- dropped "to record rationale in the clew" (mechanism detail, not needed at L0).
- `Anti-Patterns to Avoid` to `Anti-Patterns` -- "to Avoid" is redundant.
- Removed bold from anti-pattern labels. The numbered list provides structure; bold on every item is visual noise that does not aid agent parsing.
- `Active Refactoring` to `Active Refactoring` -- kept as-is (unchanged, the heading communicates temporal scope).
- Removed `(TDD-context-tier-model)` parenthetical -- source-repo TDD reference.
- `Load on demand: Skills via Skill tool, agent prompts on Task invocation, docs via Read tool.` reformatted to key-value pairs for scanability.
- `Architecture details in MEMORY.md and ecosystem-ref skill.` to `Architecture: MEMORY.md. Ecosystem patterns: ecosystem-ref skill.` -- pointer format.

---

## Part 4: Parent CLAUDE.md Specifications

### `~/.claude/CLAUDE.md` (6 lines)

**Status**: Already at target. No changes needed.

```markdown
# Global Preferences

- Use Go conventions (gofmt, golint) for Go projects
- Prefer editing existing files over creating new files
- No unnecessary documentation files unless requested
- When in a Knossos project, follow project-level CLAUDE.md instructions
```

Estimated: ~65 tokens.

### `~/Code/.claude/CLAUDE.md` (5 lines)

**Status**: Already at target. No changes needed.

```markdown
# Code Directory Preferences

- Project-specific CLAUDE.md files always take precedence
- Use language-appropriate conventions for each project
- Prefer standard library solutions when possible
```

Estimated: ~50 tokens.

### `~/CLAUDE.md` (143 lines)

**Status**: Outside Knossos domain authority. Contains general-purpose Claude Code guidance (language conventions, git workflow, tool usage, security practices). No overlap with knossos-managed CLAUDE.md content. Out of scope for C1.

**Estimated**: ~950 tokens/turn.

---

## Part 5: Template Variable Additions

### New KnossosVars

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `build_command` | string | (absent) | Project-specific build command. Rendered in `platform-infrastructure` only when present. |
| `test_command` | string | (absent) | Project-specific test command. Rendered in `platform-infrastructure` only when present. |

### Design Decision: Conditional KnossosVars vs. Remove Entirely

**Chosen**: Conditional rendering via KnossosVars.

**Rejected**: Remove build/test from knossos-managed content entirely (place in satellite `user-content` only).

**Rationale**: The build/test line has high L0 value for agents working in a project -- it prevents them from guessing or searching for build commands. Removing it entirely forces every satellite to independently add this to their `user-content`, which is worse than providing an optional template variable. The conditional pattern means:
- Source repo: sets `build_command: "cd ariadne && just build"` and `test_command: "go test ./..."` in manifest -- gets the line.
- Satellites without vars: get no build/test line -- clean, no source-repo leakage.
- Satellites with vars: set their own commands -- get their own line.

This is the standard extensibility pattern for project-specific content in knossos-managed sections.

### Source-Repo Manifest Addition

Add to `.claude/KNOSSOS_MANIFEST.yaml`:

```yaml
knossos_vars:
  build_command: "cd ariadne && just build"
  test_command: "go test ./..."
```

### Template Syntax

The `platform-infrastructure.md.tpl` template uses Go template conditionals:

```
{{ $build := index .KnossosVars "build_command" }}{{ if $build }}
Build: `{{ $build }}` | Test: `{{ index .KnossosVars "test_command" }}`
{{ end }}
```

This works because `index .KnossosVars "build_command"` returns empty string for absent keys in Go's `map[string]string`, and empty string is falsy in Go templates. The existing `RenderContext.KnossosVars` field already supports this.

### Generator Default Alignment

`getDefaultPlatformInfrastructureContent()` omits the build/test line entirely. This is correct:
- Generator default fires when no template exists and no KnossosVars are available.
- Without KnossosVars, there is no project-specific build command to render.
- The template path (which has access to KnossosVars) handles the conditional.

---

## Part 6: Implementation Specification for C2

### File Change Matrix

| File | Action | Changes |
|------|--------|---------|
| `internal/inscription/generator.go` | EDIT | Rewrite 7 default content methods + `generateQuickStartContent()` |
| `knossos/templates/sections/execution-mode.md.tpl` | EDIT | Remove PRD ref, tighten content per Section 1 |
| `knossos/templates/sections/quick-start.md.tpl` | EDIT | Replace footer per Section 2 |
| `knossos/templates/sections/agent-routing.md.tpl` | EDIT | Compress per Section 3 |
| `knossos/templates/sections/commands.md.tpl` | EDIT | Tighten per Section 4 |
| `knossos/templates/sections/agent-configurations.md.tpl` | EDIT | Tighten header per Section 5 |
| `knossos/templates/sections/platform-infrastructure.md.tpl` | EDIT | Add conditional build/test per Section 6 |
| `knossos/templates/sections/navigation.md.tpl` | EDIT | Add MEMORY.md pointer per Section 7 |
| `knossos/templates/sections/slash-commands.md.tpl` | NONE | Already optimal |
| `knossos/templates/sections/user-content.md.tpl` | EDIT | Compress placeholder comments |
| `knossos/templates/CLAUDE.md.tpl` | NONE | Master template correct from Sprint 1 |
| `.claude/KNOSSOS_MANIFEST.yaml` | EDIT | Add `knossos_vars` block |
| `.claude/CLAUDE.md` | REGENERATE + MANUAL | Regenerate via sync, then manually trim `user-content` |
| `internal/inscription/manifest.go` | NONE | Already correct from Sprint 1 |

### Generator Default Methods: Exact Target Content

Each method must produce content matching its template counterpart (minus KNOSSOS markers).

**`getDefaultExecutionModeContent()`**:
```
## Execution Mode

Three operating modes:

| Mode | Session | Rite | Behavior |
|------|---------|------|----------|
| **Native** | No | -- | Direct execution, no tracking |
| **Cross-Cutting** | Yes | No | Direct execution + session tracking |
| **Orchestrated** | Yes | Yes (ACTIVE) | Delegate via Task tool to specialist agents |

Use `/consult` if unsure. Enforcement rules: `orchestration/execution-mode.md`
```

**`getDefaultAgentRoutingContent()`**:
```
## Agent Routing

Orchestrated sessions: delegate to specialists via Task tool. No session: execute directly or use `/task`.

Routing guidance: `/consult`
```

**`getDefaultCommandsContent()`**:
```
## Commands

Invoke via the **Skill tool**. Two types:

- **Invokable** (`/name`): User-callable actions (`/start`, `/commit`, `/pr`)
- **Reference** (auto-loaded): Domain knowledge (`prompting`, `doc-artifacts`, `standards`)

Key references: `prompting` (agent patterns), `doc-artifacts` (PRD/TDD/ADR schemas), `standards` (conventions), `session/common` (lifecycle).

Full list: `.claude/commands/`
```

**`getDefaultPlatformInfrastructureContent()`**:
```
## Platform Infrastructure

Hooks auto-inject session context. CLI reference: `ari --help`.
Mutate `*_CONTEXT.md` only via `Task(moirai, "...")`.
```

**`getDefaultNavigationContent()`**:
```
## Navigation

Workflow routing: `/consult`. Domain knowledge: Skill tool. Architecture: `MEMORY.md`.
```

**`getDefaultSlashCommandsContent()`**: No change.

**`getDefaultQuickStartContent()`**:
```
## Quick Start

Multi-agent workflow:

| Agent | Role | Produces |
| ----- | ---- | -------- |

For invocation patterns: `prompting`. For new initiatives: `initiative-scoping`.
```

**`getDefaultAgentConfigsContent()`**:
```
## Agent Configurations

Agent prompts in `.claude/agents/`.
```

**`generateQuickStartContent()`**: Update footer to match template. Remove `This project uses a` when `ActiveRite` is set -- use `{N}-agent workflow ({rite}):` directly.

### Template Files: Exact Target Content

**`execution-mode.md.tpl`**:
```
{{/* execution-mode section template */}}
{{/* Owner: knossos - Always synced from Knossos templates */}}
<!-- KNOSSOS:START execution-mode -->
## Execution Mode

Three operating modes:

| Mode | Session | Rite | Behavior |
|------|---------|------|----------|
| **Native** | No | -- | Direct execution, no tracking |
| **Cross-Cutting** | Yes | No | Direct execution + session tracking |
| **Orchestrated** | Yes | Yes (ACTIVE) | Delegate via Task tool to specialist agents |

Use `/consult` if unsure. Enforcement rules: `orchestration/execution-mode.md`
<!-- KNOSSOS:END execution-mode -->
```

**`quick-start.md.tpl`**:
```
{{/* quick-start section template */}}
{{/* Owner: regenerate - Generated from ACTIVE_RITE + agents/ */}}
<!-- KNOSSOS:START quick-start regenerate=true source=ACTIVE_RITE -->
## Quick Start

{{- if .ActiveRite }}
{{ .AgentCount }}-agent workflow ({{ .ActiveRite }}):

{{include "partials/agent-table.md.tpl"}}

For invocation patterns: `prompting`. For new initiatives: `initiative-scoping`.
{{- else }}
No active rite. Run `ari team switch <rite-name>` to activate.

Use `/consult` to get started.
{{- end }}
<!-- KNOSSOS:END quick-start -->
```

**`agent-routing.md.tpl`**:
```
{{/* agent-routing section template */}}
{{/* Owner: knossos - Always synced from Knossos templates */}}
<!-- KNOSSOS:START agent-routing -->
## Agent Routing

Orchestrated sessions: delegate to specialists via Task tool. No session: execute directly or use `/task`.

Routing guidance: `/consult`
<!-- KNOSSOS:END agent-routing -->
```

**`commands.md.tpl`**:
```
{{/* commands section template */}}
{{/* Owner: knossos - Always synced from Knossos templates */}}
<!-- KNOSSOS:START commands -->
## Commands

Invoke via the **Skill tool**. Two types:

- **Invokable** (`/name`): User-callable actions (`/start`, `/commit`, `/pr`)
- **Reference** (auto-loaded): Domain knowledge (`prompting`, `doc-artifacts`, `standards`)

Key references: `prompting` (agent patterns), `doc-artifacts` (PRD/TDD/ADR schemas), `standards` (conventions), `session/common` (lifecycle).

Full list: `.claude/commands/`
<!-- KNOSSOS:END commands -->
```

**`agent-configurations.md.tpl`**:
```
{{/* agent-configurations section template */}}
{{/* Owner: regenerate - Generated from agents/*.md */}}
<!-- KNOSSOS:START agent-configurations regenerate=true source=agents/*.md -->
## Agent Configurations

{{- if .Agents }}
Agent prompts in `.claude/agents/`:

{{- range .Agents }}
- `{{ .FilePath }}` - {{ .Role }}
{{- end }}
{{- else }}
No agents installed. Run `ari team switch <rite-name>` to install.
{{- end }}
<!-- KNOSSOS:END agent-configurations -->
```

**`platform-infrastructure.md.tpl`**:
```
{{/* platform-infrastructure section template */}}
{{/* Owner: knossos - Always synced from Knossos templates */}}
<!-- KNOSSOS:START platform-infrastructure -->
## Platform Infrastructure

Hooks auto-inject session context. CLI reference: `ari --help`.
Mutate `*_CONTEXT.md` only via `Task(moirai, "...")`.
{{- $build := index .KnossosVars "build_command" }}{{ if $build }}
Build: `{{ $build }}` | Test: `{{ index .KnossosVars "test_command" }}`
{{- end }}
<!-- KNOSSOS:END platform-infrastructure -->
```

**`navigation.md.tpl`**:
```
{{/* navigation section template */}}
{{/* Owner: knossos - Always synced from Knossos templates */}}
<!-- KNOSSOS:START navigation -->
## Navigation

Workflow routing: `/consult`. Domain knowledge: Skill tool. Architecture: `MEMORY.md`.
<!-- KNOSSOS:END navigation -->
```

**`slash-commands.md.tpl`**: No change.

**`user-content.md.tpl`**:
```
<!-- KNOSSOS:START user-content -->
## Project-Specific Instructions

<!-- Add project conventions, anti-patterns, and active work here.
     This section is preserved during sync. -->
<!-- KNOSSOS:END user-content -->
```

---

## Part 7: Rendered Output Projection

Exact CLAUDE.md output for the roster project after C2 and inscription sync. This is the acceptance test artifact.

```markdown
<!-- KNOSSOS:START execution-mode -->
## Execution Mode

Three operating modes:

| Mode | Session | Rite | Behavior |
|------|---------|------|----------|
| **Native** | No | -- | Direct execution, no tracking |
| **Cross-Cutting** | Yes | No | Direct execution + session tracking |
| **Orchestrated** | Yes | Yes (ACTIVE) | Delegate via Task tool to specialist agents |

Use `/consult` if unsure. Enforcement rules: `orchestration/execution-mode.md`
<!-- KNOSSOS:END execution-mode -->

<!-- KNOSSOS:START quick-start regenerate=true source=ACTIVE_RITE+agents -->
## Quick Start

6-agent workflow (ecosystem):

| Agent | Role | Produces |
| ----- | ---- | -------- |
| **orchestrator** | Coordinates ecosystem infrastructure phases |  |
| **ecosystem-analyst** | Traces CEM/roster problems to root causes and produces gap analysis |  |
| **context-architect** | Designs context solutions, schemas, and ecosystem patterns |  |
| **integration-engineer** | Implements CEM and roster changes with integration tests |  |
| **documentation-engineer** | Creates migration runbooks and compatibility documentation |  |
| **compatibility-tester** | Validates ecosystem changes across satellite configurations |  |

For invocation patterns: `prompting`. For new initiatives: `initiative-scoping`.
<!-- KNOSSOS:END quick-start -->

<!-- KNOSSOS:START agent-routing -->
## Agent Routing

Orchestrated sessions: delegate to specialists via Task tool. No session: execute directly or use `/task`.

Routing guidance: `/consult`
<!-- KNOSSOS:END agent-routing -->

<!-- KNOSSOS:START commands -->
## Commands

Invoke via the **Skill tool**. Two types:

- **Invokable** (`/name`): User-callable actions (`/start`, `/commit`, `/pr`)
- **Reference** (auto-loaded): Domain knowledge (`prompting`, `doc-artifacts`, `standards`)

Key references: `prompting` (agent patterns), `doc-artifacts` (PRD/TDD/ADR schemas), `standards` (conventions), `session/common` (lifecycle).

Full list: `.claude/commands/`
<!-- KNOSSOS:END commands -->

<!-- KNOSSOS:START agent-configurations regenerate=true source=agents/*.md -->
## Agent Configurations

Agent prompts in `.claude/agents/`:

- `orchestrator.md` - Coordinates ecosystem infrastructure phases
- `ecosystem-analyst.md` - Traces CEM/roster problems to root causes and produces gap analysis
- `context-architect.md` - Designs context solutions, schemas, and ecosystem patterns
- `integration-engineer.md` - Implements CEM and roster changes with integration tests
- `documentation-engineer.md` - Creates migration runbooks and compatibility documentation
- `compatibility-tester.md` - Validates ecosystem changes across satellite configurations
<!-- KNOSSOS:END agent-configurations -->

<!-- KNOSSOS:START platform-infrastructure -->
## Platform Infrastructure

Hooks auto-inject session context. CLI reference: `ari --help`.
Mutate `*_CONTEXT.md` only via `Task(moirai, "...")`.
Build: `cd ariadne && just build` | Test: `go test ./...`
<!-- KNOSSOS:END platform-infrastructure -->

<!-- KNOSSOS:START navigation -->
## Navigation

Workflow routing: `/consult`. Domain knowledge: Skill tool. Architecture: `MEMORY.md`.
<!-- KNOSSOS:END navigation -->

<!-- KNOSSOS:START slash-commands -->
## Slash Commands

Always respond with outcome. "No response" is never correct for explicit user requests.
<!-- KNOSSOS:END slash-commands -->

<!-- KNOSSOS:START user-content -->
## Project-Specific Instructions

### Decision Records
Use `/stamp` on significant workflow decisions. Triggers:
- Sacred path edits (`.claude/`, `*_CONTEXT.md`, `docs/decisions/`)
- Repeated failures (same command failed 2+ times)
- Multi-file changes (5+ files modified)

### Anti-Patterns
1. Direct writes to `*_CONTEXT.md` -- use Moirai agent
2. Editing knossos-owned CLAUDE.md sections -- lost on sync
3. Skipping `/stamp` on significant decisions -- audit trail gaps
4. Shipping GRAY without QA acknowledgment -- false confidence
5. Swap-rite for temporary skill needs -- use invoke instead

### Active Refactoring (2026-02)
- `skills/` -> `commands/` unification complete (ADR-0021)
- User sync system (`internal/usersync/`) is new
- Context tier model in progress

### Context Loading
Skills: Skill tool. Agent prompts: Task tool invocation. Docs: Read tool.
Architecture: `MEMORY.md`. Ecosystem patterns: `ecosystem-ref` skill.
<!-- KNOSSOS:END user-content -->
```

**Total line count**: 105 lines (including markers and inter-section blanks).
- Knossos-managed (execution-mode through slash-commands): 80 lines (under 90)
- Satellite (user-content): 23 lines (under 50)
- Inter-section blank lines: 2
- **Under the 140-line constraint.**

---

## Part 8: Backward Compatibility

### Classification: COMPATIBLE

All changes are content-only edits within existing sections. No schema changes. No section additions or removals (Sprint 1 handled structural changes). No manifest format changes.

### KnossosVars Addition

The `knossos_vars` field in `KNOSSOS_MANIFEST.yaml` is an existing mechanism. Adding new keys is purely additive. Satellites without these keys in their manifest get the generator default behavior (no build/test line).

### Satellite Impact

| Scenario | Impact |
|----------|--------|
| Satellite runs `ari sync` after C2 | Knossos sections regenerated with tighter content. Satellite `user-content` preserved. |
| Satellite has no `knossos_vars` | No build/test line rendered. Clean output. |
| Satellite sets `knossos_vars.build_command` | Build/test line rendered with satellite's commands. |
| Satellite has custom `user-content` | Preserved by merger. No interference. |

### Heading Stability

No headings are renamed in this design. All section headings match their current names:
- `Execution Mode` (unchanged)
- `Quick Start` (unchanged)
- `Agent Routing` (unchanged)
- `Commands` (unchanged from Sprint 1 rename)
- `Agent Configurations` (unchanged)
- `Platform Infrastructure` (unchanged)
- `Navigation` (unchanged from Sprint 1 rename)
- `Slash Commands` (unchanged)
- `Project-Specific Instructions` (unchanged)

This eliminates the backward compatibility risk identified in the previous design revision where headings were renamed.

### Migration Path

No migration needed. Content changes apply on next `ari sync inscription`. No coordinated rollout required.

---

## Part 9: Integration Test Specifications

| ID | Satellite Type | Test | Expected Outcome |
|----|---------------|------|------------------|
| CT-01 | Fresh project | `ari sync inscription` on new project | 9-section CLAUDE.md. No build/test line. user-content has placeholder. Total < 90 lines (knossos only). |
| CT-02 | Source repo (roster) | `ari sync inscription` | Knossos sections match Part 7. Satellite preserved. Build/test line present (from KnossosVars). |
| CT-03 | Satellite with KnossosVars | Set `build_command: "npm run build"`, `test_command: "npm test"`, sync | Build/test line renders: `Build: \`npm run build\` \| Test: \`npm test\`` |
| CT-04 | Satellite without KnossosVars | Sync without `knossos_vars` | No build/test line. Platform-infrastructure has 2 content lines. |
| CT-05 | Existing satellite (pre-change) | Sync after Knossos upgrade | Knossos sections regenerated to new content. Satellite region preserved. |
| CT-06 | Content verification | Grep rendered output | Zero matches for: `PRD-hybrid-session-model`, `docs/philosophy`, `docs/guides`, `user-agents/moirai.md`, `.claude/skills/` (in knossos sections) |
| CT-07 | Line count | `wc -l` on rendered CLAUDE.md | Total < 140. Knossos-managed < 90. |
| CT-08 | Generator defaults | Call `getDefaultSectionContent()` for all 9 sections | Each returns non-empty. None contains source-repo paths (except via KnossosVars). |
| CT-09 | Build/test conditional | Render `platform-infrastructure.md.tpl` with empty KnossosVars | Output has exactly 2 content lines (hooks + moirai). No build/test line. |
| CT-10 | Build/test conditional | Render `platform-infrastructure.md.tpl` with `build_command` set | Output has exactly 3 content lines including build/test. |

---

## Design Decisions Log

| # | Decision | Chosen | Rejected | Rationale |
|---|----------|--------|----------|-----------|
| 1 | PRD reference in execution-mode | Remove entirely | (A) Make configurable via KnossosVar | Zero behavioral value. Agent does not need document provenance. The mode table is self-explanatory. |
| 2 | Build/test commands | Conditional via KnossosVars | (A) Remove entirely, place in satellite user-content | Build/test line has high L0 value (prevents agents from guessing). KnossosVars is the existing extensibility mechanism. Removal forces every satellite to independently reinvent this. |
| 3 | Heading renames | Keep all headings unchanged | (A) Rename `Agent Configurations` to `Agents`, `Platform Infrastructure` to `Platform` | Heading renames create backward compatibility risk (anchor references, agent prompt references) for zero material gain. The 1-2 word savings per heading is not worth the risk. |
| 4 | Key references line in commands | Keep the 4-item pointer list | (A) Remove, let agents discover from directory | These 4 commands (`prompting`, `doc-artifacts`, `standards`, `session/common`) are the most frequently reached-for references. The ~25 token cost prevents agents from needing a directory listing on most turns. |
| 5 | MEMORY.md pointer in navigation | Add `Architecture: \`MEMORY.md\`` | (A) Omit (agents discover from auto-injection) | 4 tokens. MEMORY.md is auto-loaded but agents benefit from explicit "look here for architecture" guidance. Prevents redundant Grep/Glob searches. |
| 6 | Parent CLAUDE.md files | No change (already minimal) | (A) Further reduce, (B) Empty them | Already at 6 and 5 lines respectively. Further cuts lose cross-project guidance that has independent value outside Knossos. |
| 7 | `~/CLAUDE.md` (143 lines) | Out of C1 scope | (A) Include in redesign | General-purpose guidance for non-Knossos projects. No content overlap with project CLAUDE.md. Requires separate audit. |
| 8 | Anti-pattern bold formatting | Remove bold | (A) Keep bold on each item | Numbered list provides sufficient structure. Bold on every line in a 5-item list is visual noise in token-constrained agent context. |
| 9 | Quick-start "New here?" framing | Replace with colon-pointers | (A) Keep personality framing | "New here?" addresses a human, not an agent. The pointers to `prompting` and `initiative-scoping` have value; the framing does not. |

---

## Token Budget Verification

| Component | Current (tok) | Target (tok) | Change |
|-----------|--------------|-------------|--------|
| `.claude/CLAUDE.md` (knossos) | ~1,820 | ~520 | -1,300 |
| `.claude/CLAUDE.md` (satellite) | ~1,280 | ~230 | -1,050 |
| `~/.claude/CLAUDE.md` | ~65 | ~65 | 0 |
| `~/Code/.claude/CLAUDE.md` | ~50 | ~50 | 0 |
| MEMORY.md | ~500 | ~500 | 0 |
| `~/CLAUDE.md` | ~950 | ~950 | 0 |
| **Total per turn** | **~4,665** | **~2,315** | **-2,350 (50%)** |

Note: Parent files were already trimmed in a prior pass. The TDD's original 7,950 token measurement reflected their pre-trimmed state. The savings in C1 come from project-level content re-engineering.

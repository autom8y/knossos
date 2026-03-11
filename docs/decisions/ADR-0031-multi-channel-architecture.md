# ADR-0031: Multi-Channel Architecture (claude + gemini dual-projection)

## Status

ACCEPTED

## Date

2026-03-12

## Context

Knossos orchestrates AI coding agents via the `ari` CLI. The initial implementation targeted Claude Code exclusively -- every artifact path, hook payload format, and template output assumed `.claude/` as the sole projection target. As the AI coding assistant landscape diversifies (Gemini CLI, future entrants), knossos must support multiple channels without forking the codebase or maintaining separate binaries.

The problem decomposes into four concerns:

1. **Identity**: Where does each channel's configuration live? (directory name, context file)
2. **Compilation**: How do rite artifacts (commands, skills, agents) transform into channel-native formats?
3. **Lifecycle**: How do hook payloads and responses normalize across channels with different wire protocols?
4. **Dispatch**: How does a single `ari sync` invocation project into multiple channels?

Each concern has different extensibility characteristics. Identity changes rarely. Compilation formats differ significantly (markdown vs. TOML). Lifecycle adapters must handle real-time bidirectional translation. Dispatch is the orchestration glue.

## Decision

Implement a channel abstraction layer that allows the same rite definitions, templates, and hooks to project into any supported AI assistant's configuration directory. The design consists of three interface layers plus a dispatch mechanism.

### Layer 1: TargetChannel Interface

**Package**: `internal/paths/channel.go`

Declares channel identity with three methods: `Name()`, `DirName()`, `ContextFile()`.

| Implementation | Name | DirName | ContextFile |
|---------------|------|---------|-------------|
| `ClaudeChannel` | `"claude"` | `".claude"` | `"CLAUDE.md"` |
| `GeminiChannel` | `"gemini"` | `".gemini"` | `"GEMINI.md"` |

`AllChannels()` returns all supported channels in projection order. `ChannelByName()` resolves a string to a concrete channel, treating empty string as Claude (backward compatibility).

### Layer 2: ChannelCompiler Interface

**Package**: `internal/materialize/compiler/`

Transforms mena artifacts for the target channel's expected format. The interface declares four methods: `CompileCommand()`, `CompileSkill()`, `CompileAgent()`, `ContextFilename()`. The compiler is injected into the materialization pipeline via `compilerForChannel()`.

| Artifact | ClaudeCompiler | GeminiCompiler |
|----------|---------------|----------------|
| Commands (dromena) | Markdown (`.md`) | TOML (`.toml`) with structured `name`/`description`/`prompt` fields |
| Skills (legomena) | Markdown with YAML frontmatter (`SKILL.md`) | Markdown with YAML frontmatter (`SKILL.md`) |
| Agents | YAML frontmatter + markdown body | YAML frontmatter + markdown body |
| Context file | `CLAUDE.md` | `GEMINI.md` |

The key compilation difference is commands: Claude Code expects raw markdown files, while Gemini CLI expects TOML with explicit fields. Skills and agents share the same YAML-frontmatter-plus-markdown format across both channels.

### Layer 3: LifecycleAdapter Interface

**Package**: `internal/hook/adapter.go`

Handles channel-specific hook payload parsing and response formatting.

```
type LifecycleAdapter interface {
    ParsePayload(reader io.Reader) (*Env, error)
    FormatResponse(decision string, reason string) ([]byte, error)
    ChannelName() string
}
```

`ClaudeAdapter` and `GeminiAdapter` normalize channel-specific wire formats to a common `Env` struct. Event name translation (CC `PreToolUse` to Gemini `BeforeTool`, etc.) happens at the adapter boundary via `TranslateInboundEvent()`, so all downstream hook commands operate on a uniform `Env` regardless of which CLI fired the event.

Gemini sends the same snake_case JSON fields as Claude Code (`session_id`, `hook_event_name`, `tool_name`, `tool_input`), so `StdinPayload` can parse both. Gemini-only fields (`timestamp`, `mcp_context`) are silently ignored by `json.Unmarshal`.

### Event and Tool Translation Tables

**Event mapping** (`internal/hook/events.go`):

| CC Canonical | Gemini Wire | Direction |
|-------------|-------------|-----------|
| `PreToolUse` | `BeforeTool` | Bidirectional |
| `PostToolUse` | `AfterTool` | Bidirectional |
| `SessionStart` | `SessionStart` | Passthrough |
| `SessionEnd` | `SessionEnd` | Passthrough |
| `UserPromptSubmit` | `BeforeAgent` | Bidirectional |
| `PreCompact` | `PreCompress` | Bidirectional |
| `Notification` | `Notification` | Passthrough |

CC events with no Gemini equivalent (e.g., `Stop`, `SubagentStop`) are silently skipped during Gemini hook registration.

**Tool name mapping** for matcher rewriting:

| CC Tool | Gemini Tool |
|---------|-------------|
| `Edit` | `replace` |
| `Write` | `write_file` |
| `Bash` | `run_shell_command` |
| `ReadFiles` | `read_file` |
| `Glob` | `glob` |
| `Grep` | `grep` |

`TranslateMatcherForChannel()` rewrites pipe-delimited matcher patterns per channel. Unknown tool names pass through unchanged (defensive).

### Dispatch Mechanism

`ari sync --channel=all` is resolved at `syncRiteScope()` to concrete channels. The value `"all"` is resolved ONLY at this boundary -- `MaterializeWithOptions()` never sees it; it always receives a single concrete channel name.

For Gemini projection, `MaterializeWithOptions()` temporarily overrides `claudeDirOverride` to point at `.gemini/` instead of `.claude/`, using a save-and-restore pattern (`defer func() { m.claudeDirOverride = originalOverride }()`) to prevent mutation leaking between channel passes. Each channel gets an independent materialization pass through the same pipeline.

### Provenance

Per-channel manifests stored in `.knossos/` (single directory, channel-keyed filenames):

| Channel | Manifest File |
|---------|---------------|
| Claude | `PROVENANCE_MANIFEST.yaml` (unchanged, backward compatible) |
| Gemini | `PROVENANCE_MANIFEST_GEMINI.yaml` |

`ManifestPathForChannel()` resolves the correct path. For Claude (or empty channel), it returns the default path -- identical to the pre-multi-channel `ManifestPath()` function, maintaining backward compatibility.

### Hook Registration

`hooks.yaml` remains CC-canonical (source of truth). `BuildHooksSettings()` accepts a `channel` parameter and translates event names and tool matcher patterns per channel via `TranslateEventForChannel()` and `TranslateMatcherForChannel()`. Events with no Gemini equivalent are silently skipped. The output is written to the channel-appropriate `settings.local.json` (`.claude/settings.local.json` or `.gemini/settings.local.json`).

## Alternatives Considered

### Option A: Channel-Specific Binary

Build separate `ari-claude` and `ari-gemini` binaries, each hardcoded to one channel.

- **Pros**: Simple per-binary; no abstraction overhead; channel-specific optimizations trivial
- **Cons**: Doubles build/release complexity; shared logic diverges over time; users must install and manage multiple binaries; hook commands need per-binary variants

Rejected: the maintenance cost compounds with each new channel. N channels means N binaries, N CI pipelines, N release artifacts.

### Option B: Go Plugin Architecture

Define channel support as `plugin.Plugin` modules loaded at runtime.

- **Pros**: Truly extensible; third parties could contribute channels without forking
- **Cons**: Go's `plugin` package is Linux-only (no macOS, no Windows); plugins must be compiled with the exact same Go version and module graph; no cross-compilation; debugging is significantly harder
- **Cons (practical)**: Knossos currently has 2 channels. Plugins solve for a scale of extensibility that doesn't exist.

Rejected: Go plugins are fragile and don't work cross-platform. The abstraction cost isn't justified by the current channel count.

### Option C: Separate Codebases Per Channel

Fork knossos into `knossos-claude` and `knossos-gemini`, sharing code via a library.

- **Pros**: Each codebase is fully optimized for its channel; no abstraction tax
- **Cons**: Divergence guaranteed within weeks; bug fixes must be ported; shared library versioning becomes its own project; users doing multi-channel must install both

Rejected: maintenance nightmare. Every shared concern (sessions, provenance, rite resolution) would need to be kept in sync manually.

### Option D: Channel-Agnostic Hook Event Names

Invent a third event naming convention (neither CC nor Gemini) that both channels translate to/from.

- **Pros**: Symmetric; neither channel is "canonical"
- **Cons**: Introduces a third naming vocabulary with no external documentation or user familiarity; all hook commands must learn a new name set; debugging requires three-way translation; CC is the dominant channel and its names are already established in hooks.yaml and hook command source

Rejected: invents unnecessary complexity. CC-canonical names are already the internal lingua franca. Adding a third convention increases the cognitive and debugging burden without practical benefit.

## Rationale

The interface-based approach (Option E, the chosen design) provides clean extensibility while keeping the codebase unified. Key reasoning:

1. **CC-canonical as lingua franca**: Claude Code was first. Its event names, tool names, and hook wire format are already embedded in hooks.yaml, hook command source, and developer mental models. Translating at the channel boundary keeps the core pipeline untouched.

2. **Interface segregation**: The three interfaces (`TargetChannel`, `ChannelCompiler`, `LifecycleAdapter`) correspond to the three actual variation points. Adding a new channel means implementing three interfaces -- mechanical work with clear test boundaries.

3. **Override-and-restore for dispatch**: The `claudeDirOverride` pattern reuses the existing materialization pipeline wholesale instead of forking it. This means new features added to the pipeline automatically apply to all channels.

4. **Backward-compatible provenance**: Claude's manifest path is unchanged. `ManifestPath()` and `ManifestPathForChannel("claude")` return the same value. Existing tooling that reads `PROVENANCE_MANIFEST.yaml` is unaffected.

## Consequences

### Positive

- Single binary serves all channels -- one build, one install, one `ari` command
- Adding a third channel (e.g., Cursor, Windsurf) is mechanical: implement three interfaces, add translation tables
- Existing Claude-only deployments are unaffected -- `--channel=claude` is the default
- Hook commands operate on a uniform `Env` regardless of source channel
- Provenance maintains per-channel integrity without cross-contamination

### Negative

- Per-channel manifests increase provenance complexity (two files to validate, channel-keyed paths)
- Test surface expands: each channel needs compilation, hook parsing, and dispatch validation (mitigated by regression tests added in Sprint 3.1)
- The `claudeDirOverride` save-and-restore pattern is a pragmatic hack -- a cleaner approach would thread the channel through the materializer constructor, but that would require a larger refactor
- Some templates still hardcode `.claude/` in their output text (cosmetic, tracked as follow-up)

### Neutral

- Event mapping maintenance: adding a new channel requires a new mapping table, but the work is mechanical (enumerate events, map names, add test cases)
- Tool name mapping is partial: only the 6 most common tools are mapped. Gemini tool names not in the table pass through unchanged, which may produce warnings but not failures
- `hooks.yaml` remains CC-canonical -- Gemini users reading the source see CC event names, not Gemini wire names. This is a documentation concern, not a runtime issue.

## Implementation Commits

| Sprint | Commit | Description |
|--------|--------|-------------|
| P1-P4 | Various | Channel abstraction stack (TargetChannel, ChannelCompiler, LifecycleAdapter interfaces) |
| 1.1 | `7379e58` | Dual-projection dispatch -- `ari sync --channel=all` materializes both `.claude/` and `.gemini/` |
| 1.2 | `be5cf2b` | Per-channel provenance manifests and W-001 normalization |
| 2.1 | `210c77d` | Gemini hook registration -- event/tool translation + GeminiAdapter fixes |
| 2.2 | -- | Inscription system verified (no changes needed -- context file routing already channel-aware) |
| 3.1 | -- | Integration tests validating end-to-end dual-channel sync |
| 3.2 | -- | This ADR |

## Related Decisions

- **ADR-0026** (Provenance): Established the `PROVENANCE_MANIFEST.yaml` format and checksum conventions that the per-channel manifest scheme extends
- **ADR-0030** (Processions): Cross-rite coordination that operates above the channel layer -- processions are channel-agnostic by design

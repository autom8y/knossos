# GEM-001: Knossos MCP Integration for Gemini Channel

## Status
Proposed

## Context
The current Knossos materialization pipeline correctly sets up Model Context Protocol (MCP) servers for Claude Code (CC) via `.mcp.json` (SCAR-028) and agent frontmatter injection. However, the **Gemini CLI** channel is currently broken because:

1.  **Configuration Location**: Gemini CLI reads MCP configuration from `.gemini/settings.json`, not `.mcp.json`. SCAR-028 explicitly strips `mcpServers` from `settings.local.json`, leaving Gemini with no configured servers.
2.  **Tool Reference Syntax**: Gemini CLI requires fully qualified tool names in the format `mcp_{server}_{tool}` (e.g., `mcp_github_create_issue`). Knossos uses the project-canonical syntax `mcp:server/tool` (e.g., `mcp:github/create_issue`), which is currently leaked into Gemini agent definitions without translation.
3.  **Agent Frontmatter**: The Gemini compiler correctly strips the `mcpServers` inline configuration key (as Gemini doesn't support it in agents), but since the global configuration is also missing (point 1), the tools are effectively undefined.

## Objective
Enable full MCP support for the Gemini channel by:
1.  Injecting resolved MCP server configurations into `.gemini/settings.local.json`.
2.  Translating `mcp:server/tool` references to `mcp_{server}_{tool}` format during agent compilation.

## Implementation Plan

### 1. Channel-Aware Settings Materialization
**File**: `internal/materialize/materialize_settings.go`

Modify `materializeSettingsWithManifest` to accept `resolvedMCPServers`.
*   **Logic**:
    *   Retain the SCAR-028 deletion of `mcpServers` as the default behavior (cleaning up stale CC configs).
    *   **IF** `channel == "gemini"`, re-inject the resolved MCP servers into `existingSettings["mcpServers"]`.
    *   Reuse the existing `mergeMCPServers` logic (via `hooks` or local helper) to format the server definitions correctly for `settings.json`.

### 2. Update Materialization Pipeline
**File**: `internal/materialize/materialize.go`

Update the call site in `MaterializeWithOptions`:
*   Pass the `resolvedMCPServers` (already computed in step 3.8) to `materializeSettingsWithManifest`.

### 3. MCP Tool Name Translation
**File**: `internal/channel/tools.go`

Update `TranslateTool` to handle the `mcp:` prefix.
*   **Logic**:
    *   Detect `mcp:` prefix.
    *   Strip `mcp:`.
    *   Replace `/` with `_`.
    *   Prepend `mcp_`.
    *   Example: `mcp:browserbase/browserbase_session_create` -> `mcp_browserbase_browserbase_session_create`.
*   **Edge Case**: `mcp:server` (no tool). Gemini likely requires explicit tool names. If encountered, we will translate to `mcp_{server}` but this may not be valid if the server name itself isn't a tool. However, for the current scope (fixing explicitly named tools like in `frontend-fanatic`), this translation is sufficient.

## Verification
1.  **Unit Tests**: Add test cases for `TranslateTool` with MCP strings.
2.  **Manual Verification**:
    *   Run `ari sync --channel gemini`.
    *   Inspect `.gemini/settings.local.json`: Verify `mcpServers` key exists and contains definitions.
    *   Inspect `.gemini/agents/frontend-fanatic.md`: Verify `tools` list contains `mcp_browserbase_browserbase_session_create`.
    *   Run `gemini list-tools` (or equivalent) to verify tool discovery (requires actual Gemini CLI runtime).

## Rollout
This change is purely internal to the materializer and affects only the `gemini` channel generation. It preserves existing behavior for the `claude` channel (reading from `.mcp.json`, using `mcp:server/tool` syntax).

# GEM-002: Channel-Agnostic MCP Parameterization

## Status
Proposed

## Context
The `browser-local` MCP pool uses hardcoded environment variables (e.g., `CUA_MODEL: "anthropic/claude-sonnet-4-6"`). To support multi-channel materialization (Gemini vs. Claude) with channel-appropriate models and credentials, these variables must be parameterized.

Users should be able to define channel-specific environment variables in their shell (e.g., `GEMINI_CUA_MODEL`, `ANTHROPIC_CUA_MODEL`) which `ari sync` detects and maps to the canonical configuration (e.g., `CUA_MODEL`) for the target channel.

## Objective
Enable `ari sync` to dynamically resolve MCP server environment variables based on the target channel.

**Mapping Logic:**
*   Channel `gemini` -> Prefix `GEMINI_`
*   Channel `claude` -> Prefix `ANTHROPIC_` (default/legacy)

**Mechanism:**
If `ari sync --channel gemini` is run, and `GEMINI_CUA_MODEL` exists in the host environment, the generated `mcpServers` config will use `CUA_MODEL: "${GEMINI_CUA_MODEL}"`.

## Implementation Plan

### 1. Update Pool Resolution Logic
**File**: `internal/materialize/hooks/mcp_pools.go`

Update `ResolvePoolServers` signature:
```go
func ResolvePoolServers(pools *MCPPoolsConfig, refs []MCPPoolRef, channel string) ([]MCPServerConfig, error)
```

**Logic**:
1.  Determine `channelPrefix`:
    *   `gemini` -> `GEMINI`
    *   `claude` (or other) -> `ANTHROPIC`
2.  Iterate over `pool.Server.Env` keys.
3.  For each key (e.g., `CUA_MODEL`), check if `channelPrefix + "_" + key` exists in `os.LookupEnv`.
4.  If found, add to an implicit merge map: `key: "${PREFIX_KEY}"`.
5.  Apply this merge map *before* the explicit `env_merge` from the rite manifest (allowing rites to still force overrides if needed).

### 2. Propagate Channel Context
**File**: `internal/materialize/mcp.go`

Update `resolveAllMCPServers` signature to accept `channel`:
```go
func resolveAllMCPServers(manifest *RiteManifest, poolsConfig *MCPPoolsConfig, channel string) ([]MCPServer, error)
```
Pass `channel` to `hooks.ResolvePoolServers`.

### 3. Update Materializer Call Sites
**File**: `internal/materialize/materialize.go`

Update `MaterializeWithOptions`:
```go
// Step 3.8
resolvedMCPServers, mcpResolveErr := resolveAllMCPServers(manifest, poolsConfig, opts.Channel)
```

### 4. Update Tests
**Files**: 
*   `internal/materialize/mcp_integration_test.go`
*   `internal/materialize/hooks/mcp_pools_test.go`

Update all test callers to pass a channel string (e.g., `"claude"` or `""`).
Add a new test case verifying that `GEMINI_CUA_MODEL` overrides `CUA_MODEL` when channel is `gemini`.

## Verification
1.  **Unit Tests**: Verify `ResolvePoolServers` correctly rewrites env values and args based on mock environment variables.
2.  **Manual Verification**:
    *   Export `GEMINI_CUA_MODEL="gemini-1.5-pro"`.
    *   Run `ari sync --channel gemini`.
    *   Inspect `.gemini/settings.local.json`: Confirm `CUA_MODEL` is `${GEMINI_CUA_MODEL}`.

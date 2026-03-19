// Package materialize re-exports MCP settings functions from the hooks sub-package.
package materialize

import (
	"github.com/autom8y/knossos/internal/materialize/hooks"
)

// mergeMCPServers converts core MCPServer types to hooks.MCPServerConfig
// and delegates to hooks.MergeMCPServers.
func mergeMCPServers(existingSettings map[string]any, mcpServers []MCPServer) map[string]any {
	hookServers := toHookServers(mcpServers)
	return hooks.MergeMCPServers(existingSettings, hookServers)
}

// toHookServers converts core MCPServer types to hooks.MCPServerConfig.
func toHookServers(mcpServers []MCPServer) []hooks.MCPServerConfig {
	hookServers := make([]hooks.MCPServerConfig, len(mcpServers))
	for i, s := range mcpServers {
		hookServers[i] = hooks.MCPServerConfig{
			Name:    s.Name,
			Command: s.Command,
			Args:    s.Args,
			Env:     s.Env,
			Type:    s.Type,
			URL:     s.URL,
			Headers: s.Headers,
		}
	}
	return hookServers
}

// fromHookServers converts hooks.MCPServerConfig types to core MCPServer.
func fromHookServers(hookServers []hooks.MCPServerConfig) []MCPServer {
	servers := make([]MCPServer, len(hookServers))
	for i, s := range hookServers {
		servers[i] = MCPServer{
			Name:    s.Name,
			Command: s.Command,
			Args:    s.Args,
			Env:     s.Env,
			Type:    s.Type,
			URL:     s.URL,
			Headers: s.Headers,
		}
	}
	return servers
}

// resolveAllMCPServers resolves pool refs + direct mcp_servers into a unified MCPServer list.
// Direct servers win on name collision (rite-specific override of pool canonical).
func resolveAllMCPServers(manifest *RiteManifest, poolsConfig *MCPPoolsConfig, channel string) ([]MCPServer, error) {
	if manifest == nil {
		return nil, nil
	}

	// Resolve pool references
	var poolServers []MCPServer
	if poolsConfig != nil && len(manifest.MCPPools) > 0 {
		resolved, err := hooks.ResolvePoolServers(poolsConfig, manifest.MCPPools, channel)
		if err != nil {
			return nil, err
		}
		poolServers = fromHookServers(resolved)
	}

	// Merge: pool servers first, then direct servers override on name collision
	if len(poolServers) == 0 {
		return manifest.MCPServers, nil
	}
	if len(manifest.MCPServers) == 0 {
		return poolServers, nil
	}

	// Build name→server map: pool servers as base, direct servers win
	seen := make(map[string]int) // name → index in result
	result := make([]MCPServer, 0, len(poolServers)+len(manifest.MCPServers))

	for _, s := range poolServers {
		seen[s.Name] = len(result)
		result = append(result, s)
	}
	for _, s := range manifest.MCPServers {
		if idx, exists := seen[s.Name]; exists {
			result[idx] = s // direct server overrides pool server
		} else {
			result = append(result, s)
		}
	}

	return result, nil
}

// Package materialize re-exports MCP settings functions from the hooks sub-package.
package materialize

import (
	"github.com/autom8y/knossos/internal/materialize/hooks"
)

// mergeMCPServers converts core MCPServer types to hooks.MCPServerConfig
// and delegates to hooks.MergeMCPServers.
func mergeMCPServers(existingSettings map[string]any, mcpServers []MCPServer) map[string]any {
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
	return hooks.MergeMCPServers(existingSettings, hookServers)
}

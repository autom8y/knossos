package agent

import (
	"fmt"
	"strings"
)

// ValidateAgentMCPReferences checks that MCP tool references in agent frontmatter
// resolve to declared MCP servers in the rite manifest.
// Returns WARNINGS (not errors) because servers may be satellite-provided.
// This is called when rite context is available (e.g., `ari agent validate --rite <rite>`).
func ValidateAgentMCPReferences(agent *AgentFrontmatter, mcpServerNames []string) []string {
	if agent == nil || len(agent.Tools) == 0 {
		return nil
	}

	// Build set of declared server names for quick lookup
	declaredServers := make(map[string]bool)
	for _, name := range mcpServerNames {
		declaredServers[name] = true
	}

	var warnings []string

	// Extract MCP servers from agent's tools
	agentMCPServers := agent.MCPServers()

	// Check each MCP server reference
	for _, serverName := range agentMCPServers {
		if !declaredServers[serverName] {
			warnings = append(warnings, fmt.Sprintf(
				"MCP server %q used by agent but not declared in rite manifest (may be satellite-provided)",
				serverName,
			))
		}
	}

	return warnings
}

// ValidateAgentMCPToolReferences is a more detailed version that returns warnings
// for each specific tool reference, not just server names.
// Useful for detailed validation output showing exact tool usage.
func ValidateAgentMCPToolReferences(agent *AgentFrontmatter, mcpServerNames []string) []string {
	if agent == nil || len(agent.Tools) == 0 {
		return nil
	}

	// Build set of declared server names
	declaredServers := make(map[string]bool)
	for _, name := range mcpServerNames {
		declaredServers[name] = true
	}

	var warnings []string

	// Check each tool that starts with "mcp:"
	for _, tool := range agent.Tools {
		if !strings.HasPrefix(tool, "mcp:") {
			continue
		}

		// Extract server name from tool reference
		// "mcp:github/create_issue" -> "github"
		// "mcp:github" -> "github"
		ref := strings.TrimPrefix(tool, "mcp:")
		parts := strings.SplitN(ref, "/", 2)
		serverName := parts[0]

		if !declaredServers[serverName] {
			warnings = append(warnings, fmt.Sprintf(
				"tool %q references MCP server %q not declared in rite manifest (may be satellite-provided)",
				tool, serverName,
			))
		}
	}

	return warnings
}

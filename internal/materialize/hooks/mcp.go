package hooks

import (
	"encoding/json"
	"os"

	"github.com/autom8y/knossos/internal/fileutil"
)

// MCPServerConfig represents an MCP server for settings merge.
// This is a local struct to avoid importing the parent materialize package.
// Core maps materialize.MCPServer to this type at the call site.
type MCPServerConfig struct {
	Name    string
	Command string
	Args    []string
	Env     map[string]string
	Type    string // stdio (default), sse, http
	URL     string
	Headers map[string]string
}

// MergeMCPServers merges MCP server declarations into existing settings.
// Uses union merge semantics:
//   - Servers are added/updated
//   - Existing satellite servers not in the list are preserved
//   - Output follows Claude Code's mcpServers format
func MergeMCPServers(existingSettings map[string]any, mcpServers []MCPServerConfig) map[string]any {
	// Ensure existingSettings has mcpServers key
	if existingSettings["mcpServers"] == nil {
		existingSettings["mcpServers"] = make(map[string]any)
	}

	// Get existing mcpServers (may be empty map)
	mcpServersMap, ok := existingSettings["mcpServers"].(map[string]any)
	if !ok {
		// If not a map, create new empty map
		mcpServersMap = make(map[string]any)
		existingSettings["mcpServers"] = mcpServersMap
	}

	// Merge servers (add/update)
	for _, server := range mcpServers {
		serverConfig := make(map[string]any)

		// Determine transport type: empty or "stdio" uses command/args;
		// "sse" and "http" use url/headers.
		transportType := server.Type
		if transportType == "sse" || transportType == "http" {
			serverConfig["type"] = transportType
			if server.URL != "" {
				serverConfig["url"] = server.URL
			}
			if len(server.Headers) > 0 {
				serverConfig["headers"] = server.Headers
			}
		} else {
			// stdio transport (default): emit command/args
			if server.Command != "" {
				serverConfig["command"] = server.Command
			}
			if len(server.Args) > 0 {
				serverConfig["args"] = server.Args
			}
		}

		if len(server.Env) > 0 {
			serverConfig["env"] = server.Env
		}

		mcpServersMap[server.Name] = serverConfig
	}

	return existingSettings
}

// LoadExistingSettings reads settings.local.json if it exists.
// Returns empty map if file doesn't exist or is invalid JSON.
func LoadExistingSettings(settingsPath string) (map[string]any, error) {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]any), nil
		}
		return nil, err
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		// Invalid JSON - return empty map (will be overwritten)
		return make(map[string]any), nil
	}

	return settings, nil
}

// SaveSettings writes settings to settings.local.json with pretty formatting.
// Only writes if content changed, to avoid triggering Claude Code file watcher.
func SaveSettings(settingsPath string, settings map[string]any) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	_, err = fileutil.WriteIfChanged(settingsPath, data, 0644)
	return err
}
